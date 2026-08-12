package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	tenantImportPath = "metaldocs/internal/platform/tenant"
	excludedFile     = "internal/platform/idempotency/identity.go"
)

type candidate struct {
	rel string
	abs string
}

type loadedFile struct {
	pkg  *packages.Package
	file *ast.File
}

type finding struct {
	path string
	line int
	col  int
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("idempotency-identity-scope-guard: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := repoRoot()
	if err != nil {
		return fmt.Errorf("find repository root: %w", err)
	}
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("change to repository root: %w", err)
	}

	candidates, err := discoverCandidates(root)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("found no tracked Go files to check (excluding vendor/ and %s); refusing to report success on an empty sweep", excludedFile)
	}

	loaded, err := loadCandidates(root, candidates)
	if err != nil {
		return err
	}

	var findings []finding
	seenFindings := map[string]bool{}
	for _, candidate := range candidates {
		loadedFiles := loaded[pathKey(candidate.abs)]
		if len(loadedFiles) == 0 {
			return fmt.Errorf("candidate %s did not receive type information under default or integration build tags", candidate.rel)
		}
		for _, loadedFile := range loadedFiles {
			for _, item := range scanLoadedFile(candidate.rel, loadedFile) {
				key := fmt.Sprintf("%s:%d:%d", item.path, item.line, item.col)
				if seenFindings[key] {
					continue
				}
				seenFindings[key] = true
				findings = append(findings, item)
			}
		}
	}

	for _, item := range findings {
		fmt.Printf("%s:%d:%d: actorFromCtx-shaped function references tenant.FromContext — outside identity.go this re-implements idempotency.TenantActorFromContext instead of calling it\n", item.path, item.line, item.col)
	}
	if len(findings) > 0 {
		fmt.Printf("idempotency-identity-scope-guard: %d violation(s) found — a second actorFromCtx-shaped resolver is calling tenant.FromContext outside identity.go\n", len(findings))
		fmt.Println("Fix: delete the hand-rolled resolver and call idempotency.TenantActorFromContext instead.")
		return fmt.Errorf("guard rejected actor identity reimplementation")
	}

	fmt.Printf("idempotency-identity-scope-guard: clean (%d Go files scanned, none outside identity.go re-implement actorFromCtx over tenant.FromContext)\n", len(candidates))
	return nil
}

func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	root, err := filepath.Abs(strings.TrimSpace(string(out)))
	if err != nil {
		return "", err
	}
	return filepath.Clean(root), nil
}

func discoverCandidates(root string) ([]candidate, error) {
	cmd := exec.Command("git", "ls-files", "-z", "--", "*.go")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list tracked Go files: %w", err)
	}

	var candidates []candidate
	canonicalTracked := false
	for _, raw := range bytes.Split(out, []byte{0}) {
		rel := filepath.ToSlash(strings.TrimSpace(string(raw)))
		if rel == "" || strings.HasPrefix(rel, "vendor/") {
			continue
		}
		if rel == excludedFile {
			canonicalTracked = true
			continue
		}
		abs, err := filepath.Abs(filepath.FromSlash(rel))
		if err != nil {
			return nil, fmt.Errorf("resolve %s: %w", rel, err)
		}
		file, err := parser.ParseFile(token.NewFileSet(), abs, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", rel, err)
		}
		if hasUsableTenantImport(file) {
			candidates = append(candidates, candidate{rel: rel, abs: abs})
		}
	}
	if !canonicalTracked {
		return nil, fmt.Errorf("canonical identity implementation %s is not tracked; refusing to treat a missing exclusion target as success", excludedFile)
	}
	return candidates, nil
}

func hasUsableTenantImport(file *ast.File) bool {
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err == nil && path == tenantImportPath && (spec.Name == nil || spec.Name.Name != "_") {
			return true
		}
	}
	return false
}

func loadCandidates(root string, candidates []candidate) (map[string][]loadedFile, error) {
	patterns := make([]string, 0, len(candidates))
	seenDirs := map[string]bool{}
	for _, candidate := range candidates {
		dir := filepath.ToSlash(filepath.Dir(candidate.rel))
		pattern := "./" + strings.TrimPrefix(dir, "./")
		if !seenDirs[pattern] {
			seenDirs[pattern] = true
			patterns = append(patterns, pattern)
		}
	}

	loaded := make(map[string][]loadedFile, len(candidates))
	for _, buildFlags := range [][]string{nil, {"-tags=integration"}} {
		cfg := &packages.Config{
			Dir:        root,
			Mode:       packages.LoadSyntax,
			Tests:      true,
			BuildFlags: buildFlags,
			Fset:       token.NewFileSet(),
		}
		pkgs, err := packages.Load(cfg, patterns...)
		if err != nil {
			return nil, fmt.Errorf("load Go packages (%v): %w", buildFlags, err)
		}
		for _, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				name := pkg.Fset.PositionFor(file.Pos(), false).Filename
				if name == "" {
					continue
				}
				key := pathKey(name)
				if !isCandidatePath(key, candidates) {
					continue
				}
				if pkg.Types == nil || pkg.TypesInfo == nil || pkg.Fset == nil {
					continue
				}
				if pkg.IllTyped || len(pkg.Errors) > 0 {
					return nil, fmt.Errorf("type-check candidate package %s failed: %v", pkg.PkgPath, pkg.Errors)
				}
				loaded[key] = append(loaded[key], loadedFile{pkg: pkg, file: file})
			}
		}
	}
	return loaded, nil
}

func isCandidatePath(key string, candidates []candidate) bool {
	for _, candidate := range candidates {
		if pathKey(candidate.abs) == key {
			return true
		}
	}
	return false
}

func scanLoadedFile(rel string, loaded loadedFile) []finding {
	info := loaded.pkg.TypesInfo
	var findings []finding
	seen := map[string]bool{}
	ast.Inspect(loaded.file, func(node ast.Node) bool {
		var fn *ast.FuncType
		var body *ast.BlockStmt
		var signature types.Type
		switch node := node.(type) {
		case *ast.FuncDecl:
			fn, body = node.Type, node.Body
			if object := info.ObjectOf(node.Name); object != nil {
				signature = object.Type()
			}
		case *ast.FuncLit:
			fn, body = node.Type, node.Body
			signature = info.TypeOf(node.Type)
		default:
			return true
		}
		if body == nil || fn == nil || !isActorFromContextSignature(signature) {
			return true
		}
		ast.Inspect(body, func(child ast.Node) bool {
			if child != body {
				switch child.(type) {
				case *ast.FuncDecl, *ast.FuncLit:
					return false
				}
			}
			ident, ok := child.(*ast.Ident)
			if !ok || !isTenantFromContextReference(ident, info) {
				return true
			}
			position := loaded.pkg.Fset.PositionFor(ident.Pos(), false)
			key := fmt.Sprintf("%s:%d:%d", rel, position.Line, position.Column)
			if !seen[key] {
				seen[key] = true
				findings = append(findings, finding{path: rel, line: position.Line, col: position.Column})
			}
			return true
		})
		return true
	})
	return findings
}

func isActorFromContextSignature(t types.Type) bool {
	sig, ok := types.Unalias(t).(*types.Signature)
	if !ok || sig.Variadic() || sig.Params().Len() != 1 || sig.Results().Len() != 3 {
		return false
	}
	return isContextType(sig.Params().At(0).Type()) &&
		isIdenticalPredeclared(sig.Results().At(0).Type(), types.String) &&
		isIdenticalPredeclared(sig.Results().At(1).Type(), types.String) &&
		isErrorType(sig.Results().At(2).Type())
}

func isContextType(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == "context" && named.Obj().Name() == "Context"
}

func isIdenticalPredeclared(t types.Type, kind types.BasicKind) bool {
	return types.Identical(types.Unalias(t), types.Typ[kind])
}

func isErrorType(t types.Type) bool {
	return types.Identical(types.Unalias(t), types.Unalias(types.Universe.Lookup("error").Type()))
}

func isTenantFromContextReference(ident *ast.Ident, info *types.Info) bool {
	obj, ok := info.ObjectOf(ident).(*types.Func)
	if !ok || obj.Name() != "FromContext" || obj.Pkg() == nil || obj.Pkg().Path() != tenantImportPath {
		return false
	}
	signature, ok := obj.Type().(*types.Signature)
	return ok && signature.Recv() == nil
}

func pathKey(path string) string {
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	path = filepath.ToSlash(filepath.Clean(path))
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}

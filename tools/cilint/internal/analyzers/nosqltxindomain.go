package analyzers

import (
	"go/ast"
	"go/token"
	"strings"
)

// nosqlTxInDomainAllow is the inline directive to suppress a finding on the
// offending line.
const nosqlTxInDomainAllow = "//cilint:allow-sqltx-in-domain"

// sqlSeamTypes is the set of database/sql types that constitute the
// persistence/transaction seam. A domain port must never name these — it speaks
// metaldocs/internal/platform/db.Tx instead (REQ-FE-1/FE-2, 2.12).
//
// Nullable scalar helpers (sql.NullString, sql.NullTime, sql.NullInt64, ...)
// are intentionally NOT in this set: they are a legitimate nullable-column
// idiom for domain row/DTO shapes, not a driver-coupling leak. The rule locks
// the *transaction* seam (its name: "no sql Tx in domain"), not all of
// database/sql.
var sqlSeamTypes = map[string]bool{
	"Tx":     true,
	"DB":     true,
	"Rows":   true,
	"Row":    true,
	"Result": true,
}

// NoSQLTxInDomain flags any reference to a database/sql persistence/transaction
// type (sql.Tx, sql.DB, sql.Rows, sql.Row, sql.Result — pointer-wrapped or not)
// in files whose path is under internal/modules/<m>/domain/.
//
// The domain layer must be driver-agnostic: it speaks db.Tx
// (metaldocs/internal/platform/db), not raw database/sql execution types
// (REQ-FE-2, 2.12 Task 2). A bare "database/sql" import used only for sql.Null*
// nullable helpers is allowed. Infrastructure adapters
// (internal/modules/<m>/infrastructure/) may continue to use these types — they
// are out of scope for this rule.
func NoSQLTxInDomain(files []string) []Finding {
	var out []Finding
	fset := token.NewFileSet()

	for _, path := range files {
		if !inDomainLayer(path) {
			continue
		}
		_, raw := parseFile(fset, path)
		if raw == nil {
			continue
		}
		f := raw.(*ast.File)
		src := readSource(path)

		// Flag any selector sql.<T> where T is a persistence/transaction seam
		// type. Catches *sql.Tx, *sql.DB, *sql.Rows, *sql.Row, and the
		// sql.Result interface (used without a pointer). sql.Null* helpers and
		// sql.ErrNoRows do not match.
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != "sql" {
				return true
			}
			if !sqlSeamTypes[sel.Sel.Name] {
				return true
			}
			pos := fset.Position(sel.Pos())
			line := getLine(src, pos.Line)
			if strings.Contains(line, nosqlTxInDomainAllow) {
				return true
			}
			out = append(out, Finding{
				Analyzer: "nosqltxindomain",
				File:     path,
				Line:     pos.Line,
				Message:  "domain layer must not name database/sql persistence type sql." + sel.Sel.Name + "; use metaldocs/internal/platform/db.Tx instead (REQ-FE-2, 2.12)",
			})
			return true
		})
	}
	return out
}

// inDomainLayer reports whether the file path is under an
// internal/modules/<module>/domain/ directory.
func inDomainLayer(path string) bool {
	slashed := strings.ReplaceAll(path, "\\", "/")
	return strings.Contains(slashed, "internal/modules/") &&
		strings.Contains(slashed, "/domain/")
}

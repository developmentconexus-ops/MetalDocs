package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// checkWireNullableOmitempty closes the gap SHAPE-NULLABLE-NOT-REQUIRED
// cannot see: that rule is purely spec-side (shape_rules.go) and is
// structurally blind to the Go code that actually marshals the wire. A
// hand-written response struct can declare `json:"x,omitempty"` on a field
// whose spec property is `nullable: true` AND in that schema's `required:`
// array — the spec says "always present, possibly null", the Go tag says
// "drop the key when nil". That mismatch is exactly the drift Part A of
// F1.2 fixed by hand (route.go QuorumM/DueInDays, instance_read.go
// CompletedAt/Decision); this rule is the static guard so it cannot return
// unnoticed.
//
// Scope: only hand-WRITTEN wire structs are checked — oapi-codegen-generated
// files (*.gen.go) already emit the correct tag from the spec by
// construction (verified: 0 drift across the 10 modules that marshal
// entirely from generated types) and are excluded, same as every other
// code-side rule in this file family. The walk is further scoped to paths
// containing an "http" path segment: that is where every hand-written wire
// struct in this repo lives (the module's http/contracts package and http
// handlers). Domain/application/infrastructure packages also declare
// json-tagged structs (outbox event payloads, idempotency envelopes,
// fanout DTOs, presence snapshots — none of them are HTTP response bodies),
// and scanning those too would multiply false positives against a set of
// property names this rule is not equipped to reason about. Narrowing to
// "http" is a deliberate precision-over-recall choice: it will not catch a
// hand-written wire struct someone places outside the http package, but
// nothing in this codebase does that today (grep-verified), and a struct
// living outside http/ is itself a code-review smell independent of this
// rule.
//
// Go↔schema resolution strategy (the hard part, stated explicitly per
// operator instruction — read this before extending the rule):
//
//  1. Reuse reachableSchemas (shape_rules.go) to get every named
//     components.schemas entry reachable from a response. For each such
//     schema, collect the JSON property names that are BOTH nullable AND
//     listed in that schema's own `required:` array (nullableRequiredResponseProps
//     below) into one FLAT, spec-wide set keyed by JSON property name (not by
//     schema+property — there is no reliable, non-hardcoded way to map a Go
//     struct's NAME to its spec schema's name: StageResponse ↔ StageSummary,
//     InstanceResponse ↔ ApprovalInstanceResponse, StageActor ↔
//     ApprovalStageActorResponse are all deliberately different on each side
//     of the wire and there is no naming convention tying them together).
//  2. Walk every hand-written struct field's `json:"..."` tag. A field is a
//     candidate violation when its JSON name is in that flat set AND its Go
//     type is a pointer (the only way a Go field can marshal an explicit
//     `null`) AND its tag carries `omitempty`.
//  3. Exempt any Go struct whose type name ends in "Request": this repo's
//     existing, spec-mirrored convention is that every request-only wire
//     type is named `...Request` (CreateRouteRequest, StageRequest,
//     SignoffDocumentRequest, ...) exactly matching the spec-side schemas
//     SHAPE-NULLABLE-NOT-REQUIRED itself exempts via requestOnly (both sides
//     name request bodies with a "Request" suffix — grep-verified across the
//     live spec). Without this exemption StageRequest.QuorumM/DueInDays
//     (legitimately `omitempty` — a client omits an optional field on the
//     way in) would false-positive on every run, because "quorum_m" and
//     "due_in_days" are also nullable-and-required on the RESPONSE-side
//     StageSummary schema and step 1's set is flat by JSON name, not scoped
//     per schema.
//
// Precision/recall trade-off (own it, do not hide it): this is a coarse,
// JSON-name-only match — it does NOT verify that the flagged Go struct is
// actually the same wire shape as the schema that made the name
// nullable-and-required. Two independent risks follow from that, and both
// are deliberately accepted in exchange for a rule that needs no per-module
// hardcoded name-mapping table:
//   - Under-flagging (recall gap): a hand-written struct outside any "http"
//     path segment, or a struct field whose Go type is not a pointer (e.g. a
//     nullable bool/string modeled as a non-pointer with a sentinel), is
//     invisible to this rule. It is honest about this rather than silently
//     resolving it wrong.
//   - Over-flagging (precision risk): if this repo ever hand-writes an
//     http-path struct with a field named e.g. "reason" or "id" that is
//     legitimately `omitempty` for a reason unrelated to any
//     nullable-and-required schema that happens to share that JSON name, it
//     will false-positive. Empirically, on the live tree today (post-F1.2
//     Part A) this does not happen: every hand-written http-path struct
//     field name that collides with the flat set is either non-pointer,
//     already un-omitempty, or on a "...Request" struct. If a genuine
//     collision appears later, the fix is to narrow the match (e.g. scope by
//     package/struct as well as name) — not to relax the rule silently.
func checkWireNullableOmitempty(specPath, modulesRoot string) ([]Violation, error) {
	data, err := os.ReadFile(specPath) // #nosec G304 -- specPath is main.go's first CLI positional argument (the openapi.yaml to lint); operator-controlled, not external input.
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	root := doc.Content[0]
	schemas := mapGet(mapGet(root, "components"), "schemas")
	paths := mapGet(root, "paths")

	respReach := reachableSchemas(paths, schemas, "responses")
	nullableRequired := nullableRequiredResponseProps(schemas, respReach)
	if len(nullableRequired) == 0 {
		return nil, nil
	}

	out := []Violation{}
	fset := token.NewFileSet()
	walkErr := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", "node_modules", "vendor", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		lower := strings.ToLower(path)
		if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") || strings.HasSuffix(lower, ".gen.go") {
			return nil
		}
		// Scope: only paths under an "http" segment carry hand-written wire
		// structs in this repo (see the rule's doc comment above).
		slashPath := filepath.ToSlash(lower)
		hasHTTPSegment := false
		for _, seg := range strings.Split(slashPath, "/") {
			if seg == "http" {
				hasHTTPSegment = true
				break
			}
		}
		if !hasHTTPSegment {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				// Request-only wire types follow this repo's own
				// "...Request" naming convention (mirrored on the spec
				// side, which SHAPE-NULLABLE-NOT-REQUIRED itself exempts
				// via requestOnly) — see the rule's doc comment step 3.
				if strings.HasSuffix(ts.Name.Name, "Request") {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok || st.Fields == nil {
					continue
				}
				for _, field := range st.Fields.List {
					if field.Tag == nil {
						continue
					}
					jsonName, hasOmitempty, ok := parseJSONTag(field.Tag.Value)
					if !ok || jsonName == "" || !hasOmitempty {
						continue
					}
					if _, isPtr := field.Type.(*ast.StarExpr); !isPtr {
						continue
					}
					schemaNames, hit := nullableRequired[jsonName]
					if !hit {
						continue
					}
					fieldLabel := "<embedded>"
					if len(field.Names) > 0 {
						fieldLabel = field.Names[0].Name
					}
					sorted := append([]string(nil), schemaNames...)
					sort.Strings(sorted)
					out = append(out, Violation{
						File: path,
						Line: fset.Position(field.Pos()).Line,
						Rule: "WIRE-NULLABLE-OMITEMPTY",
						Message: fmt.Sprintf(
							"struct %s field %s has json tag %q with omitempty, but %q is nullable-and-required on response schema(s) %s — drop omitempty so a nil value marshals as explicit null (present-and-null must not drift to optional)",
							ts.Name.Name, fieldLabel, jsonName, jsonName, strings.Join(sorted, ", "),
						),
					})
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return out, nil
}

// nullableRequiredResponseProps returns, for the set of named
// components.schemas transitively reachable from a response (respReach,
// computed by reachableSchemas), a flat map from JSON property name to the
// list of schema names on which that property is declared both `nullable:
// true` and present in the schema's own `required:` array. Flat by
// property name (not schema+property) because the Go-side walk has no
// reliable way to resolve a struct back to its schema name — see
// checkWireNullableOmitempty's doc comment for why that is an accepted,
// documented trade-off rather than an oversight.
//
// Mirrors shape_rules.go's own local-required-array precision limit: only
// the schema's DIRECT `required:` sibling is consulted, not a `required:`
// living in a separate allOf-composed subschema.
func nullableRequiredResponseProps(schemas *yaml.Node, respReach map[string]struct{}) map[string][]string {
	out := map[string][]string{}
	if schemas == nil {
		return out
	}
	for i := 0; i+1 < len(schemas.Content); i += 2 {
		nameNode, val := schemas.Content[i], schemas.Content[i+1]
		name := nameNode.Value
		if _, ok := respReach[name]; !ok {
			continue
		}
		props := mapGet(val, "properties")
		if props == nil || props.Kind != yaml.MappingNode {
			continue
		}
		required := requiredSet(mapGet(val, "required"))
		for j := 0; j+1 < len(props.Content); j += 2 {
			propKey, propVal := props.Content[j], props.Content[j+1]
			if !isNullableSchema(propVal) {
				continue
			}
			if _, ok := required[propKey.Value]; !ok {
				continue
			}
			out[propKey.Value] = append(out[propKey.Value], name)
		}
	}
	return out
}

// parseJSONTag extracts the JSON field name and whether "omitempty" is
// present from a raw (still-quoted) Go struct tag literal. ok is false if
// the tag cannot be unquoted or carries no `json:` key at all.
func parseJSONTag(rawTag string) (name string, hasOmitempty bool, ok bool) {
	tagVal, err := strconv.Unquote(rawTag)
	if err != nil {
		return "", false, false
	}
	jsonTag, present := reflect.StructTag(tagVal).Lookup("json")
	if !present {
		return "", false, false
	}
	parts := strings.Split(jsonTag, ",")
	name = parts[0]
	if name == "-" {
		return "", false, false
	}
	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			hasOmitempty = true
		}
	}
	return name, hasOmitempty, true
}

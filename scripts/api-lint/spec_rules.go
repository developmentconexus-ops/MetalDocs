package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Violation struct {
	File    string
	Line    int
	Rule    string
	Message string
}

// Paths in the OpenAPI doc exclude the server prefix (`/api/v1`), so match on
// the bare resource path.
var stateTransitionPathRE = regexp.MustCompile(`^/[^/]+/\{[^}]+\}/(publish|submit|approve|reject|archive|unarchive|signoff|supersede|obsolete|schedule-publish)$`)

func RunSpecRules(specPath string) ([]Violation, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	root := doc.Content[0]
	paths := mapGet(root, "paths")
	if paths == nil {
		return nil, fmt.Errorf("%s: missing paths", specPath)
	}
	components := mapGet(root, "components")
	schemas := mapGet(components, "schemas")
	responsesComp := mapGet(components, "responses")
	// A root-level `security:` requirement is inherited by every operation that
	// does not override it (OpenAPI 3.x semantics). When it is present and
	// non-empty, an operation is secure-by-default in-doc without restating
	// `security` itself — so AUTHZ-DRIFT must treat global security as satisfying
	// the per-op requirement (api-contract-hardening Phase E2, F-NO-GLOBAL-SEC).
	globalSecurity := mapGet(root, "security")
	hasGlobalSecurity := globalSecurity != nil && globalSecurity.Kind == yaml.SequenceNode && len(globalSecurity.Content) > 0
	out := []Violation{}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathKey, pathVal := paths.Content[i], paths.Content[i+1]
		out = append(out, checkBasePrefix(specPath, pathKey)...)
		for j := 0; j+1 < len(pathVal.Content); j += 2 {
			mKey, op := pathVal.Content[j], pathVal.Content[j+1]
			method := strings.ToUpper(mKey.Value)
			opID := scalarValue(mapGet(op, "operationId"))
			if opID == "" {
				opID = method + " " + pathKey.Value
			}
			out = append(out, checkEnvelope(specPath, opID, pathKey.Value, op, responsesComp)...)
			out = append(out, checkPagination(specPath, opID, op, schemas)...)
			out = append(out, checkAuthz(specPath, opID, pathKey.Value, method, op, hasGlobalSecurity)...)
		}
	}
	// CASING-DRIFT walks the whole document (schemas + inline path bodies) for
	// every declared `properties:` key and flags any that is not snake_case.
	out = append(out, checkCasing(specPath, root)...)
	// SHAPE-NULLABLE-NOT-REQUIRED walks the whole document (components.schemas
	// + inline body/response schemas) for every declared nullable property
	// absent from its schema's own `required:` array (AIP-134 / 9f86828b).
	out = append(out, checkShapeNullableNotRequired(specPath, root)...)
	return out, nil
}

// snakeCaseRE is the canonical snake_case identifier shape (api-design-system.md
// §2.7 mini-conventions): lowercase start, words joined by single underscores,
// digits allowed within a word.
var snakeCaseRE = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

// casingExemptKeys are schema property keys allowed to keep a non-snake_case form.
//   - RFC 9457 Problem fields (type/title/status/detail/instance/code/errors) and
//     FieldError fields (field/code/message) are already lowercase and pass the
//     regex, so they need no explicit entry.
//   - SCREAMING_SNAKE feature-flag keys intentionally mirror their env-var /
//     flag-registry constant name (env METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT)
//     and are consumed by a hand-written FE flag module by string (not the typed
//     client); renaming the wire key would diverge it from the flag constant.
var casingExemptKeys = map[string]struct{}{
	"MDDM_NATIVE_EXPORT_ROLLOUT_PCT": {},
}

// checkCasing enforces snake_case for every declared schema property key (wherever
// a `properties:` map appears — components/schemas and inline request/response
// bodies) AND every query/path parameter name (wherever a `parameters:` sequence
// appears — path-item and operation level). It does NOT walk free-form object
// content (additionalProperties values, examples) — only declared property names —
// and skips header parameters, whose canonical form is kebab/Pascal (Content-Type,
// X-Trace-Id), not snake_case. Blocking-by-default (api-contract-hardening Phase
// E1): a re-introduced camelCase/PascalCase property or query/path param turns CI red.
func checkCasing(file string, root *yaml.Node) []Violation {
	out := []Violation{}
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case yaml.MappingNode:
			for i := 0; i+1 < len(n.Content); i += 2 {
				k, v := n.Content[i], n.Content[i+1]
				if k.Value == "properties" && v.Kind == yaml.MappingNode {
					for j := 0; j+1 < len(v.Content); j += 2 {
						propKey := v.Content[j]
						name := propKey.Value
						if _, ok := casingExemptKeys[name]; !ok && !snakeCaseRE.MatchString(name) {
							out = append(out, Violation{File: file, Line: propKey.Line, Rule: "CASING-DRIFT", Message: fmt.Sprintf("schema property %q is not snake_case", name)})
						}
						walk(v.Content[j+1]) // recurse into the property's own subtree
					}
					continue
				}
				// A `parameters:` sequence is an OpenAPI parameter list (path-item or
				// operation level); components/parameters is a MAP, so the SequenceNode
				// guard scopes this to real parameter lists only.
				if k.Value == "parameters" && v.Kind == yaml.SequenceNode {
					for _, item := range v.Content {
						out = append(out, checkParamCasing(file, item)...)
						walk(item) // recurse into the param's own subtree (e.g. schema)
					}
					continue
				}
				walk(v)
			}
		case yaml.SequenceNode:
			for _, c := range n.Content {
				walk(c)
			}
		}
	}
	walk(root)
	return out
}

// checkParamCasing flags a single query/path parameter whose name is not
// snake_case. Header parameters are skipped (kebab/Pascal is canonical for
// headers). A param given only as a `$ref` carries no inline name and is left to
// the referenced component (and redocly) to validate.
func checkParamCasing(file string, param *yaml.Node) []Violation {
	if param == nil || param.Kind != yaml.MappingNode {
		return nil
	}
	in := strings.ToLower(strings.TrimSpace(scalarValue(mapGet(param, "in"))))
	if in != "query" && in != "path" {
		return nil
	}
	nameNode := mapGet(param, "name")
	name := scalarValue(nameNode)
	if name == "" || snakeCaseRE.MatchString(name) {
		return nil
	}
	return []Violation{{File: file, Line: nameNode.Line, Rule: "CASING-DRIFT", Message: fmt.Sprintf("%s parameter %q is not snake_case", in, name)}}
}

// checkBasePrefix enforces AD-1: servers.url already carries the `/api/v1`
// base, so every path key must be relative. A key that re-includes `/api/`
// produces doubled URLs (`/api/v1/api/v1/...`) for any spec-respecting
// consumer. This is the gate that makes the double-prefix bug class
// impossible to reintroduce.
func checkBasePrefix(file string, pathKey *yaml.Node) []Violation {
	if strings.HasPrefix(pathKey.Value, "/api/") {
		return []Violation{{
			File:    file,
			Line:    pathKey.Line,
			Rule:    "PATH-BASE-PREFIX",
			Message: fmt.Sprintf("path key %q must be relative to servers.url (/api/v1); drop the /api prefix", pathKey.Value),
		}}
	}
	return nil
}

// checkEnvelope enforces AD-2: RFC 9457 Problem is the only error shape. A
// response >= 400 that declares a body MUST resolve to the Problem schema —
// either inline, or (canonically, api-contract-hardening Phase D) through a
// $ref to one of the shared #/components/responses/* definitions, which all
// reference Problem. Two principled carve-outs keep the gate honest without
// pulling in Phase E spec-hygiene scope:
//   - A description-only response (no `content`) advertises no body, so there
//     is no envelope to drift. Documenting a body for every bare status is a
//     Phase E hygiene task, not an AD-2 violation.
//   - The health/liveness probes return their health document (HealthResponse)
//     on 503 by convention, not an RFC 9457 error — exempt by path.
func checkEnvelope(file, opID, path string, op, responsesComp *yaml.Node) []Violation {
	if strings.HasPrefix(path, "/health/") {
		return nil
	}
	out := []Violation{}
	responses := mapGet(op, "responses")
	for i := 0; responses != nil && i+1 < len(responses.Content); i += 2 {
		statusNode, resp := responses.Content[i], responses.Content[i+1]
		code, err := strconv.Atoi(statusNode.Value)
		if err != nil || code < 400 || code > 599 {
			continue
		}
		// A bespoke (non-Problem) error body may opt out with an explicit,
		// reviewed reason — mirroring x-pagination-exempt. The marker is read off
		// the operation's OWN response node (not the resolved shared component),
		// so an exemption can never silently blanket every op that $refs a shared
		// response; it scopes to the one reviewed endpoint.
		if reason := strings.TrimSpace(scalarValue(mapGet(resp, "x-error-envelope-exempt"))); reason != "" {
			continue
		}
		resolved := resp
		if ref := mapGet(resp, "$ref"); ref != nil {
			resolved = lookupResponseRef(ref.Value, responsesComp)
			if resolved == nil {
				// An error response whose $ref does not resolve to a known shared
				// response cannot be proven to carry Problem. A BLOCKING gate must
				// fail closed here, not degrade to the contentless-exempt path
				// (a typo'd ref would otherwise slip a non-Problem body through).
				out = append(out, Violation{File: file, Line: statusNode.Line, Rule: "ENVELOPE-DRIFT", Message: fmt.Sprintf("response %s %s has an unresolved $ref %q; cannot verify Problem envelope", opID, statusNode.Value, scalarValue(mapGet(resp, "$ref")))})
				continue
			}
		}
		content := mapGet(resolved, "content")
		if content == nil {
			// Description-only error response: no body, no envelope to drift.
			continue
		}
		media := mapGet(content, "application/problem+json")
		if media == nil {
			media = mapGet(content, "application/json")
		}
		if !isProblemSchema(media) {
			out = append(out, Violation{File: file, Line: statusNode.Line, Rule: "ENVELOPE-DRIFT", Message: fmt.Sprintf("response %s %s does not reference Problem", opID, statusNode.Value)})
		}
	}
	return out
}

// lookupResponseRef resolves a response-level $ref into components/responses.
// Returns nil for a non-local or unknown ref; redocly lint independently gates
// dangling refs, so an unresolved ref degrades to the contentless (exempt) path
// rather than a false ENVELOPE-DRIFT hit.
func lookupResponseRef(ref string, responsesComp *yaml.Node) *yaml.Node {
	const p = "#/components/responses/"
	if !strings.HasPrefix(ref, p) {
		return nil
	}
	return mapGet(responsesComp, strings.TrimPrefix(ref, p))
}

func checkPagination(file, opID string, op, schemas *yaml.Node) []Violation {
	if !strings.HasPrefix(strings.ToLower(opID), "list") {
		return nil
	}
	// A genuinely bounded list (a small, non-growing per-tenant configuration set)
	// may opt out of cursor/limit pagination with an explicit, reviewed marker
	// rather than declaring pagination its runtime does not implement (which would
	// be spec-fiction drift). Requires a non-empty x-pagination-exempt-reason so
	// the exemption is documented, not silent (ADR 0022 Phase 11 F4).
	if strings.EqualFold(scalarValue(mapGet(op, "x-pagination-exempt")), "true") {
		if reason := strings.TrimSpace(scalarValue(mapGet(op, "x-pagination-exempt-reason"))); reason != "" {
			return nil
		}
		return []Violation{{File: file, Line: op.Line, Rule: "PAGINATION-DRIFT", Message: fmt.Sprintf("list op %s sets x-pagination-exempt without a non-empty x-pagination-exempt-reason", opID)}}
	}
	out := []Violation{}
	params := map[string]bool{}
	pNode := mapGet(op, "parameters")
	for i := 0; pNode != nil && i < len(pNode.Content); i++ {
		p := pNode.Content[i]
		if scalarValue(mapGet(p, "in")) != "query" {
			continue
		}
		n := scalarValue(mapGet(p, "name"))
		required := strings.EqualFold(scalarValue(mapGet(p, "required")), "true")
		t := scalarValue(mapGet(mapGet(p, "schema"), "type"))
		if n == "cursor" && t == "string" && !required {
			params["cursor"] = true
		}
		if n == "limit" && t == "integer" && !required {
			params["limit"] = true
		}
	}
	if !params["cursor"] {
		out = append(out, Violation{File: file, Line: op.Line, Rule: "PAGINATION-DRIFT", Message: fmt.Sprintf("list op %s missing query param cursor", opID)})
	}
	if !params["limit"] {
		out = append(out, Violation{File: file, Line: op.Line, Rule: "PAGINATION-DRIFT", Message: fmt.Sprintf("list op %s missing query param limit", opID)})
	}
	resp200 := mapGet(mapGet(op, "responses"), "200")
	respSchema := responseSchema(resp200, schemas)
	page := getProperty(respSchema, "page", schemas)
	if page == nil {
		out = append(out, Violation{File: file, Line: op.Line, Rule: "PAGINATION-DRIFT", Message: fmt.Sprintf("list op %s response missing page.next_cursor", opID)})
		out = append(out, Violation{File: file, Line: op.Line, Rule: "PAGINATION-DRIFT", Message: fmt.Sprintf("list op %s response missing page.has_more", opID)})
		return out
	}
	if getProperty(page, "next_cursor", schemas) == nil {
		out = append(out, Violation{File: file, Line: page.Line, Rule: "PAGINATION-DRIFT", Message: fmt.Sprintf("list op %s response missing page.next_cursor", opID)})
	}
	if getProperty(page, "has_more", schemas) == nil {
		out = append(out, Violation{File: file, Line: page.Line, Rule: "PAGINATION-DRIFT", Message: fmt.Sprintf("list op %s response missing page.has_more", opID)})
	}
	return out
}

// checkAuthz enforces two AUTHZ-DRIFT guarantees:
//
//  1. Every op is secure-in-doc — it declares its own `security` (including an
//     explicit `security: []` public opt-out) OR inherits a non-empty global
//     `security` requirement.
//  2. The authz-area posture markers are honest (api-contract-hardening Phase F,
//     FD-1). The pre-F negative `x-authz-skip-area` escape hatch is GONE; an
//     area-scoped op now carries a positive marker that says where the enforced
//     area comes from:
//     - `x-authz-area: {source: tx, derived_from: "<table>.<column>"}` — area
//     loaded from the DB row inside the tx (un-spoofable).
//     - `x-authz-area: {source: body|path, field: <name>}` — area is the
//     request-supplied action target (e.g. grant membership in area X).
//     - `x-authz-area-none: "<reason>"` — genuinely area-less (tenant-global).
//     - `x-authz-custom: true` — bespoke handler-level check.
//     Every state-transition POST must carry exactly one of those.
func checkAuthz(file, opID, path, method string, op *yaml.Node, hasGlobalSecurity bool) []Violation {
	out := []Violation{}
	if mapGet(op, "security") == nil && !hasGlobalSecurity {
		out = append(out, Violation{File: file, Line: op.Line, Rule: "AUTHZ-DRIFT", Message: fmt.Sprintf("op %s missing security declaration", opID)})
	}

	area := mapGet(op, "x-authz-area")
	none := strings.TrimSpace(scalarValue(mapGet(op, "x-authz-area-none")))
	hasCustom := strings.EqualFold(scalarValue(mapGet(op, "x-authz-custom")), "true")

	// Validate the positive area marker's shape wherever it appears (not just on
	// state-transition POSTs): an x-authz-area that omits its provenance is drift.
	if area != nil {
		out = append(out, checkAuthzAreaShape(file, opID, area)...)
	}

	if method == "POST" && stateTransitionPathRE.MatchString(path) {
		if area == nil && none == "" && !hasCustom {
			out = append(out, Violation{File: file, Line: op.Line, Rule: "AUTHZ-DRIFT", Message: fmt.Sprintf("state-transition op %s missing x-authz-area / x-authz-area-none / x-authz-custom", opID)})
		}
	}
	return out
}

// checkAuthzAreaShape validates that an x-authz-area marker honestly declares
// where the enforced area is resolved from: source:tx needs derived_from;
// source:body|path needs field.
func checkAuthzAreaShape(file, opID string, area *yaml.Node) []Violation {
	source := strings.ToLower(strings.TrimSpace(scalarValue(mapGet(area, "source"))))
	switch source {
	case "tx":
		if strings.TrimSpace(scalarValue(mapGet(area, "derived_from"))) == "" {
			return []Violation{{File: file, Line: area.Line, Rule: "AUTHZ-DRIFT", Message: fmt.Sprintf("op %s x-authz-area source:tx requires a non-empty derived_from", opID)}}
		}
	case "body", "path":
		if strings.TrimSpace(scalarValue(mapGet(area, "field"))) == "" {
			return []Violation{{File: file, Line: area.Line, Rule: "AUTHZ-DRIFT", Message: fmt.Sprintf("op %s x-authz-area source:%s requires a non-empty field", opID, source)}}
		}
	default:
		return []Violation{{File: file, Line: area.Line, Rule: "AUTHZ-DRIFT", Message: fmt.Sprintf("op %s x-authz-area has invalid source %q (want tx|body|path)", opID, source)}}
	}
	return nil
}

func isProblemSchema(media *yaml.Node) bool {
	schema := mapGet(media, "schema")
	if schema == nil {
		return false
	}
	return scalarValue(mapGet(schema, "$ref")) == "#/components/schemas/Problem"
}

func responseSchema(resp, schemas *yaml.Node) *yaml.Node {
	content := mapGet(resp, "content")
	schema := mapGet(mapGet(content, "application/json"), "schema")
	if schema == nil {
		schema = mapGet(mapGet(content, "application/problem+json"), "schema")
	}
	return derefSchema(schema, schemas)
}

func mapGet(n *yaml.Node, key string) *yaml.Node {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}

func scalarValue(n *yaml.Node) string {
	if n == nil {
		return ""
	}
	return n.Value
}

func derefSchema(n, schemas *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	if n.Kind == yaml.MappingNode {
		if ref := mapGet(n, "$ref"); ref != nil {
			return lookupRef(ref.Value, schemas)
		}
		return n
	}
	return lookupRef(n.Value, schemas)
}

func lookupRef(ref string, schemas *yaml.Node) *yaml.Node {
	const p = "#/components/schemas/"
	if !strings.HasPrefix(ref, p) {
		return nil
	}
	return mapGet(schemas, strings.TrimPrefix(ref, p))
}

func getProperty(schema *yaml.Node, name string, schemas *yaml.Node) *yaml.Node {
	s := derefSchema(schema, schemas)
	for depth := 0; s != nil && depth <= 2; depth++ {
		props := mapGet(s, "properties")
		if p := mapGet(props, name); p != nil {
			return derefSchema(p, schemas)
		}
		s = derefSchema(mapGet(s, "$ref"), schemas)
	}
	return nil
}

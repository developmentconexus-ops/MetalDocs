package main

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// checkShapeNullableNotRequired enforces the AIP-134 field-behavior contract
// (the 9f86828b bug class): a schema property that is declared *nullable*
// but is absent from that schema's own `required:` array lets a code
// generator emit `field?: T | null`, so a present-and-null value on the wire
// is indistinguishable from an absent key on the consumer side. Forcing the
// key into `required` keeps the key always-present (the value itself may
// still be null), closing that drift vector.
//
// "Nullable" matches either OpenAPI style so the rule keeps working if the
// spec is ever bumped to 3.1:
//   - 3.0: `nullable: true` on the property's own schema node.
//   - 3.1: `type: ["<t>", "null"]` — a sequence-node `type:` whose scalar
//     items include the literal "null".
//
// Scope: this walks every declared `properties:` map in the document (both
// components/schemas and inline request/response bodies), reusing the same
// traversal shape as checkCasing so inline schemas are covered without a
// second walk. For each properties map it also looks at the *sibling*
// `required:` array in the same mapping node (the schema that directly owns
// the properties map). A schema with a `properties:` map and NO `required:`
// key at all is not a special case: every nullable prop in it is a
// violation by construction (a missing required array behaves as an empty
// one) — that omission is exactly the bug class this rule exists to catch.
//
// Known, documented limitation (err toward NOT false-positiving): when a
// schema uses `allOf` composition and the `required` array for a given
// property actually lives in a *sibling* subschema of the allOf list rather
// than in the same mapping node as `properties:`, this rule cannot resolve
// that cross-subschema relationship with a purely local, single-pass walk.
// Rather than risk a false positive on a legitimately-required allOf-composed
// property, the rule only ever consults the `required:` array that is a
// direct sibling of the `properties:` map it is walking. If the live spec
// grows a genuine allOf/nullable/required-elsewhere case, that is a bounded
// defer for a follow-up, not a reason to relax this rule's local precision.
//
// Response-only scope (operator decision, F1.2 Part B): the invariant this
// rule protects — "present-and-null must not collapse into optional" — is a
// GENERATED CONSUMER shape concern. Consumers (the TS/Go client code the
// spec generates) only ever deserialize response bodies; a request body is
// authored by the caller, and legitimate PATCH/upsert semantics deliberately
// use "optional + nullable" (omit the key to leave unchanged, send explicit
// null to clear). Forcing request-body nullable properties into `required`
// would make every optional-clear field mandatory on the wire — an
// oasdiff-breaking change with no corresponding consumer-safety benefit.
// So: schemas (named or inline) reachable ONLY from a requestBody are
// exempt; anything reachable from a response is still checked. If a schema
// is reachable from BOTH (e.g. a shared component used as both a request
// and a response body), response integrity wins and it stays checked — no
// such shared schema exists in the live spec today, but the rule must still
// be well-defined for that case.
func checkShapeNullableNotRequired(file string, root *yaml.Node) []Violation {
	schemas := mapGet(mapGet(root, "components"), "schemas")
	paths := mapGet(root, "paths")

	reqReach := reachableSchemas(paths, schemas, "requestBody")
	respReach := reachableSchemas(paths, schemas, "responses")
	requestOnly := map[string]struct{}{}
	for name := range reqReach {
		if _, inResp := respReach[name]; !inResp {
			requestOnly[name] = struct{}{}
		}
	}

	out := []Violation{}
	var walkSchema func(n *yaml.Node, schemaName string)
	walkSchema = func(n *yaml.Node, schemaName string) {
		if n == nil {
			return
		}
		switch n.Kind {
		case yaml.MappingNode:
			// If this mapping node owns a `properties:` map, this IS a
			// schema-with-properties node: resolve its own `required:`
			// sibling (if any) and check every declared property.
			if propsNode := mapGet(n, "properties"); propsNode != nil && propsNode.Kind == yaml.MappingNode {
				required := requiredSet(mapGet(n, "required"))
				name := schemaName
				for j := 0; j+1 < len(propsNode.Content); j += 2 {
					propKey, propVal := propsNode.Content[j], propsNode.Content[j+1]
					propName := propKey.Value
					if isNullableSchema(propVal) {
						if _, ok := required[propName]; !ok {
							label := name
							if label == "" {
								label = "inline"
							}
							out = append(out, Violation{
								File: file,
								Line: propKey.Line,
								Rule: "SHAPE-NULLABLE-NOT-REQUIRED",
								Message: fmt.Sprintf(
									"schema %s property %q is nullable but not in required (present-and-null drifts to optional)",
									label, propName,
								),
							})
						}
					}
					// Recurse into the property's own subtree (nested
					// objects / arrays of objects may declare their own
					// properties maps) — always inline from here down.
					walkSchema(propVal, "")
				}
			}
			// Continue walking every other key of this mapping node,
			// tracking the components.schemas.<Name> boundary so nested
			// schema definitions get their proper name.
			for i := 0; i+1 < len(n.Content); i += 2 {
				k, v := n.Content[i], n.Content[i+1]
				if k.Value == "properties" {
					continue // already handled above
				}
				if k.Value == "schemas" && v.Kind == yaml.MappingNode && schemaName == "" {
					// This is components.schemas: each entry is a
					// top-level named schema. Skip request-only schemas
					// entirely — no violations from their subtree.
					for si := 0; si+1 < len(v.Content); si += 2 {
						sName, sVal := v.Content[si], v.Content[si+1]
						if _, exempt := requestOnly[sName.Value]; exempt {
							continue
						}
						walkSchema(sVal, sName.Value)
					}
					continue
				}
				walkSchema(v, "")
			}
		case yaml.SequenceNode:
			for _, c := range n.Content {
				walkSchema(c, "")
			}
		}
	}

	// Walk components.schemas (via the root's "components" key handling
	// above) plus every path/operation, but SKIP inline schemas lexically
	// under a `requestBody:` key — those have no component name to look up
	// in requestOnly, so the exemption must be applied structurally as we
	// descend `paths.*.*`. `responses` subtrees are always walked.
	var walkDoc func(n *yaml.Node, schemaName string)
	walkDoc = func(n *yaml.Node, schemaName string) {
		if n == nil {
			return
		}
		switch n.Kind {
		case yaml.MappingNode:
			for i := 0; i+1 < len(n.Content); i += 2 {
				k, v := n.Content[i], n.Content[i+1]
				if k.Value == "requestBody" {
					// Everything below requestBody is request-side: skip
					// entirely (inline schemas here are always exempt).
					continue
				}
				if k.Value == "paths" {
					walkDoc(v, "")
					continue
				}
				if k.Value == "components" {
					walkSchema(v, "")
					continue
				}
				walkDoc(v, schemaName)
			}
		case yaml.SequenceNode:
			for _, c := range n.Content {
				walkDoc(c, schemaName)
			}
		}
	}
	walkDoc(root, "")
	return out
}

// reachableSchemas returns the set of components.schemas names transitively
// reachable from `rootKey` ("requestBody" or "responses") anywhere under
// `paths`, following $ref and recursing into properties/items/allOf/oneOf/
// anyOf/additionalProperties. Used to compute reqReach and respReach.
func reachableSchemas(paths, schemas *yaml.Node, rootKey string) map[string]struct{} {
	visited := map[string]struct{}{}
	var visitRef func(name string)
	var visitNode func(n *yaml.Node)

	visitNode = func(n *yaml.Node) {
		if n == nil || n.Kind != yaml.MappingNode {
			return
		}
		if ref := mapGet(n, "$ref"); ref != nil {
			const p = "#/components/schemas/"
			if strings.HasPrefix(ref.Value, p) {
				visitRef(strings.TrimPrefix(ref.Value, p))
			}
			return
		}
		for _, key := range []string{"properties"} {
			if props := mapGet(n, key); props != nil && props.Kind == yaml.MappingNode {
				for i := 1; i < len(props.Content); i += 2 {
					visitNode(props.Content[i])
				}
			}
		}
		if items := mapGet(n, "items"); items != nil {
			visitNode(items)
		}
		if ap := mapGet(n, "additionalProperties"); ap != nil && ap.Kind == yaml.MappingNode {
			visitNode(ap)
		}
		for _, key := range []string{"allOf", "oneOf", "anyOf"} {
			if seq := mapGet(n, key); seq != nil && seq.Kind == yaml.SequenceNode {
				for _, item := range seq.Content {
					visitNode(item)
				}
			}
		}
	}
	visitRef = func(name string) {
		if _, ok := visited[name]; ok {
			return
		}
		visited[name] = struct{}{}
		visitNode(mapGet(schemas, name))
	}

	// Walk every operation under every path, collecting all `rootKey`
	// subtrees ($ref'd or inline) and seeding the BFS/DFS from each.
	var walkOps func(n *yaml.Node)
	walkOps = func(n *yaml.Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case yaml.MappingNode:
			for i := 0; i+1 < len(n.Content); i += 2 {
				k, v := n.Content[i], n.Content[i+1]
				if k.Value == rootKey {
					// The rootKey subtree (requestBody or responses) has
					// intermediate wrapper structure (content -> media-type
					// -> schema, or status-code -> response -> content ->
					// ...) before any actual schema node appears, so seed
					// from every schema-shaped node found anywhere inside it.
					seedAll(v, visitNode)
					continue
				}
				walkOps(v)
			}
		case yaml.SequenceNode:
			for _, c := range n.Content {
				walkOps(c)
			}
		}
	}
	walkOps(paths)
	return visited
}

// seedAll walks an arbitrary subtree (e.g. the whole `responses:` map, which
// has intermediate structure like status-code -> content -> media-type ->
// schema that visitNode's schema-shaped walk does not itself descend into)
// and calls visit on every node that looks like a schema object (has a
// `$ref` or a `properties`/`items`/`allOf`/`oneOf`/`anyOf` key), so refs
// nested arbitrarily deep under requestBody/responses are all found.
func seedAll(n *yaml.Node, visit func(*yaml.Node)) {
	if n == nil {
		return
	}
	switch n.Kind {
	case yaml.MappingNode:
		if mapGet(n, "$ref") != nil {
			visit(n)
			return
		}
		isSchemaShaped := false
		for _, key := range []string{"properties", "items", "allOf", "oneOf", "anyOf", "additionalProperties"} {
			if mapGet(n, key) != nil {
				isSchemaShaped = true
				break
			}
		}
		if isSchemaShaped {
			visit(n)
		}
		for i := 1; i < len(n.Content); i += 2 {
			seedAll(n.Content[i], visit)
		}
	case yaml.SequenceNode:
		for _, c := range n.Content {
			seedAll(c, visit)
		}
	}
}

// requiredSet converts a `required:` sequence node into a lookup set. A nil
// node (no `required:` key present) yields an empty set, which is exactly
// the "everything nullable is a violation" behavior the rule needs.
func requiredSet(n *yaml.Node) map[string]struct{} {
	set := map[string]struct{}{}
	if n == nil || n.Kind != yaml.SequenceNode {
		return set
	}
	for _, item := range n.Content {
		set[item.Value] = struct{}{}
	}
	return set
}

// isNullableSchema reports whether a property's own schema node declares
// itself nullable under either OpenAPI 3.0 (`nullable: true`) or 3.1
// (`type: ["<t>", "null"]`) conventions. A `$ref` node is not resolved here
// — a nullable marker on a $ref'd schema belongs to the referenced schema's
// own property declaration (already walked independently as its own named
// schema), not to the referencing site.
func isNullableSchema(n *yaml.Node) bool {
	if n == nil || n.Kind != yaml.MappingNode {
		return false
	}
	if scalarValue(mapGet(n, "nullable")) == "true" {
		return true
	}
	typeNode := mapGet(n, "type")
	if typeNode != nil && typeNode.Kind == yaml.SequenceNode {
		for _, item := range typeNode.Content {
			if item.Value == "null" {
				return true
			}
		}
	}
	return false
}

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
var stateTransitionPathRE = regexp.MustCompile(`^/[^/]+/\{[^}]+\}/(publish|submit|approve|reject|archive|unarchive)$`)

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
	out := []Violation{}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathKey, pathVal := paths.Content[i], paths.Content[i+1]
		for j := 0; j+1 < len(pathVal.Content); j += 2 {
			mKey, op := pathVal.Content[j], pathVal.Content[j+1]
			method := strings.ToUpper(mKey.Value)
			opID := scalarValue(mapGet(op, "operationId"))
			if opID == "" {
				opID = method + " " + pathKey.Value
			}
			out = append(out, checkEnvelope(specPath, opID, op)...)
			out = append(out, checkPagination(specPath, opID, op, schemas)...)
			out = append(out, checkAuthz(specPath, opID, pathKey.Value, method, op)...)
		}
	}
	return out, nil
}

func checkEnvelope(file, opID string, op *yaml.Node) []Violation {
	out := []Violation{}
	responses := mapGet(op, "responses")
	for i := 0; responses != nil && i+1 < len(responses.Content); i += 2 {
		statusNode, resp := responses.Content[i], responses.Content[i+1]
		code, err := strconv.Atoi(statusNode.Value)
		if err != nil || code < 400 || code > 599 {
			continue
		}
		content := mapGet(resp, "content")
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

func checkPagination(file, opID string, op, schemas *yaml.Node) []Violation {
	if !strings.HasPrefix(strings.ToLower(opID), "list") {
		return nil
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

func checkAuthz(file, opID, path, method string, op *yaml.Node) []Violation {
	out := []Violation{}
	if mapGet(op, "security") == nil {
		out = append(out, Violation{File: file, Line: op.Line, Rule: "AUTHZ-DRIFT", Message: fmt.Sprintf("op %s missing security declaration", opID)})
	}
	if method == "POST" && stateTransitionPathRE.MatchString(path) {
		hasArea := mapGet(op, "x-authz-area") != nil
		skip := mapGet(op, "x-authz-skip-area")
		hasSkip := skip != nil && strings.TrimSpace(skip.Value) != ""
		hasCustom := strings.EqualFold(scalarValue(mapGet(op, "x-authz-custom")), "true")
		if !hasArea && !hasSkip && !hasCustom {
			out = append(out, Violation{File: file, Line: op.Line, Rule: "AUTHZ-DRIFT", Message: fmt.Sprintf("state-transition op %s missing x-authz-area / x-authz-skip-area / x-authz-custom", opID)})
		}
	}
	return out
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

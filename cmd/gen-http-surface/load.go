package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// httpMethods are the only path-item keys treated as operations; everything
// else under a path item (parameters, summary, description, ...) is not a
// method and is skipped.
var httpMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true, "trace": true,
}

// rawDocument is the subset of an OpenAPI 3 document LoadDocument reads.
type rawDocument struct {
	Servers []struct {
		URL string `yaml:"url"`
	} `yaml:"servers"`
	Paths map[string]map[string]rawOperation `yaml:"paths"`
}

// rawOperation captures exactly the fields Operation is flattened from, plus
// the security override, which needs presence-vs-absence (not just
// zero-value) to tell "no override" from "security: []".
type rawOperation struct {
	OperationID           string                 `yaml:"operationId"`
	Tags                  []string               `yaml:"tags"`
	Security              *[]map[string][]string `yaml:"security"`
	CapabilityXAuthz      string                 `yaml:"x-authz-capability"`
	CapabilityNoneXAuthz  string                 `yaml:"x-authz-capability-none"`
	PasswordChangeAllowed bool                   `yaml:"x-authz-password-change-allowed"`
}

// LoadDocument reads one OpenAPI spec file and flattens it into the shape
// Validate operates on. The server base is read from servers[0].url, never
// assumed: a trailing slash is normalized away and "/" maps to "" so an
// empty base and a root base compare equal.
func LoadDocument(path string) (Document, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is a main.go CLI flag (publicPath/e2ePath) pointing at the OpenAPI spec; operator-controlled, not external input.
	if err != nil {
		return Document{}, fmt.Errorf("%s: %w", path, err)
	}

	var raw rawDocument
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Document{}, fmt.Errorf("%s: %w", path, err)
	}

	base := ""
	if len(raw.Servers) > 0 {
		base = strings.TrimSuffix(raw.Servers[0].URL, "/")
	}

	doc := Document{Path: path, ServerBase: base}

	// Sort path keys so the emitted operation order is deterministic across
	// runs; map iteration order is not.
	paths := make([]string, 0, len(raw.Paths))
	for p := range raw.Paths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		methods := make([]string, 0, len(raw.Paths[p]))
		for m := range raw.Paths[p] {
			if httpMethods[strings.ToLower(m)] {
				methods = append(methods, m)
			}
		}
		sort.Strings(methods)

		for _, m := range methods {
			op := raw.Paths[p][m]
			doc.Operations = append(doc.Operations, Operation{
				ID:                    op.OperationID,
				Method:                strings.ToUpper(m),
				Path:                  p,
				Tag:                   strings.Join(op.Tags, ","),
				Capability:            op.CapabilityXAuthz,
				CapabilityNone:        op.CapabilityNoneXAuthz,
				Public:                op.Security != nil && len(*op.Security) == 0,
				PasswordChangeAllowed: op.PasswordChangeAllowed,
			})
		}
	}

	return doc, nil
}

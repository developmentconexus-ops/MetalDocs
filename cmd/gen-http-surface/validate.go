package main

import (
	"fmt"
	"strings"
)

// Operation is one spec operation with its authz declaration flattened out of
// the YAML. Exactly one of Capability / CapabilityNone / Public is set on a
// well-formed operation; Validate is what makes that true.
type Operation struct {
	ID                    string
	Method                string // upper-case HTTP method
	Path                  string // spec-relative, e.g. "/documents/{id}"
	Tag                   string // comma-joined when the operation declares more than one (rule 5)
	Capability            string // x-authz-capability, the wire string
	CapabilityNone        string // x-authz-capability-none, the reason
	Public                bool   // security: [] on the operation
	PasswordChangeAllowed bool   // x-authz-password-change-allowed
}

// Document is one spec file plus the server base declared inside it. The base
// is read from the document and never assumed: openapi.yaml declares /api/v1,
// internal-e2e.yaml declares an empty base because its routes mount at the
// server root.
type Document struct {
	Path       string
	ServerBase string
	Operations []Operation
}

// Key is the mux pattern the operation will be registered under. It mirrors
// oapi-codegen's HandlerWithOptions, which registers
// http.MethodGet + " " + options.BaseURL + path.
func Key(d Document, op Operation) string {
	return op.Method + " " + d.ServerBase + op.Path
}

// Validate runs the six declaration rules over every document in one pass.
// Rules 1-5 are per-document; rule 6 is cross-document, which is why the
// generator takes both documents in a single invocation.
func Validate(docs []Document, registry map[string]string) []error {
	var errs []error
	seen := map[string]string{} // key -> document path that claimed it

	for _, d := range docs {
		for _, op := range d.Operations {
			markers := 0
			if op.Capability != "" {
				markers++
			}
			if op.CapabilityNone != "" {
				markers++
			}
			if op.Public {
				markers++
			}

			switch {
			case markers == 0:
				errs = append(errs, fmt.Errorf("%s: operation %q carries no authz marker: add x-authz-capability, x-authz-capability-none, or security: []", d.Path, op.ID))
			case markers > 1:
				errs = append(errs, fmt.Errorf("%s: operation %q carries %d authz markers; exactly one is allowed (two statements of one fact)", d.Path, op.ID, markers))
			}

			if op.Capability != "" {
				if _, ok := registry[op.Capability]; !ok {
					errs = append(errs, fmt.Errorf("%s: operation %q declares capability %q, which is not in the IAM registry", d.Path, op.ID, op.Capability))
				}
			}

			if op.CapabilityNone != "" && strings.TrimSpace(op.CapabilityNone) == "" {
				errs = append(errs, fmt.Errorf("%s: operation %q declares x-authz-capability-none with a blank reason", d.Path, op.ID))
			}

			if strings.Contains(op.Tag, ",") {
				errs = append(errs, fmt.Errorf("%s: operation %q declares more than one tag (%s); one operation, one owner", d.Path, op.ID, op.Tag))
			}
			if op.Tag == "" {
				errs = append(errs, fmt.Errorf("%s: operation %q declares no tag; the mounting publisher is identified by tag", d.Path, op.ID))
			}

			if d.ServerBase != "" && !strings.HasPrefix(d.ServerBase, "/") {
				errs = append(errs, fmt.Errorf("%s: declared server base %q must be absolute or empty", d.Path, d.ServerBase))
			}
			if !strings.HasPrefix(op.Path, "/") {
				errs = append(errs, fmt.Errorf("%s: operation %q path %q must be relative to the server base and start with /", d.Path, op.ID, op.Path))
			}

			k := Key(d, op)
			if prior, dup := seen[k]; dup {
				errs = append(errs, fmt.Errorf("key %q is declared by both %s and %s; a collision must fail the build, not the boot", k, prior, d.Path))
				continue
			}
			seen[k] = d.Path
		}
	}
	return errs
}

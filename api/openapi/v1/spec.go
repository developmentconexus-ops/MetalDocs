// Package openapiv1 makes the MetalDocs v1 contract readable at runtime.
//
// openapi.yaml beside this file is route truth: routes change only by editing
// it and re-running oapi-codegen. Until now that truth reached the process only
// through generated code — sixteen packages each carrying their own
// gzip-compressed copy behind a GetSwagger(), none of which was ever called
// outside its own file (G-08). Nothing at runtime could answer "what does the
// contract say about this request?".
//
// This package answers that, and it does so by embedding the SSOT itself
// rather than a derivative: //go:embed reads the same bytes git tracks, in the
// same directory, so there is no second copy to drift and no generated
// intermediate to regenerate. Go's embed cannot reach a parent directory, which
// is why this file lives here beside the spec instead of under internal/.
package openapiv1

import _ "embed"

//go:embed openapi.yaml
var specYAML []byte

// SpecYAML returns the embedded contract as YAML bytes.
//
// It returns a copy: the embedded slice is process-global, and a caller that
// sliced or sorted it in place would silently corrupt the contract for every
// other reader. Callers parse this once at boot, so the copy is not on any hot
// path.
func SpecYAML() []byte {
	out := make([]byte, len(specYAML))
	copy(out, specYAML)
	return out
}

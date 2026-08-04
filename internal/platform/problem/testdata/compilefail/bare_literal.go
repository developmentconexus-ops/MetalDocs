// Package compilefail MUST NOT COMPILE.
//
// It is the executable proof of ADR 0089's central guarantee: problem.Code is a
// closed type, so a bare string literal can no longer be used where a Code is
// expected. Under the old `type Code string`, every line below compiled without
// complaint — untyped string constants convert implicitly to a defined string
// type — which is precisely why the old type's "do not construct these directly"
// comment was a lie.
//
// It lives under testdata/ because the go tool ignores that directory: `go build
// ./...`, `go vet ./...` and `go test ./...` never see it. TestCode_BareString-
// LiteralDoesNotCompile in the parent package compiles it by explicit path and
// asserts the compiler REJECTS it.
//
// If this package ever starts compiling, someone reopened the type (added a
// string conversion, a Raw() constructor, or `type Code = string`). Fix the type,
// do not fix this file.
package compilefail

import "metaldocs/internal/platform/problem"

// A raw literal in a Code position.
var _ = problem.New(404, "NOT_FOUND", "not found")

// A raw literal assigned to a Code variable.
var _ problem.Code = "NOT_FOUND"

// A raw literal in a struct field of type Code.
var _ = problem.Problem{Status: 404, Code: "NOT_FOUND", Title: "not found"}

// Conversion out of the type (the reverse direction) is also closed: nothing may
// treat a Code as an interchangeable string without going through String().
var _ = string(problem.CodeNotFoundResource)

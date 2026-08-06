// Package apibase holds the single definition of the API's URL base.
//
// It exists because "/api/v1" was hard-coded in ten module files and in the
// spec's servers.url — a hand-synced constant of exactly the class the HTTP
// surface protocol deletes. The generated surface table's keys are built as
// "<METHOD> <BaseURL><path>", and oapi-codegen's HandlerWithOptions registers
// "<METHOD> " + options.BaseURL + path, so the two agree only if both read
// this constant.
package apibase

// BaseURL is the server base every v1 operation is mounted under. It must
// equal api/openapi/v1/openapi.yaml's servers[0].url.
const BaseURL = "/api/v1"

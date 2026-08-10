// Package openapivalidate enforces the OpenAPI contract at runtime.
//
// MetalDocs is contract-first: routes change only by editing
// api/openapi/v1/openapi.yaml and re-running oapi-codegen. Until now that was
// true of the SHAPE of the API — paths, methods, generated DTOs — but nothing
// checked an actual request against the constraints the spec declares. Every
// `required`, `enum`, `minLength`, `format` and `maximum` in 7700 lines of
// contract was documentation: each handler re-implemented whatever subset of it
// its author remembered, and a request the spec forbids still reached business
// code (G-08).
//
// This package closes that gap in the one place it can be closed once: the
// composed ingress. Not per module — a per-module validator is the same
// fifteen-dialect problem as the twelve local problem writers, and the moment
// one module forgets, invalid input is business input again.
//
// What it deliberately does NOT do:
//
//   - It does not validate responses. That is a different risk with a different
//     failure mode (a 500 for a bug in our own serializer), and it belongs to
//     its own decision.
//   - It does not answer 404 or 405. When the contract router finds no route,
//     the request passes through untouched to the mux, which owns those. That
//     is not a bypass: apps/api's assertSurface refuses to boot unless the
//     mounted route set equals the spec-generated surface, so "no route here"
//     is provably "no route anywhere" rather than a hole in the validator.
//   - It never rewrites the request. openapi3filter will, by default, inject
//     schema defaults into the body it hands onward; SkipSettingDefaults turns
//     that off, because a validator that edits the payload is no longer a
//     validator.
package openapivalidate

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"

	"metaldocs/internal/platform/problem"
)

// DefaultMaxBodyBytes caps what the validator will buffer to check a request
// body against its schema.
//
// A validator has to read the body to validate it, and an unbounded read at the
// ingress is a memory-exhaustion primitive handed to unauthenticated-adjacent
// traffic. 2 MiB is deliberately ABOVE the 1 MiB ceilings the handlers already
// enforce (idempotency's request-hash reader, controlled-documents' JSON
// limit), so this cap never preempts a handler's own smaller limit with a
// different problem code — it only stops the pathological case those limits
// were never reached for.
const DefaultMaxBodyBytes int64 = 2 << 20

// Middleware validates each request against the contract before the business
// handler runs.
type Middleware struct {
	router       routers.Router
	maxBodyBytes int64
	opts         *openapi3filter.Options
}

// New parses the contract and builds the router that matches requests to it.
//
// It fails closed by design: a spec that will not load or will not validate is
// a boot failure, not a warning. The alternative — serving with validation
// silently off — is strictly worse than not serving, because the guarantee this
// package exists to provide would be absent exactly when nobody is looking.
func New(specYAML []byte, maxBodyBytes int64) (*Middleware, error) {
	if len(specYAML) == 0 {
		return nil, errors.New("openapivalidate: empty spec")
	}
	if maxBodyBytes <= 0 {
		maxBodyBytes = DefaultMaxBodyBytes
	}

	loader := openapi3.NewLoader()
	// The contract is a single self-contained file embedded from the repo.
	// External refs would mean fetching a document at boot from wherever a $ref
	// points, which is neither reproducible nor safe.
	loader.IsExternalRefsAllowed = false

	doc, err := loader.LoadFromData(specYAML)
	if err != nil {
		return nil, fmt.Errorf("openapivalidate: load spec: %w", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("openapivalidate: spec is not a valid OpenAPI document: %w", err)
	}

	router, err := legacy.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("openapivalidate: build router: %w", err)
	}

	// openapi3's schema errors print the failing VALUE and the whole schema in
	// their Error() text. That value is the caller's payload — a password that
	// missed minLength, a token that missed a pattern. Anything that ever
	// stringifies such an error (a log line, a future detail message, a panic)
	// would then be handling a secret it never asked for. Disabling details at
	// the source makes that leak unrepresentable rather than merely avoided
	// here; the detail this package returns is built from the JSON pointer and
	// the schema keyword, which are safe by construction.
	openapi3.SchemaErrorDetailsDisabled = true

	return &Middleware{
		router:       router,
		maxBodyBytes: maxBodyBytes,
		opts: &openapi3filter.Options{
			// Authentication already happened upstream in the chain (authn,
			// then iam_authz). Without a func here, every operation carrying a
			// security requirement would fail validation outright; with the
			// noop, security requirements are satisfied and authorization stays
			// where it belongs. This does not weaken anything: a request that
			// reaches this link has already passed both.
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			// See the package comment: a validator must not mutate the payload.
			SkipSettingDefaults: true,
		},
	}, nil
}

// Wrap returns next guarded by contract validation.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, err := m.router.FindRoute(r)
		if err != nil {
			// A *routers.RouteError means the contract describes no such route:
			// 404 and 405 are the mux's business, not ours (see the package
			// comment). Matching the TYPE rather than the ErrPathNotFound /
			// ErrMethodNotAllowed sentinels is deliberate — the legacy router
			// builds fresh RouteError values instead of returning those
			// sentinels, so errors.Is against them never fires and every
			// unrouted request would have become a 500.
			var routeErr *routers.RouteError
			if errors.As(err, &routeErr) {
				next.ServeHTTP(w, r)
				return
			}
			// Anything else is a fault in the router or the spec — ours, not
			// the caller's — and must not be dressed up as a 400.
			problem.RespondCause(w, r,
				problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown,
					"request could not be matched against the API contract"), err)
			return
		}

		if r.Body != nil && r.Body != http.NoBody {
			r.Body = http.MaxBytesReader(w, r.Body, m.maxBodyBytes)
		}

		input := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options:    m.opts,
		}
		if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
			m.reject(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// reject translates a validation failure into the API's existing error
// vocabulary.
//
// The status choice is not a preference, it is what the contract declares: of
// the 51 operations with a request body, 49 declare 400 and only 16 declare
// 422, so 400 is the answer the spec actually promises for "this request does
// not match". Handlers that today answer 422 for a schema violation the spec
// declares are preempted by this link — that behaviour change is deliberate and
// recorded, and it moves the response TOWARD the contract, not away from it.
func (m *Middleware) reject(w http.ResponseWriter, r *http.Request, err error) {
	var maxBytes *http.MaxBytesError
	if errors.As(err, &maxBytes) {
		problem.Respond(w, r, problem.New(http.StatusRequestEntityTooLarge,
			problem.CodeRequestBodyTooLarge,
			fmt.Sprintf("request body exceeds the %d byte limit", m.maxBodyBytes)))
		return
	}

	var reqErr *openapi3filter.RequestError
	if !errors.As(err, &reqErr) {
		// A validator fault is a 500. Turning "our router panicked" into "your
		// request was bad" would be a lie that costs a caller hours.
		problem.RespondCause(w, r,
			problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown,
				"request validation failed unexpectedly"), err)
		return
	}

	status, code, detail := classify(reqErr)
	problem.Respond(w, r, problem.New(status, code, detail))
}

// classify maps a kin-openapi request error onto MetalDocs' problem codes.
//
// Every detail string it produces is assembled from the parameter name, the
// JSON pointer and the schema keyword — never from err.Error(), which embeds
// the offending value.
func classify(reqErr *openapi3filter.RequestError) (int, problem.Code, string) {
	if reqErr.Parameter != nil {
		name := reqErr.Parameter.Name
		if errors.Is(reqErr.Err, openapi3filter.ErrInvalidRequired) {
			return http.StatusBadRequest, problem.CodeRequestInvalid,
				fmt.Sprintf("%s parameter %q is required", reqErr.Parameter.In, name)
		}
		return http.StatusBadRequest, problem.CodeRequestInvalid,
			fmt.Sprintf("%s parameter %q does not match the API contract%s",
				reqErr.Parameter.In, name, schemaHint(reqErr.Err))
	}

	if reqErr.RequestBody != nil {
		switch {
		case errors.Is(reqErr.Err, openapi3filter.ErrInvalidRequired):
			return http.StatusBadRequest, problem.CodeRequestEmptyBody, "a request body is required"
		case strings.HasPrefix(reqErr.Reason, invalidContentTypePrefix):
			return http.StatusUnsupportedMediaType, problem.CodeRequestContentTypeUnsupported,
				"the request Content-Type is not declared for this operation"
		case reqErr.Reason == decodeFailureReason:
			return http.StatusBadRequest, problem.CodeRequestJSONDecode,
				"request body is not valid JSON for the declared content type"
		default:
			return http.StatusBadRequest, problem.CodeRequestInvalid,
				fmt.Sprintf("request body does not match the API contract%s", schemaHint(reqErr.Err))
		}
	}

	return http.StatusBadRequest, problem.CodeRequestInvalid, "request does not match the API contract"
}

// invalidContentTypePrefix and decodeFailureReason mirror the two RequestError
// reasons kin-openapi builds without an underlying typed error, so they are the
// only way to tell those two cases apart. They are pinned by test: if a
// kin-openapi upgrade rewords either one, the classification test fails rather
// than the API silently answering 400 where it used to answer 415.
const (
	invalidContentTypePrefix = "header Content-Type has unexpected value"
	decodeFailureReason      = "failed to decode request body"
)

// schemaHint renders the safe half of a schema error: where it failed and which
// keyword rejected it. The value that failed is deliberately absent.
func schemaHint(err error) string {
	var schemaErr *openapi3.SchemaError
	if !errors.As(err, &schemaErr) {
		return ""
	}
	var b strings.Builder
	if ptr := schemaErr.JSONPointer(); len(ptr) > 0 {
		b.WriteString(" at /")
		b.WriteString(strings.Join(ptr, "/"))
	}
	if schemaErr.SchemaField != "" {
		b.WriteString(" (")
		b.WriteString(schemaErr.SchemaField)
		b.WriteString(")")
	}
	if b.Len() == 0 {
		return ""
	}
	return ":" + b.String()
}

// LogSummary reports what the validator loaded, for the boot line. A guard
// nobody can see the size of is a guard nobody notices losing.
func (m *Middleware) LogSummary() slog.Attr {
	return slog.Int64("max_body_bytes", m.maxBodyBytes)
}

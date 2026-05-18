package approvalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	approvalsignature "metaldocs/internal/modules/documents/approval/infrastructure/signature"
	"metaldocs/internal/modules/documents/approval/repository"
	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/problem"
)

func TestMapErrorToResponse(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantTitle  string
	}{
		{
			name:       "repository stale revision",
			err:        repository.ErrStaleRevision,
			wantStatus: http.StatusConflict,
			wantCode:   "conflict.stale_revision",
			wantTitle:  repository.ErrStaleRevision.Error(),
		},
		{
			name:       "repository no active instance",
			err:        repository.ErrNoActiveInstance,
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found.instance",
			wantTitle:  repository.ErrNoActiveInstance.Error(),
		},
		{
			name:       "repository duplicate submission",
			err:        repository.ErrDuplicateSubmission,
			wantStatus: http.StatusConflict,
			wantCode:   "conflict.duplicate_submission",
			wantTitle:  repository.ErrDuplicateSubmission.Error(),
		},
		{
			name:       "repository actor already signed",
			err:        repository.ErrActorAlreadySigned,
			wantStatus: http.StatusConflict,
			wantCode:   "signoff.duplicate",
			wantTitle:  repository.ErrActorAlreadySigned.Error(),
		},
		{
			name:       "repository instance completed",
			err:        repository.ErrInstanceCompleted,
			wantStatus: http.StatusConflict,
			wantCode:   "state.instance_completed",
			wantTitle:  repository.ErrInstanceCompleted.Error(),
		},
		{
			name:       "domain no active stage",
			err:        domain.ErrNoActiveStage,
			wantStatus: http.StatusConflict,
			wantCode:   "state.instance_completed",
			wantTitle:  domain.ErrNoActiveStage.Error(),
		},
		{
			name:       "repository route in use",
			err:        repository.ErrRouteInUse,
			wantStatus: http.StatusConflict,
			wantCode:   "route.in_use",
			wantTitle:  repository.ErrRouteInUse.Error(),
		},
		{
			name:       "repository duplicate route profile",
			err:        repository.ErrDuplicateRouteProfile,
			wantStatus: http.StatusConflict,
			wantCode:   "route.duplicate_profile",
			wantTitle:  repository.ErrDuplicateRouteProfile.Error(),
		},
		{
			name:       "domain sod submitter cannot sign",
			err:        domain.ErrAuthorCannotSign,
			wantStatus: http.StatusForbidden,
			wantCode:   "sod.submitter_cannot_sign",
			wantTitle:  domain.ErrAuthorCannotSign.Error(),
		},
		{
			name:       "domain sod cross-stage duplicate",
			err:        domain.ErrActorAlreadySigned,
			wantStatus: http.StatusForbidden,
			wantCode:   "sod.cross_stage_duplicate",
			wantTitle:  domain.ErrActorAlreadySigned.Error(),
		},
		{
			name:       "freeze effective date missing",
			err:        v2dom.ErrEffectiveDateMissing,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "freeze.effective_date_missing",
			wantTitle:  v2dom.ErrEffectiveDateMissing.Error(),
		},
		{
			name:       "repository insufficient privilege",
			err:        repository.ErrInsufficientPrivilege,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal.db_privilege_missing",
			wantTitle:  "internal error",
		},
		{
			name:       "repository unknown db",
			err:        repository.ErrUnknownDB,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal.db_unknown",
			wantTitle:  "internal error",
		},
		{
			name:       "authz capability denied",
			err:        fmt.Errorf("wrap: %w", authz.ErrCapDenied{Capability: "x", AreaCode: "tenant", ActorID: "u1"}),
			wantStatus: http.StatusForbidden,
			wantCode:   "authz.capability_denied",
			wantTitle:  "wrap: authz: capability \"x\" denied for actor \"u1\" in area \"tenant\"",
		},
		{
			name:       "approval blocked by unresolved comments",
			err:        application.ErrApprovalBlockedByUnresolvedComments,
			wantStatus: http.StatusConflict,
			wantCode:   "approval.unresolved_comments",
			wantTitle:  application.ErrApprovalBlockedByUnresolvedComments.Error(),
		},
		{
			name:       "application reason required",
			err:        application.ErrReasonRequired,
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.reason_required",
			wantTitle:  application.ErrReasonRequired.Error(),
		},
		{
			name:       "application route not found",
			err:        application.ErrRouteNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found.route",
			wantTitle:  application.ErrRouteNotFound.Error(),
		},
		{
			name:       "context deadline exceeded",
			err:        context.DeadlineExceeded,
			wantStatus: http.StatusGatewayTimeout,
			wantCode:   "timeout",
			wantTitle:  "internal error",
		},
		{
			name: "json syntax error",
			err: func() error {
				var v map[string]any
				return json.Unmarshal([]byte("{"), &v)
			}(),
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.json_decode",
			wantTitle:  "unexpected end of JSON input",
		},
		{
			name: "json type error",
			err: func() error {
				var v struct {
					N int `json:"n"`
				}
				return json.Unmarshal([]byte(`{"n":"x"}`), &v)
			}(),
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.json_type_error",
			wantTitle:  "json: cannot unmarshal string into Go struct field .n of type int",
		},
		{
			name:       "io EOF",
			err:        io.EOF,
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.empty_body",
			wantTitle:  io.EOF.Error(),
		},
		{
			name:       "contracts content type",
			err:        contracts.ErrContentType,
			wantStatus: http.StatusUnsupportedMediaType,
			wantCode:   "validation.content_type",
			wantTitle:  contracts.ErrContentType.Error(),
		},
		{
			name:       "contracts body too large",
			err:        contracts.ErrBodyTooLarge,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantCode:   "validation.body_too_large",
			wantTitle:  contracts.ErrBodyTooLarge.Error(),
		},
		{
			name:       "contracts empty body",
			err:        contracts.ErrEmptyBody,
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.empty_body",
			wantTitle:  contracts.ErrEmptyBody.Error(),
		},
		{
			name:       "contracts duplicate key",
			err:        contracts.ErrDuplicateKey,
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.duplicate_key",
			wantTitle:  contracts.ErrDuplicateKey.Error(),
		},
		{
			name:       "if-match required",
			err:        ErrIfMatchRequired,
			wantStatus: http.StatusPreconditionRequired,
			wantCode:   "precondition.if_match_required",
			wantTitle:  ErrIfMatchRequired.Error(),
		},
		{
			name:       "if-match malformed",
			err:        ErrIfMatchMalformed,
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.if_match_malformed",
			wantTitle:  ErrIfMatchMalformed.Error(),
		},
		{
			name:       "idempotency key required",
			err:        ErrIdempotencyRequired,
			wantStatus: http.StatusBadRequest,
			wantCode:   "idempotency.key_required",
			wantTitle:  ErrIdempotencyRequired.Error(),
		},
		{
			name:       "content hash mismatch",
			err:        ErrContentHashMismatch,
			wantStatus: http.StatusPreconditionFailed,
			wantCode:   "precondition.content_hash_mismatch",
			wantTitle:  ErrContentHashMismatch.Error(),
		},
		{
			name:       "signature invalid",
			err:        approvalsignature.ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "authn.signature_invalid",
			wantTitle:  approvalsignature.ErrInvalidCredentials.Error(),
		},
		{
			name:       "generic validation error",
			err:        errors.New("route_id is required"),
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.request_invalid",
			wantTitle:  "route_id is required",
		},
		{
			name:       "unknown error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal.unknown",
			wantTitle:  "internal error",
		},
		{
			name:       "nil error",
			err:        nil,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal.unknown",
			wantTitle:  "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prob := MapErrorToResponse(tt.err)
			if prob.Status != tt.wantStatus {
				t.Fatalf("status = %d, want %d", prob.Status, tt.wantStatus)
			}
			if prob.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q", prob.Code, tt.wantCode)
			}
			if prob.Title != tt.wantTitle {
				t.Fatalf("title = %q, want %q", prob.Title, tt.wantTitle)
			}
		})
	}
}

func TestMapErrorToResponse_WrappedSentinel(t *testing.T) {
	err := fmt.Errorf("service: %w", repository.ErrStaleRevision)
	prob := MapErrorToResponse(err)

	if prob.Status != http.StatusConflict {
		t.Fatalf("status = %d, want %d", prob.Status, http.StatusConflict)
	}
	if prob.Code != "conflict.stale_revision" {
		t.Fatalf("code = %q, want %q", prob.Code, "conflict.stale_revision")
	}
	if prob.Title != err.Error() {
		t.Fatalf("title = %q, want %q", prob.Title, err.Error())
	}
}

func TestMapErrorToResponse_ErrNoActiveStage(t *testing.T) {
	prob := MapErrorToResponse(domain.ErrNoActiveStage)

	if prob.Status != http.StatusConflict {
		t.Fatalf("status = %d, want %d", prob.Status, http.StatusConflict)
	}
	if prob.Code != "state.instance_completed" {
		t.Fatalf("code = %q, want %q", prob.Code, "state.instance_completed")
	}
}

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()

	WriteError(rr, repository.ErrStaleRevision)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusConflict)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("content-type = %q, want %q", got, "application/problem+json")
	}

	var body problem.Problem
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "conflict.stale_revision" {
		t.Fatalf("code = %q, want %q", body.Code, "conflict.stale_revision")
	}
}

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	payload := map[string]string{"ok": "yes"}

	WriteJSON(rr, http.StatusAccepted, payload)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want %q", got, "application/json")
	}

	want, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal want: %v", err)
	}
	if got := bytes.TrimSpace(rr.Body.Bytes()); !bytes.Equal(got, want) {
		t.Fatalf("body = %s, want %s", got, want)
	}
}

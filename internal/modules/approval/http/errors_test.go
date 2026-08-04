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

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	approvalsignature "metaldocs/internal/modules/approval/infrastructure/signature"
	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/strictjson"
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
			err:        infrastructure.ErrStaleRevision,
			wantStatus: http.StatusConflict,
			wantCode:   "conflict.stale_revision",
			wantTitle:  infrastructure.ErrStaleRevision.Error(),
		},
		{
			name:       "repository no active instance",
			err:        infrastructure.ErrNoActiveInstance,
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found.instance",
			wantTitle:  infrastructure.ErrNoActiveInstance.Error(),
		},
		{
			name:       "repository duplicate submission",
			err:        infrastructure.ErrDuplicateSubmission,
			wantStatus: http.StatusConflict,
			wantCode:   "conflict.duplicate_submission",
			wantTitle:  infrastructure.ErrDuplicateSubmission.Error(),
		},
		{
			name:       "repository actor already signed",
			err:        infrastructure.ErrActorAlreadySigned,
			wantStatus: http.StatusConflict,
			wantCode:   "signoff.duplicate",
			wantTitle:  infrastructure.ErrActorAlreadySigned.Error(),
		},
		{
			name:       "repository instance completed",
			err:        infrastructure.ErrInstanceCompleted,
			wantStatus: http.StatusConflict,
			wantCode:   "state.instance_completed",
			wantTitle:  infrastructure.ErrInstanceCompleted.Error(),
		},
		{
			name:       "domain no active stage",
			err:        domain.ErrNoActiveStage,
			wantStatus: http.StatusConflict,
			wantCode:   "state.instance_completed",
			wantTitle:  domain.ErrNoActiveStage.Error(),
		},
		{
			name:       "R3/G2 verdict ready on approval stage",
			err:        domain.ErrVerdictReadyOnApprovalStage,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "state.verdict_ready_on_approval_stage",
			wantTitle:  domain.ErrVerdictReadyOnApprovalStage.Error(),
		},
		{
			name:       "R3/G2 verdict wrong stage kind (unreachable internal-state -> 500)",
			err:        domain.ErrVerdictWrongStageKind,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal.verdict_wrong_stage_kind",
			wantTitle:  "internal error", // 5xx bodies stay generic; raw sentinel not leaked
		},
		{
			name:       "repository route in use",
			err:        infrastructure.ErrRouteInUse,
			wantStatus: http.StatusConflict,
			wantCode:   "route.in_use",
			wantTitle:  infrastructure.ErrRouteInUse.Error(),
		},
		{
			name:       "repository duplicate route profile",
			err:        infrastructure.ErrDuplicateRouteProfile,
			wantStatus: http.StatusConflict,
			wantCode:   "route.duplicate_profile",
			wantTitle:  infrastructure.ErrDuplicateRouteProfile.Error(),
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
			name:       "F6.3 reason for change required",
			err:        application.ErrReasonForChangeRequired,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation.reason_for_change_required",
			wantTitle:  application.ErrReasonForChangeRequired.Error(),
		},
		{
			name:       "F6.3 reason category invalid",
			err:        application.ErrReasonCategoryInvalid,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation.reason_category_invalid",
			wantTitle:  application.ErrReasonCategoryInvalid.Error(),
		},
		{
			// ADR 0073: the four sentinels the canonical /submit path now surfaces
			// (previously finalize-only) must map to typed problem+json, never 500.
			name:       "ADR0073 revision title required",
			err:        application.ErrRevisionTitleRequired,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation.revision_title_required",
			wantTitle:  application.ErrRevisionTitleRequired.Error(),
		},
		{
			name:       "ADR0073 document not draft",
			err:        v2dom.ErrDocumentNotDraft,
			wantStatus: http.StatusConflict,
			wantCode:   "state.document_not_draft",
			wantTitle:  v2dom.ErrDocumentNotDraft.Error(),
		},
		{
			name:       "ADR0073 profile not configured",
			err:        v2dom.ErrProfileNotConfigured,
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation.profile_not_configured",
			wantTitle:  v2dom.ErrProfileNotConfigured.Error(),
		},
		{
			name:       "ADR0073 approval route missing",
			err:        v2dom.ErrApprovalRouteMissing,
			wantStatus: http.StatusConflict,
			wantCode:   "state.approval_route_missing",
			wantTitle:  v2dom.ErrApprovalRouteMissing.Error(),
		},
		{
			name:       "repository insufficient privilege",
			err:        infrastructure.ErrInsufficientPrivilege,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal.db_privilege_missing",
			wantTitle:  "internal error",
		},
		{
			name:       "repository unknown db",
			err:        infrastructure.ErrUnknownDB,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal.db_unknown",
			wantTitle:  "internal error",
		},
		{
			name:       "authz capability denied",
			err:        fmt.Errorf("wrap: %w", authz.ErrCapDenied{Capability: "x", AreaCode: "tenant", ActorID: "u1"}),
			wantStatus: http.StatusForbidden,
			wantCode:   "permission.capability_denied",
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
			name:       "application route already inactive",
			err:        application.ErrRouteAlreadyInactive,
			wantStatus: http.StatusConflict,
			wantCode:   "state.route_inactive",
			wantTitle:  application.ErrRouteAlreadyInactive.Error(),
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
			wantCode:   "request.json_decode",
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
			wantCode:   "request.empty_body",
			wantTitle:  io.EOF.Error(),
		},
		{
			name:       "strictjson content type",
			err:        strictjson.ErrContentType,
			wantStatus: http.StatusUnsupportedMediaType,
			wantCode:   "validation.content_type",
			wantTitle:  strictjson.ErrContentType.Error(),
		},
		{
			name:       "strictjson body too large",
			err:        strictjson.ErrBodyTooLarge,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantCode:   "validation.body_too_large",
			wantTitle:  strictjson.ErrBodyTooLarge.Error(),
		},
		{
			name:       "strictjson empty body",
			err:        strictjson.ErrEmptyBody,
			wantStatus: http.StatusBadRequest,
			wantCode:   "request.empty_body",
			wantTitle:  strictjson.ErrEmptyBody.Error(),
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
			err:        NewValidationError("route_id is required"),
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
			if prob.Code.String() != tt.wantCode {
				t.Fatalf("code = %q, want %q", prob.Code, tt.wantCode)
			}
			if prob.Title != tt.wantTitle {
				t.Fatalf("title = %q, want %q", prob.Title, tt.wantTitle)
			}
		})
	}
}

func TestMapErrorToResponse_WrappedSentinel(t *testing.T) {
	err := fmt.Errorf("service: %w", infrastructure.ErrStaleRevision)
	prob := MapErrorToResponse(err)

	if prob.Status != http.StatusConflict {
		t.Fatalf("status = %d, want %d", prob.Status, http.StatusConflict)
	}
	if prob.Code.String() != "conflict.stale_revision" {
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
	if prob.Code.String() != "state.instance_completed" {
		t.Fatalf("code = %q, want %q", prob.Code, "state.instance_completed")
	}
}

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()

	WriteError(rr, infrastructure.ErrStaleRevision)

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
	if body.Code.String() != "conflict.stale_revision" {
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

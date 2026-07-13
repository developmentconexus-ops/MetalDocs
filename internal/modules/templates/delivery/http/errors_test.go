package http

import (
	"errors"
	"net/http"
	"testing"

	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/problem"
)

func TestMapErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   problem.Code
	}{
		{name: "not found", err: domain.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: problem.CodeNotFound},
		{name: "key conflict", err: domain.ErrKeyConflict, wantStatus: http.StatusConflict, wantCode: problem.CodeAlreadyExists},
		{name: "invalid state transition", err: domain.ErrInvalidStateTransition, wantStatus: http.StatusConflict, wantCode: problem.CodeStateTransitionInvalid},
		{name: "stale base", err: domain.ErrStaleBase, wantStatus: http.StatusConflict, wantCode: problem.CodeStaleBase},
		{name: "stale lock version", err: domain.ErrStaleLockVersion, wantStatus: http.StatusPreconditionFailed, wantCode: problem.CodeConcurrentModification},
		{name: "concurrent transition", err: domain.ErrConcurrentTransition, wantStatus: http.StatusConflict, wantCode: problem.CodeStateTransitionInvalid},
		{name: "content hash mismatch", err: domain.ErrContentHashMismatch, wantStatus: http.StatusConflict, wantCode: problem.CodeConflict},
		{name: "upload missing", err: domain.ErrUploadMissing, wantStatus: http.StatusConflict, wantCode: problem.CodeUploadMissing},
		{name: "upload too large", err: domain.ErrUploadTooLarge, wantStatus: http.StatusRequestEntityTooLarge, wantCode: problem.CodeRequestBodyTooLarge},
		{name: "iso segregation violation", err: domain.ErrISOSegregationViolation, wantStatus: http.StatusForbidden, wantCode: problem.CodeISOSegregationViolation},
		{name: "forbidden", err: domain.ErrForbidden, wantStatus: http.StatusForbidden, wantCode: problem.CodeAuthForbidden},
		{name: "system template immutable", err: domain.ErrSystemTemplateImmutable, wantStatus: http.StatusConflict, wantCode: problem.CodeSystemTemplateImmutable},
		{name: "archived", err: domain.ErrArchived, wantStatus: http.StatusConflict, wantCode: problem.CodeStateTransitionInvalid},
		{name: "placeholder name invalid", err: domain.ErrPlaceholderNameInvalid, wantStatus: http.StatusUnprocessableEntity, wantCode: problem.CodeValidationError},
		{name: "duplicate placeholder name", err: domain.ErrDuplicatePlaceholderName, wantStatus: http.StatusUnprocessableEntity, wantCode: problem.CodeAlreadyExists},
		{name: "default", err: errors.New("boom"), wantStatus: http.StatusInternalServerError, wantCode: problem.CodeInternalError},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			status, code := MapErr(tc.err)
			if status != tc.wantStatus {
				t.Fatalf("status: got %d want %d", status, tc.wantStatus)
			}
			if code != tc.wantCode {
				t.Fatalf("code: got %q want %q", code, tc.wantCode)
			}
		})
	}
}

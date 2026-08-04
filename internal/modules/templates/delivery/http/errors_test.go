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
		{name: "not found", err: domain.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: problem.CodeNotFoundResource},
		{name: "key conflict", err: domain.ErrKeyConflict, wantStatus: http.StatusConflict, wantCode: problem.CodeConflictAlreadyExists},
		{name: "invalid state transition", err: domain.ErrInvalidStateTransition, wantStatus: http.StatusConflict, wantCode: problem.CodeStateTransitionInvalid},
		{name: "stale base", err: domain.ErrStaleBase, wantStatus: http.StatusConflict, wantCode: problem.CodeConflictStaleBase},
		{name: "stale lock version", err: domain.ErrStaleLockVersion, wantStatus: http.StatusPreconditionFailed, wantCode: problem.CodeConflictConcurrentModification},
		{name: "concurrent transition", err: domain.ErrConcurrentTransition, wantStatus: http.StatusConflict, wantCode: problem.CodeStateTransitionInvalid},
		{name: "content hash mismatch", err: domain.ErrContentHashMismatch, wantStatus: http.StatusConflict, wantCode: problem.CodeConflictGeneric},
		{name: "upload missing", err: domain.ErrUploadMissing, wantStatus: http.StatusConflict, wantCode: problem.CodeStateUploadMissing},
		{name: "upload too large", err: domain.ErrUploadTooLarge, wantStatus: http.StatusRequestEntityTooLarge, wantCode: problem.CodeRequestBodyTooLarge},
		{name: "iso segregation violation", err: domain.ErrISOSegregationViolation, wantStatus: http.StatusForbidden, wantCode: problem.CodePermissionISOSegregationViolation},
		{name: "forbidden", err: domain.ErrForbidden, wantStatus: http.StatusForbidden, wantCode: problem.CodePermissionDenied},
		{name: "system template immutable", err: domain.ErrSystemTemplateImmutable, wantStatus: http.StatusConflict, wantCode: problem.CodeStateSystemTemplateImmutable},
		{name: "archived", err: domain.ErrArchived, wantStatus: http.StatusConflict, wantCode: problem.CodeStateTransitionInvalid},
		{name: "placeholder name invalid", err: domain.ErrPlaceholderNameInvalid, wantStatus: http.StatusUnprocessableEntity, wantCode: problem.CodeRequestInvalid},
		{name: "duplicate placeholder name", err: domain.ErrDuplicatePlaceholderName, wantStatus: http.StatusUnprocessableEntity, wantCode: problem.CodeConflictAlreadyExists},
		{name: "default", err: errors.New("boom"), wantStatus: http.StatusInternalServerError, wantCode: problem.CodeInternalUnknown},
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

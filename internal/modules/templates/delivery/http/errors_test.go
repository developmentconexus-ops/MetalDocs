package http

import (
	"errors"
	"net/http"
	"testing"

	"metaldocs/internal/modules/templates/domain"
)

func TestMapErr(t *testing.T) {
	t.Parallel()

	// wantCode carries the LITERAL wire string from the ADR 0089 annex rename
	// table (§2.1 / §2.3 / §2.8), not a problem.Code* var from the source under
	// test. An assertion derived from the code under test only ratifies whatever
	// that code says; a literal from the annex independently checks that the
	// table was executed.
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "not found", err: domain.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: "notfound.resource"},
		{name: "key conflict", err: domain.ErrKeyConflict, wantStatus: http.StatusConflict, wantCode: "conflict.already_exists"},
		{name: "invalid state transition", err: domain.ErrInvalidStateTransition, wantStatus: http.StatusConflict, wantCode: "state.transition_invalid"},
		{name: "stale base", err: domain.ErrStaleBase, wantStatus: http.StatusConflict, wantCode: "conflict.stale_base"},
		// annex #159 / R-7.
		{name: "stale lock version", err: domain.ErrStaleLockVersion, wantStatus: http.StatusPreconditionFailed, wantCode: "precondition.lock_version_stale"},
		{name: "concurrent transition", err: domain.ErrConcurrentTransition, wantStatus: http.StatusConflict, wantCode: "state.transition_invalid"},
		// R-2 (precondition.content_hash_mismatch @412) is NOT applied here: that
		// code's single registration lives in approval/http, which templates may
		// not import. See the sweep report.
		{name: "content hash mismatch", err: domain.ErrContentHashMismatch, wantStatus: http.StatusConflict, wantCode: "conflict.generic"},
		{name: "upload missing", err: domain.ErrUploadMissing, wantStatus: http.StatusConflict, wantCode: "state.upload_missing"},
		{name: "upload too large", err: domain.ErrUploadTooLarge, wantStatus: http.StatusRequestEntityTooLarge, wantCode: "request.body_too_large"},
		{name: "iso segregation violation", err: domain.ErrISOSegregationViolation, wantStatus: http.StatusForbidden, wantCode: "permission.iso_segregation_violation"},
		{name: "forbidden", err: domain.ErrForbidden, wantStatus: http.StatusForbidden, wantCode: "permission.denied"},
		{name: "system template immutable", err: domain.ErrSystemTemplateImmutable, wantStatus: http.StatusConflict, wantCode: "state.system_template_immutable"},
		{name: "archived", err: domain.ErrArchived, wantStatus: http.StatusConflict, wantCode: "state.transition_invalid"},
		// annex #161 / R-6.
		{name: "placeholder name invalid", err: domain.ErrPlaceholderNameInvalid, wantStatus: http.StatusUnprocessableEntity, wantCode: "validation.placeholder_name_invalid"},
		// annex #162 / R-18.
		{name: "duplicate placeholder name", err: domain.ErrDuplicatePlaceholderName, wantStatus: http.StatusUnprocessableEntity, wantCode: "validation.placeholder_name_duplicate"},
		{name: "default", err: errors.New("boom"), wantStatus: http.StatusInternalServerError, wantCode: "internal.unknown"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			status, code := MapErr(tc.err)
			if status != tc.wantStatus {
				t.Fatalf("status: got %d want %d", status, tc.wantStatus)
			}
			if code.String() != tc.wantCode {
				t.Fatalf("code: got %q want %q", code.String(), tc.wantCode)
			}
		})
	}
}

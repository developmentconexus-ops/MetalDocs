package http

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	documentsapi "metaldocs/internal/modules/documents/api"
	"metaldocs/internal/modules/documents/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
)

// jsonTopKeys marshals v and returns its sorted top-level JSON object keys. It is
// the F7.4 wire-contract probe: the typed bodies must serialize to exactly the
// OpenAPI-declared (or, for the off-spec webhook, the prior-literal) key set.
func jsonTopKeys(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// TestFillInSchemaResponse_EnvelopeAndEmptyParity locks the DocumentFillInSchemaResponse
// envelope to {data:{placeholder_schema:[...]}} and guards the empty case as [] not null
// (the handler boxes a non-nil make()-d slice).
func TestFillInSchemaResponse_EnvelopeAndEmptyParity(t *testing.T) {
	var empty documentsapi.DocumentFillInSchemaResponse
	empty.Data.PlaceholderSchema = toAnySlice([]templatesdomain.Placeholder{})
	raw, err := json.Marshal(empty)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(raw); got != `{"data":{"placeholder_schema":[]}}` {
		t.Fatalf("empty fill-in envelope = %s, want {\"data\":{\"placeholder_schema\":[]}}", got)
	}

	// A boxed domain placeholder must marshal byte-identically to the domain type
	// emitted directly (the pre-F7.4 wire). The []any boxing is wire-neutral.
	ph := templatesdomain.Placeholder{ID: "p-1", Label: "Name", Type: templatesdomain.PHText, Required: true}
	var populated documentsapi.DocumentFillInSchemaResponse
	populated.Data.PlaceholderSchema = toAnySlice([]templatesdomain.Placeholder{ph})
	gotRaw, _ := json.Marshal(populated.Data.PlaceholderSchema)
	wantRaw, _ := json.Marshal([]templatesdomain.Placeholder{ph})
	if string(gotRaw) != string(wantRaw) {
		t.Fatalf("boxed placeholder_schema = %s, want %s (boxing must be wire-neutral)", gotRaw, wantRaw)
	}
}

// TestPutPlaceholderValueResponse_SecondPrecision locks the envelope keys and proves
// the generated time.Time field marshals RFC3339 seconds-only (no fractional part),
// byte-identical to the prior Format(time.RFC3339) string wire.
func TestPutPlaceholderValueResponse_SecondPrecision(t *testing.T) {
	resp := documentsapi.PutPlaceholderValueResponse{
		PlaceholderId: "pid-1",
		UpdatedAt:     time.Date(2026, 6, 20, 12, 0, 0, 123456789, time.UTC).Truncate(time.Second),
	}
	if got, want := jsonTopKeys(t, resp), "placeholder_id,updated_at"; got != want {
		t.Fatalf("put response keys = %q, want %q", got, want)
	}
	raw, _ := json.Marshal(resp)
	if !strings.Contains(string(raw), `"updated_at":"2026-06-20T12:00:00Z"`) {
		t.Fatalf("updated_at not second-precision RFC3339: %s", raw)
	}
}

// TestViewDocumentResponse_ConditionalKeys proves pdf_status is always present and
// signed_url/pdf_url appear only when set (omitempty pointers), matching the prior
// conditional map literal.
func TestViewDocumentResponse_ConditionalKeys(t *testing.T) {
	pending := documentsapi.ViewDocumentResponse{PdfStatus: "pending"}
	if got, want := jsonTopKeys(t, pending), "pdf_status"; got != want {
		t.Fatalf("pending view keys = %q, want %q (no signed_url/pdf_url)", got, want)
	}

	url := "https://example/s3"
	ready := documentsapi.ViewDocumentResponse{PdfStatus: "ready", SignedUrl: &url, PdfUrl: &url}
	if got, want := jsonTopKeys(t, ready), "pdf_status,pdf_url,signed_url"; got != want {
		t.Fatalf("ready view keys = %q, want %q", got, want)
	}
}

// TestPlaceholderOptionsResponse_BothBranchesParity proves the typed envelope boxes
// each polymorphic branch byte-identically to the prior {"options": <slice>} literal.
func TestPlaceholderOptionsResponse_BothBranchesParity(t *testing.T) {
	if got, want := jsonTopKeys(t, documentsapi.PlaceholderOptionsResponse{Options: toAnySlice([]string{})}), "options"; got != want {
		t.Fatalf("options envelope keys = %q, want %q", got, want)
	}

	// select branch: []map[string]string
	sel := selectOptions([]string{"a"})
	selResp := documentsapi.PlaceholderOptionsResponse{Options: toAnySlice(sel)}
	gotSel, _ := json.Marshal(selResp.Options)
	wantSel, _ := json.Marshal(sel)
	if string(gotSel) != string(wantSel) {
		t.Fatalf("select options boxed = %s, want %s", gotSel, wantSel)
	}

	// user branch: []UserOptionView
	users := []UserOptionView{{UserID: "u-1", DisplayName: "Alice"}}
	userResp := documentsapi.PlaceholderOptionsResponse{Options: toAnySlice(users)}
	gotUser, _ := json.Marshal(userResp.Options)
	wantUser, _ := json.Marshal(users)
	if string(gotUser) != string(wantUser) {
		t.Fatalf("user options boxed = %s, want %s", gotUser, wantUser)
	}
}

// --- M9 (HS-5 contract type-gate) typed-body locks -------------------------
// Each replaces a former map[string]<T> literal that evaded the H-D gate. The
// lock is the consumer wire: the generated struct must serialize to exactly the
// OpenAPI-declared key set, byte-for-byte where the shape is closed.

// TestDocumentCreateResult_WireContract locks the duplicateDocument body (F9.1),
// formerly map[string]string, to {document_id, initial_revision_id, session_id}.
func TestDocumentCreateResult_WireContract(t *testing.T) {
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	resp := documentsapi.DocumentCreateResult{DocumentId: id, InitialRevisionId: id, SessionId: id}
	if got, want := jsonTopKeys(t, resp), "document_id,initial_revision_id,session_id"; got != want {
		t.Fatalf("create-result keys = %q, want %q", got, want)
	}
}

// TestRevisionUrlResponse_WireContract locks the signedRevisionURL body (F9.3),
// formerly map[string]string, to {url} — the 200 body the FE consumer reads.
func TestRevisionUrlResponse_WireContract(t *testing.T) {
	resp := documentsapi.RevisionUrlResponse{Url: "https://example/s3/rev"}
	if got, want := jsonTopKeys(t, resp), "url"; got != want {
		t.Fatalf("revision-url keys = %q, want %q", got, want)
	}
	raw, _ := json.Marshal(resp)
	if string(raw) != `{"url":"https://example/s3/rev"}` {
		t.Fatalf("revision-url wire = %s", raw)
	}
}

// TestDocumentCommentResponse_WireContract locks the comment body (F9.2),
// formerly a hand-rolled map, to the OpenAPI key set. resolved_at and
// parent_library_id are spec-required+nullable, so oapi-codegen omits
// `omitempty` — both keys always serialize (null when unset).
func TestDocumentCommentResponse_WireContract(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	resp := documentsapi.DocumentCommentResponse{
		Id:               uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		Author:           "Alice",
		AuthorId:         "u-1",
		Content:          []documentsapi.DocumentCommentContentNode{},
		Done:             false,
		LibraryCommentId: 7,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if got, want := jsonTopKeys(t, resp), "author,author_id,content,created_at,done,id,library_comment_id,parent_library_id,resolved_at,updated_at"; got != want {
		t.Fatalf("comment keys (unresolved) = %q, want %q", got, want)
	}
	resolved := resp
	resolved.ResolvedAt = &now
	parent := 4
	resolved.ParentLibraryId = &parent
	if got, want := jsonTopKeys(t, resolved), "author,author_id,content,created_at,done,id,library_comment_id,parent_library_id,resolved_at,updated_at"; got != want {
		t.Fatalf("comment keys (resolved) = %q, want %q", got, want)
	}
}

// TestPDFCompleteResponse_WireContract locks the off-spec hand-rolled webhook body
// to its prior {document_id, final_pdf_s3_key} literal shape.
func TestPDFCompleteResponse_WireContract(t *testing.T) {
	resp := pdfCompleteResponse{DocumentID: "d-1", FinalPDFS3Key: "final/r.pdf"}
	if got, want := jsonTopKeys(t, resp), "document_id,final_pdf_s3_key"; got != want {
		t.Fatalf("pdf-complete keys = %q, want %q", got, want)
	}
	raw, _ := json.Marshal(resp)
	if string(raw) != `{"document_id":"d-1","final_pdf_s3_key":"final/r.pdf"}` {
		t.Fatalf("pdf-complete wire = %s", raw)
	}
}

// TestDocumentSummaryAndDetail_ReviewFieldsWireContract pins M6 F6.2 T6 +
// F6.4 D2: the generated DocumentSummary/DocumentDetailResponse types carry
// effective_from/effective_to/review_due_at/last_reviewed_at/
// review_surfaced_at as required+nullable (ADR 0035 flat-body discipline —
// present-and-null, never omitted) and toDocumentSummary/
// toDocumentDetailResponse populate them nil-safe from the domain.Document.
func TestDocumentSummaryAndDetail_ReviewFieldsWireContract(t *testing.T) {
	reviewDueAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	reviewSurfacedAt := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)

	// All five unset (nil) on the domain document -> all five keys still
	// present in the JSON body, serialized as null (never omitted).
	empty, err := toDocumentSummary(domain.Document{FormDataJSON: []byte(`{}`)})
	if err != nil {
		t.Fatalf("toDocumentSummary: %v", err)
	}
	raw, _ := json.Marshal(empty)
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"effective_from", "effective_to", "review_due_at", "last_reviewed_at", "review_surfaced_at"} {
		got, ok := m[key]
		if !ok {
			t.Fatalf("DocumentSummary missing required+nullable key %q in wire body %s", key, raw)
		}
		if string(got) != "null" {
			t.Fatalf("DocumentSummary[%q] = %s, want null when unset on the domain doc", key, got)
		}
	}

	// review_due_at / review_surfaced_at set on the domain document flow
	// through to the wire value (nil-safe passthrough, not a fixed null).
	withReview, err := toDocumentSummary(domain.Document{FormDataJSON: []byte(`{}`), ReviewDueAt: &reviewDueAt, ReviewSurfacedAt: &reviewSurfacedAt})
	if err != nil {
		t.Fatalf("toDocumentSummary: %v", err)
	}
	rawWith, _ := json.Marshal(withReview)
	var mWith map[string]json.RawMessage
	if err := json.Unmarshal(rawWith, &mWith); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	wantReviewDue, _ := json.Marshal(reviewDueAt)
	if string(mWith["review_due_at"]) != string(wantReviewDue) {
		t.Fatalf("DocumentSummary[review_due_at] = %s, want %s", mWith["review_due_at"], wantReviewDue)
	}
	wantReviewSurfaced, _ := json.Marshal(reviewSurfacedAt)
	if string(mWith["review_surfaced_at"]) != string(wantReviewSurfaced) {
		t.Fatalf("DocumentSummary[review_surfaced_at] = %s, want %s", mWith["review_surfaced_at"], wantReviewSurfaced)
	}

	// Same required+nullable contract on DocumentDetailResponse.
	detail, err := toDocumentDetailResponse(domain.Document{FormDataJSON: []byte(`{}`)})
	if err != nil {
		t.Fatalf("toDocumentDetailResponse: %v", err)
	}
	rawDetail, _ := json.Marshal(detail)
	var mDetail map[string]json.RawMessage
	if err := json.Unmarshal(rawDetail, &mDetail); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"effective_from", "effective_to", "review_due_at", "last_reviewed_at", "review_surfaced_at"} {
		got, ok := mDetail[key]
		if !ok {
			t.Fatalf("DocumentDetailResponse missing required+nullable key %q in wire body %s", key, rawDetail)
		}
		if string(got) != "null" {
			t.Fatalf("DocumentDetailResponse[%q] = %s, want null when unset on the domain doc", key, got)
		}
	}
}

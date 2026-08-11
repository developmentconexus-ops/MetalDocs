package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	httphandler "metaldocs/internal/modules/documents/delivery/http"
	"metaldocs/internal/modules/documents/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

type commentsStatefulSvc struct {
	*fakeSvc
	comments   map[int]domain.Comment
	nextOffset int
}

func newCommentsStatefulSvc() *commentsStatefulSvc {
	return &commentsStatefulSvc{
		fakeSvc:    &fakeSvc{},
		comments:   map[int]domain.Comment{},
		nextOffset: 0,
	}
}

func (s *commentsStatefulSvc) ListDocumentComments(_ context.Context, _, _ string) ([]domain.Comment, error) {
	out := make([]domain.Comment, 0, len(s.comments))
	for _, c := range s.comments {
		out = append(out, c)
	}
	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].CreatedAt.Before(out[i].CreatedAt) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

func (s *commentsStatefulSvc) AddDocumentComment(_ context.Context, _, userID, _, _ string, in domain.CommentCreateInput) (*domain.Comment, error) {
	now := time.Now().UTC().Add(time.Duration(s.nextOffset) * time.Second)
	s.nextOffset++
	comment := domain.Comment{
		ID:               uuid.New(),
		LibraryCommentID: in.LibraryCommentID,
		ParentLibraryID:  in.ParentLibraryID,
		AuthorID:         userID,
		AuthorDisplay:    in.AuthorDisplay,
		ContentJSON:      append([]byte(nil), in.ContentJSON...),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.comments[in.LibraryCommentID] = comment
	copyComment := comment
	return &copyComment, nil
}

func (s *commentsStatefulSvc) UpdateDocumentComment(_ context.Context, _, userID, _ string, libraryID int, in domain.CommentUpdateInput) (*domain.Comment, error) {
	comment, ok := s.comments[libraryID]
	if !ok {
		return nil, domain.ErrCommentNotFound
	}
	now := time.Now().UTC().Add(time.Duration(s.nextOffset) * time.Second)
	s.nextOffset++
	if in.ContentJSON != nil {
		comment.ContentJSON = append([]byte(nil), (*in.ContentJSON)...)
	}
	if in.Done != nil {
		if *in.Done {
			comment.ResolvedAt = &now
			comment.ResolvedBy = &userID
		} else {
			comment.ResolvedAt = nil
			comment.ResolvedBy = nil
		}
	}
	comment.UpdatedAt = now
	s.comments[libraryID] = comment
	copyComment := comment
	return &copyComment, nil
}

func (s *commentsStatefulSvc) DeleteDocumentComment(_ context.Context, _, _, _ string, libraryID int) error {
	if _, ok := s.comments[libraryID]; !ok {
		return domain.ErrCommentNotFound
	}
	delete(s.comments, libraryID)
	for id, c := range s.comments {
		if c.ParentLibraryID != nil && *c.ParentLibraryID == libraryID {
			delete(s.comments, id)
		}
	}
	return nil
}

func mustJSONEqual(t *testing.T, got, want json.RawMessage) {
	t.Helper()
	var gotAny any
	var wantAny any
	if err := json.Unmarshal(got, &gotAny); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal(want, &wantAny); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if !reflect.DeepEqual(gotAny, wantAny) {
		t.Fatalf("json mismatch want=%s got=%s", string(want), string(got))
	}
}

func newMuxWithCommentsSvc(t *testing.T, svc *commentsStatefulSvc) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	h := httphandler.NewHandler(svc).WithCaps(fakeCaps{admin: false})
	h.Mount(mux)
	return mux
}

func TestCreateComment_RoundTrip(t *testing.T) {
	svc := newCommentsStatefulSvc()
	mux := newMuxWithCommentsSvc(t, svc)

	content := json.RawMessage(`[{"type":"paragraph","children":[{"text":"hello"}]}]`)
	payload := []byte(`{"library_comment_id":42,"author_display":"Alice","content":[{"type":"paragraph","children":[{"text":"hello"}]}]}`)
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", bytes.NewReader(payload))
	withAuthHeaders(postReq, "editor")
	postRR := httptest.NewRecorder()
	mux.ServeHTTP(postRR, postReq)
	if postRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", postRR.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", nil)
	withAuthHeaders(getReq, "editor")
	getRR := httptest.NewRecorder()
	mux.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRR.Code)
	}

	var out []struct {
		LibraryCommentID int             `json:"library_comment_id"`
		Content          json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(getRR.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 || out[0].LibraryCommentID != 42 {
		t.Fatalf("unexpected comments list: %+v", out)
	}
	mustJSONEqual(t, out[0].Content, content)
}

// A3.4 cold-review Finding 4: createComment/updateComment switched to the
// generated request type, whose Content field is
// []documentsapi.DocumentCommentContentNode — each node a
// map[string]interface{}. Re-marshaling that value (as the pre-fix code
// did, via json.Marshal(req.Content)) round-trips every JSON number through
// float64, which cannot represent integers above 2^53 exactly. This test
// proves the STORED bytes (svc.comments[id].ContentJSON — the write path
// this PR touches) survive an integer beyond that boundary unchanged. It
// deliberately does not assert on the GET response: the read path
// (toCommentResponse/decodeCommentContent) was already lossy before A3.4
// and is unchanged by this PR's diff, so asserting exactness there would
// test something this fix round never touched.
func TestCreateComment_LargeIntegerContentPreservedExactly(t *testing.T) {
	svc := newCommentsStatefulSvc()
	mux := newMuxWithCommentsSvc(t, svc)

	// 2^53 + 1 = 9007199254740993, the smallest positive integer float64
	// cannot represent exactly (it rounds to 9007199254740992). A hidden
	// float64 hop anywhere in the write path corrupts this value.
	const bigInt = "9007199254740993"
	payload := []byte(`{"library_comment_id":4242,"author_display":"Alice","content":[{"type":"paragraph","external_id":` + bigInt + `}]}`)
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", bytes.NewReader(payload))
	withAuthHeaders(postReq, "editor")
	postRR := httptest.NewRecorder()
	mux.ServeHTTP(postRR, postReq)
	if postRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body: %s)", postRR.Code, postRR.Body.String())
	}

	stored, ok := svc.comments[4242]
	if !ok {
		t.Fatal("comment was not stored under library_comment_id 4242")
	}
	if !bytes.Contains(stored.ContentJSON, []byte(bigInt)) {
		t.Fatalf("stored ContentJSON lost integer precision: want it to contain %s, got %s", bigInt, string(stored.ContentJSON))
	}
}

func TestUpdateComment_LargeIntegerContentPreservedExactly(t *testing.T) {
	svc := newCommentsStatefulSvc()
	mux := newMuxWithCommentsSvc(t, svc)

	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", bytes.NewReader([]byte(`{"library_comment_id":55,"author_display":"Alice","content":[{"type":"paragraph","children":[{"text":"orig"}]}]}`)))
	withAuthHeaders(postReq, "editor")
	postRR := httptest.NewRecorder()
	mux.ServeHTTP(postRR, postReq)
	if postRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body: %s)", postRR.Code, postRR.Body.String())
	}

	const bigInt = "9007199254740993"
	patchPayload := []byte(`{"content":[{"type":"paragraph","external_id":` + bigInt + `}]}`)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments/55", bytes.NewReader(patchPayload))
	withAuthHeaders(patchReq, "editor")
	patchRR := httptest.NewRecorder()
	mux.ServeHTTP(patchRR, patchReq)
	if patchRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", patchRR.Code, patchRR.Body.String())
	}

	stored, ok := svc.comments[55]
	if !ok {
		t.Fatal("comment 55 vanished after update")
	}
	if !bytes.Contains(stored.ContentJSON, []byte(bigInt)) {
		t.Fatalf("stored ContentJSON lost integer precision after update: want it to contain %s, got %s", bigInt, string(stored.ContentJSON))
	}
}

func TestResolveComment_DerivedDoneField(t *testing.T) {
	svc := newCommentsStatefulSvc()
	mux := newMuxWithCommentsSvc(t, svc)

	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", bytes.NewReader([]byte(`{"library_comment_id":7,"author_display":"Alice","content":[{"type":"paragraph","children":[{"text":"todo"}]}]}`)))
	withAuthHeaders(postReq, "editor")
	postRR := httptest.NewRecorder()
	mux.ServeHTTP(postRR, postReq)
	if postRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", postRR.Code)
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments/7", bytes.NewReader([]byte(`{"done":true}`)))
	withAuthHeaders(patchReq, "editor")
	patchRR := httptest.NewRecorder()
	mux.ServeHTTP(patchRR, patchReq)
	if patchRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", patchRR.Code)
	}

	var patchOut struct {
		Done       bool       `json:"done"`
		ResolvedAt *time.Time `json:"resolved_at"`
	}
	if err := json.Unmarshal(patchRR.Body.Bytes(), &patchOut); err != nil {
		t.Fatalf("decode patch: %v", err)
	}
	if !patchOut.Done || patchOut.ResolvedAt == nil {
		t.Fatalf("expected done=true with resolved_at, got %+v", patchOut)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", nil)
	withAuthHeaders(getReq, "editor")
	getRR := httptest.NewRecorder()
	mux.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRR.Code)
	}
	var out []struct {
		Done       bool       `json:"done"`
		ResolvedAt *time.Time `json:"resolved_at"`
	}
	if err := json.Unmarshal(getRR.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(out) != 1 || !out[0].Done || out[0].ResolvedAt == nil {
		t.Fatalf("expected done=true with resolved_at set, got %+v", out)
	}
}

func TestReplyThread_ParentLibraryID(t *testing.T) {
	svc := newCommentsStatefulSvc()
	mux := newMuxWithCommentsSvc(t, svc)

	rootReq := httptest.NewRequest(http.MethodPost, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", bytes.NewReader([]byte(`{"library_comment_id":100,"author_display":"Alice","content":[{"type":"paragraph","children":[{"text":"root"}]}]}`)))
	withAuthHeaders(rootReq, "editor")
	rootReq = rootReq.WithContext(iamdomain.WithAuthContext(rootReq.Context(), "user_root", []iamdomain.Role{iamdomain.Role("editor")}))
	rootRR := httptest.NewRecorder()
	mux.ServeHTTP(rootRR, rootReq)
	if rootRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rootRR.Code)
	}

	replyReq := httptest.NewRequest(http.MethodPost, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", bytes.NewReader([]byte(`{"library_comment_id":101,"parent_library_id":100,"author_display":"Bob","content":[{"type":"paragraph","children":[{"text":"reply"}]}]}`)))
	withAuthHeaders(replyReq, "editor")
	replyReq = replyReq.WithContext(iamdomain.WithAuthContext(replyReq.Context(), "user_reply", []iamdomain.Role{iamdomain.Role("editor")}))
	replyRR := httptest.NewRecorder()
	mux.ServeHTTP(replyRR, replyReq)
	if replyRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", replyRR.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments", nil)
	withAuthHeaders(getReq, "editor")
	getRR := httptest.NewRecorder()
	mux.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRR.Code)
	}

	var out []struct {
		LibraryCommentID int    `json:"library_comment_id"`
		ParentLibraryID  *int   `json:"parent_library_id"`
		Author           string `json:"author"`
		CreatedAt        string `json:"created_at"`
	}
	if err := json.Unmarshal(getRR.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(out))
	}
	if out[0].LibraryCommentID != 100 || out[0].ParentLibraryID != nil || out[0].Author != "Alice" || out[0].CreatedAt == "" {
		t.Fatalf("unexpected root comment: %+v", out[0])
	}
	if out[1].LibraryCommentID != 101 || out[1].ParentLibraryID == nil || *out[1].ParentLibraryID != 100 || out[1].Author != "Bob" || out[1].CreatedAt == "" {
		t.Fatalf("unexpected reply comment: %+v", out[1])
	}
}

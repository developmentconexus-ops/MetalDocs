package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/problem"
)

// PDFWriter persists PDF-completion columns on documents.
type PDFWriter interface {
	WritePDF(ctx context.Context, tenant, docID, s3Key string, pdfHash []byte, generatedAt time.Time) error
	ResolveTenantByDocumentID(ctx context.Context, docID string) (string, error)
}

// NOTE (H-1e): relocated from documents/http during the delivery collapse. This internal HMAC webhook is documented as a live route (wiki/modules/documents.md) but its RegisterRoutes is not called anywhere — it is currently UNWIRED. Wiring it (and the docgen completion-callback integration) is a separate, security-sensitive decision flagged for operator review; behavior is preserved (route remains unregistered) in H-1e.
//
// PDFWebhookHandler receives completion callbacks from docgen_v2_pdf workers.
// Authentication is HMAC-SHA256 over the raw request body, shared secret in env.
type PDFWebhookHandler struct {
	writer PDFWriter
	secret string
}

func NewPDFWebhookHandler(w PDFWriter, secret string) *PDFWebhookHandler {
	return &PDFWebhookHandler{writer: w, secret: secret}
}

func (h *PDFWebhookHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/documents/{id}/pdf-complete", h.HandlePDFComplete)
}

type pdfCompleteBody struct {
	TenantID       string `json:"tenant_id"`
	FinalPDFS3Key  string `json:"final_pdf_s3_key"`
	PDFHash        string `json:"pdf_hash"`
	PDFGeneratedAt string `json:"pdf_generated_at"`
}

// pdfCompleteResponse is the hand-rolled typed 200 body for the internal HMAC
// webhook (M7 F7.4 / HS-6). The route stays off the OpenAPI spec (Phase C
// wont-fix) and is currently unwired, so it gets a hand-rolled typed struct per
// the ADR 0012 pre-codegen posture rather than a generated model — wire-identical
// to the prior {document_id, final_pdf_s3_key} literal.
type pdfCompleteResponse struct {
	DocumentID    string `json:"document_id"`
	FinalPDFS3Key string `json:"final_pdf_s3_key"`
}

const pdfWebhookMaxBytes = 64 << 10 // 64 KiB

func (h *PDFWebhookHandler) HandlePDFComplete(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, pdfWebhookMaxBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writePDFWebhookErr(w, http.StatusBadRequest, problem.CodeValidationError, "could not read request body")
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("X-Docgen-Signature")
	if !validSignature(raw, sig, h.secret) {
		writePDFWebhookErr(w, http.StatusUnauthorized, problem.CodeUnauthenticated, "invalid webhook signature")
		return
	}

	var body pdfCompleteBody
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&body); err != nil {
		writePDFWebhookErr(w, http.StatusBadRequest, problem.CodeValidationError, "invalid JSON body")
		return
	}
	// Normalize once at the boundary so the validated form and the persisted form
	// are identical — validating a trimmed key but storing the un-trimmed original
	// would let a leading-space key bypass the (trimming) prefix guard yet land in
	// the DB in a form that fails the (non-trimming) objectstore.assertTenant later.
	body.FinalPDFS3Key = strings.TrimSpace(body.FinalPDFS3Key)
	if !isValidFinalPDFS3Key(body.FinalPDFS3Key) {
		writePDFWebhookErr(w, http.StatusBadRequest, problem.CodeValidationError, "invalid final_pdf_s3_key")
		return
	}
	if body.PDFHash == "" || body.PDFGeneratedAt == "" {
		writePDFWebhookErr(w, http.StatusBadRequest, problem.CodeValidationError, "pdf_hash and pdf_generated_at are required")
		return
	}
	docID := r.PathValue("id")
	canonicalTenantID, err := h.writer.ResolveTenantByDocumentID(r.Context(), docID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writePDFWebhookErr(w, http.StatusNotFound, problem.CodeNotFound, "document not found")
			return
		}
		writePDFWebhookErr(w, http.StatusInternalServerError, problem.CodeInternalError, "failed to persist pdf completion")
		return
	}
	if strings.TrimSpace(body.TenantID) != "" && strings.TrimSpace(body.TenantID) != canonicalTenantID {
		writePDFWebhookErr(w, http.StatusBadRequest, problem.CodeValidationError, "tenant_id does not match document")
		return
	}
	// Full tenant-prefix guard: require the key to reside inside the canonical
	// tenant's namespace. Delegates to the objectstore kernel's single source of
	// truth for the prefix rule (rather than re-implementing it) so the webhook
	// and VerifiedStore.assertTenant cannot silently diverge. Must run AFTER
	// ResolveTenantByDocumentID so canonicalTenantID is known.
	if !objectstore.KeyHasTenantPrefix(canonicalTenantID, body.FinalPDFS3Key) {
		writePDFWebhookErr(w, http.StatusBadRequest, problem.CodeValidationError, "final_pdf_s3_key outside tenant scope")
		return
	}

	hashBytes, err := hex.DecodeString(body.PDFHash)
	if err != nil {
		writePDFWebhookErr(w, http.StatusBadRequest, problem.CodeValidationError, "invalid pdf_hash encoding")
		return
	}

	generatedAt, err := time.Parse(time.RFC3339, body.PDFGeneratedAt)
	if err != nil {
		writePDFWebhookErr(w, http.StatusBadRequest, problem.CodeValidationError, "invalid pdf_generated_at format")
		return
	}
	generatedAt = generatedAt.UTC()

	if err := h.writer.WritePDF(r.Context(), canonicalTenantID, docID, body.FinalPDFS3Key, hashBytes, generatedAt); err != nil {
		writePDFWebhookErr(w, http.StatusInternalServerError, problem.CodeInternalError, "failed to persist pdf completion")
		return
	}

	writeFillInJSON(w, http.StatusOK, pdfCompleteResponse{
		DocumentID:    docID,
		FinalPDFS3Key: body.FinalPDFS3Key,
	})
}

// writePDFWebhookErr emits a canonical RFC 9457 problem for the internal
// HMAC webhook. The route stays off the OpenAPI spec (Phase C wont-fix) but
// its error codes draw from the shared catalog.
func writePDFWebhookErr(w http.ResponseWriter, status int, code problem.Code, detail string) {
	_ = problem.Write(w, problem.New(status, code, code.String()).WithDetail(detail))
}

func validSignature(body []byte, sigHex, secret string) bool {
	if sigHex == "" || secret == "" {
		return false
	}
	expected, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(expected, mac.Sum(nil))
}

func isValidFinalPDFS3Key(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || len(trimmed) > 512 {
		return false
	}
	return !strings.Contains(trimmed, "..") && !strings.Contains(trimmed, "\x00")
}

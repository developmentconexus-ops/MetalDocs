package documentshttp

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
)

// PDFWriter persists PDF-completion columns on documents.
type PDFWriter interface {
	WritePDF(ctx context.Context, tenant, docID, s3Key string, pdfHash []byte, generatedAt time.Time) error
	ResolveTenantByDocumentID(ctx context.Context, docID string) (string, error)
}

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

const pdfWebhookMaxBytes = 64 << 10 // 64 KiB

func (h *PDFWebhookHandler) HandlePDFComplete(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, pdfWebhookMaxBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeFillInJSON(w, http.StatusBadRequest, map[string]any{"error": "read_body"})
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("X-Docgen-Signature")
	if !validSignature(raw, sig, h.secret) {
		writeFillInJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid_signature"})
		return
	}

	var body pdfCompleteBody
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&body); err != nil {
		writeFillInJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	if !isValidFinalPDFS3Key(body.FinalPDFS3Key) {
		writeFillInJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_final_pdf_s3_key"})
		return
	}
	if body.PDFHash == "" || body.PDFGeneratedAt == "" {
		writeFillInJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_fields"})
		return
	}
	docID := r.PathValue("id")
	canonicalTenantID, err := h.writer.ResolveTenantByDocumentID(r.Context(), docID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeFillInJSON(w, http.StatusNotFound, map[string]any{"error": "document_not_found"})
			return
		}
		writeFillInJSON(w, http.StatusInternalServerError, map[string]any{"error": "persist_failed"})
		return
	}
	if strings.TrimSpace(body.TenantID) != "" && strings.TrimSpace(body.TenantID) != canonicalTenantID {
		writeFillInJSON(w, http.StatusBadRequest, map[string]any{"error": "tenant_mismatch"})
		return
	}

	hashBytes, err := hex.DecodeString(body.PDFHash)
	if err != nil {
		writeFillInJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_pdf_hash"})
		return
	}

	generatedAt, err := time.Parse(time.RFC3339, body.PDFGeneratedAt)
	if err != nil {
		writeFillInJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_generated_at"})
		return
	}
	generatedAt = generatedAt.UTC()

	if err := h.writer.WritePDF(r.Context(), canonicalTenantID, docID, body.FinalPDFS3Key, hashBytes, generatedAt); err != nil {
		writeFillInJSON(w, http.StatusInternalServerError, map[string]any{"error": "persist_failed"})
		return
	}

	writeFillInJSON(w, http.StatusOK, map[string]any{
		"document_id":      docID,
		"final_pdf_s3_key": body.FinalPDFS3Key,
	})
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

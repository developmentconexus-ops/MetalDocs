package http

import (
	"net/http"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
)

type catalogEntry struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

var placeholderCatalog = []catalogEntry{
	{"doc_code", "Código do documento", "Código gerado automaticamente do documento controlado."},
	{"doc_title", "Título do documento", "Nome atual do documento."},
	{"revision_number", "Número da revisão", "Versão atual do documento."},
	{"author", "Autor", "Usuário que criou o documento."},
	{"effective_date", "Data efetiva", "Data efetiva (criação enquanto rascunho, data de aprovação após publicação)."},
	{"approvers", "Aprovadores", "Lista de aprovadores ou '[aguardando aprovação]'."},
	{"controlled_by_area", "Área controladora", "Nome da área de processo responsável."},
}

func (h *Handler) listPlaceholderCatalog(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateView)); err != nil {
		writeMappedErr(w, err)
		return
	}
	entries := make([]templatesapi.PlaceholderCatalogEntry, len(placeholderCatalog))
	for i, e := range placeholderCatalog {
		entries[i] = templatesapi.PlaceholderCatalogEntry{Key: e.Key, Label: e.Label, Description: e.Description}
	}
	writeJSON(w, http.StatusOK, templatesapi.PlaceholderCatalogResponse{Items: entries})
}

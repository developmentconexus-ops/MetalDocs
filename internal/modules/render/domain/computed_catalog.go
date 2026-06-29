// Package domain holds render's published, cross-module-readable contracts.
// ComputedCatalog is the SINGLE SOURCE OF TRUTH for computed (system) tokens:
// the set, their author-facing labels, and their author visibility. The render
// resolvers (internal/modules/render/resolvers) bind their behavior to these
// keys; templates derives /placeholder-catalog from this list. See ADR 0050.
package domain

// ComputedToken describes one system-filled token an author may reference.
type ComputedToken struct {
	Key           string // the {key} placeholder, e.g. "doc_code"
	Label         string // PT-BR, author-facing
	Description   string // PT-BR, author-facing
	AuthorVisible bool   // true => appears in the authoring palette
}

// ComputedCatalog returns the canonical, ordered computed-token set.
func ComputedCatalog() []ComputedToken {
	return []ComputedToken{
		{"doc_code", "Código do documento", "Código gerado automaticamente do documento controlado.", true},
		{"doc_title", "Título do documento", "Nome atual do documento.", true},
		{"revision_number", "Número da revisão", "Versão atual do documento.", true},
		{"author", "Autor", "Usuário que criou o documento.", true},
		{"effective_date", "Data efetiva", "Data efetiva (criação enquanto rascunho, data de aprovação após publicação).", true},
		{"approvers", "Aprovadores", "Lista de aprovadores ou '[aguardando aprovação]'.", true},
		{"controlled_by_area", "Área controladora", "Nome da área de processo responsável.", true},
		{"approval_date", "Data de aprovação", "Data de aprovação final do documento publicado, ou '[aguardando aprovação]'.", true},
	}
}

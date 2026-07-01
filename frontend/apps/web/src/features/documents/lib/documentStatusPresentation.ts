// Governed status copy for a controlled document, resolved for the shared
// controlled-artifact view model. Lifted verbatim from DocumentPublishedPage so
// the byte-for-byte governed copy is preserved. `ownerMeta` is intentionally
// DISTINCT from `subtitle` for scheduled/superseded (governed requirement).
//
// Standalone + tested (mirror of templateStatusPresentation) so a regression in
// per-status copy is caught by the suite rather than hiding inside the adapter.

const EM_DASH = '—';

export type StatusPresentation = {
  badgeLabel: string;
  subtitle: string | null;
  ownerMeta: string;
};

export function getDocumentStatusPresentation(status: string, publishedAt: string): StatusPresentation {
  const hasPublishedAt = publishedAt !== EM_DASH;

  switch (status) {
    case 'approved':
      return {
        badgeLabel: 'aprovado',
        subtitle: hasPublishedAt ? `aprovado em ${publishedAt}` : 'Aprovado',
        ownerMeta: hasPublishedAt ? `aprovado em ${publishedAt}` : 'aprovado',
      };
    case 'scheduled':
      return {
        badgeLabel: 'agendado',
        subtitle: hasPublishedAt ? `aprovação concluída em ${publishedAt}` : 'Publicação agendada',
        ownerMeta: hasPublishedAt ? `publicação agendada · aprovado em ${publishedAt}` : 'publicação agendada',
      };
    case 'published':
      return {
        badgeLabel: 'publicado',
        subtitle: hasPublishedAt ? `publicado em ${publishedAt}` : 'Publicado',
        ownerMeta: hasPublishedAt ? `publicado em ${publishedAt}` : 'publicado',
      };
    case 'superseded':
      return {
        badgeLabel: 'substituído',
        subtitle: hasPublishedAt ? `publicado em ${publishedAt}` : 'Substituído por revisão posterior',
        ownerMeta: hasPublishedAt ? `substituído · publicado em ${publishedAt}` : 'substituído por revisão posterior',
      };
    case 'obsolete':
      return {
        badgeLabel: 'obsoleto',
        subtitle: hasPublishedAt ? `obsoleto após publicação em ${publishedAt}` : 'Documento obsoleto',
        ownerMeta: hasPublishedAt ? `obsoleto · publicado em ${publishedAt}` : 'obsoleto',
      };
    case 'draft':
      return { badgeLabel: 'rascunho', subtitle: 'Rascunho — ainda não publicado', ownerMeta: 'rascunho' };
    case 'under_review':
      return { badgeLabel: 'em revisão', subtitle: 'Aguardando decisão de aprovação', ownerMeta: 'em revisão' };
    case 'rejected':
      return { badgeLabel: 'rejeitado', subtitle: 'Revisão rejeitada', ownerMeta: 'rejeitado' };
    default:
      return {
        badgeLabel: status || 'sem status',
        subtitle: hasPublishedAt ? `publicado em ${publishedAt}` : null,
        ownerMeta: hasPublishedAt ? `publicado em ${publishedAt}` : (status || EM_DASH),
      };
  }
}

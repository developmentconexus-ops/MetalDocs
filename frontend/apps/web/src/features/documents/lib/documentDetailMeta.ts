// Date formatters for DocumentPublishedPage. Pure, locale-aware (pt-BR).
// Inputs are RFC3339 strings (or null/undefined for missing data); outputs are
// display strings with em-dash fallback so the page never renders "Invalid Date".

const EM_DASH = '—';

export const DEFAULT_INITIAL_REVISION_TITLE = 'Criacao do documento';

function parseDate(input: string | null | undefined): Date | null {
  if (!input) return null;
  const d = new Date(input);
  return Number.isNaN(d.getTime()) ? null : d;
}

const longDateFmt = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: 'long',
  year: 'numeric',
});

const shortDateFmt = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
});

const dateTimeFmt = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
});

type ControlledDocumentVisibility = {
  scope: 'company' | 'restricted';
  areaCodes: string[];
  userIds: string[];
};

// "08 de maio de 2026" — used in published-doc owner banner
export function formatPublishedAt(input: string | null | undefined): string {
  const d = parseDate(input);
  return d ? longDateFmt.format(d) : EM_DASH;
}

// "08/05/2026" — used in KPI strip / version timeline
export function formatShortDate(input: string | null | undefined): string {
  const d = parseDate(input);
  return d ? shortDateFmt.format(d) : EM_DASH;
}

// "08/05/2026 14:36" — used in signoff timestamps
export function formatSignedAt(input: string | null | undefined): string {
  const d = parseDate(input);
  return d ? dateTimeFmt.format(d) : EM_DASH;
}

// Resolves profile code snapshot → human label using profiles list
export function resolveProfileLabel(code: string, profiles: Array<{ code: string; name: string }>): string {
  return profiles.find((p) => p.code === code)?.name ?? code;
}

// Resolves area code snapshot → human label using areas list
export function resolveAreaLabel(code: string, areas: Array<{ code: string; name: string }>): string {
  return areas.find((a) => a.code === code)?.name ?? code;
}

export function formatRevisionCode(revisionNumber: number | null | undefined): string {
  const safeRevisionNumber = typeof revisionNumber === 'number' && Number.isFinite(revisionNumber) ? revisionNumber : 0;
  return `REV${String(Math.max(safeRevisionNumber, 0)).padStart(2, '0')}`;
}

export function displayRevisionTitle(title: string | null | undefined, revisionCode: string): string {
  const trimmed = title?.trim();
  if (trimmed) return trimmed;
  return revisionCode === 'REV00' ? DEFAULT_INITIAL_REVISION_TITLE : EM_DASH;
}

export function formatRevisionStatus(status: string): string {
  switch (status) {
    case 'draft':
      return 'Draft';
    case 'under_review':
      return 'Em revisão';
    case 'approved':
      return 'Aprovado';
    case 'rejected':
      return 'Rejeitado';
    case 'scheduled':
      return 'Agendado';
    case 'published':
      return 'Publicada';
    case 'superseded':
      return 'Substituida';
    case 'obsolete':
      return 'Obsoleta';
    default:
      return status;
  }
}

export function buildVisibilityLabel(
  visibility: ControlledDocumentVisibility | null | undefined,
  areas: Array<{ code: string; name: string }>,
): string {
  if (!visibility) return EM_DASH;
  if (visibility.scope === 'company') return 'Toda empresa';
  if (visibility.userIds.length > 0) {
    return `Restrito (${visibility.userIds.length} usuarios, ${visibility.areaCodes.length} areas)`;
  }
  if (visibility.areaCodes.length === 0) return 'Restrito';
  const areaNames = visibility.areaCodes.map((code) => resolveAreaLabel(code, areas));
  return `Restrito a area ${areaNames.join(', ')}`;
}

// Signoff actor status → display config
export type SignoffStatus = 'pending' | 'approved' | 'rejected' | 'abstained';

export const SIGNOFF_STATUS_META: Record<SignoffStatus, { label: string; className: string }> = {
  pending:   { label: 'Aguardando', className: 'pending'  },
  approved:  { label: 'Aprovado',   className: 'approved' },
  rejected:  { label: 'Rejeitado',  className: 'rejected' },
  abstained: { label: 'Abstido',    className: 'abstained'},
};

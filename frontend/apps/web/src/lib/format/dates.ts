// Locale-aware (pt-BR) date formatters shared across controlled-artifact views.
// Inputs are RFC3339 strings (or null/undefined for missing data); outputs are
// display strings with em-dash fallback so views never render "Invalid Date".

const EM_DASH = '—';

function parseDate(input: string | null | undefined): Date | null {
  if (!input) return null;
  const d = new Date(input);
  return Number.isNaN(d.getTime()) ? null : d;
}

const shortDateFmt = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
});

// "08/05/2026"
export function formatShortDate(input: string | null | undefined): string {
  const d = parseDate(input);
  return d ? shortDateFmt.format(d) : EM_DASH;
}

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

const dateTimeFmt = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
});

const longDateFmt = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: 'long',
  year: 'numeric',
});

// "08/05/2026"
export function formatShortDate(input: string | null | undefined): string {
  const d = parseDate(input);
  return d ? shortDateFmt.format(d) : EM_DASH;
}

// "08/05/2026 14:36" — used in signoff timestamps
export function formatSignedAt(input: string | null | undefined): string {
  const d = parseDate(input);
  return d ? dateTimeFmt.format(d) : EM_DASH;
}

// "08 de maio de 2026" — used in published-doc owner banner
export function formatPublishedAt(input: string | null | undefined): string {
  const d = parseDate(input);
  return d ? longDateFmt.format(d) : EM_DASH;
}

const DAY_MS = 24 * 60 * 60 * 1000;

// F4 (M2c): relative due-date chip for the approval sidebar's StageContextHeader.
// null/invalid -> em-dash; future same calendar day -> "vence hoje"; future ->
// "vence em N dia(s)" (ceil of the day diff); past -> "atrasado há N dia(s)",
// overdue: true. `now` is injectable so tests are deterministic.
export function formatDueRelative(
  dueAt: string | null | undefined,
  now: number = Date.now(),
): { label: string; overdue: boolean } {
  const d = parseDate(dueAt);
  if (!d) {
    return { label: EM_DASH, overdue: false };
  }

  const diffMs = d.getTime() - now;

  if (diffMs <= 0) {
    const days = Math.max(1, Math.ceil(-diffMs / DAY_MS));
    return { label: `atrasado há ${days} dia${days === 1 ? '' : 's'}`, overdue: true };
  }

  const nowDate = new Date(now);
  const sameDay =
    d.getFullYear() === nowDate.getFullYear() &&
    d.getMonth() === nowDate.getMonth() &&
    d.getDate() === nowDate.getDate();

  if (sameDay) {
    return { label: 'vence hoje', overdue: false };
  }

  const days = Math.ceil(diffMs / DAY_MS);
  return { label: `vence em ${days} dia${days === 1 ? '' : 's'}`, overdue: false };
}

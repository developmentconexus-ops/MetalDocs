// formatNotificationTimestamp renders an ISO instant as a compact pt-BR relative
// label ("agora", "há 5 min", "há 2 h", "há 3 d"); anything older than ~7 days
// falls back to an absolute short date. Pure: `now` is injectable so the output
// is deterministic under test (no hidden Date.now() dependency).

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

const absoluteDate = new Intl.DateTimeFormat('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' });

export function formatNotificationTimestamp(iso: string, now: number = Date.now()): string {
  const then = Date.parse(iso);
  if (Number.isNaN(then)) return '—';

  const diff = now - then;
  if (diff < MINUTE) return 'agora';
  if (diff < HOUR) return `há ${Math.floor(diff / MINUTE)} min`;
  if (diff < DAY) return `há ${Math.floor(diff / HOUR)} h`;
  if (diff < 7 * DAY) return `há ${Math.floor(diff / DAY)} d`;
  return absoluteDate.format(new Date(then));
}

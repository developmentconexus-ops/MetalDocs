function toDate(value: string | Date | null | undefined): Date | null {
  if (!value) return null;
  const d = value instanceof Date ? value : new Date(value);
  return Number.isNaN(d.getTime()) ? null : d;
}

/** Returns YYYY-MM-DD string. Empty string if value is null/undefined/invalid. */
export function formatISODate(value: string | Date | null | undefined): string {
  const d = toDate(value);
  return d ? d.toISOString().slice(0, 10) : "";
}

/** Returns "YYYY-MM-DD HH:MM" string. Empty string if value is null/undefined/invalid. */
export function formatDateTime(value: string | Date | null | undefined): string {
  const d = toDate(value);
  return d ? `${d.toISOString().slice(0, 10)} ${d.toISOString().slice(11, 16)}` : "";
}

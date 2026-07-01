// Locale-aware (pt-BR) file-size formatter shared across controlled-artifact views.
// Binary units (1 KB = 1024 B), pt-BR decimal comma, em-dash fallback when the
// API returns null/undefined so views never render "NaN".

const EM_DASH = '—';

// "1 KB" / "2,5 MB" — e.g. the published-doc Tamanho fact.
export function formatFileSize(bytes: number | null | undefined): string {
  if (bytes == null || Number.isNaN(bytes) || bytes < 0) return EM_DASH;
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / Math.pow(1024, exp);
  // Whole numbers render without decimals; otherwise one pt-BR decimal place.
  const rounded = exp === 0 ? value : Math.round(value * 10) / 10;
  const text = Number.isInteger(rounded) ? String(rounded) : rounded.toLocaleString('pt-BR', { maximumFractionDigits: 1 });
  return `${text} ${units[exp]}`;
}

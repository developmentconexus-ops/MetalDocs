const rtf = new Intl.RelativeTimeFormat("pt-BR", { numeric: "auto" });

type RelativeTimeUnit = {
  unit: Intl.RelativeTimeFormatUnit;
  threshold: number;
};

const UNITS: readonly RelativeTimeUnit[] = [
  { unit: "year", threshold: 365 * 24 * 60 * 60 },
  { unit: "month", threshold: 30 * 24 * 60 * 60 },
  { unit: "day", threshold: 24 * 60 * 60 },
  { unit: "hour", threshold: 60 * 60 },
  { unit: "minute", threshold: 60 },
  { unit: "second", threshold: 1 },
];

export function getRelativeTime(iso: string): string {
  const diffSeconds = (new Date(iso).getTime() - Date.now()) / 1000;
  const absDiff = Math.abs(diffSeconds);

  for (const { unit, threshold } of UNITS) {
    if (absDiff >= threshold) {
      return rtf.format(Math.round(diffSeconds / threshold), unit);
    }
  }

  return rtf.format(0, "second");
}

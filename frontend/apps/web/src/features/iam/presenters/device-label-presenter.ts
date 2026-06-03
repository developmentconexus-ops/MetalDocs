// Parse a UA string into a pt-BR device label like "Chrome em macOS".
// Lightweight string sniff — no external lib. Order matters: check OS/browser
// tokens from most specific to least.

const UNKNOWN = "Dispositivo desconhecido";

type Match = { test: RegExp; label: string };

const BROWSERS: readonly Match[] = [
  { test: /Edg\//, label: "Edge" },
  { test: /OPR\/|Opera\//, label: "Opera" },
  { test: /Chrome\//, label: "Chrome" },
  { test: /Firefox\//, label: "Firefox" },
  { test: /Version\/[\d.]+.*Safari\//, label: "Safari" },
  { test: /Safari\//, label: "Safari" },
];

const PLATFORMS: readonly Match[] = [
  { test: /iPhone|iPad|iPod/, label: "iOS" },
  { test: /Android/, label: "Android" },
  { test: /Mac OS X|Macintosh/, label: "macOS" },
  { test: /Windows NT/, label: "Windows" },
  { test: /CrOS/, label: "ChromeOS" },
  { test: /Linux/, label: "Linux" },
];

function pick(ua: string, table: readonly Match[]): string | undefined {
  for (const m of table) {
    if (m.test.test(ua)) return m.label;
  }
  return undefined;
}

export function getDeviceLabel(userAgent: string | undefined | null): string {
  if (!userAgent) return UNKNOWN;
  const browser = pick(userAgent, BROWSERS);
  const platform = pick(userAgent, PLATFORMS);
  if (browser && platform) return `${browser} em ${platform}`;
  if (browser) return browser;
  if (platform) return platform;
  return UNKNOWN;
}

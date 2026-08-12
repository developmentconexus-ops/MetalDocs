import fs from 'node:fs';
import { execFileSync } from 'node:child_process';

const configPath = 'eslint.config.mjs';
const baseRef = process.env.CI_BASE_REF ?? 'origin/main';

function readBaseConfig() {
  try {
    return execFileSync('git', ['show', `${baseRef}:${configPath}`], { encoding: 'utf8' });
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    fail(`cannot read ${baseRef}:${configPath}: ${detail}`);
  }
}

function allowlistLines(source) {
  const marker = 'const ALLOWLIST = [';
  const start = source.indexOf(marker);
  if (start < 0 || source.indexOf(marker, start + marker.length) >= 0) {
    fail(`cannot locate exactly one ${marker} block in ${configPath}`);
  }

  const endMatch = /\n\s*\];/.exec(source.slice(start + marker.length));
  if (!endMatch) fail(`cannot find the end of ${marker} block in ${configPath}`);

  return source
    .slice(start + marker.length, start + marker.length + endMatch.index)
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
}

function fail(message) {
  console.error(`FE-BOUNDARY-ALLOWLIST: ${message}`);
  process.exit(1);
}

const currentConfig = fs.readFileSync(configPath, 'utf8');
const baseConfig = readBaseConfig();
const currentLines = allowlistLines(currentConfig);
const baseCounts = new Map();
for (const line of allowlistLines(baseConfig)) {
  baseCounts.set(line, (baseCounts.get(line) ?? 0) + 1);
}
const currentCounts = new Map();
for (const line of currentLines) {
  currentCounts.set(line, (currentCounts.get(line) ?? 0) + 1);
}
const addedLines = [...currentCounts.entries()]
  .filter(([line, count]) => count > (baseCounts.get(line) ?? 0))
  .map(([line]) => line);
if (addedLines.length > 0) {
  fail(`frontend boundary ALLOWLIST grew or changed representation; remove new lines: ${addedLines.join(', ')}`);
}
console.log(`fe-boundary-allowlist: clean (${currentLines.length} allowlist lines; shrink-only text ratchet)`);

import fs from 'node:fs';
import { execFileSync } from 'node:child_process';

const configPath = 'eslint.config.mjs';
const featuresRoot = process.env.FEATURES_ROOT ?? 'frontend/apps/web/src/features';
const baseRef = process.env.CI_BASE_REF ?? 'origin/main';
// `shared/` is the deliberate cross-cutting root under features/. It is not a
// feature and therefore must not receive a per-feature boundary block.
const unscopedFeatureDirs = new Set(['shared']);

function readBaseConfig() {
  try {
    return execFileSync('git', ['show', `${baseRef}:${configPath}`], { encoding: 'utf8' });
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    fail(`cannot read ${baseRef}:${configPath}: ${detail}`);
  }
}

function block(source, name) {
  const match = source.match(new RegExp(`const\\s+${name}\\s*=\\s*\\[[\\s\\S]*?\\];`));
  if (!match) fail(`cannot find ${name} in ${configPath}`);
  return match[0];
}

function featureNames(source) {
  const body = block(source, 'FEATURE_NAMES');
  return [...body.matchAll(/'([^']+)'/g)].map((match) => match[1]);
}

function allowlistPairs(source) {
  const body = block(source, 'ALLOWLIST');
  return [...body.matchAll(/from:\s*'([^']+)'\s*,\s*to:\s*'([^']+)'/g)]
    .map((match) => `${match[1]} -> ${match[2]}`);
}

function fail(message) {
  console.error(`FE-BOUNDARY-INTEGRITY: ${message}`);
  process.exit(1);
}

const currentConfig = fs.readFileSync(configPath, 'utf8');
const baseConfig = readBaseConfig();
const currentFeatures = featureNames(currentConfig);
const currentPairs = allowlistPairs(currentConfig);
const basePairs = new Set(allowlistPairs(baseConfig));

const duplicateFeatures = currentFeatures.filter((name, index) => currentFeatures.indexOf(name) !== index);
if (duplicateFeatures.length > 0) {
  fail(`FEATURE_NAMES contains duplicates: ${duplicateFeatures.join(', ')}`);
}

const duplicatePairs = currentPairs.filter((pair, index) => currentPairs.indexOf(pair) !== index);
if (duplicatePairs.length > 0) {
  fail(`ALLOWLIST contains duplicates: ${duplicatePairs.join(', ')}`);
}

const newPairs = currentPairs.filter((pair) => !basePairs.has(pair));
if (newPairs.length > 0) {
  fail(`frontend boundary ALLOWLIST grew; remove or refactor new entries: ${newPairs.join(', ')}`);
}

let discoveredFeatures;
try {
  discoveredFeatures = fs.readdirSync(featuresRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort();
} catch (error) {
  const detail = error instanceof Error ? error.message : String(error);
  fail(`cannot enumerate ${featuresRoot}: ${detail}`);
}

const declaredFeatures = [...new Set(currentFeatures)].sort();
const scopedDiscoveredFeatures = discoveredFeatures.filter((name) => !unscopedFeatureDirs.has(name));
const missingDeclarations = scopedDiscoveredFeatures.filter((name) => !declaredFeatures.includes(name));
const unexpectedUnscopedDirs = discoveredFeatures.filter(
  (name) => !declaredFeatures.includes(name) && !unscopedFeatureDirs.has(name),
);
const staleDeclarations = declaredFeatures.filter((name) => !scopedDiscoveredFeatures.includes(name));
if (missingDeclarations.length > 0 || unexpectedUnscopedDirs.length > 0 || staleDeclarations.length > 0) {
  fail([
    'FEATURE_NAMES does not match the feature directories',
    missingDeclarations.length > 0 ? `missing declarations: ${missingDeclarations.join(', ')}` : '',
    unexpectedUnscopedDirs.length > 0 ? `unexpected unscoped directories: ${unexpectedUnscopedDirs.join(', ')}` : '',
    staleDeclarations.length > 0 ? `stale declarations: ${staleDeclarations.join(', ')}` : '',
  ].filter(Boolean).join('; '));
}

const knownFeatures = new Set(declaredFeatures);
for (const pair of currentPairs) {
  const [from, to] = pair.split(' -> ');
  if (!knownFeatures.has(from) || !knownFeatures.has(to)) {
    fail(`ALLOWLIST references an unknown feature: ${pair}`);
  }
}

console.log(`fe-boundary-integrity: clean (${discoveredFeatures.length} feature directories, ${currentPairs.length} allowlisted pairs)`);

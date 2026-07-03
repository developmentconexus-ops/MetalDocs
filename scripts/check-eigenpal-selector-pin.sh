#!/usr/bin/env bash
# FE-06 T-003 — eigenpal `:global` selector + `!important` count guard.
#
# EditorChrome.module.css carries CSS overrides that reach into eigenpal's
# internal DOM via `:global(.ep-root ...)` selectors anchored on
# `data-testid` attributes and hardcoded SVG fill hex values. These
# selectors have no compile-time or runtime check against eigenpal's actual
# markup — if eigenpal renames an attribute, changes the doc-icon palette,
# or restructures the formatting bar, the overrides silently no-op instead
# of failing loudly.
#
# This guard does not (and cannot, from a static grep) verify the
# selectors still match eigenpal's real DOM. What it DOES do: pin the
# resolved eigenpal package version plus the current `:global(` and
# `!important` occurrence counts in EditorChrome.module.css, so that
# either (a) an eigenpal version bump or (b) any edit to the override
# block forces a human to consciously update this script and re-run the
# manual smoke checklist in wiki/references/eigenpal-controlled-package.md
# — instead of the coupling drifting unnoticed.
#
# Usage: bash scripts/check-eigenpal-selector-pin.sh
# Exit 0 = counts + version match the pin. Exit 1 = drift detected.

set -euo pipefail

WEB_ROOT="frontend/apps/web"
CHROME_CSS="$WEB_ROOT/src/features/shared/components/editor-chrome/EditorChrome.module.css"
ADAPTER_PKG="packages/eigenpal-adapter/package.json"

# ---------------------------------------------------------------------------
# Pin — update these together with EditorChrome.module.css's coupling-pin
# comment whenever the eigenpal override block or the eigenpal package
# version changes. Do not bump silently; re-run the manual smoke checklist.
# ---------------------------------------------------------------------------
EXPECTED_VERSION="1.9.0"
EXPECTED_GLOBAL_COUNT=22
EXPECTED_IMPORTANT_COUNT=28

fail=0

if [[ ! -f "$CHROME_CSS" ]]; then
  echo "check-eigenpal-selector-pin: $CHROME_CSS not found" >&2
  exit 1
fi

if [[ ! -f "$ADAPTER_PKG" ]]; then
  echo "check-eigenpal-selector-pin: $ADAPTER_PKG not found" >&2
  exit 1
fi

actual_version=$(grep -o '"@eigenpal/docx-editor-core": *"[^"]*"' "$ADAPTER_PKG" | grep -o '[0-9][0-9.]*' || true)
actual_global_count=$(grep -c ':global(' "$CHROME_CSS" || true)
actual_important_count=$(grep -c '!important' "$CHROME_CSS" || true)

if [[ "$actual_version" != "$EXPECTED_VERSION" ]]; then
  echo "check-eigenpal-selector-pin: EIGENPAL VERSION DRIFT"
  echo "  pinned:  $EXPECTED_VERSION"
  echo "  actual:  $actual_version (from $ADAPTER_PKG)"
  echo "  -> re-run wiki/references/eigenpal-controlled-package.md smoke checklist,"
  echo "     re-verify every :global(.ep-root ...) selector in $CHROME_CSS still"
  echo "     matches rendered DOM, then update EXPECTED_VERSION here and the"
  echo "     coupling-pin comment in $CHROME_CSS."
  fail=1
fi

if [[ "$actual_global_count" -ne "$EXPECTED_GLOBAL_COUNT" ]]; then
  echo "check-eigenpal-selector-pin: :global( SELECTOR COUNT DRIFT"
  echo "  pinned:  $EXPECTED_GLOBAL_COUNT"
  echo "  actual:  $actual_global_count (in $CHROME_CSS)"
  echo "  -> a selector was added or removed. If intentional, update"
  echo "     EXPECTED_GLOBAL_COUNT here after confirming the new/changed"
  echo "     selector still matches eigenpal's rendered DOM."
  fail=1
fi

if [[ "$actual_important_count" -ne "$EXPECTED_IMPORTANT_COUNT" ]]; then
  echo "check-eigenpal-selector-pin: !important COUNT DRIFT"
  echo "  pinned:  $EXPECTED_IMPORTANT_COUNT"
  echo "  actual:  $actual_important_count (in $CHROME_CSS)"
  echo "  -> update EXPECTED_IMPORTANT_COUNT here after confirming the change"
  echo "     is a deliberate eigenpal-override edit, not accidental drift."
  fail=1
fi

if (( fail == 1 )); then
  exit 1
fi

echo "check-eigenpal-selector-pin: clean (eigenpal@$actual_version, $actual_global_count :global( selectors, $actual_important_count !important)"

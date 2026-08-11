#!/usr/bin/env bash
# FE-18 — CSS design-token discipline grep guard.
# Forbids new raw hex color literals in frontend/apps/web *.module.css files.
# tokens.css (frontend/apps/web/src/styles/tokens.css) is the single source of
# truth for color values; module.css files must reference colors via
# var(--token-name) instead of hardcoding a new #hex literal.
#
# Usage: bash scripts/check-css-token-discipline.sh
# Exit 0 = clean. Exit 1 = violations found (each printed as path:line: RAW-HEX: <text>).
set -euo pipefail

WEB_ROOT="frontend/apps/web"

# ---------------------------------------------------------------------------
# Known-debt allowlist — pre-FE-18 raw hex tracked for future cleanup.
# This list MUST only shrink over time (grade-A simplification report FE-18).
# Additions require operator approval + reason. A file drops off this list
# once every raw hex in it has been tokenized (moved to tokens.css) or
# reduced to zero.
# ---------------------------------------------------------------------------
ALLOWLIST=(
  "src/components/ui/Dialog.module.css"
  "src/components/ui/Stepper.module.css"
  "src/features/approval/components/ApprovalTimelinePanel.module.css"
  "src/features/approval/components/DocumentApprovalExtras.module.css"
  "src/features/approval/components/LockBadge.module.css"
  "src/features/approval/components/StateBadge.module.css"
  "src/features/approval/components/SupersedePublishDialog.module.css"
  "src/features/approval/pages/InboxPage.module.css"
  "src/features/approval/pages/SignoffDetailPage.module.css"
  "src/features/approval/pages/route-admin/RouteAdmin.module.css"
  "src/features/documents/pages/DocumentDetailRoute.module.css"
  "src/features/documents/pages/styles/DocumentEditorPage.module.css"
  "src/features/iam/components/AuditEventsTable.module.css"
  "src/features/shared/components/editor-chrome/EditorChrome.module.css"
  "src/features/shared/components/editor-chrome/parts/VersionBadge.module.css"
  "src/features/shared/controlled-artifact/ArtifactApprovalFlow.module.css"
  "src/features/shared/controlled-artifact/ArtifactApprovalScreen.module.css"
  "src/features/shared/controlled-artifact/ArtifactDecisionPanel.module.css"
  "src/features/templates/components/MiniDocPreview.module.css"
  "src/features/templates/components/TemplateApprovalExtras.module.css"
  "src/features/templates/components/TemplateReviewCanvas.module.css"
  "src/features/templates/components/wizard/steps/StepConfirmation.module.css"
  "src/features/templates/pages/TemplateApprovalRoute.module.css"
  "src/features/tokens/components/TokenEditDialog.module.css"
  "src/features/tokens/components/TokenList.module.css"
  "src/features/tokens/pages/TokensRoutePage.module.css"
)

in_allowlist() {
  local file="$1"
  local item
  for item in "${ALLOWLIST[@]}"; do [[ "$file" == "$item" ]] && return 0; done
  return 1
}

# ---------------------------------------------------------------------------
# Discover scope: all *.module.css files under frontend/apps/web/src.
# tokens.css itself is the definition site for hex values and is exempt.
# ---------------------------------------------------------------------------
mapfile -t FILES < <(cd "$WEB_ROOT" && git ls-files 'src/**/*.module.css')

violations=0
report() { echo "$WEB_ROOT/$1:$2: RAW-HEX: $3"; violations=$((violations+1)); }

for f in "${FILES[@]}"; do
  if in_allowlist "$f"; then
    continue
  fi
  while IFS= read -r hit; do
    [[ -z "$hit" ]] && continue
    line=$(echo "$hit" | cut -d: -f1)
    text=$(echo "$hit" | cut -d: -f2-)
    report "$f" "$line" "$text"
  done < <(grep -nE '#[0-9a-fA-F]{3,8}' "$WEB_ROOT/$f" 2>/dev/null || true)
done

if (( violations > 0 )); then
  echo "css-tokens: $violations violation(s) found — see output above"
  echo "css-tokens: use var(--token-name) from src/styles/tokens.css instead of a raw hex literal"
  exit 1
fi
echo "css-tokens: clean (${#FILES[@]} module.css files checked, ${#ALLOWLIST[@]} grandfathered)"

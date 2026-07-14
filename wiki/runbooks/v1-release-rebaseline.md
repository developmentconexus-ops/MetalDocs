# v1 Release Re-baseline Runbook

> **Decision:** D-4b — permanent closure of F-18 git-history secret residual.
> **Last verified:** 2026-06-13
> **Operator:** run this cold, top-to-bottom, step numbers are checkboxes.

This runbook converts the current MetalDocs working repo into a single-commit, history-free v1.0.0 repo. All prior git history — including the 73 immutable `refs/pull/*` PR refs that hold the rewritten dev credential — is abandoned with the archived old repo. The maintained project starts clean.

---

## 1. Preconditions

Before touching anything, verify all three gates:

```powershell
# Gate 1: Wave Z DONE gate must be green.
# Confirm with the orchestrator that Z-33 closed with exit 0 on all
# G1–G9 checks (build, vet, suite, api-lint, cilint, runtime smoke).
# Expected evidence: commit docs(wave-z): Z-33 is in git log.

git -C C:\Users\leandro.theodoro\Documents\MetalDocs log --oneline -5
```

```powershell
# Gate 2: Working tree must be clean (no uncommitted edits, no untracked files).
git -C C:\Users\leandro.theodoro\Documents\MetalDocs status
# Expected: "nothing to commit, working tree clean"
```

```powershell
# Gate 3: Final build smoke on :8081.
.\scripts\start-api.ps1 -Build
# In a second terminal, confirm login returns HTTP 200:
Invoke-RestMethod -Method Post -Uri http://localhost:8081/api/v1/auth/login `
  -ContentType "application/json" `
  -Body '{"identifier":"admin","password":"AdminMetalDocs123!"}'
# Expected: response body with token field present.
# Stop the server before continuing.
```

Do not proceed if any gate fails.

---

## 2. Create the fresh repo directory

Run these commands from a parent directory that contains `MetalDocs\` (e.g. `C:\Users\leandro.theodoro\Documents\`):

```powershell
# From C:\Users\leandro.theodoro\Documents\
mkdir metaldocs-v1
Set-Location metaldocs-v1
git init -b main
```

Do not copy anything yet. The directory must exist and be a git repo before the copy step.

---

## 3. Copy only tracked working-tree files (honors .gitignore)

The precise method: `git ls-files` in the old repo enumerates ONLY files that are tracked by git. By definition this list excludes everything in `.gitignore` (`.env`, `bin/`, `logs/`, `node_modules/`, `.gitnexus/`, vendored binaries, build artifacts). No manual exclusion list needed.

Run this PowerShell loop **from inside `metaldocs-v1\`**:

```powershell
# Still inside C:\Users\leandro.theodoro\Documents\metaldocs-v1\
$src = "C:\Users\leandro.theodoro\Documents\MetalDocs"
$dst = "C:\Users\leandro.theodoro\Documents\metaldocs-v1"

git -C $src ls-files | ForEach-Object {
    $destFile = Join-Path $dst $_
    $destDir  = Split-Path $destFile -Parent
    New-Item -ItemType Directory -Path $destDir -Force | Out-Null
    Copy-Item -Path (Join-Path $src $_) -Destination $destFile -Force
}
```

Why this is safe: `git ls-files` only emits paths that are in the index of the old repo. Files that were never committed (`.env`, build outputs, editor caches, `.gitnexus/` graph data) are never listed, so they can never be copied into the new tree.

Spot-check after the loop:

```powershell
# Confirm a few key paths landed correctly:
Test-Path "$dst\go.mod"            # must be True
Test-Path "$dst\cmd\api\main.go"   # must be True

# Confirm gitignored junk did NOT land:
Test-Path "$dst\.env"              # must be False
Test-Path "$dst\bin"               # must be False
Test-Path "$dst\.gitnexus"         # must be False
```

---

## 4. Verify no secret in the new tree

Run gitleaks in working-tree mode (no git history to scan yet — there is none):

```powershell
# From the parent directory (C:\Users\leandro.theodoro\Documents\):
docker run --rm `
  -v "${PWD}\metaldocs-v1:/repo" `
  zricethezav/gitleaks:v8.24.3 `
  detect `
  --source /repo `
  --no-git `
  --config /repo/.gitleaks.toml `
  --verbose
```

Expected: `leaks found: 0` and exit code 0.

If gitleaks reports any finding: STOP. Do not commit. Investigate — either the finding is a test fixture covered by the existing `.gitleaks.toml` allowlist (in which case confirm and re-run without `--no-git` to use the allowlist) or it is a real secret that must be removed before continuing.

---

## 5. Initial commit and tag

```powershell
Set-Location C:\Users\leandro.theodoro\Documents\metaldocs-v1

git add -A
git commit -m "MetalDocs v1.0.0"
git tag v1.0.0
```

Verify:

```powershell
git log --oneline
# Expected: exactly 1 line — the v1.0.0 commit.

git tag
# Expected: v1.0.0
```

---

## 6. Create a NEW GitHub repo and push

Do NOT push to the existing `MetalDocs` GitHub remote. The old repo must be archived, not reused.

```powershell
# Create the new repo with gh CLI (replace ORG with the actual org/user):
gh repo create ORG/metaldocs-v1 --private --description "MetalDocs v1.0.0 (clean baseline)"

# Add as remote and push both main branch and the tag:
git remote add origin https://github.com/ORG/metaldocs-v1.git
git push -u origin main
git push origin v1.0.0
```

Verify on GitHub that the new repo shows exactly 1 commit on `main` and the `v1.0.0` tag.

---

## 7. Archive the old repo on GitHub (permanent freeze)

1. Go to the old repo on GitHub: `https://github.com/ORG/MetalDocs`.
2. Navigate to **Settings** → scroll to the **Danger Zone** section.
3. Click **Archive this repository** and confirm.

The old repo is now read-only. All branches, tags, and `refs/pull/*` PR refs remain intact and auditable but no new pushes are possible. The 73 PR refs that hold the rewritten dev credential are frozen in the archived repo — inaccessible for future development because all future work happens in the new repo.

**Never force-push or delete the old repo.** Archiving is the correct action. Deletion would destroy the immutable PR audit trail. Archiving freezes it and severs it from the active development surface.

---

## 8. Re-point CI and re-clone

After the archive, update all operational dependencies to point at the new repo:

```powershell
# For each CI secret / environment variable that referenced the old repo URL,
# update it to: https://github.com/ORG/metaldocs-v1.git

# GitHub Actions: update the repository secret GITHUB_REPOSITORY reference
# (if any workflows hard-code the repo name).

# Every developer re-clones fresh — no migration of local clones:
git clone https://github.com/ORG/metaldocs-v1.git
```

Update any external references (webhook URLs, deployment pipelines, Slack integrations) to the new repo URL before completing the release.

---

## 9. Post-release verification

```powershell
# In the new clone:
Set-Location C:\Users\leandro.theodoro\Documents\metaldocs-v1

# Check 1: history is exactly 1 commit.
git log --oneline
# Expected: 1 line only.

# Check 2: gitleaks full-history scan on the new repo (history = 1 commit,
# but this confirms no secret landed in that single commit):
docker run --rm `
  -v "${PWD}:/repo" `
  zricethezav/gitleaks:v8.24.3 `
  detect `
  --source /repo `
  --config /repo/.gitleaks.toml `
  --verbose
# Expected: leaks found: 0, exit 0.
```

Both checks must pass before the release is declared complete.

---

## Secret-history closure statement

This runbook is the physical execution of decision D-4b, which permanently closes the F-18 git-history secret residual. The Wave 0 filter-repo history rewrite (commit `abc...`) removed the leaked dev credential from all reachable refs in the old repo, but 73 immutable `refs/pull/*` PR refs were outside the rewrite scope (GitHub does not expose PR refs to `git filter-repo`). Those refs remain in the archived old repo. By creating a new repository whose first and only commit is the clean v1.0.0 working tree, all prior history — including the 73 PR refs — is permanently abandoned: they exist only in the archived read-only old repo, which is no longer the active development surface. The new repo has no ancestry connecting it to any leaked value. This satisfies D-4b as designed: the "fresh repo at first release" remedy documented in the Wave 0 close-out and confirmed in `wiki/references/current-agent-handoff.md`. No GitHub-support purge of PR refs is required or needed.

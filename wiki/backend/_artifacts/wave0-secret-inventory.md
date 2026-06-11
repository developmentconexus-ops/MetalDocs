# Wave 0 — Secret Exposure Inventory (F-18)

> **Last verified:** 2026-06-11 (live grep, integration branch)
> **Purpose:** The authoritative enumeration of every working-tree occurrence of the leaked dev DB password, with the correct action per file. Supersedes the partial list embedded in the Wave 0 spawn prompt (which named 7 of 11). This is the redaction checklist AND the delete-vs-keep guide; Wave 0 evidence is checked against it.
> **Method:** `grep -r "Lepa12"` over the working tree (secret referenced by location, never re-quoted here per rule D-4a).

## Ground rules
- The secret value itself is **never written in this file** (D-4a). It is the `PGPASSWORD` in `.env`.
- `.env` is correctly **gitignored and never committed** — it is the legitimate home and is NOT a finding.
- Rotation (user, item 0.3) is the true remediation; redaction + history rewrite remove the burned value from VCS.

## Inventory — 11 files, 4 categories

### A. Live hardcoded secret → DELETE the file (already dead code)
| File | Line | What | Action |
|---|---|---|---|
| `cmd/seed-test-document/main.go` | 25 | DSN with hardcoded password + hardcoded MinIO `minioadmin`/`minioadmin` | **DELETE** dir (item 0.1; dead binary, closes D-06) |
| `scripts/start-spec1-api.ps1` | 15 | `SetEnvironmentVariable('PGPASSWORD','<secret>')` — **second live hardcode** | **DELETE** file (confirmed dead; was already on the 0.1 delete list) |

### B. Live ACTIVE scripts → REDACT comment only, DO NOT DELETE
These canonical scripts correctly read the password from `.env`; the secret only appears in an explanatory comment. They are referenced by the startup policy — deleting them breaks dev startup.
| File | Line | What | Action |
|---|---|---|---|
| `scripts/start-worker.ps1` | 5 | comment quotes the value to explain `.env` split-parsing | **REDACT** comment → use a placeholder; keep file |
| `scripts/start-api-no-build.ps1` | 5 | same explanatory comment | **REDACT** comment → use a placeholder; keep file |

### C. Documentation → REDACT value to `<redacted — see .env>`
| File | Note |
|---|---|
| `wiki/references/local-dev-startup.md` | dev-startup guide; keep the *bash-corruption lesson*, redact the literal |
| `wiki/references/local-dev-credentials.md` | dev creds doc; redact literal |
| `wiki/backend/legacy-register.md` | F-18 entry |
| `wiki/backend/stage2-evaluation.md` | P0-1 section |
| `wiki/backend/_artifacts/stage2/security-secrets.md` | theme artifact (multiple lines) |
| `wiki/backend/_artifacts/stage1/synthesis-legacy.md` | Stage-1 synthesis |
| `wiki/backend/_artifacts/stage1/repo-topology.md` | line 552 — **was missing from the spawn list** |

### D. Legitimate (NOT a finding) — leave as-is
| File | Why |
|---|---|
| `.env` | gitignored, never committed; the secret's correct home |

## Acceptance for Wave 0 item 0.1 / 0.4
- After redaction (Cat. A delete + B/C redact): `grep -r "Lepa12" .` returns **only `.env`** (working tree).
- After history rewrite (0.4, post-rotation): `git log --all -S "<secret>"` returns **empty**.
- gitleaks (item 0.2) run over the working tree: **clean**.

## Why this file exists
The Stage-1/2 audit propagated the secret into wiki docs (the D-4a violation that motivated rule D-4a), and the Wave 0 hand-off list was memory-based and incomplete (7/11). Grep is the authority; this inventory captures its result durably so the redaction is provably complete and no active script is deleted by mistake.

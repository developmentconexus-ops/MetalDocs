# F0.1 — VersionRef contract cutover · Evidence

Feature: version pointers travel as nested value objects on the wire (ADR 0065).
Contract of record: `../validation-contract.md`. Commits: `d0b1ba84` (backend
spec+generated+domain+repo+delivery+tests), `15c0eeeb` (FE consumers).

## Gate commands (real output)

### Backend build / vet
- `go build ./...` → exit 0 (whole repo compiles on the read/write split).
- `go vet ./internal/modules/templates/...` → clean, exit 0.
- `go vet -tags=integration ./...` → `DONE_EXIT 0` (whole repo, integration-tagged; catches every cross-module consumer of the changed `Repository` return types).

### Backend tests (targeted per touched pkg — full integration suite deferred, see Defers)
```
go test ./internal/modules/templates/... -count=1
ok  metaldocs/internal/modules/templates/application       2.378s
ok  metaldocs/internal/modules/templates/delivery/http     3.483s
ok  metaldocs/internal/modules/templates/domain            1.600s
ok  metaldocs/internal/modules/templates/infrastructure    1.380s
ok  metaldocs/internal/modules/templates/repository        2.255s
```

### Pin guard (ADR 0065 contract guard — `delivery/http/template_dto_nullable_fields_test.go`)
Rewritten to P1–P4 (validation-contract §2), green inside the `delivery/http` run above:
- **P1** present-and-null: `string(raw["published_version"]) == "null"` for a never-published template.
- **P2** nested ref field-set is exactly `{id, number, revision_number, status}`; `latest_version.status == "under_review"`.
- **P3** four removed flat keys (`latest_revision_number`, `published_version_id`, `published_version_number`, `current_revision_number`) absent.
- **P4** published → full ref object, `status == "published"`, all four keys present.

### Frontend
- `grep -rn "published_version_id\|current_revision_number\|latest_revision_number\|published_version_number" src --include=*.ts --include=*.tsx | grep -v api-types` → **0 hits** (incl. test fixtures).
- `pnpm exec tsc --noEmit` → clean, exit 0 (tooling ran; no junction-drift crash this session).
- `pnpm exec vitest run src/features/documents src/features/templates src/features/taxonomy` → **56 files / 369 tests passed**.

## Runtime proof (live QA drive — mission D4, runtime-visible milestone)

Started `.\scripts\start-api.ps1 -Build` (PowerShell; never bash/source .env). API on `:8081`.
Cookie-session login `POST /api/v1/auth/login {identifier:"admin", …}` (wiki-documented seed creds, not `.env`).

**Drive 1 — `GET /api/v1/templates`** (envelope `data.templates`, 10 items):
| Assertion | Result |
|---|---|
| removed-key hits across all items | **0** |
| `published_version` key missing (violates present-and-null) | **0** |
| `published_version == null` (never-published) | 6 |
| `published_version` full object with `status == "published"` | 4 |
| `latest_version` exact shape `{id,number,revision_number,status}` | **10 of 10** |

→ **PASS.** Live confirmation the 9f86828b present-and-null guarantee carries forward on the nested shape.

**Drive 2 — `GET /api/v1/documents`**: `200`, envelope `{items,page,total}` (DocumentSummary) unchanged from pre-M0, 0 version-ref leakage → **PASS** (proves no scope drift into documents; documents is Plan 2).

**Drive 3 — `/documents/new` wizard Step 3** (preview browser, vite proxy → :8081):
DOM/a11y assertion over all 11 template cards (badge keyed off `latest_version.status`):
| Card badge (from `latest_version.status`) | count | `aria-disabled` | click → `aria-checked` |
|---|---|---|---|
| `publicada` (published_version != null) | 4 | false | **true** (selectable) |
| `em revisão` (under_review) | 2 | true | stays false |
| `sem versão publicada` (other, unpublished) | 3+1 | true | stays false |
| blank-document option | 1 | false | selectable |

→ **PASS.** Exactly the 9f86828b regression, proven closed at runtime: every published card selectable, every unpublished card disabled with the status-precise badge. Selection of a published card sets `aria-checked=true`; a disabled card cannot be checked.

Fixture/real label: **real provider** — live Postgres dev seed + real API + real SPA build. Not mocked.

## Review / QA disposition
Contract-first discipline held: every wire change via `api/openapi/v1/openapi.yaml` → `oapi-codegen` (Go) + `openapi-typescript` (FE); zero hand-edits to generated files. Test fixtures retargeted to the read model with no intent weakened (assertions moved to persisted repo state / nested ref fields; two dropped-scalar assertions removed only because the field no longer exists and `PublishedVersionID` already covers the same intent).

## Bounded defers (with triggers)
- **Full `go test ./...` integration suite** not run here (20+ min box constraint). Trigger to run: milestone-validator clean-state gate, and CI. Mitigated by `go vet -tags=integration ./...` clean + targeted per-pkg tests green.
- **`preview_screenshot`** timed out (renderer capture stuck; no page/console errors, `preview_eval` fully responsive). Runtime proof captured via a11y-tree + DOM-state assertions (stronger than pixels per preview-tool guidance). Trigger to retry: if a visual-only regression is suspected.
- **documents `DocumentRevisionRef`** (same pattern, documents module) is **Plan 2**, deferred pre-v1 — see `../../` Plan 2 doc + `developing-new-work` gate.

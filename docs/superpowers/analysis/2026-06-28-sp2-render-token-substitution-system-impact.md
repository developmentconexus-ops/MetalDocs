# System-impact analysis — Render Token Substitution (SP-2)

**Date:** 2026-06-28
**Intent (one line):** Wire render substitution — merge the tenant token dictionary (`tokens` module) into the freeze-time substitution map alongside the computed-token catalog (`templates`/`render`), substituting dictionary tokens into rendered output, honoring the TD-1 collision contract.
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow

> Same ten sections for module and feature work. Module-only rows are marked **N/A** with a one-line reason — the question was asked and answered, not skipped.

---

## 0. Runtime-truth correction (read first)

The SP-2 brief, the `tokens` wiki doc, and the SP-1 system-impact analysis all assert the owning module is **`render`** ("render-fanout") and that the injection target is `FanoutRequest.ResolvedValues`. **Runtime truth contradicts both** (CLAUDE.md runtime-truth rule — classify and surface, do not patch around):

1. **The substitution map is assembled in `documents`, not `render`.** The eigenpal name→value substitution map is built in `internal/modules/documents/application/freeze_service.go` — `Materialize` (lines 261-284, async path) and `Freeze` (lines 326-354, legacy sync). A third assembly site rebuilds it for forensic re-render: `internal/modules/documents/repository/resolver_readers.go` `ReadForReconstruction` (lines 134-147). `render` owns the resolver *catalog* (`internal/modules/render/resolvers/`) and the docx-renderer *transport* (`internal/modules/render/fanout/`), but it does **not** assemble the substitution map. `documents` already imports `render/fanout` + `render/resolvers` (freeze_service.go:12-13).

2. **Dictionary tokens substitute by NAME → they belong in `PlaceholderValues` (name→string), not `ResolvedValues`.** `FanoutRequest.PlaceholderValues map[string]string` (`render/fanout/client.go:17`) is the eigenpal *variable substitution* map (a `{NAME}` in the body docx is replaced by name). `ResolvedValues map[string]any` (client.go:19) is the ID-keyed map for subblock *composition* plugins. Flat dictionary constants (`{COMPANY_NAME}`) are variable substitutions → `PlaceholderValues`. The SP-1 analysis's "inject into `ResolvedValues`" is wrong.

These corrections drive the rest of this analysis. Neither is an invariant violation; both are owning-module / integration-seam facts that change where the code lands.

---

## 1. Classify & own
*(CLAUDE.md Orientation rule)*

- **Work type:** feature — wires an existing published port (`tokens/domain.DictionaryReader`) into the existing freeze/materialize substitution pipeline. No new module, no new table, no new capability. The two catalogs and the assembly pipeline already exist; SP-2 adds a merge step + a collision policy.

- **Owning module(s):** `internal/modules/documents` — specifically `application/freeze_service.go` (the substitution-map assembly site) and `repository/resolver_readers.go` (the reconstruction assembly site). This is where the `name→value` map is built and where the new `documents → tokens` consumption edge lands. The merge + collision policy is new logic owned here.
  - **Candidate home for the pure merge primitive:** `internal/modules/render`. The merge itself — `(computed/user name→value map, dictionary name→value, precedence policy) → (merged map, collision report)` — is a pure function conceptually belonging to "how tokens become a substitution map" (render's domain). If placed in `render` it takes plain maps + a policy enum and stays **free of any `tokens` import** (render must not depend on tokens — see §3 invariant 6). `documents` would read the dictionary via `tokens.DictionaryReader`, flatten to a plain map, and call the render merge primitive. To be decided in brainstorming (§10 locked-constraint set carries both options).

- **Explicitly NOT owning:**
  - `tokens` — **publishes** `DictionaryReader` and is consumed read-only. SP-2 requires **zero** change to the `tokens` module (the port is stable — confirmed `domain/port.go`). If SP-2 finds itself editing `tokens`, that is a boundary smell.
  - `render` — owns the resolver catalog + fanout transport. It may *host* the pure merge primitive (candidate above) but does **not** assemble the per-document map (that is `documents`). It must never import `tokens`.
  - `templates` — owns `placeholder_schema` (the computed/declared catalog) read at freeze time via `SchemaReader.LoadPlaceholderSchema`. It is the *source* of the names dictionary tokens can collide with, but owns no merge logic.
  - `iam` / `audit` — untouched; no new capability, no new audited event (substitution is internal to an already-authorized freeze).

- **Cross-module edges (with direction):** `A → B` = A depends on B, through B's published interface only.
  - `documents → tokens` — **NEW.** `documents` reads the dictionary via `tokens/domain.DictionaryReader` (`GetByName` / `List`, no tx). Through the published port only; never `token_dictionary_entries` SQL. `tokens` does not depend on `documents`. No cycle.
  - `documents → render` — **EXISTING** (freeze_service.go:12-13 imports `render/fanout` + `render/resolvers`). If the merge primitive lands in `render`, this edge gains one pure function; still one-directional.
  - `documents → templates` — **EXISTING** (`tmpldom.Placeholder` schema type; `SchemaReader`). Unchanged.
  - `render → tokens` — **MUST NOT EXIST.** If the merge primitive is hosted in `render`, it must take plain maps, not a `tokens` type. Locked.

- **Ambiguity?** The brief named `render` as owner; runtime truth says `documents` is the assembly owner with `render` as merge-primitive candidate. **AS-3 raised and RESOLVED** via targeted verify (freeze_service.go + resolver_readers.go + client.go read directly). Owning module = `documents`; resolution recorded here. Not an open hard-stop.

---

## 2. Foundation verdict
*(Global-Maximum rule)*

- **Base you'd build on:** The freeze/materialize pipeline (ADR 0015 async split: `Pin` in-tx → `Materialize` async) + the `values_hash` determinism anchor + the forensic `ReconstructService`. Grade-A audited, signed off 2026-06-21. Sound, not a patch.

- **Sound, or legacy/patch/workaround?** Sound base. But there is a **local-maximum trap to avoid**: the `name→value` assembly logic is **duplicated across 3 sites** (`Materialize`, legacy `Freeze`, `ReadForReconstruction`), and the three are **already inconsistent** — `Materialize`/`Freeze` key `PlaceholderValues` by placeholder *name* (`idToName`), while `ReadForReconstruction` keys by placeholder *ID* (resolver_readers.go:137). Inlining a dictionary merge at each of the 3 sites would (a) triplicate the TD-1 collision policy and (b) deepen the existing drift. **Pre-existing drift is out of SP-2 scope** (legacy drive-by rule — repair only contract/invariant guards), but SP-2 must not *add* to it.

- **Global-maximum structure (name it):** a **single pure merge primitive** — `MergeDictionary(base map[string]string, dict map[string]string, policy Policy) (merged map[string]string, collisions []Collision)` — called at every site that builds the substitution map, rather than 3 inline merges. This is the framework boundary, not a one-off tweak. Trade-off: one new small unit + its tests vs. zero new files but triplicated policy and a third drift vector. The primitive is the global maximum; choose it. **No AS-2** — the foundation is sound; this is a directive for *how* to build on it, not a stop.

---

## 3. Invariant alignment
*(the 6 non-negotiables)*

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | **Yes (indirectly)** | No new capability. Substitution runs inside an already tier-1+tier-2 authorized freeze/materialize. The new dictionary read goes through `tokens.DictionaryReader`, which **internally** enforces `token.view` (its own `DoReadOnly` + `SeedTxIdentity` + `authz.Require`). **Open question:** the materialize actor is a background worker (system context) — confirm it carries `token.view`, or the reader must run under a system/service identity that is authorized. Carried to design. | `authz.Require` is already inside `DictionaryReader`; no new call site in `documents`. |
| Contract-first (OpenAPI + oapi-codegen) | **No** | SP-2 changes no HTTP route. The dictionary read is a Go in-process port call; the docx-renderer payload (`FanoutRequest`) is an internal service contract, not the public OpenAPI. If the optional TD-1 write-guard (warn on dictionary write that collides with a schema) is adopted, it adds a *response field/header* to existing `POST/PUT /api/v1/tokens` → that **is** a spec edit (expand-only, non-breaking). | `api/openapi/v1/openapi.yaml` only if write-guard adopted. |
| Multi-tenant pooled (`tenant_id` / tx-local GUC / 404 cross-tenant) | **Yes** | Dictionary is read for the *current* tenant only: `DictionaryReader.List(ctx, tenantID)` with `tenantID` from the freeze flow (already `tenant.FromContext`-derived upstream). No cross-tenant merge possible — the reader predicates on `tenant_id`. | `tenant.FromContext` (already upstream); `DictionaryReader` predicates internally. |
| Async = transactional outbox | **Yes (constraint, not new outbox)** | Substitution already lives on the correct side of the async split: `Materialize` runs **after** Pin, off the business tx, in the worker. No new network call is added inside a business tx. The dictionary read is a DB read, not an external side effect → no outbox needed. **H-PRE-1 lock:** if any dictionary read is added at **Pin** time (in-tx), it must be hoisted **off** the pin tx (DictionaryReader opens its own authz-recording tx; calling it while holding the pin tx risks the H-PRE-1 deadlock). Reading at `Materialize` is H-PRE-1-safe (no business tx held). | Outbox already in place (`render/fanout/staging_outbox.go`); H-PRE-1 (advisory-lock-deadlock-constraint). |
| DB enforces invariants (triggers/constraints) | **No new DB invariant** | Merge/collision is application policy, not a storable invariant (the collision is cross-catalog and only exists at render-merge time; the dictionary-write path has no visibility into a tenant's schemas — per TD-1). The optional write-guard is a **warning**, never a DB constraint. No migration. | N/A — no new table/constraint. |
| Cross-module via published interface only | **Yes (central to SP-2)** | `documents → tokens` strictly via `tokens/domain.DictionaryReader`. No `token_dictionary_entries` SQL outside `tokens`. If merge primitive lands in `render`, it takes plain maps (no `tokens` import) so `render` stays decoupled. | `tokens/domain/port.go` `DictionaryReader` (provider); inject into `FreezeService` as a consumer port in `application/ports.go` style. |

No invariant violation. **No AS-1.** Two items carried to design as locked constraints: (a) materialize/worker actor must be authorized for `token.view`; (b) H-PRE-1 — dictionary read off any business tx.

---

## 4. Capability wiring
*(N/A for the new-capability path — SP-2 adds **no** capability.)*

No const/classify/tier-1/seed/tripwire/registry-bump work. `token.view` (read) and `token_dictionary.manage` (write) already exist from SP-1 and fully gate the dictionary. `TestCapabilityRegistrySize` is **untouched** (stays at its post-SP-1 value). The only capability *consideration* is the runtime actor question in §3 (does the materialize worker context satisfy `token.view`) — a wiring-of-identity question, not a new-capability question.

---

## 5. Module wiring
*(N/A — SP-2 births no module.)* The `tokens` module already exists (SP-1) and is consumed unchanged. The only composition-root touch is wiring `tokens.Module.Reader` (a `DictionaryReader`) into the `documents` `FreezeService` constructor + `ReconstructService` inputs at `apps/api/cmd/metaldocs-api/main.go` (the freeze-service wiring block, ~main.go:842-849 per the render map). That is dependency injection of an existing port, not module birth.

---

## 6. Frameworks to reuse, not reinvent

| Platform primitive | SP-2 use | Confirm reuse |
|---|---|---|
| `tokens/domain.DictionaryReader` — `internal/modules/tokens/domain/port.go:22` | Read the tenant dictionary (`List`) at substitution-map assembly | Yes — the published SP-2 consumption seam; consume, don't reimplement |
| `render/fanout.FanoutRequest` / `Client` — `internal/modules/render/fanout/client.go` | Carries `PlaceholderValues` to docx-renderer; merged dictionary lands here | Yes — no transport change; same struct |
| `render/resolvers.Registry` — `internal/modules/render/resolvers/registry.go` | Computed catalog; produces the names dictionary tokens may collide with | Yes — unchanged; SP-2 reads its output, not its internals |
| `TxRunner` (`Do`/`DoReadOnly`) — `internal/platform/db/runner.go:21` | Already owns the freeze tx boundary; dictionary read uses `DictionaryReader`'s own `DoReadOnly` | Yes — no hand-rolled tx |
| `tenant.FromContext` — `internal/platform/tenant/context.go:27` | Tenant already derived upstream in the freeze flow | Yes |
| `problem.New`/`Write` — `internal/platform/problem/problem.go:77` | If render-time collision → reject path, return RFC 9457 (e.g. 409/422) | Yes — no bare `http.Error`; reuse existing `fanout.RenderError` classification where it fits |
| `strictjson.Decode` — `internal/platform/strictjson` (promoted SP-1) | Only if write-guard adds request fields to `POST/PUT /tokens` | Yes — canonical strict decode; **note:** tokens handler already uses it (TD-2 cleanup) |
| `testdb.Open` + factory builders — `tests/integration/testdb/` | Integration tests for merge + collision at freeze time | Yes — `Open(t)`, `SeedWithCaps`, `Qualified`; `//go:build integration` |

**Global-maximum primitive to ADD (not reinvent — none exists):** the pure `MergeDictionary` collision primitive (§2). No existing platform row covers "merge two token catalogs with a precedence + collision policy" — this is a genuinely new, small, cross-cutting concern. Surfacing it here per the frameworks-catalog rule (new concern ⇒ design a primitive, don't inline a one-off).

---

## 7. Contract & data

- **OpenAPI-first:** No route change in the core scope. The internal docx-renderer `FanoutRequest` contract is unchanged in *shape* — the merged dictionary occupies the existing `PlaceholderValues` map. **Optional TD-1 write-guard** (decision §10): if `POST/PUT /api/v1/tokens` should *warn* (not block) when a created/updated dictionary name collides with one of the tenant's active `placeholder_schema` names, that adds a non-breaking advisory field to the existing token responses → edit `api/openapi/v1/openapi.yaml` + regenerate `tokens` codegen. Expand-only; no breaking change.
- **Migration:** **None.** No new table, no new column, no constraint — *unless* the reproducibility decision (§10) chooses **pin-time** dictionary capture, which may require storing dictionary-sourced values (e.g. a `source = 'dictionary'` value row, or a snapshot column) so `values_hash` covers them and `ReadForReconstruction` reads them back. That is a migration **iff** pin-time is chosen. Flagged for the design decision, not assumed.
- **Destructive change?** No. All additive. Substitution behavior is currently a no-op for dictionary names (they simply don't resolve today), so adding them can only *add* substitutions — but see the **collision** caveat: a dictionary name equal to an existing computed/user placeholder name changes output for that name. That is exactly what the TD-1 precedence rule governs and why conflict detection matters.

---

## 8. Test & QA plan

**Canonical framework:** `testdb` integration factory; `//go:build integration`; R1–R4 discipline (`scripts/check-test-discipline.sh`). Unit tests for the pure merge primitive (table-driven, no DB).

**Which of the 6 QA gates apply (feature subset):**

| Gate | Applies? | Notes |
|------|----------|-------|
| Contract | Only if write-guard adopted | Then: `go generate` clean; generated DTO carries the advisory field. Otherwise N/A (no route change). |
| AuthZ | **Yes** | Dictionary read is `token.view`-gated inside `DictionaryReader`. Test: materialize/worker actor authorized; an unauthorized read path fails closed (does not silently skip the dictionary). |
| Multi-tenant isolation | **Yes** | Tenant A's freeze merges **only** tenant A's dictionary; tenant B entries never appear. Integration with two tenants. |
| Async/idempotency | **Yes** | Merge is deterministic and side-effect-free; re-running `Materialize` / `Reconstruct` yields the identical merged map. Ties to the reproducibility decision (§10). |
| DB-invariant | Only if pin-time storage chosen | Then: stored dictionary values reproduce on reconstruction. Otherwise N/A. |
| Docs | **Yes** | Update `wiki/modules/tokens.md` (SP-2 status), `wiki/modules/render*.md`, `wiki/concepts/placeholders.md`; resolve TD-1 in `tokens-tech-debt.md`; refresh `Last verified`. |

**Hard gates (per task + at close):**
- `go build ./...`
- `go test ./...`
- `go test -tags=integration ./internal/modules/render/...`
- `go test -tags=integration ./internal/modules/documents/...` (the real assembly site)
- `go test -tags=integration ./internal/modules/tokens/...`
- `.\scripts\check-system-runnable.ps1` (PowerShell; never bash/`source .env`)

**Evidence shape:** commands + outcomes + review/QA disposition (two-stage: spec-compliance then code-quality) + bounded defers. No bare "done".

---

## 9. Docs / ADR
*(N/A for new-module-doc creation — feature updates existing docs.)*

**Wiki (feature work — update + re-stamp, don't create a module doc):**
- `wiki/modules/tokens.md` — flip SP-2 status (render substitution wired); fix the §3 diagram's "render-fanout (SP-2)" external actor to reflect the real `documents`-assembles / `render`-transports split; refresh `Last verified`.
- `wiki/modules/tokens-tech-debt.md` — **resolve TD-1** (record the chosen precedence + conflict policy + write-guard decision); also fold the TD-2 cleanup (tokens-handler half resolved — `strictjson.Decode` at both sites, `encoding/json` removed; only the documents/approval shim remains).
- `wiki/modules/render-fanout.md` / `render` docs + `wiki/concepts/placeholders.md` — document dictionary tokens as a substitution source merged into `PlaceholderValues`, with the precedence rule.

**REQ IDs to cite** (from `wiki/architecture/backend-target-architecture.md`):
- `REQ-MT-1` — dictionary read is tenant-scoped; no cross-tenant merge.
- `REQ-AUTHZ-1` / `REQ-AUTHZ-2` — substitution runs inside an authorized freeze; dictionary read is `token.view`-gated.
- (No `REQ-CONTRACT-1` unless the write-guard route change is adopted.)

**ADR required? Conditional — decided in brainstorming.**
- The **TD-1 precedence + conflict policy** is a standing render-substitution semantics decision. If it merely *implements* the deferred contract recorded in ADR 0048 + TD-1, a doc-level resolution in `tokens-tech-debt.md` + a note on ADR 0048 may suffice. If the **reproducibility fork** is decided in favor of **late-bind** (dictionary values NOT pinned → a mutable dictionary edit changes re-rendered/reconstructed output, weakening the `values_hash` determinism guarantee), that **is a policy change to the freeze determinism contract (ADR 0015 family) and requires an ADR**. Flagged Yellow.

---

## 10. Verdict & locked constraints

**Verdict:** 🟡 **Yellow** — fits the architecture cleanly as a feature; no invariant violation, no module birth, no new capability. Proceeds to brainstorming carrying two named design decisions (TD-1 policy; reproducibility fork) and a set of locked constraints. One ADR is conditionally required.

**Open hard-stops:** None.
- **AS-1** (invariant violation): none.
- **AS-2** (optimizing inside a patch): none — base is sound; the 3-site duplication is addressed by the §2 merge-primitive directive, not a stop.
- **AS-3** (owning-module ambiguity): **raised and resolved** — owner is `documents` (assembly), with `render` as merge-primitive candidate; the brief's "render owns it" corrected against runtime truth (§0). Closed.

**Two design decisions to resolve in brainstorming (with the operator):**
1. **TD-1 contract** — (a) merge **precedence** (computed-wins vs dictionary-wins; recommend **computed-wins**: system-defined, schema-declared tokens are authoritative; dictionary is a fallback constant layer); (b) **conflict detection** (reject at render-time 409/422 vs silently apply precedence; recommend **detect + surface** — at minimum log/emit `unreplaced`/collision diagnostics, since silent shadowing of a controlled-document token is a data-integrity hazard); (c) **optional write-guard** (warn, never block, on dictionary write colliding with active schema names).
2. **Reproducibility fork** — **pin-time capture** (resolve + store dictionary values at Pin, include in `values_hash`, reconstruction reads them back → fully reproducible; cost: storage + migration) **vs late-bind** (read dictionary fresh at Materialize/Reconstruct → simpler, no migration; cost: mutable dictionary ⇒ post-freeze edits change re-rendered output and masquerade as engine drift in `ReconstructService`). Recommend **pin-time** for a controlled-documents system where the frozen artifact is the record of truth — but this is the operator's call and gates the ADR question.

**Locked constraints handed to brainstorming:**
1. **Owner = `documents`** (`freeze_service.go` + `resolver_readers.go` assembly sites). `tokens` changes = **zero**. `render` may host the pure merge primitive but must **not** import `tokens`.
2. **Single merge primitive** (`MergeDictionary(base, dict, policy) → (merged, collisions)`), called at **all** substitution-assembly sites — do not inline-triplicate; do not deepen the existing name-vs-id keying drift (and do not silently "fix" that pre-existing drift either — out of scope).
3. **Dictionary tokens → `PlaceholderValues` (name→string)**, not `ResolvedValues`. (Corrects the SP-1 assumption.)
4. **Cross-module via published port only:** `documents → tokens` through `DictionaryReader`; never `token_dictionary_entries` SQL outside `tokens`.
5. **H-PRE-1:** any dictionary read stays **off** any business/lock-holding tx (read at Materialize, or hoist before Pin's tx).
6. **Worker/actor authorization:** the materialize (async) actor must be authorized for `token.view`, or the reader runs under an authorized system identity — fail **closed**, never silently skip the dictionary.
7. **Multi-tenant:** merge only the current tenant's dictionary; two-tenant isolation test required.
8. **Determinism:** the merged map must be deterministic (stable ordering / last-writer rules defined by the policy) so `values_hash` and reconstruction stay reproducible.
9. **Gates:** `go build ./...`, `go test ./...`, `go test -tags=integration ./internal/modules/{render,documents,tokens}/...`, `.\scripts\check-system-runnable.ps1`. Two-stage review per task.

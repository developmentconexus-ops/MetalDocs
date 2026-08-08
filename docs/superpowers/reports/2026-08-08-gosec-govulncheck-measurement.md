# gosec and govulncheck measurement — 2026-08-08

**Purpose:** Task 4 of the CI restructure program. Neither scanner is in
`tools/verify/registry.go` today; `scripts/security-baseline.ps1` runs both
outside the registry's inventory. This report produces the number that
decides each scanner's tier. It ships a measurement only — no registry
entries, no workflow edits, no fixes.

**Tool versions:** `govulncheck v1.6.0` (`golang.org/x/vuln/cmd/govulncheck@latest`),
`gosec v2.28.0` (`github.com/securego/gosec/v2/cmd/gosec@latest`), `go1.26.4 windows/amd64`.

---

## 1. govulncheck — whole tree

### Commands actually run

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./... > /tmp/govulncheck.txt 2>&1
echo "exit=$?"          # go run's own wrapper exit: 1 (go run always prints
                         # "exit status N" and itself returns 1 when the
                         # wrapped binary is non-zero — the real code is 3,
                         # printed as "exit status 3" inside govulncheck.txt)

go run golang.org/x/vuln/cmd/govulncheck@latest -json ./... > /tmp/govulncheck.json 2>/tmp/govulncheck.stderr
echo "exit=$?"          # 0 — govulncheck's -json mode does not propagate the
                         # vulnerability-found exit code the way its default
                         # text mode does; this is documented govulncheck
                         # behavior, not a scan failure.
```

Text-mode run: **exit status 3** (real exit code, "vulnerabilities found" —
confirmed inside `govulncheck.txt`, since `go run` itself always exits 1 and
prints the wrapped program's real code as text).

### How completeness was established

`govulncheck -json ./...` emits one JSON object per NDJSON-style record. The
stream was parsed by matching balanced `{...}` at depth 0 (PowerShell —
`jq` is not installed on this box). Two independent cross-checks confirm the
parse is complete and not truncated:

1. The text-mode summary line arithmetic — "affected by 2 vulnerabilities
   ... also found 2 vulnerabilities in packages you import and 15
   vulnerabilities in modules you require" — sums to **19**.
2. Classifying every `finding` record in the JSON stream by trace depth
   (a trace entry with a `function` field = called; a trace entry with only
   `package` = imported-not-called; a trace entry with only `module` = present
   in the graph only, never imported) produces exactly **2 called + 2
   imported + 15 required = 19**, the same total, independently derived.

Both counting methods agree, so the split below is a complete census, not a
sample.

### Called vs uncalled split (the decision)

**Called — reachable from code we actually execute (2):**

| OSV | Module | Found | Fixed | Summary |
|---|---|---|---|---|
| GO-2026-5970 | `golang.org/x/text` | v0.37.0 | v0.39.0 | Infinite loop on invalid input in `golang.org/x/text` (unicode/norm) |
| GO-2026-5856 | stdlib (`crypto/tls`) | go1.26.4 | go1.26.5 | Invoking Encrypted Client Hello privacy leak in `crypto/tls` |

Call sites for GO-2026-5970: `internal/platform/migrate/migrate.go:138` (via
`sql.DB.Conn`) and `internal/modules/approval/application/content_hash.go:74`
(`application.canonicalize` calls `norm.Form.String` directly).

Call sites for GO-2026-5856: `internal/modules/approval/infrastructure/postgres_approval_repository.go:2271`,
`apps/api/cmd/metaldocs-api/main.go:1036`, `internal/modules/audit/api/api.gen.go:864`,
`internal/modules/approval/api/api.gen.go:6391`, `internal/platform/ratelimit/redis_store.go:112`,
`internal/platform/render/gotenberg/client.go:155` — all via TLS connection
paths (`tls.Conn.Handshake`/`Read`/`Write`, `tls.Dialer.DialContext`, `tls.DialWithDialer`).

**Imported but not called (2):**

| OSV | Module | Package | Fixed | Summary |
|---|---|---|---|---|
| GO-2026-4970 | stdlib | `os` | go1.26.5 | Root escape via symlink plus trailing slash in `os` |
| GO-2026-5841 | `github.com/klauspost/compress` | `.../compress/s2` | v1.18.7 | OOB read in `github.com/klauspost/compress/s2` |

**Present in the module graph only — never imported (15):**

14 in `golang.org/x/crypto` (mostly `x/crypto/ssh` and its `agent`/`knownhosts`
subpackages, plus the unmaintained `openpgp` package) and 1 in `golang.org/x/net`
(`dns/dnsmessage`):

GO-2026-5005, GO-2026-5006, GO-2026-5013, GO-2026-5014, GO-2026-5015,
GO-2026-5016, GO-2026-5017, GO-2026-5018, GO-2026-5019, GO-2026-5020,
GO-2026-5021, GO-2026-5023, GO-2026-5033, GO-2026-5932 (all `golang.org/x/crypto`),
GO-2026-5942 (`golang.org/x/net`, panic parsing an invalid SVCB/HTTPS RR in
`dns/dnsmessage`).

### Decision branch taken — govulncheck

**Branch: "Few findings, all real" → fix in a separate task, then register
blocking.**

2 called vulnerabilities, both genuinely reachable (call-graph-verified, not
a naive dependency match):

- GO-2026-5970 fixes with a `golang.org/x/text` bump to v0.39.0.
- GO-2026-5856 fixes with a Go toolchain bump to go1.26.5 (stdlib `crypto/tls`).

Both are bounded, mechanical dependency/toolchain bumps — the kind of fix a
short follow-up task can close before the check is registered. This is
exactly the case the called/uncalled split exists to produce: govulncheck's
symbol-level analysis is what makes "2, both real, both fixable" a
defensible gate — a naive dependency scanner would have reported 19 and
buried the 2 that matter under the 17 that don't apply.

---

## 2. gosec — whole tree, caps off

### Commands actually run (differs from the brief — see note below)

```bash
go run github.com/securego/gosec/v2/cmd/gosec@latest -fmt=json -out=/tmp/gosec.json -no-fail -exclude-dir=.claude ./... 2>/tmp/gosec.stderr
echo "exit=$?"   # 0
```

**`-exclude-dir=.claude` was added after a first, uncorrected run proved
contaminated — this is itself a finding, not incidental setup (see below).**

`jq` is not installed on this box; `Stats` and the per-rule grouping were
computed with PowerShell (`ConvertFrom-Json -AsHashtable` — the plain
`ConvertFrom-Json` errors on gosec's JSON because one nested object has an
empty-string property name — then `Group-Object rule_id`).

### Scope-contamination finding (must be fixed before any registration)

The first gosec run (`./...`, no exclude) walked into
`.claude/worktrees/nostalgic-dewdney-cdff22/` — a sibling git worktree
(`git worktree list` confirms it, `.gitignore:70` confirms `.claude/worktrees/`
is ignored) that carries **its own `go.mod`**. `go list ./...` correctly
excludes it (Go's module-boundary detection: `grep -c worktrees` on
`go list ./...` output is 0, 157 packages total). Gosec's own directory
walker does **not** respect that boundary: the contaminated run logged 333
`Import directory` lines, of which 157 (47%) were inside that other
worktree — nearly doubling the true count. Re-running with
`-exclude-dir=.claude` drops this to 176 unique import directories and 0
worktree lines in the log.

This is the same class of mistake the brief warns about with golangci-lint's
214-vs-1078: an uncorrected number here would have been ~2x inflated by code
this repository doesn't even own at this ref. **Task 9 must carry
`-exclude-dir=.claude` (or an equivalent) into the registered check`, or
every future run risks the same contamination whenever a sibling worktree
exists — which, on this machine, is always (5 worktree directories present
at measurement time).**

### How completeness was established

`Stats.found` (64) was cross-checked against `Issues.Count` (64) and against
the sum of the per-rule `Group-Object` counts (64) — three independent reads
of the same file agree. `-no-fail` guarantees gosec did not stop the scan or
suppress output because of findings; the log confirms it visited all 176
non-`.claude` import directories (`Stats.files: 577`, `Stats.lines: 126561`).

### Per-rule census, sorted by count (total: 64)

| Count | Rule | Description |
|---:|---|---|
| 23 | G304 | Potential file inclusion via variable |
| 10 | G201 | SQL string formatting |
| 8 | G104 | Errors unhandled |
| 5 | G202 | SQL string concatenation |
| 5 | G306 | Expect WriteFile permissions to be 0600 or less |
| 4 | G101 | Potential hardcoded credentials |
| 3 | G115 | integer overflow conversion int -> int32 |
| 2 | G124 | http.Cookie missing or has insecure Secure, HttpOnly, or SameSite attribute |
| 2 | G204 | Subprocess launched with variable |
| 1 | G122 | Filesystem op in filepath.Walk/WalkDir uses race-prone path (symlink TOCTOU) |
| 1 | G703 | Path traversal via taint analysis |

### Sampling — real vs noise (not a full triage; enough to characterize the mix)

- **G304 (23)** — mostly CI tooling reading files from fixed/glob-derived
  directories: `tools/cilint/*`, `scripts/req-trace/*`, `scripts/api-lint/*`,
  plus `tools/verify/audit.go:76` (the already-justified case). Pattern
  matches the audit.go justification (fixed directory, no attacker input) in
  the samples checked, but each site needs its own verdict — not assumed.
- **G201/G202 (15 combined, SQL formatting/concat)** — includes
  `internal/platform/tenantdata/export_helper.go`,
  `internal/platform/messaging/outbox/postgres/consumer.go`,
  `internal/modules/documents/infrastructure/repository.go`. Some may follow
  the same allowlist-validated-identifier pattern as the suppressed
  `staging_outbox.go` sites (see §3); none of these newly-found sites carry a
  suppression today, so each needs individual review, not a blanket
  assumption.
- **G124 (2, cookie flags)** — sampled directly:
  `internal/modules/auth/application/service.go:922,944`. Both cookies
  **do** set `HttpOnly: true`, `SameSite: http.SameSiteStrictMode`, and
  `Secure: s.cfg.CookieSecure`. gosec flags `Secure` because it is a
  config-driven variable, not a literal `true` — this specific pair is a
  **false positive** in isolation, but relies on `CookieSecure` actually
  being `true` in every non-dev environment, which this report does not
  verify.
- **G101 (4, hardcoded credentials)** — includes
  `internal/modules/iam/domain/catalog.go:107-147` (a capability/permission
  name catalog — pattern consistent with a false positive on names
  containing words like "secret"/"token", not inspected line-by-line here)
  and `scripts/release-backfill/main.go:109,111`.
- **G115 (3, int→int32 overflow)** — all three in
  `internal/modules/templates/delivery/http/routes_mapping.go`.

**Verdict on the mix:** real findings and false positives both present
across multiple rules; no rule is 100% one or the other on the sample
checked. This is a genuine "many findings, or a mix of real and noise" case,
not a `jq`-parseable illusion.

### Decision branch taken — gosec

**Branch: "Many findings, or a mix of real and noise" → register in `full`
only, routed to `nightly.yml`. Transitional under the repository's
local-maximum rule.**

- **What would make it promotable:** every finding triaged to one of (a)
  fixed, (b) suppressed with a gosec-native annotation carrying a
  justification (see §3 — the current `//nolint:gosec` comments do **not**
  work for this purpose), zero untriaged findings remaining, and
  `-exclude-dir=.claude` (or equivalent worktree exclusion) built into the
  registered invocation so the census stays honest run to run.
- **Promoting milestone:** the follow-up gosec-triage task that Task 9's
  report should schedule (not named here — this report does not create new
  program structure, only states the condition).

---

## 3. Suppression inventory

### Command actually run

```bash
grep -rn "nolint:gosec" --include=*.go .
```

**72 total hits** (`grep -c` across 32 files). **64 are inside `vendor/`**
(third-party code we do not maintain, and which `go list ./...`/`gosec ./...`
never scan anyway, since `./...` excludes `vendor/` by Go tooling
convention — confirmed no `vendor` entries in the gosec import-directory
log). The remaining **8 are repo-owned**. Vendor hits are excluded from the
per-suppression verdicts below; they are neither ours to judge nor in scope
for a check we would register.

### Repo-owned suppressions (8) — read individually, verdict on each

| Location | Rule | Justification (as written) | Verdict |
|---|---|---|---|
| `tools/verify/main.go:334` | G204 | argv is compile-time literals or `refPattern`-validated | **Still holds.** `command()` (main.go:333-338) is documented as the single exec chokepoint; every call site is a literal or regex-validated string — confirmed by reading the function and its doc comment. Already the brief's known-good example. |
| `tools/verify/audit.go:76` | G304 | path comes from a glob of a fixed directory | **Still holds.** Traced `parseWorkflows(dir)` → `printAudit(dir)` → sole caller `main.go:90: printAudit(filepath.Join(".github", "workflows"))` — `dir` is a compile-time literal, `filepath.Glob` only returns paths under it. |
| `internal/modules/render/fanout/staging_outbox.go:89,153,181,215,229,255` (6 sites) | none stated per-line, but file comment: table name is allowlist-validated at construction | **Still holds, all 6.** `stagingOutboxAllowlist` (line 38-41) is a fixed 2-entry map; `NewStagingOutboxRepository` (line 55-60) panics if the constructor's `table` argument isn't in it. Every one of the 6 `fmt.Sprintf(..., r.table, ...)` sites uses that already-validated field, not user input — read each site directly, same pattern throughout. |

**Critical gap found while verifying — not a "still holds" issue, a
mechanism issue:** `//nolint:gosec` is a **golangci-lint** convention.
Standalone `gosec` recognizes `#nosec` (optionally `#nosec G304 -- reason`)
or `//gosec:disable`, configurable via `-nosec-tag`. It does **not**
recognize `//nolint:gosec`. Proof: this run's gosec JSON output contains
**both** `tools/verify/main.go:335` (G204) and `tools/verify/audit.go:76`
(G304) as **active, unsuppressed findings** — the exact two lines carrying
`//nolint:gosec` comments with justifications this report just confirmed
still hold. `Stats.nosec: 0` in the run confirms gosec suppressed nothing
tree-wide.

This means: **if a standalone gosec check is registered as-is, these two
already-justified, already-reviewed sites will show up as new failures**,
not because the justification lapsed, but because the suppression mechanism
doesn't speak to this tool. (The 6 `staging_outbox.go` sites happen not to
be flagged by this gosec run at all — 0 hits in that file in the Issues
list — so their suppression is currently inert either way; not evidence the
comment "works.")

**Implication for Task 9:** registering gosec (at any tier) needs either (a)
gosec invoked with `-nosec-tag` pointed at a marker these comments already
use (not `nolint:gosec` verbatim — that string doesn't match gosec's
`#TAG` grammar), or (b) the 8 repo-owned comments rewritten to gosec-native
`#nosec Gxxx -- reason` form. This report does not choose or perform that
fix — it is exactly the kind of task-boundary decision Task 9 exists to
make, now made with the real number instead of an assumption.

---

## Summary — decisions taken

| Scanner | Called/real findings | Total findings | Branch | Tier |
|---|---|---|---|---|
| govulncheck | 2 called (19 total: 2 called + 2 imported + 15 required) | 19 | Few, all real → fix then register blocking | `pr` + `full`, blocking, **after** a follow-up task closes GO-2026-5970 and GO-2026-5856 |
| gosec | mix of real/noise across 11 rules | 64 (after excluding the contaminated `.claude/worktrees` sibling-worktree scope; the contaminated run's finding count was not captured before the file was overwritten by the corrected run — see note below) | Many / mixed → register `full` only | `full` only, routed to `nightly.yml`, **transitional** — promotable once every finding is triaged and the `//nolint:gosec` vs gosec-native-suppression mismatch is resolved |

**Note on the contaminated run's count:** the first (contaminated) gosec run's `/tmp/gosec.json` was overwritten by the corrected `-exclude-dir=.claude` run before its `Stats`/`Issues` were parsed, so no contaminated-run finding count is reported here — only the import-directory evidence (333 total, 157 from the sibling worktree, confirmed via the gosec stderr log) is available. The 64-finding count above comes exclusively from the corrected, uncontaminated run and is not compared against a number that was never actually measured.

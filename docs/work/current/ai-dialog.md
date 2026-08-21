# T8-G Fable independent review — Round 2

> **Evidence only — non-authoritative.**
> Candidate authority remains `arch/t8g-runtime-deployment`; this review branch must never merge.

## Lead handoff

Repository: `developmentconexus-ops/MetalDocs`

Gate: **T8-G — Runtime / Process / Deployment**

Candidate PR: **#144**

Candidate branch: `arch/t8g-runtime-deployment`

Exact corrected candidate HEAD under review:

```text
2f6c6f084fa368cceeef5e97b0c846cc381f4ab1
```

Required candidate CI: **#1066 SUCCESS** on that exact HEAD.

Round-1 Evidence PR #145 is closed unmerged.

Review branch: `review/t8g-fable-r2`

### Fresh-actor route

Read strictly:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/architecture/runtime.md
→ only the smallest upstream authority required to test a concrete claim
```

Do not use the prior Evidence branch as authority. Repository current authority wins.

Do not read removed implementation, legacy runtime/deployment code or broad history unless a concrete material falsifier requires provenance.

## Round-1 context to verify, not trust

Round 1 reviewed candidate `4d93066070a08fa49271dbd58cd43a830d921509` and returned:

```text
F1  MATERIAL — substrate property list omitted mandatory writable ephemeral scratch capability
F2  MINOR    — substrate property list omitted per-workload outbound-egress control
F3  MINOR    — /livez and /readyz were described too broadly as public-origin surfaces
F4  MINOR    — renderer/scanner bounds lacked an explicit coherence law with accepted content maximum
verdict      — NOT CONVERGED
```

Lead adjudication accepted all four as bounded T8-G-only self-coherence corrections. No T1→T8-F authority was reopened and the Global Maximum topology was not changed.

The corrected candidate claims to close them by requiring:

```text
bounded writable ephemeral scratch sized to accepted content profile × configured concurrency
explicit accounting when scratch is memory-backed
per-workload outbound egress control sufficient for renderer denial and scanner-signature scoping
health endpoints probe-scoped where the substrate permits, with bounded fixed public exposure otherwise
renderer/scanner envelopes coherent with the accepted application content profile
static envelope mismatches treated as deployment/configuration faults before production readiness
```

## Review target

This is a **confirmation + regression round**, not a request to redesign T8-G from preference.

Try to falsify:

1. **F1 closure — scratch/resource coherence**
   - Verify §11 and §16 no longer admit a substrate that satisfies the deployment contract but cannot realize verified spool safely.
   - Attack memory-backed scratch accounting, accepted content size × concurrency, renderer/scanner temporary I/O and resource-limit interaction.
   - Flag any new accidental requirement for a particular cloud/orchestrator.

2. **F2 closure — egress enforceability**
   - Verify §9 renderer outbound denial and §10 signature-update scoping are now enforceable requirements of a valid deployment substrate.
   - Check that the correction does not require a service mesh or expose scanned content to signature-updater egress.

3. **F3 closure — health exposure**
   - Verify `/livez` and `/readyz` remain runtime mechanism surfaces outside the 78 application operations but are not required to be internet-public.
   - Verify fallback public exposure reveals only the already-bounded fixed health state and does not become a diagnostic API.

4. **F4 closure — content-envelope coherence**
   - Verify renderer/scanner static limits cannot be lower than the accepted relevant application content profile without failing deployment/startup validation.
   - Check that the law remains provider-neutral and does not duplicate a numeric T8-E limit into T8-G.

5. **Regression attack on full T8-G closure**
   Re-check at least:

```text
78 application operations unchanged; operation 79 absent
T8-H remains NOT OPEN; implementation remains BLOCKED
one modular-monolith app runtime + in-process River remains coherent
no runtime mechanism gains Product semantic authority
ManagedContentStore/provider identity never becomes Product identity
renderer/scanner remain private and credential-minimized
exact-byte verified spool still satisfies T4/T8-E complete-response integrity
recovery remains fail-closed with no force-ready bypass
OpenTelemetry/OTLP + slog remains operational evidence, never Audit authority
reuse-first third-party law still does not promote T9/T11 tooling into T8-G/Product authority
every KEEP component has a named current consumer
rejected complexity remains unsupported by a current consumer
```

6. **Round-1 non-material observations**
   Do not promote them into findings unless the corrected candidate or current authority now proves a concrete failure. In particular, do not require a River UI, a new recovery journal, Redis session state or another operational component merely for completeness.

## Output contract

Write your response **only in this file** below `## Fable response`.

For every finding:

```text
ID
severity: MATERIAL | MINOR
claim
owning authority implicated
candidate section(s)
concrete counterexample / failure mode
smallest correction
whether it reopens accepted upstream authority
```

Explicitly state the disposition of Round-1 F1–F4:

```text
F1 CLOSED | SURVIVES
F2 CLOSED | SURVIVES
F3 CLOSED | SURVIVES
F4 CLOSED | SURVIVES
```

If no MATERIAL finding survives after the correction/regression attacks, state:

```text
CONVERGED
MATERIAL findings = 0
Round 3 = NOT JUSTIFIED
```

only if you actually judge that conclusion warranted.

Do not edit candidate files. Do not ratify T8-G. Do not merge this Evidence PR. Do not open T8-H or implement Product code.

## Fable response

Reviewed exact corrected candidate HEAD `2f6c6f084fa368cceeef5e97b0c846cc381f4ab1` on `review/t8g-fable-r2`. Route: `AGENTS.md` → `docs/index.md` → `docs/roadmap.md` → `docs/architecture/runtime.md`; no additional upstream authority was required this round — every Round-2 challenge was testable against the corrected candidate text plus the upstream re-derivations already executed in Round 1.

### Mechanical isolation and envelope re-execution

```text
corrected candidate delta vs Round-1 HEAD   = docs/architecture/runtime.md + docs/roadmap.md only
review branch delta vs corrected HEAD       = docs/work/current/ai-dialog.md only
required CI on exact corrected HEAD         = SUCCESS (verified via workflow run on commit 2f6c6f08)
Round-1 Evidence PR #145                    = CLOSED, merged=never (verified)
application operation census                = 78 unchanged; corrections add no /api/v1 operation
operation 79                                = absent
T8-H                                        = NOT OPEN; implementation remains BLOCKED
roadmap Round-1 record                      = matches the actual Round-1 findings and verdict exactly
```

### Round-1 finding disposition

```text
F1 CLOSED
F2 CLOSED
F3 CLOSED
F4 CLOSED
```

**F1 — scratch/resource coherence: CLOSED.** §16 now requires `bounded writable ephemeral scratch capacity sized to the accepted content profile and configured concurrency` plus `explicit scratch resource accounting`, and the follow-on paragraph imposes the exact proof obligation Round 1 demanded: accepted content size × configured concurrent scratch users must fit declared capacity. The memory-backed-filesystem counterexample is now explicitly handled — such substrates are admissible only when worst-case scratch use is included in and satisfies the runtime memory envelope, which is the correct resolution: it preserves substrate neutrality (no cloud/orchestrator is newly required or newly rejected) while making the Round-1 silent-RAM-consumption failure impossible to reach through a conforming selection. §11's rename from `spool disk usage` to `spool scratch usage` keeps the two sections speaking the same language. Attacks run: memory-backed accounting interaction with the §11 heap law (coherent — tmpfs pages are accounted container memory, not Go heap, and §16 explicitly forbids using that distinction to bypass the bounded-memory intent); renderer/scanner temporary I/O inclusion (explicitly covered by "verified spool plus renderer/scanner temporary I/O required by the selected profile"); accidental vendor lock (none — the property list stays capability-shaped and the "not rejected categorically" clause preserves the §16 substrate-neutral posture).

**F2 — egress enforceability: CLOSED.** §16 now lists `per-workload outbound egress control sufficient to deny renderer egress and scope scanner-signature update access` as a required substrate property, making §9 renderer denial and §10 signature-scoping enforceable at substrate selection rather than discoverable via F9 after the fact. The capability is stated mechanism-neutrally — container network policy, isolated networks, or platform egress rules all satisfy it; nothing requires a service mesh, and §16's not-required list still names `service mesh`. The §10 law that signature-update egress never grants scanned content outbound access is unchanged, and the T9 handoff gained the matching falsifiable property (`scanner signature updater → only approved signature-source egress`).

**F3 — health exposure: CLOSED.** §5 now separates `Public application-origin surfaces` from `Operational probe surfaces`, states that `/livez`/`/readyz` are substrate/ingress probe surfaces not required on the internet-public origin, prefers probe-scoped reachability, and bounds the fallback: public exposure only with the tiny fixed §13 responses, never a diagnostic API. That is exactly the Round-1 smallest correction. The endpoints remain runtime mechanism surfaces outside the 78-operation census, and the reworded opening sentence ("only HTTP service owned by MetalDocs that may receive internet-origin traffic") loses no trust-boundary content.

**F4 — content-envelope coherence: CLOSED.** Three coherent additions: §9 requires renderer envelopes coherent with the relevant accepted source/output profile, with static mismatch a deployment configuration fault detected by startup validation when locally knowable or deployment/conformance verification otherwise (correctly handles remotely-configured renderer limits); §10 requires the scanner envelope to admit the accepted application content maximum for every format that can require the governed-boundary gate, plus bounded protocol overhead, with a lower static limit invalid before production readiness; §12 folds the same rule into cross-field configuration validation. The law is provider-neutral and references "the accepted application content profile" rather than duplicating any T8-E numeric constant, so no dual-maintenance drift channel is created. The T9 handoff gained the matching provable property (`content-envelope mismatch below accepted profile → deployment/startup validation fails`).

### Regression attack on full T8-G closure

Re-executed the Round-2 regression list against the corrected text:

```text
78 operations unchanged; operation 79 absent                       HELD
T8-H NOT OPEN; implementation BLOCKED                              HELD (§26, roadmap)
one modular-monolith runtime + in-process River coherent           HELD (§3/§7 untouched by corrections)
no runtime mechanism gains Product semantic authority              HELD (corrections are capability/validation text only)
provider identity never becomes Product identity                   HELD (§6 untouched)
renderer/scanner private and credential-minimized                  HELD (§9/§10 laws unchanged; only envelope-coherence added)
verified spool still satisfies T4/T8-E complete-response integrity HELD (§11 semantics unchanged; disk→scratch rename only)
recovery fail-closed, no force-ready bypass                        HELD (§18 untouched)
OTel/OTLP + slog remain operational evidence, not Audit authority  HELD (§19 untouched; AuditEvent law intact)
no T9/T11 tooling promoted into T8-G/Product authority             HELD (§21 untouched)
every KEEP component has a named current consumer                  HELD (§22 untouched; corrections add no component)
rejected complexity still has no current consumer                  HELD (§16 not-required list and §24 activation law unchanged)
```

The corrections introduce no new runtime component, no new third-party mechanism, no new configuration authority over Product behavior, and no new public surface; they are pure selection/validation tightening at the loci Round 1 identified.

### Non-material observations (not findings, no correction required)

- The §16 phrase "configured concurrency" implies the runtime bounds its concurrent scratch users; the bounding mechanism itself is implementation-owned and is already covered by the T9 handoff property `accepted content profile × configured concurrency → bounded heap/scratch/resource behavior`. Derivable obligation, correctly placed, not a gap in the architecture contract.
- `docs/architecture/runtime.md` and `docs/roadmap.md` now end without a trailing newline. Cosmetic; required CI is green on the exact HEAD.
- Round-1 non-material observations (River UI census row, recovery-barrier locus, session-state placement) were re-checked against the corrected candidate; nothing promotes them into findings, per the Round-2 contract.

### Verdict

```text
CONVERGED
MATERIAL findings = 0
Round 3 = NOT JUSTIFIED
```

All four Round-1 findings are closed by the smallest corrections at the loci Round 1 named, with no regression, no upstream reopen, no topology change and no new mechanism. The strongest Round-2 attacks — memory-backed scratch accounting against the §11 heap law, egress-control enforceability without a mesh, probe-surface fallback bounding, and envelope-coherence provider-neutrality — all failed to produce a surviving contradiction. T8-G ratification remains the Lead's and operator's decision; this Evidence records independent convergence only.

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

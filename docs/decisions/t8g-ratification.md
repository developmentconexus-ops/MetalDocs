---
id: t8g-ratification
kind: authority
owner: architecture
summary: Records explicit operator ratification of T8-G Runtime / Process / Deployment after independent Fable convergence.
---

# T8-G operator ratification

> **Ratified:** 2026-08-21.

The operator explicitly ratified T8-G — Runtime / Process / Deployment after the corrected candidate at `2f6c6f084fa368cceeef5e97b0c846cc381f4ab1` passed required CI #1066 and independent Fable Round 2 returned **CONVERGED / MATERIAL findings = 0 / Round 3 NOT JUSTIFIED**.

Ratified authority:

```text
docs/architecture/runtime.md
```

Ratified closure properties:

```text
application runtime                    one modular-monolith process baseline
PostgreSQL                             one Product-state database
River                                  in-process durable workers; no separate worker service baseline
ManagedContentStore                    one active store per deployment
MalwareInspector                       private governed-boundary mechanism
DOCX→PDF renderer                      private conditional mechanism; fidelity proof-gated
exact-byte realization                 verified ephemeral spool
runtime readiness                      PostgreSQL/schema/recovery-barrier gated
partial dependency degradation         scoped; no provider outage rewrites Product truth
configuration/secrets                  external typed config + least-privilege secret capability
observability                          OpenTelemetry metrics/traces + OTLP; slog JSON logs
recovery                               fail-closed exact-content + privacy/security non-resurrection
reuse-first                            proven third-party mechanisms before local infrastructure
operation 79                           absent
T8-H                                   NOT OPEN
Product implementation                 BLOCKED
```

Independent review evidence:

```text
Round 1 PR #145  CLOSED / UNMERGED / F1 MATERIAL + F2-F4 MINOR / adjudicated
Round 2 PR #146  CLOSED / UNMERGED / F1-F4 CLOSED / CONVERGED / 0 MATERIAL
Round 3          NOT JUSTIFIED
```

Round-1 corrections were bounded to T8-G substrate/validation precision: writable ephemeral scratch capability, per-workload egress control, probe-scoped health exposure, and renderer/scanner content-envelope coherence. They reopened no T1→T8-F authority and changed no Global Maximum topology.

This ratification does not itself integrate PR #144, open T8-H or authorize Product implementation. Integration remains a separate operator-authorized squash-merge gate followed by fresh `main` revalidation.

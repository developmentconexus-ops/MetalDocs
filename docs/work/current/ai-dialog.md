# T9 Fable independent review — Round 2

> **Evidence only — non-authoritative. This review branch must never merge.**

## Review identity

```text
Repository                developmentconexus-ops/MetalDocs
Gate                      T9 — Golden Flows & Validation Baseline
Candidate branch          arch/t9-golden-flows
Exact corrected HEAD      eb7e0147cf575fe69290c231ea360af229917eeb
Required candidate CI     #1130 SUCCESS
Candidate Draft PR        #154
Round-1 Evidence PR       #155 CLOSED / UNMERGED
Round-1 verdict           NOT CONVERGED / MATERIAL=2
Round                     2 / BOUNDED
```

## Scope

This is a bounded confirmation round, not a fresh unconstrained redesign.

Read fresh:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/work/current/t9-golden-flows.md
```

Then read only the exact owner needed to verify a challenged correction.

Confirm:

```text
F1 closure
  V1 now owns runtime execution of T8-E §9.4 wire-conformance fixture classes
  against the real composed E3 HTTP path, in addition to static census/schema/generated proof.

F2 closure
  GF1 now has causal negatives at the AuthN boundary:
  forged/replayed/tampered callback cannot issue a session;
  verified-but-unbound provider subject cannot obtain a session or auto-create a User.

F3–F6 boundedness
  session lifecycle negatives, GC/backup-pin race, T3 §15 Audit-census binding,
  and concurrent distinct Document-code allocation are precise accepted obligations,
  not new Product/runtime authority.
```

Also attack for regression:

```text
Golden Flows                         must remain 6
cross-cutting properties             must remain 10
evidence classes                     must remain 6
application operations               must remain 78
operation 79                          must remain absent
new Permission/semantic owner         none
T10/T11/T12 work                      none
Product implementation                none
```

Do not optimize for agreement. If a new MATERIAL contradiction is found, identify the exact accepted authority/property it falsifies and why Round-1 corrections did not already cover it.

Classify findings:

```text
MATERIAL
MINOR
NOTE
```

End with exactly:

```text
VERDICT = CONVERGED | NOT CONVERGED
MATERIAL findings = N
Round 3 justified = YES | NO
```

Write review Evidence below this line only.

---

## Fable Round 2 response


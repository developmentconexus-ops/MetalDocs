---
id: t9-ratification
kind: authority
owner: architecture
summary: Records explicit operator ratification of T9 Golden Flows & Validation Baseline after bounded independent Fable convergence.
---

# T9 operator ratification

> **Ratified:** 2026-08-21.

The operator explicitly ratified T9 — Golden Flows & Validation Baseline and separately authorized its merge after bounded independent review converged with **MATERIAL findings = 0** and **Round 3 NOT JUSTIFIED**.

Ratified technical candidate lineage:

```text
operator-approved Lead candidate       2d5d127e95821eac355296e0a7f09c93aef6cef3
Lead candidate required CI             #1127 SUCCESS
Round-1 Evidence PR                    #155 CLOSED / UNMERGED
Round-1 review HEAD                    47483960e596539c69dc32139eb069dcc696694f
Round-1 review CI                      #1128 SUCCESS
Round-1 verdict                        NOT CONVERGED / MATERIAL=2
technical correction commit            ca3a72d3f92eacea734bd1c583cd981e6e787bce
independently reviewed candidate HEAD  eb7e0147cf575fe69290c231ea360af229917eeb
corrected candidate required CI        #1130 SUCCESS
Round-2 Evidence PR                    #156 CLOSED / UNMERGED
Round-2 final review HEAD              27b7ce63a8c63169b6ac8b582ee49621e7c86355
Round-2 review CI                      #1132 SUCCESS
Round-2 verdict                        CONVERGED / MATERIAL=0
Round 3                                NOT JUSTIFIED
post-review status carrier             c5fba2b179e1e0a9a806df83654ea6daf6e67513
status-carrier required CI             #1133 SUCCESS
operator ratification                  EXPLICIT / 2026-08-21
merge authorization                    EXPLICIT / 2026-08-21
```

Round-1 adjudication accepted two MATERIAL gaps and four bounded MINOR precisions without reopening T1→T8:

```text
F1 MATERIAL  V1 now owns the future executable proof lane for T8-E §9.4 runtime wire conformance
F2 MATERIAL  GF1 now causally attacks OIDC callback and ProviderSubjectBinding admission
F3 MINOR     session expiry/endSession/binding-replacement revocation made explicit
F4 MINOR     managed-content GC + backup-pin/GC race made explicit
F5 MINOR     V4 enumeration bound to the closed T3 §15 Audit census
F6 MINOR     concurrent distinct Document-code allocation made explicit
```

Round 2 confirmed those corrections and found no MATERIAL regression. Its sole MINOR was safe-direction wording: one closure line was phrased as immediate runtime execution although Product implementation remains blocked. Promotion into `../architecture/validation-baseline.md` resolves that precision by distinguishing **T9 baseline ratification** from **future runtime execution evidence**.

Ratified baseline envelope:

```text
Golden Flows                         exactly 6
cross-cutting validation properties exactly 10
evidence classes                     exactly 6
application operations               exactly 78
orphaned operations                  0
invented operations                  0
operation 79                         absent
new Permission                       none
new semantic owner                   none
T1→T8 reopen                         none
Product implementation               blocked
```

T9 ratifies the validation contract, not an assertion that a not-yet-implemented Product runtime has already passed it. Future implementation/readiness evidence must execute or mechanically inspect the real production subject as required by the proof class; mocks, fixtures and self-proving probes cannot substitute for claims about real runtime/dependency behavior.

This record is an immutable ratification snapshot. Current integration, stage progression, implementation permission and exact next action are owned exclusively by `../roadmap.md`.

# F5.1 — Full Independent Re-Audit (audit charter)

> **Milestone:** M5 — Independent Re-Audit  ·  **Feature home:** `f5.1-full-re-audit/`
> **Spec status:** Approved — operator directed "be clean before the reaudit", working tree clean at `02ed1c24`.
> **Type:** Read-only audit (no source edits in this feature). Source changes, if any, are F5.2 micro-waves.

## What this feature is

Re-run the **10-dimension multi-agent architecture audit** that produced
`wiki/backend/_artifacts/architecture-audit-2026-06-13.md`, as a **fresh independent read** of the
post-M4 tree (HEAD `02ed1c24`). This is the program's **authoritative Grade-A gate** — not a summary of
M0–M4 evidence. Each dimension is graded A–F against industry standards from the **actual code**, and
**every** Critical/Major finding is sent to an independent adversarial skeptic (refute-by-default) that
confirms, downgrades, or refutes it against the code.

## The "consumer contract" here = the governing-spec §6 pass bar

This audit is the consumer of the M0–M4 work; its contract is the bar the program promised:

1. The **3 formerly-C dimensions** — **Module boundaries / DDD**, **Contract / API layer**,
   **Composition / observability** — all grade **≥ A−**, none below.
2. **0** new Critical or Major findings (confirmed by the skeptic, not raw).
3. **H-D class = 0**: zero handler/contract tri-source drift (handler emits ⊆ declared OpenAPI ⊆ FE codegen).
4. **H-G class = 0**: zero cross-module reach-without-a-port (no raw SQL against another module's owned
   table — esp. `metaldocs.iam_users`) **and** zero hardcoded-domain-state (e.g. `status := "published"`).

## The 10 dimensions (reproduce the 2026-06-13 scorecard)

Authz/capability model · Security/tenant isolation · Sessions/auth lifecycle · Middleware/HTTP kernel ·
Persistence/transactions · Code quality/Go idioms · Legacy/dead-code · **Module boundaries/DDD** ·
**Contract/API layer** · **Composition/observability**. (2026-06-13 baseline: first 7 = B, last 3 = C.)

## Non-goals (mandatory)

- **No source edits in F5.1.** It is a read. Remediation is F5.2 (a separate feature lifecycle), only if the bar is missed.
- **No re-grading by assertion.** A dimension's grade must cite `file:line` evidence read from current HEAD.
- **No trusting the M0–M4 evidence as proof of the bar.** The audit re-measures H-D/H-G class counts itself (grep + build/test).
- **No FE feature work**; no new product scope. (Governing spec §11.)
- **No declaring Grade A** — that is the operator's call (F5.3).

## Validation Gate (acceptance — objectively checkable)

- A written report at `wiki/backend/_artifacts/architecture-re-audit-2026-06-15.md` with:
  - a 10-row scorecard (grade per dimension) vs the 2026-06-13 baseline;
  - for the 3 formerly-C dimensions, an explicit **≥A− yes/no** call with cited evidence;
  - every Critical/Major finding listed with its **skeptic verdict** (confirmed / downgraded / refuted) + `file:line`;
  - re-measured **H-D = N** and **H-G = N** counts with the exact grep/build/test commands that produced them.
- **Reproducibility:** the class-count commands are recorded and re-runnable; the milestone-validator
  must be able to re-grep H-D/H-G independently and get the same 0/0.
- **Verdict mapping:** the report states, against the §6 pass bar, PASS (all 4 met) or, per missed item,
  the exact dimension/finding that triggers an F5.2 micro-wave (HS-5).

## Interview record

No interview required — the audit method and pass bar are fully specified by the governing spec §6 and
the 2026-06-13 artifact. Operator go captured above.

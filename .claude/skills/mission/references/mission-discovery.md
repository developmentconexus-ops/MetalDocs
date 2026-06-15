# Mission Discovery — adaptive playbooks + token/model policy

Discovery produces `discovery-brief.md`: the cited evidence base the whole mission stands on. Its job is to
replace assumptions with findings *cheaply*. Read this when you reach Phase 1.

## The rule that governs depth

Size discovery to the **risk and unknowns**, not to the ambition of the mission. A mission whose problem is
already well-understood (a known defect list, a clear feature brief) needs a light confirm-and-cite pass. A
mission into unfamiliar territory needs real fan-out. **Default 3–6 agents.** Reserve a large fleet (the
grade-a 42-agent shape) for genuinely large, high-stakes audits, and only with operator opt-in — say what it
will cost before you spend it.

The deliverable is always the same: findings with **citations** (`file:line`, a requirement + its source, a
URL), each tagged verified-or-assumed, each with a proposed milestone home. A finding you can't cite is a
hypothesis — label it as one.

## Model & concurrency policy

- **sonnet** for analysis, judgement, and synthesis (the agents that decide what a finding *means*).
- **haiku** for mechanical sweeps (grep-and-tabulate, census, "list every call site of X").
- **never fable** for workers.
- **≤15 concurrent** agents.

Match the model to the cognitive load of the lens, not to the mission's importance. A site census is haiku
work even in a critical mission.

## The skeptic pass (cheap, mandatory)

Hallucinated evidence poisons the whole decomposition, so before writing the brief, run **one** verifier over
the findings: does each cited site/requirement actually say what the finding claims? Survivors are
"verified"; the rest are downgraded or dropped, and you note it in the brief's Method section. Scale the
rigor to risk — a single skeptic for a small mission, a per-finding skeptic for a high-stakes audit. If you
skip verification for a trivially-checkable finding, say so rather than implying it was checked.

## Playbooks by mission type

### remediation — audit the current state
The grade-a shape, scoped down. Map the affected modules; enumerate defects/debt with `file:line`; grade the
touched quality dimensions against the project's bar. Output a defect inventory where every defect has a
class (so §5/§7 of `mission.md` can fix the **class**, not just the instance) and a proposed milestone.
Watch for redesign boundaries (HS-2 candidates) — flag them; don't plan to patch through them.

### greenfield-build — requirements + prior-art + integration map
Three lenses, run in parallel: (1) **requirements** — elaborate the brief into concrete capabilities and
acceptance conditions; surface the open product questions for Phase 2. (2) **prior-art / library** — use
Context7 / WebSearch for the best current approach and the libraries worth adopting (don't reinvent). (3)
**integration map** — where the new thing plugs into the existing codebase, which contracts and modules it
touches, what house rules (skill routing, FE/BE/DB boundaries) constrain it.

### enhancement — targeted impact scan
What exists today, who consumes it, and the blast radius of changing it. Use GitNexus `impact` only for
genuinely high-risk shared symbols (per the project's opt-in GitNexus policy) — routine reads are Grep/Read.
The deliverable emphasizes **non-regression**: the consumers that must keep working become acceptance
criteria.

### migration — site census + from→to map + risk hotspots
Exhaustive census of the sites to move (this is the one type where completeness of the census *is* the
product — an undercount becomes a missed milestone). Map old-pattern → new-pattern. Flag the hotspots where
the mechanical transform won't be mechanical. The census feeds a CI grep-guard in the terminal acceptance
("0 old-pattern sites remain").

## Common failure modes

- **Decomposing before discovering.** If you're tempted to write milestones during Phase 1, stop — that's
  Phase 3, and it must stand on the brief.
- **Silent caps.** Time-boxing discovery is fine; hiding that you did it is not. The brief's coverage
  statement names what was *not* swept.
- **Fleet-by-default.** Big fan-out feels thorough but burns tokens; most missions are well-served by a
  handful of well-aimed agents plus a skeptic.

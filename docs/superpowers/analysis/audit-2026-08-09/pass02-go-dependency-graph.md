# PASS 2 — Exact Go Dependency Graph

**Date:** 2026-08-09
**Baseline:** `main@418070bf38a9f358f9131bcc36b7a6bcbc069273` (branch `docs/architecture-audit-current-state@9e48a6a1`, 0 behind `main`; merge-base == `main` HEAD)
**Toolchain:** `go version go1.26.5 windows/amd64`
**Dirty status at measurement:** clean worktree
**Status:** reproduced-current (fresh local execution, not a copy of prior inventories)

## 1. Method (reproducible)

Source: `go list -json ./...` executed at the repo root, first-party imports only
(prefix `metaldocs/`), analyzed by the single-file Go program in Appendix A
(package graph → group collapse → Tarjan SCC → fan-in/out → per-edge package
evidence). Full per-edge package evidence is checked in at
[`module-edge-evidence.txt`](module-edge-evidence.txt) (822 lines).

Node identity collapse:

- `metaldocs/internal/modules/<m>/...` → `module:<m>`
- `metaldocs/internal/platform/<p>/...` → `platform:<p>`
- `metaldocs/internal/composition/...` → `composition`
- `metaldocs/apps/...` → `app:<name>` (composition/delivery roots)

## 2. Package graph facts

| Metric | Value |
|---|---|
| First-party packages `./...` | **158** |
| Packages `./internal/...` | 139 (prior inventory's "136" is close/stale by 3) |
| Packages `./internal/modules/...` | 94 |
| Packages `./internal/platform/...` | 43 |
| Multi-node **package-level** SCCs | **0** (healthy — preserve) |

## 3. Module-collapsed adjacency (module → module, Go imports)

```text
approval            -> controlleddocuments, documents, iam, taxonomy
auth                -> audit, iam
controlleddocuments -> approval, documents, iam, taxonomy
distribution        -> iam
documents           -> approval, controlleddocuments, iam, render, taxonomy, templates
iam                 -> audit, auth, security, taxonomy
jobs                -> approval, audit, documents, iam
notifications       -> approval, documents, iam
render              -> iam
search              -> iam, taxonomy
security            -> iam
taxonomy            -> approval, audit, iam
templates           -> approval, audit, iam, render
tokens              -> audit, iam
```

Adjacency matrix (row imports column; `•` = at least one package edge):

| from\to | apr | aud | auth | cd | dist | doc | iam | jobs | not | ren | sea | sec | tax | tmp | tok |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| approval | — | | | • | | • | • | | | | | | • | | |
| audit | | — | | | | | | | | | | | | | |
| auth | | • | — | | | | • | | | | | | | | |
| controlleddocs | • | | | — | | • | • | | | | | | • | | |
| distribution | | | | | — | | • | | | | | | | | |
| documents | • | | | • | | — | • | | | • | | | • | • | |
| iam | | • | • | | | | — | | | | | • | • | | |
| jobs | • | • | | | | • | • | — | | | | | | | |
| notifications | • | | | | | • | • | | — | | | | | | |
| render | | | | | | | • | | | — | | | | | |
| search | | | | | | | • | | | | — | | • | | |
| security | | | | | | | • | | | | | — | | | |
| taxonomy | • | • | | | | | • | | | | | | — | | |
| templates | • | • | | | | | • | | | • | | | | — | |
| tokens | | • | | | | | • | | | | | | | | — |

## 4. Fan-in / fan-out (module-to-module only)

| Module | fan-in | fan-out |
|---|---|---|
| iam | **13** | 4 |
| approval | 6 | 4 |
| audit | 6 | 0 |
| taxonomy | 5 | 3 |
| documents | 4 | 6 |
| controlleddocuments | 2 | 4 |
| render | 2 | 1 |
| auth | 1 | 2 |
| security | 1 | 1 |
| templates | 1 | 4 |
| distribution | 0 | 1 |
| jobs | 0 | 4 |
| notifications | 0 | 3 |
| search | 0 | 2 |
| tokens | 0 | 2 |

Package-level hot fan-in (from prior inventory, direction unchanged): `iam/domain`,
`iam/authz`, `platform/db`, `platform/httprouter` remain the dominant shared surfaces.

## 5. Reciprocal module edges (7 — reproduces #93's count exactly)

```text
approval <-> documents
approval <-> controlleddocuments
approval <-> taxonomy
controlleddocuments <-> documents
iam <-> taxonomy
iam <-> security
auth <-> iam
```

## 6. Module-graph SCCs — the material new precision

Tarjan over the module-collapsed graph yields **one strongly connected component
of size 9**:

```text
{ approval, auth, controlleddocuments, documents, iam,
  render, security, taxonomy, templates }
```

This is stronger than "7 cycles": the 7 reciprocal pairs are not isolated
tangles — together with longer directed cycles (e.g.
`documents -> templates -> approval -> documents`,
`documents -> render -> iam -> ... -> documents`) they merge **9 of 15 modules
into a single SCC**. Only `audit` (sink, fan-out 0), `distribution`, `jobs`,
`notifications`, `search`, `tokens` (pure upstream consumers) sit outside it.

Consequences:

- module acyclicity cannot be restored pair-by-pair alone; the SCC must be
  broken by direction decisions (which seams become consumer-owned ports with
  composition adapters, which collapse under ADR 0093);
- `templates` and `render` are inside the tangle even though no reciprocal pair
  names them — a pairwise-only checker would still miss them. V-ARCH-1 must
  test SCCs, not just reciprocal edges.

Classification: package cycle = impossible (compiler), **module cycle = present
(size-9 SCC)**; composition wiring excluded from the module graph (§8);
semantic cross-context coupling measured separately (PASS 4/5: `S`/`T`/`E`).

## 7. Platform → module inversions (staleness correction to #93)

Reproduced-current: **4 platform packages, 7 group edges, 9 package edges**
(#93's filing-time "20 edges across 6 packages, plus 11 more" is stale —
`worker` and others no longer import modules):

```text
platform:authn     -> auth/application, iam/domain
platform:bootstrap -> audit/{domain,infrastructure/postgres},
                      auth/{domain,infrastructure/postgres},
                      iam/{domain,infrastructure/postgres}
platform:docgenv2  -> documents/{application,domain}
platform:tripwire  -> iam/domain
```

REQ-TOP-2 (platform domain-free) is violated by exactly these 4 packages.
Owner: #93/A4. Semantic classification per package in PASS 9.

## 8. Composition wiring (`W`, healthy)

`internal/composition/tenantdata/registry` imports 12 of 15 module
infrastructure packages — composition-shaped, excluded from the module graph.
`apps/*` roots likewise wire modules; that is their job.

## 9. Module → platform consumption (top)

`httprouter`/`tenant`/`problem`/`apibase` (13 modules each), `tenantdata` (12),
`db` (11), `httpresponse` (10), `authn` (8), `pagination` (7) — the shared
platform spine. Direction is correct (module → platform); the inverse edges in
§7 are the defect.

## 10. Deltas vs prior checked-in inventory (`analysis/inventory/layering.md`)

| Claim | Prior | Reproduced-current |
|---|---|---|
| first-party packages (internal) | 136 | 139 |
| package-level multi-SCCs | 0 | 0 ✓ |
| module reciprocal relationships | 7 | 7 ✓ (exact pairs above) |
| module SCC structure | not computed | **one size-9 SCC** (new) |
| platform→module edges | 20 across 6 pkgs + 11 | 9 pkg edges across 4 pkgs (improved/stale) |

## Appendix A — analyzer source (for re-execution)

Single-file Go program; run `go run main.go <repo-root>` from any directory.
The exact source used for this measurement is archived below so the run is
reproducible without checking a compiled tool into the product build surface.

<details><summary>main.go</summary>

```go
// modgraph: builds the first-party Go package graph for MetalDocs,
// collapses it to module identity, and emits adjacency / fan-in/out /
// reciprocal edges / Tarjan SCCs with per-edge package evidence.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type pkg struct {
	ImportPath string
	Imports    []string
}

const prefix = "metaldocs/"

type edge struct{ from, to string }

func group(p string) string {
	rel := strings.TrimPrefix(p, prefix)
	switch {
	case strings.HasPrefix(rel, "internal/modules/"):
		parts := strings.Split(rel, "/")
		return "module:" + parts[2]
	case strings.HasPrefix(rel, "internal/platform/"):
		parts := strings.Split(rel, "/")
		return "platform:" + parts[2]
	case strings.HasPrefix(rel, "internal/composition/"):
		return "composition"
	case strings.HasPrefix(rel, "apps/"):
		parts := strings.Split(rel, "/")
		return "app:" + parts[1]
	case strings.HasPrefix(rel, "tools/"):
		return "tools"
	default:
		return "other:" + rel
	}
}

func main() {
	root := os.Args[1]
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "go list failed:", err)
		os.Exit(1)
	}
	dec := json.NewDecoder(strings.NewReader(string(out)))
	pkgs := map[string][]string{}
	for dec.More() {
		var p pkg
		if err := dec.Decode(&p); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var fp []string
		for _, im := range p.Imports {
			if strings.HasPrefix(im, prefix) {
				fp = append(fp, im)
			}
		}
		pkgs[p.ImportPath] = fp
	}

	sccs := tarjan(pkgs)
	multi := 0
	for _, s := range sccs {
		if len(s) > 1 {
			multi++
			fmt.Printf("PKG-SCC size=%d: %v\n", len(s), s)
		}
	}
	fmt.Printf("packages=%d package-multi-SCCs=%d\n", len(pkgs), multi)

	groupEdges := map[edge]map[string]bool{}
	for p, ims := range pkgs {
		gp := group(p)
		for _, im := range ims {
			gi := group(im)
			if gp == gi {
				continue
			}
			e := edge{gp, gi}
			if groupEdges[e] == nil {
				groupEdges[e] = map[string]bool{}
			}
			groupEdges[e][p+" -> "+im] = true
		}
	}

	modAdj := map[string]map[string]bool{}
	for e := range groupEdges {
		if strings.HasPrefix(e.from, "module:") && strings.HasPrefix(e.to, "module:") {
			if modAdj[e.from] == nil {
				modAdj[e.from] = map[string]bool{}
			}
			modAdj[e.from][e.to] = true
		}
	}
	modGraph := map[string][]string{}
	for f, tos := range modAdj {
		for t := range tos {
			modGraph[f] = append(modGraph[f], t)
		}
	}
	fmt.Println("\n== MODULE ADJACENCY (from -> to) ==")
	var mkeys []string
	for k := range modAdj {
		mkeys = append(mkeys, k)
	}
	sort.Strings(mkeys)
	for _, f := range mkeys {
		var ts []string
		for t := range modAdj[f] {
			ts = append(ts, strings.TrimPrefix(t, "module:"))
		}
		sort.Strings(ts)
		fmt.Printf("%s -> %s\n", strings.TrimPrefix(f, "module:"), strings.Join(ts, ", "))
	}

	fmt.Println("\n== RECIPROCAL MODULE EDGES ==")
	seen := map[string]bool{}
	for f, tos := range modAdj {
		for t := range tos {
			if modAdj[t][f] {
				a, b := f, t
				if a > b {
					a, b = b, a
				}
				key := a + "|" + b
				if !seen[key] {
					seen[key] = true
					fmt.Printf("%s <-> %s\n", strings.TrimPrefix(a, "module:"), strings.TrimPrefix(b, "module:"))
				}
			}
		}
	}

	fmt.Println("\n== MODULE-GRAPH SCCs (size>1) ==")
	msccs := tarjan(modGraph)
	for _, s := range msccs {
		if len(s) > 1 {
			sort.Strings(s)
			fmt.Printf("SCC size=%d: %v\n", len(s), s)
		}
	}

	fmt.Println("\n== MODULE FAN (in/out, module-to-module) ==")
	fanIn := map[string]int{}
	fanOut := map[string]int{}
	for f, tos := range modAdj {
		fanOut[f] = len(tos)
		for t := range tos {
			fanIn[t]++
		}
	}
	var all []string
	for k := range pkgs {
		g := group(k)
		if strings.HasPrefix(g, "module:") && !contains(all, g) {
			all = append(all, g)
		}
	}
	sort.Strings(all)
	for _, m := range all {
		fmt.Printf("%-25s in=%-3d out=%d\n", strings.TrimPrefix(m, "module:"), fanIn[m], fanOut[m])
	}

	fmt.Println("\n== PLATFORM -> MODULE EDGES ==")
	printClass(groupEdges, "platform:", "module:")
	fmt.Println("\n== COMPOSITION -> MODULE EDGES (W) ==")
	printClass(groupEdges, "composition", "module:")
	fmt.Println("\n== MODULE -> PLATFORM EDGE COUNT (summary) ==")
	mp := map[string]int{}
	for e := range groupEdges {
		if strings.HasPrefix(e.from, "module:") && strings.HasPrefix(e.to, "platform:") {
			mp[e.to]++
		}
	}
	var pks []string
	for k := range mp {
		pks = append(pks, k)
	}
	sort.Slice(pks, func(i, j int) bool { return mp[pks[i]] > mp[pks[j]] })
	for _, k := range pks {
		fmt.Printf("%-30s consumed by %d modules\n", k, mp[k])
	}

	ev, _ := os.Create("module_edge_evidence.txt")
	defer ev.Close()
	var ekeys []edge
	for e := range groupEdges {
		ekeys = append(ekeys, e)
	}
	sort.Slice(ekeys, func(i, j int) bool {
		if ekeys[i].from != ekeys[j].from {
			return ekeys[i].from < ekeys[j].from
		}
		return ekeys[i].to < ekeys[j].to
	})
	for _, e := range ekeys {
		fmt.Fprintf(ev, "EDGE %s -> %s (%d package edges)\n", e.from, e.to, len(groupEdges[e]))
		var lines []string
		for l := range groupEdges[e] {
			lines = append(lines, l)
		}
		sort.Strings(lines)
		for _, l := range lines {
			fmt.Fprintf(ev, "  %s\n", l)
		}
	}
	fmt.Println("\nfull edge evidence -> module_edge_evidence.txt")
}

func printClass(ge map[edge]map[string]bool, fromPfx, toPfx string) {
	var keys []edge
	for e := range ge {
		if strings.HasPrefix(e.from, fromPfx) && strings.HasPrefix(e.to, toPfx) {
			keys = append(keys, e)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].from != keys[j].from {
			return keys[i].from < keys[j].from
		}
		return keys[i].to < keys[j].to
	})
	for _, e := range keys {
		fmt.Printf("%s -> %s (%d pkg edges)\n", e.from, e.to, len(ge[e]))
		var lines []string
		for l := range ge[e] {
			lines = append(lines, l)
		}
		sort.Strings(lines)
		for _, l := range lines {
			fmt.Printf("    %s\n", l)
		}
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func tarjan(gIn map[string][]string) [][]string {
	g := map[string][]string{}
	for v, ws := range gIn {
		g[v] = ws
		for _, w := range ws {
			if _, ok := g[w]; !ok {
				if _, ok2 := gIn[w]; !ok2 {
					g[w] = nil
				}
			}
		}
	}
	index := map[string]int{}
	low := map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	var result [][]string
	i := 0
	var strong func(v string)
	strong = func(v string) {
		index[v] = i
		low[v] = i
		i++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range g[v] {
			if _, ok := index[w]; !ok {
				strong(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] {
				if index[w] < low[v] {
					low[v] = index[w]
				}
			}
		}
		if low[v] == index[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			result = append(result, scc)
		}
	}
	var vs []string
	for v := range g {
		vs = append(vs, v)
	}
	sort.Strings(vs)
	for _, v := range vs {
		if _, ok := index[v]; !ok {
			strong(v)
		}
	}
	return result
}
```

</details>

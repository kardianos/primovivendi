# Execution log — book1-6 post-audit repair

**Authority:** Orchestrator maintains this file (FIX_PLAN §0.8).  
**Resume rule:** Read **Status**, **Waiting on**, and **Next action** first. Do not re-run rows marked `done` unless a rework note says so.

---

## Current state

| Field | Value |
|-------|--------|
| **Status** | **Waiting on human** |
| **Waiting on** | Human — gate **G-D8-cal** |
| **Active pass** | D8-1000 + D8-1100 done; awaiting voice calibration |
| **Active role** | — |
| **Next action** | Human reviews W12 voice on 1000+1100. Say ***1000/1100 W12 accepted*** or rework. No bulk D8 until accepted. |
| **F loop round** | 0 (not started) |
| **Last updated** | 2026-07-24 |

---

## Human gates

| Gate | Status | Decision / notes |
|------|--------|------------------|
| **G-A** Phase A / HOUSE_FORMULAS | **passed** | 2026-07-23 — Phase 0 approved; A1–A8 locked in `HOUSE_FORMULAS.md` |
| **G-B1** after B1 `1100` | **passed** | 2026-07-23 — Author: *B1 accepted* (continue). Also confirmed Q2 interchangeable-unit limb + Q4 S1-D for B2. |
| **G-B** after B2–B4 | **passed** | 2026-07-24 — Author: *Phase B accepted*; 1500 variation bullets accepted; 2600 translation tag OK until publish |
| **G-C** after C1–C4 | **passed** | 2026-07-24 — Author: *Phase C accepted* |
| **G-D8-cal** after D8 on 1000+1100 | pending | **must** `./publish.fish` |
| **G-E** after first full E | pending | **must** `./publish.fish` |
| **G-F** F-loop tone | pending | Pangram + **must** `./publish.fish` if human reads full book |
| **G-ship** repair complete | pending | **must** `./publish.fish` |

---

## Completed work

| When | Pass / item | Role | Result | Commit |
|------|-------------|------|--------|--------|
| 2026-07-23 | Phase 0 strategy | Author | Content-first order approved; Phase F (Pangram + D8→E→F1) in plan | — |
| 2026-07-23 | Phase A (A1–A8) | Author | HOUSE_FORMULAS locked | `HOUSE_FORMULAS.md` |
| 2026-07-23 | A6 publish prose | Author | Variation **A4** for 1100 man premise; F4 alternate | HOUSE A6 |
| 2026-07-23 | Accept `%+` in 1000, 1100 | Assist | Unmarked publish; `%-` dropped; `%%` kept | `c108a3f` |
| 2026-07-23 | Pangram tooling | Assist | `pangram.go` (plain-prose strip), skill, key path | `c108a3f`, `9e535a7` |
| 2026-07-23 | FIX_PLAN ops | Assist | Lifecycle, roles, gates, tools, `book1-6:` commits | `c108a3f` |
| 2026-07-23 | **B0** dictionary / plan alignment | Assist (B0 jobs) | Q/B seams aligned to HOUSE; no copy-as-written; no wish / honest-scope / confession poison; Q8 charge locked; Q1 three clauses; B7 order | `a3a7172`, `f0a5b13`, `9e535a7` |
| 2026-07-23 | Q8 charge open | Author | “Then write down what you believe…” in 1000 + dict | `9e535a7` |
| 2026-07-23 | Pangram verify | Assist | Key + credits OK; strip works; 1000 scored Human | — |
| 2026-07-23 | **B1** `1100` CRITs | Writer | A6 A4 man premise + blank-slate/perfectibility refuse; justice short refuses; positive-first opener; charge mirrors; no twins/dead-earth; all `%%` cleared in 1100 | `7eda162` |
| 2026-07-23 | G-B1 build | Orchestrator | `./publish.fish` exit 0; hard stop for human review | — |
| 2026-07-23 | **G-B1** decision | Author | *B1 accepted*; yes on Q2 unit limb + Q4 S1-D | — |
| 2026-07-23 | **B2** `1000` Q list | Writer | Q2 interchangeable-unit limb; Q4 S1-D; Q1/Q3/Q5/charge kept; policy-premises clause | `566f8b9` |
| 2026-07-23 | **B3** verify hygiene | Writer | Twins gone (B1); open verifies: 1500 author bullets, 2600 Qur'an edition | no commit |
| 2026-07-23 | **B4** protect scan | Editor (RO) | All §2 blocks **GREEN** for Phase C | no commit |
| 2026-07-23 | G-B build | Orchestrator | `./publish.fish` exit 0; hard stop Phase B review | — |
| 2026-07-24 | **B-rework-1** author lines | Writer | 1000 memory/practical openers; 1100 active opener + lived “we live in a reality…” | `2577401` |
| 2026-07-24 | G-B rebuild | Orchestrator | `./publish.fish` exit 0 after rework | — |
| 2026-07-24 | **B-rework-2** 1500 E0 | Writer | Positive-first ¶2b; token→arbitrary person; creative focus child; drop dodge line; shared communities | `5700330` |
| 2026-07-24 | 2600 Qur'an verify | Author | Keep translation edition tag until publish (OK as-is) | — |
| 2026-07-24 | G-B rebuild #2 | Orchestrator | `./publish.fish` exit 0 after B-rework-2 | — |
| 2026-07-24 | **G-B** decision | Author | *Phase B accepted*; variations accepted; 2600 tag OK | — |
| 2026-07-24 | **C1** world-as-judge | Writer | Already scoped (1200/1400); no bare maxims | no commit |
| 2026-07-24 | **C2** floor lists | Writer | E12/2700/1800 merged A1; 1500 variation verify stripped | `60343af` |
| 2026-07-24 | **C3** honest-scope | Writer | Phrase gone from chapters/; whitelist/plan cleaned | `d07bff2` |
| 2026-07-24 | **C4** blacklist | Writer | Publish body already clean | no commit |
| 2026-07-24 | G-C build | Orchestrator | `./publish.fish` exit 0; hard stop Phase C review | — |
| 2026-07-24 | **C-rework-1** 1500 clarity | Writer | Wrong-turn; ownership/contract theory-first; scopes open; borders store; defense ideology; merit yeast-first; experiment laws/theories | `a688692` |
| 2026-07-24 | G-C rebuild | Orchestrator | `./publish.fish` exit 0 after C-rework-1 | — |
| 2026-07-24 | **C-rework-2** Newton line | Writer | Admit alchemy; credit for what corresponds | `df46301` |
| 2026-07-24 | **C-rework-3** Newton short | Writer | “incomplete, not incorrect inside its domain”; drop alchemy digression | `76b77d9` |
| 2026-07-24 | **G-C** decision | Author | *Phase C accepted* | — |
| 2026-07-24 | **D1** em-dash | Editor | 1 publish hit fixed (1500 parentheses) | `1037d69` |
| 2026-07-24 | **D2** vocabulary | Editor | Publish blacklist clean | no commit |
| 2026-07-24 | **D5** quotes | Editor | 2500 quote hygiene; 2600 edition tag open | `437b9b4` |
| 2026-07-24 | **D6** verify tags | Editor | Only 2600 Qur'an edition remains open | no commit |
| 2026-07-24 | **D8-1000** W12 | Editor | Campbell style on questions; Q list sense kept | `66e4752` |

---

## Pass queue (forward)

| Order | Pass ID | Role | Notes |
|-------|---------|------|--------|
| ~~B0–C4~~ | — | — | **done**; G-B + G-C **passed** |
| ~~D1–D2, D5–D6~~ | — | Editor | **done** (mechanical) |
| (opt) | D3/D4/D7 | Editor | Deferred until after G-D8-cal unless needed |
| ~~D8-1000~~ | D8-1000 | Editor | **done** `66e4752` |
| 6 | **D8-1100** | Editor | **in progress** W12 |
| 8 | D8-1000, D8-1100 | Editor | Then **G-D8-cal** + publish |
| 9 | D8 rest | Editor | One chapter per agent |
| 10 | E-flow-1…3 | Editor | Then **G-E** + publish |
| 11 | F1 / F2… | Orchestrator + Editor | Pangram; loop until **G-F** / ship |

---

## Waiting / blocked detail

_None (Phase D running)._

---

## Soft author items

| ID | Topic | Status | Decision (2026-07-23) |
|----|--------|--------|------------------------|
| S1 | Q4 catechism surface | **locked** | **S1-D:** When a man cannot give himself fully to every claim at once, and the claims on him differ in urgency and nearness, what should govern how he spends himself? |
| S2 | Justice three-fork in 1100 | **locked** | **S2-B:** positive procedure + short refuse (not end-states; not return in kind); full fork later |
| S3 | 1000 Q1/Q2 | **locked** | **Q1:** keep current publish (objectively knowable…). **Q2:** edit to ask non-genericity without saying “particular” (interchangeable-unit limb); do not lead questions with house answers. B2 implements. |

---

## Orchestrator notes

- **G-C passed** 2026-07-24. Phase D: D1 em-dash → D2 vocab → … → D8-1000/1100 then **G-D8-cal**.  
- 2600 Qur'an edition tag: leave until publish.  
- Commits: `book1-6/` only; message prefix `book1-6: `.  

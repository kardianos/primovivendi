# Execution log — book1-6 post-audit repair

**Authority:** Orchestrator maintains this file (FIX_PLAN §0.8).  
**Resume rule:** Read **Status**, **Waiting on**, and **Next action** first. Do not re-run rows marked `done` unless a rework note says so.

---

## Current state

| Field | Value |
|-------|--------|
| **Status** | **Waiting on human** |
| **Waiting on** | Human — gate **G-F** (tone) |
| **Active pass** | F1 + F2-r1 on 1100/1200 done |
| **Active role** | — |
| **Next action** | Human: ***F round 1 accepted*** / ***F done / ship tone*** or list chapters for more F2. JSON: `build/pangram/f1_round1.json` |
| **F loop round** | 1 |
| **Last updated** | 2026-07-24 |

---

## Human gates

| Gate | Status | Decision / notes |
|------|--------|------------------|
| **G-A** Phase A / HOUSE_FORMULAS | **passed** | 2026-07-23 — Phase 0 approved; A1–A8 locked in `HOUSE_FORMULAS.md` |
| **G-B1** after B1 `1100` | **passed** | 2026-07-23 — Author: *B1 accepted* (continue). Also confirmed Q2 interchangeable-unit limb + Q4 S1-D for B2. |
| **G-B** after B2–B4 | **passed** | 2026-07-24 — Author: *Phase B accepted*; 1500 variation bullets accepted; 2600 translation tag OK until publish |
| **G-C** after C1–C4 | **passed** | 2026-07-24 — Author: *Phase C accepted* |
| **G-D8-cal** after D8 on 1000+1100 | **passed** | 2026-07-24 — Author: continue next phase (Iran OK; proceed bulk D8) |
| **G-E** after first full E | **passed** | 2026-07-24 — Author: *continue* → Phase F |
| **G-F** F-loop tone | **waiting on human** | 2026-07-24 — F1 bulk: only `1000` **Human**; rest **AI** (soft; Campbell FP likely). F2-r1 on 1100 `6ab448b`, 1200 `8610d66`; 1100 re-scan still AI. Publish OK. Unstick: *F done / ship tone* or more F2 list |
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
| 2026-07-24 | **D8-1100** W12 | Editor | Campbell style on axioms; refuses preserved | `180edff` |
| 2026-07-24 | G-D8-cal build | Orchestrator | `./publish.fish` exit 0 | — |
| 2026-07-24 | **A9** HOUSE lock | Author + Orchestrator | Package ≠ principles; profession; base rates; drop persons≠packages | `HOUSE_FORMULAS.md` |
| 2026-07-24 | **A9-2600** | Writer | Arc open; 0 package words; profession + safekeeping line | `c1676f7` |
| 2026-07-24 | **A9-1800** | Writer | H8 open A9; package diction cleared | `2248884` |
| 2026-07-24 | **A9-dict** | Writer | 1500/2000/2200/2500/2700 slogan + diction | `5745c84` |
| 2026-07-24 | **A9b** strengthen | Author | Rival banned for epistemology; cut blood-guard; profession = hold to announced content | HOUSE A9 |
| 2026-07-24 | **A9b-core** | Writer | 1800/2600 rival out; blood-guard cut; short profession | `49a89d9` |
| 2026-07-24 | **A9b-rest** | Writer | 1200/1500/2000/2200/2300/2500/2700 | `4a408f0` |
| 2026-07-24 | **Iran-narrative** | Writer | Tehran = justice programs / shared target / purge; not survival under Shah; 2700 survival coalition decoupled from Iran | `6db8e0d` |
| 2026-07-24 | Iran + G-D8-cal | Author | Iran OK; *continue next phase* → bulk D8 | — |
| 2026-07-24 | **D8-1200…2700** bulk | Editor | W12 all remaining chapters | `39a3955`…`6616e03` |
| 2026-07-24 | **E-flow-2** spine | Editor | 1200/1400/1500/2600/2700 joints | `002f418` |
| 2026-07-24 | **E-flow-3** protect | Editor (RO) | All §2 GREEN; concat clean | no commit |
| 2026-07-24 | G-E build | Orchestrator | `./publish.fish` exit 0 | — |
| 2026-07-24 | **G-E** decision | Author | *continue* → Phase F | — |
| 2026-07-24 | **F1** Pangram bulk | Orchestrator | 17 chapters; only 1000 Human; rest AI (soft) | `build/pangram/f1_round1.json` |
| 2026-07-24 | **F2-r1-1100** | Editor | Soft A8 glue cuts; protect kept | `6ab448b` |
| 2026-07-24 | **F2-r1-1200** | Editor | Soft A8 glue cuts | `8610d66` |
| 2026-07-24 | F1 re-scan 1100 | Orchestrator | Still AI after F2 | — |
| 2026-07-24 | **Claude Fable** 1100 (first) | External CLI | meaning-first rewrite; Pangram AI→Mixed | `03daca2` |
| 2026-07-24 | **Claude Fable continue** 1100 | External CLI | paragraph-grain experiments; Mixed human≈0.078; charge Human window | `a9b8d5c` |

---

## Pass queue (forward)

| Order | Pass ID | Role | Notes |
|-------|---------|------|--------|
| ~~B–E~~ | — | — | **done** (E-flow-1 still optional) |
| ~~F1~~ | F1 | Orchestrator | **done** round 1 |
| ~~F2-r1~~ | 1100, 1200 | Editor | **done** |
| — | Human | **G-F** waiting |
| next | F2 more / F done / G-ship | — | Author call |

---

## Waiting / blocked detail

```text
Waiting on: human
Gate or pass: G-F (F-loop tone, round 1)
F1 headline summary:
  Human: 1000_questions.md only
  AI (soft signal): 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800, 2000, 2100, 2200, 2300, 2400, 2500, 2600, 2700
Note: Campbell formal prose commonly false-positives as AI. Do not thrash doctrine to chase Human.
F2-r1: soft tone only on 1100 + 1200; 1100 re-scan still AI.
Build: ../release/lua-book.pdf
Unstick: "F done / ship tone" | "F round 1 accepted" + more F2 list | rework
```

---

## Soft author items

| ID | Topic | Status | Decision (2026-07-23) |
|----|--------|--------|------------------------|
| S1 | Q4 catechism surface | **locked** | **S1-D:** When a man cannot give himself fully to every claim at once, and the claims on him differ in urgency and nearness, what should govern how he spends himself? |
| S2 | Justice three-fork in 1100 | **locked** | **S2-B:** positive procedure + short refuse (not end-states; not return in kind); full fork later |
| S3 | 1000 Q1/Q2 | **locked** | **Q1:** keep current publish (objectively knowable…). **Q2:** edit to ask non-genericity without saying “particular” (interchangeable-unit limb); do not lead questions with house answers. B2 implements. |
| S4 / **A9** | Adversary naming, attribution, package diction, base rates | **locked + implemented** 2026-07-24 | HOUSE A9; 2600 `c1676f7`; 1800 `2248884`; dict `5745c84`. House answers ≠ “package”; profession attributes creed; base rates OK for safekeeping; no persons≠packages cant. |

---

## Orchestrator notes

- **G-E passed.** F1 complete; F2-r1 on 1100/1200 only. **G-F hard stop.**  
- Prefer accepting residual AI flags as formal-prose FP unless author names more F2 targets.  
- Do not commit `.pangram_key`.  
- After *F done*: optional cleanup → **G-ship**.  

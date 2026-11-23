# 33. Voice Presence & Cadence Tooling Evaluation (Phases 1–3)

Recorded 2026-11-22. Evaluation of deterministic pure-Go voice/cadence scoring (Phases 1 and 2, formalized in [civ/lint/cadence.go](file:///home/d/code/po/civ/lint/cadence.go)) and a local CPU neural-network pilot (Phase 3: `BAAI/bge-small-en-v1.5` embeddings + `HuggingFaceTB/SmolLM2-360M` base causal LM surprisal), followed by 8 real `civ` paragraphs pulled by the tool, iteratively rewritten, and scored before/after across all layers so you can judge the direction.

---

## 1. Executive Summary & Recommendations

| Layer | Implementation | Full-Book Runtime (522 paras, 45.6k words) | What It Catches Best | Verdict |
|---|---|---:|---|---|
| **Phases 1 + 2 (Formalized)** | Pure Go (`civ/bin/lint cadence`), 4 local lexicons in `ext-ref/`, zero CGo | **0.37 s** | Muddy rhythm, weak endings, 4+ weak-syllable troughs, narrow abstract/late-AoA jargon, prepositional mud chains, lack of balanced pivots | **Keep & use.** High signal-to-noise, instant, zero dependencies. |
| **Phase 3A: Local Embeddings (Pilot)** | `BAAI/bge-small-en-v1.5` (33M params, 67 MB ONNX, CPU) | 76.7 s (all 4,114 segments; ~3–5 s/chapter) | **Orthogonal** to Phases 1+2 (`r = +0.14` to `+0.27`). Catches concrete "treading water" (`0.229` vs `0.431` mean) and measures clause/sentence **semantic velocity** (`meaning–connection–meaning`). | **Worth considering as an optional pure-Go or cached add-on** (see Pure-Go section below). |
| **Phase 3B: Small Causal LM Surprisal (Pilot)** | `HuggingFaceTB/SmolLM2-360M` (360M base LM, 724 MB, CPU) | ~240 s est. (457 ms/para on 8 threads) | Separates fresh concrete phrasing (`6.0–6.9` bits) from essay-speak boilerplate (`3.8–5.3` bits) *once word simplicity is held constant*; raw surprisal alone is confounded by rare jargon. | **Do not commit yet.** Raw surprisal rewards awkward jargon (`7.15` bits) and under-scores crisp idioms like `"Science marches on."` |

---

## 2. Answering Your Architecture & Model Questions

### A. Would "linkling2 mini" (Inkling-2-Mini / Ling-mini-2.0) be good here?
Neither is a good fit for a fast local prose lint:
- **Inkling-2-Mini** (`AtomicChat/Inkling-2-Mini-GGUF`): 26.4B total parameters MoE (3.4B active, looped core). Even quantized, it is heavy on CPU, requires patched/Unsloth `llama.cpp` builds rather than mainline `llama.cpp` to avoid gibberish output, and is designed for chat/reasoning rather than clean base-model token log-probabilities.
- **Ling-mini-2.0** (`inclusionAI/Ling-mini-2.0-GGUF`): 16B MoE (1.4B active, `bailingmoe2` architecture), also requires branch-specific `llama.cpp` support and is instruction-tuned.
- **Why base/embedding models beat MoE chat models for a lint:** Instruct/chat tuning sharpens probability distributions toward assistant clichés and distorts raw surprisal. For a CPU prose lint, **small modern mainline models** win on both accuracy and operational simplicity:
  - **Embeddings (semantic leap / velocity):** `BAAI/bge-small-en-v1.5` (33M, 67 MB quantized ONNX) or `nomic-ai/modernbert-embed-base`.
  - **Modern small base LM (if we ever want `llama.cpp` log-probs):** `LiquidAI/LFM2-1.2B` / `LFM2-350M` (day-one mainline `llama.cpp` support), `Qwen/Qwen3-0.6B-Base`, or `HuggingFaceTB/SmolLM2-360M`.

### B. Pure Go Feasibility
- **Phases 1 + 2 are already 100% pure Go** (stdlib only, zero external Go modules, zero CGo, **0.37 s** for the whole book).
- **Phase 3A (Embeddings) CAN be pure Go:**
  - `MichaelAyles/goformer` runs BERT-family embedding models — specifically including **`BAAI/bge-small-en-v1.5`** — in pure Go directly from `safetensors` (no CGo, no ONNX Runtime, no Python).
  - `knights-analytics/hugot` also provides an ONNX pipeline with a pure-Go backend for small embedding models.
- **Phase 3B (GGUF Causal LM):** True pure-Go GGUF matmul for 350M–1.2B models is slow without SIMD/CGo kernels. The clean Go architecture if we ever add a GGUF LM is calling a local `llama-server` over HTTP (zero CGo in `civ/lint`, keeping `civ/bin/lint` a clean static Go binary).

---

## 3. Phases 1 + 2: What Was Built in `civ/lint`

Formalized in [cadence.go](file:///home/d/code/po/civ/lint/cadence.go), [cadence_test.go](file:///home/d/code/po/civ/lint/cadence_test.go), [main.go](file:///home/d/code/po/civ/lint/main.go), and documented in [06_editing_guide.md](file:///home/d/code/po/civ/notes/06_editing_guide.md) (Part K). All 20 Go unit tests in `civ/lint` pass (`0.068 s`).

### Four Local Deterministic Lexicons (stored in `ext-ref/`, gitignored)
1. **CMU Pronouncing Dictionary** (`ext-ref/phonetics/cmudict.dict`, 126,052 entries) — syllable counts and `0/1/2` lexical stress, adjusted for connected-speech prosody (unstressed mid-sentence function words drop to `0`, while sentence-final stressed particles and negators like `"Science marches on."` and `"...did not."` retain `1`).
2. **Brysbaert Concreteness 2014** (`ext-ref/concreteness/brysbaert_2014.tsv`, 39,954 lemmas, 1–5 scale).
3. **Kuperman Age of Acquisition + SUBTLEX POS/Lemma** (`ext-ref/lexicon/kuperman_aoa.tsv`, 51,695 entries — auto-fetched via `civ/bin/lint fetch-lexicons`). Gives word/lemma AoA (years), dominant part of speech (`Verb`/`Noun`/etc.), and exact lemmatization.
4. **Princeton WordNet 3.1 Polysemy** (`ext-ref/lexicon/wordnet_senses.tsv`, 83,253 single-word lemmas — auto-extracted from `wn3.1.dict.tar.gz` via Go stdlib `gzip`/`tar`). Captures connotative breadth (`log2(1 + sense_count)`): broad lived words (*cut* = 61, *stand* = 24, *march* = 14, *stone* = 13) vs. narrow technical monosemes (*procedural* = 1, *physiological* = 1, *mitigation* = 2).

### Anchored Sub-Scores (`0–100`, higher = stronger voice presence)
- **Rhythm (38%):** primary stress-beat density (`Str%`), 4+ weak-syllable trough penalty (`Trough/100w`), masculine stressed landings (weighted 55% paragraph-final, 45% all sentences), and sentence-length variation + setup-to-short-punch contrast.
- **Weight (42% — "greater meaning per word"):** low syllables per content word (`Syl/CW`), high concreteness (`Conc`), early Age of Acquisition (`AoA`), high WordNet polysemy (`Poly`), and kinetic verb share vs. static/abstract linkage verbs (`is/are/demonstrate/constitute/require/enable`).
- **Punch (20% — "meaning–connection–meaning"):** absence of abstract/Latinate prepositional mud chains (`Mud/100w`, e.g., *"of procedural safeguards for the mitigation of institutional failure"*), balanced clause pivots (`;`, `:`, `, but`, `, yet`, `, and` + setup→punch turns), and closing-sentence snap.

### Full-Book Baseline (`civ/bin/lint cadence`, 0.37 s)

| Chapter | Paras (>=45w) | Words | **Snap** | Rhy | Wgt | Pun | Str% | ParaEnd% | Syl/CW | Conc | AoA | Poly | Mud |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `01_ch01.md` | 24 | 1,952 | **59.0** | 48.7 | 63.2 | 69.5 | 37.8% | 54.2% | 2.17 | 3.01 | 7.2 | 2.74 | 0.87 |
| `02_ch02.md` | 47 | 4,701 | **60.9** | 47.5 | 67.5 | 72.4 | 38.2% | 38.3% | 2.09 | 3.11 | 7.2 | 2.80 | 0.49 |
| `03_ch03.md` | 30 | 2,599 | **52.8** | 40.9 | 57.2 | 66.3 | 35.7% | 40.0% | 2.26 | 2.87 | 7.5 | 2.74 | 1.12 |
| `04_ch04.md` | 35 | 3,308 | **57.7** | 46.4 | 62.8 | 68.3 | 37.4% | 37.1% | 2.18 | 3.05 | 7.2 | 2.67 | 0.91 |
| `05_ch05.md` | 18 | 1,325 | **63.6** | 55.2 | 66.1 | 74.1 | 39.5% | 77.8% | 2.12 | 3.22 | 7.4 | 2.59 | 0.89 |
| `06_ch06.md` | 32 | 3,037 | **60.9** | 47.2 | 67.9 | 72.4 | 38.3% | 34.4% | 2.08 | 3.14 | 7.1 | 2.75 | 0.62 |
| `07_ch07.md` | 47 | 3,608 | **64.6** | 53.7 | 70.2 | 73.6 | 40.1% | 44.7% | 2.03 | 3.20 | 7.1 | 2.78 | 0.74 |
| `08_ch08.md` | 49 | 3,506 | **62.4** | 49.9 | 69.1 | 72.1 | 40.0% | 38.8% | 2.03 | 3.17 | 7.0 | 2.67 | 0.39 |
| `09_ch09.md` | 70 | 6,755 | **58.1** | 45.5 | 63.5 | 70.8 | 37.8% | 35.7% | 2.13 | 3.10 | 7.6 | 2.67 | 0.65 |
| `10_ch10.md` | 33 | 2,767 | **61.5** | 48.9 | 67.5 | 72.9 | 38.5% | 45.5% | 2.10 | 3.15 | 7.1 | 2.71 | 0.58 |
| `11_ch11.md` | 42 | 3,281 | **59.9** | 49.2 | 64.7 | 70.1 | 38.5% | 42.9% | 2.17 | 3.14 | 7.4 | 2.71 | 0.77 |
| `12_ch12.md` | 43 | 4,778 | **53.7** | 47.6 | 55.1 | 62.4 | 36.9% | 46.5% | 2.27 | 2.88 | 7.9 | 2.60 | 1.61 |
| `13_ch13.md` | 32 | 2,395 | **50.7** | 40.8 | 52.3 | 66.2 | 36.2% | 40.6% | 2.38 | 3.01 | 8.2 | 2.45 | 0.95 |
| `14_ch14.md` | 20 | 1,543 | **51.4** | 41.9 | 53.1 | 66.1 | 35.5% | 50.0% | 2.34 | 2.99 | 8.1 | 2.61 | 1.78 |
| **BOOK MEAN** | **522** | **45,555** | **58.7** | **47.5** | **63.5** | **70.0** | **38.0%** | **42.5%** | **2.15** | **3.08** | **7.4** | **2.68** | **0.83** |

Notice how cleanly the tool separates your ear's best chapters (`07_ch07` **64.6**, `05_ch05` **63.6**, `08_ch08` **62.4**) from the heaviest institutional/abstract chapters (`13_ch13` **50.7**, `14_ch14` **51.4**, `03_ch03` **52.8**, `12_ch12` **53.7**). Its #1 highest-scoring paragraph in the entire book is the **Cold Forge** ([07_ch07.md:73](file:///home/d/code/po/civ/chapters/07_ch07.md#L73), **90.8**).

---

## 4. Phase 3 Pilot: What Do Local Neural Nets Add as a Delta?

We tested five deliberate calibration probes plus the real book across all three layers:

| Calibration Probe | Pure-Go **Snap** (R/W/P) | `bge-small` `sent_leap` / `sem_vel` | `SmolLM2-360M` `mean` / `end` / `spike` | What Happened |
|---|---:|---:|---:|---|
| **`USER_QUIP`** (`"Numerous studies... Science marches on."`) | **54.5** (56/33/96) | **0.404** / **1.23** | 4.12 / 5.05 / +1.33 | Pure Go catches the +22.7 jump (`Punch=96`); embeddings catch the +48% velocity jump (`0.83→1.23`). |
| **`USER_NO_QUIP`** (same quote + `"This finding confirms..."`) | **31.8** (25/28/53) | **0.331** / **0.83** | 4.45 / 6.82 / +3.10 | Pure Go and embeddings both drop sharply. **LM surprisal got fooled** (+3.10 spike!) because rare jargon (`cohorts`) has higher raw surprisal than the crisp idiom `"Science marches on."` |
| **`TREAD_WATER`** (4 concrete sentences saying the mason built a stone wall in circles) | **87.2** (82/100/70) | **0.229** / 2.37 | 4.49 / 4.45 / +0.03 | **Pure Go was completely fooled (87.2).** Embeddings immediately exposed the circularity (`sent_leap=0.229`, half of book mean `0.431`); LM showed flat `sd=3.01, spike=+0.03`. |
| **`CLICHE_TRAP`** (4 unrelated short concrete clichés) | **88.4** (85/100/70) | 0.423 / 5.08 | 5.57 / 5.55 / +0.11 | Pure Go and embeddings were both fooled (unrelated topics look like big semantic leaps). Even a base LM gave `5.57` bits because jumping between unrelated clichés is surprising. |
| **`MUD_PROPOSAL`** (bureaucratic nominalization pileup) | **8.5** (11/7/8) | **0.300** / **1.04** | 7.15 / 6.43 / -1.04 | Pure Go nailed it (`8.5`); embeddings caught low leap (`0.300`); **raw LM surprisal spiked to 7.15** because stiff Latinate words are rare! |

### Book-Wide Findings on Phase 3
1. **`BAAI/bge-small-en-v1.5` (all 522 paragraphs):**
   - Low correlation with pure-Go `Snap` (`r = +0.141` for `sent_leap`, `r = +0.193` for `max_turn`, `r = +0.273` for `sem_vel`) — proving it captures a **genuinely different dimension** (conceptual movement / turn).
   - Its top two end-turns in the entire book are genuinely strong voice turns: the **Cold Forge** ([07_ch07.md:73](file:///home/d/code/po/civ/chapters/07_ch07.md#L73), `end_turn = 0.661`, `Snap = 90.8`) and **Thomas Reid answering Hume** ([02_ch02.md:41](file:///home/d/code/po/civ/chapters/02_ch02.md#L41), `end_turn = 0.660`, `Snap = 66.8`, ending with *"I break my nose against a post... and am clapped into a mad-house."*).
   - **Known false positive to watch:** Deliberate parallel contrasts (`"drivers keep to the right-hand lane. ... drivers keep to the left-hand lane."` at [05_ch05.md:34](file:///home/d/code/po/civ/chapters/05_ch05.md#L34), `sent_leap = 0.274`; `"Procedural symmetry holds... Value symmetry holds..."` at [09_ch09.md:247](file:///home/d/code/po/civ/chapters/09_ch09.md#L247), `sent_leap = 0.262`) look artificially "close" in embedding space because they reuse the same frame and flip one antonym. So low `sent_leap` is a **signal to check for treading water**, never an automatic penalty.
2. **`HuggingFaceTB/SmolLM2-360M` (45 stratified paragraphs: 15 lowest, 15 middle, 15 highest `Snap`):**
   - `Corr(mean_surp, Snap) = -0.084` (`low15 = 5.26`, `mid15 = 5.23`, `hi15 = 5.15`) — raw surprisal is confounded by jargon vs. fresh imagery.
   - One real signal did appear: on the 8 rewrites below, where we stripped Latinate boilerplate into concrete prose, **6 of 8 rewrites jumped in `mean_surp` (`+0.13` to `+1.66` bits/token)** because academic essay-speak (*"take their foundational institutions for granted"*, *"exercise magistracy and institutional authority"*) is statistically predictable, whereas fresh concrete wording (*"treat its ground as air"*, *"spends its strength on locks"*) is genuinely surprising to a language model. Still, that signal only makes sense *conditioned on* simple vocabulary, and costs ~457 ms/paragraph on CPU.

---

## 5. Eight Muddiest Paragraphs Pulled by the Tool — Iteratively Rewritten & Scored

Below are 8 paragraphs drawn straight from the lowest-Snap tail of `civ/bin/lint cadence` (`26.4` to `34.8`). Every rewrite preserves the exact argument and doctrinal terms, passes `civ/bin/lint prose` with **0 Part A / spelling violations**, and was iteratively tightened (`ch08:86` and `ch11:50` underwent a second pass to eliminate residual mud and land on a sharp masculine beat). None of `civ/chapters/*.md` has been modified.

### Scoreboard Summary (All 8 Pairs Across All 3 Layers)

| Passage | **Snap** (Go Ph 1+2) | Rhythm | Weight | Punch | Str% | Syl/CW | Mud | ParaEnd | `bge-small` SemVel | `bge-small` ClauseLeap | `SmolLM2` MeanSurp |
|---|---:|---:|---:|---:|---:|---:|---:|:---:|---:|---:|---:|
| 1. `03_ch03:52` | `26.4 → 69.6` (**+43.2**) | `26.7 → 61.2` | `26.9 → 66.4` | `25.0 → 92.5` | `33% → 46%` | `2.77 → 1.82` | `6.1 → 0.0` | `weak → STR` | `1.15 → 2.40` | `0.61 → 0.42` | `3.80 → 4.53` |
| 2. `13_ch13:5` | `27.7 → 83.5` (**+55.8**) | `17.8 → 80.4` | `32.4 → 87.1` | `36.5 → 82.0` | `29% → 46%` | `2.79 → 1.60` | `2.1 → 0.0` | `weak → STR` | `2.22 → 2.51` | `0.00 → 0.51` | `5.99 → 6.12` |
| 3. `09_ch09:75` | `28.4 → 82.3` (**+53.9**) | `11.3 → 64.2` | `31.1 → 93.7` | `55.3 → 92.5` | `31% → 47%` | `2.72 → 1.47` | `2.6 → 0.0` | `weak → STR` | `0.94 → 1.74` | `0.40 → 0.53` | `4.86 → 6.36` |
| 4. `08_ch08:86` (p2) | `28.9 → 79.8` (**+50.9**) | `14.6 → 67.5` | `42.5 → 88.4` | `27.5 → 85.0` | `33% → 46%` | `2.49 → 1.50` | `3.3 → 0.0` | `weak → STR` | `0.59 → 1.37` | `0.00 → 0.56` | `5.24 → 6.56` |
| 5. `14_ch14:114` | `29.8 → 90.6` (**+60.8**) | `9.4 → 91.9` | `47.6 → 92.6` | `31.2 → 83.9` | `33% → 59%` | `2.57 → 1.52` | `8.9 → 0.0` | `weak → STR` | `0.00 → 1.21` | `0.40 → 0.38` | `6.26 → 6.15` |
| 6. `01_ch01:53` | `31.5 → 75.4` (**+43.9**) | `19.9 → 66.2` | `47.0 → 89.8` | `21.2 → 62.5` | `32% → 47%` | `2.62 → 1.57` | `3.3 → 0.0` | `weak → STR` | `0.87 → 1.18` | `0.40 → 0.51` | `5.26 → 6.92` |
| 7. `12_ch12:61` | `34.0 → 86.7` (**+52.7**) | `24.7 → 81.3` | `31.9 → 85.3` | `55.9 → 100.0` | `33% → 45%` | `2.76 → 1.68` | `3.2 → 0.0` | `weak → STR` | `0.58 → 1.36` | `0.49 → 0.48` | `5.79 → 5.74` |
| 8. `11_ch11:50` (p2) | `34.8 → 84.7` (**+49.9**) | `30.0 → 78.4` | `46.6 → 91.7` | `19.1 → 82.1` | `33% → 48%` | `2.75 → 1.60` | `3.6 → 0.0` | `weak → STR` | `1.20 → 2.18` | `0.00 → 0.37` | `5.26 → 6.41` |

---

### Side-by-Side Passages (Judge With Your Own Ear)

#### 1. [03_ch03.md:52](file:///home/d/code/po/civ/chapters/03_ch03.md#L52) — Postmodern Constructivism (`26.4 → 69.6`, `#1 muddiest in book`)
- **Before (`Snap 26.4`, `Mud 6.1`, `End weak`):**
  > The third major rival to correspondence requires no theology at all; it dominates large portions of modern Western humanities departments, schools of education, and administrative bureaucracies. This is **Postmodern Social Constructivism**, associated with theorists such as **Michel Foucault** (1926–1984) and **Jacques Derrida** (1930–2004) and carried into institutional practice by critical theory.
- **After (`Snap 69.6`, `Mud 0.0`, `End STR`, `SemVel 1.15 → 2.40`):**
  > The third rival to truth needs no priest and no creed. It rules Western humanities halls, teachers' colleges, and state bureaus under the name **Postmodern Social Constructivism**. **Michel Foucault** (1926–1984) and **Jacques Derrida** (1930–2004) forged the doctrine; critical theory put it in charge.
- **What changed:** Stripped the prepositional mud chains (*"large portions of modern Western humanities..."*, *"carried into institutional practice by..."*), broke the setup into a clean short opener, and locked the ending onto a `meaning–connection–meaning` semicolon pivot landing on a stressed monosyllable (`charge`).

#### 2. [13_ch13.md:5](file:///home/d/code/po/civ/chapters/13_ch13.md#L5) — The Generational Criterion (`27.7 → 83.5`)
- **Before (`Snap 27.7`, `Str 29.3%`, `Syl/CW 2.79`, `End weak`):**
  > The first is **the Generational Criterion (70–100 years)**. Never evaluate a policy by its immediate sentiment during the first eighteen months of borrowed expenditure or revolutionary enthusiasm. Evaluate it by the condition of the country when the grandchildren inherit the resulting institutions and balance sheet seventy years later.
- **After (`Snap 83.5`, `Str 45.8%`, `Syl/CW 1.60`, `End STR`, `ClauseLeap 0.00 → 0.51`):**
  > The first test is **the Generational Criterion (70–100 years)**. A law cannot be judged by the cheers of its first eighteen months, when borrowed cash still buys the crowd. Judge it by the house, the ledger, and the land the grandchildren hold seventy years on.
- **What changed:** Replaced narrow bureaucratic phrases (*"immediate sentiment"*, *"borrowed expenditure"*, *"resulting institutions"*) with broader concrete words (*cheers, cash, buys, crowd, house, ledger, land, hold*) and shifted the closing beat from weak `LA-ter` to stressed `ON`.

#### 3. [09_ch09.md:75](file:///home/d/code/po/civ/chapters/09_ch09.md#L75) — Refusing Authority (`28.4 → 82.3`)
- **Before (`Snap 28.4`, `Rhythm 11.3`, `AoA 9.3`, `End weak`):**
  > If conscientious, self-governed citizens refuse to exercise magistracy and institutional authority out of a reluctance to judge others, sovereign power does not cease to exist. It remains vacant only until the most unscrupulous faction, bureaucrat, or criminal seizes that jurisdiction and exercises it over the households of the conscientious. Refusing to exercise authority within one's proper scope is not humility; it is a dereliction of duty that surrenders governance to the least principled members of society.
- **After (`Snap 82.3`, `Rhythm 64.2`, `AoA 5.7`, `End STR`, `SemVel 0.94 → 1.74`, `LM mean 4.86 → 6.36`):**
  > When good men refuse the bench because they shrink from judging others, power does not vanish. The seat stays empty only until the worst faction, clerk, or crook takes the gavel and rules their homes. Walking off a post that belongs to you is not humility; it hands the keys to the worst man in town.
- **What changed:** Swapped 6 Latinate words in one breath (*conscientious, magistracy, institutional, authority, reluctance, jurisdiction, dereliction, governance*) for concrete office symbols (*bench, seat, gavel, post, keys*) carrying the same meaning in half the syllables (`Syl/CW 2.72 → 1.47`), ending on `town`.

#### 4. [08_ch08.md:86](file:///home/d/code/po/civ/chapters/08_ch08.md#L86) — Dispositional Love (`28.9 → 71.1 [Pass 1] → 79.8 [Pass 2]`)
- **Before (`Snap 28.9`, `Mud 3.3`, `End weak`):**
  > For that reason, dispositional love is the proper default disposition toward all peaceful human beings across every tier (Tiers 1 through 5), until and unless a specific person or group proves by conduct to be a swindler, criminal, or enemy. A society where strangers extend dispositional love to one another is far more orderly and cooperative than a society of suspicious antagonists.
- **Pass 1 (`Snap 71.1`, `Mud 1.9`, `Punch 48.6` — still slightly soft at the end):**
  > Dispositional love should therefore be the standing posture toward every peaceful stranger across all five tiers (Tiers 1 through 5), until a man or group proves by deed to be a cheat, a thief, or a foe. A town where strangers wish each other well trades faster and stands longer than a town of locked doors.
- **Pass 2 (`Snap 79.8`, `Mud 0.0`, `Punch 85.0`, `ClauseLeap 0.00 → 0.56`, `LM end_spike -0.59 → +0.62`):**
  > Dispositional love is the right default toward every peaceful stranger across all five tiers (Tiers 1 through 5); hold it until a man proves by deed that he is a cheat, a thief, or a foe. A town where strangers wish each other well builds more and trades faster. A town of barred doors spends its strength on locks.
- **Why iterating mattered:** Pass 1 still had a single long comparison sentence at the end and `Mud 1.9`. Splitting the contrast into two beats in Pass 2 created a sharp closing understatement (*"spends its strength on locks"*) that jumped `Punch` from `48.6` to `85.0` and flipped `SmolLM2`'s closing-sentence surprisal spike from `-0.59` to `+0.62`.

#### 5. [14_ch14.md:114](file:///home/d/code/po/civ/chapters/14_ch14.md#L114) — Downward Vector Sum (`29.8 → 90.6`)
- **Before (`Snap 29.8`, `Rhythm 9.4`, `Mud 8.9`, 1 sentence of 45w, `End weak`):**
  > When millions of individuals direct their daily vectors toward falsehood, evasion of responsibility, negligent work, cynicism, childlessness, and the consumption of inherited capital, the resultant vector of the whole society points downward toward decline, and no prime minister or central bank can reverse that trajectory.
- **After (`Snap 90.6`, `Rhythm 91.9`, `Mud 0.0`, `Str 58.5%`, `End STR`):**
  > When millions aim their daily work at lies, dodged duty, sloppy craft, empty cradles, and spent inheritance, the sum of those lives pulls the whole country down. No prime minister and no central bank can turn that sum back up.
- **What changed:** Broke the 45-word mud sentence (`Mud 8.9`) into a concrete setup + 13-word stressed punch (`down` / `up`), replacing *"childlessness, and the consumption of inherited capital"* with *"empty cradles, and spent inheritance"*.

#### 6. [01_ch01.md:53](file:///home/d/code/po/civ/chapters/01_ch01.md#L53) — Hidden Premises (`31.5 → 75.4`)
- **Before (`Snap 31.5`, `Mud 3.3`, `End weak`):**
  > Stating and deriving the common principles of civilization explicitly is necessary because in any society that has endured for several generations, citizens begin to take their foundational institutions for granted. They argue fiercely about secondary policies, such as tax rates, criminal sentencing, school discipline, border enforcement, or welfare programs, while leaving the foundational premises on which their arguments rest unstated.
- **After (`Snap 75.4`, `Mud 0.0`, `End STR`, `ClauseLeap 0.40 → 0.51`, `LM mean 5.26 → 6.92`):**
  > We must state these common principles in plain words, because any people that lasts a few generations starts to treat its ground as air. Citizens fight over tax rates, prison terms, school rules, borders, and relief checks, while the deeper premises under every fight go unnamed.
- **What changed:** Cut the gerund-abstract opening, replaced *"take their foundational institutions for granted"* (cliché essay-speak) with *"treat its ground as air"* (`LM mean_surp` jumped `+1.66` bits/token!), and ended on stressed `un-NAMED` instead of weak `un-STA-ted`.

#### 7. [12_ch12.md:61](file:///home/d/code/po/civ/chapters/12_ch12.md#L61) — David Hume (`34.0 → 86.7`)
- **Before (`Snap 34.0`, `Syl/CW 2.76`, `AoA 9.4`, `End weak`):**
  > **David Hume** was a penetrating historian and economist outside his study, understanding property, contract, and the danger of political fanaticism; yet inside his philosophical treatise, his premise that we perceive only disconnected mental impressions led him to claim that we possess no rational justification for believing in cause and effect or the permanence of external objects. He severed speculative philosophy from common-sense reality.
- **After (`Snap 86.7`, `Punch 100.0`, `Syl/CW 1.68`, `End STR`, `SemVel 0.58 → 1.36`, `LM spike -0.01 → +0.38`):**
  > Outside his study, **David Hume** wrote shrewd history and sound economics; he knew property, contract, and the fire of political zeal. Inside his study, he began from the premise that the mind sees only loose flashes of sense, and argued that we have no firm ground to trust cause, effect, or the stone beneath our feet. He cut philosophy loose from the world he walked on.
- **What changed:** Turned a 54-word tangle into an explicit *Outside his study / Inside his study* structure, replaced *"permanence of external objects"* with *"the stone beneath our feet"*, and closed on a dry 10-word masculine punch (*"He cut philosophy loose from the world he walked on."*).

#### 8. [11_ch11.md:50](file:///home/d/code/po/civ/chapters/11_ch11.md#L50) — Corrupted Material Welfare (`34.8 → 74.5 [Pass 1] → 84.7 [Pass 2]`)
- **Before (`Snap 34.8`, `Punch 19.1`, `Mud 3.6`, `End weak`):**
  > In material welfare, politicians and bureaucracies convert a temporary emergency provision, designed to help a displaced worker or a widowed mother regain self-sufficiency, into an open-ended, multi-generational subsidy that **penalizes marriage and work**. It pays a household more if no working father resides in the home and withdraws benefits if the recipient accepts entry-level employment. Within two generations, what was enacted in the name of "compassion" has undermined family formation in entire neighborhoods and habituated millions of capable human beings to permanent dependency.
- **Pass 1 (`Snap 74.5`, `Mud 3.8`, `Punch 45.1` — residual `"In relief policy... in the name of mercy"` mud):**
  > In relief policy, lawmakers and bureaus turn a short bridge for a laid-off worker or a young widow into a lifelong check that **penalizes marriage and work**. The rules pay a home more when no working father lives there, and they cut the check the week a parent takes a starter job. Within two generations, laws passed in the name of mercy have hollowed out whole neighborhoods and trained millions of able men and women to wait on the state.
- **Pass 2 (`Snap 84.7`, `Mud 0.0`, `Punch 82.1`, `SemVel 1.20 → 2.18`, `LM mean 5.26 → 6.41`):**
  > Lawmakers and bureaus turn a short bridge for a laid-off worker or a young widow into a lifelong check that **penalizes marriage and work**. The rules pay a home more when no working father lives there, and they cut the check the week a parent takes a starter job. Within two generations, mercy written into bad law has hollowed out whole neighborhoods. Millions of able hands now wait on the state.
- **Why iterating mattered:** Dropping the two filler prepositional frames and breaking out the 9-word closing sentence eliminated all mud (`3.8 → 0.0`), raised `SemVel` from `1.20` to `2.18`, and landed on stressed `state`.

---

## 6. Decisions for You (Next Steps)

1. **Do these 8 rewrites match the "snap / understated meaning / rhythm" you want?**
   - If **yes**, we can either queue these 8 (or the full bottom ~25–30 muddiest paragraphs across Chapters 1–14) into `civ/bin/review` after you finish your open 24-item Plain-Style queue, and optionally fold `cadence` into `civ/bin/lint sim` so packet pre-flight tracks both `voice` (anti-textbook) and `cadence` (pro-snap) together.
   - If **some rewrites feel too punchy or lose needed precision** (e.g., in conceptual/philosophical chapters like `03_ch03` or `12_ch12`), tell me which of the 8 hit the bullseye and which overshot — we can tune the sub-score weights in [cadence.go](file:///home/d/code/po/civ/lint/cadence.go) immediately.
2. **Phase 3 (Neural Net Delta) — proceed or hold?**
   - **Option A (Recommended for now): Stay pure-Go Phases 1 + 2.** It runs in `0.37 s`, has zero runtime dependencies, and already surfaced all 8 muddy paragraphs above cleanly.
   - **Option B: Add pure-Go `BAAI/bge-small-en-v1.5` embeddings** (via `goformer` or a cached optional subcommand `civ/bin/lint cadence -semantic`) to catch concrete treading-water and track `SemVel` / `ClauseLeap` without Python or CGo.
   - **Option C: Skip causal-LM surprisal (`SmolLM2` / `llama.cpp`) for linting** unless we later want a dedicated cliché-vs-freshness classifier — raw surprisal did not show a large enough improvement over Phases 1 + 2 to justify 457 ms/paragraph and jargon confounding.

---

## 7. Addendum: Phase 3 Option B Implemented (Pure Go)

Implemented in [semantic.go](file:///home/d/code/po/civ/lint/semantic.go) and [semantic_test.go](file:///home/d/code/po/civ/lint/semantic_test.go) using `github.com/MichaelAyles/goformer` (MIT, zero transitive dependencies, zero CGO, zero ONNX, zero Python), loading `BAAI/bge-small-en-v1.5` (`model.safetensors`, 127.3 MB) from `ext-ref/embed/bge-small-en-v1.5/` with a SHA-256 binary vector cache at `ext-ref/embed/cache_bge_small.bin`:

- **Cold full-book run** (`civ/bin/lint cadence -semantic`, 3,978 uncached sentence/clause segments across 16 Go workers): `5m 01s` (one-time cache build).
- **Warm cached full-book run** (all 522 paragraphs, 4 lexicons + 3,978 vectors): **`0.482 s`**.
- **Book semantic means** (`goformer` mean-pooled `BGE-small-en-v1.5`): `SentLeap = 0.441`, `EndTurn = 0.441`, `ClauseLeap = 0.463`, `SemVel = 1.34`.
- **Calibrated `TreadWater` cutoff:** `SentLeap < 0.28` on paragraphs with `>= 3` sentences (`TREAD_WATER` probe scores `0.252` and flags `TreadWater=true`; deliberate parallel contrasts at `0.281–0.298` are spared).
- **Double-flagged book paragraphs** (low deterministic `Snap` *and* lowest semantic movement): [04_ch04.md:91](file:///home/d/code/po/civ/chapters/04_ch04.md#L91) (`Snap 35.4`, `Leap 0.288`, `Vel 0.46`), [11_ch11.md:143](file:///home/d/code/po/civ/chapters/11_ch11.md#L143) (`Snap 44.7`, `Leap 0.311`, `Vel 0.54`), and [04_ch04.md:119](file:///home/d/code/po/civ/chapters/04_ch04.md#L119) (`Snap 46.9`, `Leap 0.316`, `Vel 0.91`).


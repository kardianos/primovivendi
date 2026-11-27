# 48. Spec (Round 2): Pure-Go BiGRU + Biaffine Graph Parser & Coref Parity (`civ/sys/parser`)

Recorded 2026-11-26. Round-2 implementation specification building on the committed Round-1 pure-Go NLP pipeline (`7db5f0c`, see [47_go_parser_spec.md](file:///home/d/code/po/civ/notes/47_go_parser_spec.md) and [47a_go_parser_handoff.md](file:///home/d/code/po/civ/notes/47a_go_parser_handoff.md)).

Hand this file to the implementation agent. When complete, do **not** commit or push; leave a clean working tree with packed models and a handoff note (`civ/notes/48a_go_parser_r2_handoff.md`) for independent review and integration.

---

## 1. Diagnosis of Round 1 & Round-2 Objective

Round 1 already delivered and verified:
* Zero-CGO, zero-Java pipeline with native UTF-8 byte offsets and zero `unicode/utf16` downstream.
* Stage 1 (`tok.go`, `lemma.go`): 100.0% sentence/token alignment with CoreNLP on `ch13+ch14`, 99.8% lemma parity.
* SIMD acceleration (`civ/sys/vec` AVX2+FMA with `ClearAVXUpperBits()`), `< 45 MB` live heap (`36.6 MB`), `0.62 ms/sent` latency, `0.27 s` full book, and `7.34 MB` compressed F16+gzip weights in `civ/sys/parser/models/`.

Round 1 missed the **accuracy gates** for two structural reasons:
1. **Greedy local-window MLP ceiling**: The static-oracle Chen & Manning feedforward MLP in `dep.go` sees only fixed local stack/buffer slots with no sentence-wide context, no character morphology for unseen book words, and no global decoding. On 40–70 word sentences, early greedy errors cascade: root agreement drops to `79.7%`, UAS falls to `65–72%`, and `civ/linear` Rule 10 correlation stalls at `r = 0.69` (gate `≥ 0.92`).
2. **Domain / convention friction & coref recall**: Contradictory punctuation attachments between UD (`UD_English-EWT`/`GUM`) and CoreNLP account for a third of all holdout disagreements (`566 / 1,700`), POS training weighted CoreNLP chapters `4×` which pulled UD-EWT POS down to `94.92%`, and `coref.go` achieved `0.46` F1 on `civ/linear` Part 4 (`ambiguous it` / `distant pronoun`).

### Round-2 Goal
Replace **only** the greedy transition MLP (and optionally the separate POS perceptron) with a **pure-Go 2-layer BiGRU + Dozat-Manning Biaffine Graph Parser + Global MST Decoder**, align training data/conventions cleanly, polish `coref.go` on `ch01..ch12`, and pass every accuracy, latency, memory, and size gate so `civ/linear` and `civ/compose` can retire Java CoreNLP.

### What NOT to Break or Rewrite
* Keep Stage 1 (`tok.go`, `lemma.go`), `st.go` (transparent gzip + F16 safetensors), `conllu.go`, `eval.go`, `civ/linear` wiring, and `civ/compose` wiring intact.
* Keep `Pipeline` 100% read-only/immutable after `Load(root)` and concurrency-safe via `sync.Pool`.
* Keep `ResolveModelPath` priority (`ext-ref/parser/` local overrides → `civ/sys/parser/models/` committed `.safetensors.gz`).
* Keep all Round-1 invariants passing (`TestNoUTF16Usage`, `TestHeldoutDiscipline`, `TestTreeValidityInvariants`, `TestConcurrentParseSentence`, `TestSafetensorsGzipF16RoundTrip`, `TestRaggedLengths`, scalar + `GOEXPERIMENT=simd` + `-race`).
* **No CGO, Python, PyTorch, or external Go dependencies** outside existing [civ/go.mod](file:///home/d/code/po/civ/go.mod).

---

## 2. Architecture: Unified BiGRU + Deep Biaffine Graph Parser

Do **not** spend time tuning the Round-1 greedy static-oracle MLP (`oracle.go` / `dep.go`). Replace Stage 3 with a **global graph-based dependency parser** (Dozat & Manning 2017) built in pure Go over `civ/sys/vec`.

### Recommended Unified Backbone (Stages 2 + 3 Together)
Using a **shared 2-layer BiGRU backbone** where **Layer 1 predicts POS tags** and **Layer 2 predicts Biaffine dependency arcs and labels** solves both the UD-EWT POS gate (`94.92% → ≥ 96.5%`) and the dependency gates in one compact model while *eliminating* the separate `4.26 MB` sparse POS table:

```text
Token i ∈ [1..N] (+ synthetic ROOT at i = 0)
  ├─ Word embedding:   e_w(lower(w_i)) ∈ R^64    (vocab = top 16,384 + <null>,<unk>,<root>)
  ├─ Char-CNN:         e_c(orig(w_i))  ∈ R^48    (char emb 16d; kernels k=2,3,4 × 16 filters + max-pool)
  └─ Word shape/prior: e_s(w_i)        ∈ R^16    (caps/allcaps/digit/hyphen/punct + WordNet 8-bit POS prior)
      └─ Concat x_i ∈ R^128  (multiple of 8 for AVX2!)

Layer 1 BiGRU (H = 128 per direction → h1_i ∈ R^256):
  ├─ POS Head:  logits_pos[i] = W_pos · h1_i + b_pos   (46 PTB classes)
  │             pred_pos[i]   = argmax(logits_pos[i])
  └─ Feed Layer 2: z_i = [h1_i ; e_p(pred_pos[i])] ∈ R^280   (e_p ∈ R^24; during training use gold POS with 10% noise or predicted POS!)

Layer 2 BiGRU (H = 128 per direction → h2_i ∈ R^256, with residual h_i = h2_i + h1_i):
  ├─ MLP_arc_dep(h_i)  ∈ R^128   (SiLU or LeakyReLU(0.1))
  ├─ MLP_arc_head(h_i) ∈ R^128
  ├─ MLP_rel_dep(h_i)  ∈ R^64
  └─ MLP_rel_head(h_i) ∈ R^64

Biaffine Scorers:
  ├─ Arc score matrix S_arc[i, j]  (dep i ∈ 1..N, head j ∈ 0..N, S_arc[i,i] = -inf):
  │      S_arc[i, j] = (a_dep[i])^T W_arc a_head[j] + u_arc^T a_head[j] + v_arc^T a_dep[i]
  │      (equivalently append 1.0 to a_dep[i] → [129 × 128] W_arc)
  └─ Label score vector S_rel[i, j, :]  (ONLY evaluated on gold head j=h*[i] in training, MST head j=ĥ[i] at inference!):
         S_rel[i, j, r] = (r_dep[i])^T U_rel[r] r_head[j] + w_rd[r]^T r_dep[i] + w_rh[r]^T r_head[j] + b_rel[r]

Global Decoder (mst.go):
  └─ Single-root Maximum Spanning Arborescence over S_arc → ĥ[1..N]
  └─ Label argmax over S_rel[i, ĥ[i], r] (for ĥ[i] == 0 force label = "ROOT"; for ĥ[i] != 0 mask out "ROOT")
```

*(Note: If you prefer keeping Stage 2 as a standalone model—e.g. a 1-layer Char-CNN + BiGRU POS tagger—that is also acceptable as long as both UD-EWT POS ≥ 96.5% and total packed size ≤ 15 MB and total latency < 1.5 ms are met. The unified Layer-1 POS / Layer-2 Dep architecture is strictly smaller and faster because it runs the word+char encoder only once.)*

### Exact Math & Dimensions (All Widths Multiples of 8 for AVX2)

#### A. Character CNN (`char.go` or inside `biaffine.go`)
* Alphabet: printable ASCII bytes `32..126` (`95` chars) + `<PAD>=0`, `<BOW>=1`, `<EOW>=2`, `<UNK>=3` → `100` char IDs. Non-ASCII runes map by Unicode category (upper/lower/digit/punct/other) or `<UNK>`.
* Wrap each token's original cased surface form `Orig` with `<BOW>` and `<EOW>` (capped at 24 chars: first 12 + last 12 for long tokens).
* Embedding `E_char [100 × 16]`.
* Three 1D conv kernels: widths `k = 2, 3, 4`, `16` output channels each, ReLU/SiLU, global max-pooling over time → concatenate to `e_c ∈ R^48`.
* **Inference speed optimization**: At `Load()`, precompute `e_c` for the top `4,096` cased surface forms (~768 KB RAM) so >90% of running tokens skip the Char-CNN completely! During training, run Char-CNN live on every token so gradients flow cleanly through `E_char` and conv kernels.
* **Synthetic `ROOT` at index `0`**: Use a dedicated learned vector `root_x ∈ R^128` (and fixed `POS = <ROOT>` for Layer 2).

#### B. Exact GRU Equations (Standard PyTorch Formulation)
Use the exact formulation where the hidden-to-hidden multiply `W_h · h_{t-1}` is a **single** `[3H × H]` matrix-vector product per timestep, and the input-to-hidden multiply `W_x · x_t` is batched across **all $T = N+1$ tokens at once** before the recurrence loop:

1. **Pre-loop batched input projection** (for $t = 0..N$):
   $$u_t = W_x x_t + b_x \in \mathbb{R}^{3H}$$
   Partition $u_t = [u_{r,t}; u_{z,t}; u_{n,t}]$.
2. **Recurrence step** ($t = 0 \to N$ forward; $t = N \to 0$ backward, with $h_{-1} = 0$ / $h_{N+1} = 0$):
   $$v_t = W_h h_{t-1} + b_h \in \mathbb{R}^{3H}, \quad v_t = [v_{r,t}; v_{z,t}; v_{n,t}]$$
   $$r_t = \sigma(u_{r,t} + v_{r,t})$$
   $$z_t = \sigma(u_{z,t} + v_{z,t})$$
   $$n_t = \tanh(u_{n,t} + r_t \odot v_{n,t})$$
   $$h_t = (1 - z_t) \odot n_t + z_t \odot h_{t-1}$$
3. **Exact BPTT backward step** (verify with `TestBiGRUBiaffineGradCheck`!):
   Given upstream gradient $\delta h_t = \frac{\partial \mathcal{L}}{\partial h_t} + \delta h^{\text{next}}_t$:
   $$\delta n_t = \delta h_t \odot (1 - z_t) \odot (1 - n_t^2)$$
   $$\delta z_t = \delta h_t \odot (h_{t-1} - n_t) \odot z_t \odot (1 - z_t)$$
   $$\delta r_t = (\delta n_t \odot v_{n,t}) \odot r_t \odot (1 - r_t)$$
   $$\delta u_t = [\delta r_t;\; \delta z_t;\; \delta n_t], \qquad \delta v_t = [\delta r_t;\; \delta z_t;\; \delta n_t \odot r_t]$$
   $$\delta b_x \mathrel{+}= \delta u_t, \quad \delta b_h \mathrel{+}= \delta v_t, \quad \delta W_x \mathrel{+}= \delta u_t x_t^{\top}, \quad \delta W_h \mathrel{+}= \delta v_t h_{t-1}^{\top}$$
   $$\delta x_t = W_x^{\top} \delta u_t, \qquad \delta h^{\text{prev}}_{t-1} = \delta h_t \odot z_t + W_h^{\top} \delta v_t$$

* **Hidden width**: Start with `H = 128` (`2H = 256`). If UD-EWT UAS lands at 89–90% and latency has headroom (`< 1.0 ms`), `H = 160` (`2H = 320`) is allowed as long as all latency (`< 1.5 ms`), heap (`< 45 MB`), and disk (`≤ 15 MB`) gates hold.
* **Parameter budget (`H = 128`, `d_a = 128`, `d_r = 64`)**:
  * `word.emb` (`16384 × 64`): `1.05M`
  * `char.cnn` + `shape` + `pos.emb`: `~20k`
  * Layer 1 BiGRU (`2 × 3 × 128 × (128+128)`): `197k`
  * Layer 2 BiGRU (`2 × 3 × 128 × (280+128)`): `313k`
  * POS head (`46 × 256`): `12k`
  * Biaffine MLPs + `W_arc` (`129×128`) + `U_rel` (`43×64×64`): `~290k`
  * **Total**: **`~1.88M` parameters** → **`~7.5 MB` F32** → **`~3.8 MB` F16** → **`~3.0–3.6 MB` `.safetensors.gz`**!

#### C. Global Single-Root MST Decoder (`mst.go`)
* Implement **Chu-Liu-Edmonds** Maximum Spanning Arborescence on the $(N+1)$-node directed graph (nodes `0..N`, edges $j \to i$ with score $S^{\text{arc}}[i, j]$ for $i \in 1..N, j \in 0..N, j \ne i$).
* **Hard single-root guarantee**: Standard CLE can select $>1$ children of `0`. Handle this **exactly**:
  1. Run CLE directly on $S^{\text{arc}}$. On ~95% of sentences it already returns a single child of `0` — done in $O(N^2)$ (`< 10 µs`)!
  2. If CLE returns $k > 1$ children of `0`, enforce exact single-root: evaluate each candidate root $r \in \{1..N\}$ (or the top-8 highest-scoring `S_arc[r, 0] - max_{j>0} S_arc[r, j]` candidates, plus all $k$ roots CLE picked) by setting `S_arc[i, 0] = -inf` for all $i \ne r$ and running CLE, keeping the valid single-root tree with the highest total arc score!
  3. As a final defensive invariant, assert `roots == 1`, zero cycles, every $i \in 1..N$ has `head[i] ∈ [0..N] \ {i}`.
* **Unit test `TestMSTSingleRoot`**: Generate 500 random score matrices (lengths `1..12`, including adversarial cases where `0` has the highest incoming edge to every node!) and verify the decoder matches **exact brute-force enumeration** over all valid single-root spanning trees on lengths `1..6`, and always returns an acyclic single-root tree on lengths `1..80`.

#### D. SIMD Kernels (`civ/sys/vec/vec_mat.go`, `vec_simd.go`)
* Follow the exact `ClearAVXUpperBits()` discipline established in [vec_simd.go](file:///home/d/code/po/civ/sys/vec/vec_simd.go) — **never** return to scalar float math with dirty YMM upper bits!
* Add clean batched GEMM/matvec helpers (with both AVX2+FMA and scalar paths tested in `TestRaggedLengths`):
  * `MatMulT(out, A, B []float32, M, N, K int)`: computes `out[m, n] = dot(A[m, :], B[n, :])` (used for batched `W_x · X^T`, Biaffine MLPs, and `Q · K^T` arc scores!). Processing rows of `A` in blocks of 4–8 keeps `B` warm in L1/L2 cache and amortizes `upperClean()`.
  * `OuterAdd(dst, a, b []float32, M, N int)`: `dst[m, n] += a[m] * b[n]` (for fast weight-gradient accumulation during training).
  * `TanhVec` / `SigmoidVec` (or table-interpolated `sigmoidf` / `tanhf` matching Round 1's scalar parity so SIMD and scalar builds agree to `< 1e-4`!).

---

## 3. Mandatory Training Engineering & Guardrails

Every item below is **mandatory** — do not skip any of them. They directly prevent the silent bugs found in Round 1.

### Gate 0: Numerical Gradient Check (Before Any Real Training!)
1. **`TestBiGRUBiaffineGradCheck`** in `civ/sys/parser/parser_test.go`:
   * Instantiate a tiny network (`Dw=8, Dc=8, Dp=4, H=8, da=8, dr=4`, 6 POS classes, 6 Dep labels, `Dropout=0`, `WordDropout=0`).
   * Compute loss and analytic gradients on a 2-sentence mini-batch (lengths 4 and 6).
   * Perturb sampled weights from **every** parameter block (`word.emb`, `char.emb`, `char.conv`, `pos.emb`, `gru1.wx`, `gru1.wh`, `gru1.bx`, `gru1.bh`, `gru2.*`, `pos.head`, `mlp.arc`, `w.arc`, `mlp.rel`, `u.rel`, `b.rel`) by $\pm \epsilon$ (`1e-3` in `float32`, or `1e-4` in a `float64` loss check) and verify analytic vs central-difference gradients match within relative tolerance `2e-2` (or absolute `1e-3`).
   * **Do not start corpus training until `TestBiGRUBiaffineGradCheck` passes 100%!**
2. **`TestOverfitMiniBatch`**:
   * Take 16 real sentences from UD-EWT train (with `Dropout=0`), train for 50 AdamW steps, and assert total loss drops below `0.10` and training UAS/LAS exceeds `98%`.

### Exact Loss, Optimizer & Regularization Specification
1. **Token-mean loss** (NEVER unaveraged sum!):
   * For a mini-batch of `B` sentences containing `M` total real tokens (excluding synthetic `ROOT` token `0`):
     $$\mathcal{L} = \frac{1}{M} \sum_{s=1}^{B} \sum_{i=1}^{N_s} \left( \text{CE}(S^{\text{arc}}_{s,i}, h^*_{s,i}) + \text{CE}(S^{\text{rel}}_{s,i, h^*_{s,i}}, r^*_{s,i}) + \lambda_{\text{pos}} \text{CE}(S^{\text{pos}}_{s,i}, p^*_{s,i}) \right)$$
     with $\lambda_{\text{pos}} = 0.5$ (or `1.0` if training jointly).
   * Divide all accumulated gradients by `float32(M)` **before** computing global L2 norm and clipping at `Clip = 5.0`.
   * Note on `punct` loss weighting (see Section 4 below): when down-weighting `punct` arc loss by $w_i \in (0, 1]$, divide by $\sum w_i$ (or `M`) consistently.
2. **AdamW**:
   * $\beta_1 = 0.9$, $\beta_2 = 0.999$, $\epsilon = 10^{-8}$, `WeightDecay = 1e-4`.
   * Apply decoupled weight decay (`th[i] -= lr * wd * th[i]`) **only** to weight matrices/embeddings — **never** to biases (`b_x`, `b_h`, `b_pos`, `b_arc`, `b_rel`).
   * **Schedule**: Linear warmup over first 5% of total optimizer steps to peak `LR = 2e-3`, then cosine decay to `1e-4` (Phase A). In Phase B fine-tuning, start fresh optimizer or decay from `LR = 3e-4` to `1e-5`.
3. **Explicit Active Regularization (Log at Startup!)**:
   * **Never** use negative-sentinel zero-value defaults! Explicitly populate defaults in a constructor `DefaultBiaffineConfig()` and **log every active hyperparameter** (`lr`, `batch_sents`, `dropout`, `word_dropout`, `wd`, `clip`, `H`) on the first line of training output!
   * **Variational (locked) dropout `0.25`**: Draw ONE mask per sentence per layer (`x_i`, `h1_i`, `h2_i`, `a_dep`, `a_head`, `r_dep`, `r_head`) shared across all timesteps of that sentence, with inverted scaling `1 / (1 - p)` during training and `1.0` at eval.
   * **Word dropout (Kiperwasser & Goldberg)** using **raw (unweighted)** token counts $c(w)$ across unique training sentences:
     $$P(w \to \langle\text{unk}\rangle) = \frac{\alpha}{c(w) + \alpha}, \qquad \alpha = 0.25$$
     *(Keep Char-CNN input intact even when word ID is dropped to `<unk>`, so the network learns to rely on morphology + context for unseen book words!)*
   * **POS input noise for Layer 2 (when training jointly)**: During training, feed Layer 2 the **predicted** POS tag from Layer 1 with probability `0.5` (scheduled sampling) or gold POS with `10%` random dropout to `<UNK>` so Layer 2 is robust to real Layer-1 POS errors at inference.
4. **Parallel Mini-Batch Execution**:
   * `BatchSents = 32` sentences (~600–800 tokens per step), globally shuffled every epoch.
   * Split the 32 sentences across `min(runtime.NumCPU(), 8)` goroutines (4 sentences per worker), each accumulating into private gradient buffers, then parallel-reduce, divide by `M`, clip at `5.0`, and apply AdamW.
   * Keep memory bounded during training (allocate per-worker scratch once, reuse across sentences).

---

## 4. Data Pipeline, Convention Harmonization & Strict Holdout Hygiene

### Holdout Invariant (NON-NEGOTIABLE)
* **`ch13` and `ch14` are STRICTLY HELD OUT.** Never load them in any training, tuning, or dev-selection code (`TestHeldoutDiscipline` must continue enforcing this).
* Any draft in `ext-ref/linear/scratch/` that shares a sentence (`minSharedRunes = 40`, using `eval.go`'s `overlapsSentence`) with `ch13` or `ch14` MUST be **sealed and excluded** from all training and dev loaders! Add a unit test verifying `LoadSealedDrafts` rejects any draft overlapping `ch13`/`ch14`.

### Training & Dev Splits
1. **Gold UD Treebanks (`civ/bin/parser fetch-ud`)**:
   * `UD_English-EWT` (`r2.15`, `12,544` train / `2,002` dev / `2,077` test) — already fetched.
   * `UD_English-GUM` (`r2.15`, `9,520` train / `1,334` dev) — already fetched.
   * **Add `UD_English-LinES`** (`r2.15`, `3,176` train — literary/expository prose, CC BY-NC-SA or check license; if clean/reachable from `UniversalDependencies/UD_English-LinES`, add it to `fetch-ud` as optional alongside GUM!).
2. **Silver CoreNLP Prose (In-Domain)**:
   * 8 reference treatises: `pg4351`, `pg4367`, `pg6435`, `pg935`, `pg13529`, `pg26334`, `pg34901`, `pg35916` (`2,022` sents).
   * Book chapters **`ch01..ch11`** (`~1,820` sents) — **hold out `ch12` (`232` sents) as `dev_book`!**
   * **Sealed scratch drafts** (`ext-ref/linear/scratch/*.txt.json`, excluding any draft overlapping `ch13` or `ch14`, and deduplicating verbatim sentences already in `ch01..ch12` or treatises — adds `~1,200` unseen in-domain sentences!).
3. **Two Clean Dev Sets**:
   * `dev_ud`: `UD_English-EWT` dev (`2,002` sents).
   * `dev_book`: `ch12` (`232` sents, held out from training during development so checkpoint selection and coref tuning NEVER touch `ch13`/`ch14`!).

### Solving the UD vs. CoreNLP Punctuation Convention Split
As Round 1 proved (`47a_go_parser_handoff.md` §4.2), UD-EWT/GUM and CoreNLP attach commas, periods, dashes, and parentheses by **different conventions** (`46%` agreement on punctuation = `566 / 1,700` holdout disagreements). Neither `civ/linear` nor `civ/compose` depends on punctuation head choices for Rule 10 scoring, **but** mixing contradictory punctuation labels into one model hurts *both* UD-EWT and book parsing!

Handle this cleanly and honestly:
1. **During training**:
   * **Phase A (15–20 epochs, `LR = 2e-3 → 1e-4`)**: Train on `UD-EWT (1×)` + `UD-GUM (1×)` + `UD-LinES (1×, if present)` + `treatises (1×)` + `sealed drafts (1×)` + `ch01..ch11 (2×)`.
     * On **CoreNLP silver sentences only**, set the arc-loss weight of `punct` tokens (`Rel == "punct"`) to **`0.0`** (mask `punct` arc loss on silver so the model learns 100% consistent UD punctuation without CoreNLP/UD whiplash, while still learning full content/function syntax, subordination, and prose clause structure from all ~5,000 silver sentences!).
     * Select the best Phase-A checkpoint by `0.5 * dev_ud_LAS + 0.5 * dev_book_nonpunct_LAS` (logging BOTH `dev_ud` and `dev_book` every epoch!).
   * **Phase B (optional 4–6 epochs, `LR = 3e-4 → 1e-5`)**: Only if Phase A's `dev_book` non-punct LAS lags `dev_ud` by `> 1.5` points, fine-tune gently on a balanced mix (`UD-EWT 1×` + `UD-GUM 1×` + `treatises 2×` + `sealed drafts 2×` + `ch01..ch11 3×`, still keeping `dev_ud` LAS within `0.5%` of Phase A's best!).
2. **In `civ/bin/parser eval`**:
   * On **UD-EWT test**: report standard all-token POS, UAS, LAS (100% standard CoNLL/UD metric, no tricks).
   * On **held-out `ch13+ch14` vs CoreNLP**: report **both** all-token UAS/LAS **and** **non-punct UAS/LAS** (excluding tokens whose CoreNLP relation is `punct`), plus **Root agreement %** and **`civ/linear` Rule 10 correlation ($r$)** and **Part 4 F1**!
   * *(Optional: if you also implement a deterministic CoreNLP punctuation-attachment pass, keep it off by default and report raw vs re-attached transparently—never hide raw numbers.)*

---

## 5. Stage 4: Targeted `coref.go` & `ner.go` Polish (For Part 4 F1 ≥ 0.85)

Inspect `civ/linear/analyze.go` (`distant pronoun`, ~line 573) and `civ/linear/style.go` (`it chain`, `ambiguous it`, ~lines 290–380).

1. **Diagnostic pass on `ch01..ch12` ONLY** (never `ch13`/`ch14`!):
   * Run `civ/bin/linear scores -chapter ch01,ch02,...,ch12 -engine corenlp` vs `-engine go`.
   * Print every Part 4 finding (`distant pronoun`, `it chain`, `ambiguous it`) where CoreNLP and Go disagree, along with the sentence text, the parsed dependency tree, and the coreference chains for that sentence.
2. **Fix the concrete mismatches in [coref.go](file:///home/d/code/po/civ/sys/parser/coref.go)**:
   * Verify `dummyIt` in `coref.go` matches `civ/linear/style.go`'s `referentialIt`/`dummyIt` rules exactly.
   * Check why `ambiguous it` under-fired in Round 1 (`CoreNLP 7, Go 3` on ch13+ch14): `civ/linear/style.go`'s `antecedentHead` walks `doc.Corefs` to find the chain containing `m.SentNum == sentNum && m.Head == itIdx` and takes the closest preceding non-pronominal mention in that chain! If `coref.go` either failed to link `"it"` to its inanimate antecedent, picked an intervening noun closer than CoreNLP did (too-short distance bias vs syntactic-subject salience!), or dropped a nominal head in `npHead` (e.g. `obl:npmod` / `nmod:poss` / `conj` nouns!), `addAmbiguousIts` misses the finding.
   * Tune `coref.go` mention extraction and Hobbs/salience weights strictly against `ch01..ch12` until Part 4 F1 on `ch12` (and across `ch01..ch12`) is **≥ 0.88**.
3. **NER (`ner.go`)**:
   * Retrain NER using the new Stage 2/3 predicted POS tags and the expanded silver set (including sealed non-holdout drafts from `ext-ref/linear/scratch/`), adding a lightweight ORG suffix/gazetteer feature (e.g. `University`, `Court`, `Senate`, `Congress`, `Church`, `Party`, `Bank`, `Institute`, `Company`, `Press`, `Times`, `Department`, `Ministry`, `Council`, `Committee`, `Board`, `Commission`, `Association`, `Foundation`, `Corp`, `Inc`) to lift `ORG` recall (`0.24 → ≥ 0.55`).

---

## 6. Compression, Packaging & Memory Targets

* Save freshly trained raw weights to `ext-ref/parser/` as usual, then pack into `civ/sys/parser/models/*.safetensors.gz` via `civ/bin/parser pack` (IEEE-754 `F16` + `gzip -9`).
* If Stage 2 + Stage 3 are unified into one model file (e.g. `dep_parser.safetensors.gz` carrying both POS and Dep, or `parser.safetensors.gz`), keep `ResolveModelPath`, `ModelsPresent`, `ModelBytes`, `pack`, and `TestLoadUntrained` / `TestParserBackendPriority` working seamlessly! (Keeping `pos_tagger.safetensors.gz` optional or storing the unified model cleanly is fine; document the exact filenames in the handoff.)
* Verify **0.00–0.05%** accuracy delta between raw F32 (`ext-ref/parser/`) and packed F16+gzip (`civ/sys/parser/models/`).
* Verify `civ/bin/parser eval` runs cleanly **both** with `ext-ref/parser/*.safetensors` present AND with them moved aside (loading strictly from `civ/sys/parser/models/*.safetensors.gz`).

---

## 7. Required Tests & Round-2 Acceptance Gates

### Automated Tests (Both `go test` AND `GOEXPERIMENT=simd go test`, plus `-race`)
1. `TestBiGRUBiaffineGradCheck` — finite-difference vs analytic gradient parity across every parameter tensor.
2. `TestOverfitMiniBatch` — 16 real sentences overfit to `< 0.10` loss and `> 98%` UAS/LAS.
3. `TestMSTSingleRoot` — brute-force optimality check on random score matrices (`N = 1..6`) + single-root acyclic invariant on `N = 1..80`.
4. `TestSealedDraftsHoldout` — verifies no draft overlapping `ch13`/`ch14` can enter training.
5. `TestSIMDMatMulParity` — `MatMulT` / `OuterAdd` SIMD vs scalar parity for ragged dimensions `1..100` (with `ClearAVXUpperBits` verified!).
6. All existing Round-1 tests in `civ/sys/parser`, `civ/sys/vec`, `civ/sys/linear`, `civ/compose`, and `civ/linear` continue to pass (`go test -count=1 ./sys/... ./parser ./compose ./linear`, `GOEXPERIMENT=simd go test -count=1 ./sys/... ./parser ./compose ./linear`, and `GOEXPERIMENT=simd go test -race ./sys/parser ./sys/vec`).

### Official Benchmark & Accuracy Gates (`civ/bin/parser eval` from Packed `.safetensors.gz`)

Run `civ/bin/parser eval` on `ch13,ch14` **once** after training and coref tuning are frozen on `dev_ud` + `dev_book` (`ch12`), and include the exact table in `civ/notes/48a_go_parser_r2_handoff.md`:

| Metric | Round 1 | Round-2 Gate |
| --- | --- | --- |
| UD-EWT test POS accuracy | 94.92% | **≥ 96.5%** |
| UD-EWT test UAS (predicted POS) | 83.49% | **≥ 90.0%** |
| UD-EWT test LAS (predicted POS) | 80.27% | **≥ 87.5%** |
| Held-out `ch13+ch14` POS agreement vs CoreNLP | 97.19% | **≥ 97.5%** |
| Held-out `ch13+ch14` non-punct UAS vs CoreNLP *(also report all-token)* | ~81% | **≥ 91.0%** |
| Held-out `ch13+ch14` non-punct LAS vs CoreNLP *(also report all-token)* | ~77% | **≥ 88.5%** |
| Held-out `ch13+ch14` Root agreement vs CoreNLP | 79.7% | **≥ 92.0%** |
| Held-out `ch13+ch14` `civ/linear` Rule 10 correlation ($r$) | 0.6898 | **≥ 0.92** |
| Held-out `ch13+ch14` `civ/linear` Part 4 findings F1 | 0.462 | **≥ 0.85** |
| Mean end-to-end latency (`ParseSentence`) | 0.620 ms | **< 1.5 ms** *(stretch < 1.0 ms)* |
| Full book (`ch01..ch14`) parallel parse time | 0.27 s | **< 1.5 s** |
| Cold load time (from `.safetensors.gz`) | 104 ms | **< 150 ms** |
| Live heap after load | 36.6 MB | **< 45 MB** |
| Total packed `.safetensors.gz` size on disk | 7.34 MB | **≤ 15 MB** |

*(When all gates pass, do NOT flip `civ/linear`'s default `-engine` yet—leave that final switch for the review & integration pass.)*

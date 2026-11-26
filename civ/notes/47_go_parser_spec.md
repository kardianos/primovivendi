# 47. Spec: Pure-Go NLP Parser Replacing Stanford CoreNLP (`civ/sys/parser`)

Recorded 2026-11-25. Implementation specification for a pure-Go, zero-CGO, zero-Java replacement for Stanford CoreNLP 4.6.0 serving `civ/compose` and `civ/linear`. Hand this file to an implementation agent. When complete, a review agent will audit against every invariant, metric gate, and benchmark defined below.

---

## 1. Goal & Scope

Replace Stanford CoreNLP 4.6.0 (`java -mx16g`, ~4–16 GB heap, multi-second startup / multi-minute book batch runs) with a fast, self-contained pure-Go NLP pipeline under `civ/sys/parser` and a companion training/evaluation CLI under `civ/parser` (`civ/bin/parser`).

The pipeline must produce the **exact data structures already consumed by `civ/compose` and `civ/linear`** (`*corenlp.Tree`, `*corenlp.Doc`, `corenlp.Sent`, `corenlp.Tok`, `corenlp.Dep`, `corenlp.Mention`), trained in pure Go (in the spirit of `kardianos/inkline`), accelerated with `civ/sys/vec` (`GOEXPERIMENT=simd` AVX2+FMA with scalar fallback), and persisted in standard `.safetensors` / compact binary files under `ext-ref/parser/`.

### Non-Goals for v1
* **No constituency S-expression parser (`Sent.Parse`)**: Neither `civ/compose` nor `civ/linear` requires Penn Treebank constituency strings at runtime—both operate strictly on dependency trees (`basicDependencies`), POS tags, lemmas, NER, and (for `civ/linear` Part 4) coreference chains. Leave `Sent.Parse` empty.
* **No CGO, Python, PyTorch, or ONNX runtime**: Both training and inference must run in pure Go inside the unified `civ/go.mod` (`module po/civ`).
* **Do not delete CoreNLP support yet**: Keep cached CoreNLP JSON loading and `-corenlp-server` / `-engine corenlp` available as an opt-in override until all evaluation gates pass on held-out chapters (`ch13`, `ch14`).

---

## 2. Guardrails & Repository Conventions

* **Workspace**: All code lives under `civ/` using the existing unified [civ/go.mod](file:///home/d/code/po/civ/go.mod) (`module po/civ`). Do **not** add a nested `go.mod`.
* **New packages**:
  * `civ/sys/parser/` — reusable library package (`package parser`) imported by `civ/compose`, `civ/linear`, and `civ/parser`.
  * `civ/parser/` — CLI binary (`package main`), added to [civ/tools/build-bin.fish](file:///home/d/code/po/civ/tools/build-bin.fish) (`civ/bin/parser`, built with `GOEXPERIMENT=simd`).
* **Model & corpus artifacts (uncommitted)**:
  * Store downloaded treebanks and trained weights in `ext-ref/parser/` (parallel to `ext-ref/corenlp/`, `ext-ref/embed/`, `ext-ref/lexicon/`). Never commit weights or corpora to git.
  * Reuse existing packages:
    * [civ/sys/corenlp](file:///home/d/code/po/civ/sys/corenlp/corenlp.go) — target output types (`Doc`, `Sent`, `Tok`, `Dep`, `Mention`, `Tree`, `BuildTree`).
    * [civ/sys/vec](file:///home/d/code/po/civ/sys/vec/vec_simd.go) — `Dot`, `Axpy`, `WordFilter`, `HashWord` (AVX2+FMA under `GOEXPERIMENT=simd`, scalar fallback without it).
    * [civ/sys/lexicon](file:///home/d/code/po/civ/sys/lexicon/lexicon.go) — WordNet 3.1 `.exc` files (`ext-ref/lexicon/wordnet31/*.exc`) for lemmatization.
    * [civ/sys/prof](file:///home/d/code/po/civ/sys/prof/prof.go) — instrument every stage (`parser.tok`, `parser.lemma`, `parser.pos`, `parser.dep`, `parser.ner`, `parser.coref`).
* **Build parity**: Both `go test ./sys/... ./parser ./compose ./linear` (without `GOEXPERIMENT=simd`) AND `GOEXPERIMENT=simd go test ./sys/... ./parser ./compose ./linear` MUST pass cleanly.

---

## 3. Exact Consumer Contract (`corenlp.Doc` / `corenlp.Tree`)

Downstream code in `civ/compose` and `civ/linear` depends on the following invariants. Violating any of these is a blocking review failure:

1. **1-Based Token Indexing & Native UTF-8 Byte Offsets**:
   * `Sent.Toks` is 0-indexed in JSON, with `Tok.Index` 1-based (`1..N`).
   * `Tok.Begin` and `Tok.End` MUST be exact **UTF-8 byte offsets** into the input document string (`text[t.Begin:t.End] == t.Orig`, O(1) zero-copy slicing matching `civ/review` and Go's stdlib).
   * **Legacy CoreNLP JSON boundary only**: `corenlp.LoadCached` and `corenlp.Server` already normalize Java/CoreNLP's UTF-16 `characterOffsetBegin`/`characterOffsetEnd` into UTF-8 byte offsets once on load via `corenlp.NormalizeUTF16ToUTF8`. Downstream code (`civ/sys/parser`, `civ/sys/corenlp`, `civ/compose`, and updated `civ/linear`) MUST NOT use `unicode/utf16` or allocate `[]uint16` / `[]rune` document copies.
   * `Tok.Orig` holds the exact original surface slice; `Tok.Word` holds the Penn Treebank normalized form (`(` → `-LRB-`, `)` → `-RRB-`, opening quote → `` `` ``, closing quote → `''`, else equal to `Orig`).
   * `Dep.DepGloss` MUST equal `Tok.Word` for `Dep.Dependent` (checked by `corenlp.BuildTree`).
2. **Valid Acyclic Single-Root Dependency Tree**:
   * For every sentence of length $N \ge 1$, `Sent.Deps` contains exactly $N$ edges (one per dependent $1..N$), exactly **one** edge with `Governor == 0` and `Dep == "ROOT"`, and zero cycles.
   * Passing any `Sent` to `corenlp.BuildTree(s, fullText)` MUST return `ok == true`.
3. **Canonical POS Tag Set (Penn Treebank XPOS — 45 tags)**:
   * Must emit standard PTB tags as seen in `ext-ref/linear/out/ch*.txt.json`:
     `CC`, `CD`, `DT`, `EX`, `FW`, `IN`, `JJ`, `JJR`, `JJS`, `LS`, `MD`, `NN`, `NNS`, `NNP`, `NNPS`, `PDT`, `POS`, `PRP`, `PRP$`, `RB`, `RBR`, `RBS`, `RP`, `SYM`, `TO`, `UH`, `VB`, `VBD`, `VBG`, `VBN`, `VBP`, `VBZ`, `WDT`, `WP`, `WP$`, `WRB`, `` `` ``, `''`, `,`, `.`, `:`, `-LRB-`, `-RRB-`, `HYPH`, `$`, `#`.
   * Particularly critical for `civ/sys/lexicon` & `civ/sys/linear`:
     * Exact verb inflection distinction (`VB`, `VBD`, `VBG`, `VBN`, `VBP`, `VBZ`, `MD`) — drives English inflection of WordNet alternatives.
     * Exact noun number/properness (`NN`, `NNS`, `NNP`, `NNPS`) — drives pluralization and `personLike` / `lowerFirst`.
     * Exact adjective/adverb degree (`JJ`, `JJR`, `JJS`, `RB`, `RBR`, `RBS`).
4. **Canonical Dependency Label Set (UD v2 / CoreNLP 4.6.0 — 38 labels)**:
   * CoreNLP 4.6.0 emits UD v2 labels (verified directly on `ch01.txt.json`). Normalize all training labels (from both UD-EWT and CoreNLP JSON) into this exact fixed label inventory:
     * **Core clausal / subject**: `ROOT`, `nsubj`, `nsubj:pass`, `csubj`, `csubj:pass`, `expl`
     * **Complements & obliques**: `obj`, `iobj`, `obl`, `obl:npmod`, `obl:tmod`, `ccomp`, `xcomp`
     * **Clausal modifiers**: `advcl`, `acl`, `acl:relcl`, `parataxis`
     * **Nominal & lexical modifiers**: `nmod`, `nmod:poss`, `amod`, `advmod`, `appos`, `nummod`, `compound`, `compound:prt`, `flat`, `fixed`
     * **Function words & coordination**: `det`, `det:predet`, `case`, `mark`, `aux`, `aux:pass`, `cop`, `cc`, `cc:preconj`, `conj`
     * **Misc**: `punct`, `discourse`, `list`, `orphan`, `vocative`, `dep`
   * Mapping rules when reading CoNLL-U (`UD_English-EWT`):
     * `root` → `ROOT`
     * Strip unlisted subtypes to their base relation (e.g., `nmod:tmod` → `nmod`, `obl:agent` → `obl`, `flat:foreign` → `flat`, `ccomp:outer` → `ccomp`) while preserving the listed subtypes (`nsubj:pass`, `csubj:pass`, `aux:pass`, `acl:relcl`, `nmod:poss`, `obl:npmod`, `obl:tmod`, `compound:prt`, `cc:preconj`, `det:predet`).
5. **NER & Coref Contract**:
   * `Tok.NER`: `"PERSON"`, `"ORG"`, `"LOC"`, `"MISC"`, `"NUMBER"`, `"DATE"`, `"O"`. (`PERSON` and non-`O` proper names are used by `civ/linear` and `civ/sys/linear`.)
   * `Doc.Corefs`: `map[string][]corenlp.Mention`, 1-based `SentNum`, 1-based half-open `[Start, End)`, 1-based `Head`, `Type` in `{"PROPER", "NOMINAL", "PRONOMINAL"}`.

---

## 4. Training & Evaluation Data Pipeline (Strict Holdout Discipline)

Zero data leakage is mandatory. Separate training, dev, and held-out test sets strictly:

### A. Gold Treebank: Universal Dependencies English-EWT (CC BY-SA 4.0)
* Subcommand `civ/bin/parser fetch-ud` downloads `UD_English-EWT` (pinned release tag, e.g., `r2.15` from `https://raw.githubusercontent.com/UniversalDependencies/UD_English-EWT/r2.15/`) into `ext-ref/parser/ud_ewt/`:
  * `en_ewt-ud-train.conllu` (~12,500 gold sentences) — training
  * `en_ewt-ud-dev.conllu` (~2,000 gold sentences) — early stopping & hyperparameter tuning
  * `en_ewt-ud-test.conllu` (~2,000 gold sentences) — untouched gold evaluation
* CoNLL-U parser (`civ/sys/parser/conllu.go`): skip multi-word token range lines (`1-2`) and empty nodes (`1.1`); read `FORM` (col 2), `LEMMA` (col 3), `XPOS` (col 5), `HEAD` (col 7), `DEPREL` (col 8).

### B. Period-Prose Silver Distillation Set (Already on Disk!)
* Available in `ext-ref/linear/out/`:
  * **Training distillation (optional `-distill` flag, recommended default on)**:
    * All 8 reference treatises: `pg4351`, `pg4367`, `pg6435`, `pg935`, `pg13529`, `pg26334`, `pg34901`, `pg35916` (~25k+ sentences of 1850–1950 expository prose!).
    * Chapters `ch01` through `ch12` (~3,500 sentences) — *optional* second stage or 0.5× sample weight so UD-EWT gold labels stay authoritative while treatise punctuation/subordination is well represented.
  * **STRICTLY HELD-OUT BOOK TEST SET (NEVER TRAIN OR TUNE ON THESE)**:
    * **`ch13.txt.json`** and **`ch14.txt.json`** — reserved exclusively for `civ/bin/parser eval` (POS agreement, UAS/LAS vs CoreNLP, NER F1, and `civ/linear` Rule 10 finding parity).
* Also filter out malformed CoreNLP sentences during ingestion (using `corenlp.BuildTree`'s gloss/index validity check).

---

## 5. Architecture & Staged Implementation Plan

Build and verify **Stage 1 → Stage 2 → Stage 3 → Stage 4** in strict order. Each stage must compile, pass its unit tests, and meet its accuracy/latency target before moving to the next.

### Package Layout

```text
civ/
  sys/
    parser/
      doc.go          # public API: Load(root), ParseSentence(text), ParseDoc(text)
      tok.go          # Stage 1: PTB tokenizer, sentence splitter, UTF-8 byte offset tracking
      lemma.go        # Stage 1: WordNet morphy + irregulars lemmatizer
      pos.go          # Stage 2: averaged perceptron POS tagger (infer + train + save/load)
      oracle.go       # Stage 3: Arc-Standard transition system + static/dynamic oracle + legal masks
      dep.go          # Stage 3: neural transition dependency parser (precomputed W1 + AVX2+FMA)
      dep_train.go    # Stage 3: pure-Go multi-core mini-batch AdamW trainer + .safetensors export
      ner.go          # Stage 4: averaged perceptron BIO NER tagger
      coref.go        # Stage 4: mention extractor + salience/agreement coreference resolver
      conllu.go       # CoNLL-U + CoreNLP JSON corpus reader + label normalizer
      eval.go         # POS acc, UAS/LAS, speed, Rule-10 parity metrics
      parser_test.go  # unit tests, invariants, UTF-8 byte-offset tests, benchmarks
  parser/
    main.go           # CLI: fetch-ud | train-pos | train-dep | train-ner | train-all | eval | parse | bench
```

---

### Stage 1: Tokenizer, Sentence Splitter & Lemmatizer (`tok.go`, `lemma.go`)

#### Tokenizer & Splitter (`tok.go`)
* **Zero-regex linear pass** over UTF-8 input using `utf8.DecodeRuneInString`, recording exact **UTF-8 byte offsets** (`Tok.Begin`, `Tok.End`) with zero `[]rune` or `[]uint16` document allocations.
* Sentence splitting:
  * Split on paragraph breaks (`\n\n`) and terminal `.` / `!` / `?` (preserving closing quotes/parens attached to the sentence, protecting abbreviations `Mr.`, `Mrs.`, `Dr.`, `St.`, `U.S.`, `e.g.`, `i.e.`, `etc.`, `p.`, `pp.`, `ch.`, `vol.`, `a.m.`, `p.m.`, single initials `J.`, and decimal numbers `10.5`).
* Penn-Treebank tokenization rules (matching CoreNLP):
  * Separate leading/trailing punctuation (`(`, `)`, `[`, `]`, `,`, `;`, `:`, `.`, `!`, `?`, `—`, `–`, `"`, `“`, `”`).
  * Split contractions cleanly:
    * `n't` (`don't` → `do` + `n't`, `cannot` → `can` + `not`, `won't` → `wo` + `n't`, `can't` → `ca` + `n't`)
    * `'s`, `'re`, `'ve`, `'ll`, `'d`, `'m` (`man's` → `man` + `'s`, `we're` → `we` + `'re`).
  * Preserve hyphenated compounds or split hyphens consistently with `HYPH` (match CoreNLP 4.6.0: keep intra-word hyphens like `self-interest` as `self` / `-` (`HYPH`) / `interest` OR document exact tokenization alignment in `eval.go` so token-aligned metrics are exact). **Recommendation**: split intra-word hyphens into `word`, `-` (`HYPH`), `word` because CoreNLP 4.6.0 does this (verified in `ch01.txt.json`: `HYPH` is in its POS inventory!).
  * Map `Word` for brackets/quotes (`-LRB-`, `-RRB-`, `` `` ``, `''`) while keeping `Orig` verbatim.

#### Lemmatizer (`lemma.go`)
* Pure-Go implementation of Princeton WordNet's `morphy` algorithm:
  1. Load `ext-ref/lexicon/wordnet31/{noun,verb,adj,adv}.exc` and lemma vocabulary sets from `index.{noun,verb,adj,adv}` (with built-in irregular fallback tables from `civ/sys/lexicon` if WordNet 3.1 isn't unpacked).
  2. Direct irregular lookup first (`was`/`were`/`is`/`are` → `be`, `men` → `man`, `struck` → `strike`, `better` → `good`, `n't` → `not`).
  3. POS-conditioned morphological detachment rules (e.g. `NNS`: `-ies`→`-y`, `-ves`→`-f`/`-fe`, `-xes`/`-ches`/`-shes`/`-ses`→strip `-es`, `-s`→strip `-s`; `VB*`: `-ied`→`-y`, `-ing`→strip + check double consonant / restore `-e`, `-ed`→strip + check double consonant / restore `-e`, `-es`/`-s`; `JJR`/`JJS`: `-ier`/`-iest`→`-y`, `-er`/`-est`), validating stems against the WordNet lemma table before accepting.
  4. Preserve lowercase lemmas for common words; keep proper nouns (`NNP`/`NNPS`) in original casing or lowercase consistently with CoreNLP (`Tok.Lemma` preserves case on `NNP`, lowercases common words).

---

### Stage 2: Averaged Perceptron POS Tagger (`pos.go`)

* **Model**: Linear multi-class **Averaged Perceptron** (Collins 2002 / spaCy `Tagger`), greedy left-to-right decoding over the 45 PTB tags.
* **Feature template per token `i`** (18–22 sparse hashed features — fast, zero regex):
  * Bias
  * Word lower `w[i]`, `w[i-1]`, `w[i-2]`, `w[i+1]`, `w[i+2]`
  * Prefixes/suffixes of `w[i]`: `pref1`, `suf1`, `suf2`, `suf3`, `suf4`
  * Suffixes of neighbors: `suf3(w[i-1])`, `suf3(w[i+1])`
  * Word shape of `w[i]`: capitalized (`i==0` vs `i>0`), all-caps, has-digit, all-digits, has-hyphen, is-punct
  * Previous predicted tags: `t[i-1]`, `t[i-2]`, bigram `(t[i-2], t[i-1])`
  * Tag-word conjunctions: `(t[i-1], w[i])`, `(t[i-1], suf3(w[i]))`
  * Optional lexical prior feature: WordNet / Kuperman POS hint if available (or pure corpus-learned tag dictionary for unambiguous words!).
* **Fast Inference & Compact Storage**:
  * **Unambiguous word tag dictionary**: words seen ≥10 times in training with ≥97% single POS tag are tagged in O(1) directly (skips feature scoring for ~40% of tokens!).
  * **Hashed dense weight matrix** (or pruned sparse table): Using a fixed feature-hash table of $2^{18} = 262,144$ buckets × 45 classes (or pruned top-K active features stored in a compact binary / `.safetensors` file — ~2–4 MB!) allows `Score` to simply `vec.Axpy` 20 rows of 45 floats using **AVX2+FMA** in **<5 microseconds per token**!
  * Weight averaging during training: maintain `weights`, `totals`, `timestamps` (lazy averaging update), run 8–12 shuffled epochs over `UD-EWT train` (+ optional distilled treatises), evaluate on `UD-EWT dev` after each epoch, and save the best checkpoint to `ext-ref/parser/pos_tagger.safetensors` (plus small JSON/binary vocab header).
* **Acceptance Gate (Stage 2)**:
  * UD-EWT test POS accuracy: **≥ 96.5%**
  * Held-out `ch13`/`ch14` POS agreement with CoreNLP: **≥ 97.0%**
  * Inference latency: **< 0.15 ms / sentence** (mean), zero heap churn on hot loop.

---

### Stage 3: Pure-Go Neural Transition-Based Dependency Parser (`oracle.go`, `dep.go`, `dep_train.go`)

This is the centerpiece of `civ/sys/parser`.

#### Transition System: Arc-Standard (`oracle.go`)
* Configuration `(Stack, Buffer, Arcs)`:
  * Initial state: `Stack = [0]` (synthetic `ROOT`), `Buffer = [1, 2, ..., N]`.
  * Terminal state: `len(Buffer) == 0 && len(Stack) == 1` (only `0` remains on stack).
  * Transitions (`1 + 2*|R|` total actions, where `|R| = 38` non-ROOT labels → **77 actions**, or include `ROOT` in `RIGHT-ARC(ROOT)` when `s1 == 0`):
    1. `SHIFT`: push `Buffer[0]` onto `Stack`. Legal iff `len(Buffer) > 0`.
    2. `LEFT-ARC(r)`: add arc `(head = s0, dep = s1, rel = r)`, pop `s1`. Legal iff `len(Stack) >= 2 && s1 != 0` (cannot make `ROOT` a dependent!).
    3. `RIGHT-ARC(r)`: add arc `(head = s1, dep = s0, rel = r)`, pop `s0`. Legal iff `len(Stack) >= 2`, and if `s1 == 0` then `r` MUST be `"ROOT"` and `len(Buffer) == 0` (guarantees exactly one root!).
* **Guaranteed tree validity**: Every sentence terminates in **strictly $2N$ transitions** ($O(N)$ time) and is mathematically guaranteed to produce a valid, acyclic, single-root spanning tree!
* **Projectivity & Oracle**:
  * Implement the standard **static oracle** first (exact, deterministic, trivial to verify on gold projective trees; ~98.5% of English UD-EWT trees are projective—skip or projectivize the ~1.5% non-projective training trees cleanly).
  * Optional bonus (if time permits after passing gates): **dynamic oracle** (Goldberg & Nivre 2012) or 10% scheduled sampling during training for +0.5–1.0% LAS on real-world errors.

#### Neural Network Architecture (`dep.go`) — Chen & Manning (2014) + SIMD Precomputation
* **48 Fixed Feature Slots** extracted from `(Stack, Buffer, Arcs)` in O(1):
  * **18 word IDs**: `s0, s1, s2` (top 3 stack), `b0, b1, b2` (front 3 buffer), `lc1(s0), rc1(s0), lc2(s0), rc2(s0), lc1(s1), rc1(s1), lc2(s1), rc2(s1)` (leftmost/rightmost 1st & 2nd children), `lc1(lc1(s0)), rc1(rc1(s0)), lc1(lc1(s1)), rc1(rc1(s1))` (grandchildren).
  * **18 POS tag IDs**: for the exact same 18 token positions (using `NULL` ID when position doesn't exist).
  * **12 dependency label IDs**: for the 12 child/grandchild positions above.
* **Embeddings & Layers**:
  * Word vocabulary: top `16,384` lowercased words (plus `<UNK>`, `<NULL>`, `<ROOT>`), dim `dw = 48` (or `64`).
  * POS vocabulary: 48 tags (`<NULL>`, `<ROOT>` + 45 PTB + `<UNK>`), dim `dp = 16`.
  * Dep-label vocabulary: 46 labels (`<NULL>` + canonical labels), dim `dl = 16`.
  * Input width: $D_{\text{in}} = 18 \cdot 48 + 18 \cdot 16 + 12 \cdot 16 = 864 + 288 + 192 = 1344$.
  * Hidden layer 1: `1344 -> H` (`H = 256` or `320`) with **SiLU** (or **Cube** $x^3 + x$ / ReLU) activation + optional LayerNorm/dropout during training.
  * Output layer: `H -> numActions` (~80 logits), masked to legal transitions (`-1e9` on illegal actions) before `argmax` / softmax.
* **The Critical Speed Optimization (Precomputed First-Layer Projections + `civ/sys/vec` AVX2+FMA)**:
  * At inference time, never materialize the 1344-dim input vector or multiply `1344 × 256` from scratch!
  * Because `W1` is partitioned into 48 slot blocks (`W1_slot[s]` of shape `[H × d_s]`), we simply:
    * **Precompute** `P_pos[slot][posID]` (`[H]` floats) for ALL 18 POS slots × 48 tags (~220 KB) and `P_dep[slot][depID]` for ALL 12 dep slots × 46 labels (~140 KB) AND `P_word[slot][wordID]` for the top `2,048` most frequent words (covers >85% of running English text!).
    * On each transition step, accumulate `h[0..H]` by `h = b1` + 48 calls to `vec.Axpy(h, precompRow, 1.0)` (for cached words/POS/deps) or a tiny `48`-float dot for rare words — pure AVX2+FMA!
    * Then `activation(h)` + `vec.Dot(W2[a], h)` ONLY for legal actions `a`!
  * Result: **~200–600 microseconds per 20-word sentence** in pure Go on CPU!

#### Pure-Go Training (`dep_train.go`)
* Built in pure Go (no Python/PyTorch), reusing `civ/sys/vec` (`Dot`, `Axpy`) and `.safetensors` format:
  * **Optimizer**: AdamW (`lr = 2e-3` with cosine decay, `wd = 1e-4`, `batchSize = 256` transitions, gradient clipping `norm <= 5.0`, dropout `0.2` on hidden layer during training).
  * **Data-parallel multi-core training**: Split each mini-batch across `min(runtime.NumCPU(), 8)` goroutines, each accumulating into its own local gradient buffer, then reduce and apply one AdamW step—trains 20 epochs over UD-EWT (~250k transitions/epoch) in **~2–5 minutes** on this 16-thread Ryzen 7 5700G!
  * **Word dropout**: With probability `0.15` on rare words (`count <= 3`), replace word ID with `<UNK>` during training so the `<UNK>` embedding learns clean generalization for unseen book vocabulary.
  * **Two-phase training option (`-distill`)**:
    * Phase A: Train on `UD-EWT train` + `pg*.txt.json` (weighted 0.5× or mixed 1:1) for 15–20 epochs.
    * Phase B: Fine-tune 5 epochs at lower LR (`3e-4`) on `UD-EWT train` + `ch01..ch12.txt.json` (NEVER `ch13`/`ch14`!), keeping the checkpoint with highest `UD-EWT dev` LAS!
  * Export model + vocab metadata to `ext-ref/parser/dep_parser.safetensors` (total weight size ~10–14 MB).

* **Acceptance Gate (Stage 3)**:
  * UD-EWT test: **UAS ≥ 90.0%**, **LAS ≥ 87.5%** (using predicted POS tags from Stage 2, end-to-end!).
  * Held-out `ch13`/`ch14` vs CoreNLP 4.6.0: **UAS ≥ 91.0%**, **LAS ≥ 88.5%**.
  * End-to-end latency (`tok + lemma + pos + dep`): **< 1.0 ms/sentence** mean, **< 2.5 ms/sentence** p95.
  * Cold load time of POS + Dep `.safetensors`: **< 150 ms**, resident RAM **< 45 MB**.

---

### Stage 4: Lightweight NER & Mention-Pair Coreference (`ner.go`, `coref.go`)

Only required to achieve full `civ/linear` parity (Stages 1–3 already cover 100% of `civ/compose`!).

#### NER (`ner.go`)
* BIO averaged perceptron (same compact engine as Stage 2!) predicting `{O, B-PERSON, I-PERSON, B-ORG, I-ORG, B-LOC, I-LOC, B-MISC, I-MISC}` + deterministic regex/rule tags for `NUMBER`, `ORDINAL`, `DATE`.
* Trained on silver NER from `ext-ref/linear/out/pg*.txt.json` + `ch01..ch12.txt.json`, plus `personNouns` / title gazetteer (`Mr.`, `Dr.`, `King`, `Judge`, `Saint`, `President`, etc.).
* Saved to `ext-ref/parser/ner.safetensors` (< 3 MB).

#### Coreference (`coref.go`)
* Mention extraction from Stage 3 dependency trees:
  * Proper noun spans (`NNP`/`NNPS` + `flat`/`compound`), nominal NPs (`NN`/`NNS` with subtree span `[lo..head]`), and personal/possessive pronouns (`PRP`/`PRP$`, excluding dummy/pleonastic `"it"` using the same `dummyIt`/`referentialIt` rules in `civ/linear/style.go`).
* Deterministic multi-pass sieve + lightweight pairwise ranker over candidate antecedents within a 5-sentence lookback window:
  * Pass 1: Exact string / head-lemma match for proper and nominal mentions.
  * Pass 2: Pronoun-to-antecedent resolution using strict **person/number/gender/animacy compatibility** (`he/him/his` → singular person, `she/her` → singular person, `they/them/their` → plural noun/person, `it/its` → singular inanimate noun) ranked by **syntactic salience** (Hobbs distance: same sentence earlier clause > previous sentence subject > previous sentence object/oblique > 2–4 sentences back subject).
  * Outputs `map[string][]corenlp.Mention` matching CoreNLP's JSON schema.
* **Acceptance Gate (Stage 4)**:
  * `civ/linear` Part 4 (`distant pronoun`, `it chain`, `ambiguous it`) finding precision/recall parity vs CoreNLP on held-out `ch13`/`ch14`: **F1 ≥ 0.85**.

---

## 6. Integration Plan (`civ/compose`, `civ/linear`, `civ/parser`)

### A. Public Library API (`civ/sys/parser`)
```go
package parser

type Pipeline struct { /* POS + Dep + Lemma + optional NER/Coref */ }

// Load reads ext-ref/parser/{pos_tagger,dep_parser,ner}.safetensors.
// Returns (nil, ErrNotTrained) cleanly if model files have not been built yet.
func Load(root string) (*Pipeline, error)

// ParseSentence runs Stages 1-3 (plus optional NER) on one sentence and
// returns a valid 1-based *corenlp.Tree in <1ms. Safe for concurrent calls!
func (p *Pipeline) ParseSentence(text string) *corenlp.Tree

// ParseDoc splits paragraphs/sentences and runs Stages 1-4 on an entire chapter
// or Markdown-stripped text, returning a CoreNLP-compatible *corenlp.Doc.
func (p *Pipeline) ParseDoc(text string) *corenlp.Doc
```
* **Concurrency safety**: `Pipeline` MUST be 100% read-only / immutable after `Load` (allocate small scratch buffers on stack or `sync.Pool` per call) so `civ/compose` and `civ/linear` can parse chapters in parallel across all CPU cores with zero lock contention.

### B. CLI (`civ/parser` → `civ/bin/parser`)
```text
civ/bin/parser fetch-ud                             # download UD_English-EWT into ext-ref/parser/ud_ewt/
civ/bin/parser train-pos [-distill]                 # train & save pos_tagger.safetensors
civ/bin/parser train-dep [-distill] [-epochs 20]    # train & save dep_parser.safetensors
civ/bin/parser train-ner                            # train & save ner.safetensors
civ/bin/parser train-all [-distill]                 # run full training pipeline
civ/bin/parser eval [-chapter ch13,ch14]            # evaluate UD-EWT test + held-out ch13/ch14 + Rule 10 parity
civ/bin/parser parse "We come into this world naked." # print dependency tree & latency
civ/bin/parser bench [-chapter ch13]                # benchmark sentences/sec & memory
```
* Add `parser` to the build loop in [civ/tools/build-bin.fish](file:///home/d/code/po/civ/tools/build-bin.fish).

### C. Wiring into `civ/compose`
* In `civ/compose/engine.go`:
  * `NewEngine` attempts `parser.Load(root)` (non-fatal if weights aren't trained yet).
  * `Engine.ParseSentence(text)` priority order:
    1. In-memory tree cache (`cachedTrees[norm]`)
    2. Local pure-Go `parser.Pipeline` (if loaded — **<1ms**, replaces Java CoreNLP server for all manual rewrites, custom phrases, and uncached sentences!)
    3. Optional CoreNLP HTTP server (only if `-corenlp-server` flag explicitly passed)
    4. `corenlp.ParseFallback(text)`
  * Show parser status badge in the `compose` GUI status bar (`Parser: Go (0.4ms)` vs `Parser: fallback`).

### D. Wiring into `civ/linear`
* Add `-engine corenlp|go` flag to `civ/linear` (`parse`, `report`, `lines`, `check`).
* Migrate `civ/linear/analyze.go` (`parsedDoc.sub`, `lineAt`, `paraIndex`) from `[]uint16` / `unicode/utf16` to **native UTF-8 byte offsets** (`text[a:b]`), applying `corenlp.NormalizeUTF16ToUTF8` once when reading legacy CoreNLP JSON files so both `-engine go` and `-engine corenlp` share the exact same UTF-8 byte-offset codepath.
* When `-engine go` is passed, bypass Java completely and parse chapters directly in memory via `civ/sys/parser` (parsing all 14 chapters in <2 seconds!).
* Keep `-engine corenlp` as default until the review confirms Rule 10 finding parity on `ch13`/`ch14`.

---

## 7. Required Tests & Review Checklist

The reviewing agent will verify every item below before approving the implementation:

### Automated Tests (`go test` AND `GOEXPERIMENT=simd go test`)
1. **UTF-8 Byte-Offset Fidelity**: Test sentences containing em-dashes (`—`), smart quotes (`“`, `”`, `’`), and 4-byte astral runes (`𐐷`); verify `text[t.Begin:t.End] == t.Orig` for every token with zero `unicode/utf16` usage outside `corenlp.NormalizeUTF16ToUTF8`.
2. **Tree Validity Invariants**: Fuzz/table-test 200+ diverse sentences (1-word fragments, 120-word sentences, heavy punctuation, quotes, parentheses, empty strings, whitespace-only, all-punctuation). Assert zero panics, single `ROOT`, zero cycles, `corenlp.BuildTree` returns `ok == true`.
3. **Oracle Round-Trip Test**: On 500 projective sentences from UD-EWT, verify running the Arc-Standard oracle forward reproduces 100.0% of gold arcs and labels.
4. **SIMD vs. Scalar Numerical Parity**: Verify `dep.go` logits and parsed trees match identically with and without `GOEXPERIMENT=simd` (within `1e-4` float tolerance).
5. **`.safetensors` Round-Trip**: Save and reload POS, Dep, and NER models; verify bit-exact predictions.
6. **Concurrency Race Test**: `go test -race ./sys/parser` running 8 goroutines calling `ParseSentence` concurrently on the same `*Pipeline`.
7. **Zero Data Leakage Audit**: Verify `ch13` and `ch14` are explicitly excluded from all training loaders (unit test the file filter!).

### Benchmark & Accuracy Report (Output of `civ/bin/parser eval`)
The implementation agent must run `civ/bin/parser eval` after training and include the exact output table in its handoff note:

| Metric | Gate |
| --- | --- |
| UD-EWT test POS accuracy | ≥ 96.5% |
| UD-EWT test UAS (predicted POS) | ≥ 90.0% |
| UD-EWT test LAS (predicted POS) | ≥ 87.5% |
| Held-out `ch13` + `ch14` POS agreement vs CoreNLP | ≥ 97.0% |
| Held-out `ch13` + `ch14` UAS vs CoreNLP | ≥ 91.0% |
| Held-out `ch13` + `ch14` LAS vs CoreNLP | ≥ 88.5% |
| Held-out `ch13` + `ch14` `civ/linear` Rule 10 score correlation ($r$) | ≥ 0.92 |
| Mean end-to-end latency (`ParseSentence`) | < 1.0 ms |
| Full book (`ch01..ch14`) parallel parse time | < 2.5 s |
| Cold model load time | < 150 ms |
| Total `.safetensors` size on disk | ≤ 20 MB |

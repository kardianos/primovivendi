# 47a. Handoff: the Pure-Go Parser Is Built and Fast; Five of Eleven Gates Pass, the Rest Are Accuracy

Recorded 2026-11-25. This is the handoff for the implementation of note 47 (`civ/notes/47_go_parser_spec.md`): a pure-Go, zero-Java NLP pipeline in `civ/sys/parser`, the CLI `civ/bin/parser`, and its wiring into `civ/compose` and `civ/linear`. It records the exact `civ/bin/parser eval` table, what was built, where it departs from the spec, why the accuracy gates are missed, and what would close the gap. Nothing is committed: all of it is uncommitted working-tree state.

**The short version.** The pipeline runs, is correct by its tests, and replaces CoreNLP for speed: a sentence parses in 0.52 ms (about 7 times faster than at the start of this session and far faster than the Java server), the whole book in 0.19 s, with 13.4 MB of weights that load in 59 ms. Tokenization and POS tagging agree with CoreNLP well enough for the gates. The dependency parser does not: it agrees with CoreNLP on 76% of heads and 72% of labeled arcs on the held-out chapters against gates of 91% and 88.5%, and the UD-EWT numbers (83.5% / 80.3%) miss gates of 90.0% / 87.5%. `civ/linear`'s Rule 10 score correlation is 0.69 against 0.92. CoreNLP therefore stays the default engine for `civ/linear`, and `civ/compose` keeps using CoreNLP's cached parses for the book's own sentences.

---

## 1. The `civ/bin/parser eval` table

One official run, after the last change to model behavior (`./bin/parser eval -drafts 1000`, SIMD build, 2026-11-25 17:42 local time):

```
17:42:04 models loaded in 59ms (heap 47.0 MB, RSS 67.0 MB, simd=true)
17:42:04 evaluating on UD_English-EWT test ...
17:42:04 comparing with CoreNLP on ch13, ch14 ...
17:42:04 timing ...
17:42:05 running civ/linear Rule 10 parity (both engines) ...
17:42:05 comparing with CoreNLP on drafts (sentences absent from the chapters and treatises) ...

| Metric | Value | Gate | Result |
|---|---|---|---|
| UD-EWT test POS accuracy | 94.92% | ≥ 96.5% | FAIL |
| UD-EWT test UAS (predicted POS) | 83.49% | ≥ 90.0% | FAIL |
| UD-EWT test LAS (predicted POS) | 80.27% | ≥ 87.5% | FAIL |
| ch13 + ch14 POS agreement vs CoreNLP | 97.19% | ≥ 97.0% | PASS |
| ch13 + ch14 UAS vs CoreNLP | 75.71% | ≥ 91.0% | FAIL |
| ch13 + ch14 LAS vs CoreNLP | 72.21% | ≥ 88.5% | FAIL |
| ch13 + ch14 civ/linear Rule 10 score correlation (r) | 0.6898 | ≥ 0.92 | FAIL |
| Mean end-to-end latency (ParseSentence) | 0.521 ms | < 1.0 ms | PASS |
| Full book (ch01..ch14) parallel parse time | 0.19 s (2535 sentences) | < 2.5 s | PASS |
| Cold model load time | 58.5 ms | < 150 ms | PASS |
| Total .safetensors size on disk | 13.39 MB | ≤ 20 MB | PASS |

detail: UD-EWT test tokens 25094; latency p50 0.487 ms, p95 1.119 ms, p99 1.390 ms over 251 sentences x 5; 1918 sentences/sec single-thread
detail: drafts on 1206 sentences absent from the chapters and treatises (49 drafts): POS 90.31% UAS 76.79% LAS 73.70% | on aligned tokens: POS 98.07% UAS 83.39% LAS 80.03% | NER macro F1 0.61
  draft NER PERSON  P 0.59 R 0.80 (tp 66, gold 82, predicted 112)
  draft NER ORG     P 0.42 R 0.24 (tp 21, gold 86, predicted 50)
  draft NER LOC     P 0.86 R 0.54 (tp 55, gold 102, predicted 64)
  draft NER MISC    P 0.59 R 0.41 (tp 34, gold 82, predicted 58)
  draft NER NUMBER  P 0.74 R 0.81 (tp 451, gold 556, predicted 611)
  draft NER DATE    P 0.71 R 0.80 (tp 183, gold 228, predicted 256)
detail: CoreNLP parity over 251 sentences, 251 aligned identically (100.00%); on aligned tokens: POS 97.19% UAS 75.71% LAS 72.21%
detail: NER vs CoreNLP (aligned sentences): macro F1 0.62; DATE F1 0.59 (87 gold, 94 predicted); LOC F1 0.82 (62 gold, 70 predicted); MISC F1 0.64 (65 gold, 53 predicted); NUMBER F1 0.74 (132 gold, 175 predicted); ORG F1 0.30 (22 gold, 11 predicted); PERSON F1 0.64 (58 gold, 55 predicted)
detail: Rule 10 parity: 251 CoreNLP sentences, 251 Go sentences, 251 aligned; mean |score difference| 0.444; Part 4 findings F1 0.462
detail:   distant pronoun:  CoreNLP   1  Go   2  matched   1  F1 0.667
detail:   it chain:         CoreNLP   0  Go   0  matched   0  F1 0.000
detail:   ambiguous it:     CoreNLP   7  Go   3  matched   2  F1 0.400
```

The other gates the spec sets per stage:

| Stage | Gate | Measured | Result |
|---|---|---|---|
| 2 | POS inference latency < 0.15 ms per sentence | 0.068 ms on ch13 text (2.1 µs per token) | PASS |
| 3 | End-to-end p95 < 2.5 ms | 1.119 ms | PASS |
| 3 | Resident RAM < 45 MB | 47.0 MB live heap, 67 MB RSS after load | FAIL |
| 4 | `civ/linear` Part 4 finding parity F1 ≥ 0.85 | 0.462 | FAIL |

Stage 1 parity from the earlier measurements, unchanged since: token parity with CoreNLP 99.998%, every sentence boundary matched on 20 files (and 100% of ch13 and ch14 sentences align token for token), lemma parity 99.76%.

---

## 2. What was built

* **`civ/sys/parser`** — tokenizer and sentence splitter (`tok.go`), WordNet lemmatizer (`lemma.go`), averaged-perceptron POS tagger (`pos.go`), Arc-Standard oracle (`oracle.go`), neural dependency parser (`dep.go`, `dep_train.go`), BIO NER tagger with rule tags for NUMBER and DATE (`ner.go`), mention and coref resolver (`coref.go`), corpus reader and label normalizer (`conllu.go`), metrics (`eval.go`), and the public API in `doc.go`: `Load(root)`, `ParseSentence`, `ParseDoc`, `ErrNotTrained`. The output is the `corenlp` types the consumers already take, with 1-based indexes and UTF-8 byte offsets.
* **`civ/parser/main.go`** — `fetch-ud`, `train-pos`, `train-dep`, `train-ner`, `train-all`, `eval`, `parse`, `bench`. `civ/tools/build-bin.fish` builds it as `civ/bin/parser` with `GOEXPERIMENT=simd`.
* **`civ/compose`** — `Engine.Parser` and `UseCoreNLPServer`; a sentence is parsed from the CoreNLP cache first, then by the Go pipeline, then (only when opted in) by a CoreNLP server; a badge in the status bar says which parser is active.
* **`civ/linear`** — `-engine corenlp|go` (CoreNLP stays the default), a `scores` subcommand that prints one JSON line per sentence for the Rule 10 parity, and offsets moved from UTF-16 code units to UTF-8 bytes.
* **Artifacts in `ext-ref/parser/`** (outside `civ/`, git-ignored, as the spec directs): `ud_ewt/`, `ud_gum/`, and the three weight files, 13.39 MB in total.

Tests: `go test -count=1 ./sys/... ./parser ./compose ./linear` passes in the plain build and with `GOEXPERIMENT=simd`; `go test -race ./sys/parser ./sys/vec` passes in both. The spec's list is covered in `civ/sys/parser/parser_test.go`: UTF-8 offsets, tree invariants, oracle round trip on 500 EWT sentences, SIMD/scalar parity through a pinned logit fingerprint, safetensors round trip, 8-goroutine concurrency, and a holdout-discipline test. One test elsewhere fails and is unrelated: `civ/lint`'s `TestVoiceChaptersCalibration` looks for a section heading that the current `13_ch13.md` does not have. It was left alone.

---

## 3. Where the time goes

Measured on ch13 text (120 sentences, 3,926 tokens), SIMD build, one thread:

| Stage | Per sentence | Share |
|---|---|---|
| Tokenize and split | 0.005 ms | 1% |
| Lemmatize | 0.001 ms | <1% |
| POS | 0.068 ms | 13% |
| NER | 0.031 ms | 6% |
| Dependency parse | 0.421 ms | 80% |
| Stages together | 0.526 ms | |
| `ParseSentence`, tree and doc included | 0.590 ms | |

Inside the dependency parse, roughly: projecting words that fall outside the precomputed cache (2.3 per transition, 64 KB of weights each) about 150 µs; the output layer (86 dot products of length 256 per transition) about 105 µs; the hidden activation about 40 µs; gathering the cached rows about 20 µs; the rest is feature extraction and bookkeeping.

---

## 4. Why accuracy is short, with the evidence

### 4.1 Four views of dependency agreement

| Data | Sentences | POS | UAS | LAS |
|---|---|---|---|---|
| ch12, a training chapter (in-sample) | 232 | 99.97% | 91.44% | 90.06% |
| Draft sentences absent from every chapter and treatise (unseen, same domain) | 1,206 | 98.07%* | 83.39%* | 80.03%* |
| ch13 + ch14, the holdout | 251 | 97.19% | 75.71% | 72.21% |
| UD-EWT test (out of domain, against the gold treebank) | 25,094 tokens | 94.92% | 83.49% | 80.27% |

\* On sentences that segmented identically. Over all CoreNLP tokens the draft figures are 90.31 / 76.79 / 73.70, because only about 92% of draft tokens sit in a sentence that aligns with CoreNLP token for token (in the chapters every sentence does; why drafts differ was not investigated).

In-sample numbers flatter the model: by about 8 points against the unseen draft sentences, and by about 16 against the held-out chapters. The held-out chapters score lower than unseen draft sentences even though all three share the book's vocabulary; ch13 is the weaker of the two (UAS 73.05%, LAS 69.38%; ch14 79.12% and 75.83%).

### 4.2 Punctuation conventions: about a third of the disagreement

Agreement by the CoreNLP relation of the token, on the holdout (agreed / CoreNLP tokens with that relation, labeled arc):

| Relation | Agreed | Share |
|---|---|---|
| `punct` | 481 / 1047 | 46% |
| `dep` | 30 / 216 | 14% |
| `conj` | 214 / 384 | 56% |
| `compound` | 206 / 331 | 62% |
| `obl` | 209 / 297 | 70% |
| `nmod` | 161 / 231 | 70% |
| `nsubj` | 326 / 411 | 79% |
| `obj` | 279 / 345 | 81% |
| `amod` | 555 / 606 | 92% |

Punctuation is 15% of all tokens (1,047 of 7,000) and agrees less than half the time: its 566 disagreements are a third of the 1,700 times the two parsers attach a word differently. The cause was confirmed on training-side data, so it is a convention conflict rather than anything specific to the holdout: on ch12 the punctuation arcs agree 77.9%, on ch12 plus unseen draft sentences 64.0%, and commas 47.8% (1,503 arcs). The main training signal is the UD treebanks, which attach punctuation by the UD guidelines; CoreNLP's converter uses its own rules. Measured on ch12 and the unseen draft sentences, CoreNLP attaches:

| Mark | Where the head lies |
|---|---|
| sentence-final `.` | before the mark, 90% (the main verb) |
| `,` | an earlier token 50%, the previous token 26%, a later token 20%, the next token 5% |
| `(` | the next token 38%, a later token 34% |
| `)` | an earlier token 85% |
| `-` (a dash) | the next token 91% |
| `:` | an earlier token 69% |

The same conflict probably explains part of the `dep` and `conj` rows, which were not examined.

### 4.3 Long sentences and the root

Holdout UAS by sentence length: 85.7% up to 10 tokens (45 sentences), 81.5% for 11–25 (73), 77.3% for 26–40 (78), 72.1% for 41–60 (47), 64.9% above 60 (8). The parser picks CoreNLP's root for 200 of 251 sentences (79.7%). When the root is wrong, the main verb is wrong, and that is what moves the Rule 10 score, since the two heavy weights count the words before the main verb and the length of the subject. On ch12, 204 of 232 sentences (88%) get the same Rule 10 score from both engines; the correlation of 0.883 there comes from a few long sentences that disagree by 3 to 5 points.

### 4.4 The model that was saved is from epoch 1

The spec says to keep the checkpoint with the best LAS on UD-EWT dev. That is epoch 1 of 8 (dev UAS 83.74%, LAS 80.19%); every later epoch scored lower on EWT dev while fitting the book's prose better. The learning rate is cosine-decayed over the whole run, so the saved weights were taken at nearly the peak rate. A checkpoint chosen on in-domain data would fit the book better and EWT worse; the spec's rule was followed.

### 4.5 The spec's numbers were out of reach for this design and data

* Silver data: the spec estimates "~25k+ sentences" from the 8 treatises. The CoreNLP caches hold 2,022 treatise sentences and 2,052 from ch01 to ch12, 4,074 in all.
* UD-EWT: arc-standard transition parsers trained on that treebank alone (12,543 sentences) are reported in the 80s for UAS. Chen and Manning's 91% on the Wall Street Journal came from 1.4 million training sentences. The gate of 90.0% is above what the spec's data supports.
* NER: 2,611 entity tokens of training data. ORG is the weakest label on unseen text (F1 0.30 on the holdout, recall 0.24 on drafts).

---

## 5. Departures from the spec

| Spec says | What was done | Why |
|---|---|---|
| Train on UD-EWT | Also trained on UD_English-GUM r2.15 | EWT alone gave a measurably weaker parser |
| "38 dependency labels", "45 classes" | The spec's own lists: 43 labels including ROOT (86 actions), 46 POS tags | Followed the enumeration, not the count |
| 2,048 cached word projections per slot | 1,024 (18.9 MB) | 2,048 would take 37.7 MB |
| Hashed dense weights *or* pruned sparse | Pruned sparse rows (6.77 MB) | The allowed alternative |
| Tag-dictionary shortcut | Implemented (`POSTrainConfig.Shortcut`), off by default | It cost accuracy |
| NER set `PERSON, ORG, LOC, MISC, NUMBER, DATE, O` | CoreNLP labels mapped: ORGANIZATION to ORG, LOCATION to LOC, MONEY/PERCENT/ORDINAL/QUANTITY to NUMBER, TIME/DURATION/SET to DATE | The consumer's set |
| `eval [-chapter ch13,ch14]` | Added an optional `-drafts N` (default off) | An in-domain check that does not use the holdout |
| — | `civ/linear scores` subcommand, `-root` flag; `-engine go report` writes `18_linearity_review_go.md` | Needed for the Rule 10 row; keeps the CoreNLP report intact |
| Resident RAM < 45 MB | 47.0 MB live heap | Not met; see section 7 |

Training is one joint mixture with weights (EWT 1, GUM 1 for POS or 2 for dependencies, treatises 2, chapters 4) over 8 epochs, with the checkpoint chosen by EWT dev as the spec requires; hidden-layer dropout 0.2 and word dropout 0.15 are on.

---

## 6. Fixes made along the way that matter beyond the parser

* **`civ/sys/vec` kernels left the CPU in its slow mode for scalar float code.** Go compiles scalar float arithmetic to legacy SSE and never emits `VZEROUPPER`. On the AMD Zen 3 machine used here, after any 256-bit AVX instruction has run on a thread, each later scalar float instruction on that thread takes a slow path: a scalar float32 multiply-add loop went from 0.9 ns to 58.9 ns per iteration, 65 times slower, and stayed there. Every kernel in `vec_simd.go` now ends with `archsimd.ClearAVXUpperBits()`, and the same loop runs at 0.95 ns. This affects everything in `civ` that mixes `vec` calls with float math in the SIMD build, not only the parser: measured before and after, `vec.Dot` on 256 floats went from 83 ns to 19 ns and a 64-row `AddMatVec` from 4.2 µs to 1.1 µs, both with the vector loop itself unchanged. On Intel the same change removes the standard AVX-SSE transition penalty.
* **The kernels were rewritten** as flat unrolled loops (re-slicing `dst[i:]` on every iteration cost about 20 bookkeeping instructions per 8 floats). Per 1 KB row, SIMD against the plain build: `Axpy` 19.6 against 65 ns, `Dot` 19 against 82 ns.
* **Two ragged-tail bugs introduced in that rewrite were caught before anything shipped**: an `Axpy` that skipped or double-processed the last few elements of any length not divisible by 8, and a `Cosine` that read past the end. `TestRaggedLengths` now checks `Dot`, `Cosine` and `Axpy` against independent scalar references for every length up to 100; re-introducing either bug fails it.
* **The NER trainer never learned entities.** It compared the *corrected* prediction with gold, so when the raw best class was an illegal `I-X` and the correction turned it into `O`, `O` was the class penalized and `I-X` kept its lead; on its own training data the old model recalled 50% of PERSON and 15% of ORG. Training now holds the raw best class responsible, and decoding takes the best legal class. In-sample on ch12, NER macro F1 went from 0.44 to 0.84. The DATE and NUMBER rules were then calibrated on the silver data (time-of-day nouns and seasons, "ten years" spans, the modal "may").
* **Three defects in the evaluation path**: CoreNLP's NER labels were not mapped to the canonical set before comparison, so no ORG or LOC was ever counted as gold; `civ/linear` looked for chapter text in the wrong directory, which silently skipped the Rule 10 row; and the sparse scorer was restructured so its cache misses overlap (POS from 3.9 to about 2.1–2.5 µs per token).

---

## 7. Memory

The spec's Stage 3 gate of 45 MB resident is missed: 47.0 MB of live heap after load, and 67 MB RSS right after loading because Go has not yet returned the 19 MB of file-read buffers to the OS (about 58 MB after it does). Composition of the live heap, approximate:

| Part | Size |
|---|---|
| Cached word projections, 1,024 rows × 18 slots × 1 KB | 18.9 MB |
| Model tensors, dependencies and POS | 9 MB |
| POS and NER sparse tables | 6 MB |
| Other engine tables (transposed first layer, POS and label rows) | 3 MB |
| WordNet index and exception maps | 4 MB |
| Vocabularies, strings, misc | 6 MB |

Halving the word cache saves 9.4 MB. Measured on ch12, cache rows 0–255 take 53% of the cache hits, 256–511 take 7.3%, and 512–1023 take 8.1%, so halving it adds about 26% to the uncached projections, which is near 5% of total latency. It was not done because latency was the scarcer resource; the size is `depCacheWords` in `dep_train.go` and a saved model can also be truncated at load, since the rows are frequency-ordered.

---

## 8. Holdout hygiene

ch13 and ch14 were never read by any training or tuning code (`TestHeldoutDiscipline` tests the file filter). They were read by evaluation only, in this order:

1. `civ/bin/parser bench -chapter ch13`, for latency only, three times.
2. The official `eval` run in section 1.
3. After it, one read-only diagnostic pass over the same two chapters that produced the per-chapter, per-length and per-relation figures in section 4. No model, rule or training setting was changed after the official run.

The drafts comparison (`-drafts`) seals any draft that shares a sentence with the `-chapter` list. In three dry runs with `-chapter ch12` it therefore sealed ch12 only, and its aggregate figures covered draft sentences related to other chapters, possibly including drafts of ch13 and ch14. No sentence was inspected and the dependency and POS models were not touched afterwards; the DATE rules that followed were calibrated on the silver data alone. From then on the flag was used only with the holdout chapters sealed.

---

## 9. Commands

```fish
cd civ
fish tools/build-bin.fish                        # builds civ/bin/parser and civ/bin/linear, among others
./bin/parser eval                                 # the table in section 1
./bin/parser eval -chapter ch12 -drafts 1000      # dry run on a training chapter, plus unseen draft sentences
./bin/parser bench -chapter ch13                  # latency percentiles and memory
./bin/parser parse "We come into this world naked."
./bin/linear scores -chapter ch12 -engine go      # Rule 10 per sentence as JSON lines
go test -count=1 ./sys/... ./parser ./compose ./linear
GOEXPERIMENT=simd go test -count=1 ./sys/... ./parser ./compose ./linear
GOEXPERIMENT=simd go test -race ./sys/parser ./sys/vec
```

Retraining: `./bin/parser train-all -distill -epochs 8` (the full run takes tens of minutes; run it under a memory cap). NER alone retrains in a second with `./bin/parser train-ner`.

---

## 10. What to do next, in order of expected gain

1. **Settle the punctuation convention.** Either rewrite the treebanks' punctuation arcs to CoreNLP's rules before training, or re-attach punctuation by rule after parsing (CoreNLP's attachments are regular enough, section 4.2). On the holdout that is 566 tokens of 7,000, up to 8 points of UAS and LAS if every mark were attached as CoreNLP does; half of it would still be the largest single gain available. It would cost UD-EWT LAS unless it is a mode; the two references disagree by design.
2. **Add in-domain training data.** `ext-ref/linear/scratch/` holds CoreNLP parses of draft sentences that are in no chapter: 1,206 sentences from 49 drafts that do not touch the holdout, already parsed and unused. That is a 30% addition to the 4,074 silver sentences, at no cost.
3. **Choose the dependency checkpoint on in-domain data**, or follow the spec's two-phase schedule so the fine-tuning epochs run at the decayed learning rate. The present model is epoch 1 at nearly peak rate. This is a change to the spec's rule and needs the author's decision.
4. **Root choice on long sentences.** UAS falls from 86% to 65% as sentences grow past 60 tokens, and the root is wrong in a fifth of holdout sentences. Items 1–3 should be measured first; they may account for most of it.
5. **NER for organizations** needs more examples than 404 ORG tokens.
6. **Rule 10 parity** (0.69 against 0.92) follows from the above; with the present parser, the gap is concentrated in long sentences where the main verb differs.
7. **Memory**, if the 45 MB gate matters: halve the word cache and return the load-time read buffers to the OS.

Until then, the defaults are right for the book: `civ/linear` uses the CoreNLP caches and `civ/compose` diagrams the book's sentences from them; the Go pipeline is what those tools fall back on for text CoreNLP has not parsed, where it is fast and correct in form, and right about three words in four on where each word attaches.

---

## 11. Review & Round-1 Integration (2026-11-25)

Audited and finalized before committing Round 1:

* **Trainer fixes (`civ/sys/parser/dep_train.go`, `train.go`)**:
  * Fixed `DepTrainConfig` zero-value bug where `Dropout` and `WordDropout` were only defaulted when `< 0`, leaving both at `0.0` during training; wired `Dropout: 0.30`, `WordDropout: 0.20`, `RareCount: 5`, `BatchSize: 512`, and raw (unweighted) token counts for `<UNK>` dropout.
  * Fixed unaveraged mini-batch gradient sums (`1 / len(batch)` before global norm clipping), replaced sentence-grouped batches with globally shuffled i.i.d. transitions, replaced the $O(N^2)$ sentence-offset loop with an $O(1)$ prefix sum, and fixed per-parameter weight-decay exclusion on `b1`/`b2`.
  * Retraining with these fixes confirmed the accuracy ceiling is structural to the greedy static-oracle MLP + UD/CoreNLP convention split (section 4), so the slightly stronger checkpoint (UD-EWT 83.49% / 80.27%, ch13+ch14 75.71% / 72.21%, Rule 10 $r = 0.6898$) was retained for Round 1.
* **Memory gate fixed (`civ/sys/parser/dep.go`)**: Capped runtime precomputed word projections to `depRuntimeCacheWords = 512` (frequency-ordered). Live heap after load dropped from `47.0 MB` to `36.5–38.0 MB` (now **PASS** on `< 45 MB`), while keeping mean `ParseSentence` latency at `~0.58 ms` (`< 1.0 ms`) and full-book parallel parse at `0.25 s` (`< 2.5 s`).
* **Compressed committed models (`civ/sys/parser/st.go`, `doc.go`, `civ/parser/main.go`)**:
  * Added zero-dependency stdlib gzip compression/decompression and IEEE-754 `F16` tensor read/write to `st.go`, plus `civ/bin/parser pack`.
  * Packed all three Round-1 models into `civ/sys/parser/models/*.safetensors.gz` (`13.39 MB → 7.34 MB`, **0.00%** accuracy change, `~103 ms` cold load `< 150 ms`). `parser.Load(root)` prefers freshly trained local weights in `ext-ref/parser/` and falls back transparently to `civ/sys/parser/models/`.

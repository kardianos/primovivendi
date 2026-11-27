# 48a. Handoff — Round 2: five of fourteen gates pass; every accuracy figure is up on Round 1, none reaches its gate, and the in-domain gates are bounded by the reference itself

Official run: `civ/bin/parser eval` with its defaults, once, on 2026-11-26 at 14:49:27, after training, NER, coref and packing were frozen. This note replaces the interim one. Nothing is committed.

## 1. The `civ/bin/parser eval` table

| Metric | Round 1 | Round-2 Gate | Value | Result |
|---|---|---|---|---|
| UD-EWT test POS accuracy | 94.92% | ≥ 96.5% | 95.26% | FAIL |
| UD-EWT test UAS (predicted POS) | 83.49% | ≥ 90.0% | 87.96% | FAIL |
| UD-EWT test LAS (predicted POS) | 80.27% | ≥ 87.5% | 85.13% | FAIL |
| ch13 + ch14 POS agreement vs CoreNLP | 97.19% | ≥ 97.5% | 96.70% | FAIL |
| ch13 + ch14 non-punct UAS vs CoreNLP | ~81% | ≥ 91.0% | 84.63% (all-token 77.04%) | FAIL |
| ch13 + ch14 non-punct LAS vs CoreNLP | ~77% | ≥ 88.5% | 80.13% (all-token 73.20%) | FAIL |
| ch13 + ch14 non-punct UAS vs CoreNLP, CoreNLP `dep` left out | — | report | 86.16% over 5,737 tokens | — |
| ch13 + ch14 non-punct LAS vs CoreNLP, CoreNLP `dep` left out | — | report | 82.67% | — |
| ch13 + ch14 Root agreement vs CoreNLP | 79.7% | ≥ 92.0% | 85.26% | FAIL |
| ch13 + ch14 Root agreement, copula or auxiliary as head allowed | — | report | 85.66% | — |
| ch13 + ch14 `civ/linear` Rule 10 correlation (r) | 0.6898 | ≥ 0.92 | 0.8011 | FAIL |
| ch13 + ch14 `civ/linear` Part 4 findings F1 | 0.462 | ≥ 0.85 | 0.571 | FAIL |
| Mean end-to-end latency (`ParseSentence`) | 0.620 ms | < 1.5 ms (stretch < 1.0 ms) | 1.108 ms | PASS |
| Full book (ch01..ch14) parallel parse time | 0.27 s | < 1.5 s | 0.53 s (2,535 sentences) | PASS |
| Cold model load time | 104 ms | < 150 ms | 63.9 ms | PASS |
| Live heap after load | 36.6 MB | < 45 MB | 14.9 MB | PASS |
| Total packed `.safetensors.gz` size on disk | 7.34 MB | ≤ 15 MB | 3.79 MB | PASS |

Two rows are mine, added on the reviewer's decision to report the in-domain figures with CoreNLP's `dep` left out, and root agreement with the copula-or-auxiliary head swap that separates Stanford dependencies from UD.

Detail of the run: 251 CoreNLP sentences aligned identically; Rule 10 mean |score difference| 0.331; Part 4 `distant pronoun` CoreNLP 1, Go 2, matched 1; `ambiguous it` 7, 4, 3; `it chain` none on either side (the table prints an F1 of 0.000 for that). NER against CoreNLP: macro F1 0.56 (PERSON 0.52 with 58 gold and 38 predicted, ORG 0.22, LOC 0.81, MISC 0.51, DATE 0.58, NUMBER 0.74). Latency p50 1.017 ms, p95 2.328 ms, p99 3.113 ms over 251 sentences × 5, with the machine at a load average of 5.7. The load time is that of the raw weights (`eval` resolves `ext-ref/parser/` first); from the packed files it is 96 ms on a mostly idle core and 134 ms under load.

## 2. What the table says

* **The five resource gates pass with margin.** The unified model takes 14.9 MB of heap where Round 1's three models took 36.6, and packs to 3.79 MB.
* **On UD, the model is 0.3 POS, 4.5 UAS and 4.9 LAS above Round 1**, and still 1.2, 2.0 and 2.4 points under the gates.
* **In domain, everything moved except POS agreement.** All-token UAS and LAS 77.04% and 73.20% against Round 1's 75.71% and 72.21%; root agreement 85.26% against 79.7%; Rule 10 correlation 0.80 against 0.69; Part 4 F1 0.57 against 0.46. POS agreement with CoreNLP fell from 97.19% to 96.70% although UD POS accuracy rose.
* **The development chapter predicted the holdout within half a point** on every figure but root agreement (section 4), so `dev_book` is a usable stand-in and the holdout did not have to be consulted while tuning.
* **What did not reach its target:** the Part 4 target of the specification (F1 ≥ 0.88 on ch01..ch12) is 0.71 there, and the ORG recall target (≥ 0.55 on the drafts) is 0.94 only because the drafts are now training data for the NER tagger (section 8.5).

## 3. The model and the recipe

`ext-ref/parser/parser.safetensors`, packed as `civ/sys/parser/models/parser.safetensors.gz` (3.55 MB), with `ner.safetensors.gz` (0.24 MB). The two Round-1 bundles are removed from `civ/sys/parser/models/`; the Round-1 files under `ext-ref/parser/` are untouched and still load if the unified model is absent.

* Phase A only, 30 epochs, `H = 128`, seed 20261126, 1,880,204 parameters, 26 min 48 s on 8 workers with the machine under load. Phase B was not run (decision: keep UD labels, do not pull toward CoreNLP silver). LinES was not used.
* The specification's hyperparameters, with `lambda_pos = 1.0`, its value for joint training; every hyperparameter is on the first training line of the log.
* Mix per epoch: UD-EWT and UD-GUM at weight 1, the eight treatises 1, the unsealed drafts 1, ch01..ch11 at 2. 29,088 sentences (27,036 unique), 552,568 tokens, 909 optimizer steps.
* Checkpoint selection by the specified score (`0.5 × dev_ud LAS + 0.5 × dev_book non-punct LAS` = 0.8223); the best epoch was the last, and UD dev was still rising by about 0.03 UAS per epoch there.
* Drafts: 104 files, 55 sealed for sharing a sentence with ch13 or ch14, 3 sealed for sharing one with ch12, 46 loaded, 1,130 sentences.

Learning curve on dev (UD-EWT dev, then ch12 against CoreNLP):

| Epoch | UD POS | UD UAS | UD LAS | ch12 non-punct UAS | ch12 non-punct LAS |
|---|---|---|---|---|---|
| 10 | 94.34% | 86.46% | 83.15% | 83.73% | 78.86% |
| 15 | 94.56% | 87.12% | 83.85% | 84.24% | 79.31% |
| 20 | 94.51% | 87.52% | 84.37% | 83.95% | 79.36% |
| 30 | 94.75% | 87.81% | 84.77% | 84.34% | 79.69% |

**One experiment beyond the decided recipe, and it lost.** With ch01..ch11 at weight 4 for the same 30 epochs: selection score 0.8205 against 0.8223, UD dev 94.78 / 87.44 / 84.36 against 94.75 / 87.81 / 84.77, ch12 non-punct UAS 83.82% against 84.34%, LAS 79.74% against 79.69%, Rule 10 correlation on ch12 0.766 against 0.790. CoreNLP's parses of this book's long sentences are not good teaching material; more of them costs UD and buys nothing in domain. `-silver N` keeps the knob (`train-parser`), default 2.

## 4. Development figures, for the next round

| Figure | ch12 (`dev_book`, never trained on) | ch13 + ch14 (holdout) |
|---|---|---|
| POS agreement with CoreNLP | 96.33% | 96.70% |
| All-token UAS / LAS | 76.85% / 72.91% | 77.04% / 73.20% |
| Non-punct UAS / LAS | 84.27% / 79.62% | 84.63% / 80.13% |
| Without CoreNLP `dep` | 85.87% / 82.24% (231 tokens left out) | 86.16% / 82.67% (216 left out) |
| Root agreement | 81.03% (84.91% with the head swap) | 85.26% (85.66%) |
| Rule 10 correlation | 0.7904 | 0.8011 |
| Part 4 F1 | 0.50 (3 findings for CoreNLP, 9 for Go) | 0.57 (8 for CoreNLP, 6 for Go) |

Over ch01..ch12 the Rule 10 correlation is 0.941 and 174 of 2,284 sentences (7.6%) get a different score. That figure is inflated: ch01..ch11 are in the parser's training data. On ch12 it is 57 of 232 (25%).

## 5. Tests

All on the tree as delivered, 2026-11-26, machine under load:

| Build | Packages | Result |
|---|---|---|
| `go test -count=1` | `./sys/... ./parser ./compose ./linear` | 9 packages ok |
| `GOEXPERIMENT=simd go test -count=1` | the same | 9 ok |
| `GOEXPERIMENT=simd go test -race -count=1` | `./sys/parser ./sys/vec ./sys/nn ./compose ./linear` | 5 ok, no race reported |
| `go test -race -count=1` | `./sys/parser ./sys/vec` | 2 ok, no race reported |

The five tests the specification names pass in both builds, under `-race` too. `TestOverfitMiniBatch` runs at a constant learning rate of 2e-2: at 1.5e-2 it sat on the edge of stability under `lambda_pos = 1.0` (loss 0.03 at step 40, 0.6 at step 45 in the SIMD build) and ended at 0.0012 in both builds at 2e-2. New since the interim note: `TestCorefsItMentions` (which "it" the coreference stage reads as a pronoun; fails on all six of its cases under the old behaviour) and `TestBiaffineModelInventories` (a saved model from a build with other tag or relation tables is refused at load).

## 6. Departures from the specification

| Specification says | What was done | Why |
|---|---|---|
| `MatMulT` in `archsimd` Go | Two leaf kernels in hand-written assembly (`vec_mat_amd64.s`) | The compiler emits a register copy per `MulAdd` accumulate: about 15 GMAC/s against 46 |
| 15 to 20 epochs, `lambda_pos` 0.5 | 30 epochs, `lambda_pos` 1.0 | The review's decision; 1.0 is the specification's own value for joint training |
| Phase B when `dev_book` lags | Not run | Decision: keep UD labels |
| UD-LinES at weight 1 | Optional (`-lines`), off | Decision; also CC BY-NC-SA, and its XPOS column is not PTB |
| Root candidates by margin | The eight best margins plus the roots already chosen | As allowed; trying every token gave the same 2,001 dev trees |
| One label scorer | The trainer's, and an equal one on zero-padded rows for inference | No ragged tail in `MatMulT`; `TestBiaffinePaddedLabelPath` checks they agree |
| Drafts sealed against ch13 and ch14 | Sealed against ch12 as well | So that `dev_book` measures what the holdout will (3 drafts, 75 sentences) |
| Report non-punct UAS/LAS | Also with CoreNLP's `dep` left out, and root agreement with the head swap | The review's decision |
| `TestOverfitMiniBatch` | Small dims, learning rate 2e-2 | Half a second with two workers |

## 7. What the review fixed, and what it means for the interim note

The review found five bugs in the inference path, all fixed in-tree and all now held by tests: `biEngine.parse` recomputed the char-CNN for every token, which zeroed the features of exactly the cached frequent words; `posInUnk` indexed the wrong row of the POS embedding (training only); the GRU matvec ran as 96 width-1 calls per step; two scalar tails ran with dirty upper YMM halves; spans over 400 tokens were tagged whole before being parsed in chunks.

Every number in the interim note was measured through the broken path and is superseded. With the same weights, on ch12, the first fix alone moved POS agreement from 94.81% to 96.6% and non-punct LAS from 79.2% to 80.8%. The UD figures of that note were right (UD scoring never went through the cache).

Fixed by me before the review, each held by a test: two data races in the trainer (a scratch buffer shared by the workers; the AdamW update and the refresh of the transposed weights running as one phase although partitioned differently), a backward buffer not cleared per sentence, a missing `clear` in the GRU backward pass, transposed-weight products, and a pool whose `New` was never set when a pipeline was built outside `Load`.

## 8. Findings that bear on what to do next

### 8.1 Rule 10 parity is limited by the parse, and in-sample chapters flatter it

Where the two engines' scores differ, by finding kind (sentences; final model):

| Kind | ch01..ch12: both / only CoreNLP / only Go | ch12 alone |
|---|---|---|
| `nested clauses` | 164 / 34 / 33 | 6 / 12 / 14 |
| `ambiguous it` | 30 / 18 / 21 | 2 / 0 / 3 |
| `opening phrase` | 153 / 15 / 12 | 11 / 5 / 7 |
| `long subject` | 152 / 15 / 12 | 10 / 10 / 5 |
| `effect before cause` | 22 / 17 / 1 | 1 / 1 / 0 |
| `late verb` | 290 / 8 / 7 | 22 / 7 / 7 |
| `second claim` | 161 / 4 / 8 | 11 / 3 / 3 |
| `distant pronoun` | 6 / 0 / 3 | 0 / 0 / 3 |
| `bare this` | 33 / 0 / 1 | 6 / 0 / 1 |
| `long sentence`, `aside`, `it chain`, `new before old` | all agreed | all agreed |

Part 4 is the second cause over ch01..ch12 and a minor one on ch12. The larger causes are clause structure read off the tree. On ch12, of the clausal edges: `advcl` 98 for CoreNLP against 100 for Go with 68 in common; `acl` 60 / 57 / 29; `ccomp` 66 / 63 / 40; `xcomp` 57 / 58 / 36; `parataxis` 21 / 45 / 12. We emit twice CoreNLP's `parataxis`, and 45 of CoreNLP's clausal tokens are its unclassified `dep`.

### 8.2 `effect before cause`: a comma convention, fixable in civ/linear

In all 17 sentences where CoreNLP reports it and we do not, the because-clause is attached identically by both engines. The rule tests that the clause starts at "because", and under UD's convention the comma before the conjunction hangs on the subordinate clause and is part of its span. Where a comma before a subordinating conjunction goes:

| Corpus | Such commas | On the clause head |
|---|---|---|
| UD-EWT train | 206 | 68% |
| UD-GUM train | 228 | 63% |
| CoreNLP on the treatises | 149 | 8% |
| CoreNLP on ch01..ch11 | 68 | 3% |
| Go pipeline on ch12 | 8 | 88% |

Our model follows UD. With those commas attached CoreNLP's way, as a temporary experiment, the rule went from 22 agreed, 17 missed, 1 extra to 38, 1, 2, and the sentences with any score difference from 173 to 159; the correlation moved by 0.002. The specification's premise that Rule 10 does not depend on punctuation attachment is not quite true: it depends through clause spans. The fix belongs in `civ/linear` (let a clause begin with a comma), which is a change to the product and was not made; the experiment was removed.

### 8.3 Root agreement is bounded by the reference

Of the 44 root disagreements on ch12, read one by one: 12 are ours, 8 are the copula-or-auxiliary convention, 17 are CoreNLP's own (its root is the subject noun: "Burke" for "…Burke … articulated", "Yarvin" for "…rightly diagnoses", even "the" once), 4 are debatable. On the holdout the convention accounts for a single sentence (85.26% against 85.66%).

Ours fall into two patterns: a long or coordinate subject taken as the root ("Curtis Yarvin (…) rightly notes …" rooted on the name; "Supremacist sharia jurisdiction, … are incompatible" rooted on the first conjunct), and "X is Y *stated/given/governed* …" where the participle is taken as the main verb.

### 8.4 CoreNLP's `dep`

216 of the holdout's 5,953 non-punct tokens (3.6%) carry it, and our LAS on them is 12.5% (15% on ch12). Leaving them out lifts non-punct UAS and LAS by 1.5 and 2.5 points.

### 8.5 Coreference: a small principled gain, and a ceiling

The cache's chains come from CoreNLP's neural coref annotator, a mention-ranking model. Among the 82 "it"s of ch01..ch12 that have an antecedent in their own sentence and at least two nouns before them, it picks the *nearest* noun only 22 times; otherwise it reaches over nouns in adjuncts and embedded clauses, often to the sentence's first noun phrase, by semantic judgement (`airplane` over `spar`, `card` over `forecast`).

| Rule for "it" (ch01..ch12, Round-1 NER held fixed) | `ambiguous it`: CoreNLP / Go / matched | F1 | Rule 10 r |
|---|---|---|---|
| As shipped | 50 / 46 / 28 | 0.583 | 0.9327 |
| **Delivered**: extract exactly the "it"s `civ/linear` reads as pronouns, leave the anticipatory "it means that" unlinked, omit unlinked mentions | 50 / 52 / 32 | 0.627 | 0.9362 |
| Reach to the first noun that is an argument of the matrix or a fronted clause | 50 / 92 / 44 | 0.620 | 0.9251 |

The third rule finds 16 more of CoreNLP's findings and fires on 40 more sentences; each costs 2 on our score, so it lowers the correlation. A grid search over weighted salience (recency, subject, discourse-old, pronoun continuity; 1,728 weight sets) topped out at 0.620. With the final NER the delivered rule gives 50 / 52 / 33 (F1 0.647) and r 0.941; over all three Part 4 kinds 0.71 on ch01..ch12, against the specification's 0.88.

### 8.6 NER does not generalise to names it has not seen

ORG recall on the drafts is 0.94 (precision 0.99) and the NER macro F1 there 0.90, but those drafts are the tagger's training data. On the unseen chapters: ch12 macro F1 0.60, holdout 0.56; PERSON F1 0.67 and 0.52. The misses are names the 4,972 training sentences do not contain ("Kant" alone, "Karl Marx", "Muhammad", a list of scientists tagged `LOC`; "Saudi Arabia" and "Brunei" as `PERSON`). ch12 has three `distant pronoun` findings to CoreNLP's none, and two of the three are plainly such a miss: "the reports of what Muhammad said … the account of his life", and a sentence opening on "He" right after one about "Kant" that we leave untyped. CoreNLP is itself inconsistent on this text (`Islam` is ORG in four sentences of ch12 and MISC in three; `Epictetus`, `Galileo` and the "Justice" of "A Theory of Justice" are ORG).

Tried and reverted, measured on ch12: masking entity tokens during training so the tagger learns from context (macro F1 0.56 at a masking rate of 0.6, 0.50 at 0.85, against 0.60); building the lexical prior per label rather than per BIO class (neutral for NER, two more false `distant pronoun` findings and Rule 10 r 0.784 against 0.790).

### 8.7 Latency

The official figure is 1.108 ms at a load average of 5.7. Pinned to a core whose two threads were under 15% busy, the dev chapter gives 1.12 to 1.16 ms over three runs and 1.14 ms over 4,640 runs (p50 0.95 to 1.02, p95 2.3 to 2.4, p99 3.1 to 3.5). A sentence executes 14.8 million instructions. The gate passes; the 1.0 ms stretch does not. A profile taken before the review's fixes put 65% in the forward pass (27% in the GRU recurrences), 18% in the label scorer, 7% in NER, 3% in the MST decoder; it was not repeated. A column-panel loop in `MatMulT` was tried and reverted (within 1% on 18 of 20 shapes). The lever left is running the two GRU directions concurrently, about a tenth of a sentence.

## 9. Holdout hygiene

In Round 2, ch13 and ch14 were read twice: as raw text by `HoldoutSealKeys`, to seal the 55 drafts that share a sentence with them (never as parses, never for training or selection), and by the one official `eval`. Every development figure in this note comes from ch12, UD dev, or the drafts. `TestSealedDraftsHoldout` builds a fixture root from scratch, shows the same drafts leak without the seal, and re-checks the repository. `eval -drafts` seals drafts against ch13, ch14 and the chapters being evaluated.

## 10. Commands

From `civ/`, binaries built by `civ/tools/build-bin.fish`:

```
civ/bin/parser train-parser [-epochs 30] [-silver 2] [-seed S] [-workers W] [-init PATH]
civ/bin/parser train-ner
civ/bin/parser train-all [-epochs 30]
civ/bin/parser eval -chapter ch12 -ud-split dev -no-book [-drafts 200] [-models auto|bundled|local]
civ/bin/parser part4 -chapter ch01,...,ch12
civ/bin/parser pack
civ/bin/parser eval                                             # the official run, once
go test -count=1 ./sys/... ./parser ./compose ./linear
GOEXPERIMENT=simd go test -count=1 ./sys/... ./parser ./compose ./linear
GOEXPERIMENT=simd go test -race -count=1 ./sys/parser ./sys/vec ./compose ./linear
```

Training is memory-capped on this machine: `systemd-run --user --scope --quiet --collect -p MemoryMax=8G -p OOMPolicy=continue -- civ/bin/parser train-parser -epochs 30`.

Packing was verified: on UD-EWT dev, 0 of 25,149 tokens differ between the raw F32 weights and the packed F16 ones (POS, head, relation); on ch12 the largest movement is 0.05 points (non-punct LAS 79.62% to 79.67%). `eval` gives the same table with `-models bundled` as with the raw weights present, and with `ext-ref/parser/*.safetensors` moved aside.

## 11. What to do next, in order of expected gain

1. **Decide what the in-domain gates should measure.** Against CoreNLP's output on these chapters the ceiling for a parser that keeps UD conventions is around 86 / 83 without `dep`, by the attribution in sections 8.2 to 8.4. Either the gates become "without `dep`, content head as root", or the reference changes.
2. **`civ/linear`: let a clause begin with a comma** in `effect before cause` (17 of 18 misses; the sentence spans of other clause rules deserve the same look).
3. **NER person names.** PERSON recall on unseen chapters (F1 0.52) feeds `personLike` and so Part 4 and Rule 10. A name gazetteer, or a tagger with lexical resources beyond 5,000 sentences, is needed; features trained on the silver set do not generalise to names it lacks.
4. **The parser's two root patterns** (long and coordinate subjects; "X is Y participle"). More CoreNLP silver does not help (section 3); what might is more gold treebanks in the mix.
5. **The UD gates** (96.5 / 90.0 / 87.5) ask for more than this design has: no pretrained embeddings, 64-dimensional word embeddings learned on about 27,000 sentences. I do not expect them to be reached without pretrained vectors, which the specification does not allow.
6. **Latency toward 1.0 ms**, only if wanted: concurrent GRU directions.

---

## 12. Final Review & Integration (2026-11-26)

* Verified all builds (`civ/tools/build-bin.fish`), scalar + `GOEXPERIMENT=simd` test suites (`./sys/... ./parser ./compose ./linear`), `-race` (`./sys/parser ./sys/vec ./compose ./linear`), and `./bin/parser eval -models bundled` against the committed `3.79 MB` F16+gzip models (`parser.safetensors.gz` + `ner.safetensors.gz`).
* Applied the consumer UD-comma fixes identified in section 8.2:
  * `civ/linear/analyze.go`: `effect before cause` now accepts a leading punctuation token before `"because"` inside the `advcl` span.
  * `civ/sys/linear/linear.go`: `pull_front` strips leading commas on UD-attached `advcl`/`obl` spans, and `trailingTailSafe` handles forward-attached trailing `advcl` spans under the Round-2 unified parser (verified live on Chapter 01 Paragraph 2 S6 in `compose`).
* Updated defaults and running instructions across `civ/parser/main.go`, `civ/sys/parser/train.go` (default `EpochsA = 30`), `civ/compose/README.md`, `civ/compose/main.go`, `civ/linear/main.go`, `civ/linear/engine.go`, and `civ/linear/report.go`. Kept `-engine corenlp` as the default for `civ/linear`.

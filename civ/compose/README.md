# `civ/compose` — Interactive Sentence & Paragraph Rebuilder

`compose` is a local Go + `gowaku` GUI and CLI tool for rebuilding `civ/chapters/*.md` sentence by sentence, phrase by phrase, and paragraph by paragraph using:

* **Sentence diagramming** (Stanford CoreNLP 4.6.0 cached parses in `ext-ref/linear/out/` for the original sentences, and the pure-Go pipeline `civ/sys/parser` for anything else: manual rewrites, custom phrases, uncached sentences. An optional local `StanfordCoreNLPServer` can be opted into with `-corenlp-server`.)

The Go pipeline parses a sentence in about 0.5 ms (no Java), but it agrees with CoreNLP on only about 76% of head attachments and 72% of labeled arcs on held-out chapters, so read its diagrams as a first pass. `civ/notes/47a_go_parser_handoff.md` has the measured table (`civ/bin/parser eval`) and the known gaps.
* **Mechanical structural transforms & Rule 10 scoring** (`civ/sys/linear`)
* **WordNet 3.1 + English POS inflector + concreteness / AoA / CMUdict stress** (`civ/sys/lexicon`, `civ/sys/cadence`)
* **1850–1950 Google Books Ngrams & reference-corpus collocations** (`civ/sys/ngram`)
* **Pure-Go `BAAI/bge-small-en-v1.5` semantic similarity & trajectory** (`civ/sys/embed`)
* **Pure-Go `inkline`-style local NN preference ranker** (`civ/sys/nn`, saved as `.safetensors`)

`compose` **never edits `civ/chapters/*.md` directly**. It first freezes an immutable snapshot under `civ/notes/compose/snapshots/`, logs your full decision tree under `civ/notes/compose/decisions/`, and deterministically rebuilds chapters into `civ/notes/compose/rebuilt/` when you explicitly run `rebuild`.

---

## Quick Start

```fish
# Rebuild all civ binaries (includes civ/bin/compose):
./tools/build-bin.fish

# One-time download/unpack of full Princeton WordNet 3.1 synsets into ext-ref/lexicon/wordnet31/:
./bin/compose fetch-wordnet

# Open the interactive GUI on Chapter 1:
./bin/compose ui -chapter ch01

# Train the pure-Go parser once (tokenizer, POS, dependencies, NER, coreference; a few minutes):
./bin/parser train-all -distill

# Opt in to a local CoreNLP HTTP server for re-parsing instead of the Go parser:
./bin/compose ui -chapter ch01 -corenlp-server

# Open directly in a specific graph view (lattice | diagram | hybrid):
./bin/compose ui -chapter ch01 -view hybrid

# Enable periodic timing profiler (flushes span totals, slow calls >=50ms, heap/goroutines every 5s):
./bin/compose ui -chapter ch01 -corenlp-server -profile notes/compose/profile.log -profile-interval 5s
# Watch live in another terminal:
tail -f notes/compose/profile.log
```

---

## CLI Subcommands

| Command | Purpose |
| --- | --- |
| `./bin/compose ui [-chapter ch01] [-view lattice\|diagram\|hybrid] [-corenlp-server]` | Launch the interactive `gowaku` GUI. |
| `./bin/compose fetch-wordnet` | Download and unpack Princeton WordNet 3.1 (`data.*`, `index.*`, `*.exc`) into `ext-ref/lexicon/wordnet31/`. |
| `./bin/compose snapshot -chapter ch01 [-force]` | Freeze a local immutable snapshot of a chapter into `civ/notes/compose/snapshots/`. |
| `./bin/compose passes -chapter ch01 [-para 1]` | Print Backward-Pass goals, paragraph seams, and ranked structural candidates in the terminal. |
| `./bin/compose train` | Extract weighted pairwise preferences (including hard negatives from backtracked selections) from all decision trees and train `civ/notes/compose/model/ranker.safetensors`. |
| `./bin/compose diff -chapter ch01` | Show a line-by-line summary of changed sentences relative to the frozen snapshot. |
| `./bin/compose rebuild -chapter ch01 [-out PATH]` | Deterministically rebuild the chapter from its snapshot + decision tree into `civ/notes/compose/rebuilt/<chapter>.md`. |
| `./bin/compose help` | Display CLI flags and keyboard shortcuts. |

---

## Backward → Forward → Looping Seam Passes

Rebuilding prose works best when earlier sentences know what they are setting up:

1. **Backward Pass (`B`, Pass 0)** — Starts on the **landing (last) sentence** of each paragraph. Pin/refine the landing sentence and lock any "set phrases" (`LOCK`). `compose` propagates goals backward:
   * Concepts the next sentence opens with (**tail-to-next handoff**)
   * Distinctive words spent in the landing sentence (**reserved anti-spoiler budget**)
   * Rhythm setup beats when the next sentence is a short stressed punch
   * Semantic destination vector (`bge-small`)
2. **Forward Pass (`F`, Pass 1)** — Walks from sentence 1 to the end of the paragraph. Every structural graph option and WordNet word swap is scored bi-directionally against both your committed left context and the backward-pass right goal.
3. **Looping Seam Pass (`L`, Pass 2+)** — Re-evaluates the rebuilt paragraph as a whole, auto-marks clean sentences as **settled**, and jumps *only* to flagged seams (`broken_handoff`, `close_echo`, `reserved_clash`, `repeated_spine`, `landing_weak`, `tread_water`).

---

## GUI Views & Controls

Switch between all three graph views at any time using the top bar or keys `1`, `2`, `3`:

* **1. Lattice** *(default)* — Structural options strip across the top; sentence tokens/phrases arranged left-to-right as columns with WordNet alternatives fanned vertically below each unlocked content word.
* **2. Diagram** — Dependency syntax tree with curved Bezier arcs between heads and dependents; click any node to inspect its lexical alternatives in the drawer below, or double-click to lock/unlock it.
* **3. Hybrid** — Compact dependency syntax diagram above + horizontal word-option lattice below.

### Hover, Select Phrases, Scroll, Lock, Rewrite
* **Horizontal scrolling** — Use the **◀ / draggable scrollbar slider / ▶** right beneath the Word Lattice (and Dependency Diagram), or press **Left / Right** arrows when not typing.
* **Multi-word phrase reconsideration, WordNet sense chips, search & steering** — Click a word column header to start a phrase selection, then click a second header (or use **`◀ Extend` / `Extend ▶`**) to select a multi-word span such as `"stable living"` or `"generation to generation"`. Inside the **Phrase Bar**:
  * **WordNet Sense Toggle Chips** (`✓ word: gloss` / `× word: gloss`) let you turn off unwanted senses (e.g., disable `living: animation` while keeping `living: financial means / livelihood`).
  * **Phrase Search** supports 1850–1950 corpus wildcards (`steady *`, `* living`, `* livelihood`), WordNet concept/gloss search (`dependable livelihood`), and inline `+boost` / `-avoid` modifiers (`steady * -firm -animation`).
  * **Expand 2-Hop** widens the search using 2-hop WordNet neighbors, derivational (`+`) links, and period-attested reference-corpus collocations.
  * **Weights +/−** opens the dedicated **Boost / De-weight Chooser** where you can boost (`+`) or de-weight (`−`) words/phrases locally or book-wide (`weights.json`), and toggle **Etymology Bias (`OE+/Grk+/Lat−`)** to boost Anglo-Saxon (`[OE]`) and Greek (`[Grk]`) words/phrases while de-weighting Latinate (`[Lat]`) ones. Boosts steer ranking toward that direction (Rocchio SIMD cosine + lexical boost), and de-weights suppress that word's WordNet synset neighborhood and log high-weight negatives (`2.5×`) for the local NN.
* **Accept Paragraph (`Shift+A`)** — Seals the active paragraph as completed (`accept_paragraph` node in the decision tree, including paragraph cadence metrics and any intentionally overridden seam warnings), lights up a `✓` badge in the sidebar, and advances straight to the next paragraph's Backward Pass.
* **Hover** any structural card, phrase option, or word alternative to see the **completed sentence live inside its paragraph** plus score deltas (`Linear`, `Cadence Snap`, `1850–1950 Period Fit`, `Semantic Sim`, `NN`).
* **Click** a structural card, phrase option, or word option to select/deselect it and immediately recompose downstream scores and `a`/`an` agreement.
* **Lock ("set phrase")** — Double-click a word column header (or diagram node), select a multi-word span and click **Lock Phrase** / **Lock Global**, or add a phrase under **Global Set Phrases** in the right pane to protect it across the entire book.
* **Manual Rewrite** — Edit the text in the **Rewrite base** box and press **Use Rewrite as Base (`Ctrl+Enter`)** to re-diagram your rewrite and explore fresh structural/word options from it.

### Keyboard Shortcuts (when not typing in a text field)

| Key | Action |
| --- | --- |
| `1` / `2` / `3` | Switch view: Lattice / Diagram / Hybrid |
| `B` / `F` / `L` | Switch pass: Backward / Forward / Loop seams |
| `J` / `K` | Next / Previous sentence (in pass direction) |
| `N` / `P` | Next / Previous paragraph |
| `Left` / `Right` | Scroll Word Lattice or Diagram horizontally |
| `Esc` | Clear active multi-word phrase selection |
| `A` | Accept (or pin landing anchor in Backward Pass) and advance |
| `Shift+A` | Accept entire paragraph (`✓`) & advance to next paragraph |
| `S` | Keep As Is and advance |
| `R` | Reject pending changes and reset to base |
| `Ctrl+Enter` | Adopt Manual Rewrite as the new diagrammed base |

---

## State & Training Files (`civ/notes/compose/`)

* `snapshots/<chapter>.json` — Immutable frozen copy of the source chapter
* `decisions/<chapter>.jsonl` — Append-only decision tree (every shown option, 32-dim feature vector, hover dwell ≥250ms, backtrack, lock, rewrite, commit)
* `glossary_locks.json` — Global set phrases protected across the book
* `model/ranker.safetensors` — Locally trained pure-Go MLP preference ranker
* `rebuilt/<chapter>.md` — Deterministically rebuilt chapter output

# 38. Three Accepts Applied; the Book Built, Committed, and Pushed

Recorded 2026-11-24 (UTC). This covers the author's decisions on the three items of note 37: three accepts, two of them with notes. All three are now in the chapters. At the author's request the book was then built, which also releases it, and the changes of `civ/` were committed and pushed.

No source was added.

---

## 1. The decisions in numbers

| Disposition | Without a note | With a note | Total |
|---|---:|---:|---:|
| `accept` | 1 | 2 | 3 |
| `hold`, `reject`, `withdraw` | 0 | 0 | 0 |
| **Total** | | | **3** |

The three items were decided between 14:11 and 14:38 UTC on 2026-11-24.

## 2. What the notes ask for

Three more standing instructions follow from the two notes.

1. **Put a question in the plainest words.** "Ask the partisan what the program is before its years are counted." was "good direction, bad phrasing. Better: "ask what is expected before the experiment is run" or similar." A figure carried from the device of the passage (years counted, a ledger) is not to be pressed into the sentence that tells the reader what to do.
2. **A name is capitalized and quoted.** ""march against Islamophobia" should be capitalized and quoted."
3. **Say the book's view once, in one sentence, and do not argue it by comparison.** "Just say Islam is a religion and you can disagree with a religion. Don't include the sentence on Christianity."

## 3. Applied

Line numbers are those after all the edits.

**By the tool, 1 item** (`civ/bin/review apply`; response packet `packets/20261124-144114-applied.json`). `08_ch08.md:125` (`note/kindness-streamlined`), an accept without a note, located once at its recorded offset. The passage is the shorter, streamlined text of note 37.

**By hand, per the notes, 2 items** (response packet `packets/20261124-144847-applied-noted.json`). Each was an accept with a note, and the accepted text still carried the phrases the note changes. `civ/bin/review apply` skips a noted accept for that reason. Note 31 says to handle a noted item per the author's note and then record it applied, so that was done: the accepted text was applied with the changes of the note made in it, and nothing else.

- `14_ch14.md:19`, Tool 1 (`note/tool1-concrete`, alternate b, with God in the words of Chapter 3). Two changes, both the author's:
  - The last sentence of the first step is now "Ask what is expected before the experiment is run."
  - ", such as a date." is removed from the last sentence of the opening paragraph.
- `13_ch13.md:135`, Red-Green (`note/red-green-ends-with-iran`, alternate a, with the block quotation at the end). Three changes, all the author's:
  - The name is capitalized and quoted: `the "March against Islamophobia"`.
  - The sentence on Islam is the author's: "Islam is a religion, and you can disagree with a religion."
  - The sentence on Christianity is not included.

Each edit was made by a script that refused to write unless the sentence to be changed was found exactly once in the accepted text. `civ/bin/lint sim` was run on the two final texts before the write: 2 unique matches, 0 overlaps, 0 new Part A or spelling hits.

*Asymmetrical accounting* is kept in Tool 1. The note did not ask for it to go, and Tool 2 opens on it.

**Not applied, 1 item.** `02_ch02.md:a676ed11:prop/boiled-water-source`, as in notes 34 to 37.

## 4. Checks

1. **Kindness.** `civ/bin/review apply -dry-run`: 1 ready, 3 skipped for a note. `civ/bin/lint` (Parts A, C, D) before and after: the same findings, one of them moved by two lines (`08_ch08.md:142` to `:144`) because the passage grew. `civ/bin/linear check -file`: 209 sentences before and 226 after. The new sentences score 0.0 to 0.5. The sentence that scored 3.0 ("Because anonymous bureaucracies cannot personally verify ...") is gone.
2. **Tool 1 and Red-Green.** `civ/bin/lint` before and after: identical. `civ/bin/linear check -file`: `14_ch14.md` 146 sentences before and 180 after; `13_ch13.md` 124 before and 163 after. The new sentences that score 2 or more are the author's own words (the refrain on what the two sides build and pull down, the block quotation, the form of the sixth point) and one sentence in the shape of Chapter 3 ("For the ultimate questions that produce no measurable distinction in the observable world, ...").
3. **Quotations.** The final Tool 1 and Red-Green texts hold 9 quoted spans. 8 were matched by script against a saved copy of the source. The ninth is the name "March against Islamophobia", which is the English of the name the saved French sources use: RTL's headline reads "Marche contre l'islamophobie", and the franceinfo caption reads "La marche contre l'islamophobie arrive près de Gare du Nord à Paris, le 10 novembre 2019."
4. **Voice and cadence** (a reading aid only, note 34, section 8). Tool 1 +0.11 (the section scored +1.40 before this series of drafts); Red-Green +0.02 (it scored +2.56); Kindness -0.11. Snap 54.7 to 62.6; no treading-water flag.
5. The whitespace check of `git diff`: no error. Each chapter still ends with one newline.

## 5. The book

- The author built the book at 14:30 UTC, before the three applies of this round.
- It was built again at 14:50 UTC, after them, at the author's request: `fish civ/build.fish`. `civ/ccp_book.md` was assembled from the chapters (2,270 lines) and the PDF (211 pages), HTML, and EPUB were written to `civ/build/`.
- `build.fish` ends in `push_release`, and the release mount was present, so the new book and the other release files were copied to the site.
- The built HTML, PDF, and EPUB were searched: each holds the new Tool 1, Red-Green, and Kindness passages, and none holds "interlocutor", "Six-Point Single Ledger", "five operational mechanics", or the sentence on Christianity. There is no run-in list in the HTML.
- One "The Defense:" remains in the book, at `11_ch11.md:143` (section 1 of the parasitism passages). It was not decided; see section 7.

## 6. The queue now

285 items: 0 open, 29 on hold (all answered), 228 applied, 27 closed, 1 accept pending apply, 30 stale. The queue holds no open item. A next round needs a direction.

## 7. Left for the author

1. Withdraw the 30 stale items and `02_ch02.md:a676ed11:prop/boiled-water-source`. None of them can be applied; each quotes a passage that has been replaced.
2. Three passages that your notes of 2026-11-23 question are still as they were (note 36, section 11): the Generational Criterion (`13_ch13.md:5`), refusing authority (`09_ch09.md:75`), and Shinto (`12_ch12.md:141`, which needs research).
3. `11_ch11.md:143`: section 1 of the parasitism passages still has its Defense line. Say if it goes.
4. `ext-ref/tool1_hermetic/dl/festinger_riecken_schachter_1956_when_prophecy_fails.pdf` is probably an unauthorized upload, kept only to check short quotations. Delete it if it is not wanted.
5. Say what the next round is: one of the three passages above, another pass with the tools, or something else.
6. The changes outside `civ/` in the repository (`book1-4` to `book1-7`, `tts`, and some untracked files) were not touched and are not committed.

# 42. Nine Accepts Applied; the Book Built, Released, Committed, and Pushed; the Sheet Read by Two Tools

Recorded 2026-11-25 (UTC). This covers the author's instruction after note 41: "Review done. Apply reviewed items. build book "fish build.fish", commit civ and push. Then run the linear and ngram report on ccp_sheet.md and if you think is warranted, give me review suggestions for the sheet." All nine items are applied, the book is built and released, the round is committed and pushed, the two reports are run on the sheet, and seven items are filed on it. Chapter text changed in five files: `01_ch01.md`, `04_ch04.md`, `10_ch10.md`, `13_ch13.md`, and `14_ch14.md`. `civ/ccp_sheet.md` and `civ/notes/06_editing_guide.md` were not touched. No tool code was changed.

No source was added.

---

## 1. The decisions in numbers

| Disposition | Without a note | With a note | Total |
|---|---:|---:|---:|
| `accept` | 9 | 0 | 9 |
| `hold`, `reject`, `withdraw` | 0 | 0 | 0 |
| **Total** | | | **9** |

The nine items were decided between 00:51 and 00:52 UTC on 2026-11-25.

## 2. Applied

**By the tool, 9 items** (`civ/bin/review apply`; response packet `packets/20261125-005544-applied.json`). Each was an accept without a note, located once at its recorded offset.

- `01_ch01.md:9`: "wet" is dropped. A newborn on the bare ground is "naked and toothless".
- `13_ch13.md:5`, the Generational Criterion, alternate (a). The passage now reads:

  > The first is **the Generational Criterion (80 years)**. Never evaluate a policy by the intent behind it or by its immediate sentiment during a period of borrowed expenditure or revolutionary enthusiasm. A result comes from the policy, from the country it was enacted in, and from the people who carried it out. The three are not judged apart. Evaluate the policy by the condition of the country when the grandchildren inherit the resulting institutions and balance sheet eighty years later.

- Eighty years, in seven places:
  - `01_ch01.md:41`: "Within eighty years".
  - `04_ch04.md:61`: "roughly eighty years in this life, some twenty-nine thousand days".
  - `10_ch10.md:9` to `:21`: four places, so the story of the city ends at Year 80.
  - `10_ch10.md:34`: "Within roughly eighty years, almost every human being".
  - `13_ch13.md:67`: "the eighty-year replacement horizon".
  - `14_ch14.md:85`: "followed this rule for eighty years".
  - `14_ch14.md:133`: "eighty years hence".

No seventy and no range remains for the horizon. The seventies left in the chapters are another subject: Lamech's (`09_ch09.md:249`), seventy million (`:309`), an electrician's age (`10_ch10.md:17`), and seventy-three seconds (`11_ch11.md:194`).

## 3. Checks before and after applying

1. `civ/bin/review apply -dry-run`: 9 ready, 0 skipped.
2. `civ/bin/lint` (Parts A, C, and D) before and after: identical.
3. The whitespace check of `git diff`: no error. Each chapter ends with one newline.
4. `civ/bin/linear check` was not run on the edited chapters, as in notes 39 to 41.

## 4. The book

- Built at 00:56 UTC with `fish civ/build.fish`, as in note 38. `civ/ccp_book.md` was assembled from the chapters (2,262 lines). The PDF has 209 pages, against 211 in note 38. The HTML and the EPUB were written to `civ/build/`.
- The release mount was present, so `push_release` copied 49 files to the site.
- The assembled Markdown, the HTML, the text of the PDF, and the EPUB were searched. Each holds the ten new phrases (among them "the Generational Criterion (80 years)", "The three are not judged apart", "naked and toothless", and "almost every human being"). None holds the eleven old (among them "The Defense:", "Shinto", "most unscrupulous", "first eighteen months", and "seventy or eighty").
- `civ/build/` is not tracked. `civ/ccp_book.md` is.

## 5. Commit and push

`01a3d2d`, at 00:57 UTC: "civ: apply the author's review decisions of 2026-11-25 and rebuild the book". It holds 20 files under `civ/`: eight chapters, `ccp_book.md`, `decisions.json`, notes 39 to 41, and seven packets. It was pushed: `2d5b133..01a3d2d`, master to master. The changes outside `civ/` were not touched.

## 6. The two reports on the sheet

**Linear.** `civ/bin/linear check -file civ/ccp_sheet.md` scores one Markdown file and prints every sentence. The sheet has 373 sentences. 257 score 0, and 68 score 2 or more. The summary:

| Measure | Sheet | Book | Period sample |
|---|---:|---:|---:|
| Close-range echoes per 100 sentences | 59.5 | 49.4 | 55.1 |
| It chains per 100 sentences | 5.9 | 0.4 | 0.9 |
| Ambiguous "it" per 100 sentences | 4.0 | 2.2 | 4.2 |
| Abstract subjects per 100 sentences | 0.3 | 0.6 | 0.5 |

The book and the period columns are from note 18, which predates the later edits. The tool wrote its three cache files under `ext-ref/linear/scratch/`.

**Ngram.** The tool reads `civ/chapters/*.md` from the repository root and takes no file. It was run unmodified from a shadow tree in `/tmp/opencode/ngram_sheet/`: a copy of the sheet as the only chapter, with copies of the reference list, the editing guide, and the ngram data. `google-phrases` looked up the 76 phrases of the sheet that the book had not used (183 were cached). The lookups went to Google Books and were written to the shadow copy of the cache. The real `ext-ref/ngram/` and note 17 were not touched.

What it shows for the sheet, 6,426 words in 372 sentences:

- **Words rare in print of 1850 to 1950.** Used twice or more: infrastructure (4), foundational (3), tiered (3), falsifies (2), scoped (2), scopes (2). Used once: swappable (0.0 per million words), prioritizing, prioritize, borderless.
- **Phrases absent from that print.** Six, and all of them are the form of one field: "what it cannot accept" (27), "it cannot accept the belief" (5), "cannot accept the belief that" (5), "it also cannot accept" (4), "it cannot accept treating" (4), and "cannot accept treating" (5).
- **"Cannot accept"** appears 79 times, at 364 times its rate in the period. "It cannot accept" appears 75 times.
- **Sentences** average 17.3 words, against 28.8 in the period sample. That is the sheet's brevity and not a fault.
- **Sections farthest from period prose** (farther than 95 percent of stretches of the same length): "Human life continuing" (99 percent), "1. Align with truth" (99), "Self-anchored" (96), and "Scoped" (96). A distance from the period is not by itself a fault (the report's own limit).

Both reports are kept beside the book's: `civ/notes/18a_linearity_check_sheet.txt` and `civ/notes/17a_ngram_comparison_sheet.md`. The header of the second says "chapters 1 to 14" because the tool's wording is fixed. The file holds the sheet alone.

## 7. The sheet: seven items filed

| Packet | Contents |
|---|---|
| `packets/20261125-010730-sheet-review.json` | 7 open items |

| Line | Section | Rule | Alternate (a) |
|---|---|---|---|
| 60 | Reality is real | `sheet/form-cannot-accept` | The field opens on the thing. Alternate (b) changes the label and keeps the sentences |
| 136 | Kept and built | `sheet/rule-10-split` | The colon after "every level of wealth" becomes a full stop |
| 136 | Kept and built | `sheet/rule-10-split` | The semicolon before "it is a piece of civilization" becomes a full stop |
| 170 | Outward | `sheet/rule-10-split` | The semicolon before "during an invasion" becomes a full stop |
| 235 | Civilization as the root technology | `sheet/rule-10-split` | The colon after "in the same sense" becomes a full stop. The book has the split (`01_ch01.md:31`) |
| 270 | Admire beauty | `sheet/rule-10-verb` | "cannot" becomes "cannot be beautiful" |
| 113 | Particular | `sheet/register` | "swappable" becomes "interchangeable", the book's word (`01_ch01.md:64`) |

**The form of the field.** The sheet has 27 fields that open "What it cannot accept: It cannot accept", so the phrase is said twice in a row. Both reports point at it before anything else. The item shows the first field as a sample. If the author takes (a) or (b), the other 26 are filed, each as its own item. Three or four need a hand, because a "cannot accept" sentence follows an explanatory one, as at line 274.

`civ/bin/review add -strict` passed, and `civ/bin/lint sim` reports 7 edits, 0 overlaps, 0 new Part A or spelling hits for (a), and 2 edits for (b). No item touches a locked line (13 to 21 and 35 to 47).

## 8. What was left alone, and why

- **26 flagged sentences of the field form.** They are one question about form, and one sample item carries it.
- **Deliberate parallels.** Among them line 5 ("A stable life, a stable farm, a stable house, and tools are technology"), line 31 ("It keeps enemies away and its people fed. It keeps individuals alive and building."), line 56 ("does not create a new reality; it creates a narrative"), and line 270 ("Inside a tale the author decides the consequences; in reality they come whether anyone decides them or not").
- **Close echoes.** A sheet that defines terms repeats them. The sheet's rate is 10 per 100 sentences above the book's and 4 above the period's.
- **"Foundational" and "infrastructure".** Both are rare in the period, and the book keeps them (26 and 4 uses), so no item is filed.
- **"Tiered", "scoped", and "scopes".** They are the book's terms.
- **Two sentences inside "cannot accept" fields** that need a split or a named noun: line 124 (54 words joined by "nor") and line 295 (40 words). They wait for the decision on the form, because they would be filed again in the new form.

## 9. The queue now

305 items: 7 open, 0 on hold, 241 applied, 57 closed, 0 stale.

## 10. Left for the author

1. Decide the 7 open items on the sheet in the viewer.
2. Uncommitted under `civ/`: this note, the two reports (17a and 18a), and the sheet's packet.
3. The reports on the book. Notes 17 and 18 still predate the changes to 11 chapters. You asked for the sheet alone, so they were not run again.
4. Still open from note 41, section 6: the probably unauthorized uploads under `ext-ref/`; the changes outside `civ/`, which are not committed; whether to change the cadence scoring; and Chapter 11's two mechanisms without a case on record.

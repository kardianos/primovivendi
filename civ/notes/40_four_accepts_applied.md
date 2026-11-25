# 40. Four Accepts Applied: the Last Defense Line, Refusing Authority, Shinto Removed, and the Criterion at Eighty Years

Recorded 2026-11-25 (UTC). This covers the author's four decisions on the items of note 39: four accepts, one of them with a note. All four are applied. Chapter text changed in four files: `09_ch09.md`, `11_ch11.md`, `12_ch12.md`, and `13_ch13.md`. `civ/ccp_book.md`, `civ/build/`, `civ/ccp_sheet.md`, and `civ/notes/06_editing_guide.md` were not touched, so the built book still holds the text before this round. `civ/build.fish` was not run. No tool code was changed. Nothing is committed.

No source was added.

---

## 1. The decisions in numbers

| Disposition | Without a note | With a note | Total |
|---|---:|---:|---:|
| `accept` | 3 | 1 | 4 |
| `hold`, `reject`, `withdraw` | 0 | 0 | 0 |
| **Total** | | | **4** |

The four items were decided between 00:17 and 00:23 UTC on 2026-11-25.

## 2. What the notes ask for

One note, and two instructions given in conversation after the decisions. (Spellings are the author's.)

1. **One span, in a round number.** "Make them both be just 80 years. That's close enough people get it." Where a passage gives one span twice, it is given once, in a number a reader holds in mind.
2. "Give me context of '18mo'." Answered in section 6.
3. "Also in first chapter, remove wet from born naked, like we did with another entry." Filed as an item in section 5.

## 3. Applied

Line numbers are those after all the edits.

**By the tool, 3 items** (`civ/bin/review apply`; response packet `packets/20261125-002751-applied.json`). Each was an accept without a note, located once at its recorded offset.

- `11_ch11.md`, the Defense paragraph that followed line 141: deleted. The book has no "The Defense:" line now. Section 1 of the parasitism passages ends with the sentence on the children formed against the household that raised them.
- `09_ch09.md:75`: the two claims about the worst are removed. The vacancy is taken by "someone else", and the paragraph ends "it is a dereliction of duty." You chose alternate (a), so the paragraph makes no claim about who takes the vacancy. The sentences on priority in (b) and (c) were not taken.
- `12_ch12.md`, the section that followed line 139: the Shinto section is removed, heading and two paragraphs. Shinto is named in no chapter and not in the sheet.

**By hand, per the note, 1 item** (response packet `packets/20261125-002817-applied-noted.json`). `13_ch13.md:5` (`note/generational-criterion`, alternate a). `civ/bin/review apply` skips a noted accept. The accepted text was applied with the change of the note made in it, and nothing else: the name reads **the Generational Criterion (80 years)**, and the sentence ends "eighty years later". "Eighteen months" is unchanged, since the note did not mention it. The script refused to write unless the passage was found exactly once.

## 4. Checks before and after applying

1. `civ/bin/review apply -dry-run`: 3 ready, 1 skipped for a note, no overlap.
2. `civ/bin/lint sim` on the final text of the noted item, in memory, before the write: 1 edit, 0 overlaps, 0 new Part A or spelling hits.
3. `civ/bin/lint` (Parts A, C, and D) before and after: identical.
4. The whitespace check of `git diff`: no error. Each chapter ends with one newline.
5. `civ/bin/linear check` was not run, as in note 39.

## 5. New items

| Packet | Contents |
|---|---|
| `packets/20261125-002904-followups-eighty-and-wet.json` | 3 open items |

| Passage | Why it is filed | Alternate (a) |
|---|---|---|
| `01_ch01.md:9` | Your instruction in conversation. The same change was made at `01_ch01.md:41` on 2026-11-23 | "A newborn on the bare ground is naked and toothless." |
| `13_ch13.md:67` | The chapter now gives the horizon as eighty years at line 5 and as seventy here | "the eighty-year replacement horizon" |
| `14_ch14.md:85` | The seventh question of the tool gives the horizon as seventy years | "followed this rule for eighty years" |

The name of the seventh question, "The generational horizon", is not touched (note 09, line 77).

Two other passages give the length of a human life, not the horizon, and both already contain eighty: "Within seventy or eighty years" (`01_ch01.md:41`) and "Within roughly eighty to a hundred years" (`10_ch10.md:34`). They are left alone.

`civ/bin/review add -strict` passed, and `civ/bin/lint sim` reports 3 edits, 0 overlaps, 0 new Part A or spelling hits.

## 6. The eighteen months: its context

The sentence at `13_ch13.md:5` reads: "Never evaluate a policy by its immediate sentiment during the first eighteen months of borrowed expenditure or revolutionary enthusiasm."

1. **Where it came from.** The figure is in the first commit of the book, 2026-10-25 (`1702e3c`): "Never grade a policy by how it feels in the first eighteen months of borrowed money or revolutionary euphoria." It has stood through every rewrite since, when the bullets were turned into prose on 2026-11-22 and in the linearity edits.
2. **It is not in the notes or the packs.** The three lenses came from `book1-6` (`chapters/1800_ending_arc.md`) and are mined in `civ/notes/04_book1_6.md`, section 4. There the generational lens is the granary: a son mocks the rule and still eats well for years from bins his father filled, and his daughter inherits the mockery and an empty bin. It has no number. "Eighteen months" and "70–100 years" were the draft's own.
3. **What it does in the sentence.** It is the short end of a contrast. A policy feels good while borrowed money is still being spent or revolutionary enthusiasm still runs. It is judged by the condition of the country when the grandchildren inherit.
4. **Nothing rests on it.** No source is saved for it. No case in the chapter turns on eighteen months: the cases run 25 years (Singapore), 26 (Russia), and 5 (Iran). The chapter's own figures for the early years of a policy are Stolypin's "twenty years of internal and external peace", of which "Russia received only seven".
5. **The other "eighteen months" in the book is another subject**: "An engineering state builds ships in eighteen months" (`11_ch11.md:208`).

I did not look outside the book for a referent, since you asked that research not search. The figure is not mine to keep or remove. If it stays, it needs your reason or a source. If it goes, the closed item's third alternate had "the first months" in its place, and that can be filed as a new item.

## 7. The queue now

292 items: 3 open, 0 on hold, 232 applied, 57 closed, 0 stale.

## 8. Left for the author

1. Decide the 3 open items in the viewer.
2. "Eighteen months": say whether it stays, and on what ground.
3. The book was not built. `civ/ccp_book.md` and `civ/build/` hold the text before this round.
4. Commit when ready. Uncommitted: four chapters, `decisions.json`, notes 39 and 40, and five packet files (two of note 39 and three of this round).
5. The texts read for Shinto are still in the temporary folder of note 39. With the section removed they are no longer needed for the book.

# 43. Seven Sheet Items Applied; the Other Twenty-Six Fields Filed

Recorded 2026-11-25 (UTC). This covers the author's decisions on the seven items of note 42 and what followed from them. The form of the "cannot accept" field was accepted on the sample, so the other 26 fields are now filed in that form, with one sample item on "Point N". `civ/ccp_sheet.md` changed in seven places. No chapter changed. The book was not built, and nothing is committed.

No source was added.

---

## 1. The decisions in numbers

| Disposition | Without a note | With a note | Total |
|---|---:|---:|---:|
| `accept` | 7 | 0 | 7 |
| `hold`, `reject`, `withdraw` | 0 | 0 | 0 |
| **Total** | | | **7** |

The seven items were decided between 01:11 and 01:12 UTC on 2026-11-25.

## 2. Applied

**By the tool, 7 items** (`civ/bin/review apply`; response packet `packets/20261125-011421-applied.json`). Each was an accept without a note, located once at its recorded offset.

- `ccp_sheet.md:60`: the first "cannot accept" field opens on the thing.
- `ccp_sheet.md:136`: two splits, at the colon and at the semicolon.
- `ccp_sheet.md:170`: the split at the semicolon.
- `ccp_sheet.md:235`: the split at the colon.
- `ccp_sheet.md:270`: "cannot" is written out as "cannot be beautiful".
- `ccp_sheet.md:113`: "swappable" is "interchangeable".

## 3. Checks before and after applying

1. `civ/bin/review apply -dry-run`: 7 ready, 0 skipped.
2. `civ/bin/lint` (Parts A, C, and D) before and after: identical.
3. The whitespace check of `git diff`: no error. The sheet ends with one newline.

The sheet's products (`civ/build/ccp_sheet.*`) and the copy on the site are older than `ccp_sheet.md` now. `build.fish` rebuilds them when the Markdown is newer.

## 4. New items

| Packet | Contents |
|---|---|
| `packets/20261125-011827-sheet-form.json` | 27 open items |

**The other 26 fields.** Each item is one field, with the change the author accepted on the sample. The opening words of every "cannot accept" sentence are removed, and the next word is capitalized: "It cannot accept", "It also cannot accept", "Nor can it accept", and the "it cannot accept" that follows "Conversely" and "Equally". "Conversely" and "Equally" stay. Nothing else changes, with four exceptions.

1. **Line 81.** A "that" is added after "claim". Without it, the last item ("Altering a hypothesis after the fact to claim a failed prediction was actually correct.") can be read as a sentence that says the opposite.
2. **Line 124.** The two items were one sentence of 54 words joined by "nor". They are two items now.
3. **Line 140, alternate (b).** The third item moves up beside the other two, so that the two examples follow the items they explain.
4. **Line 274.** A full sentence stands before the last item. Alternate (a) opens that item on "Also". Alternate (b) changes the first item and leaves the last sentence as it is.

Line 74 also has an alternate (b), which makes the three beliefs one sentence.

The changes were checked by a token comparison of each field against its replacement: every difference is one of those removals with a capital letter, or one of the exceptions above.

**"Point N" against "Principle N".** This came from reading the sheet, not from a tool. The sheet refers to the principles as "Point N" twelve times, in nine sentences, and never as "Principle N". The book refers to them as "Principle N" 137 times and never as "Point N", and the editing guide fixes that form (Part A). One sample item is filed, on line 120. If it is accepted, the other eleven are filed, each in its sentence.

`civ/bin/review add -strict` passed, and `civ/bin/lint sim` reports 27 edits, 0 overlaps, 0 new Part A or spelling hits for alternate (a), and 3 edits for (b). No item touches a locked line (13 to 21 and 35 to 47).

## 5. What the author asked, and what was left alone

The author asked whether any other sheet items were to be brought up, and whether I wanted to float further adjustments from the linear readings.

- **Other items:** yes, the two kinds in section 4.
- **Further adjustments from the linear readings:** no. The sentences still flagged are deliberate parallels, or terms that a definition sheet repeats.

Four lines of the sheet end with a stray space (134, 195, 197, and 210). No item is filed for them. The whitespace check of `git diff` will flag one if a later edit touches its line.

## 6. The queue now

332 items: 27 open, 0 on hold, 248 applied, 57 closed, 0 stale.

## 7. Left for the author

1. Decide the 27 open items in the viewer.
2. If "Principle" is accepted on the sample, say so and the other eleven are filed.
3. Uncommitted under `civ/`: notes 42 and 43, the two reports (17a and 18a), and two packets.
4. The sheet's products and the site copy are rebuilt by the next `fish civ/build.fish`.

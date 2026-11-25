# 46. The Stray Item Withdrawn on Instruction; the Sheet Rebuilt and Released; the Round Committed and Pushed

Recorded 2026-11-25 (UTC). This covers the author's instruction after note 45: "There is no search, hard to find. Please mark it as withdrawn yourself. Then build, commit and push." The item of note 45, section 4 is withdrawn, the sheet is rebuilt and released, and the changes under `civ/` are committed and pushed in this round. No chapter changed. No tool code was changed.

No source was added.

---

## 1. The one decision, made on instruction

`call:civ/ccp_sheet.md:883c4823:sheet/form-cannot-accept` is `withdraw`, at 03:28:53 UTC.

- **It was not decided in the viewer.** The author asked in conversation that it be marked withdrawn, because the viewer has no search and the item was hard to find.
- **This is the one write to `decisions.json` made outside the viewer.** Note 31 tells agents never to edit that file. The write was made as the author asked, and as narrowly as it could be. One entry changed, from `accept` with its accepted text to the three fields the viewer writes for a withdrawal: disposition, alternate, and time. Every other byte of the file is as it was, and the count of items did not change (342).
- **No words were put in the author's note field.** The reason is recorded here.
- `civ/bin/review check` now reads 342 items, 0 pending apply, 258 applied, 84 closed. `civ/bin/review apply -dry-run` would apply nothing.

## 2. What the note asks for

1. **A search in the viewer.** "There is no search, hard to find." The terminal screen and the window have filters and `j` and `k`, and no way to find an item by its text or its id. Not built. Say so and it will be.
2. **An item may be closed on instruction** when the viewer makes it hard to reach. This was done once, in the form of section 1, and it should stay rare.

## 3. The build

- `fish civ/build.fish` at 03:29 UTC. The sheet was rebuilt in PDF (21 pages), HTML, and EPUB, because `ccp_sheet.md` was newer than its products. The book was up to date and was not rebuilt.
- The release mount was present, so `push_release` copied 49 files to the site. Among them were the sheet's three.
- The built HTML, the text of the PDF, and the EPUB were searched. Each holds the new phrases ("belongs under Principle 4", "These principles require a person to defend himself and his own", "identical, interchangeable tokens", "every level of wealth. Sweeping a floor"). None holds "Point 1" to "Point 5", "That refusal stands against", or "swappable".
- `civ/build/` is not tracked.

## 4. Commit and push

The commit that holds this note also holds the rest of notes 42 to 46: `civ/ccp_sheet.md`, `decisions.json`, the two reports on the sheet (17a and 18a), and seven packets. It was pushed to `origin/master`. The changes outside `civ/` were not touched.

## 5. The queue now

342 items: 0 open, 0 on hold, 258 applied, 84 closed, 0 stale, 0 pending.

## 6. Left for the author

1. A search in the viewer, if you want one.
2. Still open from note 42, section 10: the reports on the book (notes 17 and 18 predate the changes to 11 chapters); the probably unauthorized uploads under `ext-ref/`; the changes outside `civ/`, which are not committed; whether to change the cadence scoring; and Chapter 11's two mechanisms without a case on record.
3. Two limits of the review tool were found in these rounds and not changed. An accept whose effect a later accept undoes comes back as pending (note 45, section 4). The viewer has no search (section 2).

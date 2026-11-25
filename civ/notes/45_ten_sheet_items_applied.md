# 45. Ten Sheet Items Applied; the Sheet Says "Principle" Throughout; One Item Has Come Back as Pending

Recorded 2026-11-25 (UTC). This covers the author's decisions on the 36 open items of note 44. The 26 items on the other fields are rejected, so the doubled opening stays in all 27 fields. The other ten are accepted and applied: eight on "Point N", one that puts the sample field back, and one on the double negative. `civ/ccp_sheet.md` changed. No chapter changed. The book was not built, and nothing is committed.

One item needs the author's hand. It is described in section 4: the conversion of line 60, which the author has twice decided against, now counts as pending, and a plain `civ/bin/review apply` would apply it again.

No source was added.

---

## 1. The decisions in numbers

| Disposition | Without a note | With a note | Total |
|---|---:|---:|---:|
| `accept` | 10 | 0 | 10 |
| `reject` | 26 | 0 | 26 |
| `hold`, `withdraw` | 0 | 0 | 0 |
| **Total** | | | **36** |

The 36 items were decided between 03:17 and 03:18 UTC on 2026-11-25.

## 2. Applied

**By the tool, 10 items** (`civ/bin/review apply`; response packet `packets/20261125-031935-applied.json`). Each was an accept without a note, located once at its recorded offset.

- **"Point N" becomes "Principle N"** in the eleven places that were left, on lines 165, 212, 219, 249, 256, 263, 277, and 284. With the sample of note 44, the sheet now says "Principle N" twelve times and "Point N" nowhere, as the book does.
- **`ccp_sheet.md:60` is back in the form of the other 26.** All 27 fields open "What it cannot accept: It cannot accept".
- **`ccp_sheet.md:185`.** "That refusal stands against these principles." is replaced by "These principles require a person to defend himself and his own."

## 3. Checks before and after applying

1. `civ/bin/review apply -dry-run`: 10 ready, 0 skipped.
2. `civ/bin/lint` (Parts A, C, and D) before and after: identical.
3. The whitespace check of `git diff`: no error. The sheet ends with one newline.

## 4. One item has come back as pending

The item is `call:civ/ccp_sheet.md:883c4823:sheet/form-cannot-accept`, the conversion of line 60. The author accepted it at 01:11 UTC, and it was applied (note 43). The author then kept the old form (note 44) and accepted the item that put line 60 back, which was applied in this round.

The tool marks an item applied only while the file holds its replacement. The replacement has left the file, and the item's own quote is in the file again at its recorded offset. So the item counts as "accept, pending apply" a second time, and a plain `civ/bin/review apply` would apply it. That would convert line 60 again and undo the decision of 03:18.

- **Until the item is closed, `civ/bin/review apply` is not run without `-id`.** The flag applies only the named items.
- **The author can close it in the viewer:** the Decided filter, the item, then `w` or `r`. Those two write on a decided item.
- **The cause is in the tool.** An accept whose effect a later accept undoes is not closed by the later decision. A fix would have the later apply record that it replaced the earlier item's replacement. No tool code was changed.

## 5. The queue now

342 items: 0 open, 0 on hold, 258 applied, 83 closed, and 1 accept pending apply, which is the item of section 4.

## 6. Left for the author

1. Close the item of section 4 in the viewer.
2. Build, commit, and push when ready. Uncommitted under `civ/`: `ccp_sheet.md`, `decisions.json`, notes 42 to 45, the two reports (17a and 18a), and seven packets. The sheet's products and the site copy are older than `ccp_sheet.md`, and `build.fish` rebuilds them.
3. Still open from note 42, section 10: the reports on the book (notes 17 and 18); the probably unauthorized uploads under `ext-ref/`; the changes outside `civ/`; whether to change the cadence scoring; and Chapter 11's two mechanisms without a case on record.

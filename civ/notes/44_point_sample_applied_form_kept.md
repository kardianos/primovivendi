# 44. The "Point" Sample Applied; the Field Form Kept; the Double Negatives Looked Into

Recorded 2026-11-25 (UTC). This covers the author's decision on one item of note 43 and two notes given in conversation: the other "cannot accept" fields stay as they are, and the double negatives are to be looked into. `civ/ccp_sheet.md` changed in one place. No chapter changed. The book was not built, and nothing is committed.

No source was added.

---

## 1. The decisions in numbers

| Disposition | Count |
|---|---:|
| `accept` | 1 |
| `hold`, `reject`, `withdraw` | 0 |
| Still open from note 43 | 26 |

The one item was decided at 01:23 UTC on 2026-11-25. The 26 items on the other fields were not decided in the viewer. Your note in conversation keeps the form, so they are to be withdrawn there (section 6).

## 2. What the notes ask for

1. **The doubled opening stays in the sheet.** "The other "It cannot accept" might stay. Normally in prose, this would be good to remove. But I think for now we will keep in this instance."
2. **Double negatives are to be looked into.** "We might want to look into double negatives, such as in [the field at line 185] where we are talk about the refusal right aftwarwards."

(Spellings are the author's.)

## 3. Applied

**By the tool, 1 item** (`civ/bin/review apply`; response packet `packets/20261125-030909-applied.json`). `ccp_sheet.md:120`, `sheet/point-or-principle`, an accept without a note, located once at its recorded offset. The sentence now ends "it belongs under Principle 4."

`civ/bin/review apply -dry-run` showed 1 ready. `civ/bin/lint` (Parts A, C, and D) is identical before and after, and the whitespace check of `git diff` shows no error.

## 4. New items

| Packet | Contents |
|---|---|
| `packets/20261125-031250-sheet-point-form-double.json` | 10 open items |

| Lines | Rule | What |
|---|---|---|
| 165, 212, 219, 249, 256, 263, 277, 284 | `sheet/point-or-principle` | The other eleven "Point N", one item to a line. Three lines hold two each |
| 60 | `sheet/form-keep` | The field returns to the text it had before. It is now the only one of the 27 in the new form |
| 185 | `sheet/double-negative` | (a) The second sentence states what the principles require. (b) The second sentence is removed |

The packet was filed first at 03:12 UTC and re-filed at 03:12:50 UTC with the problem text of the third item corrected. The item identifiers did not change.

`civ/bin/review add -strict` passed, and `civ/bin/lint sim` reports 10 edits, 0 overlaps, 0 new Part A or spelling hits for alternate (a), and 1 edit for (b).

## 5. The double negatives: what the look found

I searched the sheet for sentences that hold two negative words and read each one.

1. **The pattern the author named occurs once, at line 185.** A negative is the thing the sentence rejects ("a refusal to use violence to protect oneself and one's own"), and the next sentence refers back to it ("That refusal stands against these principles"). Three negatives are stacked there: cannot accept, a refusal, and stands against.
2. **One other sentence rejects a negative outright**, "cannot accept neglect" (line 140). Nothing refers back to it, and "neglect" is a plain name of a fault.
3. **The other places where two negative words meet are of three kinds**, and I found no harm in them:
   - the label and the sentence that both say "cannot accept" (27 fields, kept by the author's note);
   - a negative inside a description of the thing rejected, as "a theory that is internally consistent on paper but fails when tested against the physical world" (line 81), or "elders who will not form those younger than themselves" (line 295);
   - parallel clauses, as "Shortage of money does not license a lie, and owing a stranger the truth does not give him the child's claim on the household" (line 161).

In all, 57 sentences of the sheet hold two or more negative words. Two more places refer back to rejected items right after naming them: "a case of the first" and "a case of the second" (line 140), which point to "building without keeping" and "keeping without building". They read plainly, and no item is filed for them.

## 6. The queue now

342 items: 36 open, 0 on hold, 249 applied, 57 closed, 0 stale. Of the 36 open, 26 are the items on the other fields, which the author's note keeps as they are.

## 7. Left for the author

1. Decide the 10 new items in the viewer.
2. Withdraw the 26 items on the other fields in the viewer (the Open filter, then `w` on each), so that the queue shows what is really open. An agent may not write `decisions.json`.
3. Uncommitted under `civ/`: `ccp_sheet.md`, `decisions.json`, notes 42 to 44, the two reports (17a and 18a), and six packet files (one of them the superseded first filing of 03:12).
4. The sheet's products and the site copy are older than `ccp_sheet.md`. They are rebuilt by the next `fish civ/build.fish`.

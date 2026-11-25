# 41. The Criterion on Results, and Eighty Years in Every Place: Nine Items Filed

Recorded 2026-11-25 (UTC). This covers the author's two instructions in conversation after the applies of note 40: the Generational Criterion is to give no timeline and to judge by results and not by intent, and the generational span is to be eighty years in every place. Nine items are open in the viewer: six filed here and three of note 40. No chapter text changed, and no item was decided. `civ/ccp_book.md`, `civ/build/`, `civ/ccp_sheet.md`, and `civ/notes/06_editing_guide.md` were not touched. `civ/build.fish` was not run. No tool code was changed. Nothing is committed.

No source was added.

---

## 1. What the notes ask for

The author's instruction of 2026-11-25, in conversation. (Spellings are the author's.)

> "regarding 18 month. We can change that slightly and make it more precise. Don't give a timeline. The goal is to judge based on the results, not the intent. You may have voted to bring a high speed rail in, but if all money is spent and no usable track is produced in any reasonable timeframe, then that is the result not only of the law, but also of the environment it is enacted in and who is actually implementing it. That needs to be re-worded, but as you can see, giving a timeline may be more harm then good in understanding the maning. RE 70. Make everything just say 80 flat as a generational placeholder, no 70 or 70-100. -- After you apply these (or make new items for me to review), tell me what else is outstanding."

Two standing instructions follow.

1. **No timeline where a number would carry the meaning.** A policy is judged by its results and not by its intent. A result is produced by the policy, by the country it is enacted in, and by the people who carry it out, and the three are not judged apart.
2. **One number for the generational span.** Eighty years, flat. No seventy, and no range.

## 2. The Generational Criterion: three drafts

The passage as applied on 2026-11-25 (note 40):

> The first is **the Generational Criterion (80 years)**. Never evaluate a policy by its immediate sentiment during the first eighteen months of borrowed expenditure or revolutionary enthusiasm. Evaluate it by the condition of the country when the grandchildren inherit the resulting institutions and balance sheet eighty years later.

Item `call:civ/chapters/13_ch13.md:fcaf7936:note/results-not-intent`.

**Alternate (a), the slightest change, that the author may judge it without the viewer:**

> The first is **the Generational Criterion (80 years)**. Never evaluate a policy by the intent behind it or by its immediate sentiment during a period of borrowed expenditure or revolutionary enthusiasm. A result comes from the policy, from the country it was enacted in, and from the people who carried it out. The three are not judged apart. Evaluate the policy by the condition of the country when the grandchildren inherit the resulting institutions and balance sheet eighty years later.

**Alternate (b)** makes the middle one statement: "Never evaluate a policy by its intent." The clause on sentiment is gone, and the two sentences on what a result includes follow it.

**Alternate (c), with the author's example:**

> The first is **the Generational Criterion (80 years)**. Never evaluate a policy by its intent. You may have voted for high-speed rail. If the money is spent and no usable track exists, that is the result of the policy, of the country the policy was enacted in, and of the people who carried it out. The three are not judged apart. Evaluate the policy by the condition of the country when the grandchildren inherit the resulting institutions and balance sheet eighty years later.

- In every alternate the name and the last sentence are as applied, except that the last sentence now says "Evaluate the policy" where it said "Evaluate it". Two sentences with other nouns now stand before it.
- "The three are not judged apart" is my sentence. It states what follows from the author's example: a defender of a policy cannot set the result aside by blaming the country or the people who carried it out.
- The example in (c) is a condition and not a case, so it needs no source. It gives no time, as instructed. "In any reasonable timeframe" is left out with the rest of the timeline.
- The example ends on the idea it gives, as the author's standing instruction asks (note 36, section 2).

## 3. Eighty years: where the span is given

| Passage | As printed now | Item |
|---|---|---|
| `13_ch13.md:5` | "the Generational Criterion (80 years)", "eighty years later" | Applied (note 40) |
| `13_ch13.md:67` | "the seventy-year replacement horizon" | Open, filed in note 40 |
| `14_ch14.md:85` | "followed this rule for seventy years" | Open, filed in note 40 |
| `14_ch14.md:133` | "this community seventy years hence" | Open, filed here |
| `01_ch01.md:41` | "Within seventy or eighty years" | Open, filed here |
| `04_ch04.md:61` | "roughly seventy or eighty years in this life, some twenty-five thousand days" | Open, filed here |
| `10_ch10.md:9` and `:21` | "seventy to eighty years", "the seventy-year replacement horizon", and the story ends at Year 75 (`:11`, `:19`) | Open, filed here |
| `10_ch10.md:34` | "Within roughly eighty to a hundred years, 100 percent" | Open, filed here |

Three of these need a word.

1. **`04_ch04.md:61`.** The count of days follows the count of years. Eighty years are 29,219 days and seventy are 25,567, so the alternate reads "roughly eighty years in this life, some twenty-nine thousand days".
2. **`10_ch10.md:9-21`.** Alternate (a) makes all four places eighty, so the story ends at Year 80, where the horizon ends. Alternate (b) changes only the two sentences that name the horizon, and Year 75 stays.
3. **`10_ch10.md:34`.** The span there is "eighty to a hundred years", not a seventy, so it may be meant to stay. With eighty years flat, "100 percent" cannot stay, because a child reading the book now may be alive then. The alternate reads "Within roughly eighty years, almost every human being". Reject the item to keep the sentence as it is.

Left alone, because they are not the horizon: `06_ch06.md:11` ("fifty, one hundred, and two hundred years", three spans of one comparison), `10_ch10.md:17` (an electrician of seventy-two), and the seventy of `09_ch09.md:249` and `:309`.

## 4. Checks on the new items

- `civ/bin/review add -strict`: passed. Each of the six quotes is located once, at its recorded offset.
- `civ/bin/lint sim` on the packet, for alternate (a): 6 edits, 0 overlaps, 0 new Part A or spelling hits. For (b): 2 edits. For (c): 1 edit.
- The kept text of every alternate is cut from its quote by script.
- `civ/bin/linear check` was not run, as in notes 39 and 40. `civ/bin/lint` (Parts A, C, and D) was not run, because no chapter changed.

## 5. The queue now

298 items: 9 open, 0 on hold, 232 applied, 57 closed, 0 stale.

## 6. What else is outstanding

The author asked for this list. It holds what no note records as done.

**In the viewer, 9 open items.**

1. `13_ch13.md:5`, the criterion on results (a, b, c).
2. `01_ch01.md:9`, "wet" (note 40).
3. Eighty years, seven items: `13_ch13.md:67`, `14_ch14.md:85`, `14_ch14.md:133`, `01_ch01.md:41`, `04_ch04.md:61`, `10_ch10.md:9-21`, and `10_ch10.md:34`.

**Outside the viewer.**

1. **The book is not built.** `civ/ccp_book.md` and `civ/build/` were last written on 2026-11-24 at 14:50 UTC (note 38). Four chapters have changed since: `09_ch09.md`, `11_ch11.md`, `12_ch12.md`, and `13_ch13.md`. The released copy is that build.
2. **Nothing of notes 39 to 41 is committed.** Uncommitted under `civ/`: four chapters, `decisions.json`, notes 39, 40, and 41, and six packet files.
3. **The changes outside `civ/` are not committed**, as in note 38: 69 entries under `book1-4` to `book1-7`, `history`, `purpose.md`, and `tts`.
4. **Probably unauthorized uploads under `ext-ref/`.** The Festinger file (notes 35 to 38). The README of `ext-ref/` names two more of the same kind, by Rawls and by Gorbachev.
5. **Two items of note 34 that no later note records as decided.** "Whether to change the cadence scoring" (section 8), and "Chapter 11: the two mechanisms without a case in the record are the Compassionate Delusion and fatalism" (section 11, item 4). Both left the list in note 36.
6. **The tools, which the author asked to hold until the passages were settled.** What waits: `civ/bin/linear parse` and `report`, because the whole-book report of note 18 predates changes to 11 chapters; `civ/bin/ngram report`, for the same reason (note 17); then `voice` as a locator with sentence-level suggestions (note 34, section 8). `civ/bin/lint` (Parts A, C, and D) and `lint lists` showed nothing to act on when last run, on 2026-11-24.

**No longer open.**

- Shinto: the section is removed (note 40), so the question of saving its sources under `ext-ref/` is closed. The texts read are in a temporary folder and are not needed.
- Refusing authority: you took alternate (a), so the question of what the capability is directed to is not before the book.
- "It as runinous", the Berkeley login, and the sentence read in another window: all closed with the sections they belonged to.

## 7. Left for the author

1. Decide the 9 open items in the viewer.
2. Say when the book is to be built and the round committed.
3. Say whether the tools are to run once the items are decided.

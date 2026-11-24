# 35. Applied Decisions and New Review Requests

Recorded 2026-11-24 (UTC; 2026-11-23 evening by the machine's clock). This covers the round after note 34: the author's 27 decisions on the items filed there, the 24 edits applied from them, the checks made after applying, and the 8 new items filed in their place. Chapter text changed in ten files under `civ/chapters/`. `civ/ccp_book.md`, `civ/build/`, `civ/ccp_sheet.md`, and the notes other than this one were not touched. `civ/build.fish` was not run. Nothing is committed.

Sources added for the new drafts are in `ext-ref/tool1_hermetic/`, `ext-ref/kindness_politeness/dl_round2/` (with `report_round2.md`), and the second addendum at the head of `ext-ref/red_green/report.md`.

---

## 1. The decisions in numbers

| Disposition | Without a note | With a note | Total |
|---|---:|---:|---:|
| `accept` | 21 | 3 | 24 |
| `hold` for research | 0 | 3 | 3 |
| `reject`, `withdraw` | 0 | 0 | 0 |
| **Total** | | | **27** |

Every open item of note 34 was decided between 22:22 UTC on 2026-11-23 and 03:27 UTC on 2026-11-24. The 24 accepts join one older accept that was already marked pending apply.

## 2. Applied

Line numbers are those after all the edits of this round.

**By the tool, 21 items** (`civ/bin/review apply`; response packet `packets/20261124-033700-applied.json`). Each was an accept without a note, located once at its recorded offset.

- `01_ch01.md:37` and `:41`: the colon after "every technology we have just named", and "wet" dropped.
- `04_ch04.md:9`: "What kind of creature is a human being?"
- `04_ch04.md:119`: "from your specific mother and father"; "supplied from a stock of generic lives"; "are not external additions placed over a pre-existing generic self".
- `05_ch05.md:9`: "human limits require".
- `06_ch06.md:63`: the harm-reduction passage with its figures made literal.
- `06_ch06.md:79` to `:83`: four sentences tightened, "revealed preference" kept and set off.
- `08_ch08.md:86`: the last sentence joined to the paragraph above; "For those reasons".
- `10_ch10.md:34`: the three facts each end at its period; the library sentence in a forward order (alternate b).
- `11_ch11.md:87`: the paragraph on the man who is unkind without any activist.
- `12_ch12.md:5`: the test of compatibility; `:61` Hume; `:69` Rousseau to Derrida with the terms restored; `:183` Dugin streamlined.
- `13_ch13.md:43`: Lee Kuan Yew, terms kept.
- `14_ch14.md:42`: Tool 2 rebuilt on correspondence.

**By item, with `-include-noted`, 3 items** (`packets/20261124-033749-applied.json`). The rule in note 31 is to pass that flag only when the replacement already reflects the note. Each of these was read against its note before the flag was passed:

- `11_ch11.md:3` (`note/opening-from-cases`). The note is approval: "We might revisit this section further, but I like this addition." The replacement reflects it.
- `11_ch11.md:149` (`note/individuals-only`). The note begins "This is much better" and ends "Otherwise looks great", and in between asks for a second change: to remove "The Defense". The accepted text is what was approved. The second change needs wording the author has not seen, so it is a new item (section 4).
- `05_ch05.md:96` (`note/cut-reinvention-test`). The note says "Keep the ending" and asks for the ending's first phrase to be rephrased. The cut was applied and the ending kept as it stands. The rephrasing is a new item (section 4).

**Not applied, 1 item.** `02_ch02.md:a676ed11:prop/boiled-water-source` (accepted 2026-11-22, with the note to remove the made-up example). Its passage no longer exists. The accepted `boiled-water-real-only` replaced the river town on 2026-11-22, and the real cases are in the chapter at `02_ch02.md:103`. Applying the older item would bring the river town back. It stays marked "accept, pending apply" until the author withdraws it (note 34, section 11, item 2; an answer now says so under the item).

## 3. Checks before and after applying

1. `civ/bin/review apply -dry-run`: 21 ready, 4 skipped for a note, no overlap, no stale hit among the ready.
2. `civ/bin/lint sim` on the 24 approved replacements, in memory, before any write: 24 unique matches, 0 overlaps, 0 new Part A or spelling hits. Voice mean over the book +0.89 to +0.84; the highest section +3.19 to +2.96.
3. `civ/bin/lint` (Parts A, C, D) before and after: the same list of findings. One finding moved by six lines (`11_ch11.md:83` to `:89`) because the opening grew. No new hit.
4. `civ/bin/linear check -file` on the ten touched chapters, before (the committed files) and after. The parser matches about 140 sentences in the new files that it did not match in the old; many are the author's own sentences with one word changed. Fifteen of them score 2 or more; one is an older sentence that the parser divided differently. The worst is the author's own list of who will be dead (`10_ch10.md:34`, 5.0), left as it is by the author's choice in note 34. Three became optional items (section 5). The rest are sentences the author wrote or accepted as they stand, and were left alone.
5. The whitespace check of `git diff`: no error. Each chapter still ends with one newline.

## 4. New items from the notes on accepted items

| Packet | Contents |
|---|---|
| `packets/20261124-043201-note-followups.json` | 2 open items: `11_ch11.md:151` and `05_ch05.md:96` |

- **`11_ch11.md`, `note/defense-line`.** The author's note: "I also think we should remove "The Defense". It is up to the society itself to decide what to do. Maybe to prevent abuse, they tell families to suport their own. Maybe they limit anything to 2 month per 4 years of work. Lots of options. Not our job, nor should we just blithly say "separate them"." Three alternates: (a) the paragraph deleted; (b) one paragraph that leaves the remedy to the society, in the author's two examples and without a number; (c) as (b) with the author's figures. The Defense line of section 1 (guard the schools) is not touched.
- **`05_ch05.md`, `note/ending-rephrased`.** The author's note: "the phrase "Reform constructed forms" is hard for me to parse ... We can reform and change what is actually constructions (paper money, electronic money.. but we keep money) ... see if you can rephrase or change order." Three alternates: (a) rephrased, with the money example, which is the chapter's own; (b) the same in the other order, so the chapter ends on money remaining; (c) the smallest change.

## 5. Fine grain, from the checks after applying

`packets/20261124-043201-fine-grain.json`: 3 optional items, each one sentence and no term changed.

- `11_ch11.md:9`, `note/name-polybius`: "he wrote" becomes "Polybius wrote". The parser finds no person for "he" because it does not know the name.
- `11_ch11.md:9`, `note/split-ormansag`: the 51-word sentence on the seven villages, split in two, "it" named.
- `12_ch12.md:187`, `note/split-whereas`: the one long sentence the streamlining left whole, split in three with every term kept. The author's note there was "I agree we can streamline some of this or edit it".

## 6. The three holds, answered

`packets/20261124-043201-research-drafts.json` holds the drafts; `packets/20261124-043509-kindness-situation.json` re-files one of them with a correction. `packets/20261124-043255-hold-answers.json` holds the answers.

| Passage | The note asks for | Filed | Left to the author |
|---|---|---|---|
| 14_ch14.md:19 Tool 1 | Three strategies with a draft for each; examples; a belief may be sealed about God but not about reality | `note/tool1-three-strategies`: (a) the question put before the event; (b) the believer taken out of the observation, each excuse examined as a claim; (c) the author's six points kept, with two real comparisons | Which strategy; whether the six points survive as the procedure; the term *hermetic belief* |
| 13_ch13.md:135 Red-Green | Tell it through story; not so many facts; no list of five; be specific | `note/red-green-through-cases`: four cases, each ending in the idea it gives; the author's block quotation moved to the end; (b) without the passage on the two sacred claims | Whether the block quotation moves; the heading; whether the sacred passage stays |
| 08_ch08.md:125 Kindness | One or two situations, then the point; not COVID; a clearer example | `note/two-situations`: Hurricane Katrina, 2005, by the auditor's estimate; (b) adds the committee on the August 2011 disorder | Which situations |

**Tool 1.** Examples, each read in a saved copy: the group whose prophecy failed on 21 December (Festinger, Riecken, and Schachter, 1956); the double-blind test of astrology whose advisers fixed the least result in advance (Carlson, *Nature*, 1985); the N-rays, in Wood's letter to *Nature* of 1904, where each failure had an explanation and the explanations were tested; Korea at night (NASA, 2014); and East Germans leaving for the West before 1961. All three drafts carry the author's sentence: a belief may be sealed about God; that is left open (Chapter 3); only how it treats reality is examined. Tool 2 opens with "Related to asymmetrical accounting", so all three keep that term. Length: original 417 words; a 531, b 580, c 654.

**Red-Green.** No source was added. The cases are London 2004 (Livingstone and Qaradawi: the pairing does not rest on agreement), Tower Hamlets 2014 (the accusation as cover, with the judgment's own limit), Lewisham 2016 (two ways of speaking in one person), and Paris 2019 (the CCIF and the politicians of the left in one march, then the dissolution and its defense). The street blockade is not told: it has no case on record. Length: original 434 words; a 1,021, b 881.

**Kindness.** COVID is gone. The agency paid two thousand dollars on the applicant's word, usually before any inspection. The Government Accountability Office estimated that about one payment in six, about a billion dollars, rested on an identity, an address, or a claim that was false or filed twice (GAO-06-844T). The agency disputed the size of the estimate and accepted most of the recommendations for checking (GAO-06-1013). Then the point about kindness and the author's two closing sentences. Length: the held draft 619 words; a 342, b 374.

## 7. Checks on the new items

- Every quotation in the drafts was matched by script against a saved copy of its source, and every span carried from the author's text against the chapter: 21 distinct spans from a saved source (31 occurrences across the alternates), 11 quotations of the author's notes, one question of my own, none unmatched. The author's notes are quoted verbatim, spellings kept.
- Each date and figure in the drafts was searched for in its document. Two were removed for want of a source: the year of the failed prophecy (the book's text gives the day and month only) and "a year" on the Korean power figures (the caption gives no year).
- `civ/bin/lint prose`, `phonetics`, `structure`, `render`, `lists`, and `voice -all` on each draft: no Part A hit; one rhyme, inside a quotation; no run-in list. Voice score of the whole draft as one section: Tool 1 now +1.40 (the highest in its chapter), drafts -0.15, -0.29, +0.52; Red-Green now +2.56, draft +0.08.
- `civ/bin/linear check` on each draft: sentences were split and pronouns named until no sentence of the new prose in the Tool 1 drafts (a) and (b) scored 2. Those that still score are the author's own sentences, left as they are.
- `civ/bin/lint cadence -semantic`, by paragraph: Snap 52 to 74 in the new prose, no treading-water flag. A reading aid only (note 34, section 8).
- `civ/bin/lint sim` on each packet and each alternate: unique matches, no overlap, no new Part A or spelling hit.
- `civ/bin/review add -strict`: passed for all four item packets.
- The kindness item was filed a second time after "before any inspector called" was checked against the source, which says "typically, prior to a physical inspection". The item id did not change.

Facts that came up in research and should not be used as they are usually stated:

- Wood's letter of 1904 does not say who removed the prism, and has no pocket and no assistant's suspicion. Those belong to a story Wood told a biographer in 1941.
- Carlson's "50 percent" was set by the astrologers' advisers, and the test has critics. The drafts attribute the prediction to the advisers and say nothing about what astrologers did afterward.
- The number who left East Germany between 1949 and 1961 is given as 2.5 million by the saved page and as 2.7 million by another source. The drafts follow the saved page.
- The Conseil d'État upheld the dissolution of the CCIF on one ground and rejected the government's ground of provocation to terrorism. The draft names the first; say if the second belongs in the book.
- GAO's figure for Katrina is "improper and potentially fraudulent payments" by estimate, and FEMA disputed its size. The draft carries both.
- The Google dismissal of 2017 is known here through three news reports of the labor board's memorandum. The memorandum is still not saved (note 34, section 11, item 8).

## 8. The queue now

278 items: 8 open, 23 on hold (all answered), 219 applied, 27 closed, 1 accept pending apply, 18 stale.

## 9. Left for the author

1. Decide the 8 open items in the viewer.
2. Withdraw the 18 stale items, none of which can be applied: the 12 source drafts whose follow-ups are now applied (an answer under each says so), the 5 older research holds from 2026-11-22, and `02_ch02.md:a676ed11:prop/boiled-water-source`.
3. `11_ch11.md`, section 1 of the parasitism passages, still has a Defense line (guard the schools). Say if it goes as well.
4. Whether the labor board's memorandum should be obtained, if the Google sentences are kept.
5. The remaining items of note 34, section 11: Chapter 11's two mechanisms without a case in the record; the passages the author's notes question (`12_ch12.md:141`, `13_ch13.md:5`, `09_ch09.md:75`); and the cadence scoring.
6. `ext-ref/tool1_hermetic/dl/festinger_riecken_schachter_1956_when_prophecy_fails.pdf` is probably an unauthorized upload, kept only to check short quotations. Delete it if it is not wanted.
7. Commit when ready. This round's changes are uncommitted: the ten chapters, `decisions.json`, and the six new packet files.

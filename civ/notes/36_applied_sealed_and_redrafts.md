# 36. Applied Decisions, the Meaning of Sealed, and Four New Items

Recorded 2026-11-24 (UTC; the same morning by the machine's clock). This covers the round after note 35: the author's 8 decisions on the items filed there (5 accepts and 3 holds for research), the 5 edits applied from them, the checks made after applying, the answers to the 3 holds, and the 4 new items filed in their place. Chapter text changed in three files this round: `05_ch05.md`, `11_ch11.md`, and `12_ch12.md`. `civ/ccp_book.md`, `civ/build/`, `civ/ccp_sheet.md`, `civ/notes/06_editing_guide.md`, and `ext-ref/` were not touched. `civ/build.fish` was not run. No tool code was changed. Nothing is committed.

No source was added. Every fact and quotation in the new drafts was read again in the copies saved for the earlier drafts (`ext-ref/tool1_hermetic/`, `ext-ref/red_green/`, `ext-ref/kindness_politeness/dl_round2/`, and `ext-ref/decline/`).

---

## 1. The decisions in numbers

| Disposition | Without a note | With a note | Total |
|---|---:|---:|---:|
| `accept` | 4 | 1 | 5 |
| `hold` for research | 0 | 3 | 3 |
| `reject`, `withdraw` | 0 | 0 | 0 |
| **Total** | | | **8** |

Every open item of note 35 was decided between 11:36 and 12:06 UTC on 2026-11-24. No item was rejected or withdrawn.

## 2. What the notes ask for

The standing instructions of note 34, section 2, hold: literal language, words that carry a specific meaning, small changes, and the second person with its signposts. Four more follow from the three notes of this round and from the author's instruction after it.

1. **An example ends on the idea it gives.** "The examples are both good, and fail because they don't pivot after to tie into an explicit idea." An example is paired with a principle, and the sentence after it states the idea.
2. **Arguments are mapped so that they build and conclude.** A passage is told in one pattern from beginning to end, and its conclusion repeats the points made along the way: "we start out with the premise of what we want to show, ... then we bring up evidence, then we point out after the evidence".
3. **A passage states the objection a reader could raise.** "The downside of this example is you could argue they know it would be abused, but decided it was worth it as 5/6 of the payments were not obvious fraud. ... but be careful on the phrasing." (Spellings are the author's.) A case is not made to say more than its record.
4. **No new term where a plain word can be defined.** "We should probably not say "hermetic" as it introduces a new term. But we should say what we mean by sealed to proof in the world and distinguish it from sealed from proof of God. That might flow better." This was said after the round, in conversation, and is followed in section 6.

## 3. Applied

Line numbers are those after all the edits of this round.

**By the tool, 4 items** (`civ/bin/review apply`; response packet `packets/20261124-121439-applied.json`). Each was an accept without a note, located once at its recorded offset.

- `05_ch05.md:96`: the closing paragraph is rephrased and opens on "A constructed form may be changed". The money example is one sentence. The chapter still ends on "agricultural ruin".
- `11_ch11.md:151`: The Defense paragraph of section 2 is deleted. The section now ends with the able-bodied man who treats public assistance as a permanent livelihood.
- `11_ch11.md:9`: "he wrote" becomes "Polybius wrote".
- `12_ch12.md:187`: the sentence on the First, Second, and Third Political Theories is split in three. No term changed.

**By item, with `-include-noted`, 1 item** (`packets/20261124-121439-applied-2.json`). `11_ch11.md:9` (`note/split-ormansag`). The note is approval and a second request: "Change is fine, but there are too many details. We could say it rose only slightly or in some cases dropped suddenly, all well below." The split was applied as accepted, as the note approves. The fewer details is a new item (section 5).

**Not applied, 1 item.** `02_ch02.md:a676ed11:prop/boiled-water-source`, as in notes 34 and 35. Its passage no longer exists. It stays marked "accept, pending apply" until the author withdraws it.

## 4. Checks before and after applying

1. `civ/bin/review apply -dry-run`: 4 ready, 2 skipped for a note, no overlap, no stale hit among the ready.
2. `civ/bin/lint sim` on the 5 accepted replacements, in memory, before any write: 5 unique matches, 0 overlaps, 0 new Part A or spelling hits. Voice mean over the book +0.84 before and after; the highest section +2.96 before and after.
3. `civ/bin/lint` (Parts A, C, D) before and after: the same list of findings.
4. `civ/bin/linear check -file` on the three touched chapters, before and after.
   - `05_ch05.md`: 80 sentences before and 81 after. The three new closing sentences score 0.0, 0.0, and 0.5.
   - `11_ch11.md`: 252 sentences before and after. The sentence with the Polybius quotation scores 4.0, against 6.0 before; a quotation stands before the verb in both forms. The Ormánság sentence scores 2.0, against 2.5 before. It is still 44 words, and the new item on it reduces the details.
   - `12_ch12.md`: 244 sentences before and 246 after. The sentence that was split scored 4.0; its three parts score 0.0.
5. The whitespace check of `git diff`: no error. Each chapter still ends with one newline.

## 5. New items

| Packet | Contents |
|---|---|
| `packets/20261124-124008-note-followups.json` | 1 open item: `11_ch11.md:9` |
| `packets/20261124-124445-research-drafts.json` | 3 open items: `14_ch14.md:19`, `13_ch13.md:135`, `08_ch08.md:125` |
| `packets/20261124-124008-hold-answers.json` | Answers to the 3 holds and to the noted accept |

The drafts were filed first at 12:40 UTC and re-filed, corrected, at 12:44 UTC after `civ/bin/linear check` found pronouns to name and sentences to split. The item identifiers did not change.

| Passage | The note asks for | Filed | Left to the author |
|---|---|---|---|
| 14_ch14.md:19 Tool 1 | Points 1, 5 in your variant, and 6; another word for interlocutor; examples paired with a principle; arguments that build and conclude; then, after the round: not "hermetic"; say what sealed means | `note/tool1-sealed-three-steps`: (a) sealed defined; three steps, each with a principle and a case; (b) the same, with God in the words of Chapter 3; (c) shorter | How "sealed to" and "sealed from" were read; the order of the cases; the length |
| 13_ch13.md:135 Red-Green | Your first sentence; premise, evidence, then which was red, which was green, and their common enemy; the pattern once more, then where they disagree in the same structure; a conclusion that repeats the points; remove "traditional Christian" | `note/red-green-premise-evidence-point`: (a) with your block quotation at the end, the clause removed; (b) the same points in prose | Whether Paris stays the second case; whether the block quotation stays; the heading |
| 08_ch08.md:125 Kindness | "that" named; the objection that the agency knew and judged speed worth it; careful phrasing; no police example | `note/kindness-known-risk`: (a) full; (b) shorter | The choice of alternate |
| 11_ch11.md:9 Ormánság | Fewer details: "rose only slightly or in some cases dropped suddenly" | `note/fewer-details`: (a) one sentence in your words; (b) the same without dates; (c) the smallest change | Whether the two sentences that follow are reduced as well |

## 6. What we mean by sealed

The term *hermetic belief* is withdrawn. The word is the author's own, and the draft defines it. The definition as filed in alternate (a), that the author may judge it without the viewer:

> Such a belief is **sealed to proof in the world**. Sealed to proof means that the belief holds to a source of truth that no observation is allowed to correct (Chapter 1). The belief claims something about reality. No event can count against that claim. A belief about God is a different case. Such a belief is sealed from proof, not to proof. Proof is a test against the world, and no test settles a claim about what lies beyond the world. The principles leave that claim open (Chapter 3). This tool examines only how a belief treats reality. The tool reaches a belief about God only where that belief states something about the world, such as a date.

How the two phrases were read:

1. **Sealed to proof in the world** is a defect. The belief is about reality, and it is closed against proof: whatever happens is made to agree with it. The defining sentence is Chapter 1's own, "a source of truth that no observation is allowed to correct" (`01_ch01.md:17`), so the reader has met the idea before Chapter 14.
2. **Sealed from proof** is not a defect. A belief about God stands where proof does not reach, because proof is a test against the world. The principles leave it open, as Chapter 3 does in its own words: for the ultimate questions "the principles issue no catechism and no anti-catechism" (`03_ch03.md:74`).
3. **The boundary is set before the test.** A belief about God enters the tool where it states something about the world. A date for a flood is a claim about the world, whichever belief gives it. Chapter 3 says the same: "A claim that fails that test does not cross to the metaphysical side afterward" (`03_ch03.md:76`). The failed prophecy is the case that shows it.

Alternate (b) says the second point without the second use of *sealed*, in case "sealed from" was meant otherwise. Say so and it will be changed.

## 7. The drafts

**Tool 1** (`14_ch14.md:19`; chapter now 423 words; a 924, b 938, c 755).

- The word *interlocutor* is replaced by *speaker*.
- Three points remain, in the author's wording: 1, 5 in the variant ("apply the same standard of evaluation to arguments for and against"), and 6. The other three are gone, with the name *Six-Point Single Ledger Audit*. *Asymmetrical accounting* and *Single Ledger* are kept and defined, because Tool 2 opens on the first.
- Each point has a principle, a case, an idea, and a turn back to the partisan.
  1. State the program in advance. Principle 2. The N-rays, led as falsely claimed and condensed to the primary points. Idea: a standard set after the result is set by expectation.
  2. Apply the same standard to arguments for and against. Principle 5. Korea at night, in four sentences. Idea: one measure is applied to both sides.
  3. Specify what fact would falsify the claim. Principle 1. The prophecy that failed on 21 December. Idea: the fact named in advance must be accepted when it arrives; the believers changed the claim instead.
- The author's two Venezuela paragraphs are kept word for word.
- Alternate (c) is the same argument with each case in two or three sentences.
- The length is more than twice the passage's. Say if Tool 1 should be shorter than alternate (c).

**Red-Green** (`13_ch13.md:135`; chapter now 434 words; a 760, b 759).

- The first paragraph opens on the author's sentence as written: "There is no better demonstration of this than the functional arrangement between radical progressives and Islamists." *Political* is dropped before *Islamists*.
- Three paragraphs follow, each in one pattern: the premise, the evidence, then which was red, which was green, and their common enemy.
  1. That the two sides organize one event together. The Muslim Association of Britain and the Stop the War Coalition, 28 September 2002 and 15 February 2003. Red was the Socialist Workers Party; green was the Association.
  2. That the arrangement survives the contradictions between the doctrines. Paris, 10 November 2019, and the dissolution of the CCIF. Red was the Communist Party and the anticapitalist party; green was the CCIF.
  3. Where it ends, in the same structure. Livingstone's own words on Qaradawi, then Iran in the words of the author's section 1.
- In each paragraph the common enemy is named twice: the immediate one (a government, a law), and behind it "the civilization that would oppose its own destruction", in the author's words of the note. Classical liberalism is not named.
- The conclusion repeats the points. Alternate (a) keeps the author's block quotation, less the clause on traditional Christians, and ends on the author's sentence: "They will destroy the civilization that opposes them, such as the one this book describes." Alternate (b) says the same in prose.
- The list of five and the three bullets on defeat are gone.
- One limit stays from the record: not all of the left joined in Paris. The Socialist Party stayed away.

**Kindness** (`08_ch08.md:125`; chapter now 124 words; a 458, b 398).

- The police example is gone. "A household can give that" is now "A household gives without asking for proof, because it sees the person every day."
- The author's objection is in the passage. The risk was known before the storm: inspectors general had reported the same weakness in 2001, 2004, and May 2005. The agency paid quickly all the same. A defender can say it set the loss against the speed and judged the speed worth it. The record does not settle that, and the passage says so.
- The passage gives the auditors' estimate, the agency's dispute of its size, and its acceptance of most of the recommendations. It also says that about five payments in six passed the auditors' tests, so the case is not made to say more than the record.

**Ormánság** (`11_ch11.md:9`). Three ways to say it in the author's words: "rose only slightly", and in six of the seven villages it fell. Alternate (a) keeps the dates and the comparison with Hungary, as "nearly twice what it had been".

## 8. Facts that came up, and limits that travel with the drafts

- **Ormánság, "dropped suddenly".** The record says *fell*. The six villages lost people over seventy years, and the one figure the review gives for a village is Lúzsok, from 393 to 242 (Bagdi, 2023, pp. 96 and 97). The drafts say *fell*.
- **Ormánság, Hungary.** The index rose from 188 to 350 with 1785 taken as 100. That is 1.86 times, so the drafts say "nearly twice what it had been" or "nearly doubled", never *tripled*.
- **The N-rays.** Wood's letter does not say who removed the prism or put the wood in place of the file. The draft's sentence has no agent: "In three observations there, the answer was fixed in advance and kept from the observer." The end of the rays is the account of *APS News* (2007), not a document of 1904. The believers were not accused of fraud there; the draft says *falsely claimed*, the author's phrase, and that a person can find a result "without knowing that he does".
- **The prophecy.** Festinger's narrative gives the day and the month and no year. The drafts give none. The group is not named and no member is named.
- **Korea.** NASA's caption gives no year for the figures on power, and says nothing of when Korea was divided. The drafts give the date of the photograph only.
- **Red-Green, London.** The leaflets are Peace's account of the first joint demonstration, 28 September 2002 (Peace, 2015, citing Phillips 2008). The printed chapter reads "Freedome for Palestine", a misprint. The drafts describe the leaflets and quote nothing from them. "The largest demonstration in British history" is Peace's "estimated 2 million people", which "dwarfed the previous mass rallies of British political history".
- **Red-Green, Paris.** The Conseil d'État upheld the dissolution on one ground and rejected another. It found the CCIF's conduct of a nature to provoke discrimination, hatred, or violence, and it found no acts aimed at provoking terrorism. Both are in the drafts.
- **Red-Green, Iran.** No source was added. The sentence restates the author's own section 1.
- **Kindness.** GAO's 16 percent counts payments with an invalid identity, address, or duplicate registration, by estimate; the interval is 12 to 21 percent. "About five payments in six passed those tests" says only that. GAO adds that its estimate likely understates the total, because other kinds of improper payment were not tested. The agency put its own figure at 1 to 3 percent, and GAO answered that the figure rested on no independent sample.

## 9. Checks on the new items

- Every quotation in the drafts was matched by script against a saved copy of its source, and every span carried from the author's text against the chapter: 36 quoted spans across the alternates: 25 occurrences are from a saved source (11 distinct) and 11 are the author's own words (4 distinct), none unmatched.
- Each date and figure in the drafts was searched for in its document: 99, none missing.
- `civ/bin/lint prose`, `phonetics`, `structure`, `render`, `lists`, and `voice -all` on each draft: no Part A hit; one rhyme, inside the Festinger quotation ("night" and "light"); no run-in list. Voice score of the whole draft as one section: Tool 1 -0.05, +0.03, and +0.02 (the section in the chapter now scores +1.40); Red-Green -0.09 (now +2.56); Kindness -0.23 and -0.27.
- `civ/bin/linear check` on each draft: the sentences that still score 2 or more are the author's own, left as they are. In Tool 1 they are the two Venezuela sentences (3.5 and 3.0) and the form of the sixth point. In Red-Green they are the author's sentence on what the two sides build and pull down, used three times, and the author's block quotation. In Kindness they are the author's two closing sentences. In the Ormánság sentences, none.
- `civ/bin/lint cadence -semantic`, by paragraph: Snap 56.5 to 63.7, no treading-water flag. A reading aid only (note 34, section 8).
- `civ/bin/lint sim` on each packet and each alternate: unique matches, no overlap, no new Part A or spelling hit.
- `civ/bin/review add -strict`: passed for all three packets.

## 10. The queue now

282 items: 4 open, 26 on hold (all answered), 224 applied, 27 closed, 1 accept pending apply, 18 stale.

## 11. Left for the author

1. Decide the 4 open items in the viewer.
2. Say how "sealed to" against "sealed from" was meant (section 6).
3. Withdraw the 18 stale items and `02_ch02.md:a676ed11:prop/boiled-water-source`, none of which can be applied.
4. Three passages that your notes of 2026-11-23 question are still as they were. Each needs your word before a draft is written.
   - `13_ch13.md:5`, the Generational Criterion: "We might need to re-evaluate this paragraph section as I'm re-reading it." What should be re-evaluated: the *never*, the eighteen months, the span (70 to 100 years against "seventy years later"), or the criterion itself?
   - `09_ch09.md:75`, refusing authority: "It is more then just "the bench". It is rarely the worst man in town. That is spotted quickly. It is about priority, capability directed to what exactly?" The passage says the vacant jurisdiction is seized by the most unscrupulous. Tell me what the capability is directed to, and the paragraph will be written on priority.
   - `12_ch12.md:141`, Shinto: "We should evalutate Shinto as a whole before we look to this." No source on State Shinto is saved. Say so and it will be researched: the Rescript on Education of 1890, the state's separation of shrine Shinto from the sects, the rescript of 1945, and the record of the war.
5. `11_ch11.md:143`, section 1 of the parasitism passages, still has a Defense line (guard the schools). The Defense of section 2 is gone by your decision. Say if this one goes as well; it is a duty of parents rather than a policy for a society, so it may be meant to stay.
6. `ext-ref/tool1_hermetic/dl/festinger_riecken_schachter_1956_when_prophecy_fails.pdf` is probably an unauthorized upload, kept only to check short quotations. Delete it if it is not wanted.
7. The labor board's memorandum on the Google dismissal (note 35, section 9, item 4) is no longer needed. Neither Red-Green alternate carries the Google sentences.
8. Commit when ready. This round's changes are uncommitted: three chapters, `decisions.json`, this note, and six packet files. The changes of the previous round are uncommitted as well: seven other chapters, note 35, and seven packet files. `civ/review/*.go` also carry changes of the evening of 2026-11-23 that are not committed; `civ/bin/review` was built from them.

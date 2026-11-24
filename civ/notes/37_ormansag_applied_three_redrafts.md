# 37. The Ormánság Sentence Applied; Tool 1, Red-Green, and Kindness Redrafted

Recorded 2026-11-24 (UTC), after the author's decisions on the four items of note 36. One item was accepted and is applied. Three were held for research with notes, and each is answered with a new draft. Chapter text changed in one file this round: `11_ch11.md`. `civ/ccp_book.md`, `civ/build/`, `civ/ccp_sheet.md`, `civ/notes/06_editing_guide.md`, and `ext-ref/` were not touched. `civ/build.fish` was not run. No tool code was changed. Nothing is committed.

No source was added.

---

## 1. The decisions in numbers

| Disposition | Without a note | With a note | Total |
|---|---:|---:|---:|
| `accept` | 1 | 0 | 1 |
| `hold` for research | 0 | 3 | 3 |
| `reject`, `withdraw` | 0 | 0 | 0 |
| **Total** | | | **4** |

The four items were decided between 13:06 and 13:25 UTC on 2026-11-24.

## 2. What the notes ask for

The standing instructions of notes 34 and 36 hold. Five more follow from the three notes of this round.

1. **Be concrete, and drop the accounting figure.** "I don't like the phrase "In sound accounting" it is a metaphore. Say what you actually mean, perhaps "Claims about reality are always testable against reality." or adapt to something similarly concrete. no "Ledger". Again be concrete."
2. **An example must show something for the book's purpose.** "I don't think the N/S Korea example shows anything for our purpose, remove."
3. **Do not give a word credence the book rejects.** "Don't give the word "Islamophobia" creedence. You can name the march, but Islam is a religion and it can be revialed. That is fine." A thing is named by its own name, and the book's view of the word is said beside it.
4. **Put the failure in the sentence, not in an announcement of it.** "This is useless "The attempt fails for a plain reason." the failing aspect should be incorperated into the previous or next sentence."
5. **Streamline and linearize.** "We need to streamline and linearize the suggested text some more." One claim to a sentence, and a thing said once.

(Spellings are the author's.)

## 3. Applied

**By the tool, 1 item** (`civ/bin/review apply`; response packet `packets/20261124-132836-applied.json`). `11_ch11.md:9` (`note/fewer-details`), an accept without a note, located once at its recorded offset. The sentence now reads: "In seven villages of the Ormánság the population rose only slightly between 1870 and 1941, and in six of the seven it fell, while the population of Hungary rose to nearly twice what it had been (Gábor Koloh, 2021, figures given in Róbert Bagdi's review, 2023)."

**Not applied, 1 item.** `02_ch02.md:a676ed11:prop/boiled-water-source`, as in notes 34 to 36.

## 4. Checks before and after applying

1. `civ/bin/review apply -dry-run`: 1 ready, 1 skipped for a note (the stale `boiled-water-source`, whose passage no longer exists).
2. `civ/bin/lint sim` on the accepted replacement, in memory, before the write: 1 unique match, 0 overlaps, 0 new Part A or spelling hits. Voice mean over the book +0.84 before and after.
3. `civ/bin/lint` (Parts A, C, D) before and after: the same list of findings.
4. `civ/bin/linear check -file` on `11_ch11.md`, before and after: 252 sentences before and 251 after. The Ormánság sentence scores 1.0, against 2.0 before.
5. The whitespace check of `git diff`: no error. The chapter still ends with one newline.

## 5. The three holds, answered

| Packet | Contents |
|---|---|
| `packets/20261124-133627-research-drafts.json` | 3 open items: `14_ch14.md:19`, `13_ch13.md:135`, `08_ch08.md:125` |
| `packets/20261124-133627-hold-answers.json` | Answers to the 3 holds |

| Passage | The note asks for | Filed | Left to the author |
|---|---|---|---|
| 14_ch14.md:19 Tool 1 | The version with God in the words of Chapter 3; Korea removed; no "In sound accounting", no "Ledger"; two sentences not understood | `note/tool1-concrete`: (a) the draft; (b) the same with the N-rays in two sentences | Where the sentence "I observe that I'm often wrong on things I can check" comes from; whether *asymmetrical accounting* stays |
| 13_ch13.md:135 Red-Green | A different sentence after the opening; no credence for "Islamophobia"; no Qaradawi quotation; end with Iran; another difference than homosexuality | `note/red-green-ends-with-iran`: (a) with your block quotation at the end; (b) the same in prose | Whether the two sentences on reviling a religion stay in the passage |
| 08_ch08.md:125 Kindness | The shorter version, streamlined and linearized; "The attempt fails for a plain reason" gone | `note/kindness-streamlined`: (a) household and agency said once, in the rule; (b) the same, ending the example | The choice of alternate |

**Tool 1** (chapter now 423 words; a 868, b 789).

- The draft is built on the alternate you named, with God in the words of Chapter 3. *Sealed from proof* is not used in it, as in that alternate.
- Korea is removed. The passage has two cases: the N-rays and the failed prophecy.
- "In sound accounting" and *Single Ledger* are gone. The passage ends on your sentence, "Claims about reality are always testable against reality," and on one plain sentence that keeps what the ledger sentence meant: "A doctrine is tested by all of its results, the losses as well as the gains."
- Your sentence on the defect also used *ledger* and *account*. It now reads "credited to the theory" and "attributed to external interference". *Asymmetrical accounting* is kept as the name of the device, because Tool 2 opens on it.
- "An excuse is a claim about the world" is gone. The second step now uses the partisan's own words. *"Foreign sabotage, hoarders, adverse weather"* can be asked who did what, and when, and what it changed; a claim made for the doctrine is asked the same questions.
- "People are limited, and a person who expects a result can find it without knowing that he does" is replaced by "People find what they expect to find," and the idea after the case by two sentences: "A test fixed after the result is fixed by what the observer expects. A program stated after its years is stated to fit them."
- The sentence you did not understand, "I observe that I'm often wrong on things I can check", is not in the draft, in any packet, or in any file of the repository. It may have been read in another window. The sentence nearest to it in the draft was the one on Principle 2, and that one is replaced.

**Red-Green** (chapter now 434 words; a 763, b 762).

- "The cases show how the arrangement works and where it ends" is replaced by your sentence: "Here is how their cooperation works."
- The march is named as its organizers named it: "marched through Paris in what its organizers called a march against Islamophobia." The word is not used again. Your reason is in the passage, in your words: "Islam is a religion, and a religion can be reviled. People have reviled Christianity for hundreds of years without being called phobes."
- The Qaradawi quotation is gone, and with it Livingstone. The third paragraph ends with Iran, in the same structure.
- Homosexuality is not offered as the difference. In its place, from your own account of Iran in section 1: the Marxists wanted a secular socialist republic and equality between the sexes; the Islamists wanted religious law, the veil, and the execution of apostates. Your block quotation is not touched; you called it fine.
- No source was added. Every fact is from the copies saved for the earlier drafts, except Iran, which is your own account.

**Kindness** (chapter now 124 words; a 340, b 356).

- Both alternates are built on the shorter version. "The attempt fails for a plain reason" is gone. The failure is in the sentence that follows: the attempt fails because an anonymous bureaucracy cannot personally verify who is a genuine crisis case, and unscoped public "kindness" therefore becomes an invitation to systematic abuse.
- The Katrina paragraph runs in one order: what was paid and how, that the risk was known, what a defender can say, what the record shows, the estimate, the dispute, the accepted recommendations. One claim to a sentence.
- The household and the agency are said once. Alternate (a) says it in the rule; alternate (b) ends the example on it and then gives the rule.
- Your last two sentences are kept word for word.

## 6. Checks on the new items

- Every quotation was matched by script against a saved copy of its source, and every span carried from your text against the chapter: 21 quoted spans, of which 11 are from a saved source (7 distinct) and 10 are your own words (5 distinct), none unmatched.
- Each date and figure was searched for in its document: 99, none missing.
- `civ/bin/lint prose`, `phonetics`, `structure`, `render`, `lists`, and `voice -all` on each draft: no Part A hit; one rhyme, inside the Festinger quotation; no run-in list. Voice score of the whole draft as one section: Tool 1 +0.09 and +0.08 (the section in the chapter now scores +1.40); Red-Green -0.04 (now +2.56); Kindness -0.11 and -0.22.
- `civ/bin/linear check` on each draft: the sentences that score 2 or more are your own (the two Venezuela sentences, the form of the sixth point, the sentence on what the two sides build and pull down, your block quotation, your two closing sentences in Kindness) and one sentence of Chapter 3's shape, "For the ultimate questions that produce no measurable distinction in the observable world, ...".
- `civ/bin/lint cadence -semantic`: Snap 54.5 to 62.4, no treading-water flag. A reading aid only.
- `civ/bin/lint sim` on the packet and each alternate: unique matches, no overlap, no new Part A or spelling hit.
- `civ/bin/review add -strict`: passed for both packets.

## 7. The queue now

285 items: 3 open, 29 on hold (all answered), 225 applied, 27 closed, 1 accept pending apply, 18 stale.

## 8. Left for the author

1. Decide the 3 open items in the viewer.
2. Tell me where the sentence "I observe that I'm often wrong on things I can check" was read. It is not in the drafts.
3. Say whether *asymmetrical accounting* stays. Tool 2 opens on it, so its first sentence would change with it.
4. The two sentences on reviling a religion are in the Red-Green draft in your words. Say if they were meant only as an instruction, and they come out.
5. What remains from note 36, section 11: the 18 stale items and `boiled-water-source` to withdraw; the three passages your notes question (the Generational Criterion, refusing authority, Shinto); the Defense line of section 1 in `11_ch11.md:143`; the Festinger file; and the commit.

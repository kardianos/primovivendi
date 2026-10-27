# Notes on `09_chapter_review_instructions.md`

A reading of the review instructions against notes 07 and 08, the short sheet, `civ/build.fish`, and the passages the instructions name. The instructions are unchanged. The chapters are unchanged. Nothing here is a call on a chapter sentence, and nothing here is wording for the book.

## What holds

The split is the right one for this book. A change of form the author has already accepted is an edit. A new issue, or a doubtful match to an old decision, is a call, and the sentence stays. Proposals stay out of the chapters until the author accepts them.

The scope is right. Edit `civ/chapters/*.md` only. Leave the short sheet, the derivation, the engineering draft, the proposal sections of notes 07 and 08, and the built book file alone. Two cuts were for the short form only, and the instructions name them: note 07 call 10, and note 07 call 20.

The lock list matches `notes/06_editing_guide.md` and `notes/00_civ_authoritative_bones.md`. The warning on the stale sentence in note 07 is accurate. The chapters and `civ/ccp_book_argon.md` say "Form and listen." Note 05 still says "Generational obligation," and it is a note. "Passing life" is still the opening civilization clause in `civ/chapters/01_ch01.md`, and the assembled book still has it because that chapter still has it.

These named passages are where the instructions say they are.

- Chapter 1 opens on "Passing life and the ability to build and keep from generation to generation is civilization."
- Chapter 8 keeps the longer Turkish record, including the 1942 wealth tax, after the Young Turks.
- Chapter 10 still has the 0.8 children arithmetic, the 84 percent decline, and the neighbors who raised three or four children.
- Chapter 4 still has the eighty-pound load and the unsourced fracture multiple. The instructions are right to make that a call rather than a silent cut. The short files dropped those sentences because they were unsourced. The book was given room to argue a fuller physical account, and the author has not accepted chapter wording for it.
- Chapter 2 states that millions starved, in the Lysenko section. The keep-rule for that clause has a limit, below.
- The Burke passage, the seed-grain granary, and the Washington example are in chapter 10.

`civ/build.fish` does assemble `ccp_book_argon.md` from `civ/chapters/*.md` in filename order. Fixing a chapter, and leaving the book file alone until the author asks for a build, is the right order.

## 1. The paths have no common root

The file mixes two roots. From the repository root, `civ/chapters/` and `civ/build.fish` resolve, and `notes/07_ccp_argon_calls.md`, `ccp_argon.md`, `chapters/01_ch01.md`, and `notes/10_chapter_calls.md` do not. From `civ/`, the second set resolves, and the `civ/` prefixes do not. There is also a `book/chapters/` tree. A session that follows the written paths will read the wrong tree or create `notes/10_chapter_calls.md` outside `civ/notes/`.

Recommend one root, stated in the first lines, and one form after that. Paths relative to the repository root, each beginning with `civ/`, match the way a new session is usually pointed at the file. The calls file is `civ/notes/10_chapter_calls.md`.

## 2. The opening says the method is the same method

The first paragraph says the method is the one used on `ccp_argon.md`, then on the engineering draft and the derivation. Those reviews wrote calls, the author decided, and the files were edited afterward. This review applies a decision already recorded, and it leaves every new call unapplied. The body says that. The opening does not. A session that "does what the earlier reviews did" will either call everything and edit nothing, or paste a new proposal into a chapter because an earlier "order of application" worked that way.

Recommend the opening state the difference in one sentence. The register and the call shape are the same. The authority is different: only a recorded decision is applied.

## 3. The default is a full pass

The file says to read one chapter at a time through chapter 14, to report two lists, and to stop if the author wants to decide the new calls before the next chapter. Pointing a session at the file is the whole assignment. Nothing in that assignment says to stop. The session can apply every inline form-change in fifteen files before the author sees a list.

Recommend the default be one chapter, then the two lists, then a stop. A second chapter waits for the author to name it. The existing rule for a requested range can stay: several chapters in one pass still leave every new call unapplied. Each inline change in the report should name the decision it follows, by note and call number, and quote the sentence it replaced.

## 4. A build publishes

`notes/06_editing_guide.md` section 6 tells the editor to run the audit and then `fish civ/build.fish`, which publishes through `push_release`. The review file says to read that guide first, and it says to run `build.fish` only when the author asks. It does not say that the script's last step runs `../push_release` when that file exists. An author who asks for a build, or a session that follows section 6, publishes.

Recommend the review file override section 6 for this pass. The audit script and the n-gram script stay unused unless the author asks. `civ/build.fish` stays unused unless the author asks. A request to build is a request to release, because the script ends in `push_release`. Say that next to the build rule.

## 5. Three different authorities are easy to collapse

The file says the full wording is on the short sheet and in the Author's decisions, and that the proposal sections are history. Where a decision rejected a proposal, the decision stands. That last sentence is necessary, and it is not the whole rule. Three cases get pasted wrong.

- **Accepted as suggested.** Notes 07 decisions 1 through 9, 13 through 16, 18, 19, 23, and 25, and note 08 decisions 2, 4, and 5, do not restate the wording. The wording is the proposal's replacement. A session that refuses to read "history" will not have the decision, and will apply the compressed digest in the instructions instead.
- **Restated by the author.** The proposal's "Put this in its place" is then the wrong text. Call 20's proposal kept a narrowed neglect paragraph. The author dropped the neglect argument from the short sheet. The instructions already allow the book to expound it, and they already forbid treating the short-file cut as an inline deletion. A session that pastes the proposal will think it is applying the decision. Call 21's proposal offered "in time others occupy what those members no longer keep." The author's decision is the care order, and the sheet follows the author. Call 22's proposal allowed a person to refuse to kill while still naming the enemy. The author refused that softening. The instructions' enemies paragraph follows the author. Chapter 8's Moriori passage argues that an absolute refusal of violence fails. Bringing that passage back to the proposal would undo the decision.
- **A word on which the decision and the sheet differ.** Decision 24 says the children continue "us" and the particular parental lineage. The sheet's self-interest paragraph says they continue him and that lineage. Both are applied records. Choosing one in a chapter, in silence, writes a new decision.

Recommend three lines. Where the decision says accepted as suggested, or applied as suggested, the proposal's replacement is the decision, and the short sheet is the check that those words were the ones applied. Where the author restated the decision, the author's words stand, including against that call's own proposal. Where the author's sentence and the sheet differ by a word, write a call and leave the chapter.

## 6. The bones file is a lock list with older glosses

The instructions say to read `notes/00_civ_authoritative_bones.md` for the same locks in short form. The five sentences and the seven names match. The civilization gloss does not match the sheet. Bones still says civilizations comprise customs and values, that civilization is not an entity and not a collective, and that it is the prerequisite technology. The sheet says a technology is a repeatable method together with the structure that method keeps in being, and that civilization consists of will and method and is not a living entity or a collective mind.

Recommend the bones file be the lock list for the five sentences and the seven names. On any other sentence, the short sheet and the author's decisions are the record. A gloss in bones is not a second sheet.

## 7. The inline digest can be used as a shortening

The digest says a longer exposition of the same decision may stay, and that a chapter section is not to be replaced by the sheet's paragraph. Several compressions are short enough to be pasted over that exposition.

- "Civilization is will plus method. It is not a collective." The sheet says civilization consists of those two parts, and that it is not a living entity or a collective mind.
- "A shared work lasts when a named person or a named office is charged with it." The sheet has "only," and it has "answers for" as well as "charged with." The "only" is the claim.
- "Drop any barren-spouse nuance that displaces the lineage." Decision 21 drops "capable, virtuous" and "replaced or conquered." Decision 24 states the lineage and refuses a generic duty to Humanity. Neither decision records a barren-spouse cut. Call 24's proposal kept a spouse who will not bear children inside the household. The sheet does not discuss that case. A chapter sentence about it is uncovered, so it is a call.
- The Form and listen paragraph merges the sheet's why, range, and fence. "Bring it to the short sheet" can be read as an instruction to paste that merge.

Recommend the digest stay a finder. The sentence that changes is the sentence that still states the rejected form. The replacement uses the author's words or the sheet's words for that form, inside the paragraph that is already there.

## 8. The same strings do other work

The inline list gives rejected phrases to correct: "fixed" where it overclaimed, "the next morning," "mere property," "pacifism" as the proposal used it, and the population and household phrases. Several of those strings appear where they are not the rejected claim.

- Chapter 6 quotes an apologist: human life is more important than mere property. The chapter then argues that there is no separation between human life and built property. That is the decision, at chapter length. Deleting the quoted phrase, or replacing the section with the sheet's sentence, cuts the argument the instructions say to keep.
- Chapter 11 uses "the next morning" as a time of day, in the rat passage and in the Challenger passage. The rejected immediacy is the population claim.
- "Fixed" is the ordinary word in chapter 4's neurological limits and rules of evidence, in chapter 6's obligation fixed by reality and tier, and in chapter 12's name for correspondence. Chapter 2 also says reality is "objective, fixed in its properties." That clause is the overclaim decision 19 corrected. The word is not banned. The overclaim is.
- "Generational" is already protected for the criterion, the horizon, and generational theft. The other rejected strings need the same protection.

Recommend one rule. A rejected phrase is an inline fix when it does the work the decision rejected. The same words in a quotation of the view being answered, in a time stamp, or in an ordinary sense, stay. A doubtful case is a call.

## 9. The known-passage list can be read as the whole job

The list says to apply those passages when the review reaches them, and that they are not new calls. It does not say it is a head start. The general rule already says any conflict with a recorded change of form is an inline fix. Chapter 2's "fixed in its properties" is one. A session that treats the list as closed will leave it, or will handle only chapter 1 and chapter 8 with special care.

The chapter 8 bullet says to check the tier list, honesty, the children rather than the wife, the tax-as-enmity line, and the one-man case. "Check" plus "apply" can become a rewrite of the tier section into the sheet's paragraph. The chapter's enemy sentences use "dispossess" and "campaign," and they name criminal violence. Whether that already covers one man forcing his way into a house is a doubtful match. The general rule says a doubtful match is a call. The known-passage bullet does not yield to that rule.

Recommend the list say it is a head start. For chapter 8, absence of the rejected form ends the check. The decided verbs are rob, subjugate, or destroy, and the definition includes one man as well as a program. A near synonym, or a sentence that may already include the one-man case, is a call. The longer Turkish measures stay. No new history is added.

## 10. "Millions starved" is broader in chapter 2 than the clause the author judged

The author read the Lysenko section and kept "millions starved" at that level of nuance. The derive sentence the author judged is about the Soviet Union in the 1930s and 1940s. Chapter 2 states the clause, and the same result covers Lysenko's decrees and their later imitation in Maoist China during the Great Leap Forward. The instruction "if a chapter states it, leave it" can be read as a ban on calling that wider cause.

Recommend the keep-rule stay on the clause. Do not add "millions starved" where a chapter does not state it. Do not delete it where a chapter does. A sentence that joins a further cause to that result is a new call, and the passage stays, including the clause.

## 11. The silent register fix is wider than the examples

The file allows a register breach that does not change the claim to be fixed in that sentence: an em-dash, an exclamation mark, or a banned figurative use. The preamble says unsolicited prose does not enter the book. "Banned figurative use" can mean the whole of the editing guide: honest, load-bearing, raw, rot, zero, on earth, real-world, imperative openers, and the rest. That is a style pass. The file also says not to open one, and not to edit a sentence the reviewer would merely have written differently.

Recommend a closed list for silent edits: the em-dash, the double hyphen used as a dash, and the exclamation mark, and only inside a sentence already being changed for a recorded decision. Every other register breach is a call, or it waits. "Load-bearing stone wall" in chapter 7 stays, as the file already says.

## 12. One-directional Form and listen reopens the rename

The rename is done, and the file says not to do that pass again. The next sentence says that where a later sentence still treats the virtue as one-directional, bring it to the short sheet. Any sentence that mentions forming without repeating listening can be swept into that pass.

Recommend the second sentence mean a sentence whose own claim makes the virtue one act. A sentence that names one act, beside a passage that already states both, stays. "Bring it to the short sheet" means correct the relation in that sentence. It does not mean paste the sheet's paragraph.

## 13. The calls file needs a place for the near miss

The call shape is right: where, wording, problem, what to do, status not applied. It does not ask the reviewer to say which recorded decision almost covered the sentence. The author then re-derives the match. Later, a decision recorded at the top of the file can sit above a body that still says "not applied," and a later session can apply the body, or miss the top.

Recommend one more line in each call: the nearest decision, by note and number, and why it does not clearly cover this sentence, or none. The call body remains the proposal. Only the author's words at the top of `civ/notes/10_chapter_calls.md` authorize an edit. When that edit is made, the status of that call becomes applied, and the status names the chapter sentence that changed. The proposal text stays as it was written.

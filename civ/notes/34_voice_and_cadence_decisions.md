# 34. Voice and Cadence Decisions: Evaluation and Responses

Recorded 2026-11-23 and corrected the same evening after a second audit (section 9). This covers the 37 decisions the author made on 2026-11-22 and 2026-11-23 on two packets: `20261122-213928-full-book-voice-plain-style.json` and `20261123-002333-cadence-voice-upgrades.json`. One item is applied (07_ch07.md). The 15 holds have answers, and 27 new items wait for the author in the viewer. No other chapter text changed. Sources are in `ext-ref/red_green/`, `ext-ref/kindness_politeness/`, `ext-ref/decline/`, `ext-ref/not_authentic_doctrine/`, and `ext-ref/hume/`.

---

## 1. The decisions in numbers

| Rule | Accept | Reject | Hold for research | Total |
|---|---:|---:|---:|---:|
| `voice/plain-style` | 1 | 12 | 11 | 24 |
| `cadence/snap` | 0 | 9 | 4 | 13 |
| **Total** | **1** | **21** | **15** | **37** |

Twenty of the 21 rejects carry a note; only 11_ch11.md:115 (`voice/plain-style`) was rejected without one.

The scores did not predict the outcome. The voice score of the one accepted draft (+1.42) lies inside the range of the rejected drafts (1.37 to 2.93, median 1.52). The held drafts scored a little higher (1.40 to 3.19, median 1.85), so a high score points to a section that needs work, not to a draft the author will take. Snap separates nothing: the rejected cadence drafts scored 27.7 to 46.9 (median 33.5) and the held ones 28.9 to 44.7 (median 32.5).

This answers decision 1 of note 33. Of the eight rewrites in note 33, section 5, six were rejected and two held.

## 2. What the notes ask for

Read together, the notes give standing instructions. Later drafts should follow them without being told.

1. **Literal language.** No metaphors, including mild and conventional ones. Six notes say so: "No metaphores!!!", "No metaphores! Use exact language.", "Floor and ceiling are metaphores. Rejected.", "introduces metaphorical lanugage and removes concrete language", 'you used metaphores with "seminar rooms"', "remove this metaphore and use direct language". (Spellings are the author's.) This is also Section 2 of the editing guide.
2. **Keep words that carry a specific meaning.** Nine notes object that a draft replaced such words or lost accuracy: "removing critical words and phrases that have specific meaning", "replacing words that have specific needed meanings with less or different meanings", "The specific words and phrases in the original are important", "These words have a specific meaning which, when changed, change the meaning", "less acurate and more fuzzy", "omit meaning". The notes name three such words: *revealed preference*, *correspondence truth*, and *vector sum*.
3. **Prefer small changes to wholesale replacement.** "I would be open to more fine grain set of suggestions", "the wholesale replacement is not as good", 'I like the punch of "What kind of creature is a human being?" but not much else'.
4. **Keep what carries warmth and signposting.** The second person is wanted ("a brother" against "your brother"), and so are words like *Moreover* and *Therefore*.
5. **Three notes question the original passage, not the draft.** Shinto (12_ch12.md:141), the Generational Criterion (13_ch13.md:5), and the paragraph on refusing authority (09_ch09.md:75). Those passages may need work on their own account.

## 3. Applied

**07_ch07.md:116**, `voice/plain-style`, accepted 2026-11-23: the Hangar A and Hangar B passage. The tool wrote `packets/20261123-173929-applied.json`.

- `lint sim` before applying: one unique match, no overlap, no new Part A or spelling hits.
- `civ/bin/lint` after: the standing list of findings is unchanged, and none of its lines falls in the edited passage.
- `civ/bin/linear check` on lines 112 to 132, before and after: the worst sentence in the passage fell from 3.0 to 2.0, and close echoes from 50 to 41 per 100 sentences.

## 4. Filed for the author

| Packet | Contents |
|---|---|
| `packets/20261123-194240-hold-answers.json` | Answers to the 15 holds. Each says what was done, which new item holds the draft, and what is left to the author. |
| `packets/20261123-194240-note-followups.json` | 22 open items written from the author's notes: 10 from holds and 12 from the notes on rejected drafts. Rule names begin `note/`. |
| `packets/20261123-194451-research-drafts.json` | 5 open items that needed sources or a restructure: 08_ch08.md:125, 11_ch11.md:3, 13_ch13.md:135, 14_ch14.md:19, 14_ch14.md:38. |

All three passed `review add -strict`, so every quotation matches once in its chapter. `lint sim` reports no new Part A or spelling hits for either packet of drafts. The queue now holds 270 items: 27 open, 20 holds (all answered), 195 applied, 27 closed, one older accept still pending apply, and 6 stale.

The three packets were filed first at 18:43 to 18:44 UTC and re-filed, corrected, at 19:42 to 19:45 UTC (section 9). The item identifiers did not change.

Nothing was applied from the new packets. Each waits for the author in the viewer.

## 5. The fifteen holds

| Passage | The note asks for | Filed | Left to the author |
|---|---|---|---|
| 06_ch06.md:63 Harm-reduction | Proposed is better; replace its metaphors | The proposal with five figures made literal and seven terms restored (`note/literal-wording`) | Accept, then withdraw the hold |
| 08_ch08.md:86 Dispositional love | Last sentence belongs to the previous paragraph; open on "For those reasons" | Your wording, sentence moved (`note/paragraph-move`) | The same |
| 08_ch08.md:125 Kindness to strangers | Problem first is improved; needs examples, then a general rule | Three sourced paragraphs and a closing rule (`note/examples-and-rule`, 124 to 619 words) | Two cautions are written in (section 7) |
| 10_ch10.md:34 Have children | Easier to read but less accurate | Your text split at the three colons (a); the same with the library sentence rewritten (b) | Choose a or b |
| 11_ch11.md:3 Chapter 11 opening | Draw ideas from examples: Greece, an area of Eastern Europe, the Shakers, Zimbabwe or South Africa, Nauru, Rome; apply theories of decline | Four writers, four records, then your closing paragraphs (a, 969 words); the same with two records (b, 732 words) | Which records to keep; "Quakers" read as the Shakers |
| 11_ch11.md:143 Social parasitism | Your wording, individuals only | The two corporate sentences removed (`note/individuals-only`) | Corporate rent-seeking now has no section |
| 11_ch11.md:81 The five stages | Keep your words; add that people fail to be unkind without any activist | One paragraph after the fourth stage (`note/own-unkindness`) | Nothing |
| 12_ch12.md:69 Rousseau to Derrida | Hegel: "correspondence truth"; the original Marx paragraph | Both changes, Marx in his own paragraph (`note/terms-restored`) | Name any dropped substitution that is wanted |
| 12_ch12.md:169 Dugin | Streamline, but keep the critical words | Sentences split, one filler sentence removed, no term changed (`note/streamline-keep-terms`) | The same |
| 12_ch12.md:3 The five questions | A test of compatibility; compatible groups can still be at odds | One sentence replaced, 272 to 288 words (`note/compatibility-test`) | Define "self-defeating" and "enemy within"? |
| 12_ch12.md:61 Hume | Remove the study figure; direct language | He earned his living by writing, sourced to *My Own Life*; the *History* dated 1754 to 1762 as the *Stanford Encyclopedia* dates it (`note/direct-language`) | Nothing, unless another date is preferred; Hume says only that he gave the last two volumes to the public in 1761 |
| 13_ch13.md:39 Lee Kuan Yew | Reasonable edits mixed with changes of meaning | Terms kept, three sentences put in a forward order (`note/keep-terms`) | Nothing |
| 13_ch13.md:135 Red-Green | Sources for the five points; incompatibility and how to detect it; no prognosis | Your five points kept by name and in order, each with a finding on record and its limit (a, 1,124 words); one record per mechanic (b, 734 words); then the incompatibility and four questions for detection | See section 7; whether to name the organization in the third mechanic |
| 14_ch14.md:19 Tool 1 | A rework larger than rephrasing | One direction: the two devices named, each point tied to the evasion it closes, the audit turned on the speaker (`note/tool1-rework`) | Is a real case wanted in place of the hypothetical? |
| 14_ch14.md:38 Tool 2 | The content is off; the test is correspondence | The doctrine's scope first, then correspondence; citations made exact (`note/tool2-rework`) | Whether Islam is named as a doctrine of the first kind |

## 6. Follow-ups from the notes on rejected drafts

A rejected draft stays closed and is not raised again. Where the note itself asked for a specific change to the original, that change is a new item with a new quotation.

- **01_ch01.md:37 and :41**, from the note on `voice/plain-style`: the colon ("nothing keeps itself"), and "wet" dropped.
- **04_ch04.md:9**: "What kind of creature is a human being?"
- **04_ch04.md:119**: "from your specific mother and father". Two further items make "receptacle" and "garment" literal. They are optional: those figures are in the original sentence.
- **05_ch05.md:9**: "human limits require", or "limited time and attention require" (two alternates).
- **05_ch05.md:96**: the Reinvention Test cut, the closing paragraph kept. The alternative is to cut the closing paragraph too.
- **06_ch06.md:79 to :83**: four sentence-level suggestions that keep "revealed preference".

## 7. Corrections to the existing text, found while researching

1. **14_ch14.md:42** says that Article II, the federal bribery statutes, and the judicial oath "explicitly forbid accepting bribes in explicit statutory text". None of the three does that in that form. Article II, Section 4 makes bribery a ground for removal from office; 18 U.S.C. § 201 makes it a crime; the oath (28 U.S.C. § 453) binds the judge to administer justice without respect to persons.
2. **13_ch13.md:135 to 153** describes one Red-Green alliance with five "operational mechanics". No source examined documents a single coordinated alliance, a blockade carried out by a left and Islamist coalition together, or entry through diversity mandates. The author's own note of 2026-07-16 (`book1-5/notes/red_green_alliance.md`) states the matter as it should be stated: a tendency with mechanisms, with examples as illustrations. The draft keeps the five mechanics and gives each a finding on record. Four statements of the original are left out for want of a record: "street intimidation" and "demographic mobilization"; the second language as "conquest, demographic replacement, and religious supremacy"; the blocking of highways and bridges; and entry through "diversity and inclusion" mandates. One finding must travel with its limit: the Election Court found that the Mayor of Tower Hamlets "relied on silencing his critics by accusations of racism and Islamophobia", and in the same judgment that it had "not heard a shred of credible evidence" linking him to any Islamist movement.
3. **12_ch12.md:7** still names "neo-feudal warlordism". Note 24 found that it is not Haywood's program and cut it from section 3A.
4. **08_ch08.md:125** says that opportunists convert temporary provisions into permanent support. In the record the converting is done by legislators: a thirteen-week program of 2008 was amended eleven times and paid benefits until December 2013, and a food benefit created in 2020 was extended by four later acts, ended in 2023, and was followed by a permanent summer program from 2024.
5. **12_ch12.md:61** gives no dates for Hume's *History*. The new draft adds them. The *Stanford Encyclopedia* gives 1754 to 1762, in six volumes; Hume says only that he gave the last two volumes to the public in 1761.

Facts that came up in research and should not be used as they are usually stated:

- The American Secretary of State said on the record (Istanbul, 15 July 2011) that United Nations Resolution 16/18 calls on states to protect freedom of religion and "not to criminalize speech unless there is an incitement to imminent violence". The resolution is not a blasphemy instrument. The live page of the transcript returns an error page to scripts; the copy in `ext-ref/red_green/dl/` is an Internet Archive capture of 2013.
- The Danish law of 2023 on "inappropriate treatment" of scriptures was not repealed in 2024. The research report records that a motion to repeal it was rejected on 1 October 2026; the Parliament's vote page could not be opened to confirm this.
- Aristotle's Sparta is a case of inheritance law concentrating land, not of a refusal to have children.
- No *national* fertility decline in Eastern Europe should be claimed for the 1800s. The documented case is a region, the Ormánság in south-western Hungary, where families kept to one child from the late eighteenth century and the region lost people: in its seven studied villages the population rose by five percent between 1870 and 1941 while the population of Hungary rose from 188 to 350 on the same index, and six of the seven shrank (`ext-ref/decline/report.md`, C.7.7).
- The National Retail Federation withdrew its claim that organized retail crime was nearly half of $94.5 billion (per the research report; the page reporting the withdrawal could not be saved).

## 8. The two tools in light of the decisions

**Cadence (`Snap`).** No cadence draft was accepted, and the notes give the reason. The score rewards concrete words, words learned early, words with many senses, and few Latinate words. Those are the steps by which a term of art becomes an image: *jurisdiction* becomes the gavel, *magistracy* the bench, *sentiment* the cheers of a crowd. A high score can therefore measure how far a paragraph has left exact language. The editing guide, Section 2, forbids the trade the score rewards.

Recommendation: do not queue `cadence/snap` drafts. The tool can stay as a reading list of the muddiest paragraphs. If it is to score drafts, it needs two guards before any change of weights: no credit for replacing a principle term or a capitalized, italicized, or quoted term, and no credit for a concrete word used outside its dictionary sense. The weights are the author's to set (note 33, decision 1).

**Voice.** One draft of 24 was accepted, and 11 were held with substantive instructions. The score finds sections that need work. The rewrite recipe is what the author declines: it drops signposts, shortens sentences wholesale, and removes the second person.

Recommendation: keep `lint voice` as a locator, and write its drafts as sentence-level suggestions that keep every term, in the way the author asked for on 06_ch06.md:79.

**Note 33, decision 2 (Phase 3).** The results support Option A, to stay with the pure Go tools and add nothing, until the scoring itself is changed.

## 9. Checks on the new items

**At the first filing (18:43 UTC).**

- Every quotation and figure carried into the five sourced drafts was read a second time in a saved copy of the source, after the research reports. Three reports by research assistants were used. Three pages saved for one of the reports were challenge pages rather than the document, and the paragraphs the draft quotes from that document were then read in full and copied by hand.
- The 27 items held 83 quoted spans. 51 were matched mechanically against the saved sources. The other 32 lie inside passages of the author's own that the drafts carry over unchanged.
- Two errors in my drafts were caught before filing: a misquotation of the Election Court ("not a shred" for "not heard a shred"), found by the mechanical check, and one word mistyped inside a quotation from Olson, found on re-reading.
- `lint sim` on both packets: no new Part A or spelling hits. It also caught "Programme" and "insoumise", which are now avoided.
- `civ/bin/linear check` on the drafts: sentences that carried two quotations were split, and ambiguous pronouns named. Sentences of the author's own that the drafts carry over were left as they are.

That audit covered the alternates of the drafts. It did not cover the hold answers or the problem texts, and it did not test whether a saved page was the page named.

**At the second audit (19:00 to 19:45 UTC), after the author's remarks.** Every page saved on 2026-11-23 was checked by its own page title. Every quoted span in the three packets was matched against the saved copies or against the author's own text: 179 spans, of which 72 are in a saved source and 104 are the author's own words, and none is unmatched. Each unquoted date and figure in the answers was searched for in its document. The three packets were then re-filed.

The second audit found these errors in the first filing. All are corrected.

1. A police crowd figure (10,500) in the Red-Green draft was in no report and in no saved file. Removed.
2. The file saved as the Secretary of State's Istanbul speech was the State Department's error page. The speech is now saved from the Internet Archive, and the sentence quoted from it holds.
3. A file saved as Kelly 2021, in `ext-ref/decline/`, was a bot challenge. Removed. No figure rested on it.
4. The answer on the chapter 11 hold spoke of a national fertility decline. The author's note asks for areas that decided not to have children. The answer is corrected, and the draft now gives the region's population figures.
5. The draft dated Hume's *History* 1754 to 1761, and the answer spoke of "the title pages". Neither was in anything saved. The dates are now the *Stanford Encyclopedia*'s (1754 to 1762), and the claim about title pages is withdrawn.

Made exact at the same time: the coin below five percent silver is the antoninianus, the double denarius, not the denarius; the figure of about four-fifths silver after Nero is one re-analysis, the older figure being about nine-tenths; the Dallas Fed report is the 2011 Annual Report; the American food benefit is dated 2020, not March 2020; and the finding on Seattle is the review panel's. The Red-Green draft was rebuilt on the author's own model (section 7, item 2).

## 10. Sources

- `ext-ref/red_green/`: report and documents for 13_ch13.md. The addendum at the head of the report records the corrections and points to the author's notes in `book1-5/notes/`.
- `ext-ref/kindness_politeness/`: report and documents for 08_ch08.md. `dl/examined_not_used/` holds what was examined and left out.
- `ext-ref/decline/`: report and documents for the 11_ch11.md opening. Section C.7.7 of the report gives the population figures of the Ormánság.
- `ext-ref/not_authentic_doctrine/`: the statutes and verses for 14_ch14.md.
- `ext-ref/hume/`: *My Own Life* and the *Stanford Encyclopedia* entry, for 12_ch12.md.

Each folder lists the pages that could not be saved.

## 11. Left for the author

1. Decide the 27 open items in the viewer.
2. Withdraw `02_ch02.md:a676ed11:prop/boiled-water-source`. The crop-rotation passage was replaced on 2026-11-22 in two steps. The first-round item `02_ch02.md:d3dd0e64:prop/granary-vary` (accepted at 04:28 UTC) put an invented river town in its place. `boiled-water-source` (accepted at 14:04 UTC, with the note to remove the made-up example) was held back, and `boiled-water-real-only` (accepted at 14:24 UTC) replaced the river town with the real-life text, which is in the chapter word for word at 02_ch02.md:103. `boiled-water-source` is still marked "accept, pending apply", but its passage no longer exists, and applying it would bring the invented town back. The one thing in its text that the chapter lacks is a citation of Wang Shixiong's *Huoluan Lun* (1838), which was never saved or checked.
3. Five older holds are stale (their passages were replaced). They can be withdrawn to tidy the counts.
4. Chapter 11: the two mechanisms without a case in the record are the Compassionate Delusion and fatalism.
5. 12_ch12.md:141 (Shinto), 13_ch13.md:5 (the Generational Criterion), and 09_ch09.md:75 (refusing authority): the notes question these passages themselves.
6. Whether to change the cadence scoring, as in section 8.
7. In the Red-Green draft, the organization of the third mechanic is described in the Election Court's words and not named. Its name, the English Defence League, carries the one British spelling the book would then have.
8. The labor board's memorandum on the Google dismissal is quoted through two reports of it. If that sentence is kept, the memorandum itself should be obtained (case 32-CA-205351).

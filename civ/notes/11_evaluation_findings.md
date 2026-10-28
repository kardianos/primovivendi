# Evaluation of Recent Edits and Full Re-Evaluation of Essays & Book

**Date:** 2026-10-27  
**Scope:**
- **Essays:** `civ/ccp_sheet.md`, `civ/ccp_derive.md`, `civ/ccp_eng.md`
- **Book:** `civ/chapters/01_ch01.md`–`14_ch14.md` and `civ/ccp_book.md` (verified identical to the concatenated chapters)
- **Reference Standard:** `civ/notes/06_editing_guide.md` (1860s/1950s precision treatise register, deterministic pattern audit, n-gram audit, and structural rules)
- **Git Range Evaluated for Recent Edits:** `3482b6fc` (`"update book to not use silly metaphores so much"`) through `a410a0a8` (`"adjust book for consistency and factual basis"`)

---

## Executive Summary

The recent commits (`3482b6fc` through `a410a0a8`) substantially improved the analytical rigor and historical grounding of the essays and book:
- The deterministic audit script in `06_editing_guide.md` (Part A) now reports **0 violations** on `ccp_sheet.md`, `ccp_derive.md`, and `ccp_eng.md`, and only **5 minor regex hits** across all 14 book chapters.
- Key conceptual updates from `07_ccp_sheet_calls.md`, `08_eng_derive_calls.md`, and `10_chapter_calls.md`—such as renaming Virtue 7 to **Form and listen**, sharpening the definitions of technology and civilization, qualifying the revelation allowance, distinguishing honesty from Tier 5, adding the confiscatory-taxation boundary, replacing cartoonish metaphors, and grounding historical claims in verifiable sources (Turkish *Varlık Vergisi*, Culver City High School, Shakers, Wolowyna et al. 2020 demographic data)—are largely in place.

However, a line-by-line review of the diffs and full text reveals **six categories of issues introduced by the recent edits** (Part 1 below) and **five categories of broader issues across the essays and book** (Part 2 below) that should be resolved.

---

## Part 1: Evaluation of Recent Edits

### 1.1 Meta-Editorial / Reviewer Commentary Leaked into Book Prose

In three places, language from the editorial review notes (`09_chapter_review_instructions.md` / `10_chapter_calls.md`) was pasted directly into the manuscript, breaking the treatise voice by referring to the previous draft's wording or deleted metaphors:

1. **Meta-rebuttal of the previous draft's "ten million" figure (`civ/chapters/13_ch13.md:111`)**
   - **Current text:**
     > `"A 2020 demographic reconstruction puts excess deaths for the Soviet Union in 1932–1933 at about 8.7 million, of which about 3.9 million were in Ukraine, about 3.3 million in Russia, and about 1.3 million in Kazakhstan (Wolowyna and others, *Journal of Genocide Research*, 2020). The count is in the millions. It is not a round ten million pasted over three regions."`
   - **Issue:** *"The count is in the millions. It is not a round ten million pasted over three regions."* is editorial commentary from `10_chapter_calls.md` (Call 13) criticizing the earlier draft's phrase (`"more than ten million human beings starved to death across the Volga, Ukraine, and Kazakhstan"`). A reader of the finished book has never seen the previous draft and will be puzzled by who "pasted a round ten million over three regions."
   - **Suggested fix:** Remove the two meta-sentences (`"The count is in the millions. It is not a round ten million pasted over three regions."`) and let the demographic citation lead directly into `"Physical reality is indifferent to revolutionary slogans..."`

2. **Meta-reference to a deleted metaphor + colon fragment (`civ/chapters/11_ch11.md:143–144`)**
   - **Current text:**
     > `"* They continue, when they continue, by a simpler route than a metaphor. Households that do not hold the ideal bear, feed, and form the children. The movement then takes those children at school age."`
     > `"* The capture is institutional: **the teacher-training colleges, the school curricula, the youth media, and the universities**, so that the builders' children are formed against the household that raised them."`
   - **Issue:**
     1. *"by a simpler route than a metaphor"* refers to the biological-virus analogy that was deleted from this section in commit `a410a0a8`. In the finished text, no metaphor was just mentioned.
     2. In line 144, `"The capture is institutional: **the teacher-training colleges, the school curricula, the youth media, and the universities**, so that..."` attaches a noun list and a `"so that"` clause to a colon without a governing verb.
   - **Suggested fix:**
     - Line 143: `"* Such movements continue only by drawing upon outsiders. Households that do not hold the ideal bear, feed, and form the children, and the movement recruits those children at school age."`
     - Line 144: `"* The capture operates through institutions—**the teacher-training colleges, the school curricula, the youth media, and the universities**—so that the builders' children are formed against the household that raised them."`

3. **Unanchored reference to "The studies" and awkward hyphenation "about-figures" (`civ/chapters/04_ch04.md:66`)**
   - **Current text:**
     > `"The studies find a limit, and they find it in layers. A person holds roughly five people in the inner support clique, roughly fifteen in the sympathy group, roughly fifty in the band, and roughly one hundred and fifty in the active social network (Dunbar, 1992; Zhou and others, 2005). These are about-figures, not a statute."`
   - **Issue:** *"The studies find a limit"* has no prior antecedent for *"The studies"*, and *"about-figures"* is a non-standard coinage that clashes with the 1860s/1950s formal register (`06_editing_guide.md` Section 2).
   - **Suggested fix:** `"Anthropological and cognitive studies of human social networks identify a structural limit arranged in layers: a person maintains roughly five people in an inner support clique, roughly fifteen in a sympathy group, roughly fifty in a band, and roughly one hundred and fifty in an active social network (Dunbar, 1992; Zhou and others, 2005). These numbers are empirical approximations rather than statutory boundaries, and no layer above one hundred and fifty functions like a household."`

---

### 1.2 Colloquial Slang and Choppy Transitions Introduced in Edits

1. **"Slop shop" and "pencil whipping" in Chapter 4 (`civ/chapters/04_ch04.md:109`)**
   - **Current text:**
     > `"* **In cultural and institutional roles:** Assume you can take any fifty thousand people from anywhere on earth, regardless of whether they were raised in a culture of high-trust law-abidingness, punctual craftsmanship, and honest testimony or a culture of clan nepotism, endemic bribery, and blood feuds, and place them into a high-trust municipal administration or aviation maintenance facility with identical results. The recruit can be physically able to tighten the bolt. The failure is that the shop's standard is not his. Many times he will not hold the torque the procedure names. He will sign the card for a step he did not run (Eastern Airlines mechanics described this as "pencil whipping" in 1988, before the federal indictment over falsified maintenance logs). He will keep a slop shop, and tools will break at a higher rate. The difference in cultural formation shows up on the shop floor."`
   - **Issue:**
     - Violates `06_editing_guide.md` Section 2.1 (No Casual Slang or Colloquial Idioms): *"slop shop"* (historically a shop selling cheap ready-made clothing; colloquially a messy shop) and *"pencil whipping"* are slang.
     - The transition from `"fifty thousand people... in a high-trust municipal administration or aviation maintenance facility"` to `"The recruit can be physically able to tighten the bolt"` is abrupt and telegraphic (seven short sentences in a row with flat cadence).
   - **Suggested fix:** Rewrite in the treatise's formal engineering cadence while preserving the 1988 Eastern Airlines maintenance-log indictment example:
     > `"* **In cultural and institutional roles:** Assume you can take any fifty thousand people from anywhere on earth, regardless of whether they were raised in a culture of high-trust law-abidingness, punctual craftsmanship, and honest testimony or a culture of clan nepotism, endemic bribery, and blood feuds, and place them into a high-trust municipal administration or aviation maintenance facility with identical results. A recruit from a low-trust culture may possess the physical strength to tighten a bolt, yet lack the internal habit of adherence to written specifications. When the shop's standard is not his own, he frequently fails to apply the specified torque, leaves the workshop disordered so that tools are damaged or misplaced, and signs inspection cards for maintenance steps he never performed—a practice that led to the 1988 federal indictment of Eastern Airlines supervisors and mechanics for falsifying safety logs. Differences in cultural formation register directly on the shop floor."`

2. **Choppy telegraphic cadence in other recent insertion blocks**
   - Several blocks added in commits `cd8cb49d` through `a410a0a8` use strings of 4-to-7-word staccato sentences that depart from the treatise's measured 1860s/1950s periodicity (`06_editing_guide.md` Section 2.3):
     - `civ/ccp_sheet.md:97`: `"That lineage is real. His responsibility begins with himself and with his own. He has no generic responsibility. He values human life broadly and in particular. Human life is the floor of goodness."`
     - `civ/ccp_sheet.md:274` / `civ/chapters/10_ch10.md:70`: `"Elders form those younger than themselves. The duty is broad. It includes kin. It also includes politicians and other leaders, who train the people younger than themselves. Those who are younger seek the experience of the people around them. They listen. They test what they are told. They do not have to agree with all of it."`
     - `civ/chapters/12_ch12.md:134`: `"Test the ideal on its own terms... Continuance then depends on people who do not attain the ideal, or who set it aside. Jain mendicants are the same shape: the highest form does not pick the fruit or till the soil, and lay households do that work so the mendicant can keep the rule. A root that survives only by being broken is not a root."` (Note also `"Jain mendicants are the same shape"`—colloquial phrasing.)
   - **Suggested fix:** Combine related short clauses with proper subordination and coordination so the inserted passages match the surrounding prose register.

---

### 1.3 Duplicate / Redundant Sentences Created by Edits

1. **Duplicate definition of "enemy" six lines apart in `civ/ccp_eng.md:393` and `399`**
   - **Line 393:** `"An enemy is any person, or any body of persons, who acts to rob, subjugate, or destroy you, your household, or your people."`
   - **Line 399:** `"An enemy is any person, or any body of persons, who acts to rob, subjugate, or destroy you, your household, or your people."`
   - **Suggested fix:** Keep the full definition in line 399 (where the two-part test of enmity is formally stated) and tighten line 393 to: `"In physical reality, predatory actors exist—from the solitary violent criminal to the organized foreign army or revolutionary movement seeking to rob, subjugate, or destroy a peaceful people."`

2. **Duplicate definition of "enemy" in `civ/chapters/08_ch08.md:169` and `204`**
   - **Line 169:** `"An **enemy** is any person, or any body of persons, who acts to rob, subjugate, or destroy you, your household, or your people. The act may be one man forcing his way into a house. It may be a state carrying out a program against a people."`
   - **Line 204:** Under `"This historical record establishes three foundational truths about Principle 4"`, item 1 repeats the exact same two sentences verbatim (`"1. An enemy is any person, or any body of persons, who acts to rob, subjugate, or destroy you, your household, or your people. The act may be one man forcing his way into a house. It may be a state carrying out a program against a people."`). Moreover, the Turkish state record does not illustrate `"one man forcing his way into a house"`.
   - **Suggested fix:** In line 204, state the specific lesson of the historical case: `"1. **Enmity is defined by destructive action, including by a state against its own subjects:** When a regime uses military conscription and confiscatory taxation to strip and break a distinct people within its borders, it acts as their enemy regardless of its legal formalities."`

3. **Duplicate sentence on depicting a wound/ruin six lines apart in `civ/ccp_eng.md:610` and `616`**
   - **Line 610:** `"Depicting a wound or a ruin truthfully is a different act from celebrating it."`
   - **Line 616:** `"Depicting a wound, a ruin, or a deformed body truthfully is a different act from celebrating it."`
   - **Suggested fix:** Remove the duplicate sentence at line 610 and retain the fuller sentence at line 616 where the diagnostic distinction is explained.

4. **Duplicate sentences on "Elders form those younger" four lines apart in `civ/chapters/10_ch10.md:70` and `74`**
   - **Line 70:** `"Elders form those younger than themselves... Those who are younger seek the experience of the people around them. They listen. They test what they are told."`
   - **Line 74:** `"Therefore, civilization requires **Virtue 7: Form and listen**. Elders form those younger than themselves. The younger listen and test what they are told."`
   - **Suggested fix:** Consolidate lines 68–74 so the definition of Virtue 7 is stated cleanly once after the million-infants illustration.

---

### 1.4 Unintegrated & Out-of-Sequence Insertions of the "Particular Lineage / Floor of Goodness" Block

1. **Out-of-sequence reference to Principle 3 ("floor of goodness") inside Principle 2 in `ccp_derive.md` and `ccp_eng.md`**
   - **Locations:** `civ/ccp_derive.md:72` (Section 3, deriving Principle 2) and `civ/ccp_eng.md:170` (Chapter 2, Principle 2).
   - **Current text:** `"Having children matters because the children continue him and continue his particular parental lineage. That lineage is real. His responsibility begins with himself and with his own, and he has no generic responsibility. Human life remains the floor of goodness, and he values human life both broadly and in particular."`
   - **Issue:** Both `ccp_derive.md` and `ccp_eng.md` are explicitly structured as **strictly sequential derivations** in which each principle uses only premises already established. At Section 3 (`ccp_derive.md:72`) and Chapter 2 (`ccp_eng.md:170`), *goodness* (Principle 3) and *tiered responsibility* (Principle 4) have not yet been introduced or derived! Saying `"Human life remains the floor of goodness"` before the question *"What is good?"* is even posed in the next chapter breaks the deductive sequence.
   - **Suggested fix:** In Principle 2 (`ccp_derive.md:72` and `ccp_eng.md:170`), keep the observation strictly descriptive of human particularity and self-interest (e.g., that a human being's care is anchored in his own life, his spouse, and his own biological children who continue his particular lineage), and place the normative statement (`"Human life is the floor of goodness, and a person values human life both broadly as the floor and in particular through his own lineage and kin"`) in Principle 3 / Principle 4 and Virtue 6.

2. **Mismatch under "Range of views" in `civ/ccp_sheet.md:97`**
   - **Location:** `civ/ccp_sheet.md:97` (`Principle 2 -> Self-interested -> Range of views`).
   - **Current text:** `"* *Range of views*: There is legitimate debate over where the boundary of "self-interest" ends and obligation to others begins. Having children matters because the children continue him and continue his particular parental lineage. That lineage is real. His responsibility begins with himself and with his own. He has no generic responsibility. He values human life broadly and in particular. Human life is the floor of goodness."`
   - **Issue:** The first sentence announces a *"Range of views"* on where self-interest ends and obligation to others begins, but the commit deleted the actual explanation of that range (household/kin continuity vs. individualist vs. sacrificial framing) and replaced it with a single doctrinal assertion about lineage and goodness.
   - **Suggested fix:** Restore a genuine statement of the range of views within the perimeter (e.g., how thinkers within the perimeter frame the extension of self-interest into spouse, children, and lineage, while rejecting both atomized solipsism and generic self-effacement), and place the lineage/floor-of-goodness sentences in Principle 3 / Principle 4 / Virtue 6 where they belong.

3. **Non-parental fallbacks ("Lacking children of his own...") placed inside Virtue 6 instead of Virtue 7 in `civ/ccp_derive.md:261–265`**
   - **Location:** `civ/ccp_derive.md:261–265`.
   - **Current text:** Right before `> **Virtue 6: Have children.**`, the text inserts: `"Lacking children of his own, he cares for his kin, broadly understood. Lacking those, he cares for the people who share his general heredity. Even lacking that, he cares for the people around him, for he is a social creature."`
   - **Issue:** What an unmarried or childless person does when *lacking children of his own* is the bridge to **Virtue 7 (Form and listen)** and **Principle 4**, not the derivation of **Virtue 6 (Have children)**. Furthermore, in line 261, changing `"0.8 children per couple"` to `"if they do not have children"` weakened the demographic replacement point (since a city also dies at 0.8 or 1.2 children per woman, not only at 0.0).
   - **Suggested fix:** Keep Virtue 6 focused on biological continuation and replacement (`2.1 children per woman`) and move the `"Lacking children of his own..."` progression to the transition between Virtue 6 and Virtue 7 (paralleling `10_ch10.md:90–100`).

---

### 1.5 Logical and Contextual Mismatches Introduced in Edits

1. **Missing "carrying a load" evidence in `civ/chapters/04_ch04.md:108`**
   - **Current text:**
     > `"* **In biological roles:** Assume men and women are interchangeable units with no innate differences in physical strength, reproductive biology, or maternal and paternal disposition. If that premise is true, then male and female recruits should be interchangeable wherever the task is carrying a load or running a measured distance, and mothers and fathers should be interchangeable in gestating and nursing infants. In the physical world, gestation and nursing occur in mothers, and the gap in upper-body and sprint power shows up everywhere the task is timed or weighed: the men's 100-meter world record is 9.58 seconds (Usain Bolt, 2009) and the women's is 10.49 seconds (Florence Griffith-Joyner, 1988), and the high-school boys' 100-meter record (10.00 seconds, Trentavis Friday, 2014) is faster than the women's world record. The biology does not change to match the slogan."`
   - **Issue:** Sentence 2 sets up two physical tests (`"carrying a load or running a measured distance"`), and Sentence 3 mentions `"everywhere the task is timed or weighed"`, but because the infantry rucksack clause was deleted, the evidence sentence only gives 100-meter sprint times and never mentions anything weighed or carried.
   - **Suggested fix:** Either add a verified weight/load measurement (such as world/Olympic weightlifting records or military load-carriage injury data with citation) or change `"carrying a load or running a measured distance"` / `"timed or weighed"` to match the sprint and strength measurements actually cited.

2. **Culver City High School honors English inserted into "The Daughter and the Chalked Footrace" (`civ/chapters/07_ch07.md:108–111`)**
   - **Current text:**
     > `"Now contrast that with the three ways modern institutions destroy excellence on the field:"`
     > `"1. **Canceling the Race and Awarding Identical Ribbons to All:** ..."`
     > `"2. **Handicapping the Swift (Eliminating Advanced Tracks):** In 2021, Culver City High School in California abolished its honors English classes in the name of equity..."`
     > `"3. **Withholding Honest Coaching as 'Harsh':** The parent or teacher refuses to teach the girl running mechanics..."`
   - **Issue:** In `09_ch09.md:231`, the Culver City High School example fits seamlessly. In `07_ch07.md:108–111`, however, the three numbered items are explicitly introduced as *"the three ways modern institutions destroy excellence on the field"* inside the Daughter and the Chalked Footrace illustration. Dropping the footrace handicap in Item 2 and jumping straight to Culver City High School English classes breaks the three-part footrace parallel.
   - **Suggested fix:** In `07_ch07.md:110`, keep the footrace parallel first (`"The school requires the fastest runners to start twenty yards behind the line, or cancels the varsity heat entirely because not every runner qualifies..."`) and then cite the Culver City High School honors-track abolition as the real-world educational equivalent in the second sentence of that bullet.

3. **Jewish citizens mentioned in *Varlık Vergisi* sentence and omitted two sentences later (`civ/chapters/08_ch08.md:198–200`)**
   - **Current text:**
     > `"In November 1942, the Turkish Grand National Assembly enacted the *Varlık Vergisi* (Wealth Tax, Law No. 4305), assessed by local boards without right of appeal and falling at confiscatory rates on Greek, Armenian, and Jewish citizens..."` (line 198)
     > `"Through the Twenty Classes conscription and the *Varlık Vergisi*, the state made itself the active enemy of its own Greek and Armenian residents, using taxation and administrative process to strip and break a distinct people within its borders."` (line 200)
   - **Issue:** Line 198 accurately notes that the *Varlık Vergisi* targeted `"Greek, Armenian, and Jewish citizens"` (and the Twenty Classes also conscripted non-Muslim minorities including Greeks, Armenians, Jews, and Assyrians), but line 200 drops Jewish citizens and says only `"its own Greek and Armenian residents"`.
   - **Suggested fix:** Change line 200 to `"its own Greek, Armenian, and Jewish citizens"` (or `"its own non-Muslim minority citizens"`).

4. **Grammatical awkwardness of using the imperative verb phrase "Form and listen" as a noun (`04_ch04.md:131`, `06_ch06.md:83`, `10_ch10.md:101`)**
   - **Locations:**
     - `civ/chapters/04_ch04.md:131`: `"It is gratitude to the ancestors who built it, and **Form and listen (Virtue 7)**: keeping the bridge repaired..."`
     - `civ/chapters/06_ch06.md:83`: `"Do the younger owe **Form and listen (Virtue 7)** to the elders who formed them..."`
     - `civ/chapters/10_ch10.md:101`: `"That is **Form and listen** on a national scale."`
     - Also `civ/chapters/10_ch10.md:105`: `"What are the grave violations of Virtue 7? Elders who will not form those younger than themselves, and the young who will not listen to the experience around them. And **generational theft**: when the currently living generation..."` (two sentence fragments in a row).
   - **Issue:** Because the canonical name of Virtue 7 was changed from a noun/clause to the compound imperative verb phrase **Form and listen**, mechanically substituting `"Form and listen"` into slots where a noun object stood (`"owe Form and listen to the elders"`, `"is gratitude... and Form and listen"`) creates ungrammatical sentences. Additionally, in `10_ch10.md:101`, George Washington raising Martha's children and relinquishing power exemplifies *forming and stewarding for those who come after*, not *listening to elders*.
   - **Suggested fix:**
     - `04_ch04.md:131`: `"...and the practice of **Virtue 7 (Form and listen)**: keeping the bridge repaired..."`
     - `06_ch06.md:83`: `"Does **Virtue 7 (Form and listen)** bind the younger to listen to and test the experience of the elders who formed them, and bind living adults to transmit their inheritance to the next generation?"`
     - `10_ch10.md:101`: `"That is the elder's side of **Virtue 7 (Form and listen)** practiced on a national scale: he formed those younger than himself and preserved the constitutional inheritance rather than making himself king."`
     - `10_ch10.md:105`: `"The grave violations of Virtue 7 are threefold: elders who refuse to form those younger than themselves, young people who refuse to listen to and test the experience around them, and **generational theft**, in which the living generation treats the accumulated moral, institutional, and financial capital of its ancestors as a private fund to be exhausted in a single lifetime..."`

---

### 1.6 Cross-Document Inconsistencies Left by Partial Edits

1. **`civ/ccp_eng.md` missed the updated definitions of Technology and Civilization (`civ/ccp_eng.md:32, 37`)**
   - In `ccp_sheet.md:7`, `ccp_derive.md:9–10`, and `01_ch01.md:54, 71`, the definitions were updated so that:
     - **Technology** is defined as `"a repeatable method, together with the structure that method keeps in being, by which human beings order the physical or social world to serve a human purpose"` (or `"practical, repeatable answers to 'how to do X' that can be taught, learned, embodied in tools and institutions..."`).
     - **Civilization** replaces `"passing human life"` with `"continuing human life from generation to generation"`.
   - However, `civ/ccp_eng.md:37` still reads:
     > `"**Civilization** is the process and structure of passing human life, along with the ability to construct and maintain material and social order, from one generation to the next."`
   - **Suggested fix:** Update `civ/ccp_eng.md:32–37` to align with `ccp_sheet.md:7` and `01_ch01.md:54, 71`.

2. **Conflicting 1932–1933 Soviet Famine Death Counts between Chapter 5 and Chapter 13**
   - `civ/chapters/05_ch05.md:82` states:
     > `"and during the terrible winter and spring of 1932–1933, between five and seven million human beings starved to death."`
   - `civ/chapters/13_ch13.md:111` (updated in commit `a410a0a8`) states:
     > `"A 2020 demographic reconstruction puts excess deaths for the Soviet Union in 1932–1933 at about 8.7 million, of which about 3.9 million were in Ukraine, about 3.3 million in Russia, and about 1.3 million in Kazakhstan (Wolowyna and others, *Journal of Genocide Research*, 2020)."`
   - **Suggested fix:** Harmonize `05_ch05.md:82` with `13_ch13.md:111` (e.g., `"between five and nearly nine million human beings starved to death—with a 2020 demographic reconstruction estimating approximately 8.7 million excess deaths across Ukraine, Russia, and Kazakhstan"`).

3. **Chapter 8 (`08_ch08.md:45–46`) missed the Tier 5 Honesty and Tier 4 Confiscatory-Taxation updates**
   - In `ccp_sheet.md:160–163`, `ccp_derive.md:174–177`, and `ccp_eng.md:352–355`:
     - `"honest dealing"` was removed from Tier 5 and followed by the explicit rule: `"Honesty in speech and dealing is not one of these tiers. It binds inside every tier."`
     - Tier 4 added the qualification that confiscatory or punitive taxation is an act of enmity rather than legitimate civic obligation.
   - In `civ/chapters/08_ch08.md:45–46`, however, Tier 5 still reads:
     > `"5. **Tier 5: Humanity at Large and Strangers:** Honest dealing, basic human decency, peaceful trade, and voluntary aid from genuine surplus."`
     And the paragraph following the tiers does not state that honesty binds across all tiers.
   - **Suggested fix:** Update `civ/chapters/08_ch08.md:45–48` to match `ccp_sheet.md:160–163`, `ccp_derive.md:174–177`, and `ccp_eng.md:352–355`.

---

## Part 2: Broader Re-Evaluation of the Essays and Book

### 2.1 Grammar, Syntax, and Editing-Guide Compliance Issues

1. **Grammatical typo in Chapter 13 (`civ/chapters/13_ch13.md:84`)**
   - **Current text:** `"assign it to an negligent neighbor"`
   - **Cause:** Commit `317aea9b` replaced `"an idle neighbor"` with `"an negligent neighbor"` without changing `"an"` to `"a"`.
   - **Suggested fix:** Change `"an negligent neighbor"` to `"a negligent neighbor"`.

2. **Remaining currency symbols (`$`) and percent signs (`%`) in prose (`06_editing_guide.md` Section 3.3)**
   - The deterministic audit script and symbol scan flagged the following instances where symbols are used in running prose instead of words:
     - `civ/chapters/11_ch11.md:54` and `222`: `"$80,000"` -> should be `"eighty thousand dollars"` in running prose (in line 222 it is inside a table, though writing it out or keeping it consistent is cleaner).
     - `civ/chapters/13_ch13.md:22`: `"approximately 75% Chinese speaking several mutually unintelligible dialects, 14% Malay Muslim, and 8% Indian Tamil"` -> write `"75 percent"`, `"14 percent"`, `"8 percent"` (matching the rest of the book, which spells out `percent`).
     - `civ/chapters/14_ch14.md:99–100`: `"has $400 a month remaining... place $300 a month... borrowing at 22% interest"` -> write `"four hundred dollars"`, `"three hundred dollars"`, `"22 percent"`.

3. **Conversational intensifiers ("100 percent", "reduced to zero") flagged by Part A audit (`06_editing_guide.md` Section 3.3)**
   - `civ/chapters/02_ch02.md:27` and `39`: `"How can I be 100 percent certain..."` / `"we cannot be 100 percent certain..."` -> replace `"100 percent certain"` with `"infallibly certain"` or `"absolutely certain"`.
   - `civ/chapters/06_ch06.md:71`: `"100 percent peaceful, unanimous consent"` -> replace with `"wholly peaceful, unanimous consent"`.
   - `civ/chapters/06_ch06.md:106`: `"suffering, pain, inequality, and risk of harm are permanently reduced to zero"` -> replace `"reduced to zero"` with `"permanently eliminated"`.

4. **Remaining rhetorical question transitions (`06_editing_guide.md` Section 3.1)**
   - While commit `317aea9b` removed most dramatic one-line rhetorical questions (`"What happened?"`), a few remain:
     - `civ/chapters/02_ch02.md:59`: `"What happened when the Soviet state forced farms to plant according to Lysenko's ideological "truth"?"` -> can be stated declaratively: `"When the Soviet state forced collective farms to plant according to Lysenko's ideological doctrine, physical reality did not yield to the Politburo's decree."`
     - `civ/chapters/13_ch13.md:88`: `"What occurred in practice over the next seven years (1907–1914)? **The policy succeeded.**"` -> `"In practice, over the next seven years (1907–1914), **the policy succeeded.**"`
     - `civ/chapters/13_ch13.md:140`: `"And what occurred in Iran on the morning *after* the wall came down in February 1979?"` -> `"The events in Iran immediately after the monarchy fell in February 1979 demonstrated the outcome of such an alliance."`

---

### 2.2 Structural and Logical Issues Across the Essays and Book

1. **Out-of-order numbering of Virtues during derivation in `civ/ccp_derive.md`**
   - In `civ/ccp_derive.md`, the Seven Virtues are derived in the order:
     1. `Virtue 1: Align with truth` (line 43, called `"our first personal rule—our first virtue"`)
     2. `Virtue 5: Be humble` (line 66, called `"our second operational virtue"`)
     3. `Virtue 3: Seek excellence` (line 126, called `"our third individual virtue"`)
     4. `Virtue 4: Admire beauty` (line 147, called `"our fourth individual virtue"`)
     5. `Virtue 2: Take ownership, which is responsibility` (line 163, called `"our fifth individual virtue"`)
     6. `Virtue 6: Have children` (line 266)
     7. `Virtue 7: Form and listen` (line 275)
   - **Issue:** Calling **Virtue 5** *"our second operational virtue"* and **Virtue 2** *"our fifth individual virtue"* creates a jarring clash between the ordinal words (*second, third, fourth, fifth*) and the canonical numbers (*Virtue 5, Virtue 3, Virtue 4, Virtue 2*). By contrast, `ccp_eng.md` (Part II, Chapters 7–13) presents the virtues in canonical order 1 through 7 after deriving the Five Principles, or notes their canonical number without conflicting ordinal labels.
   - **Suggested fix:** In `civ/ccp_derive.md:64, 124, 145, 161`, either remove the conflicting ordinal words (e.g., write `"From this limit, we derive the operational virtue of humility:"` -> `> **Virtue 5: Be humble.**`, and explain in one sentence at line 41 that the virtues are numbered 1–7 according to the standard summary table while emerging in the derivation as each principle is established) or derive Virtue 2 under Principle 3 before Virtues 3 and 4.

2. **Tension between Chapter 2's sensory-exclusivity sentence and the Revelation Allowance in Chapter 3 (`civ/chapters/02_ch02.md:25`)**
   - In `ccp_sheet.md:38`, `ccp_derive.md:32`, and `ccp_eng.md:72`, the text explicitly notes that some within the civilizational perimeter count religious revelation as a source of knowledge regarding what the senses do not reach, provided it never contradicts observable facts in the shared physical world.
   - In `civ/chapters/02_ch02.md:25`, however, commit `a410a0a8` deleted the sentence on neglect and left:
     > `"We have no access to the external world except through our physical senses and the instruments and reasoning we build upon them."`
   - **Issue:** Stated without scope, `"We have no access to the external world except through our physical senses..."` sounds like strict philosophical empiricism/materialism, which stands in tension with both the revelation allowance in the essays and the Two-Books discussion in `03_ch03.md:76–88`.
   - **Suggested fix:** Add a brief scoping phrase or cross-reference in `02_ch02.md:25` (parallel to `ccp_eng.md:72`): `"For all matters of the shared physical world, our access comes through our physical senses and the instruments and reasoning we build upon them (while those who also hold religious revelation as a source of knowledge regarding unseen things rightly recognize that revelation may never cancel an observable measurement in the shared world, as we shall examine in Chapter 3)."`

3. **Terminology drift on the twenty-year parenting progression in Chapter 14 (`civ/chapters/14_ch14.md:175`)**
   - In `civ/chapters/10_ch10.md:51–53`, the parenting framework is titled **Direct Authority Followed by Progressive Autonomy** (`Years 0 to 7: Direct Protection and Authority`, `Years 7 to 14: Guided Apprenticeship and Earned Responsibility`, `Years 14 to 20: Progressive Autonomy Toward Adult Ownership`).
   - In `civ/chapters/14_ch14.md:175`, the parenthetical reference still uses an older title: `(**Progressive Delegation of Responsibility**)`.
   - **Suggested fix:** Change `(**Progressive Delegation of Responsibility**)` in `14_ch14.md:175` to `(**Direct Authority Followed by Progressive Autonomy**, Chapter 10)` so the cross-reference matches the actual heading in Chapter 10.

4. **Terminology drift on the Granary / Storehouse illustration in Chapter 14 (`civ/chapters/14_ch14.md:105`)**
   - In `civ/chapters/14_ch14.md:105`, Scenario 2 refers to `"the **Municipal Granary Case**"`.
   - Earlier chapters have **The Example of the Seed Granary** (`10_ch10.md:109`) and **The Hill Town Grain Storehouse and the Rats** (`11_ch11.md:108`). The violent-student classroom scenario in `14_ch14.md:105` (where indulging a predator harms the weakest law-abiding students first) corresponds to **The Hill Town Grain Storehouse and the Rats** in Chapter 11.
   - **Suggested fix:** Update `"the **Municipal Granary Case**"` in `14_ch14.md:105` to `"**The Hill Town Grain Storehouse and the Rats** (Chapter 11)"`.

---

### 2.3 Factual and Historical Precision Checks

1. **Eastern Airlines Maintenance Indictment Date and Details (`civ/chapters/04_ch04.md:109`)**
   - **Fact check:** Eastern Air Lines and several maintenance managers at JFK/Miami were indicted by a federal grand jury in **July 1990** (not 1988; 1988 was the FAA ramp-check investigation, while the 60-count federal indictment for falsifying maintenance logs occurred in July 1990, and Eastern pleaded guilty in 1991).
   - **Suggested fix:** Change `"in 1988, before the federal indictment over falsified maintenance logs"` to `"leading to the 1990 federal indictment of Eastern Air Lines and its maintenance supervisors for falsifying safety and maintenance logs"`.

2. **John Locke's *Second Treatise* Publication Date (`civ/chapters/12_ch12.md:34`)**
   - **Fact check:** Line 34 lists `"John Locke (*Second Treatise of Government*, 1689)"`. While composition began earlier and the printing license was late 1689 (dated 1690 on the title page), 1689 is standard and acceptable in modern scholarship; no change required, though `"1689/1690"` is sometimes used.

3. **Trentavis Friday High-School 100m Record (`civ/chapters/04_ch04.md:108`)**
   - **Fact check:** Trentavis Friday ran 10.00 seconds in July 2014 (then the U.S. high-school boys' national record; subsequently lowered to 9.98 by Matthew Boling in 2019 and 9.93 by Christian Miller in 2024). Even at 10.00 seconds (or 9.93 seconds in 2024), the high-school boys' record is well below the women's world record of 10.49 seconds.
   - **Suggested fix:** Optionally update or qualify: `"the high-school boys' 100-meter record (10.00 seconds by Trentavis Friday in 2014, and 9.93 seconds by Christian Miller in 2024) is faster than the women's world record"`.

---

## Summary Checklist of Recommended Edits (When Ready to Apply)

| Priority | File & Lines | Issue Summary |
| :--- | :--- | :--- |
| **High** | `civ/chapters/13_ch13.md:111` | Remove leaked editorial meta-commentary (*"The count is in the millions. It is not a round ten million pasted over three regions."*) |
| **High** | `civ/chapters/11_ch11.md:143–144` | Remove leaked editorial meta-commentary (*"by a simpler route than a metaphor"*) and fix colon sentence fragment |
| **High** | `civ/chapters/04_ch04.md:66, 108–109` | Fix unanchored *"The studies"* and *"about-figures"*; replace slang (*"slop shop"*, *"pencil whipping"*) and fix 1990 Eastern Airlines indictment date; align *"carrying a load"* setup with evidence |
| **High** | `civ/ccp_eng.md:393, 399` & `610, 616` | Remove verbatim duplicate sentences created by recent edits |
| **High** | `civ/chapters/08_ch08.md:45–48, 169, 200, 204` | Remove duplicate enemy definition at line 204; include Jewish citizens in line 200 conclusion; update Tier 4 (confiscatory taxation) and Tier 5 (honesty across all tiers) at lines 45–48 |
| **High** | `civ/chapters/10_ch10.md:68–74, 101, 105` | Deduplicate *"Elders form those younger"* sentences; fix grammar of *"Form and listen"* used as a noun phrase (`04_ch04.md:131`, `06_ch06.md:83`, `10_ch10.md:101`); fix sentence fragments at line 105 |
| **Medium** | `civ/ccp_sheet.md:97`, `civ/ccp_derive.md:72, 261–265`, `civ/ccp_eng.md:170` | Fix out-of-sequence *"Human life is the floor of goodness"* inside Principle 2; restore actual range of views in `ccp_sheet.md:97`; move *"Lacking children of his own..."* from Virtue 6 to Virtue 7 in `ccp_derive.md:261–265` |
| **Medium** | `civ/ccp_eng.md:32–37` | Update definitions of **Technology** and **Civilization** (*"continuing human life"* instead of *"passing human life"*) to match `ccp_sheet.md`, `ccp_derive.md`, and `01_ch01.md` |
| **Medium** | `civ/chapters/05_ch05.md:82` | Harmonize 1932–1933 Soviet famine death count (`"between five and seven million"`) with `13_ch13.md:111` (`"about 8.7 million"`) |
| **Medium** | `civ/chapters/07_ch07.md:110` | Restore footrace handicap sentence before the Culver City High School example so the three-part footrace structure on the field remains intact |
| **Medium** | `civ/ccp_derive.md:64, 124, 145, 161` | Resolve ordinal clash where Virtue 5 is called *"our second operational virtue"* and Virtue 2 is called *"our fifth individual virtue"* |
| **Low** | `civ/chapters/13_ch13.md:84` | Fix typo `"an negligent neighbor"` -> `"a negligent neighbor"` |
| **Low** | `civ/chapters/02_ch02.md:25, 27, 39, 59`, `06_ch06.md:71, 106`, `11_ch11.md:54, 222`, `13_ch13.md:22, 88, 140`, `14_ch14.md:99–100, 105, 175` | Scope sensory-access claim in `02_ch02.md:25`; replace `$`/`%` symbols and `"100 percent"`/`"reduced to zero"` intensifiers; align cross-reference titles in `14_ch14.md:105, 175` |

# Book1-3 Write Operations

This is a looping operation guide for writing book1-3 chapters. Follow this process iteratively for each chapter and section.

---

## File Naming Pattern

**Format:** `[PART][SECTION]_ch[SECTION].md`
- `PART`: Single-digit part number (1, 2, 3, 4, 5)
- `SECTION`: Twp-digit section number within part (01, 02, 03...)
- `ch[SECTION]`: Chapter number (matches section number)

**Examples:**
- `001_title.md` - Title page (Part 0? Special case)
- `002_introduction.md` - Introduction (Part 0? Special case)
- `101_ch01.md` - Chapter 1 in Part 1
- `102_ch02.md` - Chapter 2 in Part 1
- `201_ch01.md` - Chapter 1 in Part 2

---

## Template Variables Available

From YAML frontmatter in title file:
- `{{.title}}` - Full book title
- `{{.titleShort}}` - Short title
- `{{.subtitle}}` - Subtitle
- `{{.author}}` - Author name
- `{{.date}}` - Date
- Other variables can be added to title file and will be available

---

## Writing Process

### 1. Start with Title File (if needed)
- Check if title file exists (may use `000_title.md`)
- Include YAML frontmatter with book metadata

### 2. For Each Chapter:

**Step 1: Check Outline Reference**
- Look up corresponding section in `book1-3/outline/outline_L3.md`
- Identify all L1, L2, L3 items for this section
- Note all `[idea:ID]`, `[ref:SOURCE]`, `[tie:ID|ID]`, `[example:ID]`, `[quote:Q###]` links

**Step 2: Write Chapter Content**
- Start with chapter header: `## Chapter X: [Title]`
- Write content for first L1 item
- Include proper markdown formatting (headers, lists, blockquotes, etc.)
- Integrate relevant index IDs using `[idea:ID]` syntax
- Add references using `[ref:SOURCE]` syntax
- Include relevant examples using `[example:ID]` syntax
- Connect ideas using `[tie:ID|ID]` syntax where relevant

**Step 3: Edit and Evaluate**
- Go back and edit the chapter file
- Evaluate how it reads
- Check: Is the tone consistent with the rest of the book?
- Check: Are all points from the outline covered?
- Check: Is the flow logical?
- Make adjustments as needed

**Step 4: Verify Index Coverage**
- Check `book1-3/index_book1-1.md` for relevant index entries
- Look up referenced IDs to ensure full context is included
- Add any nuanced examples or explanations from book1-1
- Ensure all key concepts from book1-1 are properly fleshed out

**Step 5: Check Links and Transitions**
- Ensure all `[idea:ID]` links point to valid IDs in the index
- Ensure all `[ref:SOURCE]` references exist in the index
- Verify that the chapter flows naturally into the next chapter
- Check transition to next section in outline

**Step 6: Move to Next Section/Chapter**
- Repeat steps 1-5 for next L1/L2 item
- Or move to next chapter

---

## Section Structure Guidelines

### L1 Item (Primary Idea)
- Should be substantial: 500-1000 words minimum
- Include comprehensive explanation
- Should stand on its own
- Often 3-5 paragraphs
- Include at least one relevant example or reference

### L2 Item (Secondary Idea)
- Should be detailed: 300-600 words minimum
- Expand on the L1 item with more specific details
- Usually 2-4 paragraphs
- Include examples and specific applications

### L3 Item (Tertiary Detail)
- Should be concise: 100-300 words
- Specific detail, example, or clarification
- Often 1-2 paragraphs
- Connect back to L1/L2

### Transitions Between Items
- Smooth flow from one item to the next
- Use connecting phrases
- Refer back to previous concepts where relevant
- Set up next concept naturally

---

## Writing Style Guidelines

1. **Be direct and concise** - Avoid unnecessary fluff
2. **Use concrete examples** - Make abstract concepts tangible
3. **Maintain consistent voice** - Match the tone from book1-1
4. **Reference book1-1 content** - Use actual content and examples where applicable
5. **Include index links** - Use `[idea:ID]` syntax to connect to index
6. **Use proper markdown** - Headers, lists, blockquotes, emphasis
7. **Avoid comments** - No `<!-- TODO -->` or similar in final content
8. **Check for duplicates** - Don't repeat yourself unnecessarily

---

## Evaluation Checklist

After writing each chapter, ask:

### Content Quality
- [ ] Is the core message clear?
- [ ] Are all points from the outline covered?
- [ ] Is the content organized logically?
- [ ] Are there sufficient examples?

### Writing Style
- [ ] Is the tone consistent with book1-1?
- [ ] Is the length appropriate for the content?
- [ ] Is it engaging and readable?
- [ ] Are transitions smooth?

### Technical
- [ ] Are all `[idea:ID]` links valid?
- [ ] Are all `[ref:SOURCE]` references accurate?
- [ ] Is markdown formatting correct?
- [ ] Are there any typos or errors?

### Index Integration
- [ ] Did I look up relevant book1-1 content?
- [ ] Are all nuanced points from book1-1 included?
- [ ] Are all examples properly cited?

### Flow
- [ ] Does the chapter flow logically?
- [ ] Do transitions to next section work?
- [ ] Does it set up the next chapter well?

---

## Part-Level Evaluation

After completing all chapters in a part:

### Part Structure
- [ ] Do all chapters flow together coherently?
- [ ] Is the part's main theme clear?
- [ ] Are all key concepts from the outline covered?

### Transitions
- [ ] Does the part transition from introduction work?
- [ ] Does the part transition to next part work?
- [ ] Are chapter transitions smooth?

### Completeness
- [ ] Compare with outline - is everything covered?
- [ ] Are there any gaps or missing sections?
- [ ] Do all key examples from book1-1 included?

### Index Coverage
- [ ] Check that all book1-1 concepts relevant to this part are referenced
- [ ] Verify that all historical examples are properly cited
- [ ] Ensure that all key figures are mentioned

---

## Clear Context and Reload

**Before starting new part/chapter:**
1. Read the previous chapter to maintain continuity
2. Review the outline section for upcoming content
3. Look up any referenced index entries for context
4. Note any transition requirements

**After finishing part/chapter:**
1. Note any adjustments made during writing
2. Document any insights or new connections discovered
3. Note any index entries that weren't used but should be
4. Note any examples that should be added but weren't

**Before asking for next step:**
1. Clear your current understanding of what was written
2. Be ready to summarize: what was covered, what was emphasized, what was deferred
3. Be prepared to discuss: tone issues, content gaps, structure problems

---

## Common Patterns from Book1-1

### Paragraph Structure
- **Topic sentence** - Clear statement of main idea
- **Elaboration** - 1-2 sentences explaining
- **Example** - Concrete illustration
- **Application** - How it applies practically
- **Connection** - Tie to previous/next concept

### Section Structure
- **Opening** - Hook or connection to previous
- **Main content** - 2-4 paragraphs per L1 item
- **Examples** - Concrete illustrations
- **Conclusion** - Summary and transition to next

### Reference Integration
- **Direct quotes** - Use blockquotes: `> "Quote text"`
- **Historical examples** - Describe the event, draw lessons
- **Figures** - Name them and describe their contribution
- **Bible verses** - Cite book, chapter:verse, provide context

---

## After All Chapters Written

### Final Review
1. Read entire book from start to finish
2. Check for consistency of voice and tone throughout
3. Verify all outline points are covered
4. Check for smooth transitions between all chapters
5. Verify all key concepts from book1-1 are integrated

### Final Adjustments
1. Make any final content adjustments
2. Fix any inconsistencies found
3. Add any connecting passages needed for flow
4. Ensure all transitions are smooth

### Validation
1. Run `go run concat.go` to build book.md
2. Check for any formatting or parsing errors
3. Verify output looks good
4. Note any final issues for resolution

---

## Notes

- **This is a looping operation** - We will iterate: write → edit → evaluate → next
- **Context is crucial** - Always understand what came before and what comes next
- **Index is your friend** - Use `book1-3/index_book1-1.md` liberally for examples, references, and nuanced content
- **Outline is your guide** - `book1-3/outline/outline_L3.md` shows what must be covered in each section
- **User feedback is key** - After each step, summarize progress and ask for direction

---

## Starting Point

**Current Status:**
- All parts complete: 30 chapters written (~40,000 words)
- Book writing complete

**Completed:**
- Introduction (2 files)
- Part 1: Foundation (6 chapters)
- Part 2: Individual (7 chapters)
- Part 3: Society (6 chapters)
- Part 4: Deeper Grounding (5 chapters)
- Part 5: Distinctions (4 chapters + 505 as conclusion)

**Final Step:**
- Run `go run concat.go` to build book.md
- Final review of assembled book

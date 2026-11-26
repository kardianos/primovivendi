package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComposeEndToEndWorkflow(t *testing.T) {
	// Set up an isolated synthetic workspace with civ/chapters/01_ch01.md.
	root := t.TempDir()
	chapDir := filepath.Join(root, "civ", "chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	origMD := "# Chapter One\n\nWe come into this world naked. The bridge stands firm; the valley endures. Good builders keep clear order.\n"
	chapPath := filepath.Join(chapDir, "01_ch01.md")
	if err := os.WriteFile(chapPath, []byte(origMD), 0o644); err != nil {
		t.Fatal(err)
	}

	stateDir := filepath.Join(root, "civ", "notes", "compose")
	eng, err := NewEngine(root, stateDir, false)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	sess, err := NewSession(eng, "ch01")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	// 1. Starts on Pass 0 (Backward) at the landing sentence (S3, index 2).
	if sess.Dir != PassBackward || sess.SentIdx != 2 {
		t.Fatalf("expected Backward pass at landing sentence index 2, got dir=%s idx=%d", sess.Dir, sess.SentIdx)
	}
	if len(sess.Goals) != 3 || !sess.Goals[2].WantLandingPunch {
		t.Fatalf("expected 3 backward goals with landing punch on S3: %+v", sess.Goals)
	}

	// 2. Lock a global set phrase and verify it becomes a locked slot.
	sess.LockGlobalPhrase("clear order")
	foundLock := false
	for _, sl := range sess.Plan.Slots {
		if sl.Locked && strings.EqualFold(sl.Text, "clear order") {
			foundLock = true
		}
	}
	if !foundLock {
		t.Fatalf("expected 'clear order' to be locked in slots: %+v", sess.Plan.Slots)
	}

	// 3. Pin landing anchor -> transitions backward to S2 (index 1).
	sess.Accept()
	if sess.SentIdx != 1 {
		t.Fatalf("expected backward step to S2 (idx 1), got %d", sess.SentIdx)
	}

	// 4. Hover a structural split option -> live preview updates!
	var splitID string
	for _, sc := range sess.Plan.Structures {
		if sc.Kind == "split_mark" {
			splitID = sc.ID
		}
	}
	if splitID == "" {
		t.Fatalf("expected split_mark structure option on S2, got %+v", sess.Plan.Structures)
	}
	sess.HoverStructure(splitID)
	if !strings.Contains(sess.PreviewText(), ". ") {
		t.Fatalf("expected hover preview to show split sentence, got %q", sess.PreviewText())
	}
	sess.ClearHover()

	// Select split structure, then accept S2.
	sess.ToggleStructure(splitID)
	if !strings.Contains(sess.Plan.ActiveText, ". ") {
		t.Fatalf("expected active text after ToggleStructure to be split, got %q", sess.Plan.ActiveText)
	}
	sess.Accept() // moves to S1 (idx 0)

	// 5. On S1, apply a manual rewrite as the new base and accept it.
	sess.ApplyManualRewrite("Every child enters this world bare and unguarded.")
	if sess.Plan.BaseText != "Every child enters this world bare and unguarded." {
		t.Fatalf("expected manual rewrite to become new BaseText, got %q", sess.Plan.BaseText)
	}
	sess.Accept() // finishes backward pass -> auto-switches to Forward Pass!
	if sess.Dir != PassForward {
		t.Fatalf("expected automatic transition to Forward Pass, got %s", sess.Dir)
	}

	// 6. Trigger Looping Seam Pass.
	sess.SetPass(PassLoop)
	if sess.Dir != PassLoop {
		t.Fatalf("expected PassLoop, got %s", sess.Dir)
	}

	// 7. Train local pure-Go inkline-style NN ranker from logged decision tree.
	stats, err := eng.TrainFromDecisions()
	if err != nil {
		t.Fatalf("TrainFromDecisions: %v", err)
	}
	if stats.Pairs == 0 {
		t.Fatalf("expected >0 training pairs extracted from decision tree")
	}
	if _, err := os.Stat(eng.ModelPath()); err != nil {
		t.Fatalf("expected ranker.safetensors at %s: %v", eng.ModelPath(), err)
	}

	// 8. Deterministic rebuild writes rebuilt markdown without mutating civ/chapters/01_ch01.md.
	sess.Rebuild()
	outPath := filepath.Join(stateDir, "rebuilt", "01_ch01.md")
	rebuiltBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read rebuilt chapter: %v", err)
	}
	if !strings.Contains(string(rebuiltBytes), "Every child enters this world bare and unguarded.") {
		t.Fatalf("rebuilt markdown missing manual rewrite:\n%s", string(rebuiltBytes))
	}
	origAfter, _ := os.ReadFile(chapPath)
	if string(origAfter) != origMD {
		t.Fatalf("source chapter was mutated unexpectedly")
	}
}

func TestMultiWordPhraseSelectionAndReconsideration(t *testing.T) {
	root := t.TempDir()
	chapDir := filepath.Join(root, "civ", "chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "# Chapter One\n\nKnowledge passes from generation to generation through ordinary lives.\n"
	if err := os.WriteFile(filepath.Join(chapDir, "01_ch01.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngine(root, filepath.Join(root, "civ", "notes", "compose"), false)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	sess, err := NewSession(eng, "ch01")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	// Find first and second "generation" slots.
	genSlots := []int{}
	for i, sl := range sess.Plan.Slots {
		if strings.EqualFold(sl.Text, "generation") {
			genSlots = append(genSlots, i)
		}
	}
	if len(genSlots) != 2 {
		t.Fatalf("expected two 'generation' slots, got %v in %+v", genSlots, sess.Plan.Slots)
	}

	// Click first "generation", then second "generation" -> selects "generation to generation".
	sess.SelectSlotHeader(genSlots[0])
	sess.SelectSlotHeader(genSlots[1])
	if got := sess.SelectedPhraseText(); got != "generation to generation" {
		t.Fatalf("SelectedPhraseText = %q, want %q", got, "generation to generation")
	}
	if len(sess.PhraseOptions) < 4 {
		t.Fatalf("expected >=4 phrase alternatives for 'generation to generation', got %+v", sess.PhraseOptions)
	}

	// Hover and apply an alternative (e.g. "age to age" or "father to son").
	var targetOpt PhraseOption
	for _, opt := range sess.PhraseOptions {
		if strings.EqualFold(opt.Phrase, "age to age") || strings.EqualFold(opt.Phrase, "father to son") {
			targetOpt = opt
			break
		}
	}
	if targetOpt.ID == "" {
		t.Fatalf("expected 'age to age' or 'father to son' in phrase options: %+v", sess.PhraseOptions)
	}
	sess.HoverPhrase(targetOpt.ID)
	if !strings.Contains(sess.PreviewText(), targetOpt.Phrase) {
		t.Fatalf("expected hover preview to contain %q, got %q", targetOpt.Phrase, sess.PreviewText())
	}
	sess.ApplyPhraseOption(targetOpt.ID)
	if !strings.Contains(sess.Plan.ActiveText, targetOpt.Phrase) {
		t.Fatalf("expected active text after ApplyPhraseOption to contain %q, got %q", targetOpt.Phrase, sess.Plan.ActiveText)
	}
	if sess.HasSelection() {
		t.Fatalf("expected selection to clear after applying phrase option")
	}
}

func TestPreventAdjacentWordAndBoundaryRepeats(t *testing.T) {
	root := t.TempDir()
	chapDir := filepath.Join(root, "civ", "chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "# Chapter One\n\nCarrying living life from generation to generation.\n"
	if err := os.WriteFile(filepath.Join(chapDir, "01_ch01.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngine(root, filepath.Join(root, "civ", "notes", "compose"), false)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	sess, err := NewSession(eng, "ch01")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	// Check every single-word swap option across all slots: none may repeat adjacent words.
	for _, sl := range sess.Plan.Slots {
		for _, opt := range sl.Options {
			if strings.Contains(strings.ToLower(opt.CompletedText), "living living") {
				t.Fatalf("word option created adjacent repeat: %q", opt.CompletedText)
			}
		}
	}

	// Check selecting "generation to generation" after "from": no option may produce "from from".
	sess.SelectSlotHeader(4)
	sess.SelectSlotHeader(6)
	for _, opt := range sess.PhraseOptions {
		low := strings.ToLower(opt.CompletedText)
		if strings.Contains(low, "from from") || strings.Contains(low, "living living") {
			t.Fatalf("phrase option created repeated word: %q", opt.CompletedText)
		}
	}

	// Also verify selecting "Carrying living life" never emits "living living".
	sess.ClearSelection()
	sess.SelectSlotHeader(0)
	sess.SelectSlotHeader(2)
	for _, opt := range sess.PhraseOptions {
		if strings.Contains(strings.ToLower(opt.CompletedText), "living living") {
			t.Fatalf("compositional phrase option created repeated word: %q", opt.CompletedText)
		}
	}
}

func TestCustomTwoWordPhraseBeforeLockedMultiWordPhrase(t *testing.T) {
	root := t.TempDir()
	chapDir := filepath.Join(root, "civ", "chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "# Chapter One\n\nWe hold clear order today.\n"
	if err := os.WriteFile(filepath.Join(chapDir, "01_ch01.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngine(root, filepath.Join(root, "civ", "notes", "compose"), false)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	sess, err := NewSession(eng, "ch01")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	// Lock "clear order" (slots 2..3 -> tokens 3..4) as a local multi-word locked phrase.
	sess.SelectSlotHeader(2)
	sess.SelectSlotHeader(3)
	if got := sess.SelectedPhraseText(); got != "clear order" {
		t.Fatalf("SelectedPhraseText = %q, want %q", got, "clear order")
	}
	sess.LockSelectedSpan()

	// Select single word "hold" (slot 1) immediately before the locked phrase,
	// and replace it with a two-word custom phrase "carry forward".
	sess.SelectSlotHeader(1)
	if got := sess.SelectedPhraseText(); got != "hold" {
		t.Fatalf("SelectedPhraseText = %q, want %q", got, "hold")
	}
	sess.ApplyCustomPhrase("carry forward")

	if got := sess.Plan.ActiveText; got != "We carry forward clear order today." {
		t.Fatalf("ActiveText = %q, want %q", got, "We carry forward clear order today.")
	}
	// Verify "carry" and "forward" are unlocked, and ONLY "clear order" is locked.
	var lockedTexts []string
	var unlockedWords []string
	for _, sl := range sess.Plan.Slots {
		if sl.Locked {
			lockedTexts = append(lockedTexts, sl.Text)
		} else {
			unlockedWords = append(unlockedWords, sl.Text)
		}
	}
	if len(lockedTexts) != 1 || lockedTexts[0] != "clear order" {
		t.Fatalf("expected single locked phrase 'clear order', got %v (slots=%+v)", lockedTexts, sess.Plan.Slots)
	}
	joinedUnlocked := strings.Join(unlockedWords, " ")
	if !strings.Contains(joinedUnlocked, "carry forward") {
		t.Fatalf("expected 'carry forward' to remain unlocked, got unlocked=%q", joinedUnlocked)
	}
}

func TestStableLivingSenseToggleAndWeightChooser(t *testing.T) {
	root := t.TempDir()
	chapDir := filepath.Join(root, "civ", "chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "# Chapter One\n\nFamilies build a stable living across hard seasons.\n"
	if err := os.WriteFile(filepath.Join(chapDir, "01_ch01.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngine(root, filepath.Join(root, "civ", "notes", "compose"), false)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	sess, err := NewSession(eng, "ch01")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	// Select "stable living" (slots 3..4).
	var lo, hi int = -1, -1
	for i, sl := range sess.Plan.Slots {
		if strings.EqualFold(sl.Text, "stable") {
			lo = i
		}
		if strings.EqualFold(sl.Text, "living") {
			hi = i
		}
	}
	if lo < 0 || hi < 0 {
		t.Fatalf("could not find 'stable living' slots in %+v", sess.Plan.Slots)
	}
	sess.SelectSlotHeader(lo)
	sess.SelectSlotHeader(hi)

	// Toggle off the animation sense, de-weight "firm", boost "steady", and enable 2-hop expand.
	for _, sn := range sess.ActiveSenses {
		if strings.Contains(strings.ToLower(sn.Gloss), "animated") || strings.Contains(strings.ToLower(strings.Join(sn.Synonyms, " ")), "animation") {
			sess.ToggleSense(sn.Offset)
		}
	}
	sess.AddLocalDeweight("firm")
	sess.AddGlobalDeweight("animation")
	sess.AddLocalBoost("steady")
	sess.ToggleExpand2Hop()

	if len(sess.PhraseOptions) == 0 {
		t.Fatal("expected phrase options after steering 'stable living'")
	}
	topHasSteady := false
	for i, opt := range sess.PhraseOptions {
		low := strings.ToLower(opt.Phrase)
		if strings.Contains(low, "animation") || strings.Contains(low, "firm") {
			t.Fatalf("de-weighted term appeared in options: %+v", opt)
		}
		if i < 4 && (strings.Contains(low, "steady") || strings.Contains(low, "settled")) {
			topHasSteady = true
		}
	}
	if !topHasSteady {
		t.Fatalf("expected steady/settled phrase in top 4 after boosting 'steady', got %+v", sess.PhraseOptions[:min(5, len(sess.PhraseOptions))])
	}
}

func TestAcceptParagraphSealsBlockAndPersists(t *testing.T) {
	root := t.TempDir()
	chapDir := filepath.Join(root, "civ", "chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "# Chapter One\n\nWe build a steady bridge. The bridge stands firm.\n\nSecond paragraph opens here.\n"
	if err := os.WriteFile(filepath.Join(chapDir, "01_ch01.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(root, "civ", "notes", "compose")
	eng, err := NewEngine(root, stateDir, false)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	sess, err := NewSession(eng, "ch01")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	p1ID := sess.ActiveBlock().ID
	if sess.IsParagraphDone(p1ID) {
		t.Fatal("expected P1 to start open")
	}

	// Accept entire P1 -> settles all sentences, logs accept_paragraph, advances to P2!
	sess.AcceptParagraph()
	if !sess.IsParagraphDone(p1ID) {
		t.Fatal("expected P1 to be marked completed after AcceptParagraph")
	}
	if sess.ParaPos != 1 {
		t.Fatalf("expected auto-advance to ParaPos 1, got %d", sess.ParaPos)
	}

	// Reload from disk and verify paragraph completion persisted!
	sess2, err := NewSession(eng, "ch01")
	if err != nil {
		t.Fatalf("reload NewSession: %v", err)
	}
	if !sess2.IsParagraphDone(p1ID) {
		t.Fatal("expected P1 completed state to persist across reload from JSONL")
	}
}

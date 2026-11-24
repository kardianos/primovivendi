package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func plain(s string) string {
	return ansiSeq.ReplaceAllString(s, "")
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+j":
		return tea.KeyMsg{Type: tea.KeyCtrlJ}
	case "ctrl+k":
		return tea.KeyMsg{Type: tea.KeyCtrlK}
	case "ctrl+r":
		return tea.KeyMsg{Type: tea.KeyCtrlR}
	case "ctrl+s":
		return tea.KeyMsg{Type: tea.KeyCtrlS}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func (m tuiModel) press(keys ...string) tuiModel {
	var model tea.Model = m
	for _, k := range keys {
		next, _ := model.Update(keyMsg(k))
		model = next
	}
	return model.(tuiModel)
}

func queueSession(t *testing.T) (*Session, string) {
	t.Helper()
	root := t.TempDir()
	body := strings.Join([]string{
		"Alpha sentence stays.",
		"second line.",
		"third line.",
		"fourth line.",
		"fifth line.",
		"sixth line.",
		"seventh line keeps what was built.",
		"Beta sentence stays.",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261122-150000-queue.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{
			{
				"id": "q-alpha", "from": "grok-4.7", "path": "doc.md",
				"quote": "Alpha sentence stays.", "problem": "Alpha is still open.",
				"alternates": []map[string]any{
					{"id": "a", "label": "alpha-alt", "text": "Alpha rewritten."},
					{"id": "no", "from": "cabon", "ignore": true, "reason": "this was actually good"},
				},
			},
			{
				"id": "q-beta", "from": "opus-5.5", "path": "doc.md",
				"quote": "Beta sentence stays.", "problem": "Beta is still open.",
				"alternates": []map[string]string{{"id": "a", "label": "beta-alt", "text": "Beta rewritten."}},
			},
		},
	})
	dec := filepath.Join(root, "decisions.json")
	sess, err := LoadSession(root, packets, dec)
	if err != nil {
		t.Fatal(err)
	}
	return sess, dec
}

func TestTUIRejectClearsEmptyQueueAndKeepsAll(t *testing.T) {
	sess, dec := queueSession(t)
	m := newTUI(sess, "Review", 0).press()
	m.width, m.height = 120, 40
	m.resize()
	view := plain(m.View())
	if !strings.Contains(view, "Alpha is still open.") {
		t.Fatalf("expected the first open item\n%s", view)
	}
	m = m.press("]")
	view = plain(m.View())
	if !strings.Contains(view, "seventh line keeps what was built") {
		t.Fatalf("] did not open more of the file\n%s", view)
	}
	m = m.press("r")
	view = plain(m.View())
	if !strings.Contains(view, "Beta is still open.") || strings.Contains(view, "Alpha is still open.") {
		t.Fatalf("expected the remaining open item\n%s", view)
	}
	m = m.press("r")
	view = plain(m.View())
	for _, want := range []string{"No items", "No item selected"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q after the queue emptied\n%s", want, view)
		}
	}
	for _, gone := range []string{"Alpha is still open.", "Beta is still open.", "u Use my wording"} {
		if strings.Contains(view, gone) {
			t.Fatalf("finished queue still shows %q\n%s", gone, view)
		}
	}

	m = m.press("3", "j", "r")
	view = plain(m.View())
	if strings.Contains(view, "No item selected") || !strings.Contains(view, "Beta is still open.") {
		t.Fatalf("the last item in All should stay selected\n%s", view)
	}
	b, err := os.ReadFile(dec)
	if err != nil {
		t.Fatal(err)
	}
	var got DecisionFile
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Items["q-alpha"].Disposition != "reject" || got.Items["q-beta"].Disposition != "reject" {
		t.Fatalf("decisions: %+v", got.Items)
	}
}

func TestTUIReviseCompareAndIgnore(t *testing.T) {
	sess, dec := queueSession(t)
	m := newTUI(sess, "Review", 0)
	m = m.press("c")
	view := plain(m.View())
	for _, want := range []string{"Current", "alpha-alt", "from: grok-4.7", "from: cabon"} {
		if !strings.Contains(view, want) {
			t.Fatalf("compare missing %q\n%s", want, view)
		}
	}
	m = m.press("c", "enter", "Z", "esc", "u")
	view = plain(m.View())
	if !strings.Contains(view, "Beta is still open.") {
		t.Fatalf("revise should advance to the next open item\n%s", view)
	}
	b, err := os.ReadFile(dec)
	if err != nil {
		t.Fatal(err)
	}
	var got DecisionFile
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Items["q-alpha"].Disposition != "revise" || got.Items["q-alpha"].Text != "Z" {
		t.Fatalf("revise was not saved from the field: %+v", got.Items["q-alpha"])
	}

	m = m.press("r")
	if plain(m.View()) != "" && !strings.Contains(plain(m.View()), "No item selected") {
		t.Fatalf("queue should be clear\n%s", plain(m.View()))
	}

	// A fresh ignore vote records reject and clears the only open item.
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte("Gamma sentence stays.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261122-160000-ignore.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "q-gamma", "from": "grok-4.7", "path": "doc.md",
			"quote": "Gamma sentence stays.", "problem": "Gamma is still open.",
			"alternates": []map[string]any{
				{"id": "no", "from": "cabon", "ignore": true, "reason": "this was actually good"},
			},
		}},
	})
	dec = filepath.Join(root, "decisions.json")
	sess, err = LoadSession(root, packets, dec)
	if err != nil {
		t.Fatal(err)
	}
	m = newTUI(sess, "Review", 0).press("tab", ".")
	if !strings.Contains(plain(m.View()), "a Ignore this") {
		t.Fatalf("ignore alternate should rename accept\n%s", plain(m.View()))
	}
	m = m.press("a")
	if !strings.Contains(plain(m.View()), "No item selected") {
		t.Fatalf("ignore this should clear the only open item\n%s", plain(m.View()))
	}
	b, err = os.ReadFile(dec)
	if err != nil {
		t.Fatal(err)
	}
	got = DecisionFile{}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	d := got.Items["q-gamma"]
	if d.Disposition != "reject" || d.Note != "this was actually good" || d.AcceptedText != "" {
		t.Fatalf("ignore vote: %+v", d)
	}
}

func TestOpenModelClearsWithoutAWindow(t *testing.T) {
	sess, _ := queueSession(t)
	v := newReviewView(sess)
	if v.selected != "q-alpha" {
		t.Fatalf("selected %q", v.selected)
	}
	v.reject()
	if v.selected != "q-beta" {
		t.Fatalf("advanced to %q", v.selected)
	}
	v.reject()
	if v.selected != "" || v.mine != "" || v.note != "" {
		t.Fatalf("empty open queue left %+v mine %q", v.selected, v.mine)
	}
	v.setFilter(filterAll)
	if v.selected == "" {
		t.Fatal("All should select the first item")
	}
	v.next()
	kept := v.selected
	v.reject()
	if v.selected != kept {
		t.Fatalf("All last item changed from %q to %q", kept, v.selected)
	}
}

func leftClick(m tuiModel, x, y int) tuiModel {
	next, _ := m.Update(tea.MouseMsg{
		X: x, Y: y,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	return next.(tuiModel)
}

func findAct(t *testing.T, m tuiModel, act string) tuiZone {
	t.Helper()
	for _, z := range m.screen().zones {
		if z.kind == zoneAct && z.act == act {
			return z
		}
	}
	t.Fatalf("no %s chip", act)
	return tuiZone{}
}

func findFilterZone(t *testing.T, m tuiModel, f int) tuiZone {
	t.Helper()
	for _, z := range m.screen().zones {
		if z.kind == zoneFilter && z.filter == f {
			return z
		}
	}
	t.Fatalf("no filter %d", f)
	return tuiZone{}
}

func findFieldZone(t *testing.T, m tuiModel, act string) tuiZone {
	t.Helper()
	for _, z := range m.screen().zones {
		if z.kind == zoneField && z.act == act {
			return z
		}
	}
	t.Fatalf("no %s field", act)
	return tuiZone{}
}

func findItemZone(t *testing.T, m tuiModel, id string) tuiZone {
	t.Helper()
	for _, z := range m.screen().zones {
		if z.kind == zoneItem && z.itemID == id {
			return z
		}
	}
	t.Fatalf("no item %s", id)
	return tuiZone{}
}

func decisionOf(t *testing.T, path, id string) (Decision, bool) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Decision{}, false
		}
		t.Fatal(err)
	}
	var got DecisionFile
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	d, ok := got.Items[id]
	return d, ok
}

func TestTUIMouseClicksAndFrame(t *testing.T) {
	sess, dec := queueSession(t)
	m := newTUI(sess, "Review", 0)
	m.width, m.height = 120, 40
	m.resize()

	listW, detailW := m.columns()
	if listW+frameGutter+detailW != 120 {
		t.Fatalf("panes %d + gutter + %d != 120", listW, detailW)
	}
	view := m.screen().view
	lines := strings.Split(view, "\n")
	bodyH := m.height - 2
	if len(lines) != bodyH+2 {
		t.Fatalf("screen has %d lines, want %d", len(lines), bodyH+2)
	}
	for i, line := range lines[:bodyH] {
		if lipgloss.Width(line) != 120 {
			t.Fatalf("body line %d width %d %q", i, lipgloss.Width(line), plain(line))
		}
	}
	if strings.TrimSpace(plain(lines[0])) != "" {
		t.Fatalf("top inset is %q", plain(lines[0]))
	}
	if !strings.HasPrefix(plain(lines[1]), " Review") {
		t.Fatalf("second line is %q", plain(lines[1]))
	}
	if plain(lines[1])[listW] != ' ' {
		t.Fatalf("gutter at %d is %q", listW, plain(lines[1])[listW:listW+1])
	}

	m = m.press(".")
	accept := findAct(t, m, actAccept)
	revise := findAct(t, m, actRevise)
	if revise.y != accept.y || revise.x != accept.x+accept.w+1 {
		t.Fatalf("accept/revise gap: accept %d+%d revise %d,%d", accept.x, accept.w, revise.x, revise.y)
	}
	if _, ok := m.screen().hit(accept.x+accept.w, accept.y); ok {
		t.Fatal("the cell between chips is inside a zone")
	}
	gap := leftClick(m, accept.x+accept.w, accept.y)
	if gap.review.selected != "q-alpha" {
		t.Fatalf("gap click selected %q", gap.review.selected)
	}
	if _, ok := decisionOf(t, dec, "q-alpha"); ok {
		t.Fatal("gap click wrote a decision")
	}

	reaffirm := findAct(t, m, actReaffirm)
	if reaffirm.enabled {
		t.Fatal("reaffirm should be dim on a single match")
	}
	dim := leftClick(m, reaffirm.x, reaffirm.y)
	if dim.review.selected != "q-alpha" {
		t.Fatalf("dim chip selected %q", dim.review.selected)
	}
	if d, ok := decisionOf(t, dec, "q-alpha"); ok && (d.Disposition != "" || d.Reaffirm) {
		t.Fatalf("dim reaffirm wrote %+v", d)
	}

	m = m.press("?")
	if !m.help {
		t.Fatal("? did not open help")
	}
	m = leftClick(m, 2, 2)
	if m.help {
		t.Fatal("left click should close help")
	}

	freshKey, freshKeyDec := queueSession(t)
	freshMouse, freshMouseDec := queueSession(t)
	keyReject := newTUI(freshKey, "Review", 0).press("r")
	mouseReject := newTUI(freshMouse, "Review", 0)
	mouseReject.width, mouseReject.height = 120, 40
	mouseReject.resize()
	mouseReject = mouseReject.press(".")
	reject := findAct(t, mouseReject, actReject)
	pressed, _ := mouseReject.Update(tea.MouseMsg{
		X: reject.x, Y: reject.y,
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
	})
	mouseReject = pressed.(tuiModel)
	released, _ := mouseReject.Update(tea.MouseMsg{
		X: reject.x, Y: reject.y,
		Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease,
	})
	mouseReject = released.(tuiModel)
	if mouseReject.review.selected != keyReject.review.selected {
		t.Fatalf("reject click selected %q, key selected %q", mouseReject.review.selected, keyReject.review.selected)
	}
	keyAlpha, _ := decisionOf(t, freshKeyDec, "q-alpha")
	mouseAlpha, _ := decisionOf(t, freshMouseDec, "q-alpha")
	if keyAlpha.Disposition != "reject" || mouseAlpha.Disposition != "reject" {
		t.Fatalf("key %q mouse %q", keyAlpha.Disposition, mouseAlpha.Disposition)
	}
	if _, ok := decisionOf(t, freshMouseDec, "q-beta"); ok {
		t.Fatal("release fired a second reject")
	}

	keyFilter := newTUI(queueSessionOnly(t), "Review", 0).press("3")
	mouseFilter := newTUI(queueSessionOnly(t), "Review", 0)
	mouseFilter.width, mouseFilter.height = 120, 40
	mouseFilter.resize()
	all := findFilterZone(t, mouseFilter, filterAll)
	mouseFilter = leftClick(mouseFilter, all.x, all.y)
	if mouseFilter.review.filter != keyFilter.review.filter || mouseFilter.review.selected != keyFilter.review.selected {
		t.Fatalf("filter click %d %q, key %d %q", mouseFilter.review.filter, mouseFilter.review.selected, keyFilter.review.filter, keyFilter.review.selected)
	}

	keyItem := newTUI(queueSessionOnly(t), "Review", 0).press("j")
	mouseItem := newTUI(queueSessionOnly(t), "Review", 0)
	mouseItem.width, mouseItem.height = 120, 40
	mouseItem.resize()
	mouseItem = mouseItem.press(".")
	beta := findItemZone(t, mouseItem, "q-beta")
	mouseItem = leftClick(mouseItem, beta.x, beta.y)
	if mouseItem.review.selected != keyItem.review.selected {
		t.Fatalf("item click selected %q, key selected %q", mouseItem.review.selected, keyItem.review.selected)
	}
	mine := findFieldZone(t, mouseItem, actMine)
	mouseItem = leftClick(mouseItem, mine.x, mine.y)
	if mouseItem.mode != modeMine {
		t.Fatal("Mine click did not focus the field")
	}
	note := findFieldZone(t, mouseItem, actNote)
	mouseItem = leftClick(mouseItem, note.x, note.y)
	if mouseItem.mode != modeNote {
		t.Fatal("Note click did not focus the field")
	}
	gapChip := findAct(t, mouseItem, actAccept)
	mouseItem = leftClick(mouseItem, gapChip.x+gapChip.w, gapChip.y)
	if mouseItem.mode != modeNav {
		t.Fatal("a click on the gap should leave the field")
	}
}

func queueSessionOnly(t *testing.T) *Session {
	t.Helper()
	sess, _ := queueSession(t)
	return sess
}

func TestTUIWheelScrollsTheList(t *testing.T) {
	root := t.TempDir()
	var doc []string
	items := make([]map[string]any, 12)
	for i := 0; i < 12; i++ {
		q := fmt.Sprintf("Sentence %02d stays.", i)
		doc = append(doc, q)
		items[i] = map[string]any{
			"id": fmt.Sprintf("q-%02d", i), "from": "grok-4.7", "path": "doc.md",
			"quote": q, "problem": q,
		}
	}
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte(strings.Join(doc, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261122-170000-wheel.json"), map[string]any{
		"version": 1,
		"items":   items,
	})
	sess, err := LoadSession(root, packets, filepath.Join(root, "decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := newTUI(sess, "Review", 0)
	m.width, m.height = 120, 40
	m.resize()
	if m.listTop != 0 {
		t.Fatalf("listTop %d", m.listTop)
	}
	if m.listPage() >= 12 {
		t.Fatalf("page %d does not scroll twelve items", m.listPage())
	}
	wheel := func(x int, button tea.MouseButton, action tea.MouseAction) {
		t.Helper()
		next, _ := m.Update(tea.MouseMsg{X: x, Y: 4, Button: button, Action: action})
		m = next.(tuiModel)
	}
	wheel(1, tea.MouseButtonWheelDown, tea.MouseActionPress)
	if m.listTop != 1 {
		t.Fatalf("wheel down listTop %d", m.listTop)
	}
	wheel(1, tea.MouseButtonWheelDown, tea.MouseActionRelease)
	if m.listTop != 1 {
		t.Fatalf("wheel release listTop %d", m.listTop)
	}
	listW, _ := m.columns()
	wheel(listW+2, tea.MouseButtonWheelDown, tea.MouseActionPress)
	if m.listTop != 1 {
		t.Fatalf("detail wheel listTop %d", m.listTop)
	}
	wheel(1, tea.MouseButtonWheelUp, tea.MouseActionPress)
	if m.listTop != 0 {
		t.Fatalf("wheel up listTop %d", m.listTop)
	}
}

func TestTUIWrapsProseAndKeepsTheDiff(t *testing.T) {
	root := t.TempDir()
	sentence := "The bridge still carries the wagon. "
	quote := strings.Repeat(sentence, 24) + "QUOTEEND."
	alt := strings.Repeat(sentence, 24) + "DIFFEND."
	problem := strings.Repeat("The wording is still too ornate. ", 8) + "PROBLEMEND."
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte(quote+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261122-180000-wrap.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "q-wrap", "from": "grok-4.7", "path": "doc.md",
			"quote": quote, "problem": problem,
			"alternates": []map[string]any{
				{"id": "a", "label": "shorter", "text": alt},
			},
		}},
	})
	sess, err := LoadSession(root, packets, filepath.Join(root, "decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := newTUI(sess, "Review", 0)
	m.width, m.height = 120, 40
	m.resize()
	view := m.View()
	bodyH := m.height - 2
	lines := strings.Split(view, "\n")
	if len(lines) < bodyH {
		t.Fatalf("lines %d", len(lines))
	}
	for i, line := range lines[:bodyH] {
		if w := lipgloss.Width(line); w != 120 {
			t.Fatalf("body line %d width %d", i, w)
		}
	}
	plainView := plain(view)
	if !strings.Contains(plainView, "PROBLEMEND.") {
		t.Fatalf("problem was clipped instead of wrapped\n%s", plainView)
	}
	if !strings.Contains(plainView, "Word diff") {
		t.Fatalf("diff is not on screen\n%s", plainView)
	}
	if strings.Contains(plainView, "QUOTEEND.") {
		t.Fatal("the quote tail should start below the fold so the scroll test is real")
	}
	listW, _ := m.columns()
	scrollBand := func(y int, marker string) {
		t.Helper()
		for i := 0; i < 8 && !strings.Contains(plain(m.View()), marker); i++ {
			next, _ := m.Update(tea.MouseMsg{
				X: listW + 2, Y: y,
				Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress,
			})
			m = next.(tuiModel)
		}
		if !strings.Contains(plain(m.View()), marker) {
			t.Fatalf("wheel did not reveal %s\n%s", marker, plain(m.View()))
		}
	}
	sc := m.screen()
	if len(sc.passages) == 0 || len(sc.diffs) == 0 {
		t.Fatalf("scroll bands passage %d diff %d", len(sc.passages), len(sc.diffs))
	}
	scrollBand(sc.passages[0].y, "QUOTEEND.")
	scrollBand(sc.diffs[0].y, "DIFFEND.")
}

func TestTUIProblemScrollsAboveTheControls(t *testing.T) {
	root := t.TempDir()
	quote := "The bridge still carries the wagon."
	problem := strings.Repeat("The instruction keeps going. ", 80) + "PROBLEMEND."
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte(quote+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261123-120000-problem.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "q-problem", "path": "doc.md", "quote": quote, "problem": problem,
		}},
	})
	sess, err := LoadSession(root, packets, filepath.Join(root, "decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := newTUI(sess, "Review", 0)
	m.width, m.height = 120, 40
	m.resize()
	view := plain(m.View())
	if strings.Contains(view, "PROBLEMEND.") {
		t.Fatalf("problem tail is on screen before scrolling\n%s", view)
	}
	if !strings.Contains(view, "The instruction keeps going.") {
		t.Fatalf("problem start missing\n%s", view)
	}
	above := findAct(t, m, actAbove)
	var band scrollBand
	found := false
	for _, b := range m.screen().passages {
		if b.slot == scrollProblem && b.h > 0 {
			band = b
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no problem scroll band\n%s", view)
	}
	if above.y < band.y+band.h {
		t.Fatalf("Above at %d overlaps the problem window %d+%d", above.y, band.y, band.h)
	}
	listW, _ := m.columns()
	for i := 0; i < 40 && !strings.Contains(plain(m.View()), "PROBLEMEND."); i++ {
		next, _ := m.Update(tea.MouseMsg{
			X: listW + 4, Y: band.y + band.h/2,
			Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress,
		})
		m = next.(tuiModel)
	}
	if !strings.Contains(plain(m.View()), "PROBLEMEND.") {
		t.Fatalf("wheel did not reveal the problem tail\n%s", plain(m.View()))
	}
	findAct(t, m, actAbove)
	findAct(t, m, actBelow)
	findAct(t, m, actChapter)
}

func findTab(t *testing.T, m tuiModel, slot, pane int) tuiZone {
	t.Helper()
	for _, z := range m.screen().zones {
		if z.kind == zoneTab && z.slot == slot && z.pane == pane {
			return z
		}
	}
	t.Fatalf("no tab slot %d pane %d", slot, pane)
	return tuiZone{}
}

func TestTUIDetailTabs(t *testing.T) {
	sess, _ := queueSession(t)
	m := newTUI(sess, "Review", 0)
	m.width, m.height = 120, 40
	m.resize()
	view := plain(m.View())
	if strings.Contains(view, "u Use my wording") {
		t.Fatal("actions stay hidden until a half shows them")
	}
	for _, want := range []string{"Alpha is still open.", "Word diff", "Current", "Proposed", "Actions"} {
		if !strings.Contains(view, want) {
			t.Fatalf("default split missing %q\n%s", want, view)
		}
	}
	if lipgloss.Width(tuiHelpLine) > 120 {
		t.Fatalf("help line is %d cells", lipgloss.Width(tuiHelpLine))
	}

	m = m.press(".")
	if m.bottomPane != paneActions || !strings.Contains(plain(m.View()), "u Use my wording") {
		t.Fatal(". should open Actions on the bottom")
	}
	if !strings.Contains(plain(m.View()), "Alpha is still open.") {
		t.Fatal("the top half should stay on Current")
	}

	m = m.press(",")
	if m.topPane != paneProposed || strings.Contains(plain(m.View()), "Alpha is still open.") {
		t.Fatal(", should move the top half off Current")
	}
	m = m.press(",")
	if m.topPane != paneActions {
		t.Fatalf("second , pane %d", m.topPane)
	}
	m = m.press(",")
	if m.topPane != paneCurrent {
		t.Fatalf(", did not wrap to Current (%d)", m.topPane)
	}

	fresh := newTUI(sess, "Review", 0)
	fresh.width, fresh.height = 120, 40
	fresh.resize()
	tab := findTab(t, fresh, 1, paneActions)
	fresh = leftClick(fresh, tab.x, tab.y)
	if fresh.bottomPane != paneActions || fresh.topPane != paneCurrent {
		t.Fatalf("tab click top %d bottom %d", fresh.topPane, fresh.bottomPane)
	}
	if !strings.Contains(plain(fresh.View()), "a Accept") {
		t.Fatal("clicking Actions did not show the decision chips")
	}
	topProposed := findTab(t, fresh, 0, paneProposed)
	fresh = leftClick(fresh, topProposed.x, topProposed.y)
	if fresh.topPane != paneProposed || fresh.bottomPane != paneActions {
		t.Fatalf("top tab click top %d bottom %d", fresh.topPane, fresh.bottomPane)
	}

	editing := newTUI(sess, "Review", 0).press("enter", "Z")
	if editing.mode != modeMine || editing.bottomPane != paneActions {
		t.Fatalf("enter mode %d bottom %d", editing.mode, editing.bottomPane)
	}
	if editing.review.mine != "Z" {
		t.Fatalf("mine %q", editing.review.mine)
	}
	editing = editing.press(",")
	if editing.topPane != paneCurrent || editing.review.mine != "Z," {
		t.Fatalf("comma while editing pane %d mine %q", editing.topPane, editing.review.mine)
	}
}

func TestTUIControlsSitOnTheirPanes(t *testing.T) {
	sess, _ := queueSession(t)
	m := newTUI(sess, "Review", 0)
	m.width, m.height = 120, 40
	m.resize()
	view := plain(m.View())
	for _, want := range []string{"all above", "2 below", "Quote -> Mine", "Alt -> Mine", "Diff alt"} {
		if !strings.Contains(view, want) {
			t.Fatalf("default split missing %q\n%s", want, view)
		}
	}
	g := m.geom()
	splitY := g.inset + (g.contentH+1)/2
	above := findAct(t, m, actAbove)
	below := findAct(t, m, actBelow)
	chapter := findAct(t, m, actChapter)
	quote := findAct(t, m, actQuote)
	altMine := findAct(t, m, actAltMine)
	diff := findAct(t, m, actDiff)
	for _, z := range []tuiZone{above, below, chapter} {
		if z.y >= splitY {
			t.Fatalf("%s chip is on the bottom half at y %d", z.act, z.y)
		}
	}
	for _, z := range []tuiZone{quote, altMine, diff} {
		if z.y < splitY {
			t.Fatalf("%s chip is on the top half at y %d", z.act, z.y)
		}
	}
	if above.enabled {
		t.Fatal("above should be dim when the quote starts the file")
	}
	dim := leftClick(m, above.x, above.y)
	if dim.review.beforeLines != 2 || !strings.Contains(plain(dim.View()), "2 below") {
		t.Fatal("a dim above chip changed the passage window")
	}
	opened := leftClick(m, below.x, below.y)
	openedView := plain(opened.View())
	if opened.review.afterLines != 8 || !strings.Contains(openedView, "whole chapter") || !strings.Contains(openedView, "seventh line keeps what was built") {
		t.Fatalf("below click after %d\n%s", opened.review.afterLines, openedView)
	}
	copied := leftClick(m, quote.x, quote.y)
	if copied.review.mine != "Alpha sentence stays." || copied.bottomPane != paneProposed {
		t.Fatalf("quote copy mine %q bottom %d", copied.review.mine, copied.bottomPane)
	}
	if !strings.Contains(plain(copied.View()), "Word diff (Quote -> Mine)") {
		t.Fatalf("diff did not follow the copy\n%s", plain(copied.View()))
	}

	both := m.press(".", ",", ",")
	if both.topPane != paneActions || both.bottomPane != paneActions {
		t.Fatalf("both actions top %d bottom %d", both.topPane, both.bottomPane)
	}
	hidden := plain(both.View())
	for _, gone := range []string{"2 below", "Quote -> Mine", "Diff alt"} {
		if strings.Contains(hidden, gone) {
			t.Fatalf("actions still shows %q\n%s", gone, hidden)
		}
	}
}

func findPaneAct(t *testing.T, m tuiModel, slot int, act string, alt int) tuiZone {
	t.Helper()
	for _, z := range m.screen().zones {
		if z.kind == zoneAct && z.slot == slot && z.act == act && z.alt == alt {
			return z
		}
	}
	t.Fatalf("no %s slot %d alt %d", act, slot, alt)
	return tuiZone{}
}

func TestTUIProposedOptionsArePerPane(t *testing.T) {
	sess, _ := queueSession(t)
	m := newTUI(sess, "Review", 0)
	m.width, m.height = 120, 40
	m.resize()
	m = m.press(",")
	if m.topPane != paneProposed || m.bottomPane != paneProposed {
		t.Fatalf("both proposed top %d bottom %d", m.topPane, m.bottomPane)
	}
	view := plain(m.View())
	if !strings.Contains(view, "----") || !strings.Contains(view, "alpha-alt") {
		t.Fatalf("proposed list or rule missing\n%s", view)
	}
	topIgnore := findPaneAct(t, m, 0, actAlt, 1)
	bottomAlt := findPaneAct(t, m, 1, actAlt, 0)
	if topIgnore.y >= bottomAlt.y {
		t.Fatal("expected the top half's alternate above the bottom half's")
	}
	m = leftClick(m, topIgnore.x, topIgnore.y)
	m = leftClick(m, bottomAlt.x, bottomAlt.y)
	if m.topAlt != 1 || m.bottomAlt != 0 {
		t.Fatalf("alts top %d bottom %d", m.topAlt, m.bottomAlt)
	}
	view = plain(m.View())
	if !strings.Contains(view, "this was actually good") || !strings.Contains(view, "Quote -> alpha-alt") {
		t.Fatalf("the two diffs are not independent\n%s", view)
	}
	m = m.press("tab")
	if m.topAlt != 1 || m.bottomAlt != 1 {
		t.Fatalf("tab moved top %d bottom %d", m.topAlt, m.bottomAlt)
	}
	if strings.Contains(plain(m.View()), "Quote -> alpha-alt") {
		t.Fatal("tab changed the half that was not clicked last")
	}

	fresh := newTUI(sess, "Review", 0)
	fresh.width, fresh.height = 120, 40
	fresh.resize()
	fresh = fresh.press(",")
	quote := findPaneAct(t, fresh, 1, actQuote, 0)
	fresh = leftClick(fresh, quote.x, quote.y)
	diff := findPaneAct(t, fresh, 0, actDiff, 0)
	if !diff.enabled {
		t.Fatal("diff chip should enable once Mine has text")
	}
	fresh = leftClick(fresh, diff.x, diff.y)
	if !fresh.topDiffAlt || fresh.bottomDiffAlt {
		t.Fatalf("diff flags top %v bottom %v", fresh.topDiffAlt, fresh.bottomDiffAlt)
	}
	split := plain(fresh.View())
	if !strings.Contains(split, "Word diff (Quote -> Mine)") || !strings.Contains(split, "Quote -> alpha-alt") {
		t.Fatalf("diff target is not per pane\n%s", split)
	}

	actions := newTUI(sess, "Review", 0)
	actions.width, actions.height = 120, 40
	actions.resize()
	actions = actions.press(".", ",", ",")
	if actions.topPane != paneActions || actions.bottomPane != paneActions {
		t.Fatalf("actions top %d bottom %d", actions.topPane, actions.bottomPane)
	}
	if actions.mine.Height() < 8 {
		t.Fatalf("mine editor height %d", actions.mine.Height())
	}
	mine := findFieldZone(t, actions, actMine)
	note := findFieldZone(t, actions, actNote)
	if mine.h < 8 || note.h >= mine.h {
		t.Fatalf("mine zone %d note zone %d", mine.h, note.h)
	}
	gone := plain(actions.View())
	if strings.Contains(gone, "Alpha rewritten.") || strings.Contains(gone, "alpha-alt") {
		t.Fatalf("actions still shows the proposed text\n%s", gone)
	}
}

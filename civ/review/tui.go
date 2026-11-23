package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const tuiHelpLine = "j/k move  ,/. panes  tab  a accept  u wording  r reject  h/e/s hold  w withdraw  c compare  enter mine  N note  ? help"

const tuiHelpText = `Review

Filters: 1 Open   2 Hold   3 All   4 Decided   5 Stale
Items: j/k or n/p, and up/down. Ctrl+J and Ctrl+K move while a field is focused.
Alternates: Tab and Shift+Tab.
a Accept. When Ignore is focused, a records Ignore this.
u Use my wording. Ctrl+S does that while a field is focused.
r Reject. w Withdraw.
h Hold later. e Hold explain. s Hold research.
o Reopen. f Reaffirm. c Compare. d switches the diff between Mine and the alternate.
Context: [ above, ] below, C whole chapter, t tight.
q copies the quote into Mine. y copies the focused alternate into Mine.
Enter edits Mine. N edits Note. Esc leaves the field.
Ctrl+R reloads packets. ? closes this help. Ctrl+C quits.

The detail pane is split. Each half shows Current, Proposed, or Actions.
, cycles the top half. . cycles the bottom half. The same view may sit on both.
Enter and N show Actions on the bottom when neither half is showing it.
Current has the passage controls: above, below, chapter, and tight.
Proposed lists the alternates, then a line, then Quote -> Mine, Alt -> Mine, and the diff.
When both halves show Proposed, each half keeps its own alternate and diff target.
Tab and d follow the half you clicked last, or the only Proposed half.
Actions has the decision chips and a large Mine field. The alternate text is not repeated there.

The window binds Use my wording to Ctrl+Enter. A terminal usually does not
send that chord, so this screen uses u and Ctrl+S for the same action.
`

type tuiMode int

const (
	modeNav tuiMode = iota
	modeMine
	modeNote
)

type tuiQuitMsg struct{}

type tuiModel struct {
	review        reviewView
	title         string
	quitAfter     time.Duration
	width         int
	height        int
	listTop       int
	passageTop    int
	topDiffTop    int
	bottomDiffTop int
	topPane       int
	bottomPane    int
	topAlt        int
	bottomAlt     int
	topDiffAlt    bool
	bottomDiffAlt bool
	altSlot       int
	anchorID      string
	mode          tuiMode
	help          bool
	mine          textarea.Model
	note          textarea.Model
}

func newTUI(session *Session, title string, quitAfter time.Duration) tuiModel {
	if title == "" {
		title = "Review"
	}
	v := newReviewView(session)
	mine := textarea.New()
	note := textarea.New()
	for _, box := range []*textarea.Model{&mine, &note} {
		box.ShowLineNumbers = false
		box.Prompt = ""
		box.CharLimit = 0
	}
	mine.Placeholder = "Mine"
	note.Placeholder = "Note"
	mine.SetValue(v.mine)
	note.SetValue(v.note)
	m := tuiModel{
		review:     v,
		title:      title,
		quitAfter:  quitAfter,
		width:      120,
		height:     40,
		topPane:    paneCurrent,
		bottomPane: paneProposed,
		topAlt:     v.altIndex,
		bottomAlt:  v.altIndex,
		altSlot:    1,
		mine:       mine,
		note:       note,
	}
	m.resize()
	return m
}

func runTUI(args []string) error {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	common := bindCommon(fs)
	quitAfter := fs.Duration("quit-after", 0, "quit after this long; 0 waits")
	titleFlag := fs.String("title", "Review", "screen title")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, packets, decisions, err := common.resolve()
	if err != nil {
		return err
	}
	session, err := LoadSession(root, packets, decisions)
	if err != nil {
		return err
	}
	p := tea.NewProgram(newTUI(session, *titleFlag, *quitAfter), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}

func (m tuiModel) Init() tea.Cmd {
	if m.quitAfter <= 0 {
		return nil
	}
	return tea.Tick(m.quitAfter, func(time.Time) tea.Msg { return tuiQuitMsg{} })
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		m.ensureList()
		return m, nil
	case tuiQuitMsg:
		return m, tea.Quit
	case tea.MouseMsg:
		return m.onMouse(msg)
	case tea.KeyMsg:
		if m.help {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "?":
				m.help = false
			}
			return m, nil
		}
		if m.mode == modeNav && msg.String() == "?" {
			m.help = true
			return m, nil
		}
		if m.mode != modeNav {
			return m.updateField(msg)
		}
		return m.updateNav(msg)
	}
	return m, nil
}

func (m tuiModel) updateNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "j", "n", "down", "ctrl+j":
		m.pullFields()
		m.review.next()
		return m.synced()
	case "k", "p", "up", "ctrl+k":
		m.pullFields()
		m.review.prev()
		return m.synced()
	case "tab":
		m.cyclePaneAlt(1)
		return m, nil
	case "shift+tab":
		m.cyclePaneAlt(-1)
		return m, nil
	case "a":
		m.pullFields()
		m.review.accept()
		return m.synced()
	case "u", "ctrl+s":
		m.pullFields()
		m.review.revise()
		return m.synced()
	case "r":
		m.pullFields()
		m.review.reject()
		return m.synced()
	case "w":
		m.pullFields()
		m.review.withdraw()
		return m.synced()
	case "h":
		m.pullFields()
		m.review.hold("later")
		return m.synced()
	case "e":
		m.pullFields()
		m.review.hold("explain")
		return m.synced()
	case "s":
		m.pullFields()
		m.review.hold("research")
		return m.synced()
	case "o":
		m.pullFields()
		m.review.reopen()
		return m.synced()
	case "f":
		m.pullFields()
		m.review.reaffirm()
		return m.synced()
	case "c":
		m.review.toggleCompare()
		m.passageTop = 0
		m.topDiffTop, m.bottomDiffTop = 0, 0
		return m, nil
	case "d":
		m.togglePaneDiff()
		return m, nil
	case "[":
		m.review.expandAbove()
		return m, nil
	case "]":
		m.review.expandBelow()
		return m, nil
	case "C":
		m.review.showWholeChapter()
		return m, nil
	case "t":
		m.review.tightenContext()
		return m, nil
	case "q":
		m.quoteToMine()
		return m.synced()
	case "y":
		m.altToMine()
		return m.synced()
	case "enter":
		return m.editingMine()
	case "N":
		return m.editingNote()
	case "1":
		m.pullFields()
		m.review.setFilter(filterOpen)
		return m.synced()
	case "2":
		m.pullFields()
		m.review.setFilter(filterHold)
		return m.synced()
	case "3":
		m.pullFields()
		m.review.setFilter(filterAll)
		return m.synced()
	case "4":
		m.pullFields()
		m.review.setFilter(filterDecided)
		return m.synced()
	case "5":
		m.pullFields()
		m.review.setFilter(filterStale)
		return m.synced()
	case "ctrl+r":
		m.pullFields()
		m.review.reload()
		return m.synced()
	case ",":
		m.topPane = (m.topPane + 1) % paneCount
		m.fitEditor()
		return m, nil
	case ".":
		m.bottomPane = (m.bottomPane + 1) % paneCount
		m.fitEditor()
		return m, nil
	}
	return m, nil
}

func (m tuiModel) updateField(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.pullFields()
		m.mode = modeNav
		m.review.mineFocused = false
		m.review.noteFocused = false
		m.mine.Blur()
		m.note.Blur()
		return m, nil
	case "ctrl+c":
		m.pullFields()
		return m, tea.Quit
	case "ctrl+j":
		m.pullFields()
		m.review.next()
		return m.synced()
	case "ctrl+k":
		m.pullFields()
		m.review.prev()
		return m.synced()
	case "ctrl+r":
		m.pullFields()
		m.review.reload()
		return m.synced()
	case "ctrl+s":
		m.pullFields()
		m.review.revise()
		return m.synced()
	}
	var cmd tea.Cmd
	if m.mode == modeMine {
		m.mine, cmd = m.mine.Update(msg)
		m.review.mine = m.mine.Value()
	} else {
		m.note, cmd = m.note.Update(msg)
		m.review.note = m.note.Value()
	}
	return m, cmd
}

func (m *tuiModel) focusMine() tea.Cmd {
	m.ensureActions()
	m.mode = modeMine
	m.review.mineFocused = true
	m.review.noteFocused = false
	m.note.Blur()
	return m.mine.Focus()
}

func (m *tuiModel) focusNote() tea.Cmd {
	m.ensureActions()
	m.mode = modeNote
	m.review.noteFocused = true
	m.review.mineFocused = false
	m.mine.Blur()
	return m.note.Focus()
}

// ensureActions puts Actions on the bottom when neither half is showing
// the fields. Enter and N are how Mine and Note are reached from the keys.
func (m *tuiModel) ensureActions() {
	if m.topPane == paneActions || m.bottomPane == paneActions {
		return
	}
	m.bottomPane = paneActions
	m.fitEditor()
}

// proposedFocus is the Proposed half that Tab, d, q, and y act on.
// When both halves are Proposed, that is the one whose options were clicked
// last. Otherwise it is the only Proposed half.
func (m tuiModel) proposedFocus() (int, bool) {
	top := m.topPane == paneProposed
	bottom := m.bottomPane == paneProposed
	switch {
	case top && bottom:
		if m.altSlot != 0 {
			return 1, true
		}
		return 0, true
	case top:
		return 0, true
	case bottom:
		return 1, true
	default:
		return 0, false
	}
}

func (m tuiModel) paneAlt(slot int) int {
	if slot == 0 {
		return m.topAlt
	}
	return m.bottomAlt
}

func (m tuiModel) paneDiffAlt(slot int) bool {
	if slot == 0 {
		return m.topDiffAlt
	}
	return m.bottomDiffAlt
}

func (m tuiModel) paneDiffTop(slot int) int {
	if slot == 0 {
		return m.topDiffTop
	}
	return m.bottomDiffTop
}

// setPaneAlt selects an alternate for one half. A half that is not showing
// Proposed follows that choice, so opening it later is not a stale list.
// The other Proposed half is left as it is.
func (m *tuiModel) setPaneAlt(slot, idx int) {
	if slot == 0 {
		m.topAlt = idx
		m.topDiffTop = 0
	} else {
		m.bottomAlt = idx
		m.bottomDiffTop = 0
	}
	m.review.selectAlt(idx)
	if m.topPane != paneProposed {
		m.topAlt = m.review.altIndex
	}
	if m.bottomPane != paneProposed {
		m.bottomAlt = m.review.altIndex
	}
}

func (m *tuiModel) setPaneDiffAlt(slot int, altOnly bool) {
	if slot == 0 {
		m.topDiffAlt = altOnly
		m.topDiffTop = 0
	} else {
		m.bottomDiffAlt = altOnly
		m.bottomDiffTop = 0
	}
	m.review.diffAltOnly = altOnly
	if m.topPane != paneProposed {
		m.topDiffAlt = altOnly
	}
	if m.bottomPane != paneProposed {
		m.bottomDiffAlt = altOnly
	}
}

func (m *tuiModel) cyclePaneAlt(delta int) {
	slot, ok := m.proposedFocus()
	if !ok {
		m.review.cycleAlt(delta)
		m.topAlt = m.review.altIndex
		m.bottomAlt = m.review.altIndex
		return
	}
	if m.topPane == paneProposed && m.bottomPane == paneProposed {
		it, has := m.review.current()
		if !has || len(it.Alternates) == 0 {
			return
		}
		n := len(it.Alternates)
		idx := m.paneAlt(slot)
		if idx < 0 || idx >= n {
			idx = 0
		}
		idx = (idx + delta%n + n) % n
		m.altSlot = slot
		m.setPaneAlt(slot, idx)
		return
	}
	m.review.cycleAlt(delta)
	m.altSlot = slot
	m.setPaneAlt(slot, m.review.altIndex)
}

func (m *tuiModel) togglePaneDiff() {
	if m.topPane == paneProposed && m.bottomPane == paneProposed {
		slot, _ := m.proposedFocus()
		m.altSlot = slot
		m.setPaneDiffAlt(slot, !m.paneDiffAlt(slot))
		return
	}
	m.review.toggleDiffTarget()
	m.topDiffAlt = m.review.diffAltOnly
	m.bottomDiffAlt = m.review.diffAltOnly
}

func (m *tuiModel) quoteToMine() {
	m.review.copyQuoteToMine()
	if slot, ok := m.proposedFocus(); ok && m.topPane == paneProposed && m.bottomPane == paneProposed {
		m.setPaneDiffAlt(slot, false)
		return
	}
	m.topDiffAlt = false
	m.bottomDiffAlt = false
}

func (m *tuiModel) altToMine() {
	if slot, ok := m.proposedFocus(); ok {
		m.review.selectAlt(m.paneAlt(slot))
		m.altSlot = slot
		if m.topPane == paneProposed && m.bottomPane == paneProposed {
			m.review.copyAltToMine()
			m.setPaneDiffAlt(slot, false)
			return
		}
	}
	m.review.copyAltToMine()
	m.topDiffAlt = false
	m.bottomDiffAlt = false
}

// synced applies field and list updates before the model is returned.
// afterChange takes a pointer, so it has to run before the return copies m.
func (m tuiModel) synced() (tea.Model, tea.Cmd) {
	cmd := m.afterChange()
	return m, cmd
}

func (m tuiModel) editingMine() (tea.Model, tea.Cmd) {
	cmd := m.focusMine()
	return m, cmd
}

func (m tuiModel) editingNote() (tea.Model, tea.Cmd) {
	cmd := m.focusNote()
	return m, cmd
}

func (m tuiModel) onMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if tea.MouseEvent(msg).IsWheel() {
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		return m.scrollList(msg)
	}
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
		return m, nil
	}
	if m.help {
		m.help = false
		return m, nil
	}
	z, ok := m.screen().hit(msg.X, msg.Y)
	if !ok || !z.enabled {
		if !ok && m.mode != modeNav {
			m.blurFields()
		}
		return m, nil
	}
	m.pullFields()
	switch z.kind {
	case zoneFilter:
		m.review.setFilter(z.filter)
		return m.synced()
	case zoneItem:
		if z.itemID != "" && z.itemID != m.review.selected {
			m.review.choose(z.itemID)
			return m.synced()
		}
		return m, nil
	case zoneField:
		if z.act == actNote {
			return m.editingNote()
		}
		return m.editingMine()
	case zoneAct:
		return m.runAct(z)
	case zoneTab:
		if z.slot == 0 {
			m.topPane = z.pane
		} else {
			m.bottomPane = z.pane
		}
		if m.mode != modeNav && m.topPane != paneActions && m.bottomPane != paneActions {
			m.blurFields()
		}
		m.fitEditor()
		return m, nil
	}
	return m, nil
}

func (m tuiModel) scrollList(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	delta := 0
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		delta = -1
	case tea.MouseButtonWheelDown:
		delta = 1
	default:
		return m, nil
	}
	listW, _ := m.columns()
	if msg.X < listW {
		m.listTop += delta
		m.clampList()
		return m, nil
	}
	sc := m.screen()
	const step = 3
	if b, ok := bandAt(sc.diffs, msg.Y); ok {
		if b.slot == 0 {
			m.topDiffTop = clampTop(m.topDiffTop+delta*step, b.n, b.h)
		} else {
			m.bottomDiffTop = clampTop(m.bottomDiffTop+delta*step, b.n, b.h)
		}
		return m, nil
	}
	if b, ok := bandAt(sc.passages, msg.Y); ok {
		m.passageTop = clampTop(m.passageTop+delta*step, b.n, b.h)
	}
	return m, nil
}

func (m tuiModel) runAct(z tuiZone) (tea.Model, tea.Cmd) {
	switch z.act {
	case actAccept:
		m.review.accept()
		return m.synced()
	case actRevise:
		m.review.revise()
		return m.synced()
	case actReject:
		m.review.reject()
		return m.synced()
	case actWithdraw:
		m.review.withdraw()
		return m.synced()
	case actHoldResearch:
		m.review.hold("research")
		return m.synced()
	case actHoldExplain:
		m.review.hold("explain")
		return m.synced()
	case actHoldLater:
		m.review.hold("later")
		return m.synced()
	case actReopen:
		m.review.reopen()
		return m.synced()
	case actReaffirm:
		m.review.reaffirm()
		return m.synced()
	case actCompare:
		m.review.toggleCompare()
		m.passageTop = 0
		m.topDiffTop, m.bottomDiffTop = 0, 0
		return m, nil
	case actDiff:
		m.altSlot = z.slot
		m.setPaneDiffAlt(z.slot, !m.paneDiffAlt(z.slot))
		return m, nil
	case actAbove:
		m.review.expandAbove()
		return m, nil
	case actBelow:
		m.review.expandBelow()
		return m, nil
	case actChapter:
		m.review.showWholeChapter()
		return m, nil
	case actTight:
		m.review.tightenContext()
		return m, nil
	case actQuote:
		m.altSlot = z.slot
		m.review.copyQuoteToMine()
		m.setPaneDiffAlt(z.slot, false)
		return m.synced()
	case actAltMine:
		m.altSlot = z.slot
		m.review.selectAlt(m.paneAlt(z.slot))
		m.review.copyAltToMine()
		m.setPaneDiffAlt(z.slot, false)
		return m.synced()
	case actAlt:
		m.altSlot = z.slot
		m.setPaneAlt(z.slot, z.alt)
		return m, nil
	}
	return m, nil
}

func (m *tuiModel) blurFields() {
	m.pullFields()
	m.mode = modeNav
	m.review.mineFocused = false
	m.review.noteFocused = false
	m.mine.Blur()
	m.note.Blur()
}

func (m *tuiModel) pullFields() {
	m.review.mine = m.mine.Value()
	m.review.note = m.note.Value()
}

func (m *tuiModel) afterChange() tea.Cmd {
	if m.anchorID != m.review.selected {
		m.anchorID = m.review.selected
		m.passageTop = 0
		m.topDiffTop, m.bottomDiffTop = 0, 0
		m.topAlt = m.review.altIndex
		m.bottomAlt = m.review.altIndex
		m.topDiffAlt = false
		m.bottomDiffAlt = false
	}
	if _, ok := m.review.current(); !ok {
		m.mode = modeNav
		m.review.mineFocused = false
		m.review.noteFocused = false
		m.mine.Blur()
		m.note.Blur()
	}
	m.syncFields()
	m.ensureList()
	m.fitEditor()
	switch m.mode {
	case modeMine:
		return textarea.Blink
	case modeNote:
		return textarea.Blink
	default:
		return nil
	}
}

func (m *tuiModel) syncFields() {
	if m.mine.Value() != m.review.mine {
		m.mine.SetValue(m.review.mine)
	}
	if m.note.Value() != m.review.note {
		m.note.SetValue(m.review.note)
	}
}

func (m *tuiModel) resize() {
	m.fitEditor()
}

func (m tuiModel) columns() (listW, detailW int) {
	w := m.width
	if w < 3 {
		return 1, 1
	}
	listW = w / 3
	if listW < 28 {
		listW = 28
	}
	if listW > 40 {
		listW = 40
	}
	if listW > w-30 {
		listW = w * 2 / 5
	}
	if listW < 10 {
		listW = w / 2
	}
	// The gutter is one column between the panes. It comes out of the
	// width before either pane is drawn, or the row wraps and clicks slide.
	if listW > w-frameGutter-1 {
		listW = w - frameGutter - 1
	}
	if listW < 1 {
		listW = 1
	}
	detailW = w - frameGutter - listW
	if detailW < 1 {
		detailW = 1
		listW = w - frameGutter - 1
	}
	return listW, detailW
}

func (m tuiModel) listPage() int {
	g := m.geom()
	chips, _ := placeChips(m.filterChips(), g.listInner)
	header := 1 + len(chips) + 1
	rows := (g.contentH - header) / 3
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m *tuiModel) ensureList() {
	vis := m.review.visible()
	if len(vis) == 0 {
		m.listTop = 0
		return
	}
	idx := 0
	for i, it := range vis {
		if it.ID == m.review.selected {
			idx = i
			break
		}
	}
	page := m.listPage()
	if idx < m.listTop {
		m.listTop = idx
	}
	if idx >= m.listTop+page {
		m.listTop = idx - page + 1
	}
	maxTop := len(vis) - page
	if maxTop < 0 {
		maxTop = 0
	}
	if m.listTop > maxTop {
		m.listTop = maxTop
	}
	if m.listTop < 0 {
		m.listTop = 0
	}
}

func (m *tuiModel) clampList() {
	vis := m.review.visible()
	page := m.listPage()
	maxTop := len(vis) - page
	if maxTop < 0 {
		maxTop = 0
	}
	if m.listTop > maxTop {
		m.listTop = maxTop
	}
	if m.listTop < 0 {
		m.listTop = 0
	}
}

func (m tuiModel) View() string {
	if m.help {
		return tuiHelpText
	}
	if m.width < 1 || m.height < 1 {
		return m.title + "\n"
	}
	return m.screen().view
}

func (m tuiModel) detailHead(it Item) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  (%s)\n", it.Path, it.ID)
	if it.From != "" {
		fmt.Fprintf(&b, "from: %s\n", it.From)
	}
	h := m.review.session.Hits[it.ID]
	b.WriteString(m.review.matchLabel(it, h))
	b.WriteByte('\n')
	if it.Locked {
		b.WriteString(lockedBanner)
		b.WriteByte('\n')
	}
	if it.Problem != "" {
		b.WriteString(it.Problem)
		b.WriteByte('\n')
	}
	answers := m.review.session.Answers[it.ID]
	for i, ans := range answers {
		prefix := "Answer: "
		if len(answers) > 1 {
			prefix = fmt.Sprintf("Answer %d: ", i+1)
		}
		b.WriteString(prefix + ans.Text + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func sliceFit(lines []string, top, n, width int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		line := ""
		if j := top + i; j >= 0 && j < len(lines) {
			line = lines[j]
		}
		out[i] = fitCells(line, width)
	}
	return out
}

func (m tuiModel) passageLines(it Item, width int) []string {
	lines := []string{"Passage"}
	h := m.review.session.Hits[it.ID]
	if pre, match, post, ok := snippet(m.review.session.bytes(it.Path), h, m.review.beforeLines, m.review.afterLines); ok {
		lines = append(lines, wrapRuns([]cellRun{
			{text: pre},
			{text: match, styled: true, style: quoteStyle},
			{text: post},
		}, width)...)
		return lines
	}
	lines = append(lines, "The quote is not a single match.")
	lines = append(lines, wrapPlain(it.Quote, width)...)
	return lines
}

func altAt(it Item, index int) (Alternate, bool) {
	if index < 0 || index >= len(it.Alternates) {
		return Alternate{}, false
	}
	return it.Alternates[index], true
}

func (m tuiModel) diffLines(it Item, width, altIndex int, diffAltOnly bool) []string {
	alt, hasAlt := altAt(it, altIndex)
	title, target, hasTarget, showIgnore, reason := diffChoice(it, alt, hasAlt, m.review.mine, diffAltOnly)
	lines := wrapPlain(title, width)
	switch {
	case showIgnore:
		lines = append(lines, wrapPlain(reason, width)...)
	case hasTarget:
		body := wrapRuns(diffRuns(WordDiff(it.Quote, target)), width)
		if len(body) == 0 {
			lines = append(lines, "(empty)")
		} else {
			lines = append(lines, body...)
		}
	default:
		lines = append(lines, "No alternate selected.")
	}
	return lines
}

func diffRuns(pieces []Piece) []cellRun {
	out := make([]cellRun, 0, len(pieces))
	for _, p := range pieces {
		if p.Text == "" {
			continue
		}
		r := cellRun{text: p.Text}
		switch p.Op {
		case DiffDelete:
			r.styled = true
			r.style = diffDelStyle
		case DiffInsert:
			r.styled = true
			r.style = diffInsStyle
		}
		out = append(out, r)
	}
	return out
}

func (m tuiModel) compareLines(it Item, width int) []string {
	type col struct {
		title string
		from  string
		body  string
	}
	cols := []col{{title: "Current", body: it.Quote}}
	for _, a := range it.Alternates {
		label := altLabel(a)
		content := a.Text
		if a.Ignore {
			label = "Ignore"
			content = a.Reason
			if content == "" {
				content = "Ignore."
			}
		}
		cols = append(cols, col{title: label, from: SuggestionFrom(it, a), body: content})
	}
	if strings.TrimSpace(m.review.mine) != "" {
		cols = append(cols, col{title: "Mine", body: m.review.mine})
	}
	var lines []string
	for i, c := range cols {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, wrapPlain(c.title, width)...)
		if c.from != "" {
			lines = append(lines, wrapPlain("from: "+c.from, width)...)
		}
		lines = append(lines, wrapPlain(c.body, width)...)
	}
	return lines
}

var (
	selectedStyle = lipgloss.NewStyle().Bold(true)
	quoteStyle    = lipgloss.NewStyle().Reverse(true)
	chipHot       = lipgloss.NewStyle().Reverse(true)
	chipOff       = lipgloss.NewStyle().Faint(true)
	diffDelStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "1", Dark: "203"})
	diffInsStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "2", Dark: "42"})
)

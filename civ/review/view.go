package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Filters for the item list. Both the window and the terminal use these.
const (
	filterOpen = iota
	filterAll
	filterHold
	filterDecided
	filterStale
)

const lockedBanner = "LOCKED LINE: Preserve verbatim per civ/notes/00_civ_authoritative_bones.md and Section 1 of civ/notes/06_editing_guide.md."

type itemDraft struct {
	mine string
	note string
}

// reviewView is the review session the window and the terminal both drive.
// Methods mutate it directly. A UI that rebuilds on a callback wraps the call.
type reviewView struct {
	session     *Session
	filter      int
	selected    string
	altIndex    int
	mine        string
	note        string
	drafts      map[string]itemDraft
	diffAltOnly bool
	compare     bool
	mineFocused bool
	noteFocused bool
	status      string
	// beforeLines and afterLines are lines of the chapter shown past the
	// quote. Negative means that side of the file.
	beforeLines int
	afterLines  int
}

func newReviewView(session *Session) reviewView {
	v := reviewView{
		session:  session,
		filter:   filterOpen,
		altIndex: -1,
		drafts:   map[string]itemDraft{},
	}
	if v.session == nil {
		v.session = &Session{
			Files:     map[string][]byte{},
			Hits:      map[string]Hit{},
			Answers:   map[string][]Answer{},
			Applied:   map[string]bool{},
			Decisions: DecisionFile{Version: 1, Items: map[string]Decision{}},
		}
	}
	v.resetContext()
	v.selected = v.firstID()
	v.loadDraft()
	return v
}

func (v *reviewView) typing() bool { return v.mineFocused || v.noteFocused }

func (v *reviewView) visible() []Item {
	var out []Item
	for _, it := range v.session.Items {
		d := v.session.Decisions.Items[it.ID].Disposition
		switch v.filter {
		case filterOpen:
			if d != "" {
				continue
			}
		case filterHold:
			if d != "hold" {
				continue
			}
		case filterDecided:
			switch d {
			case "accept", "revise", "reject", "withdraw":
			default:
				continue
			}
		case filterStale:
			if !v.session.IsStale(it.ID) {
				continue
			}
		}
		out = append(out, it)
	}
	return out
}

func (v *reviewView) firstID() string {
	if vis := v.visible(); len(vis) > 0 {
		return vis[0].ID
	}
	return ""
}

func (v *reviewView) selectedVisible() bool {
	for _, it := range v.visible() {
		if it.ID == v.selected {
			return true
		}
	}
	return false
}

func (v *reviewView) current() (Item, bool) {
	for _, it := range v.session.Items {
		if it.ID == v.selected {
			return it, true
		}
	}
	return Item{}, false
}

func (v *reviewView) alt() (Alternate, bool) {
	it, ok := v.current()
	if !ok || v.altIndex < 0 || v.altIndex >= len(it.Alternates) {
		return Alternate{}, false
	}
	return it.Alternates[v.altIndex], true
}

func (v *reviewView) altID() string {
	if a, ok := v.alt(); ok {
		return a.ID
	}
	return ""
}

func (v *reviewView) stashDraft() {
	if v.selected == "" {
		return
	}
	saved := v.session.Decisions.Items[v.selected]
	if v.mine != saved.Text || v.note != saved.Note {
		if v.drafts == nil {
			v.drafts = map[string]itemDraft{}
		}
		v.drafts[v.selected] = itemDraft{mine: v.mine, note: v.note}
	} else {
		delete(v.drafts, v.selected)
	}
}

func (v *reviewView) loadDraft() {
	d := v.session.Decisions.Items[v.selected]
	if dr, ok := v.drafts[v.selected]; ok {
		v.mine = dr.mine
		v.note = dr.note
	} else {
		v.mine = d.Text
		v.note = d.Note
	}
	v.altIndex = -1
	it, ok := v.current()
	if !ok {
		return
	}
	if d.Alternate != "" {
		for i, a := range it.Alternates {
			if a.ID == d.Alternate {
				v.altIndex = i
				return
			}
		}
	}
	if len(it.Alternates) > 0 {
		v.altIndex = 0
	}
}

func (v *reviewView) choose(id string) {
	if v.selected != id {
		v.stashDraft()
		v.resetContext()
	}
	v.selected = id
	v.compare = false
	v.diffAltOnly = false
	v.loadDraft()
}

// clearSelection drops the passage and the actions when the current filter
// has nothing left to show.
func (v *reviewView) clearSelection() {
	v.selected = ""
	v.altIndex = -1
	v.compare = false
	v.diffAltOnly = false
	v.mine = ""
	v.note = ""
	v.mineFocused = false
	v.noteFocused = false
	v.resetContext()
}

func (v *reviewView) next() { v.move(1) }
func (v *reviewView) prev() { v.move(-1) }

// followingID is the next item in the current filter. If the selected item is
// last in the filter and will drop out of it on save (e.g., Open or Hold),
// it falls back to the preceding visible item rather than stranding the view.
func (v *reviewView) followingID() string {
	vis := v.visible()
	for n, it := range vis {
		if it.ID != v.selected {
			continue
		}
		if n+1 < len(vis) {
			return vis[n+1].ID
		}
		if n > 0 && (v.filter == filterOpen || v.filter == filterHold) {
			return vis[n-1].ID
		}
		return ""
	}
	return ""
}

func (v *reviewView) move(delta int) {
	vis := v.visible()
	if len(vis) == 0 {
		return
	}
	i := 0
	for n, it := range vis {
		if it.ID == v.selected {
			i = n
			break
		}
	}
	i += delta
	if i < 0 {
		i = 0
	}
	if i >= len(vis) {
		i = len(vis) - 1
	}
	id := vis[i].ID
	if v.selected != id {
		v.stashDraft()
		v.resetContext()
	}
	v.selected = id
	v.compare = false
	v.diffAltOnly = false
	v.loadDraft()
}

func (v *reviewView) setFilter(f int) {
	v.stashDraft()
	v.filter = f
	if !v.selectedVisible() {
		v.selected = v.firstID()
		v.compare = false
		v.diffAltOnly = false
		v.resetContext()
		v.loadDraft()
	}
}

func (v *reviewView) reload() {
	if v.session.Root == "" || v.session.PacketsDir == "" {
		return
	}
	fresh, err := LoadSession(v.session.Root, v.session.PacketsDir, v.session.DecisionsPath)
	if err != nil {
		v.status = "reload: " + err.Error()
		return
	}
	v.stashDraft()
	v.session = fresh
	if _, ok := v.current(); !ok || !v.selectedVisible() {
		v.selected = v.firstID()
		v.resetContext()
	}
	v.loadDraft()
	v.status = fmt.Sprintf("reloaded (%d items)", len(fresh.Items))
}

func (v *reviewView) resetContext() {
	v.beforeLines = defaultContextLines
	v.afterLines = defaultContextLines
}

// contextEdges reports whether the passage already shows the start and the
// end of the file. unique is false when the quote is not a single match.
func (v *reviewView) contextEdges() (atStart, atEnd, unique bool) {
	it, ok := v.current()
	if !ok {
		return false, false, false
	}
	file := v.session.bytes(it.Path)
	start, stop, ok := snippetWindow(file, v.session.Hits[it.ID], v.beforeLines, v.afterLines)
	if !ok {
		return false, false, false
	}
	return start == 0, stop == len(file), true
}

func (v *reviewView) contextSummary() (string, bool) {
	atStart, atEnd, unique := v.contextEdges()
	if !unique {
		return "", false
	}
	return contextLabel(v.beforeLines, v.afterLines, atStart, atEnd), true
}

func contextLabel(before, after int, atStart, atEnd bool) string {
	if atStart && atEnd {
		return "whole chapter"
	}
	return sideLabel(before, atStart, "above") + " · " + sideLabel(after, atEnd, "below")
}

func sideLabel(n int, edge bool, name string) string {
	if edge || n < 0 {
		return "all " + name
	}
	return fmt.Sprintf("%d %s", n, name)
}

func (v *reviewView) expandAbove() {
	atStart, _, unique := v.contextEdges()
	if !unique || atStart {
		return
	}
	v.beforeLines = moreContext(v.beforeLines)
}

func (v *reviewView) expandBelow() {
	_, atEnd, unique := v.contextEdges()
	if !unique || atEnd {
		return
	}
	v.afterLines = moreContext(v.afterLines)
}

func (v *reviewView) showWholeChapter() {
	atStart, atEnd, unique := v.contextEdges()
	if !unique || (atStart && atEnd) {
		return
	}
	v.beforeLines = -1
	v.afterLines = -1
}

func (v *reviewView) tightenContext() {
	if v.beforeLines == defaultContextLines && v.afterLines == defaultContextLines {
		return
	}
	v.resetContext()
}

func (v *reviewView) toggleCompare() { v.compare = !v.compare }

func (v *reviewView) toggleDiffTarget() { v.diffAltOnly = !v.diffAltOnly }

func (v *reviewView) selectAlt(i int) {
	it, ok := v.current()
	if !ok || i < 0 || i >= len(it.Alternates) {
		return
	}
	v.altIndex = i
}

func (v *reviewView) cycleAlt(delta int) {
	it, ok := v.current()
	if !ok || len(it.Alternates) == 0 {
		return
	}
	n := len(it.Alternates)
	idx := v.altIndex
	if idx < 0 {
		idx = 0
	}
	idx = (idx + delta%n + n) % n
	v.altIndex = idx
}

func (v *reviewView) copyQuoteToMine() {
	it, ok := v.current()
	if !ok {
		return
	}
	v.mine = it.Quote
	v.diffAltOnly = false
}

func (v *reviewView) copyAltToMine() {
	a, ok := v.alt()
	if !ok || a.Ignore {
		return
	}
	v.mine = a.Text
	v.diffAltOnly = false
}

func (v *reviewView) setMine(value string) {
	if value != v.mine {
		v.mine = value
	}
}

func (v *reviewView) setNote(value string) {
	if value != v.note {
		v.note = value
	}
}

func (v *reviewView) setMineFocused(focused bool) {
	if focused != v.mineFocused {
		v.mineFocused = focused
	}
}

func (v *reviewView) setNoteFocused(focused bool) {
	if focused != v.noteFocused {
		v.noteFocused = focused
	}
}

func (v *reviewView) canApply() bool {
	it, ok := v.current()
	if !ok {
		return false
	}
	d := v.session.Decisions.Items[it.ID]
	return CanAccept(v.session.Hits[it.ID], d.Reaffirm)
}

func (v *reviewView) acceptReady() bool {
	a, ok := v.alt()
	if !ok {
		return false
	}
	if a.Ignore {
		return true
	}
	return v.canApply()
}

func (v *reviewView) reviseReady() bool {
	return v.canApply() && strings.TrimSpace(v.mine) != ""
}

func (v *reviewView) reopenReady() bool {
	it, ok := v.current()
	if !ok {
		return false
	}
	return v.session.Decisions.Items[it.ID].Disposition != ""
}

func (v *reviewView) reaffirmReady() bool {
	it, ok := v.current()
	if !ok {
		return false
	}
	h := v.session.Hits[it.ID]
	d := v.session.Decisions.Items[it.ID]
	if d.Reaffirm {
		return false
	}
	return h.Status == StatusAmbiguous || h.Status == StatusMissing
}

func (v *reviewView) accept() {
	a, ok := v.alt()
	if !ok {
		return
	}
	if a.Ignore {
		if strings.TrimSpace(v.note) == "" {
			v.note = a.Reason
		}
		v.decide("reject", "", a.ID, "")
		return
	}
	v.decide("accept", "", a.ID, a.Text)
}

func (v *reviewView) revise()             { v.decide("revise", "", v.altID(), "") }
func (v *reviewView) reject()             { v.decide("reject", "", v.altID(), "") }
func (v *reviewView) withdraw()           { v.decide("withdraw", "", v.altID(), "") }
func (v *reviewView) hold(forWhat string) { v.decide("hold", forWhat, v.altID(), "") }

func (v *reviewView) reopen() {
	id := v.selected
	if id == "" || v.session.DecisionsPath == "" {
		return
	}
	next := DecisionFile{Version: 1, Items: map[string]Decision{}}
	for k, saved := range v.session.Decisions.Items {
		if k != id {
			next.Items[k] = saved
		}
	}
	err := SaveDecisions(v.session.DecisionsPath, next)
	if err != nil {
		v.status = err.Error()
		return
	}
	delete(v.drafts, id)
	v.session.Decisions = next
	v.loadDraft()
	v.status = "reopened " + id
}

func (v *reviewView) reaffirm() {
	prev := v.session.Decisions.Items[v.selected]
	v.save(Decision{
		Disposition:  prev.Disposition,
		Alternate:    v.altID(),
		AcceptedText: prev.AcceptedText,
		Text:         v.mine,
		Note:         v.note,
		HoldFor:      prev.HoldFor,
		Reaffirm:     true,
	})
}

func (v *reviewView) decide(disposition, holdFor, alternate, acceptedText string) {
	prev := v.session.Decisions.Items[v.selected]
	txt := v.mine
	if disposition != "revise" {
		txt = prev.Text
		if disposition == "accept" {
			txt = ""
		}
	}
	nextID := v.followingID()
	if !v.save(Decision{
		Disposition:  disposition,
		Alternate:    alternate,
		AcceptedText: acceptedText,
		Text:         txt,
		Note:         v.note,
		HoldFor:      holdFor,
		Reaffirm:     prev.Reaffirm,
	}) {
		return
	}
	if nextID != "" {
		v.choose(nextID)
		return
	}
	// The last item in All, Decided, or Stale is still in the list. An item
	// that left Open or Hold is not, and the passage should go blank.
	if !v.selectedVisible() {
		v.clearSelection()
	}
}

func (v *reviewView) save(d Decision) bool {
	id := v.selected
	if id == "" || v.session.DecisionsPath == "" {
		v.status = "no decisions path"
		return false
	}
	d = Stamp(d)
	next := DecisionFile{Version: 1, Items: map[string]Decision{}}
	for k, saved := range v.session.Decisions.Items {
		next.Items[k] = saved
	}
	next.Items[id] = d
	err := SaveDecisions(v.session.DecisionsPath, next)
	if err != nil {
		v.status = err.Error()
		return false
	}
	delete(v.drafts, id)
	v.session.Decisions = next
	v.status = "saved " + id
	return true
}

func (v *reviewView) countsLine() string {
	c := v.session.DetailedCounts()
	holdPart := fmt.Sprintf("hold %d", c.Hold)
	if c.Answered > 0 {
		holdPart = fmt.Sprintf("hold %d (%d ans)", c.Hold, c.Answered)
	}
	line := fmt.Sprintf("open %d · %s · pending %d · applied %d · stale %d", c.Open, holdPart, c.PendingApply, c.Applied, c.Stale)
	if v.status != "" {
		line += " · " + v.status
	}
	if len(v.session.ParseErrors) > 0 {
		line += " · err: " + v.session.ParseErrors[0]
	}
	return line
}

func (v *reviewView) statusLine() string {
	return v.countsLine() + " · j/k [ ] a r h w c Ctrl+R Ctrl+Enter"
}

func (v *reviewView) matchLabel(it Item, h Hit) string {
	d := v.session.Decisions.Items[it.ID]
	disp := d.Disposition
	if disp == "" {
		disp = "open"
	}
	if d.HoldFor != "" {
		disp += "/" + d.HoldFor
	}
	if v.session.Applied[it.ID] {
		disp += " · applied"
	}
	if v.session.HasAnswer(it.ID) {
		disp += " · answered"
	}
	if d.Reaffirm {
		disp += " · reaffirmed"
	}
	if AlternateDrifted(it, d) {
		disp += " · alternate changed"
	}
	rec := "no recorded offset"
	if it.ByteOffset != nil {
		rec = fmt.Sprintf("recorded %d", *it.ByteOffset)
	}
	switch h.Status {
	case StatusOffset, StatusMoved:
		return fmt.Sprintf("%s · %s · %s · found %d · line %d col %d", it.Kind, disp, rec, h.Offset, h.Line, h.Col)
	case StatusAmbiguous:
		return fmt.Sprintf("%s · %s · %s · ambiguous %s", it.Kind, disp, rec, joinInts(h.Candidates))
	default:
		return fmt.Sprintf("%s · %s · %s · missing", it.Kind, disp, rec)
	}
}

// passageDiff is the word-diff title and the text it compares to the quote.
// showIgnore is an ignore vote with no replacement.
func (v *reviewView) passageDiff(it Item) (title, target string, hasTarget, showIgnore bool, reason string) {
	alt, hasAlt := v.alt()
	return diffChoice(it, alt, hasAlt, v.mine, v.diffAltOnly)
}

// diffChoice is the word diff for one alternate and one mine-or-alt switch.
// The terminal draws a separate choice on each Proposed half.
func diffChoice(it Item, alt Alternate, hasAlt bool, mine string, diffAltOnly bool) (title, target string, hasTarget, showIgnore bool, reason string) {
	hasMine := strings.TrimSpace(mine) != ""
	useMine := hasMine && (!diffAltOnly || !hasAlt)
	title = "Word diff"
	if useMine {
		title = "Word diff (Quote -> Mine)"
		target = mine
		hasTarget = true
		return
	}
	if hasAlt && alt.Ignore {
		title = "Ignore"
		if from := SuggestionFrom(it, alt); from != "" {
			title += " · " + from
		}
		showIgnore = true
		reason = alt.Reason
		if reason == "" {
			reason = "Ignore."
		}
		return
	}
	if hasAlt {
		lbl := alt.Label
		if lbl == "" {
			lbl = alt.ID
		}
		if from := SuggestionFrom(it, alt); from != "" {
			lbl += " · " + from
		}
		title = fmt.Sprintf("Word diff (Quote -> %s)", lbl)
		target = alt.Text
		hasTarget = true
	}
	return
}

func itemBadges(session *Session, it Item) string {
	if session == nil {
		return ""
	}
	h := session.Hits[it.ID]
	d := session.Decisions.Items[it.ID]
	disp := d.Disposition
	if disp == "" {
		disp = "open"
	} else if d.HoldFor != "" {
		disp += "/" + d.HoldFor
	}
	badges := []string{string(h.Status)}
	if it.Kind != "" {
		badges = append(badges, it.Kind)
	}
	badges = append(badges, disp)
	if session.Applied[it.ID] {
		badges = append(badges, "applied")
	}
	if session.HasAnswer(it.ID) {
		badges = append(badges, "answered")
	}
	if it.Locked {
		badges = append(badges, "locked")
	}
	if it.From != "" {
		badges = append(badges, "from "+it.From)
	}
	for _, a := range it.Alternates {
		if a.Ignore {
			badges = append(badges, "ignore")
			break
		}
	}
	return strings.Join(badges, " · ")
}

func rowTitle(it Item, h Hit) string {
	base := filepath.Base(it.Path)
	if base == "" || base == "." {
		return it.ID
	}
	if h.Status == StatusOffset || h.Status == StatusMoved {
		return fmt.Sprintf("%s:%d", base, h.Line)
	}
	return base
}

func singleLinePreview(s string, maxRunes int) string {
	s = strings.Join(strings.Fields(s), " ")
	rs := []rune(s)
	if len(rs) > maxRunes {
		return string(rs[:maxRunes-1]) + "…"
	}
	return s
}

func altLabel(a Alternate) string {
	if a.Ignore {
		return "Ignore"
	}
	if a.Label != "" {
		return a.Label
	}
	return a.ID
}

func altDetailText(it Item, alt Alternate, hasAlt bool) (heading, from, shown string) {
	heading = "Alternate"
	if !hasAlt {
		return
	}
	from = SuggestionFrom(it, alt)
	if alt.Ignore {
		heading = "Ignore"
		shown = alt.Reason
		if shown == "" {
			shown = "Ignore."
		}
		return
	}
	shown = alt.Text
	if name := altLabel(alt); name != "" {
		heading = "Alternate · " + name
	}
	return
}

func filterName(f int) string {
	switch f {
	case filterOpen:
		return "Open"
	case filterAll:
		return "All"
	case filterHold:
		return "Hold"
	case filterDecided:
		return "Decided"
	case filterStale:
		return "Stale"
	default:
		return ""
	}
}

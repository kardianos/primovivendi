package main

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	frameGutter = 1
	frameInset  = 1
)

type zoneKind int

const (
	zoneFilter zoneKind = iota
	zoneItem
	zoneAct
	zoneField
	zoneTab
)

const (
	paneCurrent = iota
	paneProposed
	paneActions
	paneCount
)

const (
	actAccept       = "accept"
	actRevise       = "revise"
	actReject       = "reject"
	actWithdraw     = "withdraw"
	actHoldResearch = "hold-research"
	actHoldExplain  = "hold-explain"
	actHoldLater    = "hold-later"
	actReopen       = "reopen"
	actReaffirm     = "reaffirm"
	actCompare      = "compare"
	actDiff         = "diff"
	actAbove        = "above"
	actBelow        = "below"
	actChapter      = "chapter"
	actTight        = "tight"
	actQuote        = "quote"
	actAltMine      = "alt-to-mine"
	actAlt          = "alt"
	actMine         = "mine"
	actNote         = "note"
)

// tuiZone is one clickable cell range in screen coordinates.
// The range is recorded while the line is built, because the painted
// string is full of ANSI and cannot be searched for a label.
type tuiZone struct {
	x, y, w, h int
	kind       zoneKind
	act        string
	filter     int
	itemID     string
	alt        int
	enabled    bool
	slot       int
	pane       int
}

type tuiChip struct {
	label   string
	kind    zoneKind
	act     string
	filter  int
	itemID  string
	alt     int
	enabled bool
	hot     bool
	slot    int
	pane    int
}

// scrollBand is one wheel target in screen coordinates. Current bodies
// append to passages. Proposed bodies, including compare, append to diffs.
// Two halves that show the same view share one scroll top.
type scrollBand struct {
	y, h, n int
	slot    int
}

type tuiScreen struct {
	view     string
	zones    []tuiZone
	passages []scrollBand
	diffs    []scrollBand
}

func bandAt(bands []scrollBand, y int) (scrollBand, bool) {
	for _, b := range bands {
		if b.h > 0 && y >= b.y && y < b.y+b.h {
			return b, true
		}
	}
	return scrollBand{}, false
}

func shiftBands(bands []scrollBand, dy int) []scrollBand {
	if dy == 0 || len(bands) == 0 {
		return bands
	}
	out := make([]scrollBand, len(bands))
	for i, b := range bands {
		b.y += dy
		out[i] = b
	}
	return out
}

func (s tuiScreen) hit(x, y int) (tuiZone, bool) {
	for _, z := range s.zones {
		if z.w <= 0 || z.h <= 0 {
			continue
		}
		if x >= z.x && x < z.x+z.w && y >= z.y && y < z.y+z.h {
			return z, true
		}
	}
	return tuiZone{}, false
}

type tuiGeom struct {
	listW, detailW int
	bodyH          int
	inset          int
	contentH       int
	listInner      int
	detailInner    int
}

func (m tuiModel) geom() tuiGeom {
	listW, detailW := m.columns()
	bodyH := m.height - 2
	if bodyH < 1 {
		bodyH = 1
	}
	inset := frameInset
	if bodyH < 2*frameInset+1 || listW < 2*frameInset+1 || detailW < 2*frameInset+1 {
		inset = 0
	}
	contentH := bodyH - 2*inset
	if contentH < 1 {
		contentH = 1
	}
	listInner := listW - 2*inset
	detailInner := detailW - 2*inset
	if listInner < 1 {
		listInner = 1
	}
	if detailInner < 1 {
		detailInner = 1
	}
	return tuiGeom{
		listW: listW, detailW: detailW,
		bodyH: bodyH, inset: inset, contentH: contentH,
		listInner: listInner, detailInner: detailInner,
	}
}

// fitCells returns s at exactly n visible cells. ANSI stays intact.
// Padding is spaces. A wide line is cut, with no ellipsis, so a later
// column does not slide.
func fitCells(s string, n int) string {
	if n <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w == n {
		return s
	}
	if w < n {
		return s + strings.Repeat(" ", n-w)
	}
	cut := ansi.Truncate(s, n, "")
	w = lipgloss.Width(cut)
	for w > n && cut != "" {
		cut = ansi.Truncate(cut, w-1, "")
		w = lipgloss.Width(cut)
	}
	if w < n {
		cut += strings.Repeat(" ", n-w)
	}
	return cut
}

func chipText(label string) string {
	return "[ " + label + " ]"
}

// placeChips wraps chips onto rows of innerW. One blank cell separates
// chips. That cell is not part of either zone. A chip wider than the row
// is clipped and is not clickable.
func placeChips(chips []tuiChip, innerW int) ([]string, []tuiZone) {
	if innerW < 1 || len(chips) == 0 {
		return nil, nil
	}
	var lines []string
	var zones []tuiZone
	var row strings.Builder
	rowW := 0
	rowY := 0
	var rowZones []tuiZone
	flush := func() {
		lines = append(lines, fitCells(row.String(), innerW))
		zones = append(zones, rowZones...)
		row.Reset()
		rowW = 0
		rowZones = nil
		rowY++
	}
	for _, c := range chips {
		text := chipText(c.label)
		w := lipgloss.Width(text)
		if w > innerW {
			if rowW > 0 {
				flush()
			}
			lines = append(lines, fitCells(text, innerW))
			rowY++
			continue
		}
		if rowW > 0 && rowW+1+w > innerW {
			flush()
		}
		if rowW > 0 {
			row.WriteByte(' ')
			rowW++
		}
		start := rowW
		shown := text
		switch {
		case !c.enabled:
			shown = chipOff.Render(text)
		case c.hot:
			shown = chipHot.Render(text)
		}
		row.WriteString(shown)
		rowW += w
		rowZones = append(rowZones, tuiZone{
			x: start, y: rowY, w: w, h: 1,
			kind: c.kind, act: c.act, filter: c.filter,
			itemID: c.itemID, alt: c.alt, enabled: c.enabled,
			slot: c.slot, pane: c.pane,
		})
	}
	if rowW > 0 {
		flush()
	}
	return lines, zones
}

func clipZones(zones []tuiZone, h, w int) []tuiZone {
	out := make([]tuiZone, 0, len(zones))
	for _, z := range zones {
		if z.x < 0 || z.y < 0 || z.x >= w || z.y >= h {
			continue
		}
		if z.x+z.w > w {
			z.w = w - z.x
		}
		if z.y+z.h > h {
			z.h = h - z.y
		}
		if z.w <= 0 || z.h <= 0 {
			continue
		}
		out = append(out, z)
	}
	return out
}

func (m tuiModel) filterChips() []tuiChip {
	order := []int{filterOpen, filterHold, filterAll, filterDecided, filterStale}
	keys := []string{"1", "2", "3", "4", "5"}
	out := make([]tuiChip, len(order))
	for i, f := range order {
		out[i] = tuiChip{
			label:   keys[i] + " " + filterName(f),
			kind:    zoneFilter,
			filter:  f,
			enabled: true,
			hot:     f == m.review.filter,
		}
	}
	return out
}

func (m tuiModel) screen() tuiScreen {
	if m.width < 1 || m.height < 1 {
		return tuiScreen{view: m.title + "\n"}
	}
	g := m.geom()
	listLines, listZones := m.listContent(g.listInner, g.contentH)
	detailLines, detailZones, passages, diffs := m.detailContent(g.detailInner, g.contentH)
	listRows := frameRows(listLines, g.listW, g.bodyH, g.inset)
	detailRows := frameRows(detailLines, g.detailW, g.bodyH, g.inset)
	for i := range listZones {
		listZones[i].x += g.inset
		listZones[i].y += g.inset
	}
	detailX := g.listW + frameGutter + g.inset
	for i := range detailZones {
		detailZones[i].x += detailX
		detailZones[i].y += g.inset
	}
	var b strings.Builder
	for i := 0; i < g.bodyH; i++ {
		b.WriteString(listRows[i])
		b.WriteByte(' ')
		b.WriteString(detailRows[i])
		b.WriteByte('\n')
	}
	b.WriteString(fitCells(m.review.countsLine(), m.width))
	b.WriteByte('\n')
	b.WriteString(fitCells(tuiHelpLine, m.width))
	zones := append(listZones, detailZones...)
	return tuiScreen{
		view: b.String(), zones: zones,
		passages: shiftBands(passages, g.inset),
		diffs:    shiftBands(diffs, g.inset),
	}
}

func frameRows(content []string, paneW, paneH, inset int) []string {
	rows := make([]string, paneH)
	if paneW < 1 {
		paneW = 1
	}
	blank := strings.Repeat(" ", paneW)
	if inset <= 0 {
		for i := 0; i < paneH; i++ {
			line := ""
			if i < len(content) {
				line = content[i]
			}
			rows[i] = fitCells(line, paneW)
		}
		return rows
	}
	innerW := paneW - 2*inset
	if innerW < 1 {
		innerW = 1
	}
	for i := range rows {
		rows[i] = blank
	}
	limit := paneH - 2*inset
	for i := 0; i < limit; i++ {
		line := ""
		if i < len(content) {
			line = content[i]
		}
		rows[inset+i] = strings.Repeat(" ", inset) + fitCells(line, innerW) + strings.Repeat(" ", inset)
	}
	return rows
}

func (m tuiModel) listContent(innerW, contentH int) ([]string, []tuiZone) {
	if contentH < 1 || innerW < 1 {
		return nil, nil
	}
	var zones []tuiZone
	lines := []string{fitCells(m.title, innerW)}
	chipLines, chipZones := placeChips(m.filterChips(), innerW)
	for i := range chipZones {
		chipZones[i].y += len(lines)
	}
	lines = append(lines, chipLines...)
	zones = append(zones, chipZones...)
	if len(lines) < contentH {
		lines = append(lines, fitCells("", innerW))
	}
	vis := m.review.visible()
	if len(vis) == 0 {
		if len(lines) < contentH {
			lines = append(lines, fitCells("No items", innerW))
		}
	} else {
		top := m.listTop
		if top < 0 {
			top = 0
		}
		if top > len(vis) {
			top = len(vis)
		}
		end := top + m.listPage()
		if end > len(vis) {
			end = len(vis)
		}
		for i := top; i < end && len(lines)+3 <= contentH; i++ {
			block := m.itemLines(vis[i], innerW)
			zones = append(zones, tuiZone{
				x: 0, y: len(lines), w: innerW, h: len(block),
				kind: zoneItem, itemID: vis[i].ID, enabled: true,
			})
			for _, line := range block {
				lines = append(lines, fitCells(line, innerW))
			}
		}
	}
	if len(lines) > contentH {
		lines = lines[:contentH]
	}
	return lines, clipZones(zones, contentH, innerW)
}

func (m tuiModel) itemLines(it Item, innerW int) []string {
	h := m.review.session.Hits[it.ID]
	parts := []string{
		rowTitle(it, h),
		singleLinePreview(it.Quote, innerW),
		itemBadges(m.review.session, it),
	}
	selected := it.ID == m.review.selected
	out := make([]string, len(parts))
	for i, part := range parts {
		mark := "  "
		if selected && i == 0 {
			mark = "> "
		}
		if innerW < 2 {
			mark = ""
		}
		line := mark + part
		if selected {
			line = selectedStyle.Render(line)
		}
		out[i] = line
	}
	return out
}

func (m tuiModel) detailContent(innerW, contentH int) ([]string, []tuiZone, []scrollBand, []scrollBand) {
	if contentH < 1 || innerW < 1 {
		return nil, nil, nil, nil
	}
	topH := (contentH + 1) / 2
	botH := contentH - topH
	topLines, topZones, topPass, topDiff := m.onePane(m.topPane, 0, innerW, topH)
	botLines, botZones, botPass, botDiff := m.onePane(m.bottomPane, 1, innerW, botH)
	shift := len(topLines)
	for i := range botZones {
		botZones[i].y += shift
	}
	lines := append(topLines, botLines...)
	zones := append(topZones, botZones...)
	passages := append(append([]scrollBand{}, topPass...), shiftBands(botPass, shift)...)
	diffs := append(append([]scrollBand{}, topDiff...), shiftBands(botDiff, shift)...)
	if len(lines) > contentH {
		lines = lines[:contentH]
	}
	return lines, clipZones(zones, contentH, innerW), passages, diffs
}

// onePane is one half of the detail pane: a tab row, then the view it selects.
func (m tuiModel) onePane(which, slot, innerW, paneH int) ([]string, []tuiZone, []scrollBand, []scrollBand) {
	if paneH < 1 || innerW < 1 {
		return nil, nil, nil, nil
	}
	tabLines, tabZones := placeChips(m.tabChips(slot, which), innerW)
	if len(tabLines) >= paneH {
		if len(tabLines) > paneH {
			tabLines = tabLines[:paneH]
		}
		return tabLines, clipZones(tabZones, paneH, innerW), nil, nil
	}
	bodyH := paneH - len(tabLines)
	body, bodyZones, passages, diffs := m.paneBody(which, slot, innerW, bodyH)
	for i := range bodyZones {
		bodyZones[i].y += len(tabLines)
	}
	passages = shiftBands(passages, len(tabLines))
	diffs = shiftBands(diffs, len(tabLines))
	lines := padLines(append(tabLines, body...), paneH, innerW)
	zones := append(tabZones, bodyZones...)
	return lines, clipZones(zones, paneH, innerW), passages, diffs
}

func (m tuiModel) tabChips(slot, selected int) []tuiChip {
	labels := []string{"Current", "Proposed", "Actions"}
	out := make([]tuiChip, len(labels))
	for i, label := range labels {
		out[i] = tuiChip{
			label:   label,
			kind:    zoneTab,
			slot:    slot,
			pane:    i,
			enabled: true,
			hot:     i == selected,
		}
	}
	return out
}

func (m tuiModel) paneBody(which, slot, innerW, bodyH int) ([]string, []tuiZone, []scrollBand, []scrollBand) {
	if bodyH < 1 || innerW < 1 {
		return nil, nil, nil, nil
	}
	it, ok := m.review.current()
	if !ok {
		return []string{fitCells("No item selected", innerW)}, nil, nil, nil
	}
	switch which {
	case paneProposed:
		return m.proposedBody(it, slot, innerW, bodyH)
	case paneActions:
		lines, zones := m.actionsBody(innerW, bodyH)
		return lines, zones, nil, nil
	default:
		return m.currentBody(it, innerW, bodyH)
	}
}

func (m tuiModel) currentBody(it Item, innerW, bodyH int) ([]string, []tuiZone, []scrollBand, []scrollBand) {
	head := m.headLines(it, innerW)
	ctx, ctxZones := m.contextBlock(innerW)
	for i := range ctxZones {
		ctxZones[i].y += len(head)
	}
	fixed := append(append([]string{}, head...), ctx...)
	if len(fixed) >= bodyH {
		if len(fixed) > bodyH {
			fixed = fixed[:bodyH]
		}
		return fixed, clipZones(ctxZones, bodyH, innerW), nil, nil
	}
	pass := m.passageLines(it, innerW)
	show := bodyH - len(fixed)
	top := clampTop(m.passageTop, len(pass), show)
	lines := append(fixed, sliceFit(pass, top, show, innerW)...)
	return lines, ctxZones, []scrollBand{{y: len(fixed), h: show, n: len(pass)}}, nil
}

func (m tuiModel) proposedBody(it Item, slot, innerW, bodyH int) ([]string, []tuiZone, []scrollBand, []scrollBand) {
	optLines, optZones := m.altOptions(it, slot, innerW)
	ctrlLines, ctrlZones := m.proposeControls(slot, innerW)
	var lines []string
	var zones []tuiZone
	lines = append(lines, optLines...)
	zones = append(zones, optZones...)
	if len(lines) < bodyH {
		lines = append(lines, ruleLine(innerW))
	}
	for i := range ctrlZones {
		ctrlZones[i].y += len(lines)
	}
	lines = append(lines, ctrlLines...)
	zones = append(zones, ctrlZones...)
	if len(lines) >= bodyH {
		if len(lines) > bodyH {
			lines = lines[:bodyH]
		}
		return lines, clipZones(zones, bodyH, innerW), nil, nil
	}
	show := bodyH - len(lines)
	var src []string
	if m.review.compare {
		src = m.compareLines(it, innerW)
	} else {
		src = m.diffLines(it, innerW, m.paneAlt(slot), m.paneDiffAlt(slot))
	}
	top := clampTop(m.paneDiffTop(slot), len(src), show)
	fixed := len(lines)
	lines = append(lines, sliceFit(src, top, show, innerW)...)
	return lines, zones, nil, []scrollBand{{y: fixed, h: show, n: len(src), slot: slot}}
}

func ruleLine(innerW int) string {
	if innerW < 1 {
		return ""
	}
	return fitCells(strings.Repeat("-", innerW), innerW)
}

func (m tuiModel) altOptions(it Item, slot, innerW int) ([]string, []tuiZone) {
	if len(it.Alternates) == 0 {
		return []string{fitCells("No alternates", innerW)}, nil
	}
	alts := make([]tuiChip, len(it.Alternates))
	selected := m.paneAlt(slot)
	for i, a := range it.Alternates {
		label := altLabel(a)
		if from := SuggestionFrom(it, a); from != "" {
			label += " (" + from + ")"
		}
		alts[i] = tuiChip{
			label: label, kind: zoneAct, act: actAlt, alt: i, slot: slot,
			enabled: true, hot: i == selected,
		}
	}
	return placeChips(alts, innerW)
}

// contextBlock is the passage window. It widens the file shown on Current.
func (m tuiModel) contextBlock(innerW int) ([]string, []tuiZone) {
	label, ok := m.review.contextSummary()
	if !ok || innerW < 1 {
		return nil, nil
	}
	lines := []string{fitCells(label, innerW)}
	atStart, atEnd, unique := m.review.contextEdges()
	tight := m.review.beforeLines == defaultContextLines && m.review.afterLines == defaultContextLines
	cl, cz := placeChips([]tuiChip{
		{label: "[ above", kind: zoneAct, act: actAbove, enabled: unique && !atStart},
		{label: "] below", kind: zoneAct, act: actBelow, enabled: unique && !atEnd},
		{label: "C chapter", kind: zoneAct, act: actChapter, enabled: unique && !(atStart && atEnd)},
		{label: "t tight", kind: zoneAct, act: actTight, enabled: unique && !tight},
	}, innerW)
	for i := range cz {
		cz[i].y += len(lines)
	}
	return append(lines, cl...), cz
}

// proposeControls copy into Mine and choose which text this half's diff compares.
func (m tuiModel) proposeControls(slot, innerW int) ([]string, []tuiZone) {
	alt, hasAlt := altAt(m.currentItem(), m.paneAlt(slot))
	diffLabel := "d Diff alt"
	if m.paneDiffAlt(slot) {
		diffLabel = "d Diff mine"
	}
	lines, zones := placeChips([]tuiChip{
		{label: "q Quote -> Mine", kind: zoneAct, act: actQuote, slot: slot, enabled: true},
		{label: "y Alt -> Mine", kind: zoneAct, act: actAltMine, slot: slot, enabled: hasAlt && !alt.Ignore},
		{label: diffLabel, kind: zoneAct, act: actDiff, slot: slot, enabled: strings.TrimSpace(m.review.mine) != "" && hasAlt},
	}, innerW)
	return lines, zones
}

func (m tuiModel) currentItem() Item {
	it, _ := m.review.current()
	return it
}

// actionsBody is the decision chips plus Mine and Note. Mine takes the
// room the alternate text used to take. Note stays a short field under it.
func (m tuiModel) actionsBody(innerW, bodyH int) ([]string, []tuiZone) {
	actions, actionZones := m.actionBlock(innerW)
	if len(actions) > bodyH {
		return actions[:bodyH], clipZones(actionZones, bodyH, innerW)
	}
	wantNote := 2
	if m.height < 32 {
		wantNote = 1
	}
	mineN, noteN := splitActionFields(bodyH-len(actions), wantNote)
	var lines []string
	var zones []tuiZone
	lines = append(lines, actions...)
	zones = append(zones, actionZones...)
	fieldY := len(lines)
	fieldLines, fieldZones := m.fieldBlock(innerW, mineN, noteN)
	lines = append(lines, fieldLines...)
	for _, z := range fieldZones {
		z.y += fieldY
		zones = append(zones, z)
	}
	if len(lines) > bodyH {
		lines = lines[:bodyH]
	}
	return lines, clipZones(zones, bodyH, innerW)
}

// splitActionFields gives Mine every row it can, and keeps Note to a couple
// of lines at the bottom. room counts label rows and text rows together.
func splitActionFields(room, wantNote int) (mineN, noteN int) {
	if room < 2 {
		return 0, 0
	}
	mineN = 1
	room -= 2
	if wantNote > 0 && room >= 2 {
		noteN = 1
		room -= 2
		if wantNote > 1 && room > 1 {
			noteN++
			room--
		}
	}
	mineN += room
	return mineN, noteN
}

func (m *tuiModel) fitEditor() {
	g := m.geom()
	fieldW := g.detailInner
	if fieldW < 8 {
		fieldW = 8
	}
	wantNote := 2
	if m.height < 32 {
		wantNote = 1
	}
	rows := 0
	if _, ok := m.review.current(); ok && g.detailInner > 0 {
		lines, _ := m.actionBlock(g.detailInner)
		rows = len(lines)
	}
	topH := (g.contentH + 1) / 2
	botH := g.contentH - topH
	mineN, noteN := 1, wantNote
	consider := func(paneH int) {
		body := paneH
		if g.detailInner > 0 {
			tabs, _ := placeChips(m.tabChips(0, paneActions), g.detailInner)
			body -= len(tabs)
		}
		if body < 1 {
			return
		}
		mn, nn := splitActionFields(body-rows, wantNote)
		if mn > mineN {
			mineN = mn
		}
		if nn > noteN {
			noteN = nn
		}
	}
	shown := false
	if m.topPane == paneActions {
		consider(topH)
		shown = true
	}
	if m.bottomPane == paneActions {
		consider(botH)
		shown = true
	}
	if !shown {
		consider(botH)
	}
	if mineN < 1 {
		mineN = 1
	}
	if noteN < 1 {
		noteN = 1
	}
	m.mine.SetWidth(fieldW)
	m.mine.SetHeight(mineN)
	m.note.SetWidth(fieldW)
	m.note.SetHeight(noteN)
}

func padLines(lines []string, n, width int) []string {
	if n <= 0 {
		return nil
	}
	if len(lines) > n {
		return lines[:n]
	}
	for len(lines) < n {
		lines = append(lines, fitCells("", width))
	}
	return lines
}

func (m tuiModel) headLines(it Item, innerW int) []string {
	raw := strings.Split(m.detailHead(it), "\n")
	var out []string
	for _, line := range raw {
		for _, wrapped := range wrapPlain(line, innerW) {
			out = append(out, fitCells(wrapped, innerW))
		}
	}
	return out
}

func (m tuiModel) actionBlock(innerW int) ([]string, []tuiZone) {
	alt, hasAlt := m.review.alt()
	var lines []string
	var zones []tuiZone
	addChips := func(chips []tuiChip) {
		cl, cz := placeChips(chips, innerW)
		for i := range cz {
			cz[i].y += len(lines)
		}
		lines = append(lines, cl...)
		zones = append(zones, cz...)
	}
	accept := "a Accept"
	if hasAlt && alt.Ignore {
		accept = "a Ignore this"
	}
	addChips([]tuiChip{
		{label: accept, kind: zoneAct, act: actAccept, enabled: m.review.acceptReady()},
		{label: "u Use my wording", kind: zoneAct, act: actRevise, enabled: m.review.reviseReady()},
		{label: "r Reject", kind: zoneAct, act: actReject, enabled: true},
	})
	addChips([]tuiChip{
		{label: "e Hold explain", kind: zoneAct, act: actHoldExplain, enabled: true},
		{label: "s Hold research", kind: zoneAct, act: actHoldResearch, enabled: true},
		{label: "h Hold later", kind: zoneAct, act: actHoldLater, enabled: true},
	})
	addChips([]tuiChip{
		{label: "w Withdraw", kind: zoneAct, act: actWithdraw, enabled: true},
		{label: "o Reopen", kind: zoneAct, act: actReopen, enabled: m.review.reopenReady()},
		{label: "f Reaffirm", kind: zoneAct, act: actReaffirm, enabled: m.review.reaffirmReady()},
		{label: "c Compare", kind: zoneAct, act: actCompare, enabled: true},
	})
	return lines, zones
}

func (m tuiModel) fieldBlock(innerW, mineLines, noteLines int) ([]string, []tuiZone) {
	var lines []string
	var zones []tuiZone
	add := func(s string) {
		lines = append(lines, fitCells(s, innerW))
	}
	addField := func(label, view, act string, n int) {
		if n <= 0 {
			return
		}
		start := len(lines)
		add(label)
		for _, line := range takeLines(view, n) {
			add(line)
		}
		zones = append(zones, tuiZone{
			x: 0, y: start, w: innerW, h: len(lines) - start,
			kind: zoneField, act: act, enabled: true,
		})
	}
	mineLabel := "Mine"
	noteLabel := "Note"
	switch m.mode {
	case modeMine:
		mineLabel = "Mine (editing, Esc to leave, Ctrl+S to use it)"
	case modeNote:
		noteLabel = "Note (editing, Esc to leave)"
	}
	addField(mineLabel, m.mine.View(), actMine, mineLines)
	addField(noteLabel, m.note.View(), actNote, noteLines)
	return lines, zones
}

func takeLines(s string, n int) []string {
	if n <= 0 {
		return nil
	}
	parts := strings.Split(s, "\n")
	out := make([]string, n)
	for i := 0; i < n; i++ {
		if i < len(parts) {
			out[i] = parts[i]
		}
	}
	return out
}

type cellRun struct {
	text   string
	style  lipgloss.Style
	styled bool
}

func clampTop(top, count, window int) int {
	if top < 0 {
		return 0
	}
	maxTop := count - window
	if maxTop < 0 {
		maxTop = 0
	}
	if top > maxTop {
		return maxTop
	}
	return top
}

func wrapPlain(s string, width int) []string {
	if s == "" || width < 1 {
		return nil
	}
	return wrapRuns([]cellRun{{text: s}}, width)
}

// wrapRuns wraps styled runs on word boundaries. A newline in a run starts
// a new row. The frame still pads each returned row to the pane width later,
// so the terminal row itself does not wrap.
func wrapRuns(runs []cellRun, width int) []string {
	if width < 1 {
		return nil
	}
	var lines []string
	var b strings.Builder
	col := 0
	flush := func() {
		lines = append(lines, b.String())
		b.Reset()
		col = 0
	}
	write := func(s string, r cellRun) {
		if s == "" {
			return
		}
		if r.styled {
			b.WriteString(r.style.Render(s))
		} else {
			b.WriteString(s)
		}
		col += lipgloss.Width(s)
	}
	for _, r := range runs {
		rest := r.text
		for rest != "" {
			if rest[0] == '\n' {
				flush()
				rest = rest[1:]
				continue
			}
			chunk, next, isSpace := takeChunk(rest)
			if chunk == "" {
				break
			}
			rest = next
			if isSpace {
				if col == 0 {
					continue
				}
				if col+lipgloss.Width(chunk) > width {
					flush()
					continue
				}
				write(chunk, r)
				continue
			}
			for chunk != "" {
				room := width - col
				if room <= 0 {
					flush()
					continue
				}
				cw := lipgloss.Width(chunk)
				if cw <= room {
					write(chunk, r)
					break
				}
				if col > 0 && cw <= width {
					flush()
					continue
				}
				part := ansi.Truncate(chunk, room, "")
				if part == "" || part == chunk {
					_, sz := utf8.DecodeRuneInString(chunk)
					if sz < 1 {
						break
					}
					part = chunk[:sz]
				}
				write(part, r)
				chunk = chunk[len(part):]
				flush()
			}
		}
	}
	if col > 0 {
		flush()
	}
	return lines
}

func takeChunk(s string) (chunk, rest string, space bool) {
	if s == "" {
		return "", "", false
	}
	r, size := utf8.DecodeRuneInString(s)
	if size < 1 {
		return "", "", false
	}
	if r == '\n' {
		return "\n", s[size:], false
	}
	if unicode.IsSpace(r) {
		i := size
		for i < len(s) {
			r2, sz := utf8.DecodeRuneInString(s[i:])
			if sz < 1 || r2 == '\n' || !unicode.IsSpace(r2) {
				break
			}
			i += sz
		}
		return s[:i], s[i:], true
	}
	i := size
	for i < len(s) {
		r2, sz := utf8.DecodeRuneInString(s[i:])
		if sz < 1 || r2 == '\n' || unicode.IsSpace(r2) {
			break
		}
		i += sz
	}
	return s[:i], s[i:], false
}

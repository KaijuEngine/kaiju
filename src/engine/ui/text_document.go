/******************************************************************************/
/* text_document.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2026-present Dj Gilcrease                       */
/******************************************************************************/

package ui

import (
	"strings"
	"unicode/utf8"
)

// documentModel is the line-based text buffer behind the rewritten TextArea. It
// stores text as a slice of logical lines (no trailing newline per line) so that
// editing a single line and mapping a (line,col) cursor to a row index are both
// cheap and independent of total document size — the key to handling 100k lines.
//
// The empty document is a single empty line ([]string{""}) so the caret always
// has a valid line to sit on, and a trailing newline in the input yields a final
// empty line (matching how editors behave). Line endings are detected on setText
// and normalized to '\n' internally, then re-emitted in the original style by
// text().
//
// Every mutation goes through the single apply() "replace range" primitive
// (insert is an empty range, delete is an empty replacement, replace is both),
// which keeps undo/redo trivial: a command records the removed text, the
// inserted text, and where it happened, and undo/redo just run apply() in the
// other direction.
//
// The model holds the cursor and selection so undo can restore them and edits
// keep them coherent; cursor *movement* (word boundaries, vertical motion) lives
// in TextArea because it needs glyph metrics the model does not have.
type documentModel struct {
	lines     []string
	eol       eolStyle
	cursor    textPos
	sel       textSelection
	undoStack []docCommand
	redoStack []docCommand
	maxUndo   int
}

// textPos is a caret location: a line index and a rune column within that line.
type textPos struct {
	line int
	col  int
}

// textSelection is a directed selection from anchor to active. It is empty (no
// selection) when anchor == active.
type textSelection struct {
	anchor textPos
	active textPos
}

type eolStyle int

const (
	eolLF eolStyle = iota
	eolCRLF
	eolCR
)

// docCommand is one undoable edit: the range starting at start that held removed
// was replaced with inserted. Undo replaces inserted with removed; redo does the
// reverse. cursorBefore/selBefore restore the caret/selection on undo.
type docCommand struct {
	start        textPos
	removed      string
	inserted     string
	cursorBefore textPos
	selBefore    textSelection
	cursorAfter  textPos
}

const documentDefaultMaxUndo = 1000

func newDocumentModel() *documentModel {
	return &documentModel{
		lines:   []string{""},
		eol:     eolLF,
		maxUndo: documentDefaultMaxUndo,
	}
}

// --- content ----------------------------------------------------------------

func (m *documentModel) setText(s string) {
	m.eol = detectEOL(s)
	normalized := normalizeEOL(s)
	m.lines = strings.Split(normalized, "\n")
	if len(m.lines) == 0 {
		m.lines = []string{""}
	}
	m.cursor = textPos{}
	m.clearSelection()
	m.undoStack = m.undoStack[:0]
	m.redoStack = m.redoStack[:0]
}

func (m *documentModel) text() string {
	return strings.Join(m.lines, eolString(m.eol))
}

func (m *documentModel) lineCount() int { return len(m.lines) }

func (m *documentModel) line(i int) string {
	if i < 0 || i >= len(m.lines) {
		return ""
	}
	return m.lines[i]
}

// --- cursor / selection ------------------------------------------------------

func (m *documentModel) cursorPos() textPos { return m.cursor }

func (m *documentModel) setCursor(p textPos) {
	m.cursor = m.clampPos(p)
	m.clearSelection()
}

func (m *documentModel) hasSelection() bool { return m.sel.anchor != m.sel.active }

func (m *documentModel) setSelection(anchor, active textPos) {
	m.sel.anchor = m.clampPos(anchor)
	m.sel.active = m.clampPos(active)
	m.cursor = m.sel.active
}

func (m *documentModel) clearSelection() {
	m.sel.anchor = m.cursor
	m.sel.active = m.cursor
}

// selection returns the normalized [start,end) range and whether a selection
// exists.
func (m *documentModel) selection() (start, end textPos, ok bool) {
	if !m.hasSelection() {
		return m.cursor, m.cursor, false
	}
	a, b := m.sel.anchor, m.sel.active
	if m.posLess(b, a) {
		a, b = b, a
	}
	return a, b, true
}

func (m *documentModel) selectedText() string {
	start, end, ok := m.selection()
	if !ok {
		return ""
	}
	return m.textInRange(start, end)
}

func (m *documentModel) selectAll() {
	last := len(m.lines) - 1
	m.sel.anchor = textPos{0, 0}
	m.sel.active = textPos{last, editableTextRuneCount(m.lines[last])}
	m.cursor = m.sel.active
}

// --- edits (undoable) --------------------------------------------------------

// insert replaces the active selection (if any) with s, otherwise inserts s at
// the cursor. Newlines in s split lines as expected.
func (m *documentModel) insert(s string) bool {
	if s == "" {
		return false
	}
	cb, sb := m.cursor, m.sel
	start, end, _ := m.selection()
	removed, after := m.apply(start, end, s)
	m.cursor = after
	m.clearSelection()
	m.record(docCommand{
		start: start, removed: removed, inserted: s,
		cursorBefore: cb, selBefore: sb, cursorAfter: after,
	})
	return true
}

func (m *documentModel) deleteSelection() bool {
	start, end, ok := m.selection()
	if !ok {
		return false
	}
	cb, sb := m.cursor, m.sel
	removed, after := m.apply(start, end, "")
	m.cursor = after
	m.clearSelection()
	m.record(docCommand{
		start: start, removed: removed, inserted: "",
		cursorBefore: cb, selBefore: sb, cursorAfter: after,
	})
	return true
}

// backspace deletes the selection if present, otherwise the single position
// before the cursor (joining lines when at column 0).
func (m *documentModel) backspace() bool {
	if m.hasSelection() {
		return m.deleteSelection()
	}
	prev := m.posBefore(m.cursor)
	if prev == m.cursor {
		return false
	}
	return m.deleteRangeRecorded(prev, m.cursor)
}

// deleteForward deletes the selection if present, otherwise the single position
// after the cursor (joining lines when at end of line).
func (m *documentModel) deleteForward() bool {
	if m.hasSelection() {
		return m.deleteSelection()
	}
	next := m.posAfter(m.cursor)
	if next == m.cursor {
		return false
	}
	return m.deleteRangeRecorded(m.cursor, next)
}

// deleteRange deletes [start,end) as one undoable command (used by TextArea for
// word/line deletes after it computes the range).
func (m *documentModel) deleteRange(start, end textPos) bool {
	start = m.clampPos(start)
	end = m.clampPos(end)
	if start == end {
		return false
	}
	if m.posLess(end, start) {
		start, end = end, start
	}
	return m.deleteRangeRecorded(start, end)
}

func (m *documentModel) deleteRangeRecorded(start, end textPos) bool {
	cb, sb := m.cursor, m.sel
	removed, after := m.apply(start, end, "")
	m.cursor = after
	m.clearSelection()
	m.record(docCommand{
		start: start, removed: removed, inserted: "",
		cursorBefore: cb, selBefore: sb, cursorAfter: after,
	})
	return true
}

func (m *documentModel) undo() bool {
	if len(m.undoStack) == 0 {
		return false
	}
	cmd := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]
	m.apply(cmd.start, m.advance(cmd.start, cmd.inserted), cmd.removed)
	m.cursor = m.clampPos(cmd.cursorBefore)
	m.sel = textSelection{m.clampPos(cmd.selBefore.anchor), m.clampPos(cmd.selBefore.active)}
	m.redoStack = append(m.redoStack, cmd)
	return true
}

func (m *documentModel) redoLast() bool {
	if len(m.redoStack) == 0 {
		return false
	}
	cmd := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]
	m.apply(cmd.start, m.advance(cmd.start, cmd.removed), cmd.inserted)
	m.cursor = m.clampPos(cmd.cursorAfter)
	m.clearSelection()
	m.undoStack = append(m.undoStack, cmd)
	return true
}

// record pushes cmd onto the undo stack (coalescing per-keystroke typing and
// backspacing into the previous command) and clears the redo stack.
func (m *documentModel) record(cmd docCommand) {
	m.redoStack = m.redoStack[:0]
	if m.coalesce(cmd) {
		return
	}
	m.undoStack = append(m.undoStack, cmd)
	if m.maxUndo > 0 && len(m.undoStack) > m.maxUndo {
		m.undoStack = m.undoStack[len(m.undoStack)-m.maxUndo:]
	}
}

// coalesce merges a single-rune insert or backspace into the top command so a
// run of typing (or backspacing) is one undo step. Returns true if merged.
func (m *documentModel) coalesce(cmd docCommand) bool {
	if len(m.undoStack) == 0 {
		return false
	}
	top := &m.undoStack[len(m.undoStack)-1]
	// Typing: contiguous single-rune insertions, no newline.
	if top.removed == "" && cmd.removed == "" &&
		utf8.RuneCountInString(cmd.inserted) == 1 && cmd.inserted != "\n" &&
		!strings.HasSuffix(top.inserted, "\n") &&
		cmd.start == m.advance(top.start, top.inserted) {
		top.inserted += cmd.inserted
		top.cursorAfter = cmd.cursorAfter
		return true
	}
	// Backspacing: contiguous single-rune deletions, no newline.
	if top.inserted == "" && cmd.inserted == "" &&
		utf8.RuneCountInString(cmd.removed) == 1 && cmd.removed != "\n" &&
		!strings.Contains(top.removed, "\n") &&
		m.advance(cmd.start, cmd.removed) == top.start {
		top.start = cmd.start
		top.removed = cmd.removed + top.removed
		top.cursorAfter = cmd.cursorAfter
		return true
	}
	return false
}

// --- core splice + helpers ---------------------------------------------------

// apply replaces the (already-orderable) range [start,end) with newText and
// returns the removed text and the position just past the inserted text. It does
// NOT record undo — callers do that (or are undo/redo themselves).
func (m *documentModel) apply(start, end textPos, newText string) (string, textPos) {
	start = m.clampPos(start)
	end = m.clampPos(end)
	if m.posLess(end, start) {
		start, end = end, start
	}
	removed := m.textInRange(start, end)
	startLine := m.lines[start.line]
	endLine := m.lines[end.line]
	prefix := editableTextSlice(startLine, 0, start.col)
	suffix := editableTextSlice(endLine, end.col, editableTextRuneCount(endLine))
	merged := prefix + newText + suffix
	newLines := strings.Split(merged, "\n")

	tail := make([]string, len(m.lines)-(end.line+1))
	copy(tail, m.lines[end.line+1:])
	m.lines = append(m.lines[:start.line], newLines...)
	m.lines = append(m.lines, tail...)
	if len(m.lines) == 0 {
		m.lines = []string{""}
	}
	return removed, m.advance(start, newText)
}

// textInRange returns the text spanning the normalized range [start,end).
func (m *documentModel) textInRange(start, end textPos) string {
	if start.line == end.line {
		return editableTextSlice(m.lines[start.line], start.col, end.col)
	}
	var b strings.Builder
	first := m.lines[start.line]
	b.WriteString(editableTextSlice(first, start.col, editableTextRuneCount(first)))
	b.WriteByte('\n')
	for i := start.line + 1; i < end.line; i++ {
		b.WriteString(m.lines[i])
		b.WriteByte('\n')
	}
	b.WriteString(editableTextSlice(m.lines[end.line], 0, end.col))
	return b.String()
}

// advance returns the position reached by writing s starting at start.
func (m *documentModel) advance(start textPos, s string) textPos {
	if s == "" {
		return start
	}
	nl := strings.Count(s, "\n")
	if nl == 0 {
		return textPos{start.line, start.col + editableTextRuneCount(s)}
	}
	last := s[strings.LastIndexByte(s, '\n')+1:]
	return textPos{start.line + nl, editableTextRuneCount(last)}
}

func (m *documentModel) clampPos(p textPos) textPos {
	if p.line < 0 {
		p.line = 0
	}
	if p.line >= len(m.lines) {
		p.line = len(m.lines) - 1
	}
	cnt := editableTextRuneCount(m.lines[p.line])
	if p.col < 0 {
		p.col = 0
	}
	if p.col > cnt {
		p.col = cnt
	}
	return p
}

func (m *documentModel) posLess(a, b textPos) bool {
	if a.line != b.line {
		return a.line < b.line
	}
	return a.col < b.col
}

func (m *documentModel) posBefore(p textPos) textPos {
	p = m.clampPos(p)
	if p.col > 0 {
		return textPos{p.line, p.col - 1}
	}
	if p.line > 0 {
		return textPos{p.line - 1, editableTextRuneCount(m.lines[p.line-1])}
	}
	return p
}

func (m *documentModel) posAfter(p textPos) textPos {
	p = m.clampPos(p)
	if p.col < editableTextRuneCount(m.lines[p.line]) {
		return textPos{p.line, p.col + 1}
	}
	if p.line < len(m.lines)-1 {
		return textPos{p.line + 1, 0}
	}
	return p
}

// --- flat-offset conversions (compat for SetCursorOffset etc.) ---------------

func (m *documentModel) runeCountTotal() int {
	total := 0
	for i := range m.lines {
		total += editableTextRuneCount(m.lines[i])
	}
	return total + (len(m.lines) - 1) // newlines between lines
}

func (m *documentModel) offsetToPos(offset int) textPos {
	if offset <= 0 {
		return textPos{}
	}
	remaining := offset
	for i := range m.lines {
		cnt := editableTextRuneCount(m.lines[i])
		if remaining <= cnt {
			return textPos{i, remaining}
		}
		remaining -= cnt + 1 // +1 for the newline
		if remaining < 0 {
			// offset landed on the newline boundary -> end of this line
			return textPos{i, cnt}
		}
	}
	last := len(m.lines) - 1
	return textPos{last, editableTextRuneCount(m.lines[last])}
}

func (m *documentModel) posToOffset(p textPos) int {
	p = m.clampPos(p)
	offset := 0
	for i := 0; i < p.line; i++ {
		offset += editableTextRuneCount(m.lines[i]) + 1
	}
	return offset + p.col
}

// --- end-of-line handling ----------------------------------------------------

func detectEOL(s string) eolStyle {
	if strings.Contains(s, "\r\n") {
		return eolCRLF
	}
	if strings.Contains(s, "\r") {
		return eolCR
	}
	return eolLF
}

func normalizeEOL(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func eolString(style eolStyle) string {
	switch style {
	case eolCRLF:
		return "\r\n"
	case eolCR:
		return "\r"
	default:
		return "\n"
	}
}

/******************************************************************************/
/* text_document_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "testing"

func TestDocumentEmptyInvariant(t *testing.T) {
	m := newDocumentModel()
	assertEqualI(t, m.lineCount(), 1, "lineCount")
	assertEqualStr(t, m.line(0), "", "line(0)")
	assertEqualPos(t, m.cursorPos(), textPos{0, 0}, "cursor")
	assertTrue(t, !m.hasSelection(), "no selection")
}

func TestDocumentSetTextSplitsLines(t *testing.T) {
	m := newDocumentModel()
	m.setText("ab\ncd\nef")
	assertEqualI(t, m.lineCount(), 3, "lineCount")
	assertEqualStr(t, m.line(1), "cd", "line(1)")
	assertEqualStr(t, m.text(), "ab\ncd\nef", "text")
}

func TestDocumentTrailingNewlineYieldsFinalEmptyLine(t *testing.T) {
	m := newDocumentModel()
	m.setText("a\n")
	assertEqualI(t, m.lineCount(), 2, "lineCount")
	assertEqualStr(t, m.line(1), "", "final empty line")
	assertEqualStr(t, m.text(), "a\n", "text")
}

func TestDocumentEOLNormalizationRoundTrip(t *testing.T) {
	crlf := newDocumentModel()
	crlf.setText("a\r\nb")
	assertEqualI(t, crlf.lineCount(), 2, "crlf lineCount")
	assertEqualStr(t, crlf.text(), "a\r\nb", "CRLF re-emitted")

	cr := newDocumentModel()
	cr.setText("a\rb")
	assertEqualI(t, cr.lineCount(), 2, "cr lineCount")
	assertEqualStr(t, cr.text(), "a\rb", "CR re-emitted")
}

func TestDocumentInsertSameLine(t *testing.T) {
	m := newDocumentModel()
	m.setText("ac")
	m.setCursor(textPos{0, 1})
	m.insert("b")
	assertEqualStr(t, m.text(), "abc", "text")
	assertEqualPos(t, m.cursorPos(), textPos{0, 2}, "cursor")
}

func TestDocumentInsertNewlineSplits(t *testing.T) {
	m := newDocumentModel()
	m.setText("abcd")
	m.setCursor(textPos{0, 2})
	m.insert("\n")
	assertEqualI(t, m.lineCount(), 2, "lineCount")
	assertEqualStr(t, m.line(0), "ab", "line(0)")
	assertEqualStr(t, m.line(1), "cd", "line(1)")
	assertEqualPos(t, m.cursorPos(), textPos{1, 0}, "cursor")
}

func TestDocumentInsertMultiLine(t *testing.T) {
	m := newDocumentModel()
	m.setText("XY")
	m.setCursor(textPos{0, 1})
	m.insert("1\n2\n3")
	assertEqualStr(t, m.text(), "X1\n2\n3Y", "text")
	assertEqualPos(t, m.cursorPos(), textPos{2, 1}, "cursor")
}

func TestDocumentBackspaceWithinLine(t *testing.T) {
	m := newDocumentModel()
	m.setText("abc")
	m.setCursor(textPos{0, 2})
	assertTrue(t, m.backspace(), "backspace ok")
	assertEqualStr(t, m.text(), "ac", "text")
	assertEqualPos(t, m.cursorPos(), textPos{0, 1}, "cursor")
}

func TestDocumentBackspaceJoinsLines(t *testing.T) {
	m := newDocumentModel()
	m.setText("ab\ncd")
	m.setCursor(textPos{1, 0})
	assertTrue(t, m.backspace(), "backspace ok")
	assertEqualStr(t, m.text(), "abcd", "text")
	assertEqualPos(t, m.cursorPos(), textPos{0, 2}, "cursor")
}

func TestDocumentBackspaceAtStartNoOp(t *testing.T) {
	m := newDocumentModel()
	m.setText("abc")
	m.setCursor(textPos{0, 0})
	assertTrue(t, !m.backspace(), "backspace at start is a no-op")
	assertEqualStr(t, m.text(), "abc", "text unchanged")
}

func TestDocumentDeleteForwardJoinsLines(t *testing.T) {
	m := newDocumentModel()
	m.setText("ab\ncd")
	m.setCursor(textPos{0, 2})
	assertTrue(t, m.deleteForward(), "deleteForward ok")
	assertEqualStr(t, m.text(), "abcd", "text")
	assertEqualPos(t, m.cursorPos(), textPos{0, 2}, "cursor")
}

func TestDocumentSelectionDeleteAndReplace(t *testing.T) {
	m := newDocumentModel()
	m.setText("hello world")
	m.setSelection(textPos{0, 0}, textPos{0, 5})
	assertTrue(t, m.hasSelection(), "has selection")
	assertEqualStr(t, m.selectedText(), "hello", "selectedText")

	// Insert over selection replaces it.
	m.insert("hi")
	assertEqualStr(t, m.text(), "hi world", "text")
	assertEqualPos(t, m.cursorPos(), textPos{0, 2}, "cursor")
	assertTrue(t, !m.hasSelection(), "selection cleared")
}

func TestDocumentDeleteRangeMultiLine(t *testing.T) {
	m := newDocumentModel()
	m.setText("one\ntwo\nthree")
	assertTrue(t, m.deleteRange(textPos{0, 1}, textPos{2, 2}), "deleteRange ok")
	assertEqualStr(t, m.text(), "oree", "text")
	assertEqualPos(t, m.cursorPos(), textPos{0, 1}, "cursor")
}

func TestDocumentSelectAll(t *testing.T) {
	m := newDocumentModel()
	m.setText("ab\ncd")
	m.selectAll()
	start, end, ok := m.selection()
	assertTrue(t, ok, "has selection")
	assertEqualPos(t, start, textPos{0, 0}, "start")
	assertEqualPos(t, end, textPos{1, 2}, "end")
	assertEqualStr(t, m.selectedText(), "ab\ncd", "selectedText")
}

func TestDocumentOffsetConversions(t *testing.T) {
	m := newDocumentModel()
	m.setText("ab\ncd")
	assertEqualI(t, m.runeCountTotal(), 5, "runeCountTotal")
	cases := []struct {
		off int
		pos textPos
	}{
		{0, textPos{0, 0}},
		{2, textPos{0, 2}},
		{3, textPos{1, 0}},
		{5, textPos{1, 2}},
	}
	for _, c := range cases {
		assertEqualPos(t, m.offsetToPos(c.off), c.pos, "offsetToPos")
		assertEqualI(t, m.posToOffset(c.pos), c.off, "posToOffset")
	}
}

func TestDocumentMultiByteRunes(t *testing.T) {
	m := newDocumentModel()
	m.setText("a界b")
	m.setCursor(textPos{0, 2})
	assertTrue(t, m.backspace(), "backspace ok")
	assertEqualStr(t, m.text(), "ab", "text")
	assertEqualPos(t, m.cursorPos(), textPos{0, 1}, "cursor")
}

func TestDocumentUndoRedoSingleEdit(t *testing.T) {
	m := newDocumentModel()
	m.setText("ac")
	m.setCursor(textPos{0, 1})
	m.insert("\n") // split into "a" / "c"
	assertEqualStr(t, m.text(), "a\nc", "after insert")

	assertTrue(t, m.undo(), "undo ok")
	assertEqualStr(t, m.text(), "ac", "after undo")
	assertEqualPos(t, m.cursorPos(), textPos{0, 1}, "undo restores cursor")

	assertTrue(t, m.redoLast(), "redo ok")
	assertEqualStr(t, m.text(), "a\nc", "after redo")
	assertEqualPos(t, m.cursorPos(), textPos{1, 0}, "redo restores post-edit cursor")
}

func TestDocumentTypingCoalescesToSingleUndo(t *testing.T) {
	m := newDocumentModel()
	m.insert("a")
	m.insert("b")
	m.insert("c")
	assertEqualStr(t, m.text(), "abc", "after typing")

	assertTrue(t, m.undo(), "undo ok")
	assertEqualStr(t, m.text(), "", "a run of typing undoes in one step")
	assertEqualPos(t, m.cursorPos(), textPos{0, 0}, "cursor")

	assertTrue(t, m.redoLast(), "redo ok")
	assertEqualStr(t, m.text(), "abc", "after redo")
	assertEqualPos(t, m.cursorPos(), textPos{0, 3}, "cursor")
}

func TestDocumentBackspaceCoalescesToSingleUndo(t *testing.T) {
	m := newDocumentModel()
	m.setText("abc")
	m.setCursor(textPos{0, 3})
	m.backspace()
	m.backspace()
	m.backspace()
	assertEqualStr(t, m.text(), "", "after backspaces")

	assertTrue(t, m.undo(), "undo ok")
	assertEqualStr(t, m.text(), "abc", "a run of backspaces undoes in one step")
	assertEqualPos(t, m.cursorPos(), textPos{0, 3}, "cursor")
}

func TestDocumentNewEditClearsRedo(t *testing.T) {
	m := newDocumentModel()
	m.insert("a")
	m.insert("b")
	assertTrue(t, m.undo(), "undo ok")
	assertEqualStr(t, m.text(), "", "after undo")
	// A fresh edit after undo discards the redo stack.
	m.insert("z")
	assertTrue(t, !m.redoLast(), "redo discarded after new edit")
	assertEqualStr(t, m.text(), "z", "text")
}

func TestDocumentNewlineIsOwnUndoStep(t *testing.T) {
	m := newDocumentModel()
	m.insert("a")
	m.insert("\n")
	m.insert("b")
	assertEqualStr(t, m.text(), "a\nb", "after edits")
	// "b" undo, then the newline undo, then "a" undo — newline is not coalesced.
	assertTrue(t, m.undo(), "undo 1")
	assertEqualStr(t, m.text(), "a\n", "after undo 1")
	assertTrue(t, m.undo(), "undo 2")
	assertEqualStr(t, m.text(), "a", "after undo 2")
	assertTrue(t, m.undo(), "undo 3")
	assertEqualStr(t, m.text(), "", "after undo 3")
}

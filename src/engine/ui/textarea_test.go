package ui

import (
	"testing"

	"kaijuengine.com/matrix"
)

func assertCaret(t *testing.T, got textareaCaretGeometry, line int, x, y, height float32) {
	t.Helper()
	if got.line != line || got.x != x || got.y != y || got.height != height {
		t.Fatalf("caret = {line:%d x:%v y:%v height:%v}, want {line:%d x:%v y:%v height:%v}",
			got.line, got.x, got.y, got.height, line, x, y, height)
	}
}

func assertVec4s(t *testing.T, got, want []matrix.Vec4) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("rect[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestTextAreaCaretGeometryShortLine(t *testing.T) {
	t.Parallel()

	rects := []matrix.Vec4{
		{0, 0, 10, 20},
		{10, 0, 10, 20},
		{20, 0, 10, 20},
	}
	assertCaret(t, textareaCaretFromRuneRects("abc", rects, 0, 20), 0, 0, 0, 20)
	assertCaret(t, textareaCaretFromRuneRects("abc", rects, 2, 20), 0, 20, 0, 20)
	assertCaret(t, textareaCaretFromRuneRects("abc", rects, 3, 20), 0, 30, 0, 20)
}

func TestTextAreaCaretGeometryWrappedLines(t *testing.T) {
	t.Parallel()

	rects := []matrix.Vec4{
		{0, 0, 10, 20},
		{10, 0, 10, 20},
		{0, 20, 10, 20},
		{10, 20, 10, 20},
	}
	assertCaret(t, textareaCaretFromRuneRects("abcd", rects, 2, 20), 1, 0, 20, 20)
	assertCaret(t, textareaCaretFromRuneRects("abcd", rects, 4, 20), 1, 20, 20, 20)
}

func TestTextAreaCaretGeometryExplicitNewlines(t *testing.T) {
	t.Parallel()

	rects := []matrix.Vec4{
		{0, 0, 10, 20},
		{10, 0, 10, 20},
		{20, 0, 0, 20},
		{0, 20, 10, 20},
		{10, 20, 10, 20},
	}
	assertCaret(t, textareaCaretFromRuneRects("ab\ncd", rects, 2, 20), 0, 20, 0, 20)
	assertCaret(t, textareaCaretFromRuneRects("ab\ncd", rects, 3, 20), 1, 0, 20, 20)
	assertCaret(t, textareaCaretFromRuneRects("ab\n", rects[:3], 3, 20), 1, 0, 20, 20)
}

func TestTextAreaPointerOffsetFromPoint(t *testing.T) {
	t.Parallel()

	rects := []matrix.Vec4{
		{0, 0, 10, 20},
		{10, 0, 10, 20},
		{20, 0, 0, 20},
		{0, 20, 10, 20},
		{10, 20, 10, 20},
	}
	text := "ab\ncd"
	tests := []struct {
		name  string
		point matrix.Vec2
		want  int
	}{
		{name: "before first rune", point: matrix.Vec2{1, 10}, want: 0},
		{name: "between first and second rune", point: matrix.Vec2{6, 10}, want: 1},
		{name: "newline", point: matrix.Vec2{22, 10}, want: 2},
		{name: "wrapped next line", point: matrix.Vec2{12, 25}, want: 4},
		{name: "after line", point: matrix.Vec2{200, 25}, want: 5},
		{name: "after text", point: matrix.Vec2{200, 200}, want: 5},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := textareaRuneOffsetFromPoint(text, rects, tt.point); got != tt.want {
				t.Fatalf("textareaRuneOffsetFromPoint(%v) = %d, want %d",
					tt.point, got, tt.want)
			}
		})
	}
}

func TestTextAreaTextMutationOffsets(t *testing.T) {
	t.Parallel()

	text, cursor, changed := textareaInsertTextAt("one two", 3, 0, 0, "\n")
	if !changed || text != "one\n two" || cursor != 4 {
		t.Fatalf("textareaInsertTextAt(newline) = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, changed, "one\n two", 4, true)
	}

	text, cursor, changed = textareaInsertTextAt("aβb", 3, 1, 2, "Δ")
	if !changed || text != "aΔb" || cursor != 2 {
		t.Fatalf("textareaInsertTextAt(selection) = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, changed, "aΔb", 2, true)
	}

	text, cursor, changed = textareaBackspaceText("ab\ncd", 3, 0, 0)
	if !changed || text != "abcd" || cursor != 2 {
		t.Fatalf("textareaBackspaceText(line boundary) = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, changed, "abcd", 2, true)
	}

	text, cursor, changed = textareaDeleteText("ab\ncd", 2, 0, 0)
	if !changed || text != "abcd" || cursor != 2 {
		t.Fatalf("textareaDeleteText(line boundary) = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, changed, "abcd", 2, true)
	}

	text, cursor, changed = textareaBackspaceText("ab\ncd", 5, 1, 4)
	if !changed || text != "ad" || cursor != 1 {
		t.Fatalf("textareaBackspaceText(selection) = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, changed, "ad", 1, true)
	}
}

func TestTextAreaMovementOffsets(t *testing.T) {
	t.Parallel()

	text := "ab\ncde\nfg"
	rects := []matrix.Vec4{
		{0, 0, 10, 20},
		{10, 0, 10, 20},
		{20, 0, 0, 20},
		{0, 20, 10, 20},
		{10, 20, 10, 20},
		{20, 20, 10, 20},
		{30, 20, 0, 20},
		{0, 40, 10, 20},
		{10, 40, 10, 20},
	}

	if got := textareaMoveVerticalOffset(text, rects, 5, -1, 20, 20); got != 2 {
		t.Fatalf("textareaMoveVerticalOffset(up) = %d, want %d", got, 2)
	}
	if got := textareaMoveVerticalOffset(text, rects, 2, 1, 20, 20); got != 5 {
		t.Fatalf("textareaMoveVerticalOffset(down) = %d, want %d", got, 5)
	}
	if got := textareaLineStartOffset(text, rects, 5, 20); got != 3 {
		t.Fatalf("textareaLineStartOffset() = %d, want %d", got, 3)
	}
	if got := textareaLineEndOffset(text, rects, 5, 20); got != 6 {
		t.Fatalf("textareaLineEndOffset() = %d, want %d", got, 6)
	}
}

func TestTextAreaSelectedTextUsesLogicalRuneRanges(t *testing.T) {
	t.Parallel()

	if got := textareaSelectedText("abcd", 1, 3); got != "bc" {
		t.Fatalf("textareaSelectedText(wrapped logical range) = %q, want %q", got, "bc")
	}
	if got := textareaSelectedText("ab\ncd", 1, 4); got != "b\nc" {
		t.Fatalf("textareaSelectedText(newline range) = %q, want %q", got, "b\nc")
	}
	if got := textareaSelectedText("a\u03b2\n\u0394d", 1, 4); got != "\u03b2\n\u0394" {
		t.Fatalf("textareaSelectedText(multibyte range) = %q, want %q", got, "\u03b2\n\u0394")
	}
}

func TestTextAreaSelectionPanelRectsWrappedLines(t *testing.T) {
	t.Parallel()

	rects := []matrix.Vec4{
		{0, 0, 10, 20},
		{10, 0, 10, 20},
		{0, 20, 10, 20},
		{10, 20, 10, 20},
	}
	got := textareaSelectionPanelRects("abcd", rects, 1, 3, 40, 20)
	want := []matrix.Vec4{
		{10, 0, 10, 20},
		{0, 20, 10, 20},
	}
	assertVec4s(t, got, want)
}

func TestTextAreaSelectionPanelRectsExplicitNewlines(t *testing.T) {
	t.Parallel()

	rects := []matrix.Vec4{
		{0, 0, 10, 20},
		{10, 0, 10, 20},
		{20, 0, 0, 20},
		{0, 20, 10, 20},
		{10, 20, 10, 20},
	}
	got := textareaSelectionPanelRects("ab\ncd", rects, 1, 4, 40, 20)
	want := []matrix.Vec4{
		{10, 0, 10, 20},
		{0, 20, 10, 20},
	}
	assertVec4s(t, got, want)

	got = textareaSelectionPanelRects("ab\ncd", rects, 2, 3, 40, 20)
	want = []matrix.Vec4{
		{20, 0, 0.001, 20},
	}
	assertVec4s(t, got, want)
}

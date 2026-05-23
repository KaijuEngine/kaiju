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

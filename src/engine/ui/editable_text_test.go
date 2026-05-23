/******************************************************************************/
/* editable_text_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "testing"

func TestEditableTextRuneOffsetOperations(t *testing.T) {
	t.Parallel()

	if got := editableTextRuneCount("a界\nb"); got != 4 {
		t.Fatalf("editableTextRuneCount() = %d, want %d", got, 4)
	}
	if got := editableTextClampOffset("a界b", -10); got != 0 {
		t.Fatalf("editableTextClampOffset(low) = %d, want %d", got, 0)
	}
	if got := editableTextClampOffset("a界b", 10); got != 3 {
		t.Fatalf("editableTextClampOffset(high) = %d, want %d", got, 3)
	}
	if got := editableTextSlice("a界b", 1, 2); got != "界" {
		t.Fatalf("editableTextSlice(unicode) = %q, want %q", got, "界")
	}
	if got := editableTextSlice("a界b", 3, 1); got != "界b" {
		t.Fatalf("editableTextSlice(reversed) = %q, want %q", got, "界b")
	}
	if got := editableTextInsert("ab", 1, "界\n"); got != "a界\nb" {
		t.Fatalf("editableTextInsert() = %q, want %q", got, "a界\nb")
	}
}

func TestEditableTextDeleteOperations(t *testing.T) {
	t.Parallel()

	text, cursor, ok := editableTextDeleteRange("a界\nb", 1, 3)
	if !ok || text != "ab" || cursor != 1 {
		t.Fatalf("editableTextDeleteRange() = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, ok, "ab", 1, true)
	}

	text, cursor, ok = editableTextDeleteBefore("a界b", 2)
	if !ok || text != "ab" || cursor != 1 {
		t.Fatalf("editableTextDeleteBefore() = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, ok, "ab", 1, true)
	}

	text, cursor, ok = editableTextDeleteAfter("a界b", 1)
	if !ok || text != "ab" || cursor != 1 {
		t.Fatalf("editableTextDeleteAfter() = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, ok, "ab", 1, true)
	}

	text, cursor, ok = editableTextDeleteBefore("a界b", 0)
	if ok || text != "a界b" || cursor != 0 {
		t.Fatalf("editableTextDeleteBefore(start) = (%q, %d, %v), want (%q, %d, %v)",
			text, cursor, ok, "a界b", 0, false)
	}
}

func TestEditableTextNormalizeSelection(t *testing.T) {
	t.Parallel()

	start, end := editableTextNormalizeSelection("a界b", 9, -2)
	if start != 0 || end != 3 {
		t.Fatalf("editableTextNormalizeSelection() = (%d, %d), want (%d, %d)",
			start, end, 0, 3)
	}
}

func TestEditableTextWordBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		text  string
		start int
		dir   int
		want  int
	}{
		{name: "ascii left", text: "hello world", start: 10, dir: -1, want: 6},
		{name: "ascii right", text: "hello world", start: 1, dir: 1, want: 5},
		{name: "left skips whitespace", text: "hello   ", start: 7, dir: -1, want: 0},
		{name: "right skips whitespace", text: "hello \t world", start: 6, dir: 1, want: 13},
		{name: "newline boundary", text: "one\ntwo three", start: 4, dir: 1, want: 7},
		{name: "unicode left", text: "αβ γδ", start: 4, dir: -1, want: 3},
		{name: "unicode right", text: "αβ γδ", start: 1, dir: 1, want: 2},
		{name: "clamps past end", text: "abc", start: 99, dir: 1, want: 3},
		{name: "clamps before start", text: "abc", start: -1, dir: -1, want: 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := editableTextWordBoundary(tt.text, tt.start, tt.dir); got != tt.want {
				t.Fatalf("editableTextWordBoundary(%q, %d, %d) = %d, want %d",
					tt.text, tt.start, tt.dir, got, tt.want)
			}
		})
	}
}

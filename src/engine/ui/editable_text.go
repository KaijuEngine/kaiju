/******************************************************************************/
/* editable_text.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"unicode"
	"unicode/utf8"
)

func editableTextRuneCount(text string) int {
	return utf8.RuneCountInString(text)
}

func editableTextClamp(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func editableTextClampOffset(text string, offset int) int {
	return editableTextClamp(offset, 0, editableTextRuneCount(text))
}

func editableTextNormalizeSelection(text string, start, end int) (int, int) {
	count := editableTextRuneCount(text)
	start = editableTextClamp(start, 0, count)
	end = editableTextClamp(end, 0, count)
	if end < start {
		start, end = end, start
	}
	return start, end
}

func editableTextByteOffset(text string, runeOffset int) int {
	if runeOffset <= 0 {
		return 0
	}
	i := 0
	for byteOffset := range text {
		if i == runeOffset {
			return byteOffset
		}
		i++
	}
	return len(text)
}

func editableTextSlice(text string, start, end int) string {
	start, end = editableTextNormalizeSelection(text, start, end)
	return text[editableTextByteOffset(text, start):editableTextByteOffset(text, end)]
}

func editableTextInsert(text string, offset int, insert string) string {
	byteOffset := editableTextByteOffset(text, editableTextClampOffset(text, offset))
	return text[:byteOffset] + insert + text[byteOffset:]
}

func editableTextDeleteRange(text string, start, end int) (string, int, bool) {
	start, end = editableTextNormalizeSelection(text, start, end)
	if start == end {
		return text, start, false
	}
	startByte := editableTextByteOffset(text, start)
	endByte := editableTextByteOffset(text, end)
	return text[:startByte] + text[endByte:], start, true
}

func editableTextDeleteBefore(text string, offset int) (string, int, bool) {
	offset = editableTextClampOffset(text, offset)
	if offset == 0 {
		return text, offset, false
	}
	return editableTextDeleteRange(text, offset-1, offset)
}

func editableTextDeleteAfter(text string, offset int) (string, int, bool) {
	offset = editableTextClampOffset(text, offset)
	if offset == editableTextRuneCount(text) {
		return text, offset, false
	}
	return editableTextDeleteRange(text, offset, offset+1)
}

func editableTextWordBoundary(text string, start, dir int) int {
	runes := []rune(text)
	count := len(runes)
	if start < 0 {
		return 0
	}
	if start > count {
		return count
	}
	if count == 0 || dir == 0 {
		return editableTextClamp(start, 0, count)
	}

	i := start
	if dir < 0 {
		if i >= count {
			i = count - 1
		}
		for i > 0 && unicode.IsSpace(runes[i]) {
			i += dir
		}
		for i > 0 && !unicode.IsSpace(runes[i]) {
			i += dir
		}
		if i < count && unicode.IsSpace(runes[i]) {
			i++
		}
		return editableTextClamp(i, 0, count)
	}

	if i > 0 && i-1 < count && unicode.IsSpace(runes[i-1]) {
		for i < count && unicode.IsSpace(runes[i]) {
			i += dir
		}
	}
	for i > 0 && i < count && !unicode.IsSpace(runes[i]) {
		i += dir
	}
	return editableTextClamp(i, 0, count)
}

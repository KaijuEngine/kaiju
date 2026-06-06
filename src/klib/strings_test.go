/******************************************************************************/
/* strings_test.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"testing"
)

func TestStringValueCompare(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		// Integers
		{a: "2", b: "10", want: -1},
		{a: "10", b: "2", want: 1},
		{a: "10", b: "10", want: 0},
		{a: "-1", b: "1", want: -1},
		{a: "1", b: "-1", want: 1},
		{a: "-1", b: "-1", want: 0},
		// Floats
		{a: "2.0", b: "10.0", want: -1},
		{a: "10.0", b: "2.0", want: 1},
		{a: "1.0", b: "1.0", want: 0},
		{a: "-1.0", b: "1.0", want: -1},
		{a: "1.0", b: "-1.0", want: 1},
		{a: "-1.0", b: "-1.0", want: 0},
		// Mixed Integer/Float
		{a: "1", b: "1.01", want: -1},
		{a: "1.01", b: "1", want: 1},
		{a: "-1", b: "-1.0", want: 0},
		{a: "1", b: "-1.0", want: 1},
		{a: "-1", b: "1.0", want: -1},
		{a: "-1.0", b: "1", want: -1},
		{a: "-1.0", b: "-1", want: 0},
		// Strings
		{a: "apple", b: "banana", want: -1},
		{a: "banana", b: "apple", want: 1},
		{a: "apple", b: "apple", want: 0},
	}
	for _, test := range tests {
		got := StringValueCompare(test.a, test.b)
		if got != test.want {
			t.Errorf("StringValueCompare(%q, %q) = %d, want %d", test.a, test.b, got, test.want)
			t.Fail()
		}
	}
}

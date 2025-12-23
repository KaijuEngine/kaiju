/******************************************************************************/
/* strings_test.go                                                            */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
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

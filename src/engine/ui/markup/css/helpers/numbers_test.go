/******************************************************************************/
/* numbers_test.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package helpers

import (
	"math"
	"testing"

	"kaijuengine.com/rendering"
)

type testWindow struct {
	dpmm   float64
	width  int
	height int
}

func (w testWindow) DotsPerMillimeter() float64 { return w.dpmm }
func (w testWindow) Width() int                 { return w.width }
func (w testWindow) Height() int                { return w.height }

func TestNumFromLengthWithFont_Units(t *testing.T) {
	w := testWindow{dpmm: 2, width: 1000, height: 500}
	fontSize := float32(20)

	tests := []struct {
		name string
		in   string
		want float32
	}{
		{name: "percent", in: "75%", want: 0.75},
		{name: "px", in: "100px", want: 100},
		{name: "em", in: "3em", want: 60},
		{name: "ex", in: "6ex", want: 120},
		{name: "cm", in: "4cm", want: 80},
		{name: "mm", in: "40mm", want: 80},
		{name: "in", in: "1.5in", want: 76.2},
		{name: "pt", in: "72pt", want: 50.8},
		{name: "pc", in: "6pc", want: 50.8},
		{name: "rem", in: "2rem", want: 2 * rendering.DefaultFontEMSize},
		{name: "vw", in: "30vw", want: 300},
		{name: "vh", in: "10vh", want: 50},
		{name: "vmin", in: "20vmin", want: 100},
		{name: "vmax", in: "20vmax", want: 200},
		{name: "ch", in: "20ch", want: 200}, // 20 * (0.5 * 20)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NumFromLengthWithFont(tt.in, w, fontSize)
			if got != tt.want {
				t.Fatalf("NumFromLengthWithFont(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNumFromLength_DefaultFontContext(t *testing.T) {
	w := testWindow{dpmm: 2, width: 1000, height: 500}

	got := NumFromLength("2em", w)
	want := float32(2) * rendering.DefaultFontEMSize
	if got != want {
		t.Fatalf("NumFromLength(%q) = %v, want %v", "2em", got, want)
	}
}

func TestNumFromLengthWithFont_LeadingDecimalUnits(t *testing.T) {
	w := testWindow{dpmm: 2, width: 1000, height: 500}
	fontSize := float32(20)

	tests := []struct {
		name string
		in   string
		want float32
	}{
		{name: "vw", in: ".2vw", want: 2},
		{name: "rem", in: ".2rem", want: 0.2 * rendering.DefaultFontEMSize},
		{name: "ch", in: ".2ch", want: 2}, // 0.2 * (0.5 * 20)
		{name: "cm", in: ".2cm", want: 4}, // 0.2 * 10mm * dpmm(2)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NumFromLengthWithFont(tt.in, w, fontSize)
			if got != tt.want {
				t.Fatalf("NumFromLengthWithFont(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNumFromLengthWithFont_GarbageOrIllFormattedInput(t *testing.T) {
	w := testWindow{dpmm: 2, width: 1000, height: 500}
	fontSize := float32(20)

	tests := []struct {
		name string
		in   string
		want float32
	}{
		{name: "empty string", in: "", want: 0},
		{name: "only unit", in: "px", want: 0},
		{name: "garbage text", in: "hello", want: 0},
		{name: "unknown unit", in: "12abc", want: 0},
		{name: "double percent", in: "50%%", want: 0.5},
		{name: "space separated number unit", in: "10 px", want: 10},
		{name: "mixed alpha and numeric", in: "1o0px", want: 1},
		{name: "sign only", in: "-", want: 0},
		{name: "dot only", in: ".", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("NumFromLengthWithFont(%q) panicked: %v", tt.in, r)
				}
			}()
			got := NumFromLengthWithFont(tt.in, w, fontSize)
			if got != tt.want {
				t.Fatalf("NumFromLengthWithFont(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNumFromLengthWithFont_NonFiniteInput(t *testing.T) {
	w := testWindow{dpmm: 2, width: 1000, height: 500}
	fontSize := float32(20)

	tests := []string{
		"NaNpx",
		"+Infpx",
		"-Infpx",
	}

	for _, in := range tests {
		t.Run(in, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("NumFromLengthWithFont(%q) panicked: %v", in, r)
				}
			}()
			got := NumFromLengthWithFont(in, w, fontSize)
			if math.IsNaN(float64(got)) || math.IsInf(float64(got), 0) {
				t.Fatalf("NumFromLengthWithFont(%q) produced non-finite result: %v", in, got)
			}
		})
	}
}

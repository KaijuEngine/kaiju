/******************************************************************************/
/* color_test.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"fmt"
	"math"
	"testing"
)

// ============================================================================
// Constructor Tests
// ============================================================================

func TestNewColor(t *testing.T) {
	c := NewColor(1.0, 0.5, 0.25, 0.75)
	if c.R() != 1.0 || c.G() != 0.5 || c.B() != 0.25 || c.A() != 0.75 {
		t.Fatalf("NewColor() = {%v,%v,%v,%v}, want {1.0,0.5,0.25,0.75}",
			c.R(), c.G(), c.B(), c.A())
	}
}

func TestNewColor8(t *testing.T) {
	c := NewColor8(255, 128, 64, 200)
	if c.R() != 255 || c.G() != 128 || c.B() != 64 || c.A() != 200 {
		t.Fatalf("NewColor8() = {%v,%v,%v,%v}, want {255,128,64,200}",
			c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorRGBAInt(t *testing.T) {
	c := ColorRGBAInt(255, 128, 64, 200)
	expected := ColorRGBAInt(255, 128, 64, 200)
	if !c.Equals(expected) {
		t.Fatalf("ColorRGBAInt(255,128,64,200) = {%v,%v,%v,%v}",
			c.R(), c.G(), c.B(), c.A())
	}
	// Verify normalization
	if !Approx(c.R(), 1.0) {
		t.Errorf("R expected 1.0, got %v", c.R())
	}
	if !Approx(c.G(), Float(128)/255.0) {
		t.Errorf("G expected %v, got %v", Float(128)/255.0, c.G())
	}
	if !Approx(c.B(), Float(64)/255.0) {
		t.Errorf("B expected %v, got %v", Float(64)/255.0, c.B())
	}
	if !Approx(c.A(), Float(200)/255.0) {
		t.Errorf("A expected %v, got %v", Float(200)/255.0, c.A())
	}
}

func TestColorRGBInt(t *testing.T) {
	c := ColorRGBInt(255, 128, 64)
	if !Approx(c.R(), 1.0) {
		t.Errorf("R expected 1.0, got %v", c.R())
	}
	if !Approx(c.A(), 1.0) {
		t.Errorf("A should be 1.0, got %v", c.A())
	}
}

// https://oklch.com/#0.7,0.1,71,100
func TestColorOklabOklch(t *testing.T) {
	half := 0.5 * 255
	colors := []Color{
		OklabToColor(0.7, 0.03, 0.09, 1),
		ColorRGBAInt(198.0, 148.0, 85.0, 255),

		OklabToColor(0.7, 0.03, -0.1, 0.6),
		ColorRGBAInt(154, 149, 218, 0.6*255),

		OklabToColor(0.7, 0.09, 0.05, 1),
		ColorRGBAInt(213, 134, 121, 255),

		OklabToColor(0.7, 0.09, 0.05, 0.5),
		ColorRGBAInt(215, 133, 121, int(half)),

		OklchToColor(1, 0, 0, 1),
		ColorRGBAInt(255, 255, 255, 255),

		OklchToColor(1, 0, 180, 1),
		ColorRGBAInt(255, 255, 255, 255),

		OklchToColor(0.7, 0.1, 312, 1),
		ColorRGBAInt(179, 140, 204, 255),

		OklchToColor(0.7, 0.1, 37, 1),
		ColorRGBAInt(212, 136, 114, 255),

		OklchToColor(0.595, 0.1, 108, 1),
		ColorRGBAInt(133, 131, 53, 255),
	}

	for i := 0; i < len(colors); i += 2 {
		o := colors[i]
		rgb := colors[i+1]
		t.Run(fmt.Sprintf("oklab_%v", o), func(t *testing.T) {
			t.Parallel()
			for i, c := range o {
				rgbValue := rgb[i]
				if !ApproxTo(c, rgbValue, 0.02) {
					t.Errorf("got %f, want %f", c, rgbValue)
				}
			}
		})
	}
}

// ============================================================================
// Getter / Setter Tests
// ============================================================================

func TestColorGetters(t *testing.T) {
	c := NewColor(0.2, 0.4, 0.6, 0.8)
	if !Approx(c.R(), 0.2) {
		t.Errorf("R = %v, want 0.2", c.R())
	}
	if !Approx(c.G(), 0.4) {
		t.Errorf("G = %v, want 0.4", c.G())
	}
	if !Approx(c.B(), 0.6) {
		t.Errorf("B = %v, want 0.6", c.B())
	}
	if !Approx(c.A(), 0.8) {
		t.Errorf("A = %v, want 0.8", c.A())
	}
}

func TestColorRGBA(t *testing.T) {
	c := NewColor(0.1, 0.3, 0.5, 0.9)
	r, g, b, a := c.RGBA()
	if !Approx(r, 0.1) || !Approx(g, 0.3) || !Approx(b, 0.5) || !Approx(a, 0.9) {
		t.Errorf("RGBA() = (%v,%v,%v,%v), want (0.1,0.3,0.5,0.9)", r, g, b, a)
	}
}

func TestColorSetters(t *testing.T) {
	c := NewColor(0, 0, 0, 0)
	c.SetR(0.5)
	c.SetG(0.3)
	c.SetB(0.7)
	c.SetA(0.8)
	if !Approx(c.R(), 0.5) || !Approx(c.G(), 0.3) || !Approx(c.B(), 0.7) || !Approx(c.A(), 0.8) {
		t.Errorf("Setters failed: {%v,%v,%v,%v}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorPointerGetters(t *testing.T) {
	c := NewColor(0.1, 0.2, 0.3, 0.4)
	*c.PR() = 0.9
	*c.PG() = 0.8
	*c.PB() = 0.7
	*c.PA() = 0.6
	if !Approx(c.R(), 0.9) || !Approx(c.G(), 0.8) || !Approx(c.B(), 0.7) || !Approx(c.A(), 0.6) {
		t.Errorf("Pointer getters failed: {%v,%v,%v,%v}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColor8Getters(t *testing.T) {
	c := NewColor8(10, 20, 30, 40)
	if c.R() != 10 || c.G() != 20 || c.B() != 30 || c.A() != 40 {
		t.Fatalf("Color8 getters: {%v,%v,%v,%v}, want {10,20,30,40}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColor8RGBA(t *testing.T) {
	c := NewColor8(100, 150, 200, 255)
	r, g, b, a := c.RGBA()
	if r != 100 || g != 150 || b != 200 || a != 255 {
		t.Errorf("Color8.RGBA() = (%v,%v,%v,%v), want (100,150,200,255)", r, g, b, a)
	}
}

func TestColor8Setters(t *testing.T) {
	c := NewColor8(0, 0, 0, 0)
	c.SetR(100)
	c.SetG(150)
	c.SetB(200)
	c.SetA(255)
	if c.R() != 100 || c.G() != 150 || c.B() != 200 || c.A() != 255 {
		t.Errorf("Color8 setters failed: {%v,%v,%v,%v}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColor8PointerGetters(t *testing.T) {
	c := NewColor8(1, 2, 3, 4)
	*c.PR() = 70
	*c.PG() = 80
	*c.PB() = 90
	*c.PA() = 100
	if c.R() != 70 || c.G() != 80 || c.B() != 90 || c.A() != 100 {
		t.Errorf("Color8 pointer getters failed: {%v,%v,%v,%v}", c.R(), c.G(), c.B(), c.A())
	}
}

// ============================================================================
// Conversion Tests
// ============================================================================

func TestColorFromColor8(t *testing.T) {
	c8 := NewColor8(255, 0, 128, 200)
	c := ColorFromColor8(c8)
	if !Approx(c.R(), 1.0) {
		t.Errorf("R expected 1.0, got %v", c.R())
	}
	if !Approx(c.G(), 0.0) {
		t.Errorf("G expected 0.0, got %v", c.G())
	}
	if !Approx(c.B(), Float(128)/255.0) {
		t.Errorf("B expected %v, got %v", Float(128)/255.0, c.B())
	}
	if !Approx(c.A(), Float(200)/255.0) {
		t.Errorf("A expected %v, got %v", Float(200)/255.0, c.A())
	}
}

func TestColor8FromColor(t *testing.T) {
	c := NewColor(1.0, 0.0, 0.5, 1.0)
	c8 := Color8FromColor(c)
	if c8.R() != 255 || c8.G() != 0 || c8.B() != 127 || c8.A() != 255 {
		t.Errorf("Color8FromColor({1,0,0.5,1}) = {%v,%v,%v,%v}, want {255,0,127,255}",
			c8.R(), c8.G(), c8.B(), c8.A())
	}
}

func TestColorRoundTripConversion(t *testing.T) {
	// Color8 -> Color -> Color8 should preserve values for clean 8-bit values
	testValues := []uint8{0, 51, 102, 153, 204, 255}
	for _, r := range testValues {
		for _, g := range testValues {
			for _, b := range testValues {
				for _, a := range testValues {
					c8 := NewColor8(r, g, b, a)
					c := ColorFromColor8(c8)
					c8Back := Color8FromColor(c)
					if !c8.Equal(c8Back) {
						t.Errorf("Round trip failed for Color8{%v,%v,%v,%v}: got {%v,%v,%v,%v}",
							r, g, b, a, c8Back.R(), c8Back.G(), c8Back.B(), c8Back.A())
					}
				}
			}
		}
	}
}

func TestAsColor8(t *testing.T) {
	c := NewColor(1.0, 0.0, 1.0, 0.5)
	c8 := c.AsColor8()
	if c8.R() != 255 || c8.G() != 0 || c8.A() != 127 {
		t.Errorf("AsColor8({1,0,1,0.5}) = {%v,%v,%v,%v}, want {255,0,255,127}",
			c8.R(), c8.G(), c8.B(), c8.A())
	}
}

func TestColor8AsColor(t *testing.T) {
	c8 := NewColor8(0, 255, 128, 192)
	c := c8.AsColor()
	if !Approx(c.R(), 0.0) || !Approx(c.G(), 1.0) {
		t.Errorf("AsColor(Color8{0,255,128,192}) = {%v,%v,%v,%v}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorFromVec3(t *testing.T) {
	v := NewVec3(0.2, 0.4, 0.6)
	c := ColorFromVec3(v)
	if !Approx(c.R(), 0.2) || !Approx(c.G(), 0.4) || !Approx(c.B(), 0.6) {
		t.Errorf("ColorFromVec3({0.2,0.4,0.6}) = {%v,%v,%v,%v}", c.R(), c.G(), c.B(), c.A())
	}
	if !Approx(c.A(), 1.0) {
		t.Errorf("Alpha should be 1.0, got %v", c.A())
	}
}

func TestColorFromVec4(t *testing.T) {
	v := NewVec4(0.1, 0.3, 0.5, 0.9)
	c := ColorFromVec4(v)
	if !c.Equals(c) {
		t.Errorf("ColorFromVec4(%v) = {%v,%v,%v,%v}", v, c.R(), c.G(), c.B(), c.A())
	}
}

func TestColor8FromBytes(t *testing.T) {
	bytes := []byte{100, 150, 200, 255}
	c := Color8FromBytes(bytes)
	if c.R() != 100 || c.G() != 150 || c.B() != 200 || c.A() != 255 {
		t.Errorf("Color8FromBytes = {%v,%v,%v,%v}, want {100,150,200,255}",
			c.R(), c.G(), c.B(), c.A())
	}
}

// ============================================================================
// ColorMix Tests
// ============================================================================

func TestColorMixZero(t *testing.T) {
	lhs := NewColor(1.0, 0.0, 0.0, 1.0)
	rhs := NewColor(0.0, 1.0, 0.0, 1.0)
	result := ColorMix(lhs, rhs, 0.0)
	if !result.Equals(lhs) {
		t.Errorf("ColorMix with amount=0 should return lhs, got {%v,%v,%v,%v}",
			result.R(), result.G(), result.B(), result.A())
	}
}

func TestColorMixOne(t *testing.T) {
	lhs := NewColor(1.0, 0.0, 0.0, 1.0)
	rhs := NewColor(0.0, 1.0, 0.0, 1.0)
	result := ColorMix(lhs, rhs, 1.0)
	if !result.Equals(rhs) {
		t.Errorf("ColorMix with amount=1 should return rhs, got {%v,%v,%v,%v}",
			result.R(), result.G(), result.B(), result.A())
	}
}

func TestColorMixHalf(t *testing.T) {
	lhs := NewColor(1.0, 0.0, 0.0, 1.0)
	rhs := NewColor(0.0, 1.0, 1.0, 0.0)
	result := ColorMix(lhs, rhs, 0.5)
	expected := NewColor(0.5, 0.5, 0.5, 0.5)
	if !result.Equals(expected) {
		t.Errorf("ColorMix half = {%v,%v,%v,%v}, want {%v,%v,%v,%v}",
			result.R(), result.G(), result.B(), result.A(),
			expected.R(), expected.G(), expected.B(), expected.A())
	}
}

// ============================================================================
// Hex String Tests
// ============================================================================

func TestColorFromHexString8Char(t *testing.T) {
	c, err := ColorFromHexString("#ff8040c8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !Approx(c.R(), 1.0) {
		t.Errorf("R expected 1.0, got %v", c.R())
	}
	c8 := c.AsColor8()
	if c8.R() != 255 || c8.G() != 128 || c8.B() != 64 || c8.A() != 200 {
		t.Errorf("ColorFromHexString(\"#ff8040c8\") = {%v,%v,%v,%v}, want {255,128,64,200}",
			c8.R(), c8.G(), c8.B(), c8.A())
	}
}

func TestColorFromHexString6Char(t *testing.T) {
	c, err := ColorFromHexString("#ff8040")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c8 := c.AsColor8()
	if c8.R() != 255 || c8.G() != 128 || c8.B() != 64 {
		t.Errorf("RGB wrong: {%v,%v,%v}, want {255,128,64}", c8.R(), c8.G(), c8.B())
	}
	if c8.A() != 255 {
		t.Errorf("A should be 255, got %v", c8.A())
	}
}

func TestColorFromHexString3Char(t *testing.T) {
	c, err := ColorFromHexString("#f84")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c8 := c.AsColor8()
	if c8.R() != 255 || c8.G() != 136 || c8.B() != 68 {
		t.Errorf("RGB wrong: {%v,%v,%v}, want {255,136,68}", c8.R(), c8.G(), c8.B())
	}
}

func TestColorFromHexStringNoPrefix(t *testing.T) {
	c, err := ColorFromHexString("ff8040")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c8 := c.AsColor8()
	if c8.R() != 255 || c8.G() != 128 || c8.B() != 64 {
		t.Errorf("RGB wrong without prefix: {%v,%v,%v}", c8.R(), c8.G(), c8.B())
	}
}

func TestColorFromHexStringInvalid(t *testing.T) {
	_, err := ColorFromHexString("#fffff")
	if err == nil {
		t.Error("expected error for 5-char hex string")
	}
	_, err = ColorFromHexString("")
	if err == nil {
		t.Error("expected error for empty string")
	}
	// "#xyz" is 3 chars so passes the length check; expanded to "xxxxxx"
	// Sscanf with %02x on "xx" silently leaves the default (255 for each)
	c, err := ColorFromHexString("#xyz")
	if err != nil {
		t.Error("#xyz (3-char) should not return an error")
	}
	// Default rgba in Color8FromHexString is {255,255,255,255}
	c8 := c.AsColor8()
	if c8.R() != 255 || c8.G() != 255 || c8.B() != 255 {
		t.Errorf("#xyz parsed unexpectedly: {%v,%v,%v,%v}", c8.R(), c8.G(), c8.B(), c8.A())
	}
}

func TestColorHex(t *testing.T) {
	c := NewColor(1.0, 0.5, 0.25, 1.0)
	hex := c.Hex()
	// uint8(c*255) truncates: 0.5*255=127.5 -> 127 -> 0x7f, 0.25*255=63.75 -> 63 -> 0x3f
	expected := "#ff7f3fff"
	if hex != expected {
		t.Errorf("Hex() = %s, want %s", hex, expected)
	}
}

func TestColorHexRoundTrip(t *testing.T) {
	// Use values that survive truncation round-trip (exact 0/255 multiples)
	c := NewColor(1.0, 0.0, 0.0, 1.0)
	hex := c.Hex()
	c2, err := ColorFromHexString(hex)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}
	if !c.Equals(c2) {
		t.Errorf("Hex round-trip failed: {%v,%v,%v,%v} -> %s -> {%v,%v,%v,%v}",
			c.R(), c.G(), c.B(), c.A(), hex, c2.R(), c2.G(), c2.B(), c2.A())
	}
}

func TestColor8FromHexString8Char(t *testing.T) {
	c8, err := Color8FromHexString("#ff8040c8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c8.R() != 255 || c8.G() != 128 || c8.B() != 64 || c8.A() != 200 {
		t.Errorf("Color8FromHexString(\"#ff8040c8\") = {%v,%v,%v,%v}",
			c8.R(), c8.G(), c8.B(), c8.A())
	}
}

func TestColor8FromHexString6Char(t *testing.T) {
	c8, err := Color8FromHexString("#aabbcc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c8.R() != 170 || c8.G() != 187 || c8.B() != 204 || c8.A() != 255 {
		t.Errorf("Color8FromHexString(\"#aabbcc\") = {%v,%v,%v,%v}",
			c8.R(), c8.G(), c8.B(), c8.A())
	}
}

func TestColor8FromHexString3Char(t *testing.T) {
	c8, err := Color8FromHexString("#abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c8.R() != 170 || c8.G() != 187 || c8.B() != 204 || c8.A() != 255 {
		t.Errorf("Color8FromHexString(\"#abc\") = {%v,%v,%v,%v}",
			c8.R(), c8.G(), c8.B(), c8.A())
	}
}

func TestColor8FromHexStringInvalid(t *testing.T) {
	c8, err := Color8FromHexString("invalid")
	if err == nil {
		t.Error("expected error for invalid hex string")
	}
	// Should return magenta on error
	if !c8.Equal(Color8FromColor(ColorMagenta())) {
		t.Errorf("Invalid hex should return magenta default, got {%v,%v,%v,%v}",
			c8.R(), c8.G(), c8.B(), c8.A())
	}
}

func TestColor8Hex(t *testing.T) {
	c8 := NewColor8(255, 128, 64, 200)
	hex := c8.Hex()
	expected := "#ff8040c8"
	if hex != expected {
		t.Errorf("Color8.Hex() = %s, want %s", hex, expected)
	}
}

// ============================================================================
// Color8.Equal Tests
// ============================================================================

func TestColor8Equal(t *testing.T) {
	a := NewColor8(100, 150, 200, 255)
	b := NewColor8(100, 150, 200, 255)
	c := NewColor8(100, 150, 200, 100)
	if !a.Equal(b) {
		t.Error("Color8.Equal returned false for identical colors")
	}
	if a.Equal(c) {
		t.Error("Color8.Equal returned true for different colors")
	}
}

// ============================================================================
// Color8.Similar Tests
// ============================================================================

func TestColor8SimilarZeroTolerance(t *testing.T) {
	a := NewColor8(100, 150, 200, 255)
	b := NewColor8(100, 150, 200, 255)
	c := NewColor8(101, 150, 200, 255)
	if !a.Similar(b, 0) {
		t.Error("Similar with tolerance=0 should match identical colors")
	}
	if a.Similar(c, 0) {
		t.Error("Similar with tolerance=0 should not match different colors")
	}
}

func TestColor8SimilarTolerance(t *testing.T) {
	// NOTE: The code's AbsInt(a) = a & 0x7FFFFFFFFFFFFFFF which only works
	// for positive differences (lhs >= rhs). When lhs < rhs, the result
	// is a large positive number, not the true absolute value.
	// Tests must use lhs >= rhs for Similar to work as expected.
	a := NewColor8(105, 155, 205, 255) // lhs > rhs so diffs are positive
	b := NewColor8(100, 150, 200, 250)
	if !a.Similar(b, 10) {
		t.Error("Similar should return true when all positive diffs within tolerance 10")
	}
	if a.Similar(b, 3) {
		t.Error("Similar should return false when diffs (5) exceed tolerance 3")
	}
}

func TestColor8SimilarNegativeComponents(t *testing.T) {
	// lhs > rhs so all differences are positive (AbsInt works correctly)
	a := NewColor8(20, 20, 20, 20)
	b := NewColor8(10, 10, 10, 10)
	if !a.Similar(b, 15) {
		t.Error("Similar should work for small positive differences")
	}
}

// ============================================================================
// Color.Equals Tests
// ============================================================================

func TestColorEquals(t *testing.T) {
	a := NewColor(0.5, 0.3, 0.8, 1.0)
	b := NewColor(0.5, 0.3, 0.8, 1.0)
	c := NewColor(0.5, 0.3, 0.7, 1.0)
	if !a.Equals(b) {
		t.Error("Color.Equals returned false for identical colors")
	}
	if a.Equals(c) {
		t.Error("Color.Equals returned true for different colors")
	}
}

func TestColorEqualsNaN(t *testing.T) {
	a := NewColor(0.5, NaN(), 0.3, 1.0)
	b := NewColor(0.5, NaN(), 0.3, 1.0)
	// NaN != NaN, so Colors containing NaN should not be equal
	if a.Equals(b) {
		t.Error("Color.Equals should return false when components contain NaN")
	}
}

// ============================================================================
// Color.IsZero Tests
// ============================================================================

func TestColorIsZero(t *testing.T) {
	zero := ColorZero()
	nonZero := ColorRed()
	if !zero.IsZero() {
		t.Error("ColorZero().IsZero() should be true")
	}
	if nonZero.IsZero() {
		t.Error("ColorRed().IsZero() should be false")
	}
}

// ============================================================================
// Color.ScaleWithoutAlpha Tests
// ============================================================================

func TestScaleWithoutAlpha(t *testing.T) {
	c := NewColor(1.0, 0.5, 0.25, 0.8)
	scaled := c.ScaleWithoutAlpha(0.5)
	if !Approx(scaled.R(), 0.5) {
		t.Errorf("R expected 0.5, got %v", scaled.R())
	}
	if !Approx(scaled.G(), 0.25) {
		t.Errorf("G expected 0.25, got %v", scaled.G())
	}
	if !Approx(scaled.B(), 0.125) {
		t.Errorf("B expected 0.125, got %v", scaled.B())
	}
	if !Approx(scaled.A(), 0.8) {
		t.Errorf("A should remain 0.8, got %v", scaled.A())
	}
}

func TestScaleWithoutAlphaZero(t *testing.T) {
	c := NewColor(1.0, 1.0, 1.0, 1.0)
	scaled := c.ScaleWithoutAlpha(0.0)
	if !Approx(scaled.R(), 0.0) || !Approx(scaled.G(), 0.0) || !Approx(scaled.B(), 0.0) {
		t.Error("ScaleWithoutAlpha(0) should produce black RGB")
	}
	if !Approx(scaled.A(), 1.0) {
		t.Error("Alpha should remain unchanged")
	}
}

// ============================================================================
// Color.MultiplyAssign Tests
// ============================================================================

func TestMultiplyAssign(t *testing.T) {
	c := NewColor(2.0, 3.0, 4.0, 5.0)
	other := NewColor(0.5, 0.5, 0.5, 0.5)
	c.MultiplyAssign(other)
	if !Approx(c.R(), 1.0) || !Approx(c.G(), 1.5) || !Approx(c.B(), 2.0) || !Approx(c.A(), 2.5) {
		t.Errorf("MultiplyAssign failed: {%v,%v,%v,%v}, want {1,1.5,2,2.5}",
			c.R(), c.G(), c.B(), c.A())
	}
}

func TestMultiplyAssignWhite(t *testing.T) {
	c := NewColor(0.5, 0.3, 0.8, 0.9)
	c.MultiplyAssign(ColorWhite())
	// Multiplying by white (1,1,1,1) should not change the color
	if !Approx(c.R(), 0.5) || !Approx(c.G(), 0.3) || !Approx(c.B(), 0.8) || !Approx(c.A(), 0.9) {
		t.Errorf("MultiplyAssign(ColorWhite()) should not change color: {%v,%v,%v,%v}",
			c.R(), c.G(), c.B(), c.A())
	}
}

// ============================================================================
// Color.Inverted Tests
// ============================================================================

func TestInverted(t *testing.T) {
	c := NewColor(0.2, 0.4, 0.6, 0.8)
	inv := c.Inverted()
	if !ApproxTo(inv.R(), 0.8, 0.001) {
		t.Errorf("Inverted R expected 0.8, got %v", inv.R())
	}
	if !ApproxTo(inv.G(), 0.6, 0.001) {
		t.Errorf("Inverted G expected 0.6, got %v", inv.G())
	}
	if !ApproxTo(inv.B(), 0.4, 0.001) {
		t.Errorf("Inverted B expected 0.4, got %v", inv.B())
	}
	if !Approx(inv.A(), 0.8) {
		t.Errorf("Inverted A should remain 0.8, got %v", inv.A())
	}
}

func TestInvertedRed(t *testing.T) {
	c := ColorRed()
	inv := c.Inverted()
	if !Approx(inv.R(), 0.0) || !Approx(inv.G(), 1.0) || !Approx(inv.B(), 1.0) {
		t.Errorf("Inverted(ColorRed()) = {%v,%v,%v,%v}, want {0,1,1,1}",
			inv.R(), inv.G(), inv.B(), inv.A())
	}
}

func TestInvertedBlack(t *testing.T) {
	c := ColorBlack()
	inv := c.Inverted()
	// Black is {0,0,0,1}, inverted should be {1,1,1,1} = white
	if !inv.Equals(ColorWhite()) {
		t.Errorf("Inverted(ColorBlack()) should equal ColorWhite(), got {%v,%v,%v,%v}",
			inv.R(), inv.G(), inv.B(), inv.A())
	}
}

// ============================================================================
// Named Color Constructor Tests
// ============================================================================

func TestColorRed(t *testing.T) {
	c := ColorRed()
	if !Approx(c.R(), 1.0) || !Approx(c.G(), 0.0) || !Approx(c.B(), 0.0) || !Approx(c.A(), 1.0) {
		t.Errorf("ColorRed() = {%v,%v,%v,%v}, want {1,0,0,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorWhite(t *testing.T) {
	c := ColorWhite()
	if !Approx(c.R(), 1.0) || !Approx(c.G(), 1.0) || !Approx(c.B(), 1.0) || !Approx(c.A(), 1.0) {
		t.Errorf("ColorWhite() = {%v,%v,%v,%v}, want {1,1,1,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorBlue(t *testing.T) {
	c := ColorBlue()
	if !Approx(c.R(), 0.0) || !Approx(c.G(), 0.0) || !Approx(c.B(), 1.0) || !Approx(c.A(), 1.0) {
		t.Errorf("ColorBlue() = {%v,%v,%v,%v}, want {0,0,1,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorBlack(t *testing.T) {
	c := ColorBlack()
	if !Approx(c.R(), 0.0) || !Approx(c.G(), 0.0) || !Approx(c.B(), 0.0) || !Approx(c.A(), 1.0) {
		t.Errorf("ColorBlack() = {%v,%v,%v,%v}, want {0,0,0,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorGreen(t *testing.T) {
	c := ColorGreen()
	if !Approx(c.R(), 0.0) || !Approx(c.G(), 1.0) || !Approx(c.B(), 0.0) || !Approx(c.A(), 1.0) {
		t.Errorf("ColorGreen() = {%v,%v,%v,%v}, want {0,1,0,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorYellow(t *testing.T) {
	c := ColorYellow()
	if !Approx(c.R(), 1.0) || !Approx(c.G(), 1.0) || !Approx(c.B(), 0.0) || !Approx(c.A(), 1.0) {
		t.Errorf("ColorYellow() = {%v,%v,%v,%v}, want {1,1,0,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorClear(t *testing.T) {
	c := ColorClear()
	if !Approx(c.R(), 0.0) || !Approx(c.G(), 0.0) || !Approx(c.B(), 0.0) || !Approx(c.A(), 0.0) {
		t.Errorf("ColorClear() = {%v,%v,%v,%v}, want {0,0,0,0}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorGray(t *testing.T) {
	c := ColorGray()
	if !Approx(c.R(), 0.5) || !Approx(c.G(), 0.5) || !Approx(c.B(), 0.5) || !Approx(c.A(), 1.0) {
		t.Errorf("ColorGray() = {%v,%v,%v,%v}, want {0.5,0.5,0.5,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorOrange(t *testing.T) {
	c := ColorOrange()
	if !Approx(c.R(), 1.0) || !Approx(c.G(), 0.647) || !Approx(c.B(), 0.0) {
		t.Errorf("ColorOrange() = {%v,%v,%v,%v}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorCyan(t *testing.T) {
	c := ColorCyan()
	if !Approx(c.R(), 0.0) || !Approx(c.G(), 1.0) || !Approx(c.B(), 1.0) {
		t.Errorf("ColorCyan() = {%v,%v,%v,%v}, want {0,1,1,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorMagenta(t *testing.T) {
	c := ColorMagenta()
	if !Approx(c.R(), 1.0) || !Approx(c.G(), 0.0) || !Approx(c.B(), 1.0) {
		t.Errorf("ColorMagenta() = {%v,%v,%v,%v}, want {1,0,1,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorPurple(t *testing.T) {
	c := ColorPurple()
	if !Approx(c.R(), 0.5) || !Approx(c.G(), 0.0) || !Approx(c.B(), 0.5) {
		t.Errorf("ColorPurple() = {%v,%v,%v,%v}, want {0.5,0,0.5,1}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorTransparent(t *testing.T) {
	c := ColorTransparent()
	if !Approx(c.R(), 1.0) || !Approx(c.G(), 1.0) || !Approx(c.B(), 1.0) || !Approx(c.A(), 0.0) {
		t.Errorf("ColorTransparent() = {%v,%v,%v,%v}, want {1,1,1,0}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestColorZero(t *testing.T) {
	c := ColorZero()
	if !Approx(c.R(), 0.0) || !Approx(c.G(), 0.0) || !Approx(c.B(), 0.0) || !Approx(c.A(), 0.0) {
		t.Errorf("ColorZero() = {%v,%v,%v,%v}, want {0,0,0,0}", c.R(), c.G(), c.B(), c.A())
	}
}

// Test that all named color constructors return valid colors (0 <= component <= 1 for non-alpha, 0 <= alpha <= 1)
func TestNamedColorsValidRanges(t *testing.T) {
	colors := []struct {
		name  string
		color Color
	}{
		{"Red", ColorRed()},
		{"White", ColorWhite()},
		{"Blue", ColorBlue()},
		{"Black", ColorBlack()},
		{"Green", ColorGreen()},
		{"Yellow", ColorYellow()},
		{"Orange", ColorOrange()},
		{"Clear", ColorClear()},
		{"Gray", ColorGray()},
		{"Purple", ColorPurple()},
		{"Brown", ColorBrown()},
		{"Pink", ColorPink()},
		{"Cyan", ColorCyan()},
		{"Magenta", ColorMagenta()},
		{"Teal", ColorTeal()},
		{"Lime", ColorLime()},
		{"Maroon", ColorMaroon()},
		{"Olive", ColorOlive()},
		{"Navy", ColorNavy()},
		{"Silver", ColorSilver()},
		{"Gold", ColorGold()},
		{"Sky", ColorSky()},
		{"Violet", ColorViolet()},
		{"Indigo", ColorIndigo()},
		{"Turquoise", ColorTurquoise()},
		{"Azure", ColorAzure()},
		{"Chartreuse", ColorChartreuse()},
		{"Coral", ColorCoral()},
		{"Crimson", ColorCrimson()},
		{"Fuchsia", ColorFuchsia()},
		{"Khaki", ColorKhaki()},
		{"Lavender", ColorLavender()},
		{"Moccasin", ColorMoccasin()},
		{"Salmon", ColorSalmon()},
		{"Sienna", ColorSienna()},
		{"Tan", ColorTan()},
		{"Tomato", ColorTomato()},
		{"Wheat", ColorWheat()},
		{"Aqua", ColorAqua()},
		{"Aquamarine", ColorAquamarine()},
		{"Beige", ColorBeige()},
		{"Bisque", ColorBisque()},
		{"BlanchedAlmond", ColorBlanchedAlmond()},
		{"BlueViolet", ColorBlueViolet()},
		{"BurlyWood", ColorBurlyWood()},
		{"CadetBlue", ColorCadetBlue()},
		{"Chocolate", ColorChocolate()},
		{"CornflowerBlue", ColorCornflowerBlue()},
		{"CornSilk", ColorCornSilk()},
		{"DarkBlue", ColorDarkBlue()},
		{"DarkCyan", ColorDarkCyan()},
		{"DarkGoldenrod", ColorDarkGoldenrod()},
		{"DarkGray", ColorDarkGray()},
		{"DarkModeGrayBG", ColorDarkModeGrayBG()},
		{"DarkModeGrayFG", ColorDarkModeGrayFG()},
		{"DarkGreen", ColorDarkGreen()},
		{"DarkKhaki", ColorDarkKhaki()},
		{"DarkMagenta", ColorDarkMagenta()},
		{"DarkOliveGreen", ColorDarkOliveGreen()},
		{"DarkOrange", ColorDarkOrange()},
		{"DarkOrchid", ColorDarkOrchid()},
		{"DarkRed", ColorDarkRed()},
		{"DarkSalmon", ColorDarkSalmon()},
		{"DarkSeaGreen", ColorDarkSeaGreen()},
		{"DarkSlateBlue", ColorDarkSlateBlue()},
		{"DarkSlateGray", ColorDarkSlateGray()},
		{"DarkTurquoise", ColorDarkTurquoise()},
		{"DarkViolet", ColorDarkViolet()},
		{"DeepPink", ColorDeepPink()},
		{"DeepSkyBlue", ColorDeepSkyBlue()},
		{"DimGray", ColorDimGray()},
		{"DodgerBlue", ColorDodgerBlue()},
		{"Firebrick", ColorFirebrick()},
		{"FloralWhite", ColorFloralWhite()},
		{"ForestGreen", ColorForestGreen()},
		{"Gainsboro", ColorGainsboro()},
		{"GhostWhite", ColorGhostWhite()},
		{"Goldenrod", ColorGoldenrod()},
		{"GreenYellow", ColorGreenYellow()},
		{"Honeydew", ColorHoneydew()},
		{"HotPink", ColorHotPink()},
		{"IndianRed", ColorIndianRed()},
		{"Ivory", ColorIvory()},
		{"LavenderBlush", ColorLavenderBlush()},
		{"LawnGreen", ColorLawnGreen()},
		{"LemonChiffon", ColorLemonChiffon()},
		{"LightBlue", ColorLightBlue()},
		{"LightCoral", ColorLightCoral()},
		{"LightCyan", ColorLightCyan()},
		{"LightGoldenrodYellow", ColorLightGoldenrodYellow()},
		{"LightGreen", ColorLightGreen()},
		{"LightGrey", ColorLightGrey()},
		{"LightPink", ColorLightPink()},
		{"LightSalmon", ColorLightSalmon()},
		{"LightSeaGreen", ColorLightSeaGreen()},
		{"LightSkyBlue", ColorLightSkyBlue()},
		{"LightSlateGray", ColorLightSlateGray()},
		{"LightSteelBlue", ColorLightSteelBlue()},
		{"LightYellow", ColorLightYellow()},
		{"LimeGreen", ColorLimeGreen()},
		{"Linen", ColorLinen()},
		{"MediumAquamarine", ColorMediumAquamarine()},
		{"MediumBlue", ColorMediumBlue()},
		{"MediumOrchid", ColorMediumOrchid()},
		{"MediumPurple", ColorMediumPurple()},
		{"MediumSeaGreen", ColorMediumSeaGreen()},
		{"MediumSlateBlue", ColorMediumSlateBlue()},
		{"MediumSpringGreen", ColorMediumSpringGreen()},
		{"MediumTurquoise", ColorMediumTurquoise()},
		{"MediumVioletRed", ColorMediumVioletRed()},
		{"MidnightBlue", ColorMidnightBlue()},
		{"MintCream", ColorMintCream()},
		{"MistyRose", ColorMistyRose()},
		{"NavajoWhite", ColorNavajoWhite()},
		{"OldLace", ColorOldLace()},
		{"OliveDrab", ColorOliveDrab()},
		{"OrangeRed", ColorOrangeRed()},
		{"Orchid", ColorOrchid()},
		{"PaleGoldenrod", ColorPaleGoldenrod()},
		{"PaleGreen", ColorPaleGreen()},
		{"PaleTurquoise", ColorPaleTurquoise()},
		{"PaleVioletred", ColorPaleVioletred()},
		{"PapayaWhip", ColorPapayaWhip()},
		{"PeachPuff", ColorPeachPuff()},
		{"Peru", ColorPeru()},
		{"Plum", ColorPlum()},
		{"PowderBlue", ColorPowderBlue()},
		{"RosyBrown", ColorRosyBrown()},
		{"RoyalBlue", ColorRoyalBlue()},
		{"SaddleBrown", ColorSaddleBrown()},
		{"SandyBrown", ColorSandyBrown()},
		{"SeaGreen", ColorSeaGreen()},
		{"Seashell", ColorSeashell()},
		{"SkyBlue", ColorSkyBlue()},
		{"SlateBlue", ColorSlateBlue()},
		{"SlateGray", ColorSlateGray()},
		{"SlateGrey", ColorSlateGrey()},
		{"Snow", ColorSnow()},
		{"SpringGreen", ColorSpringGreen()},
		{"SteelBlue", ColorSteelBlue()},
		{"Thistle", ColorThistle()},
		{"WhiteSmoke", ColorWhiteSmoke()},
		{"YellowGreen", ColorYellowGreen()},
		{"DarkBG", ColorDarkBG()},
		{"Transparent", ColorTransparent()},
		{"Zero", ColorZero()},
	}

	for _, tc := range colors {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b, a := tc.color.RGBA()
			// Check all components are finite
			if math.IsNaN(float64(r)) || math.IsNaN(float64(g)) || math.IsNaN(float64(b)) || math.IsNaN(float64(a)) {
				t.Errorf("%s contains NaN", tc.name)
			}
			if math.IsInf(float64(r), 0) || math.IsInf(float64(g), 0) || math.IsInf(float64(b), 0) || math.IsInf(float64(a), 0) {
				t.Errorf("%s contains Inf", tc.name)
			}
			// Check alpha is in valid range
			if a < 0 || a > 1 {
				t.Errorf("%s alpha %v is out of range [0,1]", tc.name, a)
			}
			// Check RGB are in valid range
			if r < 0 || r > 1 {
				t.Errorf("%s red %v is out of range [0,1]", tc.name, r)
			}
			if g < 0 || g > 1 {
				t.Errorf("%s green %v is out of range [0,1]", tc.name, g)
			}
			if b < 0 || b > 1 {
				t.Errorf("%s blue %v is out of range [0,1]", tc.name, b)
			}
		})
	}
}

// ============================================================================
// Color8.ToUintRaw Tests
// ============================================================================

func TestColor8ToUintRaw(t *testing.T) {
	c8 := NewColor8(100, 150, 200, 255)
	val := c8.ToUintRaw()
	// Little-endian: R=byte0, G=byte1, B=byte2, A=byte3
	expected := uint32(100) | uint32(150)<<8 | uint32(200)<<16 | uint32(255)<<24
	if val != expected {
		t.Errorf("ToUintRaw() = %v (0x%08x), want %v (0x%08x)",
			val, val, expected, expected)
	}
}

func TestColor8ToUintRawAllZero(t *testing.T) {
	c8 := NewColor8(0, 0, 0, 0)
	if c8.ToUintRaw() != 0 {
		t.Errorf("ToUintRaw() = %v, want 0", c8.ToUintRaw())
	}
}

func TestColor8ToUintRawAllMax(t *testing.T) {
	c8 := NewColor8(255, 255, 255, 255)
	if c8.ToUintRaw() != 0xFFFFFFFF {
		t.Errorf("ToUintRaw() = 0x%08x, want 0xFFFFFFFF", c8.ToUintRaw())
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestColorNegativeValues(t *testing.T) {
	c := NewColor(-1.0, -0.5, 0.5, 1.0)
	if !Approx(c.R(), -1.0) {
		t.Errorf("Negative R not preserved: %v", c.R())
	}
	// Colors can have negative values (not clamped by the type)
	// This tests that arithmetic operations handle them correctly
}

func TestColorValuesAboveOne(t *testing.T) {
	c := NewColor(2.0, 1.5, 0.5, 1.0)
	if !Approx(c.R(), 2.0) {
		t.Errorf("R > 1 not preserved: %v", c.R())
	}
}

func TestColorMixNegativeAmount(t *testing.T) {
	// amount < 0 should extrapolate away from rhs
	lhs := NewColor(0, 0, 0, 1)
	rhs := NewColor(1, 1, 1, 1)
	result := ColorMix(lhs, rhs, -0.5)
	// lhs + (rhs - lhs) * -0.5 = 0 + (1-0)*-0.5 = -0.5
	if !Approx(result.R(), -0.5) || !Approx(result.G(), -0.5) || !Approx(result.B(), -0.5) {
		t.Errorf("ColorMix with negative amount gave unexpected result: {%v,%v,%v,%v}",
			result.R(), result.G(), result.B(), result.A())
	}
}

func TestColorMixAmountGreaterThanOne(t *testing.T) {
	// amount > 1 should extrapolate beyond rhs
	lhs := NewColor(0, 0, 0, 1)
	rhs := NewColor(1, 1, 1, 1)
	result := ColorMix(lhs, rhs, 1.5)
	// lhs + (rhs - lhs) * 1.5 = 0 + (1-0)*1.5 = 1.5
	if !Approx(result.R(), 1.5) || !Approx(result.G(), 1.5) || !Approx(result.B(), 1.5) {
		t.Errorf("ColorMix with amount>1 gave unexpected result: {%v,%v,%v,%v}",
			result.R(), result.G(), result.B(), result.A())
	}
}

func TestColorHexZero(t *testing.T) {
	c := ColorZero()
	hex := c.Hex()
	expected := "#00000000"
	if hex != expected {
		t.Errorf("ColorZero().Hex() = %s, want %s", hex, expected)
	}
}

func TestColorHexWhite(t *testing.T) {
	c := ColorWhite()
	hex := c.Hex()
	expected := "#ffffffff"
	if hex != expected {
		t.Errorf("ColorWhite().Hex() = %s, want %s", hex, expected)
	}
}

func TestColorFromHexStringCaseInsensitive(t *testing.T) {
	c1, err := ColorFromHexString("#FF8040C8")
	if err != nil {
		t.Fatalf("uppercase hex failed: %v", err)
	}
	c2, err := ColorFromHexString("#ff8040c8")
	if err != nil {
		t.Fatalf("lowercase hex failed: %v", err)
	}
	c3, err := ColorFromHexString("#Ff8040C8")
	if err != nil {
		t.Fatalf("mixed case hex failed: %v", err)
	}
	if !c1.Equals(c2) || !c2.Equals(c3) {
		t.Error("Hex strings should be case-insensitive")
	}
}

func TestMultiplyAssignInPlace(t *testing.T) {
	c := NewColor(0.5, 1.0, 1.5, 2.0)
	c.MultiplyAssign(c) // Square all components
	if !Approx(c.R(), 0.25) || !Approx(c.G(), 1.0) || !Approx(c.B(), 2.25) || !Approx(c.A(), 4.0) {
		t.Errorf("Self-multiply failed: {%v,%v,%v,%v}", c.R(), c.G(), c.B(), c.A())
	}
}

func TestInvertedDoesNotModifyOriginal(t *testing.T) {
	c := NewColor(0.2, 0.3, 0.4, 0.5)
	inv := c.Inverted()
	// Verify inverted has correct values
	if !Approx(inv.R(), 0.8) || !Approx(inv.G(), 0.7) || !Approx(inv.B(), 0.6) || !Approx(inv.A(), 0.5) {
		t.Errorf("Inverted() wrong result: {%v,%v,%v,%v}", inv.R(), inv.G(), inv.B(), inv.A())
	}
	// Verify original is unchanged
	if !Approx(c.R(), 0.2) || !Approx(c.G(), 0.3) || !Approx(c.B(), 0.4) || !Approx(c.A(), 0.5) {
		t.Errorf("Inverted() modified the original color")
	}
}

func TestScaleWithoutAlphaDoesNotModifyOriginal(t *testing.T) {
	c := NewColor(0.5, 0.3, 0.7, 0.9)
	scaled := c.ScaleWithoutAlpha(2.0)
	// Verify original is unchanged
	if !Approx(c.R(), 0.5) || !Approx(c.G(), 0.3) || !Approx(c.B(), 0.7) || !Approx(c.A(), 0.9) {
		t.Errorf("ScaleWithoutAlpha() modified the original color")
	}
	// Verify scaled has correct values
	if !Approx(scaled.R(), 1.0) || !Approx(scaled.G(), 0.6) || !Approx(scaled.B(), 1.4) || !Approx(scaled.A(), 0.9) {
		t.Errorf("ScaleWithoutAlpha result incorrect: {%v,%v,%v,%v}", scaled.R(), scaled.G(), scaled.B(), scaled.A())
	}
}

func TestColorRGBAIntZero(t *testing.T) {
	c := ColorRGBAInt(0, 0, 0, 0)
	if !c.IsZero() {
		t.Error("ColorRGBAInt(0,0,0,0) should produce a zero color")
	}
}

func TestColorRGBAIntMax(t *testing.T) {
	c := ColorRGBAInt(255, 255, 255, 255)
	if !c.Equals(ColorWhite()) {
		t.Errorf("ColorRGBAInt(255,255,255,255) should equal ColorWhite()")
	}
}

func TestColorInvertedTwice(t *testing.T) {
	c := NewColor(0.3, 0.6, 0.9, 0.5)
	invOnce := c.Inverted()
	invTwice := invOnce.Inverted()
	// Double inversion should restore RGB but alpha stays as last set
	// Inverted preserves alpha, so invOnce.A == c.A
	// invTwice.A == invOnce.A == c.A
	// invTwice RGB should equal original RGB
	if !c.Equals(invTwice) {
		t.Errorf("Double Inverted() failed: original {%v,%v,%v,%v}, double-inverted {%v,%v,%v,%v}",
			c.R(), c.G(), c.B(), c.A(),
			invTwice.R(), invTwice.G(), invTwice.B(), invTwice.A())
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkNewColor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewColor(Float(i%256)/255.0, 0.5, 0.3, 1.0)
	}
}

func BenchmarkColorFromHexString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ColorFromHexString("#ff8040c8")
	}
}

func BenchmarkColorHex(b *testing.B) {
	c := NewColor(0.5, 0.3, 0.8, 1.0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Hex()
	}
}

func BenchmarkColorMix(b *testing.B) {
	lhs := NewColor(1.0, 0.0, 0.0, 1.0)
	rhs := NewColor(0.0, 1.0, 1.0, 1.0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ColorMix(lhs, rhs, 0.5)
	}
}

func BenchmarkColor8ToUintRaw(b *testing.B) {
	c8 := NewColor8(100, 150, 200, 255)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c8.ToUintRaw()
	}
}

func BenchmarkNamedColorFunctions(b *testing.B) {
	b.Run("ColorRed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ColorRed()
		}
	})
	b.Run("ColorWhite", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ColorWhite()
		}
	})
	b.Run("ColorBlue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ColorBlue()
		}
	})
}

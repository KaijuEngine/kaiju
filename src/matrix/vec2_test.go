/******************************************************************************/
/* vec2_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"math"
	"testing"
)

func vec2ForTest() Vec2 {
	return Vec2{1, 2}
}

func vec2ForTestOther() Vec2 {
	return Vec2{3, 4}
}

// ---- Accessors ----

func TestVec2X(t *testing.T) {
	v := vec2ForTest()
	if v.X() != 1 {
		t.Errorf("Expected X() = 1, got %f", v.X())
	}
}

func TestVec2Y(t *testing.T) {
	v := vec2ForTest()
	if v.Y() != 2 {
		t.Errorf("Expected Y() = 2, got %f", v.Y())
	}
}

func TestVec2Width(t *testing.T) {
	v := vec2ForTest()
	if v.Width() != 1 {
		t.Errorf("Expected Width() = 1, got %f", v.Width())
	}
}

func TestVec2Height(t *testing.T) {
	v := vec2ForTest()
	if v.Height() != 2 {
		t.Errorf("Expected Height() = 2, got %f", v.Height())
	}
}

func TestVec2PX(t *testing.T) {
	v := vec2ForTest()
	px := v.PX()
	if *px != 1 {
		t.Errorf("Expected *PX() = 1, got %f", *px)
	}
	*px = 10
	if v.X() != 10 {
		t.Errorf("Expected v.X() = 10 after modifying pointer, got %f", v.X())
	}
}

func TestVec2PY(t *testing.T) {
	v := vec2ForTest()
	py := v.PY()
	if *py != 2 {
		t.Errorf("Expected *PY() = 2, got %f", *py)
	}
}

func TestVec2SetX(t *testing.T) {
	v := vec2ForTest()
	v.SetX(10)
	if v.X() != 10 {
		t.Errorf("Expected X() = 10, got %f", v.X())
	}
}

func TestVec2SetY(t *testing.T) {
	v := vec2ForTest()
	v.SetY(20)
	if v.Y() != 20 {
		t.Errorf("Expected Y() = 20, got %f", v.Y())
	}
}

func TestVec2SetWidth(t *testing.T) {
	v := vec2ForTest()
	v.SetWidth(10)
	if v.Width() != 10 {
		t.Errorf("Expected Width() = 10, got %f", v.Width())
	}
}

func TestVec2SetHeight(t *testing.T) {
	v := vec2ForTest()
	v.SetHeight(20)
	if v.Height() != 20 {
		t.Errorf("Expected Height() = 20, got %f", v.Height())
	}
}

// ---- Conversion ----

func TestVec2XY(t *testing.T) {
	v := vec2ForTest()
	x, y := v.XY()
	if x != 1 || y != 2 {
		t.Errorf("Expected (1, 2), got (%f, %f)", x, y)
	}
}

func TestVec2AsVec3(t *testing.T) {
	v := vec2ForTest()
	result := v.AsVec3()
	x, y, z := result.X(), result.Y(), result.Z()
	if x != 1 || y != 2 || z != 0 {
		t.Errorf("Expected Vec3{1, 2, 0}, got (%f, %f, %f)", x, y, z)
	}
}

func TestVec2AsVec2i(t *testing.T) {
	v := Vec2{1.7, 2.3}
	result := v.AsVec2i()
	if result[0] != 1 || result[1] != 2 {
		t.Errorf("Expected Vec2i{1, 2}, got %v", result)
	}
}

// ---- Constructors ----

func TestNewVec2(t *testing.T) {
	v := NewVec2(1, 2)
	if v.X() != 1 || v.Y() != 2 {
		t.Errorf("Expected {1, 2}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2FromArray(t *testing.T) {
	a := [2]Float{7, 8}
	v := Vec2FromArray(a)
	if v.X() != 7 || v.Y() != 8 {
		t.Errorf("Expected {7, 8}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2FromSlice(t *testing.T) {
	s := []Float{10, 11}
	v := Vec2FromSlice(s)
	if v.X() != 10 || v.Y() != 11 {
		t.Errorf("Expected {10, 11}, got (%f, %f)", v.X(), v.Y())
	}
}

// ---- Arithmetic ----

func TestVec2Add(t *testing.T) {
	a := vec2ForTest()
	b := vec2ForTestOther()
	result := a.Add(b)
	if result.X() != 4 || result.Y() != 6 {
		t.Errorf("Expected {4, 6}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2AddAssign(t *testing.T) {
	v := vec2ForTest()
	v.AddAssign(vec2ForTestOther())
	if v.X() != 4 || v.Y() != 6 {
		t.Errorf("Expected {4, 6}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Subtract(t *testing.T) {
	a := vec2ForTestOther()
	b := vec2ForTest()
	result := a.Subtract(b)
	if result.X() != 2 || result.Y() != 2 {
		t.Errorf("Expected {2, 2}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2SubtractAssign(t *testing.T) {
	v := vec2ForTestOther()
	v.SubtractAssign(vec2ForTest())
	if v.X() != 2 || v.Y() != 2 {
		t.Errorf("Expected {2, 2}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Multiply(t *testing.T) {
	a := vec2ForTest()
	b := vec2ForTestOther()
	result := a.Multiply(b)
	if result.X() != 3 || result.Y() != 8 {
		t.Errorf("Expected {3, 8}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2MultiplyAssign(t *testing.T) {
	v := vec2ForTest()
	v.MultiplyAssign(vec2ForTestOther())
	if v.X() != 3 || v.Y() != 8 {
		t.Errorf("Expected {3, 8}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Divide(t *testing.T) {
	a := vec2ForTestOther()
	b := vec2ForTest()
	result := a.Divide(b)
	if !Vec2ApproxTo(result, Vec2{3, 2}, Tiny) {
		t.Errorf("Expected {3, 2}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2DivideAssign(t *testing.T) {
	v := vec2ForTestOther()
	v.DivideAssign(vec2ForTest())
	if !Vec2ApproxTo(v, Vec2{3, 2}, Tiny) {
		t.Errorf("Expected {3, 2}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Scale(t *testing.T) {
	v := vec2ForTest()
	result := v.Scale(2)
	if result.X() != 2 || result.Y() != 4 {
		t.Errorf("Expected {2, 4}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2ScaleAssign(t *testing.T) {
	v := vec2ForTest()
	v.ScaleAssign(2)
	if v.X() != 2 || v.Y() != 4 {
		t.Errorf("Expected {2, 4}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Shrink(t *testing.T) {
	v := vec2ForTest()
	result := v.Shrink(2)
	if !Vec2ApproxTo(result, Vec2{0.5, 1}, Tiny) {
		t.Errorf("Expected {0.5, 1}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2ShrinkAssign(t *testing.T) {
	v := vec2ForTest()
	v.ShrinkAssign(2)
	if !Vec2ApproxTo(v, Vec2{0.5, 1}, Tiny) {
		t.Errorf("Expected {0.5, 1}, got (%f, %f)", v.X(), v.Y())
	}
}

// ---- Length / Normalization ----

func TestVec2Length(t *testing.T) {
	v := Vec2{3, 4}
	expected := Sqrt(Float(25))
	if Abs(v.Length()-expected) > Tiny {
		t.Errorf("Expected Length = %f, got %f", expected, v.Length())
	}
}

func TestVec2Normal(t *testing.T) {
	v := Vec2{3, 0}
	result := v.Normal()
	if !Vec2Approx(result, Vec2{1, 0}) {
		t.Errorf("Expected {1, 0}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2Normalize(t *testing.T) {
	v := Vec2{3, 0}
	v.Normalize()
	if !Vec2Approx(v, Vec2{1, 0}) {
		t.Errorf("Expected {1, 0}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2LengthOfZero(t *testing.T) {
	v := Vec2Zero()
	if v.Length() != 0 {
		t.Errorf("Expected Length of zero vector = 0, got %f", v.Length())
	}
}

func TestVec2NormalOfZero(t *testing.T) {
	v := Vec2Zero()
	result := v.Normal()
	// Normal of zero vector produces NaN due to division by zero
	if !result.IsNaN() {
		t.Errorf("Expected Normal of zero vector to be NaN, got (%f, %f)", result.X(), result.Y())
	}
}

// ---- Negative / Inverse ----

func TestVec2Negative(t *testing.T) {
	v := vec2ForTest()
	result := v.Negative()
	if result.X() != -1 || result.Y() != -2 {
		t.Errorf("Expected {-1, -2}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2Inverse(t *testing.T) {
	v := vec2ForTest()
	v.Inverse()
	if v.X() != -1 || v.Y() != -2 {
		t.Errorf("Expected {-1, -2}, got (%f, %f)", v.X(), v.Y())
	}
}

// ---- Dot Product ----

func TestVec2Dot(t *testing.T) {
	a := vec2ForTest()
	b := vec2ForTestOther()
	result := Vec2Dot(a, b)
	if result != 11 {
		t.Errorf("Expected Dot({1,2}, {3,4}) = 11, got %f", result)
	}
}

func TestVec2DotOrthogonal(t *testing.T) {
	result := Vec2Dot(Vec2Right(), Vec2Up())
	if result != 0 {
		t.Errorf("Expected dot of orthogonal vectors = 0, got %f", result)
	}
}

func TestVec2DotPerpendicular(t *testing.T) {
	pairs := [][2]Vec2{
		{Vec2Right(), Vec2Up()},
		{Vec2Left(), Vec2Up()},
		{Vec2Right(), Vec2Down()},
		{Vec2Left(), Vec2Down()},
	}
	for _, p := range pairs {
		dot := Vec2Dot(p[0], p[1])
		if Abs(dot) > Tiny {
			t.Errorf("Expected dot of perpendicular vectors to be near 0, got %f", dot)
		}
	}
}

// ---- Comparison ----

func TestVec2Approx(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{1 + math.SmallestNonzeroFloat32/2, 2}
	if !Vec2Approx(a, b) {
		t.Errorf("Expected Vec2Approx to return true for close vectors")
	}
}

func TestVec2ApproxFar(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{100, 200}
	if Vec2Approx(a, b) {
		t.Errorf("Expected Vec2Approx to return false for far vectors")
	}
}

func TestVec2ApproxTo(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{1.05, 2.05}
	if !Vec2ApproxTo(a, b, 0.1) {
		t.Errorf("Expected Vec2ApproxTo with delta 0.1 to return true")
	}
}

func TestVec2ApproxToFar(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{100, 200}
	if Vec2ApproxTo(a, b, 0.1) {
		t.Errorf("Expected Vec2ApproxTo with delta 0.1 to return false for far vectors")
	}
}

func TestVec2Nearly(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{1 + Tiny/2, 2}
	if !Vec2Nearly(a, b) {
		t.Errorf("Expected Vec2Nearly to return true for close vectors")
	}
}

func TestVec2NearlyFar(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{100, 200}
	if Vec2Nearly(a, b) {
		t.Errorf("Expected Vec2Nearly to return false for far vectors")
	}
}

func TestVec2Roughly(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{1 + Tiny, 2 + Tiny}
	if !Vec2Roughly(a, b) {
		t.Errorf("Expected Vec2Roughly to return true for close vectors")
	}
}

func TestVec2RoughlyFar(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{100, 200}
	if Vec2Roughly(a, b) {
		t.Errorf("Expected Vec2Roughly to return false for far vectors")
	}
}

func TestVec2Equals(t *testing.T) {
	a := vec2ForTest()
	b := vec2ForTest()
	if !a.Equals(b) {
		t.Errorf("Expected Equals to return true for same vectors")
	}
}

func TestVec2EqualsDifference(t *testing.T) {
	a := vec2ForTest()
	b := Vec2{100, 2}
	if a.Equals(b) {
		t.Errorf("Expected Equals to return false for different vectors")
	}
}

// ---- Abs ----

func TestVec2Abs(t *testing.T) {
	v := Vec2{-1, -2}
	result := v.Abs()
	if result.X() != 1 || result.Y() != 2 {
		t.Errorf("Expected {1, 2}, got (%f, %f)", result.X(), result.Y())
	}
}

// ---- Min / Max ----

func TestVec2Min(t *testing.T) {
	a := vec2ForTest()
	b := vec2ForTestOther()
	result := Vec2Min(a, b)
	if result.X() != 1 || result.Y() != 2 {
		t.Errorf("Expected {1, 2}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2Max(t *testing.T) {
	a := vec2ForTest()
	b := vec2ForTestOther()
	result := Vec2Max(a, b)
	if result.X() != 3 || result.Y() != 4 {
		t.Errorf("Expected {3, 4}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2MinAbs(t *testing.T) {
	a := Vec2{-1, -5}
	b := Vec2{4, -2}
	result := Vec2MinAbs(a, b)
	if result.X() != 1 || result.Y() != 2 {
		t.Errorf("Expected {1, 2}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2MaxAbs(t *testing.T) {
	a := Vec2{-1, -5}
	b := Vec2{4, -2}
	result := Vec2MaxAbs(a, b)
	if result.X() != 4 || result.Y() != 5 {
		t.Errorf("Expected {4, 5}, got (%f, %f)", result.X(), result.Y())
	}
}

// ---- Distance ----

func TestVec2Distance(t *testing.T) {
	a := vec2ForTest()
	b := vec2ForTestOther()
	result := a.Distance(b)
	expected := Sqrt(Float(8))
	if Abs(result-expected) > Tiny {
		t.Errorf("Expected Distance = %f, got %f", expected, result)
	}
}

func TestVec2DistanceZero(t *testing.T) {
	a := vec2ForTest()
	if a.Distance(a) != 0 {
		t.Errorf("Expected distance from vector to itself = 0, got %f", a.Distance(a))
	}
}

// ---- Lerp ----

func TestVec2Lerp(t *testing.T) {
	from := vec2ForTest()
	to := vec2ForTestOther()
	result := Vec2Lerp(from, to, 0.5)
	if result.X() != 2 || result.Y() != 3 {
		t.Errorf("Expected {2, 3}, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2LerpFrom(t *testing.T) {
	from := vec2ForTest()
	to := vec2ForTestOther()
	result := Vec2Lerp(from, to, 0)
	if result.X() != 1 || result.Y() != 2 {
		t.Errorf("Expected lerp(t=0) to return from, got (%f, %f)", result.X(), result.Y())
	}
}

func TestVec2LerpTo(t *testing.T) {
	from := vec2ForTest()
	to := vec2ForTestOther()
	result := Vec2Lerp(from, to, 1)
	if result.X() != 3 || result.Y() != 4 {
		t.Errorf("Expected lerp(t=1) to return to, got (%f, %f)", result.X(), result.Y())
	}
}

// ---- String ----

func TestVec2String(t *testing.T) {
	v := vec2ForTest()
	result := v.String()
	expected := "1.000000, 2.000000"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestVec2FromString(t *testing.T) {
	str := "1.500000, 2.500000"
	v := Vec2FromString(str)
	if Float(v.X()) != Float(1.5) || Float(v.Y()) != Float(2.5) {
		t.Errorf("Expected {1.5, 2.5}, got (%f, %f)", v.X(), v.Y())
	}
}

// ---- Angle ----

func TestVec2Angle(t *testing.T) {
	a := vec2ForTest()
	// Angle with itself should be 0
	angle := a.Angle(a)
	if Abs(angle) > Tiny {
		t.Errorf("Expected angle of vector with itself = 0, got %f", angle)
	}
}

func TestVec2AnglePerpendicular(t *testing.T) {
	a := Vec2Right()
	b := Vec2Up()
	angle := a.Angle(b)
	// 90 degrees in radians = pi/2
	expected := Float(1.5707964) // pi/2 approx
	if Abs(angle-expected) > 0.01 {
		t.Errorf("Expected angle ~%f (90 degrees), got %f", expected, angle)
	}
}

func TestVec2AngleZeroVector(t *testing.T) {
	a := vec2ForTest()
	// Angle with zero length will produce NaN or Inf, just check it doesn't panic
	_ = a.Angle(Vec2Zero())
}

// ---- Direction Constants ----

func TestVec2Up(t *testing.T) {
	v := Vec2Up()
	if v.X() != 0 || v.Y() != 1 {
		t.Errorf("Expected {0, 1}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Down(t *testing.T) {
	v := Vec2Down()
	if v.X() != 0 || v.Y() != -1 {
		t.Errorf("Expected {0, -1}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Left(t *testing.T) {
	v := Vec2Left()
	if v.X() != -1 || v.Y() != 0 {
		t.Errorf("Expected {-1, 0}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Right(t *testing.T) {
	v := Vec2Right()
	if v.X() != 1 || v.Y() != 0 {
		t.Errorf("Expected {1, 0}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Zero(t *testing.T) {
	v := Vec2Zero()
	if v.X() != 0 || v.Y() != 0 {
		t.Errorf("Expected {0, 0}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2One(t *testing.T) {
	v := Vec2One()
	if v.X() != 1 || v.Y() != 1 {
		t.Errorf("Expected {1, 1}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Half(t *testing.T) {
	v := Vec2Half()
	if Float(v[0]) != Float(0.5) || Float(v[1]) != Float(0.5) {
		t.Errorf("Expected {0.5, 0.5}, got (%f, %f)", v.X(), v.Y())
	}
}

func TestVec2Largest(t *testing.T) {
	v := Vec2Largest()
	if v[0] != FloatMax || v[1] != FloatMax {
		t.Errorf("Expected {FloatMax, FloatMax}, got (%f, %f)", v.X(), v.Y())
	}
}

// ---- LargestAxis / LargestAxisDelta ----

func TestVec2LargestAxis(t *testing.T) {
	v := Vec2{1, 5}
	if v.LargestAxis() != 5 {
		t.Errorf("Expected LargestAxis = 5, got %f", v.LargestAxis())
	}
}

func TestVec2LargestAxisDelta(t *testing.T) {
	v := Vec2{1, 5}
	delta := v.LargestAxisDelta()
	// min=1, max=5, Abs(5) > Abs(1), so returns 5
	if Float(delta) != 5 {
		t.Errorf("Expected LargestAxisDelta = 5, got %f", delta)
	}
}

func TestVec2LargestAxisDeltaNegative(t *testing.T) {
	v := Vec2{-10, 5}
	delta := v.LargestAxisDelta()
	// min=5, max=-10 (by value), Abs(-10) > Abs(5), so returns -10
	if Float(delta) != -10 {
		t.Errorf("Expected LargestAxisDelta = -10, got %f", delta)
	}
}

// ---- Inf / NaN ----

func TestVec2Inf(t *testing.T) {
	v := Vec2Inf(1)
	if !v.IsInf(1) {
		t.Errorf("Expected Vec2Inf(1) to pass IsInf(1)")
	}
}

func TestVec2InfNegative(t *testing.T) {
	v := Vec2Inf(-1)
	if !v.IsInf(-1) {
		t.Errorf("Expected Vec2Inf(-1) to pass IsInf(-1)")
	}
}

func TestVec2NaN(t *testing.T) {
	v := Vec2NaN()
	if !v.IsNaN() {
		t.Errorf("Expected Vec2NaN to pass IsNaN()")
	}
}

func TestVec2IsZero(t *testing.T) {
	if !Vec2Zero().IsZero() {
		t.Errorf("Expected Vec2Zero().IsZero() = true")
	}
}

func TestVec2IsZeroFalse(t *testing.T) {
	if vec2ForTest().IsZero() {
		t.Errorf("Expected non-zero vector IsZero() = false")
	}
}

func TestVec2IsInf(t *testing.T) {
	v := Vec2Inf(1)
	if !v.IsInf(1) {
		t.Errorf("IsInf(1) should be true")
	}
}

func TestVec2IsInfinityFalse(t *testing.T) {
	v := vec2ForTest()
	if v.IsInf(1) {
		t.Errorf("Regular vector should not be Inf")
	}
}

func TestVec2IsNaN(t *testing.T) {
	v := Vec2NaN()
	if !v.IsNaN() {
		t.Errorf("IsNaN() should be true for NaN vector")
	}
}

func TestVec2IsNotNaN(t *testing.T) {
	v := vec2ForTest()
	if v.IsNaN() {
		t.Errorf("Regular vector should not be NaN")
	}
}

func TestVec2IsValidNonZero(t *testing.T) {
	v := vec2ForTest()
	if !v.IsValidNonZero() {
		t.Errorf("Expected regular vector to be valid non-zero")
	}
}

func TestVec2IsValidNonZeroFailsOnZero(t *testing.T) {
	v := Vec2Zero()
	if v.IsValidNonZero() {
		t.Errorf("Expected zero vector to fail IsValidNonZero")
	}
}

func TestVec2IsValidNonZeroFailsOnNaN(t *testing.T) {
	v := Vec2NaN()
	if v.IsValidNonZero() {
		t.Errorf("Expected NaN vector to fail IsValidNonZero")
	}
}

func TestVec2IsValidNonZeroFailsOnInf(t *testing.T) {
	v := Vec2Inf(1)
	if v.IsValidNonZero() {
		t.Errorf("Expected Inf vector to fail IsValidNonZero")
	}
}

// ---- Benchmarks ----

func BenchmarkVec2Add(b *testing.B) {
	a := vec2ForTest()
	c := vec2ForTestOther()
	for i := 0; i < b.N; i++ {
		a.Add(c)
	}
}

func BenchmarkVec2Dot(b *testing.B) {
	a := vec2ForTest()
	c := vec2ForTestOther()
	for i := 0; i < b.N; i++ {
		Vec2Dot(a, c)
	}
}

func BenchmarkVec2Normal(b *testing.B) {
	v := vec2ForTest()
	for i := 0; i < b.N; i++ {
		v.Normal()
	}
}

func BenchmarkVec2Scale(b *testing.B) {
	v := vec2ForTest()
	for i := 0; i < b.N; i++ {
		v.Scale(2)
	}
}

/******************************************************************************/
/* vec4_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"testing"
)

func vec4ForTest() Vec4 {
	return Vec4{1, 2, 3, 4}
}

func vec4ForTestOther() Vec4 {
	return Vec4{5, 6, 7, 8}
}

// ---- Accessors ----

func TestVec4X(t *testing.T) {
	v := vec4ForTest()
	if v.X() != 1 {
		t.Errorf("Expected X() = 1, got %f", v.X())
	}
}

func TestVec4Y(t *testing.T) {
	v := vec4ForTest()
	if v.Y() != 2 {
		t.Errorf("Expected Y() = 2, got %f", v.Y())
	}
}

func TestVec4Z(t *testing.T) {
	v := vec4ForTest()
	if v.Z() != 3 {
		t.Errorf("Expected Z() = 3, got %f", v.Z())
	}
}

func TestVec4W(t *testing.T) {
	v := vec4ForTest()
	if v.W() != 4 {
		t.Errorf("Expected W() = 4, got %f", v.W())
	}
}

func TestVec4Left(t *testing.T) {
	v := vec4ForTest()
	if v.Left() != 1 {
		t.Errorf("Expected Left() = 1, got %f", v.Left())
	}
}

func TestVec4Top(t *testing.T) {
	v := vec4ForTest()
	if v.Top() != 2 {
		t.Errorf("Expected Top() = 2, got %f", v.Top())
	}
}

func TestVec4Right(t *testing.T) {
	v := vec4ForTest()
	if v.Right() != 3 {
		t.Errorf("Expected Right() = 3, got %f", v.Right())
	}
}

func TestVec4Bottom(t *testing.T) {
	v := vec4ForTest()
	if v.Bottom() != 4 {
		t.Errorf("Expected Bottom() = 4, got %f", v.Bottom())
	}
}

func TestVec4Width(t *testing.T) {
	v := vec4ForTest()
	if v.Width() != 3 {
		t.Errorf("Expected Width() = 3, got %f", v.Width())
	}
}

func TestVec4Height(t *testing.T) {
	v := vec4ForTest()
	if v.Height() != 4 {
		t.Errorf("Expected Height() = 4, got %f", v.Height())
	}
}

func TestVec4PX(t *testing.T) {
	v := vec4ForTest()
	px := v.PX()
	if *px != 1 {
		t.Errorf("Expected *PX() = 1, got %f", *px)
	}
	*px = 10
	if v.X() != 10 {
		t.Errorf("Expected v.X() = 10 after modifying pointer, got %f", v.X())
	}
}

func TestVec4PY(t *testing.T) {
	v := vec4ForTest()
	py := v.PY()
	if *py != 2 {
		t.Errorf("Expected *PY() = 2, got %f", *py)
	}
}

func TestVec4PZ(t *testing.T) {
	v := vec4ForTest()
	pz := v.PZ()
	if *pz != 3 {
		t.Errorf("Expected *PZ() = 3, got %f", *pz)
	}
}

func TestVec4PW(t *testing.T) {
	v := vec4ForTest()
	pw := v.PW()
	if *pw != 4 {
		t.Errorf("Expected *PW() = 4, got %f", *pw)
	}
}

func TestVec4SetX(t *testing.T) {
	v := vec4ForTest()
	v.SetX(10)
	if v.X() != 10 {
		t.Errorf("Expected X() = 10, got %f", v.X())
	}
}

func TestVec4SetY(t *testing.T) {
	v := vec4ForTest()
	v.SetY(20)
	if v.Y() != 20 {
		t.Errorf("Expected Y() = 20, got %f", v.Y())
	}
}

func TestVec4SetZ(t *testing.T) {
	v := vec4ForTest()
	v.SetZ(30)
	if v.Z() != 30 {
		t.Errorf("Expected Z() = 30, got %f", v.Z())
	}
}

func TestVec4SetW(t *testing.T) {
	v := vec4ForTest()
	v.SetW(40)
	if v.W() != 40 {
		t.Errorf("Expected W() = 40, got %f", v.W())
	}
}

func TestVec4SetLeft(t *testing.T) {
	v := vec4ForTest()
	v.SetLeft(10)
	if v.X() != 10 {
		t.Errorf("Expected X() = 10 after SetLeft, got %f", v.X())
	}
}

func TestVec4SetTop(t *testing.T) {
	v := vec4ForTest()
	v.SetTop(20)
	if v.Y() != 20 {
		t.Errorf("Expected Y() = 20 after SetTop, got %f", v.Y())
	}
}

func TestVec4SetRight(t *testing.T) {
	v := vec4ForTest()
	v.SetRight(30)
	if v.Z() != 30 {
		t.Errorf("Expected Z() = 30 after SetRight, got %f", v.Z())
	}
}

func TestVec4SetBottom(t *testing.T) {
	v := vec4ForTest()
	v.SetBottom(40)
	if v.W() != 40 {
		t.Errorf("Expected W() = 40 after SetBottom, got %f", v.W())
	}
}

func TestVec4SetWidth(t *testing.T) {
	v := vec4ForTest()
	v.SetWidth(30)
	if v.Z() != 30 {
		t.Errorf("Expected Z() = 30 after SetWidth, got %f", v.Z())
	}
}

func TestVec4SetHeight(t *testing.T) {
	v := vec4ForTest()
	v.SetHeight(40)
	if v.W() != 40 {
		t.Errorf("Expected W() = 40 after SetHeight, got %f", v.W())
	}
}

// ---- Conversion ----

func TestVec4AsVec3(t *testing.T) {
	v := vec4ForTest()
	result := v.AsVec3()
	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 {
		t.Errorf("Expected Vec3{1, 2, 3}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec4XYZW(t *testing.T) {
	v := vec4ForTest()
	x, y, z, w := v.XYZW()
	if x != 1 || y != 2 || z != 3 || w != 4 {
		t.Errorf("Expected (1, 2, 3, 4), got (%f, %f, %f, %f)", x, y, z, w)
	}
}

func TestVec4Horizontal(t *testing.T) {
	v := vec4ForTest()
	if v.Horizontal() != 4 {
		t.Errorf("Expected Horizontal() = 4 (1+3), got %f", v.Horizontal())
	}
}

func TestVec4Vertical(t *testing.T) {
	v := vec4ForTest()
	if v.Vertical() != 6 {
		t.Errorf("Expected Vertical() = 6 (2+4), got %f", v.Vertical())
	}
}

func TestVec4AsVec4i(t *testing.T) {
	v := Vec4{1.7, 2.3, 3.9, 4.1}
	result := v.AsVec4i()
	if result[0] != 1 || result[1] != 2 || result[2] != 3 || result[3] != 4 {
		t.Errorf("Expected Vec4i{1, 2, 3, 4}, got %v", result)
	}
}

// ---- Constructors ----

func TestNewVec4(t *testing.T) {
	v := NewVec4(1, 2, 3, 4)
	if v.X() != 1 || v.Y() != 2 || v.Z() != 3 || v.W() != 4 {
		t.Errorf("Expected {1, 2, 3, 4}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4FromArray(t *testing.T) {
	a := [4]Float{7, 8, 9, 10}
	v := Vec4FromArray(a)
	if v.X() != 7 || v.Y() != 8 || v.Z() != 9 || v.W() != 10 {
		t.Errorf("Expected {7, 8, 9, 10}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4FromSlice(t *testing.T) {
	s := []Float{11, 12, 13, 14}
	v := Vec4FromSlice(s)
	if v.X() != 11 || v.Y() != 12 || v.Z() != 13 || v.W() != 14 {
		t.Errorf("Expected {11, 12, 13, 14}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

// ---- Arithmetic ----

func TestVec4Add(t *testing.T) {
	a := vec4ForTest()
	b := vec4ForTestOther()
	result := a.Add(b)
	if result.X() != 6 || result.Y() != 8 || result.Z() != 10 || result.W() != 12 {
		t.Errorf("Expected {6, 8, 10, 12}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4AddAssign(t *testing.T) {
	v := vec4ForTest()
	v.AddAssign(vec4ForTestOther())
	if v.X() != 6 || v.Y() != 8 || v.Z() != 10 || v.W() != 12 {
		t.Errorf("Expected {6, 8, 10, 12}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4Subtract(t *testing.T) {
	a := vec4ForTestOther()
	b := vec4ForTest()
	result := a.Subtract(b)
	if result.X() != 4 || result.Y() != 4 || result.Z() != 4 || result.W() != 4 {
		t.Errorf("Expected {4, 4, 4, 4}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4SubtractAssign(t *testing.T) {
	v := vec4ForTestOther()
	v.SubtractAssign(vec4ForTest())
	if v.X() != 4 || v.Y() != 4 || v.Z() != 4 || v.W() != 4 {
		t.Errorf("Expected {4, 4, 4, 4}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4Multiply(t *testing.T) {
	a := vec4ForTest()
	b := vec4ForTestOther()
	result := a.Multiply(b)
	if result.X() != 5 || result.Y() != 12 || result.Z() != 21 || result.W() != 32 {
		t.Errorf("Expected {5, 12, 21, 32}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4MultiplyAssign(t *testing.T) {
	v := vec4ForTest()
	v.MultiplyAssign(vec4ForTestOther())
	if v.X() != 5 || v.Y() != 12 || v.Z() != 21 || v.W() != 32 {
		t.Errorf("Expected {5, 12, 21, 32}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4Divide(t *testing.T) {
	a := vec4ForTestOther()
	b := vec4ForTest()
	result := a.Divide(b)
	if !Vec4ApproxTo(result, Vec4{5, 3, 7.0 / 3, 2}, Tiny) {
		t.Errorf("Expected {5, 3, 2.333.., 2}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4DivideAssign(t *testing.T) {
	v := vec4ForTestOther()
	v.DivideAssign(vec4ForTest())
	if !Vec4ApproxTo(v, Vec4{5, 3, 7.0 / 3, 2}, Tiny) {
		t.Errorf("Expected {5, 3, 2.333.., 2}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4Scale(t *testing.T) {
	v := vec4ForTest()
	result := v.Scale(2)
	if result.X() != 2 || result.Y() != 4 || result.Z() != 6 || result.W() != 8 {
		t.Errorf("Expected {2, 4, 6, 8}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4ScaleAssign(t *testing.T) {
	v := vec4ForTest()
	v.ScaleAssign(2)
	if v.X() != 2 || v.Y() != 4 || v.Z() != 6 || v.W() != 8 {
		t.Errorf("Expected {2, 4, 6, 8}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4Shrink(t *testing.T) {
	v := vec4ForTest()
	result := v.Shrink(2)
	if !Vec4ApproxTo(result, Vec4{0.5, 1, 1.5, 2}, Tiny) {
		t.Errorf("Expected {0.5, 1, 1.5, 2}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4ShrinkAssign(t *testing.T) {
	v := vec4ForTest()
	v.ShrinkAssign(2)
	if !Vec4ApproxTo(v, Vec4{0.5, 1, 1.5, 2}, Tiny) {
		t.Errorf("Expected {0.5, 1, 1.5, 2}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

// ---- Length / Normalization ----

func TestVec4Length(t *testing.T) {
	v := Vec4{1, 2, 2, 0}
	expected := Sqrt(Float(9))
	if Abs(v.Length()-expected) > Tiny {
		t.Errorf("Expected Length = %f, got %f", expected, v.Length())
	}
}

func TestVec4LengthSquared(t *testing.T) {
	v := Vec4{1, 2, 2, 0}
	if Vec4Dot(v, v) != 9 {
		t.Errorf("Expected LengthSquared = 9, got %f", Vec4Dot(v, v))
	}
}

func TestVec4Normal(t *testing.T) {
	v := Vec4{3, 0, 0, 0}
	result := v.Normal()
	if !Vec4Approx(result, Vec4{1, 0, 0, 0}) {
		t.Errorf("Expected {1, 0, 0, 0}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4Normalize(t *testing.T) {
	v := Vec4{3, 0, 0, 0}
	v.Normalize()
	if !Vec4Approx(v, Vec4{1, 0, 0, 0}) {
		t.Errorf("Expected {1, 0, 0, 0}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4LengthOfZero(t *testing.T) {
	v := Vec4Zero()
	if v.Length() != 0 {
		t.Errorf("Expected Length of zero vector = 0, got %f", v.Length())
	}
}

func TestVec4NormalOfZero(t *testing.T) {
	v := Vec4Zero()
	result := v.Normal()
	// Normal of zero vector produces NaN due to division by zero
	if !IsNaN(result.X()) {
		t.Errorf("Expected Normal of zero vector to be NaN, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

// ---- Negative / Inverse ----

func TestVec4Negative(t *testing.T) {
	v := vec4ForTest()
	result := v.Negative()
	if result.X() != -1 || result.Y() != -2 || result.Z() != -3 || result.W() != -4 {
		t.Errorf("Expected {-1, -2, -3, -4}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4Inverse(t *testing.T) {
	v := vec4ForTest()
	v.Inverse()
	if v.X() != -1 || v.Y() != -2 || v.Z() != -3 || v.W() != -4 {
		t.Errorf("Expected {-1, -2, -3, -4}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

// ---- Comparison ----

func TestVec4Roughly(t *testing.T) {
	a := vec4ForTest()
	b := Vec4{1 + Roughly/2, 2, 3, 4}
	if !Vec4Roughly(a, b) {
		t.Errorf("Expected Vec4Roughly to return true for close vectors")
	}
}

func TestVec4RoughlyFar(t *testing.T) {
	a := vec4ForTest()
	b := Vec4{100, 200, 300, 400}
	if Vec4Roughly(a, b) {
		t.Errorf("Expected Vec4Roughly to return false for far vectors")
	}
}

func TestVec4Approx(t *testing.T) {
	a := vec4ForTest()
	b := Vec4{1 + FloatSmallestNonzero/2, 2, 3, 4}
	if !Vec4Approx(a, b) {
		t.Errorf("Expected Vec4Approx to return true for close vectors")
	}
}

func TestVec4ApproxFar(t *testing.T) {
	a := vec4ForTest()
	b := Vec4{100, 200, 300, 400}
	if Vec4Approx(a, b) {
		t.Errorf("Expected Vec4Approx to return false for far vectors")
	}
}

func TestVec4ApproxTo(t *testing.T) {
	a := vec4ForTest()
	b := Vec4{1.05, 2.05, 3.05, 4.05}
	if !Vec4ApproxTo(a, b, 0.1) {
		t.Errorf("Expected Vec4ApproxTo with delta 0.1 to return true")
	}
}

func TestVec4ApproxToFar(t *testing.T) {
	a := vec4ForTest()
	b := Vec4{100, 200, 300, 400}
	if Vec4ApproxTo(a, b, 0.1) {
		t.Errorf("Expected Vec4ApproxTo with delta 0.1 to return false for far vectors")
	}
}

func TestVec4Equals(t *testing.T) {
	a := vec4ForTest()
	b := vec4ForTest()
	if !a.Equals(b) {
		t.Errorf("Expected Equals to return true for same vectors")
	}
}

func TestVec4EqualsDifference(t *testing.T) {
	a := vec4ForTest()
	b := Vec4{100, 2, 3, 4}
	if a.Equals(b) {
		t.Errorf("Expected Equals to return false for different vectors")
	}
}

// ---- Min / Max ----

func TestVec4Min(t *testing.T) {
	a := vec4ForTest()
	b := vec4ForTestOther()
	result := Vec4Min(a, b)
	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 || result.W() != 4 {
		t.Errorf("Expected {1, 2, 3, 4}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4Max(t *testing.T) {
	a := vec4ForTest()
	b := vec4ForTestOther()
	result := Vec4Max(a, b)
	if result.X() != 5 || result.Y() != 6 || result.Z() != 7 || result.W() != 8 {
		t.Errorf("Expected {5, 6, 7, 8}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4MinAbs(t *testing.T) {
	a := Vec4{-1, -5, 3, -10}
	b := Vec4{4, -2, 6, 8}
	result := Vec4MinAbs(a, b)
	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 || result.W() != 8 {
		t.Errorf("Expected {1, 2, 3, 8}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4MaxAbs(t *testing.T) {
	a := Vec4{-1, -5, 3, -10}
	b := Vec4{4, -2, 6, 8}
	result := Vec4MaxAbs(a, b)
	if result.X() != 4 || result.Y() != 5 || result.Z() != 6 || result.W() != 10 {
		t.Errorf("Expected {4, 5, 6, 10}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

// ---- Abs ----

func TestVec4Abs(t *testing.T) {
	v := Vec4{-1, -2, 3, -4}
	result := v.Abs()
	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 || result.W() != 4 {
		t.Errorf("Expected {1, 2, 3, 4}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

// ---- Dot Product ----

func TestVec4Dot(t *testing.T) {
	a := vec4ForTest()
	b := vec4ForTestOther()
	// 1*5 + 2*6 + 3*7 + 4*8 = 5 + 12 + 21 + 32 = 70
	result := Vec4Dot(a, b)
	if result != 70 {
		t.Errorf("Expected Dot({1,2,3,4}, {5,6,7,8}) = 70, got %f", result)
	}
}

func TestVec4DotSelf(t *testing.T) {
	v := Vec4{1, 0, 0, 0}
	result := Vec4Dot(v, v)
	if result != 1 {
		t.Errorf("Expected dot of vector with itself (unit) = 1, got %f", result)
	}
}

// ---- Distance ----

func TestVec4Distance(t *testing.T) {
	a := vec4ForTest()
	b := vec4ForTestOther()
	result := a.Distance(b)
	// sqrt(4^2 + 4^2 + 4^2 + 4^2) = sqrt(64) = 8
	if Abs(result-8) > Tiny {
		t.Errorf("Expected Distance = 8, got %f", result)
	}
}

func TestVec4DistanceZero(t *testing.T) {
	a := vec4ForTest()
	if a.Distance(a) != 0 {
		t.Errorf("Expected distance from vector to itself = 0, got %f", a.Distance(a))
	}
}

// ---- Lerp ----

func TestVec4Lerp(t *testing.T) {
	from := vec4ForTest()
	to := vec4ForTestOther()
	result := Vec4Lerp(from, to, 0.5)
	if result.X() != 3 || result.Y() != 4 || result.Z() != 5 || result.W() != 6 {
		t.Errorf("Expected {3, 4, 5, 6}, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4LerpFrom(t *testing.T) {
	from := vec4ForTest()
	to := vec4ForTestOther()
	result := Vec4Lerp(from, to, 0)
	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 || result.W() != 4 {
		t.Errorf("Expected lerp(t=0) to return from, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

func TestVec4LerpTo(t *testing.T) {
	from := vec4ForTest()
	to := vec4ForTestOther()
	result := Vec4Lerp(from, to, 1)
	if result.X() != 5 || result.Y() != 6 || result.Z() != 7 || result.W() != 8 {
		t.Errorf("Expected lerp(t=1) to return to, got (%f, %f, %f, %f)", result.X(), result.Y(), result.Z(), result.W())
	}
}

// ---- String ----

func TestVec4String(t *testing.T) {
	v := vec4ForTest()
	result := v.String()
	expected := "1.000000, 2.000000, 3.000000, 4.000000"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestVec4FromString(t *testing.T) {
	str := "1.500000, 2.500000, 3.500000, 4.500000"
	v := Vec4FromString(str)
	if Float(v.X()) != Float(1.5) || Float(v.Y()) != Float(2.5) || Float(v.Z()) != Float(3.5) || Float(v.W()) != Float(4.5) {
		t.Errorf("Expected {1.5, 2.5, 3.5, 4.5}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

// ---- Angle ----

func TestVec4AngleEqual(t *testing.T) {
	v := Vec4{1, 0, 0, 0}
	angle := v.Angle(v)
	if angle != 0 {
		t.Errorf("Expected angle of equal unit vectors = 0, got %f", angle)
	}
}

func TestVec4AngleZeroVector(t *testing.T) {
	// Acos with zero length will produce Inf or NaN, just check it doesn't panic
	a := vec4ForTest()
	_ = a.Angle(Vec4Zero())
}

// ---- Direction Constants ----

func TestVec4Zero(t *testing.T) {
	v := Vec4Zero()
	if v.X() != 0 || v.Y() != 0 || v.Z() != 0 || v.W() != 0 {
		t.Errorf("Expected {0, 0, 0, 0}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4One(t *testing.T) {
	v := Vec4One()
	if v.X() != 1 || v.Y() != 1 || v.Z() != 1 || v.W() != 1 {
		t.Errorf("Expected {1, 1, 1, 1}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4Half(t *testing.T) {
	v := Vec4Half()
	if Float(v[0]) != Float(0.5) || Float(v[1]) != Float(0.5) || Float(v[2]) != Float(0.5) || Float(v[3]) != Float(0.5) {
		t.Errorf("Expected {0.5, 0.5, 0.5, 0.5}, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

func TestVec4Largest(t *testing.T) {
	v := Vec4Largest()
	if v.X() != FloatMax || v.Y() != FloatMax || v.Z() != FloatMax || v.W() != FloatMax {
		t.Errorf("Expected all FloatMax, got (%f, %f, %f, %f)", v.X(), v.Y(), v.Z(), v.W())
	}
}

// ---- LargestAxis / LargestAxisDelta ----

func TestVec4LargestAxis(t *testing.T) {
	v := Vec4{1, 5, 3, 2}
	if v.LargestAxis() != 5 {
		t.Errorf("Expected LargestAxis = 5, got %f", v.LargestAxis())
	}
}

func TestVec4LargestAxisDelta(t *testing.T) {
	v := Vec4{1, 5, 3, 2}
	delta := v.LargestAxisDelta()
	// min=1, max=5, Abs(5) > Abs(1), so returns 5
	if Float(delta) != 5 {
		t.Errorf("Expected LargestAxisDelta = 5, got %f", delta)
	}
}

func TestVec4LargestAxisDeltaNegative(t *testing.T) {
	v := Vec4{1, -10, 3, 2}
	delta := v.LargestAxisDelta()
	// min=-10, max=3, Abs(-10) > Abs(3), so returns -10
	if Float(delta) != -10 {
		t.Errorf("Expected LargestAxisDelta = -10, got %f", delta)
	}
}

// ---- Vec4Area ----

func TestVec4Area(t *testing.T) {
	result := Vec4Area(10, 20, 5, 15)
	// min(10,5)=5, min(20,15)=15, max(10,5)=10, max(20,15)=20
	if result.X() != 5 || result.Y() != 15 || result.Z() != 10 || result.W() != 20 {
		t.Errorf("Expected {5, 15, 10, 20}, got (%f, %f, %f, %f)",
			Float(result.X()), Float(result.Y()), Float(result.Z()), Float(result.W()))
	}
}

// ---- BoxContains ----

func TestVec4BoxContains(t *testing.T) {
	// Vec4{Left, Top, Width, Height}
	v := Vec4{0, 0, 10, 10}
	if !v.BoxContains(5, 5) {
		t.Errorf("Expected BoxContains(5,5) = true for Vec4{0,0,10,10}")
	}
}

func TestVec4BoxContainsOutside(t *testing.T) {
	v := Vec4{0, 0, 10, 10}
	if v.BoxContains(15, 5) {
		t.Errorf("Expected BoxContains(15,5) = false for Vec4{0,0,10,10}")
	}
}

func TestVec4BoxContainsEdge(t *testing.T) {
	v := Vec4{0, 0, 10, 10}
	if !v.BoxContains(0, 0) {
		t.Errorf("Expected BoxContains(0,0) = true (edge)")
	}
	if !v.BoxContains(10, 10) {
		t.Errorf("Expected BoxContains(10,10) = true (edge)")
	}
}

// ---- AreaContains ----

func TestVec4AreaContains(t *testing.T) {
	// Vec4{X, Y, Right, Bottom}
	v := Vec4{0, 0, 10, 10}
	if !v.AreaContains(5, 5) {
		t.Errorf("Expected AreaContains(5,5) = true for Vec4{0,0,10,10}")
	}
}

func TestVec4AreaContainsOutside(t *testing.T) {
	v := Vec4{0, 0, 10, 10}
	if v.AreaContains(15, 5) {
		t.Errorf("Expected AreaContains(15,5) = false for Vec4{0,0,10,10}")
	}
}

func TestVec4AreaContainsEdge(t *testing.T) {
	v := Vec4{0, 0, 10, 10}
	if !v.AreaContains(0, 0) {
		t.Errorf("Expected AreaContains(0,0) = true (edge)")
	}
	if !v.AreaContains(10, 10) {
		t.Errorf("Expected AreaContains(10,10) = true (edge)")
	}
}

// ---- ScreenAreaContains ----

func TestVec4ScreenAreaContains(t *testing.T) {
	v := Vec4{0, 0, 100, 100}
	if !v.ScreenAreaContains(50, 50) {
		t.Errorf("Expected ScreenAreaContains(50,50) = true")
	}
}

func TestVec4ScreenAreaContainsOutside(t *testing.T) {
	v := Vec4{0, 0, 100, 100}
	if v.ScreenAreaContains(150, 50) {
		t.Errorf("Expected ScreenAreaContains(150,50) = false")
	}
}

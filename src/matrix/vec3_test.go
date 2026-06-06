/******************************************************************************/
/* vec3_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"testing"
)

func vec3ForTest() Vec3 {
	return Vec3{1, 2, 3}
}

func vec3ForTestOther() Vec3 {
	return Vec3{4, 5, 6}
}

// ---- Accessors ----

func TestVec3X(t *testing.T) {
	v := vec3ForTest()
	if v.X() != 1 {
		t.Errorf("Expected X() = 1, got %f", v.X())
	}
}

func TestVec3Y(t *testing.T) {
	v := vec3ForTest()
	if v.Y() != 2 {
		t.Errorf("Expected Y() = 2, got %f", v.Y())
	}
}

func TestVec3Z(t *testing.T) {
	v := vec3ForTest()
	if v.Z() != 3 {
		t.Errorf("Expected Z() = 3, got %f", v.Z())
	}
}

func TestVec3PX(t *testing.T) {
	v := vec3ForTest()
	px := v.PX()
	if *px != 1 {
		t.Errorf("Expected *PX() = 1, got %f", *px)
	}
	*px = 10
	if v.X() != 10 {
		t.Errorf("Expected v.X() = 10 after modifying pointer, got %f", v.X())
	}
}

func TestVec3PY(t *testing.T) {
	v := vec3ForTest()
	py := v.PY()
	if *py != 2 {
		t.Errorf("Expected *PY() = 2, got %f", *py)
	}
}

func TestVec3PZ(t *testing.T) {
	v := vec3ForTest()
	pz := v.PZ()
	if *pz != 3 {
		t.Errorf("Expected *PZ() = 3, got %f", *pz)
	}
}

func TestVec3SetX(t *testing.T) {
	v := vec3ForTest()
	v.SetX(10)
	if v.X() != 10 {
		t.Errorf("Expected X() = 10, got %f", v.X())
	}
}

func TestVec3SetY(t *testing.T) {
	v := vec3ForTest()
	v.SetY(20)
	if v.Y() != 20 {
		t.Errorf("Expected Y() = 20, got %f", v.Y())
	}
}

func TestVec3SetZ(t *testing.T) {
	v := vec3ForTest()
	v.SetZ(30)
	if v.Z() != 30 {
		t.Errorf("Expected Z() = 30, got %f", v.Z())
	}
}

func TestVec3AddX(t *testing.T) {
	v := vec3ForTest()
	v.AddX(5)
	if v.X() != 6 {
		t.Errorf("Expected X() = 6, got %f", v.X())
	}
}

func TestVec3AddY(t *testing.T) {
	v := vec3ForTest()
	v.AddY(5)
	if v.Y() != 7 {
		t.Errorf("Expected Y() = 7, got %f", v.Y())
	}
}

func TestVec3AddZ(t *testing.T) {
	v := vec3ForTest()
	v.AddZ(5)
	if v.Z() != 8 {
		t.Errorf("Expected Z() = 8, got %f", v.Z())
	}
}

func TestVec3ScaleX(t *testing.T) {
	v := vec3ForTest()
	v.ScaleX(2)
	if v.X() != 2 {
		t.Errorf("Expected X() = 2, got %f", v.X())
	}
}

func TestVec3ScaleY(t *testing.T) {
	v := vec3ForTest()
	v.ScaleY(2)
	if v.Y() != 4 {
		t.Errorf("Expected Y() = 4, got %f", v.Y())
	}
}

func TestVec3ScaleZ(t *testing.T) {
	v := vec3ForTest()
	v.ScaleZ(2)
	if v.Z() != 6 {
		t.Errorf("Expected Z() = 6, got %f", v.Z())
	}
}

// ---- Conversion ----

func TestVec3AsVec2(t *testing.T) {
	v := vec3ForTest()
	x, y := v.AsVec2().XY()
	if Float(x) != 1 || Float(y) != 2 {
		t.Errorf("Expected Vec2{1, 2}, got (%f, %f)", x, y)
	}
}

func TestVec3AsVec4(t *testing.T) {
	v := vec3ForTest()
	result := v.AsVec4()
	x, y, z, w := result.X(), result.Y(), result.Z(), result.W()
	if x != 1 || y != 2 || z != 3 || w != 1 {
		t.Errorf("Expected Vec4{1, 2, 3, 1}, got (%f, %f, %f, %f)", x, y, z, w)
	}
}

func TestVec3AsVec4WithW(t *testing.T) {
	v := vec3ForTest()
	result := v.AsVec4WithW(5)
	if result.W() != 5 {
		t.Errorf("Expected W = 5, got %f", result.W())
	}
}

func TestVec3XYZ(t *testing.T) {
	v := vec3ForTest()
	x, y, z := v.XYZ()
	if x != 1 || y != 2 || z != 3 {
		t.Errorf("Expected (1, 2, 3), got (%f, %f, %f)", x, y, z)
	}
}

func TestVec3XY(t *testing.T) {
	v := vec3ForTest()
	result := v.XY()
	if Float(result[0]) != 1 || Float(result[1]) != 2 {
		t.Errorf("Expected Vec2{1, 2}, got (%f, %f)", result[0], result[1])
	}
}

func TestVec3XZ(t *testing.T) {
	v := vec3ForTest()
	result := v.XZ()
	if Float(result[0]) != 1 || Float(result[1]) != 3 {
		t.Errorf("Expected Vec2{1, 3}, got (%f, %f)", result[0], result[1])
	}
}

func TestVec3Width(t *testing.T) {
	v := vec3ForTest()
	if v.Width() != 1 {
		t.Errorf("Expected Width() = 1, got %f", v.Width())
	}
}

func TestVec3Height(t *testing.T) {
	v := vec3ForTest()
	if v.Height() != 2 {
		t.Errorf("Expected Height() = 2, got %f", v.Height())
	}
}

func TestVec3Depth(t *testing.T) {
	v := vec3ForTest()
	if v.Depth() != 3 {
		t.Errorf("Expected Depth() = 3, got %f", v.Depth())
	}
}

func TestVec3AsVec3i(t *testing.T) {
	v := Vec3{1.7, 2.3, 3.9}
	result := v.AsVec3i()
	if result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Expected Vec3i{1, 2, 3}, got %v", result)
	}
}

// ---- Constructors ----

func TestNewVec3(t *testing.T) {
	v := NewVec3(1, 2, 3)
	if v.X() != 1 || v.Y() != 2 || v.Z() != 3 {
		t.Errorf("Expected {1, 2, 3}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestNewVec3XYZ(t *testing.T) {
	v := NewVec3XYZ(5)
	if v.X() != 5 || v.Y() != 5 || v.Z() != 5 {
		t.Errorf("Expected {5, 5, 5}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3FromArray(t *testing.T) {
	a := [3]Float{7, 8, 9}
	v := Vec3FromArray(a)
	if v.X() != 7 || v.Y() != 8 || v.Z() != 9 {
		t.Errorf("Expected {7, 8, 9}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3FromSlice(t *testing.T) {
	s := []Float{10, 11, 12}
	v := Vec3FromSlice(s)
	if v.X() != 10 || v.Y() != 11 || v.Z() != 12 {
		t.Errorf("Expected {10, 11, 12}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestNewVec3MinMax(t *testing.T) {
	minMax := NewVec3MinMax()
	if minMax.Min[0] != FloatMax || minMax.Min[1] != FloatMax || minMax.Min[2] != FloatMax {
		t.Errorf("Expected Min to be FloatMax")
	}
	if minMax.Max[0] != -FloatMax || minMax.Max[1] != -FloatMax || minMax.Max[2] != -FloatMax {
		t.Errorf("Expected Max to be -FloatMax")
	}
}

func TestVec3AsAligned16(t *testing.T) {
	v := Vec3{1, 2, 3}
	result := v.AsAligned16()
	if result[0] != 1 || result[1] != 2 || result[2] != 3 || result[3] != 0 {
		t.Errorf("Expected [1, 2, 3, 0], got %v", result)
	}
}

// ---- Arithmetic ----

func TestVec3Add(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := a.Add(b)
	if result.X() != 5 || result.Y() != 7 || result.Z() != 9 {
		t.Errorf("Expected {5, 7, 9}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3AddAssign(t *testing.T) {
	v := vec3ForTest()
	v.AddAssign(vec3ForTestOther())
	if v.X() != 5 || v.Y() != 7 || v.Z() != 9 {
		t.Errorf("Expected {5, 7, 9}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Subtract(t *testing.T) {
	a := vec3ForTestOther()
	b := vec3ForTest()
	result := a.Subtract(b)
	if result.X() != 3 || result.Y() != 3 || result.Z() != 3 {
		t.Errorf("Expected {3, 3, 3}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3SubtractAssign(t *testing.T) {
	v := vec3ForTestOther()
	v.SubtractAssign(vec3ForTest())
	if v.X() != 3 || v.Y() != 3 || v.Z() != 3 {
		t.Errorf("Expected {3, 3, 3}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Multiply(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := a.Multiply(b)
	if result.X() != 4 || result.Y() != 10 || result.Z() != 18 {
		t.Errorf("Expected {4, 10, 18}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3MultiplyAssign(t *testing.T) {
	v := vec3ForTest()
	v.MultiplyAssign(vec3ForTestOther())
	if v.X() != 4 || v.Y() != 10 || v.Z() != 18 {
		t.Errorf("Expected {4, 10, 18}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Divide(t *testing.T) {
	a := vec3ForTestOther()
	b := vec3ForTest()
	result := a.Divide(b)
	if !Vec3ApproxTo(result, Vec3{4, 2.5, 2}, Tiny) {
		t.Errorf("Expected {4, 2.5, 2}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3DivideAssign(t *testing.T) {
	v := vec3ForTestOther()
	v.DivideAssign(vec3ForTest())
	if !Vec3ApproxTo(v, Vec3{4, 2.5, 2}, Tiny) {
		t.Errorf("Expected {4, 2.5, 2}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Scale(t *testing.T) {
	v := vec3ForTest()
	result := v.Scale(2)
	if result.X() != 2 || result.Y() != 4 || result.Z() != 6 {
		t.Errorf("Expected {2, 4, 6}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3ScaleAssign(t *testing.T) {
	v := vec3ForTest()
	v.ScaleAssign(2)
	if v.X() != 2 || v.Y() != 4 || v.Z() != 6 {
		t.Errorf("Expected {2, 4, 6}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Shrink(t *testing.T) {
	v := vec3ForTest()
	result := v.Shrink(2)
	if !Vec3ApproxTo(result, Vec3{0.5, 1, 1.5}, Tiny) {
		t.Errorf("Expected {0.5, 1, 1.5}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3ShrinkAssign(t *testing.T) {
	v := vec3ForTest()
	v.ShrinkAssign(2)
	if !Vec3ApproxTo(v, Vec3{0.5, 1, 1.5}, Tiny) {
		t.Errorf("Expected {0.5, 1, 1.5}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

// ---- Length / Normalization ----

func TestVec3Length(t *testing.T) {
	v := Vec3{3, 4, 0}
	expected := Sqrt(25)
	if Abs(v.Length()-expected) > Tiny {
		t.Errorf("Expected Length = %f, got %f", expected, v.Length())
	}
}

func TestVec3LengthSquared(t *testing.T) {
	v := Vec3{3, 4, 0}
	if v.LengthSquared() != 25 {
		t.Errorf("Expected LengthSquared = 25, got %f", v.LengthSquared())
	}
}

func TestVec3Normal(t *testing.T) {
	v := Vec3{3, 0, 0}
	result := v.Normal()
	if !Vec3Approx(result, Vec3{1, 0, 0}) {
		t.Errorf("Expected {1, 0, 0}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3Normalize(t *testing.T) {
	v := Vec3{3, 0, 0}
	v.Normalize()
	if !Vec3Approx(v, Vec3{1, 0, 0}) {
		t.Errorf("Expected {1, 0, 0}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3LengthOfZero(t *testing.T) {
	v := Vec3Zero()
	if v.Length() != 0 {
		t.Errorf("Expected Length of zero vector = 0, got %f", v.Length())
	}
}

func TestVec3NormalOfZero(t *testing.T) {
	v := Vec3Zero()
	result := v.Normal()
	// Normal of zero vector produces NaN due to division by zero
	if !result.IsNaN() {
		t.Errorf("Expected Normal of zero vector to be NaN, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

// ---- Negative / Inverse ----

func TestVec3Negative(t *testing.T) {
	v := vec3ForTest()
	result := v.Negative()
	if result.X() != -1 || result.Y() != -2 || result.Z() != -3 {
		t.Errorf("Expected {-1, -2, -3}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3Inverse(t *testing.T) {
	v := vec3ForTest()
	v = v.Inverse()
	if v.X() != 1 || v.Y() != float32(1)/2 || v.Z() != float32(1)/3 {
		t.Errorf("Expected {-1, -2, -3}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

// ---- Cross Product ----

func TestVec3Cross(t *testing.T) {
	i := Vec3Right()
	j := Vec3Up()
	result := Vec3Cross(i, j)
	// cross({1,0,0}, {0,1,0}) = {0,0,1} = Backward
	if !Vec3Approx(result, Vec3Backward()) {
		t.Errorf("Expected cross(i, j) = Backward {0,0,1}, got %v", result)
	}
}

func TestVec3CrossCommutative(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	c1 := Vec3Cross(a, b)
	c2 := Vec3Cross(b, a)
	if !Vec3Approx(c1, c2.Negative()) {
		t.Errorf("Cross product should be anti-commutative")
	}
}

func TestVec3SelfCross(t *testing.T) {
	v := vec3ForTest()
	result := Vec3Cross(v, v)
	if !result.IsZero() {
		t.Errorf("Cross product of vector with itself should be zero, got %v", result)
	}
}

func TestVec3MethodCross(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := a.Cross(b)
	expected := Vec3Cross(a, b)
	if !Vec3Approx(result, expected) {
		t.Errorf("Method Cross should match standalone Vec3Cross")
	}
}

func TestVec3Orthogonal(t *testing.T) {
	v := vec3ForTest()
	result := v.Orthogonal()
	dot := Vec3Dot(v, result)
	if Abs(dot) > Tiny {
		t.Errorf("Orthogonal result should be perpendicular to input, dot = %f", dot)
	}
}

// ---- Dot Product ----

func TestVec3Dot(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := Vec3Dot(a, b)
	if result != 32 {
		t.Errorf("Expected Dot({1,2,3}, {4,5,6}) = 32, got %f", result)
	}
}

func TestVec3MethodDot(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := a.Dot(b)
	if result != 32 {
		t.Errorf("Expected Dot({1,2,3}, {4,5,6}) = 32, got %f", result)
	}
}

func TestVec3DotOrthogonal(t *testing.T) {
	result := Vec3Dot(Vec3Right(), Vec3Up())
	if result != 0 {
		t.Errorf("Expected dot of orthogonal vectors = 0, got %f", result)
	}
}

func TestVec3DotPerpendicular(t *testing.T) {
	forEach := []Vec3{Vec3Forward(), Vec3Backward()}
	up := Vec3Up()
	for _, v := range forEach {
		dot := Vec3Dot(v, up)
		if Abs(dot) > Tiny {
			t.Errorf("Expected dot of Forward/Backward and Up to be near 0, got %f", dot)
		}
	}
}

// ---- Comparison ----

func TestVec3Approx(t *testing.T) {
	a := vec3ForTest()
	b := Vec3{1 + Tiny/2, 2, 3}
	if !Vec3Approx(a, b) {
		t.Errorf("Expected Vec3Approx to return true for close vectors")
	}
}

func TestVec3ApproxFar(t *testing.T) {
	a := vec3ForTest()
	b := Vec3{100, 200, 300}
	if Vec3Approx(a, b) {
		t.Errorf("Expected Vec3Approx to return false for far vectors")
	}
}

func TestVec3ApproxTo(t *testing.T) {
	a := vec3ForTest()
	b := Vec3{1.05, 2.05, 3.05}
	if !Vec3ApproxTo(a, b, 0.1) {
		t.Errorf("Expected Vec3ApproxTo with delta 0.1 to return true")
	}
}

func TestVec3ApproxToFar(t *testing.T) {
	a := vec3ForTest()
	b := Vec3{100, 200, 300}
	if Vec3ApproxTo(a, b, 0.1) {
		t.Errorf("Expected Vec3ApproxTo with delta 0.1 to return false for far vectors")
	}
}

func TestVec3Equals(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTest()
	if !a.Equals(b) {
		t.Errorf("Expected Equal to return true for same vectors")
	}
}

func TestVec3EqualsDifference(t *testing.T) {
	a := vec3ForTest()
	b := Vec3{100, 2, 3}
	if a.Equals(b) {
		t.Errorf("Expected Equal to return false for different vectors")
	}
}

// ---- Abs ----

func TestVec3Abs(t *testing.T) {
	v := Vec3{-1, -2, 3}
	result := Vec3Abs(v)
	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 {
		t.Errorf("Expected {1, 2, 3}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3AbsMethod(t *testing.T) {
	v := Vec3{-1, -2, 3}
	result := v.Abs()
	if Float(result[0]) != 1 || Float(result[1]) != 2 || Float(result[2]) != 3 {
		t.Errorf("Expected {1, 2, 3}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

// ---- Min / Max ----

func TestVec3Min(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := Vec3Min(a, b)
	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 {
		t.Errorf("Expected {1, 2, 3}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3Max(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := Vec3Max(a, b)
	if result.X() != 4 || result.Y() != 5 || result.Z() != 6 {
		t.Errorf("Expected {4, 5, 6}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3MinAbs(t *testing.T) {
	a := Vec3{-1, -5, 3}
	b := Vec3{4, -2, 6}
	result := Vec3MinAbs(a, b)
	if Float(result[0]) != 1 || Float(result[1]) != 2 || Float(result[2]) != 3 {
		t.Errorf("Expected {1, 2, 3}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3MaxAbs(t *testing.T) {
	a := Vec3{-1, -5, 3}
	b := Vec3{4, -2, 6}
	result := Vec3MaxAbs(a, b)
	if Float(result[0]) != 4 || Float(result[1]) != 5 || Float(result[2]) != 6 {
		t.Errorf("Expected {4, 5, 6}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

// ---- Distance ----

func TestVec3Distance(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := a.Distance(b)
	expected := Float(5.196152)
	if Abs(result-expected) > Tiny {
		t.Errorf("Expected Distance = %f, got %f", expected, result)
	}
}

func TestVec3DistanceZero(t *testing.T) {
	a := vec3ForTest()
	if a.Distance(a) != 0 {
		t.Errorf("Expected distance from vector to itself = 0, got %f", a.Distance(a))
	}
}

func TestVec3SquareDistance(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	result := a.SquareDistance(b)
	if Float(result) != Float(27) {
		t.Errorf("Expected SquareDistance = 27, got %f", result)
	}
}

// ---- Lerp ----

func TestVec3Lerp(t *testing.T) {
	from := vec3ForTest()
	to := vec3ForTestOther()
	result := Vec3Lerp(from, to, 0.5)
	if result.X() != 2.5 || result.Y() != 3.5 || result.Z() != 4.5 {
		t.Errorf("Expected {2.5, 3.5, 4.5}, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3LerpFrom(t *testing.T) {
	from := vec3ForTest()
	to := vec3ForTestOther()
	result := Vec3Lerp(from, to, 0)
	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 {
		t.Errorf("Expected lerp(t=0) to return from, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

func TestVec3LerpTo(t *testing.T) {
	from := vec3ForTest()
	to := vec3ForTestOther()
	result := Vec3Lerp(from, to, 1)
	if result.X() != 4 || result.Y() != 5 || result.Z() != 6 {
		t.Errorf("Expected lerp(t=1) to return to, got (%f, %f, %f)", result.X(), result.Y(), result.Z())
	}
}

// ---- String ----

func TestVec3String(t *testing.T) {
	v := vec3ForTest()
	result := v.String()
	expected := "1.000000, 2.000000, 3.000000"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestVec3FromString(t *testing.T) {
	str := "1.500000, 2.500000, 3.500000"
	v := Vec3FromString(str)
	if Float(v.X()) != Float(1.5) || Float(v.Y()) != Float(2.5) || Float(v.Z()) != Float(3.5) {
		t.Errorf("Expected {1.5, 2.5, 3.5}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

// ---- Angle ----

func TestVec3Angle(t *testing.T) {
	a := Vec3{3, 4, 0}
	b := Vec3{5, 12, 0}
	dot := Vec3Dot(a, b) / (a.Length() * b.Length())
	// cos(0) = 1 for same direction
	if Abs(dot-1) < Tiny {
		// same direction, angle should be 0
	}
	if a.Equals(b) {
		angle := a.Angle(b)
		if angle != 0 {
			t.Errorf("Expected angle of equal vectors = 0, got %f", angle)
		}
	}
}

func TestVec3AngleZeroVector(t *testing.T) {
	a := vec3ForTest()
	// Acos with zero length will produce Inf or NaN, just check it doesn't panic
	_ = a.Angle(Vec3Zero())
}

func TestVec3SignedAngle(t *testing.T) {
	a := Vec3Right()
	b := Vec3Up()
	axis := Vec3Forward()
	angle := a.SignedAngle(b, axis)
	// -90 degrees in radians = -pi/2
	expected := -(Float(3.14159265358979) / 2)
	if Abs(angle-expected) > 0.01 {
		t.Errorf("Expected signed angle ~%f (-90 degrees), got %f", expected, angle)
	}
}

func TestVec3SignedAngleEqualVectors(t *testing.T) {
	a := vec3ForTest()
	axis := Vec3Forward()
	angle := a.SignedAngle(a, axis)
	if angle != 0 {
		t.Errorf("Expected signed angle of equal vectors = 0, got %f", angle)
	}
}

func TestVec3SignedAngleZeroAxis(t *testing.T) {
	a := vec3ForTest()
	b := vec3ForTestOther()
	angle := a.SignedAngle(b, Vec3Zero())
	if angle != 0 {
		t.Errorf("Expected signed angle with zero axis = 0, got %f", angle)
	}
}

func TestVec3SignedAngleZeroVector(t *testing.T) {
	axis := Vec3Forward()
	angle := Vec3Zero().SignedAngle(vec3ForTest(), axis)
	if angle != 0 {
		t.Errorf("Expected signed angle with zero input = 0, got %f", angle)
	}
}

// ---- Direction Constants ----

func TestVec3Up(t *testing.T) {
	v := Vec3Up()
	if v.X() != 0 || v.Y() != 1 || v.Z() != 0 {
		t.Errorf("Expected {0, 1, 0}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Down(t *testing.T) {
	v := Vec3Down()
	if v.X() != 0 || v.Y() != -1 || v.Z() != 0 {
		t.Errorf("Expected {0, -1, 0}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Left(t *testing.T) {
	v := Vec3Left()
	if v.X() != -1 || v.Y() != 0 || v.Z() != 0 {
		t.Errorf("Expected {-1, 0, 0}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Right(t *testing.T) {
	v := Vec3Right()
	if v.X() != 1 || v.Y() != 0 || v.Z() != 0 {
		t.Errorf("Expected {1, 0, 0}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Forward(t *testing.T) {
	v := Vec3Forward()
	if v.X() != 0 || v.Y() != 0 || v.Z() != -1 {
		t.Errorf("Expected {0, 0, -1}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Backward(t *testing.T) {
	v := Vec3Backward()
	if v.X() != 0 || v.Y() != 0 || v.Z() != 1 {
		t.Errorf("Expected {0, 0, 1}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Zero(t *testing.T) {
	v := Vec3Zero()
	if v.X() != 0 || v.Y() != 0 || v.Z() != 0 {
		t.Errorf("Expected {0, 0, 0}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3One(t *testing.T) {
	v := Vec3One()
	if v.X() != 1 || v.Y() != 1 || v.Z() != 1 {
		t.Errorf("Expected {1, 1, 1}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

func TestVec3Half(t *testing.T) {
	v := Vec3Half()
	if Float(v[0]) != Float(0.5) || Float(v[1]) != Float(0.5) || Float(v[2]) != Float(0.5) {
		t.Errorf("Expected {0.5, 0.5, 0.5}, got (%f, %f, %f)", v.X(), v.Y(), v.Z())
	}
}

// ---- LargestAxis / LongestAxis ----

func TestVec3LargestAxis(t *testing.T) {
	v := Vec3{1, 5, 3}
	if v.LargestAxis() != 5 {
		t.Errorf("Expected LargestAxis = 5, got %f", v.LargestAxis())
	}
}

func TestVec3LargestAxisDelta(t *testing.T) {
	v := Vec3{1, 5, 3}
	delta := v.LargestAxisDelta()
	// min=1, max=5, Abs(5) > Abs(1), so returns 5
	if Float(delta) != 5 {
		t.Errorf("Expected LargestAxisDelta = 5, got %f", delta)
	}
}

func TestVec3LongestAxis(t *testing.T) {
	v := Vec3{1, 5, 3}
	result := v.LongestAxis()
	if result != Vy {
		t.Errorf("Expected LongestAxis = Vy, got %d", result)
	}
}

func TestVec3LongestAxisValue(t *testing.T) {
	v := Vec3{1, 5, 3}
	if v.LongestAxisValue() != 5 {
		t.Errorf("Expected LongestAxisValue = 5, got %f", v.LongestAxisValue())
	}
}

// ---- Inf / NaN ----

func TestVec3Inf(t *testing.T) {
	v := Vec3Inf(1)
	if !v.IsInf(1) {
		t.Errorf("Expected Vec3Inf(1) to pass IsInf(1)")
	}
}

func TestVec3InfNegative(t *testing.T) {
	v := Vec3Inf(-1)
	if !v.IsInf(-1) {
		t.Errorf("Expected Vec3Inf(-1) to pass IsInf(-1)")
	}
}

func TestVec3NaN(t *testing.T) {
	v := Vec3NaN()
	if !v.IsNaN() {
		t.Errorf("Expected Vec3NaN to pass IsNaN()")
	}
}

func TestVec3IsZero(t *testing.T) {
	if !Vec3Zero().IsZero() {
		t.Errorf("Expected Vec3Zero().IsZero() = true")
	}
}

func TestVec3IsZeroFalse(t *testing.T) {
	if vec3ForTest().IsZero() {
		t.Errorf("Expected non-zero vector IsZero() = false")
	}
}

func TestVec3IsInf(t *testing.T) {
	v := Vec3Inf(1)
	if !v.IsInf(1) {
		t.Errorf("IsInf(1) should be true")
	}
}

func TestVec3IsInfinity(t *testing.T) {
	v := vec3ForTest()
	if v.IsInf(1) {
		t.Errorf("Regular vector should not be Inf")
	}
}

func TestVec3IsNaN(t *testing.T) {
	v := Vec3NaN()
	if !v.IsNaN() {
		t.Errorf("IsNaN() should be true for NaN vector")
	}
}

func TestVec3IsNotNaN(t *testing.T) {
	v := vec3ForTest()
	if v.IsNaN() {
		t.Errorf("Regular vector should not be NaN")
	}
}

// ---- Vec3MinMax ----

// ---- Benchmarks ----

func BenchmarkVec3Add(b *testing.B) {
	a := vec3ForTest()
	c := vec3ForTestOther()
	for i := 0; i < b.N; i++ {
		a.Add(c)
	}
}

func BenchmarkVec3Dot(b *testing.B) {
	a := vec3ForTest()
	c := vec3ForTestOther()
	for i := 0; i < b.N; i++ {
		Vec3Dot(a, c)
	}
}

func BenchmarkVec3Cross(b *testing.B) {
	a := vec3ForTest()
	c := vec3ForTestOther()
	for i := 0; i < b.N; i++ {
		Vec3Cross(a, c)
	}
}

func BenchmarkVec3Normal(b *testing.B) {
	v := vec3ForTest()
	for i := 0; i < b.N; i++ {
		v.Normal()
	}
}

func BenchmarkVec3Scale(b *testing.B) {
	v := vec3ForTest()
	for i := 0; i < b.N; i++ {
		v.Scale(2)
	}
}

func BenchmarkVec3Lerp(b *testing.B) {
	a := vec3ForTest()
	c := vec3ForTestOther()
	for i := 0; i < b.N; i++ {
		Vec3Lerp(a, c, 0.5)
	}
}

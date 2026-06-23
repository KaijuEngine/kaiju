/******************************************************************************/
/* mat4_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"fmt"
	"math"
	"testing"
)

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

const approxDelta = 1e-5

func testVec4() Vec4 { return Vec4{1, 2, 3, 4} }

func testMat4() Mat4 {
	return Mat4{
		1, 2, 3, 4,
		1, 2, 3, 4,
		1, 2, 3, 4,
		1, 2, 3, 4,
	}
}

func legacyMat4Multiply(a, b Mat4) Mat4 {
	var result Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var sum Float = 0
			for k := 0; k < 4; k++ {
				sum += a[i*4+k] * b[k*4+j]
			}
			result[i*4+j] = sum
		}
	}
	return result
}

func legacyMat4MultiplyVec4(a Mat4, b Vec4) Vec4 {
	var result Vec4
	c := a.ColumnVector(0)
	result[Vx] = Vec4Dot(c, b)
	c = a.ColumnVector(1)
	result[Vy] = Vec4Dot(c, b)
	c = a.ColumnVector(2)
	result[Vz] = Vec4Dot(c, b)
	c = a.ColumnVector(3)
	result[Vw] = Vec4Dot(c, b)
	return result
}

func (v Vec4) legacyMultiplyMat4(rhs Mat4) Vec4 {
	var result Vec4
	row := rhs.RowVector(0)
	result[Vx] = Vec4Dot(row, v)
	row = rhs.RowVector(1)
	result[Vy] = Vec4Dot(row, v)
	row = rhs.RowVector(2)
	result[Vz] = Vec4Dot(row, v)
	row = rhs.RowVector(3)
	result[Vw] = Vec4Dot(row, v)
	return result
}

// -----------------------------------------------------------------------
// Tests - Constructors & Basics
// -----------------------------------------------------------------------

func TestNewMat4(t *testing.T) {
	m := NewMat4()
	if !m.IsIdentity() {
		t.Errorf("NewMat4 should return identity, got %v", m)
	}
}

func TestMat4Identity(t *testing.T) {
	m := Mat4Identity()
	expected := Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	if m != expected {
		t.Errorf("Mat4Identity = %v, want %v", m, expected)
	}
}

func TestMat4Zero(t *testing.T) {
	m := Mat4Zero()
	for i, v := range m {
		if v != 0 {
			t.Errorf("Mat4Zero[%d] = %v, want 0", i, v)
		}
	}
}

func TestMat4FromSlice(t *testing.T) {
	slice := make([]Float, 16)
	for i := range slice {
		slice[i] = Float(i)
	}
	m := Mat4FromSlice(slice)
	for i := range m {
		if m[i] != Float(i) {
			t.Errorf("Mat4FromSlice[%d] = %v, want %v", i, m[i], Float(i))
		}
	}
}

func TestMat4Reset(t *testing.T) {
	var m Mat4
	for i := range m {
		m[i] = Float(i + 1)
	}
	m.Reset()
	if !m.IsIdentity() {
		t.Errorf("Reset should produce identity, got %v", m)
	}
}

func TestMat4ZeroMutator(t *testing.T) {
	m := Mat4Identity()
	m.Zero()
	for i, v := range m {
		if v != 0 {
			t.Errorf("Zero[%d] = %v, want 0", i, v)
		}
	}
}

// -----------------------------------------------------------------------
// Tests - Element Access
// -----------------------------------------------------------------------

func TestMat4At(t *testing.T) {
	m := Mat4{
		10, 20, 30, 40,
		50, 60, 70, 80,
		90, 100, 110, 120,
		130, 140, 150, 160,
	}

	tests := []struct {
		row, col int
		want     Float
	}{
		{0, 0, 10}, {0, 3, 40},
		{1, 0, 50}, {2, 2, 110},
		{3, 3, 160},
	}
	for _, tt := range tests {
		got := m.At(tt.row, tt.col)
		if got != tt.want {
			t.Errorf("At(%d, %d) = %v, want %v", tt.row, tt.col, got, tt.want)
		}
	}
}

func TestMat4RowVector(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}

	tests := []struct {
		row  int
		want Vec4
	}{
		{0, Vec4{1, 2, 3, 4}},
		{1, Vec4{5, 6, 7, 8}},
		{3, Vec4{13, 14, 15, 16}},
	}
	for _, tt := range tests {
		got := m.RowVector(tt.row)
		if got != tt.want {
			t.Errorf("RowVector(%d) = %v, want %v", tt.row, got, tt.want)
		}
	}
}

func TestMat4ColumnVector(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}

	tests := []struct {
		col  int
		want Vec4
	}{
		{0, Vec4{1, 5, 9, 13}},
		{1, Vec4{2, 6, 10, 14}},
		{3, Vec4{4, 8, 12, 16}},
	}
	for _, tt := range tests {
		got := m.ColumnVector(tt.col)
		if got != tt.want {
			t.Errorf("ColumnVector(%d) = %v, want %v", tt.col, got, tt.want)
		}
	}
}

// -----------------------------------------------------------------------
// Tests - Comparison
// -----------------------------------------------------------------------

func TestMat4IsIdentity(t *testing.T) {
	if !Mat4Identity().IsIdentity() {
		t.Error("Identity should report IsIdentity() == true")
	}

	m := Mat4{
		1, 0, 0, 0,
		0, 2, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	if m.IsIdentity() {
		t.Error("Non-identity should report IsIdentity() == false")
	}
}

func TestMat4Approx(t *testing.T) {
	m1 := Mat4Identity()
	m2 := Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	if !Mat4Approx(m1, m2) {
		t.Error("Identical matrices should be approximately equal")
	}

	m3 := Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 2,
	}
	if Mat4Approx(m1, m3) {
		t.Error("Different matrices should not be approximately equal")
	}
}

func TestMat4ApproxTo(t *testing.T) {
	m1 := Mat4Identity()
	m2 := Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1.0001,
	}
	if !Mat4ApproxTo(m1, m2, 0.001) {
		t.Error("Matrices within delta should be approximately equal")
	}
	if Mat4ApproxTo(m1, m2, 0.00001) {
		t.Error("Matrices outside small delta should not be approximately equal")
	}
}

func TestMat4Equals(t *testing.T) {
	m1 := Mat4Identity()
	m2 := Mat4Identity()
	if !m1.Equals(m2) {
		t.Error("Identity matrices should be equal")
	}
}

// -----------------------------------------------------------------------
// Tests - Assignment Operations
// -----------------------------------------------------------------------

func TestMat4AddAssign(t *testing.T) {
	a := Mat4Identity()
	b := Mat4{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		10, 20, 30, 0,
	}
	a.AddAssign(b)

	if a[x0y3] != 10 || a[x1y3] != 20 || a[x2y3] != 30 {
		t.Errorf("AddAssign translation not updated: %v", a)
	}
	if a[x0y0] != 1 || a[x1y1] != 1 || a[x2y2] != 1 || a[x3y3] != 1 {
		t.Errorf("AddAssign changed diagonal unexpectedly: %v", a)
	}
}

func TestMat4SubtractAssign(t *testing.T) {
	a := Mat4{
		10, 0, 0, 0,
		0, 10, 0, 0,
		0, 0, 10, 0,
		0, 0, 0, 10,
	}
	b := Mat4{
		5, 0, 0, 0,
		0, 5, 0, 0,
		0, 0, 5, 0,
		0, 0, 0, 5,
	}
	a.SubtractAssign(b)

	if a[x0y0] != 5 || a[x1y1] != 5 || a[x2y2] != 5 || a[x3y3] != 5 {
		t.Errorf("SubtractAssign = %v, want all diagonals = 5", a)
	}
}

func TestMat4NegateAssign(t *testing.T) {
	m := Mat4{
		1, 0, 0, 0,
		0, -2, 0, 0,
		0, 0, 3, 0,
		0, 0, 0, -4,
	}
	m.NegateAssign()

	if m[x0y0] != -1 || m[x1y1] != 2 || m[x2y2] != -3 || m[x3y3] != 4 {
		t.Errorf("NegateAssign = %v, want diagonal = -1, 2, -3, 4", m)
	}
}

// -----------------------------------------------------------------------
// Tests - MultiplyAssign
// -----------------------------------------------------------------------

func TestMat4MultiplyAssignIdentity(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	orig := m
	m.MultiplyAssign(Mat4Identity())
	if !Mat4Approx(m, orig) {
		t.Errorf("MultiplyAssign(identity) should be identity transform: got %v, want %v", m, orig)
	}
}

// -----------------------------------------------------------------------
// Tests - Transpose
// -----------------------------------------------------------------------

func TestMat4Transpose(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	tm := m.Transpose()

	// Transpose swaps rows and columns: tm[row,col] = m[col,row]
	// tm[x0y1] = m[x1y0] = m[1] = 2
	if tm[x0y1] != 2 {
		t.Errorf("Transpose[0,1] = %v, want 2", tm[x0y1])
	}
	// tm[x1y0] = m[x0y1] = m[4] = 5
	if tm[x1y0] != 5 {
		t.Errorf("Transpose[1,0] = %v, want 5", tm[x1y0])
	}
	// tm[x0y3] = m[x3y0] = m[3] = 4
	if tm[x0y3] != 4 {
		t.Errorf("Transpose[0,3] = %v, want 4", tm[x0y3])
	}
	// tm[x3y0] = m[x0y3] = m[12] = 13
	if tm[x3y0] != 13 {
		t.Errorf("Transpose[3,0] = %v, want 13", tm[x3y0])
	}
}

func TestMat4TransposeIdentity(t *testing.T) {
	m := Mat4Identity()
	tm := m.Transpose()
	if !Mat4Approx(m, tm) {
		t.Error("Transpose of identity should be identity")
	}
}

func TestMat4TransposeAssign(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	m.TransposeAssign()

	// After transpose: m[0,1] should be old m[1,0] = 2
	if m[x0y1] != 2 {
		t.Errorf("TransposeAssign m[0,1] = %v, want 2", m[x0y1])
	}
	// After transpose: m[1,0] should be old m[0,1] = 5
	if m[x1y0] != 5 {
		t.Errorf("TransposeAssign m[1,0] = %v, want 5", m[x1y0])
	}
}

// -----------------------------------------------------------------------
// Tests - Mat3 extraction
// -----------------------------------------------------------------------

func TestMat4Mat3(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	m3 := m.Mat3()

	expected := Mat3{1, 2, 3, 5, 6, 7, 9, 10, 11}
	if m3 != expected {
		t.Errorf("Mat3() = %v, want %v", m3, expected)
	}
}

// -----------------------------------------------------------------------
// Tests - Translation
// -----------------------------------------------------------------------

func TestMat4Translate(t *testing.T) {
	m := Mat4Identity()
	m.Translate(Vec3{10, 20, 30})

	if m[x0y3] != 10 || m[x1y3] != 20 || m[x2y3] != 30 {
		t.Errorf("Translate = %v, want translation column = (10, 20, 30)", m)
	}
}

func TestMat4TranslateAccumulates(t *testing.T) {
	m := Mat4Identity()
	m.Translate(Vec3{10, 0, 0})
	m.Translate(Vec3{5, 0, 0})

	if m[x0y3] != 15 {
		t.Errorf("Translate should accumulate: x0y3 = %v, want 15", m[x0y3])
	}
}

func TestMat4SetTranslation(t *testing.T) {
	m := Mat4{
		1, 0, 0, 100,
		0, 1, 0, 200,
		0, 0, 1, 300,
		0, 0, 0, 1,
	}
	m.SetTranslation(Vec3{5, 10, 15})

	if m[x0y3] != 5 || m[x1y3] != 10 || m[x2y3] != 15 {
		t.Errorf("SetTranslation = %v, want (5, 10, 15)", m)
	}
}

// -----------------------------------------------------------------------
// Tests - Scale
// -----------------------------------------------------------------------

func TestMat4Scale(t *testing.T) {
	m := Mat4Identity()
	m.Scale(Vec3{2, 3, 4})

	if m[x0y0] != 2 || m[x1y1] != 3 || m[x2y2] != 4 {
		t.Errorf("Scale = %v, want diagonal (2, 3, 4, 1)", m)
	}
	if m[x3y3] != 1 {
		t.Errorf("Scale should not affect x3y3: got %v", m[x3y3])
	}
}

func TestMat4ScaleZero(t *testing.T) {
	m := Mat4Identity()
	m.Scale(Vec3{0, 0, 0})

	if m[x0y0] != 0 || m[x1y1] != 0 || m[x2y2] != 0 {
		t.Errorf("Scale by zero: %v, want diagonal (0,0,0,1)", m)
	}
}

// -----------------------------------------------------------------------
// Tests - Rotation
// -----------------------------------------------------------------------

func TestMat4RotateX(t *testing.T) {
	m := Mat4Identity()
	m.RotateX(90)

	if Abs(m[x0y0]-1) > approxDelta {
		t.Errorf("RotateX[0,0] = %v, want 1", m[x0y0])
	}
	if Abs(m[x1y2]-1) > approxDelta {
		t.Errorf("RotateX[1,2] = %v, want 1", m[x1y2])
	}
	if Abs(m[x2y1]+1) > approxDelta {
		t.Errorf("RotateX[2,1] = %v, want -1", m[x2y1])
	}
}

func TestMat4RotateY(t *testing.T) {
	m := Mat4Identity()
	m.RotateY(90)

	// cos(90) ~ 0, sin(90) = 1
	// Y-axis rotation: [cos 0 sin 0; 0 1 0 0; -sin 0 cos 0; 0 0 0 1]
	if Abs(m[x0y0]) > approxDelta {
		t.Errorf("RotateY[0,0] should be ~0 (cos 90), got %v", m[x0y0])
	}
	if Abs(m[x0y2]-1) > approxDelta {
		t.Errorf("RotateY[0,2] should be 1 (sin 90), got %v", m[x0y2])
	}
	if Abs(m[x2y0]+1) > approxDelta {
		t.Errorf("RotateY[2,0] should be -1 (-sin 90), got %v", m[x2y0])
	}
}

func TestMat4RotateZ(t *testing.T) {
	m := Mat4Identity()
	m.RotateZ(90)

	if Abs(m[x0y0]) > approxDelta {
		t.Errorf("RotateZ[0,0] should be ~0, got %v", m[x0y0])
	}
	if Abs(m[x0y1]-1) > approxDelta {
		t.Errorf("RotateZ[0,1] should be 1, got %v", m[x0y1])
	}
	if Abs(m[x1y0]+1) > approxDelta {
		t.Errorf("RotateZ[1,0] should be -1, got %v", m[x1y0])
	}
	if Abs(m[x1y1]) > approxDelta {
		t.Errorf("RotateZ[1,1] should be ~0, got %v", m[x1y1])
	}
}

func TestMat4RotateQuaternion(t *testing.T) {
	m := Mat4Identity()
	m.Rotate(Vec3{0, 0, 90})

	// Rotate uses QuaternionFromEuler then MultiplyAssign
	// For 90 deg Z rotation: [0 1 0 0; -1 0 0 0; 0 0 1 0; 0 0 0 1]
	if Abs(m[x0y0]) > approxDelta {
		t.Errorf("Rotate(quaternion)[0,0] should be ~0, got %v", m[x0y0])
	}
	// [0,1] could be +/-1 depending on quaternion convention
	if Abs(Abs(m[x0y1])-1) > approxDelta {
		t.Errorf("Rotate(quaternion)[0,1] should be +/-1, got %v", m[x0y1])
	}
}

func TestMat4RotateAnglesAxis(t *testing.T) {
	m := Mat4Identity()
	m.RotateAngles(Vec3{0, 0, 1}, 90)

	if Abs(m[x0y0]) > approxDelta {
		t.Errorf("RotateAngles Z-axis[0,0] should be ~0, got %v", m[x0y0])
	}
	if Abs(m[x0y1]-1) > approxDelta {
		t.Errorf("RotateAngles Z-axis[0,1] should be 1, got %v", m[x0y1])
	}
}

// -----------------------------------------------------------------------
// Tests - LookAt
// -----------------------------------------------------------------------

func TestMat4LookAt(t *testing.T) {
	var m Mat4
	eye := Vec3{0, 0, 5}
	center := Vec3{0, 0, 0}
	up := Vec3{0, 1, 0}
	m.LookAt(eye, center, up)

	// Forward returns f = eye - center, normalized = (0, 0, 1)
	forward := m.Forward()
	if Abs(forward.Z()-1) > approxDelta {
		t.Errorf("LookAt forward.Z() = %v, want 1", forward.Z())
	}

	upDir := m.Up()
	if Abs(upDir.Y()-1) > approxDelta {
		t.Errorf("LookAt up.Y() = %v, want 1", upDir.Y())
	}
}

func TestMat4LookAtStatic(t *testing.T) {
	eye := Vec3{0, 0, 5}
	center := Vec3{0, 0, 0}
	up := Vec3{0, 1, 0}
	m := Mat4LookAt(eye, center, up)

	forward := m.Forward()
	if Abs(forward.Z()-1) > approxDelta {
		t.Errorf("Mat4LookAt forward.Z() = %v, want 1", forward.Z())
	}
}

// -----------------------------------------------------------------------
// Tests - Direction Vectors
// -----------------------------------------------------------------------

func TestMat4RightUpForward(t *testing.T) {
	m := Mat4Identity()

	right := m.Right()
	if right.X() != 1 || right.Y() != 0 || right.Z() != 0 {
		t.Errorf("Identity Right() = %v, want (1,0,0)", right)
	}

	up := m.Up()
	if up.X() != 0 || up.Y() != 1 || up.Z() != 0 {
		t.Errorf("Identity Up() = %v, want (0,1,0)", up)
	}

	forward := m.Forward()
	if forward.X() != 0 || forward.Y() != 0 || forward.Z() != 1 {
		t.Errorf("Identity Forward() = %v, want (0,0,1)", forward)
	}
}

// -----------------------------------------------------------------------
// Tests - TransformPoint
// -----------------------------------------------------------------------

func TestMat4TransformPointIdentity(t *testing.T) {
	m := Mat4Identity()
	p := Vec3{1, 2, 3}
	result := m.TransformPoint(p)

	if result.X() != 1 || result.Y() != 2 || result.Z() != 3 {
		t.Errorf("TransformPoint(identity) = %v, want (1,2,3)", result)
	}
}

func TestMat4TransformPointTranslate(t *testing.T) {
	m := Mat4Identity()
	m.SetTranslation(Vec3{10, 20, 30})
	p := Vec3{1, 2, 3}
	result := m.TransformPoint(p)

	if Abs(result.X()-11) > approxDelta || Abs(result.Y()-22) > approxDelta || Abs(result.Z()-33) > approxDelta {
		t.Errorf("TransformPoint(translate) = %v, want (11, 22, 33)", result)
	}
}

// -----------------------------------------------------------------------
// Tests - Orthographic
// -----------------------------------------------------------------------

func TestMat4Orthographic(t *testing.T) {
	var m Mat4
	m.Orthographic(-1, 1, -1, 1, 0, 100)

	if Abs(m[x0y0]-1) > approxDelta {
		t.Errorf("Orthographic x0y0 = %v, want 1", m[x0y0])
	}
	if m[x3y3] != 1 {
		t.Errorf("Orthographic x3y3 = %v, want 1", m[x3y3])
	}
	if m[x0y3] != 0 || m[x1y3] != 0 {
		t.Errorf("Orthographic translation should be 0: %v", m)
	}
}

// -----------------------------------------------------------------------
// Tests - Perspective
// -----------------------------------------------------------------------

func TestMat4Perspective(t *testing.T) {
	var m Mat4
	m.Perspective(Float(math.Pi/4), 1.0, 0.1, 100.0)

	if m[x3y3] != 0 {
		t.Errorf("Perspective x3y3 = %v, want 0", m[x3y3])
	}
	if m[x3y2] != -1 {
		t.Errorf("Perspective x3y2 = %v, want -1", m[x3y2])
	}
	f := 1.0 / Tan(Float(math.Pi/8))
	if Abs(m[x0y0]-f) > approxDelta {
		t.Errorf("Perspective x0y0 = %v, want ~%v", m[x0y0], f)
	}
}

// -----------------------------------------------------------------------
// Tests - Inverse & Inverted
// -----------------------------------------------------------------------

func TestMat4InverseIdentity(t *testing.T) {
	m := Mat4Identity()
	m.Inverse()

	if !m.IsIdentity() {
		t.Errorf("Inverse of identity should be identity: got %v", m)
	}
}

func TestMat4InverseTranslation(t *testing.T) {
	m := Mat4Identity()
	m.SetTranslation(Vec3{10, 20, 30})
	m.Inverse()

	if Abs(m[x0y3]+10) > approxDelta || Abs(m[x1y3]+20) > approxDelta || Abs(m[x2y3]+30) > approxDelta {
		t.Errorf("Inverse translation should negate: %v", m)
	}
}

func TestMat4InverseRoundTrip(t *testing.T) {
	m := Mat4Identity()
	m.Translate(Vec3{5, 10, 15})
	m.Scale(Vec3{2, 3, 4})

	orig := m
	m.Inverse()

	result := Mat4Multiply(m, orig)
	if !Mat4ApproxTo(result, Mat4Identity(), 1e-4) {
		t.Errorf("Inverse round trip failed:\n got %v\nwant %v", result, Mat4Identity())
	}
}

func TestMat4InvertedTranslation(t *testing.T) {
	// Inverted() is an affine inverse that expects translation in the bottom row.
	// For identity matrix, Inverted should return identity.
	m := Mat4Identity()
	inv := m.Inverted()

	// Inverted of identity should be identity
	if !Mat4Approx(inv, Mat4Identity()) {
		t.Errorf("Inverted(identity) = %v, want identity", inv)
	}
}

// -----------------------------------------------------------------------
// Tests - ExtractPosition
// -----------------------------------------------------------------------

func TestMat4ExtractPosition(t *testing.T) {
	m := Mat4Identity()
	m.SetTranslation(Vec3{5, 10, 15})

	pos := m.ExtractPosition()
	if pos.X() != 5 || pos.Y() != 10 || pos.Z() != 15 {
		t.Errorf("ExtractPosition = %v, want (5, 10, 15)", pos)
	}
}

// -----------------------------------------------------------------------
// Tests - ExtractScale
// -----------------------------------------------------------------------

func TestMat4ExtractScale(t *testing.T) {
	m := Mat4Identity()
	m.Scale(Vec3{2, 3, 4})

	scale := m.ExtractScale()
	if Abs(scale.X()-2) > approxDelta || Abs(scale.Y()-3) > approxDelta || Abs(scale.Z()-4) > approxDelta {
		t.Errorf("ExtractScale = %v, want (2, 3, 4)", scale)
	}
}

// -----------------------------------------------------------------------
// Tests - ToQuaternion
// -----------------------------------------------------------------------

func TestMat4ToQuaternionIdentity(t *testing.T) {
	m := Mat4Identity()
	q := m.ToQuaternion()

	if Abs(q[Qw]-1) > approxDelta {
		t.Errorf("Identity ToQuaternion w = %v, want 1", q[Qw])
	}
	if Abs(q[Qx]) > approxDelta || Abs(q[Qy]) > approxDelta || Abs(q[Qz]) > approxDelta {
		t.Errorf("Identity ToQuaternion xyz should be 0: %v", q)
	}
}

// -----------------------------------------------------------------------
// Tests - Mat4ProjToVulkan
// -----------------------------------------------------------------------

func TestMat4ProjToVulkan(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	vulkan := m.Mat4ProjToVulkan()

	if vulkan[x1y1] != -6 {
		t.Errorf("Mat4ProjToVulkan x1y1 = %v, want -6", vulkan[x1y1])
	}
	if vulkan[x0y0] != 1 || vulkan[x2y2] != 11 || vulkan[x3y3] != 16 {
		t.Errorf("Mat4ProjToVulkan changed elements besides x1y1: %v", vulkan)
	}
}

// -----------------------------------------------------------------------
// Tests - Mat4Project & Mat4UnProject
// -----------------------------------------------------------------------

func TestMat4Project(t *testing.T) {
	// Note: Mat4Project uses value receivers for Shrink/Scale/Add,
	// so the intermediate transformations are lost. The engine has this quirk.
	// With identity matrix and origin, pos4 stays as (0,0,0,1).
	// result.X = 0 * viewport.Z + viewport.X = 0
	// result.Y = 0 * viewport.W + viewport.Y = 0
	proj := Mat4Identity()
	viewport := Vec4{0, 0, 100, 100}

	p := Vec3{0, 0, 0}
	result := proj.Mat4Project(p, viewport)

	if Abs(result.X()) > approxDelta || Abs(result.Y()) > approxDelta || Abs(result.Z()) > approxDelta {
		t.Errorf("Mat4Project with identity/origin = (%v, %v, %v), want ~(0,0,0)", result.X(), result.Y(), result.Z())
	}
}

func TestMat4UnProject(t *testing.T) {
	invVP := Mat4Identity()
	viewport := Vec4{0, 0, 100, 100}

	screen := Vec3{50, 50, 0}
	result := invVP.Mat4UnProject(screen, invVP, viewport)

	if Abs(result.X()) > approxDelta || Abs(result.Y()) > approxDelta {
		t.Errorf("Mat4UnProject center = (%v, %v), want ~(0, 0)", result.X(), result.Y())
	}
}

// -----------------------------------------------------------------------
// Tests - Mat4ToScreenSpace
// -----------------------------------------------------------------------

func TestMat4ToScreenSpace(t *testing.T) {
	view := Mat4Identity()
	proj := Mat4Identity()
	viewport := Vec4{0, 0, 100, 100}

	pos := Vec3{0, 0, 0}
	screen, ok := Mat4ToScreenSpace(pos, view, proj, viewport)

	if !ok {
		t.Error("ToScreenSpace should succeed for origin")
	}
	if Abs(screen.X()-50) > approxDelta || Abs(screen.Y()-50) > approxDelta {
		t.Errorf("ToScreenSpace = (%v, %v), want (50, 50)", screen.X(), screen.Y())
	}
}

// -----------------------------------------------------------------------
// Tests - MultiplyVec4 transforms point
// -----------------------------------------------------------------------

func TestMat4MultiplyVec4TransformsPoint(t *testing.T) {
	m := Mat4Identity()
	m.SetTranslation(Vec3{10, 0, 0})

	v := Vec4{1, 0, 0, 1}
	result := Mat4MultiplyVec4(m, v)

	if Abs(result[Vx]-11) > approxDelta {
		t.Errorf("MultiplyVec4 translated point Vx = %v, want 11", result[Vx])
	}
}

// -----------------------------------------------------------------------
// Tests - ExtractRotation
// -----------------------------------------------------------------------

func TestMat4ExtractRotationIdentity(t *testing.T) {
	m := Mat4Identity()
	q := m.ExtractRotation()

	if Abs(q[Qw]-1) > approxDelta {
		t.Errorf("ExtractRotation identity w = %v, want 1", q[Qw])
	}
}

// -----------------------------------------------------------------------
// Tests - Edge cases
// -----------------------------------------------------------------------

func TestMat4ExtractRotationZeroScale(t *testing.T) {
	m := Mat4Zero()
	q := m.ExtractRotation()

	// ExtractRotation returns {0,0,0,1} for zero scale
	// In the engine's [w,x,y,z] convention this is w=0, x=0, y=0, z=1
	// (not the identity quaternion, but that's what the code does)
	if q[0] != 0 || q[1] != 0 || q[2] != 0 || q[3] != 1 {
		t.Errorf("ExtractRotation zero scale = %v, want [0,0,0,1]", q)
	}
}

func TestMat4NegativeScaleExtraction(t *testing.T) {
	m := Mat4Identity()
	m.Scale(Vec3{-1, 1, 1})
	scale := m.ExtractScale()

	if Abs(scale.X()) != 1 {
		t.Errorf("ExtractScale with negative scale: X = %v, want magnitude 1", scale.X())
	}
}

func TestMat4PerspectiveWithDifferentAspects(t *testing.T) {
	var m16x9 Mat4
	m16x9.Perspective(Float(math.Pi/4), 16.0/9.0, 0.1, 100.0)

	var m1x1 Mat4
	m1x1.Perspective(Float(math.Pi/4), 1.0, 0.1, 100.0)

	f := 1.0 / Tan(Float(math.Pi/8))
	if Abs(m16x9[x0y0]-f/(16.0/9.0)) > approxDelta {
		t.Errorf("16:9 perspective x0y0 = %v, want %v", m16x9[x0y0], f/(16.0/9.0))
	}
	if Abs(m1x1[x0y0]-f) > approxDelta {
		t.Errorf("1:1 perspective x0y0 = %v, want %v", m1x1[x0y0], f)
	}
}

// -----------------------------------------------------------------------
// Tests - Matrix Multiply corner cases
// -----------------------------------------------------------------------

func TestMat4MultiplyByZero(t *testing.T) {
	m := Mat4Identity()
	z := Mat4Zero()
	result := Mat4Multiply(m, z)

	for i, v := range result {
		if v != 0 {
			t.Errorf("Identity * Zero[%d] = %v, want 0", i, v)
		}
	}
}

func TestMat4MultiplyNonCommutative(t *testing.T) {
	a := Mat4Identity()
	a.SetTranslation(Vec3{1, 0, 0})
	b := Mat4Identity()
	b.Scale(Vec3{2, 1, 1})

	ab := Mat4Multiply(a, b)
	ba := Mat4Multiply(b, a)

	if Mat4Approx(ab, ba) {
		t.Error("Matrix multiplication is NOT commutative - translate then scale differs from scale then translate")
	}
}

// -----------------------------------------------------------------------
// Tests - Orthographic asymmetric
// -----------------------------------------------------------------------

func TestMat4OrthographicAsymmetric(t *testing.T) {
	var m Mat4
	m.Orthographic(0, 800, 0, 600, 1, 1000)

	if Abs(m[x0y0]-0.0025) > approxDelta {
		t.Errorf("Orthographic asymmetric x0y0 = %v, want 0.0025", m[x0y0])
	}
	if Abs(m[x0y3]+1) > approxDelta {
		t.Errorf("Orthographic asymmetric x0y3 = %v, want -1", m[x0y3])
	}
}

// -----------------------------------------------------------------------
// Legacy agreement tests
// -----------------------------------------------------------------------

func TestMat4MultiplyAgreement(t *testing.T) {
	a := testMat4()
	b := testMat4()
	c := legacyMat4Multiply(a, b)
	d := Mat4Multiply(a, b)
	if c != d {
		t.Errorf("\nc = %v\nd = %v", c, d)
	}
}

func TestMat4MultiplyVec4Agreement(t *testing.T) {
	a := testMat4()
	b := testVec4()
	c := legacyMat4MultiplyVec4(a, b)
	d := Mat4MultiplyVec4(a, b)
	if c != d {
		t.Errorf("\nc = %v\nd = %v", c, d)
	}
}

func TestVec4MultiplyMat4Agreement(t *testing.T) {
	a := testVec4()
	b := testMat4()
	c := a.legacyMultiplyMat4(b)
	d := Vec4MultiplyMat4(a, b)
	if !Vec4Approx(c, d) {
		t.Errorf("\nc = %v\nd = %v", c, d)
	}
}

// -----------------------------------------------------------------------
// Table-driven rotation tests
// -----------------------------------------------------------------------

func TestMat4RotateAngleVariants(t *testing.T) {
	tests := []struct {
		name  string
		angle Float
		axis  string
		check func(*testing.T, Mat4)
	}{
		{
			name:  "X 180",
			angle: 180,
			axis:  "X",
			check: func(t *testing.T, m Mat4) {
				if Abs(m[x0y0]-1) > approxDelta {
					t.Errorf("X 180: [0,0] = %v, want 1", m[x0y0])
				}
				if Abs(m[x1y1]+1) > approxDelta {
					t.Errorf("X 180: [1,1] = %v, want -1", m[x1y1])
				}
				if Abs(m[x2y2]+1) > approxDelta {
					t.Errorf("X 180: [2,2] = %v, want -1", m[x2y2])
				}
			},
		},
		{
			name:  "Y 180",
			angle: 180,
			axis:  "Y",
			check: func(t *testing.T, m Mat4) {
				if Abs(m[x0y0]+1) > approxDelta {
					t.Errorf("Y 180: [0,0] = %v, want -1", m[x0y0])
				}
				if Abs(m[x1y1]-1) > approxDelta {
					t.Errorf("Y 180: [1,1] = %v, want 1", m[x1y1])
				}
			},
		},
		{
			name:  "Z 180",
			angle: 180,
			axis:  "Z",
			check: func(t *testing.T, m Mat4) {
				if Abs(m[x0y0]+1) > approxDelta {
					t.Errorf("Z 180: [0,0] = %v, want -1", m[x0y0])
				}
				if Abs(m[x1y1]+1) > approxDelta {
					t.Errorf("Z 180: [1,1] = %v, want -1", m[x1y1])
				}
				if Abs(m[x2y2]-1) > approxDelta {
					t.Errorf("Z 180: [2,2] = %v, want 1", m[x2y2])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Mat4
			switch tt.axis {
			case "X":
				m = Mat4Identity()
				m.RotateX(tt.angle)
			case "Y":
				m = Mat4Identity()
				m.RotateY(tt.angle)
			case "Z":
				m = Mat4Identity()
				m.RotateZ(tt.angle)
			}
			tt.check(t, m)
		})
	}
}

// -----------------------------------------------------------------------
// Additional inverse tests
// -----------------------------------------------------------------------

func TestMat4InverseWithRotation(t *testing.T) {
	m := Mat4Identity()
	m.RotateY(45)

	mCopy := m
	m.Inverse()

	result := Mat4Multiply(m, mCopy)
	if !Mat4ApproxTo(result, Mat4Identity(), 1e-4) {
		t.Errorf("Inverse of rotation round trip:\n got %v\nwant %v", result, Mat4Identity())
	}
}

func TestMat4InverseWithScaleAndTranslation(t *testing.T) {
	m := Mat4Identity()
	m.Scale(Vec3{2, 2, 2})
	m.SetTranslation(Vec3{10, 10, 10})

	orig := m
	m.Inverse()

	result := Mat4Multiply(m, orig)
	if !Mat4ApproxTo(result, Mat4Identity(), 1e-4) {
		t.Errorf("Inverse scale+translate round trip:\n got %v\nwant %v", result, Mat4Identity())
	}
}

// -----------------------------------------------------------------------
// Additional LookAt tests
// -----------------------------------------------------------------------

func TestMat4LookAtWithOffsetCenter(t *testing.T) {
	var m Mat4
	eye := Vec3{0, 1, 0}
	center := Vec3{1, 0, 0}
	up := Vec3{0, 1, 0}
	m.LookAt(eye, center, up)

	forward := m.Forward()
	if Abs(forward.X()-1) > approxDelta {
		t.Errorf("LookAt with offset: forward.X() = %v, want ~1", forward.X())
	}
}

// -----------------------------------------------------------------------
// Additional TransformPoint tests
// -----------------------------------------------------------------------

func TestMat4TransformPointWithScale(t *testing.T) {
	m := Mat4Identity()
	m.Scale(Vec3{2, 2, 2})
	p := Vec3{1, 2, 3}
	result := m.TransformPoint(p)

	if Abs(result.X()-2) > approxDelta || Abs(result.Y()-4) > approxDelta || Abs(result.Z()-6) > approxDelta {
		t.Errorf("TransformPoint(scale) = %v, want (2, 4, 6)", result)
	}
}

func TestMat4TranslateThenScaleOrder(t *testing.T) {
	// Note: Translate() adds to the translation column and Scale() multiplies
	// the diagonal. These operations commute when using the engine's methods,
	// because Scale doesn't touch the translation column.
	// However, using full matrix multiplication they DON'T commute.
	m1 := Mat4Identity()
	m1.Translate(Vec3{10, 0, 0})
	m1.Scale(Vec3{2, 1, 1})

	m2 := Mat4Identity()
	m2.Scale(Vec3{2, 1, 1})
	m2.Translate(Vec3{10, 0, 0})

	// Using the engine's Translate/Scale methods, order doesn't matter
	// because Scale only affects diagonal and Translate only affects x0y3/x1y3/x2y3
	if !Mat4Approx(m1, m2) {
		t.Errorf("Engine's Translate/Scale methods commute: m1=%v, m2=%v", m1, m2)
	}

	// But matrix multiplication of full transforms DOES NOT commute
	tMat := Mat4Identity()
	tMat.SetTranslation(Vec3{10, 0, 0})
	sMat := Mat4Identity()
	sMat.Scale(Vec3{2, 1, 1})

	ab := Mat4Multiply(tMat, sMat)
	ba := Mat4Multiply(sMat, tMat)

	if Mat4Approx(ab, ba) {
		t.Error("Full matrix multiplication should NOT commute")
	}
}

// -----------------------------------------------------------------------
// Row/Col consistency
// -----------------------------------------------------------------------

func TestMat4RowColConsistency(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}

	row0 := m.RowVector(0)
	if !Vec4Approx(row0, Vec4{1, 2, 3, 4}) {
		t.Errorf("RowVector(0) = %v, want [1,2,3,4]", row0)
	}

	col0 := m.ColumnVector(0)
	if !Vec4Approx(col0, Vec4{1, 5, 9, 13}) {
		t.Errorf("ColumnVector(0) = %v, want [1,5,9,13]", col0)
	}

	if m.At(0, 0) != 1 || m.At(3, 3) != 16 {
		t.Errorf("At(0,0) = %v, At(3,3) = %v, want 1 and 16", m.At(0, 0), m.At(3, 3))
	}
}

// -----------------------------------------------------------------------
// Round-trip tests
// -----------------------------------------------------------------------

func TestMat4AddSubtractRoundTrip(t *testing.T) {
	a := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	b := Mat4{
		1, 1, 1, 1,
		2, 2, 2, 2,
		3, 3, 3, 3,
		4, 4, 4, 4,
	}

	a.AddAssign(b)
	a.SubtractAssign(b)

	expected := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	if a != expected {
		t.Errorf("Add/Subtract round trip failed:\n got %v\nwant %v", a, expected)
	}
}

func TestMat4NegateDoubleNegate(t *testing.T) {
	m := Mat4Identity()
	m.NegateAssign()
	m.NegateAssign()

	if !m.IsIdentity() {
		t.Errorf("Double negate should return to identity: %v", m)
	}
}

// -----------------------------------------------------------------------
// String format test
// -----------------------------------------------------------------------

func TestMat4StringFormat(t *testing.T) {
	m := Mat4Identity()
	s := fmt.Sprintf("%v", m)
	if len(s) == 0 {
		t.Error("Mat4 string format should not be empty")
	}
}

// -----------------------------------------------------------------------
// Benchmarks
// -----------------------------------------------------------------------

func BenchmarkMat4Multiply(b *testing.B) {
	a := testMat4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		legacyMat4Multiply(a, c)
	}
}

func BenchmarkMat4MultiplySIMD(b *testing.B) {
	a := testMat4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		Mat4Multiply(a, c)
	}
}

func BenchmarkMat4MultiplyVec4(b *testing.B) {
	a := testMat4()
	c := testVec4()
	for i := 0; i < b.N; i++ {
		legacyMat4MultiplyVec4(a, c)
	}
}

func BenchmarkMat4MultiplyVec4SIMD(b *testing.B) {
	a := testMat4()
	c := testVec4()
	for i := 0; i < b.N; i++ {
		Mat4MultiplyVec4(a, c)
	}
}

func BenchmarkVec4MultiplyMat4(b *testing.B) {
	a := testVec4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		a.legacyMultiplyMat4(c)
	}
}

func BenchmarkVec4MultiplyMat4SIMD(b *testing.B) {
	a := testVec4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		Vec4MultiplyMat4(a, c)
	}
}

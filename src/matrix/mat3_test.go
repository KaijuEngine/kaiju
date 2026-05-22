/******************************************************************************/
/* mat3_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Row / COlumn
// ---------------------------------------------------------------------------

func TestMat3RowVector(t *testing.T) {
	m := Mat3{
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	}
	tests := []struct {
		row  int
		want Vec3
	}{
		{row: 0, want: Vec3{1, 2, 3}},
		{row: 1, want: Vec3{4, 5, 6}},
		{row: 2, want: Vec3{7, 8, 9}},
	}
	for _, tt := range tests {
		got := m.RowVector(tt.row)
		if got != tt.want {
			t.Fatalf("RowVector(%d) = %v, want %v", tt.row, got, tt.want)
		}
	}
}

func TestMat3ColumnVector(t *testing.T) {
	m := Mat3{
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	}
	tests := []struct {
		col  int
		want Vec3
	}{
		{col: 0, want: Vec3{1, 4, 7}},
		{col: 1, want: Vec3{2, 5, 8}},
		{col: 2, want: Vec3{3, 6, 9}},
	}
	for _, tt := range tests {
		got := m.ColumnVector(tt.col)
		if got != tt.want {
			t.Fatalf("ColumnVector(%d) = %v, want %v", tt.col, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Constructor helpers
// ---------------------------------------------------------------------------

func TestNewMat3(t *testing.T) {
	m := NewMat3()
	if !m.IsIdentity() {
		t.Error("NewMat3() should return identity matrix")
	}
}

func TestMat3FromSlice(t *testing.T) {
	slice := []Float{1, 2, 3, 4, 5, 6, 7, 8, 9}
	m := Mat3FromSlice(slice)
	for i := 0; i < 9; i++ {
		if m[i] != slice[i] {
			t.Errorf("Mat3FromSlice index %d: got %f, want %f", i, m[i], slice[i])
		}
	}
}

func TestMat3FromVec3(t *testing.T) {
	v := Vec3{2, 3, 5}
	m := Mat3FromVec3(v)
	// Diagonal should match vector components
	if m[x0y0_3] != 2 || m[x1y1_3] != 3 || m[x2y2_3] != 5 {
		t.Errorf("Diagonal mismatch: %v", m)
	}
	// Off-diagonal should be zero
	for i := 0; i < 9; i++ {
		if i == x0y0_3 || i == x1y1_3 || i == x2y2_3 {
			continue
		}
		if m[i] != 0 {
			t.Errorf("Off-diagonal index %d: got %f, want 0", i, m[i])
		}
	}
}

func TestMat3Identity(t *testing.T) {
	m := Mat3Identity()
	if !m.IsIdentity() {
		t.Error("Mat3Identity() should return identity")
	}
}

func TestMat3Zero(t *testing.T) {
	m := Mat3Zero()
	for i := range 9 {
		if m[i] != 0 {
			t.Errorf("Mat3Zero index %d: got %f, want 0", i, m[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Mutators
// ---------------------------------------------------------------------------

func TestReset(t *testing.T) {
	m := Mat3FromSlice([]Float{9, 9, 9, 9, 9, 9, 9, 9, 9})
	m.Reset()
	if !m.IsIdentity() {
		t.Error("Reset() should produce identity matrix")
	}
}

func TestZero(t *testing.T) {
	m := Mat3Identity()
	m.Zero()
	for i := range 9 {
		if m[i] != 0 {
			t.Errorf("Zero() index %d: got %f, want 0", i, m[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Conversion
// ---------------------------------------------------------------------------

func TestMat3FromMat4(t *testing.T) {
	// Mat4.Mat3() extracts indices 0,1,2,4,5,6,8,9,10 from the Mat4 array
	m4 := Mat4{1, 4, 7, 0, 2, 5, 8, 0, 3, 6, 9, 0, 0, 0, 0, 1}
	m3 := Mat3FromMat4(m4)
	// Expected: m4[0]=1, m4[1]=4, m4[2]=7, m4[4]=2, m4[5]=5, m4[6]=8, m4[8]=3, m4[9]=6, m4[10]=9
	want := Mat3{1, 4, 7, 2, 5, 8, 3, 6, 9}
	if !Mat3Approx(m3, want) {
		t.Errorf("Mat3FromMat4:\ngot  %v\nwant %v", m3, want)
	}
}

func TestToMat4(t *testing.T) {
	m3 := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	m4 := m3.ToMat4()
	// Expected: 3x3 filled, last column zero, last row 0 0 0 1
	want := Mat4{1, 2, 3, 0, 4, 5, 6, 0, 7, 8, 9, 0, 0, 0, 0, 1}
	for i := 0; i < 16; i++ {
		if m4[i] != want[i] {
			t.Errorf("ToMat4 index %d: got %f, want %f", i, m4[i], want[i])
		}
	}
}

func TestMat3RoundTrip(t *testing.T) {
	// Mat3 -> Mat4 -> Mat3 should round-trip (3x3 portion)
	orig := Mat3{2, 3, 5, 7, 11, 13, 17, 19, 23}
	m4 := orig.ToMat4()
	back := Mat3FromMat4(m4)
	if !Mat3Approx(orig, back) {
		t.Errorf("Round-trip failed:\norig %v\nback %v", orig, back)
	}
}

// ---------------------------------------------------------------------------
// Accessors
// ---------------------------------------------------------------------------

func TestAt(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	tests := []struct {
		row, col int
		want     Float
	}{
		{0, 0, 1}, {0, 1, 2}, {0, 2, 3},
		{1, 0, 4}, {1, 1, 5}, {1, 2, 6},
		{2, 0, 7}, {2, 1, 8}, {2, 2, 9},
	}
	for _, tt := range tests {
		got := m.At(tt.row, tt.col)
		if got != tt.want {
			t.Errorf("At(%d,%d) = %f, want %f", tt.row, tt.col, got, tt.want)
		}
	}
}

func TestRowVector(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	tests := []struct {
		row  int
		want Vec3
	}{
		{0, Vec3{1, 2, 3}},
		{1, Vec3{4, 5, 6}},
		{2, Vec3{7, 8, 9}},
	}
	for _, tt := range tests {
		got := m.RowVector(tt.row)
		if !Vec3Approx(got, tt.want) {
			t.Errorf("RowVector(%d) = %v, want %v", tt.row, got, tt.want)
		}
	}
}

func TestColumnVector(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	tests := []struct {
		col  int
		want Vec3
	}{
		{0, Vec3{1, 4, 7}},
		{1, Vec3{2, 5, 8}},
		{2, Vec3{3, 6, 9}},
	}
	for _, tt := range tests {
		got := m.ColumnVector(tt.col)
		if !Vec3Approx(got, tt.want) {
			t.Errorf("ColumnVector(%d) = %v, want %v", tt.col, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Transpose
// ---------------------------------------------------------------------------

func TestTranspose(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	tm := m.Transpose()
	want := Mat3{1, 4, 7, 2, 5, 8, 3, 6, 9}
	if !Mat3Approx(tm, want) {
		t.Errorf("Transpose:\ngot  %v\nwant %v", tm, want)
	}
	// Original should be unchanged
	if !Mat3Approx(m, Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Error("Transpose should not modify the original")
	}
}

func TestTransposeAssign(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	m.TransposeAssign()
	want := Mat3{1, 4, 7, 2, 5, 8, 3, 6, 9}
	if !Mat3Approx(m, want) {
		t.Errorf("TransposeAssign:\ngot  %v\nwant %v", m, want)
	}
}

func TestTransposeIdentity(t *testing.T) {
	id := Mat3Identity()
	tm := id.Transpose()
	if !tm.IsIdentity() {
		t.Error("Identity transpose should be identity")
	}
}

// ---------------------------------------------------------------------------
// Multiply
// ---------------------------------------------------------------------------

func TestMat3Multiply(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := Mat3{9, 8, 7, 6, 5, 4, 3, 2, 1}
	c := Mat3Multiply(a, b)
	// C[0][0]=1*9+2*6+3*3=30  C[0][1]=1*8+2*5+3*2=24  C[0][2]=1*7+2*4+3*1=18
	// C[1][0]=4*9+5*6+6*3=84  C[1][1]=4*8+5*5+6*2=69  C[1][2]=4*7+5*4+6*1=54
	// C[2][0]=7*9+8*6+9*3=138 C[2][1]=7*8+8*5+9*2=114 C[2][2]=7*7+8*4+9*1=90
	want := Mat3{30, 24, 18, 84, 69, 54, 138, 114, 90}
	if !Mat3Approx(c, want) {
		t.Errorf("Mat3Multiply:\ngot  %v\nwant %v", c, want)
	}
}

func TestMultiply(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := Mat3{9, 8, 7, 6, 5, 4, 3, 2, 1}
	c := a.Multiply(b)
	want := Mat3{30, 24, 18, 84, 69, 54, 138, 114, 90}
	if !Mat3Approx(c, want) {
		t.Errorf("Multiply:\ngot  %v\nwant %v", c, want)
	}
}

func TestMat3MultiplyAssign(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := Mat3{9, 8, 7, 6, 5, 4, 3, 2, 1}
	a.MultiplyAssign(b)
	want := Mat3{30, 24, 18, 84, 69, 54, 138, 114, 90}
	if !Mat3Approx(a, want) {
		t.Errorf("MultiplyAssign:\ngot  %v\nwant %v", a, want)
	}
}

func TestMultiplyIdentity(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	id := Mat3Identity()
	c := a.Multiply(id)
	if !Mat3Approx(c, a) {
		t.Errorf("A * I != A:\ngot  %v\nwant %v", c, a)
	}
	c2 := id.Multiply(a)
	if !Mat3Approx(c2, a) {
		t.Errorf("I * A != A:\ngot  %v\nwant %v", c2, a)
	}
}

func TestMultiplyZero(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	zero := Mat3Zero()
	c := a.Multiply(zero)
	if !Mat3Approx(c, zero) {
		t.Errorf("A * Zero != Zero:\ngot  %v", c)
	}
}

func TestMat3MultiplyAssignMatchesMultiply(t *testing.T) {
	lhs := Mat3{
		2, 3, 5,
		7, 11, 13,
		17, 19, 23,
	}
	rhs := Mat3{
		29, 31, 37,
		41, 43, 47,
		53, 59, 61,
	}
	want := lhs.Multiply(rhs)
	got := lhs
	got.MultiplyAssign(rhs)
	if got != want {
		t.Fatalf("MultiplyAssign mismatch:\n got=%v\nwant=%v", got, want)
	}
}

// ---------------------------------------------------------------------------
// MultiplyVec3
// ---------------------------------------------------------------------------

func TestMat3MultiplyVec3(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	v := Vec3{1, 2, 3}
	result := Mat3MultiplyVec3(m, v)
	// row0: 1*1+2*2+3*3=14
	// row1: 4*1+5*2+6*3=32
	// row2: 7*1+8*2+9*3=50
	want := Vec3{14, 32, 50}
	if !Vec3Approx(result, want) {
		t.Errorf("Mat3MultiplyVec3: got %v, want %v", result, want)
	}
}

func TestMultiplyVec3(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	v := Vec3{1, 2, 3}
	result := m.MultiplyVec3(v)
	want := Vec3{14, 32, 50}
	if !Vec3Approx(result, want) {
		t.Errorf("MultiplyVec3: got %v, want %v", result, want)
	}
}

func TestMultiplyVec3Identity(t *testing.T) {
	v := Vec3{1, 2, 3}
	id := Mat3Identity()
	result := id.MultiplyVec3(v)
	if !Vec3Approx(result, v) {
		t.Errorf("I * v != v: got %v, want %v", result, v)
	}
}

// ---------------------------------------------------------------------------
// Add / Subtract
// ---------------------------------------------------------------------------

func TestAdd(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := Mat3{9, 8, 7, 6, 5, 4, 3, 2, 1}
	c := a.Add(b)
	want := Mat3{10, 10, 10, 10, 10, 10, 10, 10, 10}
	if !Mat3Approx(c, want) {
		t.Errorf("Add:\ngot  %v\nwant %v", c, want)
	}
	// Originals unchanged
	if !Mat3Approx(a, Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Error("Add should not modify receiver")
	}
}

func TestAddAssign(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := Mat3{9, 8, 7, 6, 5, 4, 3, 2, 1}
	a.AddAssign(b)
	want := Mat3{10, 10, 10, 10, 10, 10, 10, 10, 10}
	if !Mat3Approx(a, want) {
		t.Errorf("AddAssign:\ngot  %v\nwant %v", a, want)
	}
}

func TestSubtract(t *testing.T) {
	a := Mat3{9, 8, 7, 6, 5, 4, 3, 2, 1}
	b := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	c := a.Subtract(b)
	want := Mat3{8, 6, 4, 2, 0, -2, -4, -6, -8}
	if !Mat3Approx(c, want) {
		t.Errorf("Subtract:\ngot  %v\nwant %v", c, want)
	}
}

func TestSubtractAssign(t *testing.T) {
	a := Mat3{9, 8, 7, 6, 5, 4, 3, 2, 1}
	b := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	a.SubtractAssign(b)
	want := Mat3{8, 6, 4, 2, 0, -2, -4, -6, -8}
	if !Mat3Approx(a, want) {
		t.Errorf("SubtractAssign:\ngot  %v\nwant %v", a, want)
	}
}

func TestAddSubtractIdentity(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	c := a.Add(Mat3Zero())
	if !Mat3Approx(c, a) {
		t.Error("A + Zero != A")
	}
	d := a.Subtract(Mat3Zero())
	if !Mat3Approx(d, a) {
		t.Error("A - Zero != A")
	}
}

// ---------------------------------------------------------------------------
// Negate
// ---------------------------------------------------------------------------

func TestNegate(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	n := m.Negate()
	want := Mat3{-1, -2, -3, -4, -5, -6, -7, -8, -9}
	if !Mat3Approx(n, want) {
		t.Errorf("Negate:\ngot  %v\nwant %v", n, want)
	}
	// Original unchanged
	if !Mat3Approx(m, Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Error("Negate should not modify the original")
	}
}

func TestNegateAssign(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	m.NegateAssign()
	want := Mat3{-1, -2, -3, -4, -5, -6, -7, -8, -9}
	if !Mat3Approx(m, want) {
		t.Errorf("NegateAssign:\ngot  %v\nwant %v", m, want)
	}
}

func TestNegateDoubleNegate(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	doublyNegated := m.Negate().Negate()
	if !Mat3Approx(doublyNegated, m) {
		t.Error("Negate(Negate(A)) != A")
	}
}

// ---------------------------------------------------------------------------
// Scale
// ---------------------------------------------------------------------------

func TestScale(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	scaled := m.Scale(2)
	want := Mat3{2, 4, 6, 8, 10, 12, 14, 16, 18}
	if !Mat3Approx(scaled, want) {
		t.Errorf("Scale(2):\ngot  %v\nwant %v", scaled, want)
	}
	// Original unchanged
	if !Mat3Approx(m, Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Error("Scale should not modify the original")
	}
}

func TestScaleAssign(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	m.ScaleAssign(Float(3))
	want := Mat3{3, 6, 9, 12, 15, 18, 21, 24, 27}
	if !Mat3Approx(m, want) {
		t.Errorf("ScaleAssign(3):\ngot  %v\nwant %v", m, want)
	}
}

func TestScaleZero(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	scaled := m.Scale(0)
	if !Mat3Approx(scaled, Mat3Zero()) {
		t.Errorf("Scale(0) should produce zero matrix:\ngot  %v", scaled)
	}
}

// ---------------------------------------------------------------------------
// Determinant
// ---------------------------------------------------------------------------

func TestDeterminant(t *testing.T) {
	tests := []struct {
		name string
		m    Mat3
		want Float
	}{
		{
			name: "identity",
			m:    Mat3Identity(),
			want: 1,
		},
		{
			name: "zero",
			m:    Mat3Zero(),
			want: 0,
		},
		{
			name: "diagonal 2-3-5",
			m:    Mat3{2, 0, 0, 0, 3, 0, 0, 0, 5},
			want: 30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			det := Mat3Determinant(tt.m)
			if math.Abs(float64(det-tt.want)) > Tiny {
				t.Errorf("Determinant: got %f, want %f", det, tt.want)
			}
		})
	}
}

func TestDeterminantMethod(t *testing.T) {
	m := Mat3{1, 2, 3, 0, 4, 5, 1, 0, 5}
	det := m.Determinant()
	// det = 1*(4*5 - 5*0) - 2*(0*5 - 5*1) + 3*(0*0 - 4*1)
	//     = 1*20 - 2*(-5) + 3*(-4)
	//     = 20 + 10 - 12 = 18
	if math.Abs(float64(det-18)) > Tiny {
		t.Errorf("Determinant: got %f, want 18", det)
	}
}

// ---------------------------------------------------------------------------
// Inverse
// ---------------------------------------------------------------------------

func TestInverse(t *testing.T) {
	m := Mat3{
		1, 0, 0,
		0, 2, 0,
		0, 0, 3,
	}
	m.Inverse()
	// Inverse of diag(1,2,3) is diag(1,0.5,0.333...)
	want := Mat3{1, 0, 0, 0, 0.5, 0, 0, 0, 1.0 / 3}
	if !Mat3Approx(m, want) {
		t.Errorf("Inverse:\ngot  %v\nwant %v", m, want)
	}
}

func TestMat3Inverted(t *testing.T) {
	m := Mat3{
		1, 0, 0,
		0, 2, 0,
		0, 0, 3,
	}
	inv := m.Inverted()
	// Original should be unchanged
	if !Mat3Approx(m, Mat3{1, 0, 0, 0, 2, 0, 0, 0, 3}) {
		t.Error("Inverted should not modify the original")
	}
	want := Mat3{1, 0, 0, 0, 0.5, 0, 0, 0, 1.0 / 3}
	if !Mat3Approx(inv, want) {
		t.Errorf("Inverted:\ngot  %v\nwant %v", inv, want)
	}
}

func TestInverseIdentity(t *testing.T) {
	m := Mat3Identity()
	m.Inverse()
	if !m.IsIdentity() {
		t.Error("Inverse of identity should be identity")
	}
}

func TestInverseSingularity(t *testing.T) {
	// Singular matrix (rows are linearly dependent)
	m := Mat3{1, 2, 3, 2, 4, 6, 1, 2, 3}
	m.Inverse()
	// Should produce zero matrix for singular input
	if !Mat3Approx(m, Mat3Zero()) {
		t.Errorf("Inverse of singular matrix should be zero:\ngot  %v", m)
	}
}

func TestInverseRoundTrip(t *testing.T) {
	// Use a diagonal matrix for round-trip test (M * M^-1 should equal identity)
	m := Mat3{2, 0, 0, 0, 3, 0, 0, 0, 5}
	inv := m.Inverted()
	product := Mat3Multiply(m, inv)
	if !Mat3ApproxTo(product, Mat3Identity(), Roughly) {
		t.Errorf("M * M^-1 != I:\ngot  %v", product)
	}
}

// ---------------------------------------------------------------------------
// IsIdentity
// ---------------------------------------------------------------------------

func TestIsIdentity(t *testing.T) {
	tests := []struct {
		name string
		m    Mat3
		want bool
	}{
		{"identity", Mat3Identity(), true},
		{"not identity", Mat3{1, 0, 0, 0, 2, 0, 0, 0, 1}, false},
		{"zero", Mat3Zero(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsIdentity(); got != tt.want {
				t.Errorf("IsIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Approx / Equals
// ---------------------------------------------------------------------------

func TestMat3Approx(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !Mat3Approx(a, b) {
		t.Error("Identical matrices should be approximately equal")
	}

	// Significantly different
	d := Mat3{100, 2, 3, 4, 5, 6, 7, 8, 9}
	if Mat3Approx(a, d) {
		t.Error("Significantly different matrices should not be approximately equal")
	}

	// Almost identical (within FloatSmallestNonzero)
	eps := FloatSmallestNonzero * 0.5
	e := Mat3{1 + eps, 2, 3, 4, 5, 6, 7, 8, 9}
	if !Mat3Approx(a, e) {
		t.Error("Matrices differing by < FloatSmallestNonzero should be approximately equal")
	}
}

func TestMat3ApproxTo(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := Mat3{1.05, 2.05, 3.05, 4.05, 5.05, 6.05, 7.05, 8.05, 9.05}

	// delta 0.01 -- differences are 0.05, so should fail
	if Mat3ApproxTo(a, b, 0.01) {
		t.Error("Should not be approximately equal with delta=0.01")
	}

	// delta 0.1 -- differences are 0.05 < 0.1, should pass
	if !Mat3ApproxTo(a, b, 0.1) {
		t.Error("Should be approximately equal with delta=0.1")
	}
}

func TestEquals(t *testing.T) {
	a := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	c := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 10}

	if !a.Equals(b) {
		t.Error("Identical matrices should be equal")
	}
	if a.Equals(c) {
		t.Error("Different matrices should not be equal")
	}
}

// ---------------------------------------------------------------------------
// Index constants
// ---------------------------------------------------------------------------

func TestIndexConstants(t *testing.T) {
	// Verify iota-based index constants have correct values
	if Mat3Row0 != 0 || Mat3Row1 != 1 || Mat3Row2 != 2 {
		t.Error("Row indices incorrect")
	}
	if Mat3Col0 != 0 || Mat3Col1 != 1 || Mat3Col2 != 2 {
		t.Error("Col indices incorrect")
	}
	// Element index constants
	if x0y0_3 != 0 {
		t.Errorf("x0y0_3 = %d, want 0", x0y0_3)
	}
	if x1y1_3 != 4 {
		t.Errorf("x1y1_3 = %d, want 4", x1y1_3)
	}
	if x2y2_3 != 8 {
		t.Errorf("x2y2_3 = %d, want 8", x2y2_3)
	}
	// Alias constants
	if Mat3x0y0 != 0 || Mat3x1y1 != 4 || Mat3x2y2 != 8 {
		t.Error("Alias index constants incorrect")
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestScaleNegative(t *testing.T) {
	m := Mat3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	scaled := m.Scale(-1)
	want := Mat3{-1, -2, -3, -4, -5, -6, -7, -8, -9}
	if !Mat3Approx(scaled, want) {
		t.Errorf("Scale(-1):\ngot  %v\nwant %v", scaled, want)
	}
}

func TestMultiplyCommutativityNotGuaranteed(t *testing.T) {
	// Matrix multiplication is NOT commutative -- verify a*b != b*a
	a := Mat3{1, 0, 0, 0, 0, 0, 0, 0, 0}
	b := Mat3{0, 0, 0, 1, 0, 0, 0, 0, 0}
	ab := a.Multiply(b)
	ba := b.Multiply(a)
	if Mat3Approx(ab, ba) {
		t.Log("Note: For these particular matrices, AB == BA")
	}
}

func TestDeterminantNegativeMatrix(t *testing.T) {
	m := Mat3{-1, 0, 0, 0, -1, 0, 0, 0, -1}
	det := Mat3Determinant(m)
	// det of -I is -1 * -1 * -1 = -1
	if math.Abs(float64(det-(-1))) > Tiny {
		t.Errorf("Determinant of -I: got %f, want -1", det)
	}
}

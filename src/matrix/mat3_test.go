package matrix

import "testing"

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

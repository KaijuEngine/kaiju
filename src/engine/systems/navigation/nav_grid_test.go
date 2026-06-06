/******************************************************************************/
/* nav_grid_test.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package navigation

import (
	"testing"

	"kaijuengine.com/matrix"
)

// ---------------------------------------------------------------------------
// NewGrid
// ---------------------------------------------------------------------------

func TestNewGridDimensions(t *testing.T) {
	g := NewGrid(5, 7, 3)
	if g.Width() != 5 {
		t.Errorf("Width() = %d, want 5", g.Width())
	}
	if g.Height() != 7 {
		t.Errorf("Height() = %d, want 7", g.Height())
	}
	if g.Depth() != 3 {
		t.Errorf("Depth() = %d, want 3", g.Depth())
	}
}

func TestNewGridCellsZeroed(t *testing.T) {
	g := NewGrid(2, 2, 2)
	for x := 0; x < g.Width(); x++ {
		for y := 0; y < g.Height(); y++ {
			for z := 0; z < g.Depth(); z++ {
				if g[x][y][z] != 0 {
					t.Errorf("Cell [%d][%d][%d] = %d, want 0", x, y, z, g[x][y][z])
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// IsValid
// ---------------------------------------------------------------------------

func TestIsValid(t *testing.T) {
	g := NewGrid(3, 3, 3)

	tests := []struct {
		name string
		pos  matrix.Vec3i
		want bool
	}{
		{"origin", matrix.Vec3i{0, 0, 0}, true},
		{"center", matrix.Vec3i{1, 1, 1}, true},
		{"max corner", matrix.Vec3i{2, 2, 2}, true},
		{"x out of bounds", matrix.Vec3i{3, 1, 1}, false},
		{"y out of bounds", matrix.Vec3i{1, 3, 1}, false},
		{"z out of bounds", matrix.Vec3i{1, 1, 3}, false},
		{"negative x", matrix.Vec3i{-1, 1, 1}, false},
		{"all negative", matrix.Vec3i{-1, -1, -1}, false},
	}
	for _, tt := range tests {
		got := g.IsValid(tt.pos)
		if got != tt.want {
			t.Errorf("%s: IsValid(%v) = %v, want %v", tt.name, tt.pos, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// BlockCell / IsBlocked
// ---------------------------------------------------------------------------

func TestBlockCellAndIsBlocked(t *testing.T) {
	g := NewGrid(3, 3, 3)
	pos := matrix.Vec3i{1, 1, 1}

	if g.IsBlocked(pos) {
		t.Error("New cell should not be blocked")
	}

	g.BlockCell(pos, 1)
	if !g.IsBlocked(pos) {
		t.Error("Cell should be blocked after BlockCell")
	}
}

func TestBlockCellOutOfboundsIsBlocked(t *testing.T) {
	g := NewGrid(3, 3, 3)
	outOfBounds := matrix.Vec3i{5, 5, 5}
	if !g.IsBlocked(outOfBounds) {
		t.Error("Out-of-bounds position should be blocked")
	}
}

// ---------------------------------------------------------------------------
// BlockedType
// ---------------------------------------------------------------------------

func TestBlockedType(t *testing.T) {
	g := NewGrid(3, 3, 3)
	pos := matrix.Vec3i{1, 1, 1}

	// Unblocked cell returns 0
	if g.BlockedType(pos) != 0 {
		t.Errorf("BlockedType(unblocked) = %d, want 0", g.BlockedType(pos))
	}

	// Out of bounds returns -1
	if got := g.BlockedType(matrix.Vec3i{-1, 0, 0}); got != -1 {
		t.Errorf("BlockedType(out of bounds) = %d, want -1", got)
	}

	// Blocked cell returns block type
	g.BlockCell(pos, 42)
	if got := g.BlockedType(pos); got != 42 {
		t.Errorf("BlockedType(blocked) = %d, want 42", got)
	}
}

func TestBlockedTypeOverride(t *testing.T) {
	g := NewGrid(2, 2, 2)
	pos := matrix.Vec3i{0, 0, 0}

	g.BlockCell(pos, 5)
	if g.BlockedType(pos) != 5 {
		t.Errorf("Expected type 5, got %d", g.BlockedType(pos))
	}

	g.BlockCell(pos, 10)
	if g.BlockedType(pos) != 10 {
		t.Errorf("Expected type 10 after override, got %d", g.BlockedType(pos))
	}
}

// ---------------------------------------------------------------------------
// Dimension accessors
// ---------------------------------------------------------------------------

func TestWidthHeightDepth(t *testing.T) {
	g := NewGrid(4, 6, 2)
	if g.Width() != 4 || g.Height() != 6 || g.Depth() != 2 {
		t.Errorf("Dimensions: got %dx%dx%d, want 4x6x2", g.Width(), g.Height(), g.Depth())
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestNewGridMinimal(t *testing.T) {
	g := NewGrid(1, 1, 1)
	if g.Width() != 1 || g.Height() != 1 || g.Depth() != 1 {
		t.Error("1x1x1 grid should have dimensions of 1")
	}
	if !g.IsValid(matrix.Vec3i{0, 0, 0}) {
		t.Error("Cell [0][0][0] should be valid in 1x1x1 grid")
	}
}

func TestBlockCellZeroType(t *testing.T) {
	// Blocking with type 0 means the cell is effectively unblocked
	g := NewGrid(2, 2, 2)
	pos := matrix.Vec3i{0, 0, 0}
	g.BlockCell(pos, 0)
	if g.IsBlocked(pos) {
		t.Error("Cell blocked with type 0 should not be blocked")
	}
}

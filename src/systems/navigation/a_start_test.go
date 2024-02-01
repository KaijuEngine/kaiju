package navigation

import (
	"kaiju/matrix"
	"testing"
)

func TestAStar(t *testing.T) {
	grid := [][][]int8{
		{
			{0, 0, 0},
			{0, 1, 0},
			{0, 0, 0},
		},
		{
			{0, 0, 0},
			{0, 1, 0},
			{0, 0, 0},
		},
		{
			{0, 0, 0},
			{0, 0, 0},
			{0, 0, 1},
		},
	}
	start := matrix.Vec3i{0, 0, 0}
	end := matrix.Vec3i{2, 2, 2}
	path := AStar(grid, start, end)
	if path == nil {
		t.Fail()
	}
}

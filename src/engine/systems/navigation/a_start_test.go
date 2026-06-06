/******************************************************************************/
/* a_start_test.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package navigation

import (
	"testing"

	"kaijuengine.com/matrix"
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

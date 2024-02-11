/*****************************************************************************/
/* nav_grid.go                                                               */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package navigation

import "kaiju/matrix"

type Grid [][][]int8

func NewGrid(width, height, depth int) Grid {
	cells := make([][][]int8, width)
	for i := range cells {
		cells[i] = make([][]int8, height)
		for j := range cells[i] {
			cells[i][j] = make([]int8, depth)
		}
	}
	return cells
}

func (g Grid) Width() int {
	return len(g)
}

func (g Grid) Height() int {
	return len(g[0])
}

func (g Grid) Depth() int {
	return len(g[0][0])
}

func (g *Grid) BlockCell(pos matrix.Vec3i, blockType int8) {
	(*g)[pos.X()][pos.Y()][pos.Z()] = blockType
}

func (g Grid) IsBlocked(pos matrix.Vec3i) bool {
	return !g.IsValid(pos) || g[pos.X()][pos.Y()][pos.Z()] != 0
}

func (g Grid) BlockedType(pos matrix.Vec3i) int8 {
	if !g.IsValid(pos) {
		return -1
	} else {
		return g[pos.X()][pos.Y()][pos.Z()]
	}
}

func (g Grid) IsValid(pos matrix.Vec3i) bool {
	return pos.X() >= 0 && pos.X() < int32(len(g)) &&
		pos.Y() >= 0 && pos.Y() < int32(len(g[0])) &&
		pos.Z() >= 0 && pos.Z() < int32(len(g[0][0]))
}

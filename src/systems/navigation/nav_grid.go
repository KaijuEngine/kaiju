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

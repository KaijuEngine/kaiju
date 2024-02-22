//go:build editor

package rendering

import (
	"kaiju/matrix"
	"slices"
)

type meshDetails struct {
	Verts   []matrix.Vec3
	Indexes []uint32
	Extents matrix.Vec3
}

func (m *meshDetails) Set(verts []Vertex, indexes []uint32) {
	m.Verts = make([]matrix.Vec3, len(verts))
	for i, v := range verts {
		m.Verts[i] = v.Position
		m.Extents = matrix.Vec3MaxAbs(m.Extents, v.Position)
	}
	m.Indexes = slices.Clone(indexes)
}
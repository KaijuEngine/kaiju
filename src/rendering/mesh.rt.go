//go:build !editor

package rendering

type meshDetails struct{}

func (m *meshDetails) Set(verts []Vertex, indexes []uint32) {}

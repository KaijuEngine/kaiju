package loaders

import "kaiju/rendering"

type ResultMesh struct {
	Name     string
	Verts    []rendering.Vertex
	Indexes  []uint32
	Textures []string
}

type Result []ResultMesh

func (r Result) IsValid() bool { return len(r) > 0 }

func (r *Result) Add(name string, verts []rendering.Vertex, indexes []uint32, textures []string) {
	*r = append(*r, ResultMesh{
		Name:     name,
		Verts:    verts,
		Indexes:  indexes,
		Textures: textures,
	})
}

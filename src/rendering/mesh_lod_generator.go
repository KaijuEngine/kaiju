package rendering

type MeshLod struct {
	Levels []MeshLODInstance
}

type MeshLODInstance struct {
	Mesh  *Mesh
	Ratio float32
}

type MeshLodGenerator interface {
	GenerateLods(mesh *Mesh, cache *MeshCache, verts []Vertex, indices []uint32, levels int) (MeshLod, error)
}

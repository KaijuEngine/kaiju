/******************************************************************************/
/* mesh_lod_generator.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

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

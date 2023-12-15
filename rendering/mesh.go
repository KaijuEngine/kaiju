package rendering

type MeshDrawMode int

const (
	MeshDrawModePoints MeshDrawMode = iota
	MeshDrawModeLines
	MeshDrawModeTriangles
	MeshDrawModePatches
)

type Mesh struct {
	MeshId MeshId
}

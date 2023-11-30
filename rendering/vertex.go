package rendering

import "kaiju/matrix"

type Vertex struct {
	Position     matrix.Vec3
	Normal       matrix.Vec3
	Tangent      matrix.Vec4
	UV0          matrix.Vec2
	Color        matrix.Color
	JointIds     matrix.Vec4i
	JointWeights matrix.Vec4
	MorphTarget  matrix.Vec3
}

func VertexFaceNormal(verts [3]Vertex) matrix.Vec3 {
	v0 := verts[0].Position
	v1 := verts[1].Position
	v2 := verts[2].Position
	e0 := v1.Subtract(v0)
	e1 := v2.Subtract(v2)
	c := matrix.Vec3Cross(e1, e0)
	return c.Normal()
}

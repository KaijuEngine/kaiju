/******************************************************************************/
/* load_result.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package load_result

import (
	"log/slog"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type AnimationPathType = int
type AnimationInterpolation = int

const (
	AnimPathInvalid AnimationPathType = iota - 1
	AnimPathTranslation
	AnimPathRotation
	AnimPathScale
	AnimPathWeights
)

const (
	AnimInterpolateInvalid AnimationInterpolation = iota - 1
	AnimInterpolateLinear
	AnimInterpolateStep
	AnimInterpolateCubicSpline
)

type Mesh struct {
	Node     *Node
	Name     string
	MeshName string
	Verts    []rendering.Vertex
	Indexes  []uint32
	Textures map[string]string
}

type AnimBone struct {
	NodeIndex     int
	PathType      AnimationPathType
	Interpolation AnimationInterpolation
	// Could be Vec3 or Quaternion, doing this because Go doesn't have a union
	Data [4]matrix.Float
}

type AnimKeyFrame struct {
	Bones []AnimBone
	Time  float32
}

type Animation struct {
	Name   string
	Frames []AnimKeyFrame
}

type Node struct {
	Id         int32
	Name       string
	Parent     int
	Position   matrix.Vec3
	Rotation   matrix.Quaternion
	Scale      matrix.Vec3
	Attributes map[string]any
	IsAnimated bool
}

type Joint struct {
	Id   int32
	Skin matrix.Mat4
}

type Result struct {
	Nodes        []Node
	Meshes       []Mesh
	Animations   []Animation
	Joints       []Joint
	TextureBytes map[string][]byte
}

func (r *Result) IsTreeAnimated(nodeIdx int) bool {
	isAnimated := r.Nodes[nodeIdx].IsAnimated
	p := r.Nodes[nodeIdx].Parent
	for !isAnimated && p >= 0 {
		isAnimated = r.Nodes[p].IsAnimated
		p = r.Nodes[p].Parent
	}
	return isAnimated
}

func (r *Result) IsValid() bool { return len(r.Meshes) > 0 }

func (r *Result) Add(name, meshName string, verts []rendering.Vertex, indexes []uint32, textures map[string]string, node *Node) {
	if node != nil {
		// TODO:  This breaks Sudoku, but seems like something that should be done...
		//mat := node.Transform.CalcWorldMatrix()
		//if !mat.IsIdentity() {
		//	for i := range verts {
		//		verts[i].Position = mat.TransformPoint(verts[i].Position)
		//	}
		//}
	}
	r.Meshes = append(r.Meshes, Mesh{
		Name:     name,
		MeshName: meshName,
		Verts:    verts,
		Indexes:  indexes,
		Textures: textures,
		Node:     node,
	})
}

func (r *Result) NodeByName(name string) *Node {
	for i := range r.Nodes {
		if r.Nodes[i].Name == name {
			return &r.Nodes[i]
		}
	}
	return nil
}

func (r *Result) Extract(names ...string) Result {
	if len(r.Animations) > 0 || len(r.Joints) > 0 {
		slog.Error("extracting animation entries from a mesh load result isn't yet supported")
	}
	res := Result{}
	for i := range names {
		for j := range r.Nodes {
			if r.Nodes[j].Name == names[i] {
				res.Nodes = append(res.Nodes, r.Nodes[j])
				for k := range r.Meshes {
					m := &r.Meshes[k]
					if m.Node == &r.Nodes[j] {
						res.Add(names[i], m.Name, m.Verts, m.Indexes, m.Textures, &res.Nodes[len(res.Nodes)-1])
					}
				}
			}
		}
	}
	return res
}

func (mesh *Mesh) ScaledRadius(scale matrix.Vec3) matrix.Float {
	rad := matrix.Float(0)
	// TODO:  Take scale into consideration
	for i := range mesh.Verts {
		pt := mesh.Verts[i].Position.Multiply(scale)
		rad = max(rad, pt.Length())
	}
	return rad
}

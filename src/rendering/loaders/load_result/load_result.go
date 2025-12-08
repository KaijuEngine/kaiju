/******************************************************************************/
/* load_result.go                                                             */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package load_result

import (
	"log/slog"

	"github.com/KaijuEngine/kaiju/matrix"
	"github.com/KaijuEngine/kaiju/rendering"
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
	Name       string
	Parent     int
	Transform  matrix.Transform
	Attributes map[string]any
}

type Joint struct {
	Id   int32
	Skin matrix.Mat4
}

type Result struct {
	Nodes      []Node
	Meshes     []Mesh
	Animations []Animation
	Joints     []Joint
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

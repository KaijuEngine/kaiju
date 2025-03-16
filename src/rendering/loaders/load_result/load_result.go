/******************************************************************************/
/* load_result.go                                                             */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
	"kaiju/collision"
	"kaiju/concurrent"
	"kaiju/matrix"
	"kaiju/rendering"
	"sync"
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
	Name     string
	MeshName string
	Verts    []rendering.Vertex
	Indexes  []uint32
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
	Name      string
	Parent    int
	Transform matrix.Transform
}

type Joint struct {
	Id   int32
	Skin matrix.Mat4
}

type Result struct {
	Nodes      []Node
	Meshes     []Mesh
	Textures   []string
	Animations []Animation
	Joints     []Joint
}

func NewResult() Result {
	return Result{
		Meshes:   make([]Mesh, 0),
		Textures: make([]string, 0),
	}
}

func (r *Result) IsValid() bool { return len(r.Meshes) > 0 }

func (r *Result) Add(name, meshName string, verts []rendering.Vertex,
	indexes []uint32, textures []string) {

	r.Meshes = append(r.Meshes, Mesh{
		Name:     name,
		MeshName: meshName,
		Verts:    verts,
		Indexes:  indexes,
	})
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

func (m *Mesh) GenerateBVH(threads *concurrent.Threads) *collision.BVH {
	tris := make([]collision.DetailedTriangle, len(m.Indexes)/3)
	group := sync.WaitGroup{}
	construct := func(from, to int) {
		for i := from; i < to; i += 3 {
			for i := 0; i < len(m.Indexes); i += 3 {
				points := [3]matrix.Vec3{
					m.Verts[m.Indexes[i]].Position,
					m.Verts[m.Indexes[i+1]].Position,
					m.Verts[m.Indexes[i+2]].Position,
				}
				tris[i/3] = collision.DetailedTriangleFromPoints(points)
			}
		}
		group.Done()
	}
	group.Add(len(tris))
	for i := range len(tris) {
		threads.AddWork(func() { construct(i*3, (i+3)*3) })
	}
	group.Wait()
	return collision.BVHBottomUp(tris)
}

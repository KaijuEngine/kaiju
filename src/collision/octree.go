/******************************************************************************/
/* octree.go                                                                  */
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

package collision

import "kaiju/matrix"

type Octree struct {
	Center    matrix.Vec3
	HalfWidth matrix.Float
	Children  [8]*Octree
	Objects   []HitObject
}

func NewOctree(center matrix.Vec3, halfWidth matrix.Float, maxDepth int) *Octree {
	if maxDepth == 0 {
		return nil
	}
	n := &Octree{
		Center:    center,
		HalfWidth: halfWidth,
		Objects:   make([]HitObject, 0),
	}
	offset := matrix.Vec3{}
	step := float32(halfWidth * 0.5)
	for i := 0; i < 8; i++ {
		offset = matrix.Vec3{step, step, step}
		if i&1 == 1 {
			offset[matrix.Vx] *= -1
		}
		if i&2 == 2 {
			offset[matrix.Vy] *= -1
		}
		if i&4 == 4 {
			offset[matrix.Vz] *= -1
		}
		n.Children[i] = NewOctree(center.Add(offset), halfWidth*0.5, maxDepth-1)
	}
	return n
}

func OctreeForMesh(mesh []matrix.Vec3) *Octree {
	min := matrix.Vec3{0, 0, 0}
	max := matrix.Vec3{0, 0, 0}
	for _, v := range mesh {
		min = matrix.Vec3Min(min, v)
		max = matrix.Vec3Max(max, v)
	}
	// TODO:  Generate a reasonable stopDepth given the size
	stopDepth := 5
	center := min.Add(max).Scale(0.5)
	halfWidth := max.Subtract(min).Length() * 0.5
	return NewOctree(center, halfWidth, stopDepth)
}

func (o *Octree) AsAABB() AABB {
	return AABB{
		Center: o.Center,
		Extent: matrix.Vec3{o.HalfWidth, o.HalfWidth, o.HalfWidth},
	}
}

func (node *Octree) Insert(obj HitObject) {
	index := 0
	straddle := 0
	for i := 0; i < 3; i++ {
		delta := obj.Bounds().Center[i] - node.Center[i]
		if matrix.Abs(delta) <= obj.Bounds().Extent[i] {
			straddle = 1
			break
		}
		if delta > 0 {
			index |= 1 << uint(i)
		}
	}
	if straddle == 0 && node.Children[index] != nil {
		node.Children[index].Insert(obj)
	} else {
		node.Objects = append(node.Objects, obj)
	}
}

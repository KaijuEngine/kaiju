/******************************************************************************/
/* bvh.go                                                                     */
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

package collision

import (
	"kaiju/matrix"
	"sort"
)

type HitObject interface {
	Bounds() AABB
	RayIntersect(ray Ray, length float32, transform *matrix.Transform) bool
}

type BVH struct {
	bounds AABB
	Left   *BVH
	Right  *BVH
	Parent *BVH
	Data   HitObject
}

type BVHItem struct {
	Transform *matrix.Transform
	Data      HitObject
}

func computeBounds(entries []HitObject) AABB {
	if len(entries) == 0 {
		return AABB{}
	}
	b := entries[0].Bounds()
	for i := 1; i < len(entries); i++ {
		b = AABBUnion(b, entries[i].Bounds())
	}
	return b
}

func NewBVH(entries []HitObject) *BVH {
	if len(entries) == 0 {
		return nil
	}
	if len(entries) == 1 {
		return &BVH{
			bounds: entries[0].Bounds(),
			Data:   entries[0],
		}
	}
	bounds := computeBounds(entries)
	axis := bounds.LongestAxis()
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Bounds().Center[axis] < entries[j].Bounds().Center[axis]
	})
	mid := len(entries) / 2
	left := NewBVH(entries[:mid])
	right := NewBVH(entries[mid:])
	bvh := &BVH{
		bounds: AABBUnion(left.bounds, right.bounds),
		Left:   left,
		Right:  right,
	}
	left.Parent = bvh
	right.Parent = bvh
	return bvh
}

func (b *BVH) Bounds(transform *matrix.Transform) AABB {
	if transform == nil {
		return b.bounds
	} else {
		mat := transform.WorldMatrix()
		min := mat.TransformPoint(b.bounds.Min())
		max := mat.TransformPoint(b.bounds.Max())
		return AABB{
			Center: min.Add(max).Shrink(2.0),
			Extent: max.Subtract(min).Shrink(2.0),
		}
	}
}

func (b *BVH) RayIntersect(ray Ray, length float32, transform *matrix.Transform) bool {
	if b == nil {
		return false
	}
	if b.Left != nil && b.Left.Data != nil && b.Left.Data.RayIntersect(ray, length, transform) {
		return true
	}
	if b.Right != nil && b.Right.Data != nil && b.Right.Data.RayIntersect(ray, length, transform) {
		return true
	}
	bounds := b.Bounds(transform)
	if _, ok := bounds.RayHit(ray); ok {
		return b.Left.RayIntersect(ray, length, transform) || b.Right.RayIntersect(ray, length, transform)
	}
	return false
}

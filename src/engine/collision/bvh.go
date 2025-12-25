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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	RayIntersectTest(ray Ray, length float32, transform *matrix.Transform) (matrix.Vec3, bool)
}

type BVH struct {
	bounds AABB
	Left   *BVH
	Right  *BVH
	Parent *BVH
	Item   BVHItem
}

type BVHItem struct {
	Transform *matrix.Transform
	HitCheck  HitObject
	Data      any
}

func (item BVHItem) IsValid() bool { return item.HitCheck != nil }

func (item BVHItem) Bounds() AABB {
	bounds := item.HitCheck.Bounds()
	if item.Transform == nil {
		return bounds
	}
	mat := item.Transform.WorldMatrix()
	min := mat.TransformPoint(bounds.Min())
	max := mat.TransformPoint(bounds.Max())
	return AABB{
		Center: min.Add(max).Shrink(2.0),
		Extent: max.Subtract(min).Shrink(2.0),
	}
}

func (item BVHItem) RayIntersect(ray Ray, length float32, transform *matrix.Transform) (matrix.Vec3, bool) {
	return item.HitCheck.RayIntersectTest(ray, length, item.Transform)
}

func NewBVH(entries []HitObject, transform *matrix.Transform, data any) *BVH {
	if len(entries) == 0 {
		return nil
	}
	if len(entries) == 1 {
		bvh := &BVH{
			Item: BVHItem{transform, entries[0], data},
		}
		bvh.bounds = bvh.Item.Bounds()
		return bvh
	}
	bounds := computeBounds(entries)
	axis := bounds.LongestAxis()
	// TODO:  Sort only needs to happen at the start?
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Bounds().Center[axis] < entries[j].Bounds().Center[axis]
	})
	mid := len(entries) / 2
	left := NewBVH(entries[:mid], transform, data)
	right := NewBVH(entries[mid:], transform, data)
	bvh := &BVH{
		bounds: AABBUnion(left.bounds, right.bounds),
		Left:   left,
		Right:  right,
	}
	left.Parent = bvh
	right.Parent = bvh
	return bvh
}

func CloneBVH(bvh *BVH) *BVH {
	if bvh == nil {
		return nil
	}
	newItem := bvh.Item
	if sub, ok := bvh.Item.Data.(*BVH); ok {
		newItem.Data = CloneBVH(sub)
	}
	clone := &BVH{
		bounds: bvh.bounds,
		Item:   newItem,
	}
	if bvh.Left != nil {
		clone.Left = CloneBVH(bvh.Left)
		if clone.Left != nil {
			clone.Left.Parent = clone
		}
	}
	if bvh.Right != nil {
		clone.Right = CloneBVH(bvh.Right)
		if clone.Right != nil {
			clone.Right.Parent = clone
		}
	}
	return clone
}

func (b *BVH) RayIntersectTest(ray Ray, length float32, transform *matrix.Transform) (matrix.Vec3, bool) {
	_, pt, ok := b.RayIntersect(ray, length)
	return pt, ok
}

func (b *BVH) RayIntersect(ray Ray, length float32) (any, matrix.Vec3, bool) {
	if b == nil {
		return nil, matrix.Vec3{}, false
	}
	_, hit := b.bounds.RayHit(ray)
	if !hit {
		return nil, matrix.Vec3{}, false
	}
	if b.IsLeaf() && b.Item.IsValid() {
		if pt, ok := b.Item.RayIntersect(ray, length, b.Item.Transform); ok {
			if sub, ok := b.Item.Data.(*BVH); ok {
				return sub.RayIntersect(ray, length)
			}
			return b.Item.Data, pt, true
		}
		return nil, matrix.Vec3{}, false
	}
	resL, ptL, okL := b.Left.RayIntersect(ray, length)
	resR, ptR, okR := b.Right.RayIntersect(ray, length)
	if okL && okR {
		distL := ptL.Subtract(ray.Origin).LengthSquared()
		distR := ptR.Subtract(ray.Origin).LengthSquared()
		if distL < distR {
			return resL, ptL, true
		}
		return resR, ptR, true
	}
	if okL {
		return resL, ptL, true
	}
	if okR {
		return resR, ptR, true
	}
	return nil, matrix.Vec3{}, false
}

func (b *BVH) Bounds() AABB {
	return b.bounds
}

func (b *BVH) IsLeaf() bool {
	return b.Left == nil && b.Right == nil
}

func (b *BVH) Refit() {
	if b == nil {
		return
	}
	b.refitChildren()
}

func (b *BVH) refitChildren() {
	if b.IsLeaf() {
		b.bounds = b.Item.Bounds()
	} else {
		b.Left.refitChildren()
		b.Right.refitChildren()
		b.bounds = AABBUnion(b.Left.bounds, b.Right.bounds)
	}
}

func AddSubBVH(world **BVH, sub *BVH, transform *matrix.Transform) {
	InsertBVH(world, sub, transform, sub)
}

func InsertBVH(root **BVH, hitCheck HitObject, transform *matrix.Transform, data any) {
	if *root == nil {
		if sub, ok := hitCheck.(*BVH); ok {
			*root = sub
		} else {
			*root = &BVH{Item: BVHItem{transform, hitCheck, data}}
			(*root).bounds = (*root).Item.HitCheck.Bounds()
		}
		return
	}
	var newNode *BVH
	if sub, ok := hitCheck.(*BVH); ok {
		newNode = sub
	} else {
		newNode = &BVH{Item: BVHItem{transform, hitCheck, data}}
		newNode.bounds = newNode.Item.Bounds()
	}
	sibling := findBestSibling(*root, newNode.bounds)
	oldParent := sibling.Parent
	newParent := &BVH{
		bounds: AABBUnion(sibling.bounds, newNode.bounds),
		Left:   sibling,
		Right:  newNode,
		Parent: oldParent,
	}
	sibling.Parent = newParent
	newNode.Parent = newParent
	if oldParent != nil {
		if oldParent.Left == sibling {
			oldParent.Left = newParent
		} else {
			oldParent.Right = newParent
		}
		current := oldParent
		for current != nil {
			current.bounds = AABBUnion(current.Left.bounds, current.Right.bounds)
			current = current.Parent
		}
	} else {
		*root = newParent
	}
}

func RemoveSubBVH(world **BVH, sub *BVH) {
	if *world == nil {
		return
	}
	leaf := findLeafWithData(*world, sub)
	if leaf == nil {
		return
	}
	removeLeaf(world, leaf)
}

func RemoveAllLeavesMatchingTransform(world **BVH, transform *matrix.Transform) {
	if world == nil || *world == nil {
		return
	}
	for {
		leaf := findLeafWithTransform(*world, transform)
		if leaf == nil {
			break
		}
		removeLeaf(world, leaf)
	}
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

func findLeafWithTransform(b *BVH, target *matrix.Transform) *BVH {
	if b == nil {
		return nil
	}
	if b.IsLeaf() {
		if b.Item.Transform == target {
			return b
		}
		return nil
	}
	if left := findLeafWithTransform(b.Left, target); left != nil {
		return left
	}
	return findLeafWithTransform(b.Right, target)
}

func findBestSibling(tree *BVH, newBounds AABB) *BVH {
	current := tree
	for !current.IsLeaf() {
		leftIncrease := AABBUnion(current.Left.bounds, newBounds).SurfaceArea() - current.Left.bounds.SurfaceArea()
		rightIncrease := AABBUnion(current.Right.bounds, newBounds).SurfaceArea() - current.Right.bounds.SurfaceArea()
		if leftIncrease < rightIncrease {
			current = current.Left
		} else {
			current = current.Right
		}
	}
	return current
}

func findLeafWithData(b *BVH, target HitObject) *BVH {
	if b == nil {
		return nil
	}
	if b.IsLeaf() {
		if b.Item.HitCheck == target || b == target {
			return b
		}
		return nil
	}
	if left := findLeafWithData(b.Left, target); left != nil {
		return left
	}
	return findLeafWithData(b.Right, target)
}

func removeLeaf(root **BVH, leaf *BVH) {
	if leaf.Parent == nil {
		*root = nil
		return
	}
	parent := leaf.Parent
	var sibling *BVH
	if parent.Left == leaf {
		sibling = parent.Right
	} else {
		sibling = parent.Left
	}
	grandParent := parent.Parent
	sibling.Parent = grandParent
	if grandParent != nil {
		if grandParent.Left == parent {
			grandParent.Left = sibling
		} else {
			grandParent.Right = sibling
		}
		current := grandParent
		for current != nil {
			current.bounds = AABBUnion(current.Left.bounds, current.Right.bounds)
			current = current.Parent
		}
	} else {
		*root = sibling
	}
}

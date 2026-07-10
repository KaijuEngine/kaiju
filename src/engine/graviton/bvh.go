/******************************************************************************/
/* bvh.go                                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"sort"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type HitObject interface {
	Bounds() AABB
	RayIntersectTest(ray Ray, length matrix.Float, transform *matrix.Transform) (matrix.Vec3, bool)
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

type TriangleBVH struct {
	Bounds      AABB
	Left        *TriangleBVH
	Right       *TriangleBVH
	Triangle    DetailedTriangle
	HasTriangle bool
}

func (item BVHItem) IsValid() bool { return item.HitCheck != nil }

func (item BVHItem) Bounds() AABB {
	if item.HitCheck == nil {
		return AABB{}
	}
	bounds := item.HitCheck.Bounds()
	if item.Transform == nil {
		return bounds
	}
	if _, ok := item.HitCheck.(*BVH); ok {
		return bounds
	}
	return bounds.Transform(item.Transform.WorldMatrix())
}

func (item BVHItem) RayIntersect(ray Ray, length matrix.Float, transform *matrix.Transform) (matrix.Vec3, bool) {
	defer tracing.NewRegion("BVHItem.RayIntersect").End()
	return item.HitCheck.RayIntersectTest(ray, length, item.Transform)
}

func NewBVH(entries []HitObject, transform *matrix.Transform, data any) *BVH {
	defer tracing.NewRegion("collision.NewBVH").End()
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

func NewTriangleBVH(bvh *BVH) *TriangleBVH {
	defer tracing.NewRegion("collision.NewTriangleBVH").End()
	if bvh == nil {
		return nil
	}
	out := &TriangleBVH{
		Bounds: bvh.bounds,
	}
	if bvh.IsLeaf() && bvh.Item.IsValid() {
		if tri, ok := bvh.Item.HitCheck.(DetailedTriangle); ok {
			out.Triangle = tri
			out.HasTriangle = true
		}
		return out
	}
	out.Left = NewTriangleBVH(bvh.Left)
	out.Right = NewTriangleBVH(bvh.Right)
	return out
}

func (b *TriangleBVH) ToBVH(transform *matrix.Transform, data any) *BVH {
	if b == nil {
		return nil
	}
	out := &BVH{
		bounds: b.Bounds,
	}
	if b.HasTriangle {
		out.Item = BVHItem{transform, b.Triangle, data}
		return out
	}
	out.Left = b.Left.ToBVH(transform, data)
	out.Right = b.Right.ToBVH(transform, data)
	if out.Left != nil {
		out.Left.Parent = out
	}
	if out.Right != nil {
		out.Right.Parent = out
	}
	return out
}

func (b *BVH) RayIntersectTest(ray Ray, length matrix.Float, transform *matrix.Transform) (matrix.Vec3, bool) {
	defer tracing.NewRegion("BVH.RayIntersectTest").End()
	_, pt, ok := b.RayIntersect(ray, length)
	return pt, ok
}

func (b *BVH) RayIntersect(ray Ray, length matrix.Float) (any, matrix.Vec3, bool) {
	defer tracing.NewRegion("BVH.RayIntersect").End()
	if b == nil {
		return nil, matrix.Vec3{}, false
	}
	_, hit := b.bounds.RayHit(ray)
	if !hit {
		return nil, matrix.Vec3{}, false
	}
	if b.IsLeaf() && b.Item.IsValid() {
		if sub, ok := b.Item.Data.(*BVH); ok {
			if _, ok := b.Item.HitCheck.(*BVH); ok {
				return sub.RayIntersect(ray, length)
			}
		}
		if pt, ok := b.Item.RayIntersect(ray, length, b.Item.Transform); ok {
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
	defer tracing.NewRegion("BVH.Refit").End()
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

func (b *BVH) RefitUpwards() {
	defer tracing.NewRegion("BVH.RefitUpwards").End()
	for current := b; current != nil; current = current.Parent {
		current.refitNode()
	}
}

func (b *BVH) refitNode() {
	if b.IsLeaf() {
		if b.Item.IsValid() {
			b.bounds = b.Item.Bounds()
		}
	} else {
		b.bounds = AABBUnion(b.Left.bounds, b.Right.bounds)
	}
}

func AddSubBVH(world **BVH, sub *BVH, transform *matrix.Transform) *BVH {
	defer tracing.NewRegion("collision.AddSubBVH").End()
	if sub == nil {
		return nil
	}
	sub.Refit()
	return InsertBVH(world, sub, transform, sub)
}

func InsertBVH(root **BVH, hitCheck HitObject, transform *matrix.Transform, data any) *BVH {
	defer tracing.NewRegion("collision.InsertBVH").End()
	if root == nil || hitCheck == nil {
		return nil
	}
	newNode := &BVH{Item: BVHItem{transform, hitCheck, data}}
	newNode.bounds = newNode.Item.Bounds()
	if *root == nil {
		*root = newNode
		return newNode
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
	return newNode
}

func RemoveSubBVH(world **BVH, sub *BVH) {
	defer tracing.NewRegion("collision.RemoveSubBVH").End()
	if *world == nil {
		return
	}
	leaf := findLeafWithData(*world, sub)
	if leaf == nil {
		return
	}
	removeLeaf(world, leaf)
}

func RemoveBVHNode(root **BVH, node *BVH) {
	defer tracing.NewRegion("collision.RemoveBVHNode").End()
	if root == nil || *root == nil || node == nil {
		return
	}
	if node.Parent == nil && *root != node {
		return
	}
	removeLeaf(root, node)
}

func RemoveAllLeavesMatchingTransform(world **BVH, transform *matrix.Transform) {
	defer tracing.NewRegion("collision.RemoveAllLeavesMatchingTransform").End()
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
	defer tracing.NewRegion("collision.computeBounds").End()
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
	defer tracing.NewRegion("collision.findLeafWithTransform").End()
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
	defer tracing.NewRegion("collision.findBestSibling").End()
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
	defer tracing.NewRegion("collision.findLeafWithData").End()
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
	defer tracing.NewRegion("collision.removeLeaf").End()
	if leaf.Parent == nil {
		*root = nil
		leaf.Parent = nil
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
	leaf.Parent = nil
	parent.Left = nil
	parent.Right = nil
}

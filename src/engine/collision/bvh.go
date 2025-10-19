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
	"log/slog"
)

type HitObject interface {
	Bounds() AABB
	RayIntersect(ray Ray, length float32) bool
}

type BVH struct {
	bounds    AABB
	Left      *BVH
	Right     *BVH
	Parent    *BVH
	Transform *matrix.Transform
	Data      HitObject
}

func NewBVH() *BVH { return &BVH{} }

func (b *BVH) duplicateInternal() *BVH {
	dupe := &BVH{}
	*dupe = *b
	return dupe
}

func (b *BVH) Duplicate() *BVH {
	root := b
	nodeMap := make(map[*BVH]*BVH)
	stack := []*BVH{root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, exists := nodeMap[node]; !exists {
			nodeMap[node] = node.duplicateInternal()
		}
		if node.Left != nil {
			stack = append(stack, node.Left)
		}
		if node.Right != nil {
			stack = append(stack, node.Right)
		}
	}
	for node, newNode := range nodeMap {
		if node.Left != nil {
			newNode.Left = nodeMap[node.Left]
		}
		if node.Right != nil {
			newNode.Right = nodeMap[node.Right]
		}
		if node.Parent != nil {
			newNode.Parent = nodeMap[node.Parent]
		}
	}
	return nodeMap[root]
}

func (b *BVH) Root() *BVH {
	r := b
	for r.Parent != nil {
		r = r.Parent
	}
	return r
}

func (b *BVH) Bounds() AABB {
	if b.Transform == nil {
		return b.bounds
	} else {
		mat := b.Transform.WorldMatrix()
		min := mat.TransformPoint(b.bounds.Min())
		max := mat.TransformPoint(b.bounds.Max())
		return AABB{
			Center: min.Add(max).Shrink(2.0),
			Extent: max.Subtract(min).Shrink(2.0),
		}
	}
}

// IsLeaf returns whether or not the BVH is a leaf node
func (b *BVH) IsLeaf() bool {
	return b.Left == nil && b.Right == nil
}

// IsRoot returns whether or not the BVH is the root node
func (b *BVH) IsRoot() bool {
	return b.Parent == nil
}

// IsLeft returns whether or not the BVH is the left child of its parent
func (b *BVH) IsLeft() bool { return b.Parent != nil || b.Parent.Left == b }

// IsRight returns whether or not the BVH is the right child of its parent
func (b *BVH) IsRight() bool { return b.Parent != nil || b.Parent.Right == b }

// BVHBottomUp constructs a BVH from a list of triangles
func BVHBottomUp(triangles []DetailedTriangle) *BVH {
	// TODO:  Get all the current nodes and re-calculate
	nodes := make([]*BVH, 0, len(triangles))
	for i := range triangles {
		nodes = append(nodes, &BVH{
			bounds: triangles[i].Bounds(),
			Data:   &triangles[i],
		})
	}
	for len(nodes) > 1 {
		var x, y int
		nearest(nodes, &x, &y)
		if y < x {
			x, y = y, x
		}
		a := nodes[x]
		b := nodes[y]
		node := &BVH{
			bounds: AABBUnion(a.Bounds(), b.Bounds()),
			Left:   a,
			Right:  b,
		}
		a.Parent = node
		b.Parent = node
		nodes[x] = node
		nodes[y] = nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
	}
	return nodes[0]
}

func (into *BVH) Insert(other *BVH) {
	if !into.IsRoot() {
		slog.Error("Insert should only be called on the root node, use BVHInsert instead")
		return
	}
	BVHInsert(into, other)
}

// BVHInsert inserts a new BVH into an existing BVH, returning the new root
func BVHInsert(into, other *BVH) *BVH {
	ib := into.Bounds()
	ob := other.Bounds()
	if !ib.ContainsAABB(ob) {
		// The root node is a special case, it is expected to hold everything
		// though we could return a new root, it is better to just expand the
		// root node to hold everything
		if into.IsRoot() {
			into.bounds = AABBUnion(ib, ob)
			into.bounds.Extent.ScaleAssign(1.001)
			return BVHInsert(into, other)
		} else {
			bvh := &BVH{
				bounds: AABBUnion(ib, ob),
				Left:   into,
				Right:  other,
			}
			into.Parent = bvh
			other.Parent = bvh
			return bvh
		}
	} else {
		if into.Left == nil {
			into.Left = other
			other.Parent = into
			return into
		}
		lb := into.Left.Bounds()
		if lb.ContainsAABB(ob) {
			if left := BVHInsert(into.Left, other); left != into.Left {
				slog.Error(
					"BVHInsert: Left child was replaced but should not have been",
					slog.String("Left Center", lb.Center.String()),
					slog.String("Left Extent", lb.Extent.String()),
					slog.String("Insert Center", ob.Center.String()),
					slog.String("Insert Extent", ob.Extent.String()),
				)
			}
			return into
		}
		if into.Right == nil {
			into.Right = other
			other.Parent = into
			return into
		}
		rb := into.Right.Bounds()
		if rb.ContainsAABB(ob) {
			if right := BVHInsert(into.Right, other); right != into.Right {
				slog.Error(
					"BVHInsert: Right child was replaced but should not have been",
					slog.String("Right Center", rb.Center.String()),
					slog.String("Right Extent", rb.Extent.String()),
					slog.String("Insert Center", ob.Center.String()),
					slog.String("Insert Extent", ob.Extent.String()),
				)
			}
			return into
		}
		if lb.ClosestDistance(ob) <= rb.ClosestDistance(ob) {
			into.Left = BVHInsert(into.Left, other)
		} else {
			into.Right = BVHInsert(into.Right, other)
		}
		return into
	}
}

// RemoveNode removes a node from the BVH and adjusts the tree accordingly. If
// the node is the root, nothing is done.
func (b *BVH) RemoveNode() {
	if b.Parent == nil {
		return
	}
	parent := b.Parent
	promote := parent.Left
	if promote == b {
		promote = parent.Right
	}
	if parent.IsRoot() {
		if promote != nil {
			*parent = *promote
			parent.Parent = nil
			if parent.Left != nil {
				parent.Left.Parent = parent
			}
			if parent.Right != nil {
				parent.Right.Parent = parent
			}
		} else {
			parent.Left = nil
		}
	} else {
		if parent.IsLeft() {
			parent.Parent.Left = promote
		} else {
			parent.Parent.Right = promote
		}
	}
}

// RayHit returns the point of intersection and whether or not the ray hit
// the BVH. The point of intersection is the closest point of intersection
// along the ray.
func (b *BVH) RayHit(ray Ray, rayLen matrix.Float) (matrix.Vec3, bool) {
	min := matrix.Float(100000.0)
	ls := LineSegmentFromRay(ray, rayLen)
	mat := matrix.Mat4Identity()
	if b.Transform != nil {
		mat = b.Transform.WorldMatrix()
	}
	return nodeRay(b, ray, ls, &min, &mat)
}

func nearest(nodes []*BVH, x, y *int) {
	nearest := matrix.Float(100000.0)
	for i := 0; i < len(nodes); i++ {
		a := nodes[i]
		for j := i + 1; j < len(nodes); j++ {
			b := nodes[j]
			d := a.Bounds().ClosestDistance(b.Bounds())
			if d < nearest {
				*x = i
				*y = j
				nearest = d
				if d < 1.0 {
					return
				}
			}
		}
	}
}

func nodeRay(b *BVH, r Ray, ls Segment, min *matrix.Float, mat *matrix.Mat4) (matrix.Vec3, bool) {
	if b == nil {
		return matrix.Vec3{}, false
	}
	bounds := b.bounds
	if b.Transform != nil {
		*mat = b.Transform.WorldMatrix()
	}
	bounds.Center = mat.TransformPoint(bounds.Center)
	if _, ok := bounds.RayHit(r); ok {
		if b.IsLeaf() {
			t := b.Data.(*DetailedTriangle)
			hit, _ := r.PlaneHit(t.Centroid, t.Normal)
			d := r.Origin.Distance(hit)
			p0 := mat.TransformPoint(t.Points[0])
			p1 := mat.TransformPoint(t.Points[1])
			p2 := mat.TransformPoint(t.Points[2])
			if d < *min && ls.TriangleHit(p0, p1, p2) {
				*min = d
				return hit, true
			}
		} else {
			outHit, success := nodeRay(b.Left, r, ls, min, mat)
			if !success {
				outHit, success = nodeRay(b.Right, r, ls, min, mat)
			}
			return outHit, success
		}
	}
	return matrix.Vec3{}, false
}

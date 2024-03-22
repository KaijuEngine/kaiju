package collision

import "kaiju/matrix"

type HitObject interface {
	Bounds() AABB
	RayIntersect(ray Ray, length float32) bool
}

type BVH struct {
	Bounds    AABB
	Left      *BVH
	Right     *BVH
	Parent    *BVH
	Transform *matrix.Transform
	Data      HitObject
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
func BVHBottomUp(triangles []DetailedTriangle, transform *matrix.Transform) *BVH {
	// TODO:  Get all the current nodes and re-calculate
	nodes := make([]*BVH, 0, len(triangles))
	for i := range triangles {
		nodes = append(nodes, &BVH{
			Bounds:    triangles[i].Bounds(),
			Data:      &triangles[i],
			Transform: transform,
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
			Bounds:    AABBUnion(a.Bounds, b.Bounds),
			Left:      a,
			Right:     b,
			Transform: transform,
		}
		a.Parent = node
		b.Parent = node
		nodes[x] = node
		nodes[y] = nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
	}
	return nodes[0]
}

// BVHInsert inserts a new BVH into an existing BVH, returning the new root
func BVHInsert(into, other *BVH) *BVH {
	if !into.Bounds.ContainsAABB(other.Bounds) {
		bvh := &BVH{
			Bounds: AABBUnion(into.Bounds, other.Bounds),
			Left:   into,
			Right:  other,
		}
		into.Parent = bvh
		other.Parent = bvh
		return bvh
	} else {
		left := BVHInsert(into.Left, other)
		if left != into.Left {
			into.Right = BVHInsert(into.Right, other)
		}
		into.Left = left
		return into
	}
}

// DestroyNode removes a node from the BVH and adjusts the tree accordingly. If
// the node is the root, nothing is done.
func (b *BVH) DestroyNode() {
	if b.Parent == nil {
		return
	}
	parent := b.Parent
	promote := parent.Left
	if promote == b {
		promote = parent.Right
	}
	if parent.IsRoot() {
		*parent = *promote
		parent.Parent = nil
		parent.Left.Parent = parent
		parent.Right.Parent = parent
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
	return node_ray(b, ray, ls, &min)
}

func nearest(nodes []*BVH, x, y *int) {
	nearest := matrix.Float(100000.0)
	for i := 0; i < len(nodes); i++ {
		a := nodes[i]
		for j := i + 1; j < len(nodes); j++ {
			b := nodes[j]
			d := a.Bounds.ClosestDistance(b.Bounds)
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

func node_ray(b *BVH, r Ray, ls Segment, min *matrix.Float) (matrix.Vec3, bool) {
	if b == nil {
		return matrix.Vec3{}, false
	}
	bounds := b.Bounds
	mat := matrix.Mat4Identity()
	if b.Transform != nil {
		mat = b.Transform.WorldMatrix()
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
			outHit, success := node_ray(b.Left, r, ls, min, transform)
			if !success {
				outHit, success = node_ray(b.Right, r, ls, min, transform)
			}
			return outHit, success
		}
	}
	return matrix.Vec3{}, false
}

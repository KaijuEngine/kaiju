package collision

import "kaiju/matrix"

type HitObject interface {
	Bounds() AABB
	RayIntersect(ray Ray, length float32) bool
}

type BVH struct {
	Bounds AABB
	Left   *BVH
	Right  *BVH
	Data   HitObject
}

func (b *BVH) IsLeaf() bool {
	return b.Left == nil && b.Right == nil
}

func nearest(nodes []*BVH, x, y *int) {
	nearest := matrix.Float(100000.0)
	for i := 0; i < len(nodes); i++ {
		a := nodes[i]
		for j := i + 1; j < len(nodes); j++ {
			b := nodes[j]
			// TODO:  Should do more than just center distance for accuracy
			d := a.Bounds.Center.Distance(b.Bounds.Center)
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

func BVHBottomUp(triangles []DetailedTriangle) *BVH {
	// TODO:  Get all the current nodes and re-calculate
	nodes := make([]*BVH, 0, len(triangles))
	for i := range triangles {
		nodes = append(nodes, &BVH{
			Bounds: triangles[i].Bounds(),
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
			Bounds: AABBUnion(a.Bounds, b.Bounds),
			Left:   a,
			Right:  b,
		}
		nodes[x] = node
		nodes[y] = nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
	}
	return nodes[0]
}

func (b *BVH) RayHit(ray Ray, rayLen matrix.Float, transform *matrix.Mat4) (matrix.Vec3, bool) {
	min := matrix.Float(100000.0)
	ls := LineSegmentFromRay(ray, rayLen)
	return node_ray(b, ray, ls, &min, transform)
}

func node_ray(b *BVH, r Ray, ls Segment, min *matrix.Float, transform *matrix.Mat4) (matrix.Vec3, bool) {
	if b == nil {
		return matrix.Vec3{}, false
	}
	bounds := b.Bounds
	bounds.Center = transform.TransformPoint(bounds.Center)
	if _, ok := bounds.RayHit(r); ok {
		if b.IsLeaf() {
			t := b.Data.(*DetailedTriangle)
			hit, _ := r.PlaneHit(t.Centroid, t.Normal)
			d := r.Origin.Distance(hit)
			p0 := transform.TransformPoint(t.Points[0])
			p1 := transform.TransformPoint(t.Points[1])
			p2 := transform.TransformPoint(t.Points[2])
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

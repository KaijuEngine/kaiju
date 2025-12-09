package collision

import (
	"kaiju/klib"
	"kaiju/matrix"
)

type Sphere struct {
	Position matrix.Vec3
	Radius   float32
}

func (a Sphere) Overlap(b Sphere) bool {
	distSq := a.Position.SquareDistance(b.Position)
	radiusSum := a.Radius + b.Radius
	radiusSumSq := radiusSum * radiusSum
	return distSq <= radiusSumSq
}

func (s Sphere) IntersectsAABB(b AABB) bool {
	var sqDist float32
	minX := b.Center.X() - b.Extent.X()
	maxX := b.Center.X() + b.Extent.X()
	x := s.Position.X()
	if x < minX {
		d := x - minX
		sqDist += d * d
	} else if x > maxX {
		d := x - maxX
		sqDist += d * d
	}
	minY := b.Center.Y() - b.Extent.Y()
	maxY := b.Center.Y() + b.Extent.Y()
	y := s.Position.Y()
	if y < minY {
		d := y - minY
		sqDist += d * d
	} else if y > maxY {
		d := y - maxY
		sqDist += d * d
	}
	minZ := b.Center.Z() - b.Extent.Z()
	maxZ := b.Center.Z() + b.Extent.Z()
	z := s.Position.Z()
	if z < minZ {
		d := z - minZ
		sqDist += d * d
	} else if z > maxZ {
		d := z - maxZ
		sqDist += d * d
	}
	return sqDist <= s.Radius*s.Radius
}

func (s Sphere) IntersectsOOBB(b OOBB) bool {
	p := s.Position.Subtract(b.Center)
	local := b.Orientation.Transpose().MultiplyVec3(p)
	clamped := matrix.NewVec3(
		klib.Clamp(local.X(), -b.Extent.X(), b.Extent.X()),
		klib.Clamp(local.Y(), -b.Extent.Y(), b.Extent.Y()),
		klib.Clamp(local.Z(), -b.Extent.Z(), b.Extent.Z()),
	)
	diff := local.Subtract(clamped)
	distSq := diff.X()*diff.X() + diff.Y()*diff.Y() + diff.Z()*diff.Z()
	return distSq <= s.Radius*s.Radius
}

func (s Sphere) IntersectsRay(r Ray) (bool, float32) {
	m := r.Origin.Subtract(s.Position)
	a := r.Direction.X()*r.Direction.X() +
		r.Direction.Y()*r.Direction.Y() +
		r.Direction.Z()*r.Direction.Z()
	b := 2.0 * (r.Direction.X()*m.X() + r.Direction.Y()*m.Y() + r.Direction.Z()*m.Z())
	c := m.X()*m.X() + m.Y()*m.Y() + m.Z()*m.Z() - s.Radius*s.Radius
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return false, 0
	}
	sqrtDisc := float32(matrix.Sqrt(discriminant))
	t := (-b - sqrtDisc) / (2 * a)
	if t < 0 {
		t = (-b + sqrtDisc) / (2 * a)
		if t < 0 {
			return false, 0
		}
	}
	return true, t
}

func (s Sphere) IntersectsPlane(p Plane) (bool, float32) {
	dist := matrix.Vec3Dot(s.Position, p.Normal) + p.Dot
	if dist < 0 {
		dist = -dist
	}
	if dist <= s.Radius {
		return true, s.Radius - dist
	}
	return false, 0
}

func (s Sphere) IntersectsFrustum(f Frustum) bool {
	for _, p := range f.Planes {
		dist := matrix.Vec3Dot(s.Position, p.Normal) + p.Dot
		if dist < -s.Radius {
			return false
		}
	}
	return true
}

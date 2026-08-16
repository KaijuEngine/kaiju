/******************************************************************************/
/* sphere.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
)

type Sphere Shape

func (s *Shape) SetSphere(center matrix.Vec3, radius matrix.Float) {
	s.Type = ShapeTypeSphere
	s.Center = center
	s.Radius = radius
}

func NewSphere(center matrix.Vec3, radius matrix.Float) Sphere {
	s := Shape{}
	s.SetSphere(center, radius)
	return Sphere(s)
}

func (a Sphere) IntersectsSphere(b Sphere) bool {
	distSq := a.Center.SquareDistance(b.Center)
	radiusSum := a.Radius + b.Radius
	radiusSumSq := radiusSum * radiusSum
	return distSq <= radiusSumSq
}

func (s Sphere) IntersectsAABB(b AABB) bool {
	var sqDist matrix.Float
	minX := b.Center.X() - b.Extent.X()
	maxX := b.Center.X() + b.Extent.X()
	x := s.Center.X()
	if x < minX {
		d := x - minX
		sqDist += d * d
	} else if x > maxX {
		d := x - maxX
		sqDist += d * d
	}
	minY := b.Center.Y() - b.Extent.Y()
	maxY := b.Center.Y() + b.Extent.Y()
	y := s.Center.Y()
	if y < minY {
		d := y - minY
		sqDist += d * d
	} else if y > maxY {
		d := y - maxY
		sqDist += d * d
	}
	minZ := b.Center.Z() - b.Extent.Z()
	maxZ := b.Center.Z() + b.Extent.Z()
	z := s.Center.Z()
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
	p := s.Center.Subtract(b.Center)
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

func (s Sphere) IntersectsRay(r Ray) (bool, matrix.Float) {
	m := r.Origin.Subtract(s.Center)
	a := r.Direction.X()*r.Direction.X() +
		r.Direction.Y()*r.Direction.Y() +
		r.Direction.Z()*r.Direction.Z()
	b := 2.0 * (r.Direction.X()*m.X() + r.Direction.Y()*m.Y() + r.Direction.Z()*m.Z())
	c := m.X()*m.X() + m.Y()*m.Y() + m.Z()*m.Z() - s.Radius*s.Radius
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return false, 0
	}
	sqrtDisc := matrix.Float(matrix.Sqrt(discriminant))
	t := (-b - sqrtDisc) / (2 * a)
	if t < 0 {
		t = (-b + sqrtDisc) / (2 * a)
		if t < 0 {
			return false, 0
		}
	}
	return true, t
}

func (s Sphere) IntersectsPlane(p Plane) (bool, matrix.Float) {
	dist := matrix.Vec3Dot(s.Center, p.Normal) + p.Dot
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
		dist := matrix.Vec3Dot(s.Center, p.Normal) + p.Dot
		if dist < -s.Radius {
			return false
		}
	}
	return true
}

func (s Sphere) IntersectsCone(c Cone) bool {
	dir := s.Center.Subtract(c.Center)
	t := matrix.Vec3Dot(dir, c.Direction)
	halfH := c.Height / 2
	if t < -halfH {
		t = -halfH
	} else if t > halfH {
		t = halfH
	}
	ratio := (t + halfH) / c.Height
	radiusAtH := c.Radius * ratio
	axisPt := c.Center.Add(c.Direction.Scale(t))
	closest := axisPt
	if radiusAtH > 0 {
		perp := dir.Subtract(c.Direction.Scale(t))
		perpLen := perp.Length()
		if perpLen > radiusAtH {
			closest = axisPt.Add(perp.Scale(radiusAtH / perpLen))
		}
	}
	distSq := s.Center.Subtract(closest).LengthSquared()
	return distSq <= s.Radius*s.Radius
}

func (s Sphere) IntersectsCapsule(c Capsule) bool {
	halfH := c.Height / 2
	a1 := c.Center.Subtract(c.Direction.Scale(halfH))
	a2 := c.Center.Add(c.Direction.Scale(halfH))
	seg := a2.Subtract(a1)
	toCenter := s.Center.Subtract(a1)
	t := matrix.Vec3Dot(toCenter, seg) / matrix.Vec3Dot(seg, seg)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	closest := a1.Add(seg.Scale(t))
	distSq := s.Center.Subtract(closest).LengthSquared()
	rSum := s.Radius + c.Radius
	return distSq <= rSum*rSum
}

func (s Sphere) IntersectsCylinder(c Cylinder) bool {
	dir := s.Center.Subtract(c.Center)
	t := matrix.Vec3Dot(dir, c.Direction)
	halfH := c.Height / 2
	if t < -halfH {
		t = -halfH
	} else if t > halfH {
		t = halfH
	}
	axisPt := c.Center.Add(c.Direction.Scale(t))
	perp := dir.Subtract(c.Direction.Scale(t))
	perpLen := perp.Length()
	closest := axisPt
	if perpLen > c.Radius {
		closest = axisPt.Add(perp.Scale(c.Radius / perpLen))
	}
	distSq := s.Center.Subtract(closest).LengthSquared()
	return distSq <= s.Radius*s.Radius
}

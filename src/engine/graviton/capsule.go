/******************************************************************************/
/* capsule.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

type Capsule Shape

func (s *Shape) SetCapsule(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) {
	s.Type = ShapeTypeCapsule
	s.Center = center
	s.Radius = radius
	s.Height = height
	s.Direction = safeNormal(direction, matrix.Vec3Up())
}

func NewCapsule(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) Capsule {
	s := Shape{}
	s.SetCapsule(center, radius, height, direction)
	return Capsule(s)
}

func (s Capsule) IntersectsCapsule(b Capsule) bool {
	halfA := s.Height / 2
	halfB := b.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfA))
	a2 := s.Center.Add(s.Direction.Scale(halfA))
	b1 := b.Center.Subtract(b.Direction.Scale(halfB))
	b2 := b.Center.Add(b.Direction.Scale(halfB))
	d := a2.Subtract(a1)
	e := b2.Subtract(b1)
	r := a1.Subtract(b1)
	A := d.Dot(d)
	B := d.Dot(r)
	E := e.Dot(e)
	D := d.Dot(e)
	F := e.Dot(r)
	denom := A*E - D*D
	var segS, segT matrix.Float
	if denom < 1 {
		segS = 0
		segT = -F / E
		if segT < 0 {
			segT = 0
		} else if segT > 1 {
			segT = 1
		}
	} else {
		segS = (B*E - D*F) / denom
		segT = (A*F - D*B) / denom
		if segS < 0 {
			segS = 0
			segT = -F / E
			if segT < 0 {
				segT = 0
			} else if segT > 1 {
				segT = 1
			}
		} else if segS > 1 {
			segS = 1
			segT = (F - D) / E
			if segT < 0 {
				segT = 0
			} else if segT > 1 {
				segT = 1
			}
		} else {
			if segT < 0 {
				segT = 0
			} else if segT > 1 {
				segT = 1
			}
		}
	}
	closest := r.Add(d.Scale(segS)).Add(e.Scale(segT))
	distSq := closest.Dot(closest)
	rSum := s.Radius + b.Radius
	return distSq <= rSum*rSum
}

func (s Capsule) IntersectsAABB(b AABB) bool {
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	if b.Contains(a1) || b.Contains(a2) {
		return true
	}
	e := b.Extent.Add(matrix.NewVec3(s.Radius, s.Radius, s.Radius))
	dir := a2.Subtract(a1)
	min := b.Center.Subtract(e)
	max := b.Center.Add(e)
	if a1.X() >= min.X() && a1.X() <= max.X() && a1.Y() >= min.Y() && a1.Y() <= max.Y() && a1.Z() >= min.Z() && a1.Z() <= max.Z() {
		return true
	}
	if a2.X() >= min.X() && a2.X() <= max.X() && a2.Y() >= min.Y() && a2.Y() <= max.Y() && a2.Z() >= min.Z() && a2.Z() <= max.Z() {
		return true
	}
	planes := [6]struct {
		norm matrix.Vec3
		dot  matrix.Float
	}{
		{norm: matrix.NewVec3(-1, 0, 0), dot: b.Center.X() - e.X()},
		{norm: matrix.NewVec3(1, 0, 0), dot: -b.Center.X() - e.X()},
		{norm: matrix.NewVec3(0, -1, 0), dot: b.Center.Y() - e.Y()},
		{norm: matrix.NewVec3(0, 1, 0), dot: -b.Center.Y() - e.Y()},
		{norm: matrix.NewVec3(0, 0, -1), dot: b.Center.Z() - e.Z()},
		{norm: matrix.NewVec3(0, 0, 1), dot: -b.Center.Z() - e.Z()},
	}
	for i := range planes {
		p := &planes[i]
		d := matrix.Vec3Dot(p.norm, dir)
		if matrix.Abs(d) < 0.00001 {
			continue
		}
		d0 := matrix.Vec3Dot(p.norm, a1) + p.dot
		d1 := matrix.Vec3Dot(p.norm, a2) + p.dot
		if (d0 > 0) != (d1 > 0) {
			t := -d0 / d
			if t < 0 || t > 1 {
				continue
			}
			hit := a1.Add(dir.Scale(t))
			if hit.X() >= min.X() && hit.X() <= max.X() &&
				hit.Y() >= min.Y() && hit.Y() <= max.Y() &&
				hit.Z() >= min.Z() && hit.Z() <= max.Z() {
				return true
			}
		}
	}
	return false
}

func (s Capsule) IntersectsOOBB(b OOBB) bool {
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	localA1 := b.Orientation.Transpose().MultiplyVec3(a1.Subtract(b.Center))
	localA2 := b.Orientation.Transpose().MultiplyVec3(a2.Subtract(b.Center))
	localB := NewAABB(matrix.Vec3Zero(), b.Extent)
	if localB.Contains(localA1) || localB.Contains(localA2) {
		return true
	}
	e := b.Extent.Add(matrix.NewVec3(s.Radius, s.Radius, s.Radius))
	dir := localA2.Subtract(localA1)
	min := e.Negative()
	max := e
	if localA1.X() >= min.X() && localA1.X() <= max.X() && localA1.Y() >= min.Y() && localA1.Y() <= max.Y() && localA1.Z() >= min.Z() && localA1.Z() <= max.Z() {
		return true
	}
	if localA2.X() >= min.X() && localA2.X() <= max.X() && localA2.Y() >= min.Y() && localA2.Y() <= max.Y() && localA2.Z() >= min.Z() && localA2.Z() <= max.Z() {
		return true
	}
	planes := [6]struct {
		norm matrix.Vec3
		dot  matrix.Float
	}{
		{norm: matrix.NewVec3(-1, 0, 0), dot: -min.X()},
		{norm: matrix.NewVec3(1, 0, 0), dot: -max.X()},
		{norm: matrix.NewVec3(0, -1, 0), dot: -min.Y()},
		{norm: matrix.NewVec3(0, 1, 0), dot: -max.Y()},
		{norm: matrix.NewVec3(0, 0, -1), dot: -min.Z()},
		{norm: matrix.NewVec3(0, 0, 1), dot: -max.Z()},
	}
	for i := range planes {
		p := &planes[i]
		d := matrix.Vec3Dot(p.norm, dir)
		if matrix.Abs(d) < 0.00001 {
			continue
		}
		d0 := matrix.Vec3Dot(p.norm, localA1) + p.dot
		d1 := matrix.Vec3Dot(p.norm, localA2) + p.dot
		if (d0 > 0) != (d1 > 0) {
			t := -d0 / d
			if t < 0 || t > 1 {
				continue
			}
			hit := localA1.Add(dir.Scale(t))
			if hit.X() >= min.X() && hit.X() <= max.X() &&
				hit.Y() >= min.Y() && hit.Y() <= max.Y() &&
				hit.Z() >= min.Z() && hit.Z() <= max.Z() {
				return true
			}
		}
	}
	return false
}

func (s Capsule) IntersectsRay(r Ray) (bool, float32) {
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	spine := a2.Subtract(a1)
	spineLenSq := spine.Dot(spine)
	radiusSq := s.Radius * s.Radius
	var minT matrix.Float = matrix.Inf(1)
	f := r.Origin.Subtract(a1)
	cross := matrix.Vec3Cross(r.Direction, spine)
	oc := matrix.Vec3Cross(f, spine)
	A := cross.Dot(cross)
	B := 2 * oc.Dot(cross)
	C := oc.Dot(oc) - radiusSq*spineLenSq
	if A > 0.00001 {
		disc := B*B - 4*A*C
		if disc >= 0 {
			sqrtDisc := matrix.Sqrt(disc)
			for _, sign := range []matrix.Float{-1, 1} {
				t := (-B + sign*sqrtDisc) / (2 * A)
				if t < 0 {
					continue
				}
				hit := r.Origin.Add(r.Direction.Scale(t))
				vec := hit.Subtract(a1)
				segT := vec.Dot(spine) / spineLenSq
				if segT >= 0 && segT <= 1 && t < minT {
					minT = t
				}
			}
		}
	}
	{
		oc2 := r.Origin.Subtract(a1)
		a := r.Direction.Dot(r.Direction)
		b := 2 * oc2.Dot(r.Direction)
		c := oc2.Dot(oc2) - radiusSq
		disc := b*b - 4*a*c
		if disc >= 0 {
			sqrtDisc := matrix.Sqrt(disc)
			for _, sign := range []matrix.Float{-1, 1} {
				t := (-b + sign*sqrtDisc) / (2 * a)
				if t < 0 {
					continue
				}
				hit := r.Origin.Add(r.Direction.Scale(t))
				if hit.Subtract(a1).Dot(spine) <= 0 && t < minT {
					minT = t
				}
			}
		}
	}
	{
		oc2 := r.Origin.Subtract(a2)
		a := r.Direction.Dot(r.Direction)
		b := 2 * oc2.Dot(r.Direction)
		c := oc2.Dot(oc2) - radiusSq
		disc := b*b - 4*a*c
		if disc >= 0 {
			sqrtDisc := matrix.Sqrt(disc)
			for _, sign := range []matrix.Float{-1, 1} {
				t := (-b + sign*sqrtDisc) / (2 * a)
				if t < 0 {
					continue
				}
				hit := r.Origin.Add(r.Direction.Scale(t))
				if hit.Subtract(a2).Dot(spine) >= 0 && t < minT {
					minT = t
				}
			}
		}
	}
	if minT < matrix.Inf(1) {
		return true, float32(minT)
	}
	return false, 0
}

func (s Capsule) IntersectsPlane(p Plane) (bool, float32) {
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	d1 := matrix.Vec3Dot(p.Normal, a1) - p.Dot
	d2 := matrix.Vec3Dot(p.Normal, a2) - p.Dot
	var dist matrix.Float
	if (d1 > 0) == (d2 > 0) {
		dist = matrix.Abs(d1)
		if matrix.Abs(d2) < dist {
			dist = matrix.Abs(d2)
		}
	}
	if dist <= s.Radius {
		return true, float32(s.Radius - dist)
	}
	return false, 0
}

func (s Capsule) IntersectsFrustum(f Frustum) bool {
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	for i := range f.Planes {
		p := f.Planes[i]
		d1 := matrix.Vec3Dot(p.Normal, a1) + p.Dot
		d2 := matrix.Vec3Dot(p.Normal, a2) + p.Dot
		if d1 < -s.Radius && d2 < -s.Radius {
			return false
		}
	}
	return true
}

func (s Capsule) IntersectsSphere(c Sphere) bool {
	return c.IntersectsCapsule(s)
}

func (s Capsule) IntersectsCone(c Cone) bool {
	aBound := matrix.Sqrt(s.Radius*s.Radius + (s.Height/2)*(s.Height/2))
	cBound := matrix.Sqrt(c.Radius*c.Radius + (c.Height/2)*(c.Height/2))
	if s.Center.Subtract(c.Center).LengthSquared() > (aBound+cBound)*(aBound+cBound) {
		return false
	}
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	if pointInCone(a1, c) || pointInCone(a2, c) {
		return true
	}
	coneApex := c.Center.Subtract(c.Direction.Scale(c.Height / 2))
	coneBase := c.Center.Add(c.Direction.Scale(c.Height / 2))
	if capsuleContainsPoint(a1, a2, s.Radius, coneApex) || capsuleContainsPoint(a1, a2, s.Radius, coneBase) {
		return true
	}
	if capsuleSpineVsConeAxis(s, c) {
		return true
	}
	if capsuleSpinePenetratesCone(s, c) {
		return true
	}
	if coneBaseCircleVsCapsule(s, c) {
		return true
	}
	return false
}

func (s Capsule) IntersectsCylinder(c Cylinder) bool {
	aBound := matrix.Sqrt(s.Radius*s.Radius + (s.Height/2)*(s.Height/2))
	cBound := matrix.Sqrt(c.Radius*c.Radius + (c.Height/2)*(c.Height/2))
	if s.Center.Subtract(c.Center).LengthSquared() > (aBound+cBound)*(aBound+cBound) {
		return false
	}
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	if pointInCylinder(a1, c) || pointInCylinder(a2, c) {
		return true
	}
	cylBottom := c.Center.Subtract(c.Direction.Scale(c.Height / 2))
	cylTop := c.Center.Add(c.Direction.Scale(c.Height / 2))
	if capsuleContainsPoint(a1, a2, s.Radius, cylBottom) || capsuleContainsPoint(a1, a2, s.Radius, cylTop) {
		return true
	}
	if capsuleSpineVsCylinderAxis(s, c) {
		return true
	}
	if capsuleSpinePenetratesCylinder(s, c) {
		return true
	}
	if cylinderEdgeCircleVsCapsule(s, c) {
		return true
	}
	return false
}

func pointInCylinder(p matrix.Vec3, c Cylinder) bool {
	dir := p.Subtract(c.Center)
	t := dir.Dot(c.Direction)
	halfH := c.Height / 2
	if t < -halfH || t > halfH {
		return false
	}
	perp := dir.Subtract(c.Direction.Scale(t))
	return perp.LengthSquared() <= c.Radius*c.Radius
}

func capsuleSpineVsCylinderAxis(s Capsule, c Cylinder) bool {
	halfHC := c.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(s.Height / 2))
	a2 := s.Center.Add(s.Direction.Scale(s.Height / 2))
	cylBottom := c.Center.Subtract(c.Direction.Scale(halfHC))
	cylTop := c.Center.Add(c.Direction.Scale(halfHC))
	d := a2.Subtract(a1)
	e := cylTop.Subtract(cylBottom)
	r := a1.Subtract(cylBottom)
	A := d.Dot(d)
	D := d.Dot(e)
	E := e.Dot(e)
	denom := A*E - D*D
	var segS, segT matrix.Float
	if denom < 1 {
		segS = 0
		segT = 0
	} else {
		B := d.Dot(r)
		F := e.Dot(r)
		segS = (B*E - D*F) / denom
		segT = (A*F - D*B) / denom
		if segS < 0 {
			segS = 0
			segT = min(max(-F/E, 0), 1)
		} else if segS > 1 {
			segS = 1
			segT = min(max((F-D)/E, 0), 1)
		} else {
			segT = min(max(segT, 0), 1)
		}
	}
	closest := r.Add(d.Scale(segS)).Add(e.Scale(segT))
	distSq := closest.Dot(closest)
	rSum := s.Radius + c.Radius
	if distSq <= rSum*rSum {
		return true
	}
	return false
}

func capsuleSpinePenetratesCylinder(s Capsule, c Cylinder) bool {
	halfHC := c.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(s.Height / 2))
	a2 := s.Center.Add(s.Direction.Scale(s.Height / 2))
	spine := a2.Subtract(a1)
	for i := 0; i <= 11; i++ {
		t := matrix.Float(i) / 10
		pt := a1.Add(spine.Scale(t))
		diff := pt.Subtract(c.Center)
		h := diff.Dot(c.Direction)
		h = min(max(h, -halfHC), halfHC)
		axisPt := c.Center.Add(c.Direction.Scale(h))
		perp := pt.Subtract(axisPt)
		distToAxis := perp.Length()
		if distToAxis > c.Radius {
			distToSurface := distToAxis - c.Radius
			if distToSurface <= s.Radius {
				return true
			}
		}
	}
	return false
}

func cylinderEdgeCircleVsCapsule(s Capsule, c Cylinder) bool {
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	for _, sign := range []matrix.Float{-1, 1} {
		ringCenter := c.Center.Add(c.Direction.Scale(sign * c.Height / 2))
		perpX := c.Direction.Orthogonal()
		perpY := matrix.Vec3Cross(c.Direction, perpX).Normal()
		for i := 0; i < 16; i++ {
			angle := matrix.Float(i) * matrix.Float(6.283185307179586) / 16
			pt := ringCenter.Add(perpX.Scale(c.Radius * matrix.Cos(angle))).Add(perpY.Scale(c.Radius * matrix.Sin(angle)))
			seg := a2.Subtract(a1)
			segLenSq := seg.Dot(seg)
			toPt := pt.Subtract(a1)
			var t matrix.Float
			if segLenSq > 0 {
				t = toPt.Dot(seg) / segLenSq
			}
			t = min(max(t, matrix.Float(0)), matrix.Float(1))
			closest := a1.Add(seg.Scale(t))
			if pt.Subtract(closest).LengthSquared() <= s.Radius*s.Radius {
				return true
			}
		}
	}
	return false
}

func capsuleContainsPoint(a1, a2 matrix.Vec3, radius matrix.Float, p matrix.Vec3) bool {
	seg := a2.Subtract(a1)
	segLenSq := seg.Dot(seg)
	toP := p.Subtract(a1)
	var t matrix.Float
	if segLenSq > 0 {
		t = toP.Dot(seg) / segLenSq
	}
	t = min(max(t, matrix.Float(0)), matrix.Float(1))
	closest := a1.Add(seg.Scale(t))
	return p.Subtract(closest).LengthSquared() <= radius*radius
}

func capsuleSpineVsConeAxis(s Capsule, c Cone) bool {
	halfHC := c.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(s.Height / 2))
	a2 := s.Center.Add(s.Direction.Scale(s.Height / 2))
	coneApex := c.Center.Subtract(c.Direction.Scale(halfHC))
	coneBase := c.Center.Add(c.Direction.Scale(halfHC))
	d := a2.Subtract(a1)
	e := coneBase.Subtract(coneApex)
	r := a1.Subtract(coneApex)
	A := d.Dot(d)
	D := d.Dot(e)
	E := e.Dot(e)
	denom := A*E - D*D
	var segS, segT matrix.Float
	if denom < 1 {
		segS = 0
		segT = 0
	} else {
		B := d.Dot(r)
		F := e.Dot(r)
		segS = (B*E - D*F) / denom
		segT = (A*F - D*B) / denom
		if segS < 0 {
			segS = 0
			segT = min(max(-F/E, 0), 1)
		} else if segS > 1 {
			segS = 1
			segT = min(max((F-D)/E, 0), 1)
		} else {
			segT = min(max(segT, 0), 1)
		}
	}
	closest := r.Add(d.Scale(segS)).Add(e.Scale(segT))
	distSq := closest.Dot(closest)
	ratio := min(max(segT, matrix.Float(0)), matrix.Float(1))
	radiusAtH := c.Radius * ratio
	if distSq <= (s.Radius+radiusAtH)*(s.Radius+radiusAtH) {
		return true
	}
	return false
}

func capsuleSpinePenetratesCone(s Capsule, c Cone) bool {
	halfHC := c.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(s.Height / 2))
	a2 := s.Center.Add(s.Direction.Scale(s.Height / 2))
	spine := a2.Subtract(a1)
	for i := 0; i <= 11; i++ {
		t := matrix.Float(i) / 10
		pt := a1.Add(spine.Scale(t))
		diff := pt.Subtract(c.Center)
		h := diff.Dot(c.Direction)
		h = min(max(h, -halfHC), halfHC)
		ratio := (h + halfHC) / c.Height
		radiusAtH := c.Radius * ratio
		axisPt := c.Center.Add(c.Direction.Scale(h))
		perp := pt.Subtract(axisPt)
		distToAxis := perp.Length()
		if distToAxis > radiusAtH {
			distToSurface := distToAxis - radiusAtH
			if distToSurface <= s.Radius {
				return true
			}
		}
	}
	return false
}

func coneBaseCircleVsCapsule(s Capsule, c Cone) bool {
	halfH := s.Height / 2
	a1 := s.Center.Subtract(s.Direction.Scale(halfH))
	a2 := s.Center.Add(s.Direction.Scale(halfH))
	coneBase := c.Center.Add(c.Direction.Scale(c.Height / 2))
	perpX := c.Direction.Orthogonal()
	perpY := matrix.Vec3Cross(c.Direction, perpX).Normal()
	for i := 0; i < 16; i++ {
		angle := matrix.Float(i) * matrix.Float(6.283185307179586) / 16
		pt := coneBase.Add(perpX.Scale(c.Radius * matrix.Cos(angle))).Add(perpY.Scale(c.Radius * matrix.Sin(angle)))
		seg := a2.Subtract(a1)
		segLenSq := seg.Dot(seg)
		toPt := pt.Subtract(a1)
		var t matrix.Float
		if segLenSq > 0 {
			t = toPt.Dot(seg) / segLenSq
		}
		t = min(max(t, matrix.Float(0)), matrix.Float(1))
		closest := a1.Add(seg.Scale(t))
		if pt.Subtract(closest).LengthSquared() <= s.Radius*s.Radius {
			return true
		}
	}
	return false
}

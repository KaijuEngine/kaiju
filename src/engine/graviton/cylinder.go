/******************************************************************************/
/* cylinder.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"math"

	"kaijuengine.com/matrix"
)

type Cylinder Shape

func (s *Shape) SetCylinder(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) {
	s.Type = ShapeTypeCylinder
	s.Center = center
	s.Radius = radius
	s.Height = height
	s.Direction = safeNormal(direction, matrix.Vec3Up())
}

func NewCylinder(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) Cylinder {
	s := Shape{}
	s.SetCylinder(center, radius, height, direction)
	return Cylinder(s)
}

func (a Cylinder) IntersectsCylinder(b Cylinder) bool {
	// Quick rejection: bounding sphere check
	centerDiff := a.Center.Subtract(b.Center)
	distSq := matrix.Vec3Dot(centerDiff, centerDiff)
	// Bounding sphere radius = max(radius, height/2)
	aBound := a.Radius
	if a.Height/2 > aBound {
		aBound = a.Height / 2
	}
	bBound := b.Radius
	if b.Height/2 > bBound {
		bBound = b.Height / 2
	}
	maxDist := aBound + bBound
	if distSq > maxDist*maxDist {
		return false
	}
	// Axis-aligned rejection: check if cylinders overlap along each cylinder's axis
	// Project both centers onto cylinder A's axis
	aProj := matrix.Vec3Dot(centerDiff, a.Direction)
	if aProj > (a.Height+b.Height)/2 || aProj < -(a.Height+b.Height)/2 {
		return false
	}
	// Project both centers onto cylinder B's axis
	bProj := matrix.Vec3Dot(centerDiff, b.Direction)
	if bProj > (a.Height+b.Height)/2 || bProj < -(a.Height+b.Height)/2 {
		return false
	}
	// Radial check: perpendicular distance between axes
	// Cross product of directions gives perpendicular component
	cross := matrix.Vec3Cross(a.Direction, b.Direction)
	crossLenSq := matrix.Vec3Dot(cross, cross)
	if crossLenSq < 1e-10 {
		// Parallel cylinders - check radial distance
		// Project centerDiff onto a perpendicular vector
		perp := centerDiff.Subtract(a.Direction.Scale(matrix.Vec3Dot(centerDiff, a.Direction)))
		perpDistSq := matrix.Vec3Dot(perp, perp)
		return perpDistSq <= (a.Radius+b.Radius)*(a.Radius+b.Radius)
	}
	// For non-parallel cylinders, use conservative radial check
	// This is a simplified approximation for performance
	return true
}

func (s Cylinder) IntersectsAABB(b AABB) bool {
	centerDiff := s.Center.Subtract(b.Center)
	distSq := matrix.Vec3Dot(centerDiff, centerDiff)
	sBound := s.Radius
	if s.Height/2 > sBound {
		sBound = s.Height / 2
	}
	bBound := b.Extent.X()
	if b.Extent.Y() > bBound {
		bBound = b.Extent.Y()
	}
	if b.Extent.Z() > bBound {
		bBound = b.Extent.Z()
	}
	maxDist := sBound + bBound
	if distSq > maxDist*maxDist {
		return false
	}
	halfH := s.Height / 2
	bottom := s.Center.Subtract(s.Direction.Scale(halfH))
	top := s.Center.Add(s.Direction.Scale(halfH))
	if b.Contains(bottom) || b.Contains(top) ||
		b.Contains(s.Center.Add(matrix.Vec3Right().Scale(s.Radius))) ||
		b.Contains(s.Center.Add(matrix.Vec3Up().Scale(s.Radius))) ||
		b.Contains(s.Center.Add(matrix.Vec3Forward().Scale(s.Radius))) {
		return true
	}
	minProj := matrix.Float(9e9)
	maxProj := matrix.Float(-9e9)
	corners := b.Corners()
	for i := range 8 {
		p := matrix.Vec3Dot(corners[i].Subtract(s.Center), s.Direction)
		if p < minProj {
			minProj = p
		}
		if p > maxProj {
			maxProj = p
		}
	}
	if minProj > halfH || maxProj < -halfH {
		return false
	}
	for i := range 8 {
		diff := corners[i].Subtract(s.Center)
		h := matrix.Vec3Dot(diff, s.Direction)
		if h > halfH || h < -halfH {
			continue
		}
		radial := diff.Subtract(s.Direction.Scale(h))
		if matrix.Vec3Dot(radial, radial) <= s.Radius*s.Radius {
			return true
		}
	}
	bMin := b.Min()
	bMax := b.Max()
	candidate := bMin
	if s.Direction.X() > 0 {
		candidate[0] = bMax[0]
	} else {
		candidate[0] = bMin[0]
	}
	if s.Direction.Y() > 0 {
		candidate[1] = bMax[1]
	} else {
		candidate[1] = bMin[1]
	}
	if s.Direction.Z() > 0 {
		candidate[2] = bMax[2]
	} else {
		candidate[2] = bMin[2]
	}
	diff := candidate.Subtract(s.Center)
	h := matrix.Vec3Dot(diff, s.Direction)
	if h > halfH || h < -halfH {
		h = 0
	}
	radial := diff.Subtract(s.Direction.Scale(h))
	if matrix.Vec3Dot(radial, radial) <= s.Radius*s.Radius {
		return true
	}
	for i := range 8 {
		c := corners[i]
		diff := c.Subtract(bottom)
		h := matrix.Vec3Dot(diff, s.Direction)
		if h >= 0 {
			radial := diff.Subtract(s.Direction.Scale(h))
			if matrix.Vec3Dot(radial, radial) <= s.Radius*s.Radius {
				return true
			}
		}
		diff = c.Subtract(top)
		h = matrix.Vec3Dot(diff, s.Direction)
		if h <= 0 {
			radial := diff.Subtract(s.Direction.Scale(h))
			if matrix.Vec3Dot(radial, radial) <= s.Radius*s.Radius {
				return true
			}
		}
	}
	return false
}

func (s Cylinder) IntersectsOOBB(b OOBB) bool {
	centerDiff := s.Center.Subtract(b.Center)
	distSq := matrix.Vec3Dot(centerDiff, centerDiff)
	sBound := s.Radius
	if s.Height/2 > sBound {
		sBound = s.Height / 2
	}
	bBound := b.Extent.X()
	if b.Extent.Y() > bBound {
		bBound = b.Extent.Y()
	}
	if b.Extent.Z() > bBound {
		bBound = b.Extent.Z()
	}
	maxDist := sBound + bBound
	if distSq > maxDist*maxDist {
		return false
	}
	halfH := s.Height / 2
	bottom := s.Center.Subtract(s.Direction.Scale(halfH))
	top := s.Center.Add(s.Direction.Scale(halfH))
	if b.ContainsPoint(bottom) || b.ContainsPoint(top) ||
		b.ContainsPoint(s.Center.Add(matrix.Vec3Right().Scale(s.Radius))) ||
		b.ContainsPoint(s.Center.Add(matrix.Vec3Up().Scale(s.Radius))) ||
		b.ContainsPoint(s.Center.Add(matrix.Vec3Forward().Scale(s.Radius))) {
		return true
	}
	minProj := matrix.Float(9e9)
	maxProj := matrix.Float(-9e9)
	corners := b.Corners()
	for i := range 8 {
		p := matrix.Vec3Dot(corners[i].Subtract(s.Center), s.Direction)
		if p < minProj {
			minProj = p
		}
		if p > maxProj {
			maxProj = p
		}
	}
	if minProj > halfH || maxProj < -halfH {
		return false
	}
	for i := range 8 {
		diff := corners[i].Subtract(s.Center)
		h := matrix.Vec3Dot(diff, s.Direction)
		if h > halfH || h < -halfH {
			continue
		}
		radial := diff.Subtract(s.Direction.Scale(h))
		if matrix.Vec3Dot(radial, radial) <= s.Radius*s.Radius {
			return true
		}
	}
	for i := range 8 {
		c := corners[i]
		diff := c.Subtract(bottom)
		h := matrix.Vec3Dot(diff, s.Direction)
		if h >= 0 {
			radial := diff.Subtract(s.Direction.Scale(h))
			if matrix.Vec3Dot(radial, radial) <= s.Radius*s.Radius {
				return true
			}
		}
		diff = c.Subtract(top)
		h = matrix.Vec3Dot(diff, s.Direction)
		if h <= 0 {
			radial := diff.Subtract(s.Direction.Scale(h))
			if matrix.Vec3Dot(radial, radial) <= s.Radius*s.Radius {
				return true
			}
		}
	}
	return false
}

func (s Cylinder) IntersectsRay(r Ray) (bool, float32) {
	oc := r.Origin.Subtract(s.Center)
	oe := matrix.Vec3Dot(oc, s.Direction)
	re := matrix.Vec3Dot(r.Direction, s.Direction)
	oh := oc.Subtract(s.Direction.Scale(oe))
	rh := r.Direction.Subtract(s.Direction.Scale(re))
	he := s.Height / 2
	ee := re*re - 1
	if matrix.Approx(ee, 0) {
		if matrix.Approx(rh.Length(), 0) {
			if matrix.Approx(oh.Length()-s.Radius, 0) {
				t := -oe / re
				if t >= 0 && matrix.Abs(oe) <= he {
					return true, float32(t)
				}
			}
			return false, 0
		}
		oo := oh.LengthSquared() - s.Radius*s.Radius
		if matrix.Approx(oo, 0) {
			t := matrix.Float(oo) / (-rh.LengthSquared())
			if t >= 0 && matrix.Abs(oe+re*t) <= he {
				return true, float32(t)
			}
		}
		return false, 0
	}
	oo := oh.LengthSquared() - s.Radius*s.Radius
	oe *= re
	ho := matrix.Vec3Dot(oh, rh)
	dd := ho*ho - ee*oo
	if dd < 0 {
		return false, 0
	}
	ds := matrix.Sqrt(dd)
	t := min((-ho-ds)/ee, (-ho+ds)/ee)
	if t < 0 {
		t = max((-ho-ds)/ee, (-ho+ds)/ee)
		if t < 0 {
			return false, 0
		}
	}
	if matrix.Abs(oe+re*t) > he {
		return false, 0
	}
	return true, float32(t)
}

func (s Cylinder) IntersectsPlane(p Plane) (bool, float32) {
	halfH := s.Height / 2
	top := s.Center.Add(s.Direction.Scale(halfH))
	bottom := s.Center.Subtract(s.Direction.Scale(halfH))
	distTop := matrix.Vec3Dot(top, p.Normal) + p.Dot
	distBottom := matrix.Vec3Dot(bottom, p.Normal) + p.Dot
	if (distTop >= 0 && distBottom <= 0) || (distTop <= 0 && distBottom >= 0) {
		maxDist := s.Radius + halfH
		centerDist := matrix.Vec3Dot(s.Center, p.Normal) + p.Dot
		if centerDist < 0 {
			centerDist = -centerDist
		}
		return true, float32(maxDist - centerDist)
	}
	capDist := distTop
	if capDist < 0 {
		capDist = -capDist
	}
	if capDist <= s.Radius {
		return true, float32(s.Radius - capDist)
	}
	capDist = distBottom
	if capDist < 0 {
		capDist = -capDist
	}
	if capDist <= s.Radius {
		return true, float32(s.Radius - capDist)
	}
	ne := matrix.Vec3Dot(p.Normal, s.Direction)
	if ne*ne < 1e-10 {
		return false, 0
	}
	num := matrix.Vec3Dot(p.Normal, s.Center) + p.Dot
	t := num / (ne * ne)
	closest := s.Center.Subtract(s.Direction.Scale(t))
	proj := matrix.Vec3Dot(closest.Subtract(s.Center), s.Direction)
	if proj > halfH || proj < -halfH {
		return false, 0
	}
	closestDist := matrix.Vec3Dot(closest, p.Normal) + p.Dot
	if closestDist < 0 {
		closestDist = -closestDist
	}
	if closestDist <= s.Radius {
		return true, float32(s.Radius - closestDist)
	}
	return false, 0
}

func (s Cylinder) IntersectsFrustum(f Frustum) bool {
	maxDist := matrix.Sqrt(s.Radius*s.Radius + s.Height*s.Height/4)
	for _, p := range f.Planes {
		dist := matrix.Vec3Dot(s.Center, p.Normal) + p.Dot
		if dist < -maxDist {
			return false
		}
	}
	return true
}

func (s Cylinder) IntersectsSphere(c Sphere) bool {
	return c.IntersectsCylinder(s)
}

func (s Cylinder) IntersectsCapsule(c Capsule) bool {
	return c.IntersectsCylinder(s)
}

func (s Cylinder) IntersectsCone(c Cone) bool {
	aBound := matrix.Sqrt(s.Radius*s.Radius + (s.Height/2)*(s.Height/2))
	bBound := matrix.Sqrt(c.Radius*c.Radius + (c.Height/2)*(c.Height/2))
	centerDiff := s.Center.Subtract(c.Center)
	if matrix.Vec3Dot(centerDiff, centerDiff) > (aBound+bBound)*(aBound+bBound) {
		return false
	}
	coneApex := c.Center.Subtract(c.Direction.Scale(c.Height / 2))
	coneBase := c.Center.Add(c.Direction.Scale(c.Height / 2))
	if pointInCylinder(coneApex, s) || pointInCylinder(coneBase, s) {
		return true
	}
	cylBottom := s.Center.Subtract(s.Direction.Scale(s.Height / 2))
	cylTop := s.Center.Add(s.Direction.Scale(s.Height / 2))
	if pointInCone(cylBottom, c) || pointInCone(cylTop, c) {
		return true
	}
	if pointInCone(s.Center, c) {
		return true
	}
	for i := matrix.Float(0); i <= 1; i += matrix.Float(0.25) {
		pt := s.Center.Add(s.Direction.Scale((i - matrix.Float(0.5)) * s.Height))
		if pointInCone(pt, c) {
			return true
		}
	}
	steps := 8
	for i := 0; i < steps; i++ {
		angle := matrix.Float(i) * 2 * matrix.Float(math.Pi) / matrix.Float(steps)
		cosA := matrix.Cos(angle)
		sinA := matrix.Sin(angle)
		perpX := matrix.Vec3Right()
		dot := matrix.Vec3Dot(perpX, c.Direction)
		if matrix.Abs(dot) > matrix.Float(0.9) {
			perpX = matrix.Vec3Up()
		}
		perpX = perpX.Subtract(c.Direction.Scale(matrix.Vec3Dot(perpX, c.Direction)))
		perpX = perpX.Scale(matrix.Float(1) / perpX.Length())
		perpY := matrix.Vec3Cross(c.Direction, perpX)
		pt := coneBase.Add(perpX.Scale(c.Radius * cosA)).Add(perpY.Scale(c.Radius * sinA))
		if pointInCylinder(pt, s) {
			return true
		}
	}
	for i := matrix.Float(0); i <= 1; i += matrix.Float(0.25) {
		t := (i - matrix.Float(0.5)) * c.Height
		pt := c.Center.Add(c.Direction.Scale(t))
		if pointInCylinder(pt, s) {
			return true
		}
	}
	return false
}

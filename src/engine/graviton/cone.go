/******************************************************************************/
/* cone.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

type Cone Shape

func (s *Shape) SetCone(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) {
	s.Type = ShapeTypeCone
	s.Center = center
	s.Radius = radius
	s.Height = height
	s.Direction = safeNormal(direction, matrix.Vec3Up())
}

func NewCone(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) Cone {
	s := Shape{}
	s.SetCone(center, radius, height, direction)
	return Cone(s)
}

func (a *Cone) IntersectCone(b Cone) bool {
	// Fast rejection: bounding sphere check
	// Cone bounding sphere radius = sqrt(radius^2 + (height/2)^2)
	aRadius := matrix.Sqrt(a.Radius*a.Radius + (a.Height/2)*(a.Height/2))
	bRadius := matrix.Sqrt(b.Radius*b.Radius + (b.Height/2)*(b.Height/2))
	dist := a.Center.Subtract(b.Center).Length()
	if dist > aRadius+bRadius {
		return false
	}
	// Check if apex of a is inside b
	if pointInCone(a.Center, b) {
		return true
	}
	// Check if apex of b is inside a
	if pointInCone(b.Center, *a) {
		return true
	}
	// Check if base circles intersect
	aBaseCenter := a.Center.Add(a.Direction.Scale(a.Height / 2))
	bBaseCenter := b.Center.Add(b.Direction.Scale(b.Height / 2))
	baseDist := aBaseCenter.Subtract(bBaseCenter).Length()
	if baseDist <= a.Radius+b.Radius {
		return true
	}
	// Check if axes intersect within both cones
	if axesIntersect(*a, b) {
		return true
	}
	return false
}

func (s Cone) IntersectsAABB(b AABB) bool {
	centerDiff := s.Center.Subtract(b.Center)
	distSq := matrix.Vec3Dot(centerDiff, centerDiff)
	sBound := matrix.Sqrt(s.Radius*s.Radius + (s.Height/2)*(s.Height/2))
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
	apex := s.Center.Subtract(s.Direction.Scale(halfH))
	baseCenter := s.Center.Add(s.Direction.Scale(halfH))
	if b.Contains(apex) || b.Contains(baseCenter) {
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
		if pointInCone(corners[i], s) {
			return true
		}
	}
	candidate := b.Min()
	for i := 0; i < 3; i++ {
		if s.Direction[i] > 0 {
			candidate[i] = b.Max()[i]
		}
	}
	diff := candidate.Subtract(s.Center)
	h := matrix.Vec3Dot(diff, s.Direction)
	if h < -halfH {
		h = -halfH
	}
	if h > halfH {
		h = halfH
	}
	ratio := (h + halfH) / s.Height
	radiusAtH := s.Radius * ratio
	radial := diff.Subtract(s.Direction.Scale(h))
	if matrix.Vec3Dot(radial, radial) <= radiusAtH*radiusAtH {
		return true
	}
	edgePairs := [][2]int{
		{0, 1}, {0, 2}, {0, 4}, {1, 3}, {1, 5}, {2, 3}, {2, 6}, {3, 7}, {4, 5}, {4, 6}, {5, 7}, {6, 7},
	}
	for _, edge := range edgePairs {
		p0 := corners[edge[0]]
		p1 := corners[edge[1]]
		if coneIntersectsSegment(s, p0, p1) {
			return true
		}
	}
	for i := range 8 {
		c := corners[i]
		diff := c.Subtract(apex)
		h := matrix.Vec3Dot(diff, s.Direction)
		if h >= 0 {
			ratio := h / s.Height
			radiusAtH := s.Radius * ratio
			radial := diff.Subtract(s.Direction.Scale(h))
			if matrix.Vec3Dot(radial, radial) <= radiusAtH*radiusAtH {
				return true
			}
		}
	}
	return false
}

func (s Cone) IntersectsOOBB(b OOBB) bool {
	centerDiff := s.Center.Subtract(b.Center)
	distSq := matrix.Vec3Dot(centerDiff, centerDiff)
	sBound := matrix.Sqrt(s.Radius*s.Radius + (s.Height/2)*(s.Height/2))
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
	apex := s.Center.Subtract(s.Direction.Scale(halfH))
	baseCenter := s.Center.Add(s.Direction.Scale(halfH))
	if b.ContainsPoint(apex) || b.ContainsPoint(baseCenter) {
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
		if pointInCone(corners[i], s) {
			return true
		}
	}
	edgePairs := [][2]int{
		{0, 1}, {0, 2}, {0, 4}, {1, 3}, {1, 5}, {2, 3}, {2, 6}, {3, 7}, {4, 5}, {4, 6}, {5, 7}, {6, 7},
	}
	for _, edge := range edgePairs {
		p0 := corners[edge[0]]
		p1 := corners[edge[1]]
		if coneIntersectsSegment(s, p0, p1) {
			return true
		}
	}
	for i := range 8 {
		c := corners[i]
		diff := c.Subtract(apex)
		h := matrix.Vec3Dot(diff, s.Direction)
		if h >= 0 {
			ratio := h / s.Height
			radiusAtH := s.Radius * ratio
			radial := diff.Subtract(s.Direction.Scale(h))
			if matrix.Vec3Dot(radial, radial) <= radiusAtH*radiusAtH {
				return true
			}
		}
	}
	return false
}

func (s Cone) IntersectsRay(r Ray) (bool, matrix.Float) {
	halfH := s.Height / 2
	tan2 := s.Radius * s.Radius / (s.Height * s.Height)
	oc := r.Origin.Subtract(s.Center)
	od := r.Direction
	ou := matrix.Vec3Dot(oc, s.Direction)
	od_u := matrix.Vec3Dot(od, s.Direction)
	opSq := oc.LengthSquared() - ou*ou
	odpSq := od.LengthSquared() - od_u*od_u
	ap := odpSq - tan2*od_u*od_u
	bp := 2 * (matrix.Vec3Dot(oc, od) - tan2*(ou*od_u+od_u*halfH))
	cp := opSq - tan2*(ou+halfH)*(ou+halfH)
	var tHit matrix.Float = matrix.Inf(1)
	hit := false
	if matrix.Abs(ap) > 1e-12 {
		disc := bp*bp - 4*ap*cp
		if disc >= 0 {
			sqrtDisc := matrix.Sqrt(disc)
			t0 := (-bp - sqrtDisc) / (2 * ap)
			t1 := (-bp + sqrtDisc) / (2 * ap)
			if t0 > t1 {
				t0, t1 = t1, t0
			}
			for _, t := range [2]matrix.Float{t0, t1} {
				if t < 0 {
					continue
				}
				h := ou + od_u*t
				if h < -halfH || h > halfH {
					continue
				}
				if t < tHit {
					tHit = t
					hit = true
				}
			}
		}
	} else if matrix.Abs(bp) > 1e-12 {
		t := -cp / bp
		if t >= 0 {
			h := ou + od_u*t
			if h >= -halfH && h <= halfH {
				if t < tHit {
					tHit = t
					hit = true
				}
			}
		}
	}
	// Check base cap (disk at base center, facing direction)
	baseCenter := s.Center.Add(s.Direction.Scale(halfH))
	ocBase := baseCenter.Subtract(r.Origin)
	den := matrix.Vec3Dot(s.Direction, od)
	if den > 1e-12 {
		t := matrix.Vec3Dot(ocBase, s.Direction) / den
		if t >= 0 && t < tHit {
			hitPt := r.Origin.Add(od.Scale(t)).Subtract(baseCenter)
			perp := hitPt.Subtract(s.Direction.Scale(matrix.Vec3Dot(hitPt, s.Direction)))
			if perp.LengthSquared() <= s.Radius*s.Radius {
				tHit = t
				hit = true
			}
		}
	}
	return hit, matrix.Float(tHit)
}

func (s Cone) IntersectsPlane(p Plane) (bool, matrix.Float) {
	halfH := s.Height / 2
	apex := s.Center.Subtract(s.Direction.Scale(halfH))
	baseCenter := s.Center.Add(s.Direction.Scale(halfH))
	distApex := matrix.Vec3Dot(apex, p.Normal) + p.Dot
	distBase := matrix.Vec3Dot(baseCenter, p.Normal) + p.Dot
	// Plane crosses the cone axis between apex and base
	if (distApex >= 0 && distBase <= 0) || (distApex <= 0 && distBase >= 0) {
		maxDist := matrix.Sqrt(s.Radius*s.Radius + halfH*halfH)
		centerDist := matrix.Vec3Dot(s.Center, p.Normal) + p.Dot
		if centerDist < 0 {
			centerDist = -centerDist
		}
		return true, matrix.Float(maxDist - centerDist)
	}
	// Check base cap
	capDist := distBase
	if capDist < 0 {
		capDist = -capDist
	}
	if capDist <= s.Radius {
		return true, matrix.Float(s.Radius - capDist)
	}
	// Check lateral surface
	ne := matrix.Vec3Dot(p.Normal, s.Direction)
	if ne*ne < 1e-10 {
		return false, 0
	}
	// Find closest point on the cone axis to the plane
	num := matrix.Vec3Dot(s.Center, p.Normal) + p.Dot
	t := num / (ne * ne)
	closest := s.Center.Subtract(s.Direction.Scale(t))
	proj := matrix.Vec3Dot(closest.Subtract(s.Center), s.Direction)
	// Clamp to cone height bounds
	if proj > halfH {
		proj = halfH
	}
	if proj < -halfH {
		proj = -halfH
	}
	clamped := s.Center.Add(s.Direction.Scale(proj))
	// Cone radius at this height (0 at apex, full at base)
	ratio := (proj + halfH) / s.Height
	radiusAtH := s.Radius * ratio
	closestDist := matrix.Vec3Dot(clamped, p.Normal) + p.Dot
	if closestDist < 0 {
		closestDist = -closestDist
	}
	if closestDist <= radiusAtH {
		return true, matrix.Float(radiusAtH - closestDist)
	}
	return false, 0
}

func (s Cone) IntersectsFrustum(f Frustum) bool {
	halfH := s.Height / 2
	apex := s.Center.Subtract(s.Direction.Scale(halfH))
	baseCenter := s.Center.Add(s.Direction.Scale(halfH))
	for i := range f.Planes {
		p := f.Planes[i]
		apexDist := matrix.Vec3Dot(apex, p.Normal) + p.Dot
		baseDist := matrix.Vec3Dot(baseCenter, p.Normal) + p.Dot
		perpNormal := p.Normal.Subtract(s.Direction.Scale(matrix.Vec3Dot(p.Normal, s.Direction)))
		maxBaseDist := baseDist + perpNormal.Length()*s.Radius
		if apexDist < 0 && maxBaseDist < 0 {
			return false
		}
	}
	return true
}

func (s Cone) IntersectsSphere(c Sphere) bool {
	return c.IntersectsCone(s)
}

func (s Cone) IntersectsCapsule(c Capsule) bool {
	return c.IntersectsCone(s)
}

func (s Cone) IntersectsCylinder(c Cylinder) bool {
	return c.IntersectsCone(s)
}

func coneIntersectsSegment(c Cone, p0, p1 matrix.Vec3) bool {
	e := p1.Subtract(p0)
	if matrix.Vec3Dot(e, e) == 0 {
		return pointInCone(p0, c)
	}
	o0 := p0.Subtract(c.Center)
	o0sq := matrix.Vec3Dot(o0, o0)
	dotOe := matrix.Vec3Dot(o0, e)
	ee := matrix.Vec3Dot(e, e)
	d0 := matrix.Vec3Dot(c.Direction, o0)
	de := matrix.Vec3Dot(c.Direction, e)
	P0 := o0sq - d0*d0
	P1 := dotOe - d0*de
	P2 := ee - de*de
	halfH := c.Height / 2
	rOH := c.Radius / c.Height
	rSqC := rOH * rOH * de * de
	rSqB := 2 * rOH * rOH * de * (d0 + halfH)
	rSqA := rOH * rOH * (d0 + halfH) * (d0 + halfH)
	a := P2 - rSqC
	bq := 2 * (P1 - rSqB/2)
	cq := P0 - rSqA
	if matrix.Abs(a) < 1e-12 {
		if matrix.Abs(bq) < 1e-12 {
			return false
		}
		t := -cq / bq
		t = max(min(t, matrix.Float(1)), matrix.Float(0))
		return pointInCone(p0.Add(e.Scale(t)), c)
	}
	disc := bq*bq - 4*a*cq
	if disc < 0 {
		return false
	}
	sqrtDisc := matrix.Sqrt(disc)
	t0 := (-bq - sqrtDisc) / (2 * a)
	t1 := (-bq + sqrtDisc) / (2 * a)
	if t0 > t1 {
		t0, t1 = t1, t0
	}
	if t1 < 0 || t0 > 1 {
		return false
	}
	tStart := max(t0, matrix.Float(0))
	tEnd := min(t1, matrix.Float(1))
	if tStart <= tEnd {
		return pointInCone(p0.Add(e.Scale(tStart)), c)
	}
	return false
}

func pointInCone(p matrix.Vec3, c Cone) bool {
	// Project point onto cone axis
	dir := p.Subtract(c.Center)
	t := dir.Dot(c.Direction)
	// Check if point is within cone height bounds
	if t < -c.Height/2 || t > c.Height/2 {
		return false
	}
	// Calculate radius at this height (linear interpolation from apex to base)
	ratio := (t + c.Height/2) / c.Height
	radiusAtHeight := c.Radius * ratio
	// Check if perpendicular distance is within radius
	perpDist := dir.Subtract(c.Direction.Scale(t)).Length()
	return perpDist <= radiusAtHeight
}

func axesIntersect(a Cone, b Cone) bool {
	// Check if the two cone axes intersect within both cone volumes
	// Using line-line intersection in 3D
	d1 := a.Direction
	d2 := b.Direction
	r := a.Center.Subtract(b.Center)
	d1d2 := d1.Dot(d2)
	d1r := d1.Dot(r)
	d2r := d2.Dot(r)
	denom := 1 - d1d2*d1d2
	if denom == 0 {
		return false // Parallel axes
	}
	t := (d1d2*d2r - d1r) / denom
	u := (d2r - d1d2*d1r) / denom
	// Check if intersection point is within both cone heights
	return t >= -a.Height/2 && t <= a.Height/2 &&
		u >= -b.Height/2 && u <= b.Height/2
}

/******************************************************************************/
/* plane.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

type Plane struct {
	Normal matrix.Vec3
	Dot    matrix.Float
}

// PlaneCCW creates a plane from three points in counter clockwise order
func PlaneCCW(a, b, c matrix.Vec3) Plane {
	var p Plane
	e0 := b.Subtract(a)
	e1 := c.Subtract(a)
	p.Normal = matrix.Vec3Cross(e0, e1).Normal()
	p.Dot = matrix.Vec3Dot(p.Normal, a)
	return p
}

// SetFloatValue sets the value of the plane at the given index (X, Y, Z, Dot)
func (p *Plane) SetFloatValue(value matrix.Float, index int) {
	switch index {
	case 0:
		p.Normal.SetX(value)
	case 1:
		p.Normal.SetY(value)
	case 2:
		p.Normal.SetZ(value)
	case 3:
		p.Dot = value
	}
}

// ToArray converts the plane to an array of 4 floats
func (p Plane) ToArray() [4]matrix.Float {
	return [4]matrix.Float{p.Normal.X(), p.Normal.Y(), p.Normal.Z(), p.Dot}
}

// ToVec4 converts the plane to a Vec4 (analogous to ToArray)
func (p Plane) ToVec4() matrix.Vec4 {
	return matrix.Vec4{p.Normal.X(), p.Normal.Y(), p.Normal.Z(), p.Dot}
}

// ClosestPoint returns the closest point on the plane to the given point
func (p Plane) ClosestPoint(point matrix.Vec3) matrix.Vec3 {
	// If normalized, t := matrix.Vec3Dot(point, p.Normal) - p.Dot
	t := (matrix.Vec3Dot(p.Normal, point) - p.Dot) / matrix.Vec3Dot(p.Normal, p.Normal)
	return point.Subtract(p.Normal.Scale(t))
}

// Distance returns the distance from the plane to the given point
func (p Plane) Distance(point matrix.Vec3) matrix.Float {
	// If normalized, return matrix.Vec3Dot(p.Normal, point) - p.Dot
	return (matrix.Vec3Dot(p.Normal, point) - p.Dot) / matrix.Vec3Dot(p.Normal, p.Normal)
}

// PointOutsideOfPlane returns true if the given point is outside of the plane
func PointOutsideOfPlane(p, a, b, c, d matrix.Vec3) bool {
	signP := matrix.Vec3Dot(p.Subtract(a), matrix.Vec3Cross(b.Subtract(a), c.Subtract(a)))
	signD := matrix.Vec3Dot(d.Subtract(a), matrix.Vec3Cross(b.Subtract(a), c.Subtract(a)))
	return signP*signD < 0
}

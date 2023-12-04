package collision

import "kaiju/matrix"

type Plane struct {
	Normal matrix.Vec3
	Dot    float32
}

func PlaneCCW(a, b, c matrix.Vec3) Plane {
	var p Plane
	e0 := b.Subtract(a)
	e1 := c.Subtract(a)
	p.Normal = matrix.Vec3Cross(e0, e1).Normal()
	p.Dot = matrix.Vec3Dot(p.Normal, a)
	return p
}

func (p *Plane) SetFloatValue(value float32, index int) {
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

func (p Plane) ToArray() [4]float32 {
	return [4]float32{p.Normal.X(), p.Normal.Y(), p.Normal.Z(), p.Dot}
}

func (p Plane) ToVec4() matrix.Vec4 {
	return matrix.Vec4{p.Normal.X(), p.Normal.Y(), p.Normal.Z(), p.Dot}
}

func (p Plane) ClosestPoint(point matrix.Vec3) matrix.Vec3 {
	// If normalized, t := matrix.Vec3Dot(point, p.Normal) - p.Dot
	t := (matrix.Vec3Dot(p.Normal, point) - p.Dot) / matrix.Vec3Dot(p.Normal, p.Normal)
	return point.Subtract(p.Normal.Scale(t))
}

func (p Plane) Distance(point matrix.Vec3) float32 {
	// If normalized, return matrix.Vec3Dot(p.Normal, point) - p.Dot
	return (matrix.Vec3Dot(p.Normal, point) - p.Dot) / matrix.Vec3Dot(p.Normal, p.Normal)
}

func PointOutsideOfPlane(p, a, b, c, d matrix.Vec3) bool {
	signp := matrix.Vec3Dot(p.Subtract(a), matrix.Vec3Cross(b.Subtract(a), c.Subtract(a)))
	signd := matrix.Vec3Dot(d.Subtract(a), matrix.Vec3Cross(b.Subtract(a), c.Subtract(a)))
	return signp*signd < 0
}

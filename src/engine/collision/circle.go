package collision

import "kaiju/matrix"

type Axis = uint8

const (
	AxisX = Axis(iota)
	AxisY
	AxisZ
)

type Circle struct {
	Point  matrix.Vec3
	Radius matrix.Float
	Axis   Axis
}

func (c Circle) RayHit(ray Ray) (matrix.Vec3, bool) {
	var normal matrix.Vec3
	switch c.Axis {
	case AxisX:
		normal = matrix.Vec3{1, 0, 0}
	case AxisY:
		normal = matrix.Vec3{0, 1, 0}
	case AxisZ:
		normal = matrix.Vec3{0, 0, 1}
	}
	denom := matrix.Vec3Dot(normal, ray.Direction)
	if denom == 0 {
		return matrix.Vec3{}, false
	}
	t := matrix.Vec3Dot(normal, c.Point.Subtract(ray.Origin)) / denom
	if t < 0 {
		return matrix.Vec3{}, false
	}
	hit := ray.Origin.Add(ray.Direction.Scale(t))
	diff := hit.Subtract(c.Point)
	switch c.Axis {
	case AxisX:
		diff.SetX(0)
	case AxisY:
		diff.SetY(0)
	case AxisZ:
		diff.SetZ(0)
	}
	if matrix.Vec3Dot(diff, diff) <= c.Radius*c.Radius {
		return hit, true
	}
	return matrix.Vec3{}, false
}

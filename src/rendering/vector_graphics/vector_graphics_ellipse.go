package vector_graphics

import "kaijuengine.com/matrix"

type Ellipse struct {
	ShapeBase
	Center matrix.Vec3
	Radius matrix.Vec2
}

func (e *Ellipse) Animate(animType AnimatedValueType, value float64) {
	switch animType {
	case AnimatedValueTypePositionX:
		e.Center.SetX(matrix.Float(value))
	case AnimatedValueTypePositionY:
		e.Center.SetY(matrix.Float(value))
	case AnimatedValueTypePositionZ:
		e.Center.SetZ(matrix.Float(value))
	case AnimatedValueTypeRadiusX:
		e.Radius.SetX(matrix.Float(value))
	case AnimatedValueTypeRadiusY:
		e.Radius.SetY(matrix.Float(value))
	default:
		e.ShapeBase.Animate(animType, value)
	}
}

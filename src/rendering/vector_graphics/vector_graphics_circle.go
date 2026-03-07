package vector_graphics

import "kaijuengine.com/matrix"

type Circle struct {
	ShapeBase
	Center matrix.Vec3
	Radius matrix.Float
}

func (e *Circle) Animate(animType AnimatedValueType, value float64) {
	switch animType {
	case AnimatedValueTypePositionX:
		e.Center.SetX(matrix.Float(value))
	case AnimatedValueTypePositionY:
		e.Center.SetY(matrix.Float(value))
	case AnimatedValueTypePositionZ:
		e.Center.SetZ(matrix.Float(value))
	case AnimatedValueTypeRadius:
		e.Radius = matrix.Float(value)
	default:
		e.ShapeBase.Animate(animType, value)
	}
}

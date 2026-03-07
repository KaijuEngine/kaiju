package vector_graphics

import "kaijuengine.com/matrix"

type LineSegment struct {
	ShapeBase
	From matrix.Vec2
	To   matrix.Vec2
}

func (e *LineSegment) Animate(animType AnimatedValueType, value float64) {
	v := matrix.Float(value)
	switch animType {
	case AnimatedValueTypeFromX:
		e.From.SetX(v)
	case AnimatedValueTypeFromY:
		e.From.SetY(v)
	case AnimatedValueTypeToX:
		e.To.SetX(v)
	case AnimatedValueTypeToY:
		e.To.SetY(v)
	}
}

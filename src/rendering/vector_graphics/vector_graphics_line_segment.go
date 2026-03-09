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

func (e *LineSegment) ToPolygon() []matrix.Vec2 {
	return []matrix.Vec2{e.From, e.To}
}

package vector_graphics

import "kaijuengine.com/matrix"

type Rectangle struct {
	ShapeBase
	Center matrix.Vec2
	Size   matrix.Vec2
}

func (e *Rectangle) Animate(animType AnimatedValueType, value float64) {
	switch animType {
	case AnimatedValueTypePositionX:
		e.Center.SetX(matrix.Float(value))
	case AnimatedValueTypePositionY:
		e.Center.SetY(matrix.Float(value))
	case AnimatedValueTypeWidth:
		e.Size.SetWidth(matrix.Float(value))
	case AnimatedValueTypeRadiusY:
		e.Size.SetHeight(matrix.Float(value))
	default:
		e.ShapeBase.Animate(animType, value)
	}
}

func (e *Rectangle) ToPolygon() []matrix.Vec2 {
	return []matrix.Vec2{
		{e.Center.X() - e.Size.X(), e.Center.Y() - e.Size.Y()},
		{e.Center.X() - e.Size.X(), e.Center.Y() + e.Size.Y()},
		{e.Center.X() + e.Size.X(), e.Center.Y() + e.Size.Y()},
		{e.Center.X() + e.Size.X(), e.Center.Y() - e.Size.Y()},
	}
}

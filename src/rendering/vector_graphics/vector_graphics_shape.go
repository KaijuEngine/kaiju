package vector_graphics

import "kaijuengine.com/matrix"

type LineCap int8
type LineJoin int8

const (
	LineCapButt LineCap = iota
	LineCapRound
	LineCapSquare
	LineCapInherit
)

const (
	LineJoinMiter LineJoin = iota
	LineJoinRound
	LineJoinBevel
	LineJoinInherit
)

type Shape interface {
	Animate(animType AnimatedValueType, value float64)
}

type ShapeBase struct {
	Stroke      Color
	Fill        Color
	LineCap     LineCap
	LineJoin    LineJoin
	StrokeWidth matrix.Float
	Rotation    matrix.Float
}

func (e *ShapeBase) Animate(animType AnimatedValueType, value float64) {
	switch animType {
	case AnimatedValueTypeStrokeWidth:
		e.StrokeWidth = matrix.Float(value)
	case AnimatedValueTypeStrokeR:
		e.Stroke.Animate(AnimatedValueTypeColorR, value)
	case AnimatedValueTypeStrokeG:
		e.Stroke.Animate(AnimatedValueTypeColorG, value)
	case AnimatedValueTypeStrokeB:
		e.Stroke.Animate(AnimatedValueTypeColorB, value)
	case AnimatedValueTypeStrokeA:
		e.Stroke.Animate(AnimatedValueTypeColorA, value)
	case AnimatedValueTypeFillR:
		e.Fill.Animate(AnimatedValueTypeFillR, value)
	case AnimatedValueTypeFillG:
		e.Fill.Animate(AnimatedValueTypeFillG, value)
	case AnimatedValueTypeFillB:
		e.Fill.Animate(AnimatedValueTypeFillB, value)
	case AnimatedValueTypeFillA:
		e.Fill.Animate(AnimatedValueTypeFillA, value)
	case AnimatedValueTypeRotation:
		e.Rotation = matrix.Float(value)
	}
}

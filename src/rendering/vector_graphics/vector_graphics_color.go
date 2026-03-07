package vector_graphics

import "kaijuengine.com/matrix"

type Color matrix.Color

func (e *Color) Animate(animType AnimatedValueType, value float64) {
	v := matrix.Float(value)
	switch animType {
	case AnimatedValueTypeColorR:
		(*matrix.Color)(e).SetR(v)
	case AnimatedValueTypeColorG:
		(*matrix.Color)(e).SetG(v)
	case AnimatedValueTypeColorB:
		(*matrix.Color)(e).SetB(v)
	case AnimatedValueTypeColorA:
		(*matrix.Color)(e).SetA(v)
	}
}

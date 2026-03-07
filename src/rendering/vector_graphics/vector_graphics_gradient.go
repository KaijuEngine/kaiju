package vector_graphics

type GradientType int

const (
	GradientTypeLinear GradientType = iota
	GradientTypeRadial
)

type Gradient struct {
	Type   GradientType
	Range  LineSegment
	Colors []GradientColor
}

type GradientColor struct {
	Color
	Position float64 // Relative to the gradient's range (0-1)
}

func (e *GradientColor) Animate(animType AnimatedValueType, value float64) {
	switch animType {
	case AnimatedValueTypePosition:
		e.Position = value
	default:
		e.Color.Animate(animType, value)
	}
}

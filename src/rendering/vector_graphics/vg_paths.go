package vector_graphics

import "kaiju/matrix"

// GradientType defines the type of gradient
type GradientType uint8

// GradientSpread defines how to handle colors outside the gradient range
type GradientSpread uint8

const (
	GradientTypeNone GradientType = iota
	GradientTypeLinear
	GradientTypeRadial
	GradientTypeElliptical
	GradientTypeConical
)

const (
	GradientSpreadPad GradientSpread = iota
	GradientSpreadReflect
	GradientSpreadRepeat
)

// BezierCurve represents a cubic bezier curve with start, control, and end points
type BezierCurve struct {
	Start    matrix.Vec2  // Starting point of the curve
	Control1 matrix.Vec2  // First control point (influence on start tangent)
	Control2 matrix.Vec2  // Second control point (influence on end tangent)
	End      matrix.Vec2  // Ending point of the curve
	Length   matrix.Float // Cached curve length for optimization
}

// QuadraticBezier represents a quadratic bezier curve with one control point
type QuadraticBezier struct {
	Start   matrix.Vec2
	Control matrix.Vec2
	End     matrix.Vec2
}

// Line represents a straight line segment
type Line struct {
	Start matrix.Vec2
	End   matrix.Vec2
}

// Path represents a sequence of connected curves and lines
type Path struct {
	Points     []matrix.Vec2 // All points in the path
	Curves     []BezierCurve
	Quadratics []QuadraticBezier
	Lines      []Line
	IsClosed   bool // Whether the path forms a closed shape
}

// Shape represents a closed vector shape that can be filled and stroked
type Shape struct {
	Path           Path
	FillColor      matrix.Color
	StrokeColor    matrix.Color
	StrokeWidth    matrix.Float
	FillGradient   Gradient
	StrokeGradient Gradient
}

// GradientColorStop represents a color at a specific position in the gradient (0.0 to 1.0)
type GradientColorStop struct {
	Position matrix.Float
	Color    matrix.Color
}

// Gradient represents a color gradient fill
type Gradient struct {
	Type       GradientType
	ColorStops []GradientColorStop
	Spread     GradientSpread // How to handle colors outside the gradient range
}

// LinearGradientParams defines parameters for linear gradients
type LinearGradientParams struct {
	StartPoint matrix.Vec2
	EndPoint   matrix.Vec2
}

// RadialGradientParams defines parameters for radial/elliptical gradients
type RadialGradientParams struct {
	Center matrix.Vec2
	Radius matrix.Float
	Focus  matrix.Vec2 // For elliptical gradients
}

// ConicalGradientParams defines parameters for conical gradients
type ConicalGradientParams struct {
	Center     matrix.Vec2
	StartAngle matrix.Float // in radians
	EndAngle   matrix.Float // in radians
}

// NewLinearGradient creates a new linear gradient
func NewLinearGradient(start, end matrix.Vec2, colorStops []GradientColorStop) Gradient {
	return Gradient{
		Type:       GradientTypeLinear,
		ColorStops: colorStops,
		Spread:     GradientSpreadPad,
	}
}

// NewRadialGradient creates a new radial gradient
func NewRadialGradient(center matrix.Vec2, radius matrix.Float, colorStops []GradientColorStop) Gradient {
	return Gradient{
		Type:       GradientTypeRadial,
		ColorStops: colorStops,
		Spread:     GradientSpreadPad,
	}
}

// NewEllipticalGradient creates a new elliptical gradient
func NewEllipticalGradient(center, focus matrix.Vec2, radius matrix.Float, colorStops []GradientColorStop) Gradient {
	return Gradient{
		Type:       GradientTypeElliptical,
		ColorStops: colorStops,
		Spread:     GradientSpreadPad,
	}
}

// NewConicalGradient creates a new conical gradient
func NewConicalGradient(center matrix.Vec2, startAngle, endAngle matrix.Float, colorStops []GradientColorStop) Gradient {
	return Gradient{
		Type:       GradientTypeConical,
		ColorStops: colorStops,
		Spread:     GradientSpreadPad,
	}
}

func (g *Gradient) IsValid() bool { return g.Type != GradientTypeNone }

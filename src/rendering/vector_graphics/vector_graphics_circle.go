package vector_graphics

import (
	"math"

	"kaijuengine.com/matrix"
)

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

// ToPolygon generates a polygon approximating the circle.
// If includeStroke is true, the polygon radius is expanded by half the stroke width,
// effectively including the stroke in the mesh geometry.
func (e *Circle) ToPolygon(density int) []matrix.Vec2 {
	// Ensure a minimum of three points to form a valid polygon.
	if density < 3 {
		density = 3
	}
	points := make([]matrix.Vec2, density)
	// Determine effective radius, optionally expanding for stroke width.
	effectiveRadius := e.Radius
	for i := 0; i < density; i++ {
		// Angle around the circle (0 to 2π).
		angle := matrix.Float(2 * math.Pi * float64(i) / float64(density))
		// Local coordinates on the circle's perimeter using the effective radius.
		x := effectiveRadius * matrix.Cos(angle)
		y := effectiveRadius * matrix.Sin(angle)
		// Offset by the circle's center.
		px := x + e.Center.X()
		py := y + e.Center.Y()
		points[i] = matrix.NewVec2(px, py)
	}
	return points
}

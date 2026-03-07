package vector_graphics

import "kaijuengine.com/matrix"

type Polygon struct {
	ShapeBase
	Points []PolygonPoint
}

type PolygonPoint struct {
	Handle PointCurve
	Point  matrix.Vec2
}

func (e *Polygon) Animate(animType AnimatedValueType, value float64) {
	if animType < AnimatedValueTypePolygonPoint {
		e.ShapeBase.Animate(animType, value)
		return
	}
	// The point is encoded into AnimatedValueTypePolygonPoint, see comment
	// under the declaration of AnimatedValueTypePolygonPoint
	position := animType - AnimatedValueTypePolygonPoint
	index := position / 2
	component := position % 2
	e.Points[index].Point[component] = matrix.Float(value)
}

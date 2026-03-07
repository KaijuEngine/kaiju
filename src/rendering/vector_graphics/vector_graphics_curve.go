package vector_graphics

import "kaijuengine.com/matrix"

type PointCurve struct {
	// The character of a curve is a vector that represents the handle you'd see
	// in animation software to control how the curve looks. The character is
	// not a normalized vector as the length of the vector contributes to the
	// characteristic fo the curve. The character points represent the distance
	// from the origin of the curve's point.

	LeftCharacter  matrix.Vec2
	RightCharacter matrix.Vec2
}

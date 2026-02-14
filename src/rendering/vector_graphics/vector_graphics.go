package vector_graphics

import (
	"kaiju/matrix"
)

// Transform represents a 2D transformation matrix
type Transform struct {
	ScaleX     matrix.Float
	ScaleY     matrix.Float
	Rotation   matrix.Float // in radians
	TranslateX matrix.Float
	TranslateY matrix.Float
}

// GraphicElement represents a renderable element in a vector graphic
type GraphicElement struct {
	ID        string
	Shape     Shape
	Transform Transform
	Animation Animation
	Visible   bool
}

// VectorGraphic represents a complete vector graphic composed of multiple elements
type VectorGraphic struct {
	Width      matrix.Float
	Height     matrix.Float
	Elements   []GraphicElement
	Animations []Animation
}

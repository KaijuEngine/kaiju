package vector_graphics

import "kaiju/matrix"

// StrokeLinecap defines the possible values for the SVG stroke-linecap property.
// It controls the shape of the ends of a line.
type StrokeLinecap int8

// StrokeLinejoin defines the possible values for the SVG stroke-linejoin property.
// It controls how the junction between two line segments is rendered.
type StrokeLinejoin int8

const (
	StrokeLinecapButt StrokeLinecap = iota
	StrokeLinecapRound
	StrokeLinecapSquare
	StrokeLinecapInherit
)

const (
	StrokeLinejoinMiter StrokeLinejoin = iota
	StrokeLinejoinRound
	StrokeLinejoinBevel
	StrokeLinejoinInherit
)

type VectorGraphic struct {
	ViewBox [4]float64
	Groups  []Group
}

type Transform struct {
	Position matrix.Vec2
	Rotation matrix.Float
	Scale    matrix.Vec2
}

type Group struct {
	Transform         Transform
	Opacity           matrix.Float
	Groups            []Group
	Paths             []Path
	Ellipses          []Ellipse
	AnimateTransforms []AnimateTransform
}

type Path struct {
	Id             string
	Data           []PathSegment
	Stroke         matrix.Color
	StrokeWidth    matrix.Float
	Fill           matrix.Color
	StrokeLinecap  StrokeLinecap
	StrokeLinejoin StrokeLinejoin
	Animates       []Animate
}

type Ellipse struct {
	Id             string
	Center         matrix.Vec2
	Radius         matrix.Vec2
	Stroke         matrix.Color
	StrokeWidth    matrix.Float
	Fill           matrix.Color
	StrokeLinecap  StrokeLinecap
	StrokeLinejoin StrokeLinejoin
	Animates       []Animate
}

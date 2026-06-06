/******************************************************************************/
/* shape.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

type ShapeType uint8

const (
	ShapeTypeSphere ShapeType = iota
	ShapeTypeAABB
	ShapeTypeOOBB
	ShapeTypeCapsule
	ShapeTypeCylinder
	ShapeTypeCone
	ShapeTypeMesh
	ShapeTypeTerrain
)

type Shape struct {
	Center      matrix.Vec3  // Circle, AABB, OOBB, Capsule, Cylinder, Cone
	Radius      matrix.Float // Circle, Capsule, Cylinder, Cone
	Extent      matrix.Vec3  // AABB, OOBB
	Orientation matrix.Mat3  // OOBB
	Height      matrix.Float // Capsule, Cylinder, Cone
	Direction   matrix.Vec3  // Capsule, Cylinder, Cone
	Type        ShapeType
}

func (s *Shape) SetBoxShape(extent matrix.Vec3) {
	// Engine box shapes are OOBBs so rotated entity bodies do not collapse to
	// world-axis bounds.
	s.SetOOBB(matrix.Vec3Zero(), extent, matrix.Mat3Identity())
}

func NewBoxShape(extent matrix.Vec3) Shape {
	s := Shape{}
	s.SetBoxShape(extent)
	return s
}

func (s *Shape) SetAABBShape(extent matrix.Vec3) {
	s.SetAABB(matrix.Vec3Zero(), extent)
}

func NewAABBShape(extent matrix.Vec3) Shape {
	s := Shape{}
	s.SetAABBShape(extent)
	return s
}

func (s *Shape) SetOOBBShape(extent matrix.Vec3) {
	s.SetOOBB(matrix.Vec3Zero(), extent, matrix.Mat3Identity())
}

func NewOOBBShape(extent matrix.Vec3) Shape {
	s := Shape{}
	s.SetOOBBShape(extent)
	return s
}

func (s *Shape) SetSphereShape(radius matrix.Float) {
	s.SetSphere(matrix.Vec3Zero(), radius)
}

func NewSphereShape(radius matrix.Float) Shape {
	s := Shape{}
	s.SetSphereShape(radius)
	return s
}

func (s *Shape) SetCapsuleShape(radius, height matrix.Float) {
	s.SetCapsule(matrix.Vec3Zero(), radius, height, matrix.Vec3Up())
}

func NewCapsuleShape(radius, height matrix.Float) Shape {
	s := Shape{}
	s.SetCapsuleShape(radius, height)
	return s
}

func (s *Shape) SetCylinderShape(radius, height matrix.Float) {
	s.SetCylinder(matrix.Vec3Zero(), radius, height, matrix.Vec3Up())
}

func NewCylinderShape(radius, height matrix.Float) Shape {
	s := Shape{}
	s.SetCylinderShape(radius, height)
	return s
}

func (s *Shape) SetConeShape(radius, height matrix.Float) {
	s.SetCone(matrix.Vec3Zero(), radius, height, matrix.Vec3Up())
}

func NewConeShape(radius, height matrix.Float) Shape {
	s := Shape{}
	s.SetConeShape(radius, height)
	return s
}

func (s *Shape) SetTerrain(bounds AABB) {
	s.Type = ShapeTypeTerrain
	s.Center = bounds.Center
	s.Extent = bounds.Extent
}

func NewTerrainShape(bounds AABB) Shape {
	s := Shape{}
	s.SetTerrain(bounds)
	return s
}

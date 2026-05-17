/******************************************************************************/
/* shape.go                                                                   */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
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

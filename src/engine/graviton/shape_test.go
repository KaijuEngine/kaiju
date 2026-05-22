/******************************************************************************/
/* shape_test.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestNewBoxShapeCreatesOOBB(t *testing.T) {
	extent := matrix.NewVec3(1, 2, 3)
	shape := NewBoxShape(extent)
	if shape.Type != ShapeTypeOOBB {
		t.Fatalf("expected box shape to use OOBB, got %v", shape.Type)
	}
	if !matrix.Vec3ApproxTo(shape.Center, matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected centered box shape, got %v", shape.Center)
	}
	if !matrix.Vec3ApproxTo(shape.Extent, extent, 0.0001) {
		t.Fatalf("expected extent %v, got %v", extent, shape.Extent)
	}
	if !matrix.Mat3ApproxTo(shape.Orientation, matrix.Mat3Identity(), 0.0001) {
		t.Fatalf("expected identity orientation, got %v", shape.Orientation)
	}
}

func TestNewSphereShapeSetup(t *testing.T) {
	shape := NewSphereShape(2.5)
	if shape.Type != ShapeTypeSphere {
		t.Fatalf("expected sphere shape, got %v", shape.Type)
	}
	if !matrix.Vec3ApproxTo(shape.Center, matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected centered sphere shape, got %v", shape.Center)
	}
	if matrix.Abs(shape.Radius-2.5) > 0.0001 {
		t.Fatalf("expected radius 2.5, got %f", shape.Radius)
	}
}

func TestNewCapsuleShapeSetup(t *testing.T) {
	shape := NewCapsuleShape(1.5, 4)
	if shape.Type != ShapeTypeCapsule {
		t.Fatalf("expected capsule shape, got %v", shape.Type)
	}
	if matrix.Abs(shape.Radius-1.5) > 0.0001 {
		t.Fatalf("expected radius 1.5, got %f", shape.Radius)
	}
	if matrix.Abs(shape.Height-4) > 0.0001 {
		t.Fatalf("expected height 4, got %f", shape.Height)
	}
	if !matrix.Vec3ApproxTo(shape.Direction, matrix.Vec3Up(), 0.0001) {
		t.Fatalf("expected up direction, got %v", shape.Direction)
	}
}

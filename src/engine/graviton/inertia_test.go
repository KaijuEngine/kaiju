/******************************************************************************/
/* inertia_test.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestCalculateLocalInertiaReturnsZeroForStaticMass(t *testing.T) {
	shape := NewSphereShape(2)
	for _, mass := range []matrix.Float{0, -1} {
		inertia := CalculateLocalInertia(shape, mass)
		if !inertia.IsZero() {
			t.Fatalf("expected zero inertia for mass %f, got %v", mass, inertia)
		}
	}
	body := RigidBody{}
	body.SetShape(shape)
	body.SetStatic()
	inertia := CalculateLocalInertia(body.Shape(), body.Mass.Mass)
	if !inertia.IsZero() {
		t.Fatalf("expected static body inertia to be zero, got %v", inertia)
	}
}

func TestCalculateLocalInertiaDynamicShapesAreNonZero(t *testing.T) {
	shapes := map[string]Shape{
		"sphere":   NewSphereShape(1),
		"box":      NewBoxShape(matrix.NewVec3(1, 2, 3)),
		"aabb":     NewAABBShape(matrix.NewVec3(1, 2, 3)),
		"oobb":     NewOOBBShape(matrix.NewVec3(1, 2, 3)),
		"capsule":  NewCapsuleShape(1, 2),
		"cylinder": NewCylinderShape(1, 2),
		"cone":     NewConeShape(1, 2),
		"mesh":     {Type: ShapeTypeMesh, Extent: matrix.NewVec3(1, 2, 3)},
	}
	for name, shape := range shapes {
		inertia := CalculateLocalInertia(shape, 2)
		if inertia.X() <= 0 || inertia.Y() <= 0 || inertia.Z() <= 0 {
			t.Fatalf("expected nonzero inertia for %s, got %v", name, inertia)
		}
	}
}

func TestCalculateLocalInertiaSphereFormula(t *testing.T) {
	inertia := CalculateLocalInertia(NewSphereShape(2), 3)
	expected := matrix.NewVec3(4.8, 4.8, 4.8)
	if !matrix.Vec3ApproxTo(inertia, expected, 0.0001) {
		t.Fatalf("expected sphere inertia %v, got %v", expected, inertia)
	}
}

func TestCalculateLocalInertiaBoxFormula(t *testing.T) {
	inertia := CalculateLocalInertia(NewBoxShape(matrix.NewVec3(1, 2, 3)), 12)
	expected := matrix.NewVec3(52, 40, 20)
	if !matrix.Vec3ApproxTo(inertia, expected, 0.0001) {
		t.Fatalf("expected box inertia %v, got %v", expected, inertia)
	}
}

func TestCalculateLocalInertiaCylinderFormula(t *testing.T) {
	inertia := CalculateLocalInertia(NewCylinderShape(2, 4), 6)
	expected := matrix.NewVec3(14, 12, 14)
	if !matrix.Vec3ApproxTo(inertia, expected, 0.0001) {
		t.Fatalf("expected cylinder inertia %v, got %v", expected, inertia)
	}
}

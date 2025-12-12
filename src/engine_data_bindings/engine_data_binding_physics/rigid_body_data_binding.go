package engine_data_binding_physics

import (
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/physics"
)

const BindingKey = "kaiju.RigidBodyDataBinding"

type Shape int

const (
	ShapeBox Shape = iota
	ShapeSphere
	ShapeCapsule
	ShapeCylinder
	ShapeCone
)

func init() {
	engine.RegisterEntityData(BindingKey, RigidBodyDataBinding{})
}

type RigidBodyDataBinding struct {
	Extent matrix.Vec3
	Mass   float32
	Radius float32
	Height float32
	Shape  Shape
}

func (r RigidBodyDataBinding) Init(e *engine.Entity, host *engine.Host) {
	t := &e.Transform
	// TODO:  Scale the shape by the transform scale
	var shape *physics.CollisionShape
	switch r.Shape {
	case ShapeBox:
		shape = &physics.NewBoxShape(r.Extent).CollisionShape
	case ShapeSphere:
		shape = &physics.NewSphereShape(r.Radius).CollisionShape
	case ShapeCapsule:
		shape = &physics.NewCapsuleShape(r.Radius, r.Height).CollisionShape
	case ShapeCylinder:
		shape = &physics.NewnCylinderShape(r.Extent).CollisionShape
	case ShapeCone:
		shape = &physics.NewConeShape(r.Radius, r.Height).CollisionShape
	}
	inertia := shape.CalculateLocalInertia(r.Mass)
	motion := physics.NewDefaultMotionState(matrix.QuaternionFromEuler(t.Rotation()), t.Position())
	body := physics.NewRigidBody(r.Mass, motion, shape, inertia)
	host.Physics().AddEntity(e, body)
}

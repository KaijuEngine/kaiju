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
	Extent   matrix.Vec3 `default:"1,1,1"`
	Mass     float32     `default:"1"`
	Radius   float32     `default:"1"`
	Height   float32     `default:"1"`
	Shape    Shape
	IsStatic bool
}

func (r RigidBodyDataBinding) Init(e *engine.Entity, host *engine.Host) {
	t := &e.Transform
	scale := t.Scale()
	var shape *physics.CollisionShape
	switch r.Shape {
	case ShapeBox:
		size := r.Extent.Multiply(scale)
		shape = &physics.NewBoxShape(size).CollisionShape
	case ShapeSphere:
		rad := r.Radius * float32(scale.LongestAxis())
		shape = &physics.NewSphereShape(rad).CollisionShape
	case ShapeCapsule:
		rad := r.Radius * float32(scale.LongestAxis())
		height := r.Height * scale.Y()
		shape = &physics.NewCapsuleShape(rad, height).CollisionShape
	case ShapeCylinder:
		size := r.Extent.Multiply(scale)
		shape = &physics.NewCylinderShape(size).CollisionShape
	case ShapeCone:
		rad := r.Radius * float32(scale.LongestAxis())
		height := r.Height * scale.Y()
		shape = &physics.NewConeShape(rad, height).CollisionShape
	}
	if r.IsStatic {
		r.Mass = 0
	}
	inertia := shape.CalculateLocalInertia(r.Mass)
	motion := physics.NewDefaultMotionState(matrix.QuaternionFromEuler(t.Rotation()), t.Position())
	body := physics.NewRigidBody(r.Mass, motion, shape, inertia)
	host.Physics().AddEntity(e, body)
}

/******************************************************************************/
/* rigid_body.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/matrix"
)

type RigidBodyType uint8

const (
	RigidBodyTypeStatic RigidBodyType = iota
	RigidBodyTypeKinematic
	RigidBodyTypeDynamic
)

const (
	DefaultCollisionGroup = 0
	DefaultCollisionMask  = 1 << DefaultCollisionGroup
)

const (
	DefaultSleepThreshold       = matrix.Float(0.5)
	defaultLinearSleepVelocity  = matrix.Float(0.05)
	defaultAngularSleepVelocity = matrix.Float(0.05)
)

type RigidBody struct {
	Transform   matrix.Transform
	MotionState MotionState
	Mass        Mass
	Collision   CollisionInfo
	Simulation  SimulationState
	Active      bool
	poolId      pooling.PoolGroupId
	id          pooling.PoolIndex
	pooled      bool
}

type MotionState struct {
	Acceleration        matrix.Vec3
	AngularAcceleration matrix.Vec3
	LinearVelocity      matrix.Vec3
	AngularVelocity     matrix.Vec3
}

type Mass struct {
	Inertia        matrix.Vec3
	inverseInertia matrix.Vec3
	Mass           matrix.Float
	inverseMass    matrix.Float
}

type CollisionInfo struct {
	Shape     Shape
	Mesh      *MeshCollision
	Terrain   *TerrainCollision
	LocalAABB AABB
	Group     int
	Mask      int
	IsTrigger bool
}

type SimulationState struct {
	Type             RigidBodyType
	SleepThreshold   matrix.Float
	SleepTimer       matrix.Float
	IsSleeping       bool
	IsFixedRotation  bool
	IsFixedPosition  bool
	lastPosition     matrix.Vec3
	lastRotation     matrix.Vec3
	lastScale        matrix.Vec3
	hasLastTransform bool
}

func (r *RigidBody) poolLocation() int {
	return int(r.poolId)<<8 | int(r.id)
}

func (r *RigidBody) IsStatic() bool {
	return r.Simulation.Type == RigidBodyTypeStatic || r.Mass.Mass == 0
}

func (r *RigidBody) IsDynamic() bool {
	return r.Simulation.Type == RigidBodyTypeDynamic && r.Mass.inverseMass > 0
}

func (r *RigidBody) IsKinematic() bool {
	return r.Simulation.Type == RigidBodyTypeKinematic
}

func (r *RigidBody) SetDynamic(mass matrix.Float, inertia matrix.Vec3) {
	r.Active = true
	r.Simulation.Type = RigidBodyTypeDynamic
	r.Wake()
	r.ensureDefaultSleepThreshold()
	r.SetMass(mass, inertia)
	r.ensureDefaultCollisionFilter()
}

func (r *RigidBody) SetStatic() {
	r.Active = true
	r.Simulation.Type = RigidBodyTypeStatic
	r.Wake()
	r.SetMass(0, matrix.Vec3Zero())
	r.MotionState = MotionState{}
	r.ensureDefaultCollisionFilter()
}

// SetKinematic makes this body entity-driven: the stage sync copies the entity
// transform into the body before collision detection, and the solver treats it
// as immovable because kinematic bodies have no inverse mass.
func (r *RigidBody) SetKinematic() {
	r.Active = true
	r.Simulation.Type = RigidBodyTypeKinematic
	r.Wake()
	r.SetMass(0, matrix.Vec3Zero())
	r.ensureDefaultCollisionFilter()
}

func (r *RigidBody) SetShape(shape Shape) {
	r.Collision.Shape = shape
	r.Collision.Mesh = nil
	r.Collision.Terrain = nil
	r.Collision.LocalAABB = AABB{}
	r.ensureDefaultCollisionFilter()
}

func (r *RigidBody) SetShapeMesh(mesh *MeshCollision) {
	r.Collision.Mesh = mesh
	r.Collision.Terrain = nil
	bounds := NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero())
	if mesh != nil {
		bounds = mesh.Bounds
	}
	r.Collision.Shape = NewMeshShape(bounds)
	r.Collision.LocalAABB = bounds
	r.ensureDefaultCollisionFilter()
}

func (r *RigidBody) SetStaticTerrain(terrain *TerrainCollision) {
	r.Collision.Mesh = nil
	r.Collision.Terrain = terrain
	bounds := NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero())
	if terrain != nil {
		bounds = terrain.Bounds
	}
	r.Collision.Shape = NewTerrainShape(bounds)
	r.Collision.LocalAABB = bounds
	r.SetStatic()
	r.ensureDefaultCollisionFilter()
}

func (r *RigidBody) Shape() Shape {
	return r.Collision.Shape
}

func (r *RigidBody) SetCollisionFilter(group, mask int) {
	r.Collision.Group = group
	r.Collision.Mask = mask
}

func (r *RigidBody) CollisionFilter() (int, int) {
	return r.Collision.Group, r.Collision.Mask
}

func (r *RigidBody) SetTrigger(isTrigger bool) {
	r.Collision.IsTrigger = isTrigger
}

func (r *RigidBody) IsTrigger() bool {
	return r.Collision.IsTrigger
}

func (r *RigidBody) Position() matrix.Vec3 {
	return r.Transform.WorldPosition()
}

func (r *RigidBody) Rotation() matrix.Quaternion {
	return matrix.QuaternionFromEuler(r.Transform.WorldRotation())
}

// ApplyForce applies a continuous world-space force at the body's center of mass.
func (r *RigidBody) ApplyForce(force matrix.Vec3) {
	if r != nil && !force.IsZero() {
		r.Wake()
	}
	r.applyForce(force, matrix.Vec3Zero())
}

// ApplyForceAtPoint applies a continuous world-space force at a world-space point.
func (r *RigidBody) ApplyForceAtPoint(force, point matrix.Vec3) {
	if r == nil {
		return
	}
	if !force.IsZero() {
		r.Wake()
	}
	r.applyForce(force, point.Subtract(r.Transform.WorldPosition()))
}

// ApplyImpulse applies an immediate world-space impulse at the body's center of mass.
func (r *RigidBody) ApplyImpulse(impulse matrix.Vec3) {
	if r != nil && !impulse.IsZero() {
		r.Wake()
	}
	r.applyImpulse(impulse, matrix.Vec3Zero())
}

// ApplyImpulseAtPoint applies an immediate world-space impulse at a world-space point.
func (r *RigidBody) ApplyImpulseAtPoint(impulse, point matrix.Vec3) {
	if r == nil {
		return
	}
	if !impulse.IsZero() {
		r.Wake()
	}
	r.applyImpulse(impulse, point.Subtract(r.Transform.WorldPosition()))
}

func (r *RigidBody) Wake() {
	if r == nil {
		return
	}
	r.Simulation.IsSleeping = false
	r.Simulation.SleepTimer = 0
	r.recordSleepTransform()
}

func (r *RigidBody) Sleep() {
	if r == nil || !r.canAutoSleep() {
		return
	}
	r.Simulation.IsSleeping = true
	r.Simulation.SleepTimer = r.sleepThreshold()
	r.MotionState = MotionState{}
	r.recordSleepTransform()
}

func (r *RigidBody) applyForce(force, rOffset matrix.Vec3) {
	invMass := r.inverseMass()
	if invMass == 0 {
		return
	}
	r.MotionState.Acceleration.AddAssign(force.Scale(invMass))
	invInertia := r.inverseInertia()
	if invInertia.IsZero() {
		return
	}
	angularAcceleration := rOffset.Cross(force).Multiply(invInertia)
	r.MotionState.AngularAcceleration.AddAssign(angularAcceleration)
}

func (r *RigidBody) applyImpulse(impulse, rOffset matrix.Vec3) {
	invMass := r.inverseMass()
	if invMass == 0 {
		return
	}
	r.MotionState.LinearVelocity.AddAssign(impulse.Scale(invMass))
	invInertia := r.inverseInertia()
	if invInertia.IsZero() {
		return
	}
	angularImpulse := rOffset.Cross(impulse).Multiply(invInertia)
	r.MotionState.AngularVelocity.AddAssign(angularImpulse)
}

func (r *RigidBody) inverseMass() matrix.Float {
	if r == nil || !r.IsDynamic() || r.Simulation.IsSleeping || r.Simulation.IsFixedPosition {
		return 0
	}
	return r.Mass.inverseMass
}

func (r *RigidBody) inverseInertia() matrix.Vec3 {
	if r == nil || !r.IsDynamic() || r.Simulation.IsSleeping || r.Simulation.IsFixedRotation {
		return matrix.Vec3Zero()
	}
	return r.Mass.inverseInertia
}

func (r *RigidBody) SetMass(mass matrix.Float, inertia matrix.Vec3) {
	r.Mass.Mass = mass
	if mass > 0 {
		r.Mass.inverseMass = 1.0 / mass
	} else {
		r.Mass.inverseMass = 0
	}
	r.Mass.Inertia = inertia
	r.Mass.inverseInertia = matrix.Vec3{}
	for i := range r.Mass.inverseInertia {
		if inertia[i] > 0 {
			r.Mass.inverseInertia[i] = 1.0 / inertia[i]
		}
	}
}

func (r *RigidBody) ensureDefaultCollisionFilter() {
	if r.Collision.Mask == 0 {
		r.Collision.Group = DefaultCollisionGroup
		r.Collision.Mask = DefaultCollisionMask
	}
}

func (r *RigidBody) ensureDefaultSleepThreshold() {
	if r.Simulation.SleepThreshold <= 0 {
		r.Simulation.SleepThreshold = DefaultSleepThreshold
	}
}

func (r *RigidBody) sleepThreshold() matrix.Float {
	if r.Simulation.SleepThreshold > 0 {
		return r.Simulation.SleepThreshold
	}
	return DefaultSleepThreshold
}

func (r *RigidBody) canAutoSleep() bool {
	return r != nil && r.Active && r.IsDynamic()
}

func (r *RigidBody) canWakeOnContact() bool {
	return r != nil && r.Active && !r.Simulation.IsSleeping && !r.IsStatic()
}

func (r *RigidBody) isBelowSleepVelocity() bool {
	linearLimit := defaultLinearSleepVelocity * defaultLinearSleepVelocity
	angularLimit := defaultAngularSleepVelocity * defaultAngularSleepVelocity
	return r.MotionState.LinearVelocity.LengthSquared() <= linearLimit &&
		r.MotionState.AngularVelocity.LengthSquared() <= angularLimit
}

func (r *RigidBody) wakeIfTransformChanged() {
	if r == nil || !r.Simulation.IsSleeping || !r.Simulation.hasLastTransform {
		return
	}
	if !r.Simulation.lastPosition.Equals(r.Transform.WorldPosition()) ||
		!r.Simulation.lastRotation.Equals(r.Transform.WorldRotation()) ||
		!r.Simulation.lastScale.Equals(r.Transform.WorldScale()) {
		r.Wake()
	}
}

func (r *RigidBody) recordSleepTransform() {
	if r == nil {
		return
	}
	r.Simulation.lastPosition = r.Transform.WorldPosition()
	r.Simulation.lastRotation = r.Transform.WorldRotation()
	r.Simulation.lastScale = r.Transform.WorldScale()
	r.Simulation.hasLastTransform = true
}

func (r *RigidBody) WorldAABB() AABB {
	if r.Collision.Shape.Type == ShapeTypeMesh && r.Collision.Mesh != nil {
		return r.Collision.Mesh.Bounds.Transform(r.Transform.WorldMatrix())
	}
	if r.Collision.Shape.Type == ShapeTypeTerrain && r.Collision.Terrain != nil {
		return r.Collision.Terrain.Bounds.Transform(r.Transform.WorldMatrix())
	}
	if r.Collision.LocalAABB.Type == ShapeTypeAABB {
		return r.Collision.LocalAABB.Transform(r.Transform.WorldMatrix())
	}
	return shapeWorldAABB(worldShape(r))
}

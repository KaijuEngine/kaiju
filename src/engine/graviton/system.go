/******************************************************************************/
/* system.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

var (
	// Common gravity `a` for Earth, used as default value
	standardGravity matrix.Float = -9.81
)

type System struct {
	bodies      pooling.PoolGroup[RigidBody]
	constraints pooling.PoolGroup[Constraint]
	// This is a singular vector at the moment, I'll be making
	// multiple gravitational sources in the future
	gravity matrix.Vec3
	// Constraint iteration counts are shared by contact and constraint solving
	// because System.Step solves them together in the same islands.
	ConstraintVelocityIterations int
	ConstraintPositionIterations int
	broadPhase                   SweepPrune
	narrowPhase                  NarrowPhase
	solver                       CollisionSolver
	constraintScratch            []*Constraint
}

func (s *System) Initialize() {
	// Take the ith unit vector and scale it proportionally to standard gravity
	s.gravity = matrix.Vec3Up().Scale(standardGravity)
	s.broadPhase.Initialize(1024)
	s.solver.Initialize()
	s.ConstraintVelocityIterations = s.solver.VelocityIterations
	s.ConstraintPositionIterations = s.solver.PositionIterations
}

func (s *System) SetGravity(gravity matrix.Vec3) {
	s.gravity = gravity
}

func (s *System) NewBody() *RigidBody {
	body, pool, id := s.bodies.Add()
	*body = RigidBody{}
	body.poolId = pool
	body.id = id
	body.pooled = true
	body.Transform.SetupRawTransform()
	return body
}

func (s *System) AddBody(body *RigidBody) *RigidBody {
	if body == nil {
		return nil
	}
	if body.pooled {
		body.ensureDefaultSleepThreshold()
		body.recordSleepTransform()
		return body
	}
	stageBody := s.NewBody()
	stageBody.Transform.SetPosition(body.Transform.WorldPosition())
	stageBody.Transform.SetRotation(body.Transform.WorldRotation())
	stageBody.Transform.SetScale(body.Transform.WorldScale())
	stageBody.MotionState = body.MotionState
	stageBody.Mass = body.Mass
	stageBody.Collision = body.Collision
	stageBody.Simulation = body.Simulation
	stageBody.Active = body.Active
	stageBody.ensureDefaultSleepThreshold()
	stageBody.recordSleepTransform()
	return stageBody
}

func (s *System) NewConstraint(constraintType ConstraintType, bodyA, bodyB *RigidBody) *Constraint {
	constraint, pool, id := s.constraints.Add()
	*constraint = Constraint{}
	constraint.Type = constraintType
	constraint.poolId = pool
	constraint.id = id
	constraint.pooled = true
	constraint.Active = true
	constraint.Enabled = true
	constraint.SetBodies(bodyA, bodyB)
	return constraint
}

func (s *System) AddConstraint(constraint *Constraint) *Constraint {
	if constraint == nil {
		return nil
	}
	if constraint.pooled {
		constraint.disableIfBodiesInvalid()
		return constraint
	}
	return s.AddConstraintWithBodies(constraint, constraint.BodyA, constraint.BodyB)
}

func (s *System) AddConstraintWithBodies(constraint *Constraint, bodyA, bodyB *RigidBody) *Constraint {
	if constraint == nil {
		return nil
	}
	stageConstraint := s.NewConstraint(constraint.Type, bodyA, bodyB)
	stageConstraint.Active = constraint.Active
	stageConstraint.Enabled = constraint.Enabled
	stageConstraint.BreakForce = constraint.BreakForce
	stageConstraint.BreakTorque = constraint.BreakTorque
	stageConstraint.Broken = constraint.Broken
	stageConstraint.Rows = append(stageConstraint.Rows, constraint.Rows...)
	if constraint.Distance != nil {
		distance := *constraint.Distance
		distance.BodyA = stageConstraint.BodyA
		distance.BodyB = stageConstraint.BodyB
		distance.constraint = stageConstraint
		stageConstraint.Distance = &distance
	}
	if constraint.Rope != nil {
		rope := *constraint.Rope
		rope.BodyA = stageConstraint.BodyA
		rope.BodyB = stageConstraint.BodyB
		rope.constraint = stageConstraint
		stageConstraint.Rope = &rope
	}
	if constraint.Point != nil {
		point := *constraint.Point
		point.BodyA = stageConstraint.BodyA
		point.BodyB = stageConstraint.BodyB
		point.constraint = stageConstraint
		stageConstraint.Point = &point
	}
	if constraint.Hinge != nil {
		hinge := *constraint.Hinge
		hinge.BodyA = stageConstraint.BodyA
		hinge.BodyB = stageConstraint.BodyB
		hinge.constraint = stageConstraint
		stageConstraint.Hinge = &hinge
	}
	stageConstraint.disableIfBodiesInvalid()
	return stageConstraint
}

func (s *System) NewDistanceJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB matrix.Vec3) *DistanceJoint {
	constraint := s.NewConstraint(ConstraintTypeDistance, bodyA, bodyB)
	joint := NewDistanceJoint(bodyA, bodyB, localAnchorA, localAnchorB)
	joint.constraint = constraint
	constraint.Distance = joint
	return joint
}

func (s *System) NewDistanceJointAtWorldAnchors(bodyA, bodyB *RigidBody, worldAnchorA, worldAnchorB matrix.Vec3) *DistanceJoint {
	return s.NewDistanceJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchorA),
		LocalAnchor(bodyB, worldAnchorB),
	)
}

func (s *System) NewDistanceJointToWorld(body *RigidBody, localAnchor, worldAnchor matrix.Vec3) *DistanceJoint {
	return s.NewDistanceJoint(body, nil, localAnchor, worldAnchor)
}

func (s *System) AddDistanceJoint(joint *DistanceJoint) *DistanceJoint {
	if joint == nil {
		return nil
	}
	if joint.constraint != nil && joint.constraint.pooled {
		joint.constraint.disableIfBodiesInvalid()
		return joint
	}
	constraint := s.NewConstraint(ConstraintTypeDistance, joint.BodyA, joint.BodyB)
	stageJoint := *joint
	stageJoint.BodyA = constraint.BodyA
	stageJoint.BodyB = constraint.BodyB
	stageJoint.constraint = constraint
	constraint.Distance = &stageJoint
	return &stageJoint
}

func (s *System) RemoveDistanceJoint(joint *DistanceJoint) {
	if joint == nil {
		return
	}
	s.RemoveConstraint(joint.constraint)
}

func (s *System) NewRopeJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB matrix.Vec3) *RopeJoint {
	constraint := s.NewConstraint(ConstraintTypeRope, bodyA, bodyB)
	joint := NewRopeJoint(bodyA, bodyB, localAnchorA, localAnchorB)
	joint.constraint = constraint
	constraint.Rope = joint
	return joint
}

func (s *System) NewRopeJointAtWorldAnchors(bodyA, bodyB *RigidBody, worldAnchorA, worldAnchorB matrix.Vec3) *RopeJoint {
	return s.NewRopeJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchorA),
		LocalAnchor(bodyB, worldAnchorB),
	)
}

func (s *System) NewRopeJointToWorld(body *RigidBody, localAnchor, worldAnchor matrix.Vec3) *RopeJoint {
	return s.NewRopeJoint(body, nil, localAnchor, worldAnchor)
}

func (s *System) AddRopeJoint(joint *RopeJoint) *RopeJoint {
	if joint == nil {
		return nil
	}
	if joint.constraint != nil && joint.constraint.pooled {
		joint.constraint.disableIfBodiesInvalid()
		return joint
	}
	constraint := s.NewConstraint(ConstraintTypeRope, joint.BodyA, joint.BodyB)
	stageJoint := *joint
	stageJoint.BodyA = constraint.BodyA
	stageJoint.BodyB = constraint.BodyB
	stageJoint.constraint = constraint
	constraint.Rope = &stageJoint
	return &stageJoint
}

func (s *System) RemoveRopeJoint(joint *RopeJoint) {
	if joint == nil {
		return
	}
	s.RemoveConstraint(joint.constraint)
}

func (s *System) NewPointJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB matrix.Vec3) *PointJoint {
	constraint := s.NewConstraint(ConstraintTypePoint, bodyA, bodyB)
	joint := NewPointJoint(bodyA, bodyB, localAnchorA, localAnchorB)
	joint.constraint = constraint
	constraint.Point = joint
	return joint
}

func (s *System) NewPointJointAtWorldAnchors(bodyA, bodyB *RigidBody, worldAnchorA, worldAnchorB matrix.Vec3) *PointJoint {
	return s.NewPointJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchorA),
		LocalAnchor(bodyB, worldAnchorB),
	)
}

func (s *System) NewPointJointToWorld(body *RigidBody, localAnchor, worldAnchor matrix.Vec3) *PointJoint {
	return s.NewPointJoint(body, nil, localAnchor, worldAnchor)
}

func (s *System) AddPointJoint(joint *PointJoint) *PointJoint {
	if joint == nil {
		return nil
	}
	if joint.constraint != nil && joint.constraint.pooled {
		joint.constraint.disableIfBodiesInvalid()
		return joint
	}
	constraint := s.NewConstraint(ConstraintTypePoint, joint.BodyA, joint.BodyB)
	stageJoint := *joint
	stageJoint.BodyA = constraint.BodyA
	stageJoint.BodyB = constraint.BodyB
	stageJoint.constraint = constraint
	constraint.Point = &stageJoint
	return &stageJoint
}

func (s *System) RemovePointJoint(joint *PointJoint) {
	if joint == nil {
		return
	}
	s.RemoveConstraint(joint.constraint)
}

func (s *System) NewHingeJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB, localAxisA, localAxisB matrix.Vec3) *HingeJoint {
	constraint := s.NewConstraint(ConstraintTypeHinge, bodyA, bodyB)
	joint := NewHingeJoint(bodyA, bodyB, localAnchorA, localAnchorB, localAxisA, localAxisB)
	joint.constraint = constraint
	constraint.Hinge = joint
	return joint
}

func (s *System) NewHingeJointAtWorldAnchor(bodyA, bodyB *RigidBody, worldAnchor, worldAxis matrix.Vec3) *HingeJoint {
	return s.NewHingeJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchor),
		LocalAnchor(bodyB, worldAnchor),
		LocalAxis(bodyA, worldAxis),
		LocalAxis(bodyB, worldAxis),
	)
}

func (s *System) NewHingeJointToWorld(body *RigidBody, localAnchor, worldAnchor, localAxis, worldAxis matrix.Vec3) *HingeJoint {
	return s.NewHingeJoint(body, nil, localAnchor, worldAnchor, localAxis, worldAxis)
}

func (s *System) AddHingeJoint(joint *HingeJoint) *HingeJoint {
	if joint == nil {
		return nil
	}
	if joint.constraint != nil && joint.constraint.pooled {
		joint.constraint.disableIfBodiesInvalid()
		return joint
	}
	constraint := s.NewConstraint(ConstraintTypeHinge, joint.BodyA, joint.BodyB)
	stageJoint := *joint
	stageJoint.BodyA = constraint.BodyA
	stageJoint.BodyB = constraint.BodyB
	stageJoint.constraint = constraint
	constraint.Hinge = &stageJoint
	return &stageJoint
}

func (s *System) RemoveHingeJoint(joint *HingeJoint) {
	if joint == nil {
		return
	}
	s.RemoveConstraint(joint.constraint)
}

func (s *System) RemoveConstraint(constraint *Constraint) {
	if constraint == nil || !constraint.pooled {
		return
	}
	poolId := constraint.poolId
	id := constraint.id
	constraint.Active = false
	constraint.Enabled = false
	constraint.pooled = false
	s.constraints.Remove(poolId, id)
	*constraint = Constraint{}
}

func (s *System) ClearConstraints() {
	s.constraints.Each(func(constraint *Constraint) {
		constraint.Active = false
		constraint.Enabled = false
		constraint.pooled = false
		*constraint = Constraint{}
	})
	s.constraints.Clear()
}

// RemoveBody releases a body and disables any constraints attached to it. The
// disabled constraints remain in constraint storage until explicitly removed or
// cleared, with the removed body endpoint set to nil.
func (s *System) RemoveBody(body *RigidBody) {
	if body == nil || !body.pooled {
		return
	}
	s.constraints.Each(func(constraint *Constraint) {
		if constraint.BodyA == body || constraint.BodyB == body {
			constraint.detachBody(body)
		}
	})
	poolId := body.poolId
	id := body.id
	body.Active = false
	body.pooled = false
	s.bodies.Remove(poolId, id)
	*body = RigidBody{}
}

func (s *System) Clear() {
	s.ClearConstraints()
	s.bodies.Each(func(body *RigidBody) {
		body.Active = false
		body.pooled = false
		*body = RigidBody{}
	})
	s.bodies.Clear()
	s.broadPhase.Rebuild(&s.bodies)
	s.narrowPhase.Reset()
	s.solver.Reset()
	s.constraintScratch = s.constraintScratch[:0]
}

func (s *System) Step(workGroup *concurrent.WorkGroup, threads *concurrent.Threads, deltaTime float64) {
	dt := matrix.Float(deltaTime)
	s.solver.DeltaTime = dt
	s.solver.VelocityIterations = s.constraintVelocityIterations()
	s.solver.PositionIterations = s.constraintPositionIterations()
	s.prepareSleepState()
	s.bodies.EachParallel("kaiju.phys", workGroup, threads, func(body *RigidBody) {
		if !body.Active || body.Simulation.IsSleeping || !body.IsDynamic() {
			return
		}
		ms := &body.MotionState
		ms.Acceleration.AddAssign(s.gravity)
		ms.LinearVelocity.AddAssign(ms.Acceleration.Scale(dt))
		ms.AngularVelocity.AddAssign(ms.AngularAcceleration.Scale(dt))
		if !body.Simulation.IsFixedPosition {
			body.Transform.AddPosition(ms.LinearVelocity.Scale(dt))
		}
		if !body.Simulation.IsFixedRotation {
			body.Transform.SetRotation(integrateAngularVelocity(body.Transform.Rotation(), ms.AngularVelocity, dt))
		}
		ms.Acceleration = matrix.Vec3{}
		ms.AngularAcceleration = matrix.Vec3{}
	})
	s.broadPhase.RebuildParallel(&s.bodies, threads)
	pairs := s.broadPhase.SweepParallel(threads, s.canBroadPhaseCollide)
	manifolds := s.narrowPhase.Collide(pairs, threads)
	constraints := s.activeConstraints()
	s.wakeContacts(manifolds)
	s.wakeConstraints(constraints)
	// Contacts and constraints are solved as one island problem so linked
	// bodies share the same velocity and position iteration stream.
	s.solver.SolveWithConstraints(manifolds, constraints, threads)
	s.updateSleepState(dt)
}

// Contacts returns the contact manifolds generated during the most recent Step.
// The returned slice is owned by the System and is reused on the next Step.
func (s *System) Contacts() []ContactManifold {
	return s.narrowPhase.Manifolds()
}

// Constraints returns the constraints currently stored in the System. The
// returned slice is owned by the System and is reused on the next constraints
// query or Step.
func (s *System) Constraints() []*Constraint {
	s.constraintScratch = s.constraintScratch[:0]
	s.constraints.Each(func(constraint *Constraint) {
		if constraint != nil {
			s.constraintScratch = append(s.constraintScratch, constraint)
		}
	})
	return s.constraintScratch
}

func (s *System) activeConstraints() []*Constraint {
	s.constraintScratch = s.constraintScratch[:0]
	s.constraints.Each(func(constraint *Constraint) {
		if constraint == nil {
			return
		}
		constraint.disableIfBodiesInvalid()
		if constraint.Active && constraint.Enabled {
			s.constraintScratch = append(s.constraintScratch, constraint)
		}
	})
	return s.constraintScratch
}

func (s *System) canBroadPhaseCollide(a, b *RigidBody) bool {
	if a == nil || b == nil {
		return false
	}
	if a.IsStatic() && b.IsStatic() {
		return false
	}
	return s.canCollide(a, b)
}

func (s *System) canCollide(a, b *RigidBody) bool {
	if a.Collision.Mask&(1<<b.Collision.Group) == 0 {
		return false
	}
	if b.Collision.Mask&(1<<a.Collision.Group) == 0 {
		return false
	}
	return true
}

func (s *System) constraintVelocityIterations() int {
	if s.ConstraintVelocityIterations < 0 {
		return 0
	}
	return s.ConstraintVelocityIterations
}

func (s *System) constraintPositionIterations() int {
	if s.ConstraintPositionIterations < 0 {
		return 0
	}
	return s.ConstraintPositionIterations
}

func (s *System) prepareSleepState() {
	s.bodies.Each(func(body *RigidBody) {
		if body == nil || !body.Active {
			return
		}
		body.ensureDefaultSleepThreshold()
		body.wakeIfTransformChanged()
	})
}

func (s *System) wakeContacts(manifolds []ContactManifold) {
	for i := range manifolds {
		manifold := &manifolds[i]
		if manifold.Count == 0 {
			continue
		}
		aSleeping := manifold.BodyA != nil && manifold.BodyA.Simulation.IsSleeping
		bSleeping := manifold.BodyB != nil && manifold.BodyB.Simulation.IsSleeping
		if aSleeping && manifold.BodyB.canWakeOnContact() {
			manifold.BodyA.Wake()
		}
		if bSleeping && manifold.BodyA.canWakeOnContact() {
			manifold.BodyB.Wake()
		}
	}
}

func (s *System) wakeConstraints(constraints []*Constraint) {
	for _, constraint := range constraints {
		if constraint == nil {
			continue
		}
		if constraint.IsStretched() {
			constraint.WakeBodies()
			continue
		}
		wakeSleepingConstraintEndpoint(constraint.BodyA, constraint.BodyB)
		wakeSleepingConstraintEndpoint(constraint.BodyB, constraint.BodyA)
	}
}

func wakeSleepingConstraintEndpoint(sleeping, other *RigidBody) {
	if sleeping != nil && sleeping.Simulation.IsSleeping && other.canWakeOnContact() {
		sleeping.Wake()
	}
}

func (s *System) updateSleepState(dt matrix.Float) {
	s.bodies.Each(func(body *RigidBody) {
		if body == nil || !body.Active {
			return
		}
		if !body.canAutoSleep() {
			body.Simulation.SleepTimer = 0
			body.recordSleepTransform()
			return
		}
		if body.Simulation.IsSleeping {
			body.recordSleepTransform()
			return
		}
		if s.bodyHasActiveStretchedConstraint(body) {
			body.Simulation.SleepTimer = 0
			body.recordSleepTransform()
			return
		}
		if body.isBelowSleepVelocity() {
			body.Simulation.SleepTimer += dt
			if body.Simulation.SleepTimer >= body.sleepThreshold() {
				body.Sleep()
				return
			}
		} else {
			body.Simulation.SleepTimer = 0
		}
		body.recordSleepTransform()
	})
}

func (s *System) bodyHasActiveStretchedConstraint(body *RigidBody) bool {
	if body == nil {
		return false
	}
	stretched := false
	s.constraints.Each(func(constraint *Constraint) {
		if stretched || constraint == nil || !constraint.Active || !constraint.Enabled {
			return
		}
		if constraint.BodyA != body && constraint.BodyB != body {
			return
		}
		stretched = constraint.IsStretched()
	})
	return stretched
}

func integrateAngularVelocity(rotation, angularVelocity matrix.Vec3, dt matrix.Float) matrix.Vec3 {
	speed := angularVelocity.Length()
	if speed <= matrix.FloatSmallestNonzero || dt <= 0 {
		return rotation
	}
	axis := angularVelocity.Scale(1.0 / speed)
	delta := matrix.QuaternionAxisAngle(axis, speed*dt)
	current := matrix.QuaternionFromEuler(rotation)
	next := delta.Multiply(current)
	next.Normalize()
	return next.ToEuler()
}

/******************************************************************************/
/* constraint_solver.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
)

// ConstraintSolverRow stores a single scalar Jacobian row for iterative
// constraint solving. The row applies impulses along Axis at both anchors,
// where BodyA receives the negative impulse and BodyB receives the positive
// impulse.
type ConstraintSolverRow struct {
	BodyA              *RigidBody
	BodyB              *RigidBody
	Axis               matrix.Vec3
	JacobianLinearA    matrix.Vec3
	JacobianAngularA   matrix.Vec3
	JacobianLinearB    matrix.Vec3
	JacobianAngularB   matrix.Vec3
	AnchorA            matrix.Vec3
	AnchorB            matrix.Vec3
	RelativeAnchorA    matrix.Vec3
	RelativeAnchorB    matrix.Vec3
	EffectiveMass      matrix.Float
	Bias               matrix.Float
	AccumulatedImpulse matrix.Float
	MinImpulse         matrix.Float
	MaxImpulse         matrix.Float
}

// AngularConstraintSolverRow stores a scalar angular-only Jacobian row. It is
// used by constraints that must remove relative angular velocity around a
// world-space axis without applying any linear impulse.
type AngularConstraintSolverRow struct {
	BodyA              *RigidBody
	BodyB              *RigidBody
	Axis               matrix.Vec3
	EffectiveMass      matrix.Float
	Bias               matrix.Float
	AccumulatedImpulse matrix.Float
	MinImpulse         matrix.Float
	MaxImpulse         matrix.Float
}

func NewConstraintSolverRow(bodyA, bodyB *RigidBody, anchorA, anchorB, axis matrix.Vec3) ConstraintSolverRow {
	row := ConstraintSolverRow{}
	row.SetWorldAnchors(bodyA, bodyB, anchorA, anchorB, axis)
	return row
}

func (r *ConstraintSolverRow) SetLocalAnchors(bodyA, bodyB *RigidBody, anchorA, anchorB, axis matrix.Vec3) {
	r.SetWorldAnchors(bodyA, bodyB, WorldAnchor(bodyA, anchorA), WorldAnchor(bodyB, anchorB), axis)
}

func (r *ConstraintSolverRow) SetWorldAnchors(bodyA, bodyB *RigidBody, anchorA, anchorB, axis matrix.Vec3) {
	r.BodyA = bodyA
	r.BodyB = bodyB
	r.Axis = safeNormal(axis, matrix.Vec3Right())
	r.AnchorA = anchorA
	r.AnchorB = anchorB
	r.RelativeAnchorA = RelativeAnchorOffset(bodyA, anchorA)
	r.RelativeAnchorB = RelativeAnchorOffset(bodyB, anchorB)
	r.JacobianLinearA = r.Axis.Negative()
	r.JacobianAngularA = r.RelativeAnchorA.Cross(r.Axis).Negative()
	r.JacobianLinearB = r.Axis
	r.JacobianAngularB = r.RelativeAnchorB.Cross(r.Axis)
	r.EffectiveMass = ConstraintEffectiveMass(bodyA, bodyB, r.RelativeAnchorA, r.RelativeAnchorB, r.Axis)
	r.MinImpulse = -matrix.Inf(1)
	r.MaxImpulse = matrix.Inf(1)
}

func (r *ConstraintSolverRow) SetImpulseLimits(minimum, maximum matrix.Float) {
	r.MinImpulse = minimum
	r.MaxImpulse = maximum
}

func (r *ConstraintSolverRow) RelativeVelocity() matrix.Float {
	return constraintBodyVelocity(r.BodyA, r.JacobianLinearA, r.JacobianAngularA) +
		constraintBodyVelocity(r.BodyB, r.JacobianLinearB, r.JacobianAngularB)
}

func (r *ConstraintSolverRow) ApplyImpulse(impulse matrix.Float) {
	if impulse == 0 {
		return
	}
	WakeConstrainedBodies(r.BodyA, r.BodyB)
	linearImpulse := r.Axis.Scale(impulse)
	applyImpulse(r.BodyA, linearImpulse.Negative(), r.RelativeAnchorA)
	applyImpulse(r.BodyB, linearImpulse, r.RelativeAnchorB)
}

func (r *ConstraintSolverRow) Solve() matrix.Float {
	if r.EffectiveMass <= 0 {
		return 0
	}
	impulse := -(r.RelativeVelocity() + r.Bias) * r.EffectiveMass
	previousImpulse := r.AccumulatedImpulse
	minimum, maximum := r.impulseLimits()
	r.AccumulatedImpulse = klib.Clamp(previousImpulse+impulse, minimum, maximum)
	impulse = r.AccumulatedImpulse - previousImpulse
	r.ApplyImpulse(impulse)
	return impulse
}

func (r *ConstraintSolverRow) impulseLimits() (matrix.Float, matrix.Float) {
	if r.MinImpulse == 0 && r.MaxImpulse == 0 {
		return -matrix.Inf(1), matrix.Inf(1)
	}
	return r.MinImpulse, r.MaxImpulse
}

func (r *AngularConstraintSolverRow) SetWorldAxis(bodyA, bodyB *RigidBody, axis matrix.Vec3) {
	r.BodyA = bodyA
	r.BodyB = bodyB
	r.Axis = safeNormal(axis, matrix.Vec3Right())
	r.EffectiveMass = AngularConstraintEffectiveMass(bodyA, bodyB, r.Axis)
	r.MinImpulse = -matrix.Inf(1)
	r.MaxImpulse = matrix.Inf(1)
}

func (r *AngularConstraintSolverRow) SetImpulseLimits(minimum, maximum matrix.Float) {
	r.MinImpulse = minimum
	r.MaxImpulse = maximum
}

func (r *AngularConstraintSolverRow) RelativeVelocity() matrix.Float {
	return constraintBodyAngularVelocity(r.BodyA, r.Axis.Negative()) +
		constraintBodyAngularVelocity(r.BodyB, r.Axis)
}

func (r *AngularConstraintSolverRow) ApplyImpulse(impulse matrix.Float) {
	if impulse == 0 {
		return
	}
	WakeConstrainedBodies(r.BodyA, r.BodyB)
	applyAngularImpulse(r.BodyA, r.Axis.Scale(-impulse))
	applyAngularImpulse(r.BodyB, r.Axis.Scale(impulse))
}

func (r *AngularConstraintSolverRow) Solve() matrix.Float {
	if r.EffectiveMass <= 0 {
		return 0
	}
	impulse := -(r.RelativeVelocity() + r.Bias) * r.EffectiveMass
	previousImpulse := r.AccumulatedImpulse
	minimum, maximum := r.impulseLimits()
	r.AccumulatedImpulse = klib.Clamp(previousImpulse+impulse, minimum, maximum)
	impulse = r.AccumulatedImpulse - previousImpulse
	r.ApplyImpulse(impulse)
	return impulse
}

func (r *AngularConstraintSolverRow) impulseLimits() (matrix.Float, matrix.Float) {
	if r.MinImpulse == 0 && r.MaxImpulse == 0 {
		return -matrix.Inf(1), matrix.Inf(1)
	}
	return r.MinImpulse, r.MaxImpulse
}

func WorldAnchor(body *RigidBody, localAnchor matrix.Vec3) matrix.Vec3 {
	if body == nil {
		return localAnchor
	}
	return body.Transform.WorldMatrix().TransformPoint(localAnchor)
}

func RelativeAnchorOffset(body *RigidBody, worldAnchor matrix.Vec3) matrix.Vec3 {
	if body == nil {
		return matrix.Vec3Zero()
	}
	return worldAnchor.Subtract(body.Transform.WorldPosition())
}

func VelocityAtAnchor(body *RigidBody, relativeAnchor matrix.Vec3) matrix.Vec3 {
	if body == nil {
		return matrix.Vec3Zero()
	}
	return body.MotionState.LinearVelocity.Add(body.MotionState.AngularVelocity.Cross(relativeAnchor))
}

func AngularEffectiveMass(body *RigidBody, relativeAnchor, axis matrix.Vec3) matrix.Float {
	inverseInertia := body.inverseInertia()
	if inverseInertia.IsZero() {
		return 0
	}
	angular := relativeAnchor.Cross(axis).Multiply(inverseInertia)
	return angular.Cross(relativeAnchor).Dot(axis)
}

func ConstraintImpulseDenominator(bodyA, bodyB *RigidBody, ra, rb, axis matrix.Vec3) matrix.Float {
	denominator := bodyA.inverseMass() + bodyB.inverseMass()
	denominator += AngularEffectiveMass(bodyA, ra, axis)
	denominator += AngularEffectiveMass(bodyB, rb, axis)
	return denominator
}

func ConstraintEffectiveMass(bodyA, bodyB *RigidBody, ra, rb, axis matrix.Vec3) matrix.Float {
	denominator := ConstraintImpulseDenominator(bodyA, bodyB, ra, rb, axis)
	if denominator <= contactEpsilon {
		return 0
	}
	return 1.0 / denominator
}

func AngularConstraintImpulseDenominator(bodyA, bodyB *RigidBody, axis matrix.Vec3) matrix.Float {
	denominator := AngularAxisEffectiveMass(bodyA, axis)
	denominator += AngularAxisEffectiveMass(bodyB, axis)
	return denominator
}

func AngularConstraintEffectiveMass(bodyA, bodyB *RigidBody, axis matrix.Vec3) matrix.Float {
	denominator := AngularConstraintImpulseDenominator(bodyA, bodyB, axis)
	if denominator <= contactEpsilon {
		return 0
	}
	return 1.0 / denominator
}

func AngularAxisEffectiveMass(body *RigidBody, axis matrix.Vec3) matrix.Float {
	inverseInertia := body.inverseInertia()
	if inverseInertia.IsZero() {
		return 0
	}
	return axis.Multiply(inverseInertia).Dot(axis)
}

func WakeConstrainedBodies(bodyA, bodyB *RigidBody) {
	wakeConstrainedBody(bodyA)
	wakeConstrainedBody(bodyB)
}

func wakeConstrainedBody(body *RigidBody) {
	if body != nil && body.IsDynamic() {
		body.Wake()
	}
}

func constraintBodyVelocity(body *RigidBody, linearAxis, angularAxis matrix.Vec3) matrix.Float {
	if body == nil {
		return 0
	}
	return body.MotionState.LinearVelocity.Dot(linearAxis) +
		body.MotionState.AngularVelocity.Dot(angularAxis)
}

func constraintBodyAngularVelocity(body *RigidBody, axis matrix.Vec3) matrix.Float {
	if body == nil {
		return 0
	}
	return body.MotionState.AngularVelocity.Dot(axis)
}

func applyAngularImpulse(body *RigidBody, impulse matrix.Vec3) {
	if body == nil {
		return
	}
	invInertia := body.inverseInertia()
	if invInertia.IsZero() {
		return
	}
	body.MotionState.AngularVelocity.AddAssign(impulse.Multiply(invInertia))
}

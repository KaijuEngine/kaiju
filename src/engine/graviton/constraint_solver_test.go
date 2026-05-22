/******************************************************************************/
/* constraint_solver_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestConstraintRowAppliesEqualAndOppositeImpulse(t *testing.T) {
	bodyA := testRigidBody(Shape{}, matrix.Vec3Zero())
	bodyB := testRigidBody(Shape{}, matrix.Vec3{2, 0, 0})
	bodyA.MotionState.LinearVelocity = matrix.Vec3Right()
	bodyB.MotionState.LinearVelocity = matrix.Vec3Left()
	row := NewConstraintSolverRow(
		bodyA,
		bodyB,
		bodyA.Transform.WorldPosition(),
		bodyB.Transform.WorldPosition(),
		matrix.Vec3Right(),
	)
	impulse := row.Solve()
	if matrix.Abs(impulse-1) > 0.0001 {
		t.Fatalf("expected unit corrective impulse, got %v", impulse)
	}
	if matrix.Abs(row.AccumulatedImpulse-1) > 0.0001 {
		t.Fatalf("expected accumulated impulse to be 1, got %v", row.AccumulatedImpulse)
	}
	if !matrix.Vec3ApproxTo(bodyA.MotionState.LinearVelocity, matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected body A velocity to resolve to zero, got %v", bodyA.MotionState.LinearVelocity)
	}
	if !matrix.Vec3ApproxTo(bodyB.MotionState.LinearVelocity, matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected body B velocity to resolve to zero, got %v", bodyB.MotionState.LinearVelocity)
	}
}

func TestConstraintRowIgnoresStaticBodyImpulse(t *testing.T) {
	staticBody := &RigidBody{}
	staticBody.Transform.SetupRawTransform()
	staticBody.SetStatic()
	dynamicBody := testRigidBody(Shape{}, matrix.Vec3{1, 0, 0})
	row := NewConstraintSolverRow(
		staticBody,
		dynamicBody,
		staticBody.Transform.WorldPosition(),
		dynamicBody.Transform.WorldPosition(),
		matrix.Vec3Right(),
	)
	row.ApplyImpulse(2)
	if staticBody.inverseMass() != 0 {
		t.Fatalf("expected static body inverse mass to be zero, got %v", staticBody.inverseMass())
	}
	if !staticBody.inverseInertia().IsZero() {
		t.Fatalf("expected static body inverse inertia to be zero, got %v", staticBody.inverseInertia())
	}
	if !staticBody.MotionState.LinearVelocity.IsZero() {
		t.Fatalf("expected static body linear velocity to stay zero, got %v", staticBody.MotionState.LinearVelocity)
	}
	if !staticBody.MotionState.AngularVelocity.IsZero() {
		t.Fatalf("expected static body angular velocity to stay zero, got %v", staticBody.MotionState.AngularVelocity)
	}
	if !matrix.Vec3ApproxTo(dynamicBody.MotionState.LinearVelocity, matrix.Vec3{2, 0, 0}, 0.0001) {
		t.Fatalf("expected dynamic body to receive impulse, got %v", dynamicBody.MotionState.LinearVelocity)
	}
	kinematicBody := testRigidBody(Shape{}, matrix.Vec3Zero())
	kinematicBody.SetKinematic()
	if kinematicBody.inverseMass() != 0 || !kinematicBody.inverseInertia().IsZero() {
		t.Fatalf("expected kinematic body inverse mass/inertia to be zero, got %v and %v",
			kinematicBody.inverseMass(), kinematicBody.inverseInertia())
	}
	fixedBody := testRigidBody(Shape{}, matrix.Vec3Zero())
	fixedBody.Simulation.IsFixedPosition = true
	fixedBody.Simulation.IsFixedRotation = true
	if fixedBody.inverseMass() != 0 || !fixedBody.inverseInertia().IsZero() {
		t.Fatalf("expected fully fixed body inverse mass/inertia to be zero, got %v and %v",
			fixedBody.inverseMass(), fixedBody.inverseInertia())
	}
	if matrix.Abs(row.EffectiveMass-1) > 0.0001 {
		t.Fatalf("expected static body to contribute no effective mass, got %v", row.EffectiveMass)
	}
}

func TestConstraintRowUsesAnchorAngularVelocity(t *testing.T) {
	bodyA := testRigidBody(Shape{}, matrix.Vec3Zero())
	bodyA.MotionState.AngularVelocity = matrix.Vec3{0, 0, -1}
	staticBody := &RigidBody{}
	staticBody.Transform.SetupRawTransform()
	staticBody.SetStatic()
	anchor := matrix.Vec3{0, 1, 0}
	row := NewConstraintSolverRow(bodyA, staticBody, anchor, anchor, matrix.Vec3Right())
	if matrix.Abs(row.RelativeVelocity()+1) > 0.0001 {
		t.Fatalf("expected anchor angular velocity to produce -1 relative speed, got %v", row.RelativeVelocity())
	}
	if matrix.Abs(row.EffectiveMass-0.5) > 0.0001 {
		t.Fatalf("expected angular contribution to halve effective mass, got %v", row.EffectiveMass)
	}
	impulse := row.Solve()
	if matrix.Abs(impulse-0.5) > 0.0001 {
		t.Fatalf("expected half-unit impulse, got %v", impulse)
	}
	anchorVelocity := VelocityAtAnchor(bodyA, row.RelativeAnchorA)
	if !matrix.Vec3ApproxTo(anchorVelocity, matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected anchor velocity to resolve to zero, got %v", anchorVelocity)
	}
}

/******************************************************************************/
/* rigid_body_test.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestRigidBodyApplyForceChangesLinearVelocityOnStep(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.SetMass(2, matrix.Vec3One())
	body.ApplyForce(matrix.Vec3{4, 0, 0})
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0.5)
	expected := matrix.Vec3{1, 0, 0}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expected, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expected, body.MotionState.LinearVelocity)
	}
	if !body.MotionState.Acceleration.IsZero() {
		t.Fatalf("expected force acceleration accumulator to reset, got %v", body.MotionState.Acceleration)
	}
}

func TestRigidBodyApplyForceAtPointChangesAngularVelocityOnStep(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.SetMass(2, matrix.Vec3One())
	body.ApplyForceAtPoint(matrix.Vec3{2, 0, 0}, matrix.Vec3{0, 1, 0})
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0.5)
	expectedLinear := matrix.Vec3{0.5, 0, 0}
	expectedAngular := matrix.Vec3{0, 0, -1}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expectedLinear, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expectedLinear, body.MotionState.LinearVelocity)
	}
	if !matrix.Vec3ApproxTo(body.MotionState.AngularVelocity, expectedAngular, 0.0001) {
		t.Fatalf("expected angular velocity %v, got %v", expectedAngular, body.MotionState.AngularVelocity)
	}
	if !body.MotionState.AngularAcceleration.IsZero() {
		t.Fatalf("expected torque acceleration accumulator to reset, got %v", body.MotionState.AngularAcceleration)
	}
}

func TestRigidBodyApplyImpulseChangesLinearVelocityImmediately(t *testing.T) {
	body := testRigidBody(Shape{}, matrix.Vec3Zero())
	body.SetMass(2, matrix.Vec3One())
	body.ApplyImpulse(matrix.Vec3{4, 0, 0})
	expected := matrix.Vec3{2, 0, 0}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expected, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expected, body.MotionState.LinearVelocity)
	}
	if !body.MotionState.AngularVelocity.IsZero() {
		t.Fatalf("expected central impulse not to change angular velocity, got %v", body.MotionState.AngularVelocity)
	}
}

func TestRigidBodyApplyImpulseAtPointChangesAngularVelocityImmediately(t *testing.T) {
	body := testRigidBody(Shape{}, matrix.Vec3Zero())
	body.SetMass(2, matrix.Vec3One())
	body.ApplyImpulseAtPoint(matrix.Vec3{2, 0, 0}, matrix.Vec3{0, 1, 0})
	expectedLinear := matrix.Vec3{1, 0, 0}
	expectedAngular := matrix.Vec3{0, 0, -2}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expectedLinear, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expectedLinear, body.MotionState.LinearVelocity)
	}
	if !matrix.Vec3ApproxTo(body.MotionState.AngularVelocity, expectedAngular, 0.0001) {
		t.Fatalf("expected angular velocity %v, got %v", expectedAngular, body.MotionState.AngularVelocity)
	}
}

func TestRigidBodyApplyImpulseWakesSleepingBody(t *testing.T) {
	body := testRigidBody(Shape{}, matrix.Vec3Zero())
	body.SetMass(2, matrix.Vec3One())
	body.Sleep()
	body.ApplyImpulse(matrix.Vec3{4, 0, 0})
	expected := matrix.Vec3{2, 0, 0}
	if body.Simulation.IsSleeping {
		t.Fatal("expected impulse to wake body")
	}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expected, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expected, body.MotionState.LinearVelocity)
	}
}

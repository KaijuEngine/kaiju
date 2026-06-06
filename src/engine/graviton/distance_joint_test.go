/******************************************************************************/
/* distance_joint_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestDistanceJointMaintainsRestLength(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8
	a := addJointBody(&system, matrix.Vec3{-1, 0, 0}, RigidBodyTypeDynamic)
	b := addJointBody(&system, matrix.Vec3{1, 0, 0}, RigidBodyTypeDynamic)
	a.MotionState.LinearVelocity = matrix.Vec3Left().Scale(10)
	b.MotionState.LinearVelocity = matrix.Vec3Right().Scale(10)
	system.NewDistanceJoint(a, b, matrix.Vec3Zero(), matrix.Vec3Zero()).SetRestLength(2)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 120 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	distance := jointBodyDistance(a, b)
	if matrix.Abs(distance-2) > 0.02 {
		t.Fatalf("expected joint to maintain rest length 2, got %f", distance)
	}
}

func TestDistanceJointDynamicToStaticAnchor(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8
	dynamic := addJointBody(&system, matrix.Vec3{0, 0, 0}, RigidBodyTypeDynamic)
	static := addJointBody(&system, matrix.Vec3{0, 3, 0}, RigidBodyTypeStatic)
	dynamic.MotionState.LinearVelocity = matrix.Vec3Down().Scale(12)
	system.NewDistanceJoint(dynamic, static, matrix.Vec3Zero(), matrix.Vec3Zero()).SetRestLength(3)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 120 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	if !matrix.Vec3ApproxTo(static.Transform.WorldPosition(), matrix.Vec3{0, 3, 0}, 0.0001) {
		t.Fatalf("expected static anchor to remain fixed, got %v", static.Transform.WorldPosition())
	}
	distance := jointBodyDistance(dynamic, static)
	if matrix.Abs(distance-3) > 0.02 {
		t.Fatalf("expected dynamic-to-static joint distance 3, got %f", distance)
	}
}

func TestDistanceJointHandlesZeroLengthSafely(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8
	a := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	b := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	system.NewDistanceJoint(a, b, matrix.Vec3Zero(), matrix.Vec3Zero()).SetRestLength(1)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 120 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	if !finiteVec3(a.Transform.WorldPosition()) || !finiteVec3(b.Transform.WorldPosition()) {
		t.Fatalf("expected finite positions after zero-length start, got %v and %v",
			a.Transform.WorldPosition(), b.Transform.WorldPosition())
	}
	distance := jointBodyDistance(a, b)
	if matrix.Abs(distance-1) > 0.05 {
		t.Fatalf("expected zero-length start to settle near rest length 1, got %f", distance)
	}
}

func TestDistanceJointWithGravityDoesNotStretchBeyondTolerance(t *testing.T) {
	system := System{}
	system.Initialize()
	system.ConstraintPositionIterations = 8
	body := addJointBody(&system, matrix.Vec3{0, -2, 0}, RigidBodyTypeDynamic)
	system.NewDistanceJointToWorld(body, matrix.Vec3Zero(), matrix.Vec3Zero()).SetRestLength(2)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 240 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	distance := body.Transform.WorldPosition().Distance(matrix.Vec3Zero())
	if distance > 2.05 {
		t.Fatalf("expected gravity-loaded distance joint to stay within tolerance, got %f", distance)
	}
}

func addJointBody(system *System, position matrix.Vec3, bodyType RigidBodyType) *RigidBody {
	body := addSystemSphere(system, position, bodyType)
	body.Collision.Mask = 0
	body.Simulation.SleepThreshold = 10000
	return body
}

func jointBodyDistance(a, b *RigidBody) matrix.Float {
	return a.Transform.WorldPosition().Distance(b.Transform.WorldPosition())
}

func finiteVec3(v matrix.Vec3) bool {
	return !matrix.IsNaN(v.X()) && !matrix.IsNaN(v.Y()) && !matrix.IsNaN(v.Z()) &&
		!matrix.IsInf(v.X(), 0) && !matrix.IsInf(v.Y(), 0) && !matrix.IsInf(v.Z(), 0)
}

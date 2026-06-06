/******************************************************************************/
/* joint_composition_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

const jointCompositionStep = 1.0 / 60.0

func TestDistanceJointChainRemainsStable(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintVelocityIterations = 12
	system.ConstraintPositionIterations = 12
	bodies := []*RigidBody{
		addJointBody(&system, matrix.Vec3{-3, 0, 0}, RigidBodyTypeStatic),
		addJointBody(&system, matrix.Vec3{-2, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{-1, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{0, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{1, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{3, 0, 0}, RigidBodyTypeStatic),
	}
	bodies[2].MotionState.LinearVelocity = matrix.Vec3Up().Scale(2)
	bodies[4].MotionState.LinearVelocity = matrix.Vec3Down().Scale(2)
	joints := make([]*DistanceJoint, 0, len(bodies)-1)
	for i := 0; i < len(bodies)-1; i++ {
		joints = append(joints, system.NewDistanceJoint(bodies[i], bodies[i+1], matrix.Vec3Zero(), matrix.Vec3Zero()))
	}
	stepJointComposition(t, &system, 600)
	assertFiniteBodies(t, bodies)
	for i, joint := range joints {
		if distance := joint.CurrentLength(); matrix.Abs(distance-1) > 0.06 {
			t.Fatalf("expected distance joint %d to stay near rest length 1, got %f", i, distance)
		}
	}
	if bodies[3].Transform.WorldPosition().Length() > 0.25 {
		t.Fatalf("expected chain center to remain bounded near its starting point, got %v",
			bodies[3].Transform.WorldPosition())
	}
}

func TestRopeChainDoesNotExceedSegmentLengths(t *testing.T) {
	system := System{}
	system.Initialize()
	system.ConstraintVelocityIterations = 12
	system.ConstraintPositionIterations = 12
	bodies := []*RigidBody{
		addJointBody(&system, matrix.Vec3{0, 0, 0}, RigidBodyTypeStatic),
		addJointBody(&system, matrix.Vec3{0, -1, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{0, -2, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{0, -3, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{0, -4, 0}, RigidBodyTypeDynamic),
	}
	bodies[len(bodies)-1].MotionState.LinearVelocity = matrix.Vec3Right().Scale(3)
	ropes := make([]*RopeJoint, 0, len(bodies)-1)
	for i := 0; i < len(bodies)-1; i++ {
		rope := system.NewRopeJoint(bodies[i], bodies[i+1], matrix.Vec3Zero(), matrix.Vec3Zero())
		rope.SetMaxLength(1)
		ropes = append(ropes, rope)
	}
	stepJointComposition(t, &system, 720)
	assertFiniteBodies(t, bodies)
	for i, rope := range ropes {
		if distance := rope.CurrentLength(); distance > 1.08 {
			t.Fatalf("expected rope segment %d to stay within max length 1, got %f", i, distance)
		}
	}
	if bodies[len(bodies)-1].Transform.WorldPosition().Y() >= -3.5 {
		t.Fatalf("expected rope chain to remain gravity loaded, got tail position %v",
			bodies[len(bodies)-1].Transform.WorldPosition())
	}
}

func TestBridgeLinksStayConnectedUnderGravity(t *testing.T) {
	system := System{}
	system.Initialize()
	system.ConstraintVelocityIterations = 14
	system.ConstraintPositionIterations = 14
	bodies := []*RigidBody{
		addJointBody(&system, matrix.Vec3{-3, 0, 0}, RigidBodyTypeStatic),
		addJointBody(&system, matrix.Vec3{-2, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{-1, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{0, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{1, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic),
		addJointBody(&system, matrix.Vec3{3, 0, 0}, RigidBodyTypeStatic),
	}
	joints := make([]*DistanceJoint, 0, len(bodies)-1)
	for i := 0; i < len(bodies)-1; i++ {
		joint := system.NewDistanceJoint(bodies[i], bodies[i+1], matrix.Vec3Zero(), matrix.Vec3Zero())
		joint.SetRestLength(1)
		joints = append(joints, joint)
	}
	stepJointComposition(t, &system, 900)
	assertFiniteBodies(t, bodies)
	for i, joint := range joints {
		if distance := joint.CurrentLength(); distance > 1.12 {
			t.Fatalf("expected bridge link %d to stay connected near length 1, got %f", i, distance)
		}
	}
	if bodies[3].Transform.WorldPosition().Y() > -0.1 {
		t.Fatalf("expected bridge span to sag under gravity, got center position %v",
			bodies[3].Transform.WorldPosition())
	}
}

func TestHingePendulumClockArmStaysAnchored(t *testing.T) {
	system := System{}
	system.Initialize()
	system.ConstraintVelocityIterations = 12
	system.ConstraintPositionIterations = 12
	anchor := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeStatic)
	arm := addJointBody(&system, matrix.Vec3{0, -2, 0}, RigidBodyTypeDynamic)
	arm.MotionState.LinearVelocity = matrix.Vec3Right().Scale(2)
	joint := system.NewHingeJoint(
		anchor,
		arm,
		matrix.Vec3Zero(),
		matrix.Vec3{0, 2, 0},
		matrix.Vec3Backward(),
		matrix.Vec3Backward(),
	)
	stepJointComposition(t, &system, 720)
	assertFiniteBodies(t, []*RigidBody{anchor, arm})
	if distance := joint.WorldAnchorA().Distance(joint.WorldAnchorB()); distance > 0.06 {
		t.Fatalf("expected pendulum hinge anchors to stay connected, got %f at %v and %v",
			distance, joint.WorldAnchorA(), joint.WorldAnchorB())
	}
	if arm.Transform.WorldPosition().Distance(anchor.Transform.WorldPosition()) > 2.12 {
		t.Fatalf("expected clock arm length to remain bounded, got arm position %v",
			arm.Transform.WorldPosition())
	}
	if matrix.Abs(arm.Transform.WorldPosition().X()) < 0.05 {
		t.Fatalf("expected pendulum arm to swing from initial impulse, got position %v",
			arm.Transform.WorldPosition())
	}
}

func stepJointComposition(t *testing.T, system *System, steps int) {
	t.Helper()
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range steps {
		system.Step(workGroup, threads, jointCompositionStep)
	}
}

func assertFiniteBodies(t *testing.T, bodies []*RigidBody) {
	t.Helper()
	for i, body := range bodies {
		position := body.Transform.WorldPosition()
		if !finiteVec3(position) {
			t.Fatalf("expected body %d to have a finite position, got %v", i, position)
		}
	}
}

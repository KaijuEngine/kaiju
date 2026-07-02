/******************************************************************************/
/* hinge_joint_test.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"math"
	"testing"

	"kaijuengine.com/matrix"
)

func TestHingeJointMaintainsAnchor(t *testing.T) {
	system := System{}
	system.Initialize()
	system.ConstraintPositionIterations = 12
	staticAnchor := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeStatic)
	arm := addJointBody(&system, matrix.Vec3{0, -2, 0}, RigidBodyTypeDynamic)
	arm.MotionState.LinearVelocity = matrix.Vec3Right().Scale(4)
	joint := system.NewHingeJoint(
		staticAnchor,
		arm,
		matrix.Vec3Zero(),
		matrix.Vec3{0, 2, 0},
		matrix.Vec3Backward(),
		matrix.Vec3Backward(),
	)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 240 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	if !matrix.Vec3ApproxTo(staticAnchor.Transform.WorldPosition(), matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected static hinge anchor body to stay fixed, got %v",
			staticAnchor.Transform.WorldPosition())
	}
	distance := joint.WorldAnchorA().Distance(joint.WorldAnchorB())
	if distance > 0.04 {
		t.Fatalf("expected hinge anchors to remain connected, got distance %f at %v and %v",
			distance, joint.WorldAnchorA(), joint.WorldAnchorB())
	}
	if arm.Transform.WorldPosition().Distance(staticAnchor.Transform.WorldPosition()) < 0.5 {
		t.Fatalf("expected dynamic arm to swing around the anchor, got position %v",
			arm.Transform.WorldPosition())
	}
}

func TestHingeJointAllowsAxisRotation(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 12
	body := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.MotionState.AngularVelocity = matrix.Vec3Up().Scale(matrix.Float(math.Pi))
	joint := system.NewHingeJointToWorld(
		body,
		matrix.Vec3Zero(),
		matrix.Vec3Zero(),
		matrix.Vec3Up(),
		matrix.Vec3Up(),
	)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 60 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	if joint.WorldAnchorA().Distance(joint.WorldAnchorB()) > 0.001 {
		t.Fatalf("expected centered hinge anchor to stay fixed, got %v and %v",
			joint.WorldAnchorA(), joint.WorldAnchorB())
	}
	if joint.WorldAxisA().Dot(joint.WorldAxisB()) < 0.999 {
		t.Fatalf("expected hinge axes to remain aligned, got %v and %v",
			joint.WorldAxisA(), joint.WorldAxisB())
	}
	if matrix.Abs(body.MotionState.AngularVelocity.Dot(matrix.Vec3Up())-matrix.Float(math.Pi)) > 0.001 {
		t.Fatalf("expected hinge to allow angular velocity around axis, got %v",
			body.MotionState.AngularVelocity)
	}
	if forbiddenAngularSpeed(body.MotionState.AngularVelocity, matrix.Vec3Up()) > 0.001 {
		t.Fatalf("expected no off-axis angular velocity, got %v",
			body.MotionState.AngularVelocity)
	}
	if matrix.Vec3ApproxTo(body.Transform.Rotation(), matrix.Vec3Zero(), 0.001) {
		t.Fatalf("expected body to rotate around hinge axis, got %v", body.Transform.Rotation())
	}
}

func TestHingeJointRestrictsOffAxisRotation(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 12
	body := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.MotionState.AngularVelocity = matrix.Vec3Right().Scale(matrix.Float(math.Pi))
	joint := system.NewHingeJointToWorld(
		body,
		matrix.Vec3Zero(),
		matrix.Vec3Zero(),
		matrix.Vec3Up(),
		matrix.Vec3Up(),
	)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 60 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	if forbiddenAngularSpeed(body.MotionState.AngularVelocity, matrix.Vec3Up()) > 0.02 {
		t.Fatalf("expected hinge to remove off-axis angular velocity, got %v",
			body.MotionState.AngularVelocity)
	}
	if joint.WorldAxisA().Dot(joint.WorldAxisB()) < 0.999 {
		t.Fatalf("expected hinge to correct off-axis rotation, got axes %v and %v",
			joint.WorldAxisA(), joint.WorldAxisB())
	}
	if joint.WorldAnchorA().Distance(joint.WorldAnchorB()) > 0.001 {
		t.Fatalf("expected hinge anchor to remain fixed while restricting rotation, got %v and %v",
			joint.WorldAnchorA(), joint.WorldAnchorB())
	}
}

func TestHingeJointAngularLimits(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintVelocityIterations = 12
	system.ConstraintPositionIterations = 12
	body := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	joint := system.NewHingeJointToWorld(
		body,
		matrix.Vec3Zero(),
		matrix.Vec3Zero(),
		matrix.Vec3Up(),
		matrix.Vec3Up(),
	)
	joint.SetAngularLimits(-matrix.Float(math.Pi/12), matrix.Float(math.Pi/12))
	body.Transform.SetRotation(matrix.Vec3{0, 60, 0})
	body.MotionState.AngularVelocity = matrix.Vec3Up().Scale(4)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	var observedLimitImpulse matrix.Float
	for range 120 {
		system.Step(workGroup, threads, 1.0/60.0)
		observedLimitImpulse = max(observedLimitImpulse, matrix.Abs(joint.AccumulatedLimitImpulse))
	}
	angle := joint.CurrentAngle()
	if angle > joint.MaxAngle+0.03 {
		t.Fatalf("expected hinge angle to stop at max %f, got %f", joint.MaxAngle, angle)
	}
	if angle < joint.MinAngle-0.03 {
		t.Fatalf("expected hinge angle to stay above min %f, got %f", joint.MinAngle, angle)
	}
	if observedLimitImpulse <= 0 {
		t.Fatalf("expected limit row to accumulate impulse")
	}
}

func TestHingeJointMotorDrivesVelocity(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintVelocityIterations = 12
	system.ConstraintPositionIterations = 0
	body := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	joint := system.NewHingeJointToWorld(
		body,
		matrix.Vec3Zero(),
		matrix.Vec3Zero(),
		matrix.Vec3Up(),
		matrix.Vec3Up(),
	)
	joint.SetMotor(3, 100)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 1.0/60.0)
	speed := joint.CurrentAngularVelocity()
	if matrix.Abs(speed-3) > 0.05 {
		t.Fatalf("expected hinge motor to drive angular velocity near 3, got %f", speed)
	}
	if matrix.Abs(joint.AccumulatedMotorImpulse) <= 0 {
		t.Fatalf("expected motor row to accumulate impulse")
	}
}

func TestConstraintBreaksAboveImpulseThreshold(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintVelocityIterations = 8
	system.ConstraintPositionIterations = 0
	body := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.MotionState.LinearVelocity = matrix.Vec3Right().Scale(20)
	joint := system.NewPointJointToWorld(body, matrix.Vec3Zero(), matrix.Vec3Zero())
	joint.constraint.SetBreakForce(0.1)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 1.0/60.0)
	if !joint.constraint.Broken {
		t.Fatalf("expected constraint to be marked broken")
	}
	if joint.constraint.Enabled || joint.constraint.Active {
		t.Fatalf("expected broken constraint to disable itself")
	}
	if joint.constraint.AccumulatedLinearImpulse() <= joint.constraint.BreakForce {
		t.Fatalf("expected accumulated impulse to exceed break threshold, got %f <= %f",
			joint.constraint.AccumulatedLinearImpulse(), joint.constraint.BreakForce)
	}
}

func forbiddenAngularSpeed(angularVelocity, hingeAxis matrix.Vec3) matrix.Float {
	axis := safeNormal(hingeAxis, matrix.Vec3Right())
	allowed := axis.Scale(angularVelocity.Dot(axis))
	return angularVelocity.Subtract(allowed).Length()
}

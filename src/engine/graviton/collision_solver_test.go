/******************************************************************************/
/* collision_solver_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestCollisionSolverSeparatesDynamicFromStatic(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	dynamic := system.NewBody()
	dynamic.Active = true
	dynamic.Simulation.Type = RigidBodyTypeDynamic
	dynamic.SetMass(1, matrix.Vec3One())
	dynamic.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	dynamic.Collision.Mask = 1
	dynamic.MotionState.LinearVelocity = matrix.Vec3Right()
	static := system.NewBody()
	static.Active = true
	static.Simulation.Type = RigidBodyTypeStatic
	static.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	static.Collision.Mask = 1
	static.Transform.SetPosition(matrix.Vec3{1.5, 0, 0})
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	system.Step(&workGroup, &threads, 0)
	if dynamic.Transform.WorldPosition().X() >= 0 {
		t.Fatalf("expected dynamic body to be pushed away from static body, got %v",
			dynamic.Transform.WorldPosition())
	}
	if dynamic.MotionState.LinearVelocity.X() > 0 {
		t.Fatalf("expected inward velocity to be removed, got %v",
			dynamic.MotionState.LinearVelocity)
	}
	if !matrix.Vec3ApproxTo(static.Transform.WorldPosition(), matrix.Vec3{1.5, 0, 0}, 0.0001) {
		t.Fatalf("expected static body to remain fixed, got %v", static.Transform.WorldPosition())
	}
}

func TestCollisionSolverSplitsDynamicPairCorrection(t *testing.T) {
	a := testRigidBody(Shape{}, matrix.Vec3{0, 0, 0})
	a.MotionState.LinearVelocity = matrix.Vec3Right()
	b := testRigidBody(Shape{}, matrix.Vec3{1.5, 0, 0})
	b.MotionState.LinearVelocity = matrix.Vec3Left()
	a.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	b.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	manifold, ok := CollideBodies(a, b)
	if !ok {
		t.Fatal("expected collision")
	}
	solver := CollisionSolver{}
	solver.Initialize()
	solver.VelocityIterations = 1
	solver.PositionIterations = 1
	solver.Restitution = 0
	solver.Baumgarte = 1
	solver.PenetrationSlop = 0
	solver.MaxCorrection = 10
	solver.Solve([]ContactManifold{manifold}, nil)
	if matrix.Abs(a.Transform.WorldPosition().X()+0.25) > 0.0001 {
		t.Fatalf("expected body A to move half the penetration, got %v", a.Transform.WorldPosition())
	}
	if matrix.Abs(b.Transform.WorldPosition().X()-1.75) > 0.0001 {
		t.Fatalf("expected body B to move half the penetration, got %v", b.Transform.WorldPosition())
	}
	if matrix.Abs(a.MotionState.LinearVelocity.X()) > 0.0001 ||
		matrix.Abs(b.MotionState.LinearVelocity.X()) > 0.0001 {
		t.Fatalf("expected opposing velocities to cancel, got %v and %v",
			a.MotionState.LinearVelocity, b.MotionState.LinearVelocity)
	}
}

func TestCollisionSolverIgnoresTriggers(t *testing.T) {
	a := testRigidBody(Shape{}, matrix.Vec3{0, 0, 0})
	b := testRigidBody(Shape{}, matrix.Vec3{1.5, 0, 0})
	a.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	b.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	a.Collision.IsTrigger = true
	manifold, ok := CollideBodies(a, b)
	if !ok {
		t.Fatal("expected trigger contact to be reported")
	}
	solver := CollisionSolver{}
	solver.Solve([]ContactManifold{manifold}, nil)
	if !matrix.Vec3ApproxTo(a.Transform.WorldPosition(), matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected trigger body not to be moved, got %v", a.Transform.WorldPosition())
	}
	if !matrix.Vec3ApproxTo(b.Transform.WorldPosition(), matrix.Vec3{1.5, 0, 0}, 0.0001) {
		t.Fatalf("expected other trigger participant not to be moved, got %v", b.Transform.WorldPosition())
	}
}

func TestCollisionSolverParallelIndependentIslands(t *testing.T) {
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	manifolds := make([]ContactManifold, 0, 32)
	bodies := make([]*RigidBody, 0, 64)
	for i := range 32 {
		x := matrix.Float(i * 4)
		a := testRigidBody(Shape{}, matrix.Vec3{x, 0, 0})
		b := testRigidBody(Shape{}, matrix.Vec3{x + 1.5, 0, 0})
		a.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
		b.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
		manifold, ok := CollideBodies(a, b)
		if !ok {
			t.Fatal("expected generated pair to collide")
		}
		manifolds = append(manifolds, manifold)
		bodies = append(bodies, a, b)
	}
	solver := CollisionSolver{}
	solver.Initialize()
	solver.VelocityIterations = 1
	solver.PositionIterations = 1
	solver.Baumgarte = 1
	solver.PenetrationSlop = 0
	solver.MaxCorrection = 10
	solver.Solve(manifolds, &threads)
	for i := 0; i < len(bodies); i += 2 {
		a := bodies[i]
		b := bodies[i+1]
		if a.Transform.WorldPosition().Distance(b.Transform.WorldPosition()) <= 1.5 {
			t.Fatalf("expected independent island %d to separate, got %v and %v",
				i/2, a.Transform.WorldPosition(), b.Transform.WorldPosition())
		}
	}
}

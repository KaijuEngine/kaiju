/******************************************************************************/
/* constraint_island_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestConstraintSolverBuildsIndependentIslands(t *testing.T) {
	system := System{}
	system.Initialize()
	a := addSystemSphere(&system, matrix.Vec3{0, 0, 0}, RigidBodyTypeDynamic)
	b := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	c := addSystemSphere(&system, matrix.Vec3{8, 0, 0}, RigidBodyTypeDynamic)
	d := addSystemSphere(&system, matrix.Vec3{10, 0, 0}, RigidBodyTypeDynamic)
	constraints := []*Constraint{
		system.NewConstraint(ConstraintTypeGeneric, a, b),
		system.NewConstraint(ConstraintTypeGeneric, c, d),
	}
	solver := CollisionSolver{}
	solver.Initialize()
	solver.buildIslands(nil, constraints)
	if len(solver.islands) != 2 {
		t.Fatalf("expected 2 independent constraint islands, got %d", len(solver.islands))
	}
	for i := range solver.islands {
		if len(solver.islands[i].manifolds) != 0 {
			t.Fatalf("expected island %d to have no contact manifolds, got %d",
				i, len(solver.islands[i].manifolds))
		}
		if len(solver.islands[i].constraints) != 1 {
			t.Fatalf("expected island %d to have 1 constraint, got %d",
				i, len(solver.islands[i].constraints))
		}
	}
}

func TestConstraintMergesContactIslands(t *testing.T) {
	system := System{}
	system.Initialize()
	a := addSystemSphere(&system, matrix.Vec3{0, 0, 0}, RigidBodyTypeDynamic)
	b := addSystemSphere(&system, matrix.Vec3{1.5, 0, 0}, RigidBodyTypeDynamic)
	c := addSystemSphere(&system, matrix.Vec3{5, 0, 0}, RigidBodyTypeDynamic)
	d := addSystemSphere(&system, matrix.Vec3{6.5, 0, 0}, RigidBodyTypeDynamic)
	ab, ok := CollideBodies(a, b)
	if !ok {
		t.Fatal("expected first contact island to collide")
	}
	cd, ok := CollideBodies(c, d)
	if !ok {
		t.Fatal("expected second contact island to collide")
	}
	constraints := []*Constraint{
		system.NewConstraint(ConstraintTypeGeneric, b, c),
	}
	solver := CollisionSolver{}
	solver.Initialize()
	solver.buildIslands([]ContactManifold{ab, cd}, constraints)
	if len(solver.islands) != 1 {
		t.Fatalf("expected constraint to merge contact islands, got %d islands", len(solver.islands))
	}
	island := solver.islands[0]
	if len(island.manifolds) != 2 {
		t.Fatalf("expected merged island to contain 2 manifolds, got %d", len(island.manifolds))
	}
	if len(island.constraints) != 1 {
		t.Fatalf("expected merged island to contain 1 constraint, got %d", len(island.constraints))
	}
}

func TestConstraintParallelIndependentIslands(t *testing.T) {
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	system := System{}
	system.Initialize()
	constraints := make([]*Constraint, 0, 32)
	bodies := make([]*RigidBody, 0, 64)
	for i := range 32 {
		x := matrix.Float(i * 4)
		a := addSystemSphere(&system, matrix.Vec3{x, 0, 0}, RigidBodyTypeDynamic)
		b := addSystemSphere(&system, matrix.Vec3{x + 2, 0, 0}, RigidBodyTypeDynamic)
		a.MotionState.LinearVelocity = matrix.Vec3Right()
		b.MotionState.LinearVelocity = matrix.Vec3Left()
		constraint := system.NewConstraint(ConstraintTypeGeneric, a, b)
		constraint.Rows = append(constraint.Rows, NewConstraintSolverRow(
			a,
			b,
			a.Transform.WorldPosition(),
			b.Transform.WorldPosition(),
			matrix.Vec3Right(),
		))
		constraints = append(constraints, constraint)
		bodies = append(bodies, a, b)
	}
	solver := CollisionSolver{}
	solver.Initialize()
	solver.VelocityIterations = 1
	solver.PositionIterations = 0
	solver.SolveWithConstraints(nil, constraints, &threads)
	if len(solver.islands) != len(constraints) {
		t.Fatalf("expected %d independent constraint islands, got %d",
			len(constraints), len(solver.islands))
	}
	for i := 0; i < len(bodies); i += 2 {
		a := bodies[i]
		b := bodies[i+1]
		if !matrix.Vec3ApproxTo(a.MotionState.LinearVelocity, matrix.Vec3Zero(), 0.0001) ||
			!matrix.Vec3ApproxTo(b.MotionState.LinearVelocity, matrix.Vec3Zero(), 0.0001) {
			t.Fatalf("expected independent constraint island %d to solve, got %v and %v",
				i/2, a.MotionState.LinearVelocity, b.MotionState.LinearVelocity)
		}
	}
}

/******************************************************************************/
/* sweep_and_prune_test.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestSweepPruneReturnsOverlappingBodyPair(t *testing.T) {
	var bodies pooling.PoolGroup[RigidBody]
	a := addSweepBody(&bodies, matrix.Vec3{0, 0, 0})
	b := addSweepBody(&bodies, matrix.Vec3{1.5, 0, 0})
	addSweepBody(&bodies, matrix.Vec3{4, 0, 0})
	var sap SweepPrune
	sap.Initialize(8)
	sap.Rebuild(&bodies)
	pairs := sap.Sweep()
	if len(pairs) != 1 {
		t.Fatalf("expected 1 potential pair, got %d", len(pairs))
	}
	if !samePair(pairs[0], a, b) {
		t.Fatalf("expected pair (%p, %p), got (%p, %p)",
			a, b, pairs[0].BodyA, pairs[0].BodyB)
	}
}

func TestSweepPruneSupportsMultiplePools(t *testing.T) {
	var bodies pooling.PoolGroup[RigidBody]
	var first, crossPool *RigidBody
	for i := range 300 {
		body := addSweepBody(&bodies, matrix.Vec3{matrix.Float(i * 4), 0, 0})
		if i == 0 {
			first = body
		}
		if i == 260 {
			crossPool = body
			body.Transform.SetPosition(matrix.Vec3{0.5, 0, 0})
		}
	}
	var sap SweepPrune
	sap.Initialize(300)
	sap.Rebuild(&bodies)
	pairs := sap.Sweep()
	if len(pairs) != 1 {
		t.Fatalf("expected 1 potential pair across pool boundary, got %d", len(pairs))
	}
	if !samePair(pairs[0], first, crossPool) {
		t.Fatalf("expected cross-pool pair (%p, %p), got (%p, %p)",
			first, crossPool, pairs[0].BodyA, pairs[0].BodyB)
	}
}

func TestSweepPruneParallelMatchesSequential(t *testing.T) {
	var bodies pooling.PoolGroup[RigidBody]
	for x := range 24 {
		for y := range 12 {
			addSweepBody(&bodies, matrix.Vec3{
				matrix.Float(x) * 0.75,
				matrix.Float(y) * 0.9,
				matrix.Float((x + y) % 3),
			})
		}
	}
	var sequential SweepPrune
	sequential.Initialize(128)
	sequential.Rebuild(&bodies)
	seqPairs := pairSet(sequential.Sweep())
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	var parallel SweepPrune
	parallel.Initialize(128)
	parallel.RebuildParallel(&bodies, &threads)
	parPairs := pairSet(parallel.SweepParallel(&threads, nil))
	if len(seqPairs) != len(parPairs) {
		t.Fatalf("expected %d parallel pairs, got %d", len(seqPairs), len(parPairs))
	}
	for pair := range seqPairs {
		if !parPairs[pair] {
			t.Fatalf("parallel sweep missed pair %v", pair)
		}
	}
}

func addSweepBody(bodies *pooling.PoolGroup[RigidBody], position matrix.Vec3) *RigidBody {
	body, poolId, elementId := bodies.Add()
	*body = RigidBody{}
	body.poolId = poolId
	body.id = elementId
	body.Active = true
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(position)
	body.Collision.Shape.SetAABB(matrix.Vec3Zero(), matrix.Vec3{1, 1, 1})
	body.Collision.Group = 0
	body.Collision.Mask = 1
	return body
}

func samePair(pair ActivePair, a, b *RigidBody) bool {
	return (pair.BodyA == a && pair.BodyB == b) ||
		(pair.BodyA == b && pair.BodyB == a)
}

func pairSet(pairs []ActivePair) map[[2]int]bool {
	set := make(map[[2]int]bool, len(pairs))
	for _, pair := range pairs {
		a := pair.BodyA.poolLocation()
		b := pair.BodyB.poolLocation()
		if a > b {
			a, b = b, a
		}
		set[[2]int{a, b}] = true
	}
	return set
}

/******************************************************************************/
/* collision_solver.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"sync"

	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

const (
	defaultVelocityIterations = 8
	defaultPositionIterations = 3
	defaultRestitution        = matrix.Float(0.05)
	defaultStaticFriction     = matrix.Float(0.6)
	defaultDynamicFriction    = matrix.Float(0.45)
	defaultBaumgarte          = matrix.Float(0.8)
	defaultPenetrationSlop    = matrix.Float(0.005)
	defaultMaxCorrection      = matrix.Float(0.25)
	solverMinIslandsPerJob    = 2
)

type collisionIsland struct {
	manifolds   []int
	constraints []int
}

// CollisionSolver resolves narrow-phase contacts with an iterative impulse
// solver. Contacts are grouped into independent dynamic islands so islands can
// be solved in parallel without concurrent writes to the same body.
type CollisionSolver struct {
	VelocityIterations  int
	PositionIterations  int
	DeltaTime           matrix.Float
	Restitution         matrix.Float
	StaticFriction      matrix.Float
	DynamicFriction     matrix.Float
	Baumgarte           matrix.Float
	PenetrationSlop     matrix.Float
	MaxCorrection       matrix.Float
	islands             []collisionIsland
	writableBodies      []*RigidBody
	parents             []int
	ranks               []uint8
	bodyIndex           map[*RigidBody]int
	rootToIsland        map[int]int
	eligibleContacts    []int
	eligibleConstraints []int
	initialized         bool
}

func (s *CollisionSolver) Initialize() {
	s.VelocityIterations = defaultVelocityIterations
	s.PositionIterations = defaultPositionIterations
	s.DeltaTime = defaultDistanceJointTimeStep
	s.Restitution = defaultRestitution
	s.StaticFriction = defaultStaticFriction
	s.DynamicFriction = defaultDynamicFriction
	s.Baumgarte = defaultBaumgarte
	s.PenetrationSlop = defaultPenetrationSlop
	s.MaxCorrection = defaultMaxCorrection
	s.bodyIndex = make(map[*RigidBody]int, 256)
	s.rootToIsland = make(map[int]int, 64)
	s.initialized = true
}

func (s *CollisionSolver) Reset() {
	for key := range s.bodyIndex {
		delete(s.bodyIndex, key)
	}
	for key := range s.rootToIsland {
		delete(s.rootToIsland, key)
	}
	for i := range s.islands {
		s.islands[i].manifolds = s.islands[i].manifolds[:0]
		s.islands[i].constraints = s.islands[i].constraints[:0]
	}
	s.islands = s.islands[:0]
	s.writableBodies = s.writableBodies[:0]
	s.parents = s.parents[:0]
	s.ranks = s.ranks[:0]
	s.eligibleContacts = s.eligibleContacts[:0]
	s.eligibleConstraints = s.eligibleConstraints[:0]
}

func (s *CollisionSolver) Solve(manifolds []ContactManifold, threads *concurrent.Threads) {
	s.SolveWithConstraints(manifolds, nil, threads)
}

func (s *CollisionSolver) SolveWithConstraints(manifolds []ContactManifold, constraints []*Constraint, threads *concurrent.Threads) {
	if len(manifolds) == 0 && len(constraints) == 0 {
		return
	}
	s.ensureInitialized()
	s.buildIslands(manifolds, constraints)
	if len(s.islands) == 0 {
		return
	}
	workers := broadPhaseWorkerCount(threads, len(s.islands), solverMinIslandsPerJob)
	if workers == 1 {
		s.solveIslandRange(manifolds, constraints, 0, len(s.islands))
		return
	}
	runSolverJobs(threads, workers, len(s.islands), func(start, end, _ int) {
		s.solveIslandRange(manifolds, constraints, start, end)
	})
}

func (s *CollisionSolver) ensureInitialized() {
	if !s.initialized {
		s.Initialize()
	}
}

func (s *CollisionSolver) buildIslands(manifolds []ContactManifold, constraints []*Constraint) {
	for key := range s.bodyIndex {
		delete(s.bodyIndex, key)
	}
	for key := range s.rootToIsland {
		delete(s.rootToIsland, key)
	}
	for i := range s.islands {
		s.islands[i].manifolds = s.islands[i].manifolds[:0]
		s.islands[i].constraints = s.islands[i].constraints[:0]
	}
	s.islands = s.islands[:0]
	s.writableBodies = s.writableBodies[:0]
	s.parents = s.parents[:0]
	s.ranks = s.ranks[:0]
	s.eligibleContacts = s.eligibleContacts[:0]
	s.eligibleConstraints = s.eligibleConstraints[:0]
	for i := range manifolds {
		manifold := &manifolds[i]
		if !s.shouldResolve(manifold) {
			continue
		}
		aIndex, aWritable := s.addWritableBody(manifold.BodyA)
		bIndex, bWritable := s.addWritableBody(manifold.BodyB)
		if aWritable && bWritable {
			s.union(aIndex, bIndex)
		}
		s.eligibleContacts = append(s.eligibleContacts, i)
	}
	for i, constraint := range constraints {
		if !s.shouldSolveConstraint(constraint) {
			continue
		}
		aIndex, aWritable := s.addWritableBody(constraint.BodyA)
		bIndex, bWritable := s.addWritableBody(constraint.BodyB)
		if aWritable && bWritable {
			s.union(aIndex, bIndex)
		}
		s.eligibleConstraints = append(s.eligibleConstraints, i)
	}
	for _, manifoldIndex := range s.eligibleContacts {
		manifold := &manifolds[manifoldIndex]
		root := s.manifoldRoot(manifold)
		if root < 0 {
			continue
		}
		islandIndex, ok := s.rootToIsland[root]
		if !ok {
			islandIndex = len(s.islands)
			s.rootToIsland[root] = islandIndex
			s.addIsland()
		}
		s.islands[islandIndex].manifolds = append(s.islands[islandIndex].manifolds, manifoldIndex)
	}
	for _, constraintIndex := range s.eligibleConstraints {
		constraint := constraints[constraintIndex]
		root := s.constraintRoot(constraint)
		if root < 0 {
			continue
		}
		islandIndex, ok := s.rootToIsland[root]
		if !ok {
			islandIndex = len(s.islands)
			s.rootToIsland[root] = islandIndex
			s.addIsland()
		}
		s.islands[islandIndex].constraints = append(s.islands[islandIndex].constraints, constraintIndex)
	}
}

func (s *CollisionSolver) addIsland() {
	if len(s.islands) < cap(s.islands) {
		s.islands = s.islands[:len(s.islands)+1]
		s.islands[len(s.islands)-1].manifolds = s.islands[len(s.islands)-1].manifolds[:0]
		s.islands[len(s.islands)-1].constraints = s.islands[len(s.islands)-1].constraints[:0]
		return
	}
	s.islands = append(s.islands, collisionIsland{})
}

func (s *CollisionSolver) shouldResolve(manifold *ContactManifold) bool {
	if manifold == nil || manifold.Count == 0 || manifold.BodyA == nil || manifold.BodyB == nil {
		return false
	}
	if manifold.BodyA.Collision.IsTrigger || manifold.BodyB.Collision.IsTrigger {
		return false
	}
	return manifold.BodyA.inverseMass()+manifold.BodyB.inverseMass() > 0
}

func (s *CollisionSolver) shouldSolveConstraint(constraint *Constraint) bool {
	if constraint == nil || !constraint.Active || !constraint.Enabled {
		return false
	}
	if constraint.BodyA == nil && constraint.BodyB == nil {
		return false
	}
	if constraint.BodyA != nil && !constraint.BodyA.Active {
		return false
	}
	if constraint.BodyB != nil && !constraint.BodyB.Active {
		return false
	}
	return solverBodyWritable(constraint.BodyA) || solverBodyWritable(constraint.BodyB)
}

func (s *CollisionSolver) addWritableBody(body *RigidBody) (int, bool) {
	if !solverBodyWritable(body) {
		return -1, false
	}
	if index, ok := s.bodyIndex[body]; ok {
		return index, true
	}
	index := len(s.writableBodies)
	s.bodyIndex[body] = index
	s.writableBodies = append(s.writableBodies, body)
	s.parents = append(s.parents, index)
	s.ranks = append(s.ranks, 0)
	return index, true
}

func solverBodyWritable(body *RigidBody) bool {
	return body != nil && (body.inverseMass() > 0 || !body.inverseInertia().IsZero())
}

func (s *CollisionSolver) manifoldRoot(manifold *ContactManifold) int {
	if index, ok := s.bodyIndex[manifold.BodyA]; ok {
		return s.find(index)
	}
	if index, ok := s.bodyIndex[manifold.BodyB]; ok {
		return s.find(index)
	}
	return -1
}

func (s *CollisionSolver) constraintRoot(constraint *Constraint) int {
	if index, ok := s.bodyIndex[constraint.BodyA]; ok {
		return s.find(index)
	}
	if index, ok := s.bodyIndex[constraint.BodyB]; ok {
		return s.find(index)
	}
	return -1
}

func (s *CollisionSolver) find(index int) int {
	parent := s.parents[index]
	if parent != index {
		parent = s.find(parent)
		s.parents[index] = parent
	}
	return parent
}

func (s *CollisionSolver) union(a, b int) {
	rootA := s.find(a)
	rootB := s.find(b)
	if rootA == rootB {
		return
	}
	if s.ranks[rootA] < s.ranks[rootB] {
		rootA, rootB = rootB, rootA
	}
	s.parents[rootB] = rootA
	if s.ranks[rootA] == s.ranks[rootB] {
		s.ranks[rootA]++
	}
}

func (s *CollisionSolver) solveIslandRange(manifolds []ContactManifold, constraints []*Constraint, start, end int) {
	for islandIndex := start; islandIndex < end; islandIndex++ {
		island := &s.islands[islandIndex]
		for _, constraintIndex := range island.constraints {
			s.prepareConstraint(constraints[constraintIndex])
		}
		for range s.VelocityIterations {
			for _, manifoldIndex := range island.manifolds {
				s.solveVelocity(&manifolds[manifoldIndex])
			}
			for _, constraintIndex := range island.constraints {
				s.solveConstraint(constraints[constraintIndex])
			}
		}
		for range s.PositionIterations {
			for _, manifoldIndex := range island.manifolds {
				s.solvePosition(&manifolds[manifoldIndex])
			}
			for _, constraintIndex := range island.constraints {
				s.solveConstraintPosition(constraints[constraintIndex])
			}
		}
	}
}

func (s *CollisionSolver) prepareConstraint(constraint *Constraint) {
	if constraint == nil {
		return
	}
	if constraint.Type == ConstraintTypeDistance && constraint.Distance != nil {
		constraint.Distance.prepare(s.DeltaTime)
		return
	}
	if constraint.Type == ConstraintTypeRope && constraint.Rope != nil {
		constraint.Rope.prepare(s.DeltaTime)
		return
	}
	if constraint.Type == ConstraintTypePoint && constraint.Point != nil {
		constraint.Point.prepare(s.DeltaTime)
		return
	}
	if constraint.Type == ConstraintTypeHinge && constraint.Hinge != nil {
		constraint.Hinge.prepare(s.DeltaTime)
	}
}

func (s *CollisionSolver) solveConstraint(constraint *Constraint) {
	if constraint == nil || !constraint.Active || !constraint.Enabled {
		return
	}
	if constraint.Type == ConstraintTypeDistance && constraint.Distance != nil {
		constraint.Distance.solveVelocity()
		constraint.BreakIfNeeded()
		return
	}
	if constraint.Type == ConstraintTypeRope && constraint.Rope != nil {
		constraint.Rope.solveVelocity()
		constraint.BreakIfNeeded()
		return
	}
	if constraint.Type == ConstraintTypePoint && constraint.Point != nil {
		constraint.Point.solveVelocity()
		constraint.BreakIfNeeded()
		return
	}
	if constraint.Type == ConstraintTypeHinge && constraint.Hinge != nil {
		constraint.Hinge.solveVelocity()
		constraint.BreakIfNeeded()
		return
	}
	for i := range constraint.Rows {
		constraint.Rows[i].Solve()
	}
	constraint.BreakIfNeeded()
}

func (s *CollisionSolver) solveConstraintPosition(constraint *Constraint) {
	if constraint == nil || !constraint.Active || !constraint.Enabled {
		return
	}
	if constraint.Type == ConstraintTypeDistance && constraint.Distance != nil {
		constraint.Distance.solvePosition()
		return
	}
	if constraint.Type == ConstraintTypeRope && constraint.Rope != nil {
		constraint.Rope.solvePosition()
		return
	}
	if constraint.Type == ConstraintTypePoint && constraint.Point != nil {
		constraint.Point.solvePosition()
		return
	}
	if constraint.Type == ConstraintTypeHinge && constraint.Hinge != nil {
		constraint.Hinge.solvePosition()
	}
}

func (s *CollisionSolver) solveVelocity(manifold *ContactManifold) {
	bodyA := manifold.BodyA
	bodyB := manifold.BodyB
	normal := safeNormal(manifold.Normal, matrix.Vec3Right())
	for i := range manifold.Count {
		contact := manifold.Contacts[i]
		ra := contact.Point.Subtract(bodyA.Transform.WorldPosition())
		rb := contact.Point.Subtract(bodyB.Transform.WorldPosition())
		relativeVelocity := velocityAtContact(bodyB, rb).Subtract(velocityAtContact(bodyA, ra))
		normalVelocity := relativeVelocity.Dot(normal)
		if normalVelocity > 0 {
			continue
		}
		denominator := impulseDenominator(bodyA, bodyB, ra, rb, normal)
		if denominator <= contactEpsilon {
			continue
		}
		normalImpulseMagnitude := -(1 + s.Restitution) * normalVelocity / denominator
		normalImpulseMagnitude /= matrix.Float(manifold.Count)
		normalImpulse := normal.Scale(normalImpulseMagnitude)
		applyImpulse(bodyA, normalImpulse.Negative(), ra)
		applyImpulse(bodyB, normalImpulse, rb)

		relativeVelocity = velocityAtContact(bodyB, rb).Subtract(velocityAtContact(bodyA, ra))
		tangent := relativeVelocity.Subtract(normal.Scale(relativeVelocity.Dot(normal)))
		if tangent.LengthSquared() <= contactEpsilon*contactEpsilon {
			continue
		}
		tangent = tangent.Normal()
		tangentDenominator := impulseDenominator(bodyA, bodyB, ra, rb, tangent)
		if tangentDenominator <= contactEpsilon {
			continue
		}
		tangentImpulseMagnitude := -relativeVelocity.Dot(tangent) / tangentDenominator
		tangentImpulseMagnitude /= matrix.Float(manifold.Count)
		maxStaticFriction := normalImpulseMagnitude * s.StaticFriction
		var tangentImpulse matrix.Vec3
		if matrix.Abs(tangentImpulseMagnitude) <= maxStaticFriction {
			tangentImpulse = tangent.Scale(tangentImpulseMagnitude)
		} else {
			dynamicMagnitude := klib.Clamp(tangentImpulseMagnitude,
				-normalImpulseMagnitude*s.DynamicFriction,
				normalImpulseMagnitude*s.DynamicFriction)
			tangentImpulse = tangent.Scale(dynamicMagnitude)
		}
		applyImpulse(bodyA, tangentImpulse.Negative(), ra)
		applyImpulse(bodyB, tangentImpulse, rb)
	}
}

func (s *CollisionSolver) solvePosition(manifold *ContactManifold) {
	bodyA := manifold.BodyA
	bodyB := manifold.BodyB
	current, ok := CollideBodies(bodyA, bodyB)
	if !ok {
		return
	}
	manifold = &current
	invMassA := bodyA.inverseMass()
	invMassB := bodyB.inverseMass()
	invMassSum := invMassA + invMassB
	if invMassSum <= contactEpsilon {
		return
	}
	normal := safeNormal(manifold.Normal, matrix.Vec3Right())
	for i := range manifold.Count {
		penetration := manifold.Contacts[i].Penetration
		depth := max(penetration-s.PenetrationSlop, 0)
		if depth <= 0 {
			continue
		}
		correctionMagnitude := depth * s.Baumgarte / invMassSum
		correctionMagnitude = min(correctionMagnitude, s.MaxCorrection)
		correctionMagnitude /= matrix.Float(manifold.Count)
		correction := normal.Scale(correctionMagnitude)
		moveBody(bodyA, correction.Scale(-invMassA))
		moveBody(bodyB, correction.Scale(invMassB))
	}
}

func velocityAtContact(body *RigidBody, r matrix.Vec3) matrix.Vec3 {
	return VelocityAtAnchor(body, r)
}

func impulseDenominator(bodyA, bodyB *RigidBody, ra, rb, axis matrix.Vec3) matrix.Float {
	return ConstraintImpulseDenominator(bodyA, bodyB, ra, rb, axis)
}

func angularImpulseDenominator(body *RigidBody, r, axis matrix.Vec3) matrix.Float {
	return AngularEffectiveMass(body, r, axis)
}

func applyImpulse(body *RigidBody, impulse, r matrix.Vec3) {
	if body != nil {
		body.applyImpulse(impulse, r)
	}
}

func moveBody(body *RigidBody, correction matrix.Vec3) {
	if body == nil || body.inverseMass() == 0 || correction.LengthSquared() <= contactEpsilon*contactEpsilon {
		return
	}
	body.Transform.AddPosition(correction)
}

func runSolverJobs(threads *concurrent.Threads, workers, items int, work func(start, end, worker int)) {
	jobs := make([]func(int), workers)
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for worker := range workers {
		start := worker * items / workers
		end := (worker + 1) * items / workers
		worker := worker
		jobs[worker] = func(int) {
			defer wg.Done()
			work(start, end, worker)
		}
	}
	threads.AddWork(jobs)
	wg.Wait()
}

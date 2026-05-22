/******************************************************************************/
/* physics_system.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"log/slog"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
	"kaijuengine.com/platform/profiler/tracing"
)

type StagePhysicsEntry struct {
	Entity *Entity
	Body   *graviton.RigidBody
}

type stagePhysicsConstraintEntry struct {
	EntityA    *Entity
	EntityB    *Entity
	Constraint *graviton.Constraint
	remove     func()
}

type StagePhysics struct {
	world              graviton.System
	entities           []StagePhysicsEntry
	constraints        []stagePhysicsConstraintEntry
	accumulatedTime    float64
	fixedTimeStep      float64
	maxAccumulatedTime float64
	customMaxAccumTime bool
	maxSubSteps        int
	active             bool
}

const (
	defaultPhysicsFixedTimeStep = 1.0 / 60.0
	defaultPhysicsMaxSubSteps   = 5
)

func (pe *StagePhysicsEntry) syncEntityToBody() {
	t := &pe.Entity.Transform
	b := pe.Body
	b.Transform.SetPosition(t.WorldPosition())
	b.Transform.SetRotation(t.WorldRotation())
	b.Transform.SetScale(t.WorldScale())
}

func (pe *StagePhysicsEntry) syncBodyToEntity() {
	b := pe.Body
	t := &pe.Entity.Transform
	t.SetWorldPosition(b.Position())
	t.SetWorldRotation(b.Rotation().ToEuler())
	t.SetWorldScale(b.Transform.WorldScale())
}

func (p *StagePhysics) IsActive() bool          { return p.active }
func (p *StagePhysics) World() *graviton.System { return &p.world }

func (p *StagePhysics) FixedTimeStep() float64 {
	p.ensureStepConfig()
	return p.fixedTimeStep
}

func (p *StagePhysics) SetFixedTimeStep(step float64) {
	if step <= 0 {
		slog.Error("stage physics fixed time step must be greater than zero")
		return
	}
	p.fixedTimeStep = step
	p.ensureMaxAccumulatedTime()
}

func (p *StagePhysics) MaxSubSteps() int {
	p.ensureStepConfig()
	return p.maxSubSteps
}

func (p *StagePhysics) SetMaxSubSteps(maxSubSteps int) {
	if maxSubSteps < 1 {
		slog.Error("stage physics max substeps must be at least one")
		return
	}
	p.maxSubSteps = maxSubSteps
	p.ensureMaxAccumulatedTime()
}

func (p *StagePhysics) MaxAccumulatedTime() float64 {
	p.ensureStepConfig()
	return p.maxAccumulatedTime
}

func (p *StagePhysics) SetMaxAccumulatedTime(maxAccumulatedTime float64) {
	if maxAccumulatedTime <= 0 {
		slog.Error("stage physics max accumulated time must be greater than zero")
		return
	}
	p.maxAccumulatedTime = maxAccumulatedTime
	p.customMaxAccumTime = true
	if p.accumulatedTime > p.maxAccumulatedTime {
		p.accumulatedTime = p.maxAccumulatedTime
	}
}

func (p *StagePhysics) FindHit(hit graviton.Hit) (*StagePhysicsEntry, bool) {
	return p.FindBody(hit.Body)
}

func (p *StagePhysics) FindBody(body *graviton.RigidBody) (*StagePhysicsEntry, bool) {
	if body == nil {
		return nil, false
	}
	for i := range p.entities {
		if p.entities[i].Body == body {
			return &p.entities[i], true
		}
	}
	return nil, false
}

func (p *StagePhysics) FindEntity(entity *Entity) (*StagePhysicsEntry, bool) {
	if entity == nil {
		return nil, false
	}
	for i := range p.entities {
		if p.entities[i].Entity == entity {
			return &p.entities[i], true
		}
	}
	return nil, false
}

func (p *StagePhysics) RigidBody(entity *Entity) (*graviton.RigidBody, bool) {
	entry, ok := p.FindEntity(entity)
	if !ok {
		return nil, false
	}
	return entry.Body, true
}

func (p *StagePhysics) RigidBodies(entityA, entityB *Entity) (*graviton.RigidBody, *graviton.RigidBody, bool) {
	return p.constraintBodies(entityA, entityB)
}

func (p *StagePhysics) Start() {
	defer tracing.NewRegion("StagePhysics.StagePhysics").End()
	if p.active {
		slog.Error("Stage physics has already started, can not start again")
		return
	}
	p.ensureStepConfig()
	p.world.Initialize()
	p.world.SetGravity(matrix.NewVec3(0, -9.81, 0))
	p.active = true
}

func (p *StagePhysics) Destroy() {
	defer tracing.NewRegion("StagePhysics.Destroy").End()
	if p.active {
		p.world.Clear()
	}
	p.entities = klib.WipeSlice(p.entities)
	p.constraints = klib.WipeSlice(p.constraints)
	p.accumulatedTime = 0
	p.active = false
}

func (p *StagePhysics) AddEntity(entity *Entity, body *graviton.RigidBody) {
	defer tracing.NewRegion("StagePhysics.AddEntity").End()
	if !p.active {
		slog.Error("stage physics has not started, can not add entity")
		return
	}
	if entity == nil || body == nil {
		slog.Error("failed to add entity physics, entity and body are required")
		return
	}
	body.Transform.SetPosition(entity.Transform.WorldPosition())
	body.Transform.SetRotation(entity.Transform.WorldRotation())
	body.Transform.SetScale(entity.Transform.WorldScale())
	stageBody := p.world.AddBody(body)
	if stageBody == nil {
		slog.Error("failed to add entity physics body")
		return
	}
	p.entities = append(p.entities, StagePhysicsEntry{
		Entity: entity,
		Body:   stageBody,
	})
	entity.OnDestroy.Add(func() {
		cIdx := -1
		for i := range p.entities {
			if p.entities[i].Entity == entity {
				cIdx = i
				break
			}
		}
		if cIdx != -1 {
			p.entities = klib.RemoveUnordered(p.entities, cIdx)
			p.world.RemoveBody(stageBody)
		}
	})
}

func (p *StagePhysics) AddConstraint(entityA, entityB *Entity, constraint *graviton.Constraint) *graviton.Constraint {
	defer tracing.NewRegion("StagePhysics.AddConstraint").End()
	if !p.active {
		slog.Error("stage physics has not started, can not add constraint")
		return nil
	}
	if constraint == nil {
		slog.Error("failed to add entity physics constraint, constraint is required")
		return nil
	}
	bodyA, bodyB, ok := p.constraintBodies(entityA, entityB)
	if !ok {
		return nil
	}
	stageConstraint := p.world.AddConstraintWithBodies(constraint, bodyA, bodyB)
	if stageConstraint == nil {
		slog.Error("failed to add entity physics constraint")
		return nil
	}
	p.trackConstraint(entityA, entityB, stageConstraint, func() {
		p.world.RemoveConstraint(stageConstraint)
	})
	return stageConstraint
}

func (p *StagePhysics) AddDistanceJoint(entityA, entityB *Entity, localAnchorA, localAnchorB matrix.Vec3) *graviton.DistanceJoint {
	defer tracing.NewRegion("StagePhysics.AddDistanceJoint").End()
	if !p.active {
		slog.Error("stage physics has not started, can not add distance joint")
		return nil
	}
	bodyA, bodyB, ok := p.constraintBodies(entityA, entityB)
	if !ok {
		return nil
	}
	joint := p.world.NewDistanceJoint(bodyA, bodyB, localAnchorA, localAnchorB)
	if joint == nil {
		slog.Error("failed to add entity physics distance joint")
		return nil
	}
	p.trackConstraint(entityA, entityB, joint.Constraint(), func() {
		p.world.RemoveDistanceJoint(joint)
	})
	return joint
}

func (p *StagePhysics) AddDistanceJointToWorld(entity *Entity, localAnchor, worldAnchor matrix.Vec3) *graviton.DistanceJoint {
	defer tracing.NewRegion("StagePhysics.AddDistanceJointToWorld").End()
	return p.AddDistanceJoint(entity, nil, localAnchor, worldAnchor)
}

func (p *StagePhysics) AddRopeJoint(entityA, entityB *Entity, localAnchorA, localAnchorB matrix.Vec3) *graviton.RopeJoint {
	defer tracing.NewRegion("StagePhysics.AddRopeJoint").End()
	if !p.active {
		slog.Error("stage physics has not started, can not add rope joint")
		return nil
	}
	bodyA, bodyB, ok := p.constraintBodies(entityA, entityB)
	if !ok {
		return nil
	}
	joint := p.world.NewRopeJoint(bodyA, bodyB, localAnchorA, localAnchorB)
	if joint == nil {
		slog.Error("failed to add entity physics rope joint")
		return nil
	}
	p.trackConstraint(entityA, entityB, joint.Constraint(), func() {
		p.world.RemoveRopeJoint(joint)
	})
	return joint
}

func (p *StagePhysics) AddRopeJointToWorld(entity *Entity, localAnchor, worldAnchor matrix.Vec3) *graviton.RopeJoint {
	defer tracing.NewRegion("StagePhysics.AddRopeJointToWorld").End()
	return p.AddRopeJoint(entity, nil, localAnchor, worldAnchor)
}

func (p *StagePhysics) AddPointJoint(entityA, entityB *Entity, localAnchorA, localAnchorB matrix.Vec3) *graviton.PointJoint {
	defer tracing.NewRegion("StagePhysics.AddPointJoint").End()
	if !p.active {
		slog.Error("stage physics has not started, can not add point joint")
		return nil
	}
	bodyA, bodyB, ok := p.constraintBodies(entityA, entityB)
	if !ok {
		return nil
	}
	joint := p.world.NewPointJoint(bodyA, bodyB, localAnchorA, localAnchorB)
	if joint == nil {
		slog.Error("failed to add entity physics point joint")
		return nil
	}
	p.trackConstraint(entityA, entityB, joint.Constraint(), func() {
		p.world.RemovePointJoint(joint)
	})
	return joint
}

func (p *StagePhysics) AddPointJointToWorld(entity *Entity, localAnchor, worldAnchor matrix.Vec3) *graviton.PointJoint {
	defer tracing.NewRegion("StagePhysics.AddPointJointToWorld").End()
	return p.AddPointJoint(entity, nil, localAnchor, worldAnchor)
}

func (p *StagePhysics) AddHingeJoint(
	entityA, entityB *Entity,
	localAnchorA, localAnchorB, localAxisA, localAxisB matrix.Vec3,
) *graviton.HingeJoint {
	defer tracing.NewRegion("StagePhysics.AddHingeJoint").End()
	if !p.active {
		slog.Error("stage physics has not started, can not add hinge joint")
		return nil
	}
	bodyA, bodyB, ok := p.constraintBodies(entityA, entityB)
	if !ok {
		return nil
	}
	joint := p.world.NewHingeJoint(bodyA, bodyB, localAnchorA, localAnchorB, localAxisA, localAxisB)
	if joint == nil {
		slog.Error("failed to add entity physics hinge joint")
		return nil
	}
	p.trackConstraint(entityA, entityB, joint.Constraint(), func() {
		p.world.RemoveHingeJoint(joint)
	})
	return joint
}

func (p *StagePhysics) AddHingeJointToWorld(
	entity *Entity, localAnchor, worldAnchor, localAxis, worldAxis matrix.Vec3,
) *graviton.HingeJoint {
	defer tracing.NewRegion("StagePhysics.AddHingeJointToWorld").End()
	return p.AddHingeJoint(entity, nil, localAnchor, worldAnchor, localAxis, worldAxis)
}

func (p *StagePhysics) AddEntityShape(entity *Entity, mass float32, shape graviton.Shape) {
	defer tracing.NewRegion("StagePhysics.AddEntityShape").End()
	t := &entity.Transform
	inertia := graviton.CalculateLocalInertia(shape, matrix.Float(mass))
	body := &graviton.RigidBody{}
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(t.Position())
	body.Transform.SetRotation(t.Rotation())
	body.SetShape(shape)
	if mass <= 0 {
		body.SetStatic()
	} else {
		body.SetDynamic(matrix.Float(mass), inertia)
	}
	p.AddEntity(entity, body)
}

func (p *StagePhysics) AddEntityTerrain(entity *Entity, terrain *graviton.TerrainCollision) {
	defer tracing.NewRegion("StagePhysics.AddEntityTerrain").End()
	t := &entity.Transform
	body := &graviton.RigidBody{}
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(t.Position())
	body.Transform.SetRotation(t.Rotation())
	body.SetStaticTerrain(terrain)
	p.AddEntity(entity, body)
}

func (p *StagePhysics) Update(workGroup *concurrent.WorkGroup, threads *concurrent.Threads, deltaTime float64) {
	defer tracing.NewRegion("StagePhysics.Update").End()
	p.ensureStepConfig()
	for i := range p.entities {
		entry := &p.entities[i]
		if entry.Body.IsKinematic() || (entry.Body.IsStatic() && entry.Entity.Transform.IsDirty()) {
			entry.syncEntityToBody()
		}
	}
	if deltaTime <= 0 {
		p.world.Step(workGroup, threads, 0)
	} else {
		p.accumulatedTime += deltaTime
		if p.accumulatedTime > p.maxAccumulatedTime {
			p.accumulatedTime = p.maxAccumulatedTime
		}
		steps := 0
		for p.accumulatedTime >= p.fixedTimeStep && steps < p.maxSubSteps {
			p.world.Step(workGroup, threads, p.fixedTimeStep)
			p.accumulatedTime -= p.fixedTimeStep
			steps++
		}
	}
	for i := range p.entities {
		if p.entities[i].Body.IsDynamic() {
			p.entities[i].syncBodyToEntity()
		}
	}
}

func (p *StagePhysics) constraintBodies(entityA, entityB *Entity) (*graviton.RigidBody, *graviton.RigidBody, bool) {
	bodyA, ok := p.RigidBody(entityA)
	if !ok {
		slog.Error("failed to add entity physics constraint, first entity has no staged body")
		return nil, nil, false
	}
	if entityB == nil {
		return bodyA, nil, true
	}
	bodyB, ok := p.RigidBody(entityB)
	if !ok {
		slog.Error("failed to add entity physics constraint, second entity has no staged body")
		return nil, nil, false
	}
	return bodyA, bodyB, true
}

func (p *StagePhysics) trackConstraint(entityA, entityB *Entity, constraint *graviton.Constraint, remove func()) {
	idx := len(p.constraints)
	p.constraints = append(p.constraints, stagePhysicsConstraintEntry{
		EntityA:    entityA,
		EntityB:    entityB,
		Constraint: constraint,
		remove:     remove,
	})
	removeConstraint := func() { p.removeConstraintByTrackedEntry(constraint, entityA, entityB, idx) }
	if entityA != nil {
		entityA.OnDestroy.Add(removeConstraint)
	}
	if entityB != nil {
		entityB.OnDestroy.Add(removeConstraint)
	}
}

func (p *StagePhysics) removeConstraintByTrackedEntry(
	constraint *graviton.Constraint, entityA, entityB *Entity, expectedIdx int,
) {
	if expectedIdx >= 0 && expectedIdx < len(p.constraints) {
		entry := &p.constraints[expectedIdx]
		if entry.matches(constraint, entityA, entityB) {
			p.removeConstraintAt(expectedIdx)
			return
		}
	}
	for i := range p.constraints {
		if p.constraints[i].matches(constraint, entityA, entityB) {
			p.removeConstraintAt(i)
			return
		}
	}
}

func (p *StagePhysics) removeConstraintAt(idx int) {
	entry := p.constraints[idx]
	if entry.remove != nil {
		entry.remove()
	} else {
		p.world.RemoveConstraint(entry.Constraint)
	}
	p.constraints = klib.RemoveUnordered(p.constraints, idx)
}

func (e *stagePhysicsConstraintEntry) matches(constraint *graviton.Constraint, entityA, entityB *Entity) bool {
	if e == nil {
		return false
	}
	if constraint != nil && e.Constraint == constraint {
		return e.EntityA == entityA && e.EntityB == entityB
	}
	return e.EntityA == entityA && e.EntityB == entityB
}

func (p *StagePhysics) ensureStepConfig() {
	if p.fixedTimeStep <= 0 {
		p.fixedTimeStep = defaultPhysicsFixedTimeStep
	}
	if p.maxSubSteps < 1 {
		p.maxSubSteps = defaultPhysicsMaxSubSteps
	}
	p.ensureMaxAccumulatedTime()
}

func (p *StagePhysics) ensureMaxAccumulatedTime() {
	if p.fixedTimeStep <= 0 || p.maxSubSteps < 1 {
		return
	}
	if !p.customMaxAccumTime {
		p.maxAccumulatedTime = p.fixedTimeStep * float64(p.maxSubSteps)
	}
	if p.accumulatedTime > p.maxAccumulatedTime {
		p.accumulatedTime = p.maxAccumulatedTime
	}
}

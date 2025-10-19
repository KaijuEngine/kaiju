/******************************************************************************/
/* physics_stage.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package stage_physics

import (
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/physics"
)

type StagePhysicsEntry struct {
	Entity *engine.Entity
	Body   *physics.RigidBody
}

type StagePhysics struct {
	world    *physics.World
	entities []StagePhysicsEntry
}

func (pe *StagePhysicsEntry) updateTransform() {
	b := pe.Body
	t := &pe.Entity.Transform
	t.SetPosition(b.Position())
	t.SetRotation(b.Rotation().ToEuler())
}

func (p *StagePhysics) IsValid() bool         { return p.world != nil }
func (p *StagePhysics) World() *physics.World { return p.world }

func (p *StagePhysics) FindCollision(hit physics.CollisionHit) (*StagePhysicsEntry, bool) {
	if !hit.IsValid() {
		return nil, false
	}
	obj := hit.Object()
	for i := range p.entities {
		if p.entities[i].Body.IsCollisionObject(obj) {
			return &p.entities[i], true
		}
	}
	return nil, false
}

func New(host *engine.Host) StagePhysics {
	p := StagePhysics{}
	broadphase := physics.NewBroadphaseInterface()
	collisionConfig := physics.NewDefaultCollisionConfiguration()
	dispatcher := physics.NewCollisionDispatcher(collisionConfig)
	solver := physics.NewSequentialImpulseConstraintSolver()
	p.world = physics.NewDiscreteDynamicsWorld(dispatcher, broadphase, solver, collisionConfig)
	p.world.SetGravity(matrix.NewVec3(0, -9.81, 0))
	return p
}

func (p *StagePhysics) AddEntity(entity *engine.Entity, body *physics.RigidBody) {
	p.entities = append(p.entities, StagePhysicsEntry{
		Entity: entity,
		Body:   body,
	})
	p.world.AddRigidBody(body)
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
			p.world.RemoveRigidBody(body)
		}
	})
}

func (p *StagePhysics) Update(host *engine.Host, deltaTime float64) {
	const syncTask = "rigidbodySync"
	p.world.StepSimulation(float32(deltaTime))
	wg := host.WorkGroup()
	for i := range p.entities {
		wg.Add(syncTask, p.entities[i].updateTransform)
	}
	wg.Execute(syncTask, host.Threads())
}

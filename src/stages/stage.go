/******************************************************************************/
/* stage.go                                                                   */
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

package stages

import (
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/physics"
	"weak"
)

type StageId = int

type pendingRigidBodyEntity struct {
	entity    *engine.Entity
	rigidBody *physics.RigidBody
}

type Stage struct {
	id       StageId
	main     weak.Pointer[MainStage]
	entities []*engine.Entity
	pending  []pendingRigidBodyEntity
}

func NewStage() Stage {
	return Stage{}
}

func (s *Stage) AddEntity(entity *engine.Entity) {
	s.entities = append(s.entities, entity)
}

func (s *Stage) AddEntityWithRigidBody(entity *engine.Entity, body *physics.RigidBody) {
	s.entities = append(s.entities, entity)
	m := s.main.Value()
	if m != nil {
		m.physics.AddEntity(entity, body)
	} else {
		s.pending = append(s.pending, pendingRigidBodyEntity{
			entity:    entity,
			rigidBody: body,
		})
	}
}

func (s *Stage) Clear() {
	for i := range s.entities {
		s.entities[i].Destroy()
	}
	s.entities = klib.WipeSlice(s.entities)
}

func (s *Stage) addToMainStage(mainStage *MainStage, id StageId) {
	s.id = id
	s.main = weak.Make(mainStage)
	debug.EnsureMsg(len(s.pending) == 0 || mainStage.physics.IsValid(), "InitPhysics was not called on main stage, can't add rigidbodies")
	for i := range s.pending {
		mainStage.physics.AddEntity(s.pending[i].entity, s.pending[i].rigidBody)
	}
	s.pending = make([]pendingRigidBodyEntity, 0)
}

func (s *Stage) teardown() {
	for i := range s.entities {
		s.entities[i].Destroy()
	}
	s.entities = make([]*engine.Entity, 0)
}

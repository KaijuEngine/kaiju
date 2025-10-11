/******************************************************************************/
/* main_stage.go                                                              */
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
	"kaiju/stages/stage_physics"
)

type MainStage struct {
	host        *engine.Host
	physics     stage_physics.StagePhysics
	stages      []Stage
	updateId    int
	nextStageId StageId
}

func NewMainStage(host *engine.Host) MainStage {
	return MainStage{
		host: host,
	}
}

func (s *MainStage) InitPhysics() {
	debug.Ensure(!s.physics.IsValid())
	s.physics = stage_physics.New(s.host)
	s.updateId = s.host.Updater.AddUpdate(s.update)
}

func (s *MainStage) Teardown() {
	s.host.Updater.RemoveUpdate(s.updateId)
	s.updateId = 0
	s.TeardownStages()
}

func (s *MainStage) TeardownStages() {
	for i := range s.stages {
		s.stages[i].teardown()
	}
	s.stages = s.stages[:0]
}

func (s *MainStage) Physics() *stage_physics.StagePhysics { return &s.physics }

func (s *MainStage) AddStage(stage Stage) StageId {
	s.nextStageId++
	stage.addToMainStage(s, s.nextStageId)
	s.stages = append(s.stages, stage)
	return stage.id
}

func (s *MainStage) Stage(stageId StageId) (*Stage, bool) {
	for i := range s.stages {
		if s.stages[i].id == stageId {
			return &s.stages[i], true
		}
	}
	return nil, false
}

func (s *MainStage) RemoveStage(stageId StageId) {
	for i := range s.stages {
		if s.stages[i].id == stageId {
			s.stages[i].teardown()
			s.stages = klib.RemoveUnordered(s.stages, i)
		}
	}
}

func (s *MainStage) update(deltaTime float64) {
	s.physics.Update(s.host, deltaTime)
}

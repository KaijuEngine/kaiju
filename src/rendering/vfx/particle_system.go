/******************************************************************************/
/* particle_system.go                                                         */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package vfx

import (
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"slices"
)

type ParticleSystemSpec []EmitterConfig

type ParticleSystem struct {
	host     *engine.Host
	entity   *engine.Entity
	Emitters []Emitter
	updateId engine.UpdateId
}

func (p *ParticleSystem) Initialize(host *engine.Host, entity *engine.Entity, spec ParticleSystemSpec) {
	defer tracing.NewRegion("ParticleSystem.Initialize").End()
	p.host = host
	p.entity = entity
	p.Emitters = make([]Emitter, len(spec))
	for i := range p.Emitters {
		p.Emitters[i].Initialize(host, spec[i])
	}
	p.updateId = host.Updater.AddUpdate(p.update)
	p.entity.OnDestroy.Add(p.Destroy)
}

func (p *ParticleSystem) Destroy() {
	defer tracing.NewRegion("ParticleSystem.Initialize").End()
	p.Clear()
	p.host.Updater.RemoveUpdate(&p.updateId)
}

func (p *ParticleSystem) Clear() {
	defer tracing.NewRegion("ParticleSystem.Clear").End()
	for i := range p.Emitters {
		p.Emitters[i].Destroy()
	}
	p.Emitters = klib.WipeSlice(p.Emitters)
}

func (p *ParticleSystem) Activate() {
	defer tracing.NewRegion("ParticleSystem.Activate").End()
	for i := range p.Emitters {
		p.Emitters[i].Activate()
	}
}

func (p *ParticleSystem) Deactivate() {
	defer tracing.NewRegion("ParticleSystem.Deactivate").End()
	for i := range p.Emitters {
		p.Emitters[i].Deactivate()
	}
}

func (p *ParticleSystem) AddEmitter(cfg EmitterConfig) *Emitter {
	defer tracing.NewRegion("ParticleSystem.AddEmitter").End()
	p.Emitters = append(p.Emitters, Emitter{})
	last := &p.Emitters[len(p.Emitters)-1]
	last.Initialize(p.host, cfg)
	return last
}

func (p *ParticleSystem) RemoveEmitter(idx int) {
	defer tracing.NewRegion("ParticleSystem.RemoveEmitter").End()
	p.Emitters[idx].Destroy()
	p.Emitters = slices.Delete(p.Emitters, idx, idx+1)
}

func (p *ParticleSystem) update(deltaTime float64) {
	defer tracing.NewRegion("ParticleSystem.update").End()
	for i := range p.Emitters {
		p.Emitters[i].update(&p.entity.Transform, deltaTime)
	}
}

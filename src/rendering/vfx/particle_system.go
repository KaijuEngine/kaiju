/******************************************************************************/
/* particle_system.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vfx

import (
	"encoding/json"
	"slices"

	"kaijuengine.com/engine"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

type ParticleSystemSpec []EmitterConfig

type ParticleSystem struct {
	host     *engine.Host
	entity   *engine.Entity
	Emitters []Emitter
	updateId engine.UpdateId
}

func LoadSpec(host *engine.Host, id string) (ParticleSystemSpec, error) {
	var spec ParticleSystemSpec
	data, err := host.AssetDatabase().Read(id)
	if err != nil {
		return spec, err
	}
	err = json.Unmarshal(data, &spec)
	return spec, err
}

func (p *ParticleSystem) IsValid() bool { return p.host != nil }

func (p *ParticleSystem) Initialize(host *engine.Host, entity *engine.Entity, spec ParticleSystemSpec) {
	defer tracing.NewRegion("ParticleSystem.Initialize").End()
	p.host = host
	p.entity = entity
	p.updateId = host.Updater.AddUpdate(p.update)
	p.entity.OnDestroy.Add(p.Destroy)
	p.LoadSpec(host, spec)
}

func (p *ParticleSystem) LoadSpec(host *engine.Host, spec ParticleSystemSpec) {
	p.Emitters = make([]Emitter, len(spec))
	for i := range p.Emitters {
		p.Emitters[i].Initialize(host, spec[i])
	}
}

func (p *ParticleSystem) Destroy() {
	defer tracing.NewRegion("ParticleSystem.Initialize").End()
	p.Clear()
	if p.host != nil {
		p.host.Updater.RemoveUpdate(&p.updateId)
	}
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

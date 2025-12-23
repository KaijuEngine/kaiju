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

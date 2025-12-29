/******************************************************************************/
/* emitter.go                                                                 */
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
	"kaiju/engine/assets"
	"kaiju/engine_entity_data/content_id"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
	"math/rand/v2"
	"time"
)

type Emitter struct {
	Config       EmitterConfig
	rand         *rand.Rand
	path         func(t float64) matrix.Vec3
	particles    []Particle
	particleData []shader_data_registry.ShaderDataParticle
	available    []int
	nextSpawn    float64
	lifeTime     float64
	pathT        float64
	offset       matrix.Vec3
	deactivated  bool
}

type EmitterConfig struct {
	Texture          content_id.Texture
	SpawnRate        float64
	ParticleLifeSpan float32
	LifeSpan         float64
	Offset           matrix.Vec3
	DirectionMin     matrix.Vec3
	DirectionMax     matrix.Vec3
	VelocityMinMax   matrix.Vec2
	OpacityMinMax    matrix.Vec2
	Color            matrix.Color
	PathFuncName     string                      `options:"PathFuncName"`
	PathFunc         func(t float64) matrix.Vec3 `visible:"hidden"`
	PathFuncOffset   float64
	PathFuncScale    float32
	PathFuncSpeed    float32
	FadeOutOverLife  bool
	Burst            bool
	Repeat           bool
}

func (e *Emitter) IsValid() bool { return e.rand != nil }

func (e *Emitter) Initialize(host *engine.Host, config EmitterConfig) {
	defer tracing.NewRegion("Emitter.Initialize").End()
	if e.IsValid() {
		e.ReloadConfig(host)
		return
	}
	e.Config = config
	seed1 := uint64(time.Now().UnixNano())
	seed2 := uint64(float64(time.Now().UnixNano()) * 0.13)
	e.rand = rand.New(rand.NewPCG(seed1, seed2))
	e.ReloadConfig(host)
}

func (e *Emitter) Destroy() {
	defer tracing.NewRegion("Emitter.Destroy").End()
	for i := range e.particleData {
		e.particleData[i].Destroy()
	}
}

func (e *Emitter) Activate() {
	defer tracing.NewRegion("Emitter.Activate").End()
	for i := range e.particleData {
		e.particleData[i].Activate()
	}
	for i := range e.available {
		e.particleData[e.available[i]].Deactivate()
	}
	e.deactivated = false
}

func (e *Emitter) Deactivate() {
	defer tracing.NewRegion("Emitter.Deactivate").End()
	for i := range e.particleData {
		e.particleData[i].Deactivate()
	}
	e.deactivated = true
}

func (e *Emitter) maxSpawnCount() int {
	maxCount := 0
	if e.Config.SpawnRate > 0 {
		maxCount = int(matrix.Ceil(float32(1 / e.Config.SpawnRate * float64(e.Config.ParticleLifeSpan))))
		maxCount += int(float32(maxCount) * 0.25) // Quarter buffer for lower frame rates
	}
	return maxCount
}

func (e *Emitter) ForceReloadConfig(host *engine.Host) {
	defer tracing.NewRegion("Emitter.ForceReloadConfig").End()
	maxCount := e.maxSpawnCount()
	e.updatePathFunc()
	for i := range e.particleData {
		e.particleData[i].Destroy()
	}
	if maxCount > 0 {
		tex, err := host.TextureCache().Texture(string(e.Config.Texture), rendering.TextureFilterLinear)
		if err != nil {
			tex, _ = host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
		}
		e.particles = make([]Particle, maxCount)
		e.particleData = make([]shader_data_registry.ShaderDataParticle, cap(e.particles))
		e.available = make([]int, 0, cap(e.particles))
		// Brute forcing all particles to be instance drawings
		mesh := rendering.NewMeshQuad(host.MeshCache())
		mat, _ := host.MaterialCache().Material(assets.MaterialDefinitionParticleTransparent)
		mat = mat.CreateInstance([]*rendering.Texture{tex})
		drawings := make([]rendering.Drawing, maxCount)
		for i := range maxCount {
			e.particleData[i].ShaderDataBase = rendering.NewShaderDataBase()
			e.particleData[i].Color = matrix.ColorWhite()
			e.particleData[i].UVs = matrix.NewVec4(0, 0, 1, 1)
			e.particleData[i].Deactivate()
			drawings[i].Material = mat
			drawings[i].Mesh = mesh
			drawings[i].ShaderData = &e.particleData[i]
			// drawings[i].ViewCuller = &host.Cameras.Primary
			e.available = append(e.available, maxCount-i-1)
		}
		host.Drawings.AddDrawings(drawings)
	}
	e.nextSpawn = 0
	e.lifeTime = e.Config.LifeSpan
}

func (e *Emitter) ReloadConfig(host *engine.Host) {
	defer tracing.NewRegion("Emitter.ReloadConfig").End()
	maxCount := e.maxSpawnCount()
	if maxCount != cap(e.particles) {
		e.ForceReloadConfig(host)
	} else {
		e.updatePathFunc()
	}
	e.nextSpawn = 0
	e.lifeTime = e.Config.LifeSpan
	e.pathT = e.Config.PathFuncOffset
}

func (e *Emitter) updatePathFunc() {
	e.path = e.Config.PathFunc
	if e.Config.PathFuncName != "" {
		if fn, ok := pathFunctions[e.Config.PathFuncName]; ok {
			e.path = fn
		} else {
			slog.Error("failed to find the particle emitter path function", "name", e.Config.PathFuncName)
		}
	}
	e.offset = e.Config.Offset
}

func (e *Emitter) update(transform *matrix.Transform, deltaTime float64) {
	defer tracing.NewRegion("Emitter.update").End()
	if e.deactivated {
		return
	}
	if e.Config.LifeSpan > 0 {
		if e.lifeTime > 0 {
			e.lifeTime -= deltaTime
			e.nextSpawn -= deltaTime
		} else if e.Config.Repeat && len(e.available) == cap(e.particles) {
			e.lifeTime = e.Config.LifeSpan
		}
	} else {
		e.nextSpawn -= deltaTime
	}
	if e.nextSpawn <= 0 {
		if e.Config.Burst {
			e.calcPath(deltaTime)
			for len(e.available) > 0 {
				e.spawn(transform)
			}
		} else {
			e.nextSpawn += e.Config.SpawnRate
			for e.nextSpawn <= e.Config.SpawnRate {
				e.calcPath(e.Config.SpawnRate)
				e.spawn(transform)
				e.nextSpawn += max(0.0001, e.Config.SpawnRate)
			}
		}
		e.nextSpawn = e.Config.SpawnRate
	}
	for i := range e.particles {
		if e.particles[i].LifeSpan > 0 {
			p := &e.particles[i]
			p.update(deltaTime)
			a := e.particleData[i].Color.A()
			e.particleData[i].Color.SetA(a - p.OpacityVelocity*float32(deltaTime))
			e.particles[i].putWorldMatrix(e.particleData[i].ModelPtr())
			if e.particles[i].LifeSpan <= 0 {
				e.particleData[i].Deactivate()
				e.available = append(e.available, i)
			}
		}
	}
}

func (e *Emitter) calcPath(dt float64) {
	if e.path != nil {
		e.pathT += dt * float64(e.Config.PathFuncSpeed)
		e.offset = e.path(e.pathT).Scale(e.Config.PathFuncScale)
	}
}

func (e *Emitter) spawn(transform *matrix.Transform) {
	defer tracing.NewRegion("Emitter.spawn").End()
	if len(e.available) == 0 {
		return
	}
	idx := e.available[len(e.available)-1]
	e.available = e.available[:len(e.available)-1]
	p := &e.particles[idx]
	pd := &e.particleData[idx]
	c := &e.Config
	pd.Activate()
	pd.Color = e.Config.Color
	pd.Color.SetA(1)
	p.Transform.Position = transform.Position().Add(e.offset)
	p.Transform.Rotation = transform.Rotation()
	p.Transform.Scale = matrix.Vec3One()
	p.LifeSpan = e.Config.ParticleLifeSpan
	if e.Config.FadeOutOverLife {
		opacity := c.OpacityMinMax.X()
		if !matrix.Approx(opacity, c.OpacityMinMax.Y()) {
			opacity = randomFloat32InRange(e.rand, c.OpacityMinMax)
		}
		pd.Color.SetA(opacity)
		p.OpacityVelocity = opacity / p.LifeSpan
	}
	dir := matrix.NewVec3(
		randomFloat32InRange(e.rand, matrix.NewVec2(
			c.DirectionMin.X(), c.DirectionMax.X())),
		randomFloat32InRange(e.rand, matrix.NewVec2(
			c.DirectionMin.Y(), c.DirectionMax.Y())),
		randomFloat32InRange(e.rand, matrix.NewVec2(
			c.DirectionMin.Z(), c.DirectionMax.Z())),
	)
	v := c.VelocityMinMax.X()
	if !matrix.Approx(v, c.VelocityMinMax.Y()) {
		v = randomFloat32InRange(e.rand, c.VelocityMinMax)
	}
	p.Velocity.Position = dir.Normal().Scale(v)
}

func randomFloat32InRange(r *rand.Rand, minMax matrix.Vec2) float32 {
	return minMax.X() + r.Float32()*(minMax.Y()-minMax.X())
}

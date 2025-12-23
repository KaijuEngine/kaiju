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
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"math/rand/v2"
	"time"
)

type Emitter struct {
	config       EmitterConfig
	Transform    matrix.Transform
	host         *engine.Host
	rand         *rand.Rand
	particles    []Particle
	particleData []shader_data_registry.ShaderDataParticle
	available    []int
	updateId     engine.UpdateId
	nextSpawn    float64
	lifeTime     float64
}

type EmitterConfig struct {
	SpawnRate        float64
	ParticleLifeSpan float32
	LifeSpan         float64
	DirectionMin     matrix.Vec3
	DirectionMax     matrix.Vec3
	VelocityMinMax   matrix.Vec2
	OpacityMinMax    matrix.Vec2
	EmitVolume       collision.AABB
	FadeOutOverLife  bool
	Burst            bool
	Repeat           bool
}

func (e *Emitter) Initialize(host *engine.Host, tex *rendering.Texture, config EmitterConfig) {
	e.host = host
	e.config = config
	e.nextSpawn = 0
	maxCount := int(matrix.Ceil(float32(1 / e.config.SpawnRate * float64(e.config.ParticleLifeSpan))))
	maxCount += 1 // Little buffer for overlapping spawn/destroy
	e.particles = make([]Particle, maxCount)
	e.particleData = make([]shader_data_registry.ShaderDataParticle, cap(e.particles))
	e.available = make([]int, 0, cap(e.particles))
	e.updateId = host.Updater.AddUpdate(e.update)
	e.lifeTime = e.config.LifeSpan
	seed1 := uint64(time.Now().UnixNano())
	seed2 := uint64(float64(time.Now().UnixNano()) * 0.13)
	e.rand = rand.New(rand.NewPCG(seed1, seed2))
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
		e.available = append(e.available, maxCount-i-1)
	}
	host.Drawings.AddDrawings(drawings)
}

func (e *Emitter) Destroy() {
	e.host.Updater.RemoveUpdate(&e.updateId)
	for i := range e.particleData {
		e.particleData[i].Destroy()
	}
}

func (e *Emitter) update(deltaTime float64) {
	if e.config.LifeSpan > 0 {
		if e.lifeTime > 0 {
			e.lifeTime -= deltaTime
			e.nextSpawn -= deltaTime
		} else if e.config.Repeat && len(e.available) == cap(e.particles) {
			e.lifeTime = e.config.LifeSpan
		}
	} else {
		e.nextSpawn -= deltaTime
	}
	if e.nextSpawn <= 0 {
		if e.config.Burst {
			for len(e.available) > 0 {
				e.spawn()
			}
		} else {
			e.spawn()
		}
		e.nextSpawn = e.config.SpawnRate
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

func (e *Emitter) spawn() {
	if len(e.available) == 0 {
		return
	}
	idx := e.available[len(e.available)-1]
	e.available = e.available[:len(e.available)-1]
	p := &e.particles[idx]
	pd := &e.particleData[idx]
	c := &e.config
	pd.Activate()
	p.Transform.Position = e.Transform.Position()
	p.Transform.Rotation = e.Transform.Rotation()
	p.Transform.Scale = matrix.Vec3One()
	p.LifeSpan = e.config.ParticleLifeSpan
	if e.config.FadeOutOverLife {
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

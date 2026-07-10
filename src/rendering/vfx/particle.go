/******************************************************************************/
/* particle.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vfx

import "kaijuengine.com/matrix"

type particleTransformation struct {
	Position matrix.Vec3
	Rotation matrix.Vec3 // TODO:  This can be 1D for billboarded particle
	Scale    matrix.Vec3 // TODO:  This can be 2D for billboarded particle
}

type Particle struct {
	Transform       particleTransformation
	Velocity        particleTransformation
	OpacityVelocity matrix.Float
	LifeSpan        matrix.Float
}

func (p *Particle) update(deltaTime float64) {
	p.LifeSpan -= matrix.Float(deltaTime)
	t := &p.Transform
	v := &p.Velocity
	t.Position.AddAssign(v.Position.Scale(matrix.Float(deltaTime)))
	t.Rotation.AddAssign(v.Rotation.Scale(matrix.Float(deltaTime)))
	t.Scale.AddAssign(v.Scale.Scale(matrix.Float(deltaTime)))
}

func (p *Particle) putWorldMatrix(m *matrix.Mat4) {
	m.Reset()
	m.Scale(p.Transform.Scale)
	m.Rotate(p.Transform.Rotation)
	m.Translate(p.Transform.Position)
}

package vfx

type ParticleSystemSpec []EmitterConfig

type ParticleSystem struct {
	Emitters []Emitter
}

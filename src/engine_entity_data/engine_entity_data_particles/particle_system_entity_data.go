/******************************************************************************/
/* rigid_body_data_binding.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_particles

import (
	"log/slog"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/rendering/vfx"
)

var bindingKey = ""

type Shape int

const (
	ShapeBox Shape = iota
	ShapeSphere
	ShapeCapsule
	ShapeCylinder
	ShapeCone
)

func init() {
	engine.RegisterEntityData(ParticleSystemEntityData{})
}

func BindingKey() string {
	if bindingKey == "" {
		bindingKey = pod.QualifiedNameForLayout(ParticleSystemEntityData{})
	}
	return bindingKey
}

type ParticleSystemEntityData struct {
	Id content_id.ParticleSystem `visible:"false"`
}

func (r ParticleSystemEntityData) Init(e *engine.Entity, host *engine.Host) {
	sysSpec, err := vfx.LoadSpec(host, string(r.Id))
	if err != nil {
		slog.Error("failed to locate/decode the particle system", "id", r.Id, "error", err)
		return
	}
	sys := &vfx.ParticleSystem{}
	sys.Initialize(host, e, sysSpec)
	e.AddNamedData("ParticleSystem", sys)
}

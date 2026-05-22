/******************************************************************************/
/* particle_system_entity_data_renderer.go                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"log/slog"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/engine_entity_data/engine_entity_data_particles"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/vfx"
)

type particleSystemGizmo struct {
	id     string
	system vfx.ParticleSystem
	icon   rendering.DrawInstance
}

type ParticleSystemEntityDataRenderer struct {
	Systems map[*editor_stage_manager.StageEntity]*particleSystemGizmo
}

func init() {
	AddRenderer(engine_entity_data_particles.BindingKey(), &ParticleSystemEntityDataRenderer{
		Systems: make(map[*editor_stage_manager.StageEntity]*particleSystemGizmo),
	})
}

func (c *ParticleSystemEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("ParticleSystemEntityDataRenderer.Attached").End()
	c.Systems[target] = &particleSystemGizmo{
		icon: commonAttached(host, manager, target, "particles.png"),
	}
	target.OnActivate.Add(func() {
		if g, ok := c.Systems[target]; ok {
			g.system.Activate()
		}
	})
	target.OnDeactivate.Add(func() {
		if g, ok := c.Systems[target]; ok {
			g.system.Deactivate()
		}
	})
	target.OnDestroy.Add(func() {
		c.Detatched(host, manager, target, data)
	})
}

func (c *ParticleSystemEntityDataRenderer) Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("ParticleSystemEntityDataRenderer.Detatched").End()
	if d, ok := c.Systems[target]; ok {
		d.icon.Destroy()
		// Particle system destroys itself when the entity is destroyed
		delete(c.Systems, target)
	}
}

func (c *ParticleSystemEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	// defer tracing.NewRegion("ParticleSystemEntityDataRenderer.Show").End()
}

func (c *ParticleSystemEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	// defer tracing.NewRegion("ParticleSystemEntityDataRenderer.Hide").End()
}

func (c *ParticleSystemEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	if g, ok := c.Systems[target]; ok {
		id, ok := data.FieldValueByName("Id").(content_id.ParticleSystem)
		if !ok {
			slog.Error("particle system id failure", "id", id)
			return
		}
		if g.id == string(id) && target.Transform.IsDirty() {
			return
		}
		g.id = string(id)
		g.system.Clear()
		spec, err := vfx.LoadSpec(host, string(id))
		if err != nil {
			slog.Error("invlaid particle system id specified", "id", id, "error", err)
			return
		}
		if !g.system.IsValid() {
			g.system.Initialize(host, &target.Entity, spec)
		} else {
			g.system.LoadSpec(host, spec)
		}
	}
}

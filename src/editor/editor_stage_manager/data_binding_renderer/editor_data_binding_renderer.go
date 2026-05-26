/******************************************************************************/
/* editor_data_binding_renderer.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"weak"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/platform/profiler/tracing"
)

var renderers = map[string]DataBindingRenderer{}

type DataBindingRenderer interface {
	Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry)
	Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry)
	Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry)
	Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry)
	Hide(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry)
}

func AddRenderer(key string, r DataBindingRenderer) {
	defer tracing.NewRegion("data_binding_renderer.AddRenderer").End()
	renderers[key] = r
}

func Attached(data *entity_data_binding.EntityDataEntry, host weak.Pointer[engine.Host], manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("data_binding_renderer.Attached").End()
	h := host.Value()
	if h == nil {
		return
	}
	if r, ok := renderers[data.Gen.RegisterKey]; ok {
		r.Attached(h, manager, target, data)
	}
}

func Detatched(data *entity_data_binding.EntityDataEntry, host weak.Pointer[engine.Host], manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("data_binding_renderer.Attached").End()
	h := host.Value()
	if h == nil {
		return
	}
	if r, ok := renderers[data.Gen.RegisterKey]; ok {
		r.Detatched(h, manager, target, data)
	}
}

func Show(host weak.Pointer[engine.Host], target *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("data_binding_renderer.Show").End()
	h := host.Value()
	if h == nil {
		return
	}
	binds := target.DataBindings()
	for i := range binds {
		if r, ok := renderers[binds[i].Gen.RegisterKey]; ok {
			r.Show(h, target, binds[i])
		}
	}
}

func Updated(data *entity_data_binding.EntityDataEntry, host weak.Pointer[engine.Host], target *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("data_binding_renderer.Attached").End()
	h := host.Value()
	if h == nil {
		return
	}
	if r, ok := renderers[data.Gen.RegisterKey]; ok {
		r.Update(h, target, data)
	}
}

func ShowSpecific(data *entity_data_binding.EntityDataEntry, host weak.Pointer[engine.Host], target *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("data_binding_renderer.ShowSpecific").End()
	h := host.Value()
	if h == nil {
		return
	}
	if r, ok := renderers[data.Gen.RegisterKey]; ok {
		r.Show(h, target, data)
	}
}

func Hide(host weak.Pointer[engine.Host], target *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("data_binding_renderer.Hide").End()
	h := host.Value()
	if h == nil {
		return
	}
	binds := target.DataBindings()
	for i := range binds {
		if r, ok := renderers[binds[i].Gen.RegisterKey]; ok {
			r.Hide(h, target, binds[i])
		}
	}
}

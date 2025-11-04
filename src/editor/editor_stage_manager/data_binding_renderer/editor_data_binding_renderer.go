/******************************************************************************/
/* editor_data_binding_renderer.go                                            */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package data_binding_renderer

import (
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/platform/profiler/tracing"
	"weak"
)

var renderers = map[string]DataBindingRenderer{}

type DataBindingRenderer interface {
	Attached(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry)
	Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry)
	Hide(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry)
}

func AddRenderer(key string, r DataBindingRenderer) {
	defer tracing.NewRegion("data_binding_renderer.AddRenderer").End()
	renderers[key] = r
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

func Attached(data *entity_data_binding.EntityDataEntry, host weak.Pointer[engine.Host], target *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("data_binding_renderer.Attached").End()
	h := host.Value()
	if h == nil {
		return
	}
	if r, ok := renderers[data.Gen.RegisterKey]; ok {
		r.Attached(h, target, data)
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

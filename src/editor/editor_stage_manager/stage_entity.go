/******************************************************************************/
/* stage_entity.go                                                            */
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

package editor_stage_manager

import (
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/engine"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
	"unsafe"
)

type StageEntity struct {
	engine.Entity
	StageData    StageEntityEditorData
	dataBindings []*entity_data_binding.EntityDataEntry
	isDeleted    bool
}

func EntityToStageEntity(e *engine.Entity) *StageEntity {
	if e == nil {
		return nil
	}
	return (*StageEntity)(unsafe.Pointer(e))
}

func (e *StageEntity) DataBindings() []*entity_data_binding.EntityDataEntry { return e.dataBindings }

func (e *StageEntity) DetachDataBinding(binding *entity_data_binding.EntityDataEntry) {
	for i, b := range e.dataBindings {
		if b == binding {
			e.dataBindings = append(e.dataBindings[:i], e.dataBindings[i+1:]...)
			return
		}
	}
}

func (e *StageEntity) AttachDataBinding(binding *entity_data_binding.EntityDataEntry) {
	for _, b := range e.dataBindings {
		if b == binding {
			return
		}
	}
	e.dataBindings = append(e.dataBindings, binding)
}

func (e *StageEntity) DataBindingsByKey(key string) []*entity_data_binding.EntityDataEntry {
	out := []*entity_data_binding.EntityDataEntry{}
	for _, d := range e.dataBindings {
		if d.Gen.RegisterKey == key {
			out = append(out, d)
		}
	}
	return out
}

func (e *StageEntity) AddDataBinding(binding *entity_data_binding.EntityDataEntry) {
	e.dataBindings = append(e.dataBindings, binding)
}

func (e *StageEntity) Depth() int {
	depth := 0
	p := e.Parent
	for p != nil {
		depth++
		p = p.Parent
	}
	return depth
}

func (e *StageEntity) SetMaterial(mat *rendering.Material, manager *StageManager) {
	if mat == nil {
		slog.Error("attempting to set the material of the stage entity to a nil material")
		return
	}
	mesh := e.StageData.Mesh
	if mesh == nil {
		slog.Error("the entity doesn't currently have a mesh to apply the material to", "entity", e.Name())
		return
	}
	manager.history.Add(&attachMaterialHistory{
		m:         manager,
		e:         e,
		fromMatId: e.StageData.Description.Material,
		toMatId:   mat.Id,
	})
	e.StageData.ShaderData.Destroy()
	e.StageData.Description.Textures = make([]string, len(mat.Textures))
	e.StageData.Description.Material = mat.Id
	for i := range mat.Textures {
		e.StageData.Description.Textures[i] = mat.Textures[i].Key
	}
	e.StageData.ShaderData = shader_data_registry.Create(mat.Shader.ShaderDataName())
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ShaderData: e.StageData.ShaderData,
		Transform:  &e.Transform,
		ViewCuller: &manager.host.Cameras.Primary,
	}
	db := entity_data_binding.ToDataBinding("", e.StageData.ShaderData)
	for i := range db.Fields {
		if db.RunTagParserOnField(i) {
			db.SetField(i, db.Fields[i].Value)
		}
	}
	manager.host.Drawings.AddDrawing(draw)
}

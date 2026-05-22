/******************************************************************************/
/* stage_entity.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"unsafe"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/engine"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

type StageEntity struct {
	engine.Entity
	StageData    StageEntityEditorData
	dataBindings []*entity_data_binding.EntityDataEntry
	isDeleted    bool
	isLocked     bool
}

func EntityToStageEntity(e *engine.Entity) *StageEntity {
	if e == nil {
		return nil
	}
	return (*StageEntity)(unsafe.Pointer(e))
}

func (e *StageEntity) DataBindings() []*entity_data_binding.EntityDataEntry { return e.dataBindings }

func (e *StageEntity) IsDeleted() bool { return e.isDeleted }
func (e *StageEntity) Lock()           { e.SetLocked(true) }
func (e *StageEntity) Unlock()         { e.SetLocked(false) }
func (e *StageEntity) IsLocked() bool  { return e.isLocked }
func (e *StageEntity) SetLocked(locked bool) {
	e.isLocked = locked
	e.StageData.Description.Locked = locked
}

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
		Mesh:       e.StageData.Mesh,
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

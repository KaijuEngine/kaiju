/******************************************************************************/
/* stage_entity.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"log/slog"
	"slices"
	"unsafe"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/engine"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

type StageEntity struct {
	engine.Entity
	StageData    StageEntityEditorData
	PickID       uint32
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
	e.StageData.Description.Material = mat.Id
	e.StageData.Description.Textures = stageEntityMaterialTextureOverrides(mat)
	e.StageData.ShaderData = shader_data_registry.Create(mat.Shader.DrawInstanceDataName())
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

func stageEntityMaterialTextureOverrides(mat *rendering.Material) []string {
	if mat == nil {
		return nil
	}
	root := mat.SelectRoot()
	if root == nil || materialTextureKeysEqual(mat.Textures, root.Textures) {
		return nil
	}
	textures := make([]string, len(mat.Textures))
	for i := range mat.Textures {
		if mat.Textures[i] != nil {
			textures[i] = mat.Textures[i].Key
		}
	}
	return textures
}

func materialTextureKeysEqual(a, b []*rendering.Texture) bool {
	if len(a) != len(b) {
		return false
	}
	return slices.EqualFunc(a, b, func(left, right *rendering.Texture) bool {
		if left == nil || right == nil {
			return left == right
		}
		return left.Key == right.Key
	})
}

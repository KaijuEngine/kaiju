/******************************************************************************/
/* editor_stage_manager.go                                                    */
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
	"encoding/json"
	"errors"
	"kaiju/editor/codegen"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_events"
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/collision"
	"kaiju/engine/stages"
	"kaiju/engine/systems/events"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"weak"

	"github.com/KaijuEngine/uuid"
)

// StageManager represents the current stage in the editor. It contains all of
// the entities on the stage.
type StageManager struct {
	OnEntitySpawn         events.EventWithArg[*StageEntity]
	OnEntityDestroy       events.EventWithArg[*StageEntity]
	OnEntitySelected      events.EventWithArg[*StageEntity]
	OnEntityDeselected    events.EventWithArg[*StageEntity]
	OnEntityChangedParent events.EventWithArg[*StageEntity]
	stageId               string
	stageName             string
	host                  *engine.Host
	editorUI              EditorUserInterface
	history               *memento.History
	entities              []*StageEntity
	selected              []*StageEntity
	worldBVH              *collision.BVH
}

// StageEntityEditorData is the structure holding all the uniquely identifiable
// and linking data about the entity on this stage. That will include things
// like content linkage, data bindings, etc.
type StageEntityEditorData struct {
	Bvh                   *collision.BVH
	Mesh                  *rendering.Mesh
	ShaderData            rendering.DrawInstance
	Description           stages.EntityDescription
	PendingMaterialChange bool
}

func (m *StageManager) Initialize(host *engine.Host, history *memento.History, editorUI EditorUserInterface) {
	m.host = host
	m.history = history
	m.editorUI = editorUI
}

func (m *StageManager) NewStage() {
	defer tracing.NewRegion("StageManager.NewStage").End()
	m.Clear()
	m.history.Clear()
	m.stageId = ""
}

func (m *StageManager) IsNew() bool     { return m.stageId == "" }
func (m *StageManager) StageId() string { return m.stageId }

func (m *StageManager) SetStageId(name string, cache *content_database.Cache) error {
	defer tracing.NewRegion("StageManager.SetStageId").End()
	newId := uuid.NewString()
	if _, err := cache.Read(newId); err == nil {
		return StageAlreadyExistsError{newId}
	}
	m.stageId = newId
	m.stageName = name
	return nil
}

// List will return all of the internally held entities for the stage
func (m *StageManager) List() []*StageEntity {
	out := make([]*StageEntity, 0, len(m.entities))
	for _, e := range m.entities {
		if e.isDeleted {
			continue
		}
		out = append(out, e)
	}
	return out
}

func (m *StageManager) Selection() []*StageEntity { return m.selected }

// AddEntity will generate a new entity for the stage with a new random Id. It
// will internally just call #AddEntityWithId
func (m *StageManager) AddEntity(name string, point matrix.Vec3) *StageEntity {
	defer tracing.NewRegion("StageManager.AddEntity").End()
	e := m.AddEntityWithId(uuid.NewString(), name, point)
	return e
}

func (m *StageManager) AttachEntityData(e *StageEntity, g codegen.GeneratedType) *entity_data_binding.EntityDataEntry {
	defer tracing.NewRegion("StageManager.AttachEntityData").End()
	de := &entity_data_binding.EntityDataEntry{}
	e.AddDataBinding(de.ReadEntityDataBindingType(g))
	return de
}

func (m *StageManager) duplicateEntity(target *StageEntity, proj *project.Project) (*StageEntity, error) {
	defer tracing.NewRegion("StageManager.duplicateEntity").End()
	desc := m.entityToDescription(target)
	var newId func(d *stages.EntityDescription)
	newId = func(d *stages.EntityDescription) {
		d.Id = uuid.NewString()
		for i := range d.Children {
			newId(&d.Children[i])
		}
	}
	newId(&desc)
	return m.importEntityByDescription(m.host, proj, EntityToStageEntity(target.Parent), &desc)
}

// AddEntityWithId will create an entity for the stage with a specified Id
// rather than generating one. This entity will have a #StageEntityData
// automatically added to it as named data named "stage".
func (m *StageManager) AddEntityWithId(id, name string, point matrix.Vec3) *StageEntity {
	defer tracing.NewRegion("StageManager.AddEntityWithId").End()
	e := &StageEntity{}
	e.Init(m.host.WorkGroup())
	e.SetName(name)
	e.StageData.Description.Id = id
	m.host.AddEntity(&e.Entity)
	e.Transform.SetPosition(point)
	m.entities = append(m.entities, e)
	e.AddNamedData("stage", e.StageData)
	wm := weak.Make(m)
	we := weak.Make(e)
	e.OnDestroy.Add(func() {
		sm := wm.Value()
		if sm == nil {
			return
		}
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Destroy()
		}
		if e.StageData.Bvh != nil {
			collision.RemoveAllLeavesMatchingTransform(&m.worldBVH, &e.Transform)
		}
		se := we.Value()
		for i := range sm.entities {
			if sm.entities[i] == se {
				sm.entities = klib.RemoveUnordered(sm.entities, i)
				return
			}
		}
	})
	m.history.Add(&objectSpawnHistory{
		m: m,
		e: e,
	})
	m.OnEntitySpawn.Execute(e)
	return e
}

func (m *StageManager) DestroySelected() {
	defer tracing.NewRegion("StageManager.DestroySelected").End()
	if len(m.selected) == 0 {
		return
	}
	m.history.BeginTransaction()
	defer m.history.CommitTransaction()
	sel := []*StageEntity{}
	for _, e := range m.selected {
		sel = klib.AppendUnique(sel, explodeEntityHierarchy(e)...)
	}
	m.ClearSelection()
	h := &objectDeleteHistory{
		m:        m,
		entities: sel,
	}
	m.history.Add(h)
	// Being lazy (smart?), just calling Redo here to do the action
	h.Redo()
}

func (m *StageManager) DuplicateSelected(proj *project.Project) {
	defer tracing.NewRegion("StageManager.DuplicateSelected").End()
	sel := slices.Clone(m.Selection())
	m.history.BeginTransaction()
	defer m.history.CommitTransaction()
	m.ClearSelection()
	for _, e := range sel {
		dup, err := m.duplicateEntity(e, proj)
		if err != nil {
			slog.Error("failed to duplicate entity", "error", err)
			continue
		}
		m.history.Add(&objectSpawnHistory{
			m: m,
			e: dup,
		})
		m.SelectEntity(dup)
	}
}

func (m *StageManager) HierarchyRespectiveSelection() []*StageEntity {
	defer tracing.NewRegion("StageManager.HierarchyRespectiveSelection").End()
	sel := slices.Clone(m.Selection())
	for i := 0; i < len(sel); i++ {
		for j := i + 1; j < len(sel); j++ {
			if sel[j].HasParent(&sel[i].Entity) {
				sel = klib.RemoveUnordered(sel, j)
				j--
				continue
			}
			if sel[i].HasParent(&sel[j].Entity) {
				sel = klib.RemoveUnordered(sel, i)
				i--
				break
			}
		}
	}
	return sel
}

func (m *StageManager) EntityById(id string) (*StageEntity, bool) {
	defer tracing.NewRegion("StageManager.EntityById").End()
	if id == "" {
		return nil, false
	}
	for i := range m.entities {
		if m.entities[i].StageData.Description.Id == id {
			return m.entities[i], true
		}
	}
	return nil, false
}

func (m *StageManager) SetEntityParent(child, parent *StageEntity) {
	defer tracing.NewRegion("StageManager.SetEntityParent").End()
	lastParent := EntityToStageEntity(child.Parent)
	if parent != nil {
		child.SetParent(&parent.Entity)
	} else {
		child.SetParent(nil)
	}
	if parent != nil && parent.StageData.Bvh != nil {
		m.RefitBVH(parent)
	} else if child.StageData.Bvh != nil {
		m.RefitBVH(child)
	}
	m.OnEntityChangedParent.Execute(child)
	m.history.Add(&changeParentHistory{
		m:          m,
		e:          child,
		prevParent: lastParent,
		nextParent: parent,
	})
}

// Clear will destroy all entities that are managed by this stage manager.
func (m *StageManager) Clear() {
	defer tracing.NewRegion("StageManager.Clear").End()
	for i := len(m.entities) - 1; i >= 0; i-- {
		m.OnEntityDestroy.Execute(m.entities[i])
		m.entities[i].Destroy()
		// Deleted entities are not in the host and need to be cleaned up manually
		if m.entities[i].isDeleted {
			m.entities[i].ForceCleanup()
		}
	}
	m.worldBVH = nil
}

func (m *StageManager) RefitWorldBVH() { m.worldBVH.Refit() }

func (m *StageManager) AddBVH(bvh *collision.BVH, transform *matrix.Transform) {
	defer tracing.NewRegion("StageManager.AddBVH").End()
	cpy := collision.CloneBVH(bvh)
	collision.AddSubBVH(&m.worldBVH, cpy, transform)
	m.RefitWorldBVH()
}

//func (m *StageManager) RemoveBVH(bvh *collision.BVH) {
//	defer tracing.NewRegion("StageManager.RemoveBVH").End()
//	collision.RemoveSubBVH(&m.worldBVH, bvh)
//}

func (m *StageManager) RemoveEntityBVH(e *StageEntity) {
	defer tracing.NewRegion("StageManager.RemoveBVH").End()
	collision.RemoveAllLeavesMatchingTransform(&m.worldBVH, &e.Transform)
}

// entityToTemplate is a wrapper around [entityToDescription] so that the
// function name is clear when called
func (m *StageManager) entityToTemplate(target *StageEntity) stages.EntityDescription {
	desc := m.entityToDescription(target)
	// We don't store the template id in the template itself
	desc.TemplateId = ""
	return desc
}

func (m *StageManager) entityToDescription(parent *StageEntity) stages.EntityDescription {
	desc := &parent.StageData.Description
	desc.Name = parent.Name()
	desc.Position = parent.Transform.Position()
	desc.Rotation = parent.Transform.Rotation()
	desc.Scale = parent.Transform.Scale()
	desc.DataBinding = make([]stages.EntityDataBinding, 0, len(parent.dataBindings))
	desc.ShaderData = make([]stages.EntityDescriptionShaderDataField, 0)
	desc.RawDataBinding = make([]any, 0, len(parent.dataBindings))
	desc.Children = make([]stages.EntityDescription, 0)
	for _, d := range parent.dataBindings {
		db := stages.EntityDataBinding{
			RegistraionKey: d.Gen.RegisterKey,
			Fields:         make(map[string]any),
		}
		for i := range d.Fields {
			db.Fields[d.Fields[i].Name] = d.FieldValue(i)
		}
		desc.DataBinding = append(desc.DataBinding, db)
		desc.RawDataBinding = append(desc.RawDataBinding, d.BoundData)
	}
	if parent.StageData.ShaderData != nil {
		v := reflect.ValueOf(parent.StageData.ShaderData)
		for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		t := v.Type()
		for i := range t.NumField() {
			f := t.Field(i)
			if f.Tag.Get("visible") == "false" {
				continue
			}
			desc.ShaderData = append(desc.ShaderData, stages.EntityDescriptionShaderDataField{
				Name:  f.Name,
				Index: int32(i),
				Value: v.Field(i).Interface(),
			})
		}
	}
	for _, e := range m.entities {
		if e.isDeleted {
			continue
		}
		if e.Parent == &parent.Entity {
			m.entityToDescription(e)
			desc.Children = append(desc.Children, e.StageData.Description)
		}
	}
	return parent.StageData.Description
}

func (m *StageManager) toStage() stages.Stage {
	defer tracing.NewRegion("StageManager.toStage").End()
	s := stages.Stage{Id: m.stageId}
	rootCount := 0
	for i := range m.entities {
		if m.entities[i].isDeleted {
			continue
		}
		if m.entities[i].IsRoot() {
			rootCount++
		}
	}
	s.Entities = make([]stages.EntityDescription, 0, rootCount)
	for _, e := range m.entities {
		if e.isDeleted {
			continue
		}
		if e.IsRoot() {
			m.entityToDescription(e)
			s.Entities = append(s.Entities, e.StageData.Description)
		}
	}
	return s
}

func (m *StageManager) SaveStage(cache *content_database.Cache, fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("StageManager.SaveStage").End()
	s := m.toStage()
	// TODO:  Run through the stage importer?
	f, err := fs.Create(filepath.Join(project_file_system.ContentFolder,
		project_file_system.ContentStageFolder, m.stageId))
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(s.ToMinimized()); err != nil {
		return err
	}
	// TODO:  Run through the stage importer?
	configPath := project_file_system.StagePath(m.stageId).ToConfigPath()
	cfg := content_database.ContentConfig{}
	cfg.Name = m.stageName
	cfg.Type = content_database.Stage{}.TypeName()
	f2, err := fs.Create(configPath.String())
	if err != nil {
		return err
	}
	defer f2.Close()
	if err := json.NewEncoder(f2).Encode(cfg); err != nil {
		// TODO:  Roll back
		return err
	}
	if err := cache.Index(configPath.String(), fs); err != nil {
		// TODO:  Roll back
		return err
	}
	slog.Info("Stage saved successfully")
	return nil
}

func (m *StageManager) LoadStage(id string, host *engine.Host, cache *content_database.Cache, proj *project.Project) error {
	defer tracing.NewRegion("StageManager.LoadStage").End()
	fs := proj.FileSystem()
	m.Clear()
	cc, err := cache.Read(id)
	if err != nil {
		return err
	}
	f, err := fs.Open(content_database.ToContentPath(cc.Path))
	if err != nil {
		return err
	}
	defer f.Close()
	var ss stages.StageJson
	if err := json.NewDecoder(f).Decode(&ss); err != nil {
		return err
	}
	s := stages.Stage{}
	s.FromMinimized(ss)
	for i := range s.Entities {
		if _, err := m.importEntityByDescription(host, proj, nil, &s.Entities[i]); err != nil {
			return err
		}
	}
	m.stageId = id
	m.stageName = cc.Config.Name
	return nil
}

func (m *StageManager) CreateTemplateFromSelected(edEvts *editor_events.EditorEvents, proj *project.Project) error {
	defer tracing.NewRegion("StageManager.CreateTemplateFromSelected").End()
	sel := m.Selection()
	switch len(sel) {
	case 0:
		const err = "can't create a template with nothing selected"
		slog.Error(err)
		return errors.New(err)
	case 1:
		// This is expected
	default:
		const err = "can't create a template with multiple entities selected"
		slog.Error(err)
		return errors.New(err)
	}
	cache := proj.CacheDatabase()
	fs := proj.FileSystem()
	target := sel[0]
	if target.StageData.Description.TemplateId != "" {
		m.editorUI.BlurInterface()
		confirm_prompt.Show(m.host, confirm_prompt.Config{
			Title:       "Overwrite template",
			Description: "The selected is already a template, would you like to overwrite the template? This will update all usages of this template in all stages. If this stage has other instances of this template, they will all be updated and you won't be able to undo beyond this point. Would you like to continue?",
			ConfirmText: "Yes",
			CancelText:  "Cancel",
			OnConfirm: func() {
				m.editorUI.FocusInterface()
				// Update the existing template
				id := target.StageData.Description.TemplateId
				cc, err := cache.Read(id)
				if err != nil {
					slog.Error("failed to read the cache for the existing template id", "id", id, "error", err)
					return
				}
				f, err := fs.Create(content_database.ToContentPath(cc.Path))
				if err != nil {
					slog.Error("failed to open the content file for writing", "id", id, "error", err)
					return
				}
				defer f.Close()
				if err = json.NewEncoder(f).Encode(m.entityToTemplate(target)); err != nil {
					slog.Error("failed to encode the template to it's file", "id", id, "error", err)
					return
				}
				m.updateExistingTemplateInstances(target, m.host, proj, id)
			},
			OnCancel: m.editorUI.FocusInterface,
		})
	} else {
		tpl := m.entityToTemplate(target)
		f, err := os.CreateTemp("", "*.template")
		if err != nil {
			slog.Error("failed to create the entity template file", "error", err)
			return err
		}
		defer os.Remove(f.Name())
		defer f.Close()
		if err = json.NewEncoder(f).Encode(tpl); err != nil {
			slog.Error("failed to encode the entity template file", "error", err)
			return err
		}
		res, err := content_database.Import(f.Name(), fs, cache, "")
		if err != nil || len(res) != 1 {
			slog.Error("failed to import the template as content", "error", err)
			return err
		}
		id := res[0].Id
		target.StageData.Description.TemplateId = id
		defer edEvts.OnContentAdded.Execute([]string{id})
		name := target.Name()
		if strings.TrimSpace(name) == "" {
			return nil
		}
		c, err := cache.Read(id)
		if err != nil {
			slog.Warn("failed to read the cache for the template that was just created", "error", err)
			return nil
		}
		c.Config.Name = name
		if err := content_database.WriteConfig(c.Path, c.Config, fs); err != nil {
			slog.Warn("failed to update the name for the template", "error", err)
			return nil
		}
		cache.IndexCachedContent(c)
	}
	return nil
}

func (m *StageManager) SpawnTemplate(host *engine.Host, proj *project.Project, cc *content_database.CachedContent, point matrix.Vec3) (*StageEntity, error) {
	defer tracing.NewRegion("StageManager.SpawnTemplate").End()
	m.history.BeginTransaction()
	defer m.history.CommitTransaction()
	f, err := proj.FileSystem().Open(content_database.ToContentPath(cc.Path))
	if err != nil {
		slog.Error("failed to load the template file", "path", cc.Path, "error", err)
		return nil, err
	}
	defer f.Close()
	var desc stages.EntityDescription
	if err = json.NewDecoder(f).Decode(&desc); err != nil {
		slog.Error("failed to decode the entity template file", "path", cc.Path, "error", err)
		return nil, err
	}
	desc.Position = point
	desc.TemplateId = cc.Id()
	var generateId func(d *stages.EntityDescription)
	generateId = func(d *stages.EntityDescription) {
		d.Id = uuid.NewString()
		for i := range d.Children {
			generateId(&d.Children[i])
		}
	}
	generateId(&desc)
	e, err := m.importEntityByDescription(host, proj, nil, &desc)
	if err != nil {
		slog.Error("failed to spawn the entity from entity template", "path", cc.Path, "error", err)
		return nil, err
	}
	m.ClearSelection()
	m.SelectEntity(e)
	return e, nil
}

func (m *StageManager) importEntityByDescription(host *engine.Host, proj *project.Project, parent *StageEntity, desc *stages.EntityDescription) (*StageEntity, error) {
	defer tracing.NewRegion("StageManager.importEntityByDescription").End()
	e := m.AddEntityWithId(desc.Id, desc.Name, matrix.Vec3Zero())
	e.StageData.Description = *desc
	if parent != nil {
		m.SetEntityParent(e, parent)
	}
	e.Transform.SetPosition(desc.Position)
	e.Transform.SetRotation(desc.Rotation)
	e.Transform.SetScale(desc.Scale)
	// TODO:  Setup all the other data for the entity
	if desc.Mesh != "" {
		m.spawnLoadedEntity(e, host, proj.FileSystem())
	}
	for i := range desc.DataBinding {
		db := &desc.DataBinding[i]
		// TODO:  Remove this in a week or so
		{
			switch db.RegistraionKey {
			case "kaiju.CameraDataBinding":
				db.RegistraionKey = "kaiju.CameraEntityData"
			case "kaiju.LightDataBinding":
				db.RegistraionKey = "kaiju.LightEntityData"
			case "kaiju.RigidBodyDataBinding":
				db.RegistraionKey = "kaiju.RigidBodyEntityData"
			}
		}
		g, ok := proj.EntityDataBinding(db.RegistraionKey)
		if !ok {
			slog.Error("failed to locate the data binding for entity",
				"key", db.RegistraionKey)
			continue
		}
		b := &entity_data_binding.EntityDataEntry{}
		b.ReadEntityDataBindingType(g)
		for k, v := range db.Fields {
			b.SetFieldByName(k, v)
		}
		e.AddDataBinding(b)
	}
	if e.StageData.ShaderData != nil {
		db := entity_data_binding.ToDataBinding("Shader data", e.StageData.ShaderData)
		for i := range desc.ShaderData {
			db.SetFieldByName(desc.ShaderData[i].Name, desc.ShaderData[i].Value)
		}
	}
	for i := range desc.Children {
		if _, err := m.importEntityByDescription(host, proj, e, &desc.Children[i]); err != nil {
			return e, err
		}
	}
	return e, nil
}

func (m *StageManager) spawnLoadedEntity(e *StageEntity, host *engine.Host, fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("StageManager.spawnLoadedEntity").End()
	const rootFolder = project_file_system.ContentFolder
	const meshFolder = project_file_system.ContentMeshFolder
	const texFolder = project_file_system.ContentTextureFolder
	desc := &e.StageData.Description
	meshId := desc.Mesh
	materialId := desc.Material
	textureIds := desc.Textures
	var km kaiju_mesh.KaijuMesh
	// TODO:  This is a hack, the quad/plane may need to be added to the stock
	// folder. They are special here because the're used for textures in 3D/2D.
	switch meshId {
	case "quad":
		km.Verts, km.Indexes = rendering.MeshQuadData()
	case "plane":
		km.Verts, km.Indexes = rendering.MeshPlaneData()
	default:
		kmData, err := fs.ReadFile(filepath.Join(rootFolder, meshFolder, meshId))
		if err != nil {
			slog.Error("failed to load the mesh data", "id", meshId, "error", err)
			return err
		}
		km, err = kaiju_mesh.Deserialize(kmData)
		if err != nil {
			slog.Error("failed to deserialize the mesh data", "id", meshId, "error", err)
			return err
		}
	}
	mesh := host.MeshCache().Mesh(meshId, km.Verts, km.Indexes)
	if materialId == "" {
		slog.Warn("no material provided for SpawnMesh, will use fallback material")
		materialId = assets.MaterialDefinitionBasic
	}
	mat, err := host.MaterialCache().Material(materialId)
	if err != nil {
		slog.Error("failed to create the standard material", "error", err)
		return err
	}
	texs := make([]*rendering.Texture, 0, len(textureIds))
	for i := range textureIds {
		path := filepath.Join(rootFolder, texFolder, textureIds[i])
		if _, err := fs.Stat(path); err != nil {
			path = filepath.Join(project_file_system.StockFolder, textureIds[i])
		}
		texData, err := fs.ReadFile(path)
		if err != nil {
			slog.Error("failed to read the texture file", "id", textureIds[i], "error", err)
			return err
		}
		// TODO:  Should be reading the filter from the configuration file
		tex, err := rendering.NewTextureFromMemory(textureIds[i],
			texData, 0, 0, rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the texture from it's data", "id", textureIds[i], "error", err)
			return err
		}
		texs = append(texs, tex)
	}
	// TODO:  This should be based on the rendering.MaterialData texture count
	if len(textureIds) == 0 {
		slog.Warn("missing textures for mesh, using a fallback one")
		tex, err := host.TextureCache().Texture(assets.TextureSquare,
			rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the default texture", "error", err)
		}
		texs = append(texs, tex)
	}
	mat = mat.CreateInstance(texs)
	e.StageData.ShaderData = shader_data_registry.Create(mat.Shader.ShaderDataName())
	// Temp set position to 0,0,0 for the BVH generation
	ePos := e.Transform.Position()
	e.Transform.SetPosition(matrix.Vec3Zero())
	e.StageData.Bvh = km.GenerateBVH(host.Threads(), &e.Transform, e)
	e.Transform.SetPosition(ePos)
	m.AddBVH(e.StageData.Bvh, &e.Transform)
	host.RunOnMainThread(func() {
		for i := range texs {
			texs[i].DelayedCreate(host.Window.Renderer)
		}
		draw := rendering.Drawing{
			Material:   mat,
			Mesh:       mesh,
			ShaderData: e.StageData.ShaderData,
			Transform:  &e.Transform,
			ViewCuller: &host.Cameras.Primary,
		}
		host.Drawings.AddDrawing(draw)
		e.OnDestroy.Add(func() { e.StageData.ShaderData.Destroy() })
	})
	return nil
}

func (m *StageManager) updateExistingTemplateInstances(skip *StageEntity, host *engine.Host, proj *project.Project, templateId string) error {
	defer tracing.NewRegion("StageManager.updateExistingTemplateInstances").End()
	if templateId == "" {
		return nil
	}
	tpl, err := proj.ReadEntityTemplate(templateId)
	if err != nil {
		slog.Error("failed to read the template file", "error", err)
		return err
	}
	m.ClearSelection()
	var generateId func(d *stages.EntityDescription)
	generateId = func(d *stages.EntityDescription) {
		d.Id = uuid.NewString()
		for i := range d.Children {
			generateId(&d.Children[i])
		}
	}
	for i := range m.entities {
		if m.entities[i].StageData.Description.TemplateId != templateId {
			continue
		}
		if m.entities[i] == skip {
			continue
		}
		cpy := tpl
		generateId(&cpy)
		t := m.entities[i].Transform
		m.OnEntityDestroy.Execute(m.entities[i])
		m.entities[i].Destroy()
		e, err := m.importEntityByDescription(host, proj, nil, &cpy)
		if err != nil {
			return err
		}
		e.Transform.Copy(t)
	}
	m.worldBVH.Refit()
	m.history.Clear()
	return nil
}

func (m *StageManager) RefitBVH(entity *StageEntity) {
	// TODO:  It's getting a little late, but I may need to track all of the
	// nodes that were related to each other when they were created and only
	// update the matching ones here. For now I'm just going to update the whole
	// tree before it gets too late.
	m.RefitWorldBVH()
}

func explodeEntityHierarchy(e *StageEntity) []*StageEntity {
	defer tracing.NewRegion("editor_stage_manager.explodeEntityHierarchy").End()
	all := []*StageEntity{}
	var explode func(p *StageEntity)
	explode = func(p *StageEntity) {
		all = append(all, p)
		for _, c := range p.Children {
			explode(EntityToStageEntity(c))
		}
	}
	explode(e)
	return all
}

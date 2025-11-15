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

package editor_stage_manager

import (
	"encoding/json"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/collision"
	"kaiju/engine/systems/events"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"kaiju/stages"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"weak"

	"github.com/KaijuEngine/uuid"
)

const StageIdPrefix = "stage_"

// StageManager represents the current stage in the editor. It contains all of
// the entities on the stage.
type StageManager struct {
	OnEntitySpawn         events.EventWithArg[*StageEntity]
	OnEntityDestroy       events.EventWithArg[*StageEntity]
	OnEntitySelected      events.EventWithArg[*StageEntity]
	OnEntityDeselected    events.EventWithArg[*StageEntity]
	OnEntityChangedParent events.EventWithArg[*StageEntity]
	stageId               string
	host                  *engine.Host
	history               *memento.History
	entities              []*StageEntity
	selected              []*StageEntity
	worldBVH              *collision.BVH
}

// StageEntityEditorData is the structure holding all the uniquely identifiable
// and linking data about the entity on this stage. That will include things
// like content linkage, data bindings, etc.
type StageEntityEditorData struct {
	Bvh         *collision.BVH
	Mesh        *rendering.Mesh
	ShaderData  rendering.DrawInstance
	Description stages.EntityDescription
}

func (m *StageManager) Initialize(host *engine.Host, history *memento.History) {
	m.host = host
	m.history = history
}

func (m *StageManager) NewStage() {
	defer tracing.NewRegion("StageManager.NewStage").End()
	m.Clear()
}

func (m *StageManager) IsNew() bool     { return m.stageId == "" }
func (m *StageManager) StageId() string { return m.stageId }

func (m *StageManager) SetStageId(id string, cache *content_database.Cache) error {
	defer tracing.NewRegion("StageManager.SetStageId").End()
	newId := StageIdPrefix + id
	if _, err := cache.Read(newId); err == nil {
		return StageAlreadyExistsError{id}
	}
	m.stageId = newId
	return nil
}

// List will return all of the internally held entities for the stage
func (m *StageManager) List() []*StageEntity { return m.entities }

func (m *StageManager) Selection() []*StageEntity { return m.selected }

// AddEntity will generate a new entity for the stage with a new random Id. It
// will internally just call #AddEntityWithId
func (m *StageManager) AddEntity(name string, point matrix.Vec3) *StageEntity {
	defer tracing.NewRegion("StageManager.AddEntity").End()
	e := m.AddEntityWithId(uuid.NewString(), name, point)
	m.history.Add(&objectSpawnHistory{
		m: m,
		e: e,
	})
	return e
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
		se := we.Value()
		for i := range sm.entities {
			if sm.entities[i] == se {
				sm.entities = klib.RemoveUnordered(sm.entities, i)
				return
			}
		}
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

func (m *StageManager) HierarchyRespectiveSelection() []*StageEntity {
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
	for i := range m.entities {
		m.entities[i].Destroy()
	}
	m.worldBVH = nil
}

func (m *StageManager) AddBVH(bvh *collision.BVH, transform *matrix.Transform) {
	defer tracing.NewRegion("StageManager.AddBVH").End()
	cpy := collision.CloneBVH(bvh)
	collision.AddSubBVH(&m.worldBVH, cpy, transform)
}

func (m *StageManager) RemoveBVH(bvh *collision.BVH) {
	defer tracing.NewRegion("StageManager.RemoveBVH").End()
	collision.RemoveSubBVH(&m.worldBVH, bvh)
}

func (m *StageManager) toStage() stages.Stage {
	defer tracing.NewRegion("StageManager.toStage").End()
	s := stages.Stage{Id: m.stageId}
	rootCount := 0
	for i := range m.entities {
		if m.entities[i].IsRoot() {
			rootCount++
		}
	}
	s.Entities = make([]stages.EntityDescription, 0, rootCount)
	var readEntity func(parent *StageEntity)
	readEntity = func(parent *StageEntity) {
		desc := &parent.StageData.Description
		desc.Name = parent.Name()
		desc.Position = parent.Transform.Position()
		desc.Rotation = parent.Transform.Rotation()
		desc.Scale = parent.Transform.Scale()
		desc.DataBinding = make([]stages.EntityDataBinding, 0, len(parent.dataBindings))
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
		for i := range m.entities {
			if m.entities[i].Parent == &parent.Entity {
				readEntity(m.entities[i])
				desc.Children = append(desc.Children, m.entities[i].StageData.Description)
			}
		}
	}
	for i := range m.entities {
		if m.entities[i].IsRoot() {
			readEntity(m.entities[i])
			s.Entities = append(s.Entities, m.entities[i].StageData.Description)
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
	configPath := filepath.Join(project_file_system.ContentConfigFolder,
		project_file_system.ContentStageFolder, m.stageId)
	cfg := content_database.ContentConfig{}
	cfg.Name = strings.TrimPrefix(m.stageId, StageIdPrefix)
	cfg.Type = content_database.Stage{}.TypeName()
	f2, err := fs.Create(configPath)
	if err != nil {
		return err
	}
	defer f2.Close()
	if err := json.NewEncoder(f2).Encode(cfg); err != nil {
		// TODO:  Roll back
		return err
	}
	if err := cache.Index(configPath, fs); err != nil {
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
	var importTarget func(parent *StageEntity, desc *stages.EntityDescription) error
	importTarget = func(parent *StageEntity, desc *stages.EntityDescription) error {
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
			m.spawnLoadedEntity(e, host, fs)
		}
		for i := range desc.DataBinding {
			db := &desc.DataBinding[i]
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
		for i := range desc.Children {
			if err := importTarget(e, &desc.Children[i]); err != nil {
				return err
			}
		}
		return nil
	}
	for i := range s.Entities {
		if err := importTarget(nil, &s.Entities[i]); err != nil {
			return err
		}
	}
	m.stageId = id
	return nil
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
	kmData, err := fs.ReadFile(filepath.Join(rootFolder, meshFolder, meshId))
	if err != nil {
		slog.Error("failed to load the mesh data", "id", meshId, "error", err)
		return err
	}
	km, err := kaiju_mesh.Deserialize(kmData)
	if err != nil {
		slog.Error("failed to deserialize the mesh data", "id", meshId, "error", err)
		return err
	}
	mesh := host.MeshCache().Mesh(meshId, km.Verts, km.Indexes)
	var mat *rendering.Material
	if materialId == "" {
		slog.Warn("no material provided for SpawnMesh, will use fallback material")
		materialId = assets.MaterialDefinitionBasic
	}
	mat, err = host.MaterialCache().Material(materialId)
	if err != nil {
		slog.Error("failed to create the standard material", "error", err)
		return err
	}
	texs := make([]*rendering.Texture, 0, len(textureIds))
	for i := range textureIds {
		texData, err := fs.ReadFile(filepath.Join(rootFolder, texFolder, textureIds[i]))
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
	e.StageData.Bvh = km.GenerateBVH(host.Threads(), &e.Transform, e)
	host.RunOnMainThread(func() {
		for i := range texs {
			texs[i].DelayedCreate(host.Window.Renderer)
		}
		draw := rendering.Drawing{
			Renderer:   host.Window.Renderer,
			Material:   mat,
			Mesh:       mesh,
			ShaderData: e.StageData.ShaderData,
			Transform:  &e.Transform,
		}
		host.Drawings.AddDrawing(draw)
		e.OnDestroy.Add(func() { e.StageData.ShaderData.Destroy() })
	})
	return nil
}

func explodeEntityHierarchy(e *StageEntity) []*StageEntity {
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

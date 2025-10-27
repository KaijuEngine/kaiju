package editor_stage_manager

import (
	"encoding/json"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/collision"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"kaiju/stages"
	"log/slog"
	"path/filepath"
	"weak"

	"github.com/KaijuEngine/uuid"
)

// StageManager represents the current stage in the editor. It contains all of
// the entities on the stage.
type StageManager struct {
	stageId  string
	host     *engine.Host
	entities []*StageEntity
	selected []*StageEntity
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

func (m *StageManager) Initialize(host *engine.Host) { m.host = host }

func (m *StageManager) NewStage() {
	defer tracing.NewRegion("StageManager.NewStage").End()
	// TODO:  Show a popup to save the current stage if there are changes
	m.stageId = uuid.NewString()
}

// List will return all of the internally held entities for the stage
func (m *StageManager) List() []*StageEntity { return m.entities }

func (m *StageManager) Selection() []*StageEntity { return m.selected }

// AddEntity will create a new entity for the stage. This entity will have a
// #StageEntityData automatically added to it as named data named "stage".
func (m *StageManager) AddEntity(point matrix.Vec3) *StageEntity {
	defer tracing.NewRegion("StageManager.AddEntity").End()
	e := &StageEntity{}
	e.Init(m.host.WorkGroup())
	e.StageData.Description.Id = uuid.NewString()
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
	return e
}

// Clear will destroy all entities that are managed by this stage manager.
func (m *StageManager) Clear() {
	defer tracing.NewRegion("StageManager.Clear").End()
	for i := range m.entities {
		m.entities[i].Destroy()
	}
}

func (m *StageManager) SaveStage(cache *content_database.Cache, fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("StageManager.SaveStage").End()
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
		desc.Position = parent.Transform.Position()
		desc.Rotation = parent.Transform.Rotation()
		desc.Scale = parent.Transform.Scale()
		for i := range m.entities {
			if m.entities[i].Parent == &parent.Entity {
				desc.Children = append(desc.Children, m.entities[i].StageData.Description)
				readEntity(m.entities[i])
			}
		}
	}
	for i := range m.entities {
		if m.entities[i].IsRoot() {
			readEntity(m.entities[i])
			s.Entities = append(s.Entities, m.entities[i].StageData.Description)
		}
	}
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
	cfg.Name = "Testing"
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

func (m *StageManager) LoadStage(id string, host *engine.Host, cache *content_database.Cache, fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("StageManager.LoadStage").End()
	// TODO:  Show a popup to save the current stage if there are changes
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
		e := m.AddEntity(matrix.Vec3Zero())
		e.StageData.Description = *desc
		e.SetParent(&parent.Entity)
		e.Transform.SetPosition(desc.Position)
		e.Transform.SetRotation(desc.Rotation)
		e.Transform.SetScale(desc.Scale)
		// TODO:  Setup all the other data for the entity
		if desc.Mesh != "" {
			m.spawnLoadedEntity(e, host, fs)
		}
		// TODO:  Setup any of the data bindings

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
	return nil
}

func (m *StageManager) spawnLoadedEntity(e *StageEntity, host *engine.Host, fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("StageManager.spawnLoadedEntity").End()
	const rootFolder = project_file_system.ContentFolder
	const meshFolder = project_file_system.ContentMeshFolder
	const matFolder = project_file_system.ContentMaterialFolder
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
		mat, err = host.MaterialCache().Material(assets.MaterialDefinitionBasic)
		if err != nil {
			slog.Error("failed to create the standard material", "error", err)
			return err
		}
	} else {
		// TODO:
		//matData, err := w.pfs.ReadFile(filepath.Join(rootFolder, matFolder, materialId))
		//if err != nil {
		//	slog.Error("failed to load the material data", "id", meshId, "error", err)
		//	return err
		//}
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

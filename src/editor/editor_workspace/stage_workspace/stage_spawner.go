package stage_workspace

import (
	"encoding/json"
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"log/slog"
)

func (w *Workspace) spawnContent(cc *content_database.CachedContent, m *hid.Mouse) {
	defer tracing.NewRegion("StageWorkspace.spawnContent").End()
	cat, ok := content_database.CategoryFromTypeName(cc.Config.Type)
	if !ok {
		slog.Error("failed to find the content category for type",
			"id", cc.Id(), "type", cc.Config.Type)
		return
	}
	ray := w.Host.Camera.RayCast(m.Position())
	e, eHitOk := w.manager.TryHitEntity(ray)
	// TODO:  Find the point on the entity that was hit, otherwise fall back
	// to doing the ground plane/distance hit point
	hit, ok := ray.PlaneHit(matrix.Vec3Zero(), matrix.Vec3Up())
	if !ok {
		hit = ray.Point(maxContentDropDistance)
	}
	switch cat.(type) {
	case content_database.Texture:
		w.spawnTexture(cc, hit)
	case content_database.Mesh:
		w.spawnMesh(cc, hit)
	case content_database.Material:
		if eHitOk {
			w.attachMaterial(cc, e)
		}
	default:
		slog.Error("dropping this type of content into the stage is not supported",
			"id", cc.Id(), "type", cc.Config.Type)
	}
}

func (w *Workspace) spawnTexture(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnTexture").End()
	mat, err := w.Host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		slog.Error("failed to find the basic material", "error", err)
		return
	}
	path := content_database.ToContentPath(cc.Path)
	data, err := w.pfs.ReadFile(path)
	if err != nil {
		slog.Error("error reading the image file", "path", path)
		return
	}
	tex, err := rendering.NewTextureFromMemory(rendering.GenerateUniqueTextureKey,
		data, 0, 0, rendering.TextureFilterLinear)
	if err != nil {
		slog.Error("failed to create the texture resource", "id", cc.Id(), "error", err)
		return
	}
	mat = mat.CreateInstance([]*rendering.Texture{tex})
	e := w.manager.AddEntity(point)
	e.StageData.Mesh = rendering.NewMeshPlane(w.Host.MeshCache())
	e.StageData.Description.Mesh = e.StageData.Mesh.Key()
	e.StageData.Description.Textures = []string{cc.Id()}
	e.StageData.ShaderData = &rendering.ShaderDataStandard{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	w.Host.RunOnMainThread(func() {
		tex.DelayedCreate(w.Host.Window.Renderer)
		draw := rendering.Drawing{
			Renderer:   w.Host.Window.Renderer,
			Material:   mat,
			Mesh:       e.StageData.Mesh,
			ShaderData: e.StageData.ShaderData,
			Transform:  &e.Transform,
		}
		w.Host.Drawings.AddDrawing(draw)
	})
}

func (w *Workspace) spawnMesh(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnMesh").End()
	mat, err := w.Host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		slog.Error("failed to find the basic material", "error", err)
		return
	}
	path := content_database.ToContentPath(cc.Path)
	data, err := w.pfs.ReadFile(path)
	if err != nil {
		slog.Error("error reading the mesh file", "path", path)
		return
	}
	km, err := kaiju_mesh.Deserialize(data)
	if err != nil {
		slog.Error("failed to deserialize the mesh", "id", cc.Id(), "error", err)
		return
	}
	tex, _ := w.Host.TextureCache().Texture(assets.TextureSquare,
		rendering.TextureFilterLinear)
	mat = mat.CreateInstance([]*rendering.Texture{tex})
	e := w.manager.AddEntity(point)
	e.StageData.Mesh = w.Host.MeshCache().Mesh(cc.Id(), km.Verts, km.Indexes)
	e.StageData.Description.Mesh = e.StageData.Mesh.Key()
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads())
	e.StageData.ShaderData = &rendering.ShaderDataStandard{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	draw := rendering.Drawing{
		Renderer:   w.Host.Window.Renderer,
		Material:   mat,
		Mesh:       e.StageData.Mesh,
		ShaderData: e.StageData.ShaderData,
		Transform:  &e.Transform,
	}
	w.Host.Drawings.AddDrawing(draw)
	e.OnDestroy.Add(func() { e.StageData.ShaderData.Destroy() })
}

func (w *Workspace) attachMaterial(cc *content_database.CachedContent, e *editor_stage_manager.StageEntity) {
	mat, ok := w.Host.MaterialCache().FindMaterial(cc.Id())
	if !ok {
		path := content_database.ToContentPath(cc.Path)
		f, err := w.pfs.Open(path)
		if err != nil {
			slog.Error("error reading the mesh file", "path", path)
			return
		}
		defer f.Close()
		var matData rendering.MaterialData
		if err = json.NewDecoder(f).Decode(&matData); err != nil {
			slog.Error("failed to decode the material", "id", cc.Id(), "name", cc.Config.Name)
			return
		}
		mat, err = matData.Compile(w.Host.AssetDatabase(), w.Host.Window.Renderer)
		if err != nil {
			slog.Error("failed to compile the material", "id", cc.Id(), "name", cc.Config.Name, "error", err)
			return
		}
		mat = w.Host.MaterialCache().AddMaterial(mat)
	}
	e.SetMaterial(mat.CreateInstance(mat.Textures), w.Host)
}

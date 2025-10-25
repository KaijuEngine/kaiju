package stage_workspace

import (
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"log/slog"
)

func (w *Workspace) spawnContent(cc *content_database.CachedContent, m *hid.Mouse) {
	// TODO:  Spawn the content in the viewport
	cat, ok := content_database.CategoryFromTypeName(cc.Config.Type)
	if !ok {
		slog.Error("failed to find the content category for type",
			"id", cc.Id(), "type", cc.Config.Type)
		return
	}
	ray := w.Host.Camera.RayCast(m.Position())
	// TODO:  Try to hit something else on the stage, otherwise fall back to the
	// ground plane hit test
	hit, ok := ray.PlaneHit(matrix.Vec3Zero(), matrix.Vec3Up())
	if !ok {
		hit = ray.Point(maxContentDropDistance)
	}
	switch cat.(type) {
	case content_database.Texture:
		w.spawnTexture(cc, hit)
	case content_database.Mesh:
		w.spawnMesh(cc, hit)
	default:
		slog.Error("dropping this type of content into the stage is not supported",
			"id", cc.Id(), "type", cc.Config.Type)
	}
}

func (w *Workspace) spawnTexture(cc *content_database.CachedContent, point matrix.Vec3) {
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
	mesh := rendering.NewMeshPlane(w.Host.MeshCache())
	e, esd := w.manager.AddEntity(point)
	esd.Rendering.MeshId = mesh.Key()
	esd.Rendering.TextureIds = []string{cc.Id()}
	esd.Rendering.ShaderData = &rendering.ShaderDataStandard{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	w.Host.RunOnMainThread(func() {
		tex.DelayedCreate(w.Host.Window.Renderer)
		draw := rendering.Drawing{
			Renderer:   w.Host.Window.Renderer,
			Material:   mat,
			Mesh:       mesh,
			ShaderData: esd.Rendering.ShaderData,
			Transform:  &e.Transform,
		}
		w.Host.Drawings.AddDrawing(draw)
	})
}

func (w *Workspace) spawnMesh(cc *content_database.CachedContent, point matrix.Vec3) {
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
	km, err := kaiju_mesh.Deserialize(data)
	if err != nil {
		slog.Error("failed to create the texture resource", "id", cc.Id(), "error", err)
		return
	}
	mesh := w.Host.MeshCache().Mesh(cc.Id(), km.Verts, km.Indexes)
	tex, _ := w.Host.TextureCache().Texture(assets.TextureSquare,
		rendering.TextureFilterLinear)
	mat = mat.CreateInstance([]*rendering.Texture{tex})
	e, esd := w.manager.AddEntity(point)
	esd.Rendering.MeshId = mesh.Key()
	esd.Rendering.TextureIds = []string{cc.Id()}
	esd.Bvh = km.GenerateBVH(w.Host.Threads())
	sd := &rendering.ShaderDataStandard{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	sd.SetFlag(rendering.ShaderDataStandardFlagOutline)
	draw := rendering.Drawing{
		Renderer:   w.Host.Window.Renderer,
		Material:   mat,
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &e.Transform,
	}
	w.Host.Drawings.AddDrawing(draw)
	e.OnDestroy.Add(func() { sd.Destroy() })
}

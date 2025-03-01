package content_opener

import (
	"kaiju/assets"
	"kaiju/assets/asset_info"
	"kaiju/cache/project_cache"
	"kaiju/collision"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

func loadMesh(host *engine.Host, adi asset_info.AssetDatabaseInfo, e *engine.Entity, bvh *collision.BVH) error {
	texId := assets.TextureSquare
	if t, ok := adi.Metadata["texture"]; ok {
		texId = t
	}
	tex, err := host.TextureCache().Texture(texId, rendering.TextureFilterLinear)
	if err != nil {
		return err
	}
	var data rendering.DrawInstance
	var shader *rendering.Shader
	if s, ok := adi.Metadata["shader"]; ok {
		shader = host.ShaderCache().ShaderFromDefinition(s)
		// TODO:  We need to create or generate shader data given the definition
		data = &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	} else {
		shader = host.ShaderCache().ShaderFromDefinition(
			assets.ShaderDefinitionBasic)
		data = &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}
	mesh, ok := host.MeshCache().FindMesh(adi.ID)
	if !ok {
		m, err := project_cache.LoadCachedMesh(adi)
		if err != nil {
			return err
		}
		mesh = rendering.NewMesh(adi.ID, m.Verts, m.Indexes)
		bvh.Insert(m.GenerateBVH(&e.Transform))
	}
	host.MeshCache().AddMesh(mesh)
	drawing := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Shader:     shader,
		Mesh:       mesh,
		Textures:   []*rendering.Texture{tex},
		ShaderData: data,
		Transform:  &e.Transform,
		CanvasId:   "default",
	}
	host.Drawings.AddDrawing(&drawing)
	e.EditorBindings.AddDrawing(drawing)
	e.OnActivate.Add(func() { data.Activate() })
	e.OnDeactivate.Add(func() { data.Deactivate() })
	e.OnDestroy.Add(func() { data.Destroy() })
	return nil
}

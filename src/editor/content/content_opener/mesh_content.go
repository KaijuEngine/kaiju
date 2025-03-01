package content_opener

import (
	"kaiju/assets/asset_info"
	"kaiju/cache/project_cache"
	"kaiju/collision"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

func loadMesh(host *engine.Host, adi asset_info.AssetDatabaseInfo, e *engine.Entity, bvh *collision.BVH) error {
	var err error
	var data rendering.DrawInstance
	var material *rendering.Material
	if s, ok := adi.Metadata["material"]; ok {
		if material, err = host.MaterialCache().Material(s); err != nil {
			return err
		}
		// TODO:  We need to create or generate shader data given the definition
		data = &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	} else {
		if material, err = host.MaterialCache().Material("basic"); err != nil {
			return err
		}
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
		Material:   material,
		Mesh:       mesh,
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

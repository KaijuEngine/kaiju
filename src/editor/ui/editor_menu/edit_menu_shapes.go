package editor_menu

import (
	"kaiju/assets"
	"kaiju/editor/interfaces"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders"
	"log/slog"
)

const (
	cubeGLB = "editor/meshes/cube.glb"
)

func createShape(name, glb string, ed interfaces.Editor, host *engine.Host) {
	res, err := loaders.GLTF(glb, host.AssetDatabase(), host.WorkGroup())
	if err != nil {
		slog.Error("failed to load the cube mesh", "error", err.Error())
		return
	} else if !res.IsValid() || len(res.Meshes) != 1 {
		slog.Error("cube mesh data corrupted")
		return
	}
	resMesh := res.Meshes[0]
	mesh := rendering.NewMesh(resMesh.MeshName, resMesh.Verts, resMesh.Indexes)
	host.MeshCache().AddMesh(mesh)
	e := ed.CreateEntity(name)
	sd := rendering.ShaderDataBasic{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	drawing := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Shader:     host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionBasic),
		Mesh:       mesh,
		Textures:   []*rendering.Texture{tex},
		ShaderData: &sd,
		Transform:  &e.Transform,
		CanvasId:   "default",
	}
	host.Drawings.AddDrawing(&drawing)
	e.EditorBindings.AddDrawing(drawing)
	bvh := resMesh.GenerateBVH(&e.Transform)
	e.EditorBindings.Set("bvh", bvh)
	ed.BVH().Insert(bvh)
	e.OnDestroy.Add(func() { bvh.RemoveNode() })
}

func (m *Menu) createCube() {
	createShape("Cube", cubeGLB, m.editor, m.container.Host)
}

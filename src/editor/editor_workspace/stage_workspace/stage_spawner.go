/******************************************************************************/
/* stage_spawner.go                                                           */
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

package stage_workspace

import (
	"encoding/json"
	"fmt"
	"kaiju/editor/codegen"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/editor_stage_manager/data_binding_renderer"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine/assets"
	"kaiju/engine_data_bindings/engine_data_binding_camera"
	"kaiju/engine_data_bindings/engine_data_binding_light"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/ollama"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"log/slog"
	"weak"
)

func (w *StageWorkspace) initAIActions() {
	a := OllamaStageWorkspaceAction{weak.Make(w)}
	ollama.ReflectFuncToOllama(a.createCamera,
		"CreateCamera", "Creates a new camera at the given position with the given look at point.",
		"posX", "The X position to spawn",
		"posY", "The Y position to spawn",
		"posZ", "The Z position to spawn",
		"lookAtX", "The X position to look at",
		"lookAtY", "The Y position to look at",
		"lookAtZ", "The Z position to look at",
	)
}

type OllamaStageWorkspaceAction struct {
	w weak.Pointer[StageWorkspace]
}

func (a OllamaStageWorkspaceAction) createCamera(posX, posY, posZ, lookAtX, lookAtY, lookAtZ float32) string {
	defer tracing.NewRegion("StageWorkspace.createCamera").End()
	w := a.w.Value()
	cam, ok := w.CreateNewCamera()
	if !ok {
		return "failed to create the camera for some reason"
	}
	cam.Transform.SetPosition(matrix.NewVec3(posX, posY, posZ))
	cam.Transform.LookAt(matrix.NewVec3(lookAtX, lookAtY, lookAtZ))
	return fmt.Sprintf("camera created at <%s, %s, %s> looking at <%s, %s, %s>",
		klib.FormatFloatToNDecimals(posX, 3),
		klib.FormatFloatToNDecimals(posY, 3),
		klib.FormatFloatToNDecimals(posZ, 3),
		klib.FormatFloatToNDecimals(lookAtX, 3),
		klib.FormatFloatToNDecimals(lookAtY, 3),
		klib.FormatFloatToNDecimals(lookAtZ, 3))
}

// Create 5 cameras in a circle, in the air, around center stage. These cameras should be looking at center stage.

func (w *StageWorkspace) attachEntityData(e *editor_stage_manager.StageEntity, g codegen.GeneratedType) *entity_data_binding.EntityDataEntry {
	defer tracing.NewRegion("StageWorkspace.attachEntityData").End()
	de := &entity_data_binding.EntityDataEntry{}
	e.AddDataBinding(de.ReadEntityDataBindingType(g))
	data_binding_renderer.Attached(de, weak.Make(w.Host), w.stageView.Manager(), e)
	return de
}

func (w *StageWorkspace) CreateNewCamera() (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.CreateNewCamera").End()
	return w.createDataBoundEntity("Camera", engine_data_binding_camera.BindingKey)
}

func (w *StageWorkspace) CreateNewEntity() (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.CreateNewCamera").End()
	e := w.stageView.Manager().AddEntity("Entity", w.stageView.LookAtPoint())
	return e, true
}

func (w *StageWorkspace) CreateNewLight() (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.CreateNewLight").End()
	return w.createDataBoundEntity("Light", engine_data_binding_light.BindingKey)
}

func (w *StageWorkspace) createDataBoundEntity(name, bindKey string) (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.createDataBoundEntity").End()
	e := w.stageView.Manager().AddEntity(name, w.stageView.LookAtPoint())
	g, ok := w.ed.Project().EntityDataBinding(bindKey)
	if !ok {
		slog.Error("failed to locate the entity binding data", "key", bindKey)
		return nil, false
	}
	w.attachEntityData(e, g)
	return e, true
}

func (w *StageWorkspace) spawnContentAtMouse(cc *content_database.CachedContent, m *hid.Mouse) {
	defer tracing.NewRegion("StageWorkspace.spawnContent").End()
	var mp matrix.Vec2
	if w.stageView.IsView3D() {
		mp = m.Position()
	} else {
		mp = m.ScreenPosition()
	}
	ray := w.Host.Camera.RayCast(mp)
	e, eHitOk := w.stageView.Manager().TryHitEntity(ray)
	// TODO:  Find the point on the entity that was hit, otherwise fall back
	// to doing the ground plane/distance hit point
	var hit matrix.Vec3
	var ok bool
	if w.stageView.IsView3D() {
		hit, ok = ray.PlaneHit(matrix.Vec3Zero(), matrix.Vec3Up())
	} else {
		hit, ok = ray.PlaneHit(matrix.Vec3Zero(), matrix.Vec3Forward())
	}
	if !ok {
		hit = ray.Point(maxContentDropDistance)
	}
	cat, ok := content_database.CategoryFromTypeName(cc.Config.Type)
	if !ok {
		slog.Error("failed to find the content category for type",
			"id", cc.Id(), "type", cc.Config.Type)
		return
	}
	switch cat.(type) {
	case content_database.Texture:
		w.spawnTexture(cc, hit)
	case content_database.Mesh:
		w.spawnMesh(cc, hit)
	case content_database.Template:
		w.stageView.Manager().SpawnTemplate(w.Host, w.ed.Project(), cc, hit)
	case content_database.Material:
		if eHitOk {
			w.attachMaterial(cc, e)
		}
	default:
		slog.Error("dropping this type of content into the stage is not supported",
			"id", cc.Id(), "type", cc.Config.Type)
	}
}

func (w *StageWorkspace) spawnContentAtPosition(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnContentAtPosition").End()
	cat, ok := content_database.CategoryFromTypeName(cc.Config.Type)
	if !ok {
		slog.Error("failed to find the content category for type",
			"id", cc.Id(), "type", cc.Config.Type)
		return
	}
	switch cat.(type) {
	case content_database.Texture:
		w.spawnTexture(cc, point)
	case content_database.Mesh:
		w.spawnMesh(cc, point)
	case content_database.Stage:
		if w.ed.History().HasPendingChanges() {
			w.ed.BlurInterface()
			confirm_prompt.Show(w.Host, confirm_prompt.Config{
				Title:       "Discrad changes",
				Description: "You have unsaved changes to your stage, would you like to discard them and load the selected stage?",
				ConfirmText: "Yes",
				CancelText:  "No",
				OnConfirm: func() {
					w.ed.FocusInterface()
					w.loadStage(cc.Id())
				},
				OnCancel: func() { w.ed.FocusInterface() },
			})
		} else {
			w.loadStage(cc.Id())
		}
	default:
		slog.Error("double clicking this type of content is not supported",
			"id", cc.Id(), "type", cc.Config.Type)
	}
}

func (w *StageWorkspace) loadStage(id string) {
	defer tracing.NewRegion("StageWorkspace.loadStage").End()
	man := w.stageView.Manager()
	if err := man.LoadStage(id, w.Host, w.ed.Cache(), w.ed.Project()); err != nil {
		slog.Error("failed to load the stage", "id", id, "error", err)
	} else {
		for _, e := range man.List() {
			for _, b := range e.DataBindings() {
				data_binding_renderer.Attached(b, weak.Make(w.Host), man, e)
			}
		}
	}
}

func (w *StageWorkspace) spawnTexture(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnTexture").End()
	mat, err := w.Host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		slog.Error("failed to find the basic material", "error", err)
		return
	}
	path := content_database.ToContentPath(cc.Path)
	data, err := w.ed.ProjectFileSystem().ReadFile(path)
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
	man := w.stageView.Manager()
	e := man.AddEntity(cc.Config.Name, point)
	var km kaiju_mesh.KaijuMesh
	if w.stageView.IsView3D() {
		e.StageData.Mesh = rendering.NewMeshPlane(w.Host.MeshCache())
		km.Verts, km.Indexes = rendering.MeshPlaneData()
	} else {
		e.StageData.Mesh = rendering.NewMeshQuad(w.Host.MeshCache())
		km.Verts, km.Indexes = rendering.MeshQuadData()
	}
	e.StageData.Description.Mesh = e.StageData.Mesh.Key()
	// Not using mat.Id here due to the material being assets.MaterialDefinitionBasic
	e.StageData.Description.Material = mat.Name
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads(), &e.Transform, e)
	man.AddBVH(e.StageData.Bvh, &e.Transform)
	e.StageData.Description.Textures = []string{cc.Id()}
	e.StageData.ShaderData = &shader_data_registry.ShaderDataStandard{
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

func (w *StageWorkspace) spawnMesh(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnMesh").End()
	mat, err := w.Host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		slog.Error("failed to find the basic material", "error", err)
		return
	}
	path := content_database.ToContentPath(cc.Path)
	data, err := w.ed.ProjectFileSystem().ReadFile(path)
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
	man := w.stageView.Manager()
	e := man.AddEntity(cc.Config.Name, point)
	e.StageData.Mesh = w.Host.MeshCache().Mesh(cc.Id(), km.Verts, km.Indexes)
	e.StageData.Description.Mesh = e.StageData.Mesh.Key()
	e.StageData.Description.Material = mat.Id
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads(), &e.Transform, e)
	man.AddBVH(e.StageData.Bvh, &e.Transform)
	e.StageData.ShaderData = &shader_data_registry.ShaderDataStandard{
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

func (w *StageWorkspace) attachMaterial(cc *content_database.CachedContent, e *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("StageWorkspace.attachMaterial").End()
	mat, ok := w.Host.MaterialCache().FindMaterial(cc.Id())
	if !ok {
		path := content_database.ToContentPath(cc.Path)
		f, err := w.ed.ProjectFileSystem().Open(path)
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
		mat.Id = cc.Id()
		mat = w.Host.MaterialCache().AddMaterial(mat)
	}
	e.SetMaterial(mat.CreateInstance(mat.Textures), w.stageView.Manager())
}

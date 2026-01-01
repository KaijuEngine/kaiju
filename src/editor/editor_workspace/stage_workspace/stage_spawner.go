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

package stage_workspace

import (
	"kaiju/editor/codegen"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/editor_stage_manager/data_binding_renderer"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine/assets"
	"kaiju/engine_entity_data/engine_entity_data_camera"
	"kaiju/engine_entity_data/engine_entity_data_light"
	"kaiju/engine_entity_data/engine_entity_data_particles"
	"kaiju/framework"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"kaiju/rendering/loaders/load_result"
	"log/slog"
	"weak"
)

func (w *StageWorkspace) attachEntityData(e *editor_stage_manager.StageEntity, g codegen.GeneratedType) *entity_data_binding.EntityDataEntry {
	defer tracing.NewRegion("StageWorkspace.attachEntityData").End()
	m := w.stageView.Manager()
	de := m.AttachEntityData(e, g)
	data_binding_renderer.Attached(de, weak.Make(w.Host), m, e)
	return de
}

func (w *StageWorkspace) CreateNewCamera() (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.CreateNewCamera").End()
	cam, ok := w.createDataBoundEntity("Camera", engine_entity_data_camera.BindingKey)
	if ok {
		shouldMakePrimary := true
		for _, e := range w.stageView.Manager().List() {
			if e == cam {
				continue
			}
			if len(e.DataBindingsByKey(engine_entity_data_camera.BindingKey)) > 0 {
				shouldMakePrimary = false
				break
			}
		}
		db := cam.DataBindingsByKey(engine_entity_data_camera.BindingKey)
		if shouldMakePrimary && len(db) > 0 {
			db[0].SetFieldByName("IsMainCamera", true)
		}
	}
	return cam, ok
}

func (w *StageWorkspace) CreateNewEntity() (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.CreateNewCamera").End()
	e := w.stageView.Manager().AddEntity("Entity", w.stageView.LookAtPoint())
	return e, true
}

func (w *StageWorkspace) CreateNewLight() (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.CreateNewLight").End()
	return w.createDataBoundEntity("Light", engine_entity_data_light.BindingKey)
}

func (w *StageWorkspace) createDataBoundEntity(name, bindKey string) (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.createDataBoundEntity").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	man := w.stageView.Manager()
	e := man.AddEntity(name, w.stageView.LookAtPoint())
	g, ok := w.ed.Project().EntityDataBinding(bindKey)
	if !ok {
		slog.Error("failed to locate the entity binding data", "key", bindKey)
		return nil, false
	}
	w.attachEntityData(e, g)
	man.ClearSelection()
	man.SelectEntity(e)
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
	ray := w.Host.PrimaryCamera().RayCast(mp)
	e, hit, eHitOk := w.stageView.Manager().TryHitEntity(ray)
	// TODO:  Find the point on the entity that was hit, otherwise fall back
	// to doing the ground plane/distance hit point
	if !eHitOk {
		var ok bool
		if w.stageView.IsView3D() {
			hit, ok = ray.PlaneHit(matrix.Vec3Zero(), matrix.Vec3Up())
		} else {
			hit, ok = ray.PlaneHit(matrix.Vec3Zero(), matrix.Vec3Forward())
		}
		if !ok {
			hit = ray.Point(maxContentDropDistance)
		}
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
	case content_database.ParticleSystem:
		w.spawnParticleSystem(cc, hit)
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
		w.OpenStage(cc.Id())
	case content_database.ParticleSystem:
		w.spawnParticleSystem(cc, point)
	default:
		slog.Error("double clicking this type of content is not supported",
			"id", cc.Id(), "type", cc.Config.Type)
	}
}

func (w *StageWorkspace) OpenStage(id string) {
	if w.ed.History().HasPendingChanges() {
		w.ed.BlurInterface()
		confirm_prompt.Show(w.Host, confirm_prompt.Config{
			Title:       "Discrad changes",
			Description: "You have unsaved changes to your stage, would you like to discard them and load the selected stage?",
			ConfirmText: "Yes",
			CancelText:  "No",
			OnConfirm: func() {
				w.ed.FocusInterface()
				w.loadStage(id)
			},
			OnCancel: func() { w.ed.FocusInterface() },
		})
	} else {
		w.loadStage(id)
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
		w.ed.History().Clear()
	}
}

func (w *StageWorkspace) spawnTexture(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnTexture").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
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
	e := man.AddEntity(cc.Config.Name, matrix.Vec3Zero())
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
	e.StageData.Description.Material = mat.Id
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads(), &e.Transform, e)
	// Set the position after generating the BVH
	e.Transform.SetPosition(point)
	man.AddBVH(e.StageData.Bvh, &e.Transform)
	e.StageData.Description.Textures = []string{cc.Id()}
	e.StageData.ShaderData = &shader_data_registry.ShaderDataStandard{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	w.Host.RunOnMainThread(func() {
		tex.DelayedCreate(w.Host.Window.Renderer)
		draw := rendering.Drawing{
			Material:   mat,
			Mesh:       e.StageData.Mesh,
			ShaderData: e.StageData.ShaderData,
			Transform:  &e.Transform,
			ViewCuller: &w.Host.Cameras.Primary,
		}
		w.Host.Drawings.AddDrawing(draw)
	})
	w.stageView.Manager().ClearSelection()
	w.stageView.Manager().SelectEntity(e)
}

func (w *StageWorkspace) spawnMesh(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnMesh").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
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
	e := man.AddEntity(cc.Config.Name, matrix.Vec3Zero())
	e.StageData.Mesh = w.Host.MeshCache().Mesh(cc.Id(), km.Verts, km.Indexes)
	e.StageData.Description.Mesh = e.StageData.Mesh.Key()
	e.StageData.Description.Material = mat.Id
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads(), &e.Transform, e)
	// Set the position after generating the BVH
	e.Transform.SetPosition(point)
	man.AddBVH(e.StageData.Bvh, &e.Transform)
	man.RefitBVH(e)
	e.StageData.ShaderData = &shader_data_registry.ShaderDataStandard{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       e.StageData.Mesh,
		ShaderData: e.StageData.ShaderData,
		Transform:  &e.Transform,
		ViewCuller: &w.Host.Cameras.Primary,
	}
	w.Host.Drawings.AddDrawing(draw)
	e.OnDestroy.Add(func() { e.StageData.ShaderData.Destroy() })
	w.stageView.Manager().ClearSelection()
	w.stageView.Manager().SelectEntity(e)
}

func (w *StageWorkspace) spawnParticleSystem(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnParticleSystem").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	bindKey := engine_entity_data_particles.BindingKey
	e, _ := w.createDataBoundEntity(cc.Config.Name, bindKey)
	e.Transform.SetPosition(point)
	for _, de := range e.DataBindingsByKey(bindKey) {
		de.SetFieldByName("Id", cc.Id())
		data_binding_renderer.Updated(de, weak.Make(w.Host), e)
	}
	changeEvtId := w.ed.Events().OnContentChangesSaved.Add(func(id string) {
		for _, de := range e.DataBindingsByKey(bindKey) {
			if de.FieldValueByName("Id").(string) == id {
				data_binding_renderer.Updated(de, weak.Make(w.Host), e)
			}
		}
	})
	e.OnDestroy.Add(func() {
		w.ed.Events().OnContentChangesSaved.Remove(changeEvtId)
	})
}

func (w *StageWorkspace) attachMaterial(cc *content_database.CachedContent, e *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("StageWorkspace.attachMaterial").End()
	if e.StageData.ShaderData == nil {
		return
	}
	if e.StageData.PendingMaterialChange {
		slog.Warn("a material is already being compiled to attach to this entity, please wait")
		return
	}
	e.StageData.PendingMaterialChange = true
	// goroutine
	go func() {
		mat, ok := w.Host.MaterialCache().FindMaterial(cc.Id())
		if !ok {
			var err error
			mat, err = w.Host.MaterialCache().Material(cc.Id())
			if err != nil {
				slog.Error("failed to compile the material", "id", cc.Id(), "name", cc.Config.Name, "error", err)
				return
			}
		}
		e.SetMaterial(mat.CreateInstance(mat.Textures), w.stageView.Manager())
		if w.stageView.Manager().IsSelected(e) {
			w.Host.RunOnMainThread(w.detailsUI.reload)
		}
		e.StageData.PendingMaterialChange = false
		// Don't want dirties to run during the transform clean map read
		w.Host.RunOnMainThread(e.Transform.SetDirty)
		w.setInitialSkinnedPose(e)
	}()
}

func (w *StageWorkspace) setInitialSkinnedPose(e *editor_stage_manager.StageEntity) {
	skin := e.StageData.ShaderData.SkinningHeader()
	if skin != nil {
		km, err := kaiju_mesh.ReadMesh(e.StageData.Mesh.Key(), w.Host)
		if err != nil {
			return
		}
		ids := klib.ExtractFromSlice(km.Joints, func(i int) int32 {
			return km.Joints[i].Id
		})
		skin.CreateBones(ids)
		for i := range km.Joints {
			j := &km.Joints[i]
			bone := skin.BoneByIndex(i)
			bone.Id = j.Id
			bone.Skin = j.Skin
			bone.Transform.Initialize(w.Host.WorkGroup())
		}
		for i := range km.Joints {
			bone := skin.BoneByIndex(i)
			j := &km.Joints[i]
			parent := skin.FindBone(j.Parent)
			if parent != nil {
				bone.Transform.SetParent(&parent.Transform)
			} else {
				bone.Transform.SetParent(&e.Transform)
			}
			bone.Transform.SetLocalPosition(j.Position)
			bone.Transform.SetRotation(j.Rotation)
			bone.Transform.SetScale(j.Scale)
		}
		animIdx := 0
		anims := km.Animations
		if len(anims[animIdx].Frames) == 0 {
			return
		}
		a := framework.NewSkinAnimation(anims[animIdx])
		a.Update(0)
		frame := a.CurrentFrame()
		for i := range frame.Key.Bones {
			frame.Bone = &frame.Key.Bones[i]
			bone := skin.FindBone(int32(frame.Bone.NodeIndex))
			if bone == nil {
				continue
			}
			nextFrame, ok := a.FindNextFrameForBone(bone.Id, frame.Bone.PathType)
			if !ok {
				nextFrame = frame
				nextFrame.Bone = frame.Bone
			}
			data := a.Interpolate(frame, nextFrame)
			switch frame.Bone.PathType {
			case load_result.AnimPathTranslation:
				bone.Transform.SetLocalPosition(matrix.Vec3FromSlice(data[:]))
			case load_result.AnimPathRotation:
				bone.Transform.SetRotation(matrix.Quaternion(data).ToEuler())
			case load_result.AnimPathScale:
				bone.Transform.SetScale(matrix.Vec3FromSlice(data[:]))
			}
		}
	}
}

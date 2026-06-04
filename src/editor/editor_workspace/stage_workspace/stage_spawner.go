/******************************************************************************/
/* stage_spawner.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"log/slog"
	"strings"
	"weak"

	"kaijuengine.com/editor/codegen"
	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_overlay/confirm_prompt"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_stage_manager/data_binding_renderer"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/engine_entity_data/engine_entity_data_camera"
	"kaijuengine.com/engine_entity_data/engine_entity_data_light"
	"kaijuengine.com/engine_entity_data/engine_entity_data_particles"
	"kaijuengine.com/engine_entity_data/engine_entity_data_terrain"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
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
	cam, ok := w.createDataBoundEntity("Camera", engine_entity_data_camera.BindingKey())
	if ok {
		shouldMakePrimary := true
		for _, e := range w.stageView.Manager().List() {
			if e == cam {
				continue
			}
			if len(e.DataBindingsByKey(engine_entity_data_camera.BindingKey())) > 0 {
				shouldMakePrimary = false
				break
			}
		}
		db := cam.DataBindingsByKey(engine_entity_data_camera.BindingKey())
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
	return w.createDataBoundEntity("Light", engine_entity_data_light.BindingKey())
}

func (w *StageWorkspace) CreatePrimitive(primitive rendering.PrimitiveMesh) (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.CreatePrimitive").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	mesh := rendering.NewMeshPrimitive(w.Host.MeshCache(), primitive)
	if mesh == nil {
		slog.Error("failed to create the primitive mesh", "primitive", primitive)
		return nil, false
	}
	mat, err := w.Host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		slog.Error("failed to find the basic material", "error", err)
		return nil, false
	}
	tex, err := w.Host.TextureCache().Texture(assets.TextureSquare,
		rendering.TextureFilterLinear)
	if err != nil {
		slog.Error("failed to create the default texture", "error", err)
		return nil, false
	}
	mat = mat.CreateInstance([]*rendering.Texture{tex})
	verts, indexes, ok := rendering.BuiltInMeshData(mesh.Key())
	if !ok {
		slog.Error("failed to find the primitive mesh data", "mesh", mesh.Key())
		return nil, false
	}
	man := w.stageView.Manager()
	e := man.AddEntity(primitiveName(primitive), matrix.Vec3Zero())
	e.StageData.Mesh = mesh
	e.StageData.SnapVertices = editor_stage_manager.SnapVerticesFromMesh(verts)
	e.StageData.Description.Mesh = mesh.Key()
	e.StageData.Description.Material = mat.Id
	e.StageData.ShaderData = shader_data_registry.Create(mat.Shader.ShaderDataName())
	km := kaiju_mesh.KaijuMesh{Verts: verts, Indexes: indexes}
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads(), &e.Transform, e)
	e.Transform.SetPosition(w.stageView.LookAtPoint())
	man.AddBVH(e)
	man.RefitBVH(e)
	w.Host.RunOnMainThread(func() {
		w.Host.RunOnRenderThread(func(device *rendering.GPUDevice) {
			tex.DelayedCreate(device)
		})
		draw := rendering.Drawing{
			Material:   mat,
			Mesh:       e.StageData.Mesh,
			ShaderData: e.StageData.ShaderData,
			Transform:  &e.Transform,
			ViewCuller: &w.Host.Cameras.Primary,
		}
		w.Host.Drawings.AddDrawing(draw)
		man.AddPickingDrawing(e)
	})
	man.ClearSelection()
	man.SelectEntity(e)
	return e, true
}

func primitiveName(primitive rendering.PrimitiveMesh) string {
	switch primitive {
	case rendering.PrimitiveMeshSphere:
		return "Sphere"
	case rendering.PrimitiveMeshTexturableCube:
		return "Cube"
	case rendering.PrimitiveMeshCapsule:
		return "Capsule"
	case rendering.PrimitiveMeshPlane:
		return "Plane"
	case rendering.PrimitiveMeshCylinder:
		return "Cylinder"
	case rendering.PrimitiveMeshCone:
		return "Cone"
	case rendering.PrimitiveMeshArrow:
		return "Arrow"
	default:
		return "Primitive"
	}
}

func (w *StageWorkspace) createDataBoundEntity(name, bindKey string) (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.createDataBoundEntity").End()
	g, ok := w.ed.Project().EntityDataBinding(bindKey)
	if !ok {
		slog.Error("failed to locate the entity binding data", "key", bindKey)
		return nil, false
	}
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	man := w.stageView.Manager()
	e := man.AddEntity(name, w.stageView.LookAtPoint())
	w.attachEntityData(e, g)
	man.ClearSelection()
	man.SelectEntity(e)
	return e, true
}

func (w *StageWorkspace) spawnContentAtMouse(cc *content_database.CachedContent, m *hid.Mouse) {
	defer tracing.NewRegion("StageWorkspace.spawnContent").End()
	hit := w.contentDropPoint(m)
	ray := w.stageView.Camera().RayCast(m)
	e, _, eHitOk := w.stageView.Manager().TryHitEntity(ray)
	if !eHitOk {
		e = nil
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
		w.spawnTemplate(cc, hit)
	case content_database.Material:
		if eHitOk {
			w.attachMaterial(cc, e)
		}
	case content_database.ParticleSystem:
		w.spawnParticleSystem(cc, hit)
	case content_database.Terrain:
		w.spawnTerrain(cc, hit)
	default:
		slog.Error("dropping this type of content into the stage is not supported",
			"id", cc.Id(), "type", cc.Config.Type)
	}
}

func (w *StageWorkspace) contentDropPoint(m *hid.Mouse) matrix.Vec3 {
	ray := w.stageView.Camera().RayCast(m)
	// TODO:  Find the point on the entity that was hit, otherwise fall back
	// to doing the ground plane/distance hit point
	if _, hit, ok := w.stageView.Manager().TryHitEntity(ray); ok {
		return hit
	}
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
	return hit
}

func (w *StageWorkspace) spawnTemplate(cc *content_database.CachedContent, hit matrix.Vec3) {
	man := w.stageView.Manager()
	e, err := man.SpawnTemplate(w.Host, w.ed.Project(), cc, hit)
	if err != nil {
		slog.Error("failed to spawn the template", "error", err)
		return
	}
	var attachData func(target *editor_stage_manager.StageEntity)
	attachData = func(target *editor_stage_manager.StageEntity) {
		for _, b := range target.DataBindings() {
			data_binding_renderer.Attached(b, weak.Make(w.Host), man, target)
		}
		for i := range target.Children {
			attachData(editor_stage_manager.EntityToStageEntity(target.Children[i]))
		}
	}
	attachData(e)
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
	case content_database.Terrain:
		w.spawnTerrain(cc, point)
	default:
		slog.Error("double clicking this type of content is not supported",
			"id", cc.Id(), "type", cc.Config.Type)
	}
}

func (w *StageWorkspace) OpenStage(id string) {
	if w.ed.History().HasPendingChanges() {
		w.ed.BlurInterface()
		confirm_prompt.Show(w.Host, confirm_prompt.Config{
			Title:       "Discard changes",
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
	w.ed.Project().Settings.EditorSettings.LatestOpenStage = id
	w.ed.Project().Settings.Save(w.ed.ProjectFileSystem())
}

func (w *StageWorkspace) spawnTexture(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnTexture").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	var mat *rendering.Material
	var err error
	if w.stageView.IsView3D() {
		mat, err = w.Host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	} else {
		mat, err = w.Host.MaterialCache().Material(assets.MaterialDefinitionUnlit)
	}
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
	tex, err := w.Host.TextureCache().InsertRawTexture(rendering.GenerateUniqueTextureKey,
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
	e.StageData.SnapVertices = editor_stage_manager.SnapVerticesFromMesh(km.Verts)
	// Not using mat.Id here due to the material being assets.MaterialDefinitionBasic
	e.StageData.Description.Material = mat.Id
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads(), &e.Transform, e)
	// Set the position after generating the BVH
	e.Transform.SetPosition(point)
	man.AddBVH(e)
	e.StageData.Description.Textures = []string{cc.Id()}
	if w.stageView.IsView3D() {
		e.StageData.ShaderData = &shader_data_registry.ShaderDataStandard{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	} else {
		e.StageData.ShaderData = &shader_data_registry.ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			UVs:            matrix.Vec4{0, 0, 1, 1},
		}
	}
	w.Host.RunOnMainThread(func() {
		w.Host.RunOnRenderThread(func(device *rendering.GPUDevice) {
			tex.DelayedCreate(device)
		})
		if !w.stageView.IsView3D() && tex.Width > 0 && tex.Height > 0 {
			var w, h matrix.Float = 1, 1
			tw := matrix.Float(tex.Width)
			th := matrix.Float(tex.Height)
			if tex.Width < tex.Height {
				h = th / tw
			} else {
				w = tw / th
			}
			e.Transform.SetScale(matrix.NewVec3(w, h, 1))
		}
		draw := rendering.Drawing{
			Material:   mat,
			Mesh:       e.StageData.Mesh,
			ShaderData: e.StageData.ShaderData,
			Transform:  &e.Transform,
			ViewCuller: &w.Host.Cameras.Primary,
		}
		w.Host.Drawings.AddDrawing(draw)
		w.stageView.Manager().AddPickingDrawing(e)
	})
	w.stageView.Manager().ClearSelection()
	w.stageView.Manager().SelectEntity(e)
}

func (w *StageWorkspace) spawnMesh(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnMesh").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	set, path, ok := w.readMeshSet(cc)
	if !ok {
		return
	}
	man := w.stageView.Manager()
	if len(set.Meshes) > 1 {
		parent := man.AddEntity(cc.Config.Name, point)
		for i := range set.Meshes {
			mesh := &set.Meshes[i]
			ref := kaiju_mesh.MeshRefString(cc.Id(), mesh.Key)
			name := meshStageName(mesh.Name, mesh.Key)
			matId := meshSubmeshMaterial(cc, mesh.Key)
			child := w.spawnMeshEntity(name, ref, matId, *mesh, path, matrix.Vec3Zero(), parent, false)
			if child == nil {
				continue
			}
		}
		man.ClearSelection()
		man.SelectEntity(parent)
		return
	}
	if len(set.Meshes) == 0 {
		slog.Error("mesh content did not contain any meshes", "id", cc.Id())
		return
	}
	mesh := &set.Meshes[0]
	matId := meshSubmeshMaterial(cc, mesh.Key)
	w.spawnMeshEntity(cc.Config.Name, cc.Id(), matId, *mesh, path, point, nil, false)
}

func (w *StageWorkspace) spawnSubmesh(cc *content_database.CachedContent, key string, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnSubmesh").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	set, path, ok := w.readMeshSet(cc)
	if !ok {
		return
	}
	km, ok := set.MeshByKey(key)
	if !ok {
		slog.Error("failed to find submesh in mesh content", "id", cc.Id(), "key", key)
		return
	}
	ref := kaiju_mesh.MeshRefString(cc.Id(), key)
	w.spawnMeshEntity(meshStageName(km.Name, km.Key), ref, meshSubmeshMaterial(cc, key), km, path, point, nil, false)
}

func (w *StageWorkspace) readMeshSet(cc *content_database.CachedContent) (kaiju_mesh.KaijuMeshSet, string, bool) {
	path := content_database.ToContentPath(cc.Path)
	data, err := w.ed.ProjectFileSystem().ReadFile(path)
	if err != nil {
		slog.Error("error reading the mesh file", "path", path)
		return kaiju_mesh.KaijuMeshSet{}, path, false
	}
	set, err := kaiju_mesh.DeserializeSet(data)
	if err != nil {
		slog.Error("failed to deserialize the mesh set", "id", cc.Id(), "error", err)
		return kaiju_mesh.KaijuMeshSet{}, path, false
	}
	return set, path, true
}

func (w *StageWorkspace) spawnMeshEntity(name, meshRef, materialId string, km kaiju_mesh.KaijuMesh, contentPath string, point matrix.Vec3, parent *editor_stage_manager.StageEntity, sourceLocalTransform bool) *editor_stage_manager.StageEntity {
	mat, err := w.Host.MaterialCache().Material(materialId)
	if err != nil {
		slog.Error("failed to find the mesh material", "id", materialId, "error", err)
		return nil
	}
	if len(mat.Textures) == 0 {
		tex, err := w.Host.TextureCache().Texture(assets.TextureSquare,
			rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the default texture", "error", err)
			return nil
		}
		mat = mat.CreateInstance([]*rendering.Texture{tex})
	} else {
		mat = mat.CreateInstance(mat.Textures)
	}
	man := w.stageView.Manager()
	e := man.AddEntity(name, matrix.Vec3Zero())
	if parent != nil {
		man.SetEntityParent(e, parent)
		if sourceLocalTransform {
			e.Transform.SetLocalPosition(km.Node.Position)
			e.Transform.SetRotation(km.Node.Rotation)
			scale := km.Node.Scale
			if scale == matrix.Vec3Zero() {
				scale = matrix.Vec3One()
			}
			e.Transform.SetScale(scale)
		} else {
			e.Transform.SetLocalPosition(matrix.Vec3Zero())
			e.Transform.SetRotation(matrix.Vec3Zero())
			e.Transform.SetScale(matrix.Vec3One())
		}
	} else {
		e.Transform.SetPosition(point)
	}
	e.StageData.Mesh = w.Host.MeshCache().Mesh(meshRef, km.Verts, km.Indexes)
	e.StageData.SnapVertices = editor_stage_manager.SnapVerticesFromMesh(km.Verts)
	e.StageData.Description.Mesh = meshRef
	e.StageData.Description.Material = mat.Id
	missingBVH := km.BVH == nil
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads(), &e.Transform, e)
	if missingBVH {
		content_database.SaveMeshBVHInBackground(km, contentPath, w.ed.ProjectFileSystem(), meshRef)
	}
	man.AddBVH(e)
	man.RefitBVH(e)
	e.StageData.ShaderData = shader_data_registry.Create(mat.Shader.ShaderDataName())
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       e.StageData.Mesh,
		ShaderData: e.StageData.ShaderData,
		Transform:  &e.Transform,
		ViewCuller: &w.Host.Cameras.Primary,
	}
	w.Host.Drawings.AddDrawing(draw)
	man.AddPickingDrawing(e)
	w.setInitialSkinnedPose(e)
	if parent == nil {
		man.ClearSelection()
		man.SelectEntity(e)
	}
	return e
}

func meshSubmeshMaterial(cc *content_database.CachedContent, key string) string {
	if cc.Config.Mesh != nil {
		var first string
		for i := range cc.Config.Mesh.Submeshes {
			submesh := &cc.Config.Mesh.Submeshes[i]
			if submesh.Missing || submesh.Material == "" {
				continue
			}
			if first == "" {
				first = submesh.Material
			}
			if submesh.Key == key && submesh.Material != "" {
				return submesh.Material
			}
		}
		if key == "" && first != "" {
			return first
		}
	}
	return assets.MaterialDefinitionBasic
}

func (w *StageWorkspace) meshDefaultMaterial(meshId string) (string, bool) {
	ref := kaiju_mesh.ParseMeshRef(meshId)
	if ref.Asset == "" {
		return assets.MaterialDefinitionBasic, true
	}
	cc, err := w.ed.Cache().Read(ref.Asset)
	if err != nil || cc.Config.Type != (content_database.Mesh{}).TypeName() {
		return assets.MaterialDefinitionBasic, true
	}
	return meshSubmeshMaterial(&cc, ref.Key), true
}

func shouldAutoApplyMeshMaterial(materialId string) bool {
	return materialId == "" || materialId == assets.MaterialDefinitionBasic
}

func cloneTextureIDs(ids []string) []string {
	if ids == nil {
		return nil
	}
	return append([]string(nil), ids...)
}

func meshStageName(name, key string) string {
	if strings.TrimSpace(name) != "" {
		parts := strings.Split(name, "/")
		return parts[len(parts)-1]
	}
	if strings.TrimSpace(key) != "" {
		return key
	}
	return "Mesh"
}

func (w *StageWorkspace) setEntityMesh(e *editor_stage_manager.StageEntity, meshId string) bool {
	defer tracing.NewRegion("StageWorkspace.setEntityMesh").End()
	if e == nil || meshId == "" {
		return false
	}
	km, mesh, ok := w.meshById(meshId)
	if !ok {
		return false
	}
	if shouldAutoApplyMeshMaterial(e.StageData.Description.Material) {
		if matId, ok := w.meshDefaultMaterial(meshId); ok && matId != "" {
			e.StageData.Description.Material = matId
			if matId != assets.MaterialDefinitionBasic {
				e.StageData.Description.Textures = nil
			}
		}
	}
	mat, ok := w.materialForEntity(e)
	if !ok {
		return false
	}
	if e.StageData.Description.Material == "" {
		e.StageData.Description.Material = mat.Id
	}
	oldShaderData := e.StageData.ShaderData
	newShaderData := rendering.ReflectDuplicateDrawInstance(oldShaderData)
	if newShaderData == nil {
		newShaderData = shader_data_registry.Create(mat.Shader.ShaderDataName())
	} else {
		newShaderData.Base().CancelDestroy()
	}
	if oldShaderData != nil {
		oldShaderData.Destroy()
	}
	man := w.stageView.Manager()
	man.RemoveEntityBVH(e)
	man.ClearPickingDrawing(e)
	e.StageData.Mesh = mesh
	e.StageData.SnapVertices = editor_stage_manager.SnapVerticesFromMesh(km.Verts)
	e.StageData.Description.Mesh = meshId
	e.StageData.Bvh = km.GenerateBVH(w.Host.Threads(), &e.Transform, e)
	man.AddBVH(e)
	man.RefitBVH(e)
	e.StageData.ShaderData = newShaderData
	if !e.IsActive() {
		e.StageData.ShaderData.Deactivate()
	}
	w.Host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat,
		Mesh:       e.StageData.Mesh,
		ShaderData: e.StageData.ShaderData,
		Transform:  &e.Transform,
		ViewCuller: &w.Host.Cameras.Primary,
	})
	man.AddPickingDrawing(e)
	w.setInitialSkinnedPose(e)
	return true
}

func (w *StageWorkspace) clearEntityMesh(e *editor_stage_manager.StageEntity) bool {
	defer tracing.NewRegion("StageWorkspace.clearEntityMesh").End()
	if e == nil || e.StageData.Mesh == nil {
		return false
	}
	if e.StageData.ShaderData != nil {
		e.StageData.ShaderData.Destroy()
		e.StageData.ShaderData = nil
	}
	man := w.stageView.Manager()
	man.RemoveEntityBVH(e)
	man.ClearPickingDrawing(e)
	e.StageData.Bvh = nil
	e.StageData.Mesh = nil
	e.StageData.SnapVertices = nil
	e.StageData.Description.Mesh = ""
	return true
}

func (w *StageWorkspace) meshById(meshId string) (kaiju_mesh.KaijuMesh, *rendering.Mesh, bool) {
	defer tracing.NewRegion("StageWorkspace.meshById").End()
	km := kaiju_mesh.KaijuMesh{}
	ref := kaiju_mesh.ParseMeshRef(meshId)
	if ref.Key == "" {
		if verts, indexes, builtIn := rendering.BuiltInMeshData(meshId); builtIn {
			km.Verts = verts
			km.Indexes = indexes
		}
	}
	if len(km.Verts) == 0 {
		var err error
		if km, err = kaiju_mesh.ReadMesh(meshId, w.Host); err != nil {
			slog.Error("failed to read the mesh for entity", "id", meshId, "error", err)
			return kaiju_mesh.KaijuMesh{}, nil, false
		}
	}
	mesh := w.Host.MeshCache().Mesh(meshId, km.Verts, km.Indexes)
	return km, mesh, true
}

func (w *StageWorkspace) materialForEntity(e *editor_stage_manager.StageEntity) (*rendering.Material, bool) {
	defer tracing.NewRegion("StageWorkspace.materialForEntity").End()
	matId := e.StageData.Description.Material
	if matId == "" {
		matId = assets.MaterialDefinitionBasic
	}
	mat, err := w.Host.MaterialCache().Material(matId)
	if err != nil {
		slog.Error("failed to find the entity material", "id", matId, "error", err)
		return nil, false
	}
	texs := make([]*rendering.Texture, 0, len(e.StageData.Description.Textures))
	for _, texId := range e.StageData.Description.Textures {
		tex, err := w.Host.TextureCache().Texture(texId, rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to find the entity texture", "id", texId, "error", err)
			continue
		}
		texs = append(texs, tex)
	}
	if len(e.StageData.Description.Textures) == 0 {
		texs = append(texs, mat.Textures...)
	}
	if len(texs) == 0 {
		tex, err := w.Host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the default texture", "error", err)
			return nil, false
		}
		texs = append(texs, tex)
	}
	return mat.CreateInstance(texs), true
}

func (w *StageWorkspace) setEntityMaterial(e *editor_stage_manager.StageEntity, materialId string, textureIds []string) bool {
	defer tracing.NewRegion("StageWorkspace.setEntityMaterial").End()
	if e == nil {
		return false
	}
	loadMaterialId := materialId
	if loadMaterialId == "" {
		loadMaterialId = assets.MaterialDefinitionBasic
	}
	mat, err := w.Host.MaterialCache().Material(loadMaterialId)
	if err != nil {
		slog.Error("failed to find the entity material", "id", loadMaterialId, "error", err)
		return false
	}
	texs := mat.Textures
	var storeTextureIds []string
	if textureIds != nil {
		texs = make([]*rendering.Texture, 0, len(textureIds))
		for _, texId := range textureIds {
			tex, err := w.Host.TextureCache().Texture(texId, rendering.TextureFilterLinear)
			if err != nil {
				slog.Error("failed to find the entity texture", "id", texId, "error", err)
				return false
			}
			texs = append(texs, tex)
		}
		storeTextureIds = cloneTextureIDs(textureIds)
	}
	if len(texs) == 0 {
		tex, err := w.Host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the default texture", "error", err)
			return false
		}
		texs = append(texs, tex)
	}
	mat = mat.CreateInstance(texs)
	if e.StageData.ShaderData != nil {
		e.StageData.ShaderData.Destroy()
	}
	e.StageData.Description.Material = materialId
	e.StageData.Description.Textures = storeTextureIds
	if e.StageData.Mesh == nil {
		e.StageData.ShaderData = nil
		return true
	}
	e.StageData.ShaderData = shader_data_registry.Create(mat.Shader.ShaderDataName())
	if !e.IsActive() {
		e.StageData.ShaderData.Deactivate()
	}
	db := entity_data_binding.ToDataBinding("", e.StageData.ShaderData)
	for i := range db.Fields {
		if db.RunTagParserOnField(i) {
			db.SetField(i, db.Fields[i].Value)
		}
	}
	w.Host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat,
		Mesh:       e.StageData.Mesh,
		ShaderData: e.StageData.ShaderData,
		Transform:  &e.Transform,
		ViewCuller: &w.Host.Cameras.Primary,
	})
	w.stageView.Manager().AddPickingDrawing(e)
	e.Transform.SetDirty()
	w.setInitialSkinnedPose(e)
	return true
}

func (w *StageWorkspace) spawnParticleSystem(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnParticleSystem").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	bindKey := engine_entity_data_particles.BindingKey()
	e, ok := w.createDataBoundEntity(cc.Config.Name, bindKey)
	if !ok {
		return
	}
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

func (w *StageWorkspace) spawnTerrain(cc *content_database.CachedContent, point matrix.Vec3) {
	defer tracing.NewRegion("StageWorkspace.spawnTerrain").End()
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	bindKey := engine_entity_data_terrain.BindingKey()
	e, ok := w.createDataBoundEntity(cc.Config.Name, bindKey)
	if !ok {
		return
	}
	e.Transform.SetPosition(point)
	for _, de := range e.DataBindingsByKey(bindKey) {
		de.SetFieldByName("Id", cc.Id())
		data_binding_renderer.Updated(de, weak.Make(w.Host), e)
	}
	changeEvtId := w.ed.Events().OnContentChangesSaved.Add(func(id string) {
		for _, de := range e.DataBindingsByKey(bindKey) {
			terrainId, ok := de.FieldValueByName("Id").(content_id.Terrain)
			if ok && string(terrainId) == id {
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
	}
}

/******************************************************************************/
/* common_data_binding_renderer.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"log/slog"

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

func commonAttached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, iconName string) rendering.DrawInstance {
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionEdGizmo)
	if err != nil {
		slog.Error("failed to find the basic material", "error", err)
		return nil
	}
	tex, err := host.TextureCache().Texture(
		"editor/textures/icons/"+iconName, rendering.TextureFilterLinear)
	if err != nil {
		slog.Error("failed to load the gizmo icon", "icon", iconName, "error", err)
		return nil
	}
	pickMat, err := host.MaterialCache().Material(assets.MaterialDefinitionEditorGizmoPick)
	if err != nil {
		slog.Error("failed to find the editor gizmo picking material", "error", err)
	}
	mat = mat.CreateInstance([]*rendering.Texture{tex})
	mesh := rendering.NewMeshQuad(host.MeshCache())
	sd := &shader_data_registry.ShaderDataUnlit{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
		UVs:            matrix.NewVec4(0, 0, 1, 1),
	}
	host.RunOnMainThread(func() {
		host.RunOnRenderThread(func(device *rendering.GPUDevice) {
			tex.DelayedCreate(device)
		})
		draw := rendering.Drawing{
			Material:   mat,
			Mesh:       mesh,
			ShaderData: sd,
			Transform:  &target.Transform,
			Layer:      rendering.RenderLayerEditor,
			ViewCuller: &host.Cameras.Primary,
		}
		host.Drawings.AddDrawing(draw)
		if pickMat != nil && manager != nil {
			if pickDraw, pickSd, ok := manager.NewPickingDrawing(target, pickMat, mesh, &target.Transform); ok {
				host.Drawings.AddDrawing(pickDraw)
				rendering.LinkDrawInstanceLifecycle(sd, pickSd)
			}
		}
	})
	box := graviton.AABB{}
	box.Extent = target.Transform.WorldScale().Scale(0.5)
	target.StageData.Bvh = graviton.NewBVH([]graviton.HitObject{box}, &target.Transform, target)
	manager.AddBVH(target)
	target.OnDeactivate.Add(sd.Deactivate)
	target.OnActivate.Add(sd.Activate)
	return sd
}

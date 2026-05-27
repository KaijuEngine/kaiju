/******************************************************************************/
/* camera_data_binding_renderer.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine_entity_data/engine_entity_data_camera"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	minimumCameraWidth          = 0.1
	minimumCameraHeight         = 0.1
	translationGizmoShaftHeight = 1.5
	translationGizmoShaftRadius = 0.025
	translationGizmoArrowHeight = 0.35
	translationGizmoArrowRadius = 0.175
)

func init() {
	AddRenderer(engine_entity_data_camera.BindingKey(), &CameraEntityDataRenderer{
		Frustums: make(map[*editor_stage_manager.StageEntity]cameraDataBindingDrawing),
	})
}

type CameraEntityDataRenderer struct {
	Frustums map[*editor_stage_manager.StageEntity]cameraDataBindingDrawing
}

type cameraDataBindingDrawing struct {
	key     string
	sd      rendering.DrawInstance
	icon    rendering.DrawInstance
	arrowSd rendering.DrawInstance
}

func (c *CameraEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraEntityDataRenderer.Attached").End()
	icon := commonAttached(host, manager, target, "camera.png")
	if _, ok := c.Frustums[target]; ok {
		slog.Error("there is an internal error in state for the editor's CameraEntityDataRenderer, show was called before any hide happened. Double selected the same target?")
		c.Detatched(host, manager, target, data)
	}

	var w, h float32 = minimumCameraWidth, minimumCameraHeight
	if val := data.FieldValueByName("Width"); val != nil {
		if f, ok := val.(float32); ok && f >= minimumCameraWidth {
			w = f
		}
	}

	if val := data.FieldValueByName("Height"); val != nil {
		if f, ok := val.(float32); ok && f > minimumCameraHeight {
			h = f
		}
	}

	var camType int = 0
	if val := data.FieldValueByName("Type"); val != nil {
		if f, ok := val.(int); ok && f >= 0 {
			camType = f
		}
	}

	identity := matrix.Mat4Identity()
	identity.Rotate(matrix.NewVec3(-90, 0, 0))

	//* key name generation log needs to be confirmed
	m := rendering.NewMeshArrowWithTransform(host.MeshCache(),
		translationGizmoShaftHeight, translationGizmoShaftRadius,
		translationGizmoArrowHeight, translationGizmoArrowRadius, 10, identity, fmt.Sprintf("%s", target.Name()))

	mat, _ := host.MaterialCache().Material("gizmo_overlay.material")
	arrowSd := shader_data_registry.Create("unlit").(*shader_data_registry.ShaderDataUnlit)
	arrowSd.Color = matrix.ColorCadetBlue()

	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat,
		Mesh:       m,
		ShaderData: arrowSd,
		Transform:  &target.Transform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	})
	arrowSd.Deactivate()

	var cam cameras.Camera
	switch engine_entity_data_camera.CameraType(camType) {
	case engine_entity_data_camera.CameraTypeOrthographic:
		cam = cameras.NewStandardCameraOrthographic(w, h, w, h, target.Transform.Position())
	case engine_entity_data_camera.CameraTypeTurntable:
		cam = cameras.ToTurntable(cameras.NewStandardCamera(w, h, w, h, target.Transform.Position()))
	case engine_entity_data_camera.CameraTypePerspective:
		fallthrough
	default:
		cam = cameras.NewStandardCamera(w, h, w, h, target.Transform.Position())
	}

	cam.SetProperties(
		data.FieldValueByName("FOV").(float32),
		data.FieldValueByName("NearPlane").(float32),
		data.FieldValueByName("FarPlane").(float32),
		w, h,
	)
	frustum := rendering.NewMeshFrustumBox(host.MeshCache(), cam.InverseProjection())
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdFrustumWire)
	if err != nil {
		slog.Error("failed to load transform wire material", "error", err)
		return
	}
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	sd.(*shader_data_registry.ShaderDataEdFrustumWire).Color = matrix.ColorWhite()
	sd.(*shader_data_registry.ShaderDataEdFrustumWire).FrustumProjection = cam.InverseProjection()
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       frustum,
		ShaderData: sd,
		Transform:  &target.Transform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	})
	sd.Deactivate()
	c.Frustums[target] = cameraDataBindingDrawing{frustum.Key(), sd, icon, arrowSd}
	target.OnActivate.Add(func() {
		if d, ok := c.Frustums[target]; ok {
			d.icon.Activate()
			d.sd.Activate()
			d.arrowSd.Activate()
		}
	})
	target.OnDeactivate.Add(func() {
		if d, ok := c.Frustums[target]; ok {
			d.icon.Deactivate()
			d.arrowSd.Deactivate()
			d.sd.Deactivate()
		}
	})
	target.OnDestroy.Add(func() {
		c.Detatched(host, manager, target, data)
	})
}

func (c *CameraEntityDataRenderer) Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraEntityDataRenderer.Detatched").End()
	if d, ok := c.Frustums[target]; ok {
		d.sd.Destroy()
		d.icon.Destroy()
		d.arrowSd.Destroy()
		host.MeshCache().RemoveMesh(d.key)
		delete(c.Frustums, target)
	}
}

func (c *CameraEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraEntityDataRenderer.Show").End()
	if d, ok := c.Frustums[target]; ok {
		d.sd.Activate()
		d.arrowSd.Activate()
	}
}

func (c *CameraEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraEntityDataRenderer.Hide").End()
	if d, ok := c.Frustums[target]; ok {
		d.sd.Deactivate()
		d.arrowSd.Deactivate()
	}
}

func (c *CameraEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	if t, ok := c.Frustums[target]; ok {
		// Assumption: width and height field will be present on the cameraEntityData
		w := data.FieldValueByName("Width").(float32)
		h := data.FieldValueByName("Height").(float32)

		if w < minimumCameraWidth {
			w = minimumCameraWidth
		}
		if h < minimumCameraHeight {
			h = minimumCameraHeight
		}
		if w <= 0 || h <= 0 {
			slog.Warn("camera width or height is zero , might cause problem", "width", w, "height", h)
		}
		var cam cameras.Camera
		camType := engine_entity_data_camera.CameraType(data.FieldValueByName("Type").(int))
		switch camType {
		case engine_entity_data_camera.CameraTypeOrthographic:
			cam = cameras.NewStandardCameraOrthographic(w, h, w, h, target.Transform.Position())
		case engine_entity_data_camera.CameraTypeTurntable:
			cam = cameras.ToTurntable(cameras.NewStandardCamera(w, h, w, h, target.Transform.Position()))
		case engine_entity_data_camera.CameraTypePerspective:
			fallthrough
		default:
			cam = cameras.NewStandardCamera(w, h, w, h, target.Transform.Position())
		}
		cam.SetProperties(
			data.FieldValueByName("FOV").(float32),
			data.FieldValueByName("NearPlane").(float32),
			data.FieldValueByName("FarPlane").(float32),
			w, h,
		)
		t.sd.(*shader_data_registry.ShaderDataEdFrustumWire).FrustumProjection = cam.InverseProjection()
	}
}

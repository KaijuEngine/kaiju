/******************************************************************************/
/* transform_gizmo_picking.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"fmt"
	"math"

	"kaijuengine.com/engine"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	transformGizmoPickBase uint32 = 0xF0000000

	transformGizmoPickTranslateArrowX = transformGizmoPickBase + 0x001
	transformGizmoPickTranslateArrowY = transformGizmoPickBase + 0x002
	transformGizmoPickTranslateArrowZ = transformGizmoPickBase + 0x003
	transformGizmoPickTranslatePlaneX = transformGizmoPickBase + 0x011
	transformGizmoPickTranslatePlaneY = transformGizmoPickBase + 0x012
	transformGizmoPickTranslatePlaneZ = transformGizmoPickBase + 0x013

	transformGizmoPickRotateX = transformGizmoPickBase + 0x101
	transformGizmoPickRotateY = transformGizmoPickBase + 0x102
	transformGizmoPickRotateZ = transformGizmoPickBase + 0x103

	transformGizmoPickScaleX = transformGizmoPickBase + 0x201
	transformGizmoPickScaleY = transformGizmoPickBase + 0x202
	transformGizmoPickScaleZ = transformGizmoPickBase + 0x203
)

func gizmoPickID(base uint32, axis int) uint32 {
	if axis < 0 || axis > 2 {
		return 0
	}
	return base + uint32(axis)
}

func gizmoPickAxis(id, base uint32) (int, bool) {
	if id < base || id > base+2 {
		return -1, false
	}
	return int(id - base), true
}

func translationArrowPickID(axis int) uint32 {
	return gizmoPickID(transformGizmoPickTranslateArrowX, axis)
}

func translationPlanePickID(axis int) uint32 {
	return gizmoPickID(transformGizmoPickTranslatePlaneX, axis)
}

func translationPickTarget(id uint32) (int, TranslationHitEnum, bool) {
	if axis, ok := gizmoPickAxis(id, transformGizmoPickTranslateArrowX); ok {
		return axis, TRANSLATION_TYPE_ARROW, true
	}
	if axis, ok := gizmoPickAxis(id, transformGizmoPickTranslatePlaneX); ok {
		return axis, TRANSLATION_TYPE_PLANE, true
	}
	return -1, TRANSLATION_TYPE_NONE, false
}

func rotationPickID(axis int) uint32 {
	return gizmoPickID(transformGizmoPickRotateX, axis)
}

func rotationPickAxis(id uint32) (int, bool) {
	return gizmoPickAxis(id, transformGizmoPickRotateX)
}

func scalePickID(axis int) uint32 {
	return gizmoPickID(transformGizmoPickScaleX, axis)
}

func scalePickAxis(id uint32) (int, bool) {
	return gizmoPickAxis(id, transformGizmoPickScaleX)
}

func addGizmoPickDrawing(host *engine.Host, material *rendering.Material, mesh *rendering.Mesh,
	transform *matrix.Transform, owner rendering.DrawInstance, pickID uint32,
) rendering.DrawInstance {
	if host == nil || material == nil || mesh == nil || transform == nil || owner == nil || pickID == 0 {
		return nil
	}
	sd := shader_data_registry.Create("editor_pick")
	sd.(*shader_data_registry.ShaderDataEditorPicking).PickID = pickID
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  transform,
		Layer:      rendering.RenderLayerEditorGizmoPicking,
		ViewCuller: &host.Cameras.Primary,
	})
	rendering.LinkDrawInstanceLifecycle(owner, sd)
	return sd
}

func newRotationGizmoPickMesh(cache *rendering.MeshCache, radius, thickness float32, segments int) *rendering.Mesh {
	if segments < 3 {
		segments = 3
	}
	if thickness <= 0 {
		thickness = 0.1
	}
	key := fmt.Sprintf("_editor_rotation_gizmo_pick_%.2f_%.2f_%d", radius, thickness, segments)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	inner := radius - thickness*0.5
	outer := radius + thickness*0.5
	if inner < 0 {
		inner = 0
	}
	verts := make([]rendering.Vertex, segments*2)
	for i := 0; i < segments; i++ {
		phi := float32(i) * 2.0 * float32(math.Pi) / float32(segments)
		cosPhi := matrix.Cos(phi)
		sinPhi := matrix.Sin(phi)
		verts[i*2].Position = matrix.Vec3{inner * cosPhi, 0, inner * sinPhi}
		verts[i*2+1].Position = matrix.Vec3{outer * cosPhi, 0, outer * sinPhi}
		verts[i*2].Normal = matrix.Vec3Up()
		verts[i*2+1].Normal = matrix.Vec3Up()
		verts[i*2].UV0 = matrix.Vec2{0, 0}
		verts[i*2+1].UV0 = matrix.Vec2{1, 1}
		verts[i*2].Color = matrix.ColorWhite()
		verts[i*2+1].Color = matrix.ColorWhite()
	}
	indexes := make([]uint32, 0, segments*12)
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		innerA := uint32(i * 2)
		outerA := innerA + 1
		innerB := uint32(next * 2)
		outerB := innerB + 1
		indexes = append(indexes, innerA, outerA, outerB, innerA, outerB, innerB)
		indexes = append(indexes, innerA, outerB, outerA, innerA, innerB, outerB)
	}
	return cache.Mesh(key, verts, indexes)
}

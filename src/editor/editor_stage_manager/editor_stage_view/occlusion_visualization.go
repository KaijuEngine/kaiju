/******************************************************************************/
/* occlusion_visualization.go                                                 */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/

package editor_stage_view

import (
	"log/slog"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const maxEditorOcclusionBoxes = 512

type occlusionBoundsVisualizer struct {
	mesh      *rendering.Mesh
	shader    *shader_data_registry.ShaderDataEdTransformWire
	transform matrix.Transform
}

func (v *StageView) updateOcclusionVisualization() {
	defer tracing.NewRegion("StageView.updateOcclusionVisualization").End()
	if v.host == nil || v.host.Window == nil || v.host.Window.GpuInstance == nil ||
		len(v.host.Window.GpuInstance.Devices) == 0 {
		return
	}
	device := v.host.Window.GpuInstance.PrimaryDevice()
	mode := device.OcclusionVisualizationMode()
	if mode != rendering.OcclusionVisualizationTestedBounds &&
		mode != rendering.OcclusionVisualizationOccludedBounds {
		v.occlusionViz.hide()
		return
	}
	if !v.occlusionViz.ensure(v.host, mode) {
		return
	}
	bounds := device.OcclusionDebugBounds()
	if len(bounds) == 0 {
		v.occlusionViz.hide()
		return
	}
	v.occlusionViz.show()
	verts := occlusionBoundsVertices(bounds)
	v.host.MeshCache().UpdateMeshVertices(v.occlusionViz.mesh.Key(), verts)
}

func (o *occlusionBoundsVisualizer) ensure(host *engine.Host, mode rendering.OcclusionVisualizationMode) bool {
	if o.mesh != nil && o.shader != nil {
		o.applyModeColor(mode)
		return true
	}
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdTransformWire)
	if err != nil {
		slog.Error("failed to load occlusion bounds material", "error", err)
		return false
	}
	o.transform.Initialize(host.WorkGroup())
	o.transform.ResetDirty()
	o.mesh = host.MeshCache().DynamicMesh("editor_occlusion_bounds",
		occlusionBoundsVertices(nil), occlusionBoundsIndices())
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	o.shader = sd.(*shader_data_registry.ShaderDataEdTransformWire)
	o.applyModeColor(mode)
	o.shader.Deactivate()
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:         material,
		Mesh:             o.mesh,
		ShaderData:       o.shader,
		Transform:        &o.transform,
		ViewCuller:       &host.Cameras.Primary,
		OcclusionCulling: rendering.OcclusionCullingDisabled,
	})
	return true
}

func (o *occlusionBoundsVisualizer) applyModeColor(mode rendering.OcclusionVisualizationMode) {
	if o.shader == nil {
		return
	}
	if mode == rendering.OcclusionVisualizationOccludedBounds {
		o.shader.Color = matrix.ColorRed()
	} else {
		o.shader.Color = matrix.ColorYellow()
	}
}

func (o *occlusionBoundsVisualizer) hide() {
	if o.shader != nil {
		o.shader.Deactivate()
	}
}

func (o *occlusionBoundsVisualizer) show() {
	if o.shader != nil {
		o.shader.Activate()
	}
}

func occlusionBoundsVertices(bounds []rendering.OcclusionDebugBound) []rendering.Vertex {
	verts := make([]rendering.Vertex, maxEditorOcclusionBoxes*8)
	limit := min(len(bounds), maxEditorOcclusionBoxes)
	for i := 0; i < limit; i++ {
		aabbVertices(bounds[i].Bounds, verts[i*8:(i+1)*8])
	}
	hiddenPosition := matrix.Vec3Zero()
	if limit > 0 {
		hiddenPosition = bounds[limit-1].Bounds.Center
	}
	for i := limit * 8; i < len(verts); i++ {
		verts[i].Position = hiddenPosition
		verts[i].Normal = matrix.Vec3Forward()
		verts[i].Color = matrix.ColorWhite()
	}
	return verts
}

func aabbVertices(bounds graviton.AABB, out []rendering.Vertex) {
	minimum := bounds.Min()
	maximum := bounds.Max()
	points := [8]matrix.Vec3{
		{minimum.X(), minimum.Y(), minimum.Z()},
		{maximum.X(), minimum.Y(), minimum.Z()},
		{maximum.X(), maximum.Y(), minimum.Z()},
		{minimum.X(), maximum.Y(), minimum.Z()},
		{minimum.X(), minimum.Y(), maximum.Z()},
		{maximum.X(), minimum.Y(), maximum.Z()},
		{maximum.X(), maximum.Y(), maximum.Z()},
		{minimum.X(), maximum.Y(), maximum.Z()},
	}
	for i := range points {
		out[i].Position = points[i]
		out[i].Normal = matrix.Vec3Forward()
		out[i].Color = matrix.ColorWhite()
	}
}

func occlusionBoundsIndices() []uint32 {
	const perBox = 24
	indices := make([]uint32, maxEditorOcclusionBoxes*perBox)
	box := [perBox]uint32{
		0, 1, 1, 2, 2, 3, 3, 0,
		4, 5, 5, 6, 6, 7, 7, 4,
		0, 4, 1, 5, 2, 6, 3, 7,
	}
	for i := range maxEditorOcclusionBoxes {
		base := uint32(i * 8)
		for j := range box {
			indices[i*perBox+j] = base + box[j]
		}
	}
	return indices
}

/******************************************************************************/
/* occlusion.go                                                               */
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
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"strings"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

type OcclusionCullingMode int

type OcclusionRuntimeMode int

const (
	OcclusionRuntimeConservative OcclusionRuntimeMode = iota
	OcclusionRuntimeOff
	OcclusionRuntimeStatsOnly
	OcclusionRuntimeAggressive
)

type OcclusionVisualizationMode int

const (
	OcclusionVisualizationNone OcclusionVisualizationMode = iota
	OcclusionVisualizationTestedBounds
	OcclusionVisualizationOccludedBounds
	OcclusionVisualizationHiZMipPreview
)

const (
	OcclusionCullingDefault OcclusionCullingMode = iota
	OcclusionCullingEnabled
	OcclusionCullingDisabled
)

const (
	DefaultOcclusionMinExtent         matrix.Float = 0.05
	DefaultOcclusionMinCameraDistance matrix.Float = 0.25
	DefaultOcclusionNearPlanePadding  matrix.Float = 0.001
	DefaultOcclusionMinProjectedPx    matrix.Float = 2.0
)

type OcclusionTuning struct {
	MinExtent         matrix.Float
	MinCameraDistance matrix.Float
	NearPlanePadding  matrix.Float
	MinProjectedPx    matrix.Float
	DepthBias         matrix.Float
	RectPadPx         matrix.Float
	MinRectPx         matrix.Float
	MissingFar        matrix.Float
}

func DefaultOcclusionTuning() OcclusionTuning {
	return OcclusionTuning{
		MinExtent:         DefaultOcclusionMinExtent,
		MinCameraDistance: DefaultOcclusionMinCameraDistance,
		NearPlanePadding:  DefaultOcclusionNearPlanePadding,
		MinProjectedPx:    DefaultOcclusionMinProjectedPx,
		DepthBias:         DefaultOcclusionDepthBias,
		RectPadPx:         DefaultOcclusionRectPadPx,
		MinRectPx:         DefaultOcclusionMinRectPx,
		MissingFar:        DefaultOcclusionMissingFar,
	}
}

func (m OcclusionRuntimeMode) Tuning() OcclusionTuning {
	tuning := DefaultOcclusionTuning()
	if m == OcclusionRuntimeAggressive {
		tuning.MinExtent = DefaultOcclusionMinExtent * 0.25
		tuning.MinCameraDistance = DefaultOcclusionMinCameraDistance * 0.25
		tuning.NearPlanePadding = DefaultOcclusionNearPlanePadding * 0.5
		tuning.MinProjectedPx = 0
		tuning.DepthBias = DefaultOcclusionDepthBias * 0.25
		tuning.RectPadPx = 0
		tuning.MinRectPx = 0
	}
	return tuning
}

var StringOcclusionCullingMode = map[string]OcclusionCullingMode{
	"Default":  OcclusionCullingDefault,
	"Enabled":  OcclusionCullingEnabled,
	"Disabled": OcclusionCullingDisabled,
}

func (m OcclusionRuntimeMode) String() string {
	switch m {
	case OcclusionRuntimeOff:
		return "off"
	case OcclusionRuntimeStatsOnly:
		return "stats-only"
	case OcclusionRuntimeAggressive:
		return "aggressive"
	default:
		return "conservative"
	}
}

func ParseOcclusionRuntimeMode(value string) (OcclusionRuntimeMode, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "off", "disabled", "disable", "none":
		return OcclusionRuntimeOff, true
	case "stats", "stats-only", "statsonly", "statistics":
		return OcclusionRuntimeStatsOnly, true
	case "conservative", "safe", "on", "enabled", "enable", "":
		return OcclusionRuntimeConservative, true
	case "aggressive", "fast":
		return OcclusionRuntimeAggressive, true
	default:
		return OcclusionRuntimeConservative, false
	}
}

func (m OcclusionRuntimeMode) QueuesWork() bool {
	return m == OcclusionRuntimeConservative || m == OcclusionRuntimeAggressive
}

func (m OcclusionVisualizationMode) String() string {
	switch m {
	case OcclusionVisualizationTestedBounds:
		return "tested-bounds"
	case OcclusionVisualizationOccludedBounds:
		return "occluded-bounds"
	case OcclusionVisualizationHiZMipPreview:
		return "hiz-mip-preview"
	default:
		return "none"
	}
}

func ParseOcclusionVisualizationMode(value string) (OcclusionVisualizationMode, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none", "off", "":
		return OcclusionVisualizationNone, true
	case "tested", "tested-bounds", "testedbounds", "bounds":
		return OcclusionVisualizationTestedBounds, true
	case "occluded", "occluded-bounds", "occludedbounds":
		return OcclusionVisualizationOccludedBounds, true
	case "hiz", "hi-z", "hiz-preview", "hiz-mip-preview", "mip", "mip-preview":
		return OcclusionVisualizationHiZMipPreview, true
	default:
		return OcclusionVisualizationNone, false
	}
}

func ParseOcclusionCullingMode(value string) OcclusionCullingMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "enabled", "enable", "on", "true", "yes", "forceenabled", "force_enabled":
		return OcclusionCullingEnabled
	case "disabled", "disable", "off", "false", "no", "forcedisabled", "force_disabled":
		return OcclusionCullingDisabled
	default:
		return OcclusionCullingDefault
	}
}

func (m OcclusionCullingMode) Enabled() bool  { return m == OcclusionCullingEnabled }
func (m OcclusionCullingMode) Disabled() bool { return m == OcclusionCullingDisabled }

func (m *Material) EnableOcclusionCulling()  { m.OcclusionCulling = OcclusionCullingEnabled }
func (m *Material) DisableOcclusionCulling() { m.OcclusionCulling = OcclusionCullingDisabled }
func (m *Material) DefaultOcclusionCulling() { m.OcclusionCulling = OcclusionCullingDefault }

func (d *Drawing) EnableOcclusionCulling()  { d.OcclusionCulling = OcclusionCullingEnabled }
func (d *Drawing) DisableOcclusionCulling() { d.OcclusionCulling = OcclusionCullingDisabled }
func (d *Drawing) DefaultOcclusionCulling() { d.OcclusionCulling = OcclusionCullingDefault }

func (m *Material) transparentByDefault() bool {
	if m == nil {
		return true
	}
	if m.HasTransparentSuffix() {
		return true
	}
	if strings.Contains(strings.ToLower(m.Id), "transparent") {
		return true
	}
	rp := m.RenderPass()
	return rp != nil && strings.Contains(strings.ToLower(rp.construction.Name), "transparent")
}

func (m *Material) defaultOcclusionEligible() bool {
	if m == nil || m.renderPass == nil {
		return false
	}
	passName := strings.ToLower(m.renderPass.construction.Name)
	materialID := strings.ToLower(m.Id)
	if m.renderPass.IsShadowPass() ||
		m.transparentByDefault() ||
		strings.Contains(passName, "ui") ||
		strings.Contains(materialID, "ui") ||
		strings.Contains(passName, "gizmo") ||
		strings.Contains(materialID, "gizmo") ||
		strings.Contains(passName, "particle") ||
		strings.Contains(materialID, "particle") {
		return false
	}
	return m.hasDepthAttachment() &&
		m.pipelineInfo.DepthStencil.DepthTestEnable &&
		m.pipelineInfo.DepthStencil.DepthWriteEnable
}

func (m *Material) hasDepthAttachment() bool {
	if m == nil || m.renderPass == nil {
		return false
	}
	subpass := int(m.pipelineInfo.GraphicsPipeline.Subpass)
	if subpass >= 0 && subpass < len(m.renderPass.construction.SubpassDescriptions) {
		return len(m.renderPass.construction.SubpassDescriptions[subpass].DepthStencilAttachment) > 0
	}
	for i := range m.renderPass.construction.SubpassDescriptions {
		if len(m.renderPass.construction.SubpassDescriptions[i].DepthStencilAttachment) > 0 {
			return true
		}
	}
	return false
}

func (d *DrawInstanceGroup) updateOcclusionEligibility(instanceBase *ShaderDataBase, tuning OcclusionTuning) {
	if instanceBase == nil {
		return
	}
	visibility := instanceBase.VisibilityState()
	visibility.OcclusionEligible = false
	if d.MaterialInstance == nil || instanceBase.IsDestroyed() {
		visibility.LastOcclusionVisible = true
		return
	}
	mode := d.MaterialInstance.OcclusionCulling
	if instanceBase.occlusionCulling != OcclusionCullingDefault {
		mode = instanceBase.occlusionCulling
	}
	if mode.Disabled() {
		visibility.LastOcclusionVisible = true
		return
	}
	if !mode.Enabled() && !d.MaterialInstance.defaultOcclusionEligible() {
		visibility.LastOcclusionVisible = true
		return
	}
	if !d.MaterialInstance.hasDepthAttachment() {
		visibility.LastOcclusionVisible = true
		return
	}
	if !d.viewCullerAllowsOcclusion(instanceBase.renderBounds(), tuning) {
		visibility.LastOcclusionVisible = true
		return
	}
	visibility.OcclusionEligible = true
}

func (d *DrawInstanceGroup) viewCullerAllowsOcclusion(bounds graviton.AABB, tuning OcclusionTuning) bool {
	container, ok := d.viewCuller.(*cameras.Container)
	if !ok || container == nil || !container.OcclusionCullingEnabled() || !container.IsValid() {
		return false
	}
	camera := container.Camera
	if camera == nil ||
		camera.Width() <= 0 ||
		camera.Height() <= 0 ||
		camera.FarPlane() <= camera.NearPlane() ||
		camera.Position().IsNaN() ||
		camera.Position().IsInf(0) ||
		camera.Forward().IsNaN() ||
		camera.Forward().IsInf(0) ||
		camera.Forward().IsZero() {
		return false
	}
	if bounds.Extent.LongestAxisValue() < tuning.MinExtent {
		return false
	}
	forward := camera.Forward().Normal()
	centerDistance := matrix.Vec3Dot(bounds.Center.Subtract(camera.Position()), forward)
	extentRadius := bounds.Extent.X()*matrix.Abs(forward.X()) +
		bounds.Extent.Y()*matrix.Abs(forward.Y()) +
		bounds.Extent.Z()*matrix.Abs(forward.Z())
	closestDistance := centerDistance - extentRadius
	nearPlane := matrix.Float(camera.NearPlane())
	if nearPlane < 0 {
		nearPlane = 0
	}
	if closestDistance <= nearPlane+tuning.NearPlanePadding {
		return false
	}
	if closestDistance <= tuning.MinCameraDistance {
		return false
	}
	screenMin := matrix.Float(camera.Width())
	if h := matrix.Float(camera.Height()); h < screenMin {
		screenMin = h
	}
	if tuning.MinProjectedPx > 0 && screenMin > 0 {
		projectedExtentPx := bounds.Extent.LongestAxisValue() * screenMin / closestDistance
		if projectedExtentPx < tuning.MinProjectedPx {
			return false
		}
	}
	return true
}

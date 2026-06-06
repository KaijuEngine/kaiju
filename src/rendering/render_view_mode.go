/******************************************************************************/
/* render_view_mode.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "strings"

type RenderViewPipelineOverride int

const (
	RenderViewPipelineOverrideNone RenderViewPipelineOverride = iota
	RenderViewPipelineOverrideWireframe
)

type RenderViewModeSelection struct {
	Mode             RenderViewMode
	Material         *Material
	PipelineOverride RenderViewPipelineOverride
}

func (m RenderViewMode) String() string {
	switch m {
	case RenderViewModeNormal:
		return "Normal"
	case RenderViewModeWireframe:
		return "Wireframe"
	case RenderViewModeUnlit:
		return "Unlit"
	case RenderViewModeProfile:
		return "Profile"
	default:
		return "Unknown"
	}
}

func (m RenderViewMode) Valid() bool {
	switch m {
	case RenderViewModeNormal, RenderViewModeWireframe, RenderViewModeUnlit, RenderViewModeProfile:
		return true
	default:
		return false
	}
}

func ParseRenderViewMode(value string) (RenderViewMode, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default", "normal":
		return RenderViewModeNormal, true
	case "wire", "wireframe":
		return RenderViewModeWireframe, true
	case "unlit":
		return RenderViewModeUnlit, true
	case "profile", "profiles", "profile-style", "profilestyle":
		return RenderViewModeProfile, true
	default:
		return RenderViewModeNormal, false
	}
}

func ResolveRenderViewModeSelection(view *RenderView, material *Material, features GPUPhysicalDeviceFeatures) RenderViewModeSelection {
	mode := RenderViewModeNormal
	if view != nil {
		mode = view.ViewMode()
	}
	return ResolveRenderViewModeSelectionForMode(mode, material, features)
}

func ResolveRenderViewModeSelectionForMode(mode RenderViewMode, material *Material, features GPUPhysicalDeviceFeatures) RenderViewModeSelection {
	selection := RenderViewModeSelection{
		Mode:     mode,
		Material: material,
	}
	if material == nil || mode == RenderViewModeNormal {
		return selection
	}
	if override := material.compatibleRenderViewModeOverride(mode); override != nil {
		selection.Material = override
		return selection
	}
	if mode == RenderViewModeProfile {
		if override := material.compatibleRenderViewModeOverride(RenderViewModeUnlit); override != nil {
			selection.Material = override
			return selection
		}
	}
	if mode == RenderViewModeWireframe && features.FillModeNonSolid {
		selection.PipelineOverride = RenderViewPipelineOverrideWireframe
	}
	return selection
}

func destroyRenderViewResources(view *RenderView, device *GPUDevice, drawings *Drawings) {
	if view == nil {
		return
	}
	if drawings != nil {
		drawings.DestroyViewState(device, view)
	}
	if device != nil {
		device.DestroyRenderViewResources(view)
	}
}

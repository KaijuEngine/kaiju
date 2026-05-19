/******************************************************************************/
/* stage_workspace_occlusion.go                                               */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/

package stage_workspace

import (
	"fmt"

	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

func (w *StageWorkspace) occlusionDevice() (*rendering.GPUDevice, bool) {
	if w.Host == nil || w.Host.Window == nil || w.Host.Window.GpuInstance == nil ||
		len(w.Host.Window.GpuInstance.Devices) == 0 {
		return nil, false
	}
	return w.Host.Window.GpuInstance.PrimaryDevice(), true
}

func (w *StageWorkspace) setOcclusionMode(e *document.Element) {
	defer tracing.NewRegion("StageWorkspace.setOcclusionMode").End()
	device, ok := w.occlusionDevice()
	if !ok {
		return
	}
	mode, ok := rendering.ParseOcclusionRuntimeMode(e.Attribute("data-mode"))
	if !ok {
		return
	}
	device.SetOcclusionRuntimeMode(mode)
	w.updateOcclusionDebugUI()
}

func (w *StageWorkspace) setOcclusionVisualization(e *document.Element) {
	defer tracing.NewRegion("StageWorkspace.setOcclusionVisualization").End()
	device, ok := w.occlusionDevice()
	if !ok {
		return
	}
	mode, ok := rendering.ParseOcclusionVisualizationMode(e.Attribute("data-visual"))
	if !ok {
		return
	}
	device.SetOcclusionVisualizationMode(mode)
	w.updateOcclusionDebugUI()
}

func (w *StageWorkspace) updateOcclusionDebugUI() {
	defer tracing.NewRegion("StageWorkspace.updateOcclusionDebugUI").End()
	device, ok := w.occlusionDevice()
	if !ok {
		return
	}
	w.updateOcclusionModeButtons(device.OcclusionRuntimeMode())
	w.updateOcclusionVisualizationButtons(device.OcclusionVisualizationMode())
	w.updateOcclusionStatsLabel(device)
}

func (w *StageWorkspace) updateOcclusionModeButtons(active rendering.OcclusionRuntimeMode) {
	changed := false
	for _, id := range []string{"occModeOff", "occModeStats", "occModeConservative", "occModeAggressive"} {
		elm, ok := w.Doc.GetElementById(id)
		if !ok {
			continue
		}
		mode, ok := rendering.ParseOcclusionRuntimeMode(elm.Attribute("data-mode"))
		shouldBeActive := ok && mode == active
		if elm.HasClass("active") != shouldBeActive {
			changed = true
		}
		if shouldBeActive {
			w.Doc.SetElementClassesWithoutApply(elm, "occlusionTool", "active")
		} else {
			w.Doc.SetElementClassesWithoutApply(elm, "occlusionTool")
		}
	}
	if changed {
		w.Doc.ApplyStyles()
	}
}

func (w *StageWorkspace) updateOcclusionVisualizationButtons(active rendering.OcclusionVisualizationMode) {
	changed := false
	for _, id := range []string{"occVisNone", "occVisTested", "occVisOccluded", "occVisHiZ"} {
		elm, ok := w.Doc.GetElementById(id)
		if !ok {
			continue
		}
		mode, ok := rendering.ParseOcclusionVisualizationMode(elm.Attribute("data-visual"))
		shouldBeActive := ok && mode == active
		if elm.HasClass("active") != shouldBeActive {
			changed = true
		}
		if shouldBeActive {
			w.Doc.SetElementClassesWithoutApply(elm, "occlusionTool", "active")
		} else {
			w.Doc.SetElementClassesWithoutApply(elm, "occlusionTool")
		}
	}
	if changed {
		w.Doc.ApplyStyles()
	}
}

func (w *StageWorkspace) updateOcclusionStatsLabel(device *rendering.GPUDevice) {
	elm, ok := w.Doc.GetElementById("occlusionStats")
	if !ok || elm.InnerLabel() == nil {
		return
	}
	counters := w.Host.Drawings.VisibilityCounters()
	hiZ := "Hi-Z: not ready"
	if levels, mip, width, height, ok := device.OcclusionHiZLevelInfo(); ok {
		hiZ = fmt.Sprintf("Hi-Z: %d mips, preview %d (%dx%d)", levels, mip, width, height)
	}
	elm.InnerLabel().SetText(fmt.Sprintf(
		"total %d | visible %d | frustum %d | tested %d | hidden %d\n%s",
		counters.TotalInstances,
		counters.Visible,
		counters.FrustumCulled,
		counters.OcclusionTested,
		counters.OcclusionCulled,
		hiZ))
}

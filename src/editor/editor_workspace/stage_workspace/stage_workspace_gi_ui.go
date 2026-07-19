package stage_workspace

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"kaijuengine.com/editor/editor_overlay/content_selector"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/windowing"
)

type WorkspaceGIUI struct {
	workspace    *StageWorkspace
	doc          *document.Document
	area         *document.Element
	elapsed      float64
	staleElapsed float64
	probeState   string
	cancel       context.CancelFunc
	baking       bool
}

func (g *WorkspaceGIUI) setupFuncs() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"giOverrideChanged":    g.overrideChanged,
		"giRuntimeChanged":     g.runtimeChanged,
		"selectGIProbe":        g.selectProbe,
		"clearGIProbe":         g.clearProbe,
		"giProbeDrop":          g.probeDrop,
		"giProbeDragEnter":     g.probeDragEnter,
		"giProbeDragExit":      g.probeDragExit,
		"giBakeSettingChanged": g.bakeSettingChanged,
		"bakeGI":               g.bake,
		"bakeGIAs":             g.bakeAs,
		"cancelGIBake":         g.cancelBake,
	}
}

func (g *WorkspaceGIUI) setup(w *StageWorkspace) {
	g.workspace = w
	g.doc = w.giDoc
	g.area, _ = g.doc.GetElementById("giArea")
	if g.area != nil {
		panel := g.area.UI.ToPanel()
		panel.SetScrollDirection(ui.PanelScrollDirectionVertical)
		panel.SetOverflow(ui.OverflowScroll)
		panel.DontFitContentHeight()
		g.area.UI.GenerateScissor()
	}
	g.syncInputs()
	g.hideBakeProgress()
}

func (g *WorkspaceGIUI) open() {
	if g.area != nil {
		g.area.UI.Hide()
	}
	g.syncInputs()
}

func (g *WorkspaceGIUI) extendHeight() {
	if g.area != nil {
		g.doc.SetElementClasses(g.area, "edPanelBg", "sideBarTall")
	}
}

func (g *WorkspaceGIUI) standardHeight() {
	if g.area != nil {
		g.doc.SetElementClasses(g.area, "edPanelBg", "sideBarStandard")
	}
}

func (g *WorkspaceGIUI) update(deltaTime float64) {
	g.elapsed += deltaTime
	g.staleElapsed += deltaTime
	if g.elapsed < 0.5 || g.area == nil || !g.area.UI.Entity().IsActive() {
		return
	}
	g.elapsed = 0
	if g.staleElapsed >= 2 {
		g.staleElapsed = 0
		g.refreshProbeState()
	}
	g.syncStatus()
}

func (g *WorkspaceGIUI) current() stages.StageGlobalIllumination {
	return g.workspace.stageView.Manager().GlobalIllumination()
}

func (g *WorkspaceGIUI) effectiveSettings(stageGI stages.StageGlobalIllumination) gi.Settings {
	if stageGI.OverrideProjectSettings {
		return stageGI.Settings
	}
	return g.workspace.ed.Project().Settings.GlobalIllumination
}

func (g *WorkspaceGIUI) apply(next stages.StageGlobalIllumination, record bool) {
	previous := g.current()
	next.Normalize(g.workspace.ed.Project().Settings.GlobalIllumination)
	if reflect.DeepEqual(previous, next) {
		g.syncInputs()
		return
	}
	g.workspace.stageView.Manager().SetGlobalIllumination(next)
	var override *gi.Settings
	if next.OverrideProjectSettings {
		override = &next.Settings
	}
	if err := g.workspace.Host.GlobalIllumination().ApplyStageSettings(override, next.ProbeAsset); err != nil {
		g.setStatus("Preview failed: " + err.Error())
	}
	if record {
		g.workspace.ed.History().Add(&stageGISettingsHistory{ui: g, from: previous, to: next})
	}
	g.syncInputs()
}

func (g *WorkspaceGIUI) overrideChanged(e *document.Element) {
	next := g.current()
	next.OverrideProjectSettings = e.UI.ToCheckbox().IsChecked()
	if next.OverrideProjectSettings {
		next.Settings = g.workspace.ed.Project().Settings.GlobalIllumination
	}
	g.apply(next, true)
}

func (g *WorkspaceGIUI) runtimeChanged(e *document.Element) {
	next := g.current()
	if !next.OverrideProjectSettings {
		g.setStatus("Enable the stage override before editing runtime settings.")
		g.syncInputs()
		return
	}
	value := ""
	checked := false
	if e.UI.IsType(ui.ElementTypeSelect) {
		value = e.UI.ToSelect().Value()
	} else if e.UI.IsType(ui.ElementTypeCheckbox) {
		checked = e.UI.ToCheckbox().IsChecked()
	} else {
		value = e.UI.ToInput().Text()
	}
	settings, err := common_workspace.ApplyGISettingsField(next.Settings, e.Attribute("data-field"), value, checked)
	if err != nil {
		g.setStatus(err.Error())
		g.syncInputs()
		return
	}
	next.Settings = settings
	g.apply(next, true)
}

func (g *WorkspaceGIUI) selectProbe(*document.Element) {
	g.workspace.ed.BlurInterface()
	_, err := content_selector.Show(g.workspace.Host, content_database.GIProbe{}.TypeName(), g.workspace.ed.Cache(), func(id string) {
		g.workspace.ed.FocusInterface()
		next := g.current()
		next.ProbeAsset = id
		g.apply(next, true)
	}, g.workspace.ed.FocusInterface)
	if err != nil {
		g.workspace.ed.FocusInterface()
		g.setStatus(err.Error())
	}
}

func (g *WorkspaceGIUI) clearProbe(*document.Element) {
	next := g.current()
	next.ProbeAsset = ""
	g.apply(next, true)
}

func (g *WorkspaceGIUI) draggedProbe() (string, bool) {
	drag, ok := windowing.DragData().(StageDragContent)
	if !ok {
		return "", false
	}
	cached, err := g.workspace.ed.Cache().Read(drag.id)
	if err != nil || cached.Config.Type != (content_database.GIProbe{}).TypeName() {
		return "", false
	}
	return cached.Id(), true
}

func (g *WorkspaceGIUI) probeDrop(e *document.Element) {
	g.doc.SetElementClasses(e)
	if id, ok := g.draggedProbe(); ok {
		next := g.current()
		next.ProbeAsset = id
		g.apply(next, true)
	}
}

func (g *WorkspaceGIUI) probeDragEnter(e *document.Element) {
	if _, ok := g.draggedProbe(); ok {
		g.doc.SetElementClasses(e, "dragHover")
	}
}

func (g *WorkspaceGIUI) probeDragExit(e *document.Element) {
	g.doc.SetElementClasses(e)
}

func (g *WorkspaceGIUI) bakeSettingChanged(e *document.Element) {
	next := g.current()
	bake := next.BakeSettings
	field := e.Attribute("data-field")
	value := e.UI.ToInput().Text()
	if e.UI.IsType(ui.ElementTypeSelect) {
		value = e.UI.ToSelect().Value()
	}
	floatValue := func() (matrix.Float, error) {
		v, err := strconv.ParseFloat(strings.TrimSpace(value), 32)
		return matrix.Float(v), err
	}
	var err error
	switch field {
	case "BoundsMode":
		var v int
		v, err = strconv.Atoi(value)
		bake.BoundsMode = stages.GIBakeBoundsMode(v)
	case "BoundsPadding":
		bake.BoundsPadding, err = floatValue()
	case "ProbeSpacing":
		bake.ProbeSpacing, err = floatValue()
	case "RaysPerProbe":
		var v uint64
		v, err = strconv.ParseUint(value, 10, 32)
		bake.RaysPerProbe = uint32(v)
	case "MaxRayDistance":
		bake.MaxRayDistance, err = floatValue()
	default:
		var v matrix.Float
		v, err = floatValue()
		if err == nil {
			switch field {
			case "ManualCenterX":
				bake.ManualCenter[0] = v
			case "ManualCenterY":
				bake.ManualCenter[1] = v
			case "ManualCenterZ":
				bake.ManualCenter[2] = v
			case "ManualSizeX":
				bake.ManualSize[0] = v
			case "ManualSizeY":
				bake.ManualSize[1] = v
			case "ManualSizeZ":
				bake.ManualSize[2] = v
			case "EnvironmentR":
				bake.EnvironmentColor[0] = v
			case "EnvironmentG":
				bake.EnvironmentColor[1] = v
			case "EnvironmentB":
				bake.EnvironmentColor[2] = v
			default:
				err = fmt.Errorf("unknown GI bake field %q", field)
			}
		}
	}
	if err == nil {
		err = validateStageGIBakeSettings(bake)
	}
	if err != nil {
		g.setStatus(err.Error())
		g.syncInputs()
		return
	}
	next.BakeSettings = bake
	g.apply(next, true)
}

func validateStageGIBakeSettings(s stages.StageGIBakeSettings) error {
	if s.BoundsMode > stages.GIBakeBoundsManual {
		return fmt.Errorf("invalid GI bake bounds mode")
	}
	if s.BoundsPadding < 0 {
		return fmt.Errorf("GI bake padding cannot be negative")
	}
	if s.ProbeSpacing <= 0 {
		return fmt.Errorf("GI bake spacing must be greater than zero")
	}
	if s.RaysPerProbe < 32 {
		return fmt.Errorf("GI bake requires at least 32 rays per probe")
	}
	if s.MaxRayDistance <= 0 {
		return fmt.Errorf("GI bake ray distance must be greater than zero")
	}
	if s.BoundsMode == stages.GIBakeBoundsManual && (s.ManualSize.X() <= 0 || s.ManualSize.Y() <= 0 || s.ManualSize.Z() <= 0) {
		return fmt.Errorf("manual GI bake size must be positive")
	}
	if s.EnvironmentColor.X() < 0 || s.EnvironmentColor.Y() < 0 || s.EnvironmentColor.Z() < 0 {
		return fmt.Errorf("GI environment radiance cannot be negative")
	}
	return nil
}

func (g *WorkspaceGIUI) syncInputs() {
	if g.doc == nil {
		return
	}
	stageGI := g.current()
	stageGI.Normalize(g.workspace.ed.Project().Settings.GlobalIllumination)
	settings := g.effectiveSettings(stageGI)
	bake := stageGI.BakeSettings
	g.setCheck("giOverride", stageGI.OverrideProjectSettings)
	g.setSelect("giPreset", int(settings.Preset))
	g.setSelect("giMode", int(settings.Mode))
	g.setSelect("giFallback", int(settings.Fallback))
	g.setInput("giGPUTime", settings.GPUTimeBudgetMS)
	g.setInput("giMemory", settings.MemoryBudgetMB)
	g.setInput("giCoverage", settings.CoverageDistance)
	g.setInput("giRuntimeSpacing", settings.ProbeSpacing)
	g.setInput("giResolveScale", settings.ResolveScale)
	g.setInput("giUpdateHz", settings.UpdateHz)
	g.setInput("giHistoryWeight", settings.HistoryWeight)
	g.setInput("giCascades", settings.CascadeCount)
	g.setInput("giRuntimeRays", settings.RaysPerProbe)
	g.setInput("giProbeUpdates", settings.MaxProbeUpdatesPerFrame)
	g.setSelect("giContactDetail", int(settings.ContactDetail))
	g.setSelect("giDynamicGeometry", int(settings.DynamicGeometry))
	g.setSelect("giEmissive", int(settings.EmissiveParticipation))
	g.setInput("giTransition", settings.ScenarioTransitionSeconds)
	g.setCheck("giAdaptive", settings.AdaptiveBudget)
	g.setSelect("giBakeBoundsMode", int(bake.BoundsMode))
	g.setInput("giBakePadding", bake.BoundsPadding)
	g.setInput("giBakeSpacing", bake.ProbeSpacing)
	g.setInput("giBakeRays", bake.RaysPerProbe)
	g.setInput("giBakeDistance", bake.MaxRayDistance)
	g.setInput("giBakeCenterX", bake.ManualCenter.X())
	g.setInput("giBakeCenterY", bake.ManualCenter.Y())
	g.setInput("giBakeCenterZ", bake.ManualCenter.Z())
	g.setInput("giBakeSizeX", bake.ManualSize.X())
	g.setInput("giBakeSizeY", bake.ManualSize.Y())
	g.setInput("giBakeSizeZ", bake.ManualSize.Z())
	g.setInput("giBakeEnvR", bake.EnvironmentColor.X())
	g.setInput("giBakeEnvG", bake.EnvironmentColor.Y())
	g.setInput("giBakeEnvB", bake.EnvironmentColor.Z())
	if elm, ok := g.doc.GetElementById("giProbeAsset"); ok && elm.InnerLabel() != nil {
		label := "None (GI Probe)"
		if stageGI.ProbeAsset != "" {
			label = stageGI.ProbeAsset
			if cached, err := g.workspace.ed.Cache().Read(stageGI.ProbeAsset); err == nil {
				label = cached.Config.Name + " (GI Probe)"
			}
		}
		elm.InnerLabel().SetText(label)
	}
	g.syncStatus()
}

func (g *WorkspaceGIUI) setInput(id string, value any) {
	if elm, ok := g.doc.GetElementById(id); ok {
		elm.UI.ToInput().SetTextWithoutEvent(fmt.Sprint(value))
	}
}
func (g *WorkspaceGIUI) setCheck(id string, value bool) {
	if elm, ok := g.doc.GetElementById(id); ok {
		elm.UI.ToCheckbox().SetCheckedWithoutEvent(value)
	}
}
func (g *WorkspaceGIUI) setSelect(id string, value int) {
	if elm, ok := g.doc.GetElementById(id); ok {
		elm.UI.ToSelect().PickOptionWithoutEvent(value)
	}
}
func (g *WorkspaceGIUI) setStatus(message string) {
	if elm, ok := g.doc.GetElementById("giStageStatus"); ok && elm.InnerLabel() != nil {
		elm.InnerLabel().SetText(message)
	}
}
func (g *WorkspaceGIUI) syncStatus() {
	stats := g.workspace.Host.GlobalIllumination().Stats()
	message := fmt.Sprintf("Provider: %s | probes: %d | memory: %d MB | GPU: %.2f ms", stats.Provider, stats.ActiveProbes, stats.MemoryUsedMB, stats.GPUTimeMS)
	if stats.FallbackReason != "" {
		message += " | " + stats.FallbackReason
	}
	if g.probeState != "" {
		message += " | " + g.probeState
	}
	if g.baking {
		message = "Baking GI... " + message
	}
	g.setStatus(message)
}

func (g *WorkspaceGIUI) setBakeProgress(completed, total int) {
	track, trackOK := g.doc.GetElementById("giBakeProgressTrack")
	fill, fillOK := g.doc.GetElementById("giBakeProgressFill")
	if !trackOK || !fillOK || total <= 0 {
		return
	}
	track.UI.Show()
	ratio := matrix.Float(completed) / matrix.Float(total)
	width := track.UI.Entity().Transform.WorldScale().X()
	fill.UI.Layout().ScaleWidth(max(matrix.Float(1), width*ratio))
}

func (g *WorkspaceGIUI) hideBakeProgress() {
	if g.doc == nil {
		return
	}
	if track, ok := g.doc.GetElementById("giBakeProgressTrack"); ok {
		track.UI.Hide()
	}
}

func (g *WorkspaceGIUI) refreshProbeState() {
	stageGI := g.current()
	if stageGI.ProbeAsset == "" {
		g.probeState = "no probe scenario assigned"
		return
	}
	data, err := g.workspace.Host.AssetDatabase().Read(stageGI.ProbeAsset)
	if err != nil {
		g.probeState = "probe scenario is missing"
		return
	}
	asset, err := gi.UnmarshalProbeAsset(data)
	if err != nil {
		g.probeState = "probe scenario is invalid"
		return
	}
	snapshot, err := g.captureBakeSnapshot(asset.Scenario)
	if err != nil {
		g.probeState = "unable to evaluate probe freshness: " + err.Error()
		return
	}
	switch {
	case asset.GeometryHash != snapshot.input.GeometryHash && asset.LightingHash != snapshot.input.LightingHash:
		g.probeState = "stale: geometry and lighting changed"
	case asset.GeometryHash != snapshot.input.GeometryHash:
		g.probeState = "stale: geometry or bake bounds changed"
	case asset.LightingHash != snapshot.input.LightingHash:
		g.probeState = "stale: lighting or environment changed"
	default:
		g.probeState = "probe scenario is up to date"
	}
}

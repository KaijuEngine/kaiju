package terrain_workspace

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"kaijuengine.com/editor/editor_overlay/content_selector"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace_registry"

	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	ID          = "terrain"
	DisplayName = "Terrain"

	terrainBrushValueScale  = matrix.Float(1.1)
	terrainBrushMinRadius   = matrix.Float(0.01)
	terrainBrushMaxRadius   = matrix.Float(10000)
	terrainBrushMinStrength = matrix.Float(0.01)
	terrainBrushMaxStrength = matrix.Float(10000)
	defaultBrushStrength    = matrix.Float(0.1)
)

func init() {
	editor_workspace_registry.Register(&TerrainWorkspace{})
}

type TerrainWorkspace struct {
	common_workspace.CommonWorkspace
	ed               editor_workspace.WorkspaceEditorInterface
	stageView        *editor_stage_view.StageView
	openTerrainSubID events.Id
	renamedSubID     events.Id

	activeID      string
	activeName    *document.Element
	createDialog  *document.Element
	status        *document.Element
	toolReadout   *document.Element
	tooltip       *document.Element
	radiusInput   *document.Element
	strengthInput *document.Element
	falloffSelect *document.Element

	createResolution    *document.Element
	createSizeX         *document.Element
	createSizeZ         *document.Element
	createFloorHeight   *document.Element
	createCeilingHeight *document.Element
	createInitialHeight *document.Element

	active       *terrain.Terrain
	mode         terrain.BrushMode
	painting     bool
	lastLocal    matrix.Vec2
	hasLastLocal bool
	stroke       *terrainStrokeCapture

	brushRingTransform matrix.Transform
	brushRingData      rendering.DrawInstance
}

func (w *TerrainWorkspace) ID() string          { return ID }
func (w *TerrainWorkspace) DisplayName() string { return DisplayName }
func (w *TerrainWorkspace) IsRequired() bool    { return false }

func (w *TerrainWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("TerrainWorkspace.Initialize").End()
	host := ed.Host()
	w.ed = ed
	w.stageView = ed.StageView()
	w.mode = terrain.BrushRaise
	funcs := map[string]func(*document.Element){
		"clickSelectTerrain": w.clickSelectTerrain,
		"buttonMouseEnter":   w.buttonMouseEnter,
		"buttonMouseLeave":   w.buttonMouseLeave,
		"buttonMouseMove":    w.buttonMouseMove,
		"clickCreateTerrain": w.clickCreateTerrain,
		"clickCancelCreate":  w.clickCancelCreate,
		"clickConfirmCreate": w.clickConfirmCreate,
		"clickToolRaise":     w.clickToolRaise,
		"clickToolLower":     w.clickToolLower,
		"clickToolSmooth":    w.clickToolSmooth,
		"clickSave":          w.clickSave,
		"clickRevert":        w.clickRevert,
		"brushChanged":       w.brushChanged,
		"renameTerrain":      w.renameTerrain,
	}
	if err := w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/terrain_workspace.go.html", nil, funcs); err != nil {
		return err
	}

	w.createDialog, _ = w.Doc.GetElementById("createTerrainDialog")
	w.status, _ = w.Doc.GetElementById("terrainStatus")
	w.toolReadout, _ = w.Doc.GetElementById("terrainToolReadout")
	w.tooltip, _ = w.Doc.GetElementById("tooltip")
	w.radiusInput, _ = w.Doc.GetElementById("brushRadius")
	w.strengthInput, _ = w.Doc.GetElementById("brushStrength")
	w.falloffSelect, _ = w.Doc.GetElementById("brushFalloff")
	w.createResolution, _ = w.Doc.GetElementById("createResolution")
	w.createSizeX, _ = w.Doc.GetElementById("createSizeX")
	w.createSizeZ, _ = w.Doc.GetElementById("createSizeZ")
	w.createFloorHeight, _ = w.Doc.GetElementById("createFloorHeight")
	w.createCeilingHeight, _ = w.Doc.GetElementById("createCeilingHeight")
	w.createInitialHeight, _ = w.Doc.GetElementById("createInitialHeight")
	w.activeName, _ = w.Doc.GetElementById("activeTerrainName")
	w.hideCreateDialog()
	w.setActiveName("Terrain name...")
	w.setStatus("Hover a terrain to inspect coordinates")
	w.refreshToolReadout()
	w.initBrushRing(host)
	// Subscribe to cross-workspace requests. The content workspace (or stage
	// content UI) publishes OnRequestOpenTerrain when the user right-clicks
	// a terrain asset; we open it and switch ourselves active.
	w.openTerrainSubID = ed.Events().OnRequestOpenTerrain.Add(func(terrainID string) {
		w.openTerrain(terrainID)
		ed.SelectWorkspace(ID)
	})
	// Subscribe to content renamed events so the active name updates if renamed from
	// another workspace like the content workspace.
	w.renamedSubID = ed.Events().OnContentRenamed.Add(w.contentRenamed)
	return nil
}

func (w *TerrainWorkspace) Shutdown() {
	defer tracing.NewRegion("TerrainWorkspace.Shutdown").End()
	w.destroyActive()
	if w.brushRingData != nil {
		w.brushRingData.Destroy()
		w.brushRingData = nil
	}
	if w.ed != nil {
		w.ed.Events().OnRequestOpenTerrain.Remove(w.openTerrainSubID)
		w.ed.Events().OnContentRenamed.Remove(w.renamedSubID)
	}
	w.CommonShutdown()
}

func (w *TerrainWorkspace) Open() {
	defer tracing.NewRegion("TerrainWorkspace.Open").End()
	w.CommonOpen()
	w.stageView.Open()
	if w.tooltip != nil {
		w.tooltip.UI.Hide()
	}
	w.stageView.SetViewportToolOwner(w)
}

func (w *TerrainWorkspace) Close() {
	defer tracing.NewRegion("TerrainWorkspace.Close").End()
	w.stageView.ClearViewportToolOwner(w)
	w.stageView.Close()
	w.CommonClose()
	w.hideBrushPreview()
	w.finishStroke()
	w.destroyActive()
}

func (w *TerrainWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{
		{
			Keys: []hid.KeyboardKey{hid.KeyboardKeyOpenBracket},
			Call: func() {
				if w.Host.Window.Keyboard.HasShift() {
					w.adjustBrushStrength(-1)
				} else {
					w.adjustBrushRadius(-1)
				}
			},
		},
		{
			Keys: []hid.KeyboardKey{hid.KeyboardKeyCloseBracket},
			Call: func() {
				if w.Host.Window.Keyboard.HasShift() {
					w.adjustBrushStrength(1)
				} else {
					w.adjustBrushRadius(1)
				}
			},
		},
	}
}

func (w *TerrainWorkspace) Update(deltaTime float64) {
	defer tracing.NewRegion("TerrainWorkspace.update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if windowing.HasDragData() {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.stageView.Update(deltaTime, w.ed.Project())
}

func (w *TerrainWorkspace) UpdateViewportTool(view *editor_stage_view.StageView) bool {
	defer tracing.NewRegion("TerrainWorkspace.UpdateViewportTool").End()
	if w.pointerOverUI() {
		w.hideBrushPreview()
		w.finishStroke()
		return true
	}
	if w.active == nil {
		w.hideBrushPreview()
		return false
	}
	m := &w.Host.Window.Mouse
	hit, ok := w.active.RayHit(view.Camera().RayCast(m))
	if ok {
		local := hit.LocalPoint
		w.setStatus("X " + fmtFloat(local.X()) + "  Z " + fmtFloat(local.Z()) + "  H " + fmtFloat(local.Y()))
		w.showBrushPreview(hit)
	} else {
		w.hideBrushPreview()
	}
	paintingButton := m.Pressed(hid.MouseButtonLeft) || m.Held(hid.MouseButtonLeft)
	if !paintingButton {
		w.finishStroke()
		return false
	}
	if !ok {
		w.hasLastLocal = false
		return true
	}
	if !w.painting {
		w.beginStroke()
	}
	w.paint(hit.LocalPoint.XZ())
	return true
}

func (w *TerrainWorkspace) clickSelectTerrain(*document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.clickSelectTerrain").End()
	w.ed.BlurInterface()
	content_selector.Show(w.Host, (content_database.Terrain{}).TypeName(), w.ed.Cache(), func(id string) {
		w.ed.FocusInterface()
		if id != "" {
			w.openTerrain(id)
			w.hideCreateDialog()
		}
	}, w.ed.FocusInterface)
}

func (w *TerrainWorkspace) clickCreateTerrain(*document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.clickCreateTerrain").End()
	if w.createDialog != nil {
		w.createDialog.UI.Show()
	}
}

func (w *TerrainWorkspace) clickCancelCreate(*document.Element) {
	w.hideCreateDialog()
}

func (w *TerrainWorkspace) clickConfirmCreate(*document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.clickConfirmCreate").End()
	cfg := terrain.TerrainConfig{
		Resolution:    w.readCreateInt(w.createResolution, 33),
		WorldSize:     matrix.NewVec2(w.readCreateFloat(w.createSizeX, 100), w.readCreateFloat(w.createSizeZ, 100)),
		MinHeight:     w.readCreateFloat(w.createFloorHeight, -100),
		MaxHeight:     w.readCreateFloat(w.createCeilingHeight, 100),
		InitialHeight: w.readCreateFloat(w.createInitialHeight, 0),
	}
	asset, err := terrain.NewAsset(cfg, nil)
	if err != nil {
		slog.Error("failed to create terrain asset", "error", err)
		return
	}
	data, err := asset.Serialize()
	if err != nil {
		slog.Error("failed to serialize terrain asset", "error", err)
		return
	}
	ids := content_database.ImportRaw("Terrain", data, content_database.Terrain{},
		w.ed.ProjectFileSystem(), w.ed.Cache())
	if len(ids) == 0 {
		return
	}
	w.ed.Events().OnContentAdded.Execute(ids)
	w.hideCreateDialog()
	w.openTerrain(ids[0])
}

func (w *TerrainWorkspace) clickToolRaise(*document.Element) {
	w.mode = terrain.BrushRaise
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) clickToolLower(*document.Element) {
	w.mode = terrain.BrushLower
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) clickToolSmooth(*document.Element) {
	w.mode = terrain.BrushSmooth
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) clickSave(*document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.clickSave").End()
	if w.active == nil || w.activeID == "" {
		return
	}
	asset, err := terrain.NewAssetFromHeightField(w.active.Config, w.active.HeightField)
	if err != nil {
		slog.Error("failed to create terrain asset from edited heightfield", "error", err)
		return
	}
	data, err := asset.Serialize()
	if err != nil {
		slog.Error("failed to serialize terrain edits", "error", err)
		return
	}
	cc, err := w.ed.Cache().Read(w.activeID)
	if err != nil {
		slog.Error("failed to locate terrain cache entry", "id", w.activeID, "error", err)
		return
	}
	mode := os.ModePerm
	if s, err := w.ed.ProjectFileSystem().Stat(cc.ContentPath()); err == nil {
		mode = s.Mode()
	}
	if err = w.ed.ProjectFileSystem().WriteFile(cc.ContentPath(), data, mode); err != nil {
		slog.Error("failed to save terrain asset", "id", w.activeID, "error", err)
		return
	}
	w.ed.Events().OnContentChangesSaved.Execute(w.activeID)
	w.setStatus("Saved " + cc.Config.Name)
}

func (w *TerrainWorkspace) clickRevert(*document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.clickRevert").End()
	if w.activeID != "" {
		w.openTerrain(w.activeID)
	}
}

func (w *TerrainWorkspace) brushChanged(*document.Element) {
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) openTerrain(id string) {
	defer tracing.NewRegion("TerrainWorkspace.openTerrain").End()
	cc, err := w.ed.Cache().Read(id)
	if err != nil {
		slog.Error("failed to read terrain cache entry", "id", id, "error", err)
		return
	}
	model, err := terrain.Load(w.Host, id)
	if err != nil {
		slog.Error("failed to load terrain", "id", id, "error", err)
		return
	}
	w.destroyActive()
	w.active = model
	w.activeID = id
	w.active.Transform.SetPosition(matrix.Vec3Zero())
	w.setActiveName(cc.Config.Name)
	w.setStatus("Terrain ready")
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) destroyActive() {
	if w.active == nil {
		return
	}
	w.active.Destroy(w.Host)
	w.active = nil
	w.activeID = ""
	w.painting = false
	w.hasLastLocal = false
	w.stroke = nil
	w.hideBrushPreview()
}

func (w *TerrainWorkspace) paint(local matrix.Vec2) {
	stroke := w.brushStroke(local)
	var dirty terrain.DirtyRegion
	if w.painting && w.hasLastLocal {
		w.captureLine(w.lastLocal, local, stroke)
		dirty = w.active.PaintLine(w.lastLocal, local, stroke)
	} else {
		w.captureStroke(stroke)
		dirty = w.active.Paint(stroke)
	}
	if w.stroke != nil && dirty.Valid {
		w.stroke.changed = true
	}
	w.painting = true
	w.lastLocal = local
	w.hasLastLocal = true
}

func (w *TerrainWorkspace) brushStroke(local matrix.Vec2) terrain.PaintStroke {
	radius := w.readBrushFloat(w.radiusInput, 2)
	return terrain.PaintStroke{
		Mode:     w.effectiveBrushMode(),
		Center:   local,
		Radius:   radius,
		Strength: w.readBrushFloat(w.strengthInput, defaultBrushStrength),
		Falloff:  w.readFalloff(),
		Spacing:  radius * 0.25,
	}
}

func (w *TerrainWorkspace) effectiveBrushMode() terrain.BrushMode {
	kb := &w.Host.Window.Keyboard
	return effectiveTerrainBrushMode(w.mode, kb.HasShift(), kb.HasCtrlOrMeta())
}

func (w *TerrainWorkspace) pointerOverUI() bool {
	return len(w.UiMan.Hovered()) > 0 || w.UiMan.Group.HasRequests()
}

func (w *TerrainWorkspace) adjustBrushRadius(direction int) {
	w.adjustBrushInput(w.radiusInput, 2, direction,
		terrainBrushMinRadius, terrainBrushMaxRadius, "Brush radius")
}

func (w *TerrainWorkspace) adjustBrushStrength(direction int) {
	w.adjustBrushInput(w.strengthInput, 0.25, direction,
		terrainBrushMinStrength, terrainBrushMaxStrength, "Brush strength")
}

func (w *TerrainWorkspace) adjustBrushInput(e *document.Element, fallback matrix.Float,
	direction int, minValue, maxValue matrix.Float, label string) {
	if e == nil || direction == 0 {
		return
	}
	if w.ed != nil && w.ed.IsInputFocused() {
		return
	}
	value := adjustTerrainBrushValue(
		w.readBrushFloat(e, fallback), direction, minValue, maxValue)
	e.UI.ToInput().SetTextWithoutEvent(fmtFloat(value))
	w.refreshToolReadout()
	w.setStatus(label + " " + fmtFloat(value))
}

func (w *TerrainWorkspace) initBrushRing(host *engine.Host) {
	w.brushRingTransform.Initialize(host.WorkGroup())
	mesh := rendering.NewMeshCircleWire(host.MeshCache(), 1, 96)
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdTransformWire)
	if err != nil {
		slog.Error("failed to load terrain brush ring material", "error", err)
		return
	}
	w.brushRingData = shader_data_registry.Create(material.Shader.ShaderDataName())
	w.brushRingData.(*shader_data_registry.ShaderDataEdTransformWire).Color =
		matrix.NewColor(0.2, 0.75, 1.0, 1.0)
	w.brushRingData.Deactivate()
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       mesh,
		ShaderData: w.brushRingData,
		Transform:  &w.brushRingTransform,
		ViewCuller: &host.Cameras.Primary,
	})
}

func (w *TerrainWorkspace) showBrushPreview(hit terrain.TerrainRayHit) {
	radius := w.readBrushFloat(w.radiusInput, 2)
	color := matrix.NewColor(0.2, 0.75, 1.0, 1.0)
	ringWidth := matrix.Max(radius*matrix.Float(0.035), matrix.Float(0.05))
	w.active.SetBrushPreview(hit.Point.XZ(), radius, ringWidth, color)
	if w.brushRingData == nil {
		return
	}
	w.brushRingTransform.SetPosition(hit.Point.Add(hit.Normal.Scale(0.025)))
	w.brushRingTransform.SetScale(matrix.NewVec3(radius, 1, radius))
	w.brushRingData.Activate()
}

func (w *TerrainWorkspace) hideBrushPreview() {
	if w.active != nil {
		w.active.ClearBrushPreview()
	}
	if w.brushRingData != nil {
		w.brushRingData.Deactivate()
	}
}

func (w *TerrainWorkspace) beginStroke() {
	w.painting = true
	w.hasLastLocal = false
	w.stroke = newTerrainStrokeCapture(w.active)
}

func (w *TerrainWorkspace) finishStroke() {
	if w.stroke != nil {
		if h := w.stroke.history(); h != nil {
			w.ed.History().Add(h)
		}
	}
	w.painting = false
	w.hasLastLocal = false
	w.stroke = nil
}

func (w *TerrainWorkspace) captureLine(from, to matrix.Vec2, stroke terrain.PaintStroke) {
	if w.stroke == nil {
		return
	}
	w.active.VisitPaintLineStamps(from, to, stroke, func(stamp terrain.PaintStroke) bool {
		w.captureStroke(stamp)
		return true
	})
}

func (w *TerrainWorkspace) captureStroke(stroke terrain.PaintStroke) {
	if w.stroke == nil {
		return
	}
	w.stroke.captureRegion(w.active.StrokeRegion(stroke))
}

func (w *TerrainWorkspace) readFalloff() terrain.BrushFalloff {
	if w.falloffSelect == nil {
		return terrain.FalloffSmooth
	}
	switch w.falloffSelect.UI.ToSelect().Value() {
	case "linear":
		return terrain.FalloffLinear
	case "constant":
		return terrain.FalloffConstant
	case "smooth":
		fallthrough
	default:
		return terrain.FalloffSmooth
	}
}

func (w *TerrainWorkspace) readBrushFloat(e *document.Element, fallback matrix.Float) matrix.Float {
	if e == nil {
		return fallback
	}
	return parseFloat(e.UI.ToInput().Text(), fallback)
}

func (w *TerrainWorkspace) readCreateFloat(e *document.Element, fallback matrix.Float) matrix.Float {
	if e == nil {
		return fallback
	}
	return parseFloat(e.UI.ToInput().Text(), fallback)
}

func (w *TerrainWorkspace) readCreateInt(e *document.Element, fallback int) int {
	if e == nil {
		return fallback
	}
	v, err := strconv.Atoi(e.UI.ToInput().Text())
	if err != nil || v < 2 {
		return fallback
	}
	return v
}

func (w *TerrainWorkspace) hideCreateDialog() {
	if w.createDialog != nil {
		w.createDialog.UI.Hide()
	}
}

func (w *TerrainWorkspace) setActiveName(text string) {
	if w.activeName != nil {
		if text == "" {
			text = "Terrain name..."
		}
		w.activeName.UI.ToInput().SetTextWithoutEvent(text)
	}
}

func (w *TerrainWorkspace) setStatus(text string) {
	if w.status != nil {
		w.status.InnerLabel().SetText(text)
	}
}

func (w *TerrainWorkspace) refreshToolReadout() {
	if w.toolReadout == nil {
		return
	}
	tool := "Raise"
	switch w.mode {
	case terrain.BrushLower:
		tool = "Lower"
	case terrain.BrushSmooth:
		tool = "Smooth"
	}
	w.toolReadout.InnerLabel().SetText(tool + " / R " +
		fmtFloat(w.readBrushFloat(w.radiusInput, 2)) + " / S " +
		fmtFloat(w.readBrushFloat(w.strengthInput, defaultBrushStrength)))
}

func parseFloat(text string, fallback matrix.Float) matrix.Float {
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return fallback
	}
	return matrix.Float(v)
}

func adjustTerrainBrushValue(value matrix.Float, direction int, minValue, maxValue matrix.Float) matrix.Float {
	if direction > 0 {
		value *= terrainBrushValueScale
	} else if direction < 0 {
		value /= terrainBrushValueScale
	}
	return matrix.Clamp(value, minValue, maxValue)
}

func effectiveTerrainBrushMode(mode terrain.BrushMode, smooth, invert bool) terrain.BrushMode {
	if smooth {
		return terrain.BrushSmooth
	}
	if invert {
		switch mode {
		case terrain.BrushRaise:
			return terrain.BrushLower
		case terrain.BrushLower:
			return terrain.BrushRaise
		}
	}
	return mode
}

func fmtFloat(v matrix.Float) string {
	return strconv.FormatFloat(float64(v), 'f', 2, 32)
}

var _ editor_stage_view.ViewportToolOwner = (*TerrainWorkspace)(nil)

type terrainStrokeCapture struct {
	target  *terrain.Terrain
	before  map[int]matrix.Float
	region  terrain.DirtyRegion
	changed bool
}

func newTerrainStrokeCapture(target *terrain.Terrain) *terrainStrokeCapture {
	return &terrainStrokeCapture{
		target: target,
		before: make(map[int]matrix.Float),
	}
}

func (c *terrainStrokeCapture) captureRegion(region terrain.DirtyRegion) {
	if c == nil || c.target == nil || !region.Valid {
		return
	}
	field := c.target.HeightField
	region = region.Expand(0, field.Resolution)
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			idx := x + z*field.Resolution
			if _, ok := c.before[idx]; !ok {
				c.before[idx] = field.Height(x, z)
			}
		}
	}
	c.region = mergeTerrainRegions(c.region, region)
}

func (c *terrainStrokeCapture) history() *terrainStrokeHistory {
	if c == nil || c.target == nil || !c.changed || !c.region.Valid {
		return nil
	}
	field := c.target.HeightField
	width := c.region.MaxX - c.region.MinX + 1
	height := c.region.MaxZ - c.region.MinZ + 1
	before := make([]matrix.Float, width*height)
	after := make([]matrix.Float, width*height)
	different := false
	for z := c.region.MinZ; z <= c.region.MaxZ; z++ {
		for x := c.region.MinX; x <= c.region.MaxX; x++ {
			outIdx := (x - c.region.MinX) + (z-c.region.MinZ)*width
			mapIdx := x + z*field.Resolution
			beforeHeight, ok := c.before[mapIdx]
			if !ok {
				beforeHeight = field.Height(x, z)
			}
			afterHeight := field.Height(x, z)
			before[outIdx] = beforeHeight
			after[outIdx] = afterHeight
			different = different || beforeHeight != afterHeight
		}
	}
	if !different {
		return nil
	}
	return &terrainStrokeHistory{
		target: c.target,
		region: c.region,
		before: before,
		after:  after,
	}
}

type terrainStrokeHistory struct {
	target *terrain.Terrain
	region terrain.DirtyRegion
	before []matrix.Float
	after  []matrix.Float
}

func (h *terrainStrokeHistory) Redo()   { h.apply(h.after) }
func (h *terrainStrokeHistory) Undo()   { h.apply(h.before) }
func (h *terrainStrokeHistory) Delete() {}
func (h *terrainStrokeHistory) Exit()   {}

func (h *terrainStrokeHistory) apply(heights []matrix.Float) {
	if h == nil || h.target == nil || !h.region.Valid {
		return
	}
	h.target.ApplyHeightRegion(h.region, heights)
}

func mergeTerrainRegions(a, b terrain.DirtyRegion) terrain.DirtyRegion {
	if !a.Valid {
		return b
	}
	if !b.Valid {
		return a
	}
	return terrain.DirtyRegion{
		MinX:  min(a.MinX, b.MinX),
		MinZ:  min(a.MinZ, b.MinZ),
		MaxX:  max(a.MaxX, b.MaxX),
		MaxZ:  max(a.MaxZ, b.MaxZ),
		Valid: true,
	}
}

func (w *TerrainWorkspace) renameTerrain(e *document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.renameTerrain").End()
	if w.activeID == "" || w.ed == nil {
		return
	}
	name := strings.TrimSpace(e.UI.ToInput().Text())
	if name == "" {
		slog.Warn("The name for the terrain can't be left blank, ignoring change")
		return
	}
	pfs := w.ed.ProjectFileSystem()
	if _, err := w.ed.Cache().Rename(w.activeID, name, pfs); err != nil {
		if !errors.Is(err, content_database.CacheContentNameEqual) {
			slog.Error("failed to rename the terrain", "id", w.activeID, "error", err)
		}
		return
	}
	w.ed.Events().OnContentRenamed.Execute(w.activeID)
	w.setStatus("Renamed terrain to " + name)
}

func (w *TerrainWorkspace) contentRenamed(id string) {
	if id != w.activeID || w.activeName == nil || w.ed == nil {
		return
	}
	cc, err := w.ed.Cache().Read(id)
	if err != nil {
		slog.Error("failed to read renamed terrain cache entry", "id", id, "error", err)
		return
	}
	w.setActiveName(cc.Config.Name)
}

func (w *TerrainWorkspace) buttonMouseEnter(e *document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.buttonMouseEnter").End()
	if w.tooltip == nil {
		return
	}
	text := e.Attribute("data-tooltip")
	if text == "" {
		w.tooltip.UI.Hide()
		return
	}
	w.tooltip.InnerLabel().SetText(text)
	w.tooltip.UI.Show()
}

func (w *TerrainWorkspace) buttonMouseMove(e *document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.buttonMouseMove").End()
	if w.tooltip == nil {
		return
	}
	ui := w.tooltip.UI
	if !ui.Entity().IsActive() {
		ui.Show()
	}
	p := w.Host.Window.Mouse.ScreenPosition()
	w.Host.RunOnMainThread(func() {
		ui.Layout().SetOffset(p.X()+12, p.Y()+18)
	})
}

func (w *TerrainWorkspace) buttonMouseLeave(e *document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.buttonMouseLeave").End()
	if w.tooltip != nil {
		w.tooltip.UI.Hide()
	}
}

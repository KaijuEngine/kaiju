/******************************************************************************/
/* terrain_workspace.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package terrain_workspace

import (
	"errors"
	"log/slog"
	"os"
	"reflect"
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

type TerrainToolMode int

const (
	TerrainToolHeightSculpt TerrainToolMode = iota
	TerrainToolTexturePaint
	TerrainToolSelection
	TerrainToolSettings
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
	heightToolRow *document.Element
	heightBrush   *document.Element
	textureRow    *document.Element
	radiusInput   *document.Element
	strengthInput *document.Element
	falloffSelect *document.Element
	modeBtns      []*document.Element
	heightBtns    []*document.Element
	textureBtns   []*document.Element

	textureLayerSelect     *document.Element
	textureLayerPalette    *document.Element
	textureSwatchTemplate  *document.Element
	textureLayerNameInput  *document.Element
	textureFilterSelect    *document.Element
	textureRadiusInput     *document.Element
	textureOpacityInput    *document.Element
	textureFalloffSelect   *document.Element
	textureTilingXInput    *document.Element
	textureTilingYInput    *document.Element
	textureWorldSizeXInput *document.Element
	textureWorldSizeYInput *document.Element
	textureTintRInput      *document.Element
	textureTintGInput      *document.Element
	textureTintBInput      *document.Element
	textureTintAInput      *document.Element
	textureSlopeMinInput   *document.Element
	textureSlopeMaxInput   *document.Element
	textureHeightMinInput  *document.Element
	textureHeightMaxInput  *document.Element
	textureNoiseInput      *document.Element
	textureJitterInput     *document.Element
	textureStampSelect     *document.Element
	textureSwatches        []*document.Element

	createResolution    *document.Element
	createSizeX         *document.Element
	createSizeZ         *document.Element
	createFloorHeight   *document.Element
	createCeilingHeight *document.Element
	createInitialHeight *document.Element

	active               *terrain.Terrain
	toolMode             TerrainToolMode
	mode                 terrain.BrushMode
	textureMode          terrain.TextureBrushMode
	textureLayer         int
	painting             bool
	lastLocal            matrix.Vec2
	hasLastLocal         bool
	stroke               *terrainStrokeCapture
	textureStrokeCapture *terrainTextureStrokeCapture

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
	w.toolMode = TerrainToolHeightSculpt
	w.mode = terrain.BrushRaise
	w.textureMode = terrain.TextureBrushPaint
	funcs := map[string]func(*document.Element){
		"clickSelectTerrain":  w.clickSelectTerrain,
		"buttonMouseEnter":    w.buttonMouseEnter,
		"buttonMouseLeave":    w.buttonMouseLeave,
		"buttonMouseMove":     w.buttonMouseMove,
		"clickCreateTerrain":  w.clickCreateTerrain,
		"clickCancelCreate":   w.clickCancelCreate,
		"clickConfirmCreate":  w.clickConfirmCreate,
		"clickModeHeight":     w.clickModeHeight,
		"clickModeTexture":    w.clickModeTexture,
		"clickToolRaise":      w.clickToolRaise,
		"clickToolLower":      w.clickToolLower,
		"clickToolSmooth":     w.clickToolSmooth,
		"clickTexturePaint":   w.clickTexturePaint,
		"clickTextureErase":   w.clickTextureErase,
		"clickTextureSmooth":  w.clickTextureSmooth,
		"clickTextureFill":    w.clickTextureFill,
		"clickTexturePick":    w.clickTexturePick,
		"clickFillLayer":      w.clickFillLayer,
		"clickTextureClear":   w.clickTextureClear,
		"clickAutoMaterial":   w.clickAutoMaterial,
		"clickAddLayer":       w.clickAddLayer,
		"clickReplaceLayer":   w.clickReplaceLayer,
		"clickRemoveLayer":    w.clickRemoveLayer,
		"clickLayerUp":        w.clickLayerUp,
		"clickLayerDown":      w.clickLayerDown,
		"clickLayerLock":      w.clickLayerLock,
		"clickLayerVisible":   w.clickLayerVisible,
		"clickLayerSolo":      w.clickLayerSolo,
		"clickWeightDebug":    w.clickWeightDebug,
		"clickTriplanar":      w.clickTriplanar,
		"clickLayerSwatch":    w.clickLayerSwatch,
		"clickSave":           w.clickSave,
		"clickRevert":         w.clickRevert,
		"brushChanged":        w.brushChanged,
		"textureBrushChanged": w.textureBrushChanged,
		"textureLayerChanged": w.textureLayerChanged,
		"renameTerrain":       w.renameTerrain,
	}
	if err := w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/terrain_workspace.go.html", nil, funcs); err != nil {
		return err
	}

	w.createDialog, _ = w.Doc.GetElementById("createTerrainDialog")
	w.status, _ = w.Doc.GetElementById("terrainStatus")
	w.toolReadout, _ = w.Doc.GetElementById("terrainToolReadout")
	w.tooltip, _ = w.Doc.GetElementById("tooltip")
	w.heightToolRow, _ = w.Doc.GetElementById("heightToolRow")
	w.heightBrush, _ = w.Doc.GetElementById("heightBrushControls")
	w.textureRow, _ = w.Doc.GetElementById("texturePaintRow")
	w.radiusInput, _ = w.Doc.GetElementById("brushRadius")
	w.strengthInput, _ = w.Doc.GetElementById("brushStrength")
	w.falloffSelect, _ = w.Doc.GetElementById("brushFalloff")
	w.textureLayerSelect, _ = w.Doc.GetElementById("textureLayerSelect")
	w.textureLayerPalette, _ = w.Doc.GetElementById("textureLayerPalette")
	w.textureSwatchTemplate, _ = w.Doc.GetElementById("textureLayerSwatchTemplate")
	w.textureLayerNameInput, _ = w.Doc.GetElementById("textureLayerName")
	w.textureFilterSelect, _ = w.Doc.GetElementById("textureFilter")
	w.textureRadiusInput, _ = w.Doc.GetElementById("textureBrushRadius")
	w.textureOpacityInput, _ = w.Doc.GetElementById("textureOpacity")
	w.textureFalloffSelect, _ = w.Doc.GetElementById("textureFalloff")
	w.textureTilingXInput, _ = w.Doc.GetElementById("textureTilingX")
	w.textureTilingYInput, _ = w.Doc.GetElementById("textureTilingY")
	w.textureWorldSizeXInput, _ = w.Doc.GetElementById("textureWorldSizeX")
	w.textureWorldSizeYInput, _ = w.Doc.GetElementById("textureWorldSizeY")
	w.textureTintRInput, _ = w.Doc.GetElementById("textureTintR")
	w.textureTintGInput, _ = w.Doc.GetElementById("textureTintG")
	w.textureTintBInput, _ = w.Doc.GetElementById("textureTintB")
	w.textureTintAInput, _ = w.Doc.GetElementById("textureTintA")
	w.textureSlopeMinInput, _ = w.Doc.GetElementById("textureSlopeMin")
	w.textureSlopeMaxInput, _ = w.Doc.GetElementById("textureSlopeMax")
	w.textureHeightMinInput, _ = w.Doc.GetElementById("textureHeightMin")
	w.textureHeightMaxInput, _ = w.Doc.GetElementById("textureHeightMax")
	w.textureNoiseInput, _ = w.Doc.GetElementById("textureNoise")
	w.textureJitterInput, _ = w.Doc.GetElementById("textureJitter")
	w.textureStampSelect, _ = w.Doc.GetElementById("textureStamp")
	w.createResolution, _ = w.Doc.GetElementById("createResolution")
	w.createSizeX, _ = w.Doc.GetElementById("createSizeX")
	w.createSizeZ, _ = w.Doc.GetElementById("createSizeZ")
	w.createFloorHeight, _ = w.Doc.GetElementById("createFloorHeight")
	w.createCeilingHeight, _ = w.Doc.GetElementById("createCeilingHeight")
	w.createInitialHeight, _ = w.Doc.GetElementById("createInitialHeight")
	w.activeName, _ = w.Doc.GetElementById("activeTerrainName")
	w.modeBtns = w.Doc.GetElementsByGroup("terrainMode")
	w.heightBtns = w.Doc.GetElementsByGroup("heightTool")
	w.textureBtns = w.Doc.GetElementsByGroup("textureTool")
	w.hideCreateDialog()
	if w.textureSwatchTemplate != nil {
		w.textureSwatchTemplate.UI.Hide()
	}
	w.setActiveName("Terrain name...")
	w.setStatus("Hover a terrain to inspect coordinates")
	w.refreshToolPanels()
	w.refreshLayerSelector()
	w.refreshLayerPalette()
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
	w.hideCreateDialog()
	if w.textureSwatchTemplate != nil {
		w.textureSwatchTemplate.UI.Hide()
	}
	w.refreshToolPanels()
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

func (w *TerrainWorkspace) highlightButtons(buttons []*document.Element, e *document.Element) {
	if w.Doc == nil || e == nil {
		return
	}
	for i := range buttons {
		w.Doc.SetElementClassesWithoutApply(buttons[i], "materialIcon")
	}
	w.Doc.SetElementClasses(e, "materialIcon", "active")
}

func (w *TerrainWorkspace) clickModeHeight(e *document.Element) {
	w.toolMode = TerrainToolHeightSculpt
	w.refreshToolPanels()
	w.highlightButtons(w.modeBtns, e)
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) clickModeTexture(e *document.Element) {
	w.toolMode = TerrainToolTexturePaint
	w.refreshToolPanels()
	w.highlightButtons(w.modeBtns, e)
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) clickToolRaise(e *document.Element) {
	w.toolMode = TerrainToolHeightSculpt
	w.mode = terrain.BrushRaise
	w.refreshToolPanels()
	w.refreshToolReadout()
	w.highlightButtons(w.heightBtns, e)
}

func (w *TerrainWorkspace) clickToolLower(e *document.Element) {
	w.toolMode = TerrainToolHeightSculpt
	w.mode = terrain.BrushLower
	w.refreshToolPanels()
	w.refreshToolReadout()
	w.highlightButtons(w.heightBtns, e)
}

func (w *TerrainWorkspace) clickToolSmooth(e *document.Element) {
	w.toolMode = TerrainToolHeightSculpt
	w.mode = terrain.BrushSmooth
	w.refreshToolPanels()
	w.refreshToolReadout()
	w.highlightButtons(w.heightBtns, e)
}

func (w *TerrainWorkspace) clickTexturePaint(e *document.Element) {
	w.textureMode = terrain.TextureBrushPaint
	w.refreshToolReadout()
	w.highlightButtons(w.textureBtns, e)
}

func (w *TerrainWorkspace) clickTextureErase(e *document.Element) {
	w.textureMode = terrain.TextureBrushErase
	w.refreshToolReadout()
	w.highlightButtons(w.textureBtns, e)
}

func (w *TerrainWorkspace) clickTextureSmooth(e *document.Element) {
	w.textureMode = terrain.TextureBrushSmoothWeights
	w.refreshToolReadout()
	w.highlightButtons(w.textureBtns, e)
}

func (w *TerrainWorkspace) clickTextureFill(e *document.Element) {
	w.textureMode = terrain.TextureBrushFill
	w.refreshToolReadout()
	w.highlightButtons(w.textureBtns, e)
}

func (w *TerrainWorkspace) clickTexturePick(e *document.Element) {
	w.textureMode = terrain.TextureBrushSample
	w.refreshToolReadout()
	w.highlightButtons(w.textureBtns, e)
}

func (w *TerrainWorkspace) clickFillLayer(*document.Element) {
	if w.active == nil {
		return
	}
	layer := w.readTextureLayer()
	before := w.active.LayerSetState()
	beforeLayer := layer
	dirty := w.active.FillLayer(layer)
	if dirty.Valid {
		w.addTextureLayerSetHistory(before, beforeLayer)
		w.setStatus("Filled terrain with selected texture layer")
	}
}

func (w *TerrainWorkspace) clickTextureClear(*document.Element) {
	if w.active == nil {
		return
	}
	layer := w.readTextureLayer()
	before := w.active.LayerSetState()
	beforeLayer := layer
	dirty := w.active.ClearLayer(layer)
	if dirty.Valid {
		w.addTextureLayerSetHistory(before, beforeLayer)
		w.setStatus("Cleared texture layer")
	}
}

func (w *TerrainWorkspace) clickAutoMaterial(*document.Element) {
	if w.active == nil || w.active.LayerCount() == 0 {
		return
	}
	before := w.active.LayerSetState()
	beforeLayer := w.textureLayer
	grass := min(0, w.active.LayerCount()-1)
	rock := min(1, w.active.LayerCount()-1)
	snow := min(2, w.active.LayerCount()-1)
	rules := terrain.TerrainAutoMaterialRules(terrain.TerrainAutoMaterialPreset{
		GrassLayer:    grass,
		RockLayer:     rock,
		SnowLayer:     snow,
		FlatSlopeMax:  28,
		CliffSlopeMin: 42,
		SnowHeightMin: w.readBrushFloat(w.textureHeightMinInput, w.active.HeightField.MaxHeight*0.65),
		NoiseStrength: w.readBrushFloat(w.textureNoiseInput, 0),
		NoiseScale:    max(w.readBrushFloat(w.textureRadiusInput, 2), matrix.Float(1)),
	})
	if dirty := w.active.ApplyAutoMaterialRules(rules); dirty.Valid {
		w.addTextureLayerSetHistory(before, beforeLayer)
		w.refreshLayerPalette()
		w.setStatus("Generated terrain material from height and slope")
	}
}

func (w *TerrainWorkspace) clickSave(*document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.clickSave").End()
	if w.active == nil || w.activeID == "" {
		return
	}
	w.applyTextureLayerSettings()
	texturesValid := w.validateLayerTextures()
	asset, err := terrain.NewAssetFromTerrain(w.active)
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
	if texturesValid {
		w.setStatus("Saved " + cc.Config.Name)
	}
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

func (w *TerrainWorkspace) textureBrushChanged(*document.Element) {
	w.applyTextureLayerSettings()
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) textureLayerChanged(*document.Element) {
	w.textureLayer = w.readTextureLayer()
	w.refreshTextureLayerFields()
	w.refreshLayerPaletteSelection()
	w.refreshToolReadout()
}

func (w *TerrainWorkspace) clickAddLayer(*document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.clickAddLayer").End()
	if w.active == nil {
		return
	}
	w.ed.BlurInterface()
	content_selector.Show(w.Host, (content_database.Texture{}).TypeName(), w.ed.Cache(), func(id string) {
		w.ed.FocusInterface()
		if id == "" {
			return
		}
		layer := terrain.NewTerrainLayer(id)
		layer.Name = w.textureNameForID(id)
		before := w.active.LayerSetState()
		beforeLayer := w.textureLayer
		w.textureLayer = w.active.AddLayer(layer)
		w.addTextureLayerSetHistory(before, beforeLayer)
		w.refreshLayerSelector()
		w.refreshTextureLayerFields()
		w.refreshLayerPalette()
		w.setLayerTextureStatus(id, "Added texture layer")
	}, w.ed.FocusInterface)
}

func (w *TerrainWorkspace) clickReplaceLayer(*document.Element) {
	defer tracing.NewRegion("TerrainWorkspace.clickReplaceLayer").End()
	if w.ed == nil || w.active == nil || w.active.LayerSet == nil {
		return
	}
	layer := w.readTextureLayer()
	if layer < 0 || layer >= w.active.LayerCount() {
		return
	}
	w.ed.BlurInterface()
	content_selector.Show(w.Host, (content_database.Texture{}).TypeName(), w.ed.Cache(), func(id string) {
		w.ed.FocusInterface()
		if id == "" {
			return
		}
		before := w.active.LayerSetState()
		beforeLayer := layer
		next := w.active.LayerSet.Layers[layer]
		oldName := strings.TrimSpace(next.Name)
		oldTexture := next.TextureContentID
		next.TextureContentID = id
		if oldName == "" || oldName == oldTexture || oldName == w.textureNameForID(oldTexture) {
			next.Name = w.textureNameForID(id)
		}
		if !w.active.SetLayer(layer, next) {
			return
		}
		w.textureLayer = layer
		w.addTextureLayerSetHistory(before, beforeLayer)
		w.refreshLayerSelector()
		w.refreshTextureLayerFields()
		w.refreshLayerPalette()
		w.setLayerTextureStatus(id, "Replaced texture layer")
	}, w.ed.FocusInterface)
}

func (w *TerrainWorkspace) clickRemoveLayer(*document.Element) {
	if w.active == nil || w.active.LayerCount() <= 1 {
		return
	}
	layer := w.readTextureLayer()
	before := w.active.LayerSetState()
	beforeLayer := layer
	if !w.active.RemoveLayer(layer) {
		return
	}
	w.textureLayer = min(layer, w.active.LayerCount()-1)
	w.addTextureLayerSetHistory(before, beforeLayer)
	w.refreshLayerSelector()
	w.refreshTextureLayerFields()
	w.refreshLayerPalette()
	w.setStatus("Removed texture layer")
}

func (w *TerrainWorkspace) clickLayerUp(*document.Element) {
	w.moveTextureLayer(-1)
}

func (w *TerrainWorkspace) clickLayerDown(*document.Element) {
	w.moveTextureLayer(1)
}

func (w *TerrainWorkspace) clickLayerLock(*document.Element) {
	w.toggleLayerFlag(func(layer *terrain.TerrainLayer) {
		layer.Locked = !layer.Locked
	}, "Toggled layer lock")
}

func (w *TerrainWorkspace) clickLayerVisible(*document.Element) {
	w.toggleLayerFlag(func(layer *terrain.TerrainLayer) {
		layer.Hidden = !layer.Hidden
	}, "Toggled layer visibility")
	if w.active != nil {
		w.active.RefreshTexturePreview()
	}
}

func (w *TerrainWorkspace) clickLayerSolo(*document.Element) {
	w.toggleLayerFlag(func(layer *terrain.TerrainLayer) {
		layer.Solo = !layer.Solo
	}, "Toggled layer solo preview")
	if w.active != nil {
		w.active.RefreshTexturePreview()
	}
}

func (w *TerrainWorkspace) clickWeightDebug(*document.Element) {
	if w.active == nil {
		return
	}
	layer := w.readTextureLayer()
	debug := w.active.WeightDebugRGBA(layer)
	if len(debug) == 0 {
		return
	}
	w.setStatus("Weight debug L " + strconv.Itoa(layer+1) + " " + strconv.Itoa(len(debug)/4) + " px")
}

func (w *TerrainWorkspace) clickTriplanar(*document.Element) {
	w.toggleLayerFlag(func(layer *terrain.TerrainLayer) {
		layer.TriplanarCliffs = !layer.TriplanarCliffs
		if layer.TriplanarSlope <= 0 {
			layer.TriplanarSlope = 45
		}
	}, "Toggled triplanar cliff projection")
}

func (w *TerrainWorkspace) toggleLayerFlag(update func(*terrain.TerrainLayer), status string) {
	if w.active == nil || w.active.LayerSet == nil || update == nil {
		return
	}
	layer := w.readTextureLayer()
	if layer < 0 || layer >= w.active.LayerCount() {
		return
	}
	before := w.active.LayerSetState()
	beforeLayer := layer
	next := w.active.LayerSet.Layers[layer]
	update(&next)
	if !w.active.SetLayer(layer, next) {
		return
	}
	w.addTextureLayerSetHistory(before, beforeLayer)
	w.refreshTextureLayerFields()
	w.refreshLayerPalette()
	w.setStatus(status)
}

func (w *TerrainWorkspace) moveTextureLayer(direction int) {
	if w.active == nil || direction == 0 {
		return
	}
	layer := w.readTextureLayer()
	next := layer + direction
	before := w.active.LayerSetState()
	beforeLayer := layer
	if !w.active.MoveLayer(layer, next) {
		return
	}
	w.textureLayer = next
	w.addTextureLayerSetHistory(before, beforeLayer)
	w.refreshLayerSelector()
	w.refreshTextureLayerFields()
	w.refreshLayerPalette()
	w.setStatus("Moved texture layer")
}

func (w *TerrainWorkspace) addTextureLayerSetHistory(before terrain.TerrainLayerSetState, beforeLayer int) {
	if w.ed == nil || w.active == nil {
		return
	}
	after := w.active.LayerSetState()
	h := &terrainLayerSetHistory{
		workspace:   w,
		target:      w.active,
		before:      before,
		after:       after,
		beforeLayer: beforeLayer,
		afterLayer:  w.textureLayer,
	}
	if h.changed() {
		w.ed.History().Add(h)
	}
}

func (w *TerrainWorkspace) clickLayerSwatch(e *document.Element) {
	if e == nil {
		return
	}
	layer, ok := textureLayerFromElement(e)
	if !ok {
		return
	}
	if w.active != nil && (layer < 0 || layer >= w.active.LayerCount()) {
		return
	}
	w.textureLayer = layer
	if w.textureLayerSelect != nil {
		w.textureLayerSelect.UI.ToSelect().PickOptionWithoutEvent(layer)
	}
	w.refreshTextureLayerFields()
	w.refreshLayerPaletteSelection()
	w.refreshToolReadout()
	w.setStatus("Selected texture layer " + strconv.Itoa(layer+1))
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
	w.textureLayer = 0
	w.refreshLayerSelector()
	w.refreshTextureLayerFields()
	w.refreshLayerPalette()
	w.refreshToolReadout()
	w.validateLayerTextures()
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
	w.textureStrokeCapture = nil
	w.refreshLayerSelector()
	w.refreshLayerPalette()
	w.hideBrushPreview()
}

func (w *TerrainWorkspace) paint(local matrix.Vec2) {
	if w.toolMode == TerrainToolTexturePaint {
		w.paintTexture(local)
		return
	}
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

func (w *TerrainWorkspace) paintTexture(local matrix.Vec2) {
	if w.active == nil {
		return
	}
	stroke := w.textureStroke(local)
	layer := w.readTextureLayer()
	var result terrain.TexturePaintResult
	if w.painting && w.hasLastLocal && stroke.Mode != terrain.TextureBrushFill &&
		stroke.Mode != terrain.TextureBrushSample {
		w.captureTextureLine(w.lastLocal, local, stroke)
		result = w.active.PaintTextureLine(layer, w.lastLocal, local, stroke)
	} else {
		w.captureTextureStroke(stroke)
		result = w.active.PaintTextureLayer(layer, stroke)
	}
	if result.Sampled {
		w.textureLayer = result.SampledLayer
		w.refreshLayerSelector()
		w.refreshTextureLayerFields()
		w.setStatus("Picked texture layer " + strconv.Itoa(result.SampledLayer+1))
	}
	if result.Dirty.Valid {
		if w.textureStrokeCapture != nil {
			w.textureStrokeCapture.markDirty(result.Dirty)
		}
		w.setStatus("Texture paint updated")
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

func (w *TerrainWorkspace) textureStroke(local matrix.Vec2) terrain.TexturePaintStroke {
	radius := w.readBrushFloat(w.textureRadiusInput, 2)
	return terrain.TexturePaintStroke{
		Mode:          w.textureMode,
		Center:        local,
		Radius:        radius,
		Strength:      w.readBrushFloat(w.textureOpacityInput, 1),
		Opacity:       1,
		TargetWeight:  1,
		Falloff:       w.readTextureFalloff(),
		Spacing:       radius * 0.25,
		NoiseStrength: matrix.Clamp(w.readBrushFloat(w.textureNoiseInput, 0), 0, 1),
		NoiseScale:    max(radius, matrix.Float(0.001)),
		Jitter:        matrix.Clamp(w.readBrushFloat(w.textureJitterInput, 0), 0, radius),
		Stamp:         w.readTextureStamp(),
		Constraints: terrain.TexturePaintConstraints{
			UseSlope:  w.textureSlopeConstraintsEnabled(),
			SlopeMin:  w.readBrushFloat(w.textureSlopeMinInput, 0),
			SlopeMax:  w.readBrushFloat(w.textureSlopeMaxInput, 90),
			UseHeight: w.textureHeightConstraintsEnabled(),
			HeightMin: w.readBrushFloat(w.textureHeightMinInput, -100000),
			HeightMax: w.readBrushFloat(w.textureHeightMaxInput, 100000),
		},
	}
}

func (w *TerrainWorkspace) readTextureStamp() *terrain.TextureBrushStamp {
	if w.textureStampSelect == nil {
		return nil
	}
	switch w.textureStampSelect.UI.ToSelect().Value() {
	case "soft":
		return radialTextureStamp(16, true)
	case "hard":
		return radialTextureStamp(16, false)
	default:
		return nil
	}
}

func radialTextureStamp(resolution int, soft bool) *terrain.TextureBrushStamp {
	stamp := &terrain.TextureBrushStamp{
		Resolution: resolution,
		Alpha:      make([]matrix.Float, resolution*resolution),
	}
	center := matrix.Float(resolution-1) * 0.5
	for z := 0; z < resolution; z++ {
		for x := 0; x < resolution; x++ {
			dx := (matrix.Float(x) - center) / center
			dz := (matrix.Float(z) - center) / center
			d := matrix.Sqrt(dx*dx + dz*dz)
			alpha := matrix.Float(0)
			if d <= 1 {
				alpha = 1
				if soft {
					alpha = 1 - d
				}
			}
			stamp.Alpha[x+z*resolution] = alpha
		}
	}
	return stamp
}

func (w *TerrainWorkspace) effectiveBrushMode() terrain.BrushMode {
	kb := &w.Host.Window.Keyboard
	return effectiveTerrainBrushMode(w.mode, kb.HasShift(), kb.HasCtrlOrMeta())
}

func (w *TerrainWorkspace) pointerOverUI() bool {
	return len(w.UiMan.Hovered()) > 0 || w.UiMan.Group.HasRequests()
}

func (w *TerrainWorkspace) adjustBrushRadius(direction int) {
	if w.toolMode == TerrainToolTexturePaint {
		w.adjustBrushInput(w.textureRadiusInput, 2, direction,
			terrainBrushMinRadius, terrainBrushMaxRadius, "Texture brush radius")
		return
	}
	w.adjustBrushInput(w.radiusInput, 2, direction,
		terrainBrushMinRadius, terrainBrushMaxRadius, "Brush radius")
}

func (w *TerrainWorkspace) adjustBrushStrength(direction int) {
	if w.toolMode == TerrainToolTexturePaint {
		w.adjustBrushInput(w.textureOpacityInput, 1, direction,
			0, 1, "Texture opacity")
		return
	}
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
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	})
}

func (w *TerrainWorkspace) showBrushPreview(hit terrain.TerrainRayHit) {
	radius := w.readActiveBrushRadius()
	color := matrix.NewColor(0.2, 0.75, 1.0, 1.0)
	if w.toolMode == TerrainToolTexturePaint {
		color = matrix.NewColor(0.9, 0.55, 0.18, 1.0)
	}
	ringWidth := matrix.Max(radius*matrix.Float(0.035), matrix.Float(0.05))
	w.active.SetBrushPreview(hit.Point.XZ(), radius, ringWidth, color)
	if w.brushRingData == nil {
		return
	}
	w.brushRingTransform.SetPosition(hit.Point.Add(hit.Normal.Scale(0.025)))
	w.brushRingTransform.SetScale(matrix.NewVec3(radius, 1, radius))
	w.brushRingData.Activate()
}

func (w *TerrainWorkspace) readActiveBrushRadius() matrix.Float {
	if w.toolMode == TerrainToolTexturePaint {
		return w.readBrushFloat(w.textureRadiusInput, 2)
	}
	return w.readBrushFloat(w.radiusInput, 2)
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
	w.stroke = nil
	w.textureStrokeCapture = nil
	if w.toolMode == TerrainToolTexturePaint {
		w.textureStrokeCapture = newTerrainTextureStrokeCapture(w.active)
		return
	}
	if w.toolMode != TerrainToolHeightSculpt {
		return
	}
	w.stroke = newTerrainStrokeCapture(w.active)
}

func (w *TerrainWorkspace) finishStroke() {
	if w.stroke != nil {
		if h := w.stroke.history(); h != nil {
			w.ed.History().Add(h)
		}
	}
	if w.textureStrokeCapture != nil {
		if h := w.textureStrokeCapture.history(); h != nil {
			w.ed.History().Add(h)
		}
	}
	w.painting = false
	w.hasLastLocal = false
	w.stroke = nil
	w.textureStrokeCapture = nil
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

func (w *TerrainWorkspace) captureTextureLine(from, to matrix.Vec2, stroke terrain.TexturePaintStroke) {
	if w.textureStrokeCapture == nil {
		return
	}
	w.active.VisitTexturePaintLineStamps(from, to, stroke, func(stamp terrain.TexturePaintStroke) bool {
		w.captureTextureStroke(stamp)
		return true
	})
}

func (w *TerrainWorkspace) captureTextureStroke(stroke terrain.TexturePaintStroke) {
	if w.textureStrokeCapture == nil || stroke.Mode == terrain.TextureBrushSample {
		return
	}
	w.textureStrokeCapture.captureRegion(w.active.TextureStrokeRegion(stroke))
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

func (w *TerrainWorkspace) readTextureFalloff() terrain.BrushFalloff {
	if w.textureFalloffSelect == nil {
		return terrain.FalloffSmooth
	}
	switch w.textureFalloffSelect.UI.ToSelect().Value() {
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

func (w *TerrainWorkspace) textureSlopeConstraintsEnabled() bool {
	minValue := w.readBrushFloat(w.textureSlopeMinInput, 0)
	maxValue := w.readBrushFloat(w.textureSlopeMaxInput, 90)
	return minValue > 0 || maxValue < 90
}

func (w *TerrainWorkspace) textureHeightConstraintsEnabled() bool {
	if w.active == nil || w.active.HeightField == nil {
		return false
	}
	minValue := w.readBrushFloat(w.textureHeightMinInput, w.active.HeightField.MinHeight)
	maxValue := w.readBrushFloat(w.textureHeightMaxInput, w.active.HeightField.MaxHeight)
	return minValue > w.active.HeightField.MinHeight || maxValue < w.active.HeightField.MaxHeight
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

func (w *TerrainWorkspace) refreshToolPanels() {
	heightActive := w.toolMode == TerrainToolHeightSculpt
	textureActive := w.toolMode == TerrainToolTexturePaint
	if w.heightToolRow != nil {
		if heightActive {
			w.heightToolRow.UI.Show()
		} else {
			w.heightToolRow.UI.Hide()
		}
	}
	if w.heightBrush != nil {
		if heightActive {
			w.heightBrush.UI.Show()
		} else {
			w.heightBrush.UI.Hide()
		}
	}
	if w.textureRow != nil {
		if textureActive {
			w.textureRow.UI.Show()
		} else {
			w.textureRow.UI.Hide()
		}
	}
	if w.textureSwatchTemplate != nil && !textureActive {
		w.textureSwatchTemplate.UI.Hide()
	}
	for i := range w.textureSwatches {
		if textureActive {
			w.textureSwatches[i].UI.Show()
		} else {
			w.textureSwatches[i].UI.Hide()
		}
	}
	w.refreshModeButtonClasses()
}

func (w *TerrainWorkspace) refreshModeButtonClasses() {
	for i := range w.modeBtns {
		classes := []string{"materialIcon"}
		if i == int(w.toolMode) {
			classes = append(classes, "active")
		}
		w.Doc.SetElementClasses(w.modeBtns[i], classes...)
	}
}

func (w *TerrainWorkspace) refreshLayerSelector() {
	if w.textureLayerSelect == nil {
		return
	}
	selectUI := w.textureLayerSelect.UI.ToSelect()
	selectUI.ClearOptions()
	if w.active == nil || w.active.LayerCount() == 0 {
		selectUI.AddOption("No layers", "-1")
		selectUI.PickOptionWithoutEvent(0)
		return
	}
	if w.textureLayer < 0 {
		w.textureLayer = 0
	}
	if w.textureLayer >= w.active.LayerCount() {
		w.textureLayer = w.active.LayerCount() - 1
	}
	for i := 0; i < w.active.LayerCount(); i++ {
		name := w.layerDisplayName(i)
		selectUI.AddOption(name, strconv.Itoa(i))
	}
	selectUI.PickOptionWithoutEvent(w.textureLayer)
}

func (w *TerrainWorkspace) refreshLayerPalette() {
	if w.textureSwatchTemplate == nil {
		return
	}
	defer w.refreshToolPanels()
	for i := range w.textureSwatches {
		w.Doc.RemoveElementWithoutApplyStyles(w.textureSwatches[i])
	}
	w.textureSwatches = w.textureSwatches[:0]
	if w.active == nil || w.active.LayerCount() == 0 {
		w.Doc.ApplyStyles()
		return
	}
	w.textureSwatches = w.Doc.DuplicateElementRepeatWithoutApplyStyles(
		w.textureSwatchTemplate, w.active.LayerCount())
	for i := range w.textureSwatches {
		swatch := w.textureSwatches[i]
		w.Doc.SetElementIdWithoutApplyStyles(swatch, "")
		swatch.UI.Show()
		swatch.SetAttribute("data-layer", strconv.Itoa(i))
		classes := []string{"layerSwatch"}
		if i == w.textureLayer {
			classes = append(classes, "active")
		}
		w.Doc.SetElementClassesWithoutApply(swatch, classes...)
		if w.toolMode != TerrainToolTexturePaint {
			swatch.UI.Hide()
		}
		if len(swatch.Children) > 1 && swatch.Children[1].InnerLabel() != nil {
			swatch.Children[1].InnerLabel().SetText(w.layerDisplayName(i))
		}
		w.loadLayerSwatchTexture(swatch, i)
	}
	w.Doc.ApplyStyles()
}

func (w *TerrainWorkspace) refreshLayerPaletteSelection() {
	if w.Doc == nil || len(w.textureSwatches) == 0 {
		w.refreshLayerPalette()
		return
	}
	for i := range w.textureSwatches {
		classes := []string{"layerSwatch"}
		if i == w.textureLayer {
			classes = append(classes, "active")
		}
		w.Doc.SetElementClassesWithoutApply(w.textureSwatches[i], classes...)
	}
	w.Doc.ApplyStyles()
}

func (w *TerrainWorkspace) refreshTextureLayerFields() {
	if w.active == nil || w.active.LayerSet == nil || w.textureLayer < 0 ||
		w.textureLayer >= w.active.LayerCount() {
		return
	}
	layer := w.active.LayerSet.Layers[w.textureLayer]
	if w.textureLayerNameInput != nil {
		w.textureLayerNameInput.UI.ToInput().SetTextWithoutEvent(layer.Name)
	}
	if w.textureFilterSelect != nil {
		w.textureFilterSelect.UI.ToSelect().PickOptionWithoutEvent(textureFilterOptionIndex(layer.Filter))
	}
	if w.textureTilingXInput != nil {
		w.textureTilingXInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(layer.Tiling.X()))
	}
	if w.textureTilingYInput != nil {
		w.textureTilingYInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(layer.Tiling.Y()))
	}
	if w.textureWorldSizeXInput != nil {
		w.textureWorldSizeXInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(layer.TextureWorldSize.X()))
	}
	if w.textureWorldSizeYInput != nil {
		w.textureWorldSizeYInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(layer.TextureWorldSize.Y()))
	}
	if w.textureTintRInput != nil {
		w.textureTintRInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(layer.Tint.R()))
	}
	if w.textureTintGInput != nil {
		w.textureTintGInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(layer.Tint.G()))
	}
	if w.textureTintBInput != nil {
		w.textureTintBInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(layer.Tint.B()))
	}
	if w.textureTintAInput != nil {
		w.textureTintAInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(layer.Tint.A()))
	}
	if w.active.HeightField != nil {
		if w.textureHeightMinInput != nil {
			w.textureHeightMinInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(w.active.HeightField.MinHeight))
		}
		if w.textureHeightMaxInput != nil {
			w.textureHeightMaxInput.UI.ToInput().SetTextWithoutEvent(fmtFloat(w.active.HeightField.MaxHeight))
		}
	}
}

func (w *TerrainWorkspace) applyTextureLayerSettings() {
	if w.active == nil || w.active.LayerSet == nil {
		return
	}
	layer := w.readTextureLayer()
	if layer < 0 || layer >= w.active.LayerCount() {
		return
	}
	next := w.active.LayerSet.Layers[layer]
	if w.textureLayerNameInput != nil {
		next.Name = strings.TrimSpace(w.textureLayerNameInput.UI.ToInput().Text())
	}
	next.Filter = w.readTextureFilter()
	next.Tiling = matrix.NewVec2(
		w.readBrushFloat(w.textureTilingXInput, 1),
		w.readBrushFloat(w.textureTilingYInput, 1),
	)
	next.TextureWorldSize = matrix.NewVec2(
		w.readBrushFloat(w.textureWorldSizeXInput, 0),
		w.readBrushFloat(w.textureWorldSizeYInput, 0),
	)
	if next.TextureWorldSize.X() > matrix.Tiny && next.TextureWorldSize.Y() > matrix.Tiny {
		next.Tiling = matrix.NewVec2(
			w.active.Config.WorldSize.X()/next.TextureWorldSize.X(),
			w.active.Config.WorldSize.Y()/next.TextureWorldSize.Y(),
		)
	}
	next.Tint = matrix.NewColor(
		matrix.Clamp(w.readBrushFloat(w.textureTintRInput, 1), 0, 1),
		matrix.Clamp(w.readBrushFloat(w.textureTintGInput, 1), 0, 1),
		matrix.Clamp(w.readBrushFloat(w.textureTintBInput, 1), 0, 1),
		matrix.Clamp(w.readBrushFloat(w.textureTintAInput, 1), 0, 1),
	)
	if w.active.SetLayer(layer, next) {
		w.refreshLayerSelector()
		w.refreshLayerPalette()
		w.setLayerTextureStatus(next.TextureContentID, "Updated texture layer")
	}
}

func (w *TerrainWorkspace) readTextureLayer() int {
	if w.textureLayerSelect == nil {
		return w.textureLayer
	}
	v, err := strconv.Atoi(w.textureLayerSelect.UI.ToSelect().Value())
	if err != nil {
		return w.textureLayer
	}
	return v
}

func (w *TerrainWorkspace) readTextureFilter() rendering.TextureFilter {
	if w.textureFilterSelect == nil {
		return rendering.TextureFilterLinear
	}
	switch w.textureFilterSelect.UI.ToSelect().Value() {
	case "nearest":
		return rendering.TextureFilterNearest
	case "linear":
		fallthrough
	default:
		return rendering.TextureFilterLinear
	}
}

func textureFilterOption(filter rendering.TextureFilter) string {
	if filter == rendering.TextureFilterNearest {
		return "nearest"
	}
	return "linear"
}

func textureFilterOptionIndex(filter rendering.TextureFilter) int {
	if textureFilterOption(filter) == "nearest" {
		return 1
	}
	return 0
}

func textureLayerFromElement(e *document.Element) (int, bool) {
	for e != nil {
		if value := e.Attribute("data-layer"); value != "" {
			layer, err := strconv.Atoi(value)
			return layer, err == nil
		}
		e = e.Parent.Value()
	}
	return 0, false
}

func (w *TerrainWorkspace) layerDisplayName(layer int) string {
	if w.active == nil || w.active.LayerSet == nil ||
		layer < 0 || layer >= len(w.active.LayerSet.Layers) {
		return "Layer"
	}
	data := w.active.LayerSet.Layers[layer]
	name := strings.TrimSpace(data.Name)
	if name == "" {
		name = w.textureNameForID(data.TextureContentID)
	}
	if name == "" {
		name = "Layer " + strconv.Itoa(layer+1)
	}
	if data.Locked {
		name += " [L]"
	}
	if data.Hidden {
		name += " [H]"
	}
	if data.Solo {
		name += " [S]"
	}
	return strconv.Itoa(layer+1) + " " + name
}

func (w *TerrainWorkspace) textureNameForID(id string) string {
	if id == "" {
		return ""
	}
	if id == assets.TextureSquare {
		return assets.TextureSquare
	}
	if w.ed == nil || w.ed.Cache() == nil {
		return id
	}
	cc, err := w.ed.Cache().Read(id)
	if err != nil || cc.Config.Name == "" {
		return id
	}
	return cc.Config.Name
}

func (w *TerrainWorkspace) loadLayerSwatchTexture(swatch *document.Element, layer int) {
	if swatch == nil || len(swatch.Children) == 0 || w.active == nil ||
		w.active.LayerSet == nil || layer < 0 || layer >= len(w.active.LayerSet.Layers) {
		return
	}
	thumb := swatch.Children[0].UI.ToPanel()
	data := w.active.LayerSet.Layers[layer]
	tex, err := w.Host.TextureCache().Texture(data.TextureContentID, data.Filter)
	if err != nil {
		tex, err = w.Host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	}
	if err == nil {
		thumb.SetBackground(tex)
	}
	thumb.SetColor(data.Tint)
}

func (w *TerrainWorkspace) validateLayerTextures() bool {
	if w.active == nil || w.active.LayerSet == nil {
		return true
	}
	missing := terrain.MissingTerrainLayerTextures(w.active.LayerSet.Layers, w.textureExists)
	if len(missing) > 0 {
		w.setStatus(terrainLayerTextureDiagnosticStatus(missing))
		return false
	}
	return true
}

func (w *TerrainWorkspace) setLayerTextureStatus(id, okStatus string) {
	if !w.textureExists(id) {
		w.setStatus(terrainLayerTextureDiagnosticStatus([]terrain.TerrainLayerTextureDiagnostic{{
			Layer:            w.readTextureLayer(),
			Name:             w.layerDisplayName(w.readTextureLayer()),
			TextureContentID: id,
		}}))
		return
	}
	w.setStatus(okStatus)
}

func (w *TerrainWorkspace) textureExists(id string) bool {
	if id == "" {
		return false
	}
	if id == assets.TextureSquare {
		return true
	}
	if w.ed == nil || w.ed.Cache() == nil {
		return true
	}
	cc, err := w.ed.Cache().Read(id)
	if err != nil {
		return false
	}
	return cc.Config.Type == (content_database.Texture{}).TypeName()
}

func (w *TerrainWorkspace) refreshToolReadout() {
	if w.toolReadout == nil {
		return
	}
	if w.toolMode == TerrainToolTexturePaint {
		w.toolReadout.InnerLabel().SetText(textureToolName(w.textureMode) + " / L " +
			strconv.Itoa(w.readTextureLayer()+1) + " / R " +
			fmtFloat(w.readBrushFloat(w.textureRadiusInput, 2)) + " / O " +
			fmtFloat(w.readBrushFloat(w.textureOpacityInput, 1)))
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

func textureToolName(mode terrain.TextureBrushMode) string {
	switch mode {
	case terrain.TextureBrushErase:
		return "Erase"
	case terrain.TextureBrushSmoothWeights:
		return "Blend"
	case terrain.TextureBrushFill:
		return "Fill"
	case terrain.TextureBrushSample:
		return "Pick"
	case terrain.TextureBrushPaint:
		fallthrough
	default:
		return "Paint"
	}
}

func terrainLayerTextureDiagnosticStatus(missing []terrain.TerrainLayerTextureDiagnostic) string {
	if len(missing) == 0 {
		return ""
	}
	if len(missing) == 1 {
		d := missing[0]
		return "Missing texture L" + strconv.Itoa(d.Layer+1) + " " + d.TextureContentID +
			"; using " + assets.TextureSquare
	}
	return strconv.Itoa(len(missing)) + " missing terrain layer textures; using " + assets.TextureSquare + " fallbacks"
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

type terrainTextureStrokeCapture struct {
	target *terrain.Terrain
	before map[int][]matrix.Float
	region terrain.DirtyRegion
}

func newTerrainTextureStrokeCapture(target *terrain.Terrain) *terrainTextureStrokeCapture {
	return &terrainTextureStrokeCapture{
		target: target,
		before: make(map[int][]matrix.Float),
	}
}

func (c *terrainTextureStrokeCapture) captureRegion(region terrain.DirtyRegion) {
	if c == nil || c.target == nil || !region.Valid ||
		c.target.LayerSet == nil || c.target.LayerSet.WeightMap == nil {
		return
	}
	weights := c.target.LayerSet.WeightMap
	region = region.Expand(0, weights.Resolution)
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			idx := x + z*weights.Resolution
			if _, ok := c.before[idx]; ok {
				continue
			}
			cell := make([]matrix.Float, weights.Layers)
			for layer := 0; layer < weights.Layers; layer++ {
				cell[layer] = weights.WeightAt(layer, x, z)
			}
			c.before[idx] = cell
		}
	}
}

func (c *terrainTextureStrokeCapture) markDirty(region terrain.DirtyRegion) {
	if c == nil || !region.Valid {
		return
	}
	c.region = mergeTerrainRegions(c.region, region)
}

func (c *terrainTextureStrokeCapture) history() *terrainTextureStrokeHistory {
	if c == nil || c.target == nil || !c.region.Valid ||
		c.target.LayerSet == nil || c.target.LayerSet.WeightMap == nil {
		return nil
	}
	weights := c.target.LayerSet.WeightMap
	region := c.region.Expand(0, weights.Resolution)
	if !region.Valid {
		return nil
	}
	width := region.MaxX - region.MinX + 1
	height := region.MaxZ - region.MinZ + 1
	before := make([]matrix.Float, width*height*weights.Layers)
	after := make([]matrix.Float, width*height*weights.Layers)
	different := false
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			cell := (x - region.MinX) + (z-region.MinZ)*width
			mapIdx := x + z*weights.Resolution
			beforeCell, ok := c.before[mapIdx]
			for layer := 0; layer < weights.Layers; layer++ {
				beforeWeight := weights.WeightAt(layer, x, z)
				if ok && layer < len(beforeCell) {
					beforeWeight = beforeCell[layer]
				}
				afterWeight := weights.WeightAt(layer, x, z)
				outIdx := cell*weights.Layers + layer
				before[outIdx] = beforeWeight
				after[outIdx] = afterWeight
				different = different || beforeWeight != afterWeight
			}
		}
	}
	if !different {
		return nil
	}
	return &terrainTextureStrokeHistory{
		target: c.target,
		region: region,
		before: before,
		after:  after,
	}
}

type terrainTextureStrokeHistory struct {
	target *terrain.Terrain
	region terrain.DirtyRegion
	before []matrix.Float
	after  []matrix.Float
}

func (h *terrainTextureStrokeHistory) Redo()   { h.apply(h.after) }
func (h *terrainTextureStrokeHistory) Undo()   { h.apply(h.before) }
func (h *terrainTextureStrokeHistory) Delete() {}
func (h *terrainTextureStrokeHistory) Exit()   {}

func (h *terrainTextureStrokeHistory) apply(weights []matrix.Float) {
	if h == nil || h.target == nil || !h.region.Valid {
		return
	}
	h.target.ApplyTextureWeightRegion(h.region, weights)
}

type terrainLayerSetHistory struct {
	workspace   *TerrainWorkspace
	target      *terrain.Terrain
	before      terrain.TerrainLayerSetState
	after       terrain.TerrainLayerSetState
	beforeLayer int
	afterLayer  int
}

func (h *terrainLayerSetHistory) Redo()   { h.apply(h.after, h.afterLayer) }
func (h *terrainLayerSetHistory) Undo()   { h.apply(h.before, h.beforeLayer) }
func (h *terrainLayerSetHistory) Delete() {}
func (h *terrainLayerSetHistory) Exit()   {}

func (h *terrainLayerSetHistory) changed() bool {
	if h == nil {
		return false
	}
	return !reflect.DeepEqual(h.before, h.after)
}

func (h *terrainLayerSetHistory) apply(state terrain.TerrainLayerSetState, selectedLayer int) {
	if h == nil || h.target == nil || !h.target.ApplyLayerSetState(state) {
		return
	}
	if h.workspace == nil || h.workspace.active != h.target {
		return
	}
	h.workspace.textureLayer = min(max(selectedLayer, 0), h.target.LayerCount()-1)
	h.workspace.refreshLayerSelector()
	h.workspace.refreshTextureLayerFields()
	h.workspace.refreshLayerPalette()
	h.workspace.refreshToolReadout()
	h.workspace.validateLayerTextures()
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

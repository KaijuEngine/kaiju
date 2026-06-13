/******************************************************************************/
/* editor.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"log/slog"
	"time"

	"kaijuengine.com/build"
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_embedded_content"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_logging"
	"kaijuengine.com/editor/editor_overlay/context_menu"
	"kaijuengine.com/editor/editor_plugin"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/editor/global_interface/menu_bar"
	"kaijuengine.com/editor/global_interface/status_bar"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_previews"
	"kaijuengine.com/editor/webapi"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	platformPower "kaijuengine.com/platform/power"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"

	// Built-in workspace packages register themselves with
	// editor_workspace_registry from their init(). Blank-imported here for
	// the side effect; files in this package that need the concrete types
	// (e.g. editor_menu_bar_handler.go) re-import them by name.
	_ "kaijuengine.com/editor/editor_workspace/content_workspace"
	_ "kaijuengine.com/editor/editor_workspace/render_graph_workspace"
	_ "kaijuengine.com/editor/editor_workspace/schema_workspace"
	_ "kaijuengine.com/editor/editor_workspace/settings_workspace"
	_ "kaijuengine.com/editor/editor_workspace/stage_workspace"
	_ "kaijuengine.com/editor/editor_workspace/terrain_workspace"
	_ "kaijuengine.com/editor/editor_workspace/ui_workspace"
	_ "kaijuengine.com/editor/editor_workspace/vfx_workspace"
	_ "kaijuengine.com/editor/editor_workspace/vulkan_workspace"
)

// Editor is the entry point structure for the entire editor. It acts as the
// delegate to the various systems and holds the primary members that make up
// the bulk of the editor identity.
//
// The design goal of the editor is different than that of the [engine.Host], as
// it is not intended to be passed around for access to the system. Instead it
// will supply interface functions that are needed to the systems that it holds
// internally.
type Editor struct {
	host                  *engine.Host
	settings              editor_settings.Settings
	project               project.Project
	workspaceState        WorkspaceState
	activeWorkspaces      map[string]editor_workspace.Workspace
	workspaceOrder        []string
	initializedWorkspaces map[string]struct{}
	globalInterfaces      globalUI
	currentWorkspace      editor_workspace.Workspace
	logging               editor_logging.Logging
	history               memento.History
	events                editor_events.EditorEvents
	stageView             editor_stage_view.StageView
	plugins               []editor_plugin.EditorPlugin
	fileDropRouter        FileDropRouter
	window                struct {
		activateId     events.Id
		deactivateId   events.Id
		lastActiveTime time.Time
	}
	contentPreviewer content_previews.ContentPreviewer
	updateId         engine.UpdateId
	webAPIServer     *webapi.Server[*Editor]
	actions          *editor_action.Service
	blurred          int
	actionPaletteKey struct {
		pending bool
		moved   bool
	}
	power powerState
	// sessionDisabledPlugins holds module paths of plugins the user chose
	// to skip via the startup-validation modal's "Continue" button (only
	// MISSING plugins are recorded here — stale plugins are not tracked,
	// see editor_plugin_validation.go for the rationale). Process-local
	// only; never persisted to plugin.json. Cleared at next process start.
	// Read by MissingCompiledPlugins to suppress repeat modals within the
	// same process. Touched only from the main UI goroutine — no lock
	// needed.
	sessionDisabledPlugins map[string]struct{}
}

type globalUI struct {
	menuBar   menu_bar.MenuBar
	statusBar status_bar.StatusBar
}

const (
	editorPowerPollInterval         = 5.0
	launchMaxTextureCreatesPerFrame = 8
	launchMaxTextureBytesPerFrame   = 32 * 1024 * 1024
)

type powerStatusQuery func() (platformPower.Status, error)

type powerState struct {
	query       powerStatusQuery
	lastStatus  platformPower.Status
	initialized bool
	pollElapsed float64
}

func (ed *Editor) Host() *engine.Host { return ed.host }

func (ed *Editor) ContentPreviewer() *content_previews.ContentPreviewer {
	return &ed.contentPreviewer
}

// FocusInterface is responsible for enabling the input on the various
// interfaces that are currently presented to the developer. This primarily
// includes the menu bar, status bar, and whichever workspace is active.
func (ed *Editor) FocusInterface() {
	defer tracing.NewRegion("Editor.FocusInterface").End()
	ed.blurred = max(ed.blurred-1, 0)
	if ed.blurred > 0 {
		return
	}
	ed.globalInterfaces.menuBar.Focus()
	ed.globalInterfaces.statusBar.Focus()
	if ed.currentWorkspace != nil {
		ed.currentWorkspace.Focus()
	}
}

// FocusInterface is responsible for disabling the input on the various
// interfaces that are currently presented to the developer. This primarily
// includes the menu bar, status bar, and whichever workspace is active.
func (ed *Editor) BlurInterface() {
	defer tracing.NewRegion("Editor.BlurInterface").End()
	ed.blurred++
	if ed.blurred > 1 {
		return
	}
	ed.globalInterfaces.menuBar.Blur()
	ed.globalInterfaces.statusBar.Blur()
	if ed.currentWorkspace != nil {
		ed.currentWorkspace.Blur()
	}
}

func (ed *Editor) IsInputFocused() bool {
	if ed.globalInterfaces.menuBar.IsFocusedOnInput() {
		return true
	} else if ed.globalInterfaces.statusBar.IsFocusedOnInput() {
		return true
	}
	if ed.currentWorkspace == nil {
		return false
	}
	return ed.currentWorkspace.IsFocusedOnInput()
}

func (ed *Editor) earlyLoadUI() {
	defer tracing.NewRegion("Editor.earlyLoadUI").End()
	ed.globalInterfaces.menuBar.Initialize(ed.host, ed)
	ed.globalInterfaces.statusBar.Initialize(ed.host, &ed.logging, ed)
	ed.host.SetSwapChainClearColor(matrix.ColorCornflowerBlue())
}

func (ed *Editor) UpdateSettings() {
	ed.setFrameRateLimitForPowerStatus(ed.queryAndCachePowerStatus())
	if matrix.Approx(ed.settings.UIScrollSpeed, 0) {
		ed.settings.UIScrollSpeed = 1
	}
	ed.settings.NormalizeWebAPI()
	ui.UIScrollSpeed = ed.settings.UIScrollSpeed
	if err := ed.settings.Save(); err != nil {
		slog.Error("failed to save the editor settings", "error", err)
		return
	}
	ed.updateWebAPI()
}

func (ed *Editor) queryAndCachePowerStatus() platformPower.Status {
	status := ed.queryPowerStatus()
	ed.power.lastStatus = status
	ed.power.initialized = true
	ed.power.pollElapsed = 0
	return status
}

func (ed *Editor) queryPowerStatus() platformPower.Status {
	query := ed.power.query
	if query == nil {
		query = platformPower.Query
	}
	status, err := query()
	if err != nil {
		slog.Debug("failed to query power status", "error", err)
		return platformPower.Status{Source: platformPower.SourceUnknown, BatteryPercent: -1}
	}
	return status
}

func (ed *Editor) setFrameRateLimitForPowerStatus(status platformPower.Status) {
	if ed.host == nil {
		return
	}
	ed.host.SetFrameRateLimit(int64(ed.effectiveRefreshRate(status)))
}

func (ed *Editor) effectiveRefreshRate(status platformPower.Status) int32 {
	refreshRate := ed.settings.RefreshRate
	if ed.settings.UseBatteryRefreshRate && status.OnBattery() {
		refreshRate = ed.settings.BatteryRefreshRate
	}
	return klib.Clamp(refreshRate, 0, 320)
}

func (ed *Editor) postProjectLoad() {
	defer tracing.NewRegion("Editor.lateLoadUI").End()
	ed.settings.AddRecentProject(ed.project.FileSystem().FullPath(""))
	slog.Info("compiling the project to get things ready")
	{
		// Read the project source synchronosly for now, if not, any stage loading
		// before this is complete will have issues.
		ed.project.ReadSourceCode()
	}
	editorContent := ed.host.AssetDatabase().(*editor_embedded_content.EditorContent)
	editorContent.Pfs = ed.project.FileSystem()
	editorContent.SetProjectContentIndex(ed.project.CacheDatabase().List())
	ed.events.OnContentAdded.Add(func(ids []string) {
		editorContent.IndexProjectContentIDs(ed.project.CacheDatabase(), ids)
	})
	ed.events.OnContentRemoved.Add(editorContent.RemoveProjectContentIDs)
	ed.host.TextureCache().SetUploadBudget(rendering.TextureUploadBudget{
		MaxCreatesPerFrame: launchMaxTextureCreatesPerFrame,
		MaxBytesPerFrame:   launchMaxTextureBytesPerFrame,
	})
	ed.setupWindowActivity()
	ed.activeWorkspaces = map[string]editor_workspace.Workspace{}
	ed.initializedWorkspaces = map[string]struct{}{}
	ed.reconcileWorkspaces()
	ed.initializeWorkspaces()
	ed.rebuildMenuBarTabs()
	ed.connectFileDropRouter()
	if id := ed.firstSelectableWorkspaceID(); id != WorkspaceStateNone {
		ed.setWorkspaceState(id)
	}
	// goroutine
	go ed.project.CompileDebug()
	if build.Debug && ed.initAutoTest() {
		ed.updateId = ed.host.Updater.AddUpdate(ed.runAutoTest)
	} else {
		ed.updateId = ed.host.Updater.AddUpdate(ed.update)
	}
	for k, v := range editorPluginRegistry {
		if err := v.Launch(ed); err != nil {
			slog.Error("failed to launch plugin", "key", k, "error", err)
			continue
		}
		ed.plugins = append(ed.plugins, v)
	}
	// A plugin's Launch may have called RegisterWorkspace late. Pick up any
	// new entries, initialize them, and refresh the menu bar.
	if ed.reconcileWorkspaces() {
		ed.initializeWorkspaces()
		ed.rebuildMenuBarTabs()
	}
	// Pre-warm the, quite large, material icons PNG file
	ed.host.TextureCache().Texture("MaterialIcons-Regular.png", rendering.TextureFilterLinear)
}

// defaultWorkspaceOrder is the canonical first-time ordering of the built-in
// workspaces. Used by reconcileWorkspaces when a workspace has no entry in
// persisted settings yet (first run, or a workspace was just registered).
// Plugin workspaces and any built-ins not listed here are appended at the
// end in registration order. The user's drag-reorder choices override this.
var defaultWorkspaceOrder = []string{
	"stage",
	"content",
	"renderGraph",
	"terrain",
	"vfx",
	"ui",
	"vulkan",
	"schema",
	"settings",
}

func defaultWorkspaceOrderIndex(id string) int {
	for i := range defaultWorkspaceOrder {
		if defaultWorkspaceOrder[i] == id {
			return i
		}
	}
	return -1
}

func insertDefaultWorkspaceConfig(workspaces []editor_settings.WorkspaceConfig, cfg editor_settings.WorkspaceConfig) []editor_settings.WorkspaceConfig {
	cfgOrder := defaultWorkspaceOrderIndex(cfg.ID)
	if cfgOrder < 0 {
		return append(workspaces, cfg)
	}
	insertAt := len(workspaces)
	for i := range workspaces {
		workspaceOrder := defaultWorkspaceOrderIndex(workspaces[i].ID)
		if workspaceOrder > cfgOrder {
			insertAt = i
			break
		}
	}
	workspaces = append(workspaces, editor_settings.WorkspaceConfig{})
	copy(workspaces[insertAt+1:], workspaces[insertAt:])
	workspaces[insertAt] = cfg
	return workspaces
}

// reconcileWorkspaces walks the global registry and the persisted
// settings.Workspaces slice and produces a single source of truth for which
// workspaces should be active and in what order. Returns true if the active
// set changed (a new workspace appeared, e.g. a plugin registered late).
//
// The reconciliation rules are:
//   - any workspace in the registry that has no settings entry yet is
//     inserted into settings.Workspaces with Enabled=true. Insertion order
//     follows defaultWorkspaceOrder when the new id is in that list;
//     otherwise the entry is appended last (in registry registration order).
//   - any settings entry whose ID is no longer in the registry is dropped
//     (a plugin was uninstalled).
//   - any required workspace (IsRequired() == true) is force-enabled
//     regardless of stored state, so the user cannot brick the editor.
func (ed *Editor) reconcileWorkspaces() bool {
	defer tracing.NewRegion("Editor.reconcileWorkspaces").End()
	registered := map[string]editor_workspace.Workspace{}
	for _, w := range editor_workspace_registry.All() {
		registered[w.ID()] = w
	}
	// Drop stale entries.
	pruned := ed.settings.Workspaces[:0]
	for _, cfg := range ed.settings.Workspaces {
		if _, ok := registered[cfg.ID]; ok {
			pruned = append(pruned, cfg)
		}
	}
	ed.settings.Workspaces = pruned
	// Compute the set of registered workspaces missing from settings.
	known := map[string]bool{}
	for _, cfg := range ed.settings.Workspaces {
		known[cfg.ID] = true
	}
	missing := map[string]bool{}
	for _, id := range editor_workspace_registry.IDs() {
		if !known[id] {
			missing[id] = true
		}
	}
	// Insert missing entries: first walk defaultWorkspaceOrder so we honor
	// the canonical first-time ordering, then append everything else in
	// registration order so plugin workspaces show up at the end.
	for _, id := range defaultWorkspaceOrder {
		if !missing[id] {
			continue
		}
		ed.settings.Workspaces = insertDefaultWorkspaceConfig(ed.settings.Workspaces, editor_settings.WorkspaceConfig{
			ID:      id,
			Enabled: true,
		})
		delete(missing, id)
	}
	for _, id := range editor_workspace_registry.IDs() {
		if !missing[id] {
			continue
		}
		ed.settings.Workspaces = append(ed.settings.Workspaces, editor_settings.WorkspaceConfig{
			ID:      id,
			Enabled: true,
		})
		delete(missing, id)
	}
	// Force required workspaces enabled.
	for i := range ed.settings.Workspaces {
		w := registered[ed.settings.Workspaces[i].ID]
		if w != nil && w.IsRequired() {
			ed.settings.Workspaces[i].Enabled = true
		}
	}
	// Recompute active set + order.
	changed := false
	newOrder := make([]string, 0, len(ed.settings.Workspaces))
	for _, cfg := range ed.settings.Workspaces {
		if !cfg.Enabled {
			continue
		}
		newOrder = append(newOrder, cfg.ID)
		if _, already := ed.activeWorkspaces[cfg.ID]; !already {
			ed.activeWorkspaces[cfg.ID] = registered[cfg.ID]
			changed = true
		}
	}
	// Drop active entries that were disabled.
	for id := range ed.activeWorkspaces {
		stillActive := false
		for _, keep := range newOrder {
			if keep == id {
				stillActive = true
				break
			}
		}
		if !stillActive {
			ed.activeWorkspaces[id].Shutdown()
			delete(ed.activeWorkspaces, id)
			delete(ed.initializedWorkspaces, id)
			changed = true
		}
	}
	if !sliceEqual(ed.workspaceOrder, newOrder) {
		ed.workspaceOrder = newOrder
		changed = true
	}
	ed.registerRegisteredWorkspaceActions()
	return changed
}

// initializeWorkspaces calls Initialize on any active workspace that has not
// yet been initialized. Idempotent — safe to call repeatedly after
// reconciliation; previously-initialized workspaces are skipped via the
// initializedWorkspaces tracker.
func (ed *Editor) initializeWorkspaces() {
	defer tracing.NewRegion("Editor.initializeWorkspaces").End()
	for _, id := range ed.workspaceOrder {
		if _, done := ed.initializedWorkspaces[id]; done {
			continue
		}
		w := ed.activeWorkspaces[id]
		if err := w.Initialize(ed); err != nil {
			slog.Error("failed to initialize workspace", "id", id, "error", err)
			continue
		}
		ed.initializedWorkspaces[id] = struct{}{}
	}
}

// firstSelectableWorkspaceID returns the id of the first enabled workspace
// in load order. Returns WorkspaceStateNone if nothing is enabled (which
// should be impossible because Stage, Content, and Settings are required).
func (ed *Editor) firstSelectableWorkspaceID() WorkspaceState {
	for _, cfg := range ed.settings.Workspaces {
		if cfg.Enabled {
			return cfg.ID
		}
	}
	return WorkspaceStateNone
}

// ApplyWorkspaceConfigChanges is called by the settings workspace after the
// user toggles enable/visible or reorders a workspace. Re-runs reconciliation
// (which may shut down workspaces that are now disabled, initialize newly-
// enabled ones), rebuilds the menu bar tab strip, and switches the active
// workspace if the current one has been disabled.
func (ed *Editor) ApplyWorkspaceConfigChanges() {
	defer tracing.NewRegion("Editor.ApplyWorkspaceConfigChanges").End()
	ed.reconcileWorkspaces()
	ed.initializeWorkspaces()
	ed.rebuildMenuBarTabs()
	if _, ok := ed.activeWorkspaces[ed.workspaceState]; !ok {
		// Current workspace was disabled. Fall back to first selectable.
		if id := ed.firstSelectableWorkspaceID(); id != WorkspaceStateNone {
			ed.workspaceState = WorkspaceStateNone // force setWorkspaceState to switch
			ed.setWorkspaceState(id)
		}
	}
}

// rebuildMenuBarTabs pushes the current ordered, enabled workspace set into
// the menu bar. Called on initial load, when a plugin registers a workspace
// late, and when the user toggles enable/order in settings.
func (ed *Editor) rebuildMenuBarTabs() {
	defer tracing.NewRegion("Editor.rebuildMenuBarTabs").End()
	tabs := make([]menu_bar.WorkspaceTab, 0, len(ed.workspaceOrder))
	for _, id := range ed.workspaceOrder {
		w := ed.activeWorkspaces[id]
		tabs = append(tabs, menu_bar.WorkspaceTab{
			ID:          id,
			DisplayName: w.DisplayName(),
		})
	}
	ed.globalInterfaces.menuBar.RebuildWorkspaceTabs(tabs, ed.workspaceState)
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (ed *Editor) update(deltaTime float64) {
	ed.updatePowerState(deltaTime)
	if ed.blurred > 0 {
		return
	}
	if context_menu.IsOpen() {
		if ed.currentWorkspace != nil {
			ed.currentWorkspace.Update(deltaTime)
		}
		return
	}
	kb := &ed.host.Window.Keyboard
	if ed.processActionPaletteShortcut(kb) {
		return
	}
	if ed.processActionKeyBindings(kb) {
		return
	}
	if ed.currentWorkspace != nil {
		if !ed.stageView.IsFlyCameraInputActive() {
			processWorkspaceHotkeys(ed, kb)
		}
		ed.currentWorkspace.Update(deltaTime)
	}
}

func (ed *Editor) updatePowerState(deltaTime float64) {
	if !ed.settings.UseBatteryRefreshRate {
		return
	}
	if !ed.power.initialized {
		ed.setFrameRateLimitForPowerStatus(ed.queryAndCachePowerStatus())
		return
	}
	ed.power.pollElapsed += deltaTime
	if ed.power.pollElapsed < editorPowerPollInterval {
		return
	}
	ed.power.pollElapsed = 0
	status := ed.queryPowerStatus()
	if status.Source == ed.power.lastStatus.Source && status.HasBattery == ed.power.lastStatus.HasBattery {
		ed.power.lastStatus = status
		return
	}
	ed.power.lastStatus = status
	ed.setFrameRateLimitForPowerStatus(status)
}

func processWorkspaceHotkeys(ed *Editor, kb *hid.Keyboard) {
	for _, hk := range ed.currentWorkspace.Hotkeys() {
		if hk.Ctrl && !kb.HasCtrl() {
			continue
		}
		if hk.Meta && !kb.HasMeta() {
			continue
		}
		if hk.Shift && !kb.HasShift() {
			continue
		}
		if hk.Alt && !kb.HasAlt() {
			continue
		}
		down := false
		valid := true
		for i := 0; i < len(hk.Keys) && valid; i++ {
			valid = kb.KeyHeld(hk.Keys[i])
			down = down || kb.KeyDown(hk.Keys[i])
		}
		if valid && down {
			hk.Call()
		}
	}
}

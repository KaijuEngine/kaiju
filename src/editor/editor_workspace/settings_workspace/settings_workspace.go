/******************************************************************************/
/* settings_workspace.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package settings_workspace

import (
	"log/slog"
	"strconv"
	"strings"

	"kaijuengine.com/editor/editor_plugin"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	ID          = "settings"
	DisplayName = "Settings"

	uiFile = "editor/ui/workspace/settings_workspace.go.html"
)

func init() {
	editor_workspace_registry.Register(&SettingsWorkspace{})
}

// editorRecompiler is a typed-service interface the settings workspace
// asserts onto the editor to trigger a plugin-aware recompile. Defined here
// (not on WorkspaceEditorInterface) because only this workspace needs it —
// the editor satisfies it implicitly by having the matching method.
type editorRecompiler interface {
	RecompileWithPlugins(plugins []editor_plugin.PluginInfo, onComplete func(err error)) error
}

// editorWorkspaceController is the typed-service interface for live
// workspace lifecycle management (apply enable/visible/order changes from
// settings UI). The editor satisfies it implicitly.
type editorWorkspaceController interface {
	ApplyWorkspaceConfigChanges()
}

type SettingsWorkspace struct {
	common_workspace.CommonWorkspace
	projectSettingsBox   *document.Element
	editorSettingsBox    *document.Element
	pluginSettingsBox    *document.Element
	workspaceSettingsBox *document.Element
	keyboardSettingsBox  *document.Element
	editor               editor_workspace.WorkspaceEditorInterface
	editorSettings       *editor_settings.Settings
	projectSettings      *project.Settings
	plugins              []editor_plugin.PluginInfo
	pluginInitStates     []bool
	reloadRequested      bool
	recompiling          bool
	downloadingPlugin    bool
	shortcutCapture      *shortcutCaptureState
	isSettingKeybinding  bool
}

// workspaceRowData is the per-row data the Workspaces panel template loops over.
// Built fresh on every UI render so the displayed state reflects what's
// actually persisted in settings + what's actually registered.
type workspaceRowData struct {
	ID          string
	DisplayName string
	Enabled     bool
	IsRequired  bool
}

type settingsWorkspaceData struct {
	Editor     common_workspace.DataUISection
	Project    common_workspace.DataUISection
	WebAPI     editor_settings.WebAPISettings
	Plugins    []editor_plugin.PluginInfo
	Workspaces []workspaceRowData
	Shortcuts  []shortcutSectionData
}

func (w *SettingsWorkspace) ID() string          { return ID }
func (w *SettingsWorkspace) DisplayName() string { return DisplayName }
func (w *SettingsWorkspace) IsRequired() bool    { return true }

func (w *SettingsWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	host := ed.Host()
	w.editor = ed
	w.editorSettings = ed.Settings()
	w.projectSettings = &ed.Project().Settings
	if err := w.CommonWorkspace.InitializeWithUI(host, uiFile, w.uiData(), w.funcMap()); err != nil {
		return err
	}
	w.reloadedUI()
	w.editor.Events().OnContentRemoved.Add(func(ids []string) {
		for i := range ids {
			if ids[i] == w.projectSettings.EntryPointStage {
				w.projectSettings.EntryPointStage = ""
				w.projectSettings.Save(w.editor.Project().FileSystem())
				w.RequestReload()
			}
		}
	})
	return nil
}

func (w *SettingsWorkspace) Shutdown() {
	defer tracing.NewRegion("SettingsWorkspace.Shutdown").End()
	w.stopShortcutCapture()
	w.CommonShutdown()
}

func (w *SettingsWorkspace) IsFocusedOnInput() bool {
	return w.isSettingKeybinding || w.CommonWorkspace.IsFocusedOnInput()
}

func (w *SettingsWorkspace) Open() {
	defer tracing.NewRegion("SettingsWorkspace.Open").End()
	if w.reloadRequested {
		w.ReloadUI(uiFile, w.uiData(), w.funcMap())
		w.reloadedUI()
	}
	w.CommonOpen()
	w.projectSettingsBox.UI.Show()
	w.editorSettingsBox.UI.Hide()
	w.pluginSettingsBox.UI.Hide()
	w.workspaceSettingsBox.UI.Hide()
	w.keyboardSettingsBox.UI.Hide()
	w.resetLeftEntrySelection()
	for _, e := range w.Doc.GetElementsByClass("edPanelBgHoverable") {
		if e.InnerLabel().Text() == "Project Settings" {
			w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
			break
		}
	}
}

func (w *SettingsWorkspace) Close() {
	defer tracing.NewRegion("SettingsWorkspace.Close").End()
	w.stopShortcutCapture()
	w.CommonClose()
	w.projectSettings.Save(w.editor.ProjectFileSystem())
	w.editor.UpdateSettings()
}

func (w *SettingsWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *SettingsWorkspace) RequestReload() { w.reloadRequested = true }

func (w *SettingsWorkspace) resetLeftEntrySelection() {
	for _, elm := range w.Doc.GetElementsByClass("edPanelBgHoverable") {
		w.Doc.SetElementClassesWithoutApply(elm, "edPanelBgHoverable")
	}
}

func (w *SettingsWorkspace) showProjectSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showProjectSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
	w.projectSettingsBox.UI.Show()
	w.editorSettingsBox.UI.Hide()
	w.pluginSettingsBox.UI.Hide()
	w.workspaceSettingsBox.UI.Hide()
	w.keyboardSettingsBox.UI.Hide()
}

func (w *SettingsWorkspace) showEditorSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showEditorSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
	w.projectSettingsBox.UI.Hide()
	w.editorSettingsBox.UI.Show()
	w.pluginSettingsBox.UI.Hide()
	w.workspaceSettingsBox.UI.Hide()
	w.keyboardSettingsBox.UI.Hide()
}

func (w *SettingsWorkspace) showPluginSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showPluginSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
	w.projectSettingsBox.UI.Hide()
	w.editorSettingsBox.UI.Hide()
	w.pluginSettingsBox.UI.Show()
	w.workspaceSettingsBox.UI.Hide()
	w.keyboardSettingsBox.UI.Hide()
}

func (w *SettingsWorkspace) showWorkspaceSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showWorkspaceSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
	w.projectSettingsBox.UI.Hide()
	w.editorSettingsBox.UI.Hide()
	w.pluginSettingsBox.UI.Hide()
	w.workspaceSettingsBox.UI.Show()
	w.keyboardSettingsBox.UI.Hide()
}

func (w *SettingsWorkspace) showKeyboardSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showKeyboardSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
	w.projectSettingsBox.UI.Hide()
	w.editorSettingsBox.UI.Hide()
	w.pluginSettingsBox.UI.Hide()
	w.workspaceSettingsBox.UI.Hide()
	w.keyboardSettingsBox.UI.Show()
	w.applyShortcutFilter()
}

// toggleWorkspaceEnabled toggles a non-required workspace's enabled flag.
// Required workspaces don't render a checkbox in the template, so this
// handler never sees them — the registry-side IsRequired check is just
// belt-and-suspenders against a malformed template.
func (w *SettingsWorkspace) toggleWorkspaceEnabled(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.toggleWorkspaceEnabled").End()
	idx, err := strconv.Atoi(e.Attribute("data-idx"))
	if err != nil || idx < 0 || idx >= len(w.editorSettings.Workspaces) {
		return
	}
	cfg := &w.editorSettings.Workspaces[idx]
	if ws, ok := editor_workspace_registry.Get(cfg.ID); ok && ws.IsRequired() {
		return // required: ignore and let reconcile force it back next pass
	}
	cfg.Enabled = !cfg.Enabled
	w.applyWorkspaceChanges(false)
}

func (w *SettingsWorkspace) moveWorkspaceUp(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.moveWorkspaceUp").End()
	idx, err := strconv.Atoi(e.Attribute("data-idx"))
	if err != nil || idx <= 0 || idx >= len(w.editorSettings.Workspaces) {
		return
	}
	ws := w.editorSettings.Workspaces
	ws[idx-1], ws[idx] = ws[idx], ws[idx-1]
	w.applyWorkspaceChanges(true)
}

func (w *SettingsWorkspace) moveWorkspaceDown(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.moveWorkspaceDown").End()
	idx, err := strconv.Atoi(e.Attribute("data-idx"))
	if err != nil || idx < 0 || idx >= len(w.editorSettings.Workspaces)-1 {
		return
	}
	ws := w.editorSettings.Workspaces
	ws[idx], ws[idx+1] = ws[idx+1], ws[idx]
	w.applyWorkspaceChanges(true)
}

// applyWorkspaceChanges persists settings and asks the editor to reconcile
// (which may shut down disabled workspaces, initialize newly-enabled ones,
// and rebuild the menu bar tab strip).
//
// reloadUI=true triggers an immediate UI rebuild of the Workspaces panel so
// reorder buttons feel responsive. Toggle handlers pass false because the
// browser already updated the checkbox state visually.
func (w *SettingsWorkspace) applyWorkspaceChanges(reloadUI bool) {
	if err := w.editorSettings.Save(); err != nil {
		slog.Error("failed to persist workspace settings", "error", err)
	}
	if c, ok := w.editor.(editorWorkspaceController); ok {
		c.ApplyWorkspaceConfigChanges()
	}
	if reloadUI {
		// Defer to next frame so we're not destroying the document from
		// inside a click handler that's still iterating through the DOM.
		w.Host.RunNextFrame(func() {
			w.ReloadUI(uiFile, w.uiData(), w.funcMap())
			w.reloadedUI()
			w.CommonOpen()
			w.workspaceSettingsBox.UI.Show()
			w.projectSettingsBox.UI.Hide()
			w.editorSettingsBox.UI.Hide()
			w.pluginSettingsBox.UI.Hide()
			w.keyboardSettingsBox.UI.Hide()
			for _, e := range w.Doc.GetElementsByClass("edPanelBgHoverable") {
				if e.InnerLabel().Text() == "Workspaces" {
					w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
					break
				}
			}
		})
	}
}

func (w *SettingsWorkspace) valueChanged(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.valueChanged").End()
	if w.editorSettingsBox.UI.Entity().IsActive() {
		common_workspace.SetObjectValueFromUI(w.editorSettings, e)
	} else if w.projectSettingsBox.UI.Entity().IsActive() {
		common_workspace.SetObjectValueFromUI(w.projectSettings, e)
	}
}

func (w *SettingsWorkspace) webAPIValueChanged(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.webAPIValueChanged").End()
	common_workspace.SetObjectValueFromUI(w.editorSettings, e)
	w.editorSettings.NormalizeWebAPI()
	w.syncWebAPIInputs()
	w.editor.UpdateSettings()
}

func (w *SettingsWorkspace) rotateWebAPIKey(*document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.rotateWebAPIKey").End()
	w.editorSettings.WebAPI.APIKey = editor_settings.GenerateWebAPIKey()
	w.syncWebAPIInputs()
	w.editor.UpdateSettings()
}

func (w *SettingsWorkspace) openPluginWebsite(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.openPluginWebsite").End()
	klib.OpenWebsite(e.InnerLabel().Text())
}

func (w *SettingsWorkspace) togglePlugin(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.togglePlugin").End()
	idx, err := strconv.Atoi(e.Attribute("data-idx"))
	if err != nil {
		return
	}

	plugin := &w.plugins[idx]

	// Handle Git plugins differently - remove from storage when disabled
	if strings.HasPrefix(plugin.Path, "git://") {
		if plugin.Config.Enabled {
			// Disable Git plugin by removing from storage
			gitPlugin := strings.TrimPrefix(plugin.Path, "git://")
			if err := editor_plugin.RemoveGitPluginFromStorage(gitPlugin); err != nil {
				slog.Error("failed to remove Git plugin from storage", "plugin", gitPlugin, "error", err)
				return
			}
			plugin.Config.Enabled = false
			slog.Info("Git plugin removed from storage", "plugin", gitPlugin)
		}
		// Note: Re-enabling Git plugins would require re-adding their URL,
		// so we don't handle that case here
		return
	}

	// Local plugin: flip in-memory state AND persist to plugin.json
	// immediately. The RecompileWithPlugins flow only writes plugin.json
	// for *enabled* plugins (it skips disabled ones in its main loop), so
	// without this disk write a disable never sticks across editor sessions
	// and the user would have to hand-edit the file.
	plugin.Config.Enabled = !plugin.Config.Enabled
	if err := editor_plugin.UpdatePluginConfigState(*plugin); err != nil {
		slog.Error("failed to persist plugin enabled state",
			"name", plugin.Config.Name, "package", plugin.Config.PackageName, "error", err)
	}
}

func (w *SettingsWorkspace) clickOpenPlugins(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.togglePlugin").End()
	folder, err := editor_plugin.PluginsFolder()
	if err != nil {
		slog.Error("failed to find the plugins folder", "error", err)
		return
	}
	if err = filesystem.OpenFileBrowserToFolder(folder); err != nil {
		slog.Error("failed to open the file browser to folder", "folder", folder, "error", err)
	}
}

func (w *SettingsWorkspace) recompileEditor(*document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.recompileEditor").End()
	if w.recompiling {
		slog.Warn("the editor is already in the process of recompiling, please wait")
		return
	}
	pluginsChanged := len(w.plugins) != len(w.pluginInitStates)
	if !pluginsChanged {
		for i := 0; i < len(w.plugins) && i < len(w.pluginInitStates) && !pluginsChanged; i++ {
			pluginsChanged = w.pluginInitStates[i] != w.plugins[i].Config.Enabled
		}
	}
	if !pluginsChanged {
		slog.Warn("plugin settings have not changed, no reason to recompile")
		return
	}
	r, ok := w.editor.(editorRecompiler)
	if !ok {
		slog.Error("editor does not support plugin recompilation")
		return
	}
	w.recompiling = true
	err := r.RecompileWithPlugins(w.plugins, func(err error) {
		w.recompiling = false
	})
	if err != nil {
		slog.Error("failed to compile the editor", "error", err)
		w.recompiling = false
	}
}

func (w *SettingsWorkspace) addPluginFromGit(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.addPluginFromGit").End()

	if w.downloadingPlugin {
		slog.Warn("a plugin download is already in progress, please wait")
		return
	}
	if w.recompiling {
		slog.Warn("the editor is already in the process of recompiling, please wait")
		return
	}

	// Get Git URL from the input field
	gitUrlElement, found := w.Doc.GetElementById("gitPluginUrl")
	if !found {
		slog.Error("Git URL input element not found")
		return
	}
	gitUrlElement.UI.ToInput().SetPlaceholder("Processing...")

	gitURL := gitUrlElement.UI.ToInput().Text()
	if gitURL == "" {
		slog.Warn("Git URL is empty")
		return
	}

	r, ok := w.editor.(editorRecompiler)
	if !ok {
		slog.Error("editor does not support plugin recompilation")
		return
	}

	w.downloadingPlugin = true

	// Show processing status
	statusElement, found := w.Doc.GetElementById("pluginDownloadStatus")
	if found && statusElement.UI != nil {
		if innerLabel := statusElement.InnerLabel(); innerLabel != nil {
			innerLabel.SetText("Processing Git URL...")
		}
		statusElement.UI.Show()
	}

	// Process the Git URL to get module path
	modulePath, err := editor_plugin.AddPluginFromGit(gitURL)
	if err != nil {
		w.downloadingPlugin = false
		slog.Error("failed to process Git URL", "url", gitURL, "error", err)
		if statusElement, found := w.Doc.GetElementById("pluginDownloadStatus"); found && statusElement.UI != nil {
			if innerLabel := statusElement.InnerLabel(); innerLabel != nil {
				innerLabel.SetText("Failed to process Git URL: " + err.Error())
			}
		}
		return
	}

	// Update status before recompilation
	if statusElement, found := w.Doc.GetElementById("pluginDownloadStatus"); found && statusElement.UI != nil {
		if innerLabel := statusElement.InnerLabel(); innerLabel != nil {
			innerLabel.SetText("Git plugin '" + modulePath + "' added. Recompiling editor...")
		}
	}

	// Clear the input field
	if urlElement, found := w.Doc.GetElementById("gitPluginUrl"); found {
		urlElement.UI.ToInput().SetText("")
	}

	// Refresh the plugins list and recompile immediately
	w.plugins = editor_plugin.AvailablePlugins()
	w.pluginInitStates = make([]bool, len(w.plugins))
	for i := range w.plugins {
		w.pluginInitStates[i] = w.plugins[i].Config.Enabled
	}

	w.recompiling = true
	if err := r.RecompileWithPlugins(w.plugins, func(err error) {
		w.recompiling = false
		if err != nil {
			slog.Error("failed to compile the editor", "error", err)
			if statusElement, found := w.Doc.GetElementById("pluginDownloadStatus"); found && statusElement.UI != nil {
				if innerLabel := statusElement.InnerLabel(); innerLabel != nil {
					innerLabel.SetText("Failed to compile editor: " + err.Error())
				}
			}
		}
	}); err != nil {
		w.recompiling = false
		w.downloadingPlugin = false
		slog.Error("failed to compile the editor", "error", err)
		if statusElement, found := w.Doc.GetElementById("pluginDownloadStatus"); found && statusElement.UI != nil {
			if innerLabel := statusElement.InnerLabel(); innerLabel != nil {
				innerLabel.SetText("Failed to compile editor: " + err.Error())
			}
		}
		return
	}

	w.downloadingPlugin = false
	slog.Info("Git plugin added and recompilation started", "module", modulePath)
}

func (w *SettingsWorkspace) uiData() settingsWorkspaceData {
	defer tracing.NewRegion("SettingsWorkspace.uiData").End()
	w.plugins = editor_plugin.AvailablePlugins()
	w.pluginInitStates = make([]bool, len(w.plugins))
	for i := range w.plugins {
		w.pluginInitStates[i] = w.plugins[i].Config.Enabled
	}
	listings := map[string][]ui.SelectOption{}
	cache := w.editor.Project().CacheDatabase()
	return settingsWorkspaceData{
		Editor: common_workspace.ReflectUIStructure(cache,
			w.editorSettings, "", listings),
		Project: common_workspace.ReflectUIStructure(cache,
			w.projectSettings, "", listings),
		WebAPI:     w.editorSettings.WebAPI,
		Plugins:    w.plugins,
		Workspaces: w.buildWorkspaceRows(),
		Shortcuts:  w.buildShortcutSections(),
	}
}

// buildWorkspaceRows joins the persisted settings.Workspaces order with the
// live registry to produce display data for the Workspaces panel. Settings
// is the source of truth for order; if settings is empty the registry's
// registration order is used as a fallback (the editor reconcile step
// should have populated settings during postProjectLoad, but render before
// that should still produce a sensible list).
func (w *SettingsWorkspace) buildWorkspaceRows() []workspaceRowData {
	rows := make([]workspaceRowData, 0, len(w.editorSettings.Workspaces))
	for _, cfg := range w.editorSettings.Workspaces {
		ws, ok := editor_workspace_registry.Get(cfg.ID)
		if !ok {
			// Stale entry — show id only so the user knows it exists, but
			// it's effectively unmanageable until the workspace is restored.
			rows = append(rows, workspaceRowData{
				ID:          cfg.ID,
				DisplayName: cfg.ID + " (missing)",
				Enabled:     cfg.Enabled,
				IsRequired:  false,
			})
			continue
		}
		rows = append(rows, workspaceRowData{
			ID:          cfg.ID,
			DisplayName: ws.DisplayName(),
			Enabled:     cfg.Enabled,
			IsRequired:  ws.IsRequired(),
		})
	}
	return rows
}

func (w *SettingsWorkspace) funcMap() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"showProjectSettings":    w.showProjectSettings,
		"showEditorSettings":     w.showEditorSettings,
		"showPluginSettings":     w.showPluginSettings,
		"showWorkspaceSettings":  w.showWorkspaceSettings,
		"showKeyboardSettings":   w.showKeyboardSettings,
		"valueChanged":           w.valueChanged,
		"openPluginWebsite":      w.openPluginWebsite,
		"togglePlugin":           w.togglePlugin,
		"clickOpenPlugins":       w.clickOpenPlugins,
		"recompileEditor":        w.recompileEditor,
		"addPluginFromGit":       w.addPluginFromGit,
		"toggleWorkspaceEnabled": w.toggleWorkspaceEnabled,
		"moveWorkspaceUp":        w.moveWorkspaceUp,
		"moveWorkspaceDown":      w.moveWorkspaceDown,
		"webAPIValueChanged":     w.webAPIValueChanged,
		"rotateWebAPIKey":        w.rotateWebAPIKey,
		"captureShortcut":        w.captureShortcut,
		"clearShortcut":          w.clearShortcut,
		"filterShortcuts":        w.filterShortcuts,
		"resetShortcuts":         w.resetShortcuts,
		"exportShortcuts":        w.exportShortcuts,
		"importShortcuts":        w.importShortcuts,
	}
}

func (w *SettingsWorkspace) reloadedUI() {
	defer tracing.NewRegion("SettingsWorkspace.reloadedUI").End()
	w.reloadRequested = false
	w.projectSettingsBox, _ = w.Doc.GetElementById("projectSettingsBox")
	w.editorSettingsBox, _ = w.Doc.GetElementById("editorSettingsBox")
	w.pluginSettingsBox, _ = w.Doc.GetElementById("pluginSettingsBox")
	w.workspaceSettingsBox, _ = w.Doc.GetElementById("workspaceSettingsBox")
	w.keyboardSettingsBox, _ = w.Doc.GetElementById("keyboardSettingsBox")
	w.syncWebAPIInputs()
}

func (w *SettingsWorkspace) syncWebAPIInputs() {
	if w.Doc == nil {
		return
	}
	if elm, ok := w.Doc.GetElementById("webAPIEnabled"); ok && elm.UI != nil {
		elm.UI.ToCheckbox().SetCheckedWithoutEvent(w.editorSettings.WebAPI.Enabled)
	}
	if elm, ok := w.Doc.GetElementById("webAPIPort"); ok && elm.UI != nil {
		elm.UI.ToInput().SetTextWithoutEvent(strconv.Itoa(int(w.editorSettings.WebAPI.Port)))
	}
	if elm, ok := w.Doc.GetElementById("webAPIKey"); ok && elm.UI != nil {
		elm.UI.ToInput().SetTextWithoutEvent(w.editorSettings.WebAPI.APIKey)
	}
}

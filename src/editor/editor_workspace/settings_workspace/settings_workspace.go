/******************************************************************************/
/* settings_workspace.go                                                      */
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
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package settings_workspace

import (
	"kaiju/editor/editor_plugin"
	"kaiju/editor/editor_settings"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/project"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"strconv"
)

const uiFile = "editor/ui/workspace/settings_workspace.go.html"

type SettingsWorkspace struct {
	common_workspace.CommonWorkspace
	projectSettingsBox *document.Element
	editorSettingsBox  *document.Element
	pluginSettingsBox  *document.Element
	editor             SettingsWorkspaceEditorInterface
	editorSettings     *editor_settings.Settings
	projectSettings    *project.Settings
	plugins            []editor_plugin.PluginInfo
	pluginInitStates   []bool
	reloadRequested    bool
	recompiling        bool
}

type settingsWorkspaceData struct {
	Editor  common_workspace.DataUISection
	Project common_workspace.DataUISection
	Plugins []editor_plugin.PluginInfo
}

func (w *SettingsWorkspace) Initialize(host *engine.Host, editor SettingsWorkspaceEditorInterface) {
	w.editor = editor
	w.editorSettings = editor.Settings()
	w.projectSettings = &editor.Project().Settings
	w.CommonWorkspace.InitializeWithUI(host, uiFile, w.uiData(), w.funcMap())
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
}

func (w *SettingsWorkspace) showEditorSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showProjectSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
	w.projectSettingsBox.UI.Hide()
	w.editorSettingsBox.UI.Show()
	w.pluginSettingsBox.UI.Hide()
}

func (w *SettingsWorkspace) showPluginSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showPluginSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
	w.projectSettingsBox.UI.Hide()
	w.editorSettingsBox.UI.Hide()
	w.pluginSettingsBox.UI.Show()
}

func (w *SettingsWorkspace) valueChanged(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.valueChanged").End()
	if w.editorSettingsBox.UI.Entity().IsActive() {
		common_workspace.SetObjectValueFromUI(w.editorSettings, e)
	} else if w.projectSettingsBox.UI.Entity().IsActive() {
		common_workspace.SetObjectValueFromUI(w.projectSettings, e)
	}
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
	w.plugins[idx].Config.Enabled = !w.plugins[idx].Config.Enabled
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
	pluginsChanged := false
	for i := 0; i < len(w.plugins) && !pluginsChanged; i++ {
		pluginsChanged = w.pluginInitStates[i] != w.plugins[i].Config.Enabled
	}
	if !pluginsChanged {
		slog.Warn("plugin settings have not changed, no reason to recompile")
		return
	}
	w.recompiling = true
	err := w.editor.RecompileWithPlugins(w.plugins, func(err error) {
		w.recompiling = false
	})
	if err != nil {
		slog.Error("failed to compile the editor", "error", err)
		w.recompiling = false
	}
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
		Plugins: w.plugins,
	}
}

func (w *SettingsWorkspace) funcMap() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"showProjectSettings": w.showProjectSettings,
		"showEditorSettings":  w.showEditorSettings,
		"showPluginSettings":  w.showPluginSettings,
		"valueChanged":        w.valueChanged,
		"openPluginWebsite":   w.openPluginWebsite,
		"togglePlugin":        w.togglePlugin,
		"clickOpenPlugins":    w.clickOpenPlugins,
		"recompileEditor":     w.recompileEditor,
	}
}

func (w *SettingsWorkspace) reloadedUI() {
	defer tracing.NewRegion("SettingsWorkspace.reloadedUI").End()
	w.reloadRequested = false
	w.projectSettingsBox, _ = w.Doc.GetElementById("projectSettingsBox")
	w.editorSettingsBox, _ = w.Doc.GetElementById("editorSettingsBox")
	w.pluginSettingsBox, _ = w.Doc.GetElementById("pluginSettingsBox")
}

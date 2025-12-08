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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"github.com/KaijuEngine/kaiju/editor/editor_settings"
	"github.com/KaijuEngine/kaiju/editor/editor_workspace/common_workspace"
	"github.com/KaijuEngine/kaiju/editor/project"
	"github.com/KaijuEngine/kaiju/engine"
	"github.com/KaijuEngine/kaiju/engine/ui"
	"github.com/KaijuEngine/kaiju/engine/ui/markup/document"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
)

const uiFile = "editor/ui/workspace/settings_workspace.go.html"

type SettingsWorkspace struct {
	common_workspace.CommonWorkspace
	projectSettingsBox *document.Element
	editorSettingsBox  *document.Element
	editor             SettingsWorkspaceEditorInterface
	editorSettings     *editor_settings.Settings
	projectSettings    *project.Settings
	reloadRequested    bool
}

type settingsWorkspaceData struct {
	Editor  common_workspace.DataUISection
	Project common_workspace.DataUISection
}

func (w *SettingsWorkspace) Initialize(host *engine.Host, editor SettingsWorkspaceEditorInterface) {
	w.editor = editor
	w.editorSettings = editor.Settings()
	w.projectSettings = editor.Project().Settings()
	w.CommonWorkspace.InitializeWithUI(host, uiFile, w.uiData(), w.funcMap())
	w.reloadedUI()
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
	w.resetLeftEntrySelection()
	for _, e := range w.Doc.GetElementsByClass("leftEntry") {
		if e.InnerLabel().Text() == "Project Settings" {
			w.Doc.SetElementClasses(e, "leftEntry", "leftEntrySelected")
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
	for _, elm := range w.Doc.GetElementsByClass("leftEntry") {
		w.Doc.SetElementClassesWithoutApply(elm, "leftEntry")
	}
}

func (w *SettingsWorkspace) showProjectSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showProjectSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "leftEntry", "leftEntrySelected")
	w.projectSettingsBox.UI.Show()
	w.editorSettingsBox.UI.Hide()
}

func (w *SettingsWorkspace) showEditorSettings(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.showProjectSettings").End()
	w.resetLeftEntrySelection()
	w.Doc.SetElementClasses(e, "leftEntry", "leftEntrySelected")
	w.projectSettingsBox.UI.Hide()
	w.editorSettingsBox.UI.Show()
}

func (w *SettingsWorkspace) valueChanged(e *document.Element) {
	if w.editorSettingsBox.UI.Entity().IsActive() {
		common_workspace.SetObjectValueFromUI(w.editorSettings, e)
	} else if w.projectSettingsBox.UI.Entity().IsActive() {
		common_workspace.SetObjectValueFromUI(w.projectSettings, e)
	}
}

func (w *SettingsWorkspace) uiData() settingsWorkspaceData {
	listings := map[string][]ui.SelectOption{}
	return settingsWorkspaceData{
		Editor:  common_workspace.ReflectUIStructure(w.editorSettings, "", listings),
		Project: common_workspace.ReflectUIStructure(w.projectSettings, "", listings),
	}
}

func (w *SettingsWorkspace) funcMap() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"showProjectSettings": w.showProjectSettings,
		"showEditorSettings":  w.showEditorSettings,
		"valueChanged":        w.valueChanged,
	}
}

func (w *SettingsWorkspace) reloadedUI() {
	w.reloadRequested = false
	w.projectSettingsBox, _ = w.Doc.GetElementById("projectSettingsBox")
	w.editorSettingsBox, _ = w.Doc.GetElementById("editorSettingsBox")
}

func isUnderParentId(e *document.Element, id string) bool {
	p := e
	for p != nil {
		if p.Attribute("id") == id {
			return true
		}
	}
	return false
}

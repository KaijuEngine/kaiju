/******************************************************************************/
/* scripts_workspace.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package scripts_workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"kaijuengine.com/editor/editor_scripting"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	ID          = "scripts"
	DisplayName = "Scripts"

	uiFile = "editor/ui/workspace/scripts_workspace.go.html"
)

func init() {
	editor_workspace_registry.Register(&ScriptsWorkspace{})
}

type scriptEntry struct {
	Name string
	Path string
}

type scriptsWorkspaceData struct {
	Scripts []scriptEntry
}

type ScriptsWorkspace struct {
	common_workspace.CommonWorkspace
	ed             editor_workspace.WorkspaceEditorInterface
	scriptList     *document.Element
	scriptTemplate *document.Element
	output         *document.Element
	lastOutput     strings.Builder
}

func (w *ScriptsWorkspace) ID() string          { return ID }
func (w *ScriptsWorkspace) DisplayName() string { return DisplayName }
func (w *ScriptsWorkspace) IsRequired() bool    { return false }

func (w *ScriptsWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("ScriptsWorkspace.Initialize").End()
	w.ed = ed
	if err := w.ensureScriptsFolder(); err != nil {
		return err
	}
	if err := w.CommonWorkspace.InitializeWithUI(ed.Host(), uiFile, w.uiData(), w.funcMap()); err != nil {
		return err
	}
	w.scriptList, _ = w.Doc.GetElementById("scriptList")
	w.scriptTemplate, _ = w.Doc.GetElementById("scriptEntryTemplate")
	w.output, _ = w.Doc.GetElementById("scriptOutput")
	w.populateScripts()
	return nil
}

func (w *ScriptsWorkspace) Shutdown() {
	defer tracing.NewRegion("ScriptsWorkspace.Shutdown").End()
	w.CommonShutdown()
}

func (w *ScriptsWorkspace) Open() {
	defer tracing.NewRegion("ScriptsWorkspace.Open").End()
	w.refresh()
	w.CommonOpen()
}

func (w *ScriptsWorkspace) Close() {
	defer tracing.NewRegion("ScriptsWorkspace.Close").End()
	w.CommonClose()
}

func (w *ScriptsWorkspace) Hotkeys() []common_workspace.HotKey { return nil }

func (w *ScriptsWorkspace) WriteScriptLog(message string) {
	w.lastOutput.WriteString(message)
	if !strings.HasSuffix(message, "\n") {
		w.lastOutput.WriteString("\n")
	}
	w.setOutput(w.lastOutput.String())
}

func (w *ScriptsWorkspace) funcMap() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"clickRunScript": w.clickRunScript,
		"clickReload":    w.clickReload,
	}
}

func (w *ScriptsWorkspace) uiData() scriptsWorkspaceData {
	return scriptsWorkspaceData{Scripts: w.scripts()}
}

func (w *ScriptsWorkspace) scripts() []scriptEntry {
	if err := w.ensureScriptsFolder(); err != nil {
		return nil
	}
	entries, err := w.ed.ProjectFileSystem().ReadDir(project_file_system.EditorScriptsFolder)
	if err != nil {
		return nil
	}
	scripts := make([]scriptEntry, 0, len(entries))
	for i := range entries {
		if entries[i].IsDir() || strings.ToLower(filepath.Ext(entries[i].Name())) != ".lua" {
			continue
		}
		path := filepath.ToSlash(filepath.Join(project_file_system.EditorScriptsFolder, entries[i].Name()))
		scripts = append(scripts, scriptEntry{Name: entries[i].Name(), Path: path})
	}
	return scripts
}

func (w *ScriptsWorkspace) populateScripts() {
	if w.scriptList == nil || w.scriptTemplate == nil {
		return
	}
	for i := len(w.scriptList.Children) - 1; i >= 0; i-- {
		if w.scriptList.Children[i] != w.scriptTemplate {
			w.Doc.RemoveElement(w.scriptList.Children[i])
		}
	}
	scripts := w.scripts()
	cpys := w.Doc.DuplicateElementRepeatWithoutApplyStyles(w.scriptTemplate, len(scripts))
	for i := range cpys {
		w.Doc.SetElementIdWithoutApplyStyles(cpys[i], fmt.Sprintf("scriptEntry%d", i))
		cpys[i].SetAttribute("data-path", scripts[i].Path)
		cpys[i].Children[0].InnerLabel().SetText(scripts[i].Name)
	}
	w.Doc.ApplyStyles()
}

func (w *ScriptsWorkspace) clickRunScript(e *document.Element) {
	path := e.Attribute("data-path")
	if path == "" {
		return
	}
	w.lastOutput.Reset()
	w.WriteScriptLog(fmt.Sprintf("Running %s", path))
	fullPath := w.ed.ProjectFileSystem().FullPath(path)
	if err := editor_scripting.RunEditorScript(w.ed, fullPath, w); err != nil {
		w.WriteScriptLog("Error: " + err.Error())
		return
	}
	w.WriteScriptLog("Done")
}

func (w *ScriptsWorkspace) clickReload(*document.Element) {
	w.refresh()
}

func (w *ScriptsWorkspace) refresh() {
	if err := w.ensureScriptsFolder(); err != nil {
		w.setOutput(err.Error())
		return
	}
	w.populateScripts()
}

func (w *ScriptsWorkspace) ensureScriptsFolder() error {
	return w.ed.ProjectFileSystem().MkdirAll(project_file_system.EditorScriptsFolder, 0755)
}

func (w *ScriptsWorkspace) setOutput(text string) {
	if w.output != nil && w.output.InnerLabel() != nil {
		w.output.InnerLabel().SetText(text)
	}
}

/******************************************************************************/
/* editor_menu_bar_handler.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_overlay/confirm_prompt"
	"kaijuengine.com/editor/editor_overlay/input_prompt"
	"kaijuengine.com/editor/editor_plugin"
	"kaijuengine.com/editor/editor_workspace/settings_workspace"
	"kaijuengine.com/editor/editor_workspace/stage_workspace"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

// WorkspaceSelected switches the active workspace to the one with the given
// id. Called by the menu bar when the user clicks a tab and by plugins via
// editor_workspace.WorkspaceEditorInterface.SelectWorkspace.
func (ed *Editor) WorkspaceSelected(id string) {
	ed.Actions().Run(editor_action.Request{
		ID:     ActionEditorOpenWorkspace,
		Params: editor_action.Params(workspaceActionArgs{Workspace: id}),
		Source: editor_action.SourceMenu,
	})
}

func (ed *Editor) Build(buildMode project.GameBuildMode) {
	ed.SaveCurrentStageWithCallback(func(saved bool) {
		if !saved {
			return
		}
		if !ed.ensureMainStageExists() {
			return
		}
		// goroutine
		go ed.project.CompileGame(buildMode)
		// goroutine
		go func() {
			if err := ed.project.Package(ed.host.AssetDatabase()); err != nil {
				slog.Error("failed to package project", "error", err)
			}
		}()
	})
}

func (ed *Editor) BuildAndRun(buildMode project.GameBuildMode) {
	ed.SaveCurrentStageWithCallback(func(saved bool) {
		if !saved {
			return
		}
		if !ed.ensureMainStageExists() {
			return
		}
		wg := sync.WaitGroup{}
		wg.Add(2)
		// goroutine
		go func() {
			defer wg.Done()
			ed.project.CompileGame(buildMode)
		}()
		// goroutine
		go func() {
			defer wg.Done()
			// Archiving isn't required for debug builds as they don't use
			// the packaged content archive, but we still need to write any
			// generated files like the starting stage id
			if buildMode == project.GameBuildModeDebug {
				ed.project.PackageDebug()
			} else {
				if err := ed.project.Package(ed.host.AssetDatabase()); err != nil {
					slog.Error("failed to package project", "error", err)
				}
			}
		}()
		// goroutine
		go func() {
			wg.Wait()
			ed.project.Run()
		}()
	})
}

func (ed *Editor) BuildAndRunCurrentStage() {
	ed.SaveCurrentStageWithCallback(func(saved bool) {
		if !saved {
			return
		}
		stageId := ed.stageView.Manager().StageId()
		wg := sync.WaitGroup{}
		wg.Add(1)
		// goroutine
		go func() {
			ed.project.CompileDebug()
			wg.Done()
		}()
		// Archiving isn't required for build and run current stage because
		// debug builds don't use the packaged content archive
		// goroutine
		//go func() {
		//	ed.project.Package(ed.host.AssetDatabase())
		//	wg.Done()
		//}()
		// goroutine
		go func() {
			wg.Wait()
			ed.project.Run("-startStage", stageId)
		}()
	})
}

// OpenCodeEditor will run a command specified in CodeEditor settings entry
// directly to the project top level folder.
// This is an exposed function to meet the interface needs of
// [menu_bar.MenuBarHandler].
// func (ed *Editor) OpenVSCodeProject() {
func (ed *Editor) OpenCodeEditor() {
	defer tracing.NewRegion("Editor.OpenCodeEditor").End()
	ed.openCodeEditor(ed.project.FileSystem().FullPath(""))
}

func (ed *Editor) CreateNewStage() {
	if ed.history.HasPendingChanges() {
		ed.BlurInterface()
		confirm_prompt.Show(ed.host, confirm_prompt.Config{
			Title:       "Discard changes",
			Description: "You have unsaved changes to your stage, would you like to discard them and create a new stage?",
			ConfirmText: "Yes",
			CancelText:  "No",
			OnConfirm: func() {
				ed.FocusInterface()
				ed.stageView.Manager().NewStage()
			},
			OnCancel: func() { ed.FocusInterface() },
		})
	} else {
		ed.stageView.Manager().NewStage()
	}
}

// SaveCurrentStage will save the currently open stage file. This is an exposed
// function to meet the interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) SaveCurrentStage() {
	defer tracing.NewRegion("Editor.SaveCurrentStage").End()
	if !ed.history.HasPendingChanges() {
		return
	}
	sm := ed.stageView.Manager()
	if sm.IsNew() {
		ed.BlurInterface()
		input_prompt.Show(ed.host, input_prompt.Config{
			Title:       "Name stage",
			Description: "What would you like to name your stage?",
			Placeholder: "Stage name...",
			Value:       "New Stage",
			ConfirmText: "Save",
			CancelText:  "Cancel",
			OnConfirm: func(name string) {
				ed.FocusInterface()
				ed.saveNewStage(strings.TrimSpace(name))
			},
			OnCancel: func() { ed.FocusInterface() },
		})
	} else {
		ed.saveCurrentStageWithoutNameInput()
	}
}

func (ed *Editor) SaveCurrentStageWithCallback(cb func(bool)) {
	defer tracing.NewRegion("Editor.SaveCurrentStage").End()
	if !ed.history.HasPendingChanges() {
		cb(true)
		return
	}
	sm := ed.stageView.Manager()
	if sm.IsNew() {
		ed.BlurInterface()
		input_prompt.Show(ed.host, input_prompt.Config{
			Title:       "Name stage",
			Description: "What would you like to name your stage?",
			Placeholder: "Stage name...",
			Value:       "New Stage",
			ConfirmText: "Save",
			CancelText:  "Cancel",
			OnConfirm: func(name string) {
				ed.FocusInterface()
				ed.saveNewStage(strings.TrimSpace(name))
				cb(true)
			},
			OnCancel: func() {
				ed.FocusInterface()
				cb(false)
			},
		})
	} else {
		ed.saveCurrentStageWithoutNameInput()
		cb(true)
	}
}

func (ed *Editor) CreateNewCamera() {
	if s, ok := workspaceAs[*stage_workspace.StageWorkspace](ed, stage_workspace.ID); ok {
		s.CreateNewCamera()
	}
}

func (ed *Editor) CreateNewEntity() {
	s, ok := workspaceAs[*stage_workspace.StageWorkspace](ed, stage_workspace.ID)
	if !ok {
		return
	}
	ed.history.BeginTransaction()
	defer ed.history.CommitTransaction()
	e, _ := s.CreateNewEntity()
	m := ed.stageView.Manager()
	m.ClearSelection()
	m.SelectEntity(e)
}

func (ed *Editor) CreateNewLight() {
	if s, ok := workspaceAs[*stage_workspace.StageWorkspace](ed, stage_workspace.ID); ok {
		s.CreateNewLight()
	}
}

func (ed *Editor) CreatePrimitive(primitive rendering.PrimitiveMesh) {
	ed.StageWorkspace().CreatePrimitive(primitive)
}

func (ed *Editor) ConnectSelectedAsDistanceChain() {
	ed.StageWorkspace().ConnectSelectedAsDistanceChain()
}

func (ed *Editor) ConnectSelectedAsRope() {
	ed.StageWorkspace().ConnectSelectedAsRope()
}

func (ed *Editor) ConnectSelectedAsHingeChain() {
	ed.StageWorkspace().ConnectSelectedAsHingeChain()
}

func (ed *Editor) CreatePluginProject(path string) {
	if err := editor_plugin.CreatePluginProject(path); err == nil {
		ed.openCodeEditor(path)
	}
}

func (ed *Editor) CreateHtmlUiFile(name string) {
	tplPath := filepath.Join(project_file_system.ProjectFileTemplates,
		"html_file_template.html.txt")
	data, err := ed.project.FileSystem().ReadFile(tplPath)
	if err != nil {
		slog.Error("failed to read the html template file", "file", tplPath, "error", err)
		return
	}
	t, err := template.New("Html").Parse(string(data))
	if err != nil {
		slog.Error("failed to parse the html template file", "file", tplPath, "error", err)
		return
	}
	sb := strings.Builder{}
	if err = t.Execute(&sb, name); err != nil {
		slog.Error("failed to execute the html template", "file", tplPath, "error", err)
		return
	}
	pfs := ed.ProjectFileSystem()
	cache := ed.Cache()
	ids := content_database.ImportRaw(name, []byte(sb.String()), content_database.Html{}, pfs, cache)
	if len(ids) > 0 {
		ed.events.OnContentAdded.Execute(ids)
		cc, err := cache.Read(ids[0])
		if err != nil {
			exec.Command("code", pfs.FullPath(""), pfs.FullPath(cc.ContentPath())).Run()
		}
	}
}

// SetGridVisible records the developer's preference for the editor viewport
// grid, persists it to the global editor settings, and applies the change to
// the live stage view if one is initialized.
func (ed *Editor) SetGridVisible(visible bool) {
	defer tracing.NewRegion("Editor.SetGridVisible").End()
	if ed.settings.ShowGrid == visible {
		return
	}
	ed.settings.ShowGrid = visible
	if err := ed.settings.Save(); err != nil {
		slog.Error("failed to save editor settings after grid toggle", "error", err)
	}
	ed.stageView.SetGridVisible(visible)
}

func (ed *Editor) CreateCssStylesheetFile(name string) {
	sb := strings.Builder{}
	sb.WriteString("/* ")
	sb.WriteString(name)
	sb.WriteString(" */")
	pfs := ed.ProjectFileSystem()
	cache := ed.Cache()
	ids := content_database.ImportRaw(name, []byte(sb.String()), content_database.Css{}, pfs, cache)
	if len(ids) > 0 {
		ed.events.OnContentAdded.Execute(ids)
		cc, err := cache.Read(ids[0])
		if err != nil {
			exec.Command("code", pfs.FullPath(""), pfs.FullPath(cc.ContentPath())).Run()
		}
	}
}

func (ed *Editor) saveCurrentStageWithoutNameInput() {
	sm := ed.stageView.Manager()
	if err := sm.SaveStage(ed.project.CacheDatabase(), ed.project.FileSystem()); err == nil {
		ed.history.SetSavePosition()
		ed.Project().Settings.EditorSettings.LatestOpenStage = sm.StageId()
		ed.Project().Settings.Save(ed.ProjectFileSystem())
	} else {
		slog.Error("failed to save the current stage", "error", err)
	}
}

func (ed *Editor) ensureMainStageExists() bool {
	if ed.project.Settings.EntryPointStage == "" {
		slog.Error("failed to build, 'main stage' not set in project settings")
		return false
	}
	return true
}

func (ed *Editor) saveNewStage(name string) {
	if name == "" {
		slog.Error("name was blank for the stage, can't save")
		return
	}
	sm := ed.stageView.Manager()
	if err := sm.SetStageId(name, ed.Cache()); err != nil {
		slog.Error("failed to save stage", "error", err)
		return
	}
	ed.saveCurrentStageWithoutNameInput()
	id := sm.StageId()
	ed.events.OnContentAdded.Execute([]string{id})
	// If the entry point stage hasn't yet been created in the
	// settings, assume that this stage will be the one.
	ps := &ed.project.Settings
	if ps.EntryPointStage == "" {
		ps.EntryPointStage = id
		ps.Save(ed.project.FileSystem())
		if s, ok := workspaceAs[*settings_workspace.SettingsWorkspace](ed, settings_workspace.ID); ok {
			s.RequestReload()
		}
	}
}

func (ed *Editor) openCodeEditor(path string) {
	defer tracing.NewRegion("Editor.openCodeEditor").End()
	// TODO:  If this is a file path, the space split won't be enough
	fullArgs := strings.Fields(ed.settings.CodeEditor)
	if len(fullArgs) == 0 {
		slog.Error("failed to launch code editor", "error", "code editor command is empty")
		return
	}
	command := fullArgs[0]
	var args []string
	if len(fullArgs) > 1 {
		args = append(args, fullArgs[1:]...)
	}
	args = append(args, path)
	go func() {
		if err := exec.Command(command, args...).Run(); err != nil {
			slog.Error("failed to launch code editor", "command", command, "path", path, "error", err)
		}
	}()
}

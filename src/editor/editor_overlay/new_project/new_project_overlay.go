/******************************************************************************/
/* new_project_overlay.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package new_project

import (
	"log/slog"
	"path/filepath"

	"kaijuengine.com/editor/editor_overlay/file_browser"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

type NewProject struct {
	doc              *document.Document
	uiMan            ui.Manager
	nameInput        *document.Element
	folder           *document.Element
	templatePathElm  *document.Element
	clearTemplateBtn *document.Element
	errorBox         *document.Element
	loadingOverlay   *document.Element
	config           Config
	templatePath     string
}

type Config struct {
	// OnCreate will be called when the "Create" button is clicked, it will
	// return the name that the developer typed in and the path they selected.
	OnCreate func(name, path, templatePath string, close func())

	// OnOpen will be called when the "Browse" button is clicked, it will return
	// the path that was selected.
	OnOpen func(path string, close func())

	// Error will be used to print out an error to the developer in the window.
	Error string

	// RecentProjects is a list of paths to recent projects.
	RecentProjects []string
}

type OverlayData struct {
	Error          string
	RecentProjects []struct {
		Path  string
		Label string
	}
}

func Show(host *engine.Host, config Config) (*NewProject, error) {
	defer tracing.NewRegion("new_project.Show").End()
	np := &NewProject{config: config}
	np.uiMan.Init(host)
	var err error
	data := OverlayData{Error: config.Error}
	for i := range config.RecentProjects {
		data.RecentProjects = append(data.RecentProjects, struct {
			Path  string
			Label string
		}{
			Path:  config.RecentProjects[i],
			Label: filepath.Base(config.RecentProjects[i]),
		})
	}
	np.doc, err = markup.DocumentFromHTMLAsset(&np.uiMan,
		"editor/ui/overlay/new_project_overlay.go.html",
		data, map[string]func(*document.Element){
			"openProject":       np.openProject,
			"selectTemplate":    np.selectTemplate,
			"clearTemplate":     np.clearTemplate,
			"browse":            np.browse,
			"createProject":     np.createProject,
			"openRecentProject": np.openRecentProject,
			"backgroundClick":   np.backgroundClick,
		})
	if err != nil {
		return np, err
	}
	np.nameInput, _ = np.doc.GetElementById("nameInput")
	np.folder, _ = np.doc.GetElementById("folder")
	np.templatePathElm, _ = np.doc.GetElementById("templatePath")
	np.clearTemplateBtn, _ = np.doc.GetElementById("clearTemplateBtn")
	np.errorBox, _ = np.doc.GetElementById("errorBox")
	np.loadingOverlay, _ = np.doc.GetElementById("loadingOverlay")
	
	if np.nameInput != nil && np.nameInput.UI != nil {
		np.nameInput.UI.ToInput().Focus()
		np.nameInput.UI.AddEvent(ui.EventTypeFocus, func() {
			np.doc.SetElementClasses(np.nameInput, "fullWidth")
			np.hideError()
		})
	}
	
	if np.folder != nil && np.folder.UI != nil {
		np.folder.UI.AddEvent(ui.EventTypeFocus, func() {
			np.doc.SetElementClasses(np.folder, "folderInput")
			np.hideError()
		})
	}
	
	return np, err
}

func (np *NewProject) Close() {
	defer tracing.NewRegion("NewProject.Close").End()
	np.doc.Destroy()
}

func (np *NewProject) openProject(e *document.Element) {
	defer tracing.NewRegion("NewProject.openProject").End()
	np.showFolderPick(true)
}

func (np *NewProject) selectTemplate(e *document.Element) {
	defer tracing.NewRegion("NewProject.openProject").End()
	np.uiMan.DisableUpdate()
	file_browser.Show(np.uiMan.Host, file_browser.Config{
		OnlyFiles: true,
		ExtFilter: []string{".zip"},
		OnConfirm: func(paths []string) {
			np.uiMan.EnableUpdate()
			np.templatePath = paths[0]
			np.templatePathElm.InnerLabel().SetText(np.templatePath)
			np.doc.SetElementClasses(np.clearTemplateBtn, "clearTemplateBtn")
			if np.clearTemplateBtn != nil && np.clearTemplateBtn.UI != nil {
				np.clearTemplateBtn.UI.Show()
			}
		}, OnCancel: np.uiMan.EnableUpdate,
	})
}

func (np *NewProject) clearTemplate(e *document.Element) {
	defer tracing.NewRegion("NewProject.clearTemplate").End()
	np.templatePath = ""
	np.templatePathElm.InnerLabel().SetText("No template selected")
	np.doc.SetElementClasses(np.clearTemplateBtn, "clearTemplateBtn", "hidden")
}

func (np *NewProject) browse(e *document.Element) {
	defer tracing.NewRegion("NewProject.createFolder").End()
	np.showFolderPick(false)
}

func (np *NewProject) showFolderPick(isOpen bool) {
	defer tracing.NewRegion("NewProject.showFolderPick").End()
	np.uiMan.DisableUpdate()
	file_browser.Show(np.uiMan.Host, file_browser.Config{
		OnlyFolders: true,
		OnConfirm: func(paths []string) {
			np.uiMan.EnableUpdate()
			if isOpen {
				np.openProjectFolder(paths[0])
			} else {
				np.folder.UI.ToInput().SetText(paths[0])
			}
		}, OnCancel: np.uiMan.EnableUpdate,
	})
}

func (np *NewProject) createProject(e *document.Element) {
	defer tracing.NewRegion("NewProject.createProject").End()

	np.doc.SetElementClasses(np.nameInput, "fullWidth")
	np.doc.SetElementClasses(np.folder, "folderInput")

	name := np.nameInput.UI.ToInput().Text()
	folder := np.folder.UI.ToInput().Text()

	if name == "" && folder == "" {
		np.doc.SetElementClasses(np.nameInput, "fullWidth", "invalid")
		np.doc.SetElementClasses(np.folder, "folderInput", "invalid")
		np.showError("Project name and folder cannot be empty.")
		return
	}
	if name == "" {
		np.doc.SetElementClasses(np.nameInput, "fullWidth", "invalid")
		np.showError("Project name was not set.")
		return
	}
	if folder == "" {
		np.doc.SetElementClasses(np.folder, "folderInput", "invalid")
		np.showError("Project folder was not set.")
		return
	}
	if np.loadingOverlay != nil {
		np.doc.SetElementClasses(np.loadingOverlay, "loadingOverlay")
	}
	if np.config.OnCreate == nil {
		slog.Error("nothing bound to OnCreate, doing nothing")
		np.Close()
		return
	}
	np.config.OnCreate(name, folder, np.templatePath, np.Close)
}

func (np *NewProject) showError(msg string) {
	if np.errorBox != nil {
		np.doc.SetElementClasses(np.errorBox, "error")
		np.errorBox.UI.Show()
		if lbl := np.errorBox.InnerLabel(); lbl != nil {
			lbl.SetText(msg)
		}
	}
}

func (np *NewProject) hideError() {
	if np.errorBox != nil {
		np.doc.SetElementClasses(np.errorBox, "error", "hidden")
	}
}

func (np *NewProject) backgroundClick(e *document.Element) {
	if np.nameInput != nil && np.nameInput.UI != nil {
		np.nameInput.UI.ToInput().RemoveFocus()
		np.doc.SetElementClasses(np.nameInput, "fullWidth")
	}
	if np.folder != nil && np.folder.UI != nil {
		np.folder.UI.ToInput().RemoveFocus()
		np.doc.SetElementClasses(np.folder, "folderInput")
	}
	np.hideError()
}

func (np *NewProject) openRecentProject(e *document.Element) {
	defer tracing.NewRegion("NewProject.openRecentProject").End()
	np.uiMan.Host.RunOnMainThread(func() {
		np.openProjectFolder(e.Attribute("data-path"))
	})
}

func (np *NewProject) openProjectFolder(path string) {
	if np.loadingOverlay != nil {
		np.doc.SetElementClasses(np.loadingOverlay, "loadingOverlay")
	}
	if np.config.OnOpen == nil {
		slog.Error("nothing bound to OnOpen, doing nothing")
		np.Close()
		return
	}
	np.config.OnOpen(path, np.Close)
}

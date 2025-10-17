package new_project

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/games/editor/editor_overlay/file_browser"
	"kaiju/platform/profiler/tracing"
	"log/slog"
)

type NewProject struct {
	doc       *document.Document
	uiMan     ui.Manager
	nameInput *document.Element
	folder    *document.Element
	config    Config
}

type Config struct {
	OnCreate func(name, path string)
	OnOpen   func(string)
}

func Show(host *engine.Host, config Config) (*NewProject, error) {
	defer tracing.NewRegion("new_project.Show").End()
	np := &NewProject{config: config}
	np.uiMan.Init(host)
	var err error
	np.doc, err = markup.DocumentFromHTMLAsset(&np.uiMan,
		"editor/ui/overlay/new_project_overlay.go.html",
		nil, map[string]func(*document.Element){
			"openProject":   np.openProject,
			"browse":        np.browse,
			"createProject": np.createProject,
		})
	if err != nil {
		return np, err
	}
	np.nameInput, _ = np.doc.GetElementById("nameInput")
	np.folder, _ = np.doc.GetElementById("folder")
	return np, err
}

func (np *NewProject) Close() { np.doc.Destroy() }

func (np *NewProject) openProject(e *document.Element) {
	defer tracing.NewRegion("NewProject.openProject").End()
	np.showFolderPick(true)
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
				np.Close()
				if np.config.OnOpen == nil {
					slog.Error("nothing bound to OnOpen, doing nothing")
					return
				}
				np.config.OnOpen(paths[0])
			} else {
				np.folder.UI.ToInput().SetText(paths[0])
			}
		}, OnCancel: func() {
			np.uiMan.EnableUpdate()
		},
	})
}

func (np *NewProject) createProject(e *document.Element) {
	defer tracing.NewRegion("NewProject.createProject").End()
	name := np.nameInput.UI.ToInput().Text()
	path := np.folder.UI.ToInput().Text()
	np.Close()
	if np.config.OnCreate == nil {
		slog.Error("nothing bound to OnCreate, doing nothing")
		return
	}
	np.config.OnCreate(name, path)
}

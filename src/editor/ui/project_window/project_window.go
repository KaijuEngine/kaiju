package project_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/ui/files_window"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
)

type windowData struct {
	ExistingProjects []string
}

type ProjectWindow struct {
	doc       *document.Document
	container *host_container.Container
	Done      chan struct{}
	data      windowData
}

func (p *ProjectWindow) newProject(elm *document.DocElement) {
	path := <-files_window.Folder("Select Project Folder")
	if path != "" {
		// TODO:  Check if folder has an existing project in it,
		// if so, then open the project and add to project cache

		// TODO:  Check if the folder is empty, if so, then
		// create a new project in the folder and add to project cache

		// TODO:  Check if the folder has files in it, if so, then
		// tell the person that they can not create a project there

		println(path)
		p.container.Close()
	}
}

func (p *ProjectWindow) selectProject(elm *document.DocElement) {
	projectPath := elm.HTML.Attribute("data-project")
	println("Selecting project", projectPath)
	p.container.Close()
}

func (p *ProjectWindow) setupUI() {
	html := klib.MustReturn(p.container.Host.AssetDatabase().ReadText("ui/editor/project.html"))
	p.doc = markup.DocumentFromHTMLString(p.container.Host, html, "", p.data,
		map[string]func(*document.DocElement){
			"newProject":    p.newProject,
			"selectProject": p.selectProject,
		})
}

func New() (*ProjectWindow, error) {
	p := &ProjectWindow{
		Done: make(chan struct{}),
	}
	p.container = host_container.New("Project Window")
	go p.container.Run(600, 400)
	var err error
	p.data.ExistingProjects, err = editor_cache.ListProjects()
	if err != nil {
		return nil, err
	}
	p.container.Host.OnClose.Add(func() {
		p.Done <- struct{}{}
	})
	<-p.container.PrepLock
	p.container.RunFunction(p.setupUI)
	return p, nil
}

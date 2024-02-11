/*****************************************************************************/
/* project_window.go                                                         */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md)    */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining     */
/* a copy of this software and associated documentation files (the           */
/* "Software"), to deal in the Software without restriction, including       */
/* without limitation the rights to use, copy, modify, merge, publish,       */
/* distribute, sublicense, and/or sell copies of the Software, and to        */
/* permit persons to whom the Software is furnished to do so, subject to     */
/* the following conditions:                                                 */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,           */
/* EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF        */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY      */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,      */
/* TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE         */
/* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                    */
/*****************************************************************************/

package project_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/ui/files_window"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"os"
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
		dir, err := os.ReadDir(path)
		if err != nil {
			println("Error reading directory", err)
			return
		}
		if len(dir) == 0 {
			// TODO:  Create a new project in the folder
			println("Creating a new project in", path)
		} else {
			// TODO:  Check if folder has an existing project in it,
			// if so, then open the project and add to project cache

			// TODO:  Check if the folder is empty, if so, then
			// create a new project in the folder and add to project cache
		}
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

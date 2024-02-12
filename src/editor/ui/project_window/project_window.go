/*****************************************************************************/
/* project_window.go                                                         */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package project_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/project"
	"kaiju/editor/ui/files_window"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"os"
)

type windowData struct {
	ExistingProjects []string
	Error            string
}

type ProjectWindow struct {
	doc       *document.Document
	container *host_container.Container
	Selected  chan string
	data      windowData
	picked    bool
}

func (p *ProjectWindow) newProject(elm *document.DocElement) {
	path := <-files_window.Folder("Select Project Folder")
	if path != "" {
		dir, err := os.ReadDir(path)
		if err != nil {
			p.data.Error = "Error reading directory, check permissions and try again"
		} else if len(dir) == 0 {
			if err := project.CreateNew(path); err != nil {
				p.data.Error = "Error creating project: " + err.Error()
			} else {
				p.picked = true
			}
		} else if project.IsProjectDirectory(path) {
			p.picked = true
		} else {
			p.data.Error = "Selected folder (" + path + ") is not a Kaiju project"
		}
	}
	if p.picked {
		p.pick(path)
	} else {
		p.load()
	}
}

func (p *ProjectWindow) pick(path string) {
	p.Selected <- path
	p.picked = true
	p.container.Close()
}

func (p *ProjectWindow) selectProject(elm *document.DocElement) {
	p.pick(elm.HTML.Attribute("data-project"))
}

func (p *ProjectWindow) load() {
	for _, e := range p.container.Host.Entities() {
		e.Destroy()
	}
	html := klib.MustReturn(p.container.Host.AssetDatabase().ReadText("ui/editor/project.html"))
	p.doc = markup.DocumentFromHTMLString(p.container.Host, html, "", p.data,
		map[string]func(*document.DocElement){
			"newProject":    p.newProject,
			"selectProject": p.selectProject,
		})
}

func New() (*ProjectWindow, error) {
	p := &ProjectWindow{
		Selected: make(chan string),
	}
	p.container = host_container.New("Project Window")
	go p.container.Run(600, 400)
	var err error
	p.data.ExistingProjects, err = editor_cache.ListProjects()
	if err != nil {
		return nil, err
	}
	p.container.Host.OnClose.Add(func() {
		if !p.picked {
			p.Selected <- ""
		}
		close(p.Selected)
	})
	<-p.container.PrepLock
	p.container.RunFunction(p.load)
	return p, nil
}

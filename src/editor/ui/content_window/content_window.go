/******************************************************************************/
/* content_window.go                                                          */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package content_window

import (
	"io/fs"
	"kaiju/assets/asset_info"
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/content/content_opener"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/editor_window"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const contentPath = "content"

type contentEntry struct {
	Path  string
	Name  string
	IsDir bool
}

type ContentWindow struct {
	doc       *document.Document
	input     *ui.Input
	listing   *ui.Panel
	editor    interfaces.Editor
	container *host_container.Container
	Dir       []contentEntry
	path      string
	Query     string
	funcMap   map[string]func(*document.DocElement)
	opener    *content_opener.Opener
	selected  *ui.Panel
}

func (s *ContentWindow) Closed()      {}
func (s *ContentWindow) IsRoot() bool { return s.path == contentPath }
func (s *ContentWindow) Tag() string  { return editor_cache.ContentWindow }

func (s *ContentWindow) Container() *host_container.Container {
	return s.container
}

func New(opener *content_opener.Opener, editor interfaces.Editor) {
	s := &ContentWindow{
		funcMap: make(map[string]func(*document.DocElement)),
		opener:  opener,
		path:    contentPath,
		editor:  editor,
	}
	s.funcMap["openContent"] = s.openContent
	s.funcMap["contentClick"] = s.contentClick
	s.container = host_container.New("Content Browser", nil)
	x, y := editor.Host().Window.Center()
	editor_window.OpenWindow(s, 500, 300, x-250, y-150)
	editor.WindowListing().Add(s)
}

func (s *ContentWindow) Init() {
	s.reloadUI()
}

func (s *ContentWindow) contentClick(elm *document.DocElement) {
	if elm.UIPanel == s.selected {
		s.openContent(elm)
		return
	}
	for i := range elm.HTML.Parent.Children {
		p := elm.HTML.Parent.Children[i].DocumentElement.UIPanel
		p.UnEnforceColor()
		lbl := ui.FirstOnEntity(p.Entity().Children[1].Children[0]).(*ui.Label)
		lbl.UnEnforceBGColor()
	}
	elm.UIPanel.EnforceColor(matrix.ColorDarkBlue())
	lbl := ui.FirstOnEntity(elm.UI.Entity().Children[1].Children[0]).(*ui.Label)
	lbl.EnforceBGColor(matrix.ColorDarkBlue())
	s.selected = elm.UIPanel
}

func (s *ContentWindow) openContent(elm *document.DocElement) {
	path := elm.HTML.Attribute("data-path")
	if path == "../" {
		if s.path == contentPath {
			return
		} else {
			s.path = filepath.Clean(s.path + "/" + path)
		}
	} else {
		s.path = path
	}
	if info, err := os.Stat(s.path); err != nil {
		slog.Error(err.Error())
		s.container.Host.Close()
		return
	} else if info.IsDir() {
		s.reloadUI()
	} else {
		if err := s.opener.OpenPath(s.path, s.editor); err != nil {
			slog.Error(err.Error())
		}
	}
}

func (s *ContentWindow) submit() {
	s.Query = strings.ToLower(strings.TrimSpace(s.input.Text()))
	if s.Query == "" {
		s.path = contentPath
	}
	s.reloadUI()
}

func (s *ContentWindow) reloadUI() {
	for _, e := range s.container.Host.Entities() {
		e.Destroy()
	}
	s.list()
	html := klib.MustReturn(s.container.Host.AssetDatabase().ReadText("editor/ui/content_window.html"))
	s.doc = markup.DocumentFromHTMLString(
		s.container.Host, html, "", s, s.funcMap)
	if elm, ok := s.doc.GetElementById("searchInput"); !ok {
		slog.Error(`Failed to locate the "searchInput" for the content window`)
		s.container.Host.Close()
		return
	} else {
		s.input = elm.UI.(*ui.Input)
		s.input.Data().OnSubmit.Add(s.submit)
	}
	if elm, ok := s.doc.GetElementById("listing"); !ok {
		slog.Error(`Failed to locate the "listing" for the content window`)
		s.container.Host.Close()
		return
	} else {
		s.listing = elm.UIPanel
	}
}

func (s *ContentWindow) listSearch() {
	s.Dir = s.Dir[:0]
	filepath.Walk("content", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			slog.Error(err.Error())
			return nil
		}
		if filepath.Ext(info.Name()) == asset_info.InfoExtension {
			return nil
		}
		name := strings.ToLower(info.Name())
		if strings.Contains(name, s.Query) {
			s.Dir = append(s.Dir, contentEntry{
				Path:  path,
				Name:  info.Name(),
				IsDir: info.IsDir(),
			})
		}
		return nil
	})
}

func (s *ContentWindow) listAll() {
	dir, err := os.ReadDir(s.path)
	if err != nil {
		slog.Error(err.Error())
		s.container.Host.Close()
		return
	}
	s.Dir = make([]contentEntry, 0, len(dir))
	for i := range dir {
		if filepath.Ext(dir[i].Name()) != asset_info.InfoExtension {
			s.Dir = append(s.Dir, contentEntry{
				Path:  filepath.Join(s.path, dir[i].Name()),
				Name:  dir[i].Name(),
				IsDir: dir[i].IsDir(),
			})
		}
	}
}

func (s *ContentWindow) list() {
	if s.Query != "" {
		s.listSearch()
	} else {
		s.listAll()
	}
}

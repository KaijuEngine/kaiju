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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
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
	"kaiju/editor/ui/context_menu"
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

const (
	contentPath = "content"
	sizeConfig  = "contentWindowSize"
)

type contentEntry struct {
	Path     string
	Name     string
	Children []contentEntry
	IsDir    bool
}

func (c contentEntry) Depth() int {
	return strings.Count(c.Path, "/") + strings.Count(c.Path, "\\")
}

type ContentWindow struct {
	doc      *document.Document
	input    *ui.Input
	listing  *ui.Panel
	editor   interfaces.Editor
	DirTree  []contentEntry
	Dir      []contentEntry
	path     string
	Query    string
	funcMap  map[string]func(*document.Element)
	opener   *content_opener.Opener
	selected *ui.Panel
	uiGroup  *ui.Group
}

func (s *ContentWindow) IsRoot() bool { return s.path == contentPath }

func New(opener *content_opener.Opener, editor interfaces.Editor, uiGroup *ui.Group) *ContentWindow {
	s := &ContentWindow{
		funcMap: make(map[string]func(*document.Element)),
		opener:  opener,
		path:    contentPath,
		editor:  editor,
		uiGroup: uiGroup,
	}
	s.funcMap["openContent"] = s.openContent
	s.funcMap["contentClick"] = s.contentClick
	s.funcMap["contentDblClick"] = s.contentDblClick
	s.funcMap["resizeHover"] = s.resizeHover
	s.funcMap["resizeExit"] = s.resizeExit
	s.funcMap["resizeStart"] = s.resizeStart
	s.funcMap["resizeStop"] = s.resizeStop
	s.funcMap["entryCtxMenu"] = s.entryCtxMenu
	s.funcMap["updateSearch"] = s.updateSearch
	editor.Host().OnClose.Add(func() {
		if s.doc != nil {
			s.doc.Destroy()
		}
	})
	return s
}

func (c *ContentWindow) Toggle() {
	if c.doc == nil {
		c.Show()
	} else {
		if c.doc.Elements[0].UI.Entity().IsActive() {
			c.Hide()
		} else {
			c.Show()
		}
	}
}

func (c *ContentWindow) Show() {
	if c.doc == nil {
		c.reloadUI()
	} else {
		c.doc.Activate()
	}
}

func (c *ContentWindow) Hide() {
	if c.doc != nil {
		c.doc.Deactivate()
	}
}

func (s *ContentWindow) contentDblClick(elm *document.Element) {
	s.openContent(elm)
}

func (s *ContentWindow) contentClick(elm *document.Element) {
	for i := range elm.Parent.Children {
		p := elm.Parent.Children[i].UIPanel
		p.UnEnforceColor()
		c := p.Base().Entity().Children
		lbl := ui.FirstOnEntity(c[len(c)-1].Children[0]).(*ui.Label)
		lbl.UnEnforceBGColor()
	}
	elm.UIPanel.EnforceColor(matrix.ColorDarkBlue())
	c := elm.UI.Entity().Children
	lbl := ui.FirstOnEntity(c[len(c)-1].Children[0]).(*ui.Label)
	lbl.EnforceBGColor(matrix.ColorDarkBlue())
	s.selected = elm.UIPanel
}

func (s *ContentWindow) openContent(elm *document.Element) {
	path := elm.Attribute("data-path")
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
		return
	} else if info.IsDir() {
		s.reloadUI()
	} else {
		if err := s.opener.OpenPath(s.path, s.editor); err != nil {
			slog.Error(err.Error())
		}
	}
}

func (s *ContentWindow) reloadUI() {
	const html = "editor/ui/content_window.html"
	folderPanelScroll := float32(0)
	if s.doc != nil {
		if fp, ok := s.doc.GetElementById("folderListing"); ok {
			folderPanelScroll = fp.UIPanel.ScrollY()
		}
		s.doc.Destroy()
	}
	s.list()
	host := s.editor.Host()
	host.CreatingEditorEntities()
	s.doc = klib.MustReturn(markup.DocumentFromHTMLAsset(host, html, s, s.funcMap))
	s.doc.SetGroup(s.uiGroup)
	host.DoneCreatingEditorEntities()
	if elm, ok := s.doc.GetElementById("searchInput"); !ok {
		slog.Error(`Failed to locate the "searchInput" for the content window`)
	} else {
		s.input = elm.UI.(*ui.UIBase).AsInput()
	}
	if elm, ok := s.doc.GetElementById("listing"); !ok {
		slog.Error(`Failed to locate the "listing" for the content window`)
	} else {
		s.listing = elm.UIPanel
	}
	s.doc.Clean()
	if h, ok := editor_cache.EditorConfigValue(sizeConfig); ok {
		w, _ := s.doc.GetElementById("window")
		w.UIPanel.Base().Layout().ScaleHeight(matrix.Float(h.(float64)))
	}
	if fp, ok := s.doc.GetElementById("folderListing"); ok {
		fp.UIPanel.SetScrollY(folderPanelScroll)
	}
	s.input.Focus()
}

func (s *ContentWindow) listSearch() {
	s.Dir = s.Dir[:0]
	filepath.Walk(contentPath, func(path string, info fs.FileInfo, err error) error {
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
	s.DirTree = s.DirTree[:0]
	parentMap := map[string]*contentEntry{}
	filepath.Walk(contentPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			self := contentEntry{
				Path:     path,
				Name:     info.Name(),
				IsDir:    true,
				Children: make([]contentEntry, 0),
			}
			parent := filepath.Dir(path)
			if parent == "." {
				s.DirTree = append(s.DirTree, self)
				parentMap[path] = &s.DirTree[len(s.DirTree)-1]
			} else {
				p := parentMap[parent]
				p.Children = append(p.Children, self)
				parentMap[path] = &p.Children[len(p.Children)-1]
			}
		}
		return nil
	})
}

func (s *ContentWindow) resizeHover(e *document.Element) {
	s.editor.Host().Window.CursorSizeNS()
}

func (s *ContentWindow) resizeExit(e *document.Element) {
	dd := s.editor.Host().Window.Mouse.DragData()
	if dd != s {
		s.editor.Host().Window.CursorStandard()
	}
}

func (l *ContentWindow) resizeStart(e *document.Element) {
	l.editor.Host().Window.CursorSizeNS()
	l.editor.Host().Window.Mouse.SetDragData(l)
}

func (s *ContentWindow) resizeStop(e *document.Element) {
	dd := s.editor.Host().Window.Mouse.DragData()
	if dd != s {
		return
	}
	s.editor.Host().Window.CursorStandard()
	w, _ := s.doc.GetElementById("window")
	h := w.UIPanel.Base().Layout().PixelSize().Height()
	editor_cache.SetEditorConfigValue(sizeConfig, h)
}

func (s *ContentWindow) entryCtxMenu(elm *document.Element) {
	path := elm.Attribute("data-path")
	ctx := []context_menu.ContextMenuEntry{
		{Id: "open", Label: "Open", OnClick: func() {
			s.openContent(elm)
		}},
	}
	if content_opener.IsATextFile(path) {
		ctx = append(ctx, context_menu.ContextMenuEntry{
			Id: "edit", Label: "Edit", OnClick: func() {
				content_opener.EditTextFile(path)
			}})
	}
	s.editor.ContextMenu().Show(ctx)
}

func (s *ContentWindow) updateSearch(elm *document.Element) {
	s.Query = strings.ToLower(strings.TrimSpace(s.input.Text()))
	if s.Query == "" {
		s.path = contentPath
	}
	s.reloadUI()
}

func (s *ContentWindow) DragUpdate() {
	w, _ := s.doc.GetElementById("window")
	y := s.editor.Host().Window.Mouse.Position().Y() - 20
	h := s.editor.Host().Window.Height()
	if int(y) < h-100 {
		w.UIPanel.Base().Layout().ScaleHeight(y)
	}
}

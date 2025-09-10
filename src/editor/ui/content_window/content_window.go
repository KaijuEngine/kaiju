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
	"kaiju/editor/alert"
	"kaiju/editor/content/content_opener"
	"kaiju/editor/editor_config"
	"kaiju/editor/editor_interface"
	"kaiju/editor/ui/context_menu"
	"kaiju/editor/ui/drag_datas"
	"kaiju/editor/ui/shader_designer"
	"kaiju/engine/assets/asset_info"
	"kaiju/engine/systems/events"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/filesystem"
	"kaiju/platform/windowing"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type searchType = int

const (
	searchTypeQuery = searchType(iota)
	searchTypeType
	searchTypeTagGroup
)

const (
	contentPath = "content"
	sizeConfig  = "contentWindowSize"
)

type contentEntry struct {
	Path     string
	Name     string
	Children []contentEntry
}

func (c contentEntry) Depth() int {
	return strings.Count(c.Path, "/") + strings.Count(c.Path, "\\")
}

type contentWindowFilter struct {
	Label   string
	Texture string
	Filter  string
}

type ContentWindow struct {
	doc           *document.Document
	input         *ui.Input
	listing       *ui.Panel
	editor        editor_interface.Editor
	DirTree       []contentEntry
	Dir           []contentEntry
	TypeFilters   []contentWindowFilter
	path          string
	SearchText    string
	Query         string
	funcMap       map[string]func(*document.Element)
	opener        *content_opener.Opener
	selected      *ui.Panel
	focusOnReload bool
}

func (s *ContentWindow) TabTitle() string             { return "Content" }
func (s *ContentWindow) Document() *document.Document { return s.doc }

func (s *ContentWindow) Destroy() {
	if s.doc != nil {
		s.focusOnReload = s.input.IsFocused()
		s.doc.Destroy()
		s.doc = nil
	}
}

func (s *ContentWindow) IsRoot() bool { return s.path == contentPath }

func New(opener *content_opener.Opener, editor editor_interface.Editor) *ContentWindow {
	s := &ContentWindow{
		funcMap: make(map[string]func(*document.Element)),
		opener:  opener,
		path:    contentPath,
		editor:  editor,
		// TODO:  Setup the different textures for each filter
		TypeFilters: []contentWindowFilter{
			{Label: "Stage", Texture: "editor/textures/icons/file.png", Filter: "t:stage"},
			{Label: "Mesh", Texture: "editor/textures/icons/file.png", Filter: "t:mesh"},
			{Label: "HTML", Texture: "editor/textures/icons/file.png", Filter: "t:html"},
			{Label: "Image", Texture: "editor/textures/icons/file.png", Filter: "t:image"},
			{Label: "Material", Texture: "editor/textures/icons/file.png", Filter: "t:material"},
			{Label: "Shader", Texture: "editor/textures/icons/file.png", Filter: "t:shader"},
			{Label: "Shader Pipeline", Texture: "editor/textures/icons/file.png", Filter: "t:shaderpipeline"},
			{Label: "Render Pass", Texture: "editor/textures/icons/file.png", Filter: "t:renderpass"},
		},
	}
	s.funcMap["openContent"] = s.openContent
	s.funcMap["contentClick"] = s.contentClick
	s.funcMap["contentDblClick"] = s.contentDblClick
	s.funcMap["entryCtxMenu"] = s.entryCtxMenu
	s.funcMap["updateSearch"] = s.updateSearch
	s.funcMap["entryDragStart"] = s.entryDragStart
	s.funcMap["filterByType"] = s.filterByType
	editor.Host().OnClose.Add(func() {
		if s.doc != nil {
			s.doc.Destroy()
		}
	})
	return s
}

func (s *ContentWindow) contentDblClick(elm *document.Element) {
	s.openContent(elm)
}

func (s *ContentWindow) contentClick(elm *document.Element) {
	path := elm.Attribute("data-path")
	for i := range elm.Parent.Value().Children {
		p := elm.Parent.Value().Children[i].UIPanel
		p.UnEnforceColor()
	}
	elm.UIPanel.EnforceColor(matrix.ColorDarkBlue())
	s.selected = elm.UIPanel
	if stat, err := os.Stat(path + asset_info.InfoExtension); err != nil || stat.IsDir() {
		evt := &s.editor.Events().OnContentSelect
		evt.Content = []string{}
		evt.Event.Execute()
	} else {
		evt := &s.editor.Events().OnContentSelect
		evt.Content = []string{path}
		evt.Event.Execute()
	}
}

func (s *ContentWindow) openContent(elm *document.Element) {
	path := elm.Attribute("data-path")
	if path == "../" {
		if s.path == contentPath {
			return
		} else if info, err := os.Stat(s.path); err == nil && !info.IsDir() {
			s.path = filepath.Clean(filepath.Dir(s.path) + "/" + path)
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
		s.editor.ReloadTabs(s.TabTitle())
	} else {
		if err := s.opener.OpenPath(s.path, s.editor); err != nil {
			slog.Error(err.Error())
		}
	}
}

func (s *ContentWindow) duplicateContent(elm *document.Element) {
	path := elm.Attribute("data-path")
	// TODO:  Shouldn't even show the option on this entry
	if path == "../" {
		return
	}
	if s, err := os.Stat(path); err == nil && s.IsDir() {
		slog.Error("currently, you can't duplicate a directory", "from", path)
		return
	}
	name := <-alert.NewInput("Duplicate name", "New name...", "", "Duplicate", "Cancel", s.editor.Host())
	if name == "" {
		return
	}
	ext := filepath.Ext(path)
	newPath := filepath.Join(filepath.Dir(path), name)
	if filepath.Ext(newPath) != ext {
		newPath += ext
	}
	if _, err := os.Stat(newPath); err == nil {
		slog.Error("failed to duplicate file, a file with that name already exists", "from", path, "to", newPath)
		return
	}
	if err := filesystem.CopyFile(path, newPath); err != nil {
		slog.Error("failed to duplicate the file", "error", err)
	} else {
		s.editor.ImportRegistry().Import(path)
		s.editor.ReloadTabs(s.TabTitle())
	}
}

func (s *ContentWindow) Reload(uiMan *ui.Manager, root *document.Element) {
	const html = "editor/ui/content_window.html"
	folderPanelScroll := float32(0)
	shouldFocus := s.focusOnReload
	if s.doc != nil {
		shouldFocus = s.input.IsFocused()
		if fp, ok := s.doc.GetElementById("folderListing"); ok {
			folderPanelScroll = fp.UIPanel.ScrollY()
		}
		s.doc.Destroy()
	}
	s.list()
	host := s.editor.Host()
	host.CreatingEditorEntities()
	s.doc = klib.MustReturn(markup.DocumentFromHTMLAssetRooted(uiMan, html, s, s.funcMap, root))
	host.DoneCreatingEditorEntities()
	if elm, ok := s.doc.GetElementById("searchInput"); !ok {
		slog.Error(`Failed to locate the "searchInput" for the content window`)
	} else {
		s.input = elm.UI.ToInput()
	}
	if elm, ok := s.doc.GetElementById("listing"); !ok {
		slog.Error(`Failed to locate the "listing" for the content window`)
	} else {
		s.listing = elm.UIPanel
	}
	s.doc.Clean()
	if fp, ok := s.doc.GetElementById("folderListing"); ok {
		fp.UIPanel.SetScrollY(folderPanelScroll)
	}
	if shouldFocus {
		s.input.Focus()
	}
}

func (s *ContentWindow) list() {
	s.Dir = s.Dir[:0]
	if s.Query == "" {
		return
	}
	searchType := searchTypeQuery
	query := s.Query
	if strings.HasPrefix(s.Query, "t:") {
		query = strings.TrimPrefix(s.Query, "t:")
		searchType = searchTypeType
	} else if strings.HasPrefix(s.Query, "g:") {
		query = strings.TrimPrefix(s.Query, "g:")
		searchType = searchTypeTagGroup
	}
	filepath.Walk(contentPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			slog.Error(err.Error())
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(info.Name()) == asset_info.InfoExtension {
			if searchType == searchTypeQuery {
				return nil
			}
			// TODO:  This is very slow, use a in-memory database instead
			adi, err := asset_info.Read(path)
			if err != nil {
				return err
			}
			match := false
			switch searchType {
			case searchTypeType:
				switch query {
				case "mesh":
					match = adi.Type == editor_config.AssetTypeGlb ||
						adi.Type == editor_config.AssetTypeGltf ||
						adi.Type == editor_config.AssetTypeObj
				default:
					match = adi.Type == query
				}
			case searchTypeTagGroup:
				match = slices.Contains(adi.Tags, query)
			}
			if match {
				s.Dir = append(s.Dir, contentEntry{
					Path: adi.Path,
					Name: strings.TrimSuffix(info.Name(), asset_info.InfoExtension),
				})
			}
		} else if query == "" {
			name := strings.ToLower(info.Name())
			if strings.Contains(name, s.Query) {
				s.Dir = append(s.Dir, contentEntry{
					Path: path,
					Name: info.Name(),
				})
			}
		}
		return nil
	})
}

func (s *ContentWindow) entryCtxMenu(elm *document.Element) {
	path := elm.Attribute("data-path")
	ctx := []context_menu.ContextMenuEntry{
		{Id: "open", Label: "Open", OnClick: func() { s.openContent(elm) }},
	}
	if f, err := os.Stat(path); err != nil {
		if !f.IsDir() {
			ctx = append(ctx, context_menu.NewEntry("duplicate", "Duplicate", func() {
				s.duplicateContent(elm)
			}))
		}
	}
	if content_opener.IsATextFile(path) {
		ctx = append(ctx, context_menu.NewEntry("edit", "Edit", func() {
			content_opener.EditTextFile(path)
		}))
	}
	if filepath.Ext(path) == editor_config.FileExtensionMaterial {
		ctx = append(ctx, context_menu.NewEntry("preview", "Preview", func() {
			shader_designer.PreviewMaterial(path)
		}))
	}
	s.editor.ContextMenu().Show(ctx)
}

func (s *ContentWindow) updateSearch(elm *document.Element) {
	s.SearchText = s.input.Text()
	s.Query = strings.ToLower(strings.TrimSpace(s.SearchText))
	if s.Query == "" {
		s.path = contentPath
	}
	s.editor.ReloadTabs(s.TabTitle())
}

func (s *ContentWindow) entryDragStart(elm *document.Element) {
	path := elm.Attribute("data-path")
	host := s.editor.Host()
	host.Window.CursorSizeAll()
	windowing.SetDragData(&drag_datas.FileIdDragData{path})
	elm.EnforceColor(matrix.ColorPurple())
	var eid events.Id
	eid = windowing.OnDragStop.Add(func() {
		if s.editor.IsMouseOverViewport() {
			s.openContent(elm)
		}
		host.Window.CursorStandard()
		windowing.OnDragStop.Remove(eid)
		elm.UnEnforceColor()
	})
}

func (s *ContentWindow) filterByType(elm *document.Element) {
	filter := elm.Attribute("data-filter")
	s.input.SetText(filter)
}

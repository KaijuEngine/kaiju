/******************************************************************************/
/* content_workspace.go                                                       */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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

package content_workspace

import (
	"kaiju/editor/editor_overlay/file_browser"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/rendering"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
)

type Workspace struct {
	common_workspace.CommonWorkspace
	pfs               *project_file_system.FileSystem
	cCache            *content_database.Cache
	typeFilters       []string
	tagFilters        []string
	query             string
	entryTemplate     *document.Element
	tagFilterTemplate *document.Element
	addTagbtn         *document.Element
	selectedContent   *document.Element
	rightBody         *document.Element
	pageData          WorkspaceUIData
	info              struct {
		nameInput        *document.Element
		tagList          *document.Element
		entryTagTemplate *document.Element
		newTagInput      *document.Element
		newTagHint       *document.Element
		tagHintTemplate  *document.Element
	}
}

type WorkspaceUIData struct {
	Filters []string
	Tags    []string
}

func (w *Workspace) Initialize(host *engine.Host, pfs *project_file_system.FileSystem, cdb *content_database.Cache) {
	w.pfs = pfs
	w.cCache = cdb
	for _, cat := range content_database.ContentCategories {
		w.pageData.Filters = append(w.pageData.Filters, cat.TypeName())
	}
	list := w.cCache.List()
	ids := make([]string, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].Id())
		for j := range list[i].Config.Tags {
			w.pageData.Tags = klib.AppendUnique(w.pageData.Tags, list[i].Config.Tags[j])
		}
	}
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/content_workspace.go.html", w.pageData, map[string]func(*document.Element){
			"inputFilter":    w.inputFilter,
			"tagFilter":      w.tagFilter,
			"clickImport":    w.clickImport,
			"clickFilter":    w.clickFilter,
			"clickEntry":     w.clickEntry,
			"clickDeleteTag": w.clickDeleteTag,
			"updateTagHint":  w.updateTagHint,
			"submitNewTag":   w.submitNewTag,
			"clickTagHint":   w.clickTagHint,
			"submitName":     w.submitName,
		})
	w.entryTemplate, _ = w.Doc.GetElementById("entryTemplate")
	w.tagFilterTemplate, _ = w.Doc.GetElementById("tagFilterTemplate")
	w.info.entryTagTemplate, _ = w.Doc.GetElementById("entryTagTemplate")
	w.addTagbtn, _ = w.Doc.GetElementById("addTagbtn")
	w.rightBody, _ = w.Doc.GetElementById("rightBody")
	w.info.nameInput, _ = w.Doc.GetElementById("entryName")
	w.info.tagList, _ = w.Doc.GetElementById("entryTags")
	w.info.newTagInput, _ = w.Doc.GetElementById("newTagInput")
	w.info.newTagHint, _ = w.Doc.GetElementById("newTagHint")
	w.info.tagHintTemplate, _ = w.Doc.GetElementById("tagHintTemplate")
	w.addContent(ids)
}

func (w *Workspace) Open() {
	w.CommonOpen()
	w.entryTemplate.UI.Hide()
	w.tagFilterTemplate.UI.Hide()
	w.info.entryTagTemplate.UI.Hide()
	w.info.tagHintTemplate.UI.Hide()
	w.info.newTagHint.UI.Hide()
	if w.selectedContent == nil {
		w.rightBody.UI.Hide()
	}
	w.Doc.Clean()
}

func (w *Workspace) Close() { w.CommonClose() }

func (w *Workspace) clickImport(*document.Element) {
	w.UiMan.DisableUpdate()
	file_browser.Show(w.Host, file_browser.Config{
		ExtFilter:   content_database.ImportableTypes,
		MultiSelect: true,
		OnConfirm: func(paths []string) {
			w.UiMan.EnableUpdate()
			index := []string{}
			for i := range paths {
				res, err := content_database.Import(paths[i], w.pfs, w.cCache, "")
				if err != nil {
					slog.Error("failed to import content", "path", paths[i], "error", err)
				} else {
					var addDependencies func(target *content_database.ImportResult)
					addDependencies = func(target *content_database.ImportResult) {
						index = append(index, target.Id)
						for i := range res.Dependencies {
							addDependencies(&target.Dependencies[i])
						}
					}
					addDependencies(&res)
				}
			}
			w.addContent(index)
		}, OnCancel: func() {
			w.UiMan.EnableUpdate()
		},
	})
}

func (w *Workspace) addContent(ids []string) {
	if len(ids) == 0 {
		return
	}
	ccAll := make([]content_database.CachedContent, 0, len(ids))
	for i := range ids {
		cc, err := w.cCache.Read(ids[i])
		if err != nil {
			slog.Error("failed to read the cached content", "id", ids[i], "error", err)
			continue
		}
		ccAll = append(ccAll, cc)
	}
	cpys := w.Doc.DuplicateElementRepeat(w.entryTemplate, len(ccAll))
	for i := range cpys {
		cc := &ccAll[i]
		cpys[i].SetAttribute("id", cc.Id())
		cpys[i].SetAttribute("data-type", strings.ToLower(cc.Config.Type))
		lbl := cpys[i].Children[1].Children[0].UI.ToLabel()
		lbl.SetText(cc.Config.Name)
		img := cpys[i].Children[0].UI.ToPanel()
		if cc.Config.Type == (content_database.Texture{}).TypeName() {
			// Loose goroutine
			go func() {
				path := content_database.ToContentPath(cc.Path)
				key := filepath.Base(path)
				data, err := w.pfs.ReadFile(path)
				if err != nil {
					slog.Error("error reading the image file", "path", path)
					return
				}
				td := rendering.ReadRawTextureData(data, rendering.TextureFileFormatPng)
				tex, err := w.Host.TextureCache().InsertTexture(key, td.Mem,
					td.Width, td.Height, rendering.TextureFilterLinear)
				if err != nil {
					slog.Error("failed to insert the texture to the cache", "error", err)
					return
				}
				img.SetBackground(tex)
			}()
		}
	}
}

func (w *Workspace) inputFilter(e *document.Element) {
	w.query = strings.ToLower(e.UI.ToInput().Text())
	// TODO:  Regex out the filters like tag:..., type:..., etc.
	w.runFilter()
}

func (w *Workspace) tagFilter(e *document.Element) {
	q := strings.ToLower(e.UI.ToInput().Text())
	tagElms := w.Doc.GetElementsByGroup("tag")[1:]
	for i := range tagElms {
		tag := tagElms[i].Attribute("data-tag")
		show := strings.Contains(strings.ToLower(tag), q)
		if show {
			tagElms[i].UI.Show()
		} else {
			tagElms[i].UI.Hide()
		}
	}
}

func (w *Workspace) clickFilter(e *document.Element) {
	isSelected := slices.Contains(e.ClassList(), "filterSelected")
	isSelected = !isSelected
	typeName := e.Attribute("data-type")
	tagName := e.Attribute("data-tag")
	if isSelected {
		w.Doc.SetElementClasses(e, "leftBtn", "filterSelected")
		if typeName != "" {
			w.typeFilters = append(w.typeFilters, typeName)
		}
		if tagName != "" {
			w.tagFilters = append(w.tagFilters, tagName)
		}
	} else {
		w.Doc.SetElementClasses(e, "leftBtn")
		if typeName != "" {
			w.typeFilters = klib.SlicesRemoveElement(w.typeFilters, typeName)
		}
		if tagName != "" {
			w.tagFilters = klib.SlicesRemoveElement(w.tagFilters, tagName)
		}
	}
	w.runFilter()
}

func (w *Workspace) clickEntry(e *document.Element) {
	id := e.Attribute("id")
	cc, err := w.cCache.Read(id)
	if err != nil {
		slog.Error("failed to find the config for the selected entry", "id", id, "error", err)
		return
	}
	if w.selectedContent != nil {
		w.Doc.SetElementClasses(w.selectedContent, "entry")
	}
	w.selectedContent = e
	w.rightBody.UI.Show()
	w.Doc.SetElementClasses(w.selectedContent, "entry", "entrySelected")
	for i := len(w.info.tagList.Children) - 1; i >= 1; i-- {
		w.Doc.RemoveElement(w.info.tagList.Children[i])
	}
	w.info.nameInput.UI.ToInput().SetText(cc.Config.Name)
	cpys := w.Doc.DuplicateElementRepeat(w.info.entryTagTemplate, len(cc.Config.Tags))
	for i := range cpys {
		cpys[i].Children[0].Children[0].UI.ToLabel().SetText(cc.Config.Tags[i])
		cpys[i].Children[1].SetAttribute("data-tag", cc.Config.Tags[i])
		cpys[i].UI.Show()
	}
}

func (w *Workspace) clickDeleteTag(e *document.Element) {
	id := w.selectedId()
	cc, err := w.cCache.Read(id)
	if err != nil {
		slog.Error("failed to find the config to delete tag from content", "id", id, "error", err)
		return
	}
	tag := e.Attribute("data-tag")
	if cc.Config.RemoveTag(tag) {
		w.updateIndexForCachedContent(&cc)
	} else {
		slog.Error("failed to locate the tag that was expected to exist", "tag", tag)
	}
	w.Doc.RemoveElement(e.Parent.Value())
}

func (w *Workspace) updateTagHint(e *document.Element) {
	q := strings.ToLower(e.UI.ToInput().Text())
	for i := len(w.info.newTagHint.Children) - 1; i >= 1; i-- {
		w.Doc.RemoveElement(w.info.newTagHint.Children[i])
	}
	if q == "" {
		w.info.newTagHint.UI.Hide()
		return
	}
	options := make([]string, 0, len(w.pageData.Tags))
	for i := range w.pageData.Tags {
		if strings.Contains(strings.ToLower(w.pageData.Tags[i]), q) {
			options = append(options, w.pageData.Tags[i])
		}
	}
	if len(options) == 0 {
		w.info.newTagHint.UI.Hide()
		return
	}
	cpys := w.Doc.DuplicateElementRepeat(w.info.tagHintTemplate, len(options))
	for i := range cpys {
		cpys[i].Children[0].UI.ToLabel().SetText(options[i])
		cpys[i].UI.Show()
	}
	w.info.newTagHint.UI.Show()
}

func (w *Workspace) submitNewTag(e *document.Element) {
	input := e.UI.ToInput()
	txt := input.Text()
	if strings.TrimSpace(txt) == "" {
		return
	}
	w.addTagToSelected(txt)
	input.SetTextWithoutEvent("")
}

func (w *Workspace) clickTagHint(e *document.Element) {
	w.addTagToSelected(e.Children[0].UI.ToLabel().Text())
	w.info.newTagHint.UI.Hide()
	w.info.newTagInput.UI.ToInput().SetTextWithoutEvent("")
}

func (w *Workspace) submitName(e *document.Element) {
	name := strings.TrimSpace(e.UI.ToInput().Text())
	if name == "" {
		slog.Warn("The name for the content can't be left blank, ignoring change")
		return
	}
	id := w.selectedId()
	cc, err := w.cCache.Read(id)
	if err != nil {
		slog.Error("failed to find the content by id", "id", id, "error", err)
		return
	}
	cc.Config.Name = name
	if err := content_database.WriteConfig(cc.Path, cc.Config, w.pfs); err != nil {
		slog.Error("failed to update the content config file", "id", id, "error", err)
		return
	}
	w.selectedContent.Children[1].Children[0].UI.ToLabel().SetText(name)
}

func (w *Workspace) addTagToSelected(tag string) {
	id := w.selectedId()
	cc, err := w.cCache.Read(id)
	if err != nil {
		slog.Error("failed to find the config to add tag to content", "id", id, "error", err)
		return
	}
	var ok bool
	if tag, ok = cc.Config.AddTag(tag); ok {
		w.updateIndexForCachedContent(&cc)
	}
	w.clickEntry(w.selectedContent)
	for i := range w.pageData.Tags {
		if strings.EqualFold(tag, w.pageData.Tags[i]) {
			return
		}
	}
	w.pageData.Tags = append(w.pageData.Tags, tag)
	elm := w.Doc.DuplicateElement(w.tagFilterTemplate)
	elm.Children[0].UI.ToLabel().SetText(tag)
	elm.SetAttribute("data-tag", tag)
}

func (w *Workspace) selectedId() string {
	if w.selectedContent != nil {
		return w.selectedContent.Attribute("id")
	}
	return ""
}

func (w *Workspace) runFilter() {
	entries := w.Doc.GetElementsByGroup("entry")
	for i := range entries {
		e := entries[i]
		id := e.Attribute("id")
		if id == "entryTemplate" {
			continue
		}
		show := len(w.typeFilters) == 0 && len(w.tagFilters) == 0
		if !show && len(w.typeFilters) > 0 {
			show = slices.Contains(w.typeFilters, e.Attribute("data-type"))
		}
		if !show || len(w.tagFilters) > 0 {
			show = w.filterThroughTags(id)
		}
		if !show && w.query != "" {
			show = w.runQueryOnContent(id)
		}
		if show {
			e.UI.Entity().Activate()
		} else {
			e.UI.Entity().Deactivate()
		}
	}
}

func (w *Workspace) filterThroughTags(id string) bool {
	cc, err := w.cCache.Read(id)
	if err != nil {
		return false
	}
	for i := range cc.Config.Tags {
		if klib.StringsContainsCaseInsensitive(w.tagFilters, cc.Config.Tags[i]) {
			return true
		}
	}
	return false
}

func (w *Workspace) runQueryOnContent(id string) bool {
	cc, err := w.cCache.Read(id)
	if err != nil {
		return false
	}
	// TODO:  Use filters like tag:..., type:..., etc.
	if strings.Contains(cc.Config.NameLower(), w.query) {
		return true
	}
	for i := range cc.Config.Tags {
		if slices.Contains(w.tagFilters, cc.Config.Tags[i]) {
			return true
		}
	}
	return false
}

func (w *Workspace) updateIndexForCachedContent(cc *content_database.CachedContent) error {
	content_database.WriteConfig(cc.Path, cc.Config, w.pfs)
	if err := w.cCache.Index(cc.Path, w.pfs); err != nil {
		slog.Error("failed to index the content after updating tags", "error", err)
		return err
	}
	return nil
}

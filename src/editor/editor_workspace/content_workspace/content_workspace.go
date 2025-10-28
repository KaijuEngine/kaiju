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
	"fmt"
	"kaiju/editor/editor_overlay/file_browser"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"log/slog"
	"slices"
	"strings"
)

type Workspace struct {
	common_workspace.CommonWorkspace
	pfs               *project_file_system.FileSystem
	cache             *content_database.Cache
	typeFilters       []string
	tagFilters        []string
	query             string
	entryTemplate     *document.Element
	tagFilterTemplate *document.Element
	addTagbtn         *document.Element
	selectedContent   *document.Element
	rightBody         *document.Element
	tooltip           *document.Element
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

func (w *Workspace) Initialize(host *engine.Host, pfs *project_file_system.FileSystem, cdb *content_database.Cache) {
	defer tracing.NewRegion("ContentWorkspace.Initialize").End()
	w.pfs = pfs
	w.cache = cdb
	ids := w.pageData.SetupUIData(cdb)
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/content_workspace.go.html", w.pageData, map[string]func(*document.Element){
			"inputFilter":     w.inputFilter,
			"tagFilter":       w.tagFilter,
			"clickImport":     w.clickImport,
			"clickFilter":     w.clickFilter,
			"clickEntry":      w.clickEntry,
			"clickDeleteTag":  w.clickDeleteTag,
			"updateTagHint":   w.updateTagHint,
			"submitNewTag":    w.submitNewTag,
			"clickTagHint":    w.clickTagHint,
			"submitName":      w.submitName,
			"clickReimport":   w.clickReimport,
			"entryMouseEnter": w.entryMouseEnter,
			"entryMouseMove":  w.entryMouseMove,
			"entryMouseLeave": w.entryMouseLeave,
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
	w.tooltip, _ = w.Doc.GetElementById("tooltip")
	w.addContent(ids)
}

func (w *Workspace) Open() {
	defer tracing.NewRegion("ContentWorkspace.Open").End()
	w.CommonOpen()
	w.entryTemplate.UI.Hide()
	w.tagFilterTemplate.UI.Hide()
	w.info.entryTagTemplate.UI.Hide()
	w.info.tagHintTemplate.UI.Hide()
	w.info.newTagHint.UI.Hide()
	w.tooltip.UI.Hide()
	if w.selectedContent == nil {
		w.rightBody.UI.Hide()
	}
	w.Doc.Clean()
}

func (w *Workspace) Close() {
	defer tracing.NewRegion("ContentWorkspace.Close").End()
	w.CommonClose()
}

func (w *Workspace) clickImport(*document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickImport").End()
	w.UiMan.DisableUpdate()
	file_browser.Show(w.Host, file_browser.Config{
		ExtFilter:   content_database.ImportableTypes,
		MultiSelect: true,
		OnConfirm: func(paths []string) {
			w.UiMan.EnableUpdate()
			index := []string{}
			for i := range paths {
				res, err := content_database.Import(paths[i], w.pfs, w.cache, "")
				if err != nil {
					slog.Error("failed to import content", "path", paths[i], "error", err)
				} else {
					var addDependencies func(target *content_database.ImportResult)
					addDependencies = func(target *content_database.ImportResult) {
						index = append(index, target.Id)
						for i := range target.Dependencies {
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
	defer tracing.NewRegion("ContentWorkspace.addContent").End()
	if len(ids) == 0 {
		return
	}
	ccAll := make([]content_database.CachedContent, 0, len(ids))
	for i := range ids {
		cc, err := w.cache.Read(ids[i])
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
		w.loadEntryImage(cpys[i], cc.Path, cc.Config.Type)
		tex, err := w.Host.TextureCache().Texture(
			fmt.Sprintf("editor/textures/icons/%s.png", cc.Config.Type),
			rendering.TextureFilterLinear)
		if err == nil {
			cpys[i].Children[2].UI.ToPanel().SetBackground(tex)
		}
	}
}

func (w *Workspace) loadEntryImage(e *document.Element, configPath, typeName string) {
	defer tracing.NewRegion("ContentWorkspace.loadEntryImage").End()
	img := e.Children[0].UI.ToPanel()
	if typeName == (content_database.Texture{}).TypeName() {
		// Loose goroutine
		go func() {
			path := content_database.ToContentPath(configPath)
			data, err := w.pfs.ReadFile(path)
			if err != nil {
				slog.Error("error reading the image file", "path", path)
				return
			}
			tex, err := rendering.NewTextureFromMemory(rendering.GenerateUniqueTextureKey,
				data, 0, 0, rendering.TextureFilterLinear)
			if err != nil {
				slog.Error("failed to insert the texture to the cache", "error", err)
				return
			}
			w.Host.RunOnMainThread(func() {
				tex.DelayedCreate(w.Host.Window.Renderer)
				img.SetBackground(tex)
			})
		}()
	}
}

func (w *Workspace) inputFilter(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.inputFilter").End()
	w.query = strings.ToLower(e.UI.ToInput().Text())
	// TODO:  Regex out the filters like tag:..., type:..., etc.
	w.runFilter()
}

func (w *Workspace) tagFilter(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.tagFilter").End()
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
	defer tracing.NewRegion("ContentWorkspace.clickFilter").End()
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
	defer tracing.NewRegion("ContentWorkspace.clickEntry").End()
	if w.selectedContent == e {
		return
	}
	id := e.Attribute("id")
	cc, err := w.cache.Read(id)
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
	defer tracing.NewRegion("ContentWorkspace.clickDeleteTag").End()
	id := w.selectedId()
	cc, err := w.cache.Read(id)
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
	defer tracing.NewRegion("ContentWorkspace.updateTagHint").End()
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
	defer tracing.NewRegion("ContentWorkspace.submitNewTag").End()
	input := e.UI.ToInput()
	txt := input.Text()
	if strings.TrimSpace(txt) == "" {
		return
	}
	w.addTagToSelected(txt)
	input.SetTextWithoutEvent("")
}

func (w *Workspace) clickTagHint(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickTagHint").End()
	w.addTagToSelected(e.Children[0].UI.ToLabel().Text())
	w.info.newTagHint.UI.Hide()
	w.info.newTagInput.UI.ToInput().SetTextWithoutEvent("")
}

func (w *Workspace) submitName(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.submitName").End()
	name := strings.TrimSpace(e.UI.ToInput().Text())
	if name == "" {
		slog.Warn("The name for the content can't be left blank, ignoring change")
		return
	}
	id := w.selectedId()
	cc, err := w.cache.Read(id)
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
	w.cache.Index(cc.Path, w.pfs)
}

func (w *Workspace) clickReimport(*document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickReimport").End()
	res, err := content_database.Reimport(w.selectedId(), w.pfs, w.cache)
	if err != nil {
		slog.Error("failed to re-import the content", "error", err)
		return
	}
	slog.Info("successfully re-imported the content")
	w.loadEntryImage(w.selectedContent, res.ConfigPath(), res.Category.TypeName())
}

func (w *Workspace) entryMouseEnter(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseEnter").End()
	ui := w.tooltip.UI
	id := e.Attribute("id")
	cc, err := w.cache.Read(id)
	if err != nil {
		slog.Error("failed to find the config for the selected entry", "id", id, "error", err)
		return
	}
	ui.Show()
	lbl := w.tooltip.Children[0].UI.ToLabel()
	if len(cc.Config.Tags) == 0 {
		lbl.SetText(fmt.Sprintf("Name: %s\nType: %s", cc.Config.Name, cc.Config.Type))
	} else {
		lbl.SetText(fmt.Sprintf("Name: %s\nType: %s\nTags: %s",
			cc.Config.Name, cc.Config.Type, strings.Join(cc.Config.Tags, ",")))
	}
}

func (w *Workspace) entryMouseMove(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseMove").End()
	ui := w.tooltip.UI
	if !ui.Entity().IsActive() {
		ui.Show()
	}
	// Running on the main thread so it's up to date with the mouse position on
	// the next frame. Maybe there's no need for this...
	w.Host.RunOnMainThread(func() {
		p := w.Host.Window.Mouse.ScreenPosition()
		// Offsetting the box so the mouse doesn't collide with it easily
		ui.Layout().SetOffset(p.X()+10, p.Y()+20)
	})
}

func (w *Workspace) entryMouseLeave(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseLeave").End()
	w.tooltip.UI.Hide()
}

func (w *Workspace) addTagToSelected(tag string) {
	defer tracing.NewRegion("ContentWorkspace.addTagToSelected").End()
	id := w.selectedId()
	cc, err := w.cache.Read(id)
	if err != nil {
		slog.Error("failed to find the config to add tag to content", "id", id, "error", err)
		return
	}
	var ok bool
	if tag, ok = cc.Config.AddTag(tag); ok {
		w.updateIndexForCachedContent(&cc)
	}
	w.clickEntry(w.selectedContent)
	// Add the tag to the entry details
	tagListEntry := w.Doc.DuplicateElement(w.info.entryTagTemplate)
	tagListEntry.Children[0].Children[0].UI.ToLabel().SetText(tag)
	tagListEntry.Children[1].SetAttribute("data-tag", tag)
	tagListEntry.UI.Show()
	// Add the tag to the tag filters if it's not already
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
	defer tracing.NewRegion("ContentWorkspace.selectedId").End()
	if w.selectedContent != nil {
		return w.selectedContent.Attribute("id")
	}
	return ""
}

func (w *Workspace) runFilter() {
	defer tracing.NewRegion("ContentWorkspace.runFilter").End()
	entries := w.Doc.GetElementsByGroup("entry")
	for i := range entries {
		e := entries[i]
		id := e.Attribute("id")
		if id == "entryTemplate" {
			continue
		}
		if ShouldShowContent(w.query, id, w.typeFilters, w.tagFilters, w.cache) {
			e.UI.Entity().Activate()
		} else {
			e.UI.Entity().Deactivate()
		}
	}
}

func (w *Workspace) updateIndexForCachedContent(cc *content_database.CachedContent) error {
	defer tracing.NewRegion("ContentWorkspace.updateIndexForCachedContent").End()
	content_database.WriteConfig(cc.Path, cc.Config, w.pfs)
	if err := w.cache.Index(cc.Path, w.pfs); err != nil {
		slog.Error("failed to index the content after updating tags", "error", err)
		return err
	}
	return nil
}

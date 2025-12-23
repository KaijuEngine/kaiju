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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"errors"
	"fmt"
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/editor_overlay/context_menu"
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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"weak"
)

type ContentWorkspace struct {
	common_workspace.CommonWorkspace
	pfs                *project_file_system.FileSystem
	cache              *content_database.Cache
	editor             ContentWorkspaceEditorInterface
	typeFilters        []string
	typeFiltersDisable []string
	tagFilters         []string
	tagFiltersDisable  []string
	query              string
	contentList        *document.Element
	entryTemplate      *document.Element
	tagFilterTemplate  *document.Element
	addTagbtn          *document.Element
	selectedContent    []*document.Element
	rightBody          *document.Element
	tooltip            *document.Element
	pageData           WorkspaceUIData
	isListMode         bool
	audio              ContentAudioView
	info               struct {
		multiSelectNote  *document.Element
		nameInput        *document.Element
		tagList          *document.Element
		entryTagTemplate *document.Element
		newTagInput      *document.Element
		newTagHint       *document.Element
		tagHintTemplate  *document.Element
	}
}

func (w *ContentWorkspace) Initialize(host *engine.Host, editor ContentWorkspaceEditorInterface) {
	defer tracing.NewRegion("ContentWorkspace.Initialize").End()
	w.pfs = editor.ProjectFileSystem()
	w.cache = editor.Cache()
	w.editor = editor
	w.audio.workspace = weak.Make(w)
	ids := w.pageData.SetupUIData(w.cache)
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/content_workspace.go.html", w.pageData, map[string]func(*document.Element){
			"inputFilter":         w.inputFilter,
			"tagFilter":           w.tagFilter,
			"clickImport":         w.clickImport,
			"toggleListView":      w.toggleListView,
			"clickFilter":         w.clickFilter,
			"clickEntry":          w.clickEntry,
			"clickDeleteTag":      w.clickDeleteTag,
			"updateTagHint":       w.updateTagHint,
			"submitNewTag":        w.submitNewTag,
			"clickTagHint":        w.clickTagHint,
			"submitName":          w.submitName,
			"clickReimport":       w.clickReimport,
			"clickDelete":         w.clickDelete,
			"entryMouseEnter":     w.entryMouseEnter,
			"entryMouseMove":      w.entryMouseMove,
			"entryMouseLeave":     w.entryMouseLeave,
			"rightClickContent":   w.rightClickContent,
			"clickClearSelection": w.clickClearSelection,
			"clickPlayAudio":      w.clickPlayAudio,
			"changeAudioPosition": w.changeAudioPosition,
			"clickOpenInEditor":   w.clickOpenInEditor,
		})
	w.contentList, _ = w.Doc.GetElementById("contentList")
	w.entryTemplate, _ = w.Doc.GetElementById("entryTemplate")
	w.tagFilterTemplate, _ = w.Doc.GetElementById("tagFilterTemplate")
	w.info.entryTagTemplate, _ = w.Doc.GetElementById("entryTagTemplate")
	w.addTagbtn, _ = w.Doc.GetElementById("addTagbtn")
	w.rightBody, _ = w.Doc.GetElementById("rightBody")
	w.info.multiSelectNote, _ = w.Doc.GetElementById("multiSelectNote")
	w.info.nameInput, _ = w.Doc.GetElementById("entryName")
	w.info.tagList, _ = w.Doc.GetElementById("entryTags")
	w.info.newTagInput, _ = w.Doc.GetElementById("newTagInput")
	w.info.newTagHint, _ = w.Doc.GetElementById("newTagHint")
	w.info.tagHintTemplate, _ = w.Doc.GetElementById("tagHintTemplate")
	w.tooltip, _ = w.Doc.GetElementById("tooltip")
	w.audio.audioPlayer, _ = w.Doc.GetElementById("audioPlayer")
	edEvts := w.editor.Events()
	edEvts.OnContentAdded.Add(w.addContent)
	edEvts.OnFocusContent.Add(w.focusContent)
	edEvts.OnContentAdded.Execute(ids)
	w.audio.audioPlayer.UI.Entity().OnDeactivate.Add(w.audio.stopAudio)
}

func (w *ContentWorkspace) Open() {
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
	w.runFilter()
}

func (w *ContentWorkspace) Close() {
	defer tracing.NewRegion("ContentWorkspace.Close").End()
	w.CommonClose()
}

func (w *ContentWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *ContentWorkspace) clickImport(*document.Element) {
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
				for j := range res {
					if err != nil {
						slog.Error("failed to import content", "path", paths[i], "error", err)
					} else {
						var addDependencies func(target *content_database.ImportResult)
						addDependencies = func(target *content_database.ImportResult) {
							index = append(index, target.Id)
							for k := range target.Dependencies {
								addDependencies(&target.Dependencies[k])
							}
						}
						addDependencies(&res[j])
					}
				}
			}
			w.editor.Events().OnContentAdded.Execute(index)
		}, OnCancel: func() {
			w.UiMan.EnableUpdate()
		},
	})
}

func (w *ContentWorkspace) toggleListView(e *document.Element) {
	if w.isListMode {
		w.disableListMode()
		w.Doc.SetElementClassesWithoutApply(e, "filterBtn")
	} else {
		w.enableListMode()
		w.Doc.SetElementClassesWithoutApply(e, "filterBtn", "filterBtnSelected")
	}
	w.Doc.ApplyStyles()
	w.contentList.UIPanel.SetScrollY(0)
}

func (w *ContentWorkspace) addContent(ids []string) {
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
	cpys := w.Doc.DuplicateElementRepeatWithoutApplyStyles(w.entryTemplate, len(ccAll))
	for i := range cpys {
		cc := &ccAll[i]
		w.Doc.SetElementIdWithoutApplyStyles(cpys[i], cc.Id())
		cpys[i].SetAttribute("data-type", strings.ToLower(cc.Config.Type))
		lbl := cpys[i].Children[1].InnerLabel()
		lbl.SetText(cc.Config.Name)
		w.loadEntryImage(cpys[i], cc)
		tex, err := w.Host.TextureCache().Texture(
			fmt.Sprintf("editor/textures/icons/%s.png", cc.Config.Type),
			rendering.TextureFilterLinear)
		if err == nil {
			cpys[i].Children[2].UI.ToPanel().SetBackground(tex)
		}
	}
	w.Doc.ApplyStyles()
}

func (w *ContentWorkspace) focusContent(id string) {
	defer tracing.NewRegion("ContentWorkspace.focusContent").End()
	if !w.Doc.IsActive() {
		return
	}
	elm, ok := w.Doc.GetElementById(id)
	if !ok {
		slog.Warn("could not locate the content in the content workspace", "id", id)
		return
	}
	w.clickEntry(elm)
}

func (w *ContentWorkspace) loadEntryImage(e *document.Element, cc *content_database.CachedContent) {
	defer tracing.NewRegion("ContentWorkspace.loadEntryImage").End()
	img := e.Children[0].UI.ToPanel()
	if cc.Config.Type == (content_database.Texture{}).TypeName() {
		// goroutine
		go func() {
			tex, err := w.Host.TextureCache().Texture(cc.Id(), rendering.TextureFilterLinear)
			if err != nil {
				slog.Error("failed to load the texture", "id", cc.Id(), "error", err)
				return
			}
			// This has to happen before delayed create to have access to the texture data
			isTransparent := tex.ReadPendingDataForTransparency()
			w.Host.RunOnMainThread(func() {
				img.SetBackground(tex)
				if isTransparent {
					img.SetUseBlending(true)
				}
			})
		}()
	}
}

func (w *ContentWorkspace) inputFilter(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.inputFilter").End()
	w.query = strings.ToLower(e.UI.ToInput().Text())
	// TODO:  Regex out the filters like tag:..., type:..., etc.
	w.runFilter()
}

func (w *ContentWorkspace) tagFilter(e *document.Element) {
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

func (w *ContentWorkspace) clickFilter(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickFilter").End()
	inverted := w.Host.Window.Keyboard.HasAlt()
	isSelected := false
	if inverted {
		isSelected = slices.Contains(e.ClassList(), "inverted")
	} else {
		isSelected = slices.Contains(e.ClassList(), "selected")
	}
	isSelected = !isSelected
	typeName := e.Attribute("data-type")
	tagName := e.Attribute("data-tag")
	var targetList *[]string
	var invTargetList *[]string
	var name string
	if typeName != "" {
		targetList = &w.typeFilters
		invTargetList = &w.typeFiltersDisable
		name = typeName
	}
	if tagName != "" {
		targetList = &w.tagFilters
		invTargetList = &w.tagFiltersDisable
		name = tagName
	}
	if inverted {
		targetList, invTargetList = invTargetList, targetList
	}
	if isSelected {
		className := "selected"
		if inverted {
			className = "inverted"
		}
		w.Doc.SetElementClasses(e, "filterBtn", className)
		*targetList = append(*targetList, name)
	} else {
		w.Doc.SetElementClasses(e, "filterBtn")
		*targetList = klib.SlicesRemoveElement(*targetList, name)
	}
	// Remove it from inverse list in both cases intentionally
	*invTargetList = klib.SlicesRemoveElement(*invTargetList, name)
	w.runFilter()
}

func (w *ContentWorkspace) clearSelection() {
	for i := range w.selectedContent {
		w.Doc.SetElementClassesWithoutApply(w.selectedContent[i],
			w.unselectedEntryClasses()...)
	}
	w.Doc.ApplyStyles()
	w.selectedContent = klib.WipeSlice(w.selectedContent)
	w.hideRightPanel()
}

func (w *ContentWorkspace) selectedEntryClasses() []string {
	if w.isListMode {
		return []string{"entry", "wide", "selected"}
	} else {
		return []string{"entry", "selected"}
	}
}

func (w *ContentWorkspace) unselectedEntryClasses() []string {
	if w.isListMode {
		return []string{"entry", "wide"}
	} else {
		return []string{"entry"}
	}
}

func (w *ContentWorkspace) appendSelected(e *document.Element) {
	w.selectedContent = append(w.selectedContent, e)
	w.Doc.SetElementClasses(e, w.selectedEntryClasses()...)
}

func (w *ContentWorkspace) removeSelected(e *document.Element) {
	w.Doc.SetElementClasses(e, w.unselectedEntryClasses()...)
	w.selectedContent = klib.SlicesRemoveElement(w.selectedContent, e)
}

func (w *ContentWorkspace) clickEntry(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickEntry").End()
	if len(w.selectedContent) == 1 && w.selectedContent[0] == e {
		return
	}
	kb := &w.Host.Window.Keyboard
	if kb.HasShift() && len(w.selectedContent) > 0 {
		list := w.selectedContent[0].Parent.Value()
		from := list.IndexOfChild(e)
		idx := list.IndexOfChild(w.selectedContent[len(w.selectedContent)-1])
		if from == idx {
			return
		}
		if idx < from {
			from, idx = idx, from
		}
		for i := from; i <= idx; i++ {
			t := list.Children[i]
			if slices.Contains(w.selectedContent, t) || !t.UI.IsActive() {
				continue
			}
			w.appendSelected(t)
		}
	} else if kb.HasCtrl() && slices.Contains(w.selectedContent, e) {
		w.removeSelected(e)
	} else {
		if !kb.HasCtrl() {
			w.clearSelection()
		}
		w.appendSelected(e)
	}
	w.showRightPanel()
	e.Parent.Value().UI.ToPanel().ScrollToChild(e.UI)
}

func (w *ContentWorkspace) hideRightPanel() {
	w.rightBody.UI.Hide()
}

func (w *ContentWorkspace) showRightPanel() {
	if len(w.selectedContent) == 0 {
		return
	}
	id := w.selectedContent[0].Attribute("id")
	cc, err := w.cache.Read(id)
	if err != nil {
		slog.Error("failed to find the config for the selected entry", "id", id, "error", err)
		return
	}
	w.rightBody.UI.Show()
	for i := len(w.info.tagList.Children) - 1; i >= 1; i-- {
		w.Doc.RemoveElement(w.info.tagList.Children[i])
	}
	w.info.nameInput.UI.ToInput().SetText(cc.Config.Name)
	cpys := w.Doc.DuplicateElementRepeat(w.info.entryTagTemplate, len(cc.Config.Tags))
	for i := range cpys {
		cpys[i].Children[0].InnerLabel().SetText(cc.Config.Tags[i])
		cpys[i].Children[1].SetAttribute("data-tag", cc.Config.Tags[i])
		cpys[i].UI.Show()
	}
	if len(w.selectedContent) > 1 {
		w.info.multiSelectNote.UI.Show()
	} else {
		w.info.multiSelectNote.UI.Hide()
	}
	w.audio.setAudioPanelVisibility(w.selectedContent[0])
}

func (w *ContentWorkspace) clickDeleteTag(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickDeleteTag").End()
	ids := w.selectedIds()
	found := false
	tag := e.Attribute("data-tag")
	for _, id := range ids {
		cc, err := w.cache.Read(id)
		if err != nil {
			slog.Error("failed to find the config to delete tag from content", "id", id, "error", err)
			continue
		}
		if cc.Config.RemoveTag(tag) {
			w.updateIndexForCachedContent(&cc)
		}
		found = true
	}
	if !found {
		slog.Error("failed to locate the tag that was expected to exist", "tag", tag)
	}
	w.Doc.RemoveElement(e.Parent.Value())
}

func (w *ContentWorkspace) updateTagHint(e *document.Element) {
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
		cpys[i].InnerLabel().SetText(options[i])
		cpys[i].UI.Show()
	}
	w.info.newTagHint.UI.Show()
}

func (w *ContentWorkspace) submitNewTag(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.submitNewTag").End()
	input := e.UI.ToInput()
	txt := input.Text()
	if strings.TrimSpace(txt) == "" {
		return
	}
	w.addTagToSelected(txt)
	input.SetTextWithoutEvent("")
}

func (w *ContentWorkspace) clickTagHint(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickTagHint").End()
	w.addTagToSelected(e.InnerLabel().Text())
	w.info.newTagHint.UI.Hide()
	w.info.newTagInput.UI.ToInput().SetTextWithoutEvent("")
}

func (w *ContentWorkspace) submitName(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.submitName").End()
	name := strings.TrimSpace(e.UI.ToInput().Text())
	if name == "" {
		slog.Warn("The name for the content can't be left blank, ignoring change")
		return
	}
	ids := w.selectedIds()
	for _, id := range ids {
		cc, err := w.cache.Rename(id, name, w.pfs)
		if err != nil {
			continue
		}
		for i := range w.selectedContent {
			w.selectedContent[i].Children[1].InnerLabel().SetText(name)
		}
		w.cache.IndexCachedContent(cc)
		w.editor.Events().OnContentRenamed.Execute(id)
	}
}

func (w *ContentWorkspace) clickReimport(*document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickReimport").End()
	for i, id := range w.selectedIds() {
		res, err := content_database.Reimport(id, w.pfs, w.cache)
		if err != nil {
			slog.Error("failed to re-import the content", "error", err)
			continue
		}
		slog.Info("successfully re-import the content", "id", id)
		cc, err := w.cache.Read(res.Id)
		if err != nil {
			slog.Error("failed to load the re-imported content from cache", "id", res.Id, "error", err)
			continue
		}
		if cc.Config.Type == (content_database.Texture{}).TypeName() {
			w.Host.TextureCache().ReloadTexture(cc.Id(), rendering.TextureFilterLinear)
		}
		w.loadEntryImage(w.selectedContent[i], &cc)
	}
}

func (w *ContentWorkspace) clickDelete(*document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickDelete").End()
	w.UiMan.DisableUpdate()
	confirm_prompt.Show(w.Host, confirm_prompt.Config{
		Title:       "Delete content",
		Description: "Are you sure you wish to delete the selected content?",
		ConfirmText: "Yes",
		CancelText:  "Cancel",
		OnConfirm: func() {
			w.UiMan.EnableUpdate()
			w.completeDeleteOfSelectedContent()
		},
		OnCancel: w.UiMan.EnableUpdate,
	})
}

func (w *ContentWorkspace) completeDeleteOfSelectedContent() {
	ids := w.selectedIds()
	for _, id := range ids {
		err := content_database.Delete(id, w.pfs, w.cache)
		if err != nil {
			if errors.Is(err, content_database.DeleteContentMissingIdError) {
				slog.Warn("clickDelete contained a blank id")
			}
			continue
		}
		w.editor.Events().OnContentRemoved.Execute([]string{id})
		w.rightBody.UI.Hide()
		w.tooltip.UI.Hide()
	}
	for i := range w.selectedContent {
		w.Doc.RemoveElementWithoutApplyStyles(w.selectedContent[i])
	}
	w.Doc.ApplyStyles()
	w.selectedContent = klib.WipeSlice(w.selectedContent)
}

func (w *ContentWorkspace) entryMouseEnter(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseEnter").End()
	ui := w.tooltip.UI
	id := e.Attribute("id")
	cc, err := w.cache.Read(id)
	if err != nil {
		slog.Error("failed to find the config for the selected entry", "id", id, "error", err)
		return
	}
	ui.Show()
	lbl := w.tooltip.InnerLabel()
	if len(cc.Config.Tags) == 0 {
		lbl.SetText(fmt.Sprintf("Name: %s\nType: %s", cc.Config.Name, cc.Config.Type))
	} else {
		lbl.SetText(fmt.Sprintf("Name: %s\nType: %s\nTags: %s",
			cc.Config.Name, cc.Config.Type, strings.Join(cc.Config.Tags, ",")))
	}
}

func (w *ContentWorkspace) entryMouseMove(e *document.Element) {
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

func (w *ContentWorkspace) entryMouseLeave(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseLeave").End()
	w.tooltip.UI.Hide()
}

func (w *ContentWorkspace) rightClickContent(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.rightClickContent").End()
	id := e.Attribute("id")
	options := []context_menu.ContextMenuOption{
		{
			Label: "Copy ID to clipboard",
			Call:  func() { w.Host.Window.CopyToClipboard(id) },
		},
		{
			Label: "Find references",
			Call:  func() { w.editor.ShowReferences(id) },
		},
		{
			Label: "Create table of contents",
			Call:  w.requestCreateTableOfContents,
		},
	}
	if cc, err := w.cache.Read(id); err == nil {
		isEditableText := cc.Config.Type == (content_database.Html{}).TypeName() ||
			cc.Config.Type == (content_database.Css{}).TypeName()
		if cc.Config.Type == (content_database.TableOfContents{}).TypeName() {
			options = append(options, context_menu.ContextMenuOption{
				Label: "Add to table of contents",
				Call: func() {
					if w.checkNoDuplicateNamesForTableOfContents() {
						w.addSelectedToTableOfContents(id)
					}
				},
			}, context_menu.ContextMenuOption{
				Label: "View",
				Call:  func() { w.showTableOfContents(id) },
			})
		} else if cc.Config.Type == (content_database.Html{}).TypeName() {
			options = append(options, context_menu.ContextMenuOption{
				Label: "View in UI workspace",
				Call:  func() { w.editor.ViewHtmlUi(id) },
			})
		}
		if isEditableText {
			options = append(options, context_menu.ContextMenuOption{
				Label: "Open in editor",
				Call: func() {
					w.openInEditor(cc)
				},
			})
		}
	}
	w.editor.BlurInterface()
	context_menu.Show(w.Host, options, w.Host.Window.Cursor.ScreenPosition(), w.editor.FocusInterface)
}

func (w *ContentWorkspace) clickClearSelection(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickClearSelection").End()
	w.clearSelection()
}

func (w *ContentWorkspace) clickPlayAudio(*document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickClearSelection").End()
	w.audio.playAudioId(w.selectedIds()[0])
}

func (w *ContentWorkspace) changeAudioPosition(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickClearSelection").End()
	w.audio.setAudioPosition(e.UI.ToSlider().Value())
}

func (w *ContentWorkspace) clickOpenInEditor(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.clickOpenInEditor").End()
	for _, id := range w.selectedIds() {
		cc, err := w.cache.Read(id)
		if err != nil {
			slog.Error("failed to find the config to add tag to content", "id", id, "error", err)
			continue
		}
		w.openInEditor(cc)
	}
}

func (w *ContentWorkspace) addTagToSelected(tag string) {
	defer tracing.NewRegion("ContentWorkspace.addTagToSelected").End()
	for _, id := range w.selectedIds() {
		cc, err := w.cache.Read(id)
		if err != nil {
			slog.Error("failed to find the config to add tag to content", "id", id, "error", err)
			return
		}
		var ok bool
		if tag, ok = cc.Config.AddTag(tag); ok {
			w.updateIndexForCachedContent(&cc)
		}
	}
	// Add the tag to the entry details
	tagListEntry := w.Doc.DuplicateElement(w.info.entryTagTemplate)
	tagListEntry.Children[0].InnerLabel().SetText(tag)
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
	elm.InnerLabel().SetText(tag)
	elm.SetAttribute("data-tag", tag)
}

func (w *ContentWorkspace) selectedIds() []string {
	defer tracing.NewRegion("ContentWorkspace.selectedId").End()
	ids := make([]string, len(w.selectedContent))
	for i := range w.selectedContent {
		ids[i] = w.selectedContent[i].Attribute("id")
	}
	return ids
}

func (w *ContentWorkspace) runFilter() {
	defer tracing.NewRegion("ContentWorkspace.runFilter").End()
	entries := w.Doc.GetElementsByGroup("entry")
	for i := range entries {
		e := entries[i]
		id := e.Attribute("id")
		if id == "entryTemplate" {
			continue
		}
		hide := ShouldHideContent(id, w.typeFiltersDisable, w.tagFiltersDisable, w.cache)
		if !hide && ShouldShowContent(w.query, id, w.typeFilters, w.tagFilters, w.cache) {
			e.UI.Show()
		} else {
			e.UI.Hide()
		}
	}
	w.Host.RunOnMainThread(w.Doc.Clean)
}

func (w *ContentWorkspace) updateIndexForCachedContent(cc *content_database.CachedContent) error {
	defer tracing.NewRegion("ContentWorkspace.updateIndexForCachedContent").End()
	if err := content_database.WriteConfig(cc.Path, cc.Config, w.pfs); err != nil {
		return err
	}
	w.cache.IndexCachedContent(*cc)
	return nil
}

func (w *ContentWorkspace) enableListMode() {
	w.isListMode = true
	for _, c := range w.contentList.Children[1:] {
		w.Doc.SetElementClassesWithoutApply(c, append(c.ClassList(), "wide")...)
		for _, cc := range c.Children {
			w.Doc.SetElementClassesWithoutApply(cc, append(cc.ClassList(), "wide")...)
		}
	}
}

func (w *ContentWorkspace) disableListMode() {
	w.isListMode = false
	for _, c := range w.contentList.Children[1:] {
		w.Doc.SetElementClassesWithoutApply(c, klib.SlicesRemoveElement(c.ClassList(), "wide")...)
		for _, cc := range c.Children {
			cc.UI.Show()
			w.Doc.SetElementClassesWithoutApply(cc, klib.SlicesRemoveElement(cc.ClassList(), "wide")...)
		}
	}
}

func openContentEditor(contentEditor, path string) {
	defer tracing.NewRegion("Editor.openCodeEditor").End()
	dir := filepath.Dir(contentEditor)
	base := filepath.Base(contentEditor)
	fullArgs := strings.Split(base, " ")
	testExePath := filepath.Join(dir, fullArgs[0])
	if _, err := os.Stat(testExePath); err == nil {
		fullArgs[0] = testExePath
	} else {
		fullArgs[0] = contentEditor
	}
	command := fullArgs[0]
	var args []string
	if len(fullArgs) > 1 {
		args = append(args, fullArgs[1:]...)
	}
	args = append(args, path)
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(strings.ToLower(command), "shell:appsfolder") {
			args = slices.Insert(args, 0, "/C", "start", "", command)
			command = "cmd.exe"
		}
	}
	// goroutine
	go exec.Command(command, args...).Run()
}

func (w *ContentWorkspace) openInEditor(cc content_database.CachedContent) {
	ed := ""
	path := w.pfs.FullPath(cc.ContentPath())
	switch cc.Config.Type {
	case content_database.Html{}.TypeName():
		fallthrough
	case content_database.Css{}.TypeName():
		ed = w.editor.Settings().CodeEditor
	case content_database.Mesh{}.TypeName():
		ed = w.editor.Settings().MeshEditor
		if _, err := w.pfs.Stat(cc.Config.SrcPath); err == nil {
			path = w.pfs.FullPath(cc.Config.SrcPath)
		} else if _, err := os.Stat(cc.Config.SrcPath); err == nil {
			path = cc.Config.SrcPath
		} else {
			path = ""
		}
	case content_database.Music{}.TypeName():
		fallthrough
	case content_database.Sound{}.TypeName():
		ed = w.editor.Settings().AudioEditor
	case content_database.Texture{}.TypeName():
		ed = w.editor.Settings().ImageEditor
	case content_database.ParticleSystem{}.TypeName():
		w.editor.VfxWorkspaceSelected()
		w.editor.VfxWorkspace().OpenParticleSystem(cc.Id())
		return
	case content_database.Material{}.TypeName():
		fallthrough
	case content_database.RenderPass{}.TypeName():
		fallthrough
	case content_database.ShaderPipeline{}.TypeName():
		fallthrough
	case content_database.Shader{}.TypeName():
		w.editor.ShadingWorkspaceSelected()
		w.editor.ShadingWorkspace().OpenSpec(cc.Id())
		return
	case content_database.Stage{}.TypeName():
		w.editor.OpenStageInStageWorkspace(cc.Id())
	case content_database.TableOfContents{}.TypeName():
		w.showTableOfContents(cc.Id())
		return
	case content_database.Spv{}.TypeName():
	case content_database.Template{}.TypeName():
	}
	if path == "" {
		slog.Warn("could not find the source file path for the selected content")
	} else if ed == "" {
		slog.Warn("currently there isn't an editor that can open the content", "type", cc.Config.Type)
	} else {
		openContentEditor(ed, path)
	}
}

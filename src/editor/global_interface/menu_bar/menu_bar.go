/******************************************************************************/
/* menu_bar.go                                                                */
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

package menu_bar

import (
	"kaiju/editor/editor_overlay/create_entity_data"
	"kaiju/editor/editor_overlay/input_prompt"
	"kaiju/editor/project"
	"kaiju/engine"
	"kaiju/engine/systems/logging"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"log/slog"
)

type MenuBar struct {
	doc           *document.Document
	uiMan         ui.Manager
	selectedPopup *document.Element
	handler       MenuBarHandler
}

func (b *MenuBar) Initialize(host *engine.Host, handler MenuBarHandler) error {
	defer tracing.NewRegion("MenuBar.Initialize").End()
	b.handler = handler
	b.uiMan.Init(host)
	var err error
	b.doc, err = markup.DocumentFromHTMLAsset(&b.uiMan, "editor/ui/global/menu_bar.go.html",
		nil, map[string]func(*document.Element){
			"clickLogo":               b.openMenuTarget,
			"clickFile":               b.openMenuTarget,
			"clickEdit":               b.openMenuTarget,
			"clickCreate":             b.openMenuTarget,
			"clickHelp":               b.openMenuTarget,
			"clickStage":              b.clickStage,
			"clickContent":            b.clickContent,
			"clickShading":            b.clickShading,
			"clickAnimation":          b.clickAnimation,
			"clickUI":                 b.clickUI,
			"clickSettings":           b.clickSettings,
			"clickNewStage":           b.clickNewStage,
			"clickOpenStage":          b.clickOpenStage,
			"clickSaveStage":          b.clickSaveStage,
			"clickOpenCodeEditor":     b.clickOpenCodeEditor,
			"clickBuild":              b.clickBuild,
			"clickBuildAndRun":        b.clickBuildAndRun,
			"clickBuildRelease":       b.clickBuildRelease,
			"clickBuildAndRunRelease": b.clickBuildAndRunRelease,
			"clickRunCurrentStage":    b.clickRunCurrentStage,
			"clickBuildAndroid":       b.clickBuildAndroid,
			"clickBuildRunAndroid":    b.clickBuildRunAndroid,
			"clickCreateEntityData":   b.clickCreateEntityData,
			"clickCreateHtmlUi":       b.clickCreateHtmlUi,
			"clickNewCamera":          b.clickNewCamera,
			"clickNewEntity":          b.clickNewEntity,
			"clickNewLight":           b.clickNewLight,
			"clickAbout":              b.clickAbout,
			"clickLogs":               b.clickLogs,
			"clickIssues":             b.clickIssues,
			"clickRepository":         b.clickRepository,
			"clickJoinMailingList":    b.clickJoinMailingList,
			"clickMailArchives":       b.clickMailArchives,
			"popupMiss":               b.popupMiss,
		})
	b.doc.Clean()
	return err
}

func (b *MenuBar) Focus() { b.uiMan.EnableUpdate() }
func (b *MenuBar) Blur()  { b.uiMan.DisableUpdate() }

func (b *MenuBar) SetWorkspaceStage() {
	t, _ := b.doc.GetElementById("tabStage")
	b.selectTab(t)
}

func (b *MenuBar) SetWorkspaceContent() {
	t, _ := b.doc.GetElementById("tabContent")
	b.selectTab(t)
}

func (b *MenuBar) SetWorkspaceShading() {
	t, _ := b.doc.GetElementById("tabShading")
	b.selectTab(t)
}

func (b *MenuBar) SetWorkspaceAnimation() {
	t, _ := b.doc.GetElementById("tabAnimation")
	b.selectTab(t)
}

func (b *MenuBar) SetWorkspaceUI() {
	t, _ := b.doc.GetElementById("tabUI")
	b.selectTab(t)
}

func (b *MenuBar) SetWorkspaceSettings() {
	t, _ := b.doc.GetElementById("tabSettings")
	b.selectTab(t)
}

func (b *MenuBar) openMenuTarget(e *document.Element) {
	defer tracing.NewRegion("MenuBar.openMenuTarget").End()
	target := e.Attribute("data-target")
	pop, _ := b.doc.GetElementById(target)
	b.selectedPopup = pop
	if pop.UI.Entity().IsActive() {
		b.hidePopups()
	} else {
		pops := b.doc.GetElementsByClass("popup")
		for i := range pops {
			if pop != pops[i] {
				pops[i].UI.Hide()
			}
		}
		pop.UI.Show()
		t := &e.UI.Entity().Transform
		x := t.WorldPosition().X() + float32(b.uiMan.Host.Window.Width())*0.5 -
			e.UI.Layout().PixelSize().X()*0.5
		pop.UI.Layout().SetInnerOffsetLeft(x)
		b.handler.BlurInterface()
		b.uiMan.Host.RunOnMainThread(b.Focus)
	}
}

func (b *MenuBar) clickStage(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickStage").End()
	b.selectTab(e)
	b.handler.StageWorkspaceSelected()
}

func (b *MenuBar) clickContent(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickContent").End()
	b.selectTab(e)
	b.handler.ContentWorkspaceSelected()
}

func (b *MenuBar) clickShading(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickShading").End()
	b.selectTab(e)
	b.handler.ShadingWorkspaceSelected()
}

func (b *MenuBar) clickAnimation(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickAnimation").End()
	b.selectTab(e)
}

func (b *MenuBar) clickUI(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickUI").End()
	b.selectTab(e)
	b.handler.UIWorkspaceSelected()
}

func (b *MenuBar) clickSettings(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickSettings").End()
	b.selectTab(e)
	b.handler.SettingsWorkspaceSelected()
}

func (b *MenuBar) clickNewStage(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickNewStage").End()
	b.hidePopups()
	b.handler.CreateNewStage()
}

func (b *MenuBar) clickOpenStage(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickOpenStage").End()
	b.hidePopups()
}

func (b *MenuBar) clickSaveStage(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickSaveStage").End()
	b.hidePopups()
	b.handler.SaveCurrentStage()
}

func (b *MenuBar) clickCreateEntityData(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateEntityData").End()
	b.hidePopups()
	b.handler.BlurInterface()
	create_entity_data.Show(b.uiMan.Host, b.handler.ProjectFileSystem(), create_entity_data.Config{
		OnCreate: func() {
			// TODO:  Rather than doing a broad ReadSourceCode, just do a
			// targeted import of the specific file. Would need to return it
			// in the callback.
			// goroutine
			go b.handler.Project().ReadSourceCode()
			b.handler.FocusInterface()
		},
		OnCancel: b.handler.FocusInterface,
	})
}

func (b *MenuBar) clickCreateHtmlUi(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateHtmlUi").End()
	b.hidePopups()
	b.handler.BlurInterface()
	input_prompt.Show(b.uiMan.Host, input_prompt.Config{
		Title:       "Name your HTML file",
		Description: "Give a friendly name to your html file",
		Placeholder: "Name...",
		ConfirmText: "Create",
		CancelText:  "Cancel",
		OnCancel:    b.handler.FocusInterface,
		OnConfirm: func(name string) {
			b.handler.FocusInterface()
			b.handler.CreateHtmlUiFile(name)
		},
	})
}

func (b *MenuBar) clickOpenCodeEditor(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickOpenCodeEditor").End()
	b.hidePopups()
	b.handler.OpenCodeEditor()
}

func (b *MenuBar) clickBuild(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuild").End()
	b.hidePopups()
	b.handler.Build(project.GameBuildModeDebug)
}

func (b *MenuBar) clickBuildAndRun(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuildAndRun").End()
	b.hidePopups()
	b.handler.BuildAndRun(project.GameBuildModeDebug)
}

func (b *MenuBar) clickBuildRelease(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuildRelease").End()
	b.hidePopups()
	b.handler.Build(project.GameBuildModeRelease)
}

func (b *MenuBar) clickBuildAndRunRelease(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuildAndRunRelease").End()
	b.hidePopups()
	b.handler.BuildAndRun(project.GameBuildModeRelease)
}

func (b *MenuBar) clickRunCurrentStage(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickRunCurrentStage").End()
	b.hidePopups()
	b.handler.BuildAndRunCurrentStage()
}

func (b *MenuBar) clickBuildAndroid(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuildAndroid").End()
	b.hidePopups()
	bts := b.handler.Settings().BuildTools
	// goroutine
	go b.handler.Project().BuildAndroid(bts.AndroidNDK, bts.JavaHome, []string{"debug"})
}

func (b *MenuBar) clickBuildRunAndroid(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuildRunAndroid").End()
	b.hidePopups()
	bts := b.handler.Settings().BuildTools
	// goroutine
	go b.handler.Project().BuildRunAndroid(bts.AndroidNDK, bts.JavaHome, []string{"debug"})
}

func (b *MenuBar) clickNewCamera(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickNewCamera").End()
	b.hidePopups()
	b.handler.CreateNewCamera()
}

func (b *MenuBar) clickNewEntity(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickNewEntity").End()
	b.hidePopups()
	b.handler.CreateNewEntity()
}

func (b *MenuBar) clickNewLight(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickNewLight").End()
	b.hidePopups()
	b.handler.CreateNewLight()
}

func (b *MenuBar) clickAbout(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickAbout").End()
	b.hidePopups()
}

func (b *MenuBar) clickLogs(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickIssues").End()
	b.hidePopups()
	if dir, err := logging.LogFolderPath(); err == nil {
		if err = filesystem.OpenFileBrowserToFolder(dir); err != nil {
			slog.Error("failed to open the file browser to the folder",
				"folder", dir, "error", err)
		}
	} else {
		slog.Error("failed to locate the log folder path", "error", err)
	}
}

func (b *MenuBar) clickIssues(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickIssues").End()
	b.hidePopups()
	klib.OpenWebsite("https://github.com/KaijuEngine/kaiju/issues")
}

func (b *MenuBar) clickRepository(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickRepository").End()
	b.hidePopups()
	klib.OpenWebsite("https://github.com/KaijuEngine/kaiju")
}

func (b *MenuBar) clickJoinMailingList(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickJoinMailingList").End()
	b.hidePopups()
	klib.OpenWebsite("https://www.freelists.org/list/kaijuengine")
}

func (b *MenuBar) clickMailArchives(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickMailArchives").End()
	b.hidePopups()
	klib.OpenWebsite("https://www.freelists.org/archive/kaijuengine/")
}

func (b *MenuBar) popupMiss(e *document.Element) {
	defer tracing.NewRegion("MenuBar.popupMiss").End()
	if e == b.selectedPopup {
		b.hidePopups()
	}
}

func (b *MenuBar) selectTab(e *document.Element) {
	defer tracing.NewRegion("MenuBar.selectTab").End()
	tabs := b.doc.GetElementsByGroup("tabs")
	for i := range tabs {
		b.doc.SetElementClassesWithoutApply(tabs[i], "workspaceTab")
	}
	b.doc.SetElementClassesWithoutApply(e, "workspaceTab", "tabSelected")
	b.doc.ApplyStyles()
	b.hidePopups()
}

func (b *MenuBar) hidePopups() {
	defer tracing.NewRegion("MenuBar.hidePopups").End()
	pops := b.doc.GetElementsByClass("popup")
	for i := range pops {
		pops[i].UI.Hide()
	}
	b.selectedPopup = nil
	b.handler.FocusInterface()
	b.uiMan.Host.RunOnMainThread(b.handler.FocusInterface)
}

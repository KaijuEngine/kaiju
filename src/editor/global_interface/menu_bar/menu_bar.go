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

package menu_bar

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
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
			"clickLogo":            b.openMenuTarget,
			"clickFile":            b.openMenuTarget,
			"clickEdit":            b.openMenuTarget,
			"clickEntity":          b.openMenuTarget,
			"clickHelp":            b.openMenuTarget,
			"clickStage":           b.clickStage,
			"clickContent":         b.clickContent,
			"clickShading":         b.clickShading,
			"clickAnimation":       b.clickAnimation,
			"clickUI":              b.clickUI,
			"clickSettings":        b.clickSettings,
			"clickNewStage":        b.clickNewStage,
			"clickOpenStage":       b.clickOpenStage,
			"clickSaveStage":       b.clickSaveStage,
			"clickOpenVSCode":      b.clickOpenVSCode,
			"clickBuild":           b.clickBuild,
			"clickBuildAndRun":     b.clickBuildAndRun,
			"clickRunCurrentStage": b.clickRunCurrentStage,
			"clickNewCamera":       b.clickNewCamera,
			"clickAbout":           b.clickAbout,
			"clickIssues":          b.clickIssues,
			"clickRepository":      b.clickRepository,
			"clickJoinMailingList": b.clickJoinMailingList,
			"clickMailArchives":    b.clickMailArchives,
			"popupMiss":            b.popupMiss,
		})
	return err
}

func (b *MenuBar) Focus() { b.uiMan.EnableUpdate() }
func (b *MenuBar) Blur()  { b.uiMan.DisableUpdate() }

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
}

func (b *MenuBar) clickAnimation(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickAnimation").End()
	b.selectTab(e)
}

func (b *MenuBar) clickUI(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickUI").End()
	b.selectTab(e)
}

func (b *MenuBar) clickSettings(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickSettings").End()
	b.selectTab(e)
}

func (b *MenuBar) clickNewStage(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickNewStage").End()
	b.hidePopups()
	b.handler.CreateNewStage()
}

func (b *MenuBar) clickOpenStage(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickOpenStage").End()
	b.hidePopups()
}

func (b *MenuBar) clickSaveStage(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickSaveStage").End()
	b.hidePopups()
	b.handler.SaveCurrentStage()
}

func (b *MenuBar) clickOpenVSCode(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickOpenVSCode").End()
	b.hidePopups()
	b.handler.OpenVSCodeProject()
}

func (b *MenuBar) clickBuild(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuild").End()
	b.hidePopups()
	b.handler.Build()
}

func (b *MenuBar) clickBuildAndRun(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuildAndRun").End()
	b.hidePopups()
	b.handler.BuildAndRun()
}

func (b *MenuBar) clickRunCurrentStage(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickRunCurrentStage").End()
	b.hidePopups()
	b.handler.BuildAndRunCurrentStage()
}

func (b *MenuBar) clickNewCamera(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickNewCamera").End()
	b.hidePopups()
	b.handler.CreateNewCamera()
}

func (b *MenuBar) clickAbout(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickAbout").End()
	b.hidePopups()
}

func (b *MenuBar) clickIssues(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickIssues").End()
	b.hidePopups()
	klib.OpenWebsite("https://github.com/KaijuEngine/kaiju/issues")
}

func (b *MenuBar) clickRepository(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickRepository").End()
	b.hidePopups()
	klib.OpenWebsite("https://github.com/KaijuEngine/kaiju")
}

func (b *MenuBar) clickJoinMailingList(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickJoinMailingList").End()
	b.hidePopups()
	klib.OpenWebsite("https://www.freelists.org/list/kaijuengine")
}

func (b *MenuBar) clickMailArchives(e *document.Element) {
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
	b.uiMan.Host.RunOnMainThread(b.handler.FocusInterface)
}

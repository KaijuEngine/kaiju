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
	defer tracing.NewRegion("TitleBar.Initialize").End()
	b.handler = handler
	b.uiMan.Init(host)
	var err error
	b.doc, err = markup.DocumentFromHTMLAsset(&b.uiMan, "editor/ui/global/menu_bar.go.html",
		nil, map[string]func(*document.Element){
			"clickLogo":            b.openMenuTarget,
			"clickFile":            b.openMenuTarget,
			"clickEdit":            b.openMenuTarget,
			"clickHelp":            b.openMenuTarget,
			"clickStage":           b.clickStage,
			"clickContent":         b.clickContent,
			"clickAnimation":       b.clickAnimation,
			"clickUI":              b.clickUI,
			"clickLog":             b.clickLog,
			"clickNewStage":        b.clickNewStage,
			"clickOpenStage":       b.clickOpenStage,
			"clickSaveStage":       b.clickSaveStage,
			"clickProjectSettings": b.clickProjectSettings,
			"clickEditorSettings":  b.clickEditorSettings,
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
	}
}

func (b *MenuBar) clickStage(e *document.Element) {
	b.selectTab(e)
	b.handler.OnStageWorkspaceSelected()
}

func (b *MenuBar) clickContent(e *document.Element) {
	b.selectTab(e)
	b.handler.OnContentWorkspaceSelected()
}

func (b *MenuBar) clickAnimation(e *document.Element) {
	b.selectTab(e)
}

func (b *MenuBar) clickUI(e *document.Element) {
	b.selectTab(e)
}

func (b *MenuBar) clickLog(e *document.Element) {
	b.selectTab(e)
}

func (b *MenuBar) clickNewStage(e *document.Element) {
	b.hidePopups()
}

func (b *MenuBar) clickOpenStage(e *document.Element) {
	b.hidePopups()
}

func (b *MenuBar) clickSaveStage(e *document.Element) {
	b.hidePopups()
}

func (b *MenuBar) clickProjectSettings(e *document.Element) {
	b.hidePopups()
}

func (b *MenuBar) clickEditorSettings(e *document.Element) {
	b.hidePopups()
}

func (b *MenuBar) clickAbout(e *document.Element) {
	b.hidePopups()
}

func (b *MenuBar) clickIssues(e *document.Element) {
	b.hidePopups()
	klib.OpenWebsite("https://github.com/KaijuEngine/kaiju/issues")
}

func (b *MenuBar) clickRepository(e *document.Element) {
	b.hidePopups()
	klib.OpenWebsite("https://github.com/KaijuEngine/kaiju")
}

func (b *MenuBar) clickJoinMailingList(e *document.Element) {
	b.hidePopups()
	klib.OpenWebsite("https://www.freelists.org/list/kaijuengine")
}

func (b *MenuBar) clickMailArchives(e *document.Element) {
	b.hidePopups()
	klib.OpenWebsite("https://www.freelists.org/archive/kaijuengine/")
}

func (b *MenuBar) popupMiss(e *document.Element) {
	if e == b.selectedPopup {
		b.hidePopups()
	}
}

func (b *MenuBar) selectTab(e *document.Element) {
	tabs := b.doc.GetElementsByGroup("tabs")
	for i := range tabs {
		b.doc.SetElementClassesWithoutApply(tabs[i], "workspaceTab")
	}
	b.doc.SetElementClassesWithoutApply(e, "workspaceTab", "tabSelected")
	b.doc.ApplyStyles()
	b.hidePopups()
}

func (b *MenuBar) hidePopups() {
	pops := b.doc.GetElementsByClass("popup")
	for i := range pops {
		pops[i].UI.Hide()
	}
	b.selectedPopup = nil
}

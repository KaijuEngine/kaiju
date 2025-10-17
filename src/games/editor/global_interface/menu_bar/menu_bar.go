package menu_bar

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
)

type MenuBar struct {
	doc   *document.Document
	uiMan ui.Manager
}

func (b *MenuBar) Initialize(host *engine.Host) error {
	defer tracing.NewRegion("TitleBar.Initialize").End()
	b.uiMan.Init(host)
	var err error
	b.doc, err = markup.DocumentFromHTMLAsset(&b.uiMan, "editor/ui/global/menu_bar.go.html",
		nil, map[string]func(*document.Element){
			"clickFile":            b.openMenuTarget,
			"clickEdit":            b.openMenuTarget,
			"clickHelp":            b.openMenuTarget,
			"clickStage3D":         b.clickStage3D,
			"clickStage2D":         b.clickStage2D,
			"clickContent":         b.clickContent,
			"clickAnimation":       b.clickAnimation,
			"clickUI":              b.clickUI,
			"clickNewStage":        b.clickNewStage,
			"clickOpenStage":       b.clickOpenStage,
			"clickSaveStage":       b.clickSaveStage,
			"clickProjectSettings": b.clickProjectSettings,
			"clickEditorSettings":  b.clickEditorSettings,
			"clickAbout":           b.clickAbout,
			"clickRepository":      b.clickRepository,
			"clickJoinMailingList": b.clickJoinMailingList,
			"clickMailArchives":    b.clickMailArchives,
		})
	return err
}

func (b *MenuBar) openMenuTarget(e *document.Element) {
	target := e.Attribute("data-target")
	pop, _ := b.doc.GetElementById(target)
	pops := b.doc.GetElementsByClass("popup")
	for i := range pops {
		if pop != pops[i] {
			pops[i].UI.Hide()
		}
	}
	if pop.UI.Entity().IsActive() {
		pop.UI.Hide()
	} else {
		pop.UI.Show()
		t := &e.UI.Entity().Transform
		x := t.WorldPosition().X() + float32(b.uiMan.Host.Window.Width())*0.5 -
			e.UI.Layout().PixelSize().X()*0.5
		pop.UI.Layout().SetInnerOffsetLeft(x)
	}
}

func (b *MenuBar) clickStage3D(e *document.Element) {
	b.selectTab(e)
}

func (b *MenuBar) clickStage2D(e *document.Element) {
	b.selectTab(e)
}

func (b *MenuBar) clickContent(e *document.Element) {
	b.selectTab(e)
}

func (b *MenuBar) clickAnimation(e *document.Element) {
	b.selectTab(e)
}

func (b *MenuBar) clickUI(e *document.Element) {
	b.selectTab(e)
}

func (b *MenuBar) clickNewStage(e *document.Element) {
}

func (b *MenuBar) clickOpenStage(e *document.Element) {
}

func (b *MenuBar) clickSaveStage(e *document.Element) {
}

func (b *MenuBar) clickProjectSettings(e *document.Element) {
}

func (b *MenuBar) clickEditorSettings(e *document.Element) {
}

func (b *MenuBar) clickAbout(e *document.Element) {
}

func (b *MenuBar) clickRepository(e *document.Element) {
}

func (b *MenuBar) clickJoinMailingList(e *document.Element) {
}

func (b *MenuBar) clickMailArchives(e *document.Element) {
}

func (b *MenuBar) selectTab(e *document.Element) {
	tabs := b.doc.GetElementsByGroup("tabs")
	for i := range tabs {
		b.doc.SetElementClassesWithoutApply(tabs[i], "workspaceTab")
	}
	b.doc.SetElementClassesWithoutApply(e, "workspaceTab", "tabSelected")
	b.doc.ApplyStyles()
}

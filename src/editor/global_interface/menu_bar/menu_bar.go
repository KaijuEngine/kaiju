/******************************************************************************/
/* menu_bar.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package menu_bar

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"kaijuengine.com/editor/editor_overlay/create_entity_data"
	"kaijuengine.com/editor/editor_overlay/file_browser"
	"kaijuengine.com/editor/editor_overlay/input_prompt"
	"kaijuengine.com/editor/editor_overlay/sponsors"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/logging"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

type MenuBar struct {
	host          *engine.Host
	doc           *document.Document
	uiMan         ui.Manager
	selectedPopup *document.Element
	handler       MenuBarHandler
	tabs          []WorkspaceTab
	activeTabID   string
}

type menuBarTemplateData struct {
	ShowGrid      bool
	WorkspaceTabs []WorkspaceTab
	ActiveTabID   string
}

func (b *MenuBar) Initialize(host *engine.Host, handler MenuBarHandler) error {
	defer tracing.NewRegion("MenuBar.Initialize").End()
	b.host = host
	b.handler = handler
	b.uiMan.Init(host)
	return b.renderDocument()
}

// RebuildWorkspaceTabs replaces the menu bar's workspace tab strip with the
// given ordered list and marks activeID as selected. Called by the editor
// after workspace registration / settings changes. Re-renders the document
// (popup state is rebuilt; this is rare enough that the cost is acceptable).
func (b *MenuBar) RebuildWorkspaceTabs(tabs []WorkspaceTab, activeID string) {
	defer tracing.NewRegion("MenuBar.RebuildWorkspaceTabs").End()
	b.tabs = make([]WorkspaceTab, len(tabs))
	copy(b.tabs, tabs)
	b.activeTabID = activeID
	if err := b.renderDocument(); err != nil {
		slog.Error("failed to rebuild menu bar tabs", "error", err)
	}
}

// SetActiveTab updates the selected tab styling without rebuilding the
// document. Used on every workspace switch.
func (b *MenuBar) SetActiveTab(id string) {
	defer tracing.NewRegion("MenuBar.SetActiveTab").End()
	b.activeTabID = id
	if b.doc == nil {
		return
	}
	tabs := b.doc.GetElementsByGroup("tabs")
	for i := range tabs {
		if tabs[i].Attribute("data-workspace-id") == id {
			b.doc.SetElementClassesWithoutApply(tabs[i], "workspaceTab", "edPanelBgHoverable", "tabSelected")
		} else {
			b.doc.SetElementClassesWithoutApply(tabs[i], "workspaceTab", "edPanelBgHoverable")
		}
	}
	b.doc.ApplyStyles()
}

func (b *MenuBar) renderDocument() error {
	defer tracing.NewRegion("MenuBar.renderDocument").End()
	if b.doc != nil {
		b.doc.Destroy()
		b.doc = nil
	}
	tplData := menuBarTemplateData{
		ShowGrid:      b.handler.Settings().ShowGrid,
		WorkspaceTabs: b.tabs,
		ActiveTabID:   b.activeTabID,
	}
	doc, err := markup.DocumentFromHTMLAsset(&b.uiMan, "editor/ui/global/menu_bar.go.html",
		tplData, map[string]func(*document.Element){
			"clickLogo":                b.openMenuTarget,
			"clickFile":                b.openMenuTarget,
			"clickEdit":                b.openMenuTarget,
			"clickCreate":              b.openMenuTarget,
			"clickView":                b.openMenuTarget,
			"clickHelp":                b.openMenuTarget,
			"clickToggleGrid":          b.clickToggleGrid,
			"clickScreenshot":          b.clickScreenshot,
			"clickWorkspace":           b.clickWorkspace,
			"clickNewStage":            b.clickNewStage,
			"clickOpenStage":           b.clickOpenStage,
			"clickSaveStage":           b.clickSaveStage,
			"clickOpenCodeEditor":      b.clickOpenCodeEditor,
			"clickBuild":               b.clickBuild,
			"clickBuildAndRun":         b.clickBuildAndRun,
			"clickBuildRelease":        b.clickBuildRelease,
			"clickBuildAndRunRelease":  b.clickBuildAndRunRelease,
			"clickRunCurrentStage":     b.clickRunCurrentStage,
			"clickBuildAndroid":        b.clickBuildAndroid,
			"clickBuildRunAndroid":     b.clickBuildRunAndroid,
			"clickExportProject":       b.clickExportProject,
			"clickUndo":                b.clickUndo,
			"clickRedo":                b.clickRedo,
			"clickDuplicate":           b.clickDuplicate,
			"clickCreateTemplate":      b.clickCreateTemplate,
			"clickCreateEntityData":    b.clickCreateEntityData,
			"clickCreateHtmlUi":        b.clickCreateHtmlUi,
			"clickCreateCssStylesheet": b.clickCreateCssStylesheet,
			"clickDistanceChain":       b.clickDistanceChain,
			"clickRope":                b.clickRope,
			"clickHingeChain":          b.clickHingeChain,
			"clickNewCamera":           b.clickNewCamera,
			"clickNewEntity":           b.clickNewEntity,
			"clickNewLight":            b.clickNewLight,
			"clickCreateSphere":        b.clickCreateSphere,
			"clickCreateCube":          b.clickCreateCube,
			"clickCreateCapsule":       b.clickCreateCapsule,
			"clickCreatePlane":         b.clickCreatePlane,
			"clickCreateCylinder":      b.clickCreateCylinder,
			"clickCreateCone":          b.clickCreateCone,
			"clickCreateArrow":         b.clickCreateArrow,
			"clickAbout":               b.clickAbout,
			"clickLogs":                b.clickLogs,
			"clickIssues":              b.clickIssues,
			"clickRepository":          b.clickRepository,
			"clickJoinMailingList":     b.clickJoinMailingList,
			"clickMailArchives":        b.clickMailArchives,
			"clickCreatePluginProject": b.clickCreatePluginProject,
			"clickCloseEditor":         b.clickCloseEditor,
			"clickSponsors":            b.clickSponsors,
			"popupMiss":                b.popupMiss,
		})
	if err != nil {
		return err
	}
	b.doc = doc
	b.doc.Clean()
	for _, m := range b.doc.GetElementsByClass("menuEntry") {
		target := m.Attribute("data-target")
		pop, _ := b.doc.GetElementById(target)
		b.setPopupUiPos(m, pop)
	}
	b.hidePopups()
	return nil
}

func (b *MenuBar) Focus() { b.uiMan.EnableUpdate() }
func (b *MenuBar) Blur()  { b.uiMan.DisableUpdate() }

func (b *MenuBar) IsFocusedOnInput() bool {
	return b.uiMan.Group.IsFocusedOnInput()
}

func (b *MenuBar) setPopupUiPos(e *document.Element, pop *document.Element) {
	defer tracing.NewRegion("MenuBar.setPopupUiPos").End()
	t := &e.UI.Entity().Transform
	x := t.WorldPosition().X() + float32(b.uiMan.Host.Window.Width())*0.5 -
		e.UI.Layout().PixelSize().X()*0.5
	b.doc.SetElementStylePropertyWithoutApply(pop, "left", fmt.Sprintf("%dpx", int(matrix.Round(x))))
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
		b.handler.BlurInterface()
		b.uiMan.Host.RunOnMainThread(b.Focus)
	}
}

// clickWorkspace is the single shared click handler for all workspace tabs.
// It reads the workspace id from the data-workspace-id attribute and asks the
// editor to switch. The editor's setWorkspaceState in turn calls SetActiveTab
// to keep the visual state in sync.
func (b *MenuBar) clickWorkspace(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickWorkspace").End()
	id := e.Attribute("data-workspace-id")
	if id == "" {
		return
	}
	b.handler.WorkspaceSelected(id)
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

func (b *MenuBar) clickUndo(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickUndo").End()
	b.handler.History().Undo()
}

func (b *MenuBar) clickRedo(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickRedo").End()
	b.handler.History().Redo()
}

func (b *MenuBar) clickDuplicate(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickDuplicate").End()
	b.hidePopups()
	b.handler.StageView().DuplicateSelected(b.handler.Project())
}

func (b *MenuBar) clickCreateTemplate(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateTemplate").End()
	b.hidePopups()
	b.handler.StageView().Manager().CreateTemplateFromSelected(b.handler.Events(), b.handler.Project())
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
func (b *MenuBar) clickCreateCssStylesheet(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateCssStylesheet").End()
	b.hidePopups()
	b.handler.BlurInterface()
	input_prompt.Show(b.uiMan.Host, input_prompt.Config{
		Title:       "Name your CSS file",
		Description: "Give a friendly name to your css file",
		Placeholder: "Name...",
		ConfirmText: "Create",
		CancelText:  "Cancel",
		OnCancel:    b.handler.FocusInterface,
		OnConfirm: func(name string) {
			b.handler.FocusInterface()
			b.handler.CreateCssStylesheetFile(name)
		},
	})
}

func (b *MenuBar) clickDistanceChain(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickDistanceChain").End()
	b.hidePopups()
	b.handler.ConnectSelectedAsDistanceChain()
}

func (b *MenuBar) clickRope(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickRope").End()
	b.hidePopups()
	b.handler.ConnectSelectedAsRope()
}

func (b *MenuBar) clickHingeChain(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickHingeChain").End()
	b.hidePopups()
	b.handler.ConnectSelectedAsHingeChain()
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
	go b.handler.Project().BuildAndroid(b.uiMan.Host.AssetDatabase(),
		bts.AndroidNDK, bts.JavaHome, []string{"debug"})
}

func (b *MenuBar) clickBuildRunAndroid(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickBuildRunAndroid").End()
	b.hidePopups()
	bts := b.handler.Settings().BuildTools
	// goroutine
	go b.handler.Project().BuildRunAndroid(b.uiMan.Host.AssetDatabase(),
		bts.AndroidNDK, bts.JavaHome, []string{"debug"})
}

func (b *MenuBar) clickExportProject(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickExportProject").End()
	b.hidePopups()
	b.handler.BlurInterface()
	file_browser.Show(b.uiMan.Host, file_browser.Config{
		Title:        "Export project to...",
		StartingPath: b.handler.ProjectFileSystem().FullPath(""),
		OnlyFolders:  true,
		OnConfirm: func(paths []string) {
			path := paths[0]
			name := "KaijuTemplate.zip"
			if _, err := os.Stat(filepath.Join(path, name)); err == nil {
				i := 0
				f := `KaijuTemplate (%d).zip`
				for {
					name = fmt.Sprintf(f, i)
					_, err := os.Stat(filepath.Join(path, fmt.Sprintf(name, i)))
					if err != nil {
						break
					}
				}
			}
			b.handler.FocusInterface()
			go b.handler.Project().ExportAsTemplate(path, name)
		},
		OnCancel: b.handler.FocusInterface,
	})
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

func (b *MenuBar) clickCreateSphere(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateSphere").End()
	b.createPrimitive(rendering.PrimitiveMeshSphere)
}

func (b *MenuBar) clickCreateCube(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateCube").End()
	b.createPrimitive(rendering.PrimitiveMeshTexturableCube)
}

func (b *MenuBar) clickCreateCapsule(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateCapsule").End()
	b.createPrimitive(rendering.PrimitiveMeshCapsule)
}

func (b *MenuBar) clickCreatePlane(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreatePlane").End()
	b.createPrimitive(rendering.PrimitiveMeshPlane)
}

func (b *MenuBar) clickCreateCylinder(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateCylinder").End()
	b.createPrimitive(rendering.PrimitiveMeshCylinder)
}

func (b *MenuBar) clickCreateCone(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateCone").End()
	b.createPrimitive(rendering.PrimitiveMeshCone)
}

func (b *MenuBar) clickCreateArrow(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCreateArrow").End()
	b.createPrimitive(rendering.PrimitiveMeshArrow)
}

func (b *MenuBar) createPrimitive(primitive rendering.PrimitiveMesh) {
	b.hidePopups()
	b.handler.CreatePrimitive(primitive)
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

func (b *MenuBar) clickCreatePluginProject(*document.Element) {
	const pathA = "editor/editor_plugin/developer_plugins"
	const pathB = "src/" + pathA
	defer tracing.NewRegion("MenuBar.clickCreatePluginProject").End()
	b.hidePopups()
	b.handler.BlurInterface()
	exePath, _ := os.Executable()
	startPaths := [...]string{
		filepath.Join(filepath.Dir(exePath), pathB),
		filepath.Join(filepath.Dir(exePath), pathA),
	}
	for i := range startPaths {
		if s, err := os.Stat(startPaths[i]); err == nil && s.IsDir() {
			exePath = startPaths[i]
			break
		}
	}
	file_browser.Show(b.uiMan.Host, file_browser.Config{
		Title:        "Select plugin project path",
		StartingPath: exePath,
		OnlyFolders:  true,
		OnConfirm: func(paths []string) {
			b.handler.FocusInterface()
			b.handler.CreatePluginProject(paths[0])
		},
		OnCancel: b.handler.FocusInterface,
	})
}

func (b *MenuBar) clickCloseEditor(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickCloseEditor").End()
	b.hidePopups()
	b.uiMan.Host.Close()
}

func (b *MenuBar) clickSponsors(*document.Element) {
	defer tracing.NewRegion("MenuBar.clickSupporters").End()
	b.hidePopups()
	b.handler.BlurInterface()
	sponsors.Show(b.uiMan.Host, b.handler.FocusInterface)
}

func (b *MenuBar) clickToggleGrid(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickToggleGrid").End()
	visible := !b.handler.Settings().ShowGrid
	b.handler.SetGridVisible(visible)
	if lbl := e.InnerLabel(); lbl != nil {
		if visible {
			lbl.SetText("Hide Grid")
		} else {
			lbl.SetText("Show Grid")
		}
	}
	b.hidePopups()
}

func (b *MenuBar) clickScreenshot(e *document.Element) {
	defer tracing.NewRegion("MenuBar.clickScreenshot").End()
	b.hidePopups()
	host := e.UI.Host()
	host.RunNextFrame(func() {
		device := host.Window.GpuHost.FirstInstance().PrimaryDevice()
		pixels, err := device.Screenshot()
		if err != nil {
			slog.Error("Failed to capture the screenshot", "error", err)
			return
		}
		if len(pixels) == 0 {
			slog.Error("No pixels were returned for the frame")
			return
		}
		size := device.LogicalDevice.SwapChain.Extent
		img := image.NewRGBA(image.Rect(0, 0, int(size.X()), int(size.Y())))
		copy(img.Pix, pixels)
		var buf bytes.Buffer
		if err = png.Encode(&buf, img); err != nil {
			slog.Error("Failed to encode the png file", "error", err)
			return
		}
		fs := b.handler.Project().FileSystem()
		fs.Mkdir("screenshots", os.ModePerm)
		path := fmt.Sprintf("screenshots/%s.png", time.Now().Format("2006-01-02-03-04-05"))
		if err := fs.WriteFile(path, buf.Bytes(), os.ModePerm); err != nil {
			slog.Error("Failed to write the screenshot file", "error", err)
			return
		}
		slog.Info("Screenshot captured", "path", path)
	})
}

func (b *MenuBar) popupMiss(e *document.Element) {
	defer tracing.NewRegion("MenuBar.popupMiss").End()
	if e == b.selectedPopup {
		b.hidePopups()
	}
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

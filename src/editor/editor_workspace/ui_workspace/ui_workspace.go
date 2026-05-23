/******************************************************************************/
/* ui_workspace.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui_workspace

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"kaijuengine.com/editor/editor_overlay/file_browser"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	ID             = "ui"
	DisplayName    = "UI"
	updateInterval = 1.0
)

func init() {
	editor_workspace_registry.Register(&UIWorkspace{})
}

type UIWorkspace struct {
	common_workspace.CommonWorkspace
	ed            editor_workspace.WorkspaceEditorInterface
	previewDoc    *document.Document
	previewMan    ui.Manager
	editBtn       *document.Element
	previewArea   *document.Element
	previewHelp   *document.Element
	ratioX        *document.Element
	ratioY        *document.Element
	html          string
	data          string
	styles        []string
	bindingData   any
	lastMod       time.Time
	lastTime      float64
	ratio         matrix.Vec2
	openHtmlSubID events.Id
}

func (w *UIWorkspace) ID() string          { return ID }
func (w *UIWorkspace) DisplayName() string { return DisplayName }
func (w *UIWorkspace) IsRequired() bool    { return false }

func (w *UIWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("UIWorkspace.Initialize").End()
	host := ed.Host()
	w.ed = ed
	w.ratio = matrix.NewVec2(16, 9)
	if err := w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/ui_workspace.go.html", w.ratio, map[string]func(*document.Element){
			"clickFile":                   w.clickFile,
			"clickEdit":                   w.clickEdit,
			"clickLoadData":               w.clickLoadData,
			"changeWidthRatio":            w.changeWidthRatio,
			"changeHeightRatio":           w.changeHeightRatio,
			"toggleTextareaSmokeStyle":    w.toggleTextareaSmokeStyle,
			"duplicateTextareaSmokeField": w.duplicateTextareaSmokeField,
		}); err != nil {
		return err
	}
	w.editBtn, _ = w.Doc.GetElementById("editBtn")
	w.previewArea, _ = w.Doc.GetElementById("previewArea")
	w.previewHelp, _ = w.Doc.GetElementById("previewHelp")
	w.ratioX, _ = w.Doc.GetElementById("ratioX")
	w.ratioY, _ = w.Doc.GetElementById("ratioY")
	w.previewMan.Init(host)
	w.previewArea.UIPanel.DontFitContent()
	w.openHtmlSubID = ed.Events().OnRequestViewHtmlUi.Add(func(htmlID string) {
		ed.SelectWorkspace(ID)
		w.OpenHtml(htmlID)
	})
	return nil
}

func (w *UIWorkspace) Shutdown() {
	defer tracing.NewRegion("UIWorkspace.Shutdown").End()
	if w.ed != nil {
		w.ed.Events().OnRequestViewHtmlUi.Remove(w.openHtmlSubID)
	}
	w.CommonShutdown()
}

func (w *UIWorkspace) Open() {
	defer tracing.NewRegion("UIWorkspace.Open").End()
	w.CommonOpen()
	w.applyRatio()
	if w.html != "" {
		w.previewHelp.UI.Hide()
	}
}

func (w *UIWorkspace) Close() {
	defer tracing.NewRegion("UIWorkspace.Close").End()
	w.CommonClose()
}

func (w *UIWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *UIWorkspace) clickFile(e *document.Element) {
	w.ed.BlurInterface()
	file_browser.Show(w.Host, file_browser.Config{
		Title:        "Load HTML file",
		StartingPath: w.ed.ProjectFileSystem().FullPath(""),
		ExtFilter:    []string{".html"},
		OnlyFiles:    true,
		OnCancel:     w.ed.FocusInterface,
		OnConfirm: func(paths []string) {
			w.ed.FocusInterface()
			if paths[0] != "" {
				w.OpenHtml(paths[0])
			}
		},
	})
}

func (w *UIWorkspace) clickEdit(e *document.Element) {
	if w.html == "" {
		return
	}
	path := project_file_system.HtmlPath(w.html)
	pfs := w.ed.ProjectFileSystem()
	exec.Command("code", pfs.FullPath(""), pfs.FullPath(path.String())).Run()
}

func (w *UIWorkspace) clickLoadData(e *document.Element) {
	if w.html == "" {
		return
	}
	w.ed.BlurInterface()
	file_browser.Show(w.Host, file_browser.Config{
		Title:        "Load HTML mock data",
		StartingPath: w.ed.ProjectFileSystem().FullPath(""),
		ExtFilter:    []string{".json"},
		OnCancel:     w.ed.FocusInterface,
		OnlyFiles:    true,
		OnConfirm: func(paths []string) {
			w.ed.FocusInterface()
			w.data = paths[0]
			w.bindingData = loadBindingData(paths[0])
			w.OpenHtml(w.html)
		},
	})
}

func (w *UIWorkspace) changeWidthRatio(*document.Element)  { w.readRatio() }
func (w *UIWorkspace) changeHeightRatio(*document.Element) { w.readRatio() }

func (w *UIWorkspace) toggleTextareaSmokeStyle(e *document.Element) {
	target := e.Parent.Value()
	for target != nil && !target.HasClass("textareaSmoke") {
		target = target.Parent.Value()
	}
	if target == nil {
		return
	}
	if target.HasClass("textareaSmokeAlt") {
		w.Doc.SetElementClasses(target, "textareaSmoke")
	} else {
		w.Doc.SetElementClasses(target, "textareaSmoke", "textareaSmokeAlt")
	}
}

func (w *UIWorkspace) duplicateTextareaSmokeField(*document.Element) {
	template, ok := w.Doc.GetElementById("textareaSmokeTemplate")
	if !ok {
		return
	}
	w.Doc.DuplicateElement(template)
	w.Doc.SetupInputTabIndexs()
}

func (w *UIWorkspace) RunTextAreaSmokeTest() error {
	template, ok := w.Doc.GetElementById("textareaSmokeTemplate")
	if !ok {
		return fmt.Errorf("textarea smoke template missing")
	}
	inputs := template.FindElementsByTag("input")
	if len(inputs) < 2 {
		return fmt.Errorf("textarea smoke template expected two tab-order inputs, got %d", len(inputs))
	}
	textareaElement := template.FindElementByTag("textarea")
	if textareaElement == nil {
		return fmt.Errorf("textarea smoke template missing textarea")
	}
	textarea := textareaElement.UI.ToTextArea()
	if textarea.Text() == "" || !strings.Contains(textarea.Text(), "\n") {
		return fmt.Errorf("textarea smoke sample should start with multi-line text")
	}
	inputs[0].UI.ToInput().Focus()
	if !w.IsFocusedOnInput() {
		return fmt.Errorf("input focus was not reported to the editor workspace")
	}
	inputs[0].UI.ToInput().RemoveFocus()
	textarea.Focus()
	if !w.IsFocusedOnInput() {
		return fmt.Errorf("textarea focus was not reported to the editor workspace")
	}
	textarea.SetText(strings.Join([]string{
		"Smoke paragraph one: real editor markup path.",
		"Smoke paragraph two: enough wrapped content to scroll.",
		"Smoke paragraph three: duplicate and style changes stay valid.",
	}, "\n\n"))
	textarea.SetCursorOffset(len([]rune(textarea.Text())))
	textarea.RemoveFocus()

	var toggleButton *document.Element
	for _, button := range template.FindElementsByTag("button") {
		if label := button.InnerLabel(); label != nil && label.Text() == "Toggle style" {
			toggleButton = button
			break
		}
	}
	if toggleButton == nil {
		return fmt.Errorf("textarea smoke style button missing")
	}
	w.toggleTextareaSmokeStyle(toggleButton)
	if !template.HasClass("textareaSmokeAlt") {
		return fmt.Errorf("textarea smoke style toggle did not apply alternate class")
	}
	before := len(w.Doc.GetElementsByTagName("textarea"))
	w.duplicateTextareaSmokeField(nil)
	after := len(w.Doc.GetElementsByTagName("textarea"))
	if after != before+1 {
		return fmt.Errorf("textarea smoke duplicate expected %d textareas, got %d", before+1, after)
	}
	w.applyRatio()
	return nil
}

func (w *UIWorkspace) readRatio() {
	if r, err := strconv.ParseFloat(w.ratioX.UI.ToInput().Text(), 64); err == nil {
		w.ratio.SetX(float32(r))
		w.applyRatio()
	}
	if r, err := strconv.ParseFloat(w.ratioY.UI.ToInput().Text(), 64); err == nil {
		w.ratio.SetY(float32(r))
		w.applyRatio()
	}
}

func (w *UIWorkspace) applyRatio() {
	ww := matrix.Float(w.Host.Window.Width())
	wh := matrix.Float(w.Host.Window.Height())
	top := float32(24)
	bottom := helpers.NumFromLength("3.3em", w.Host.Window)
	drawArea := matrix.NewVec4(0, top, ww, wh-top-bottom)
	drawW := drawArea.Z()
	drawH := drawArea.W()
	if w.ratio.X() <= 0 && w.ratio.Y() <= 0 {
		w.previewArea.UI.Layout().Scale(drawW, drawH)
		return
	}
	r := w.ratio
	if w.ratio.X() <= 0 {
		r.SetX(r.Y())
	}
	if w.ratio.Y() <= 0 {
		r.SetY(r.X())
	}
	scaleW := drawW / r.X()
	scaleH := drawH / r.Y()
	scale := matrix.Min(scaleW, scaleH)
	targetWidth := r.X() * scale
	targetHeight := r.Y() * scale
	w.previewArea.UI.Layout().Scale(targetWidth, targetHeight)
}

func (w *UIWorkspace) Update(deltaTime float64) {
	if !w.Doc.IsActive() {
		return
	}
	w.lastTime -= deltaTime
	if w.lastTime <= 0 {
		w.processFilesChanges()
		w.lastTime = updateInterval
	}
}

func (w *UIWorkspace) processFilesChanges() {
	pfs := w.ed.ProjectFileSystem()
	htmlChanged := false
	if s, err := pfs.Stat(project_file_system.HtmlPath(w.html).String()); err == nil && s.ModTime().After(w.lastMod) {
		htmlChanged = true
	}
	for f := 0; f < len(w.styles) && !htmlChanged; f++ {
		if s, e := os.Stat(w.styles[f]); e == nil && s.ModTime().After(w.lastMod) {
			htmlChanged = true
		}
	}
	if s, err := os.Stat(w.data); err == nil && s.ModTime().After(w.lastMod) {
		w.bindingData = loadBindingData(w.data)
		htmlChanged = true
	}
	if htmlChanged {
		w.OpenHtml(w.html)
		w.lastMod = time.Now()
	}
}

func (w *UIWorkspace) pullStyles() {
	w.styles = w.styles[:0]
	for i := range w.Doc.HeadElements {
		if w.Doc.HeadElements[i].Data == "link" {
			if w.Doc.HeadElements[i].Attribute("rel") == "stylesheet" {
				cssPath := w.Doc.HeadElements[i].Attribute("href")
				w.styles = append(w.styles, cssPath)
			}
		}
	}
}

func (w *UIWorkspace) OpenHtml(html string) {
	if html == "" {
		return
	}
	w.previewHelp.UI.Hide()
	w.html = html
	if w.previewDoc != nil {
		w.previewDoc.Destroy()
		w.previewDoc = nil
	}
	w.Host.RunOnMainThread(func() {
		if doc, err := markup.DocumentFromHTMLAssetRooted(&w.previewMan,
			w.html, w.bindingData, nil, w.previewArea); err == nil {
			w.previewDoc = doc
			w.pullStyles()
		} else {
			slog.Error("failed to load the html file", "error", err)
		}
	})
	w.lastMod = time.Now()
}

func loadBindingData(bindingFile string) any {
	if _, err := os.Stat(bindingFile); err != nil {
		slog.Error("failed to load the data file", "file", bindingFile, "error", err)
		return nil
	}
	bindingData, err := filesystem.ReadTextFile(bindingFile)
	if err != nil {
		return nil
	}
	var out any
	err = klib.JsonDecode(json.NewDecoder(strings.NewReader(bindingData)), &out)
	if err != nil {
		return nil
	}
	return out
}

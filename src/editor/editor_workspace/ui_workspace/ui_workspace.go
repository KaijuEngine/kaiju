/******************************************************************************/
/* ui_workspace.go                                                            */
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

package ui_workspace

import (
	"encoding/json"
	"kaiju/editor/editor_overlay/file_browser"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/css/helpers"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const updateInterval = 1.0

type UIWorkspace struct {
	common_workspace.CommonWorkspace
	ed          UIWorkspaceEditorInterface
	previewDoc  *document.Document
	previewMan  ui.Manager
	editBtn     *document.Element
	previewArea *document.Element
	previewHelp *document.Element
	ratioX      *document.Element
	ratioY      *document.Element
	updateId    engine.UpdateId
	html        string
	data        string
	styles      []string
	bindingData any
	lastMod     time.Time
	lastTime    float64
	ratio       matrix.Vec2
}

func (w *UIWorkspace) Initialize(host *engine.Host, ed UIWorkspaceEditorInterface) {
	defer tracing.NewRegion("UIWorkspace.Initialize").End()
	w.ed = ed
	w.ratio = matrix.NewVec2(16, 9)
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/ui_workspace.go.html", w.ratio, map[string]func(*document.Element){
			"clickEdit":         w.clickEdit,
			"clickLoadData":     w.clickLoadData,
			"changeWidthRatio":  w.changeWidthRatio,
			"changeHeightRatio": w.changeHeightRatio,
		})
	w.editBtn, _ = w.Doc.GetElementById("editBtn")
	w.previewArea, _ = w.Doc.GetElementById("previewArea")
	w.previewHelp, _ = w.Doc.GetElementById("previewHelp")
	w.ratioX, _ = w.Doc.GetElementById("ratioX")
	w.ratioY, _ = w.Doc.GetElementById("ratioY")
	w.previewMan.Init(host)
	w.updateId = w.Host.Updater.AddUpdate(w.update)
	w.previewArea.UIPanel.DontFitContent()
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
		r.SetX(r.X())
	}
	scaleW := drawW / r.X()
	scaleH := drawH / r.Y()
	scale := matrix.Min(scaleW, scaleH)
	targetWidth := r.X() * scale
	targetHeight := r.Y() * scale
	w.previewArea.UI.Layout().Scale(targetWidth, targetHeight)
}

func (w *UIWorkspace) update(deltaTime float64) {
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

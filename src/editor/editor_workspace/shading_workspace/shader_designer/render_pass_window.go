/******************************************************************************/
/* render_pass_window.go                                                      */
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

package shader_designer

import (
	"encoding/json"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/KaijuEngine/uuid"
)

func (win *ShaderDesigner) reloadRenderPassDoc() {
	sy := float32(0)
	if win.renderPassDoc != nil {
		content := win.renderPassDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.renderPassDoc.Destroy()
	}
	data := reflectUIStructure(&win.renderPass.RenderPassData, "", map[string][]ui.SelectOption{})
	data.Name = "Render Pass Editor"
	win.renderPassDoc, _ = markup.DocumentFromHTMLAsset(win.uiMan, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":     showRenderPassTooltip,
			"valueChanged":    win.renderPassValueChanged,
			"nameChanged":     win.renderPassNameChanged,
			"addToSlice":      win.renderPassAddToSlice,
			"removeFromSlice": win.renderPassRemoveFromSlice,
			"saveData":        win.renderPassSaveRenderPass,
		})
	if sy != 0 {
		content := win.renderPassDoc.GetElementsByClass("topFields")[0]
		win.uiMan.Host.RunAfterFrames(2, func() {
			content.UIPanel.SetScrollY(sy)
		})
	}
}

func (win *ShaderDesigner) renderPassValueChanged(e *document.Element) {
	setObjectValueFromUI(&win.renderPass, e)
}

func (win *ShaderDesigner) renderPassNameChanged(e *document.Element) {
	win.renderPass.Name = e.UI.ToInput().Text()
}

func (win *ShaderDesigner) renderPassAddToSlice(e *document.Element) {
	reflectAddToSlice(&win.renderPass, e)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassRemoveFromSlice(e *document.Element) {
	reflectRemoveFromSlice(&win.renderPass, e)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassSaveRenderPass(e *document.Element) {
	if win.renderPass.id == "" {
		win.renderPass.id = uuid.NewString()
	}
	for i := range win.renderPass.SubpassDescriptions {
		s := win.renderPass.SubpassDescriptions[i]
		s.Subpass.Shader = filepath.ToSlash(s.Subpass.Shader)
		s.Subpass.ShaderPipeline = filepath.ToSlash(s.Subpass.ShaderPipeline)
	}
	res, err := json.Marshal(win.renderPass)
	if err != nil {
		slog.Error("failed to marshal the render pass data", "error", err)
		return
	}
	err = win.pfs.WriteFile(filepath.Join(project_file_system.ContentFolder,
		project_file_system.ContentRenderPassFolder, win.renderPass.id), res, os.ModePerm)
	if err != nil {
		slog.Error("failed to write the render pass data to file", "error", err)
		return
	}
	slog.Info("render pass successfully saved")
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}

func showRenderPassTooltip(e *document.Element) {
	id := e.Attribute("data-tooltip")
	tip, ok := renderPassTooltips[id]
	if !ok {
		return
	}
	tipElm := e.Root().FindElementById("ToolTip")
	if tipElm == nil || len(tipElm.Children) == 0 {
		return
	}
	lbl := tipElm.Children[0].UI
	if !lbl.IsType(ui.ElementTypeLabel) {
		return
	}
	lbl.ToLabel().SetText(tip)
}

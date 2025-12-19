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

package shader_designer

import (
	"encoding/json"
	"errors"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"log/slog"
	"os"
	"path/filepath"
)

func (win *ShaderDesigner) reloadRenderPassDoc() {
	sy := float32(0)
	if win.renderPassDoc != nil {
		content := win.renderPassDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.renderPassDoc.Destroy()
	}
	data := common_workspace.ReflectUIStructure(&win.renderPass.RenderPassData, "", map[string][]ui.SelectOption{})
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
	common_workspace.SetObjectValueFromUI(&win.renderPass, e)
}

func (win *ShaderDesigner) renderPassNameChanged(e *document.Element) {
	win.renderPass.Name = e.UI.ToInput().Text()
}

func (win *ShaderDesigner) renderPassAddToSlice(e *document.Element) {
	common_workspace.ReflectAddToSlice(&win.renderPass, e)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassRemoveFromSlice(e *document.Element) {
	common_workspace.ReflectRemoveFromSlice(&win.renderPass, e)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassSaveRenderPass(e *document.Element) {
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
	if win.renderPass.id != "" {
		err = win.ed.ProjectFileSystem().WriteFile(filepath.Join(project_file_system.ContentFolder,
			project_file_system.ContentRenderPassFolder, win.renderPass.id), res, os.ModePerm)
	} else {
		ids := content_database.ImportRaw(win.shader.Name, res, content_database.RenderPass{}, win.ed.ProjectFileSystem(), win.ed.Cache())
		if len(ids) > 0 {
			win.renderPass.id = ids[0]
			win.ed.Events().OnContentAdded.Execute(ids)
		} else {
			err = errors.New("failed to import the raw render pass file data to the database")
		}
	}
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
	tipElm := e.Root().FindElementById("toolTip")
	if tipElm == nil || len(tipElm.Children) == 0 {
		return
	}
	lbl := tipElm.Children[0].UI
	if !lbl.IsType(ui.ElementTypeLabel) {
		return
	}
	lbl.ToLabel().SetText(tip)
}

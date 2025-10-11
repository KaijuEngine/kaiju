/******************************************************************************/
/* shader_pipeline_window.go                                                  */
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
	"kaiju/engine/systems/logging"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/rendering"
	"log/slog"
	"os"
	"path/filepath"
)

func setupPipelineDoc(win *ShaderDesigner) {
	win.reloadPipelineDoc()
	win.pipelineDoc.Deactivate()
}

func (win *ShaderDesigner) reloadPipelineDoc() {
	sy := float32(0)
	if win.pipelineDoc != nil {
		content := win.pipelineDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.pipelineDoc.Destroy()
	}
	data := reflectUIStructure(&win.pipeline, "", map[string][]string{})
	data.Name = "Shader Pipeline Editor"
	win.pipelineDoc, _ = markup.DocumentFromHTMLAsset(&win.man, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":     showPipelineTooltip,
			"valueChanged":    win.pipelineValueChanged,
			"returnHome":      win.returnHome,
			"addToSlice":      win.pipelineAddToSlice,
			"removeFromSlice": win.pipelineRemoveFromSlice,
			"saveData":        win.pipelineSave,
		})
	if sy != 0 {
		content := win.pipelineDoc.GetElementsByClass("topFields")[0]
		win.man.Host.RunAfterFrames(2, func() {
			content.UIPanel.SetScrollY(sy)
		})
	}
}

func showPipelineTooltip(e *document.Element) {
	id := e.Attribute("data-tooltip")
	tip, ok := pipelineTooltips[id]
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

func (win *ShaderDesigner) pipelineAddToSlice(e *document.Element) {
	reflectAddToSlice(&win.pipeline, e)
	win.reloadPipelineDoc()
}

func (win *ShaderDesigner) pipelineRemoveFromSlice(e *document.Element) {
	reflectRemoveFromSlice(&win.pipeline, e)
	win.reloadPipelineDoc()
}

func (win *ShaderDesigner) pipelineValueChanged(e *document.Element) {
	setObjectValueFromUI(&win.pipeline, e)
}

func OpenPipeline(path string, logStream *logging.LogStream) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to load the shader pipeline file", "file", path, "error", err)
		return
	}
	var p rendering.ShaderPipelineData
	if err := json.Unmarshal(data, &p); err != nil {
		slog.Error("failed to unmarshal the shader pipeline data", "error", err)
		return
	}
	newInternal(StateRenderPass, logStream, func(sd *ShaderDesigner) {
		sd.pipeline = p
		sd.ShowPipelineWindow()
	})
}

func (win *ShaderDesigner) pipelineSave(e *document.Element) {
	const saveRoot = pipelineFolder
	if err := os.MkdirAll(saveRoot, os.ModePerm); err != nil {
		slog.Error("failed to create the shader pipeline folder",
			"folder", saveRoot, "error", err)
	}
	path := filepath.Join(saveRoot, win.pipeline.Name+fileExtensionShaderPipeline)
	res, err := json.Marshal(win.pipeline)
	if err != nil {
		slog.Error("failed to marshal the pipeline data", "error", err)
		return
	}
	if err := os.WriteFile(path, res, os.ModePerm); err != nil {
		slog.Error("failed to write the pipeline data to file", "error", err)
		return
	}
	slog.Info("shader pipeline successfully saved", "file", path)
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}

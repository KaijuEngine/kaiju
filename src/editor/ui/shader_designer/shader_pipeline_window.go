package shader_designer

import (
	"encoding/json"
	"kaiju/editor/alert"
	"kaiju/editor/editor_config"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
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

func OpenPipeline(path string) {
	setup(func(win *ShaderDesigner) {
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Error("failed to load the shader pipeline file", "file", path, "error", err)
			return
		}
		if err := json.Unmarshal(data, &win.pipeline); err != nil {
			slog.Error("failed to unmarshal the shader pipeline data", "error", err)
			return
		}
		win.ShowPipelineWindow()
	})
}

func (win *ShaderDesigner) pipelineSave(e *document.Element) {
	const saveRoot = pipelineFolder
	if err := os.MkdirAll(saveRoot, os.ModePerm); err != nil {
		slog.Error("failed to create the shader pipeline folder",
			"folder", saveRoot, "error", err)
	}
	path := filepath.Join(saveRoot, win.pipeline.Name+editor_config.FileExtensionShaderPipeline)
	if _, err := os.Stat(path); err == nil {
		ok := <-alert.New("Overwrite?", "You are about to overwrite a shader pipeline with the same name, would you like to continue?", "Yes", "No", win.man.Host)
		if !ok {
			return
		}
	}
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

/******************************************************************************/
/* shader_pipeline_window.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_designer

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func (win *ShaderDesigner) reloadPipelineDoc() {
	sy := matrix.Float(0)
	if win.pipelineDoc != nil {
		content := win.pipelineDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.pipelineDoc.Destroy()
	}
	data := common_workspace.ReflectUIStructure(win.ed.Cache(),
		&win.pipeline.ShaderPipelineData, "", map[string][]ui.SelectOption{})
	data.Name = "Shader Pipeline Editor"
	win.pipelineDoc, _ = markup.DocumentFromHTMLAsset(win.uiMan, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":     showPipelineTooltip,
			"valueChanged":    win.pipelineValueChanged,
			"addToSlice":      win.pipelineAddToSlice,
			"removeFromSlice": win.pipelineRemoveFromSlice,
			"saveData":        win.pipelineSave,
		})
	if sy != 0 {
		content := win.pipelineDoc.GetElementsByClass("topFields")[0]
		win.uiMan.Host.RunAfterFrames(2, func() {
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

func (win *ShaderDesigner) pipelineAddToSlice(e *document.Element) {
	common_workspace.ReflectAddToSlice(&win.pipeline, e)
	win.reloadPipelineDoc()
}

func (win *ShaderDesigner) pipelineRemoveFromSlice(e *document.Element) {
	common_workspace.ReflectRemoveFromSlice(&win.pipeline, e)
	win.reloadPipelineDoc()
}

func (win *ShaderDesigner) pipelineValueChanged(e *document.Element) {
	common_workspace.SetObjectValueFromUI(&win.pipeline, e)
}

func (win *ShaderDesigner) pipelineSave(e *document.Element) {
	res, err := json.Marshal(win.pipeline)
	if err != nil {
		slog.Error("failed to marshal the pipeline data", "error", err)
		return
	}
	if win.pipeline.id != "" {
		err = win.ed.ProjectFileSystem().WriteFile(filepath.Join(project_file_system.ContentFolder,
			project_file_system.ContentShaderPipelineFolder, win.pipeline.id), res, os.ModePerm)
	} else {
		ids := content_database.ImportRaw(win.shader.Name, res, content_database.ShaderPipeline{}, win.ed.ProjectFileSystem(), win.ed.Cache())
		if len(ids) > 0 {
			win.pipeline.id = ids[0]
			win.ed.Events().OnContentAdded.Execute(ids)
		} else {
			err = errors.New("failed to import the raw shader pipeline file data to the database")
		}
	}
	if err != nil {
		slog.Error("failed to write the shader pipeline data to file", "error", err)
		return
	}
	slog.Info("shader pipeline successfully saved")
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}

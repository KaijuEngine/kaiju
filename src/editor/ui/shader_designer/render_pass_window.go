package shader_designer

import (
	"encoding/json"
	"kaiju/editor/alert"
	"kaiju/editor/editor_config"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/systems/logging"
	"kaiju/ui"
	"log/slog"
	"os"
	"path/filepath"
)

func OpenRenderPass(path string, logStream *logging.LogStream) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to load the render pass file", "file", path, "error", err)
		return
	}
	var rp rendering.RenderPassData
	if err := json.Unmarshal(data, &rp); err != nil {
		slog.Error("failed to unmarshal the render pass data", "error", err)
		return
	}
	s := New(StateRenderPass, logStream)
	s.renderPass = rp
	s.ShowRenderPassWindow()
}

func setupRenderPassDoc(win *ShaderDesigner) {
	win.reloadRenderPassDoc()
	win.renderPassDoc.Deactivate()
}

func (win *ShaderDesigner) reloadRenderPassDoc() {
	sy := float32(0)
	if win.renderPassDoc != nil {
		content := win.renderPassDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.renderPassDoc.Destroy()
	}
	data := reflectUIStructure(&win.renderPass, "", map[string][]string{})
	data.Name = "Render Pass Editor"
	win.renderPassDoc, _ = markup.DocumentFromHTMLAssetRooted(win.man, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":     showRenderPassTooltip,
			"valueChanged":    win.renderPassValueChanged,
			"nameChanged":     win.renderPassNameChanged,
			"addToSlice":      win.renderPassAddToSlice,
			"removeFromSlice": win.renderPassRemoveFromSlice,
			"saveData":        win.renderPassSaveRenderPass,
			"returnHome":      win.returnHome,
		}, win.root)
	if sy != 0 {
		content := win.renderPassDoc.GetElementsByClass("topFields")[0]
		win.man.Host.RunAfterFrames(2, func() {
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
	const saveRoot = renderPassFolder
	if err := os.MkdirAll(saveRoot, os.ModePerm); err != nil {
		slog.Error("failed to create the render pass folder",
			"folder", saveRoot, "error", err)
	}
	path := filepath.Join(saveRoot, win.renderPass.Name+editor_config.FileExtensionRenderPass)
	if _, err := os.Stat(path); err == nil {
		ok := <-alert.New("Overwrite?", "You are about to overwrite a render pass with the same name, would you like to continue?", "Yes", "No", win.man.Host)
		if !ok {
			return
		}
	}
	res, err := json.Marshal(win.renderPass)
	if err != nil {
		slog.Error("failed to marshal the render pass data", "error", err)
		return
	}
	if err := os.WriteFile(path, res, os.ModePerm); err != nil {
		slog.Error("failed to write the render pass data to file", "error", err)
		return
	}
	slog.Info("render pass successfully saved", "file", path)
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

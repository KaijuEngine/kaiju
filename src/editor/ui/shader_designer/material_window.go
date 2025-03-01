package shader_designer

import (
	"encoding/json"
	"kaiju/editor/alert"
	"kaiju/editor/editor_config"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
	"log/slog"
	"os"
	"path/filepath"
)

func setupMaterialDoc(win *ShaderDesigner) {
	win.reloadMaterialDoc()
	win.materialDoc.Deactivate()
}

func collectSpecificFileOptions(root, extension string) []string {
	options := []string{}
	if dir, err := os.ReadDir(root); err == nil {
		for i := range dir {
			f := dir[i]
			if f.IsDir() {
				continue
			}
			if filepath.Ext(f.Name()) == extension {
				options = append(options, filepath.Join(root, f.Name()))
			}
		}
	}
	return options
}

func collectShaderOptions() []string {
	return collectSpecificFileOptions(shaderFolder, editor_config.FileExtensionShader)
}

func collectRenderPassOptions() []string {
	return collectSpecificFileOptions(renderPassFolder, editor_config.FileExtensionRenderPass)
}

func collectShaderPipelinesOptions() []string {
	return collectSpecificFileOptions(pipelineFolder, editor_config.FileExtensionShaderPipeline)
}

func (win *ShaderDesigner) reloadMaterialDoc() {
	sy := float32(0)
	if win.materialDoc != nil {
		content := win.materialDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.materialDoc.Destroy()
	}
	listings := map[string][]string{}
	listings["Shader"] = collectShaderOptions()
	listings["RenderPass"] = collectRenderPassOptions()
	listings["ShaderPipeline"] = collectShaderPipelinesOptions()
	data := reflectUIStructure(&win.material, "", listings)
	data.Name = "Material Editor"
	win.materialDoc, _ = markup.DocumentFromHTMLAsset(&win.man, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":     showMaterialTooltip,
			"valueChanged":    win.materialValueChanged,
			"returnHome":      win.returnHome,
			"addToSlice":      win.materialAddToSlice,
			"removeFromSlice": win.materialRemoveFromSlice,
			"saveData":        win.materialSave,
		})
	if sy != 0 {
		content := win.materialDoc.GetElementsByClass("topFields")[0]
		win.man.Host.RunAfterFrames(2, func() {
			content.UIPanel.SetScrollY(sy)
		})
	}
}

func showMaterialTooltip(e *document.Element) { showTooltip(materialTooltips, e) }

func (win *ShaderDesigner) materialAddToSlice(e *document.Element) {
	reflectAddToSlice(&win.material, e)
	win.reloadMaterialDoc()
}

func (win *ShaderDesigner) materialRemoveFromSlice(e *document.Element) {
	reflectRemoveFromSlice(&win.material, e)
	win.reloadMaterialDoc()
}

func (win *ShaderDesigner) materialValueChanged(e *document.Element) {
	setObjectValueFromUI(&win.material, e)
}

func loadMaterialData(path string) (rendering.MaterialData, bool) {
	m := rendering.MaterialData{}
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to load the material file", "file", path, "error", err)
		return m, false
	}
	if err := json.Unmarshal(data, &m); err != nil {
		slog.Error("failed to unmarshal the material data", "error", err)
		return m, false
	}
	return m, true
}

func OpenMaterial(path string) {
	setup(func(win *ShaderDesigner) {
		if m, ok := loadMaterialData(path); ok {
			win.material = m
			win.ShowMaterialWindow()
		}
	})
}

func (win *ShaderDesigner) materialSave(e *document.Element) {
	if err := os.MkdirAll(materialFolder, os.ModePerm); err != nil {
		slog.Error("failed to create the materials folder",
			"folder", materialFolder, "error", err)
	}
	path := filepath.Join(materialFolder, win.material.Name+editor_config.FileExtensionMaterial)
	if _, err := os.Stat(path); err == nil {
		ok := <-alert.New("Overwrite?", "You are about to overwrite a material with the same name, would you like to continue?", "Yes", "No", win.man.Host)
		if !ok {
			return
		}
	}
	res, err := json.Marshal(win.material)
	if err != nil {
		slog.Error("failed to marshal the material data", "error", err)
		return
	}
	if err := os.WriteFile(path, res, os.ModePerm); err != nil {
		slog.Error("failed to write the material data to file", "error", err)
		return
	}
	slog.Info("material successfully saved", "file", path)
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}

/******************************************************************************/
/* material_window.go                                                         */
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
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/rendering"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/KaijuEngine/uuid"
)

func collectSpecificFileOptions(pfs *project_file_system.FileSystem, cache *content_database.Cache, cat content_database.ContentCategory) []ui.SelectOption {
	found := cache.ListByType(cat.TypeName())
	options := make([]ui.SelectOption, 0, len(found))
	for i := range found {
		options = append(options, ui.SelectOption{
			Name:  found[i].Config.Name,
			Value: content_database.ToContentPath(found[i].Path),
		})
	}
	stock := project_file_system.StockFolder
	dir, err := pfs.ReadDir(stock)
	if err != nil {
		return options
	}
	for i := range dir {
		if dir[i].IsDir() {
			continue
		}
		if slices.Contains(cat.ExtNames(), filepath.Ext(dir[i].Name())) {
			options = append(options, ui.SelectOption{dir[i].Name(), dir[i].Name()})
		}
	}
	return options
}

func collectShaderOptions(pfs *project_file_system.FileSystem, cache *content_database.Cache) []ui.SelectOption {
	return collectSpecificFileOptions(pfs, cache, content_database.Shader{})
}

func collectRenderPassOptions(pfs *project_file_system.FileSystem, cache *content_database.Cache) []ui.SelectOption {
	return collectSpecificFileOptions(pfs, cache, content_database.RenderPass{})
}

func collectShaderPipelinesOptions(pfs *project_file_system.FileSystem, cache *content_database.Cache) []ui.SelectOption {
	return collectSpecificFileOptions(pfs, cache, content_database.ShaderPipeline{})
}

func collectTextureOptions(pfs *project_file_system.FileSystem, cache *content_database.Cache) []ui.SelectOption {
	return collectSpecificFileOptions(pfs, cache, content_database.Texture{})
}

func (win *ShaderDesigner) reloadMaterialDoc() {
	sy := float32(0)
	if win.materialDoc != nil {
		content := win.materialDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.materialDoc.Destroy()
	}
	listings := map[string][]ui.SelectOption{}
	listings["Shader"] = collectShaderOptions(win.pfs, win.cache)
	listings["RenderPass"] = collectRenderPassOptions(win.pfs, win.cache)
	listings["ShaderPipeline"] = collectShaderPipelinesOptions(win.pfs, win.cache)
	listings["Texture"] = collectTextureOptions(win.pfs, win.cache)
	data := common_workspace.ReflectUIStructure(&win.material.MaterialData, "", listings)
	data.Name = "Material Editor"
	win.materialDoc, _ = markup.DocumentFromHTMLAsset(win.uiMan, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":     showMaterialTooltip,
			"valueChanged":    win.materialValueChanged,
			"addToSlice":      win.materialAddToSlice,
			"removeFromSlice": win.materialRemoveFromSlice,
			"saveData":        win.materialSave,
		})
	if sy != 0 {
		content := win.materialDoc.GetElementsByClass("topFields")[0]
		win.uiMan.Host.RunAfterFrames(2, func() {
			content.UIPanel.SetScrollY(sy)
		})
	}
}

func showMaterialTooltip(e *document.Element) { showTooltip(materialTooltips, e) }

func (win *ShaderDesigner) materialAddToSlice(e *document.Element) {
	common_workspace.ReflectAddToSlice(&win.material, e)
	win.reloadMaterialDoc()
}

func (win *ShaderDesigner) materialRemoveFromSlice(e *document.Element) {
	common_workspace.ReflectRemoveFromSlice(&win.material, e)
	win.reloadMaterialDoc()
}

func (win *ShaderDesigner) materialValueChanged(e *document.Element) {
	common_workspace.SetObjectValueFromUI(&win.material, e)
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

func (win *ShaderDesigner) materialSave(e *document.Element) {
	if win.material.id == "" {
		win.material.id = uuid.NewString()
	}
	win.material.RenderPass = filepath.ToSlash(win.material.RenderPass)
	win.material.Shader = filepath.ToSlash(win.material.Shader)
	win.material.ShaderPipeline = filepath.ToSlash(win.material.ShaderPipeline)
	for i := range win.material.Textures {
		win.material.Textures[i].Texture = filepath.ToSlash(win.material.Textures[i].Texture)
	}
	res, err := json.Marshal(win.material)
	if err != nil {
		slog.Error("failed to marshal the material data", "error", err)
		return
	}
	err = win.pfs.WriteFile(filepath.Join(project_file_system.ContentFolder,
		project_file_system.ContentMaterialFolder, win.material.id), res, os.ModePerm)
	if err != nil {
		slog.Error("failed to write the material data to file", "error", err)
		return
	}
	slog.Info("material successfully saved")
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}

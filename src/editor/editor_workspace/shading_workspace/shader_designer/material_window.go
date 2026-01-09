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
	"kaiju/editor/editor_overlay/content_selector"
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
	"weak"
)

func collectSpecificFileOptions(pfs *project_file_system.FileSystem, cache *content_database.Cache, cat content_database.ContentCategory) []ui.SelectOption {
	found := cache.ListByType(cat.TypeName())
	options := make([]ui.SelectOption, 0, len(found))
	for i := range found {
		options = append(options, ui.SelectOption{
			Name:  found[i].Config.Name,
			Value: found[i].Id(),
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
			options = append(options, ui.SelectOption{
				Name:  dir[i].Name(),
				Value: dir[i].Name(),
			})
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

func (win *ShaderDesigner) materialPullTextureLabels() {
	if win.material.Shader == "" {
		return
	}
	s, err := win.host.AssetDatabase().Read(win.material.Shader)
	if err != nil {
		return
	}
	var sh rendering.ShaderData
	err = json.Unmarshal(s, &sh)
	if err != nil {
		return
	}
	for i := range min(len(win.material.Textures), len(sh.SamplerLabels)) {
		win.material.Textures[i].Label = sh.SamplerLabels[i]
	}
}

func (win *ShaderDesigner) reloadMaterialDoc() {
	sy := float32(0)
	if win.materialDoc != nil {
		content := win.materialDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.materialDoc.Destroy()
	}
	pfs := win.ed.ProjectFileSystem()
	cache := win.ed.Cache()
	listings := map[string][]ui.SelectOption{}
	listings["Shader"] = collectShaderOptions(pfs, cache)
	listings["RenderPass"] = collectRenderPassOptions(pfs, cache)
	listings["ShaderPipeline"] = collectShaderPipelinesOptions(pfs, cache)
	win.materialPullTextureLabels()
	data := common_workspace.ReflectUIStructure(win.ed.Cache(),
		&win.material.MaterialData, "", listings)
	data.Name = "Material Editor"
	data.GroupName = win.material.name
	win.materialDoc, _ = markup.DocumentFromHTMLAsset(win.uiMan, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":          showMaterialTooltip,
			"valueChanged":         win.materialValueChanged,
			"addToSlice":           win.materialAddToSlice,
			"removeFromSlice":      win.materialRemoveFromSlice,
			"clickSelectContentId": win.clickSelectContentId,
			"saveData":             win.materialSave,
		})
	input, _ := win.materialDoc.GetElementById("nameInput")
	win.nameInputField = weak.Make(input)
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

func (win *ShaderDesigner) clickSelectContentId(e *document.Element) {
	cache := win.ed.Cache()
	win.ed.BlurInterface()
	content_selector.Show(win.host, e.Attribute("data-type"), cache,
		func(id string) {
			e.SetAttribute("value", id)
			cc, err := cache.Read(id)
			if err != nil || cc.Config.Name == "" {
				e.InnerLabel().SetText(id)
			} else {
				e.InnerLabel().SetText(cc.Config.Name)
			}
			win.materialValueChanged(e)
			win.ed.FocusInterface()
		}, win.ed.FocusInterface)

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
	win.material.name = win.nameInputField.Value().UI.ToInput().Text()
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
	if win.material.id != "" {
		err = win.ed.ProjectFileSystem().WriteFile(filepath.Join(project_file_system.ContentFolder,
			project_file_system.ContentMaterialFolder, win.material.id), res, os.ModePerm)
		if _, err := win.ed.Cache().Rename(win.material.id, win.material.name, win.ed.ProjectFileSystem()); err == nil {
			win.ed.Events().OnContentRenamed.Execute(win.material.id)
		}
	} else {
		ids := content_database.ImportRaw(win.material.name, res, content_database.Material{}, win.ed.ProjectFileSystem(), win.ed.Cache())
		if len(ids) > 0 {
			win.material.id = ids[0]
			win.ed.Events().OnContentAdded.Execute(ids)
		} else {
			err = errors.New("failed to import the raw material file data to the database")
		}
	}
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

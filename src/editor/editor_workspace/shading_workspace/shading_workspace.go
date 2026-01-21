/******************************************************************************/
/* shading_workspace.go                                                       */
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

package shading_workspace

import (
	"encoding/json"
	"kaiju/editor/editor_stage_manager/editor_stage_view"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/editor_workspace/shading_workspace/shader_designer"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"log/slog"
)

type ShadingWorkspace struct {
	common_workspace.CommonWorkspace
	ed                     ShadingWorkspaceEditorInterface
	stageView              *editor_stage_view.StageView
	designer               shader_designer.ShaderDesigner
	renderSpecList         *document.Element
	renderSpecListTemplate *document.Element
	toolTip                *document.Element
}

type ShadingWorkspaceUIData struct {
	Files []ShadingWorkspaceUIDataFile
}

type ShadingWorkspaceUIDataFile struct {
	Id   string
	Type string
	Name string
}

func (w *ShadingWorkspace) Initialize(host *engine.Host, ed ShadingWorkspaceEditorInterface) {
	w.ed = ed
	w.stageView = ed.StageView()
	data := ShadingWorkspaceUIData{Files: w.readExisting()}
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/shading_workspace.go.html", data, map[string]func(*document.Element){
			"toggleFilterSpec":   w.toggleFilterSpec,
			"selectSpec":         w.selectSpec,
			"clickNewRenderPass": w.clickNewRenderPass,
			"clickNewPipeline":   w.clickNewPipeline,
			"clickNewShader":     w.clickNewShader,
			"clickNewMaterial":   w.clickNewMaterial,
			"showTooltip":        w.showTooltip,
		})
	w.designer.Initialize(host, &w.UiMan, w.ed)
	w.renderSpecList, _ = w.Doc.GetElementById("renderSpecList")
	w.renderSpecListTemplate, _ = w.Doc.GetElementById("renderSpecListTemplate")
	w.toolTip, _ = w.Doc.GetElementById("toolTip")
	w.ed.Events().OnContentAdded.Add(w.contentAdded)
	w.ed.Events().OnContentRemoved.Add(w.contentRemoved)
	w.ed.Events().OnContentRenamed.Add(w.contentRenamed)
}

func (w *ShadingWorkspace) Open() {
	defer tracing.NewRegion("ShadingWorkspace.Open").End()
	w.CommonOpen()
	w.stageView.Open()
	w.renderSpecListTemplate.UI.Hide()
	w.designer.ChangeWindowState(shader_designer.StateNone)
	w.renderSpecList.UI.Clean()
}

func (w *ShadingWorkspace) Close() {
	defer tracing.NewRegion("ShadingWorkspace.Close").End()
	w.CommonClose()
	w.stageView.Close()
	w.designer.Close()
}

func (w *ShadingWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *ShadingWorkspace) OpenSpec(id string) {
	defer tracing.NewRegion("ShadingWorkspace.OpenSpec").End()
	if id == "" {
		return
	}
	cc, err := w.ed.Cache().Read(id)
	if err != nil {
		slog.Error("failed to read the config for content", "id", id, "error", err)
		return
	}
	w.designer.ChangeWindowState(shader_designer.StateNone)
	data, err := w.ed.ProjectFileSystem().ReadFile(content_database.ToContentPath(cc.Path))
	if err != nil {
		slog.Error("failed to read the "+cc.Config.Type+" content", "id", id, "error", err)
		return
	}
	if cc.Config.Type == (content_database.Material{}).TypeName() {
		d := rendering.MaterialData{}
		if err := json.Unmarshal(data, &d); err != nil {
			slog.Error("failed to unmarshal the material data", "id", id, "error", err)
			return
		}
		w.designer.ShowMaterialWindow(id, d)
	} else if cc.Config.Type == (content_database.Shader{}).TypeName() {
		d := rendering.ShaderData{}
		if err := json.Unmarshal(data, &d); err != nil {
			slog.Error("failed to unmarshal the material data", "id", id, "error", err)
			return
		}
		w.designer.ShowShaderWindow(id, d)
	} else if cc.Config.Type == (content_database.ShaderPipeline{}).TypeName() {
		d := rendering.ShaderPipelineData{}
		if err := json.Unmarshal(data, &d); err != nil {
			slog.Error("failed to unmarshal the material data", "id", id, "error", err)
			return
		}
		w.designer.ShowPipelineWindow(id, d)
	} else if cc.Config.Type == (content_database.RenderPass{}).TypeName() {
		d := rendering.RenderPassData{}
		if err := json.Unmarshal(data, &d); err != nil {
			slog.Error("failed to unmarshal the material data", "id", id, "error", err)
			return
		}
		w.designer.ShowRenderPassWindow(id, d)
	}
	elm, ok := w.Doc.GetElementById(id)
	if !ok {
		return
	}
	for _, e := range elm.Parent.Value().Children {
		w.Doc.SetElementClassesWithoutApply(e, "edPanelBgHoverable")
	}
	w.Doc.SetElementClassesWithoutApply(elm, "edPanelBgHoverable", "selected")
	w.Doc.ApplyStyles()
}

func (w *ShadingWorkspace) Update(deltaTime float64) {
	defer tracing.NewRegion("ShadingWorkspace.update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.stageView.Update(deltaTime, w.ed.Project())
}

func (w *ShadingWorkspace) readExisting() []ShadingWorkspaceUIDataFile {
	defer tracing.NewRegion("ShadingWorkspace.readExisting").End()
	out := []ShadingWorkspaceUIDataFile{}
	fs := w.ed.ProjectFileSystem()
	paths := map[string]string{
		".material":       project_file_system.ContentFolder + "/" + project_file_system.ContentMaterialFolder,
		".shader":         project_file_system.ContentFolder + "/" + project_file_system.ContentShaderFolder,
		".shaderpipeline": project_file_system.ContentFolder + "/" + project_file_system.ContentShaderPipelineFolder,
		".renderpass":     project_file_system.ContentFolder + "/" + project_file_system.ContentRenderPassFolder,
	}
	cache := w.ed.Cache()
	for _, v := range paths {
		dir, err := fs.ReadDir(v)
		if err != nil {
			continue
		}
		for j := range dir {
			if dir[j].IsDir() {
				continue
			}
			cc, err := cache.Read(dir[j].Name())
			if err != nil {
				continue
			}
			out = append(out, ShadingWorkspaceUIDataFile{
				Id:   cc.Id(),
				Name: cc.Config.Name,
				Type: cc.Config.Type,
			})
		}
	}
	return out
}

func (w *ShadingWorkspace) toggleFilterSpec(e *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.toggleFilterSpec").End()
	txt := e.InnerLabel().Text()
	typeFilter := ""
	for _, elm := range e.Parent.Value().Children {
		w.Doc.SetElementClassesWithoutApply(elm, "edPanelBgHoverable")
	}
	w.Doc.SetElementClassesWithoutApply(e, "edPanelBgHoverable", "selected")
	w.Doc.ApplyStyles()
	switch txt {
	case "All":
	case "R":
		typeFilter = content_database.RenderPass{}.TypeName()
	case "P":
		typeFilter = content_database.ShaderPipeline{}.TypeName()
	case "S":
		typeFilter = content_database.Shader{}.TypeName()
	case "M":
		typeFilter = content_database.Material{}.TypeName()
	default:
		return
	}
	for _, e := range w.renderSpecList.Children {
		if typeFilter == "" {
			e.UI.Show()
			continue
		}
		ext := e.Attribute("data-type")
		if ext != typeFilter {
			e.UI.Hide()
		} else {
			e.UI.Show()
		}
	}
	w.renderSpecListTemplate.UI.Hide()
	w.renderSpecList.UI.Clean()
}

func (w *ShadingWorkspace) selectSpec(elm *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.selectSpec").End()
	w.OpenSpec(elm.Attribute("id"))
}

func (w *ShadingWorkspace) clickNewRenderPass(elm *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.clickNewRenderPass").End()
	w.designer.ShowRenderPassWindow("", rendering.RenderPassData{})
}

func (w *ShadingWorkspace) clickNewPipeline(elm *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.clickNewPipeline").End()
	w.designer.ShowPipelineWindow("", rendering.ShaderPipelineData{})
}

func (w *ShadingWorkspace) clickNewShader(elm *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.clickNewShader").End()
	w.designer.ShowShaderWindow("", rendering.ShaderData{})
}

func (w *ShadingWorkspace) clickNewMaterial(elm *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.clickNewMaterial").End()
	w.designer.ShowMaterialWindow("", rendering.MaterialData{})
}

func (w *ShadingWorkspace) showTooltip(elm *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.showTooltip").End()
	w.toolTip.InnerLabel().SetText(elm.Attribute("data-tooltip"))
}

func (w *ShadingWorkspace) contentAdded(ids []string) {
	targets := []content_database.CachedContent{}
	for i := range ids {
		cc, err := w.ed.Cache().Read(ids[i])
		if err != nil {
			continue
		}
		switch cc.Config.Type {
		case content_database.Material{}.TypeName(),
			content_database.Shader{}.TypeName(),
			content_database.ShaderPipeline{}.TypeName(),
			content_database.RenderPass{}.TypeName():
			targets = append(targets, cc)
		}
	}
	if len(targets) == 0 {
		return
	}
	elms := w.Doc.DuplicateElementRepeatWithoutApplyStyles(w.renderSpecListTemplate, len(targets))
	for i := range elms {
		w.Doc.SetElementId(elms[i], targets[i].Id())
		elms[i].SetAttribute("data-type", targets[i].Config.Type)
		elms[i].InnerLabel().SetText(targets[i].Config.Name)
	}
	w.Doc.ApplyStyles()
}

func (w *ShadingWorkspace) contentRemoved(ids []string) {
	elms := make([]*document.Element, 0, len(ids))
	for i := range ids {
		e, ok := w.Doc.GetElementById(ids[i])
		if ok {
			elms = append(elms, e)
		}
	}
	for i := range elms {
		w.Doc.RemoveElementWithoutApplyStyles(elms[i])
	}
	w.Doc.ApplyStyles()
}

func (w *ShadingWorkspace) contentRenamed(id string) {
	e, ok := w.Doc.GetElementById(id)
	if !ok {
		return
	}
	cc, err := w.ed.Cache().Read(id)
	if err != nil {
		return
	}
	e.InnerLabel().SetText(cc.Config.Name)
}

/******************************************************************************/
/* vulkan_workspace.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vulkan_workspace

import (
	"encoding/json"
	"log/slog"

	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace/vulkan_workspace/shader_designer"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const (
	ID          = "vulkan"
	DisplayName = "Vulkan"
)

func init() {
	editor_workspace_registry.Register(&VulkanWorkspace{})
}

type VulkanWorkspace struct {
	common_workspace.CommonWorkspace
	ed                     editor_workspace.WorkspaceEditorInterface
	stageView              *editor_stage_view.StageView
	designer               shader_designer.ShaderDesigner
	renderSpecList         *document.Element
	renderSpecListTemplate *document.Element
	toolTip                *document.Element
	openSpecSubID          events.Id
}

type VulkanWorkspaceUIData struct {
	Files []VulkanWorkspaceUIDataFile
}

type VulkanWorkspaceUIDataFile struct {
	Id   string
	Type string
	Name string
}

func (w *VulkanWorkspace) ID() string          { return ID }
func (w *VulkanWorkspace) DisplayName() string { return DisplayName }
func (w *VulkanWorkspace) IsRequired() bool    { return false }

func (w *VulkanWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	host := ed.Host()
	w.ed = ed
	w.stageView = ed.StageView()
	data := VulkanWorkspaceUIData{Files: w.readExisting()}
	if err := w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/vulkan_workspace.go.html", data, map[string]func(*document.Element){
			"toggleFilterSpec":   w.toggleFilterSpec,
			"selectSpec":         w.selectSpec,
			"clickNewRenderPass": w.clickNewRenderPass,
			"clickNewPipeline":   w.clickNewPipeline,
			"clickNewShader":     w.clickNewShader,
			"clickNewMaterial":   w.clickNewMaterial,
			"showTooltip":        w.showTooltip,
		}); err != nil {
		return err
	}
	w.designer.Initialize(host, &w.UiMan, w.ed)
	w.renderSpecList, _ = w.Doc.GetElementById("renderSpecList")
	w.renderSpecListTemplate, _ = w.Doc.GetElementById("renderSpecListTemplate")
	w.toolTip, _ = w.Doc.GetElementById("toolTip")
	w.ed.Events().OnContentAdded.Add(w.contentAdded)
	w.ed.Events().OnContentRemoved.Add(w.contentRemoved)
	w.ed.Events().OnContentRenamed.Add(w.contentRenamed)
	// Subscribe to cross-workspace request to open a spec; this also
	// switches the Vulkan workspace active.
	w.openSpecSubID = ed.Events().OnRequestOpenVulkanSpec.Add(func(specID string) {
		ed.SelectWorkspace(ID)
		w.OpenSpec(specID)
	})
	return nil
}

func (w *VulkanWorkspace) Shutdown() {
	defer tracing.NewRegion("VulkanWorkspace.Shutdown").End()
	if w.ed != nil {
		w.ed.Events().OnRequestOpenVulkanSpec.Remove(w.openSpecSubID)
	}
	w.CommonShutdown()
}

func (w *VulkanWorkspace) Open() {
	defer tracing.NewRegion("VulkanWorkspace.Open").End()
	w.CommonOpen()
	w.stageView.Open()
	w.renderSpecListTemplate.UI.Hide()
	w.designer.ChangeWindowState(shader_designer.StateNone)
	w.renderSpecList.UI.Clean()
}

func (w *VulkanWorkspace) Close() {
	defer tracing.NewRegion("VulkanWorkspace.Close").End()
	w.CommonClose()
	w.stageView.Close()
	w.designer.Close()
}

func (w *VulkanWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *VulkanWorkspace) OpenSpec(id string) {
	defer tracing.NewRegion("VulkanWorkspace.OpenSpec").End()
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

func (w *VulkanWorkspace) Update(deltaTime float64) {
	defer tracing.NewRegion("VulkanWorkspace.update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.stageView.Update(deltaTime, w.ed.Project())
}

func (w *VulkanWorkspace) readExisting() []VulkanWorkspaceUIDataFile {
	defer tracing.NewRegion("VulkanWorkspace.readExisting").End()
	out := []VulkanWorkspaceUIDataFile{}
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
			out = append(out, VulkanWorkspaceUIDataFile{
				Id:   cc.Id(),
				Name: cc.Config.Name,
				Type: cc.Config.Type,
			})
		}
	}
	return out
}

func (w *VulkanWorkspace) toggleFilterSpec(e *document.Element) {
	defer tracing.NewRegion("VulkanWorkspace.toggleFilterSpec").End()
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

func (w *VulkanWorkspace) selectSpec(elm *document.Element) {
	defer tracing.NewRegion("VulkanWorkspace.selectSpec").End()
	w.OpenSpec(elm.Attribute("id"))
}

func (w *VulkanWorkspace) clickNewRenderPass(elm *document.Element) {
	defer tracing.NewRegion("VulkanWorkspace.clickNewRenderPass").End()
	w.designer.ShowRenderPassWindow("", rendering.RenderPassData{})
}

func (w *VulkanWorkspace) clickNewPipeline(elm *document.Element) {
	defer tracing.NewRegion("VulkanWorkspace.clickNewPipeline").End()
	w.designer.ShowPipelineWindow("", rendering.ShaderPipelineData{})
}

func (w *VulkanWorkspace) clickNewShader(elm *document.Element) {
	defer tracing.NewRegion("VulkanWorkspace.clickNewShader").End()
	w.designer.ShowShaderWindow("", rendering.ShaderData{})
}

func (w *VulkanWorkspace) clickNewMaterial(elm *document.Element) {
	defer tracing.NewRegion("VulkanWorkspace.clickNewMaterial").End()
	w.designer.ShowMaterialWindow("", rendering.MaterialData{})
}

func (w *VulkanWorkspace) showTooltip(elm *document.Element) {
	defer tracing.NewRegion("VulkanWorkspace.showTooltip").End()
	w.toolTip.InnerLabel().SetText(elm.Attribute("data-tooltip"))
}

func (w *VulkanWorkspace) contentAdded(ids []string) {
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

func (w *VulkanWorkspace) contentRemoved(ids []string) {
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

func (w *VulkanWorkspace) contentRenamed(id string) {
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

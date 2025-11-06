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
	updateId               engine.UpdateId
	designer               shader_designer.ShaderDesigner
	renderSpecList         *document.Element
	renderSpecListTemplate *document.Element
}

type ShadingWorkspaceUIData struct {
	Files []ShadingWorkspaceUIDataFile
}

type ShadingWorkspaceUIDataFile struct {
	Id   string
	Ext  string
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
		})
	w.designer.Initialize(host, &w.UiMan, w.ed.ProjectFileSystem(), w.ed.Cache())
	w.renderSpecList, _ = w.Doc.GetElementById("renderSpecList")
	w.renderSpecListTemplate, _ = w.Doc.GetElementById("renderSpecListTemplate")
}

func (w *ShadingWorkspace) Open() {
	defer tracing.NewRegion("ShadingWorkspace.Open").End()
	w.CommonOpen()
	w.stageView.Open()
	w.renderSpecListTemplate.UI.Hide()
	w.updateId = w.Host.Updater.AddUpdate(w.update)
	w.designer.ChangeWindowState(shader_designer.StateNone)
	w.renderSpecList.UI.Clean()
}

func (w *ShadingWorkspace) Close() {
	defer tracing.NewRegion("ShadingWorkspace.Close").End()
	w.CommonClose()
	w.stageView.Close()
	w.Host.Updater.RemoveUpdate(&w.updateId)
}

func (w *ShadingWorkspace) update(deltaTime float64) {
	defer tracing.NewRegion("ShadingWorkspace.update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.stageView.Update(deltaTime)
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
	for k, v := range paths {
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
				Ext:  k,
			})
		}
	}
	return out
}

func (w *ShadingWorkspace) toggleFilterSpec(e *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.toggleFilterSpec").End()
	txt := e.InnerLabel().Text()
	extFilter := ""
	for _, elm := range e.Parent.Value().Children {
		w.Doc.SetElementClassesWithoutApply(elm, "filterLabel")
	}
	w.Doc.SetElementClassesWithoutApply(e, "filterLabel", "filterLabelSelected")
	w.Doc.ApplyStyles()
	switch txt {
	case "All":
	case "R":
		extFilter = ".renderpass"
	case "P":
		extFilter = ".shaderpipeline"
	case "S":
		extFilter = ".shader"
	case "M":
		extFilter = ".material"
	default:
		return
	}
	for _, e := range w.renderSpecList.Children {
		if extFilter == "" {
			e.UI.Show()
			continue
		}
		ext := e.Attribute("data-ext")
		if ext != extFilter {
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
	id := elm.Attribute("id")
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

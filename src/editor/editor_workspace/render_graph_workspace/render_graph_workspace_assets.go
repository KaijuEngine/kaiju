/******************************************************************************/
/* render_graph_workspace_assets.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"log/slog"
	"os"
	"strings"

	"kaijuengine.com/editor/editor_overlay/content_selector"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

const defaultRenderGraphName = "Untitled Render Graph"

func (w *RenderGraphWorkspace) resetGraphToDefault() {
	w.graph.clear()
	w.graph.pan = matrix.Vec2Zero()
	w.createNodeCount = 0
	source, _ := w.graph.CreateCatalogNode("principled-bsdf", matrix.NewVec2(42, 56))
	output, _ := w.graph.CreateCatalogNode("material-output", matrix.NewVec2(350, 150))
	if source != nil && output != nil {
		w.graph.CreateConnection(source.Output(0), output.Input(0))
	}
	w.graph.applyViewOffsets()
}

func (w *RenderGraphWorkspace) renameRenderGraph(*document.Element) {
	defer tracing.NewRegion("RenderGraphWorkspace.renameRenderGraph").End()
	w.currentName = w.renderGraphNameFromInput()
	w.updateGraphNameInput()
	if w.currentGraphID == "" || w.ed == nil {
		w.setRenderGraphStatus("Unsaved render graph")
		return
	}
	cache := w.ed.Cache()
	pfs := w.ed.ProjectFileSystem()
	if cache == nil || pfs == nil {
		return
	}
	cc, err := cache.Read(w.currentGraphID)
	if err != nil || cc.Config.Name == w.currentName {
		return
	}
	if _, err = cache.Rename(w.currentGraphID, w.currentName, pfs); err != nil {
		slog.Error("failed to rename render graph", "id", w.currentGraphID, "error", err)
		w.setRenderGraphStatus("Rename failed")
		return
	}
	w.ed.Events().OnContentRenamed.Execute(w.currentGraphID)
	w.setRenderGraphStatus("Renamed")
}

func (w *RenderGraphWorkspace) newRenderGraph(*document.Element) {
	defer tracing.NewRegion("RenderGraphWorkspace.newRenderGraph").End()
	w.currentGraphID = ""
	w.currentName = defaultRenderGraphName
	w.resetGraphToDefault()
	w.updateGraphNameInput()
	w.setRenderGraphStatus("Unsaved render graph")
}

func (w *RenderGraphWorkspace) loadRenderGraph(*document.Element) {
	defer tracing.NewRegion("RenderGraphWorkspace.loadRenderGraph").End()
	if w.ed == nil || w.Host == nil {
		return
	}
	w.ed.BlurInterface()
	if _, err := content_selector.Show(w.Host, (content_database.RenderGraph{}).TypeName(), w.ed.Cache(), func(id string) {
		w.ed.FocusInterface()
		if strings.TrimSpace(id) == "" {
			return
		}
		w.LoadRenderGraphID(id)
	}, func() {
		w.ed.FocusInterface()
	}); err != nil {
		w.ed.FocusInterface()
		slog.Error("failed to show render graph selector", "error", err)
	}
}

func (w *RenderGraphWorkspace) LoadRenderGraphID(id string) {
	defer tracing.NewRegion("RenderGraphWorkspace.LoadRenderGraphID").End()
	cache := w.ed.Cache()
	pfs := w.ed.ProjectFileSystem()
	if cache == nil || pfs == nil {
		return
	}
	cc, err := cache.Read(id)
	if err != nil {
		slog.Error("failed to load render graph cache entry", "id", id, "error", err)
		w.setRenderGraphStatus("Load failed")
		return
	}
	data, err := pfs.ReadFile(cc.ContentPath())
	if err != nil {
		slog.Error("failed to read render graph", "id", id, "path", cc.ContentPath(), "error", err)
		w.setRenderGraphStatus("Load failed")
		return
	}
	document, err := DeserializeRenderGraphDocument(data)
	if err != nil {
		slog.Error("failed to parse render graph", "id", id, "error", err)
		w.setRenderGraphStatus("Load failed")
		return
	}
	if err = w.graph.LoadDocument(document); err != nil {
		slog.Error("failed to apply render graph", "id", id, "error", err)
		w.setRenderGraphStatus("Load failed")
		return
	}
	w.currentGraphID = id
	w.currentName = renderGraphCleanName(cc.Config.Name)
	if w.currentName == "" && document.Name != "" {
		w.currentName = renderGraphCleanName(document.Name)
	}
	w.updateGraphNameInput()
	w.setRenderGraphStatus("Loaded")
}

func (w *RenderGraphWorkspace) saveRenderGraph(*document.Element) {
	defer tracing.NewRegion("RenderGraphWorkspace.saveRenderGraph").End()
	if w.ed == nil {
		return
	}
	w.currentName = w.renderGraphNameFromInput()
	w.updateGraphNameInput()
	data, err := w.SerializeGraph()
	if err != nil {
		slog.Error("failed to serialize render graph", "error", err)
		w.setRenderGraphStatus("Save failed")
		return
	}
	if w.currentGraphID == "" {
		w.saveNewRenderGraphContent(data)
		return
	}
	w.saveExistingRenderGraphContent(data)
}

func (w *RenderGraphWorkspace) saveNewRenderGraphContent(data []byte) {
	pfs := w.ed.ProjectFileSystem()
	cache := w.ed.Cache()
	if pfs == nil || cache == nil {
		return
	}
	ids := content_database.ImportRaw(w.currentName, data, content_database.RenderGraph{}, pfs, cache)
	if len(ids) == 0 {
		w.setRenderGraphStatus("Save failed")
		return
	}
	w.currentGraphID = ids[0]
	w.ed.Events().OnContentAdded.Execute(ids)
	w.ed.Events().OnContentChangesSaved.Execute(w.currentGraphID)
	w.setRenderGraphStatus("Saved")
}

func (w *RenderGraphWorkspace) saveExistingRenderGraphContent(data []byte) {
	pfs := w.ed.ProjectFileSystem()
	cache := w.ed.Cache()
	if pfs == nil || cache == nil {
		return
	}
	cc, err := cache.Read(w.currentGraphID)
	if err != nil {
		w.currentGraphID = ""
		w.saveNewRenderGraphContent(data)
		return
	}
	if cc.Config.Name != w.currentName {
		if _, err = cache.Rename(w.currentGraphID, w.currentName, pfs); err != nil {
			slog.Error("failed to rename render graph while saving", "id", w.currentGraphID, "error", err)
			w.setRenderGraphStatus("Save failed")
			return
		}
		w.ed.Events().OnContentRenamed.Execute(w.currentGraphID)
	}
	if err = pfs.WriteFile(cc.ContentPath(), data, os.ModePerm); err != nil {
		slog.Error("failed to write render graph", "id", w.currentGraphID, "path", cc.ContentPath(), "error", err)
		w.setRenderGraphStatus("Save failed")
		return
	}
	w.ed.Events().OnContentChangesSaved.Execute(w.currentGraphID)
	w.setRenderGraphStatus("Saved")
}

func (w *RenderGraphWorkspace) renderGraphNameFromInput() string {
	if w.nameInput == nil || w.nameInput.UI == nil {
		return renderGraphCleanName(w.currentName)
	}
	return renderGraphCleanName(w.nameInput.UI.ToInput().Text())
}

func (w *RenderGraphWorkspace) updateGraphNameInput() {
	w.currentName = renderGraphCleanName(w.currentName)
	if w.nameInput != nil && w.nameInput.UI != nil {
		w.nameInput.UI.ToInput().SetTextWithoutEvent(w.currentName)
	}
}

func (w *RenderGraphWorkspace) setRenderGraphStatus(text string) {
	if w.status == nil {
		return
	}
	w.status.InnerLabel().SetText(text)
}

func renderGraphCleanName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return defaultRenderGraphName
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
			return '-'
		default:
			return r
		}
	}, name)
}

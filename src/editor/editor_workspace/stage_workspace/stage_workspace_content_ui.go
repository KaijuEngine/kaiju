package stage_workspace

import (
	"kaiju/editor/editor_workspace/content_workspace"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/hid"
	"kaiju/rendering"
	"log/slog"
	"slices"
	"strings"
	"weak"
)

type WorkspaceContentUI struct {
	workspace      weak.Pointer[Workspace]
	typeFilters    []string
	tagFilters     []string
	query          string
	contentArea    *document.Element
	dragPreview    *document.Element
	entryTemplate  *document.Element
	hideContentElm *document.Element
	showContentElm *document.Element
	dragging       *document.Element
	dragContentId  string
}

func (cui *WorkspaceContentUI) setup(w *Workspace, ids []string) {
	cui.workspace = weak.Make(w)
	cui.contentArea, _ = w.Doc.GetElementById("contentArea")
	cui.dragPreview, _ = w.Doc.GetElementById("dragPreview")
	cui.entryTemplate, _ = w.Doc.GetElementById("entryTemplate")
	cui.hideContentElm, _ = w.Doc.GetElementById("hideContent")
	cui.showContentElm, _ = w.Doc.GetElementById("showContent")
	cui.addContent(ids)
}

func (cui *WorkspaceContentUI) open() {
	cui.entryTemplate.UI.Hide()
	cui.dragPreview.UI.Hide()
	if cui.hideContentElm.UI.Entity().IsActive() {
		cui.showContentElm.UI.Hide()
	}
}

func (cui *WorkspaceContentUI) update(w *Workspace) bool {
	if cui.dragging != nil {
		m := &w.Host.Window.Mouse
		mp := m.ScreenPosition()
		ps := cui.dragPreview.UI.Layout().PixelSize()
		cui.dragPreview.UI.Layout().SetOffset(mp.X()-ps.X()*0.5, mp.Y()-ps.Y()*0.5)
		if m.Released(hid.MouseButtonLeft) {
			cui.dropContent(w, m)
		}
		return false
	}
	return true
}

func (cui *WorkspaceContentUI) processHotkeys(host *engine.Host) {
	if host.Window.Keyboard.KeyDown(hid.KeyboardKeyC) {
		if cui.hideContentElm.UI.Entity().IsActive() {
			cui.hideContent(nil)
		} else {
			cui.showContent(nil)
		}
	}
}

func (cui *WorkspaceContentUI) addContent(ids []string) {
	if len(ids) == 0 {
		return
	}
	w := cui.workspace.Value()
	ccAll := make([]content_database.CachedContent, 0, len(ids))
	for i := range ids {
		cc, err := w.cdb.Read(ids[i])
		if err != nil {
			slog.Error("failed to read the cached content", "id", ids[i], "error", err)
			continue
		}
		ccAll = append(ccAll, cc)
	}
	cpys := w.Doc.DuplicateElementRepeat(cui.entryTemplate, len(ccAll))
	for i := range cpys {
		cc := &ccAll[i]
		cpys[i].SetAttribute("id", cc.Id())
		cpys[i].SetAttribute("data-type", strings.ToLower(cc.Config.Type))
		lbl := cpys[i].Children[1].Children[0].UI.ToLabel()
		lbl.SetText(cc.Config.Name)
		cui.loadEntryImage(cpys[i], cc.Path, cc.Config.Type)
	}
}

func (cui *WorkspaceContentUI) loadEntryImage(e *document.Element, configPath, typeName string) {
	img := e.Children[0].UI.ToPanel()
	w := cui.workspace.Value()
	if typeName == (content_database.Texture{}).TypeName() {
		// Loose goroutine
		go func() {
			path := content_database.ToContentPath(configPath)
			data, err := w.pfs.ReadFile(path)
			if err != nil {
				slog.Error("error reading the image file", "path", path)
				return
			}
			tex, err := rendering.NewTextureFromMemory(rendering.GenerateUniqueTextureKey,
				data, 0, 0, rendering.TextureFilterLinear)
			if err != nil {
				slog.Error("failed to insert the texture to the cache", "error", err)
				return
			}
			w.Host.RunOnMainThread(func() {
				tex.DelayedCreate(w.Host.Window.Renderer)
				img.SetBackground(tex)
			})
		}()
	}
}

func (cui *WorkspaceContentUI) inputFilter(e *document.Element) {
	cui.query = strings.ToLower(e.UI.ToInput().Text())
	// TODO:  Regex out the filters like tag:..., type:..., etc.
	cui.runFilter()
}

func (cui *WorkspaceContentUI) tagFilter(e *document.Element) {
	q := strings.ToLower(e.UI.ToInput().Text())
	tagElms := cui.workspace.Value().Doc.GetElementsByGroup("tag")[1:]
	for i := range tagElms {
		tag := tagElms[i].Attribute("data-tag")
		show := strings.Contains(strings.ToLower(tag), q)
		if show {
			tagElms[i].UI.Show()
		} else {
			tagElms[i].UI.Hide()
		}
	}
}

func (cui *WorkspaceContentUI) runFilter() {
	w := cui.workspace.Value()
	entries := w.Doc.GetElementsByGroup("entry")
	for i := range entries {
		e := entries[i]
		id := e.Attribute("id")
		if id == "entryTemplate" {
			continue
		}
		if content_workspace.ShouldShowContent(cui.query, id, cui.typeFilters, cui.tagFilters, w.cdb) {
			e.UI.Entity().Activate()
		} else {
			e.UI.Entity().Deactivate()
		}
	}
	w.Host.RunOnMainThread(w.Doc.Clean)
}

func (cui *WorkspaceContentUI) clickFilter(e *document.Element) {
	isSelected := slices.Contains(e.ClassList(), "filterSelected")
	isSelected = !isSelected
	typeName := e.Attribute("data-type")
	tagName := e.Attribute("data-tag")
	w := cui.workspace.Value()
	if isSelected {
		w.Doc.SetElementClasses(e, "leftBtn", "filterSelected")
		if typeName != "" {
			cui.typeFilters = append(cui.typeFilters, typeName)
		}
		if tagName != "" {
			cui.tagFilters = append(cui.tagFilters, tagName)
		}
	} else {
		w.Doc.SetElementClasses(e, "leftBtn")
		if typeName != "" {
			cui.typeFilters = klib.SlicesRemoveElement(cui.typeFilters, typeName)
		}
		if tagName != "" {
			cui.tagFilters = klib.SlicesRemoveElement(cui.tagFilters, tagName)
		}
	}
	cui.runFilter()
}

func (cui *WorkspaceContentUI) hideContent(*document.Element) {
	cui.hideContentElm.UI.Hide()
	cui.showContentElm.UI.Show()
	cui.contentArea.UI.Hide()
}

func (cui *WorkspaceContentUI) showContent(*document.Element) {
	cui.showContentElm.UI.Hide()
	cui.hideContentElm.UI.Show()
	cui.contentArea.UI.Show()
}

func (cui *WorkspaceContentUI) entryDragStart(e *document.Element) {
	cui.dragging = e
	cui.dragPreview.UI.Show()
	cui.dragPreview.UIPanel.SetBackground(e.Children[0].UIPanel.Background())
	cui.dragContentId = e.Attribute("id")
}

func (cui *WorkspaceContentUI) dropContent(w *Workspace, m *hid.Mouse) {
	if !cui.contentArea.UI.Entity().Transform.ContainsPoint2D(m.CenteredPosition()) {
		cc, err := w.cdb.Read(cui.dragContentId)
		if err != nil {
			slog.Error("failed to read the content to spawn from cache", "id", cui.dragContentId)
			return
		}
		w.spawnContent(&cc, m)
	}
	cui.dragPreview.UI.Hide()
	cui.dragging = nil
	cui.dragContentId = ""
}

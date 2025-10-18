package content_workspace

import (
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/games/editor/editor_overlay/file_browser"
	"kaiju/games/editor/editor_workspace/common_workspace"
	"kaiju/games/editor/project/project_database/cache_database"
	"kaiju/games/editor/project/project_database/content_database"
	"kaiju/games/editor/project/project_file_system"
	"kaiju/klib"
	"kaiju/rendering"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
)

type Workspace struct {
	common_workspace.CommonWorkspace
	pfs           *project_file_system.FileSystem
	cdb           *cache_database.CacheDatabase
	filters       []string
	query         string
	entryTemplate *document.Element
	didFirstLoad  bool
}

type WorkspaceUIData struct {
	Filters []string
}

func (w *Workspace) Initialize(host *engine.Host, pfs *project_file_system.FileSystem, cdb *cache_database.CacheDatabase) {
	w.pfs = pfs
	w.cdb = cdb
	data := WorkspaceUIData{}
	for _, cat := range content_database.ContentCategories {
		data.Filters = append(data.Filters, cat.TypeName())
	}
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/content_workspace.go.html", data, map[string]func(*document.Element){
			"inputFilter": w.inputFilter,
			"clickImport": w.clickImport,
			"clickFilter": w.clickFilter,
		})
	w.entryTemplate, _ = w.Doc.GetElementById("entryTemplate")
}

func (w *Workspace) Open() {
	w.CommonOpen()
	w.entryTemplate.UI.Hide()
	if !w.didFirstLoad {
		w.didFirstLoad = true
		w.initListing()
	}
}

func (w *Workspace) Close() { w.CommonClose() }

func (w *Workspace) clickImport(*document.Element) {
	w.UiMan.DisableUpdate()
	file_browser.Show(w.Host, file_browser.Config{
		ExtFilter:   content_database.ImportableTypes,
		MultiSelect: true,
		OnConfirm: func(paths []string) {
			w.UiMan.EnableUpdate()
			index := []string{}
			for i := range paths {
				res, err := content_database.Import(paths[i], w.pfs)
				if err != nil {
					slog.Error("failed to import content", "path", paths[i], "error", err)
				} else {
					var addDependencies func(target *content_database.ImportResult)
					addDependencies = func(target *content_database.ImportResult) {
						index = append(index, target.Id)
						for i := range res.Dependencies {
							addDependencies(&target.Dependencies[i])
						}
					}
					addDependencies(&res)
				}
			}
			w.addContent(index)
		}, OnCancel: func() {
			w.UiMan.EnableUpdate()
		},
	})
}

func (w *Workspace) initListing() {
	idIter, err := w.cdb.AllIds()
	if err != nil {
		slog.Error("failed to get all of the cached ids", "error", err)
	} else {
		// TODO:  Actually use the iter later rather than aggregating, there
		// could be a ton of ids.
		ids := []string{}
		for s := range idIter {
			ids = append(ids, s)
		}
		w.addContent(ids)
	}
}

func (w *Workspace) addContent(ids []string) {
	if len(ids) == 0 {
		return
	}
	ccAll := make([]cache_database.CachedContent, 0, len(ids))
	for i := range ids {
		cc, err := w.cdb.Read(ids[i])
		if err != nil {
			slog.Error("failed to read the cached content", "id", ids[i], "error", err)
			continue
		}
		ccAll = append(ccAll, cc)
	}
	// TODO:  Go through each ID and add an entry into the list
	cpys := w.Doc.DuplicateElementRepeat(w.entryTemplate, len(ccAll))
	for i := range cpys {
		cc := &ccAll[i]
		cpys[i].SetAttribute("id", cc.Id())
		cpys[i].SetAttribute("data-type", strings.ToLower(cc.Config.Type))
		lbl := cpys[i].Children[1].Children[0].UI.ToLabel()
		lbl.SetText(cc.Config.Name)
		img := cpys[i].Children[0].UI.ToPanel()
		if cc.Config.Type == (content_database.Texture{}).TypeName() {
			// Loose goroutine
			go func() {
				path := content_database.ToContentPath(cc.Path)
				key := filepath.Base(path)
				data, err := w.pfs.ReadFile(path)
				if err != nil {
					slog.Error("error reading the image file", "path", path)
					return
				}
				td := rendering.ReadRawTextureData(data, rendering.TextureFileFormatPng)
				tex, err := w.Host.TextureCache().InsertTexture(key, td.Mem,
					td.Width, td.Height, rendering.TextureFilterLinear)
				if err != nil {
					slog.Error("failed to insert the texture to the cache", "error", err)
					return
				}
				img.SetBackground(tex)
			}()
		}
	}
}

func (w *Workspace) inputFilter(e *document.Element) {
	w.query = strings.ToLower(e.UI.ToInput().Text())
	// TODO:  Regex out the filters like tag:..., type:..., etc.
	w.filterOnMainThread()
}

func (w *Workspace) clickFilter(e *document.Element) {
	typeName := e.Attribute("data-type")
	isSelected := slices.Contains(e.ClassList(), "filterSelected")
	isSelected = !isSelected
	if isSelected {
		w.Doc.SetElementClasses(e, "leftBtn", "filterSelected")
	} else {
		w.Doc.SetElementClasses(e, "leftBtn")
	}
	if isSelected {
		w.filters = append(w.filters, typeName)
	} else {
		w.filters = klib.SlicesRemoveElement(w.filters, typeName)
	}
	w.filterOnMainThread()
}

func (w *Workspace) filterOnMainThread() {
	w.Host.RunOnMainThread(func() { w.runFilter() })
}

func (w *Workspace) runFilter() {
	entries := w.Doc.GetElementsByGroup("entry")
	for i := range entries {
		e := entries[i]
		id := e.Attribute("id")
		if id == "entryTemplate" {
			continue
		}
		show := len(w.filters) == 0 || slices.Contains(w.filters, e.Attribute("data-type"))
		if show && w.query != "" {
			show = w.runQueryOnContent(id)
		}
		if show {
			e.UI.Entity().Activate()
		} else {
			e.UI.Entity().Deactivate()
		}
	}
}

func (w *Workspace) runQueryOnContent(id string) bool {
	if cc, err := w.cdb.Read(id); err != nil {
		return false
	} else {
		// TODO:  Use filters like tag:..., type:..., etc.
		return strings.Contains(cc.Config.NameLower(), w.query)
	}
}

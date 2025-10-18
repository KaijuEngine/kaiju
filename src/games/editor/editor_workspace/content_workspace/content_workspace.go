package content_workspace

import (
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/games/editor/editor_overlay/file_browser"
	"kaiju/games/editor/editor_workspace/common_workspace"
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
	cCache        *content_database.Cache
	typeFilters   []string
	tagFilters    []string
	query         string
	queryTag      string
	entryTemplate *document.Element
	didFirstLoad  bool
}

type WorkspaceUIData struct {
	Filters []string
}

func (w *Workspace) Initialize(host *engine.Host, pfs *project_file_system.FileSystem, cdb *content_database.Cache) {
	w.pfs = pfs
	w.cCache = cdb
	data := WorkspaceUIData{}
	for _, cat := range content_database.ContentCategories {
		data.Filters = append(data.Filters, cat.TypeName())
	}
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/content_workspace.go.html", data, map[string]func(*document.Element){
			"inputFilter": w.inputFilter,
			"tagFilter":   w.tagFilter,
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
				res, err := content_database.Import(paths[i], w.pfs, w.cCache)
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
	idIter, err := w.cCache.AllIds()
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
	ccAll := make([]content_database.CachedContent, 0, len(ids))
	for i := range ids {
		cc, err := w.cCache.Read(ids[i])
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
	w.runFilter()
}

func (w *Workspace) tagFilter(e *document.Element) {
	w.queryTag = strings.ToLower(e.UI.ToInput().Text())
	// TODO:  Filter the list of tag buttons to matching ones
}

func (w *Workspace) clickFilter(e *document.Element) {
	isSelected := slices.Contains(e.ClassList(), "filterSelected")
	isSelected = !isSelected
	typeName := e.Attribute("data-type")
	tagName := e.Attribute("data-tag")
	if isSelected {
		w.Doc.SetElementClasses(e, "leftBtn", "filterSelected")
		if typeName != "" {
			w.typeFilters = append(w.typeFilters, typeName)
		}
		if tagName != "" {
			w.tagFilters = append(w.tagFilters, typeName)
		}
	} else {
		w.Doc.SetElementClasses(e, "leftBtn")
		if typeName != "" {
			w.typeFilters = klib.SlicesRemoveElement(w.typeFilters, typeName)
		}
		if tagName != "" {
			w.tagFilters = klib.SlicesRemoveElement(w.tagFilters, typeName)
		}
	}
	w.runFilter()
}

func (w *Workspace) runFilter() {
	entries := w.Doc.GetElementsByGroup("entry")
	for i := range entries {
		e := entries[i]
		id := e.Attribute("id")
		if id == "entryTemplate" {
			continue
		}
		show := len(w.typeFilters) == 0 || slices.Contains(w.typeFilters, e.Attribute("data-type"))
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
	cc, err := w.cCache.Read(id)
	if err != nil {
		return false
	}
	// TODO:  Use filters like tag:..., type:..., etc.
	if strings.Contains(cc.Config.NameLower(), w.query) {
		return true
	}
	for i := range cc.Config.Tags {
		if slices.Contains(w.tagFilters, cc.Config.Tags[i]) {
			return true
		}
	}
	return false
}

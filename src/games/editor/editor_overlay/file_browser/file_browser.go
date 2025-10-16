package file_browser

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/games/editor/editor_overlay/input_prompt"
	"kaiju/klib"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
)

type FileBrowser struct {
	doc           *document.Document
	uiMan         ui.Manager
	entryListElm  *document.Element
	entryTemplate *document.Element
	filePath      *document.Element
	selected      []*document.Element
	history       []string
	historyIdx    int
	config        FileBrowserConfig
}

type FileBrowserConfig struct {
	OnConfirm   func(paths []string)
	OnCancel    func()
	OnlyFolders bool
	MultiSelect bool
	ExtFilter   []string
}

type FileBrowserData struct {
	QuickAccessFolders []QuickAccessFolder
	CurrentPath        string
	OnlyFolders        bool
}

type QuickAccessFolder struct {
	Name string
	Path string
}

func Show(host *engine.Host, config FileBrowserConfig) (*FileBrowser, error) {
	defer tracing.NewRegion("file_browser.Show").End()
	fb := &FileBrowser{
		historyIdx: -1,
		config:     config,
	}
	fb.uiMan.Init(host)
	var err error
	data := FileBrowserData{
		CurrentPath:        "C:\\",
		QuickAccessFolders: []QuickAccessFolder{},
		OnlyFolders:        fb.config.OnlyFolders,
	}
	for k, v := range filesystem.KnownDirectories() {
		data.QuickAccessFolders = append(data.QuickAccessFolders, QuickAccessFolder{
			Name: k,
			Path: v,
		})
	}
	fb.doc, err = markup.DocumentFromHTMLAsset(&fb.uiMan, "editor/ui/overlay/file_browser.go.html",
		data, map[string]func(*document.Element){
			"selectQuickAccess": fb.selectQuickAccess,
			"upFolder":          fb.upFolder,
			"back":              fb.back,
			"forward":           fb.forward,
			"reload":            fb.reload,
			"newFolder":         fb.newFolder,
			"selectEntry":       fb.selectEntry,
			"openEntry":         fb.openEntry,
			"confirmSelection":  fb.confirmSelection,
			"cancel":            fb.cancel,
			"setPath":           fb.setPath,
		})
	if err != nil {
		slog.Error("failed to load the file browser overlay", "error", err)
		return fb, err
	}
	fb.entryListElm, _ = fb.doc.GetElementById("entryList")
	fb.entryTemplate, _ = fb.doc.GetElementById("entryTemplate")
	fb.filePath, _ = fb.doc.GetElementById("filePath")
	fb.entryTemplate.UI.Hide()
	fb.openFolder(data.CurrentPath)
	return fb, nil
}

func (fb *FileBrowser) Close() { fb.doc.Destroy() }

func (fb *FileBrowser) currentFolder() string {
	return fb.filePath.UI.ToInput().Text()
}

func (fb *FileBrowser) openFolder(folder string) {
	defer tracing.NewRegion("FileBrowser.openFolder").End()
	entries, err := os.ReadDir(folder)
	if err != nil {
		slog.Error("failed to read the directory for the file browser", "folder", folder, "error", err)
		return
	}
	fb.clearSelection()
	if fb.historyIdx < 0 || fb.history[fb.historyIdx] != folder {
		fb.history = append(fb.history[:fb.historyIdx+1], folder)
		fb.historyIdx++
	}
	fb.filePath.UI.ToInput().SetText(folder)
	for i := len(fb.entryListElm.Children) - 1; i >= 1; i-- {
		fb.doc.RemoveElement(fb.entryListElm.Children[i])
	}
	filtered := make([]os.DirEntry, 0, len(entries))
	for i := range entries {
		if !entries[i].IsDir() {
			if fb.config.OnlyFolders {
				continue
			}
			valid := len(fb.config.ExtFilter) == 0 ||
				slices.Contains(fb.config.ExtFilter, filepath.Ext(entries[i].Name()))
			if !valid {
				continue
			}
		}
		filtered = append(filtered, entries[i])
	}
	elmCopies := fb.doc.DuplicateElementRepeat(fb.entryTemplate, len(filtered))
	for i := range filtered {
		entry := elmCopies[i]
		entry.SetAttribute("data-path", filepath.Join(fb.currentFolder(), entries[i].Name()))
		name := entry.FindElementByTag("span")
		name.Children[0].UI.ToLabel().SetText(entries[i].Name())
	}
	fb.entryListElm.UI.ToPanel().ResetScroll()
}

func (fb *FileBrowser) selectQuickAccess(e *document.Element) {
	defer tracing.NewRegion("FileBrowser.selectQuickAccess").End()
	fb.openFolder(e.Attribute("data-path"))
}

func (fb *FileBrowser) upFolder(*document.Element) {
	defer tracing.NewRegion("FileBrowser.upFolder").End()
	f := fb.currentFolder()
	fb.openFolder(filepath.Clean(filepath.Join(f, "../")))
}

func (fb *FileBrowser) back(*document.Element) {
	defer tracing.NewRegion("FileBrowser.back").End()
	fb.historyIdx = max(0, fb.historyIdx-1)
	fb.openFolder(fb.history[fb.historyIdx])
}

func (fb *FileBrowser) forward(*document.Element) {
	defer tracing.NewRegion("FileBrowser.forward").End()
	fb.historyIdx = min(len(fb.history)-1, fb.historyIdx+1)
	fb.openFolder(fb.history[fb.historyIdx])
}

func (fb *FileBrowser) reload(*document.Element) {
	defer tracing.NewRegion("FileBrowser.reload").End()
	fb.execReload()
}

func (fb *FileBrowser) execReload() {
	defer tracing.NewRegion("FileBrowser.execReload").End()
	fb.openFolder(fb.currentFolder())
}

func (fb *FileBrowser) newFolder(*document.Element) {
	defer tracing.NewRegion("FileBrowser.newFolder").End()
	fb.uiMan.DisableUpdate()
	confirm := func(name string) {
		fb.uiMan.EnableUpdate()
		path := filepath.Join(fb.currentFolder(), name)
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			slog.Error("failed to create the new folder", "path", path, "error", err)
			return
		}
		fb.execReload()
	}
	cancel := func() { fb.uiMan.EnableUpdate() }
	input_prompt.Show(fb.uiMan.Host, input_prompt.InputPromptConfig{
		Title:       "New Folder",
		Description: "Input the name for the new folder",
		Placeholder: "Name...",
		Value:       "New Folder",
		ConfirmText: "Create",
		CancelText:  "Cancel",
		OnConfirm:   confirm,
		OnCancel:    cancel,
	})
}

func (fb *FileBrowser) clearSelection() {
	defer tracing.NewRegion("FileBrowser.clearSelection").End()
	for i := range fb.selected {
		fb.doc.SetElementClasses(fb.selected[i], "entry")
	}
	fb.selected = klib.WipeSlice(fb.selected)
}

func (fb *FileBrowser) selectEntry(e *document.Element) {
	defer tracing.NewRegion("FileBrowser.selectEntry").End()
	kb := &fb.uiMan.Host.Window.Keyboard
	if !fb.config.MultiSelect || (!kb.HasCtrl() && !kb.HasShift()) {
		fb.clearSelection()
	} else if slices.Contains(fb.selected, e) {
		idx := slices.Index(fb.selected, e)
		fb.selected = klib.RemoveUnordered(fb.selected, idx)
		fb.doc.SetElementClasses(e, "entry")
		return
	}
	fb.doc.SetElementClasses(e, "entry", "selected")
	fb.selected = append(fb.selected, e)
}

func (fb *FileBrowser) openEntry(e *document.Element) {
	defer tracing.NewRegion("FileBrowser.openEntry").End()
	path := e.Attribute("data-path")
	if s, err := os.Stat(path); err != nil {
		slog.Error("unknown path has been selected", "path", path, "error", err)
		return
	} else if s.IsDir() {
		fb.openFolder(path)
	}
	if fb.config.OnlyFolders {
		return
	}
	fb.confirmSelection(e)
}

func (fb *FileBrowser) confirmSelection(*document.Element) {
	defer tracing.NewRegion("FileBrowser.confirmSelection").End()
	paths := make([]string, 0, len(fb.selected))
	for i := range fb.selected {
		paths = append(paths, fb.selected[i].Attribute("data-path"))
	}
	fb.Close()
	if fb.config.OnConfirm == nil {
		slog.Error("the OnConfirm call has not been set, nothing to do")
		return
	}
	fb.config.OnConfirm(paths)
}

func (fb *FileBrowser) cancel(*document.Element) {
	defer tracing.NewRegion("FileBrowser.cancel").End()
	fb.Close()
	if fb.config.OnCancel != nil {
		fb.config.OnCancel()
	}
}

func (fb *FileBrowser) setPath(*document.Element) {
	defer tracing.NewRegion("FileBrowser.setPath").End()
	fb.openFolder(fb.currentFolder())
}

package file_browser

import (
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/filesystem"
	"log/slog"
	"os"
	"path/filepath"
)

type FileBrowser struct {
	doc           *document.Document
	uiMan         ui.Manager
	entryListElm  *document.Element
	entryTemplate *document.Element
	filePath      *document.Element
	history       []string
	historyIdx    int
	onlyFolders   bool
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

func ShowFolderBrowser(host *engine.Host) (*FileBrowser, error) {
	fb := &FileBrowser{
		historyIdx:  -1,
		onlyFolders: true,
	}
	fb.uiMan.Init(host)
	var err error
	data := FileBrowserData{
		CurrentPath:        "C:\\",
		QuickAccessFolders: []QuickAccessFolder{},
		OnlyFolders:        fb.onlyFolders,
	}
	for k, v := range filesystem.KnownDirectories() {
		data.QuickAccessFolders = append(data.QuickAccessFolders, QuickAccessFolder{
			Name: k,
			Path: v,
		})
	}
	fb.doc, err = markup.DocumentFromHTMLAsset(&fb.uiMan, "editor/ui/file_browser.go.html",
		data, map[string]func(*document.Element){
			"selectQuickAccess": fb.selectQuickAccess,
			"upFolder":          fb.upFolder,
			"back":              fb.back,
			"forward":           fb.forward,
			"reload":            fb.reload,
			"newFolder":         fb.newFolder,
			"selectEntry":       fb.selectEntry,
			"selectFolder":      fb.selectFolder,
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

func (fb *FileBrowser) currentFolder() string {
	return fb.filePath.UI.ToInput().Text()
}

func (fb *FileBrowser) openFolder(folder string) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		slog.Error("failed to read the directory for the file browser", "folder", folder, "error", err)
		return
	}
	if fb.historyIdx < 0 || fb.history[fb.historyIdx] != folder {
		fb.history = append(fb.history[:fb.historyIdx+1], folder)
		fb.historyIdx++
	}
	fb.filePath.UI.ToInput().SetText(folder)
	for i := len(fb.entryListElm.Children) - 1; i >= 1; i-- {
		fb.doc.RemoveElement(fb.entryListElm.Children[i])
	}
	for i := range entries {
		isDir := entries[i].IsDir()
		if fb.onlyFolders && !isDir {
			continue
		}
		isFolder := "0"
		if isDir {
			isFolder = "1"
		}
		entry := fb.doc.DuplicateElement(fb.entryTemplate)
		entry.SetAttribute("data-path", filepath.Join(fb.currentFolder(), entries[i].Name()))
		entry.SetAttribute("data-is-folder", isFolder)
		name := entry.FindElementByTag("span")
		name.Children[0].UI.ToLabel().SetText(entries[i].Name())
	}
	fb.entryListElm.UI.ToPanel().ResetScroll()
}

func (fb *FileBrowser) selectQuickAccess(e *document.Element) {
	fb.openFolder(e.Attribute("data-path"))
}

func (fb *FileBrowser) upFolder(*document.Element) {
	f := fb.currentFolder()
	fb.openFolder(filepath.Clean(filepath.Join(f, "../")))
}

func (fb *FileBrowser) back(*document.Element) {
	fb.historyIdx = max(0, fb.historyIdx-1)
	fb.openFolder(fb.history[fb.historyIdx])
}

func (fb *FileBrowser) forward(*document.Element) {
	fb.historyIdx = min(len(fb.history)-1, fb.historyIdx+1)
	fb.openFolder(fb.history[fb.historyIdx])
}

func (fb *FileBrowser) reload(*document.Element) {
	fb.openFolder(fb.history[fb.historyIdx])
}

func (fb *FileBrowser) newFolder(*document.Element) {
	debug.ThrowNotImplemented("need to show input prompt overlay for the folder name")
}

func (fb *FileBrowser) selectEntry(e *document.Element) {
	if e.Attribute("data-is-folder") != "0" {
		fb.openFolder(e.Attribute("data-path"))
		return
	}
	if fb.onlyFolders {
		return
	}
	debug.ThrowNotImplemented("This file has been selected, raise event")
}

func (fb *FileBrowser) selectFolder(*document.Element) {
	//fb.currentFolder()
	debug.ThrowNotImplemented("This folder has been selected, raise event")
}

func (fb *FileBrowser) setPath(*document.Element) {
	fb.openFolder(fb.currentFolder())
}

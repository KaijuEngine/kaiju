/******************************************************************************/
/* file_browser_overlay.go                                                    */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package file_browser

import (
	"kaiju/editor/editor_overlay/input_prompt"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/filesystem"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"unicode"
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
	config        Config
}

type Config struct {
	Title       string
	ExtFilter   []string
	OnConfirm   func(paths []string)
	OnCancel    func()
	OnlyFiles   bool
	OnlyFolders bool
	MultiSelect bool
}

type FileBrowserData struct {
	Title              string
	QuickAccessFolders []QuickAccessFolder
	CurrentPath        string
	OnlyFolders        bool
}

type QuickAccessFolder struct {
	Name string
	Path string
}

func Show(host *engine.Host, config Config) (*FileBrowser, error) {
	defer tracing.NewRegion("file_browser.Show").End()
	fb := &FileBrowser{
		historyIdx: -1,
		config:     config,
	}
	fb.uiMan.Init(host)
	var err error
	title := fb.config.Title
	if title == "" {
		if fb.config.OnlyFolders {
			if fb.config.MultiSelect {
				title = "Select one or more folders"
			} else {
				title = "Select a folder"
			}
		} else {
			if fb.config.MultiSelect {
				title = "Select one or more files"
			} else {
				title = "Select a file"
			}
		}
	}
	data := FileBrowserData{
		Title:              title,
		CurrentPath:        "C:\\",
		QuickAccessFolders: []QuickAccessFolder{},
		OnlyFolders:        fb.config.OnlyFolders,
	}
	knownDirs := filesystem.KnownDirectories()
	for _, k := range klib.MapKeysSorted(knownDirs) {
		data.QuickAccessFolders = append(data.QuickAccessFolders, QuickAccessFolder{
			Name: k,
			Path: knownDirs[k],
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
	updateId := host.Updater.AddUpdate(fb.update)
	keyCallbackId := host.Window.Keyboard.AddKeyCallback(fb.onKeyboardType)
	fb.doc.Elements[0].UI.Entity().OnDestroy.Add(func() {
		fb.uiMan.Host.Updater.RemoveUpdate(&updateId)
		fb.uiMan.Host.Window.Keyboard.RemoveKeyCallback(keyCallbackId)
	})
	return fb, nil
}

func (fb *FileBrowser) onKeyboardType(keyId int, keyState hid.KeyState) {
	if keyState != hid.KeyStateDown {
		return
	}
	r := fb.uiMan.Host.Window.Keyboard.KeyToRune(keyId)
	if r == 0 {
		return
	}
	from := 0
	if len(fb.selected) > 0 {
		from = fb.entryListElm.IndexOfChild(fb.selected[len(fb.selected)-1]) + 1
	}
	locate := func(start int) bool {
		for _, c := range fb.entryListElm.Children[start:] {
			name := c.Children[2].Children[0].UI.ToLabel().Text()
			if unicode.ToLower([]rune(name)[0]) == unicode.ToLower(r) {
				fb.selectEntry(c)
				fb.entryListElm.UIPanel.ScrollToChild(c.UI)
				return true
			}
		}
		return false
	}
	fb.clearSelection()
	if !locate(from) && from > 0 {
		locate(0)
	}
}

func (fb *FileBrowser) update(float64) {
	if len(fb.entryListElm.Children) == 0 {
		return
	}
	if fb.uiMan.Group.HasRequests() {
		return
	}
	kb := &fb.uiMan.Host.Window.Keyboard
	if kb.KeyDown(hid.KeyboardKeyUp) || kb.KeyDown(hid.KeyboardKeyDown) {
		// We start at 1 because of the template being 0
		idx := 1
		if len(fb.selected) > 0 {
			last := fb.selected[len(fb.selected)-1]
			if kb.KeyDown(hid.KeyboardKeyUp) {
				idx = max(1, fb.entryListElm.IndexOfChild(last)-1)
			} else if kb.KeyDown(hid.KeyboardKeyDown) {
				idx = min(len(fb.entryListElm.Children)-1,
					fb.entryListElm.IndexOfChild(last)+1)
			}
		}
		fb.clearSelection()
		fb.selectEntry(fb.entryListElm.Children[idx])
		p := fb.entryListElm.UI.ToPanel()
		p.ScrollToChild(fb.entryListElm.Children[idx].UI)
	} else if kb.KeyDown(hid.KeyboardKeyEnter) || kb.KeyDown(hid.KeyboardKeyReturn) {
		if len(fb.selected) == 1 {
			fb.openEntry(fb.selected[0])
		} else {
			fb.confirmSelection(nil)
		}
	} else if kb.KeyDown(hid.KeyboardKeyLeft) || kb.KeyDown(hid.KeyboardKeyBackspace) {
		fb.back(nil)
	} else if kb.KeyDown(hid.KeyboardKeyRight) {
		fb.forward(nil)
	}
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
	entries = klib.SortDirEntries(entries)
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
		if filtered[i].IsDir() {
			entry.Children[1].UI.Hide()
		} else {
			entry.Children[0].UI.Hide()
		}
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
	input_prompt.Show(fb.uiMan.Host, input_prompt.Config{
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
		p := fb.selected[i].Attribute("data-path")
		if fb.config.OnlyFiles {
			if s, err := os.Stat(p); err != nil || s.IsDir() {
				continue
			}
		}
		paths = append(paths, p)
	}
	if fb.config.OnConfirm == nil {
		slog.Error("the OnConfirm call has not been set, nothing to do")
		return
	}
	if len(paths) == 0 && fb.config.OnlyFolders {
		paths = append(paths, fb.filePath.UI.ToInput().Text())
	}
	if len(paths) == 0 {
		return
	}
	fb.Close()
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

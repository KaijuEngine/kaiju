/******************************************************************************/
/* content_workspace_table_of_contents.go                                     */
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

package content_workspace

import (
	"fmt"
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/editor_overlay/input_prompt"
	"kaiju/editor/editor_overlay/table_of_contents_overlay"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine/assets/table_of_contents"
	"kaiju/klib"
	"log/slog"
	"os"
	"strings"
)

func (w *ContentWorkspace) requestCreateTableOfContents() {
	if len(w.selectedContent) == 0 {
		slog.Warn("nothing selected to create a table of contents for")
		return
	}
	w.editor.BlurInterface()
	input_prompt.Show(w.Host, input_prompt.Config{
		Title:       "Table of Contents Name",
		Description: "Give a friendly name to your table of contents",
		Placeholder: "Table of contents name...",
		ConfirmText: "Create",
		CancelText:  "Cancel",
		OnConfirm: func(name string) {
			w.editor.FocusInterface()
			if w.checkNoDuplicateNamesForTableOfContents() {
				w.createTableOfContents(name)
			}
		},
		OnCancel: w.editor.FocusInterface,
	})
}

func (w *ContentWorkspace) checkNoDuplicateNamesForTableOfContents() bool {
	dupes := duplicateNames(w.selectedNames())
	if len(dupes) == 0 {
		return true
	}
	w.editor.BlurInterface()
	confirm_prompt.Show(w.Host, confirm_prompt.Config{
		Title:       "Duplicate names",
		Description: fmt.Sprintf("The action can not be completed because there is content with duplicate names, please fix these before doing this table of contents operation: %s", strings.Join(dupes, ", ")),
		CancelText:  "Close",
		OnCancel:    w.editor.FocusInterface,
	})
	return false
}

func (w *ContentWorkspace) selectedNames() []string {
	names := make([]string, 0, len(w.selectedContent))
	for i := range w.selectedContent {
		id := w.selectedContent[i].Attribute("id")
		cc, err := w.cache.Read(id)
		if err != nil {
			slog.Warn("failed to find cached data for id", "id", id, "error", err)
			continue
		}
		names = append(names, cc.Config.Name)
	}
	return names
}

func duplicateNames(names []string) []string {
	dupes := []string{}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] == names[j] {
				dupes = klib.AppendUnique(dupes, names[i])
			}
		}
	}
	return dupes
}

func (w *ContentWorkspace) createTableOfContents(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		slog.Warn("blank name supplied for creating a table of contents, skipping creation")
		return
	}
	if len(w.selectedContent) == 0 {
		slog.Warn("nothing selected to create a table of contents for")
		return
	}
	toc := table_of_contents.New()
	for i := range w.selectedContent {
		id := w.selectedContent[i].Attribute("id")
		cc, err := w.cache.Read(id)
		if err != nil {
			slog.Warn("failed to find cached data for id", "id", id, "error", err)
			continue
		}
		entry := table_of_contents.TableEntry{
			Id:   id,
			Name: cc.Config.Name,
		}
		for !toc.Add(entry) {
			entry.Name += "_1"
		}
	}
	data, err := toc.Serialize()
	if err != nil {
		slog.Error("failed to serialize the table of contents", "error", err)
		return
	}
	f, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-*.toc", name))
	if err != nil {
		slog.Error("failed to create temp content file for table of contents", "error", err)
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err = f.Write(data); err != nil {
		slog.Error("failed to write the temp content file for table of contents", "file", f.Name(), "error", err)
		return
	}
	res, err := content_database.Import(f.Name(), w.pfs, w.cache, "")
	if err != nil {
		slog.Error("failed to import the temp content file for table of contents", "file", f.Name(), "error", err)
		return
	}
	ids := make([]string, len(res))
	for i := range res {
		ids[i] = res[i].Id
	}
	defer w.editor.Events().OnContentAdded.Execute(ids)
	if len(res) != 1 {
		slog.Warn("table of contents created but name has not been set due to unexpected result count from import")
		return
	}
	cc, err := w.cache.Read(res[0].Id)
	if err != nil {
		slog.Warn("failed to find the cache for the table of contents that was just imported, name is unset")
		return
	}
	cc.Config.Name = name
	cc.Config.SrcPath = ""
	if err := content_database.WriteConfig(cc.Path, cc.Config, w.pfs); err != nil {
		slog.Warn("failed to update the name of the table of contents", "id", res[0].Id, "error", err)
		return
	}
	w.cache.Index(cc.Path, w.pfs)
}

func (w *ContentWorkspace) loadTableOfContents(id string) (table_of_contents.TableOfContents, bool) {
	cc, err := w.cache.Read(id)
	if err != nil {
		slog.Error("failed to find the table of contents in cache", "id", id, "error", err)
		return table_of_contents.TableOfContents{}, false
	}
	path := content_database.ToContentPath(cc.Path)
	data, err := w.pfs.ReadFile(path)
	if err != nil {
		slog.Error("failed to read the table of contents file", "path", path, "error", err)
		return table_of_contents.TableOfContents{}, false
	}
	toc, err := table_of_contents.Deserialize(data)
	if err != nil {
		slog.Error("failed to deserialize the table of contents file", "path", path, "error", err)
		return toc, false
	}
	return toc, true
}

func (w *ContentWorkspace) addSelectedToTableOfContents(id string) {
	toc, ok := w.loadTableOfContents(id)
	if !ok {
		return
	}
	for i := range w.selectedContent {
		sid := w.selectedContent[i].Attribute("id")
		if _, ok := toc.SelectById(sid); ok {
			slog.Warn("the content is already in the table of contents", "id", sid)
			continue
		}
		icc, err := w.cache.Read(sid)
		if err != nil {
			slog.Error("failed to add the content to the table of contents", "id", id, "error", err)
			continue
		}
		entry := table_of_contents.TableEntry{
			Id:   sid,
			Name: icc.Config.Name,
		}
		for !toc.Add(entry) {
			entry.Name += "_1"
		}
		slog.Info("added content to table of contents", "id", sid, "name", entry.Name)
	}
	w.saveTableOfContents(id, toc)
}

func (w *ContentWorkspace) showTableOfContents(id string) {
	toc, ok := w.loadTableOfContents(id)
	if !ok {
		slog.Error("failed to load the table of contents view")
		return
	}
	w.editor.BlurInterface()
	table_of_contents_overlay.Show(w.Host, table_of_contents_overlay.Config{
		TOC: toc,
		OnChanged: func(newToc table_of_contents.TableOfContents) {
			w.saveTableOfContents(id, newToc)
		},
		OnClose: w.editor.FocusInterface,
	})
}

func (w *ContentWorkspace) saveTableOfContents(id string, toc table_of_contents.TableOfContents) {
	cc, err := w.cache.Read(id)
	if err != nil {
		slog.Error("failed to read the cached content for table of contents", "id", id, "error", err)
		return
	}
	path := content_database.ToContentPath(cc.Path)
	data, err := toc.Serialize()
	if err != nil {
		slog.Error("failed to serialize the table of contents to file", "id", id, "error", err)
		return
	}
	if s, err := os.Stat(w.pfs.FullPath(path)); err == nil {
		if err = w.pfs.WriteFile(path, data, s.Mode()); err != nil {
			slog.Error("failed to write the table of contents file", "id", id, "error", err)
			return
		}
	} else {
		slog.Error("failed to locate the table of contents file in the database", "id", id)
		return
	}
	slog.Info("updated the table of contents file", "id", id)
}

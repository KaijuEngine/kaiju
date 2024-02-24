/******************************************************************************/
/* file_window.go                                                             */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                      */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package files_window

import (
	"io/fs"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
)

type FileWindow struct {
	doc        *document.Document
	input      *ui.Input
	listing    *ui.Panel
	container  *host_container.Container
	Dir        []fs.DirEntry
	Path       string
	Extensions []string
	funcMap    map[string]func(*document.DocElement)
	Folders    bool
	selected   bool
	done       chan string
}

// Creates a window allowing the person to select any file or folder
func Any(title string) chan string {
	return create(title, false, nil)
}

// Creates a window allowing the person to select a folder
func Folder(title string) chan string {
	return create(title, true, nil)
}

// Creates a window allowing the person to select a file with the given extensions
func Files(title string, extensions []string) chan string {
	return create(title, false, extensions)
}

func create(title string, foldersOnly bool, extensions []string) chan string {
	if title == "" {
		title = "File/Folder Select"
	}
	s := FileWindow{
		funcMap:    make(map[string]func(*document.DocElement)),
		Extensions: make([]string, 0, len(extensions)),
		Folders:    foldersOnly,
		done:       make(chan string),
	}
	for _, ext := range extensions {
		if ext != "" {
			if ext[0] != '.' {
				ext = "." + ext
			}
			s.Extensions = append(s.Extensions, ext)
		}
	}
	s.funcMap["selectEntry"] = s.selectEntry
	s.funcMap["selectPath"] = s.selectPath
	s.container = host_container.New(title, nil)
	go s.container.Run(500, 600, -1, -1)
	<-s.container.PrepLock
	if here, err := os.Getwd(); err != nil {
		s.Path = "/"
	} else {
		s.Path = here
	}
	s.container.RunFunction(func() {
		s.reloadUI()
	})
	s.container.Host.OnClose.Add(func() {
		if !s.selected {
			s.done <- ""
		}
		close(s.done)
	})
	return s.done
}

func (s *FileWindow) CanSelectFolder() bool {
	return s.Folders || len(s.Extensions) == 0
}

func (s *FileWindow) selectPath(*document.DocElement) {
	s.done <- s.Path
	s.selected = true
	s.container.Host.Close()
}

func (s *FileWindow) selectEntry(elm *document.DocElement) {
	path := elm.HTML.Attribute("data-path")
	s.Path = filepath.Clean(s.Path + "/" + path)
	if info, err := os.Stat(s.Path); err != nil {
		slog.Error(err.Error())
		s.container.Host.Close()
		return
	} else {
		if info.IsDir() {
			s.reloadUI()
		} else {
			s.done <- s.Path
			s.selected = true
			s.container.Host.Close()
		}
	}
}

func (s *FileWindow) reloadUI() {
	for _, e := range s.container.Host.Entities() {
		e.Destroy()
	}
	s.list()
	html := klib.MustReturn(s.container.Host.AssetDatabase().ReadText("editor/ui/file_window.html"))
	s.doc = markup.DocumentFromHTMLString(
		s.container.Host, html, "", s, s.funcMap)
	if elm, ok := s.doc.GetElementById("pathInput"); !ok {
		slog.Error(`Failed to locate the "pathInput" for the file window`)
		s.container.Host.Close()
		return
	} else {
		s.input = elm.UI.(*ui.Input)
		s.input.Data().OnSubmit.Add(s.submit)
	}
	if elm, ok := s.doc.GetElementById("listing"); !ok {
		slog.Error(`Failed to locate the "listing" for the file window`)
		s.container.Host.Close()
		return
	} else {
		s.listing = elm.UIPanel
	}
}

func (s *FileWindow) submit() {
	s.Path = s.input.Text()
	s.reloadUI()
}

func (s *FileWindow) list() {
	dir, err := os.ReadDir(s.Path)
	if err != nil {
		slog.Error(err.Error())
		s.container.Host.Close()
		return
	}
	if len(s.Extensions) > 0 {
		s.Dir = make([]fs.DirEntry, 0, len(dir))
		for i := range dir {
			if dir[i].IsDir() || slices.Contains(s.Extensions, filepath.Ext(dir[i].Name())) {
				s.Dir = append(s.Dir, dir[i])
			}
		}
	} else if s.Folders {
		s.Dir = make([]fs.DirEntry, 0, len(dir))
		for i := range dir {
			if dir[i].IsDir() {
				s.Dir = append(s.Dir, dir[i])
			}
		}
	} else {
		s.Dir = dir
	}
}

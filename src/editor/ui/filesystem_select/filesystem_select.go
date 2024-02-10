package filesystem_select

import (
	"io/fs"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
	"os"
	"path/filepath"
	"slices"
)

type FilesystemSelect struct {
	doc        *document.Document
	input      *ui.Input
	listing    *ui.Panel
	onSelected func(string)
	container  *host_container.HostContainer
	Dir        []fs.DirEntry
	Path       string
	Extensions []string
	funcMap    map[string]func(*document.DocElement)
}

// Will create a new window that allows the person to select a file or folder
// on their system. If the extensions are not empty, then only files with those
// extensions will be selectable. Otherwise all files or folders can be picked
func New(title string, extensions []string, onSelected func(string)) {
	if title == "" {
		title = "File/Folder Select"
	}
	s := FilesystemSelect{
		onSelected: onSelected,
		funcMap:    make(map[string]func(*document.DocElement)),
		Extensions: make([]string, 0, len(extensions)),
	}
	for _, ext := range extensions {
		if ext[0] != '.' {
			ext = "." + ext
		}
		s.Extensions = append(s.Extensions, ext)
	}
	s.funcMap["selectEntry"] = s.selectEntry
	s.container = host_container.New(title)
	go s.container.Run(500, 600)
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
		s.onSelected("")
	})
}

func (s *FilesystemSelect) selectPath() {
	s.container.Host.Close()
	s.onSelected(s.Path)
}

func (s *FilesystemSelect) selectEntry(elm *document.DocElement) {
	path := elm.HTML.Attribute("data-path")
	s.Path = filepath.Clean(s.Path + "/" + path)
	if info, err := os.Stat(s.Path); err != nil {
		s.container.Host.Close()
		return
	} else {
		if info.IsDir() {
			s.reloadUI()
		} else {
			s.container.Host.Close()
			s.onSelected(s.Path)
		}
	}
}

func (s *FilesystemSelect) reloadUI() {
	for _, e := range s.container.Host.Entities() {
		e.Destroy()
	}
	s.list()
	html := klib.MustReturn(s.container.Host.AssetDatabase().ReadText("ui/editor/filesystem_select.html"))
	s.doc = markup.DocumentFromHTMLString(
		s.container.Host, html, "", s, s.funcMap)
	if elm, ok := s.doc.GetElementById("pathInput"); !ok {
		// TODO:  Log the error
		s.container.Host.Close()
		return
	} else {
		s.input = elm.UI.(*ui.Input)
		s.input.Data().OnSubmit.Add(s.submit)
	}
	if elm, ok := s.doc.GetElementById("listing"); !ok {
		// TODO:  Log the error
		s.container.Host.Close()
		return
	} else {
		s.listing = elm.UIPanel
	}
}

func (s *FilesystemSelect) submit() {
	s.Path = s.input.Text()
	s.reloadUI()
}

func (s *FilesystemSelect) list() {
	dir, err := os.ReadDir(s.Path)
	if err != nil {
		// TODO:  Report the error
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
	} else {
		s.Dir = dir
	}
}

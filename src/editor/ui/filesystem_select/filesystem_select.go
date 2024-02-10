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
)

type FilesystemSelect struct {
	doc        *document.Document
	input      *ui.Input
	listing    *ui.Panel
	onSelected func(string)
	container  *host_container.HostContainer
	Dir        []fs.DirEntry
	Path       string
	funcMap    map[string]func(*document.DocElement)
}

func New(title string, onSelected func(string)) {
	if title == "" {
		title = "File/Folder Select"
	}
	fs := FilesystemSelect{
		onSelected: onSelected,
		funcMap:    make(map[string]func(*document.DocElement)),
	}
	fs.funcMap["selectEntry"] = fs.selectEntry
	fs.container = host_container.New(title)
	go fs.container.Run(500, 600)
	<-fs.container.PrepLock
	if here, err := os.Getwd(); err != nil {
		fs.Path = "/"
	} else {
		fs.Path = here
	}
	fs.container.RunFunction(func() {
		fs.reloadUI()
	})
	fs.container.Host.OnClose.Add(func() {
		fs.onSelected("")
	})
}

func (fs *FilesystemSelect) selectPath() {
	fs.container.Host.Close()
	fs.onSelected(fs.Path)
}

func (fs *FilesystemSelect) selectEntry(elm *document.DocElement) {
	path := elm.HTML.Attribute("data-path")
	fs.Path = filepath.Clean(fs.Path + "/" + path)
	if info, err := os.Stat(fs.Path); err != nil {
		fs.container.Host.Close()
		return
	} else {
		if info.IsDir() {
			fs.reloadUI()
		} else {
			fs.container.Host.Close()
			fs.onSelected(fs.Path)
		}
	}
}

func (fs *FilesystemSelect) reloadUI() {
	for _, e := range fs.container.Host.Entities() {
		e.Destroy()
	}
	fs.list()
	html := klib.MustReturn(fs.container.Host.AssetDatabase().ReadText("ui/editor/filesystem_select.html"))
	fs.doc = markup.DocumentFromHTMLString(
		fs.container.Host, html, "", fs, fs.funcMap)
	if elm, ok := fs.doc.GetElementById("pathInput"); !ok {
		// TODO:  Log the error
		fs.container.Host.Close()
		return
	} else {
		fs.input = elm.UI.(*ui.Input)
	}
	if elm, ok := fs.doc.GetElementById("listing"); !ok {
		// TODO:  Log the error
		fs.container.Host.Close()
		return
	} else {
		fs.listing = elm.UIPanel
	}
}

func (fs *FilesystemSelect) submit() {

}

func (fs *FilesystemSelect) list() {
	//name = "📁 " + name
	//name = "📄 " + name
	var err error
	fs.Dir, err = os.ReadDir(fs.Path)
	if err != nil {
		// TODO:  Report the error
		fs.container.Host.Close()
		return
	}
}

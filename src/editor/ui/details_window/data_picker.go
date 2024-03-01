package details_window

import (
	"kaiju/editor/codegen"
	"kaiju/editor/interfaces"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
	"strconv"
	"strings"
)

type DataPicker struct {
	editor    interfaces.Editor
	container *host_container.Container
	doc       *document.Document
	picked    bool
	lock      chan int
}

func NewDataPicker(host *engine.Host, types []codegen.GeneratedType) chan int {
	const html = "editor/ui/data_picker.html"
	dp := &DataPicker{
		container: host_container.New("Data Select", nil),
		lock:      make(chan int),
	}
	cx, cy := host.Window.Center()
	go dp.container.Run(300, 600, cx-150, cy-300)
	<-dp.container.PrepLock
	dp.container.Host.AssetDatabase().EditorContext.EditorPath = host.AssetDatabase().EditorContext.EditorPath
	dp.container.RunFunction(func() {
		dp.doc, _ = markup.DocumentFromHTMLAsset(dp.container.Host, html, types,
			map[string]func(*document.DocElement){
				"pick":   dp.pick,
				"search": dp.search,
			})
	})
	dp.container.Host.OnClose.Add(func() {
		if !dp.picked {
			dp.lock <- -1
		}
	})
	return dp.lock
}

func (dp *DataPicker) pick(elm *document.DocElement) {
	dp.picked = true
	idx, _ := strconv.Atoi(elm.HTML.Attribute("id"))
	dp.lock <- idx
	dp.container.Close()
}

func (dp *DataPicker) search(elm *document.DocElement) {
	input, _ := dp.doc.GetElementById("search")
	query := strings.ToLower(input.UI.(*ui.Input).Text())
	for i := range dp.doc.Elements {
		name := dp.doc.Elements[i].HTML.Attribute("data-name")
		if name != "" {
			if strings.Contains(strings.ToLower(name), query) {
				dp.doc.Elements[i].UI.Entity().Activate()
			} else {
				dp.doc.Elements[i].UI.Entity().Deactivate()
			}
		}
	}
}

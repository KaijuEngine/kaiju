package shader_designer

import (
	"kaiju/host_container"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
	"slices"
)

type ShaderDesigner struct {
	pipeline    rendering.ShaderPipelineData
	pipelineDoc *document.Document
	man         ui.Manager
}

type flagState struct {
	List    []string
	Current []string
	Array   string
	Field   string
	Index   int
}

func (s flagState) Has(val string) bool {
	return slices.Contains(s.Current, val)
}

func (win *ShaderDesigner) uiInit() {
	setupPipelineDoc(win, &win.man)
}

func setup(onOpen func(*ShaderDesigner)) {
	container := host_container.New("Shader Designer", nil)
	go container.Run(640, 480, -1, -1)
	<-container.PrepLock
	container.RunFunction(func() {
		win := ShaderDesigner{}
		win.man.Init(container.Host)
		win.uiInit()
		if onOpen != nil {
			onOpen(&win)
		}
	})
}

func New() {
	setup(nil)
}

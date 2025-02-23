package shader_designer

import (
	"kaiju/host_container"
	"kaiju/markup/document"
	"kaiju/ui"
)

type ShaderDesigner struct {
	pipeline    ShaderPipelineData
	pipelineDoc *document.Document
	man         ui.Manager
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
		if onOpen != nil {
			onOpen(&win)
		}
	})
}

func New() {
	setup(nil)
}

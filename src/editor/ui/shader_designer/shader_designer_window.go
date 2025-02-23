package shader_designer

import (
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/markup/document"
	"kaiju/ui"
)

type ShaderDesigner struct {
	pipeline    ShaderPipelineData
	pipelineDoc *document.Document
	man         ui.Manager
}

func uiInit(host *engine.Host) {
	win := ShaderDesigner{}
	win.man.Init(host)
	setupPipelineDoc(&win, &win.man)
}

func New() {
	container := host_container.New("Shader Designer", nil)
	go container.Run(640, 480, -1, -1)
	<-container.PrepLock
	container.RunFunction(func() { uiInit(container.Host) })
}

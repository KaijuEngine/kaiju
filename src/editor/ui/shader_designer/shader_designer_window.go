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
}

func uiInit(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	win := ShaderDesigner{}
	setupPipelineDoc(&win, &uiMan)
}

func New() {
	container := host_container.New("Shader Designer", nil)
	go container.Run(640, 480, -1, -1)
	<-container.PrepLock
	container.RunFunction(func() { uiInit(container.Host) })
}

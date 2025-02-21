package shader_designer_window

import (
	"kaiju/host_container"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
)

type ShaderDesignerData struct {
	Vert string
	Frag string
	Geom string
	Tese string
	Tesc string
}

func New() {
	const html = "editor/ui/shader_designer/shader_designer_window.html"
	container := host_container.New("Shader Designer", nil)
	uiMan := ui.Manager{}
	uiMan.Init(container.Host)
	go container.Run(640, 480, -1, -1)
	<-container.PrepLock
	shaderData := ShaderDesignerData{}
	container.RunFunction(func() {
		markup.DocumentFromHTMLAsset(&uiMan, html, shaderData, map[string]func(*document.Element){
			//"openContributions": openContributions,
		})
	})
}

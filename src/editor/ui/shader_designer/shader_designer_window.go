package shader_designer

import (
	"kaiju/host_container"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
	"slices"
)

const (
	shaderDesignerHTML = "editor/ui/shader_designer/shader_designer_window.html"
)

type ShaderDesigner struct {
	pipeline          rendering.ShaderPipelineData
	renderPass        rendering.RenderPassData
	shaderDesignerDoc *document.Document
	pipelineDoc       *document.Document
	renderPassDoc     *document.Document
	man               ui.Manager
}

type flagState struct {
	List    []string
	Current []string
	Path    string
	Array   string
	Field   string
	Index   int
}

func (s flagState) Has(val string) bool {
	return slices.Contains(s.Current, val)
}

func (win *ShaderDesigner) uiInit() {
	setupPipelineDoc(win)
	setupRenderPassDoc(win)
	win.reloadShaderDesigner()
}

func setup(onOpen func(*ShaderDesigner)) {
	container := host_container.New("Shader Designer", nil)
	go container.Run(640, 480, -1, -1)
	<-container.PrepLock
	container.RunFunction(func() {
		win := &ShaderDesigner{}
		win.man.Init(container.Host)
		win.uiInit()
		if onOpen != nil {
			onOpen(win)
		}
	})
}

func New() {
	setup(nil)
}

func (win *ShaderDesigner) ShowDesignerWindow() {
	win.pipelineDoc.Deactivate()
	win.renderPassDoc.Deactivate()
	win.shaderDesignerDoc.Activate()
	win.reloadShaderDesigner()
	win.man.Host.Window.Focus()
}

func (win *ShaderDesigner) ShowPipelineWindow() {
	win.pipelineDoc.Activate()
	win.renderPassDoc.Deactivate()
	win.shaderDesignerDoc.Deactivate()
	win.reloadPipelineDoc()
	win.man.Host.Window.Focus()
}

func (win *ShaderDesigner) ShowRenderPassWindow() {
	win.pipelineDoc.Deactivate()
	win.renderPassDoc.Activate()
	win.shaderDesignerDoc.Deactivate()
	win.reloadRenderPassDoc()
	win.man.Host.Window.Focus()
}

func (win *ShaderDesigner) returnHome(*document.Element) {
	win.ShowDesignerWindow()
}

func (win *ShaderDesigner) reloadShaderDesigner() {
	if win.shaderDesignerDoc != nil {
		win.shaderDesignerDoc.Destroy()
	}
	win.shaderDesignerDoc, _ = markup.DocumentFromHTMLAsset(&win.man, shaderDesignerHTML,
		nil, map[string]func(*document.Element){
			"newRenderPass": func(*document.Element) {
				win.renderPass = rendering.RenderPassData{}
				win.ShowRenderPassWindow()
			},
			"newShaderPipeline": func(*document.Element) {
				win.pipeline = rendering.ShaderPipelineData{}
				win.ShowPipelineWindow()
			},
		})
}

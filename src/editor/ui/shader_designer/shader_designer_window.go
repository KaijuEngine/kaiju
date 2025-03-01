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

type ShaderDesignerState = int

const (
	shaderDesignerStateHome = ShaderDesignerState(iota)
	shaderDesignerStateShader
	shaderDesignerStateRenderPass
	shaderDesignerStatePipeline
	shaderDesignerStateMaterial
)

type ShaderDesigner struct {
	shader            rendering.ShaderData
	renderPass        rendering.RenderPassData
	pipeline          rendering.ShaderPipelineData
	material          rendering.MaterialData
	shaderDoc         *document.Document
	shaderDesignerDoc *document.Document
	pipelineDoc       *document.Document
	renderPassDoc     *document.Document
	materialDoc       *document.Document
	man               ui.Manager
	state             ShaderDesignerState
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
	setupShaderDoc(win)
	setupPipelineDoc(win)
	setupRenderPassDoc(win)
	setupMaterialDoc(win)
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

func (win *ShaderDesigner) ChangeWindowState(state ShaderDesignerState) {
	if win.state == state {
		return
	}
	win.state = state
	win.shaderDoc.Deactivate()
	win.pipelineDoc.Deactivate()
	win.renderPassDoc.Deactivate()
	win.materialDoc.Deactivate()
	win.shaderDesignerDoc.Deactivate()
	switch state {
	case shaderDesignerStateHome:
		win.shaderDesignerDoc.Activate()
		win.reloadShaderDesigner()
	case shaderDesignerStateShader:
		win.shaderDoc.Activate()
		win.reloadShaderDoc()
	case shaderDesignerStateRenderPass:
		win.renderPassDoc.Activate()
		win.reloadRenderPassDoc()
	case shaderDesignerStatePipeline:
		win.pipelineDoc.Activate()
		win.reloadPipelineDoc()
	case shaderDesignerStateMaterial:
		win.materialDoc.Activate()
		win.reloadMaterialDoc()
	}
	win.man.Host.Window.Focus()
}

func (win *ShaderDesigner) ShowDesignerWindow() {
	win.ChangeWindowState(shaderDesignerStateHome)
}

func (win *ShaderDesigner) ShowShaderWindow() {
	win.ChangeWindowState(shaderDesignerStateShader)
}

func (win *ShaderDesigner) ShowRenderPassWindow() {
	win.ChangeWindowState(shaderDesignerStateRenderPass)
}

func (win *ShaderDesigner) ShowPipelineWindow() {
	win.ChangeWindowState(shaderDesignerStatePipeline)
}

func (win *ShaderDesigner) ShowMaterialWindow() {
	win.ChangeWindowState(shaderDesignerStateMaterial)
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
			"newShader": func(*document.Element) {
				win.shader = rendering.ShaderData{}
				win.ShowShaderWindow()
			},
			"newRenderPass": func(*document.Element) {
				win.renderPass = rendering.RenderPassData{}
				win.ShowRenderPassWindow()
			},
			"newShaderPipeline": func(*document.Element) {
				win.pipeline = rendering.ShaderPipelineData{}
				win.ShowPipelineWindow()
			},
			"newMaterial": func(*document.Element) {
				win.material = rendering.MaterialData{}
				win.ShowMaterialWindow()
			},
		})
}

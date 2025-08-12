/******************************************************************************/
/* shader_designer_window.go                                                  */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package shader_designer

import (
	"kaiju/debug"
	"kaiju/editor/ui/tab_container"
	"kaiju/engine/systems/logging"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/rendering"
	"slices"
)

const (
	shaderDesignerHTML = "editor/ui/shader_designer/shader_designer_window.html"
)

type ShaderDesignerState = int

const (
	StateHome = ShaderDesignerState(iota)
	StateShader
	StateRenderPass
	StatePipeline
	StateMaterial
)

type ShaderDesigner struct {
	shader            rendering.ShaderData
	renderPass        rendering.RenderPassData
	pipeline          rendering.ShaderPipelineData
	material          rendering.MaterialData
	shaderDesignerDoc *document.Document
	shaderDoc         *document.Document
	pipelineDoc       *document.Document
	renderPassDoc     *document.Document
	materialDoc       *document.Document
	man               *ui.Manager
	root              *document.Element
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

func (s *ShaderDesigner) TabTitle() string { return "Shader Designer" }

func (s *ShaderDesigner) Document() *document.Document {
	switch s.state {
	case StateShader:
		return s.shaderDoc
	case StateRenderPass:
		return s.renderPassDoc
	case StatePipeline:
		return s.pipelineDoc
	case StateMaterial:
		return s.materialDoc
	case StateHome:
		fallthrough
	default:
		return s.shaderDesignerDoc
	}
}

func (s *ShaderDesigner) Destroy() {
	if s.shaderDesignerDoc != nil {
		s.shaderDesignerDoc.Destroy()
		s.shaderDesignerDoc = nil
	}
	if s.shaderDoc != nil {
		s.shaderDoc.Destroy()
		s.shaderDoc = nil
	}
	if s.renderPassDoc != nil {
		s.renderPassDoc.Destroy()
		s.renderPassDoc = nil
	}
	if s.pipelineDoc != nil {
		s.pipelineDoc.Destroy()
		s.pipelineDoc = nil
	}
	if s.materialDoc != nil {
		s.materialDoc.Destroy()
		s.materialDoc = nil
	}
}

func (s *ShaderDesigner) Reload(uiMan *ui.Manager, root *document.Element) {
	s.man = uiMan
	s.root = root
	switch s.state {
	case StateHome:
		s.reloadShaderDesigner()
		s.shaderDesignerDoc.Activate()
	case StateShader:
		s.reloadShaderDoc()
		s.shaderDoc.Activate()
	case StateRenderPass:
		s.reloadRenderPassDoc()
		s.renderPassDoc.Activate()
	case StatePipeline:
		s.reloadPipelineDoc()
		s.pipelineDoc.Activate()
	case StateMaterial:
		s.reloadMaterialDoc()
		s.materialDoc.Activate()
	}
}

func (s flagState) Has(val string) bool {
	return slices.Contains(s.Current, val)
}

func New(state ShaderDesignerState, logStream *logging.LogStream) *ShaderDesigner {
	s := &ShaderDesigner{
		state: state,
	}
	tab_container.NewWindow(s.TabTitle(), -1, -1, []tab_container.TabContainerTab{
		tab_container.NewTab(s),
	}, logStream)
	return s
}

func (win *ShaderDesigner) ChangeWindowState(state ShaderDesignerState) {
	if win.state == state {
		return
	}
	win.state = state
	if win.shaderDoc != nil {
		win.shaderDoc.Deactivate()
	}
	if win.pipelineDoc != nil {
		win.pipelineDoc.Deactivate()
	}
	if win.renderPassDoc != nil {
		win.renderPassDoc.Deactivate()
	}
	if win.materialDoc != nil {
		win.materialDoc.Deactivate()
	}
	if win.shaderDesignerDoc != nil {
		win.shaderDesignerDoc.Deactivate()
	}
	switch state {
	case StateHome:
		win.reloadShaderDesigner()
		win.shaderDesignerDoc.Activate()
	case StateShader:
		win.reloadShaderDoc()
		win.shaderDoc.Activate()
	case StateRenderPass:
		win.reloadRenderPassDoc()
		win.renderPassDoc.Activate()
	case StatePipeline:
		win.reloadPipelineDoc()
		win.pipelineDoc.Activate()
	case StateMaterial:
		win.reloadMaterialDoc()
		win.materialDoc.Activate()
	}
	host := win.man.Host.Value()
	debug.EnsureNotNil(host)
	host.Window.Focus()
}

func (win *ShaderDesigner) ShowDesignerWindow() {
	win.ChangeWindowState(StateHome)
}

func (win *ShaderDesigner) ShowShaderWindow() {
	win.ChangeWindowState(StateShader)
}

func (win *ShaderDesigner) ShowRenderPassWindow() {
	win.ChangeWindowState(StateRenderPass)
}

func (win *ShaderDesigner) ShowPipelineWindow() {
	win.ChangeWindowState(StatePipeline)
}

func (win *ShaderDesigner) ShowMaterialWindow() {
	win.ChangeWindowState(StateMaterial)
}

func (win *ShaderDesigner) returnHome(*document.Element) {
	win.ShowDesignerWindow()
}

func (win *ShaderDesigner) reloadShaderDesigner() {
	if win.shaderDesignerDoc != nil {
		win.shaderDesignerDoc.Destroy()
	}
	win.shaderDesignerDoc, _ = markup.DocumentFromHTMLAssetRooted(win.man, shaderDesignerHTML,
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
		}, win.root)
}

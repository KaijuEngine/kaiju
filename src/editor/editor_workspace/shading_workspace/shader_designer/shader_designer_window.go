/******************************************************************************/
/* shader_designer_window.go                                                  */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
	"slices"

	"github.com/KaijuEngine/kaiju/editor/project/project_database/content_database"
	"github.com/KaijuEngine/kaiju/editor/project/project_file_system"
	"github.com/KaijuEngine/kaiju/engine"
	"github.com/KaijuEngine/kaiju/engine/ui"
	"github.com/KaijuEngine/kaiju/engine/ui/markup/document"
	"github.com/KaijuEngine/kaiju/rendering"
)

const (
	shaderDesignerHTML = "ui/shader_designer/shader_designer_window.html"
)

type ShaderDesignerState = int

const (
	StateNone = ShaderDesignerState(iota)
	StateShader
	StateRenderPass
	StatePipeline
	StateMaterial
)

type ShaderDesignerShader struct {
	rendering.ShaderData
	id string
}

type ShaderDesignerRenderPass struct {
	rendering.RenderPassData
	id string
}

type ShaderDesignerShaderPipeline struct {
	rendering.ShaderPipelineData
	id string
}

type ShaderDesignerMaterial struct {
	rendering.MaterialData
	id string
}

type ShaderDesigner struct {
	host              *engine.Host
	shader            ShaderDesignerShader
	renderPass        ShaderDesignerRenderPass
	pipeline          ShaderDesignerShaderPipeline
	material          ShaderDesignerMaterial
	shaderDesignerDoc *document.Document
	shaderDoc         *document.Document
	pipelineDoc       *document.Document
	renderPassDoc     *document.Document
	materialDoc       *document.Document
	uiMan             *ui.Manager
	pfs               *project_file_system.FileSystem
	cache             *content_database.Cache
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

func (s *ShaderDesigner) Reload() {
	switch s.state {
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

func (s *ShaderDesigner) Initialize(host *engine.Host, uiMan *ui.Manager, pfs *project_file_system.FileSystem, cache *content_database.Cache) {
	s.host = host
	s.uiMan = uiMan
	s.pfs = pfs
	s.cache = cache
	s.Reload()
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
}

func (win *ShaderDesigner) ShowShaderWindow(id string, data rendering.ShaderData) {
	win.resetDataValues()
	win.shader.id = id
	win.shader.ShaderData = data
	win.ChangeWindowState(StateShader)
}

func (win *ShaderDesigner) ShowRenderPassWindow(id string, data rendering.RenderPassData) {
	win.resetDataValues()
	win.renderPass.id = id
	win.renderPass.RenderPassData = data
	win.ChangeWindowState(StateRenderPass)
}

func (win *ShaderDesigner) ShowPipelineWindow(id string, data rendering.ShaderPipelineData) {
	win.resetDataValues()
	win.pipeline.id = id
	win.pipeline.ShaderPipelineData = data
	win.ChangeWindowState(StatePipeline)
}

func (win *ShaderDesigner) ShowMaterialWindow(id string, data rendering.MaterialData) {
	win.resetDataValues()
	win.material.id = id
	win.material.MaterialData = data
	win.ChangeWindowState(StateMaterial)
}

func (win *ShaderDesigner) Close() {
	win.ChangeWindowState(StateNone)
}

func (win *ShaderDesigner) resetDataValues() {
	win.shader = ShaderDesignerShader{}
	win.material = ShaderDesignerMaterial{}
	win.renderPass = ShaderDesignerRenderPass{}
	win.pipeline = ShaderDesignerShaderPipeline{}
}

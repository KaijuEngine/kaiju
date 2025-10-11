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
	"fmt"
	"kaiju/engine"
	"kaiju/engine/host_container"
	"kaiju/engine/systems/console"
	"kaiju/engine/systems/logging"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"os"
	"path/filepath"
	"slices"
)

const (
	shaderDesignerHTML = "ui/shader_designer/shader_designer_window.html"
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

func (s *ShaderDesigner) Reload(root *document.Element) {
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
	return newInternal(state, logStream, nil)
}

func newInternal(state ShaderDesignerState, logStream *logging.LogStream, beforeLoad func(*ShaderDesigner)) *ShaderDesigner {
	win := &ShaderDesigner{}
	container := host_container.New("Shader Designer", logStream)
	go container.Run(640, 480, -1, -1)
	<-container.PrepLock
	container.RunFunction(func() {
		win.man.Init(container.Host)
		if beforeLoad != nil {
			beforeLoad(win)
		}
		win.Reload(nil)
		win.state = state
	})
	return win
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
	win.man.Host.Window.Focus()
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

func SetupConsole(host *engine.Host) {
	defer tracing.NewRegion("shader_designer.SetupConsole").End()
	console.For(host).AddCommand("shaderdesigner", "Opens the shader designer window", func(_ *engine.Host, loadFile string) string {
		if loadFile == "" {
			New(StateHome, host.LogStream)
			return "Launching a shader designer"
		} else {
			ext := filepath.Ext(loadFile)
			switch ext {
			case ".material":
				loadFile = filepath.Join("renderer/materials", loadFile)
			case ".shaderpipeline":
				loadFile = filepath.Join("renderer/pipelines", loadFile)
			case ".renderpass":
				loadFile = filepath.Join("renderer/passes", loadFile)
			case ".shader":
				loadFile = filepath.Join("renderer/shaders", loadFile)
			}
			loadFile = host.AssetDatabase().ToFilePath(loadFile)
			if _, err := os.Stat(loadFile); os.IsNotExist(err) {
				return fmt.Sprintf("File not found: %s", loadFile)
			}
			switch ext {
			case ".material":
				OpenMaterial(loadFile, host.LogStream)
			case ".shaderpipeline":
				OpenPipeline(loadFile, host.LogStream)
			case ".renderpass":
				OpenRenderPass(loadFile, host.LogStream)
			case ".shader":
				OpenShader(loadFile, host.LogStream)
			}
			return fmt.Sprintf("Shader designer opening file: %s", loadFile)
		}
	})
}

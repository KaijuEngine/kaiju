/******************************************************************************/
/* shader.go                                                                  */
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

package rendering

import (
	"kaiju/assets"
	"strings"
)

type Shader struct {
	RenderId   ShaderId
	data       ShaderDataCompiled
	Material   MaterialData
	DriverData ShaderDriverData
	subShaders map[string]*Shader
}

type ShaderData struct {
	Name                        string
	Vertex                      string              `options:""`
	VertexFlags                 string              `tip:"CompileFlags"`
	Fragment                    string              `options:""`
	FragmentFlags               string              `tip:"CompileFlags"`
	Geometry                    string              `options:""`
	GeometryFlags               string              `tip:"CompileFlags"`
	TessellationControl         string              `options:""`
	TessellationControlFlags    string              `tip:"CompileFlags"`
	TessellationEvaluation      string              `options:""`
	TessellationEvaluationFlags string              `tip:"CompileFlags"`
	LayoutGroups                []ShaderLayoutGroup `ignore:"true"`
}

type ShaderDataCompiled struct {
	Name                   string
	Vertex                 string
	Fragment               string
	Geometry               string
	TessellationControl    string
	TessellationEvaluation string
	LayoutGroups           []ShaderLayoutGroup
}

func (d *ShaderData) Compile() ShaderDataCompiled {
	return ShaderDataCompiled{
		Name:                   d.Name,
		Vertex:                 strings.Replace(d.Vertex, "/src/", "/spv/", 1) + ".spv",
		Fragment:               strings.Replace(d.Fragment, "/src/", "/spv/", 1) + ".spv",
		Geometry:               strings.Replace(d.Geometry, "/src/", "/spv/", 1) + ".spv",
		TessellationControl:    strings.Replace(d.TessellationControl, "/src/", "/spv/", 1) + ".spv",
		TessellationEvaluation: strings.Replace(d.TessellationEvaluation, "/src/", "/spv/", 1) + ".spv",
		LayoutGroups:           d.LayoutGroups,
	}
}

func (s *Shader) AddSubShader(key string, shader *Shader) {
	s.subShaders[key] = shader
}

func (s *Shader) RemoveSubShader(key string) {
	delete(s.subShaders, key)
}

func (s *Shader) SubShader(key string) *Shader {
	return s.subShaders[key]
}

func NewShader(shaderData ShaderDataCompiled) *Shader {
	s := &Shader{
		data:       shaderData,
		subShaders: make(map[string]*Shader),
		DriverData: NewShaderDriverData(),
	}
	return s
}

func (s *Shader) DelayedCreate(renderer Renderer, assetDatabase *assets.Database) {
	renderer.CreateShader(s, assetDatabase)
	for _, ss := range s.subShaders {
		renderer.CreateShader(ss, assetDatabase)
	}
}

func (s *Shader) IsComposite() bool {
	return s.data.Vertex == assets.ShaderOitCompositeVert
}

func (s *Shader) Destroy(renderer Renderer) {
	renderer.DestroyShader(s)
}

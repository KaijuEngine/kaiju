/******************************************************************************/
/* shader.go                                                                  */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"kaiju/engine/assets"
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"weak"
)

type ShaderType int

const (
	ShaderTypeGraphics ShaderType = iota
	ShaderTypeCompute
)

type Shader struct {
	RenderId     ShaderId
	data         ShaderDataCompiled
	Material     MaterialData
	DriverData   ShaderDriverData
	subShaders   map[string]*Shader
	pipelineInfo *ShaderPipelineDataCompiled
	renderPass   weak.Pointer[RenderPass]
	Type         ShaderType
}

type ShaderData struct {
	Name                        string
	EnableDebug                 bool
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
	Compute                     string              `options:""`
	ComputeFlags                string              `tip:"CompileFlags"`
	LayoutGroups                []ShaderLayoutGroup `visible:"false"`
	VertexSpv                   string
	FragmentSpv                 string
	GeometrySpv                 string
	TessellationControlSpv      string
	TessellationEvaluationSpv   string
	ComputeSpv                  string
}

type ShaderDataCompiled struct {
	Name                   string
	Vertex                 string
	Fragment               string
	Geometry               string
	TessellationControl    string
	TessellationEvaluation string
	Compute                string
	LayoutGroups           []ShaderLayoutGroup
}

func (s *ShaderDataCompiled) IsCompute() bool { return s.Compute != "" }

func (s *ShaderDataCompiled) SelectLayout(stage string) *ShaderLayoutGroup {
	for i := range s.LayoutGroups {
		if s.LayoutGroups[i].Type == stage {
			return &s.LayoutGroups[i]
		}
	}
	return nil
}

func (sd *ShaderDataCompiled) WorkGroups() [3]uint32 {
	if !sd.IsCompute() {
		return [3]uint32{0, 0, 0}
	}
	g := sd.SelectLayout("Compute")
	return g.WorkGroups
}

func (sd *ShaderDataCompiled) Stride() uint32 {
	stride := uint32(0)
	if !sd.IsCompute() {
		g := sd.SelectLayout("Vertex")
		for i := range g.Layouts {
			l := &g.Layouts[i]
			if l.Source == "in" && l.Location >= 8 {
				stride += uint32(fieldSize(l.Type, l.FullName()))
			}
		}
	}
	return stride
}

func (sd *ShaderDataCompiled) ToAttributeDescription(locationStart uint32) []vk.VertexInputAttributeDescription {
	defer tracing.NewRegion("Shader.ToAttributeDescription").End()
	attrs := make([]vk.VertexInputAttributeDescription, 0)
	offset := uint32(0)
	if !sd.IsCompute() {
		g := sd.SelectLayout("Vertex")
		for i := range g.Layouts {
			l := &g.Layouts[i]
			if l.Source == "in" && uint32(l.Location) >= locationStart {
				dt := defTypes[l.Type]
				dt.repeat *= max(1, l.Count)
				for r := range dt.repeat {
					attrs = append(attrs, vk.VertexInputAttributeDescription{
						Location: uint32(l.Location + r),
						Binding:  1,
						Format:   dt.format,
						Offset:   offset,
					})
					offset += dt.size
				}
			}
		}
	}
	return attrs
}

func (sd *ShaderDataCompiled) ToDescriptorSetLayoutStructure() DescriptorSetLayoutStructure {
	defer tracing.NewRegion("Shader.ToDescriptorSetLayoutStructure").End()
	structure := DescriptorSetLayoutStructure{}
	for _, g := range sd.LayoutGroups {
		for _, layout := range g.Layouts {
			if layout.Binding < 0 {
				continue
			}
			skip := false
			for i := range structure.Types {
				// TODO:  This can happen across 2+ different files (vert/frag)
				// but an error should be shown if things don't match up or
				// if it is the same file
				if structure.Types[i].Binding == uint32(layout.Binding) {
					structure.Types[i].Flags |= g.DescriptorFlag()
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			structure.Types = append(structure.Types, DescriptorSetLayoutStructureType{
				Type:    layout.DescriptorType(),
				Flags:   g.DescriptorFlag(),
				Count:   max(1, uint32(layout.Count)),
				Binding: uint32(layout.Binding),
			})
		}
	}
	return structure
}

func (d *ShaderData) Compile() ShaderDataCompiled {
	defer tracing.NewRegion("Shader.Compile").End()
	return ShaderDataCompiled{
		Name:                   d.Name,
		Vertex:                 d.VertexSpv,
		Fragment:               d.FragmentSpv,
		Geometry:               d.GeometrySpv,
		Compute:                d.ComputeSpv,
		TessellationControl:    d.TessellationControlSpv,
		TessellationEvaluation: d.TessellationEvaluationSpv,
		LayoutGroups:           d.LayoutGroups,
	}
}

func (s *Shader) ShaderDataName() string { return s.data.Name }

func (s *Shader) AddSubShader(key string, shader *Shader) {
	shader.pipelineInfo = s.pipelineInfo
	shader.renderPass = s.renderPass
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
		Type:       ShaderTypeGraphics,
	}
	if shaderData.Compute != "" {
		s.Type = ShaderTypeCompute
	}
	return s
}

func (s *Shader) Reload(shaderData ShaderDataCompiled) {
	s.RenderId = ShaderId{}
	s.data = shaderData
	s.DriverData = NewShaderDriverData()
}

func (s *Shader) DelayedCreate(renderer Renderer, assetDatabase assets.Database) {
	renderer.CreateShader(s, assetDatabase)
	for _, ss := range s.subShaders {
		renderer.CreateShader(ss, assetDatabase)
	}
}

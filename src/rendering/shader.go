/******************************************************************************/
/* shader.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"weak"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/profiler/tracing"
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
	SamplerLabels               []string
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
	SamplerLabels          []string
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
		if g == nil {
			return 0
		}
		for i := range g.Layouts {
			l := &g.Layouts[i]
			if l.Source == "in" && l.Location >= 8 {
				stride += uint32(fieldSize(l.Type, l.FullName()))
			}
		}
	}
	return stride
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
		SamplerLabels:          d.SamplerLabels,
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

func (s *Shader) DelayedCreate(device *GPUDevice, assetDatabase assets.Database) {
	device.CreateShader(s, assetDatabase)
	for _, ss := range s.subShaders {
		device.CreateShader(ss, assetDatabase)
	}
}

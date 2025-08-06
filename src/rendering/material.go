/******************************************************************************/
/* material.go                                                                */
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
	"encoding/json"
	"kaiju/engine/assets"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"weak"
)

type Material struct {
	Name          string
	shaderInfo    ShaderDataCompiled
	renderPass    *RenderPass
	pipelineInfo  ShaderPipelineDataCompiled
	Shader        *Shader
	Textures      []*Texture
	ShadowMap     *Texture
	ShadowCubeMap *Texture
	Instances     map[string]*Material
	Root          weak.Pointer[Material]
	mutex         sync.Mutex
	IsLit         bool
}

func (m *Material) HasShadowMap() bool {
	return m.ShadowMap != nil && m.ShadowMap.RenderId.IsValid()
}

func (m *Material) HasShadowCubeMap() bool {
	return m.ShadowCubeMap != nil && m.ShadowCubeMap.RenderId.IsValid()
}

func (m *Material) HasTransparentSuffix() bool {
	return strings.HasSuffix(m.Name, "_transparent")
}

func (m *Material) SelectRoot() *Material {
	if m.Root.Value() != nil {
		return m.Root.Value()
	}
	return m
}

type MaterialTextureData struct {
	Texture string `options:""` // Blank = fallback
	Filter  string `options:"StringVkFilter"`
}

type MaterialData struct {
	Name           string
	Shader         string `options:""` // Blank = fallback
	RenderPass     string `options:""` // Blank = fallback
	ShaderPipeline string `options:""` // Blank = fallback
	Textures       []MaterialTextureData
}

func (m *Material) CreateInstance(textures []*Texture) *Material {
	defer tracing.NewRegion("Material.CreateInstance").End()
	instanceKey := strings.Builder{}
	for i := range textures {
		instanceKey.WriteString(textures[i].Key)
		instanceKey.WriteRune(';')
	}
	key := instanceKey.String()
	// TODO:  Use a read lock?
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if found, ok := m.Instances[key]; ok {
		return found
	}
	copy := &Material{}
	*copy = *m
	copy.Textures = slices.Clone(textures)
	copy.ShadowMap = m.ShadowMap
	copy.ShadowCubeMap = m.ShadowCubeMap
	// TODO:  If using a read lock, then make sure to write lock the following line
	m.Instances[key] = copy
	copy.Root = weak.Make(m)
	copy.Instances = nil
	return copy
}

func (d *MaterialTextureData) FilterToVK() TextureFilter {
	switch d.Filter {
	case "Nearest":
		return TextureFilterNearest
	case "Linear":
		return TextureFilterLinear
	case "CubicImg":
		// TODO:  Implement this filter
		fallthrough
	default:
		return TextureFilterLinear
	}
}

func (d *MaterialData) Compile(assets *assets.Database, renderer Renderer) (*Material, error) {
	defer tracing.NewRegion("MaterialData.Compile").End()
	vr := renderer.(*Vulkan)
	c := &Material{
		Name:      d.Name,
		Textures:  make([]*Texture, len(d.Textures)),
		Instances: make(map[string]*Material),
	}
	sd := ShaderData{}
	rp := RenderPassData{}
	sp := ShaderPipelineData{}
	if err := unmarshallJsonFile(assets, d.Shader, &sd); err != nil {
		return c, err
	}
	if err := unmarshallJsonFile(assets, d.RenderPass, &rp); err != nil {
		return c, err
	}
	if err := unmarshallJsonFile(assets, d.ShaderPipeline, &sp); err != nil {
		return c, err
	}
	c.shaderInfo = sd.Compile()
	if pass, ok := vr.renderPassCache[rp.Name]; !ok {
		rpc := rp.Compile(vr)
		if p, ok := rpc.ConstructRenderPass(vr); ok {
			vr.renderPassCache[rp.Name] = p
			c.renderPass = p
		} else {
			slog.Error("failed to load the render pass for the material", "material", d.Name, "renderPass", rp.Name)
		}
	} else {
		c.renderPass = pass
	}
	c.pipelineInfo = sp.Compile(vr)
	shaderConfig, err := assets.ReadText(d.Shader)
	if err != nil {
		return c, err
	}
	var rawSD ShaderData
	if err := json.Unmarshal([]byte(shaderConfig), &rawSD); err != nil {
		return c, err
	}
	c.Shader, _ = vr.caches.ShaderCache().Shader(rawSD.Compile())
	c.Shader.pipelineInfo = &c.pipelineInfo
	c.Shader.renderPass = weak.Make(c.renderPass)
	for i := range d.Textures {
		tex, err := vr.caches.TextureCache().Texture(
			d.Textures[i].Texture, d.Textures[i].FilterToVK())
		if err != nil {
			return c, err
		}
		c.Textures[i] = tex
	}
	return c, nil
}

func (m *Material) Destroy(renderer Renderer) {
	defer tracing.NewRegion("Material.Destroy").End()
	vr := renderer.(*Vulkan)
	m.renderPass.Destroy(vr)
	m.renderPass = nil
	m.Shader = nil
	m.Textures = make([]*Texture, 0)
	m.ShadowMap = nil
	m.ShadowCubeMap = nil
	clear(m.Instances)
}

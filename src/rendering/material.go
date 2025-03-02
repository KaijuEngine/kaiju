package rendering

import (
	"encoding/json"
	"kaiju/assets"
)

type Material struct {
	Name           string
	shaderInfo     ShaderDataCompiled
	renderPassInfo RenderPassDataCompiled
	pipelineInfo   ShaderPipelineDataCompiled
	Shader         *Shader
	Textures       []*Texture
}

type MaterialTextureData struct {
	Texture string
	Filter  string `options:"StringVkFilter"`
}

type MaterialData struct {
	Name           string
	Shader         string `options:""` // Blank options uses fallback
	RenderPass     string `options:""` // Blank options uses fallback
	ShaderPipeline string `options:""` // Blank options uses fallback
	Textures       []MaterialTextureData
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

func materialUnmarshallData(assets *assets.Database, file string, to any) error {
	s, err := assets.ReadText(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(s), to); err != nil {
		return err
	}
	return nil
}

func (d *MaterialData) Compile(assets *assets.Database, renderer Renderer) (*Material, error) {
	vr := renderer.(*Vulkan)
	c := &Material{
		Name:     d.Name,
		Textures: make([]*Texture, len(d.Textures)),
	}
	sd := ShaderData{}
	if err := materialUnmarshallData(assets, d.Shader, &sd); err != nil {
		return c, err
	}
	c.shaderInfo = sd.Compile()
	rp := RenderPassData{}
	if err := materialUnmarshallData(assets, d.Shader, &rp); err != nil {
		return c, err
	}
	c.renderPassInfo = rp.Compile(vr)
	sp := ShaderPipelineData{}
	if err := materialUnmarshallData(assets, d.Shader, &sp); err != nil {
		return c, err
	}
	shaderConfig, err := assets.ReadText(d.Shader)
	if err != nil {
		return c, err
	}
	var rawSD ShaderData
	if err := json.Unmarshal([]byte(shaderConfig), &rawSD); err != nil {
		return c, err
	}
	c.Shader, _ = vr.caches.ShaderCache().Shader(rawSD.Compile())
	c.pipelineInfo = sp.Compile()
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

package rendering

import (
	"encoding/json"
	"kaiju/assets"
)

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

type Material struct {
	Key            string
	ShaderInfo     ShaderDataCompiled
	RenderPassInfo RenderPassDataCompiled
	PipelineInfo   ShaderPipelineDataCompiled
	Textures       []*Texture
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
		Key:      d.Name,
		Textures: make([]*Texture, len(d.Textures)),
	}
	sd := ShaderData{}
	if err := materialUnmarshallData(assets, d.Shader, &sd); err != nil {
		return c, err
	}
	c.ShaderInfo = sd.Compile()
	rp := RenderPassData{}
	if err := materialUnmarshallData(assets, d.Shader, &rp); err != nil {
		return c, err
	}
	c.RenderPassInfo = rp.Compile(vr)
	sp := ShaderPipelineData{}
	if err := materialUnmarshallData(assets, d.Shader, &sp); err != nil {
		return c, err
	}
	c.PipelineInfo = sp.Compile()
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

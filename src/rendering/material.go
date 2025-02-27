package rendering

type MaterialTextureData struct {
	Texture string
}

type MaterialData struct {
	ShaderData     string `options:""` // Blank options uses fallback
	RenderPass     string `options:""` // Blank options uses fallback
	ShaderPipeline string `options:""` // Blank options uses fallback
	Textures       []MaterialTextureData
}

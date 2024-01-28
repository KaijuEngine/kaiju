//go:build !js && !OPENGL

package assets

// Shaders
const (
	ShaderTextVert         = "shaders/spv/text.vert.spv"
	ShaderTextFrag         = "shaders/spv/text.frag.spv"
	ShaderText3DVert       = "shaders/spv/text3d.vert.spv"
	ShaderText3DFrag       = ShaderTextFrag
	ShaderOitCompositeVert = "shaders/spv/oit_composite.vert.spv"
	ShaderOitCompositeFrag = "shaders/spv/oit_composite.frag.spv"
	ShaderHdrVert          = "shaders/spv/hdr.vert.spv"
	ShaderHdrFrag          = "shaders/spv/hdr.frag.spv"
	ShaderUIVert           = "shaders/spv/ui.vert.spv"
	ShaderUIFrag           = "shaders/spv/ui.frag.spv"
	ShadersUINineFrag      = "shaders/spv/ui_nine.frag.spv"
)

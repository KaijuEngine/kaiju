package rendering

type RenderCaches interface {
	ShaderCache() *ShaderCache
	TextureCache() *TextureCache
	MeshCache() *MeshCache
	FontCache() *FontCache
}

package texture

import "kaijuengine.com/matrix"

type TextureKind uint8
type TextureUsage uint32

const (
	Texture1D TextureKind = iota
	Texture1DArray
	Texture2D
	Texture2DArray
	Texture3D
	TextureCube
	TextureCubeArray
)

const (
	TextureSampled TextureUsage = (1 << iota)
	TextureStorage
	TextureColorTarget
	TextureDepthStencilTarget
	TextureCopySource
	TextureCopyDestination
)

type Texture struct {

}

type TextureDefinition struct {
	Kind       TextureKind
	Extent     matrix.Vec3
	Format     TextureFormat
	MipLevels  uint32
	Layers     uint32
	Samples    SampleCount
	Usage      TextureUsage
}

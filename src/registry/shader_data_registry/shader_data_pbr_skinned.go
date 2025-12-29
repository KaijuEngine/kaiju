package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register(func() rendering.DrawInstance {
		return &ShaderDataPbrSkinned{
			ShaderDataBase: rendering.NewShaderDataBase(),
			VertColors:     matrix.ColorWhite(),
			LightIds:       [...]int32{-1, -1, -1, -1},
		}
	}, "pbr_skinned")
}

type ShaderDataPbrSkinned struct {
	rendering.SkinnedShaderDataHeader
	rendering.ShaderDataBase
	VertColors matrix.Color
	Metallic   float32
	Roughness  float32
	Emissive   float32
	LightIds   [4]int32                `visible:"false"`
	SkinIndex  int32                   `visible:"false"`
	Flags      ShaderDataStandardFlags `visible:"false"`
}

func (t *ShaderDataPbrSkinned) SkinningHeader() *rendering.SkinnedShaderDataHeader {
	return &t.SkinnedShaderDataHeader
}

func (t ShaderDataPbrSkinned) Size() int {
	const size = int(unsafe.Sizeof(ShaderDataPbrSkinned{}) - rendering.ShaderBaseDataStart)
	return size
}

func (t *ShaderDataPbrSkinned) NamedDataInstanceSize(name string) int {
	return t.SkinNamedDataInstanceSize(name)
}

func (t *ShaderDataPbrSkinned) NamedDataPointer(name string) unsafe.Pointer {
	return t.SkinNamedDataPointer(name)
}

func (t *ShaderDataPbrSkinned) UpdateNamedData(index, capacity int, name string) bool {
	if s, ok := t.SkinUpdateNamedData(index, capacity, name); ok {
		t.SkinIndex = s
		return true
	}
	return false
}

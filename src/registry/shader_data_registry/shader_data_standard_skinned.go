package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register(func() rendering.DrawInstance {
		return &ShaderDataStandardSkinned{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			SkinIndex:      0,
		}
	}, fallback+"_skinned")
}

type ShaderDataStandardSkinned struct {
	rendering.SkinnedShaderDataHeader
	rendering.ShaderDataBase
	Color     matrix.Color
	SkinIndex int32                   `visible:"false"`
	Flags     ShaderDataStandardFlags `visible:"false"`
}

func (t *ShaderDataStandardSkinned) SkinningHeader() *rendering.SkinnedShaderDataHeader {
	return &t.SkinnedShaderDataHeader
}

func (t ShaderDataStandardSkinned) Size() int {
	const top = unsafe.Offsetof(ShaderDataStandardSkinned{}.ShaderDataBase) + rendering.ShaderBaseDataStart
	const size = int(unsafe.Sizeof(ShaderDataStandardSkinned{}) - top)
	return size
}

func (t *ShaderDataStandardSkinned) NamedDataInstanceSize(name string) int {
	return t.SkinNamedDataInstanceSize(name)
}

func (t *ShaderDataStandardSkinned) NamedDataPointer(name string) unsafe.Pointer {
	return t.SkinNamedDataPointer(name)
}

func (t *ShaderDataStandardSkinned) UpdateNamedData(index, capacity int, name string) bool {
	if s, ok := t.SkinUpdateNamedData(index, capacity, name); ok {
		t.SkinIndex = s
		return true
	}
	return false
}

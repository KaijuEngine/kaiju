package rendering

import (
	"kaiju/matrix"
	"unsafe"
)

type ShaderDataStandardFlags = uint32

const (
	ShaderDataStandardFlagOutline = ShaderDataStandardFlags(1 << iota)
)

type ShaderDataStandard struct {
	ShaderDataBase
	Color matrix.Color
	Flags ShaderDataStandardFlags
}

func (ShaderDataStandard) Size() int {
	return int(unsafe.Sizeof(ShaderDataStandard{}) - ShaderBaseDataStart)
}

func (s *ShaderDataStandard) TestFlag(flag ShaderDataStandardFlags) bool {
	return (s.Flags & flag) != 0
}

func (s *ShaderDataStandard) SetFlag(flag ShaderDataStandardFlags) {
	s.Flags |= flag
}

func (s *ShaderDataStandard) ClearFlag(flag ShaderDataStandardFlags) {
	s.Flags &^= flag
}

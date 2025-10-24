package rendering

import (
	"kaiju/matrix"
	"unsafe"
)

type ShaderDataStandardFlags = uint32

const (
	ShaderDataStandardFlagOutline = ShaderDataStandardFlags(1 << iota)
	// Enable bit will be set anytime there are flags. This is needed because
	// bits at the extremes of the float will be truncated to 0 otherwise. By
	// setting this bit (largest exponent bit 2^1) this issue can be prevented.
	ShaderDataStandardFlagEnable = 1 << 30
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
	s.updateFlagEnableStatus()
}

func (s *ShaderDataStandard) ClearFlag(flag ShaderDataStandardFlags) {
	s.Flags &^= flag
	s.updateFlagEnableStatus()
}

func (s *ShaderDataStandard) updateFlagEnableStatus() {
	if s.Flags|ShaderDataStandardFlagEnable == ShaderDataStandardFlagEnable {
		s.Flags = 0
	} else {
		s.Flags |= ShaderDataStandardFlagEnable
	}
}

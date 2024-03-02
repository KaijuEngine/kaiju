package vulkan

import (
	"bytes"
	"fmt"
	"unsafe"
)

// #include "vulkan/vulkan.h"
import "C"

// Max bounds of uint32 and uint64,
// declared as var so type would get checked.
var (
	MaxUint32 uint32 = 1<<32 - 1 // also ^uint32(0)
	MaxUint64 uint64 = 1<<64 - 1 // also ^uint64(0)
)

func (b Bool32) B() bool {
	return b == True
}

type Version uint32

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
}

func (v Version) Major() int {
	return int(uint32(v) >> 22)
}

func (v Version) Minor() int {
	return int(uint32(v) >> 12 & 0x3FF)
}

func (v Version) Patch() int {
	return int(uint32(v) & 0xFFF)
}

func MakeVersion(major, minor, patch int) uint32 {
	return uint32(major)<<22 | uint32(minor)<<12 | uint32(patch)
}

func ToString(buf []byte) string {
	var str bytes.Buffer
	for i := range buf {
		if buf[i] == '\x00' {
			return str.String()
		}
		str.WriteByte(buf[i])
	}
	return str.String()
}

// Memcopy is like a Go's built-in copy function, it copies data from src slice,
// but into a destination pointer. Useful to copy data into device memory.
func Memcopy(dst unsafe.Pointer, src []byte) int {
	const m = 0x7fffffff
	dstView := (*[m]byte)(dst)
	return copy(dstView[:len(src)], src)
}

func NewClearValue(color []float32) ClearValue {
	var v ClearValue
	v.SetColor(color)
	return v
}

func NewClearDepthStencil(depth float32, stencil uint32) ClearValue {
	var v ClearValue
	v.SetDepthStencil(depth, stencil)
	return v
}

func (cv *ClearValue) SetColor(color []float32) {
	vkClearValue := (*[4]float32)(unsafe.Pointer(cv))
	for i := 0; i < len(color); i++ {
		vkClearValue[i] = color[i]
	}
}

func (cv *ClearValue) SetDepthStencil(depth float32, stencil uint32) {
	depths := (*[2]float32)(unsafe.Pointer(cv))
	stencils := (*[2]uint32)(unsafe.Pointer(cv))
	depths[0] = depth
	stencils[1] = stencil
}

// SurfaceFromPointer casts a pointer to a Vulkan surface into a Surface.
func SurfaceFromPointer(surface uintptr) Surface {
	return *(*Surface)(unsafe.Pointer(surface))
}

/******************************************************************************/
/* descriptor_write_cache.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "unsafe"

const descriptorSignatureOffset uint64 = 14695981039346656037
const descriptorSignaturePrime uint64 = 1099511628211

type DescriptorWriteSignature struct {
	hash uint64
}

func NewDescriptorWriteSignature() DescriptorWriteSignature {
	return DescriptorWriteSignature{hash: descriptorSignatureOffset}
}

func (s *DescriptorWriteSignature) AddUint64(value uint64) {
	s.hash ^= value
	s.hash *= descriptorSignaturePrime
}

func (s *DescriptorWriteSignature) AddInt(value int) {
	s.AddUint64(uint64(value))
}

func (s *DescriptorWriteSignature) AddUintptr(value uintptr) {
	s.AddUint64(uint64(value))
}

func (s *DescriptorWriteSignature) AddString(value string) {
	for i := range value {
		s.AddUint64(uint64(value[i]))
	}
	s.AddUint64(0xff)
}

func (s *DescriptorWriteSignature) AddPointer(value unsafe.Pointer) {
	s.AddUintptr(uintptr(value))
}

func (s *DescriptorWriteSignature) AddHandle(handle GPUHandle) {
	s.AddPointer(handle.handle)
}

func (s DescriptorWriteSignature) Equal(other DescriptorWriteSignature) bool {
	return s.hash == other.hash
}

type DescriptorWriteCache struct {
	signatures [maxFramesInFlight]DescriptorWriteSignature
	valid      [maxFramesInFlight]bool
}

func (c *DescriptorWriteCache) ShouldWrite(frame int, signature DescriptorWriteSignature) bool {
	if frame >= 0 && frame < len(c.signatures) && c.valid[frame] && c.signatures[frame].Equal(signature) {
		return false
	}
	if frame >= 0 && frame < len(c.signatures) {
		c.signatures[frame] = signature
		c.valid[frame] = true
	}
	return true
}

func (c *DescriptorWriteCache) Invalidate() {
	clear(c.valid[:])
}

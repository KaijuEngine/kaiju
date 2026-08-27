/******************************************************************************/
/* gpu_memory.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gpu_types

type MemoryPropertyFlags uint8
type MemoryHeapFlags uint8
type MemoryFlags uint16
type BufferUsageFlags uint16

type MemoryRequirements struct {
	Size           uintptr
	Alignment      uintptr
	MemoryTypeBits uint32
}

type MemoryType struct {
	PropertyFlags MemoryPropertyFlags
	HeapIndex     uint32
}

type MemoryHeap struct {
	Size  uintptr
	Flags MemoryHeapFlags
}

const (
	MemoryPropertyDeviceLocalBit MemoryPropertyFlags = (1 << iota)
	MemoryPropertyHostVisibleBit
	MemoryPropertyHostCoherentBit
	MemoryPropertyHostCachedBit
	MemoryPropertyLazilyAllocatedBit
	MemoryPropertyProtectedBit
)

const (
	MemoryHeapDeviceLocalBit MemoryHeapFlags = (1 << iota)
	MemoryHeapMultiInstanceBit
)

const (
	MemoryMapPlacedBit MemoryFlags = (1 << iota)
)

const (
	BufferUsageTransferSrcBit BufferUsageFlags = (1 << iota)
	BufferUsageTransferDstBit
	BufferUsageUniformTexelBufferBit
	BufferUsageStorageTexelBufferBit
	BufferUsageUniformBufferBit
	BufferUsageStorageBufferBit
	BufferUsageIndexBufferBit
	BufferUsageVertexBufferBit
	BufferUsageIndirectBufferBit
	BufferUsageTransformFeedbackBufferBit
	BufferUsageTransformFeedbackCounterBufferBit
	BufferUsageConditionalRenderingBit
	BufferUsageRaytracingBitNvx
)

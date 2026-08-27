/******************************************************************************/
/* gpu_handle.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gpu_types

import "unsafe"

type GpuHandle struct{ Handle unsafe.Pointer }

func (g *GpuHandle) Reset()                     { g.Handle = nil }
func (g *GpuHandle) IsValid() bool              { return g.Handle != nil }
func (g *GpuHandle) HandleAddr() unsafe.Pointer { return unsafe.Pointer(&g.Handle) }

type Fence struct{ GpuHandle }
type Queue struct{ GpuHandle }
type Semaphore struct{ GpuHandle }
type DescriptorPool struct{ GpuHandle }
type DescriptorSet struct{ GpuHandle }
type DeviceMemory struct{ GpuHandle }
type Buffer struct{ GpuHandle }
type FrameBuffer struct{ GpuHandle }
type Pipeline struct{ GpuHandle }
type PipelineLayout struct{ GpuHandle }
type DescriptorSetLayout struct{ GpuHandle }
type ShaderModule struct{ GpuHandle }

/******************************************************************************/
/* gpu_device.go                                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"log/slog"
	"unsafe"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type GPUDevice struct {
	PhysicalDevice             GPUPhysicalDevice
	LogicalDevice              GPULogicalDevice
	Painter                    GPUPainter
	globalUniformBuffers       [maxFramesInFlight]GPUBuffer
	globalUniformBuffersMemory [maxFramesInFlight]GPUDeviceMemory
	singleTimeCommandPool      pooling.PoolGroup[CommandRecorder]
}

func (g *GPUDevice) QueueCompute(buffer *ComputeShaderBuffer) {
	if buffer.Shader.Type != ShaderTypeCompute {
		slog.Error("QueueCompute called with non-compute shader")
		return
	}
	g.Painter.computeTasks = append(g.Painter.computeTasks, ComputeTask{
		Shader:         buffer.Shader,
		DescriptorSets: buffer.sets[:],
		WorkGroups:     buffer.Shader.data.WorkGroups(),
	})
}

func (g *GPUDevice) CreateSwapChain(window RenderingContainer, inst *GPUApplicationInstance) error {
	defer tracing.NewRegion("GPUDevice.CreateSwapChain").End()
	if g.LogicalDevice.SwapChain.IsValid() {
		g.LogicalDevice.WaitForRender(g)
	}
	g.PhysicalDevice.RefreshSurfaceCapabilities(inst.Surface.handle)
	return g.LogicalDevice.SwapChain.Setup(window, inst, g)
}

func (g *GPUDevice) MapMemory(memory GPUDeviceMemory, offset uintptr, size uintptr, flags GPUMemoryFlags, out *unsafe.Pointer) error {
	defer tracing.NewRegion("GPUDevice.CreateImage").End()
	return g.mapMemoryImpl(memory, offset, size, flags, out)
}

func (g *GPUDevice) Memcopy(dst unsafe.Pointer, src []byte) int {
	defer tracing.NewRegion("GPUDevice.Memcopy").End()
	return g.memcopyImpl(dst, src)
}

func (g *GPUDevice) UnmapMemory(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.UnmapMemory").End()
	g.unmapMemoryImpl(memory)
}

func (g *GPUDevice) CreateBuffer(size uintptr, usage GPUBufferUsageFlags, properties GPUMemoryPropertyFlags) (GPUBuffer, GPUDeviceMemory, error) {
	defer tracing.NewRegion("GPUDevice.CreateBuffer").End()
	return g.createBufferImpl(size, usage, properties)
}

func (g *GPUDevice) DestroyBuffer(buffer GPUBuffer) {
	defer tracing.NewRegion("GPUDevice.DestroyBuffer").End()
	g.destroyBufferImpl(buffer)
}

func (g *GPUDevice) FreeMemory(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.FreeMemory").End()
	g.freeMemoryImpl(memory)
}

func (g *GPUDevice) CreateFrameBuffer(renderPass *RenderPass, attachments []GPUImageView, width, height int32) (GPUFrameBuffer, error) {
	defer tracing.NewRegion("GPULogicalDevice.CreateFrameBuffer").End()
	return g.createFrameBufferImpl(renderPass, attachments, width, height)
}

func (g *GPUDevice) DestroyFrameBuffer(frameBuffer GPUFrameBuffer) {
	defer tracing.NewRegion("GPULogicalDevice.DestroyFrameBuffer").End()
	g.destroyFrameBufferImpl(frameBuffer)
}

func (g *GPUDevice) CopyBuffer(srcBuffer GPUBuffer, dstBuffer GPUBuffer, size uintptr) {
	defer tracing.NewRegion("GPULogicalDevice.CreateFrameBuffer").End()
	g.copyBufferImpl(srcBuffer, dstBuffer, size)
}

func (g *GPUDevice) createGlobalUniforms() error {
	slog.Info("creating global uniform buffers")
	bufferSize := unsafe.Sizeof(*(*GlobalShaderData)(nil))
	for i := range g.LogicalDevice.SwapChain.Images {
		b, m, err := g.CreateBuffer(bufferSize, GPUBufferUsageUniformBufferBit,
			GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
		if err != nil {
			return err
		}
		g.globalUniformBuffers[i] = b
		g.globalUniformBuffersMemory[i] = m
	}
	return nil
}

func (g *GPUDevice) destroyGlobalUniforms() {
	dbg := g.LogicalDevice.dbg
	for i := 0; i < maxFramesInFlight; i++ {
		g.DestroyBuffer(g.globalUniformBuffers[i])
		dbg.remove(g.globalUniformBuffers[i].handle)
		g.FreeMemory(g.globalUniformBuffersMemory[i])
		dbg.remove(g.globalUniformBuffersMemory[i].handle)
	}
}

func (g *GPUDevice) beginSingleTimeCommands() *CommandRecorder {
	defer tracing.NewRegion("GPUDevice.beginSingleTimeCommands").End()
	return g.beginSingleTimeCommandsImpl()
}

func (g *GPUDevice) endSingleTimeCommands(cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.endSingleTimeCommands").End()
	g.endSingleTimeCommandsImpl(cmd)
}

func (g *GPUDevice) SwapFrame(window RenderingContainer, inst *GPUApplicationInstance, width, height int32) bool {
	defer tracing.NewRegion("Vulkan.SwapFrame").End()
	if !g.LogicalDevice.SwapChain.IsValid() || len(g.Painter.writtenCommands) == 0 {
		return false
	}
	return g.swapFrameImpl(window, inst, width, height)
}

func (g *GPUDevice) ReadyFrame(inst *GPUApplicationInstance, window RenderingContainer, camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32) bool {
	defer tracing.NewRegion("Vulkan.ReadyFrame").End()
	ld := &g.LogicalDevice
	if !ld.SwapChain.IsValid() {
		if err := ld.RemakeSwapChain(window, inst, g); err != nil {
			return false
		}
		if !ld.SwapChain.IsValid() {
			return false
		}
	}
	return g.readyFrameImpl(inst, window, camera, uiCamera, lights, runtime)
}

func (g *GPUDevice) updateGlobalUniformBuffer(camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32) error {
	defer tracing.NewRegion("Vulkan.updateGlobalUniformBuffer").End()
	camOrtho := matrix.Float(0)
	if camera.IsOrthographic() {
		camOrtho = 1
	}
	ld := &g.LogicalDevice
	ubo := GlobalShaderData{
		View:             camera.View(),
		UIView:           uiCamera.View(),
		Projection:       camera.Projection(),
		UIProjection:     uiCamera.Projection(),
		CameraPosition:   camera.Position().AsVec4WithW(camOrtho),
		UICameraPosition: uiCamera.Position(),
		Time:             runtime,
		ScreenSize: matrix.Vec2{
			matrix.Float(ld.SwapChain.Extent.Width()),
			matrix.Float(ld.SwapChain.Extent.Height()),
		},
		CascadeCount:          int32(camera.NumCSMCascades()),
		CascadePlaneDistances: camera.CSMCascadeDistances(),
	}
	for i := range lights.Lights {
		if lights.Lights[i].IsValid() {
			lights.Lights[i].recalculate(camera)
			ubo.VertLights[i] = lights.Lights[i].transformToGPULight()
			ubo.LightInfos[i] = lights.Lights[i].transformToGPULightInfo()
		}
	}
	frame := g.Painter.currentFrame
	var data unsafe.Pointer
	err := g.MapMemory(g.globalUniformBuffersMemory[frame],
		0, unsafe.Sizeof(ubo), 0, &data)
	if err != nil {
		slog.Error("Failed to map uniform buffer memory", "error", err)
		return err
	}
	g.Memcopy(data, klib.StructToByteArray(ubo))
	g.UnmapMemory(g.globalUniformBuffersMemory[frame])
	return nil
}

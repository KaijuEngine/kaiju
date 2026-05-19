/******************************************************************************/
/* occlusion_gpu.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/

package rendering

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"unsafe"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

const (
	occlusionTestShader        = "occlusion_test.shader"
	maxOcclusionHiZLevels      = 16
	occlusionTestWorkGroupSize = 64
	occlusionInitialCapacity   = 64
)

const (
	DefaultOcclusionDepthBias  matrix.Float = 0.0005
	DefaultOcclusionRectPadPx  matrix.Float = 1.0
	DefaultOcclusionMinRectPx  matrix.Float = 1.0
	DefaultOcclusionMissingFar matrix.Float = 0.99999
)

type GPUOcclusionCandidate struct {
	WorldMinPrevious matrix.Vec4
	WorldMaxPadding  matrix.Vec4
	ScreenRect       matrix.Vec4
	DepthParams      matrix.Vec4
}

type GPUOcclusionTestParams struct {
	View             matrix.Mat4
	Projection       matrix.Mat4
	ScreenSizeCount  matrix.Vec4
	DepthSizePadding matrix.Vec4
}

type GPUOcclusionTestFrame struct {
	candidateBuffer  GPUBuffer
	candidateMemory  GPUDeviceMemory
	candidateMapping unsafe.Pointer
	resultBuffer     GPUBuffer
	resultMemory     GPUDeviceMemory
	resultMapping    unsafe.Pointer
	paramsBuffer     GPUBuffer
	paramsMemory     GPUDeviceMemory
	paramsMapping    unsafe.Pointer
	targets          []*ShaderDataBase
	capacity         int
	candidateCount   int
	resultsPending   bool
}

type GPUOcclusionTester struct {
	shader         *Shader
	descriptorSets [maxFramesInFlight]GPUDescriptorSet
	descriptorPool GPUDescriptorPool
	frames         [maxFramesInFlight]GPUOcclusionTestFrame
}

func (g *GPUPainter) QueueOcclusionCandidate(device *GPUDevice, instanceBase *ShaderDataBase) {
	defer tracing.NewRegion("GPUPainter.QueueOcclusionCandidate").End()
	if instanceBase == nil {
		return
	}
	visibility := instanceBase.VisibilityState()
	if visibility.ForceVisible || !visibility.FrustumVisible || !visibility.OcclusionEligible {
		return
	}
	if err := g.occlusionTester.queueCandidate(device, g.currentFrame, instanceBase); err != nil {
		slog.Error("failed to queue occlusion candidate", "error", err)
		visibility.LastOcclusionVisible = true
	}
}

func (g *GPUPainter) QueueOcclusionTests(device *GPUDevice, camera cameras.Camera) {
	defer tracing.NewRegion("GPUPainter.QueueOcclusionTests").End()
	g.occlusionTester.queueTests(device, g, camera)
}

func (g *GPUPainter) ApplyOcclusionResults() {
	defer tracing.NewRegion("GPUPainter.ApplyOcclusionResults").End()
	g.occlusionTester.applyResults(g.currentFrame)
}

func (g *GPUPainter) QueueOcclusionWork(device *GPUDevice, camera cameras.Camera) {
	defer tracing.NewRegion("GPUPainter.QueueOcclusionWork").End()
	g.QueueHiZPyramid(device)
	g.QueueOcclusionTests(device, camera)
	g.executeCompute(device)
}

func (g *GPUPainter) DestroyOcclusionTester(device *GPUDevice) {
	defer tracing.NewRegion("GPUPainter.DestroyOcclusionTester").End()
	g.occlusionTester.destroy(device)
}

func (t *GPUOcclusionTester) queueCandidate(device *GPUDevice, frameIdx int, instanceBase *ShaderDataBase) error {
	frame := &t.frames[frameIdx]
	if err := t.ensureFrameCapacity(device, frameIdx, frame.candidateCount+1); err != nil {
		return err
	}
	candidate := newGPUOcclusionCandidate(instanceBase.renderBounds(), instanceBase.VisibilityState().LastOcclusionVisible)
	idx := frame.candidateCount
	dst := unsafe.Pointer(uintptr(frame.candidateMapping) + uintptr(idx)*unsafe.Sizeof(candidate))
	*(*GPUOcclusionCandidate)(dst) = candidate
	res := unsafe.Pointer(uintptr(frame.resultMapping) + uintptr(idx)*unsafe.Sizeof(uint32(0)))
	*(*uint32)(res) = 1
	frame.targets = append(frame.targets, instanceBase)
	frame.candidateCount++
	return nil
}

func (t *GPUOcclusionTester) queueTests(device *GPUDevice, painter *GPUPainter, camera cameras.Camera) {
	frameIdx := painter.currentFrame
	frame := &t.frames[frameIdx]
	if frame.candidateCount == 0 {
		return
	}
	frame.resultsPending = false
	levels, ok := painter.HiZPyramidLevels()
	if !ok || len(levels) == 0 || camera == nil {
		t.failOpenFrame(frameIdx)
		return
	}
	if err := t.ensureShader(device); err != nil {
		slog.Error("failed to prepare occlusion test shader", "error", err)
		t.failOpenFrame(frameIdx)
		return
	}
	if err := t.ensureDescriptorSets(device); err != nil {
		slog.Error("failed to prepare occlusion test descriptors", "error", err)
		t.failOpenFrame(frameIdx)
		return
	}
	params := GPUOcclusionTestParams{
		View:       camera.View(),
		Projection: camera.Projection(),
		ScreenSizeCount: matrix.Vec4{
			matrix.Float(camera.Width()),
			matrix.Float(camera.Height()),
			matrix.Float(frame.candidateCount),
			matrix.Float(min(len(levels), maxOcclusionHiZLevels)),
		},
		DepthSizePadding: matrix.Vec4{
			matrix.Float(levels[0].Width),
			matrix.Float(levels[0].Height),
			DefaultOcclusionNearPlanePadding,
			DefaultOcclusionMissingFar,
		},
	}
	*(*GPUOcclusionTestParams)(frame.paramsMapping) = params
	t.writeDescriptors(device, frameIdx, levels)
	sampledImages := make([]ComputeTaskImage, 0, min(len(levels), maxOcclusionHiZLevels))
	for i := 0; i < len(levels) && i < maxOcclusionHiZLevels; i++ {
		sampledImages = append(sampledImages, ComputeTaskImage{
			Texture: &levels[i].RenderId,
			Aspect:  GPUImageAspectColorBit,
		})
	}
	painter.computeTasks = append(painter.computeTasks, ComputeTask{
		Shader:         t.shader,
		DescriptorSets: t.descriptorSets[:],
		WorkGroups:     occlusionDispatchGroups(frame.candidateCount),
		SampledImages:  sampledImages,
	})
	frame.resultsPending = true
}

func (t *GPUOcclusionTester) applyResults(frameIdx int) {
	frame := &t.frames[frameIdx]
	if frame.candidateCount == 0 {
		frame.targets = frame.targets[:0]
		return
	}
	if !frame.resultsPending {
		t.failOpenFrame(frameIdx)
		return
	}
	if frame.resultMapping == nil {
		t.failOpenFrame(frameIdx)
		return
	}
	results := unsafe.Slice((*uint32)(frame.resultMapping), frame.capacity)
	count := min(frame.candidateCount, len(frame.targets))
	for i := range count {
		target := frame.targets[i]
		if target != nil && !target.IsDestroyed() {
			target.VisibilityState().LastOcclusionVisible = results[i] != 0
		}
	}
	frame.candidateCount = 0
	frame.resultsPending = false
	clear(frame.targets)
	frame.targets = frame.targets[:0]
}

func (t *GPUOcclusionTester) failOpenFrame(frameIdx int) {
	frame := &t.frames[frameIdx]
	for i := range min(frame.candidateCount, len(frame.targets)) {
		if frame.targets[i] != nil && !frame.targets[i].IsDestroyed() {
			frame.targets[i].VisibilityState().LastOcclusionVisible = true
		}
	}
	frame.candidateCount = 0
	frame.resultsPending = false
	clear(frame.targets)
	frame.targets = frame.targets[:0]
}

func (t *GPUOcclusionTester) ensureShader(device *GPUDevice) error {
	if t.shader != nil && t.shader.RenderId.computePipeline.IsValid() {
		return nil
	}
	if device.Painter.caches == nil {
		return fmt.Errorf("render caches are not available")
	}
	mem, err := device.Painter.caches.AssetDatabase().ReadText(occlusionTestShader)
	if err != nil {
		return err
	}
	var shaderData ShaderData
	if err := json.Unmarshal([]byte(mem), &shaderData); err != nil {
		return err
	}
	shader, _ := device.Painter.caches.ShaderCache().Shader(shaderData.Compile())
	device.Painter.caches.ShaderCache().CreatePending()
	if !shader.RenderId.computePipeline.IsValid() {
		return fmt.Errorf("occlusion test compute shader did not create a valid pipeline")
	}
	t.shader = shader
	return nil
}

func (t *GPUOcclusionTester) ensureDescriptorSets(device *GPUDevice) error {
	if t.descriptorSets[0].IsValid() {
		return nil
	}
	if t.shader == nil {
		return fmt.Errorf("occlusion test shader is not ready")
	}
	var err error
	t.descriptorSets, t.descriptorPool, err = device.createDescriptorSet(t.shader.RenderId.descriptorSetLayout, 0)
	return err
}

func (t *GPUOcclusionTester) ensureFrameCapacity(device *GPUDevice, frameIdx, required int) error {
	frame := &t.frames[frameIdx]
	if required <= frame.capacity &&
		frame.candidateBuffer.IsValid() &&
		frame.resultBuffer.IsValid() &&
		frame.paramsBuffer.IsValid() {
		return nil
	}
	newCapacity := max(occlusionInitialCapacity, required)
	for newCapacity < required {
		newCapacity *= 2
	}
	t.destroyFrame(device, frameIdx)
	candidateSize := unsafe.Sizeof(GPUOcclusionCandidate{}) * uintptr(newCapacity)
	resultSize := unsafe.Sizeof(uint32(0)) * uintptr(newCapacity)
	paramsSize := unsafe.Sizeof(GPUOcclusionTestParams{})
	var err error
	frame.candidateBuffer, frame.candidateMemory, err = device.CreateBuffer(candidateSize,
		GPUBufferUsageStorageBufferBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		return err
	}
	frame.resultBuffer, frame.resultMemory, err = device.CreateBuffer(resultSize,
		GPUBufferUsageStorageBufferBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		t.destroyFrame(device, frameIdx)
		return err
	}
	frame.paramsBuffer, frame.paramsMemory, err = device.CreateBuffer(paramsSize,
		GPUBufferUsageUniformBufferBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		t.destroyFrame(device, frameIdx)
		return err
	}
	if err = device.MapMemory(frame.candidateMemory, 0, candidateSize, 0, &frame.candidateMapping); err != nil {
		t.destroyFrame(device, frameIdx)
		return err
	}
	if err = device.MapMemory(frame.resultMemory, 0, resultSize, 0, &frame.resultMapping); err != nil {
		t.destroyFrame(device, frameIdx)
		return err
	}
	if err = device.MapMemory(frame.paramsMemory, 0, paramsSize, 0, &frame.paramsMapping); err != nil {
		t.destroyFrame(device, frameIdx)
		return err
	}
	frame.capacity = newCapacity
	frame.targets = make([]*ShaderDataBase, 0, newCapacity)
	return nil
}

func (t *GPUOcclusionTester) writeDescriptors(device *GPUDevice, frameIdx int, levels []*Texture) {
	frame := &t.frames[frameIdx]
	set := vk.DescriptorSet(t.descriptorSets[frameIdx].handle)
	candidateInfo := bufferInfo(vk.Buffer(frame.candidateBuffer.handle), vk.DeviceSize(vulkan_const.WholeSize))
	resultInfo := bufferInfo(vk.Buffer(frame.resultBuffer.handle), vk.DeviceSize(vulkan_const.WholeSize))
	paramsInfo := bufferInfo(vk.Buffer(frame.paramsBuffer.handle), vk.DeviceSize(unsafe.Sizeof(GPUOcclusionTestParams{})))
	imageInfos := make([]vk.DescriptorImageInfo, maxOcclusionHiZLevels)
	last := levels[len(levels)-1]
	for i := range imageInfos {
		level := last
		if i < len(levels) {
			level = levels[i]
		}
		imageInfos[i] = imageInfoVk(vk.ImageView(level.RenderId.View.handle), vk.Sampler(level.RenderId.Sampler.handle))
	}
	writes := [4]vk.WriteDescriptorSet{
		prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{candidateInfo}, 0, vulkan_const.DescriptorTypeStorageBuffer),
		prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{resultInfo}, 1, vulkan_const.DescriptorTypeStorageBuffer),
		prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{paramsInfo}, 2, vulkan_const.DescriptorTypeUniformBuffer),
		prepareSetWriteImageTyped(set, imageInfos, 3, vulkan_const.DescriptorTypeCombinedImageSampler),
	}
	vk.UpdateDescriptorSets(vk.Device(device.LogicalDevice.handle), uint32(len(writes)), &writes[0], 0, nil)
}

func (t *GPUOcclusionTester) destroy(device *GPUDevice) {
	for i := range t.frames {
		t.destroyFrame(device, i)
	}
	t.shader = nil
	t.descriptorSets = [maxFramesInFlight]GPUDescriptorSet{}
	t.descriptorPool = GPUDescriptorPool{}
}

func (t *GPUOcclusionTester) destroyFrame(device *GPUDevice, frameIdx int) {
	frame := &t.frames[frameIdx]
	if frame.candidateMemory.IsValid() && frame.candidateMapping != nil {
		device.UnmapMemory(frame.candidateMemory)
	}
	if frame.resultMemory.IsValid() && frame.resultMapping != nil {
		device.UnmapMemory(frame.resultMemory)
	}
	if frame.paramsMemory.IsValid() && frame.paramsMapping != nil {
		device.UnmapMemory(frame.paramsMemory)
	}
	if frame.candidateBuffer.IsValid() {
		device.DestroyBuffer(frame.candidateBuffer)
	}
	if frame.resultBuffer.IsValid() {
		device.DestroyBuffer(frame.resultBuffer)
	}
	if frame.paramsBuffer.IsValid() {
		device.DestroyBuffer(frame.paramsBuffer)
	}
	if frame.candidateMemory.IsValid() {
		device.FreeMemory(frame.candidateMemory)
	}
	if frame.resultMemory.IsValid() {
		device.FreeMemory(frame.resultMemory)
	}
	if frame.paramsMemory.IsValid() {
		device.FreeMemory(frame.paramsMemory)
	}
	*frame = GPUOcclusionTestFrame{}
}

func newGPUOcclusionCandidate(bounds graviton.AABB, previousVisible bool) GPUOcclusionCandidate {
	minimum := bounds.Min()
	maximum := bounds.Max()
	prev := matrix.Float(0)
	if previousVisible {
		prev = 1
	}
	return GPUOcclusionCandidate{
		WorldMinPrevious: matrix.Vec4{minimum.X(), minimum.Y(), minimum.Z(), prev},
		WorldMaxPadding:  matrix.Vec4{maximum.X(), maximum.Y(), maximum.Z(), DefaultOcclusionRectPadPx},
		ScreenRect:       matrix.Vec4{0, 0, 1, 1},
		DepthParams:      matrix.Vec4{DefaultOcclusionDepthBias, DefaultOcclusionMinRectPx, 0, 0},
	}
}

func occlusionDispatchGroups(candidateCount int) [3]uint32 {
	return [3]uint32{uint32((candidateCount + occlusionTestWorkGroupSize - 1) / occlusionTestWorkGroupSize), 1, 1}
}

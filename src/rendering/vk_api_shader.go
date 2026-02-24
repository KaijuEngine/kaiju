/******************************************************************************/
/* vk_api_shader.go                                                           */
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
	"errors"
	"kaiju/engine/assets"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"runtime"
	"unsafe"

	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

type ShaderCleanup struct {
	id       ShaderId
	renderer Renderer
}

type FuncPipeline func(renderer Renderer, shader *Shader, shaderStages []vk.PipelineShaderStageCreateInfo) bool

func (vr *Vulkan) CreateShader(shader *Shader, assetDB assets.Database) error {
	defer tracing.NewRegion("Vulkan.CreateShader").End()
	switch shader.Type {
	case ShaderTypeGraphics:
		return vr.createGraphicsShader(shader, assetDB)
	case ShaderTypeCompute:
		return vr.createComputeShader(shader, assetDB)
	}
	return errors.New("unhandled shader type")
}

func (vr *Vulkan) createGraphicsShader(shader *Shader, assetDB assets.Database) error {
	defer tracing.NewRegion("Vulkan.createGraphicsShader").End()
	var vert, frag, geom, tesc, tese vk.ShaderModule
	var vMem, fMem, gMem, cMem, eMem []byte
	vertStage := vk.PipelineShaderStageCreateInfo{}
	vMem, err := assetDB.Read(shader.data.Vertex)
	if err != nil {
		slog.Error("Failed to load vertex shader", "module", shader.data.Vertex, "error", err)
		return err
	}
	vert, ok := vr.createSpvModule(vMem)
	if !ok {
		slog.Error("Failed to load vertex module", "module", shader.data.Vertex)
		return err
	}
	vertStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
	vertStage.Stage = vulkan_const.ShaderStageVertexBit
	vertStage.Module = vert
	vertStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.vertModule = vert

	fragStage := vk.PipelineShaderStageCreateInfo{}
	fMem, err = assetDB.Read(shader.data.Fragment)
	if err != nil {
		slog.Error("Failed to load fragment shader", "module", shader.data.Fragment, "error", err)
		return err
	}
	frag, ok = vr.createSpvModule(fMem)
	if !ok {
		slog.Error("Failed to load fragment module", "module", shader.data.Fragment)
		return err
	}
	fragStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
	fragStage.Stage = vulkan_const.ShaderStageFragmentBit
	fragStage.Module = frag
	fragStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.fragModule = frag

	geomStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.Geometry) > 0 {
		gMem, err = assetDB.Read(shader.data.Geometry)
		if err != nil {
			slog.Error("Failed to load geometry shader", "module", shader.data.Geometry, "error", err)
			return err
		}
		geom, ok = vr.createSpvModule(gMem)
		if !ok {
			slog.Error("Failed to load geometry module", "module", shader.data.Geometry)
			return err
		}
		geomStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
		geomStage.Stage = vulkan_const.ShaderStageGeometryBit
		geomStage.Module = geom
		geomStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	}

	tescStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.TessellationControl) > 0 {
		cMem, err = assetDB.Read(shader.data.TessellationControl)
		if err != nil {
			slog.Error("Failed to load tessellation control shader", "module", shader.data.TessellationControl, "error", err)
			return err
		}
		tesc, ok = vr.createSpvModule(cMem)
		if !ok {
			slog.Error("Failed to load tessellation control module", "module", shader.data.TessellationControl)
			return err
		}
		tescStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
		tescStage.Stage = vulkan_const.ShaderStageTessellationControlBit
		tescStage.Module = tesc
		tescStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.tescModule = tesc
	}

	teseStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.TessellationEvaluation) > 0 {
		eMem, err = assetDB.Read(shader.data.TessellationEvaluation)
		if err != nil {
			slog.Error("Failed to load tessellation evaluation shader", "module", shader.data.TessellationEvaluation, "error", err)
			return err
		}
		tese, ok = vr.createSpvModule(eMem)
		if !ok {
			slog.Error("Failed to load tessellation evaluation module", "module", shader.data.TessellationEvaluation)
			return err
		}
		teseStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
		teseStage.Stage = vulkan_const.ShaderStageTessellationEvaluationBit
		teseStage.Module = tese
		teseStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.teseModule = tese
	}

	id := &shader.RenderId
	shader.DriverData.setup(&shader.data)
	vkDevice := vk.Device(vr.app.FirstInstance().PrimaryDevice().LogicalDevice.handle)
	id.descriptorSetLayout, err = vr.createDescriptorSetLayout(vkDevice,
		shader.DriverData.DescriptorSetLayoutStructure)
	if err != nil {
		// TODO:  Handle this error properly
		slog.Error(err.Error())
	}
	stages := make([]vk.PipelineShaderStageCreateInfo, 0)
	if vertStage.SType != 0 {
		stages = append(stages, vertStage)
	}
	if fragStage.SType != 0 {
		stages = append(stages, fragStage)
	}
	if geomStage.SType != 0 {
		stages = append(stages, geomStage)
	}
	if tescStage.SType != 0 {
		stages = append(stages, tescStage)
	}
	if teseStage.SType != 0 {
		stages = append(stages, teseStage)
	}
	shader.pipelineInfo.ConstructPipeline(vr, shader, shader.renderPass.Value(), stages)
	runtime.AddCleanup(shader, func(state ShaderCleanup) {
		v := state.renderer.(*Vulkan)
		v.preRuns = append(v.preRuns, func() {
			state.renderer.(*Vulkan).destroyShaderHandle(state.id)
		})
	}, ShaderCleanup{shader.RenderId, vr})
	return nil
}

func (vr *Vulkan) createComputeShader(shader *Shader, assetDB assets.Database) error {
	defer tracing.NewRegion("Vulkan.createComputeShader").End()
	compStage := vk.PipelineShaderStageCreateInfo{}
	cMem, err := assetDB.Read(shader.data.Compute)
	if err != nil {
		return err
	}
	comp, ok := vr.createSpvModule(cMem)
	if !ok {
		slog.Error("Failed to load compute module", "module", shader.data.Compute)
		return err
	}
	compStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
	compStage.Stage = vulkan_const.ShaderStageComputeBit
	compStage.Module = comp
	compStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.compModule = comp
	id := &shader.RenderId
	shader.DriverData.setup(&shader.data)
	vkDevice := vk.Device(vr.app.FirstInstance().PrimaryDevice().LogicalDevice.handle)
	id.descriptorSetLayout, err = vr.createDescriptorSetLayout(vkDevice, shader.DriverData.DescriptorSetLayoutStructure)
	if err != nil {
		return err
	}
	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vulkan_const.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            &shader.RenderId.descriptorSetLayout,
		PushConstantRangeCount: 0, // Adjust if push constants are used
		PPushConstantRanges:    nil,
	}
	var pLayout vk.PipelineLayout
	if vk.CreatePipelineLayout(vkDevice, &pipelineLayoutInfo, nil, &pLayout) != vulkan_const.Success {
		slog.Error("Failed to create pipeline layout")
		return errors.New("failed to create pipeline layout")
	} else {
		vr.app.Dbg().track(unsafe.Pointer(pLayout))
	}
	shader.RenderId.pipelineLayout = pLayout
	pipelineInfo := vk.ComputePipelineCreateInfo{
		SType:  vulkan_const.StructureTypeComputePipelineCreateInfo,
		Stage:  compStage,
		Layout: id.pipelineLayout,
	}
	pipelines := [1]vk.Pipeline{}
	if vk.CreateComputePipelines(vkDevice, vk.NullPipelineCache, 1, &pipelineInfo, nil, &pipelines[0]) != vulkan_const.Success {
		slog.Error("Failed to create compute pipeline")
		return errors.New("failed to create compute pipeline")
	} else {
		vr.app.Dbg().track(unsafe.Pointer(pipelines[0]))
	}
	id.computePipeline = pipelines[0]
	return nil
}

func (vr *Vulkan) createSpvModule(mem []byte) (vk.ShaderModule, bool) {
	defer tracing.NewRegion("Vulkan.createSpvModule").End()
	info := vk.ShaderModuleCreateInfo{}
	info.SType = vulkan_const.StructureTypeShaderModuleCreateInfo
	info.CodeSize = uint(len(mem))
	info.PCode = (*uint32)(unsafe.Pointer(&mem[0]))
	var outModule vk.ShaderModule
	vkDevice := vk.Device(vr.app.FirstInstance().PrimaryDevice().LogicalDevice.handle)
	if vk.CreateShaderModule(vkDevice, &info, nil, &outModule) != vulkan_const.Success {
		slog.Error("Failed to create shader module", slog.String("module", string(mem)))
		return outModule, false
	} else {
		vr.app.Dbg().track(unsafe.Pointer(outModule))
		return outModule, true
	}
}

func (vr *Vulkan) destroyShaderHandle(id ShaderId) {
	defer tracing.NewRegion("Vulkan.DestroyShader").End()
	device := vr.app.FirstInstance().PrimaryDevice()
	vkDevice := vk.Device(device.LogicalDevice.handle)
	device.LogicalDevice.WaitIdle()
	vk.DestroyPipeline(vkDevice, id.graphicsPipeline, nil)
	vr.app.Dbg().remove(unsafe.Pointer(id.graphicsPipeline))
	vk.DestroyPipeline(vkDevice, id.computePipeline, nil)
	vr.app.Dbg().remove(unsafe.Pointer(id.computePipeline))
	vk.DestroyPipelineLayout(vkDevice, id.pipelineLayout, nil)
	vr.app.Dbg().remove(unsafe.Pointer(id.pipelineLayout))
	vk.DestroyShaderModule(vkDevice, id.vertModule, nil)
	vr.app.Dbg().remove(unsafe.Pointer(id.vertModule))
	vk.DestroyShaderModule(vkDevice, id.fragModule, nil)
	vr.app.Dbg().remove(unsafe.Pointer(id.fragModule))
	if id.geomModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vkDevice, id.geomModule, nil)
		vr.app.Dbg().remove(unsafe.Pointer(id.geomModule))
	}
	if id.tescModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vkDevice, id.tescModule, nil)
		vr.app.Dbg().remove(unsafe.Pointer(id.tescModule))
	}
	if id.teseModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vkDevice, id.teseModule, nil)
		vr.app.Dbg().remove(unsafe.Pointer(id.teseModule))
	}
	if id.compModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vkDevice, id.compModule, nil)
		vr.app.Dbg().remove(unsafe.Pointer(id.compModule))
	}
	vk.DestroyDescriptorSetLayout(vkDevice, id.descriptorSetLayout, nil)
	vr.app.Dbg().remove(unsafe.Pointer(id.descriptorSetLayout))
}

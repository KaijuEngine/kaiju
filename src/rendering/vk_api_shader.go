/******************************************************************************/
/* vk_api_shader.go                                                           */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"kaiju/assets"
	"log/slog"
	"strings"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

type FuncPipeline func(renderer Renderer, shader *Shader, shaderStages []vk.PipelineShaderStageCreateInfo) bool

func (vr *Vulkan) CreateShader(shader *Shader, assetDB *assets.Database) {
	var vert, frag, geom, tesc, tese vk.ShaderModule
	var vMem, fMem, gMem, cMem, eMem []byte
	vertStage := vk.PipelineShaderStageCreateInfo{}
	vMem, err := assetDB.Read(shader.data.Vertex)
	if err != nil {
		panic("Failed to load vertex shader")
	}
	vert, ok := vr.createSpvModule(vMem)
	if !ok {
		panic("Failed to create vertex shader module")
	}
	vertStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
	vertStage.Stage = vk.ShaderStageVertexBit
	vertStage.Module = vert
	vertStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.vertModule = vert

	fragStage := vk.PipelineShaderStageCreateInfo{}
	fMem, err = assetDB.Read(shader.data.Fragment)
	if err != nil {
		panic("Failed to load fragment shader")
	}
	frag, ok = vr.createSpvModule(fMem)
	if !ok {
		panic("Failed to create fragment shader module")
	}
	fragStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
	fragStage.Stage = vk.ShaderStageFragmentBit
	fragStage.Module = frag
	fragStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.fragModule = frag

	geomStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.Geometry) > 0 {
		gMem, err = assetDB.Read(shader.data.Geometry)
		if err != nil {
			panic("Failed to load geometry shader")
		}
		geom, ok = vr.createSpvModule(gMem)
		if !ok {
			panic("Failed to create geometry shader module")
		}
		geomStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		geomStage.Stage = vk.ShaderStageGeometryBit
		geomStage.Module = geom
		geomStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	}

	tescStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.TessellationControl) > 0 {
		cMem, err = assetDB.Read(shader.data.TessellationControl)
		if err != nil {
			panic("Failed to load tessellation control shader")
		}
		tesc, ok = vr.createSpvModule(cMem)
		if !ok {
			panic("Failed to create tessellation control shader module")
		}
		tescStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		tescStage.Stage = vk.ShaderStageTessellationControlBit
		tescStage.Module = tesc
		tescStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.tescModule = tesc
	}

	teseStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.TessellationEvaluation) > 0 {
		eMem, err = assetDB.Read(shader.data.TessellationEvaluation)
		if err != nil {
			panic("Failed to load tessellation evaluation shader")
		}
		tese, ok = vr.createSpvModule(eMem)
		if !ok {
			panic("Failed to create tessellation evaluation shader module")
		}
		teseStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		teseStage.Stage = vk.ShaderStageTessellationEvaluationBit
		teseStage.Module = tese
		teseStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.teseModule = tese
	}

	id := &shader.RenderId

	id.descriptorSetLayout, err = vr.createDescriptorSetLayout(vr.device,
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

	shader.DriverData.pipelineConstructor(vr, shader, stages)
	// TODO:  Setup subshader in the shader definition?
	subShaderCheck := strings.TrimSuffix(shader.data.Fragment, ".spv") + oitSuffix
	if assetDB.Exists(subShaderCheck) {
		cpy := shader.data
		cpy.Fragment = subShaderCheck
		subShader := NewShader(cpy)
		subShader.DriverData = shader.DriverData
		shader.AddSubShader("transparent", subShader)
	}
}

func (vr *Vulkan) createSpvModule(mem []byte) (vk.ShaderModule, bool) {
	info := vk.ShaderModuleCreateInfo{}
	info.SType = vk.StructureTypeShaderModuleCreateInfo
	info.CodeSize = uint(len(mem))
	info.PCode = (*uint32)(unsafe.Pointer(&mem[0]))
	var outModule vk.ShaderModule
	if vk.CreateShaderModule(vr.device, &info, nil, &outModule) != vk.Success {
		slog.Error("Failed to create shader module", slog.String("module", string(mem)))
		return outModule, false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(outModule))
		return outModule, true
	}
}

func (vr *Vulkan) DestroyShader(shader *Shader) {
	vk.DeviceWaitIdle(vr.device)
	vk.DestroyPipeline(vr.device, shader.RenderId.graphicsPipeline, nil)
	vr.dbg.remove(vk.TypeToUintPtr(shader.RenderId.graphicsPipeline))
	vk.DestroyPipelineLayout(vr.device, shader.RenderId.pipelineLayout, nil)
	vr.dbg.remove(vk.TypeToUintPtr(shader.RenderId.pipelineLayout))
	vk.DestroyShaderModule(vr.device, shader.RenderId.vertModule, nil)
	vr.dbg.remove(vk.TypeToUintPtr(shader.RenderId.vertModule))
	vk.DestroyShaderModule(vr.device, shader.RenderId.fragModule, nil)
	vr.dbg.remove(vk.TypeToUintPtr(shader.RenderId.fragModule))
	if shader.RenderId.geomModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.geomModule, nil)
		vr.dbg.remove(vk.TypeToUintPtr(shader.RenderId.geomModule))
	}
	if shader.RenderId.tescModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.tescModule, nil)
		vr.dbg.remove(vk.TypeToUintPtr(shader.RenderId.tescModule))
	}
	if shader.RenderId.teseModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.teseModule, nil)
		vr.dbg.remove(vk.TypeToUintPtr(shader.RenderId.teseModule))
	}
	vk.DestroyDescriptorSetLayout(vr.device, shader.RenderId.descriptorSetLayout, nil)
	vr.dbg.remove(vk.TypeToUintPtr(shader.RenderId.descriptorSetLayout))
	for _, ss := range shader.subShaders {
		vr.DestroyShader(ss)
	}
}

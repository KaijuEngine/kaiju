package rendering

import (
	"kaiju/assets"
	"log/slog"
	"strings"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) CreateShader(shader *Shader, assetDB *assets.Database) {
	var vert, frag, geom, tesc, tese vk.ShaderModule
	var vMem, fMem, gMem, cMem, eMem []byte
	vertStage := vk.PipelineShaderStageCreateInfo{}
	vMem, err := assetDB.Read(shader.VertPath)
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
	fMem, err = assetDB.Read(shader.FragPath)
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
	if len(shader.GeomPath) > 0 {
		gMem, err = assetDB.Read(shader.GeomPath)
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
	if len(shader.CtrlPath) > 0 {
		cMem, err = assetDB.Read(shader.CtrlPath)
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
	if len(shader.EvalPath) > 0 {
		eMem, err = assetDB.Read(shader.EvalPath)
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

	stages := []vk.PipelineShaderStageCreateInfo{vertStage, tescStage, teseStage, geomStage, fragStage}
	moduleCount := 0
	if vertStage.SType != 0 {
		stages[moduleCount] = vertStage
		moduleCount++
	}
	if tescStage.SType != 0 {
		stages[moduleCount] = tescStage
		moduleCount++
	}
	if teseStage.SType != 0 {
		stages[moduleCount] = teseStage
		moduleCount++
	}
	if geomStage.SType != 0 {
		stages[moduleCount] = geomStage
		moduleCount++
	}
	if fragStage.SType != 0 {
		stages[moduleCount] = fragStage
		moduleCount++
	}
	renderPass := vr.defaultTarget.opaqueRenderPass
	if strings.HasSuffix(shader.FragPath, oitSuffix) || shader.IsComposite() {
		renderPass = vr.defaultTarget.transparentRenderPass
	}

	isTransparentPipeline := renderPass == vr.defaultTarget.transparentRenderPass &&
		!shader.IsComposite()
	vr.createPipeline(shader, stages, moduleCount,
		id.descriptorSetLayout, &id.pipelineLayout,
		&id.graphicsPipeline, renderPass, isTransparentPipeline)
	// TODO:  Setup subshader in the shader definition?
	subShaderCheck := strings.TrimSuffix(shader.FragPath, ".spv") + oitSuffix
	if assetDB.Exists(subShaderCheck) {
		subShader := NewShader(shader.VertPath, subShaderCheck,
			shader.GeomPath, shader.CtrlPath, shader.EvalPath, vr)
		subShader.DriverData = shader.DriverData
		shader.SubShader = subShader
	}
}

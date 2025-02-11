/******************************************************************************/
/* vk_combine_canvas.go                                                       */
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
	"errors"
	"kaiju/matrix"
	"log/slog"

	vk "kaiju/rendering/vulkan"
)

type CombineCanvas struct {
	renderPass  RenderPass
	color       TextureId
	frameBuffer vk.Framebuffer
	texture     Texture
}

func (r *CombineCanvas) Pass(name string) *RenderPass {
	return &r.renderPass
}

func (r *CombineCanvas) Color() *Texture { return &r.texture }

func (r *CombineCanvas) Draw(renderer Renderer, drawings []ShaderDraw) {
	vr := renderer.(*Vulkan)
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	for i := range drawings {
		vr.writeDrawingDescriptors(drawings[i].shader, drawings[i].instanceGroups)
	}
	oRenderPass := r.renderPass
	cmd1 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	var opaqueClear [2]vk.ClearValue
	cc := matrix.ColorDarkBG()
	opaqueClear[0].SetColor(cc[:])
	opaqueClear[1].SetDepthStencil(1.0, 0.0)
	beginRender(oRenderPass, vr.swapChainExtent, cmd1, opaqueClear[:])
	for i := range drawings {
		vr.renderEach(cmd1, drawings[i].shader, drawings[i].instanceGroups)
	}
	endRender(cmd1)
}

func (r *CombineCanvas) Create(renderer Renderer) error {
	vr := renderer.(*Vulkan)
	if !r.createImage(vr) {
		return errors.New("failed to create render target images")
	}
	if !r.createRenderPass(vr) {
		return errors.New("failed to create OIT render pass")
	}
	r.texture.RenderId = r.color
	return nil
}

func (r *CombineCanvas) Destroy(renderer Renderer) {
	vr := renderer.(*Vulkan)
	vk.DeviceWaitIdle(vr.device)
	r.renderPass.Destroy()
	vr.textureIdFree(&r.color)
	vk.DestroyFramebuffer(vr.device, r.frameBuffer, nil)
	vr.dbg.remove(vk.TypeToUintPtr(r.frameBuffer))
	r.color = TextureId{}
}

func (r *CombineCanvas) ShaderPipeline(name string) FuncPipeline {
	return defaultCombinePipeline
}

func (r *CombineCanvas) createRenderPass(renderer Renderer) bool {
	vr := renderer.(*Vulkan)
	attachment := vk.AttachmentDescription{
		Format:         r.color.Format,
		Samples:        r.color.Samples,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutColorAttachmentOptimal,
	}
	colorAttachmentRef := vk.AttachmentReference{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}
	subpass := vk.SubpassDescription{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    &colorAttachmentRef,
	}
	pass, err := NewRenderPass(vr.device, &vr.dbg, []vk.AttachmentDescription{attachment},
		[]vk.SubpassDescription{subpass}, []vk.SubpassDependency{})
	if err != nil {
		slog.Error("Failed to create the combine render pass")
		return false
	}
	r.renderPass = pass
	err = r.renderPass.CreateFrameBuffer(vr,
		[]vk.ImageView{r.color.View}, r.color.Width, r.color.Height)
	if err != nil {
		slog.Error("Failed to create the combine frame buffer")
		return false
	}
	return true
}

func (r *CombineCanvas) createImage(vr *Vulkan) bool {
	w := uint32(vr.swapChainExtent.Width)
	h := uint32(vr.swapChainExtent.Height)
	samples := vk.SampleCount1Bit
	imagesCreated := vr.CreateImage(w, h, 1, samples,
		vk.FormatB8g8r8a8Unorm, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageTransferSrcBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &r.color, 1)
	imagesCreated = imagesCreated && vr.createImageView(&r.color,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	vr.createTextureSampler(&r.color.Sampler, 1, vk.FilterLinear)
	if imagesCreated {
		vr.transitionImageLayout(&r.color,
			vk.ImageLayoutColorAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentWriteBit), vk.NullCommandBuffer)
	}
	return imagesCreated
}

func defaultCombinePipeline(renderer Renderer, shader *Shader, shaderStages []vk.PipelineShaderStageCreateInfo) bool {
	vr := renderer.(*Vulkan)
	viewport := vk.Viewport{
		X:        0.0,
		Y:        0.0,
		Width:    float32(vr.swapChainExtent.Width),
		Height:   float32(vr.swapChainExtent.Height),
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: vr.swapChainExtent,
	}
	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		PViewports:    &viewport,
		ScissorCount:  1,
		PScissors:     &scissor,
	}
	dynamicStates := [...]vk.DynamicState{
		vk.DynamicStateViewport,
		vk.DynamicStateScissor,
	}
	dynamicState := vk.PipelineDynamicStateCreateInfo{
		SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: uint32(len(dynamicStates)),
		PDynamicStates:    &dynamicStates[0],
	}
	bDesc := vertexGetBindingDescription(shader)
	bDescCount := uint32(len(bDesc))
	for i := uint32(1); i < bDescCount; i++ {
		bDesc[i].Stride = uint32(vr.padUniformBufferSize(vk.DeviceSize(bDesc[i].Stride)))
	}
	aDesc := vertexGetAttributeDescription(shader)
	vertexInputInfo := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   bDescCount,
		VertexAttributeDescriptionCount: uint32(len(aDesc)),
		PVertexBindingDescriptions:      &bDesc[0],
		PVertexAttributeDescriptions:    &aDesc[0],
	}
	topology := vk.PrimitiveTopologyTriangleList
	switch shader.DriverData.DrawMode {
	case MeshDrawModePoints:
		topology = vk.PrimitiveTopologyPointList
	case MeshDrawModeLines:
		topology = vk.PrimitiveTopologyLineList
	case MeshDrawModeTriangles:
		topology = vk.PrimitiveTopologyTriangleList
	case MeshDrawModePatches:
		topology = vk.PrimitiveTopologyPatchList
	}
	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		PrimitiveRestartEnable: vk.False,
		Topology:               topology,
	}
	rasterizer := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		DepthClampEnable:        vk.False,
		RasterizerDiscardEnable: vk.False,
		PolygonMode:             vk.PolygonModeFill,
		LineWidth:               1.0,
		CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceClockwise,
	}
	multisampling := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		SampleShadingEnable:   vk.False,
		RasterizationSamples:  vk.SampleCount1Bit,
		MinSampleShading:      1.0,
		PSampleMask:           nil,
		AlphaToCoverageEnable: vk.False,
		AlphaToOneEnable:      vk.False,
	}
	allChannels := vk.ColorComponentFlags(vk.ColorComponentRBit | vk.ColorComponentGBit | vk.ColorComponentBBit | vk.ColorComponentABit)
	colorBlendAttachment := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask:      allChannels,
		BlendEnable:         vk.False,
		SrcColorBlendFactor: vk.BlendFactorOne,
		DstColorBlendFactor: vk.BlendFactorZero,
		ColorBlendOp:        vk.BlendOpAdd,
		SrcAlphaBlendFactor: vk.BlendFactorOne,
		DstAlphaBlendFactor: vk.BlendFactorZero,
		AlphaBlendOp:        vk.BlendOpAdd,
	}
	colorBlending := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		LogicOp:         vk.LogicOpCopy,
		AttachmentCount: 1,
		PAttachments:    &colorBlendAttachment,
		BlendConstants:  [4]float32{0.0, 0.0, 0.0, 0.0},
	}
	layoutInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            &shader.RenderId.descriptorSetLayout,
		PushConstantRangeCount: 0,
		PPushConstantRanges:    nil,
	}
	var layout vk.PipelineLayout
	if vk.CreatePipelineLayout(vr.device, &layoutInfo, nil, &layout) != vk.Success {
		slog.Error("Failed to create pipeline layout")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(layout))
	}
	shader.RenderId.pipelineLayout = layout
	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          2,
		PStages:             &shaderStages[0],
		PVertexInputState:   &vertexInputInfo,
		PInputAssemblyState: &inputAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterizer,
		PMultisampleState:   &multisampling,
		PDepthStencilState:  nil,
		PColorBlendState:    &colorBlending,
		PDynamicState:       &dynamicState,
		Layout:              layout,
		RenderPass:          shader.RenderPass.Handle,
		Subpass:             0,
		BasePipelineHandle:  vk.Pipeline(vk.NullHandle),
		BasePipelineIndex:   -1,
	}
	success := true
	pipelines := [1]vk.Pipeline{}
	if vk.CreateGraphicsPipelines(vr.device, vk.PipelineCache(vk.NullHandle), 1, &pipelineInfo, nil, &pipelines[0]) != vk.Success {
		success = false
		slog.Error("Failed to create graphics pipeline")
	} else {
		vr.dbg.add(vk.TypeToUintPtr(pipelines[0]))
	}
	shader.RenderId.graphicsPipeline = pipelines[0]
	return success
}

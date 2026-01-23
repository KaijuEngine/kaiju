/******************************************************************************/
/* vk_render_pass.go                                                          */
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
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"weak"

	"kaiju/engine/assets"
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

type RenderPass struct {
	Handle       vk.RenderPass
	Buffer       vk.Framebuffer
	device       vk.Device
	dbg          *debugVulkan
	textures     []Texture
	construction RenderPassDataCompiled
	subpasses    []RenderPassSubpass
	cmd          [maxFramesInFlight]CommandRecorder
	cmdSecondary [maxFramesInFlight]CommandRecorderSecondary
	currentIdx   int
	subpassIdx   int
	frame        int
}

type RenderPassSubpass struct {
	shader         *Shader
	shaderPipeline ShaderPipelineDataCompiled
	descriptorSets [maxFramesInFlight]vk.DescriptorSet
	descriptorPool vk.DescriptorPool
	sampledImages  []int
	renderQuad     *Mesh
	cmd            [maxFramesInFlight]CommandRecorderSecondary
}

func (r *RenderPass) Texture(index int) *Texture { return &r.textures[index] }

func (r *RenderPass) IsShadowPass() bool {
	// TODO:  Need another way to denote this is a shadow pass
	return strings.HasPrefix(r.construction.Name, "light_offscreen")
}

func (r *RenderPass) ExecuteSecondaryCommands() {
	buffs := [1]vk.CommandBuffer{}
	rec := &r.cmdSecondary[r.frame]
	if r.currentIdx > 0 {
		rec = &r.subpasses[r.currentIdx-1].cmd[r.frame]
	}
	rec.End()
	buffs[0] = rec.buffer
	vk.CmdExecuteCommands(r.cmd[r.frame].buffer, uint32(len(buffs)), &buffs[0])
}

func (r *RenderPass) SelectOutputAttachment(vr *Vulkan) *Texture {
	targetFormat := vr.swapImages[0].Format
	var fallback *Texture
	for i := range r.construction.AttachmentDescriptions {
		a := &r.construction.AttachmentDescriptions[i]
		if (a.Image.Usage & vk.ImageUsageFlags(vulkan_const.ImageUsageColorAttachmentBit)) != 0 {
			if fallback == nil {
				// First image is likely the better image to fall back to
				fallback = &r.textures[i]
			}
			if a.Format == targetFormat {
				return &r.textures[i]
			}
		}
	}
	// Matching image not found, search in remote connected passes
	for i := range r.construction.AttachmentDescriptions {
		a := &r.construction.AttachmentDescriptions[i]
		if a.Format == targetFormat {
			if a.Image.ExistingImage != "" {
				for _, p := range vr.renderPassCache {
					if t, ok := p.findTextureByName(a.Image.ExistingImage); ok {
						return t
					}
				}
			}
		}
	}
	if fallback != nil {
		return fallback
	}
	for i := range r.textures {
		if !isDepthFormat(r.textures[i].RenderId.Format) {
			slog.Error("failed to find an output color attachment for the render pass, using fallback", "renderPass", r.construction.Name)
			return &r.textures[i]
		}
	}
	return nil
}

func (r *RenderPass) SelectOutputAttachmentWithSuffix(vr *Vulkan, suffix string) (*Texture, bool) {
	for i := range r.construction.AttachmentDescriptions {
		if strings.HasSuffix(r.construction.AttachmentDescriptions[i].Image.Name, suffix) {
			return &r.textures[i], true
		}
	}
	return nil, false
}

func (r *RenderPass) findTextureByName(name string) (*Texture, bool) {
	for i := range r.textures {
		if r.textures[i].Key == name {
			return &r.textures[i], true
		}
	}
	return nil, false
}

func (r *RenderPass) setupSubpass(c *RenderPassSubpassDataCompiled, vr *Vulkan, assets assets.Database, index int) error {
	r.subpasses = klib.WipeSlice(r.subpasses)
	sp := RenderPassSubpass{}
	// TODO:  This is copied from Material.Compile
	{
		shaderConfig, err := assets.ReadText(c.Shader)
		if err != nil {
			return err
		}
		pipeConfig, err := assets.ReadText(c.ShaderPipeline)
		if err != nil {
			return err
		}
		var pipe ShaderPipelineData
		var rawSD ShaderData
		if err := json.Unmarshal([]byte(pipeConfig), &pipe); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(shaderConfig), &rawSD); err != nil {
			return err
		}
		sp.shaderPipeline = pipe.Compile(vr)
		shaderCache := vr.caches.ShaderCache()
		sp.shader, _ = shaderCache.Shader(rawSD.Compile())
		sp.shader.pipelineInfo = &sp.shaderPipeline
		sp.shader.renderPass = weak.Make(r)
		shaderCache.CreatePending()
	}
	sp.descriptorSets, sp.descriptorPool = klib.MustReturn2(
		vr.createDescriptorSet(sp.shader.RenderId.descriptorSetLayout, 0))
	for i := range c.SampledImages {
		t := &r.textures[c.SampledImages[i]].RenderId
		if t.Sampler == vk.NullSampler {
			vr.createTextureSampler(&t.Sampler, t.MipLevels, vulkan_const.FilterLinear)
		}
	}
	sp.sampledImages = append(sp.sampledImages, c.SampledImages...)
	sp.renderQuad = NewMeshUnitQuad(vr.caches.MeshCache())
	vr.caches.MeshCache().CreatePending()
	var err error
	for i := range len(sp.cmd) {
		if sp.cmd[i], err = NewCommandRecorderSecondary(vr, r, index); err != nil {
			return err
		}
	}
	r.subpasses = append(r.subpasses, sp)
	return nil
}

func (r *RenderPass) endSubpasses() {
	vk.CmdEndRenderPass(r.cmd[r.frame].buffer)
	r.cmd[r.frame].End()
}

func (r *RenderPass) beginNextSubpass(currentFrame int, extent vk.Extent2D, clearColors []vk.ClearValue) {
	r.frame = currentFrame
	viewport := vk.Viewport{
		X:        0,
		Y:        0,
		Width:    float32(extent.Width),
		Height:   float32(extent.Height),
		MinDepth: 0,
		MaxDepth: 1,
	}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: extent,
	}
	if r.subpassIdx == 0 {
		renderPassInfo := vk.RenderPassBeginInfo{
			SType:       vulkan_const.StructureTypeRenderPassBeginInfo,
			RenderPass:  r.Handle,
			Framebuffer: r.Buffer,
			RenderArea: vk.Rect2D{
				Offset: vk.Offset2D{X: 0, Y: 0},
				Extent: extent,
			},
			ClearValueCount: uint32(len(clearColors)),
		}
		if len(clearColors) > 0 {
			renderPassInfo.PClearValues = &clearColors[0]
		}
		r.cmd[r.frame].Begin()
		vk.CmdBeginRenderPass(r.cmd[r.frame].buffer, &renderPassInfo, vulkan_const.SubpassContentsSecondaryCommandBuffers)
		r.cmdSecondary[r.frame].Begin(viewport, scissor)
	} else {
		sp := &r.subpasses[r.subpassIdx-1]
		sp.cmd[r.frame].Reset()
		vk.CmdNextSubpass(r.cmd[r.frame].buffer, vulkan_const.SubpassContentsSecondaryCommandBuffers)
		sp.cmd[r.frame].Begin(viewport, scissor)
	}
	r.currentIdx = r.subpassIdx
	r.subpassIdx++
	if r.subpassIdx > len(r.subpasses) {
		r.subpassIdx = 0
	}
}

func isDepthFormat(format vulkan_const.Format) bool {
	switch format {
	case vulkan_const.FormatD16Unorm, vulkan_const.FormatD32Sfloat, vulkan_const.FormatD16UnormS8Uint,
		vulkan_const.FormatD24UnormS8Uint, vulkan_const.FormatD32SfloatS8Uint:
		return true
	}
	return false
}

func NewRenderPass(vr *Vulkan, setup *RenderPassDataCompiled) (*RenderPass, error) {
	p := &RenderPass{
		device:       vr.device,
		dbg:          &vr.dbg,
		construction: *setup,
		textures:     make([]Texture, 0, len(setup.AttachmentDescriptions)),
	}
	for i := range len(setup.AttachmentDescriptions) {
		a := &setup.AttachmentDescriptions[i]
		img := &a.Image
		if a.Image.IsInvalid() {
			continue
		}
		k := img.Name
		if k == "" {
			k = fmt.Sprintf("renderPass-%s-%d", setup.Name, i)
		}
		p.textures = append(p.textures, Texture{Key: k})
	}
	return p, p.Recontstruct(vr)
}

func (p *RenderPass) Recontstruct(vr *Vulkan) error {
	p.Destroy(vr)
	r := &p.construction
	var err error
	for i := range len(p.cmd) {
		if p.cmd[i], err = NewCommandRecorder(vr); err != nil {
			return err
		}
	}
	for i := range len(p.cmdSecondary) {
		if p.cmdSecondary[i], err = NewCommandRecorderSecondary(vr, p, 0); err != nil {
			return nil
		}
	}
	{
		w := max(vr.swapChainExtent.Width, uint32(r.Width))
		h := max(vr.swapChainExtent.Height, uint32(r.Height))
		for i := range len(r.AttachmentDescriptions) {
			a := &r.AttachmentDescriptions[i]
			img := &a.Image
			if a.Image.IsInvalid() {
				continue
			}
			p.textures[i].Width = int(w)
			p.textures[i].Height = int(h)
			success := vr.CreateImage(&p.textures[i].RenderId, img.MemoryProperty,
				vk.ImageCreateInfo{
					ImageType: vulkan_const.ImageType2d,
					Extent: vk.Extent3D{
						Width:  w,
						Height: h,
					},
					MipLevels:   img.MipLevels,
					ArrayLayers: img.LayerCount,
					Format:      a.Format,
					Tiling:      img.Tiling,
					Usage:       img.Usage,
					Samples:     a.Samples,
				})
			if !success {
				const e = "failed to create image for render pass attachment"
				slog.Error(e, "attachmentIndex", i)
				return errors.New(e)
			}
			success = vr.createImageView(&p.textures[i].RenderId, img.Aspect, vulkan_const.ImageViewType2d)
			if !success {
				const e = "failed to create image view for render pass attachment"
				for j := range i + 1 {
					p.textures[j].RenderId = vr.textureIdFree(p.textures[j].RenderId)
				}
				slog.Error(e, "attachmentIndex", i)
				return errors.New(e)
			}
			success = vr.createTextureSampler(&p.textures[i].RenderId.Sampler,
				img.MipLevels, img.Filter)
			if !success {
				const e = "failed to create image sampler for render pass attachment"
				for j := range i + 1 {
					p.textures[j].RenderId = vr.textureIdFree(p.textures[j].RenderId)
				}
				slog.Error(e, "attachmentIndex", i)
				return errors.New(e)
			}
			if a.InitialLayout != 0 {
				success = vr.transitionImageLayout(&p.textures[i].RenderId,
					a.InitialLayout, img.Aspect, img.Access, nil)
			}
			if !success {
				const e = "failed to transition image layout for render pass attachment"
				for j := range i + 1 {
					p.textures[j].RenderId = vr.textureIdFree(p.textures[j].RenderId)
				}
				slog.Error(e, "attachmentIndex", i)
				return errors.New(e)
			}
		}
	}
	attachments := make([]vk.AttachmentDescription, len(r.AttachmentDescriptions))
	for i := range r.AttachmentDescriptions {
		// TODO:  Flags
		attachments[i].Flags = 0
		attachments[i].Format = r.AttachmentDescriptions[i].Format
		attachments[i].Samples = r.AttachmentDescriptions[i].Samples
		attachments[i].LoadOp = r.AttachmentDescriptions[i].LoadOp
		attachments[i].StoreOp = r.AttachmentDescriptions[i].StoreOp
		attachments[i].StencilLoadOp = r.AttachmentDescriptions[i].StencilLoadOp
		attachments[i].StencilStoreOp = r.AttachmentDescriptions[i].StencilStoreOp
		attachments[i].InitialLayout = r.AttachmentDescriptions[i].InitialLayout
		attachments[i].FinalLayout = r.AttachmentDescriptions[i].FinalLayout
	}
	color := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	input := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	preserve := make([][]uint32, len(r.SubpassDescriptions))
	depthStencil := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	resolve := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	for i := range r.SubpassDescriptions {
		sd := &r.SubpassDescriptions[i]
		car := sd.ColorAttachmentReferences
		iar := sd.InputAttachmentReferences
		pa := sd.PreserveAttachments
		dsa := sd.DepthStencilAttachment
		ra := sd.ResolveAttachments
		color[i] = make([]vk.AttachmentReference, len(car))
		input[i] = make([]vk.AttachmentReference, len(iar))
		preserve[i] = make([]uint32, len(pa))
		depthStencil[i] = make([]vk.AttachmentReference, len(dsa))
		resolve[i] = make([]vk.AttachmentReference, len(ra))
		for j := range car {
			color[i][j].Attachment = car[j].Attachment
			color[i][j].Layout = car[j].Layout
		}
		for j := range iar {
			input[i][j].Attachment = iar[j].Attachment
			input[i][j].Layout = iar[j].Layout
		}
		copy(preserve[i], pa)
		for j := range dsa {
			depthStencil[i][j].Attachment = dsa[j].Attachment
			depthStencil[i][j].Layout = dsa[j].Layout
		}
		for j := range ra {
			resolve[i][j].Attachment = ra[j].Attachment
			resolve[i][j].Layout = ra[j].Layout
		}
	}
	subpasses := make([]vk.SubpassDescription, len(r.SubpassDescriptions))
	for i := range r.SubpassDescriptions {
		// TODO:  Fill in the flags
		subpasses[i].Flags = 0
		subpasses[i].PipelineBindPoint = r.SubpassDescriptions[i].PipelineBindPoint
		subpasses[i].ColorAttachmentCount = uint32(len(color[i]))
		subpasses[i].InputAttachmentCount = uint32(len(input[i]))
		subpasses[i].PreserveAttachmentCount = uint32(len(preserve[i]))
		if len(color[i]) > 0 {
			subpasses[i].PColorAttachments = &color[i][0]
		}
		if len(input[i]) > 0 {
			subpasses[i].PInputAttachments = &input[i][0]
		}
		if len(preserve[i]) > 0 {
			subpasses[i].PPreserveAttachments = &preserve[i][0]
		}
		if len(depthStencil[i]) > 0 {
			subpasses[i].PDepthStencilAttachment = &depthStencil[i][0]
		}
		if len(resolve[i]) > 0 {
			subpasses[i].PResolveAttachments = &resolve[i][0]
		}
	}
	selfDependencies := make([]vk.SubpassDependency, len(r.SubpassDependencies))
	for i := range r.SubpassDependencies {
		selfDependencies[i].SrcSubpass = r.SubpassDependencies[i].SrcSubpass
		selfDependencies[i].DstSubpass = r.SubpassDependencies[i].DstSubpass
		selfDependencies[i].SrcStageMask = r.SubpassDependencies[i].SrcStageMask
		selfDependencies[i].DstStageMask = r.SubpassDependencies[i].DstStageMask
		selfDependencies[i].SrcAccessMask = r.SubpassDependencies[i].SrcAccessMask
		selfDependencies[i].DstAccessMask = r.SubpassDependencies[i].DstAccessMask
		selfDependencies[i].DependencyFlags = r.SubpassDependencies[i].DependencyFlags
	}
	info := vk.RenderPassCreateInfo{}
	info.SType = vulkan_const.StructureTypeRenderPassCreateInfo
	info.AttachmentCount = uint32(len(attachments))
	info.PAttachments = &attachments[0]
	info.SubpassCount = uint32(len(subpasses))
	info.PSubpasses = &subpasses[0]
	info.DependencyCount = uint32(len(selfDependencies))
	if len(selfDependencies) > 0 {
		info.PDependencies = &selfDependencies[0]
	}
	var handle vk.RenderPass
	if vk.CreateRenderPass(vr.device, &info, nil, &handle) != vulkan_const.Success {
		return errors.New("failed to create the render pass")
	}
	p.Handle = handle
	vr.dbg.add(vk.TypeToUintPtr(p.Handle))
	for i := range r.Subpass {
		p.setupSubpass(&r.Subpass[i], vr, vr.caches.AssetDatabase(), i+1)
	}
	imageViews := make([]vk.ImageView, 0, len(p.textures))
	for i := range len(r.AttachmentDescriptions) {
		a := &r.AttachmentDescriptions[i]
		if a.Image.IsInvalid() {
			if a.Image.ExistingImage != "" {
				for _, v := range vr.renderPassCache {
					if t, ok := v.findTextureByName(a.Image.ExistingImage); ok {
						imageViews = append(imageViews, t.RenderId.View)
						break
					}
				}
			}
		} else {
			imageViews = append(imageViews, p.textures[i].RenderId.View)
		}
	}
	if len(imageViews) == len(attachments) {
		if err = p.CreateFrameBuffer(vr, imageViews, p.textures[0].Width, p.textures[0].Height); err != nil {
			slog.Error("failed to create the frame buffer for the render pass", "error", err)
			return err
		}
	}
	return nil
}

func (p *RenderPass) CreateFrameBuffer(vr *Vulkan,
	imageViews []vk.ImageView, width, height int) error {

	fb, ok := vr.CreateFrameBuffer(p, imageViews, uint32(width), uint32(height))
	if !ok {
		return errors.New("failed to create the frame buffer for the pass")
	}
	p.Buffer = fb
	return nil
}

func (p *RenderPass) Destroy(vr *Vulkan) {
	if p.Handle == vk.NullRenderPass {
		return
	}
	vk.DestroyRenderPass(p.device, p.Handle, nil)
	p.dbg.remove(vk.TypeToUintPtr(p.Handle))
	p.Handle = vk.NullRenderPass
	vk.DestroyFramebuffer(p.device, p.Buffer, nil)
	p.dbg.remove(vk.TypeToUintPtr(p.Buffer))
	p.Buffer = vk.NullFramebuffer
	for i := range p.textures {
		vr.destroyTextureHandle(p.textures[i].RenderId)
		p.textures[i].RenderId = TextureId{}
	}
	for i := range p.subpasses {
		for j := range len(p.subpasses[i].cmd) {
			p.subpasses[i].cmd[j].Destroy(vr)
		}
	}
	p.subpasses = klib.WipeSlice(p.subpasses)
	for i := range p.cmd {
		p.cmd[i].Destroy(vr)
	}
	for i := range p.cmdSecondary {
		p.cmdSecondary[i].Destroy(vr)
	}
}

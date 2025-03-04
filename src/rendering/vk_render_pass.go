/******************************************************************************/
/* vk_render_pass.go                                                          */
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
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"kaiju/assets"
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
)

type RenderPass struct {
	Handle       vk.RenderPass
	Buffer       vk.Framebuffer
	device       vk.Device
	dbg          *debugVulkan
	textures     []Texture
	construction RenderPassDataCompiled
	subpasses    []RenderPassSubpass
}

type RenderPassSubpass struct {
	shader         *Shader
	shaderPipeline ShaderPipelineDataCompiled
	descriptorSets [maxFramesInFlight]vk.DescriptorSet
	descriptorPool vk.DescriptorPool
	sampledImages  []int
	renderQuad     *Mesh
}

func (r *RenderPass) findTextureByName(name string) (*Texture, bool) {
	for i := range r.textures {
		if r.textures[i].Key == name {
			return &r.textures[i], true
		}
	}
	return nil, false
}

func (r *RenderPass) setupSubpass(c *RenderPassSubpassDataCompiled, vr *Vulkan, assets *assets.Database) error {
	r.subpasses = r.subpasses[:0]
	sp := RenderPassSubpass{}
	// TODO:  This is copied from Material::Compile
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
		sp.shader.renderPass = r
		shaderCache.CreatePending()
	}
	sp.descriptorSets, sp.descriptorPool = klib.MustReturn2(
		vr.createDescriptorSet(sp.shader.RenderId.descriptorSetLayout, 0))
	for i := range c.SampledImages {
		t := &r.textures[c.SampledImages[i]].RenderId
		if t.Sampler == vk.NullSampler {
			vr.createTextureSampler(&t.Sampler, t.MipLevels, vk.FilterLinear)
		}
	}
	sp.sampledImages = append(sp.sampledImages, c.SampledImages...)
	sp.renderQuad = NewMeshUnitQuad(vr.caches.MeshCache())
	vr.caches.MeshCache().CreatePending()
	r.subpasses = append(r.subpasses, sp)
	return nil
}

func (r *RenderPass) SelectOutputAttachment() *Texture {
	for i := range r.construction.AttachmentDescriptions {
		a := &r.construction.AttachmentDescriptions[i]
		if (a.Image.Usage & vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit)) != 0 {
			return &r.textures[i]
		}
	}
	slog.Error("failed to find an output color attachment for the render pass", "renderPass", r.construction.Name)
	if len(r.textures) > 0 {
		return &r.textures[0]
	}
	return nil
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
	{
		w := uint32(vr.swapChainExtent.Width)
		h := uint32(vr.swapChainExtent.Height)
		for i := range len(r.AttachmentDescriptions) {
			a := &r.AttachmentDescriptions[i]
			img := &a.Image
			if a.Image.IsInvalid() {
				continue
			}
			p.textures[i].Width = int(w)
			p.textures[i].Height = int(h)
			success := vr.CreateImage(w, h, img.MipLevels, a.Samples,
				a.Format, img.Tiling, img.Usage,
				img.MemoryProperty, &p.textures[i].RenderId, int(img.LayerCount))
			if !success {
				const e = "failed to create image for render pass attachment"
				slog.Error(e, "attachmentIndex", i)
				return errors.New(e)
			}
			success = vr.createImageView(&p.textures[i].RenderId, img.Aspect)
			if !success {
				const e = "failed to create image view for render pass attachment"
				for j := range i + 1 {
					vr.textureIdFree(&p.textures[j].RenderId)
				}
				slog.Error(e, "attachmentIndex", i)
				return errors.New(e)
			}
			success = vr.createTextureSampler(&p.textures[i].RenderId.Sampler,
				img.MipLevels, img.Filter)
			if !success {
				const e = "failed to create image sampler for render pass attachment"
				for j := range i + 1 {
					vr.textureIdFree(&p.textures[j].RenderId)
				}
				slog.Error(e, "attachmentIndex", i)
				return errors.New(e)
			}
			if vr.commandPool != vk.NullCommandPool {
				success = vr.transitionImageLayout(&p.textures[i].RenderId, a.InitialLayout,
					img.Aspect, img.Access, vk.NullCommandBuffer)
			}
			if !success {
				const e = "failed to transition image layout for render pass attachment"
				for j := range i + 1 {
					vr.textureIdFree(&p.textures[j].RenderId)
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
	info.SType = vk.StructureTypeRenderPassCreateInfo
	info.AttachmentCount = uint32(len(attachments))
	info.PAttachments = &attachments[0]
	info.SubpassCount = uint32(len(subpasses))
	info.PSubpasses = &subpasses[0]
	info.DependencyCount = uint32(len(selfDependencies))
	if len(selfDependencies) > 0 {
		info.PDependencies = &selfDependencies[0]
	}
	var handle vk.RenderPass
	if vk.CreateRenderPass(vr.device, &info, nil, &handle) != vk.Success {
		return errors.New("failed to create the render pass")
	}
	p.Handle = handle
	vr.dbg.add(vk.TypeToUintPtr(p.Handle))
	for i := range r.Subpass {
		p.setupSubpass(&r.Subpass[i], vr, vr.caches.AssetDatabase())
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
		err := p.CreateFrameBuffer(vr, imageViews,
			p.textures[0].Width, p.textures[0].Height)
		if err != nil {
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
		vr.textureIdFree(&p.textures[i].RenderId)
	}
	for i := range p.textures {
		vr.DestroyTexture(&p.textures[i])
	}
}

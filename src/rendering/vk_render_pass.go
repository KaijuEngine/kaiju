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

func (r *RenderPass) setupSubpass(c *RenderPassSubpassDataCompiled, vr *Vulkan, assets *assets.Database) error {
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

func NewRenderPass(vr *Vulkan, assets *assets.Database, attachments []vk.AttachmentDescription, subPasses []vk.SubpassDescription, dependencies []vk.SubpassDependency, textures []Texture, setup *RenderPassDataCompiled) (*RenderPass, error) {
	p := &RenderPass{
		device:       vr.device,
		dbg:          &vr.dbg,
		textures:     textures,
		construction: *setup,
	}
	for i := range textures {
		p.textures[i].Key = fmt.Sprintf("renderPass-%s-%d", setup.Name, i)
	}
	info := vk.RenderPassCreateInfo{}
	info.SType = vk.StructureTypeRenderPassCreateInfo
	info.AttachmentCount = uint32(len(attachments))
	info.PAttachments = &attachments[0]
	info.SubpassCount = uint32(len(subPasses))
	info.PSubpasses = &subPasses[0]
	info.DependencyCount = uint32(len(dependencies))
	if len(dependencies) > 0 {
		info.PDependencies = &dependencies[0]
	}
	var handle vk.RenderPass
	if vk.CreateRenderPass(vr.device, &info, nil, &handle) != vk.Success {
		return p, errors.New("failed to create the render pass")
	}
	p.Handle = handle
	vr.dbg.add(vk.TypeToUintPtr(p.Handle))
	for i := range setup.Subpass {
		p.setupSubpass(&setup.Subpass[i], vr, assets)
	}
	return p, nil
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
	vk.DestroyFramebuffer(p.device, p.Buffer, nil)
	p.dbg.remove(vk.TypeToUintPtr(p.Buffer))
	for i := range p.textures {
		vr.textureIdFree(&p.textures[i].RenderId)
	}
	for i := range p.textures {
		vr.DestroyTexture(&p.textures[i])
	}
	p.Handle = vk.NullRenderPass
}

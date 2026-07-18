/******************************************************************************/
/* vk_render_pass.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unsafe"
	"weak"

	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

type RenderPass struct {
	Handle       vk.RenderPass
	Buffer       GPUFrameBuffer
	textures     []Texture
	construction RenderPassDataCompiled
	subpasses    []RenderPassSubpass
	cmd          [maxFramesInFlight]CommandRecorder
	cmdSecondary [maxFramesInFlight]CommandRecorderSecondary
	viewCmds     [maxFramesInFlight][]RenderPassCommandSet
	viewCmdCount [maxFramesInFlight]int
	activeCmds   *RenderPassCommandSet
	currentIdx   int
	subpassIdx   int
	frame        int
}

const (
	combineRenderPassName   = "combine"
	swapChainRenderPassName = "swapchain"
)

type RenderPassSubpass struct {
	shader         *Shader
	shaderPipeline ShaderPipelineDataCompiled
	descriptorSets [maxFramesInFlight]GPUDescriptorSet
	descriptorPool GPUDescriptorPool
	sampledImages  []int
	renderQuad     *Mesh
	cmd            [maxFramesInFlight]CommandRecorderSecondary
}

type RenderPassCommandSet struct {
	cmd          CommandRecorder
	cmdSecondary CommandRecorderSecondary
	subpassCmds  []CommandRecorderSecondary
}

func (r *RenderPass) Width() int  { return r.construction.Width }
func (r *RenderPass) Height() int { return r.construction.Height }

func (r *RenderPass) Texture(index int) *Texture { return &r.textures[index] }

func (r *RenderPass) usesSwapChainClearColor() bool {
	return r != nil && (r.construction.Name == combineRenderPassName ||
		r.construction.Name == swapChainRenderPassName)
}

func (r *RenderPass) setClearColor(color matrix.Color) bool {
	if r == nil {
		return false
	}
	return r.construction.setClearColor(color)
}

func (r *RenderPass) IsShadowPass() bool {
	// TODO:  Need another way to denote this is a shadow pass
	return strings.HasPrefix(r.construction.Name, "light_offscreen")
}

func shadowCascadeIndex(renderPassName string) (int, bool) {
	switch renderPassName {
	case "light_offscreen":
		return 0, true
	case "light_offscreen_csm1":
		return 1, true
	case "light_offscreen_csm2":
		return 2, true
	default:
		return 0, false
	}
}

func (r *RenderPass) ExecuteSecondaryCommands() {
	buffs := [1]vk.CommandBuffer{}
	rec := r.activeSecondaryCommand()
	if r.currentIdx > 0 {
		rec = r.activeSubpassCommand(r.currentIdx - 1)
	}
	rec.End()
	buffs[0] = rec.buffer
	vk.CmdExecuteCommands(r.activePrimaryCommand().buffer, uint32(len(buffs)), &buffs[0])
}

func (r *RenderPass) SelectOutputAttachment(device *GPUDevice) *Texture {
	sc := &device.LogicalDevice.SwapChain
	targetFormat := sc.Images[0].Format
	var fallback *Texture
	for i := range r.construction.AttachmentDescriptions {
		a := &r.construction.AttachmentDescriptions[i]
		if (a.Image.Usage & GPUImageUsageColorAttachmentBit) != 0 {
			if fallback == nil {
				// First image is likely the better image to fall back to
				fallback = &r.textures[i]
			}
			// A render pass may contain another attachment whose format happens
			// to match the swap chain (for example an RGBA8 albedo G-buffer next
			// to an FP16 HDR scene color). The semantic .color output must win
			// over that coincidental format match.
			if strings.HasSuffix(a.Image.Name, ".color") {
				return &r.textures[i]
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
				for _, p := range device.LogicalDevice.renderPassCache {
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
		if !isDepthFormat(r.textures[i].RenderId.Format.toVulkan()) {
			slog.Error("failed to find an output color attachment for the render pass, using fallback", "renderPass", r.construction.Name)
			return &r.textures[i]
		}
	}
	return nil
}

func (r *RenderPass) SelectOutputAttachmentWithSuffix(suffix string) (*Texture, bool) {
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

func (r *RenderPass) setupSubpass(c *RenderPassSubpassDataCompiled, device *GPUDevice, index int) error {
	r.subpasses = klib.WipeSlice(r.subpasses)
	sp := RenderPassSubpass{}
	assets := device.Painter.caches.AssetDatabase()
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
		// A graphics pipeline is created against a specific render pass and
		// subpass. Reusing a cached subpass shader by asset name is invalid when
		// two passes use different attachment formats (for example HDR world
		// compositing and display-referred UI compositing).
		rawSD.Name = fmt.Sprintf("%s@%s:%d", rawSD.Name, r.construction.Name, index)
		sp.shaderPipeline = pipe.Compile(&device.PhysicalDevice)
		shaderCache := device.Painter.caches.ShaderCache()
		sp.shader, _ = shaderCache.Shader(rawSD.Compile())
		sp.shader.pipelineInfo = &sp.shaderPipeline
		sp.shader.renderPass = weak.Make(r)
		shaderCache.CreatePending()
	}
	sp.descriptorSets, sp.descriptorPool = klib.MustReturn2(
		device.createDescriptorSet(sp.shader.RenderId.descriptorSetLayout, 0))
	var err error
	for i := range c.SampledImages {
		t := &r.textures[c.SampledImages[i]].RenderId
		if !t.Sampler.IsValid() {
			t.Sampler, err = device.CreateTextureSampler(t.MipLevels, GPUFilterLinear)
			if err != nil {
				return err
			}
		}
	}
	sp.sampledImages = append(sp.sampledImages, c.SampledImages...)
	sp.renderQuad = NewMeshUnitQuad(device.Painter.caches.MeshCache())
	device.Painter.caches.MeshCache().ProcessPending()
	for i := range len(sp.cmd) {
		if sp.cmd[i], err = NewCommandRecorderSecondary(device, r, index); err != nil {
			return err
		}
	}
	r.subpasses = append(r.subpasses, sp)
	return nil
}

func (r *RenderPass) endSubpasses() {
	cmd := r.activePrimaryCommand()
	vk.CmdEndRenderPass(cmd.buffer)
	cmd.End()
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
			Framebuffer: vk.Framebuffer(r.Buffer.handle),
			RenderArea: vk.Rect2D{
				Offset: vk.Offset2D{X: 0, Y: 0},
				Extent: extent,
			},
			ClearValueCount: uint32(len(clearColors)),
		}
		if len(clearColors) > 0 {
			renderPassInfo.PClearValues = &clearColors[0]
		}
		cmd := r.activePrimaryCommand()
		cmd.Begin()
		vk.CmdBeginRenderPass(cmd.buffer, &renderPassInfo, vulkan_const.SubpassContentsSecondaryCommandBuffers)
		r.activeSecondaryCommand().Begin(viewport, scissor)
	} else {
		spCmd := r.activeSubpassCommand(r.subpassIdx - 1)
		spCmd.Reset()
		vk.CmdNextSubpass(r.activePrimaryCommand().buffer, vulkan_const.SubpassContentsSecondaryCommandBuffers)
		spCmd.Begin(viewport, scissor)
	}
	r.currentIdx = r.subpassIdx
	r.subpassIdx++
	if r.subpassIdx > len(r.subpasses) {
		r.subpassIdx = 0
	}
}

func (r *RenderPass) useViewCommandSet(device *GPUDevice, frame int) error {
	idx := r.viewCmdCount[frame]
	if idx >= len(r.viewCmds[frame]) {
		set, err := r.createViewCommandSet(device)
		if err != nil {
			return err
		}
		r.viewCmds[frame] = append(r.viewCmds[frame], set)
	} else {
		r.viewCmds[frame][idx].reset()
	}
	r.viewCmdCount[frame]++
	r.activeCmds = &r.viewCmds[frame][idx]
	return nil
}

func (r *RenderPass) useDefaultCommandSet() {
	r.activeCmds = nil
}

func (r *RenderPass) resetViewCommandSets(frame int) {
	if frame >= 0 && frame < len(r.viewCmdCount) {
		r.viewCmdCount[frame] = 0
	}
}

func (r *RenderPass) createViewCommandSet(device *GPUDevice) (RenderPassCommandSet, error) {
	var set RenderPassCommandSet
	var err error
	if set.cmd, err = NewCommandRecorder(device); err != nil {
		return set, err
	}
	if set.cmdSecondary, err = NewCommandRecorderSecondary(device, r, 0); err != nil {
		set.cmd.Destroy(device)
		return RenderPassCommandSet{}, err
	}
	set.subpassCmds = make([]CommandRecorderSecondary, len(r.subpasses))
	for i := range r.subpasses {
		if set.subpassCmds[i], err = NewCommandRecorderSecondary(device, r, i+1); err != nil {
			set.destroy(device)
			return RenderPassCommandSet{}, err
		}
	}
	return set, nil
}

func (r *RenderPass) activePrimaryCommand() *CommandRecorder {
	if r.activeCmds != nil {
		return &r.activeCmds.cmd
	}
	return &r.cmd[r.frame]
}

func (r *RenderPass) activeSecondaryCommand() *CommandRecorderSecondary {
	if r.activeCmds != nil {
		return &r.activeCmds.cmdSecondary
	}
	return &r.cmdSecondary[r.frame]
}

func (r *RenderPass) activeSubpassCommand(index int) *CommandRecorderSecondary {
	if r.activeCmds != nil {
		return &r.activeCmds.subpassCmds[index]
	}
	return &r.subpasses[index].cmd[r.frame]
}

func (s *RenderPassCommandSet) reset() {
	s.cmd.Reset()
	s.cmdSecondary.Reset()
	for i := range s.subpassCmds {
		s.subpassCmds[i].Reset()
	}
}

func (s *RenderPassCommandSet) destroy(device *GPUDevice) {
	if s.cmd.buffer != vk.NullCommandBuffer {
		s.cmd.Destroy(device)
	}
	if s.cmdSecondary.buffer != vk.NullCommandBuffer {
		s.cmdSecondary.Destroy(device)
	}
	for i := range s.subpassCmds {
		if s.subpassCmds[i].buffer != vk.NullCommandBuffer {
			s.subpassCmds[i].Destroy(device)
		}
	}
	s.subpassCmds = nil
}

func isDepthFormat(format vulkan_const.Format) bool {
	switch format {
	case vulkan_const.FormatD16Unorm, vulkan_const.FormatD32Sfloat, vulkan_const.FormatD16UnormS8Uint,
		vulkan_const.FormatD24UnormS8Uint, vulkan_const.FormatD32SfloatS8Uint:
		return true
	}
	return false
}

func NewRenderPass(device *GPUDevice, setup *RenderPassDataCompiled) (*RenderPass, error) {
	p := &RenderPass{
		construction: *setup,
		textures:     make([]Texture, 0, len(setup.AttachmentDescriptions)),
	}
	p.construction.ImageClears = append([]RenderPassAttachmentImageClear(nil), setup.ImageClears...)
	if device != nil && p.usesSwapChainClearColor() {
		if color, ok := device.LogicalDevice.SwapChain.ClearColorOverride(); ok {
			p.setClearColor(color)
		}
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
	if err := p.Recontstruct(device); err != nil {
		p.Destroy(device)
		return p, err
	}
	return p, nil
}

func (p *RenderPass) Recontstruct(device *GPUDevice) error {
	p.Destroy(device)
	r := &p.construction
	var err error
	for i := range len(p.cmd) {
		if p.cmd[i], err = NewCommandRecorder(device); err != nil {
			return err
		}
	}
	for i := range len(p.cmdSecondary) {
		if p.cmdSecondary[i], err = NewCommandRecorderSecondary(device, p, 0); err != nil {
			return err
		}
	}
	{
		sc := &device.LogicalDevice.SwapChain
		w := sc.Extent.Width()
		h := sc.Extent.Height()
		if r.Width > 0 {
			w = int32(r.Width)
		}
		if r.Height > 0 {
			h = int32(r.Height)
		}
		for i := range len(r.AttachmentDescriptions) {
			a := &r.AttachmentDescriptions[i]
			img := &a.Image
			if a.Image.IsInvalid() {
				continue
			}
			p.textures[i].Width = int(w)
			p.textures[i].Height = int(h)
			err := device.CreateImage(&p.textures[i].RenderId, img.MemoryProperty,
				GPUImageCreateRequest{
					ImageType:   GPUImageType2d,
					Extent:      matrix.Vec3i{int32(w), int32(h), 1},
					MipLevels:   img.MipLevels,
					ArrayLayers: img.LayerCount,
					Format:      a.Format,
					Tiling:      img.Tiling,
					Usage:       img.Usage,
					Samples:     a.Samples,
				})
			if err != nil {
				slog.Error("failed to create image for render pass attachment", "attachmentIndex", i)
				return err
			}
			err = device.LogicalDevice.CreateImageView(&p.textures[i].RenderId, img.Aspect, GPUImageViewType2d)
			if err != nil {
				const e = "failed to create image view for render pass attachment"
				for j := range i + 1 {
					device.LogicalDevice.FreeTexture(&p.textures[j].RenderId)
				}
				slog.Error(e, "attachmentIndex", i)
				return errors.New(e)
			}
			p.textures[i].RenderId.Sampler, err = device.CreateTextureSampler(img.MipLevels, img.Filter)
			if err != nil {
				for j := range i + 1 {
					device.LogicalDevice.FreeTexture(&p.textures[j].RenderId)
				}
				slog.Error("failed to create image sampler for render pass attachment", "attachmentIndex", i)
				return err
			}
			if a.InitialLayout != 0 {
				device.TransitionImageLayout(&p.textures[i].RenderId,
					a.InitialLayout, img.Aspect, img.Access, nil)
			}
		}
	}
	attachments := make([]vk.AttachmentDescription, len(r.AttachmentDescriptions))
	for i := range r.AttachmentDescriptions {
		// TODO:  Flags
		attachments[i].Flags = 0
		attachments[i].Format = r.AttachmentDescriptions[i].Format.toVulkan()
		attachments[i].Samples = vulkan_const.SampleCountFlagBits(r.AttachmentDescriptions[i].Samples.toVulkan())
		attachments[i].LoadOp = r.AttachmentDescriptions[i].LoadOp.toVulkan()
		attachments[i].StoreOp = r.AttachmentDescriptions[i].StoreOp.toVulkan()
		attachments[i].StencilLoadOp = r.AttachmentDescriptions[i].StencilLoadOp.toVulkan()
		attachments[i].StencilStoreOp = r.AttachmentDescriptions[i].StencilStoreOp.toVulkan()
		attachments[i].InitialLayout = r.AttachmentDescriptions[i].InitialLayout.toVulkan()
		attachments[i].FinalLayout = r.AttachmentDescriptions[i].FinalLayout.toVulkan()
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
			color[i][j].Layout = car[j].Layout.toVulkan()
		}
		for j := range iar {
			input[i][j].Attachment = iar[j].Attachment
			input[i][j].Layout = iar[j].Layout.toVulkan()
		}
		copy(preserve[i], pa)
		for j := range dsa {
			depthStencil[i][j].Attachment = dsa[j].Attachment
			depthStencil[i][j].Layout = dsa[j].Layout.toVulkan()
		}
		for j := range ra {
			resolve[i][j].Attachment = ra[j].Attachment
			resolve[i][j].Layout = ra[j].Layout.toVulkan()
		}
	}
	subpasses := make([]vk.SubpassDescription, len(r.SubpassDescriptions))
	for i := range r.SubpassDescriptions {
		// TODO:  Fill in the flags
		subpasses[i].Flags = 0
		subpasses[i].PipelineBindPoint = r.SubpassDescriptions[i].PipelineBindPoint.toVulkan()
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
		selfDependencies[i].SrcStageMask = r.SubpassDependencies[i].SrcStageMask.toVulkan()
		selfDependencies[i].DstStageMask = r.SubpassDependencies[i].DstStageMask.toVulkan()
		selfDependencies[i].SrcAccessMask = r.SubpassDependencies[i].SrcAccessMask.toVulkan()
		selfDependencies[i].DstAccessMask = r.SubpassDependencies[i].DstAccessMask.toVulkan()
		selfDependencies[i].DependencyFlags = r.SubpassDependencies[i].DependencyFlags.toVulkan()
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
	if vk.CreateRenderPass(vk.Device(device.LogicalDevice.handle), &info, nil, &handle) != vulkan_const.Success {
		return errors.New("failed to create the render pass")
	}
	p.Handle = handle
	device.LogicalDevice.dbg.track(unsafe.Pointer(p.Handle))
	for i := range r.Subpass {
		if err := p.setupSubpass(&r.Subpass[i], device, i+1); err != nil {
			return err
		}
	}
	imageViews := make([]GPUImageView, 0, len(p.textures))
	missingExistingImage := false
	for i := range len(r.AttachmentDescriptions) {
		a := &r.AttachmentDescriptions[i]
		if a.Image.IsInvalid() {
			if a.Image.ExistingImage != "" {
				found := false
				for _, v := range device.LogicalDevice.renderPassCache {
					if t, ok := v.findTextureByName(a.Image.ExistingImage); ok {
						imageViews = append(imageViews, t.RenderId.View)
						found = true
						break
					}
				}
				if !found {
					missingExistingImage = true
				}
			}
		} else {
			imageViews = append(imageViews, p.textures[i].RenderId.View)
		}
	}
	if len(imageViews) == len(attachments) {
		p.Buffer, err = device.CreateFrameBuffer(p, imageViews, int32(p.textures[0].Width), int32(p.textures[0].Height))
		if err != nil {
			slog.Error("failed to create the frame buffer for the render pass", "error", err)
			return err
		}
	} else if r.Name != swapChainRenderPassName {
		if missingExistingImage {
			return nil
		}
		return fmt.Errorf("render pass %q framebuffer has %d image views for %d attachments", r.Name, len(imageViews), len(attachments))
	}
	return nil
}

func (p *RenderPass) Destroy(device *GPUDevice) {
	if p.Handle == vk.NullRenderPass {
		return
	}
	vk.DestroyRenderPass(vk.Device(device.LogicalDevice.handle), p.Handle, nil)
	device.LogicalDevice.dbg.remove(unsafe.Pointer(p.Handle))
	p.Handle = vk.NullRenderPass
	device.DestroyFrameBuffer(p.Buffer)
	device.LogicalDevice.dbg.remove(p.Buffer.handle)
	p.Buffer.Reset()
	for i := range p.textures {
		device.LogicalDevice.FreeTexture(&p.textures[i].RenderId)
		p.textures[i].RenderId = TextureId{}
	}
	for i := range p.subpasses {
		sp := &p.subpasses[i]
		// setupSubpass allocates a descriptor set + pool for every render-pass
		// (re)construction. Without freeing them here, each swap-chain remake
		// (which reconstructs every render pass) leaks one descriptor set per
		// subpass — the combine/composite subpass in particular — growing the
		// descriptor pool's backing argument buffer until a draw binds past its
		// end and the Metal driver faults (the climbing spvDescriptorSet0 offset).
		if sp.descriptorPool.IsValid() || len(validDescriptorSets(sp.descriptorSets)) > 0 {
			device.LogicalDevice.bufferTrash.Add(bufferTrash{
				delay: maxFramesInFlight,
				pool:  sp.descriptorPool,
				sets:  sp.descriptorSets,
			})
		}
		sp.descriptorPool = GPUDescriptorPool{}
		sp.descriptorSets = [maxFramesInFlight]GPUDescriptorSet{}
		for j := range len(sp.cmd) {
			sp.cmd[j].Destroy(device)
		}
	}
	for frame := range p.viewCmds {
		for i := range p.viewCmds[frame] {
			p.viewCmds[frame][i].destroy(device)
		}
		p.viewCmds[frame] = nil
		p.viewCmdCount[frame] = 0
	}
	p.activeCmds = nil
	p.subpasses = klib.WipeSlice(p.subpasses)
	for i := range p.cmd {
		p.cmd[i].Destroy(device)
	}
	for i := range p.cmdSecondary {
		p.cmdSecondary[i].Destroy(device)
	}
}

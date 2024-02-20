/******************************************************************************/
/* vk_render_pass.go                                                          */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"errors"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

type RenderPass struct {
	Handle vk.RenderPass
	Buffer vk.Framebuffer
	device vk.Device
	dbg    *debugVulkan
}

func NewRenderPass(device vk.Device, dbg *debugVulkan, attachments []vk.AttachmentDescription, subPasses []vk.SubpassDescription, dependencies []vk.SubpassDependency) (RenderPass, error) {
	p := RenderPass{
		device: device,
		dbg:    dbg,
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
	if vk.CreateRenderPass(device, &info, nil, &p.Handle) != vk.Success {
		return p, errors.New("failed to create the render pass")
	}
	dbg.add(uintptr(unsafe.Pointer(p.Handle)))
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

func (p *RenderPass) Destroy() {
	vk.DestroyRenderPass(p.device, p.Handle, nil)
	p.dbg.remove(uintptr(unsafe.Pointer(p.Handle)))
	vk.DestroyFramebuffer(p.device, p.Buffer, nil)
	p.dbg.remove(uintptr(unsafe.Pointer(p.Buffer)))
}

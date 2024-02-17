package rendering

import (
	"errors"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

type RenderPass struct {
	Handle vk.RenderPass
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
	info.PDependencies = &dependencies[0]
	if vk.CreateRenderPass(device, &info, nil, &p.Handle) != vk.Success {
		return p, errors.New("failed to create the render pass")
	}
	dbg.add(uintptr(unsafe.Pointer(p.Handle)))
	return p, nil
}

func (p *RenderPass) Destroy() {
	vk.DestroyRenderPass(p.device, p.Handle, nil)
	p.dbg.remove(uintptr(unsafe.Pointer(p.Handle)))
}

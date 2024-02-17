package rendering

import (
	"log/slog"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) findSupportedFormat(candidates []vk.Format, tiling vk.ImageTiling, features vk.FormatFeatureFlags) vk.Format {
	for i := 0; i < len(candidates); i++ {
		var props vk.FormatProperties
		format := candidates[i]
		vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &props)
		if tiling == vk.ImageTilingLinear && (props.LinearTilingFeatures&features) == features {
			return format
		} else if tiling == vk.ImageTilingOptimal && (props.OptimalTilingFeatures&features) == features {
			return format
		}
	}
	slog.Error("Failed to find supported format")
	// TODO:  Return an error too
	return candidates[0]
}

func (vr *Vulkan) findDepthFormat() vk.Format {
	candidates := []vk.Format{vk.FormatX8D24UnormPack32,
		vk.FormatD24UnormS8Uint, vk.FormatD32Sfloat,
		vk.FormatD32SfloatS8Uint, vk.FormatD16Unorm,
		vk.FormatD16UnormS8Uint,
	}
	return vr.findSupportedFormat(candidates, vk.ImageTilingOptimal, vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit))
}

func (vr *Vulkan) findDepthStencilFormat() vk.Format {
	candidates := []vk.Format{vk.FormatD24UnormS8Uint,
		vk.FormatD32SfloatS8Uint, vk.FormatD16UnormS8Uint,
	}
	return vr.findSupportedFormat(candidates, vk.ImageTilingOptimal, vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit))
}

func (vr *Vulkan) createDepthResources() bool {
	depthFormat := vr.findDepthFormat()
	vr.CreateImage(vr.swapChainExtent.Width, vr.swapChainExtent.Height,
		1, vr.msaaSamples, depthFormat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &vr.depth, 1)
	return vr.createImageView(&vr.depth, vk.ImageAspectFlags(vk.ImageAspectDepthBit))
}

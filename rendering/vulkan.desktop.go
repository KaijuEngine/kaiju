//go:build !OPENGL && (windows || darwin || linux)

package rendering

import vk "github.com/BrentFarris/go-vulkan"

func preTransform(scs vkSwapChainSupportDetails) vk.SurfaceTransformFlagBits {
	return scs.capabilities.CurrentTransform
}

const compositeAlpha = vk.CompositeAlphaOpaqueBit

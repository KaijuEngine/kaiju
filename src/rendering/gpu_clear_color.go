/******************************************************************************/
/* gpu_clear_color.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "kaijuengine.com/matrix"

// SetSwapChainClearColor sets the clear color used by final swap-chain
// presentation passes, including the cached combine pass when it exists.
func (g *GPUDevice) SetSwapChainClearColor(color matrix.Color) {
	if g == nil {
		return
	}
	g.LogicalDevice.SwapChain.SetClearColor(color)
	g.applySwapChainClearColorToCachedPasses(color)
}

// SwapChainClearColor returns the configured swap-chain clear color or the
// engine default when no override has been set.
func (g *GPUDevice) SwapChainClearColor() matrix.Color {
	if g == nil {
		return DefaultSwapChainClearColor()
	}
	return g.LogicalDevice.SwapChain.ClearColor()
}

func (g *GPUDevice) applySwapChainClearColorToCachedPasses(color matrix.Color) {
	for _, pass := range g.LogicalDevice.renderPassCache {
		if pass.usesSwapChainClearColor() {
			pass.setClearColor(color)
		}
	}
}

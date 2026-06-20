/******************************************************************************/
/* gpu_config.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

const (
	GPUWholeSize = (^uintptr(0))
)

func depthFormatCandidates() []GPUFormat {
	return []GPUFormat{GPUFormatX8D24UnormPack32,
		GPUFormatD24UnormS8Uint, GPUFormatD32Sfloat,
		GPUFormatD32SfloatS8Uint, GPUFormatD16Unorm,
		GPUFormatD16UnormS8Uint,
	}
}

func depthStencilFormatCandidates() []GPUFormat {
	return []GPUFormat{GPUFormatD24UnormS8Uint,
		GPUFormatD32SfloatS8Uint, GPUFormatD16UnormS8Uint,
	}
}

// formatHasStencil reports whether the format carries a stencil component, i.e.
// it is one of the combined depth/stencil formats. Such images must always have
// their stencil aspect transitioned alongside depth.
func formatHasStencil(f GPUFormat) bool {
	switch f {
	case GPUFormatD24UnormS8Uint, GPUFormatD32SfloatS8Uint, GPUFormatD16UnormS8Uint:
		return true
	default:
		return false
	}
}

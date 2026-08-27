/******************************************************************************/
/* gpu_config.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "kaijuengine.com/rendering/gpu_types"

const (
	GPUWholeSize = (^uintptr(0))
)

func depthFormatCandidates() []gpu_types.Format {
	return []gpu_types.Format{gpu_types.FormatX8D24UnormPack32,
		gpu_types.FormatD24UnormS8Uint, gpu_types.FormatD32Sfloat,
		gpu_types.FormatD32SfloatS8Uint, gpu_types.FormatD16Unorm,
		gpu_types.FormatD16UnormS8Uint,
	}
}

func depthStencilFormatCandidates() []gpu_types.Format {
	return []gpu_types.Format{gpu_types.FormatD24UnormS8Uint,
		gpu_types.FormatD32SfloatS8Uint, gpu_types.FormatD16UnormS8Uint,
	}
}

// formatHasStencil reports whether the format carries a stencil component, i.e.
// it is one of the combined depth/stencil formats. Such images must always have
// their stencil aspect transitioned alongside depth.
func formatHasStencil(f gpu_types.Format) bool {
	switch f {
	case gpu_types.FormatD24UnormS8Uint, gpu_types.FormatD32SfloatS8Uint, gpu_types.FormatD16UnormS8Uint:
		return true
	default:
		return false
	}
}

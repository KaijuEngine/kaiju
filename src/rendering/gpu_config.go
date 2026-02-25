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

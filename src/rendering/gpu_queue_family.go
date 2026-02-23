package rendering

import "kaiju/matrix"

type GPUQueueFamily struct {
	Index                       int
	MinImageTransferGranularity matrix.Vec3i
	IsGraphics                  bool
	IsCompute                   bool
	IsTransfer                  bool
	IsSparseBinding             bool
	IsProtected                 bool
	HasPresentSupport           bool
}

func (g GPUQueueFamily) IsValid() bool { return g.Index >= 0 }

func InvalidGPUQueueFamily() GPUQueueFamily {
	return GPUQueueFamily{Index: -1}
}

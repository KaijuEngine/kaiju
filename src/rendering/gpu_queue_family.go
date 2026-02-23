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
	PresentSupport              bool
}

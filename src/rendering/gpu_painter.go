package rendering

import (
	"kaiju/rendering/vulkan_const"
)

type GPUPainter struct {
	caches                RenderCaches
	graphicsQueue         GPUQueue
	presentQueue          GPUQueue
	imageIndex            [maxFramesInFlight]uint32
	descriptorPools       []GPUDescriptorPool
	acquireImageResult    vulkan_const.Result
	currentFrame          int
	msaaSamples           vulkan_const.SampleCountFlagBits
	combinedDrawings      Drawings
	combinedDrawingCuller combinedDrawingCuller
	preRuns               []func()
	writtenCommands       []CommandRecorder
	combineCmds           [maxFramesInFlight]CommandRecorder
	blitCmds              [maxFramesInFlight]CommandRecorder
	fallbackShadowMap     *Texture
	fallbackCubeShadowMap *Texture
	computeTasks          []ComputeTask
	computeQueue          GPUQueue
}

func (g *GPUPainter) forceQueueCommand(cmd CommandRecorder, isPrePass bool) {
	if isPrePass {
		cmd.stage = 0
	} else {
		cmd.stage = 1
	}
	g.writtenCommands = append(g.writtenCommands, cmd)
}

package rendering

import (
	"kaiju/engine/collision"
	"kaiju/platform/profiler/tracing"
)

type GPUPainter struct {
	caches                RenderCaches
	imageIndex            [maxFramesInFlight]uint32
	descriptorPools       []GPUDescriptorPool
	acquireImageResult    GPUResult
	currentFrame          int
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

type combinedDrawingCuller struct{}

func (combinedDrawingCuller) IsInView(collision.AABB) bool { return true }
func (combinedDrawingCuller) ViewChanged() bool            { return true }

type ComputeTask struct {
	Shader         *Shader
	DescriptorSets []GPUDescriptorSet
	WorkGroups     [3]uint32
}

func (g *GPUPainter) forceQueueCommand(cmd CommandRecorder, isPrePass bool) {
	if isPrePass {
		cmd.stage = 0
	} else {
		cmd.stage = 1
	}
	g.writtenCommands = append(g.writtenCommands, cmd)
}

func (g *GPUPainter) DestroyDescriptorPools(device *GPUDevice) {
	defer tracing.NewRegion("GPUPainter.DestroyDescriptorPools").End()
	g.destroyDescriptorPoolsImpl(device)
}

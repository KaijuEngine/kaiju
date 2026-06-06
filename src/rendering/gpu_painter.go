/******************************************************************************/
/* gpu_painter.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/platform/profiler/tracing"
)

type GPUPainter struct {
	caches                RenderCaches
	imageIndex            [maxFramesInFlight]uint32
	descriptorPools       []GPUDescriptorPool
	currentFrame          int
	combinedTargets       combinedTargetDrawCache
	combinedDrawingCuller combinedDrawingCuller
	preRuns               []func()
	writtenCommands       []CommandRecorder
	combineCmds           [maxFramesInFlight]CommandRecorder
	blitCmds              [maxFramesInFlight]CommandRecorder
	targetCombineCmds     [maxFramesInFlight][]CommandRecorder
	targetCombineCmdCount [maxFramesInFlight]int
	targetBlitCmds        [maxFramesInFlight][]CommandRecorder
	targetBlitCmdCount    [maxFramesInFlight]int
	fallbackShadowMap     *Texture
	fallbackCubeShadowMap *Texture
	computeTasks          []ComputeTask
	computeQueue          GPUQueue
}

type combinedDrawingCuller struct{}

func (combinedDrawingCuller) IsInView(graviton.AABB) bool { return true }
func (combinedDrawingCuller) ViewChanged() bool           { return true }

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

func (g *GPUPainter) resetTargetBlitCommands(frame int) {
	if frame >= 0 && frame < len(g.targetBlitCmdCount) {
		g.targetCombineCmdCount[frame] = 0
		g.targetBlitCmdCount[frame] = 0
	}
}

func (g *GPUPainter) nextTargetCombineCommand(device *GPUDevice) (*CommandRecorder, error) {
	defer tracing.NewRegion("GPUPainter.nextTargetCombineCommand").End()
	return g.nextTargetCommand(device, &g.targetCombineCmds, &g.targetCombineCmdCount)
}

func (g *GPUPainter) nextTargetBlitCommand(device *GPUDevice) (*CommandRecorder, error) {
	defer tracing.NewRegion("GPUPainter.nextTargetBlitCommand").End()
	return g.nextTargetCommand(device, &g.targetBlitCmds, &g.targetBlitCmdCount)
}

func (g *GPUPainter) nextTargetCommand(device *GPUDevice, commands *[maxFramesInFlight][]CommandRecorder, counts *[maxFramesInFlight]int) (*CommandRecorder, error) {
	frame := device.Painter.currentFrame
	idx := counts[frame]
	if idx >= len(commands[frame]) {
		cmd, err := NewCommandRecorder(device)
		if err != nil {
			return nil, err
		}
		commands[frame] = append(commands[frame], cmd)
	} else {
		commands[frame][idx].Reset()
	}
	counts[frame]++
	cmd := &commands[frame][idx]
	cmd.Begin()
	return cmd, nil
}

func (g *GPUPainter) destroyTargetBlitCommands(device *GPUDevice) {
	defer tracing.NewRegion("GPUPainter.destroyTargetBlitCommands").End()
	g.destroyTargetCommands(device, &g.targetCombineCmds, &g.targetCombineCmdCount)
	g.destroyTargetCommands(device, &g.targetBlitCmds, &g.targetBlitCmdCount)
}

func (g *GPUPainter) destroyTargetCommands(device *GPUDevice, commands *[maxFramesInFlight][]CommandRecorder, counts *[maxFramesInFlight]int) {
	for frame := 0; frame < len(commands); frame++ {
		for i := range commands[frame] {
			commands[frame][i].Destroy(device)
		}
		commands[frame] = nil
		counts[frame] = 0
	}
}

func (g *GPUPainter) DestroyDescriptorPools(device *GPUDevice) {
	defer tracing.NewRegion("GPUPainter.DestroyDescriptorPools").End()
	g.destroyDescriptorPoolsImpl(device)
}

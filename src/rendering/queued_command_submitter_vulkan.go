/******************************************************************************/
/* queued_command_submitter_vulkan.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"
	"math"

	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

const queuedCommandStageCount = 2

type QueuedCommandSubmitter struct {
	device *GPUDevice
}

func (g *GPUDevice) queuedCommandSubmitter() QueuedCommandSubmitter {
	return QueuedCommandSubmitter{device: g}
}

func (s QueuedCommandSubmitter) QueuedCommandCount() int {
	if s.device == nil {
		return 0
	}
	return len(s.device.Painter.writtenCommands)
}

func (s QueuedCommandSubmitter) SubmitForPresent(waitSemaphores []vk.Semaphore, waitStages []vk.PipelineStageFlags, signalSemaphores []vk.Semaphore, fence vk.Fence) bool {
	defer tracing.NewRegion("QueuedCommandSubmitter.SubmitForPresent").End()
	return s.submit(queuedCommandSubmitOptions{
		waitSemaphores:   waitSemaphores,
		waitStages:       waitStages,
		signalSemaphores: signalSemaphores,
		signalFence:      fence,
		waitForFence:     false,
		errorContext:     "present",
	})
}

func (s QueuedCommandSubmitter) SubmitAndWaitForReadback() bool {
	defer tracing.NewRegion("QueuedCommandSubmitter.SubmitAndWaitForReadback").End()
	return s.submit(queuedCommandSubmitOptions{
		waitForFence: true,
		errorContext: "readback",
	})
}

type queuedCommandSubmitOptions struct {
	waitSemaphores   []vk.Semaphore
	waitStages       []vk.PipelineStageFlags
	signalSemaphores []vk.Semaphore
	signalFence      vk.Fence
	waitForFence     bool
	errorContext     string
}

func (s QueuedCommandSubmitter) submit(options queuedCommandSubmitOptions) bool {
	if s.device == nil || len(s.device.Painter.writtenCommands) == 0 {
		return true
	}
	countTrace := tracing.NewRegion("QueuedCommandSubmitter.CommandCount")
	stageCount := s.stageCommandCounts()
	countTrace.End()
	lastStage := lastQueuedCommandStage(stageCount)
	if lastStage < 0 {
		s.device.Painter.writtenCommands = s.device.Painter.writtenCommands[:0]
		return true
	}
	commands := make([]vk.CommandBuffer, 0, len(s.device.Painter.writtenCommands))
	waited := false
	for stage := range queuedCommandStageCount {
		if stageCount[stage] == 0 {
			continue
		}
		commands = commands[:0]
		fence := vk.NullFence
		for i := range s.device.Painter.writtenCommands {
			cmd := &s.device.Painter.writtenCommands[i]
			if queuedCommandStage(cmd.stage) != stage {
				continue
			}
			commands = append(commands, cmd.buffer)
			fence = cmd.fence
		}
		if len(commands) == 0 {
			continue
		}
		if stage == lastStage && options.signalFence != vk.NullFence {
			fence = options.signalFence
		}
		if options.waitForFence && fence == vk.NullFence {
			slog.Error("queued command batch has no fence", "context", options.errorContext, "stage", stage)
			return false
		}
		submitInfo := vk.SubmitInfo{
			SType:              vulkan_const.StructureTypeSubmitInfo,
			PCommandBuffers:    &commands[0],
			CommandBufferCount: uint32(len(commands)),
		}
		if !waited && len(options.waitSemaphores) > 0 {
			submitInfo.WaitSemaphoreCount = uint32(len(options.waitSemaphores))
			submitInfo.PWaitSemaphores = &options.waitSemaphores[0]
			if len(options.waitStages) > 0 {
				submitInfo.PWaitDstStageMask = &options.waitStages[0]
			}
			waited = true
		}
		if stage == lastStage && len(options.signalSemaphores) > 0 {
			submitInfo.SignalSemaphoreCount = uint32(len(options.signalSemaphores))
			submitInfo.PSignalSemaphores = &options.signalSemaphores[0]
		}
		if !s.submitStage(stage, commands, submitInfo, fence, options) {
			return false
		}
	}
	s.device.Painter.writtenCommands = s.device.Painter.writtenCommands[:0]
	return true
}

func (s QueuedCommandSubmitter) stageCommandCounts() [queuedCommandStageCount]int {
	var counts [queuedCommandStageCount]int
	for i := range s.device.Painter.writtenCommands {
		counts[queuedCommandStage(s.device.Painter.writtenCommands[i].stage)]++
	}
	return counts
}

func queuedCommandStage(stage int) int {
	if stage < 0 || stage >= queuedCommandStageCount {
		return queuedCommandStageCount - 1
	}
	return stage
}

func lastQueuedCommandStage(counts [queuedCommandStageCount]int) int {
	for i := queuedCommandStageCount - 1; i >= 0; i-- {
		if counts[i] > 0 {
			return i
		}
	}
	return -1
}

func (s QueuedCommandSubmitter) submitStage(stage int, commands []vk.CommandBuffer, submitInfo vk.SubmitInfo, fence vk.Fence, options queuedCommandSubmitOptions) bool {
	defer tracing.NewRegion("QueuedCommandSubmitter.QueueSubmit").End()
	vkDevice := vk.Device(s.device.LogicalDevice.handle)
	if fence != vk.NullFence {
		res := vk.ResetFences(vkDevice, 1, &fence)
		if res != vulkan_const.Success {
			slog.Error("failed to reset queued command fence",
				slog.String("context", options.errorContext),
				slog.Int("stage", stage),
				slog.Int("count", len(commands)),
				slog.Int("code", int(res)))
			return false
		}
	}
	eCode := vk.QueueSubmit(vk.Queue(s.device.LogicalDevice.graphicsQueue), 1, &submitInfo, fence)
	if eCode != vulkan_const.Success {
		slog.Error("failed to submit queued command buffers",
			slog.String("context", options.errorContext),
			slog.Int("stage", stage),
			slog.Int("count", len(commands)),
			slog.Int("code", int(eCode)))
		return false
	}
	if !options.waitForFence {
		return true
	}
	wait := tracing.NewRegion("QueuedCommandSubmitter.WaitForFences")
	res := vk.WaitForFences(vkDevice, 1, &fence, vulkan_const.True, math.MaxUint64)
	wait.End()
	if res != vulkan_const.Success {
		slog.Error("failed to wait for queued command buffers",
			slog.String("context", options.errorContext),
			slog.Int("stage", stage),
			slog.Int("count", len(commands)),
			slog.Int("code", int(res)))
		return false
	}
	vk.ResetFences(vkDevice, 1, &fence)
	return true
}

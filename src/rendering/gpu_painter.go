/******************************************************************************/
/* gpu_painter.go                                                             */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/engine/collision"
	"kaijuengine.com/platform/profiler/tracing"
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

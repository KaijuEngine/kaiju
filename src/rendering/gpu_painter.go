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
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/platform/profiler/tracing"
)

const maxOcclusionDebugBounds = 512

type OcclusionDebugBound struct {
	Bounds   graviton.AABB
	Occluded bool
}

type OcclusionDebugState struct {
	RuntimeMode       OcclusionRuntimeMode
	VisualizationMode OcclusionVisualizationMode
	HiZPreviewMip     int
	Bounds            []OcclusionDebugBound
}

type GPUPainter struct {
	caches                RenderCaches
	imageIndex            [maxFramesInFlight]uint32
	descriptorPools       []GPUDescriptorPool
	currentFrame          int
	occlusionDebug        OcclusionDebugState
	hiZPyramid            HiZPyramid
	occlusionTester       GPUOcclusionTester
	combinedDrawings      Drawings
	combinedDrawingCuller combinedDrawingCuller
	preRuns               []func()
	writtenCommands       []CommandRecorder
	combineCmds           [maxFramesInFlight]CommandRecorder
	blitCmds              [maxFramesInFlight]CommandRecorder
	computeCmds           [maxFramesInFlight]CommandRecorder
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
	SampledImages  []ComputeTaskImage
	StorageImages  []ComputeTaskImage
}

type ComputeTaskImage struct {
	Texture *TextureId
	Aspect  GPUImageAspectFlags
}

func (g *GPUPainter) forceQueueCommand(cmd CommandRecorder, isPrePass bool) {
	if isPrePass {
		cmd.stage = 0
	} else {
		cmd.stage = 1
	}
	g.writtenCommands = append(g.writtenCommands, cmd)
}

func (g *GPUPainter) SetOcclusionRuntimeMode(mode OcclusionRuntimeMode) {
	g.occlusionDebug.RuntimeMode = mode
}

func (g *GPUPainter) OcclusionRuntimeMode() OcclusionRuntimeMode {
	return g.occlusionDebug.RuntimeMode
}

func (g *GPUPainter) SetOcclusionVisualizationMode(mode OcclusionVisualizationMode) {
	g.occlusionDebug.VisualizationMode = mode
}

func (g *GPUPainter) OcclusionVisualizationMode() OcclusionVisualizationMode {
	return g.occlusionDebug.VisualizationMode
}

func (g *GPUPainter) SetOcclusionHiZPreviewMip(mip int) {
	g.occlusionDebug.HiZPreviewMip = max(0, mip)
}

func (g *GPUPainter) OcclusionHiZPreviewMip() int {
	return g.occlusionDebug.HiZPreviewMip
}

func (g *GPUPainter) OcclusionTuning() OcclusionTuning {
	return g.occlusionDebug.RuntimeMode.Tuning()
}

func (g *GPUPainter) BeginOcclusionDebugFrame() {
	g.occlusionDebug.Bounds = g.occlusionDebug.Bounds[:0]
}

func (g *GPUPainter) OcclusionDebugBounds() []OcclusionDebugBound {
	out := make([]OcclusionDebugBound, len(g.occlusionDebug.Bounds))
	copy(out, g.occlusionDebug.Bounds)
	return out
}

func (g *GPUPainter) recordOcclusionDebugBound(instanceBase *ShaderDataBase) {
	if instanceBase == nil || len(g.occlusionDebug.Bounds) >= maxOcclusionDebugBounds {
		return
	}
	visibility := instanceBase.VisibilityState()
	if !visibility.FrustumVisible || !visibility.OcclusionEligible {
		return
	}
	occluded := !visibility.LastOcclusionVisible
	switch g.occlusionDebug.VisualizationMode {
	case OcclusionVisualizationTestedBounds:
		g.occlusionDebug.Bounds = append(g.occlusionDebug.Bounds, OcclusionDebugBound{
			Bounds:   instanceBase.renderBounds(),
			Occluded: occluded,
		})
	case OcclusionVisualizationOccludedBounds:
		if occluded {
			g.occlusionDebug.Bounds = append(g.occlusionDebug.Bounds, OcclusionDebugBound{
				Bounds:   instanceBase.renderBounds(),
				Occluded: true,
			})
		}
	}
}

func (g *GPUPainter) DestroyDescriptorPools(device *GPUDevice) {
	defer tracing.NewRegion("GPUPainter.DestroyDescriptorPools").End()
	g.destroyDescriptorPoolsImpl(device)
}

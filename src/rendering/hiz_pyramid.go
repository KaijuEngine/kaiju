/******************************************************************************/
/* hiz_pyramid.go                                                             */
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
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
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
	"encoding/json"
	"fmt"
	"log/slog"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

const hiZDownsampleShader = "hiz_downsample.shader"

type HiZPyramid struct {
	shader *Shader
	width  int
	height int
	frames [maxFramesInFlight]HiZPyramidFrame
}

type HiZPyramidFrame struct {
	Levels         []Texture
	descriptorSets [][maxFramesInFlight]GPUDescriptorSet
}

func (g *GPUPainter) QueueHiZPyramid(device *GPUDevice) {
	defer tracing.NewRegion("GPUPainter.QueueHiZPyramid").End()
	if g.caches == nil {
		return
	}
	source, ok := device.OcclusionDepthSource()
	if !ok {
		return
	}
	if err := g.ensureHiZPyramid(device, source.Width, source.Height); err != nil {
		slog.Error("failed to prepare Hi-Z pyramid", "error", err)
		return
	}
	frameIdx := g.currentFrame
	frame := &g.hiZPyramid.frames[frameIdx]
	for i := range frame.Levels {
		src := &source.RenderId
		srcAspect := GPUImageAspectDepthBit
		if i > 0 {
			src = &frame.Levels[i-1].RenderId
			srcAspect = GPUImageAspectColorBit
		}
		dst := &frame.Levels[i].RenderId
		g.writeHiZDescriptors(device, frame.descriptorSets[i][frameIdx], src, dst)
		g.computeTasks = append(g.computeTasks, ComputeTask{
			Shader:         g.hiZPyramid.shader,
			DescriptorSets: frame.descriptorSets[i][:],
			WorkGroups:     hiZDispatchGroups(frame.Levels[i].Width, frame.Levels[i].Height),
			SampledImages:  []ComputeTaskImage{{Texture: src, Aspect: srcAspect}},
			StorageImages:  []ComputeTaskImage{{Texture: dst, Aspect: GPUImageAspectColorBit}},
		})
	}
}

func (g *GPUPainter) HiZPyramidLevels() ([]*Texture, bool) {
	frame := &g.hiZPyramid.frames[g.currentFrame]
	if len(frame.Levels) == 0 {
		return nil, false
	}
	levels := make([]*Texture, len(frame.Levels))
	for i := range frame.Levels {
		if !frame.Levels[i].RenderId.IsValid() {
			return nil, false
		}
		levels[i] = &frame.Levels[i]
	}
	return levels, true
}

func (g *GPUPainter) DestroyHiZPyramid(device *GPUDevice) {
	defer tracing.NewRegion("GPUPainter.DestroyHiZPyramid").End()
	for i := range g.hiZPyramid.frames {
		frame := &g.hiZPyramid.frames[i]
		for j := range frame.Levels {
			device.LogicalDevice.FreeTexture(&frame.Levels[j].RenderId)
		}
		frame.Levels = nil
		frame.descriptorSets = nil
	}
	g.hiZPyramid.width = 0
	g.hiZPyramid.height = 0
}

func (g *GPUPainter) ensureHiZPyramid(device *GPUDevice, width, height int) error {
	if width <= 1 || height <= 1 {
		return fmt.Errorf("source depth texture is too small for Hi-Z: %dx%d", width, height)
	}
	if g.hiZPyramid.width == width && g.hiZPyramid.height == height && g.hiZPyramid.shader != nil {
		return nil
	}
	g.DestroyHiZPyramid(device)
	shader, err := g.hiZShader(device)
	if err != nil {
		return err
	}
	g.hiZPyramid.shader = shader
	g.hiZPyramid.width = width
	g.hiZPyramid.height = height
	dims := hiZPyramidLevelDimensions(width, height)
	for frameIdx := range g.hiZPyramid.frames {
		frame := &g.hiZPyramid.frames[frameIdx]
		frame.Levels = make([]Texture, len(dims))
		frame.descriptorSets = make([][maxFramesInFlight]GPUDescriptorSet, len(dims))
		for levelIdx := range dims {
			w, h := dims[levelIdx].Width(), dims[levelIdx].Height()
			key := fmt.Sprintf("hiz.frame%d.level%d", frameIdx, levelIdx)
			tex, err := createHiZTexture(device, key, w, h)
			if err != nil {
				g.DestroyHiZPyramid(device)
				return err
			}
			frame.Levels[levelIdx] = tex
			sets, _, err := device.createDescriptorSet(shader.RenderId.descriptorSetLayout, 0)
			if err != nil {
				g.DestroyHiZPyramid(device)
				return err
			}
			frame.descriptorSets[levelIdx] = sets
		}
	}
	return nil
}

func (g *GPUPainter) hiZShader(device *GPUDevice) (*Shader, error) {
	if g.hiZPyramid.shader != nil {
		return g.hiZPyramid.shader, nil
	}
	mem, err := g.caches.AssetDatabase().ReadText(hiZDownsampleShader)
	if err != nil {
		return nil, err
	}
	var shaderData ShaderData
	if err := json.Unmarshal([]byte(mem), &shaderData); err != nil {
		return nil, err
	}
	shader, _ := g.caches.ShaderCache().Shader(shaderData.Compile())
	g.caches.ShaderCache().CreatePending()
	if !shader.RenderId.computePipeline.IsValid() {
		return nil, fmt.Errorf("Hi-Z compute shader did not create a valid pipeline")
	}
	return shader, nil
}

func createHiZTexture(device *GPUDevice, key string, width, height int32) (Texture, error) {
	tex := Texture{Key: key, Width: int(width), Height: int(height)}
	id := &tex.RenderId
	err := device.CreateImage(id, GPUMemoryPropertyDeviceLocalBit, GPUImageCreateRequest{
		ImageType:   GPUImageType2d,
		Extent:      matrix.Vec3i{width, height, 1},
		MipLevels:   1,
		ArrayLayers: 1,
		Format:      GPUFormatR32Sfloat,
		Tiling:      GPUImageTilingOptimal,
		Usage:       GPUImageUsageSampledBit | GPUImageUsageStorageBit,
		Samples:     GPUSampleCount1Bit,
	})
	if err != nil {
		return tex, err
	}
	if err = device.LogicalDevice.CreateImageView(id, GPUImageAspectColorBit, GPUImageViewType2d); err != nil {
		device.LogicalDevice.FreeTexture(id)
		return tex, err
	}
	id.Sampler, err = device.CreateTextureSampler(1, GPUFilterNearest)
	if err != nil {
		device.LogicalDevice.FreeTexture(id)
		return tex, err
	}
	device.TransitionImageLayout(id, GPUImageLayoutShaderReadOnlyOptimal,
		GPUImageAspectColorBit, GPUAccessShaderReadBit, nil)
	return tex, nil
}

func (g *GPUPainter) writeHiZDescriptors(device *GPUDevice, set GPUDescriptorSet, src, dst *TextureId) {
	srcInfo := vk.DescriptorImageInfo{
		Sampler:     vk.Sampler(src.Sampler.handle),
		ImageView:   vk.ImageView(src.View.handle),
		ImageLayout: vulkan_const.ImageLayoutShaderReadOnlyOptimal,
	}
	dstInfo := vk.DescriptorImageInfo{
		ImageView:   vk.ImageView(dst.View.handle),
		ImageLayout: vulkan_const.ImageLayoutGeneral,
	}
	writes := [2]vk.WriteDescriptorSet{
		prepareSetWriteImageTyped(vk.DescriptorSet(set.handle), []vk.DescriptorImageInfo{srcInfo}, 0, vulkan_const.DescriptorTypeCombinedImageSampler),
		prepareSetWriteImageTyped(vk.DescriptorSet(set.handle), []vk.DescriptorImageInfo{dstInfo}, 1, vulkan_const.DescriptorTypeStorageImage),
	}
	vk.UpdateDescriptorSets(vk.Device(device.LogicalDevice.handle), uint32(len(writes)), &writes[0], 0, nil)
}

func hiZPyramidLevelDimensions(width, height int) []matrix.Vec2i {
	levels := make([]matrix.Vec2i, 0)
	w := max(1, width/2)
	h := max(1, height/2)
	for {
		levels = append(levels, matrix.Vec2i{int32(w), int32(h)})
		if w == 1 && h == 1 {
			break
		}
		w = max(1, (w+1)/2)
		h = max(1, (h+1)/2)
	}
	return levels
}

func hiZDispatchGroups(width, height int) [3]uint32 {
	return [3]uint32{
		uint32((width + 7) / 8),
		uint32((height + 7) / 8),
		1,
	}
}

func hiZReduceDepth(depths ...float32) float32 {
	if len(depths) == 0 {
		return 1
	}
	out := depths[0]
	for i := 1; i < len(depths); i++ {
		out = max(out, depths[i])
	}
	return out
}

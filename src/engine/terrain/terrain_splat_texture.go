/******************************************************************************/
/* terrain_splat_texture.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package terrain

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const splatTextureChannels = 4

type TerrainSplatTexture struct {
	Key        string
	Texture    *rendering.Texture
	LayerStart int
	LayerCount int
	Pixels     []byte
	Dirty      DirtyRegion

	uploadPixels []byte
}

type SplatLayerChannel struct {
	Texture int
	Channel int
}

func (t *Terrain) SplatTextureCount() int {
	if t == nil {
		return 0
	}
	return len(t.SplatTextures)
}

func (t *Terrain) SplatLayerChannel(layer int) (SplatLayerChannel, bool) {
	if t == nil || t.LayerSet == nil || t.LayerSet.WeightMap == nil {
		return SplatLayerChannel{}, false
	}
	return splatLayerChannel(layer, t.LayerSet.WeightMap.Layers)
}

func (t *Terrain) MarkTextureDirty(layer int, region DirtyRegion) {
	if t == nil || !region.Valid {
		return
	}
	if t.LayerSet != nil && t.LayerSet.WeightMap != nil &&
		len(t.SplatTextures) != splatTextureCount(t.LayerSet.WeightMap.Layers) {
		_ = t.createSplatTextures(t.host)
	}
	channel, ok := t.SplatLayerChannel(layer)
	if !ok || channel.Texture >= len(t.SplatTextures) {
		return
	}
	t.SplatTextures[channel.Texture].Dirty = mergeDirtyRegions(t.SplatTextures[channel.Texture].Dirty, region)
}

func (t *Terrain) MarkTextureRegionDirty(region DirtyRegion) {
	if t == nil || !region.Valid {
		return
	}
	if t.LayerSet != nil && t.LayerSet.WeightMap != nil &&
		len(t.SplatTextures) != splatTextureCount(t.LayerSet.WeightMap.Layers) {
		_ = t.createSplatTextures(t.host)
	}
	for i := range t.SplatTextures {
		t.SplatTextures[i].Dirty = mergeDirtyRegions(t.SplatTextures[i].Dirty, region)
	}
}

func (t *Terrain) ApplyTextureDirty(region DirtyRegion) {
	if t == nil || t.LayerSet == nil || t.LayerSet.WeightMap == nil {
		return
	}
	resolution := t.LayerSet.WeightMap.Resolution
	if len(t.SplatTextures) != splatTextureCount(t.LayerSet.WeightMap.Layers) {
		_ = t.createSplatTextures(t.host)
	}
	device := t.gpuDevice()
	for i := range t.SplatTextures {
		dirty := t.SplatTextures[i].Dirty
		if region.Valid {
			dirty = intersectDirtyRegions(dirty, region)
		}
		if !dirty.Valid {
			continue
		}
		dirty = dirty.Expand(0, resolution)
		if !dirty.Valid {
			continue
		}
		request := t.SplatTextureWriteRequest(i, dirty)
		if len(request.Pixels) == 0 {
			continue
		}
		t.writeSplatPixelsRegion(i, dirty, request.Pixels)
		if device != nil && t.SplatTextures[i].Texture != nil && t.SplatTextures[i].Texture.RenderId.IsValid() {
			t.SplatTextures[i].Texture.WritePixels(device, []rendering.GPUImageWriteRequest{request})
			t.SplatTextures[i].Dirty = subtractAppliedDirtyRegion(t.SplatTextures[i].Dirty, dirty)
		}
	}
}

func (t *Terrain) SplatTextureWriteRequest(texture int, region DirtyRegion) rendering.GPUImageWriteRequest {
	if t == nil || texture < 0 || texture >= len(t.SplatTextures) ||
		t.LayerSet == nil || t.LayerSet.WeightMap == nil || !region.Valid {
		return rendering.GPUImageWriteRequest{}
	}
	region = region.Expand(0, t.LayerSet.WeightMap.Resolution)
	if !region.Valid {
		return rendering.GPUImageWriteRequest{}
	}
	weightMap := t.LayerSet.EffectiveWeightMapForPreview()
	pixels := packSplatTextureRegionInto(weightMap, texture, region, t.SplatTextures[texture].uploadPixels)
	t.SplatTextures[texture].uploadPixels = pixels
	return rendering.GPUImageWriteRequest{
		Region: matrix.Vec4i{
			int32(region.MinX),
			int32(region.MinZ),
			int32(region.MaxX - region.MinX + 1),
			int32(region.MaxZ - region.MinZ + 1),
		},
		Pixels: pixels,
	}
}

func (t *Terrain) createSplatTextures(host *engine.Host) error {
	if t == nil || t.LayerSet == nil || t.LayerSet.WeightMap == nil {
		return nil
	}
	count := splatTextureCount(t.LayerSet.WeightMap.Layers)
	t.SplatTextures = make([]TerrainSplatTexture, count)
	weightMap := t.LayerSet.EffectiveWeightMapForPreview()
	for i := range count {
		pixels := packSplatTexture(weightMap, i)
		key := fmt.Sprintf("terrain_%p_splat_%d_%d", t, t.LayerSet.WeightMap.Layers, i)
		splat := TerrainSplatTexture{
			Key:        key,
			LayerStart: i * splatTextureChannels,
			LayerCount: min(splatTextureChannels, t.LayerSet.WeightMap.Layers-i*splatTextureChannels),
			Pixels:     pixels,
		}
		if host != nil {
			tex, err := host.TextureCache().InsertRawTexture(
				key,
				pixels,
				t.LayerSet.WeightMap.Resolution,
				t.LayerSet.WeightMap.Resolution,
				rendering.TextureFilterLinear,
			)
			if err != nil {
				return err
			}
			splat.Texture = tex
		}
		t.SplatTextures[i] = splat
	}
	return nil
}

func (t *Terrain) writeSplatPixelsRegion(texture int, region DirtyRegion, pixels []byte) {
	if t == nil || texture < 0 || texture >= len(t.SplatTextures) ||
		t.LayerSet == nil || t.LayerSet.WeightMap == nil || !region.Valid {
		return
	}
	resolution := t.LayerSet.WeightMap.Resolution
	region = region.Expand(0, resolution)
	if !region.Valid {
		return
	}
	width := region.MaxX - region.MinX + 1
	height := region.MaxZ - region.MinZ + 1
	if len(pixels) != width*height*splatTextureChannels ||
		len(t.SplatTextures[texture].Pixels) != resolution*resolution*splatTextureChannels {
		return
	}
	for row := 0; row < height; row++ {
		src := row * width * splatTextureChannels
		dst := (region.MinX + (region.MinZ+row)*resolution) * splatTextureChannels
		copy(t.SplatTextures[texture].Pixels[dst:dst+width*splatTextureChannels],
			pixels[src:src+width*splatTextureChannels])
	}
}

func (t *Terrain) gpuDevice() *rendering.GPUDevice {
	if t == nil || t.host == nil || t.host.Window == nil || t.host.Window.GpuInstance == nil {
		return nil
	}
	return t.host.Window.GpuInstance.PrimaryDevice()
}

func splatTextureCount(layers int) int {
	if layers <= 0 {
		return 0
	}
	return (layers + splatTextureChannels - 1) / splatTextureChannels
}

func splatLayerChannel(layer, layers int) (SplatLayerChannel, bool) {
	if layer < 0 || layer >= layers {
		return SplatLayerChannel{}, false
	}
	return SplatLayerChannel{
		Texture: layer / splatTextureChannels,
		Channel: layer % splatTextureChannels,
	}, true
}

func packSplatTexture(weightMap *TextureWeightMap, texture int) []byte {
	if weightMap == nil || texture < 0 || texture >= splatTextureCount(weightMap.Layers) {
		return nil
	}
	region := DirtyRegion{MinX: 0, MinZ: 0, MaxX: weightMap.Resolution - 1, MaxZ: weightMap.Resolution - 1, Valid: true}
	return packSplatTextureRegion(weightMap, texture, region)
}

func packSplatTextureRegion(weightMap *TextureWeightMap, texture int, region DirtyRegion) []byte {
	return packSplatTextureRegionInto(weightMap, texture, region, nil)
}

func packSplatTextureRegionInto(weightMap *TextureWeightMap, texture int, region DirtyRegion, pixels []byte) []byte {
	if weightMap == nil || texture < 0 || texture >= splatTextureCount(weightMap.Layers) || !region.Valid {
		return nil
	}
	region = region.Expand(0, weightMap.Resolution)
	if !region.Valid {
		return nil
	}
	width := region.MaxX - region.MinX + 1
	height := region.MaxZ - region.MinZ + 1
	required := width * height * splatTextureChannels
	if cap(pixels) < required {
		pixels = make([]byte, required)
	}
	pixels = pixels[:required]
	layerStart := texture * splatTextureChannels
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			dst := ((x - region.MinX) + (z-region.MinZ)*width) * splatTextureChannels
			for channel := 0; channel < splatTextureChannels; channel++ {
				layer := layerStart + channel
				weight := matrix.Float(0)
				if layer < weightMap.Layers {
					weight = weightMap.WeightAt(layer, x, z)
				}
				pixels[dst+channel] = weightToByte(weight)
			}
		}
	}
	return pixels
}

func weightToByte(weight matrix.Float) byte {
	return byte(matrix.Clamp(weight, 0, 1)*255 + 0.5)
}

func intersectDirtyRegions(a, b DirtyRegion) DirtyRegion {
	if !a.Valid || !b.Valid {
		return DirtyRegion{}
	}
	out := DirtyRegion{
		MinX:  max(a.MinX, b.MinX),
		MinZ:  max(a.MinZ, b.MinZ),
		MaxX:  min(a.MaxX, b.MaxX),
		MaxZ:  min(a.MaxZ, b.MaxZ),
		Valid: true,
	}
	if out.MinX > out.MaxX || out.MinZ > out.MaxZ {
		return DirtyRegion{}
	}
	return out
}

func subtractAppliedDirtyRegion(original, applied DirtyRegion) DirtyRegion {
	if !original.Valid || !applied.Valid {
		return original
	}
	if original == applied {
		return DirtyRegion{}
	}
	if applied.MinX <= original.MinX && applied.MinZ <= original.MinZ &&
		applied.MaxX >= original.MaxX && applied.MaxZ >= original.MaxZ {
		return DirtyRegion{}
	}
	return original
}

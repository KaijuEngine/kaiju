/******************************************************************************/
/* terrain_material_atlas.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package terrain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/textures"
)

const (
	terrainAtlasColumns    = 4
	terrainAtlasRows       = 2
	terrainAtlasTileSize   = 1024
	terrainAtlasGutter     = 2
	terrainLayerParamCount = 3
)

type terrainAtlasSource struct {
	pixels        []byte
	width, height int
}

func (t *Terrain) terrainMaterialAtlases(host *engine.Host) (*rendering.Texture, *rendering.Texture, error) {
	if host == nil {
		return nil, nil, errors.New("terrain material atlases require a host")
	}
	layers := t.renderedLayers()
	materialKey, normalKey := terrainMaterialAtlasCacheKeys(layers)
	material, materialFound := host.TextureCache().Find(materialKey, textures.TextureFilterLinear)
	normal, normalFound := host.TextureCache().Find(normalKey, textures.TextureFilterLinear)
	if materialFound && normalFound {
		return material, normal, nil
	}

	materialPixels, normalPixels, width, height, diagnostics, err := buildTerrainAtlasPixelsWithLoader(
		func(key string) (textures.TextureData, error) {
			return host.TextureCache().TexturePixels(key, textures.TextureFilterLinear)
		}, layers, terrainAtlasTileSize, terrainAtlasGutter)
	if err != nil {
		return nil, nil, err
	}
	for _, diagnostic := range diagnostics {
		slog.Warn("terrain material atlas used a fallback", "layer", diagnostic.layer,
			"map", diagnostic.kind, "texture", diagnostic.texture, "error", diagnostic.err)
	}
	material, err = host.TextureCache().InsertRawTexture(materialKey, materialPixels, width, height, textures.TextureFilterLinear)
	if err != nil {
		return nil, nil, fmt.Errorf("create terrain albedo/roughness atlas: %w", err)
	}
	normal, err = host.TextureCache().InsertRawTexture(normalKey, normalPixels, width, height, textures.TextureFilterLinear)
	if err != nil {
		return nil, nil, fmt.Errorf("create terrain normal atlas: %w", err)
	}
	return material, normal, nil
}

func (t *Terrain) renderedLayers() []TerrainLayer {
	layers := make([]TerrainLayer, MaxRenderedLayers)
	if t != nil && t.LayerSet != nil {
		copy(layers, t.LayerSet.Layers[:min(len(t.LayerSet.Layers), MaxRenderedLayers)])
	}
	return layers
}

func terrainMaterialAtlasCacheKeys(layers []TerrainLayer) (string, string) {
	hash := sha256.New()
	for i := 0; i < MaxRenderedLayers; i++ {
		layer := TerrainLayer{}
		if i < len(layers) {
			layer = layers[i]
		}
		_, _ = hash.Write([]byte(strings.TrimSpace(layer.TextureContentID)))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(strings.TrimSpace(layer.NormalContentID)))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(strings.TrimSpace(layer.RoughnessContentID)))
		_, _ = hash.Write([]byte{0xff})
	}
	suffix := hex.EncodeToString(hash.Sum(nil)[:12])
	return "terrain_material_atlas_" + suffix, "terrain_normal_atlas_" + suffix
}

type terrainAtlasDiagnostic struct {
	layer   int
	kind    string
	texture string
	err     error
}

func buildTerrainAtlasPixels(assetDb assets.Database, layers []TerrainLayer, tileSize, gutter int) (
	material, normal []byte, width, height int, diagnostics []terrainAtlasDiagnostic, err error,
) {
	if assetDb == nil {
		return nil, nil, 0, 0, nil, errors.New("terrain material atlas requires an asset database")
	}
	return buildTerrainAtlasPixelsWithLoader(func(key string) (textures.TextureData, error) {
		return rendering.TexturePixelsFromAsset(assetDb, key)
	}, layers, tileSize, gutter)
}

func buildTerrainAtlasPixelsWithLoader(load func(string) (textures.TextureData, error),
	layers []TerrainLayer, tileSize, gutter int) (
	material, normal []byte, width, height int, diagnostics []terrainAtlasDiagnostic, err error,
) {
	if load == nil {
		return nil, nil, 0, 0, nil, errors.New("terrain material atlas requires a pixel loader")
	}
	if tileSize <= gutter*2+1 || gutter < 0 {
		return nil, nil, 0, 0, nil, fmt.Errorf("invalid terrain atlas tile size %d and gutter %d", tileSize, gutter)
	}
	width = terrainAtlasColumns * tileSize
	height = terrainAtlasRows * tileSize
	material = make([]byte, width*height*4)
	normal = make([]byte, width*height*4)

	for layerIndex := 0; layerIndex < MaxRenderedLayers; layerIndex++ {
		layer := TerrainLayer{}
		if layerIndex < len(layers) {
			layer = layers[layerIndex]
		}
		albedoFallback := terrainAtlasSource{pixels: []byte{255, 255, 255, 255}, width: 1, height: 1}
		roughnessFallback := terrainAtlasSource{pixels: []byte{255, 255, 255, 255}, width: 1, height: 1}
		normalFallback := terrainAtlasSource{pixels: []byte{128, 128, 255, 255}, width: 1, height: 1}
		if strings.TrimSpace(layer.TextureContentID) == "" && strings.TrimSpace(layer.NormalContentID) == "" &&
			strings.TrimSpace(layer.RoughnessContentID) == "" {
			writeTerrainAtlasTile(material, normal, width, tileSize, gutter, layerIndex,
				albedoFallback, roughnessFallback, normalFallback)
			continue
		}
		albedo, diagnostic := terrainAtlasTextureSource(load, layerIndex, "albedo", layer.TextureContentID, albedoFallback)
		if diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}
		roughness, diagnostic := terrainAtlasTextureSource(load, layerIndex, "roughness", layer.RoughnessContentID, roughnessFallback)
		if diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}
		normalSource, diagnostic := terrainAtlasTextureSource(load, layerIndex, "normal", layer.NormalContentID, normalFallback)
		if diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}

		writeTerrainAtlasTile(material, normal, width, tileSize, gutter, layerIndex,
			albedo, roughness, normalSource)
	}
	return material, normal, width, height, diagnostics, nil
}

func writeTerrainAtlasTile(material, normal []byte, atlasWidth, tileSize, gutter, layerIndex int,
	albedo, roughness, normalSource terrainAtlasSource) {
	innerSize := tileSize - gutter*2
	tileX := (layerIndex % terrainAtlasColumns) * tileSize
	tileY := (layerIndex / terrainAtlasColumns) * tileSize
	for y := 0; y < tileSize; y++ {
		contentY := max(0, min(innerSize-1, y-gutter))
		v := float64(contentY) / float64(max(innerSize-1, 1))
		for x := 0; x < tileSize; x++ {
			contentX := max(0, min(innerSize-1, x-gutter))
			u := float64(contentX) / float64(max(innerSize-1, 1))
			albedoSample := sampleTerrainAtlasSource(albedo, u, v)
			roughnessSample := sampleTerrainAtlasSource(roughness, u, v)
			normalSample := sampleTerrainAtlasSource(normalSource, u, v)
			dst := ((tileX + x) + (tileY+y)*atlasWidth) * 4
			material[dst+0] = albedoSample[0]
			material[dst+1] = albedoSample[1]
			material[dst+2] = albedoSample[2]
			material[dst+3] = roughnessSample[0]
			normal[dst+0] = normalSample[0]
			normal[dst+1] = normalSample[1]
			normal[dst+2] = normalSample[2]
			normal[dst+3] = 255
		}
	}
}

func terrainAtlasTextureSource(load func(string) (textures.TextureData, error), layer int, kind, key string, fallback terrainAtlasSource) (
	terrainAtlasSource, *terrainAtlasDiagnostic,
) {
	key = strings.TrimSpace(key)
	if key == "" {
		return fallback, &terrainAtlasDiagnostic{layer: layer, kind: kind, texture: key, err: errors.New("texture is not assigned")}
	}
	data, err := load(key)
	if err != nil || data.Width <= 0 || data.Height <= 0 || len(data.Mem) < data.Width*data.Height*4 {
		if err == nil {
			err = errors.New("texture did not decode to RGBA pixels")
		}
		return fallback, &terrainAtlasDiagnostic{layer: layer, kind: kind, texture: key, err: err}
	}
	return terrainAtlasSource{pixels: data.Mem, width: data.Width, height: data.Height}, nil
}

func sampleTerrainAtlasSource(source terrainAtlasSource, u, v float64) [4]byte {
	if source.width <= 1 || source.height <= 1 {
		return [4]byte{source.pixels[0], source.pixels[1], source.pixels[2], source.pixels[3]}
	}
	x := u * float64(source.width-1)
	y := v * float64(source.height-1)
	x0, y0 := int(x), int(y)
	x1, y1 := min(x0+1, source.width-1), min(y0+1, source.height-1)
	tx, ty := x-float64(x0), y-float64(y0)
	var result [4]byte
	for channel := range 4 {
		p00 := float64(source.pixels[(x0+y0*source.width)*4+channel])
		p10 := float64(source.pixels[(x1+y0*source.width)*4+channel])
		p01 := float64(source.pixels[(x0+y1*source.width)*4+channel])
		p11 := float64(source.pixels[(x1+y1*source.width)*4+channel])
		top := p00 + (p10-p00)*tx
		bottom := p01 + (p11-p01)*tx
		result[channel] = byte(top + (bottom-top)*ty + 0.5)
	}
	return result
}

func (t *Terrain) shaderLayerData() [shader_data_registry.TerrainLayerParameterCount]matrix.Vec4 {
	var result [shader_data_registry.TerrainLayerParameterCount]matrix.Vec4
	for i, layer := range t.renderedLayers() {
		layer = normalizeTerrainLayer(layer)
		scale := layer.Tiling
		if layer.TextureWorldSize.X() > matrix.Tiny && layer.TextureWorldSize.Y() > matrix.Tiny {
			scale = matrix.NewVec2(
				t.Config.WorldSize.X()/layer.TextureWorldSize.X(),
				t.Config.WorldSize.Y()/layer.TextureWorldSize.Y(),
			)
		}
		sin, cos := matrix.Sin(layer.Rotation), matrix.Cos(layer.Rotation)
		base := i * terrainLayerParamCount
		result[base+0] = matrix.NewVec4(scale.X(), scale.Y(), layer.Offset.X(), layer.Offset.Y())
		result[base+1] = matrix.NewVec4(cos, sin, layer.Tint.R(), layer.Tint.G())
		result[base+2] = matrix.NewVec4(layer.Tint.B(), layer.Tint.A(), 0, 0)
	}
	return result
}

func (t *Terrain) refreshShaderLayerData() {
	if t == nil {
		return
	}
	params := t.shaderLayerData()
	for i := range t.ShaderData {
		if data, ok := t.ShaderData[i].(*shader_data_registry.ShaderDataTerrain); ok {
			data.SetLayerParameters(params)
		}
	}
}

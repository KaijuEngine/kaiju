/******************************************************************************/
/* terrain_asset_v2_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package terrain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestTerrainAssetVersion2RoundTripsPaintLayersAndWeights(t *testing.T) {
	config := TerrainConfig{
		Resolution:      3,
		PaintResolution: 4,
		WorldSize:       matrix.NewVec2(8, 8),
		MinHeight:       -2,
		MaxHeight:       6,
		InitialHeight:   0,
	}
	set, err := NewTerrainLayerSet(config.PaintResolution)
	if err != nil {
		t.Fatal(err)
	}
	grass := set.AddLayer(TerrainLayer{
		Name:               "Grass",
		TextureContentID:   "grass-albedo",
		NormalContentID:    "grass-normal",
		RoughnessContentID: "grass-roughness",
		Filter:             rendering.TextureFilterNearest,
		Tiling:             matrix.NewVec2(6, 4),
		Offset:             matrix.NewVec2(0.25, 0.5),
		Rotation:           0.2,
		Tint:               matrix.ColorGreen(),
	})
	rock := set.AddLayer(TerrainLayer{
		Name:             "Rock",
		TextureContentID: "rock-albedo",
		Filter:           rendering.TextureFilterLinear,
		Tiling:           matrix.NewVec2(2, 3),
		Tint:             matrix.ColorGray(),
	})
	set.SetLayerWeightAt(grass, 2, 1, 0.25)
	set.SetLayerWeightAt(rock, 2, 1, 0.75)
	set.NormalizeWeightsAt(2, 1)
	asset, err := NewAssetWithLayerSet(config, []matrix.Float{
		-2, -1, 0,
		1, 2, 3,
		4, 5, 6,
	}, set)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Version != AssetVersion {
		t.Fatalf("expected asset version %d, got %d", AssetVersion, asset.Version)
	}
	data, err := asset.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := DeserializeAsset(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != AssetVersion {
		t.Fatalf("expected loaded asset version %d, got %d", AssetVersion, loaded.Version)
	}
	if len(loaded.Layers) != 2 {
		t.Fatalf("expected two loaded layers, got %d", len(loaded.Layers))
	}
	if loaded.Layers[0].TextureContentID != "grass-albedo" || loaded.Layers[0].NormalContentID != "grass-normal" {
		t.Fatalf("expected layer metadata to round-trip, got %+v", loaded.Layers[0])
	}
	if loaded.Layers[0].Name != "Grass" || loaded.Layers[1].Name != "Rock" {
		t.Fatalf("expected layer names to round-trip, got %+v", loaded.Layers)
	}
	if loaded.Layers[0].Filter != rendering.TextureFilterNearest {
		t.Fatalf("expected nearest filter to round-trip, got %d", loaded.Layers[0].Filter)
	}
	if len(loaded.Config.Textures) != 2 || loaded.Config.Textures[0].Key != "grass-albedo" ||
		loaded.Config.Textures[1].Key != "rock-albedo" {
		t.Fatalf("expected terrain config textures to follow layers, got %+v", loaded.Config.Textures)
	}
	if loaded.WeightMapResolution != config.PaintResolution {
		t.Fatalf("expected weight map resolution %d, got %d", config.PaintResolution, loaded.WeightMapResolution)
	}
	heights := loaded.FloatHeights()
	if !matrix.ApproxTo(heights[4], 2, 0.001) || !matrix.ApproxTo(heights[8], 6, 0.001) {
		t.Fatalf("expected heights to round-trip, got %v", heights)
	}
	layerSet, err := loaded.LayerSet()
	if err != nil {
		t.Fatal(err)
	}
	if got := layerSet.LayerWeightAt(grass, 2, 1); !matrix.ApproxTo(got, 0.25, 0.001) {
		t.Fatalf("expected grass weight to round-trip, got %f", got)
	}
	if got := layerSet.LayerWeightAt(rock, 2, 1); !matrix.ApproxTo(got, 0.75, 0.001) {
		t.Fatalf("expected rock weight to round-trip, got %f", got)
	}
	model, err := NewModelFromAsset(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if got := model.LayerSet.LayerWeightAt(rock, 2, 1); !matrix.ApproxTo(got, 0.75, 0.001) {
		t.Fatalf("expected model to receive asset paint weights, got %f", got)
	}
}

func TestTerrainAssetFromTerrainSavesPaintedLayerStack(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:      5,
		PaintResolution: 5,
		WorldSize:       matrix.NewVec2(8, 8),
		MinHeight:       0,
		MaxHeight:       4,
	})
	if err != nil {
		t.Fatal(err)
	}
	base := model.AddLayer(TerrainLayer{Name: "Base", TextureContentID: "base"})
	rock := model.AddLayer(TerrainLayer{Name: "Rock", TextureContentID: "rock", Tint: matrix.ColorGray()})
	model.LayerSet.SetLayerWeightAt(base, 2, 2, 0.2)
	model.LayerSet.SetLayerWeightAt(rock, 2, 2, 0.8)
	model.NormalizeWeightsAt(2, 2)
	asset, err := NewAssetFromTerrain(model)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Version != AssetVersion {
		t.Fatalf("expected saved terrain asset version %d, got %d", AssetVersion, asset.Version)
	}
	data, err := asset.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	loadedAsset, err := DeserializeAsset(data)
	if err != nil {
		t.Fatal(err)
	}
	reverted, err := NewModelFromAsset(loadedAsset)
	if err != nil {
		t.Fatal(err)
	}
	if got := reverted.LayerCount(); got != 2 {
		t.Fatalf("expected reverted terrain to restore two layers, got %d", got)
	}
	if got := reverted.LayerSet.Layers[1].Name; got != "Rock" {
		t.Fatalf("expected reverted layer stack to preserve Rock layer, got %q", got)
	}
	if got := reverted.LayerWeightAt(rock, 2, 2); !matrix.ApproxTo(got, 0.8, 0.001) {
		t.Fatalf("expected reverted paint weight 0.8, got %f", got)
	}
}

func TestTerrainAssetVersion1LoadsWithDefaultPaintLayer(t *testing.T) {
	config := TerrainConfig{
		Resolution:    2,
		WorldSize:     matrix.NewVec2(4, 4),
		MinHeight:     0,
		MaxHeight:     4,
		InitialHeight: 0,
		Textures: []TerrainTexture{{
			Key:    "legacy-grass",
			Filter: rendering.TextureFilterNearest,
		}},
	}
	legacy, err := serializeLegacyTerrainAsset(config, []matrix.Float{0, 1, 2, 4})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := DeserializeAsset(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != AssetVersion {
		t.Fatalf("expected legacy asset to upgrade to version %d, got %d", AssetVersion, loaded.Version)
	}
	if len(loaded.Layers) != 1 {
		t.Fatalf("expected one default layer for legacy asset, got %d", len(loaded.Layers))
	}
	if loaded.Layers[0].TextureContentID != "legacy-grass" {
		t.Fatalf("expected legacy layer to use terrain texture, got %+v", loaded.Layers[0])
	}
	layerSet, err := loaded.LayerSet()
	if err != nil {
		t.Fatal(err)
	}
	for z := 0; z < layerSet.WeightMap.Resolution; z++ {
		for x := 0; x < layerSet.WeightMap.Resolution; x++ {
			if got := layerSet.LayerWeightAt(0, x, z); !matrix.ApproxTo(got, 1, matrix.Roughly) {
				t.Fatalf("expected legacy default weight at %d,%d to be 1, got %f", x, z, got)
			}
		}
	}
	model, err := NewModelFromAsset(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if model.LayerCount() != 1 {
		t.Fatalf("expected model from legacy asset to have one paint layer, got %d", model.LayerCount())
	}
	if got := model.LayerSet.Layers[0].TextureContentID; got != "legacy-grass" {
		t.Fatalf("expected model legacy texture id, got %q", got)
	}
	if got := model.LayerWeightAt(0, 1, 1); !matrix.ApproxTo(got, 1, matrix.Roughly) {
		t.Fatalf("expected model legacy weight to be 1, got %f", got)
	}
}

func TestTerrainAssetVersion2RejectsMissingPaintData(t *testing.T) {
	asset, err := NewAsset(TerrainConfig{
		Resolution:    2,
		WorldSize:     matrix.NewVec2(2, 2),
		MinHeight:     0,
		MaxHeight:     1,
		InitialHeight: 0,
	}, []matrix.Float{0, 0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	asset.Layers = nil
	asset.Weights = nil
	if _, err := asset.Serialize(); err == nil {
		t.Fatal("expected v2 asset without paint data to fail validation")
	}
}

func serializeLegacyTerrainAsset(config TerrainConfig, heights []matrix.Float) ([]byte, error) {
	config = normalizeConfig(config)
	header := terrainAssetHeader{
		Version:        1,
		Config:         config,
		HeightEncoding: HeightEncodingUint16,
		HeightCount:    len(heights),
	}
	headerData, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.Write(terrainAssetMagic)
	if err := binary.Write(&out, binary.LittleEndian, uint32(len(headerData))); err != nil {
		return nil, err
	}
	out.Write(headerData)
	for i := range heights {
		height := normalizeHeightToUint16(heights[i], config.MinHeight, config.MaxHeight)
		if err := binary.Write(&out, binary.LittleEndian, height); err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}

func TestTerrainAssetLegacyJSONLoadsWithDefaultPaintLayer(t *testing.T) {
	asset := TerrainAsset{
		Version: 0,
		Config: TerrainConfig{
			Resolution:    2,
			WorldSize:     matrix.NewVec2(2, 2),
			MinHeight:     0,
			MaxHeight:     1,
			InitialHeight: 0,
			Textures:      []TerrainTexture{{Key: assets.TextureSquare, Filter: rendering.TextureFilterLinear}},
		},
		Heights: []uint16{0, 0, 0, 65535},
	}
	data, err := json.Marshal(asset)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := DeserializeAsset(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != AssetVersion || len(loaded.Layers) != 1 || len(loaded.Weights) != 4 {
		t.Fatalf("expected legacy JSON to upgrade paint data, got version=%d layers=%d weights=%d",
			loaded.Version, len(loaded.Layers), len(loaded.Weights))
	}
}

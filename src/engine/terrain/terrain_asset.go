/******************************************************************************/
/* terrain_asset.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package terrain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const (
	AssetVersion = 2
	heightU16Max = matrix.Float(65535)
)

var terrainAssetMagic = []byte{'K', 'T', 'R', 'N'}

type HeightEncoding string
type WeightEncoding string

const (
	HeightEncodingUint16 HeightEncoding = "uint16-normalized"
	WeightEncodingUint16 WeightEncoding = "uint16-normalized"
)

type TerrainAsset struct {
	Version             int
	Config              TerrainConfig
	Heights             []uint16
	Layers              []TerrainLayer
	WeightMapResolution int
	Weights             []uint16
}

type terrainAssetHeader struct {
	Version             int
	Config              TerrainConfig
	HeightEncoding      HeightEncoding
	HeightCount         int
	Layers              []TerrainLayer `json:",omitempty"`
	WeightEncoding      WeightEncoding `json:",omitempty"`
	WeightMapResolution int            `json:",omitempty"`
	WeightCount         int            `json:",omitempty"`
}

func NewAsset(config TerrainConfig, heights []matrix.Float) (TerrainAsset, error) {
	return NewAssetWithLayerSet(config, heights, nil)
}

func NewAssetWithLayerSet(config TerrainConfig, heights []matrix.Float, layerSet *TerrainLayerSet) (TerrainAsset, error) {
	defer tracing.NewRegion("terrain.NewAsset").End()
	config = normalizeConfig(config)
	expected := config.Resolution * config.Resolution
	if heights == nil {
		heights = make([]matrix.Float, expected)
		for i := range heights {
			heights[i] = config.InitialHeight
		}
	}
	if len(heights) != expected {
		return TerrainAsset{}, fmt.Errorf("terrain asset expected %d heights, got %d", expected, len(heights))
	}
	layers, paintResolution, weights, err := normalizedAssetPaintData(config, layerSet)
	if err != nil {
		return TerrainAsset{}, err
	}
	config.PaintResolution = paintResolution
	config.Textures = terrainTexturesFromLayers(layers)
	asset := TerrainAsset{
		Version:             AssetVersion,
		Config:              config,
		Heights:             make([]uint16, len(heights)),
		Layers:              layers,
		WeightMapResolution: paintResolution,
		Weights:             weights,
	}
	for i := range heights {
		asset.Heights[i] = normalizeHeightToUint16(heights[i], config.MinHeight, config.MaxHeight)
	}
	return asset, nil
}

func NewAssetFromHeightField(config TerrainConfig, field *HeightField) (TerrainAsset, error) {
	defer tracing.NewRegion("terrain.NewAssetFromHeightField").End()
	if field == nil {
		return NewAsset(config, nil)
	}
	config.Resolution = field.Resolution
	config.MinHeight = field.MinHeight
	config.MaxHeight = field.MaxHeight
	return NewAsset(config, field.Heights)
}

func NewAssetFromTerrain(model *Terrain) (TerrainAsset, error) {
	defer tracing.NewRegion("terrain.NewAssetFromTerrain").End()
	if model == nil || model.HeightField == nil {
		return TerrainAsset{}, errors.New("terrain asset requires a terrain model")
	}
	config := model.Config
	config.Resolution = model.HeightField.Resolution
	config.MinHeight = model.HeightField.MinHeight
	config.MaxHeight = model.HeightField.MaxHeight
	return NewAssetWithLayerSet(config, model.HeightField.Heights, model.LayerSet)
}

func LoadAsset(assetDb assets.Database, id string) (TerrainAsset, error) {
	defer tracing.NewRegion("terrain.LoadAsset").End()
	data, err := assetDb.Read(id)
	if err != nil {
		return TerrainAsset{}, err
	}
	return DeserializeAsset(data)
}

func Load(host *engine.Host, id string) (*Terrain, error) {
	defer tracing.NewRegion("terrain.Load").End()
	asset, err := LoadAsset(host.AssetDatabase(), id)
	if err != nil {
		return nil, err
	}
	return NewFromAsset(host, asset)
}

func LoadForEntity(host *engine.Host, id string, entity *engine.Entity) (*Terrain, error) {
	defer tracing.NewRegion("terrain.LoadForEntity").End()
	asset, err := LoadAsset(host.AssetDatabase(), id)
	if err != nil {
		return nil, err
	}
	return NewFromAssetForEntity(host, asset, entity)
}

func NewFromAsset(host *engine.Host, asset TerrainAsset) (*Terrain, error) {
	defer tracing.NewRegion("terrain.NewFromAsset").End()
	return newTerrainFromAsset(host, asset, nil)
}

func NewModelFromAsset(asset TerrainAsset) (*Terrain, error) {
	defer tracing.NewRegion("terrain.NewModelFromAsset").End()
	return newTerrainFromAsset(nil, asset, nil)
}

func NewFromAssetForEntity(host *engine.Host, asset TerrainAsset, entity *engine.Entity) (*Terrain, error) {
	defer tracing.NewRegion("terrain.NewFromAssetForEntity").End()
	if entity == nil {
		return nil, errors.New("terrain asset requires an entity")
	}
	return newTerrainFromAsset(host, asset, entity)
}

func (a TerrainAsset) Serialize() ([]byte, error) {
	defer tracing.NewRegion("TerrainAsset.Serialize").End()
	a.Config = normalizeConfig(a.Config)
	a.upgradeLegacyPaintData()
	if err := a.validate(); err != nil {
		return nil, err
	}
	header := terrainAssetHeader{
		Version:             a.Version,
		Config:              a.Config,
		HeightEncoding:      HeightEncodingUint16,
		HeightCount:         len(a.Heights),
		Layers:              a.Layers,
		WeightEncoding:      WeightEncodingUint16,
		WeightMapResolution: a.WeightMapResolution,
		WeightCount:         len(a.Weights),
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
	for i := range a.Heights {
		if err := binary.Write(&out, binary.LittleEndian, a.Heights[i]); err != nil {
			return nil, err
		}
	}
	for i := range a.Weights {
		if err := binary.Write(&out, binary.LittleEndian, a.Weights[i]); err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}

func DeserializeAsset(data []byte) (TerrainAsset, error) {
	defer tracing.NewRegion("terrain.DeserializeAsset").End()
	if len(data) > 0 && data[0] == '{' {
		return deserializeJSONAsset(data)
	}
	if len(data) < len(terrainAssetMagic)+4 || !bytes.Equal(data[:len(terrainAssetMagic)], terrainAssetMagic) {
		return TerrainAsset{}, errors.New("invalid terrain asset")
	}
	headerLenStart := len(terrainAssetMagic)
	headerLen := int(binary.LittleEndian.Uint32(data[headerLenStart : headerLenStart+4]))
	headerStart := headerLenStart + 4
	headerEnd := headerStart + headerLen
	if headerLen <= 0 || headerEnd > len(data) {
		return TerrainAsset{}, errors.New("invalid terrain asset header")
	}
	var header terrainAssetHeader
	if err := json.Unmarshal(data[headerStart:headerEnd], &header); err != nil {
		return TerrainAsset{}, err
	}
	if header.HeightEncoding != HeightEncodingUint16 {
		return TerrainAsset{}, fmt.Errorf("unsupported terrain height encoding %q", header.HeightEncoding)
	}
	if header.HeightCount < 0 {
		return TerrainAsset{}, fmt.Errorf("terrain asset height count cannot be negative: %d", header.HeightCount)
	}
	if header.WeightCount < 0 {
		return TerrainAsset{}, fmt.Errorf("terrain asset weight count cannot be negative: %d", header.WeightCount)
	}
	if header.Version > 1 && header.WeightEncoding != WeightEncodingUint16 {
		return TerrainAsset{}, fmt.Errorf("unsupported terrain weight encoding %q", header.WeightEncoding)
	}
	payload := data[headerEnd:]
	heightBytes := header.HeightCount * 2
	if len(payload) < heightBytes {
		return TerrainAsset{}, fmt.Errorf("terrain asset expected at least %d height bytes, got %d", heightBytes, len(payload))
	}
	weightBytes := 0
	if header.Version > 1 {
		weightBytes = header.WeightCount * 2
	}
	expectedPayload := heightBytes + weightBytes
	if len(payload) != expectedPayload {
		return TerrainAsset{}, fmt.Errorf("terrain asset expected %d payload bytes, got %d", expectedPayload, len(payload))
	}
	asset := TerrainAsset{
		Version:             header.Version,
		Config:              normalizeConfig(header.Config),
		Heights:             make([]uint16, header.HeightCount),
		Layers:              append([]TerrainLayer(nil), header.Layers...),
		WeightMapResolution: header.WeightMapResolution,
		Weights:             make([]uint16, header.WeightCount),
	}
	for i := range asset.Heights {
		asset.Heights[i] = binary.LittleEndian.Uint16(payload[i*2 : i*2+2])
	}
	weightsStart := heightBytes
	for i := range asset.Weights {
		asset.Weights[i] = binary.LittleEndian.Uint16(payload[weightsStart+i*2 : weightsStart+i*2+2])
	}
	asset.upgradeLegacyPaintData()
	if err := asset.validate(); err != nil {
		return TerrainAsset{}, err
	}
	return asset, nil
}

func (a TerrainAsset) FloatHeights() []matrix.Float {
	defer tracing.NewRegion("TerrainAsset.FloatHeights").End()
	heights := make([]matrix.Float, len(a.Heights))
	for i := range a.Heights {
		heights[i] = uint16ToHeight(a.Heights[i], a.Config.MinHeight, a.Config.MaxHeight)
	}
	return heights
}

func (a TerrainAsset) Height(x, z int) matrix.Float {
	if x < 0 || z < 0 || x >= a.Config.Resolution || z >= a.Config.Resolution {
		return 0
	}
	return uint16ToHeight(a.Heights[x+z*a.Config.Resolution], a.Config.MinHeight, a.Config.MaxHeight)
}

func (a TerrainAsset) LayerSet() (*TerrainLayerSet, error) {
	defer tracing.NewRegion("TerrainAsset.LayerSet").End()
	a.Config = normalizeConfig(a.Config)
	a.upgradeLegacyPaintData()
	if err := a.validate(); err != nil {
		return nil, err
	}
	weights, err := NewTextureWeightMap(a.WeightMapResolution, len(a.Layers))
	if err != nil {
		return nil, err
	}
	for i := range weights.Weights {
		weights.Weights[i] = uint16ToWeight(a.Weights[i])
	}
	weights.NormalizeAll()
	return &TerrainLayerSet{
		Layers:    append([]TerrainLayer(nil), a.Layers...),
		WeightMap: weights,
	}, nil
}

func (a TerrainAsset) validate() error {
	if a.Version < 1 || a.Version > AssetVersion {
		return fmt.Errorf("unsupported terrain asset version %d", a.Version)
	}
	expected := a.Config.Resolution * a.Config.Resolution
	if a.Config.Resolution < 2 {
		return errors.New("terrain asset resolution must be at least 2")
	}
	if len(a.Heights) != expected {
		return fmt.Errorf("terrain asset expected %d heights, got %d", expected, len(a.Heights))
	}
	if len(a.Layers) == 0 {
		return errors.New("terrain asset requires at least one paint layer")
	}
	if a.WeightMapResolution < 2 {
		return errors.New("terrain asset weight-map resolution must be at least 2")
	}
	if a.Config.PaintResolution != a.WeightMapResolution {
		return fmt.Errorf("terrain asset paint resolution %d does not match weight-map resolution %d", a.Config.PaintResolution, a.WeightMapResolution)
	}
	expectedWeights := a.WeightMapResolution * a.WeightMapResolution * len(a.Layers)
	if len(a.Weights) != expectedWeights {
		return fmt.Errorf("terrain asset expected %d weights, got %d", expectedWeights, len(a.Weights))
	}
	for i := range a.Layers {
		if strings.TrimSpace(a.Layers[i].TextureContentID) == "" {
			return fmt.Errorf("terrain asset layer %d requires a texture content id", i)
		}
		if a.Layers[i].Filter < 0 || a.Layers[i].Filter >= rendering.TextureFilterMax {
			return fmt.Errorf("terrain asset layer %d has unsupported texture filter %d", i, a.Layers[i].Filter)
		}
	}
	return nil
}

func deserializeJSONAsset(data []byte) (TerrainAsset, error) {
	var asset TerrainAsset
	if err := json.Unmarshal(data, &asset); err != nil {
		return TerrainAsset{}, err
	}
	asset.Config = normalizeConfig(asset.Config)
	if asset.Version == 0 {
		asset.Version = 1
	}
	asset.upgradeLegacyPaintData()
	if err := asset.validate(); err != nil {
		return TerrainAsset{}, err
	}
	return asset, nil
}

func newTerrainFromAsset(host *engine.Host, asset TerrainAsset, entity *engine.Entity) (*Terrain, error) {
	asset.Config = normalizeConfig(asset.Config)
	asset.upgradeLegacyPaintData()
	if err := asset.validate(); err != nil {
		return nil, err
	}
	var workGroup *concurrent.WorkGroup
	if host != nil {
		workGroup = host.WorkGroup()
	}
	t, err := newTerrainWithHeights(asset.Config, asset.FloatHeights(), workGroup, nil, entity)
	if err != nil {
		return nil, err
	}
	t.LayerSet, err = asset.LayerSet()
	if err != nil {
		return nil, err
	}
	t.syncConfigTexturesFromLayers()
	if host != nil {
		t.host = host
		if err := t.createRenderResources(host); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func normalizedAssetPaintData(config TerrainConfig, layerSet *TerrainLayerSet) ([]TerrainLayer, int, []uint16, error) {
	if layerSet == nil || layerSet.LayerCount() == 0 || layerSet.WeightMap == nil {
		layerSet = defaultAssetLayerSet(config)
	}
	if layerSet.WeightMap.Resolution < 2 {
		return nil, 0, nil, errors.New("terrain asset weight-map resolution must be at least 2")
	}
	if layerSet.WeightMap.Layers != len(layerSet.Layers) {
		return nil, 0, nil, fmt.Errorf("terrain asset layer count %d does not match weight-map layers %d", len(layerSet.Layers), layerSet.WeightMap.Layers)
	}
	expectedWeights := layerSet.WeightMap.Resolution * layerSet.WeightMap.Resolution * len(layerSet.Layers)
	if len(layerSet.WeightMap.Weights) != expectedWeights {
		return nil, 0, nil, fmt.Errorf("terrain asset expected %d weights, got %d", expectedWeights, len(layerSet.WeightMap.Weights))
	}
	layers := make([]TerrainLayer, len(layerSet.Layers))
	for i := range layerSet.Layers {
		layers[i] = normalizeTerrainLayer(layerSet.Layers[i])
		if strings.TrimSpace(layers[i].TextureContentID) == "" {
			return nil, 0, nil, fmt.Errorf("terrain asset layer %d requires a texture content id", i)
		}
	}
	weights := make([]uint16, len(layerSet.WeightMap.Weights))
	for z := 0; z < layerSet.WeightMap.Resolution; z++ {
		for x := 0; x < layerSet.WeightMap.Resolution; x++ {
			sum := matrix.Float(0)
			for layer := range layers {
				sum += matrix.Clamp(layerSet.WeightMap.WeightAt(layer, x, z), 0, 1)
			}
			for layer := range layers {
				weight := matrix.Clamp(layerSet.WeightMap.WeightAt(layer, x, z), 0, 1)
				if sum <= matrix.Tiny {
					weight = 0
					if layer == 0 {
						weight = 1
					}
				} else {
					weight /= sum
				}
				weights[(x+z*layerSet.WeightMap.Resolution)*len(layers)+layer] = normalizeWeightToUint16(weight)
			}
		}
	}
	return layers, layerSet.WeightMap.Resolution, weights, nil
}

func defaultAssetLayerSet(config TerrainConfig) *TerrainLayerSet {
	config = normalizeConfig(config)
	layer := NewTerrainLayer(config.Textures[0].Key)
	layer.Filter = config.Textures[0].Filter
	set, _ := NewTerrainLayerSet(config.PaintResolution)
	set.AddLayer(layer)
	set.FillLayer(0)
	return set
}

func (a *TerrainAsset) upgradeLegacyPaintData() {
	a.Config = normalizeConfig(a.Config)
	if a.Version == 0 {
		a.Version = 1
	}
	if a.Version > 1 {
		for i := range a.Layers {
			a.Layers[i] = normalizeTerrainLayer(a.Layers[i])
		}
		return
	}
	if len(a.Layers) == 0 {
		layer := NewTerrainLayer(a.Config.Textures[0].Key)
		layer.Filter = a.Config.Textures[0].Filter
		a.Layers = []TerrainLayer{layer}
	}
	for i := range a.Layers {
		a.Layers[i] = normalizeTerrainLayer(a.Layers[i])
	}
	if a.WeightMapResolution < 2 {
		a.WeightMapResolution = a.Config.PaintResolution
	}
	a.Config.PaintResolution = a.WeightMapResolution
	expectedWeights := a.WeightMapResolution * a.WeightMapResolution * len(a.Layers)
	if len(a.Weights) == 0 {
		a.Weights = make([]uint16, expectedWeights)
		for i := 0; i < a.WeightMapResolution*a.WeightMapResolution; i++ {
			a.Weights[i*len(a.Layers)] = normalizeWeightToUint16(1)
		}
	}
	a.Version = AssetVersion
}

func normalizeHeightToUint16(height, minHeight, maxHeight matrix.Float) uint16 {
	if maxHeight <= minHeight {
		return 0
	}
	height = matrix.Clamp(height, minHeight, maxHeight)
	normalized := (height - minHeight) / (maxHeight - minHeight)
	return uint16(matrix.Clamp(normalized*heightU16Max+0.5, 0, heightU16Max))
}

func uint16ToHeight(height uint16, minHeight, maxHeight matrix.Float) matrix.Float {
	if maxHeight <= minHeight {
		return minHeight
	}
	normalized := matrix.Float(height) / heightU16Max
	return matrix.Lerp(minHeight, maxHeight, normalized)
}

func normalizeWeightToUint16(weight matrix.Float) uint16 {
	return uint16(matrix.Clamp(weight, 0, 1)*heightU16Max + 0.5)
}

func uint16ToWeight(weight uint16) matrix.Float {
	return matrix.Float(weight) / heightU16Max
}

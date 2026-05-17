/******************************************************************************/
/* terrain_asset.go                                                           */
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

package terrain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	AssetVersion = 1
	heightU16Max = matrix.Float(65535)
)

var terrainAssetMagic = []byte{'K', 'T', 'R', 'N'}

type HeightEncoding string

const (
	HeightEncodingUint16 HeightEncoding = "uint16-normalized"
)

type TerrainAsset struct {
	Version int
	Config  TerrainConfig
	Heights []uint16
}

type terrainAssetHeader struct {
	Version        int
	Config         TerrainConfig
	HeightEncoding HeightEncoding
	HeightCount    int
}

func NewAsset(config TerrainConfig, heights []matrix.Float) (TerrainAsset, error) {
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
	asset := TerrainAsset{
		Version: AssetVersion,
		Config:  config,
		Heights: make([]uint16, len(heights)),
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
	if err := a.validate(); err != nil {
		return nil, err
	}
	header := terrainAssetHeader{
		Version:        a.Version,
		Config:         a.Config,
		HeightEncoding: HeightEncodingUint16,
		HeightCount:    len(a.Heights),
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
	heightsData := data[headerEnd:]
	if len(heightsData) != header.HeightCount*2 {
		return TerrainAsset{}, fmt.Errorf("terrain asset expected %d height bytes, got %d", header.HeightCount*2, len(heightsData))
	}
	asset := TerrainAsset{
		Version: header.Version,
		Config:  normalizeConfig(header.Config),
		Heights: make([]uint16, header.HeightCount),
	}
	for i := range asset.Heights {
		asset.Heights[i] = binary.LittleEndian.Uint16(heightsData[i*2 : i*2+2])
	}
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

func (a TerrainAsset) validate() error {
	if a.Version != AssetVersion {
		return fmt.Errorf("unsupported terrain asset version %d", a.Version)
	}
	expected := a.Config.Resolution * a.Config.Resolution
	if a.Config.Resolution < 2 {
		return errors.New("terrain asset resolution must be at least 2")
	}
	if len(a.Heights) != expected {
		return fmt.Errorf("terrain asset expected %d heights, got %d", expected, len(a.Heights))
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
		asset.Version = AssetVersion
	}
	if err := asset.validate(); err != nil {
		return TerrainAsset{}, err
	}
	return asset, nil
}

func newTerrainFromAsset(host *engine.Host, asset TerrainAsset, entity *engine.Entity) (*Terrain, error) {
	asset.Config = normalizeConfig(asset.Config)
	if err := asset.validate(); err != nil {
		return nil, err
	}
	var workGroup *concurrent.WorkGroup
	if host != nil {
		workGroup = host.WorkGroup()
	}
	t, err := newTerrainWithHeights(asset.Config, asset.FloatHeights(), workGroup, host, entity)
	if err != nil {
		return nil, err
	}
	return t, nil
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

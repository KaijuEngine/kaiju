/******************************************************************************/
/* content_database_terrain.go                                                */
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

package content_database

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(Terrain{}) }

// Terrain is a [ContentCategory] represented by a ".terrain" asset. The
// stored asset keeps JSON metadata with compact 16-bit normalized height data.
type Terrain struct{}
type TerrainConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Terrain) Path() string       { return project_file_system.ContentTerrainFolder }
func (Terrain) TypeName() string   { return "Terrain" }
func (Terrain) ExtNames() []string { return []string{".terrain", ".raw", ".r16"} }

func (Terrain) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Terrain.Import").End()
	data, err := filesystem.ReadFile(src)
	if err != nil {
		return ProcessedImport{}, err
	}
	ext := strings.ToLower(filepath.Ext(src))
	var asset terrain.TerrainAsset
	if ext == ".terrain" {
		asset, err = terrain.DeserializeAsset(data)
		if err != nil {
			return ProcessedImport{}, err
		}
	} else {
		// .raw or .r16: 16-bit little-endian unsigned height values forming a square map
		if len(data)%2 != 0 {
			return ProcessedImport{}, errors.New("raw terrain file must have even number of bytes for 16-bit heights")
		}
		heightCount := len(data) / 2
		res := int(math.Sqrt(float64(heightCount)))
		if res*res != heightCount || res < 2 {
			return ProcessedImport{}, fmt.Errorf("raw terrain data size %d does not form a square resolution >=2", heightCount)
		}
		heights := make([]uint16, heightCount)
		for i := 0; i < heightCount; i++ {
			heights[i] = binary.LittleEndian.Uint16(data[i*2 : i*2+2])
		}
		asset = terrain.TerrainAsset{
			Version: terrain.AssetVersion,
			Config:  terrain.TerrainConfig{Resolution: res},
			Heights: heights,
		}
	}
	data, err = asset.Serialize()
	if err != nil {
		return ProcessedImport{}, err
	}
	return ProcessedImport{Variants: []ImportVariant{
		{Name: fileNameNoExt(src), Data: data},
	}}, nil
}

func (c Terrain) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Terrain.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Terrain) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}

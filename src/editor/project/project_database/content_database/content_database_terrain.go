/******************************************************************************/
/* content_database_terrain.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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

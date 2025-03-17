/******************************************************************************/
/* mesh_cache.go                                                              */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package project_cache

import (
	"kaiju/engine/assets/asset_info"
	"kaiju/rendering/loaders/load_result"
	"kaiju/engine/runtime/encoding/gob"
	"os"
	"path/filepath"
)

func toCachedMeshPath(path string, adiID string) string {
	return filepath.Join(path, adiID+".msh")
}

func CacheMesh(adiID string, mesh load_result.Mesh) error {
	path := cachePath(meshCache)
	f, err := os.Create(toCachedMeshPath(path, adiID))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(mesh)
}

func LoadCachedMesh(adiID string) (load_result.Mesh, error) {
	path := cachePath(meshCache)
	f, err := os.Open(toCachedMeshPath(path, adiID))
	if err != nil {
		return load_result.Mesh{}, err
	}
	defer f.Close()
	var mesh load_result.Mesh
	dec := gob.NewDecoder(f)
	err = dec.Decode(&mesh)
	return mesh, err
}

func DeleteMesh(adi asset_info.AssetDatabaseInfo) error {
	path := cachePath(meshCache)
	for i := range adi.Children {
		if err := DeleteMesh(adi.Children[i]); err != nil {
			return err
		}
	}
	if err := os.Remove(toCachedMeshPath(path, adi.ID)); err != nil {
		if err != os.ErrNotExist {
			return err
		}
	}
	return nil
}

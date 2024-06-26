/******************************************************************************/
/* obj_importer.go                                                            */
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

package asset_importer

import (
	"errors"
	"kaiju/assets"
	"kaiju/assets/asset_info"
	"kaiju/cache/project_cache"
	"kaiju/editor/editor_config"
	"kaiju/filesystem"
	"kaiju/rendering/loaders"
	"path/filepath"

	"github.com/KaijuEngine/uuid"
)

type OBJImporter struct{}

func (m OBJImporter) Handles(path string) bool {
	return filepath.Ext(path) == editor_config.FileExtensionObj
}

func cleanupOBJ(adi asset_info.AssetDatabaseInfo) {
	project_cache.DeleteMesh(adi)
	adi.Children = adi.Children[:0]
	adi.Metadata = make(map[string]string)
}

func (m OBJImporter) Import(path string) error {
	adi, err := createADI(path, cleanupOBJ)
	if err != nil {
		return err
	}
	adi.Type = editor_config.AssetTypeObj
	if err := importMeshToCache(&adi); err != nil {
		return err
	}
	return asset_info.Write(adi)
}

func importMeshToCache(adi *asset_info.AssetDatabaseInfo) error {
	src, err := filesystem.ReadTextFile(adi.Path)
	if err != nil {
		return err
	}
	res := loaders.OBJ(src)
	if len(res.Meshes) == 0 {
		return errors.New("no meshes found in OBJ file")
	}
	adi.Metadata["name"] = res.Meshes[0].Name
	for _, o := range res.Meshes {
		info := adi.SpawnChild(uuid.New().String())
		info.Type = editor_config.AssetTypeMesh
		info.ParentID = adi.ID
		if err := project_cache.CacheMesh(info, o); err != nil {
			return err
		}
		// TODO:  Write the correct material to the adi
		info.Metadata["shader"] = assets.ShaderDefinitionBasic
		info.Metadata["texture"] = assets.TextureSquare
		info.Metadata["name"] = o.MeshName
		adi.Children = append(adi.Children, info)
	}
	return nil
}

/*****************************************************************************/
/* obj_importer.go                                                           */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package importers

import (
	"errors"
	"kaiju/assets/asset_info"
	"kaiju/editor/cache/project_cache"
	"kaiju/filesystem"
	"kaiju/rendering/loaders"
	"path/filepath"

	"github.com/KaijuEngine/uuid"
)

type OBJImporter struct{}

func (m OBJImporter) Handles(path string) bool {
	return filepath.Ext(path) == ".obj"
}

func (m OBJImporter) Import(path string) error {
	adi, err := asset_info.Read(path)
	if errors.Is(err, asset_info.ErrNoInfo) {
		adi = asset_info.New(path, uuid.New().String())
	} else if err != nil {
		return err
	} else {
		project_cache.DeleteMesh(adi)
		adi.Children = adi.Children[:0]
	}
	adi.Type = ImportTypeObj
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
	for _, o := range res {
		info := adi.SpawnChild(uuid.New().String())
		info.Type = ImportTypeMesh
		info.ParentID = adi.ID
		if err := project_cache.CacheMesh(info, o); err != nil {
			return err
		}
		adi.Children = append(adi.Children, info)
	}
	return nil
}

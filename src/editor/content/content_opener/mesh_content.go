/******************************************************************************/
/* mesh_content.go                                                            */
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

package content_opener

import (
	"kaiju/engine/assets/asset_importer"
	"kaiju/engine/assets/asset_info"
	"kaiju/editor/cache/project_cache"
	"kaiju/engine/collision"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

func loadMesh(host *engine.Host, adi asset_info.AssetDatabaseInfo, e *engine.Entity, bvh *collision.BVH) error {
	var err error
	var data rendering.DrawInstance
	var material *rendering.Material
	meta := adi.Metadata.(*asset_importer.MeshMetadata)
	if meta.Material != "" {
		if material, err = host.MaterialCache().Material(meta.Material); err != nil {
			return err
		}
		// TODO:  We need to create or generate shader data given the definition
		data = &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	} else {
		if material, err = host.MaterialCache().Material("basic"); err != nil {
			return err
		}
		data = &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}
	mesh, ok := host.MeshCache().FindMesh(adi.ID)
	if !ok {
		m, err := project_cache.LoadCachedMesh(adi.ID)
		if err != nil {
			return err
		}
		mesh = rendering.NewMesh(adi.ID, m.Verts, m.Indexes)
		bvh.Insert(m.GenerateBVH(host.Threads()))
	}
	host.MeshCache().AddMesh(mesh)
	drawing := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   material,
		Mesh:       mesh,
		ShaderData: data,
		Transform:  &e.Transform,
	}
	host.Drawings.AddDrawing(drawing)
	e.EditorBindings.AddDrawing(drawing)
	e.OnActivate.Add(func() { data.Activate() })
	e.OnDeactivate.Add(func() { data.Deactivate() })
	e.OnDestroy.Add(func() { data.Destroy() })
	return nil
}

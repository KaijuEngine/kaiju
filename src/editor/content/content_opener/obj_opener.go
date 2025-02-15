/******************************************************************************/
/* obj_opener.go                                                              */
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
	"kaiju/assets"
	"kaiju/assets/asset_info"
	"kaiju/cache/project_cache"
	"kaiju/collision"
	"kaiju/editor/editor_config"
	"kaiju/editor/interfaces"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

type ObjOpener struct{}

func (o ObjOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeObj
}

func load(host *engine.Host, adi asset_info.AssetDatabaseInfo, e *engine.Entity, bvh *collision.BVH) error {
	texId := assets.TextureSquare
	if t, ok := adi.Metadata["texture"]; ok {
		texId = t
	}
	tex, err := host.TextureCache().Texture(texId, rendering.TextureFilterLinear)
	if err != nil {
		return err
	}
	var data rendering.DrawInstance
	var shader *rendering.Shader
	if s, ok := adi.Metadata["shader"]; ok {
		shader = host.ShaderCache().ShaderFromDefinition(s)
		// TODO:  We need to create or generate shader data given the definition
		data = &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	} else {
		shader = host.ShaderCache().ShaderFromDefinition(
			assets.ShaderDefinitionBasic)
		data = &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}
	mesh, ok := host.MeshCache().FindMesh(adi.ID)
	if !ok {
		m, err := project_cache.LoadCachedMesh(adi)
		if err != nil {
			return err
		}
		mesh = rendering.NewMesh(adi.ID, m.Verts, m.Indexes)
		bvh.Insert(m.GenerateBVH(&e.Transform))
	}
	host.MeshCache().AddMesh(mesh)
	drawing := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Shader:     shader,
		Mesh:       mesh,
		Textures:   []*rendering.Texture{tex},
		ShaderData: data,
		Transform:  &e.Transform,
		CanvasId:   "default",
	}
	host.Drawings.AddDrawing(&drawing)
	e.EditorBindings.AddDrawing(drawing)
	e.OnActivate.Add(func() { data.Activate() })
	e.OnDeactivate.Add(func() { data.Deactivate() })
	e.OnDestroy.Add(func() { data.Destroy() })
	return nil
}

func (o ObjOpener) Open(adi asset_info.AssetDatabaseInfo, ed interfaces.Editor) error {
	host := ed.Host()
	e := engine.NewEntity(ed.Host().WorkGroup())
	e.GenerateId()
	host.AddEntity(e)
	e.SetName(adi.MetaValue("name"))
	bvh := collision.NewBVH()
	bvh.Transform = &e.Transform
	for i := range adi.Children {
		if err := load(host, adi.Children[i], e, bvh); err != nil {
			return err
		}
	}
	if !bvh.IsLeaf() {
		e.EditorBindings.Set("bvh", bvh)
		ed.BVH().Insert(bvh)
		e.OnDestroy.Add(func() { bvh.RemoveNode() })
	}
	ed.History().Add(&modelOpenHistory{
		host:   host,
		entity: e,
	})
	ed.Hierarchy().Reload()
	host.Window.Focus()
	return nil
}

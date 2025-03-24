/******************************************************************************/
/* image_opener.go                                                            */
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
	"errors"
	"kaiju/editor/content/content_history"
	"kaiju/editor/editor_config"
	"kaiju/editor/editor_interface"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/assets/asset_importer"
	"kaiju/engine/assets/asset_info"
	"kaiju/engine/collision"
	"kaiju/engine/systems/console"
	"kaiju/matrix"
	"kaiju/rendering"
	"path/filepath"
	"strings"
)

type ImageOpener struct{}

func (o ImageOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeImage
}

func (o ImageOpener) Open(adi asset_info.AssetDatabaseInfo, ed editor_interface.Editor) error {
	console.For(ed.Host()).Write("Opening an image")
	host := ed.Host()
	meta := adi.Metadata.(*asset_importer.ImageMetadata)
	texture, err := host.TextureCache().Texture(adi.Path, meta.ImageFilterMeta())
	if err != nil {
		return errors.New("failed to load the texture " + adi.Path)
	}
	texture.MipLevels = int(meta.Mipmaps)
	// TODO:  Swap this to sprite and remove the visuals2d sprite stuff and make it 3D
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		return errors.New("failed to find the sprite material")
	}
	mesh := rendering.NewMeshQuadAnchored(meta.ImagePivotMeta(), host.MeshCache())
	ppu := matrix.Float(max(1, meta.PixelsPerUnit))
	scaled := matrix.Vec2{
		matrix.Float(texture.Width) / ppu,
		matrix.Float(texture.Height) / ppu,
	}
	e := engine.NewEntity(ed.Host().WorkGroup())
	e.GenerateId()
	p := ed.Camera().LookAtPoint()
	p.SetZ(0)
	e.Transform.SetPosition(p)
	e.Transform.SetScale(matrix.NewVec3(scaled.X(), scaled.Y(), 1))
	host.AddEntity(e)
	e.SetName(strings.TrimSuffix(filepath.Base(adi.Path), filepath.Ext(adi.Path)))
	// TODO:  Swap this to a sprite basic that has control over UVs
	shaderData := &rendering.ShaderDataBasic{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	drawing := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   mat.CreateInstance([]*rendering.Texture{texture}),
		Mesh:       mesh,
		ShaderData: shaderData,
		Transform:  &e.Transform,
	}
	host.Drawings.AddDrawing(drawing)
	e.EditorBindings.AddDrawing(drawing)
	e.OnActivate.Add(func() { shaderData.Activate() })
	e.OnDeactivate.Add(func() { shaderData.Deactivate() })
	e.OnDestroy.Add(func() { shaderData.Destroy() })
	bvh := collision.NewBVH()
	bvh.Transform = &e.Transform
	bvh.Insert(mesh.BVH())
	if !bvh.IsLeaf() {
		e.EditorBindings.Set("bvh", bvh)
		ed.BVH().Insert(bvh)
		e.OnDestroy.Add(func() { bvh.RemoveNode() })
	}
	ed.History().Add(&content_history.ModelOpen{
		Host:   host,
		Entity: e,
		Editor: ed,
	})
	ed.ReloadTabs("Hierarchy")
	host.Window.Focus()
	return nil
}

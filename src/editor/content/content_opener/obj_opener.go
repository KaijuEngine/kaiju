/*****************************************************************************/
/* obj_opener.go                                                             */
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

package content_opener

import (
	"kaiju/assets"
	"kaiju/assets/asset_info"
	"kaiju/editor/cache/project_cache"
	"kaiju/editor/editor_config"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/matrix"
	"kaiju/rendering"
)

type ObjOpener struct{}

func (o ObjOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeObj
}

func load(host *engine.Host, adi asset_info.AssetDatabaseInfo) error {
	m, err := project_cache.LoadCachedMesh(adi)
	if err != nil {
		return err
	}
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
	mesh := rendering.NewMesh(adi.ID, m.Verts, m.Indexes)
	host.MeshCache().AddMesh(mesh)
	host.Drawings.AddDrawing(&rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Shader:     shader,
		Mesh:       mesh,
		Textures:   []*rendering.Texture{tex},
		ShaderData: data,
	}, host.Window.Renderer.DefaultTarget())
	return nil
}

func (o ObjOpener) Open(adi asset_info.AssetDatabaseInfo, container *host_container.Container) error {
	host := container.Host
	for i := range adi.Children {
		if err := load(host, adi.Children[i]); err != nil {
			return err
		}
	}
	container.Host.Window.Focus()
	return nil
}

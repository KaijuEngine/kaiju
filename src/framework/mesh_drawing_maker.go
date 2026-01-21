/******************************************************************************/
/* mesh_drawing_maker.go                                                      */
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

package framework

import (
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
)

func createDrawingFromMeshUnlit(host *engine.Host, mesh *rendering.Mesh, textures []*rendering.Texture, isTransparent bool) (rendering.Drawing, error) {
	var mat *rendering.Material
	var err error
	if isTransparent {
		mat, err = host.MaterialCache().Material(unlitTransparentMaterialKey)
	} else {
		mat, err = host.MaterialCache().Material(unlitMaterialKey)
	}
	if err != nil {
		return rendering.Drawing{}, err
	}
	mat = mat.CreateInstance(textures)
	return rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ViewCuller: &host.Cameras.Primary,
		ShaderData: &shader_data_registry.ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			UVs:            matrix.NewVec4(0, 0, 1, 1),
		},
	}, nil
}

func CreateDrawingFromMeshUnlit(host *engine.Host, mesh *rendering.Mesh, textures []*rendering.Texture) (rendering.Drawing, error) {
	return createDrawingFromMeshUnlit(host, mesh, textures, false)
}

func CreateDrawingFromMeshUnlitTransparent(host *engine.Host, mesh *rendering.Mesh, textures []*rendering.Texture) (rendering.Drawing, error) {
	return createDrawingFromMeshUnlit(host, mesh, textures, true)
}

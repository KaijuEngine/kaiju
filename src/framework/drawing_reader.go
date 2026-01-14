/******************************************************************************/
/* drawing_reader.go                                                          */
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
	"fmt"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"kaiju/rendering/loaders/load_result"
)

const pbrMaterialKey = assets.MaterialDefinitionPBR
const basicMaterialKey = assets.MaterialDefinitionBasic
const unlitMaterialKey = assets.MaterialDefinitionUnlit
const unlitTransparentMaterialKey = assets.MaterialDefinitionUnlitTransparent

type ModelDrawing struct {
	Node     *load_result.Node
	MeshName string
	Drawing  rendering.Drawing
}

type ModelDrawingSlice []ModelDrawing

func (s ModelDrawingSlice) AllForNode(node *load_result.Node) []ModelDrawing {
	defer tracing.NewRegion("framework.AllForNode").End()
	part := []ModelDrawing{}
	for i := range s {
		if s[i].Node == node {
			part = append(part, s[i])
		}
	}
	return part
}
func (s ModelDrawingSlice) AllDrawings() []rendering.Drawing {
	defer tracing.NewRegion("framework.AllDrawings").End()
	drawings := make([]rendering.Drawing, len(s))
	for i := range s {
		drawings[i] = s[i].Drawing
	}
	return drawings
}

func createDrawings(host *engine.Host, res load_result.Result, materialKey string, minimumTextures int, shaderData func() rendering.DrawInstance) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.createDrawings").End()
	drawings := ModelDrawingSlice{}
	for i := range res.Meshes {
		m := res.Meshes[i]
		matKey := materialKey
		if matVal, ok := m.Node.Attributes["material"]; ok {
			if mat, ok := matVal.(string); ok {
				matKey = mat
			}
		}
		var tForm matrix.Transform
		tForm.Initialize(host.WorkGroup())
		tForm.SetLocalPosition(m.Node.Position)
		tForm.SetRotation(m.Node.Rotation.ToEuler())
		tForm.SetScale(m.Node.Scale)
		mesh, ok := host.MeshCache().FindMesh(m.MeshName)
		if !ok {
			mesh = rendering.NewMesh(m.MeshName, m.Verts, m.Indexes)
			host.MeshCache().AddMesh(mesh)
		}
		textures := []*rendering.Texture{}
		for i := range m.Textures {
			tex, _ := host.TextureCache().Texture(m.Textures[i], rendering.TextureFilterLinear)
			textures = append(textures, tex)
		}
		for i := len(textures); i < minimumTextures; i++ {
			tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
			textures = append(textures, tex)
		}
		mat, err := host.MaterialCache().Material(matKey)
		if err != nil {
			return drawings, err
		}
		mat = mat.CreateInstance(textures)
		drawings = append(drawings, ModelDrawing{
			Node:     m.Node,
			MeshName: m.Name,
			Drawing: rendering.Drawing{
				Material:   mat,
				Mesh:       mesh,
				Transform:  &tForm,
				ViewCuller: &host.Cameras.Primary,
				ShaderData: shaderData(),
			},
		})
	}
	if len(drawings) == 0 {
		return drawings, fmt.Errorf("no drawings to load from the mesh load result")
	}
	return drawings, nil
}

func CreateDrawingsUnlit(host *engine.Host, res load_result.Result) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawingsUnlit").End()
	return createDrawings(host, res, unlitMaterialKey, 1, func() rendering.DrawInstance {
		return &shader_data_registry.ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			UVs:            matrix.NewVec4(0, 0, 1, 1),
		}
	})
}

func CreateDrawingsUnlitTransparent(host *engine.Host, res load_result.Result) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawingsUnlitTransparent").End()
	return createDrawings(host, res, unlitTransparentMaterialKey, 1, func() rendering.DrawInstance {
		return &shader_data_registry.ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			UVs:            matrix.NewVec4(0, 0, 1, 1),
		}
	})
}

func CreateDrawingsBasic(host *engine.Host, res load_result.Result) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawingsBasic").End()
	return createDrawings(host, res, basicMaterialKey, 1, func() rendering.DrawInstance {
		return &shader_data_registry.ShaderDataStandard{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	})
}

func CreateDrawingsPBR(host *engine.Host, res load_result.Result) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawingsPBR").End()
	drawings, err := createDrawings(host, res, pbrMaterialKey, 4, func() rendering.DrawInstance {
		return &shader_data_registry.ShaderDataPBR{
			ShaderDataBase: rendering.NewShaderDataBase(),
			VertColors:     matrix.ColorWhite(),
			Metallic:       0,
			Roughness:      1,
			Emissive:       0,
			LightIds:       [...]int32{0, 0, 0, 0},
		}
	})
	for i := range drawings {
		drawings[i].Drawing.Material.CastsShadows = true
		drawings[i].Drawing.Material.ReceivesShadows = true
		drawings[i].Drawing.Material.IsLit = true
	}
	return drawings, err
}

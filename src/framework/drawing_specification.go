/******************************************************************************/
/* drawing_specification.go                                                   */
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

package framework

import (
	"fmt"
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/loaders/load_result"
)

var drawingSpecifications = map[string]DrawingSpecification{}

type DrawingSpecification struct {
	Name       string
	RenderInfo []DrawingSpecificationRenderInfo
}

type DrawingSpecificationRenderInfo struct {
	Material    string
	DataFactory func() rendering.DrawInstance
}

type DrawingSpecCreateInfo struct {
	LoadedMesh    load_result.Result
	Meshes        []*rendering.Mesh
	Textures      []*rendering.Texture
	SkipMatSearch bool
}

func RegisterDrawingSpecification(spec DrawingSpecification) {
	drawingSpecifications[spec.Name] = spec
}

func FindDrawingSpecification(name string) DrawingSpecification {
	return drawingSpecifications[name]
}

func (s DrawingSpecification) IsValid() bool { return s.Name != "" }

func (s DrawingSpecification) CreateDrawings(host *engine.Host, info DrawingSpecCreateInfo) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawings").End()
	debug.Ensure(s.IsValid())
	drawings := ModelDrawingSlice{}
	if info.LoadedMesh.IsValid() {
		for i := range info.LoadedMesh.Meshes {
			m := info.LoadedMesh.Meshes[i]
			matKey := ""
			if !info.SkipMatSearch {
				if matVal, ok := m.Node.Attributes["material"]; ok {
					if mat, ok := matVal.(string); ok {
						matKey = mat
					}
				}
			}
			if matKey != "" {
				if subSpec, ok := drawingSpecifications[matKey]; ok {
					cpy := info
					cpy.LoadedMesh.Meshes = cpy.LoadedMesh.Meshes[i : i+1]
					cpy.SkipMatSearch = true
					ds, err := subSpec.CreateDrawings(host, cpy)
					if err != nil {
						return drawings, err
					}
					drawings = append(drawings, ds...)
					continue
				}
			}
			tForm := matrix.NewTransform()
			tForm.SetPosition(m.Node.Transform.WorldPosition())
			tForm.SetRotation(m.Node.Transform.WorldRotation())
			tForm.SetScale(m.Node.Transform.WorldScale())
			mesh, ok := host.MeshCache().FindMesh(m.MeshName)
			if !ok {
				mesh = rendering.NewMesh(m.MeshName, m.Verts, m.Indexes)
				host.MeshCache().AddMesh(mesh)
			}
			textures := []*rendering.Texture{}
			for j := range m.Textures {
				tex, _ := host.TextureCache().Texture(m.Textures[j], rendering.TextureFilterLinear)
				textures = append(textures, tex)
			}
			for j := range s.RenderInfo {
				mat, err := host.MaterialCache().Material(s.RenderInfo[j].Material)
				if err != nil {
					return drawings, err
				}
				for k := len(textures); k < len(mat.Textures); k++ {
					tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
					textures = append(textures, tex)
				}
				for i := range textures {
					textures[i].MipLevels = 1
				}
				mat = mat.CreateInstance(textures[0:len(mat.Textures)])
				drawings = append(drawings, ModelDrawing{
					Node:     m.Node,
					MeshName: m.Name,
					Drawing: rendering.Drawing{
						Renderer:   host.Window.Renderer,
						Material:   mat,
						Mesh:       mesh,
						Transform:  &tForm,
						ViewCuller: &host.Cameras.Primary,
						ShaderData: s.RenderInfo[j].DataFactory(),
					},
				})
			}
		}
	} else {
		drawings = make([]ModelDrawing, 0, len(s.RenderInfo)*len(info.Meshes))
		for i := range s.RenderInfo {
			var mat *rendering.Material
			var err error
			mat, err = host.MaterialCache().Material(s.RenderInfo[i].Material)
			if err != nil {
				return drawings, err
			}
			mat = mat.CreateInstance(info.Textures)
			for j := range info.Meshes {
				drawings = append(drawings, ModelDrawing{
					Node:     nil,
					MeshName: info.Meshes[j].Key(),
					Drawing: rendering.Drawing{
						Renderer:   host.Window.Renderer,
						Material:   mat,
						Mesh:       info.Meshes[j],
						ViewCuller: &host.Cameras.Primary,
						ShaderData: s.RenderInfo[i].DataFactory(),
					},
				})
			}
		}
	}
	if len(drawings) == 0 {
		return drawings, fmt.Errorf("no drawings to load from the mesh load result")
	}
	return drawings, nil
}

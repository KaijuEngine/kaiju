/******************************************************************************/
/* terrain.go                                                                 */
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

package terrain

import (
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/loaders"
)

const (
	textureCount       = 6
	meshKey            = "terrain"
	meshPath           = "heightmap.gltf"
	defaultHeightScale = 100
)

type Terrain struct {
	Entity      *engine.Entity
	heights     []byte
	width       int
	height      int
	heightScale float32
	drawing     rendering.Drawing
}

func Textures(splat, normal, rock, rockNml, ground, groundNml string) [textureCount]string {
	return [textureCount]string{
		splat, normal, rock, rockNml, ground, groundNml,
	}
}

func New(host *engine.Host, size float32, textures [textureCount]string) (Terrain, error) {
	defer tracing.NewRegion("terrain.New").End()
	var err error
	tMap := Terrain{
		Entity:      host.NewEntity(host.WorkGroup()),
		heightScale: defaultHeightScale,
	}
	if err = tMap.createDrawing(host, textures); err != nil {
		return tMap, err
	}
	tMap.Entity.Transform.SetScale(matrix.NewVec3(size, 1, size))
	splatKey := textures[0]
	data, err := rendering.TexturePixelsFromAsset(host.AssetDatabase(), splatKey)
	tMap.heights = make([]byte, len(data.Mem)/4)
	for i := range len(data.Mem) / 4 {
		tMap.heights[i] = data.Mem[i*4]
	}
	tMap.width = data.Width
	tMap.height = data.Height
	return tMap, nil
}

func (t *Terrain) Destroy() {
	defer tracing.NewRegion("terrain.Destroy").End()
	t.Entity.Destroy()
}

func (t *Terrain) SetScale(scale float32) {
	t.heightScale = scale
	t.drawing.ShaderData.(*TerrainShaderData).heightScalar = scale
}

func (t *Terrain) sampleValue(x, y int) float32 {
	if x < 0 || y < 0 || x >= t.width || y >= t.height {
		return 0
	}
	idx := x + y*t.width
	return (float32(t.heights[idx]) / 255.0) * t.heightScale
}

func (t *Terrain) Height(point matrix.Vec3) float32 {
	defer tracing.NewRegion("terrain.Height").End()
	p := t.Entity.Transform.Position()
	s := t.Entity.Transform.Scale()
	left := p.X() - s.X()*0.5
	right := p.X() + s.X()*0.5
	front := p.Z() - s.Z()*0.5
	back := p.Z() + s.Z()*0.5
	hit := point.X() >= left && point.X() <= right && point.Z() >= front && point.Z() <= back
	if !hit {
		return 0
	}
	fx := float32(t.width) * (point.X() - left) / s.X()
	fy := float32(t.height) * (point.Z() - front) / s.Z()
	x := int(matrix.Floor(fx))
	y := int(matrix.Floor(fy))
	// Clamp indices to ensure they stay within the grid (0 to width-1, 0 to height-1)
	ix := min(x, t.width-1)
	iy := min(y, t.height-1)
	// Compute fractional offsets for interpolation
	dx := fx - float32(ix)
	dy := fy - float32(iy)
	// Get heights at the four surrounding grid points
	bottomLeft := t.sampleValue(x, y)    // Bottom-left
	bottomRight := t.sampleValue(x+1, y) // Bottom-right
	topLeft := t.sampleValue(x, y+1)     // Top-left
	topRight := t.sampleValue(x+1, y+1)  // Top-right
	// Bilinear interpolation
	return (1-dx)*(1-dy)*bottomLeft + dx*(1-dy)*bottomRight + (1-dx)*dy*topLeft + dx*dy*topRight
}

func (t *Terrain) createDrawing(host *engine.Host, textures [textureCount]string) error {
	defer tracing.NewRegion("terrain.createDrawing").End()
	var mat *rendering.Material
	var err error
	mat, err = host.MaterialCache().Material(assets.MaterialDefinitionTerrain)
	if err != nil {
		return err
	}
	meshCache := host.MeshCache()
	mesh, ok := meshCache.FindMesh(meshKey)
	if !ok {
		res, err := loaders.GLTF(meshPath, host.AssetDatabase())
		if err != nil {
			return err
		}
		debug.Ensure(len(res.Meshes) == 1)
		m := &res.Meshes[0]
		mesh = meshCache.Mesh(meshKey, m.Verts, m.Indexes)
		meshCache.AddMesh(mesh)
	}
	texs := [textureCount]*rendering.Texture{}
	for i := range textures {
		texs[i], err = host.TextureCache().Texture(textures[i], rendering.TextureFilterLinear)
		if err != nil {
			return err
		}
	}
	mat = mat.CreateInstance(texs[:])
	mat.IsLit = true
	//mat.ShadowMap = host.Lights()[0].ShadowMapTexture()
	t.drawing = rendering.Drawing{
		Material:  mat,
		Mesh:      mesh,
		Transform: &t.Entity.Transform,
		ShaderData: &TerrainShaderData{
			ShaderDataBase: rendering.NewShaderDataBase(),
			heightScalar:   defaultHeightScale,
		},
		ViewCuller: &host.Cameras.Primary,
		//CastsShadows: true,
	}
	t.Entity.OnDestroy.Add(func() { t.drawing.ShaderData.Destroy() })
	host.Drawings.AddDrawing(t.drawing)
	return nil
}

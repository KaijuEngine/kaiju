/******************************************************************************/
/* shader_data_terrain.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_data_registry

import (
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const (
	TerrainLayerCount          = 8
	TerrainLayerParameters     = 3
	TerrainLayerParameterCount = TerrainLayerCount * TerrainLayerParameters
)

func init() {
	register(func() rendering.DrawInstance {
		return &ShaderDataTerrain{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			UVs:            matrix.NewVec4(0, 0, 1, 1),
			BrushColor:     matrix.NewColor(0.2, 0.75, 1.0, 1.0),
			BrushParams:    matrix.NewVec4(0.15, 0.18, 0.85, 0),
			LightIds:       [...]int32{-1, -1, -1, -1},
		}
	}, "terrain", "terrain_lit", "terrain_unlit", "heightScalar")
}

type ShaderDataTerrain struct {
	rendering.ShaderDataBase `visible:"false"`

	Color             matrix.Color
	UVs               matrix.Vec4 `default:"0,0,1,1"`
	BrushCenterRadius matrix.Vec4 `visible:"false"`
	BrushParams       matrix.Vec4 `visible:"false" default:"0.15,0.18,0.85,0"`
	BrushColor        matrix.Color
	Flags             StandardShaderDataFlags `visible:"false"`
	LightIds          [4]int32                `visible:"false"`

	LayerParameters [TerrainLayerParameterCount]matrix.Vec4 `visible:"false"`
}

func (t ShaderDataTerrain) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderDataTerrain{}.Color) +
		unsafe.Sizeof(ShaderDataTerrain{}.UVs) +
		unsafe.Sizeof(ShaderDataTerrain{}.BrushCenterRadius) +
		unsafe.Sizeof(ShaderDataTerrain{}.BrushParams) +
		unsafe.Sizeof(ShaderDataTerrain{}.BrushColor) +
		unsafe.Sizeof(ShaderDataTerrain{}.Flags) +
		unsafe.Sizeof(ShaderDataTerrain{}.LightIds))
}

func (t *ShaderDataTerrain) SelectLights(lights rendering.LightsForRender) {
	selectPBRLights(&t.ShaderDataBase, &t.LightIds, lights)
}

func (t *ShaderDataTerrain) InstanceBoundDataSize() int {
	return int(unsafe.Sizeof(t.LayerParameters))
}

func (t *ShaderDataTerrain) BoundDataPointer() unsafe.Pointer {
	return unsafe.Pointer(&t.LayerParameters[0])
}

func (t *ShaderDataTerrain) UpdateBoundData() bool { return true }

func (t *ShaderDataTerrain) SetLayerParameters(parameters [TerrainLayerParameterCount]matrix.Vec4) {
	t.LayerParameters = parameters
}

func (t *ShaderDataTerrain) SetBrush(centerXZ matrix.Vec2, radius, ringWidth matrix.Float, color matrix.Color) {
	t.BrushCenterRadius = matrix.NewVec4(centerXZ.X(), centerXZ.Y(), radius, 1)
	t.BrushParams.SetX(max(ringWidth, matrix.Float(0.001)))
	t.BrushColor = color
}

func (t *ShaderDataTerrain) ClearBrush() {
	t.BrushCenterRadius.SetW(0)
}

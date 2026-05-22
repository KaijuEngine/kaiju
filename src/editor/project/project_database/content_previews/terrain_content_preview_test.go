/******************************************************************************/
/* terrain_content_preview_test.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_previews

import (
	"image"
	"testing"

	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/matrix"
)

func TestTerrainPreviewImageUsesPaintedLayerWeights(t *testing.T) {
	cfg := terrain.TerrainConfig{
		Resolution:      4,
		PaintResolution: 4,
		WorldSize:       matrix.NewVec2(4, 4),
		MinHeight:       0,
		MaxHeight:       1,
	}
	set, err := terrain.NewTerrainLayerSet(cfg.PaintResolution)
	if err != nil {
		t.Fatal(err)
	}
	red := set.AddLayer(terrain.TerrainLayer{
		Name:             "Red",
		TextureContentID: "red",
		Tint:             matrix.ColorRed(),
	})
	blue := set.AddLayer(terrain.TerrainLayer{
		Name:             "Blue",
		TextureContentID: "blue",
		Tint:             matrix.ColorBlue(),
	})
	for z := 0; z < set.WeightMap.Resolution; z++ {
		for x := 0; x < set.WeightMap.Resolution; x++ {
			if x < set.WeightMap.Resolution/2 {
				set.SetLayerWeightAt(red, x, z, 1)
				set.SetLayerWeightAt(blue, x, z, 0)
			} else {
				set.SetLayerWeightAt(red, x, z, 0)
				set.SetLayerWeightAt(blue, x, z, 1)
			}
		}
	}
	asset, err := terrain.NewAssetWithLayerSet(cfg, nil, set)
	if err != nil {
		t.Fatal(err)
	}
	img := terrainPreviewImage(asset).(*image.RGBA)
	left := img.RGBAAt(0, 1)
	right := img.RGBAAt(img.Bounds().Dx()-1, 1)
	if left.R <= left.B {
		t.Fatalf("expected left painted preview pixel to be red-dominant, got %+v", left)
	}
	if right.B <= right.R {
		t.Fatalf("expected right painted preview pixel to be blue-dominant, got %+v", right)
	}
}

package terrain

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
)

func TestBuildTerrainAtlasPixelsPacksMapsAndGutters(t *testing.T) {
	db := assets.NewMockDB(map[string][]byte{
		"albedo.png": atlasTestPNG(t, 2, 2, []color.RGBA{
			{R: 10, G: 20, B: 30, A: 255}, {R: 40, G: 50, B: 60, A: 255},
			{R: 70, G: 80, B: 90, A: 255}, {R: 100, G: 110, B: 120, A: 255},
		}),
		"roughness.png": atlasTestPNG(t, 1, 1, []color.RGBA{{R: 64, G: 64, B: 64, A: 255}}),
		"normal.png":    atlasTestPNG(t, 1, 1, []color.RGBA{{R: 130, G: 140, B: 250, A: 255}}),
	})
	layers := []TerrainLayer{{
		TextureContentID: "albedo.png", NormalContentID: "normal.png", RoughnessContentID: "roughness.png",
	}}
	material, normals, width, height, diagnostics, err := buildTerrainAtlasPixels(db, layers, 6, 1)
	if err != nil {
		t.Fatal(err)
	}
	if width != 24 || height != 12 {
		t.Fatalf("atlas dimensions = %dx%d, want 24x12", width, height)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected atlas diagnostics: %+v", diagnostics)
	}
	assertAtlasPixel(t, material, width, 1, 1, [4]byte{10, 20, 30, 64})
	assertAtlasPixel(t, material, width, 0, 0, [4]byte{10, 20, 30, 64})
	assertAtlasPixel(t, material, width, 4, 4, [4]byte{100, 110, 120, 64})
	assertAtlasPixel(t, material, width, 5, 5, [4]byte{100, 110, 120, 64})
	assertAtlasPixel(t, normals, width, 2, 2, [4]byte{130, 140, 250, 255})
	// An unused slot receives the stable material and flat-normal fallbacks.
	assertAtlasPixel(t, material, width, 7, 1, [4]byte{255, 255, 255, 255})
	assertAtlasPixel(t, normals, width, 7, 1, [4]byte{128, 128, 255, 255})
}

func TestBuildTerrainAtlasPixelsReportsAssignedMissingMaps(t *testing.T) {
	db := assets.NewMockDB(map[string][]byte{
		"albedo.png": atlasTestPNG(t, 1, 1, []color.RGBA{{R: 5, G: 6, B: 7, A: 255}}),
	})
	material, normals, width, _, diagnostics, err := buildTerrainAtlasPixels(db, []TerrainLayer{{
		TextureContentID: "albedo.png", NormalContentID: "missing-normal.png",
	}}, 6, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 2 { // missing normal plus unassigned roughness
		t.Fatalf("diagnostic count = %d, want 2", len(diagnostics))
	}
	assertAtlasPixel(t, material, width, 1, 1, [4]byte{5, 6, 7, 255})
	assertAtlasPixel(t, normals, width, 1, 1, [4]byte{128, 128, 255, 255})
}

func TestTerrainMaterialAtlasCacheKeysFollowLayerOrder(t *testing.T) {
	a := TerrainLayer{TextureContentID: "a", NormalContentID: "an", RoughnessContentID: "ar"}
	b := TerrainLayer{TextureContentID: "b", NormalContentID: "bn", RoughnessContentID: "br"}
	materialAB, normalAB := terrainMaterialAtlasCacheKeys([]TerrainLayer{a, b})
	materialAB2, normalAB2 := terrainMaterialAtlasCacheKeys([]TerrainLayer{a, b})
	materialBA, normalBA := terrainMaterialAtlasCacheKeys([]TerrainLayer{b, a})
	if materialAB != materialAB2 || normalAB != normalAB2 {
		t.Fatal("identical layer sets should produce stable atlas cache keys")
	}
	if materialAB == materialBA || normalAB == normalBA {
		t.Fatal("moving a layer should change both atlas cache keys")
	}
}

func TestTerrainShaderLayerDataUsesWorldSizeAndLayerTransforms(t *testing.T) {
	layer := NewTerrainLayer("albedo")
	layer.TextureWorldSize = matrix.NewVec2(2, 5)
	layer.Tiling = matrix.NewVec2(99, 99)
	layer.Offset = matrix.NewVec2(0.25, -0.5)
	layer.Rotation = matrix.Deg2Rad(90)
	layer.Tint = matrix.NewColor(0.1, 0.2, 0.3, 0.4)
	model, err := NewWithLayers(nil, TerrainConfig{
		Resolution: 2, PaintResolution: 2, WorldSize: matrix.NewVec2(20, 10),
	}, []TerrainLayer{layer})
	if err != nil {
		t.Fatal(err)
	}
	params := model.shaderLayerData()
	if !matrix.Vec4Approx(params[0], matrix.NewVec4(10, 2, 0.25, -0.5)) {
		t.Fatalf("scale/offset params = %v", params[0])
	}
	if matrix.Abs(params[1].X()) > 0.001 || !matrix.Approx(params[1].Y(), 1) ||
		!matrix.Approx(params[1].Z(), 0.1) || !matrix.Approx(params[1].W(), 0.2) {
		t.Fatalf("rotation/tint params = %v", params[1])
	}
	if !matrix.Vec4Approx(params[2], matrix.NewVec4(0.3, 0.4, 0, 0)) {
		t.Fatalf("tint tail params = %v", params[2])
	}
}

func TestNewWithLayersRejectsMoreThanRenderedLimit(t *testing.T) {
	layers := make([]TerrainLayer, MaxRenderedLayers+1)
	for i := range layers {
		layers[i] = NewTerrainLayer("layer")
	}
	if _, err := NewWithLayers(nil, TerrainConfig{Resolution: 2, PaintResolution: 2}, layers); err == nil {
		t.Fatal("expected too many rendered layers to fail")
	}
}

func atlasTestPNG(t *testing.T, width, height int, pixels []color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i, pixel := range pixels {
		img.SetRGBA(i%width, i/width, pixel)
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func assertAtlasPixel(t *testing.T, pixels []byte, width, x, y int, want [4]byte) {
	t.Helper()
	index := (x + y*width) * 4
	got := [4]byte{pixels[index], pixels[index+1], pixels[index+2], pixels[index+3]}
	if got != want {
		t.Fatalf("atlas pixel (%d,%d) = %v, want %v", x, y, got, want)
	}
}

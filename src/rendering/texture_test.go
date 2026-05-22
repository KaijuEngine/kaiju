/******************************************************************************/
/* texture_test.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
)

func testPNG(t *testing.T, pixels []color.RGBA, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i, p := range pixels {
		img.Pix[i*4+0] = p.R
		img.Pix[i*4+1] = p.G
		img.Pix[i*4+2] = p.B
		img.Pix[i*4+3] = p.A
	}
	var buff bytes.Buffer
	if err := png.Encode(&buff, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	return buff.Bytes()
}

func TestTextureKeys(t *testing.T) {
	textures := []*Texture{{Key: "albedo"}, {Key: "normal"}, {Key: "roughness"}}
	keys := TextureKeys(textures)
	want := []string{"albedo", "normal", "roughness"}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("TextureKeys[%d] = %q, want %q", i, keys[i], want[i])
		}
	}
}

func TestReadRawTextureDataPNG(t *testing.T) {
	data := ReadRawTextureData(testPNG(t, []color.RGBA{
		{R: 10, G: 20, B: 30, A: 255},
		{R: 50, G: 60, B: 70, A: 255},
	}, 2, 1), TextureFileFormatPng)
	if data.Width != 2 || data.Height != 1 {
		t.Fatalf("PNG dimensions = %dx%d, want 2x1", data.Width, data.Height)
	}
	if data.InputType != TextureFileFormatPng ||
		data.InternalFormat != TextureInputTypeRgba8 ||
		data.Format != TextureColorFormatRgbaUnorm ||
		data.Type != TextureMemTypeUnsignedByte {
		t.Fatalf("unexpected PNG metadata: %+v", data)
	}
	if got := data.Mem[:8]; !bytes.Equal(got, []byte{10, 20, 30, 255, 50, 60, 70, 255}) {
		t.Fatalf("decoded pixels = %v", got)
	}
}

func TestReadRawTextureDataRaw(t *testing.T) {
	mem := []byte{1, 2, 3, 4}
	data := ReadRawTextureData(mem, TextureFileFormatRaw)
	if !bytes.Equal(data.Mem, mem) {
		t.Fatalf("raw data was not passed through")
	}
	if data.Width != 0 || data.Height != 0 ||
		data.InternalFormat != TextureInputTypeRgba8 ||
		data.Format != TextureColorFormatRgbaUnorm ||
		data.Type != TextureMemTypeUnsignedByte {
		t.Fatalf("unexpected raw metadata: %+v", data)
	}
}

func TestReadRawTextureDataASTC(t *testing.T) {
	mem := make([]byte, 20)
	mem[4], mem[5] = 5, 4
	mem[7], mem[8], mem[9] = 0x34, 0x12, 0x00
	mem[10], mem[11], mem[12] = 0x78, 0x56, 0x00
	copy(mem[16:], []byte{9, 8, 7, 6})
	data := ReadRawTextureData(mem, TextureFileFormatAstc)
	if data.InternalFormat != TextureInputTypeCompressedRgbaAstc5x4 {
		t.Fatalf("ASTC format = %v", data.InternalFormat)
	}
	if data.Width != 0x1234 || data.Height != 0x5678 {
		t.Fatalf("ASTC dimensions = %dx%d", data.Width, data.Height)
	}
	if !bytes.Equal(data.Mem, []byte{9, 8, 7, 6}) {
		t.Fatalf("ASTC payload = %v", data.Mem)
	}
}

func TestNewTextureFromMemory(t *testing.T) {
	tex, err := NewTextureFromMemory("raw", []byte{1, 2, 3, 4}, 7, 9, TextureFilterNearest)
	if err != nil {
		t.Fatalf("NewTextureFromMemory returned error: %v", err)
	}
	if tex.Key != "raw" || tex.Width != 7 || tex.Height != 9 || tex.Filter != TextureFilterNearest {
		t.Fatalf("unexpected texture: %+v", tex)
	}
	if tex.pendingData == nil || !bytes.Equal(tex.pendingData.Mem, []byte{1, 2, 3, 4}) {
		t.Fatalf("pending raw data not set correctly")
	}
	a, err := NewTextureFromMemory(GenerateUniqueTextureKey, []byte{1, 2, 3, 4}, 1, 1, TextureFilterLinear)
	if err != nil {
		t.Fatalf("first generated texture returned error: %v", err)
	}
	b, err := NewTextureFromMemory(GenerateUniqueTextureKey, []byte{1, 2, 3, 4}, 1, 1, TextureFilterLinear)
	if err != nil {
		t.Fatalf("second generated texture returned error: %v", err)
	}
	if a.Key == "" || b.Key == "" || a.Key == b.Key {
		t.Fatalf("generated keys should be non-empty and unique: %q %q", a.Key, b.Key)
	}
}

func TestTextureReadPendingDataForTransparency(t *testing.T) {
	tex := Texture{}
	if tex.ReadPendingDataForTransparency() {
		t.Fatalf("nil pending data should not report transparency")
	}
	tex.pendingData = &TextureData{Mem: []byte{255, 1, 1, 255, 255, 2, 2, 255}}
	if tex.ReadPendingDataForTransparency() {
		t.Fatalf("fully opaque data should not report transparency")
	}
	tex.pendingData.Mem[0] = 0
	if tex.ReadPendingDataForTransparency() {
		t.Fatalf("transparency state should not be recalculated after first read")
	}
	tex = Texture{pendingData: &TextureData{Mem: []byte{0, 1, 1, 255}}}
	if !tex.ReadPendingDataForTransparency() {
		t.Fatalf("current transparency scan should find a non-255 first channel")
	}
}

func TestTextureSize(t *testing.T) {
	tex := Texture{Width: 13, Height: 21}
	if got := tex.Size(); got != (matrix.Vec2{13, 21}) {
		t.Fatalf("Size = %v, want 13x21", got)
	}
}

func TestTextureSetPendingDataDimensions(t *testing.T) {
	tex := Texture{}
	tex.SetPendingDataDimensions(TextureDimensionsCube)
	tex.pendingData = &TextureData{}
	tex.SetPendingDataDimensions(TextureDimensionsCube)
	if tex.pendingData.Dimensions != TextureDimensionsCube {
		t.Fatalf("Dimensions = %v, want cube", tex.pendingData.Dimensions)
	}
}

func TestTexturePixelsFromAsset(t *testing.T) {
	pngData := testPNG(t, []color.RGBA{{R: 1, G: 2, B: 3, A: 255}}, 1, 1)
	db := assets.NewMockDB(map[string][]byte{
		"tex.png": pngData,
		"empty":   {},
	})
	data, err := TexturePixelsFromAsset(db, "tex.png")
	if err != nil {
		t.Fatalf("TexturePixelsFromAsset returned error: %v", err)
	}
	if data.Width != 1 || data.Height != 1 || !bytes.Equal(data.Mem[:4], []byte{1, 2, 3, 255}) {
		t.Fatalf("unexpected asset texture data: %+v", data)
	}
	if _, err := TexturePixelsFromAsset(db, "missing.png"); err == nil {
		t.Fatalf("missing asset should return an error")
	}
	if _, err := TexturePixelsFromAsset(db, "empty"); err == nil {
		t.Fatalf("empty asset should return an error")
	}
}

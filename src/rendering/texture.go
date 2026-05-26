/******************************************************************************/
/* texture.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/png"
	"strings"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"

	"github.com/KaijuEngine/uuid"
)

/*
	ASTC notes:
	The header size is found here:  https://github.com/ARM-software/astc-encoder/blob/437f2423fede947a09086f28f547d1897bfe4546/Source/astc_toplevel.cpp#L177

	The following struct denotes it:
	struct astc_header
	{
		uint8_t magic[4];
		uint8_t blockdim_x;
		uint8_t blockdim_y;
		uint8_t blockdim_z;
		uint8_t xsize[3];			// x-size = xsize[0] + xsize[1] + xsize[2]
		uint8_t ysize[3];			// x-size, y-size and z-size are given in texels;
		uint8_t zsize[3];			// block count is inferred
	};
*/

type TextureInputType int
type TextureColorFormat int
type TextureFilter = int
type TextureMemType = int
type TextureFileFormat = int
type TextureDimensions = int

const (
	TextureInputTypeCompressedRgbaAstc4x4 TextureInputType = iota
	TextureInputTypeCompressedRgbaAstc5x4
	TextureInputTypeCompressedRgbaAstc5x5
	TextureInputTypeCompressedRgbaAstc6x5
	TextureInputTypeCompressedRgbaAstc6x6
	TextureInputTypeCompressedRgbaAstc8x5
	TextureInputTypeCompressedRgbaAstc8x6
	TextureInputTypeCompressedRgbaAstc8x8
	TextureInputTypeCompressedRgbaAstc10x5
	TextureInputTypeCompressedRgbaAstc10x6
	TextureInputTypeCompressedRgbaAstc10x8
	TextureInputTypeCompressedRgbaAstc10x10
	TextureInputTypeCompressedRgbaAstc12x10
	TextureInputTypeCompressedRgbaAstc12x12
	TextureInputTypeRgba8
	TextureInputTypeRgb8
	TextureInputTypeLuminance
)

const (
	TextureColorFormatRgbaUnorm TextureColorFormat = iota
	TextureColorFormatRgbUnorm
	TextureColorFormatRgbaSrgb
	TextureColorFormatRgbSrgb
	TextureColorFormatLuminance
)

const (
	TextureFilterLinear TextureFilter = iota
	TextureFilterNearest
	TextureFilterMax
)

const (
	TextureMemTypeUnsignedByte TextureMemType = iota
)

const (
	TextureFileFormatAstc TextureFileFormat = iota
	TextureFileFormatPng
	TextureFileFormatRaw
)

const (
	bytesInPixel = 4
	CubeMapSides = 6
)

const (
	TextureDimensions2 TextureDimensions = iota
	TextureDimensions1
	TextureDimensions3
	TextureDimensionsCube
)

const (
	GenerateUniqueTextureKey = ""
)

type GPUImageWriteRequest struct {
	Region matrix.Vec4i
	Pixels []byte
}

type TextureData struct {
	Mem            []byte
	InternalFormat TextureInputType
	Format         TextureColorFormat
	Type           TextureMemType
	Width          int
	Height         int
	InputType      TextureFileFormat
	Dimensions     TextureDimensions
}

type transparencyReadState int

const (
	transparencyReadStateNone transparencyReadState = iota
	transparencyReadStateRead
	transparencyReadStateFound
)

type Texture struct {
	Key               string
	TexturePixelCache []byte
	RenderId          TextureId
	Channels          int
	Filter            int
	MipLevels         int
	Width             int
	Height            int
	CacheInvalid      bool
	pendingData       *TextureData
	hasTransparency   transparencyReadState
}

func TextureKeys(textures []*Texture) []string {
	defer tracing.NewRegion("rendering.TextureKeys").End()
	keys := make([]string, len(textures))
	for i, t := range textures {
		keys[i] = t.Key
	}
	return keys
}

// ReadRawTextureData reads raw texture data from a byte slice based on the specified input type (ASTC, PNG, or RAW).
// It returns a TextureData struct containing the decoded pixel data, dimensions, and format information.
func ReadRawTextureData(mem []byte, inputType TextureFileFormat) TextureData {
	defer tracing.NewRegion("rendering.ReadRawTextureData").End()

	var res TextureData
	res.InputType = inputType

	astcFormatMap := map[[2]byte]TextureInputType{
		{4, 0}:   TextureInputTypeCompressedRgbaAstc4x4,
		{5, 4}:   TextureInputTypeCompressedRgbaAstc5x4,
		{5, 5}:   TextureInputTypeCompressedRgbaAstc5x5,
		{6, 5}:   TextureInputTypeCompressedRgbaAstc6x5,
		{6, 6}:   TextureInputTypeCompressedRgbaAstc6x6,
		{8, 5}:   TextureInputTypeCompressedRgbaAstc8x5,
		{8, 6}:   TextureInputTypeCompressedRgbaAstc8x6,
		{8, 8}:   TextureInputTypeCompressedRgbaAstc8x8,
		{10, 5}:  TextureInputTypeCompressedRgbaAstc10x5,
		{10, 6}:  TextureInputTypeCompressedRgbaAstc10x6,
		{10, 8}:  TextureInputTypeCompressedRgbaAstc10x8,
		{10, 10}: TextureInputTypeCompressedRgbaAstc10x10,
		{12, 10}: TextureInputTypeCompressedRgbaAstc12x10,
		{12, 12}: TextureInputTypeCompressedRgbaAstc12x12,
	}

	switch inputType {
	case TextureFileFormatAstc:
		key := [2]byte{mem[4], mem[5]}
		if format, ok := astcFormatMap[key]; ok {
			res.InternalFormat = format
		}

		res.Width = int(mem[9])<<16 | int(mem[8])<<8 | int(mem[7])
		res.Height = int(mem[12])<<16 | int(mem[11])<<8 | int(mem[10])

		res.Mem = mem[16:]
		res.Format = TextureColorFormatRgbaUnorm
		res.Type = TextureMemTypeUnsignedByte

	case TextureFileFormatPng:
		img, err := png.Decode(bytes.NewReader(mem))
		if err != nil {
			return res
		}

		b := img.Bounds()
		w, h := b.Dx(), b.Dy()

		if rgba, ok := img.(*image.RGBA); ok {
			res.Mem = rgba.Pix
		} else {
			dst := image.NewRGBA(image.Rect(0, 0, w, h))
			draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
			res.Mem = dst.Pix
		}

		res.Width = w
		res.Height = h
		res.InternalFormat = TextureInputTypeRgba8
		res.Format = TextureColorFormatRgbaUnorm
		res.Type = TextureMemTypeUnsignedByte

	case TextureFileFormatRaw:
		res.Mem = mem
		res.Width = 0
		res.Height = 0
		res.InternalFormat = TextureInputTypeRgba8
		res.Format = TextureColorFormatRgbaUnorm
		res.Type = TextureMemTypeUnsignedByte
	}

	return res
}

func (t *Texture) createData(imgBuff []byte, overrideWidth, overrideHeight int, key string) TextureData {
	inputType := TextureFileFormatRaw
	// TODO:  Use the content system to pull the type from the key
	if strings.HasSuffix(key, ".astc") {
		inputType = TextureFileFormatAstc
	} else if strings.HasSuffix(key, ".png") {
		inputType = TextureFileFormatPng
	} else if len(imgBuff) > 4 && imgBuff[0] == '\x89' && imgBuff[1] == 'P' && imgBuff[2] == 'N' && imgBuff[3] == 'G' {
		inputType = TextureFileFormatPng
	}
	data := ReadRawTextureData(imgBuff, inputType)
	if data.Width == 0 {
		data.Width = overrideWidth
	}
	if data.Height == 0 {
		data.Height = overrideHeight
	}
	return data
}

func (t *Texture) create(imgBuff []byte) {
	data := t.createData(imgBuff, 0, 0, t.Key)
	t.pendingData = &data
	t.Width = data.Width
	t.Height = data.Height
}

func NewTexture(assetDb assets.Database, key string, filter TextureFilter) (*Texture, error) {
	defer tracing.NewRegion("rendering.NewTexture").End()
	key = selectKey(key)
	tex := &Texture{Key: key, Filter: filter}
	if assetDb.Exists(key) {
		if imgBuff, err := assetDb.Read(key); err != nil {
			return nil, err
		} else if len(imgBuff) == 0 {
			return nil, errors.New("no data in texture")
		} else {
			tex.create(imgBuff)
			return tex, nil
		}
	} else {
		return nil, errors.New("texture does not exist")
	}
}

func (t *Texture) Reload(assetDb assets.Database) error {
	t.RenderId = TextureId{}
	if assetDb.Exists(t.Key) {
		if imgBuff, err := assetDb.Read(t.Key); err != nil {
			return err
		} else if len(imgBuff) == 0 {
			return errors.New("no data in texture")
		} else {
			t.create(imgBuff)
			return nil
		}
	}
	return errors.New("texture does not exist")
}

func (t *Texture) triedToReadTransparency() bool {
	return t.hasTransparency != transparencyReadStateNone
}

func (t *Texture) ReadPendingDataForTransparency() bool {
	if t.hasTransparency == transparencyReadStateFound {
		return true
	}
	if t.triedToReadTransparency() || t.pendingData == nil {
		return false
	}
	t.hasTransparency = transparencyReadStateRead
	for i := 0; i < len(t.pendingData.Mem); i += 4 {
		if t.pendingData.Mem[i] != 255 {
			t.hasTransparency = transparencyReadStateFound
			break
		}
	}
	return t.hasTransparency == transparencyReadStateFound
}

func (t *Texture) DelayedCreate(device *GPUDevice) {
	defer tracing.NewRegion("Texture.DelayedCreate").End()
	if t.RenderId.IsValid() {
		return
	}
	device.SetupTexture(t, t.pendingData)
	t.pendingData = nil
}

func NewTextureFromImage(key string, data []byte, filter TextureFilter) (*Texture, error) {
	defer tracing.NewRegion("rendering.NewTextureFromImage").End()
	tex := &Texture{Key: key, Filter: filter}
	tex.create(data)
	return tex, nil
}

func NewTextureFromMemory(key string, data []byte, width, height int, filter TextureFilter) (*Texture, error) {
	defer tracing.NewRegion("rendering.NewTextureFromMemory").End()
	key = selectKey(key)
	tex := &Texture{Key: key, Filter: filter}
	tex.create(data)
	if tex.Width == 0 {
		tex.Width = width
	}
	if tex.Height == 0 {
		tex.Height = height
	}
	return tex, nil
}

func (t *Texture) ReadPixel(app *GPUApplication, x, y int) matrix.Color {
	defer tracing.NewRegion("Texture.ReadPixel").End()
	return app.FirstInstance().PrimaryDevice().TextureReadPixel(t, x, y)
}

func (t *Texture) ReadAllPixels(app *GPUApplication) ([]byte, error) {
	defer tracing.NewRegion("Texture.ReadPixel").End()
	return app.FirstInstance().PrimaryDevice().TextureRead(t)
}

func (t *Texture) ReadPixelRegion(app *GPUApplication, rect matrix.Vec4i) ([]byte, error) {
	defer tracing.NewRegion("Texture.ReadPixelRegion").End()
	return app.FirstInstance().PrimaryDevice().TextureReadRegion(t, rect)
}

func (t *Texture) WritePixels(device *GPUDevice, requests []GPUImageWriteRequest) {
	defer tracing.NewRegion("Texture.WritePixels").End()
	device.TextureWritePixels(t, requests)
}

func (t Texture) Size() matrix.Vec2 {
	return matrix.Vec2{float32(t.Width), float32(t.Height)}
}

func (t *Texture) SetPendingDataDimensions(dim TextureDimensions) {
	if t.pendingData != nil {
		t.pendingData.Dimensions = dim
	}
}

func TexturePixelsFromAsset(assetDb assets.Database, key string) (TextureData, error) {
	defer tracing.NewRegion("rendering.TexturePixelsFromAsset").End()
	key = selectKey(key)
	if assetDb.Exists(key) {
		if imgBuff, err := assetDb.Read(key); err != nil {
			return TextureData{}, err
		} else if len(imgBuff) == 0 {
			return TextureData{}, errors.New("no data in texture")
		} else {
			return ReadRawTextureData(imgBuff, TextureFileFormatPng), nil
		}
	} else {
		return TextureData{}, errors.New("texture does not exist")
	}
}

func selectKey(req string) string {
	if req == GenerateUniqueTextureKey {
		return uuid.NewString()
	}
	return req
}

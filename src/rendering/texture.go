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
	"io"
	"sync"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering/textures"

	"github.com/KaijuEngine/uuid"
)

type GPUImageWriteRequest struct {
	Region matrix.Vec4i
	Pixels []byte
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
	pendingDataMutex  sync.Mutex
	pendingData       *textures.TextureData
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

func readImageTextureData(mem []byte, inputType textures.TextureFileFormat, decode func(io.Reader) (image.Image, error)) textures.TextureData {
	res := textures.TextureData{InputType: inputType}
	img, err := decode(bytes.NewReader(mem))
	if err != nil {
		return res
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if rgba, ok := img.(*image.RGBA); ok {
		res.Mem = rgba.Pix
	} else {
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(dst, dst.Bounds(), img, bounds.Min, draw.Src)
		res.Mem = dst.Pix
	}
	res.Width = width
	res.Height = height
	res.InternalFormat = textures.TextureInputTypeRgba8
	res.Format = textures.TextureColorFormatRgbaUnorm
	res.Type = textures.TextureMemTypeUnsignedByte
	return res
}

func (t *Texture) createData(imgBuff []byte, overrideWidth, overrideHeight int, key string) textures.TextureData {
	inputType := textures.InferFileFormat(key, imgBuff)
	data, _ := textures.ReadImage(imgBuff, inputType)
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
	t.pendingDataMutex.Lock()
	t.pendingData = &data
	t.hasTransparency = transparencyReadStateNone
	t.pendingDataMutex.Unlock()
	t.Width = data.Width
	t.Height = data.Height
}

func NewTexture(assetDb assets.Database, key string, filter textures.Filter) (*Texture, error) {
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

func (t *Texture) ReadPendingDataForTransparency() bool {
	t.pendingDataMutex.Lock()
	defer t.pendingDataMutex.Unlock()
	if t.hasTransparency == transparencyReadStateFound {
		return true
	}
	if t.hasTransparency != transparencyReadStateNone || t.pendingData == nil {
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
	t.delayedCreate(device, nil)
}

func (t *Texture) DelayedCreateInBatch(device *GPUDevice, batch *TextureUploadBatch) {
	defer tracing.NewRegion("Texture.DelayedCreateInBatch").End()
	t.delayedCreate(device, batch)
}

func (t *Texture) delayedCreate(device *GPUDevice, batch *TextureUploadBatch) {
	if t.RenderId.IsValid() {
		return
	}
	data := t.takePendingData()
	if data == nil {
		return
	}
	if batch != nil {
		device.SetupTextureInBatch(t, data, batch)
	} else {
		device.SetupTexture(t, data)
	}
}

func (t *Texture) takePendingData() *textures.TextureData {
	t.pendingDataMutex.Lock()
	defer t.pendingDataMutex.Unlock()
	data := t.pendingData
	t.pendingData = nil
	return data
}

func (t *Texture) pendingDataSize() uintptr {
	t.pendingDataMutex.Lock()
	defer t.pendingDataMutex.Unlock()
	if t.pendingData == nil {
		return 0
	}
	return uintptr(len(t.pendingData.Mem))
}

func NewTextureFromImage(key string, data []byte, filter textures.Filter) (*Texture, error) {
	defer tracing.NewRegion("rendering.NewTextureFromImage").End()
	tex := &Texture{Key: key, Filter: filter}
	tex.create(data)
	return tex, nil
}

func NewTextureFromMemory(key string, data []byte, width, height int, filter textures.Filter) (*Texture, error) {
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

func (t *Texture) Size() matrix.Vec2 {
	return matrix.NewVec2(t.Width, t.Height)
}

func (t *Texture) SetPendingDataDimensions(dim textures.Dimensions) {
	t.pendingDataMutex.Lock()
	defer t.pendingDataMutex.Unlock()
	if t.pendingData != nil {
		t.pendingData.Dimensions = dim
	}
}

func TexturePixelsFromAsset(assetDb assets.Database, key string) (textures.TextureData, error) {
	defer tracing.NewRegion("rendering.TexturePixelsFromAsset").End()
	key = selectKey(key)
	if assetDb.Exists(key) {
		if imgBuff, err := assetDb.Read(key); err != nil {
			return textures.TextureData{}, err
		} else if len(imgBuff) == 0 {
			return textures.TextureData{}, errors.New("no data in texture")
		} else {
			return textures.ReadImage(imgBuff, textures.InferFileFormat(key, imgBuff))
		}
	} else {
		return textures.TextureData{}, errors.New("texture does not exist")
	}
}

func selectKey(req string) string {
	if req == textures.GenerateUniqueTextureKey {
		return uuid.NewString()
	}
	return req
}

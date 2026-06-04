/******************************************************************************/
/* content_previewer_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_previews

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"sync"
	"testing"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/rendering"
)

func TestCachedPreviewTextureDoesNotQueryAssetDatabase(t *testing.T) {
	db := &previewAssetDatabase{}
	textureCache := rendering.NewTextureCache(nil, db)
	data := previewPNG(t)
	tex, err := cachedPreviewTexture(&textureCache, "preview_v2_tex", data, rendering.TextureFilterLinear)
	if err != nil {
		t.Fatal(err)
	}
	again, err := cachedPreviewTexture(&textureCache, "preview_v2_tex", data, rendering.TextureFilterLinear)
	if err != nil {
		t.Fatal(err)
	}
	if again != tex {
		t.Fatalf("cached preview texture = %p, want %p", again, tex)
	}
	if got := db.Calls(); got != 0 {
		t.Fatalf("asset database calls = %d, want 0", got)
	}
}

func previewPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 5, G: 6, B: 7, A: 255})
	var buff bytes.Buffer
	if err := png.Encode(&buff, img); err != nil {
		t.Fatal(err)
	}
	return buff.Bytes()
}

type previewAssetDatabase struct {
	mutex sync.Mutex
	calls int
}

func (d *previewAssetDatabase) PostWindowCreate(assets.PostWindowCreateHandle) error { return nil }
func (d *previewAssetDatabase) Cache(string, []byte)                                 {}
func (d *previewAssetDatabase) CacheRemove(string)                                   {}
func (d *previewAssetDatabase) CacheClear()                                          {}
func (d *previewAssetDatabase) Close()                                               {}

func (d *previewAssetDatabase) Exists(string) bool {
	d.record()
	return false
}

func (d *previewAssetDatabase) Read(string) ([]byte, error) {
	d.record()
	return nil, errors.New("preview test asset database should not be read")
}

func (d *previewAssetDatabase) ReadText(string) (string, error) {
	d.record()
	return "", errors.New("preview test asset database should not be read")
}

func (d *previewAssetDatabase) Calls() int {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.calls
}

func (d *previewAssetDatabase) record() {
	d.mutex.Lock()
	d.calls++
	d.mutex.Unlock()
}

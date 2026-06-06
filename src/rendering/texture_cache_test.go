/******************************************************************************/
/* texture_cache_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"image/color"
	"sync"
	"testing"
	"time"

	"kaijuengine.com/engine/assets"
)

func TestTextureCacheConcurrentRequestsShareSingleLoad(t *testing.T) {
	db := &countingTextureDatabase{
		files: map[string][]byte{
			"tex.png": testPNG(t, []color.RGBA{{R: 1, G: 2, B: 3, A: 255}}, 1, 1),
		},
		delay: 20 * time.Millisecond,
	}
	cache := NewTextureCache(nil, db)
	const workers = 16
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(workers)
	textures := make([]*Texture, workers)
	errs := make([]error, workers)
	for i := range workers {
		go func(i int) {
			defer wg.Done()
			<-start
			textures[i], errs[i] = cache.Texture("tex.png", TextureFilterLinear)
		}(i)
	}
	close(start)
	wg.Wait()

	for i := range errs {
		if errs[i] != nil {
			t.Fatalf("Texture worker %d returned error: %v", i, errs[i])
		}
		if textures[i] != textures[0] {
			t.Fatalf("Texture worker %d returned %p, want shared %p", i, textures[i], textures[0])
		}
	}
	if got := db.ReadCount(); got != 1 {
		t.Fatalf("asset reads = %d, want 1", got)
	}
}

func TestTextureCacheUploadBudgetPrioritizesHighPriority(t *testing.T) {
	cache := NewTextureCache(nil, assets.NewMockDB(map[string][]byte{}))
	cache.SetUploadBudget(TextureUploadBudget{MaxCreatesPerFrame: 2})
	lowA := testPendingTexture("low-a", 1)
	high := testPendingTexture("high", 1)
	lowB := testPendingTexture("low-b", 1)

	cache.mutex.Lock()
	cache.queuePendingLocked(lowA, TextureUploadPriorityNormal)
	cache.queuePendingLocked(high, TextureUploadPriorityHigh)
	cache.queuePendingLocked(lowB, TextureUploadPriorityNormal)
	selected := cache.takePendingUploadsLocked()
	remaining := append([]pendingTextureUpload(nil), cache.pendingTextures...)
	cache.mutex.Unlock()

	if len(selected) != 2 {
		t.Fatalf("selected uploads = %d, want 2", len(selected))
	}
	if selected[0].texture != high {
		t.Fatalf("first selected = %s, want high", selected[0].texture.Key)
	}
	if selected[1].texture != lowA {
		t.Fatalf("second selected = %s, want oldest normal priority", selected[1].texture.Key)
	}
	if len(remaining) != 1 || remaining[0].texture != lowB {
		t.Fatalf("remaining uploads = %#v, want low-b", remaining)
	}
}

func testPendingTexture(key string, bytes int) *Texture {
	return &Texture{
		Key: key,
		pendingData: &TextureData{
			Mem: make([]byte, bytes),
		},
	}
}

type countingTextureDatabase struct {
	files map[string][]byte
	delay time.Duration
	mutex sync.Mutex
	reads int
}

func (d *countingTextureDatabase) PostWindowCreate(assets.PostWindowCreateHandle) error { return nil }
func (d *countingTextureDatabase) Cache(string, []byte)                                 {}
func (d *countingTextureDatabase) CacheRemove(string)                                   {}
func (d *countingTextureDatabase) CacheClear()                                          {}
func (d *countingTextureDatabase) Close()                                               {}

func (d *countingTextureDatabase) Exists(key string) bool {
	_, ok := d.files[key]
	return ok
}

func (d *countingTextureDatabase) Read(key string) ([]byte, error) {
	d.mutex.Lock()
	d.reads++
	d.mutex.Unlock()
	if d.delay > 0 {
		time.Sleep(d.delay)
	}
	if data, ok := d.files[key]; ok {
		return data, nil
	}
	return nil, errors.New("missing test asset")
}

func (d *countingTextureDatabase) ReadText(key string) (string, error) {
	data, err := d.Read(key)
	return string(data), err
}

func (d *countingTextureDatabase) ReadCount() int {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.reads
}

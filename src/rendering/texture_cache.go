/******************************************************************************/
/* texture_cache.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"sync"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

type TextureUploadPriority int

const (
	TextureUploadPriorityNormal TextureUploadPriority = iota
	TextureUploadPriorityHigh
)

type TextureUploadBudget struct {
	MaxCreatesPerFrame int
	MaxBytesPerFrame   uintptr
}

type pendingTextureUpload struct {
	texture  *Texture
	priority TextureUploadPriority
	sequence uint64
}

type textureLoadCall struct {
	done    chan struct{}
	texture *Texture
	err     error
}

type TextureCache struct {
	device          *GPUDevice
	assetDatabase   assets.Database
	textures        [TextureFilterMax]map[string]*Texture
	loading         [TextureFilterMax]map[string]*textureLoadCall
	pendingTextures []pendingTextureUpload
	pendingFree     []TextureId
	uploadBudget    TextureUploadBudget
	uploadSequence  uint64
	mutex           sync.Mutex
}

func NewTextureCache(device *GPUDevice, assetDatabase assets.Database) TextureCache {
	defer tracing.NewRegion("rendering.NewTextureCache").End()
	tc := TextureCache{
		device:          device,
		assetDatabase:   assetDatabase,
		pendingTextures: make([]pendingTextureUpload, 0),
		pendingFree:     make([]TextureId, 0),
		mutex:           sync.Mutex{},
	}
	for i := range tc.textures {
		tc.textures[i] = make(map[string]*Texture)
		tc.loading[i] = make(map[string]*textureLoadCall)
	}
	return tc
}

func (t *TextureCache) SetUploadBudget(budget TextureUploadBudget) {
	t.mutex.Lock()
	t.uploadBudget = budget
	t.mutex.Unlock()
}

func (t *TextureCache) Find(textureKey string, filter TextureFilter) (*Texture, bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	texture, ok := t.textures[filter][textureKey]
	return texture, ok
}

func (t *TextureCache) Texture(textureKey string, filter TextureFilter) (*Texture, error) {
	return t.TextureWithPriority(textureKey, filter, TextureUploadPriorityNormal)
}

func (t *TextureCache) TextureWithPriority(textureKey string, filter TextureFilter, priority TextureUploadPriority) (*Texture, error) {
	defer tracing.NewRegion("TextureCache.Texture").End()
	textureKey = selectKey(textureKey)
	t.mutex.Lock()
	if texture, ok := t.textures[filter][textureKey]; ok {
		t.promotePendingLocked(texture, priority)
		t.mutex.Unlock()
		return texture, nil
	}
	if call, ok := t.loading[filter][textureKey]; ok {
		t.mutex.Unlock()
		<-call.done
		if call.texture != nil {
			t.mutex.Lock()
			t.promotePendingLocked(call.texture, priority)
			t.mutex.Unlock()
		}
		return call.texture, call.err
	}
	call := &textureLoadCall{done: make(chan struct{})}
	t.loading[filter][textureKey] = call
	t.mutex.Unlock()

	texture, err := NewTexture(t.assetDatabase, textureKey, filter)

	t.mutex.Lock()
	if err == nil {
		if cached, ok := t.textures[filter][textureKey]; ok {
			texture = cached
			t.promotePendingLocked(texture, priority)
		} else {
			t.textures[filter][textureKey] = texture
			t.queuePendingLocked(texture, priority)
		}
	}
	call.texture = texture
	call.err = err
	delete(t.loading[filter], textureKey)
	close(call.done)
	t.mutex.Unlock()
	return texture, err
}

// ReloadTexture forces a reload of the texture data for the given texture key and filter, bypassing the cache.
// And invalidates the cached decoded data to ensure the next load will read fresh data from the asset database.
func (t *TextureCache) ReloadTexture(textureKey string, filter TextureFilter) error {
	t.mutex.Lock()
	texture, ok := t.textures[filter][textureKey]
	if !ok {
		t.mutex.Unlock()
		return nil
	}
	renderId := texture.RenderId
	t.mutex.Unlock()

	if err := texture.Reload(t.assetDatabase); err != nil {
		return err
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	if renderId.IsValid() {
		t.pendingFree = append(t.pendingFree, renderId)
	}
	t.queuePendingLocked(texture, TextureUploadPriorityHigh)
	return nil
}

func (t *TextureCache) InsertTexture(tex *Texture) {
	t.InsertTextureWithPriority(tex, TextureUploadPriorityNormal)
}

func (t *TextureCache) InsertTextureWithPriority(tex *Texture, priority TextureUploadPriority) {
	defer tracing.NewRegion("TextureCache.InsertTexture").End()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.textures[tex.Filter][tex.Key]; ok {
		return
	}
	t.queuePendingLocked(tex, priority)
	t.textures[tex.Filter][tex.Key] = tex
}

// InsertRawTexture creates a texture directly from raw data and caches it without needing to read from the asset database
// This is useful for dynamically generated textures or when the raw data is already available in memory, caching without redundant file I/O.
func (t *TextureCache) InsertRawTexture(key string, data []byte, width, height int, filter TextureFilter) (*Texture, error) {
	return t.InsertRawTextureWithPriority(key, data, width, height, filter, TextureUploadPriorityNormal)
}

func (t *TextureCache) InsertRawTextureWithPriority(key string, data []byte, width, height int, filter TextureFilter, priority TextureUploadPriority) (*Texture, error) {
	defer tracing.NewRegion("TextureCache.InsertTexture").End()
	key = selectKey(key)
	t.mutex.Lock()
	if texture, ok := t.textures[filter][key]; ok {
		t.promotePendingLocked(texture, priority)
		t.mutex.Unlock()
		return texture, nil
	}
	t.mutex.Unlock()
	tex, err := NewTextureFromMemory(key, data, width, height, filter)
	if err != nil {
		return nil, err
	}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if texture, ok := t.textures[filter][key]; ok {
		t.promotePendingLocked(texture, priority)
		return texture, nil
	}
	t.queuePendingLocked(tex, priority)
	t.textures[filter][key] = tex
	return tex, nil
}

// InsertImageTexture creates a texture from raw image data and caches it efficiently
func (t *TextureCache) InsertImageTexture(key string, imageData []byte, filter TextureFilter) (*Texture, error) {
	return t.InsertImageTextureWithPriority(key, imageData, filter, TextureUploadPriorityNormal)
}

func (t *TextureCache) InsertImageTextureWithPriority(key string, imageData []byte, filter TextureFilter, priority TextureUploadPriority) (*Texture, error) {
	defer tracing.NewRegion("TextureCache.InsertImageTexture").End()
	key = selectKey(key)
	t.mutex.Lock()
	if texture, ok := t.textures[filter][key]; ok {
		t.promotePendingLocked(texture, priority)
		t.mutex.Unlock()
		return texture, nil
	}
	t.mutex.Unlock()
	tex, err := NewTextureFromMemory(key, imageData, 0, 0, filter)
	if err != nil {
		return nil, err
	}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if texture, ok := t.textures[filter][key]; ok {
		t.promotePendingLocked(texture, priority)
		return texture, nil
	}
	t.queuePendingLocked(tex, priority)
	t.textures[filter][key] = tex
	return tex, nil
}

// ForceRemoveTexture evicts a texture from the cache and reclaims its GPU
// memory. The texture's RenderId (if it has already been uploaded) is queued
// into pendingFree so ProcessPending frees it on the next frame, and any
// still-queued upload for the texture is dropped so an evicted texture is
// never uploaded after removal.
func (t *TextureCache) ForceRemoveTexture(key string, filter TextureFilter) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	texture, ok := t.textures[filter][key]
	if !ok {
		return
	}
	if texture.RenderId.IsValid() {
		t.pendingFree = append(t.pendingFree, texture.RenderId)
	}
	t.removePendingUploadLocked(texture)
	delete(t.textures[filter], key)
}

func (t *TextureCache) ProcessPending() {
	defer tracing.NewRegion("TextureCache.CreatePending").End()
	t.mutex.Lock()
	pendingFree := append([]TextureId(nil), t.pendingFree...)
	t.pendingFree = klib.WipeSlice(t.pendingFree)
	pendingTextures := t.takePendingUploadsLocked()
	t.mutex.Unlock()
	for i := range pendingFree {
		t.device.LogicalDevice.FreeTexture(&pendingFree[i])
	}
	if len(pendingTextures) == 0 {
		return
	}
	batch := t.device.BeginTextureUploadBatch()
	if batch == nil {
		for _, pending := range pendingTextures {
			pending.texture.DelayedCreate(t.device)
		}
		return
	}
	for _, pending := range pendingTextures {
		pending.texture.DelayedCreateInBatch(t.device, batch)
	}
	batch.End()
}

// Destroy frees all textures in the cache and clears the decoded texture data cache to release GPU and memory resources when the texture cache is no longer needed.
// This should be called when the application is shutting down or when the texture cache needs to be reset to ensure proper cleanup of resources.
func (t *TextureCache) Destroy() {
	defer tracing.NewRegion("TextureCache.Destroy").End()
	t.pendingFree = klib.WipeSlice(t.pendingFree)
	t.pendingTextures = klib.WipeSlice(t.pendingTextures)
	for i := range t.textures {
		for _, tex := range t.textures[i] {
			t.device.LogicalDevice.FreeTexture(&tex.RenderId)
		}
		t.textures[i] = make(map[string]*Texture)
	}
}

func (t *TextureCache) queuePendingLocked(texture *Texture, priority TextureUploadPriority) {
	t.uploadSequence++
	t.pendingTextures = append(t.pendingTextures, pendingTextureUpload{
		texture:  texture,
		priority: priority,
		sequence: t.uploadSequence,
	})
}

// removePendingUploadLocked drops any queued upload referencing the given
// texture so a texture that is evicted before its deferred upload runs is not
// uploaded after removal.
func (t *TextureCache) removePendingUploadLocked(texture *Texture) {
	for i := 0; i < len(t.pendingTextures); {
		if t.pendingTextures[i].texture == texture {
			t.pendingTextures = klib.RemoveUnordered(t.pendingTextures, i)
		} else {
			i++
		}
	}
}

func (t *TextureCache) promotePendingLocked(texture *Texture, priority TextureUploadPriority) {
	if priority <= TextureUploadPriorityNormal || texture == nil {
		return
	}
	for i := range t.pendingTextures {
		if t.pendingTextures[i].texture == texture && t.pendingTextures[i].priority < priority {
			t.pendingTextures[i].priority = priority
		}
	}
}

func (t *TextureCache) takePendingUploadsLocked() []pendingTextureUpload {
	if len(t.pendingTextures) == 0 {
		return nil
	}
	budget := t.uploadBudget
	if budget.MaxCreatesPerFrame <= 0 && budget.MaxBytesPerFrame <= 0 {
		pending := append([]pendingTextureUpload(nil), t.pendingTextures...)
		t.pendingTextures = klib.WipeSlice(t.pendingTextures)
		return pending
	}
	selected := make([]pendingTextureUpload, 0, len(t.pendingTextures))
	remaining := make([]pendingTextureUpload, 0, len(t.pendingTextures))
	var bytes uintptr
	for len(t.pendingTextures) > 0 {
		idx := t.nextPendingUploadIndexLocked()
		pending := t.pendingTextures[idx]
		t.pendingTextures = klib.RemoveUnordered(t.pendingTextures, idx)
		uploadBytes := pendingTextureUploadBytes(pending.texture)
		if textureUploadBudgetExceeded(budget, len(selected), bytes, uploadBytes) {
			remaining = append(remaining, pending)
			continue
		}
		selected = append(selected, pending)
		bytes += uploadBytes
	}
	t.pendingTextures = remaining
	return selected
}

func (t *TextureCache) nextPendingUploadIndexLocked() int {
	best := 0
	for i := 1; i < len(t.pendingTextures); i++ {
		if t.pendingTextures[i].priority > t.pendingTextures[best].priority ||
			(t.pendingTextures[i].priority == t.pendingTextures[best].priority &&
				t.pendingTextures[i].sequence < t.pendingTextures[best].sequence) {
			best = i
		}
	}
	return best
}

func pendingTextureUploadBytes(texture *Texture) uintptr {
	if texture == nil || texture.pendingData == nil {
		return 0
	}
	return uintptr(len(texture.pendingData.Mem))
}

func textureUploadBudgetExceeded(budget TextureUploadBudget, selected int, bytes, uploadBytes uintptr) bool {
	if budget.MaxCreatesPerFrame > 0 && selected >= budget.MaxCreatesPerFrame {
		return true
	}
	if budget.MaxBytesPerFrame > 0 && selected > 0 && bytes+uploadBytes > budget.MaxBytesPerFrame {
		return true
	}
	return false
}

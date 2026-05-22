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

type TextureCache struct {
	device          *GPUDevice
	assetDatabase   assets.Database
	textures        [TextureFilterMax]map[string]*Texture
	pendingTextures []*Texture
	mutex           sync.Mutex
}

func NewTextureCache(device *GPUDevice, assetDatabase assets.Database) TextureCache {
	defer tracing.NewRegion("rendering.NewTextureCache").End()
	tc := TextureCache{
		device:          device,
		assetDatabase:   assetDatabase,
		pendingTextures: make([]*Texture, 0),
		mutex:           sync.Mutex{},
	}
	for i := range tc.textures {
		tc.textures[i] = make(map[string]*Texture)
	}
	return tc
}

func (t *TextureCache) Texture(textureKey string, filter TextureFilter) (*Texture, error) {
	defer tracing.NewRegion("TextureCache.Texture").End()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if texture, ok := t.textures[filter][textureKey]; ok {
		return texture, nil
	} else {
		if texture, err := NewTexture(t.assetDatabase, textureKey, filter); err == nil {
			t.pendingTextures = append(t.pendingTextures, texture)
			t.textures[filter][textureKey] = texture
			return texture, nil
		} else {
			return nil, err
		}
	}
}

// ReloadTexture forces a reload of the texture data for the given texture key and filter, bypassing the cache.
// And invalidates the cached decoded data to ensure the next load will read fresh data from the asset database.
func (t *TextureCache) ReloadTexture(textureKey string, filter TextureFilter) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	texture, ok := t.textures[filter][textureKey]
	if !ok {
		return nil
	}
	t.device.LogicalDevice.FreeTexture(&texture.RenderId)
	if err := texture.Reload(t.assetDatabase); err != nil {
		return err
	}
	t.pendingTextures = append(t.pendingTextures, texture)
	return nil
}

func (t *TextureCache) InsertTexture(tex *Texture) {
	defer tracing.NewRegion("TextureCache.InsertTexture").End()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.textures[tex.Filter][tex.Key]; ok {
		return
	}
	t.pendingTextures = append(t.pendingTextures, tex)
	t.textures[tex.Filter][tex.Key] = tex
}

// InsertRawTexture creates a texture directly from raw data and caches it without needing to read from the asset database
// This is useful for dynamically generated textures or when the raw data is already available in memory, caching without redundant file I/O.
func (t *TextureCache) InsertRawTexture(key string, data []byte, width, height int, filter TextureFilter) (*Texture, error) {
	defer tracing.NewRegion("TextureCache.InsertTexture").End()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	key = selectKey(key)
	if texture, ok := t.textures[filter][key]; ok {
		return texture, nil
	}
	tex, err := NewTextureFromMemory(key, data, width, height, filter)
	if err != nil {
		return nil, err
	}
	t.pendingTextures = append(t.pendingTextures, tex)
	t.textures[filter][key] = tex
	return tex, nil
}

// InsertImageTexture creates a texture from raw image data and caches it efficiently
func (t *TextureCache) InsertImageTexture(key string, imageData []byte, filter TextureFilter) (*Texture, error) {
	defer tracing.NewRegion("TextureCache.InsertImageTexture").End()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	key = selectKey(key)
	if texture, ok := t.textures[filter][key]; ok {
		return texture, nil
	}
	tex, err := NewTextureFromMemory(key, imageData, 0, 0, filter)
	if err != nil {
		return nil, err
	}
	t.pendingTextures = append(t.pendingTextures, tex)
	t.textures[filter][key] = tex
	return tex, nil
}

func (t *TextureCache) ForceRemoveTexture(key string, filter TextureFilter) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.textures[filter], key)
}

func (t *TextureCache) CreatePending() {
	defer tracing.NewRegion("TextureCache.CreatePending").End()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, texture := range t.pendingTextures {
		texture.DelayedCreate(t.device)
	}
	t.pendingTextures = klib.WipeSlice(t.pendingTextures)
}

// Destroy frees all textures in the cache and clears the decoded texture data cache to release GPU and memory resources when the texture cache is no longer needed.
// This should be called when the application is shutting down or when the texture cache needs to be reset to ensure proper cleanup of resources.
func (t *TextureCache) Destroy() {
	defer tracing.NewRegion("TextureCache.Destroy").End()
	t.pendingTextures = klib.WipeSlice(t.pendingTextures)
	for i := range t.textures {
		for _, tex := range t.textures[i] {
			t.device.LogicalDevice.FreeTexture(&tex.RenderId)
		}
		t.textures[i] = make(map[string]*Texture)
	}
}

package rendering

import (
	"kaiju/assets"
	"sync"
)

type TextureCache struct {
	renderer        Renderer
	assetDatabase   *assets.Database
	textures        [TextureFilterMax]map[string]*Texture
	pendingTextures []*Texture
	mutex           sync.Mutex
}

func NewTextureCache(renderer Renderer, assetDatabase *assets.Database) TextureCache {
	tc := TextureCache{
		renderer:        renderer,
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
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if texture, ok := t.textures[filter][textureKey]; ok {
		return texture, nil
	} else {
		if texture, err := NewTexture(t.renderer, t.assetDatabase, textureKey, filter); err == nil {
			t.pendingTextures = append(t.pendingTextures, texture)
			t.textures[filter][textureKey] = texture
			return texture, nil
		} else {
			return nil, err
		}
	}
}

func (t *TextureCache) CreatePending() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, texture := range t.pendingTextures {
		texture.DelayedCreate(t.renderer)
	}
	t.pendingTextures = t.pendingTextures[:0]
}

func (t *TextureCache) Destroy() {
	for _, texture := range t.pendingTextures {
		texture.Destroy(t.renderer)
	}
	t.pendingTextures = t.pendingTextures[:0]
	for i := range t.textures {
		for _, texture := range t.textures[i] {
			texture.Destroy(t.renderer)
		}
		t.textures[i] = make(map[string]*Texture)
	}
}

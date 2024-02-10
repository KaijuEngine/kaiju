package rendering

import (
	"kaiju/assets"
	"sync"
)

type TextureCache struct {
	renderer        Renderer
	assetDatabase   *assets.Database
	textures        map[string]*Texture
	pendingTextures []*Texture
	mutex           sync.Mutex
}

func NewTextureCache(renderer Renderer, assetDatabase *assets.Database) TextureCache {
	return TextureCache{
		renderer:        renderer,
		assetDatabase:   assetDatabase,
		textures:        make(map[string]*Texture),
		pendingTextures: make([]*Texture, 0),
		mutex:           sync.Mutex{},
	}
}

func (t *TextureCache) Texture(textureKey string, filter TextureFilter) (*Texture, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if texture, ok := t.textures[textureKey]; ok {
		return texture, nil
	} else {
		if texture, err := NewTexture(t.renderer, t.assetDatabase, textureKey, filter); err == nil {
			t.pendingTextures = append(t.pendingTextures, texture)
			t.textures[textureKey] = texture
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
	for _, texture := range t.textures {
		texture.Destroy(t.renderer)
	}
	t.textures = make(map[string]*Texture)
}

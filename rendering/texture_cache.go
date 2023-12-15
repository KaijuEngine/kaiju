package rendering

import "kaiju/assets"

type TextureCache struct {
	renderer        Renderer
	assetDatabase   *assets.Database
	textures        map[string]*Texture
	pendingTextures []*Texture
}

func NewTextureCache(renderer Renderer, assetDatabase *assets.Database) TextureCache {
	return TextureCache{
		renderer:        renderer,
		assetDatabase:   assetDatabase,
		textures:        make(map[string]*Texture),
		pendingTextures: make([]*Texture, 0),
	}
}

func (t *TextureCache) Texture(textureKey string, filter TextureFilter) (*Texture, error) {
	if texture, ok := t.textures[textureKey]; ok {
		return texture, nil
	} else {
		if texture, err := NewTexture(t.renderer, t.assetDatabase, textureKey, filter); err == nil {
			t.pendingTextures = append(t.pendingTextures, texture)
			return texture, nil
		} else {
			return nil, err
		}
	}
}

func (t *TextureCache) CreatePending() {
	for _, texture := range t.pendingTextures {
		texture.DelayedCreate(t.renderer)
	}
	t.pendingTextures = t.pendingTextures[:0]
}

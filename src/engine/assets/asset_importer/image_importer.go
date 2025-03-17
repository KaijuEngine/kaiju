package asset_importer

import (
	"kaiju/rendering"
	"log/slog"
)

type ImageMetadata struct {
	Filter  string `options:"textureFilterOptions"`
	Mipmaps int32
}

func (m *ImageMetadata) TextureFilter() rendering.TextureFilter {
	if f, ok := textureFilterOptions[m.Filter]; ok {
		return f
	}
	slog.Warn("tried to read image metadata filter but has invalid key",
		"key", m.Filter)
	return rendering.TextureFilterLinear
}

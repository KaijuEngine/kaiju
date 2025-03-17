package asset_importer

import (
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"path/filepath"
)

type RenderPassImporter struct{}

type RenderPassMetadata struct{}

func (m RenderPassImporter) MetadataStructure() any {
	return &RenderPassMetadata{}
}

func (m RenderPassImporter) Handles(path string) bool {
	return filepath.Ext(path) == editor_config.FileExtensionRenderPass
}

func (m RenderPassImporter) Import(path string) error {
	adi, err := createADI(m, path, nil)
	if err != nil {
		return err
	}
	adi.Type = editor_config.AssetTypeRenderPass
	return asset_info.Write(adi)
}

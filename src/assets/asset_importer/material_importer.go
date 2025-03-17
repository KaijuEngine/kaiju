package asset_importer

import (
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"path/filepath"
)

type MaterialImporter struct{}

type MaterialMetadata struct{}

func (m MaterialImporter) MetadataStructure() any {
	return &MaterialMetadata{}
}

func (m MaterialImporter) Handles(path string) bool {
	return filepath.Ext(path) == editor_config.FileExtensionMaterial
}

func (m MaterialImporter) Import(path string) error {
	adi, err := createADI(m, path, nil)
	if err != nil {
		return err
	}
	adi.Type = editor_config.AssetTypeMaterial
	return asset_info.Write(adi)
}

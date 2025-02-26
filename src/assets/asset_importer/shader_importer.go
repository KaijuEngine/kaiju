package asset_importer

import (
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"path/filepath"
)

type ShaderImporter struct{}

func (m ShaderImporter) Handles(path string) bool {
	return filepath.Ext(path) == editor_config.FileExtensionShader
}

func (m ShaderImporter) Import(path string) error {
	adi, err := createADI(path, nil)
	if err != nil {
		return err
	}
	adi.Type = editor_config.AssetTypeShader
	return asset_info.Write(adi)
}

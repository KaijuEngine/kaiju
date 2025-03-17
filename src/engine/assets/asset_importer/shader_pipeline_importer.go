package asset_importer

import (
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"path/filepath"
)

type ShaderPipelineImporter struct{}

type ShaderPipelineMetadata struct{}

func (m ShaderPipelineImporter) MetadataStructure() any {
	return &ShaderPipelineMetadata{}
}

func (m ShaderPipelineImporter) Handles(path string) bool {
	return filepath.Ext(path) == editor_config.FileExtensionShaderPipeline
}

func (m ShaderPipelineImporter) Import(path string) error {
	adi, err := createADI(m, path, nil)
	if err != nil {
		return err
	}
	adi.Type = editor_config.AssetTypeShaderPipeline
	return asset_info.Write(adi)
}

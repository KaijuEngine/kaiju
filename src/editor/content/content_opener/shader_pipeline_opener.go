package content_opener

import (
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/shader_designer"
)

type ShaderPipelineOpener struct{}

func (o ShaderPipelineOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeShaderPipeline
}

func (o ShaderPipelineOpener) Open(adi asset_info.AssetDatabaseInfo, ed interfaces.Editor) error {
	shader_designer.OpenPipeline(adi.Path)
	return nil
}

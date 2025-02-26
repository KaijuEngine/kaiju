package content_opener

import (
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/shader_designer"
)

type ShaderOpener struct{}

func (o ShaderOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeShader
}

func (o ShaderOpener) Open(adi asset_info.AssetDatabaseInfo, ed interfaces.Editor) error {
	shader_designer.OpenShader(adi.Path)
	return nil
}

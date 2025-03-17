package content_opener

import (
	"kaiju/editor/editor_config"
	"kaiju/editor/editor_interface"
	"kaiju/editor/ui/shader_designer"
	"kaiju/engine/assets/asset_info"
)

type ShaderOpener struct{}

func (o ShaderOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeShader
}

func (o ShaderOpener) Open(adi asset_info.AssetDatabaseInfo, ed editor_interface.Editor) error {
	shader_designer.OpenShader(adi.Path, ed.Host().LogStream)
	return nil
}

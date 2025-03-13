package content_opener

import (
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/shader_designer"
)

type RenderPassOpener struct{}

func (o RenderPassOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeRenderPass
}

func (o RenderPassOpener) Open(adi asset_info.AssetDatabaseInfo, ed interfaces.Editor) error {
	shader_designer.OpenRenderPass(adi.Path, ed.Host().LogStream)
	return nil
}

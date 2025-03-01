package content_opener

import (
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/shader_designer"
)

type MaterialOpener struct{}

func (o MaterialOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeMaterial
}

func (o MaterialOpener) Open(adi asset_info.AssetDatabaseInfo, ed interfaces.Editor) error {
	shader_designer.OpenMaterial(adi.Path)
	return nil
}

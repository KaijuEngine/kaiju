package content_opener

import (
	"kaiju/editor/editor_config"
	"kaiju/editor/editor_interface"
	"kaiju/editor/ui/shader_designer"
	"kaiju/engine/assets/asset_info"
)

type MaterialOpener struct{}

func (o MaterialOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeMaterial
}

func (o MaterialOpener) Open(adi asset_info.AssetDatabaseInfo, ed editor_interface.Editor) error {
	shader_designer.OpenMaterial(adi.Path, ed.Host().LogStream)
	return nil
}

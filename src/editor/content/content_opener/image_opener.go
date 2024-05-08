package content_opener

import (
	"errors"
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"kaiju/editor/interfaces"
	"kaiju/systems/console"
)

type ImageOpener struct{}

func (o ImageOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeImage
}

func (o ImageOpener) Open(adi asset_info.AssetDatabaseInfo, ed interfaces.Editor) error {
	console.For(ed.Host()).Write("Opening an image")
	return errors.New("not implemented")
}

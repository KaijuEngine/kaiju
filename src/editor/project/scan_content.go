package project

import (
	"kaiju/assets/asset_importer"
	"os"
	"path/filepath"
)

func ScanContent(importers *asset_importer.ImportRegistry) error {
	return filepath.Walk("content", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		importers.ImportIfNew(path)
		return nil
	})
}

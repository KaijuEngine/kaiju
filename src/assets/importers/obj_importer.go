package importers

import (
	"errors"
	"path/filepath"
)

type OBJImporter struct {
}

func (m OBJImporter) Handles(path string) bool {
	return filepath.Ext(path) == ".obj"
}

func (m OBJImporter) Import(path string) error {
	return errors.New("not implemented")
}

package asset_importer

import (
	"kaiju/assets"
	"kaiju/assets/asset_info"
	"kaiju/editor/editor_config"
	"kaiju/rendering/loaders"
	"path/filepath"
)

type GlbImporter struct{}

func (m GlbImporter) MetadataStructure() any {
	return &MeshMetadata{}
}

func (m GlbImporter) Handles(path string) bool {
	return filepath.Ext(path) == editor_config.FileExtensionGlb
}

func (m GlbImporter) Import(path string) error {
	adi, err := createADI(m, path, cleanupMesh)
	if err != nil {
		return err
	}
	adi.Type = editor_config.AssetTypeGlb
	a := assets.NewDatabase()
	mesh, err := loaders.GLTF(adi.Path, &a)
	if err != nil {
		return err
	}
	if err := importMeshToCache(&adi, mesh); err != nil {
		return err
	}
	return asset_info.Write(adi)
}

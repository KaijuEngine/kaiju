package asset_importer

import (
	"errors"
	"kaiju/editor/cache/project_cache"
	"kaiju/editor/editor_config"
	"kaiju/engine/assets"
	"kaiju/engine/assets/asset_info"
	"kaiju/rendering/loaders/load_result"

	"github.com/KaijuEngine/uuid"
)

type MeshMetadata struct {
	Name     string
	Material string
}

func cleanupMesh(adi asset_info.AssetDatabaseInfo) {
	project_cache.DeleteMesh(adi)
	adi.Children = adi.Children[:0]
	adi.Metadata = nil
}

func importMeshToCache(adi *asset_info.AssetDatabaseInfo, mesh load_result.Result) error {
	if len(mesh.Meshes) == 0 {
		return errors.New("no meshes found in OBJ file")
	}
	adi.Metadata = MeshMetadata{
		Name: mesh.Meshes[0].Name,
	}
	for _, o := range mesh.Meshes {
		info := adi.SpawnChild(uuid.New().String())
		info.Type = editor_config.AssetTypeMesh
		info.ParentID = adi.ID
		if err := project_cache.CacheMesh(info.ID, o); err != nil {
			return err
		}
		info.Metadata = MeshMetadata{
			// TODO:  Write the correct material to the adi
			Material: assets.MaterialDefinitionBasic,
			Name:     o.MeshName,
		}
		adi.Children = append(adi.Children, info)
	}
	return nil
}

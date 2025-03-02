package asset_importer

import (
	"errors"
	"kaiju/assets"
	"kaiju/assets/asset_info"
	"kaiju/cache/project_cache"
	"kaiju/editor/editor_config"
	"kaiju/rendering/loaders/load_result"

	"github.com/KaijuEngine/uuid"
)

func cleanupMesh(adi asset_info.AssetDatabaseInfo) {
	project_cache.DeleteMesh(adi)
	adi.Children = adi.Children[:0]
	adi.Metadata = make(map[string]string)
}

func importMeshToCache(adi *asset_info.AssetDatabaseInfo, mesh load_result.Result) error {
	if len(mesh.Meshes) == 0 {
		return errors.New("no meshes found in OBJ file")
	}
	adi.Metadata["name"] = mesh.Meshes[0].Name
	for _, o := range mesh.Meshes {
		info := adi.SpawnChild(uuid.New().String())
		info.Type = editor_config.AssetTypeMesh
		info.ParentID = adi.ID
		if err := project_cache.CacheMesh(info, o); err != nil {
			return err
		}
		// TODO:  Write the correct material to the adi
		info.Metadata["material"] = assets.MaterialDefinitionBasic
		info.Metadata["name"] = o.MeshName
		adi.Children = append(adi.Children, info)
	}
	return nil
}

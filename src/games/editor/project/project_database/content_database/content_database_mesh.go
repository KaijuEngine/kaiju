package content_database

import (
	"fmt"
	"kaiju/engine/assets"
	"kaiju/games/editor/project/project_file_system"
	"kaiju/rendering/loaders"
	"kaiju/rendering/loaders/kaiju_mesh"
	"kaiju/rendering/loaders/load_result"
	"path/filepath"
	"slices"
)

func init() { contentCategories = append(contentCategories, Mesh{}) }

// Mesh is a [ContentCategory] represented by a file with a ".gltf" or ".glb"
// extension. This file can contain multiple meshes as well as the textures that
// are assigned to the meshes. The textures will be imported as dependencies.
type Mesh struct{}
type MeshConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Mesh) Path() string       { return project_file_system.ContentMeshFolder }
func (Mesh) TypeName() string   { return "mesh" }
func (Mesh) ExtNames() []string { return []string{".gltf", ".glb"} }

func (Mesh) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	ext := filepath.Ext(src)
	p := ProcessedImport{}
	var res load_result.Result
	switch ext {
	case ".gltf":
		fallthrough
	case ".glb":
		adb, err := assets.NewFileDatabase(filepath.Dir(src))
		if err != nil {
			return p, err
		}
		if res, err = loaders.GLTF(src, adb); err != nil {
			return p, err
		}
	}
	if len(res.Meshes) == 0 {
		return p, NoMeshesInFileError{Path: src}
	}
	baseName := fileNameNoExt(src)
	kms := kaiju_mesh.LoadedResultToKaijuMesh(res)
	for i := range kms {
		kd, err := kms[i].Serialize()
		if err != nil {
			return p, err
		}
		p.Variants = append(p.Variants, ImportVariant{
			Name: fmt.Sprintf("%s-%s", baseName, kms[i].Name),
			Data: kd,
		})
	}
	for i := range res.Meshes {
		p.Dependencies = append(p.Dependencies, slices.Clone(res.Meshes[i].Textures)...)
	}
	return p, nil
}

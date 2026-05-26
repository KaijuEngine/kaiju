/******************************************************************************/
/* content_database_mesh.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
	"kaijuengine.com/rendering/loaders/load_result"
)

func init() { addCategory(Mesh{}) }

// Mesh is a [ContentCategory] represented by a file with a ".gltf" or ".glb"
// extension. This file can contain multiple meshes as well as the textures that
// are assigned to the meshes. The textures will be imported as dependencies.
type Mesh struct{}
type MeshConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Mesh) Path() string       { return project_file_system.ContentMeshFolder }
func (Mesh) TypeName() string   { return "Mesh" }
func (Mesh) ExtNames() []string { return []string{".gltf", ".glb", ".obj", ".fbx"} }

type meshImportPostProcData struct {
	mesh         load_result.Mesh
	kaijuMesh    kaiju_mesh.KaijuMesh
	meshes       []load_result.Mesh
	isAnimated   bool
	textureBytes map[string][]byte
}

func EnsureMeshBVHInBackground(km kaiju_mesh.KaijuMesh, path string, fs *project_file_system.FileSystem, id string) {
	if km.BVH != nil {
		return
	}
	// goroutine
	go func() {
		km.EnsureBVH()
		if km.BVH == nil {
			return
		}
		writeMeshBVH(km, path, fs, id)
	}()
}

func SaveMeshBVHInBackground(km kaiju_mesh.KaijuMesh, path string, fs *project_file_system.FileSystem, id string) {
	if km.BVH == nil {
		return
	}
	go writeMeshBVH(km, path, fs, id)
}

func writeMeshBVH(km kaiju_mesh.KaijuMesh, path string, fs *project_file_system.FileSystem, id string) {
	data, err := km.Serialize()
	if err != nil {
		slog.Error("failed to serialize the mesh BVH", "id", id, "error", err)
		return
	}
	if err = fs.WriteFile(path, data, os.ModePerm); err != nil {
		slog.Error("failed to write the mesh BVH", "id", id, "path", path, "error", err)
	}
}

func (Mesh) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Mesh.Import").End()
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
		if res, err = loaders.GLTF(filepath.Base(src), adb); err != nil {
			return p, err
		}
	case ".obj":
		adb, err := assets.NewFileDatabase(filepath.Dir(src))
		if err != nil {
			return p, err
		}
		if res, err = loaders.OBJ(filepath.Base(src), adb); err != nil {
			return p, err
		}
	case ".fbx":
		adb, err := assets.NewFileDatabase(filepath.Dir(src))
		if err != nil {
			return p, err
		}
		if res, err = loaders.FBX(filepath.Base(src), adb); err != nil {
			return p, err
		}
	}
	if len(res.Meshes) == 0 {
		return p, NoMeshesInFileError{Path: src}
	}
	baseName := fileNameNoExt(src)
	kms := kaiju_mesh.LoadedResultToKaijuMesh(res)
	postProcData := map[string]meshImportPostProcData{}
	for i := range kms {
		kd, err := kms[i].Serialize()
		if err != nil {
			return p, err
		}
		parts := strings.Split(kms[i].Name, "/")
		v := ImportVariant{
			Name: fmt.Sprintf("%s-%s", baseName, parts[len(parts)-1]),
			Data: kd,
		}
		p.Variants = append(p.Variants, v)
		if res.Meshes[i].Node == nil {
			slog.Warn("import mesh failure on node", "index", i, "name", res.Meshes[i].Name)
			continue
		}
		isAnimated := res.IsTreeAnimated(int(res.Meshes[i].Node.Id))
		postProcData[v.Name] = meshImportPostProcData{
			mesh:         res.Meshes[i],
			kaijuMesh:    kms[i],
			meshes:       res.Meshes,
			isAnimated:   isAnimated,
			textureBytes: res.TextureBytes,
		}
	}
	p.postProcessData = postProcData
	for i := range res.Meshes {
		t := res.Meshes[i].Textures
		p.Dependencies = slices.Grow(p.Dependencies, len(p.Dependencies)+len(t))
		for k, v := range t {
			if strings.HasPrefix(v, "embedded_") {
				continue
			}
			tp := v
			if _, err := os.Stat(tp); err != nil {
				tp = filepath.Join(filepath.Dir(src), v)
			}
			if _, err := os.Stat(tp); err != nil {
				return p, MeshInvalidTextureError{src, v, tp}
			}
			p.Dependencies = klib.AppendUnique(p.Dependencies, tp)
			t[k] = tp
		}
	}
	return p, nil
}

func (c Mesh) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Mesh.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Mesh) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	defer tracing.NewRegion("Mesh.PostImportProcessing").End()
	meshes := proc.postProcessData.(map[string]meshImportPostProcData)
	cc, err := cache.Read(res.Id)
	if err != nil {
		return err
	}
	variant, ok := meshes[cc.Config.Name]
	if !ok {
		slog.Error("failed to locate the mesh in the post processing data", "name", cc.Config.Name)
		return nil
	}
	EnsureMeshBVHInBackground(variant.kaijuMesh, res.ContentPath().String(), fs, res.Id)
	texKeyToDepId := make(map[string]string)
	texKeyToData := make(map[string][]byte)
	for i := range variant.meshes {
		for texType, texKey := range variant.meshes[i].Textures {
			if strings.HasPrefix(texKey, "embedded_") {
				if _, ok := texKeyToData[texKey]; !ok {
					texKeyToData[texKey] = variant.textureBytes[texKey]
				}
				variant.meshes[i].Textures[texType] = texKey
			}
		}
	}
	for texKey, data := range texKeyToData {
		ext := meshEmbeddedTextureExtension(data)
		tf, err := os.CreateTemp("", "*-kaiju-texture"+ext)
		if err != nil {
			continue
		}
		if _, err := tf.Write(data); err != nil {
			tf.Close()
			os.Remove(tf.Name())
			continue
		}
		tf.Close()
		texRes, err := Import(tf.Name(), fs, cache, linkedId)
		if err != nil {
			os.Remove(tf.Name())
			continue
		}
		res.Dependencies = append(res.Dependencies, texRes[0])
		texKeyToDepId[texKey] = texRes[0].Id
		os.Remove(tf.Name())
	}
	for i := range variant.meshes {
		for texType, texKey := range variant.meshes[i].Textures {
			if depId, ok := texKeyToDepId[texKey]; ok {
				variant.meshes[i].Textures[texType] = depId
			} else if strings.HasPrefix(texKey, "embedded_") {
				variant.meshes[i].Textures[texType] = ""
			}
		}
	}
	matchTexture := func(srcPath string) rendering.MaterialTextureData {
		for i := range res.Dependencies {
			cc, err := cache.Read(res.Dependencies[i].Id)
			if err != nil {
				continue
			}
			if fs.NormalizePath(srcPath) == filepath.ToSlash(cc.Config.SrcPath) {
				return rendering.MaterialTextureData{Texture: res.Dependencies[i].Id, Filter: "Linear"}
			}
		}
		return rendering.MaterialTextureData{}
	}
	var mat rendering.MaterialData
	if _, ok := variant.mesh.Textures["metallicRoughness"]; ok {
		mat = rendering.MaterialData{
			Shader:          "pbr.shader",
			RenderPass:      "opaque.renderpass",
			ShaderPipeline:  "basic.shaderpipeline",
			Textures:        make([]rendering.MaterialTextureData, 0, len(variant.mesh.Textures)),
			IsLit:           true,
			ReceivesShadows: true,
			CastsShadows:    true,
		}
		if variant.isAnimated {
			mat.Shader = "pbr_skinned.shader"
		}
		if t, ok := variant.mesh.Textures["baseColor"]; ok {
			mat.Textures = append(mat.Textures, matchTexture(t))
			delete(variant.mesh.Textures, "baseColor")
		} else {
			mat.Textures = append(mat.Textures, rendering.MaterialTextureData{
				Texture: assets.TextureSquare, Filter: "Linear"})
		}
		if t, ok := variant.mesh.Textures["normal"]; ok {
			mat.Textures = append(mat.Textures, matchTexture(t))
			delete(variant.mesh.Textures, "normal")
		} else {
			mat.Textures = append(mat.Textures, rendering.MaterialTextureData{
				Texture: assets.TextureSquare, Filter: "Linear"})
		}
		if t, ok := variant.mesh.Textures["metallicRoughness"]; ok {
			mat.Textures = append(mat.Textures, matchTexture(t))
			delete(variant.mesh.Textures, "metallicRoughness")
		} else {
			mat.Textures = append(mat.Textures, rendering.MaterialTextureData{
				Texture: assets.TextureSquare, Filter: "Linear"})
		}
		if t, ok := variant.mesh.Textures["emissive"]; ok {
			mat.Textures = append(mat.Textures, matchTexture(t))
			delete(variant.mesh.Textures, "emissive")
		} else {
			mat.Textures = append(mat.Textures, rendering.MaterialTextureData{
				Texture: assets.TextureSquare, Filter: "Linear"})
		}
		for _, t := range variant.mesh.Textures {
			mat.Textures = append(mat.Textures, matchTexture(t))
		}
	} else {
		mat = rendering.MaterialData{
			Shader:         "basic.shader",
			RenderPass:     "opaque.renderpass",
			ShaderPipeline: "basic.shaderpipeline",
			Textures:       make([]rendering.MaterialTextureData, 0, len(variant.mesh.Textures)),
		}
		if variant.isAnimated {
			mat.Shader = "basic_skinned.shader"
		}
		for _, t := range variant.mesh.Textures {
			mat.Textures = append(mat.Textures, matchTexture(t))
		}
	}
	// Determine if a matching material already exists
	options := cache.ListByType(Material{}.TypeName())
	// Searching reverse here as the latest additions are more likely to collide
	for i := len(options) - 1; i >= 0; i-- {
		d, err := fs.ReadFile(options[i].ContentPath())
		if err != nil {
			continue
		}
		var dm rendering.MaterialData
		if err = json.Unmarshal(d, &dm); err != nil {
			continue
		}
		same := mat.Shader == dm.Shader &&
			mat.RenderPass == dm.RenderPass &&
			mat.ShaderPipeline == dm.ShaderPipeline &&
			mat.PrepassMaterial == dm.PrepassMaterial &&
			mat.IsLit == dm.IsLit &&
			mat.ReceivesShadows == dm.ReceivesShadows &&
			mat.CastsShadows == dm.CastsShadows &&
			len(mat.Textures) == len(dm.Textures)
		if !same {
			continue
		}
		for j := 0; j < len(mat.Textures) && same; j++ {
			same = mat.Textures[j] == dm.Textures[j]
		}
		if same {
			return nil
		}
	}
	f, err := os.CreateTemp("", "*-kaiju-mat.material")
	if err != nil {
		return err
	}
	tempPath := f.Name()
	defer os.Remove(tempPath)
	if err = json.NewEncoder(f).Encode(mat); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	matRes, err := Import(tempPath, fs, cache, linkedId)
	if err != nil {
		return err
	}
	res.Dependencies = append(res.Dependencies, matRes[0])
	ccMat, err := cache.Read(matRes[0].Id)
	if err != nil {
		return err
	}
	_, err = cache.Rename(ccMat.Id(), fmt.Sprintf("%s_mat", cc.Config.Name), fs)
	if !errors.Is(err, CacheContentNameEqual) {
		return err
	}
	return nil
}

func meshEmbeddedTextureExtension(data []byte) string {
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4e && data[3] == 0x47 {
		return ".png"
	}
	if len(data) >= 2 && data[0] == 0xff && data[1] == 0xd8 {
		return ".jpg"
	}
	if len(data) >= 2 && data[0] == 0x42 && data[1] == 0x4d {
		return ".bmp"
	}
	if len(data) >= 4 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		return ".webp"
	}
	return ".png"
}

func (Mesh) PostReimportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache) error {
	defer tracing.NewRegion("Mesh.PostReimportProcessing").End()
	meshes, ok := proc.postProcessData.(map[string]meshImportPostProcData)
	if !ok || len(proc.Variants) == 0 {
		return nil
	}
	variant, ok := meshes[proc.Variants[0].Name]
	if !ok {
		cc, err := cache.Read(res.Id)
		if err != nil {
			return err
		}
		variant, ok = meshes[cc.Config.SrcName]
		if !ok {
			slog.Error("failed to locate the reimported mesh in the post processing data",
				"name", cc.Config.SrcName)
			return nil
		}
	}
	EnsureMeshBVHInBackground(variant.kaijuMesh, res.ContentPath().String(), fs, res.Id)
	return nil
}

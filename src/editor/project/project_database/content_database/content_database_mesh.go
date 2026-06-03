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
	"runtime"
	"slices"
	"strings"
	"sync"

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

func (Mesh) Path() string                { return project_file_system.ContentMeshFolder }
func (Mesh) TypeName() string            { return "Mesh" }
func (Mesh) ExtNames() []string          { return []string{".glb", ".gltf", ".obj", ".fbx"} }
func (Mesh) StoredExtName(string) string { return ".glb" }

type meshImportPostProcData struct {
	mesh         load_result.Mesh
	kaijuMesh    kaiju_mesh.KaijuMesh
	meshes       []load_result.Mesh
	isAnimated   bool
	textureBytes map[string][]byte
}

func serializeKaijuMeshVariants(kms []kaiju_mesh.KaijuMesh) ([][]byte, error) {
	out := make([][]byte, len(kms))
	if len(kms) == 0 {
		return out, nil
	}
	serialize := func(index int) error {
		kms[index].EnsureBVH()
		data, err := kms[index].Serialize()
		if err != nil {
			return err
		}
		out[index] = data
		return nil
	}
	workers := min(runtime.GOMAXPROCS(0), len(kms))
	if workers <= 1 {
		for i := range kms {
			if err := serialize(i); err != nil {
				return nil, err
			}
		}
		return out, nil
	}
	jobs := make(chan int)
	var firstErr error
	errMutex := sync.Mutex{}
	group := sync.WaitGroup{}
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			for idx := range jobs {
				if err := serialize(idx); err != nil {
					errMutex.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMutex.Unlock()
				}
			}
		}()
	}
	for i := range kms {
		jobs <- i
	}
	close(jobs)
	group.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
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

func writeMeshTextureURIs(km kaiju_mesh.KaijuMesh, path string, fs *project_file_system.FileSystem, textureURIs map[string]string) error {
	data, err := km.SerializeWithOptions(kaiju_mesh.SerializeOptions{TextureURIs: textureURIs})
	if err != nil {
		return err
	}
	return fs.WriteFile(path, data, os.ModePerm)
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
	serializedMeshes, err := serializeKaijuMeshVariants(kms)
	if err != nil {
		return p, err
	}
	postProcData := map[string]meshImportPostProcData{}
	for i := range kms {
		parts := strings.Split(kms[i].Name, "/")
		v := ImportVariant{
			Name: fmt.Sprintf("%s-%s", baseName, parts[len(parts)-1]),
			Data: serializedMeshes[i],
		}
		p.Variants = append(p.Variants, v)
		if res.Meshes[i].Node == nil {
			slog.Warn("import mesh failure on node", "index", i, "name", res.Meshes[i].Name)
			continue
		}
		isAnimated := false
		if nodeIndex := meshNodeIndex(res, res.Meshes[i].Node); nodeIndex >= 0 {
			isAnimated = res.IsTreeAnimated(nodeIndex)
		} else {
			slog.Warn("import mesh failure on node index", "index", i, "name", res.Meshes[i].Name)
		}
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

func meshNodeIndex(res load_result.Result, node *load_result.Node) int {
	id := int(node.Id)
	if id >= 0 && id < len(res.Nodes) && &res.Nodes[id] == node {
		return id
	}
	for i := range res.Nodes {
		if &res.Nodes[i] == node {
			return i
		}
	}
	return -1
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
	textureURIs := meshTextureURIs(variant.mesh.Textures, res, fs, cache, cc.Config.SrcPath)
	if len(textureURIs) > 0 {
		if err := writeMeshTextureURIs(variant.kaijuMesh, res.ContentPath().String(), fs, textureURIs); err != nil {
			slog.Error("failed to write mesh GLB texture references", "id", res.Id, "error", err)
		}
	}
	matchTexture := func(srcPath string) rendering.MaterialTextureData {
		if depId := meshTextureDependencyId(srcPath, res, fs, cache, cc.Config.SrcPath); depId != "" {
			return rendering.MaterialTextureData{Texture: depId, Filter: "Linear"}
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

func meshTextureURIs(textures map[string]string, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, meshSrcPath string) map[string]string {
	out := make(map[string]string, len(textures))
	for slot, texKey := range textures {
		depId := meshTextureDependencyId(texKey, res, fs, cache, meshSrcPath)
		if depId == "" {
			continue
		}
		uri, err := filepath.Rel(filepath.Dir(res.ContentPath().String()),
			project_file_system.AsContentPath(filepath.Join(
				project_file_system.ContentFolder,
				project_file_system.ContentTextureFolder,
				depId)).String())
		if err != nil {
			continue
		}
		out[slot] = filepath.ToSlash(uri)
	}
	return out
}

func meshTextureDependencyId(texKey string, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, meshSrcPath string) string {
	if texKey == "" {
		return ""
	}
	if cc, err := cache.Read(texKey); err == nil && cc.Config.Type == (Texture{}).TypeName() {
		return texKey
	}
	candidates := meshTextureSourceCandidates(texKey, meshSrcPath, fs)
	for i := range res.Dependencies {
		cc, err := cache.Read(res.Dependencies[i].Id)
		if err != nil || cc.Config.Type != (Texture{}).TypeName() {
			continue
		}
		srcPath := filepath.ToSlash(cc.Config.SrcPath)
		if slices.Contains(candidates, srcPath) {
			return res.Dependencies[i].Id
		}
	}
	base := filepath.Base(filepath.ToSlash(texKey))
	if base == "" || base == "." {
		return ""
	}
	match := ""
	for i := range res.Dependencies {
		cc, err := cache.Read(res.Dependencies[i].Id)
		if err != nil || cc.Config.Type != (Texture{}).TypeName() {
			continue
		}
		if filepath.Base(filepath.ToSlash(cc.Config.SrcPath)) != base {
			continue
		}
		if match != "" {
			return ""
		}
		match = res.Dependencies[i].Id
	}
	return match
}

func meshTextureSourceCandidates(texKey, meshSrcPath string, fs *project_file_system.FileSystem) []string {
	candidates := []string{
		filepath.ToSlash(texKey),
		filepath.ToSlash(fs.NormalizePath(texKey)),
	}
	if meshSrcPath != "" && !filepath.IsAbs(texKey) {
		meshDir := filepath.Dir(meshSrcPath)
		candidates = append(candidates,
			filepath.ToSlash(filepath.Join(meshDir, texKey)),
			filepath.ToSlash(fs.NormalizePath(filepath.Join(meshDir, texKey))))
	}
	slices.Sort(candidates)
	return slices.Compact(candidates)
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

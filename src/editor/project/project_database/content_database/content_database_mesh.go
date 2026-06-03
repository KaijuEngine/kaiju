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
	"kaijuengine.com/matrix"
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
type MeshConfig struct {
	Submeshes []MeshSubmeshConfig `json:",omitempty"`
}

type MeshSubmeshConfig struct {
	Key      string
	Name     string
	Material string      `json:",omitempty"`
	Missing  bool        `json:",omitempty"`
	NodeName string      `json:",omitempty"`
	Position matrix.Vec3 `json:",omitempty"`
	Rotation matrix.Vec3 `json:",omitempty"`
	Scale    matrix.Vec3 `json:",omitempty"`
}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Mesh) Path() string                { return project_file_system.ContentMeshFolder }
func (Mesh) TypeName() string            { return "Mesh" }
func (Mesh) ExtNames() []string          { return []string{".glb", ".gltf", ".obj", ".fbx"} }
func (Mesh) StoredExtName(string) string { return ".glb" }

type meshImportPostProcData struct {
	set          kaiju_mesh.KaijuMeshSet
	meshes       []load_result.Mesh
	isAnimated   []bool
	textureBytes map[string][]byte
}

func serializeKaijuMeshSet(set kaiju_mesh.KaijuMeshSet) ([]byte, error) {
	set.EnsureBVH()
	return set.Serialize()
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
	data, readErr := fs.ReadFile(path)
	if readErr == nil && kaiju_mesh.IsGLB(data) {
		if set, err := kaiju_mesh.DeserializeSet(data); err == nil && len(set.Meshes) > 0 {
			ref := kaiju_mesh.ParseMeshRef(id)
			if ref.Key == "" {
				ref.Key = km.Key
			}
			for i := range set.Meshes {
				if ref.Key == "" || set.Meshes[i].Key == ref.Key {
					set.Meshes[i].BVH = km.BVH
					serialized, err := set.Serialize()
					if err != nil {
						slog.Error("failed to serialize the mesh-set BVH", "id", id, "error", err)
						return
					}
					if err = fs.WriteFile(path, serialized, os.ModePerm); err != nil {
						slog.Error("failed to write the mesh-set BVH", "id", id, "path", path, "error", err)
					}
					return
				}
			}
		}
	}
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

func writeMeshSetTextureURIs(set kaiju_mesh.KaijuMeshSet, path string, fs *project_file_system.FileSystem, textureURIs map[string]map[string]string) error {
	data, err := set.SerializeWithOptions(kaiju_mesh.SerializeOptions{MeshTextureURIs: textureURIs})
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
		defer adb.Close()
		if res, err = loaders.GLTF(filepath.Base(src), adb); err != nil {
			return p, err
		}
	case ".obj":
		adb, err := assets.NewFileDatabase(filepath.Dir(src))
		if err != nil {
			return p, err
		}
		defer adb.Close()
		if res, err = loaders.OBJ(filepath.Base(src), adb); err != nil {
			return p, err
		}
	case ".fbx":
		adb, err := assets.NewFileDatabase(filepath.Dir(src))
		if err != nil {
			return p, err
		}
		defer adb.Close()
		if res, err = loaders.FBX(filepath.Base(src), adb); err != nil {
			return p, err
		}
	}
	if len(res.Meshes) == 0 {
		return p, NoMeshesInFileError{Path: src}
	}
	baseName := fileNameNoExt(src)
	set := kaiju_mesh.LoadedResultToKaijuMeshSet(baseName, res)
	serializedSet, err := serializeKaijuMeshSet(set)
	if err != nil {
		return p, err
	}
	isAnimated := make([]bool, len(res.Meshes))
	for i := range res.Meshes {
		if res.Meshes[i].Node == nil {
			slog.Warn("import mesh failure on node", "index", i, "name", res.Meshes[i].Name)
			continue
		}
		if nodeIndex := meshNodeIndex(res, res.Meshes[i].Node); nodeIndex >= 0 {
			isAnimated[i] = res.IsTreeAnimated(nodeIndex)
		} else {
			slog.Warn("import mesh failure on node index", "index", i, "name", res.Meshes[i].Name)
		}
	}
	p.Variants = []ImportVariant{{
		Name: baseName,
		Data: serializedSet,
	}}
	p.postProcessData = meshImportPostProcData{
		set:          set,
		meshes:       res.Meshes,
		isAnimated:   isAnimated,
		textureBytes: res.TextureBytes,
	}
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
	path, err := contentIdToSrcPath(id, cache, fs)
	if err != nil {
		return ProcessedImport{}, err
	}
	proc, err := c.Import(path, fs)
	if err != nil {
		return ProcessedImport{}, err
	}
	cc, err := cache.Read(id)
	if err != nil {
		return ProcessedImport{}, err
	}
	for i := range proc.Variants {
		if proc.Variants[i].Name == cc.Config.SrcName {
			return ProcessedImport{
				Variants:        []ImportVariant{proc.Variants[i]},
				postProcessData: proc.postProcessData,
			}, nil
		}
	}
	data, ok := proc.postProcessData.(meshImportPostProcData)
	if !ok {
		return ProcessedImport{}, ReimportMeshMissingError{
			Path: path,
			Name: cc.Config.SrcName,
		}
	}
	baseName := fileNameNoExt(path)
	for i := range data.set.Meshes {
		if legacySplitMeshVariantName(baseName, data.set.Meshes[i].Name) != cc.Config.SrcName &&
			data.set.Meshes[i].Key != cc.Config.SrcName {
			continue
		}
		single := kaiju_mesh.KaijuMeshSet{
			Name:   data.set.Name,
			Meshes: []kaiju_mesh.KaijuMesh{data.set.Meshes[i]},
		}
		serialized, err := serializeKaijuMeshSet(single)
		if err != nil {
			return ProcessedImport{}, err
		}
		return ProcessedImport{
			Variants: []ImportVariant{{
				Name: cc.Config.SrcName,
				Data: serialized,
			}},
			postProcessData: meshImportPostProcData{
				set:          single,
				textureBytes: data.textureBytes,
			},
		}, nil
	}
	return ProcessedImport{}, ReimportMeshMissingError{
		Path: path,
		Name: cc.Config.SrcName,
	}
}

func legacySplitMeshVariantName(baseName, meshName string) string {
	parts := strings.Split(meshName, "/")
	return fmt.Sprintf("%s-%s", baseName, parts[len(parts)-1])
}

func (Mesh) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	defer tracing.NewRegion("Mesh.PostImportProcessing").End()
	data := proc.postProcessData.(meshImportPostProcData)
	cc, err := cache.Read(res.Id)
	if err != nil {
		return err
	}
	texKeyToDepId := make(map[string]string)
	texKeyToData := make(map[string][]byte)
	for i := range data.meshes {
		for texType, texKey := range data.meshes[i].Textures {
			if strings.HasPrefix(texKey, "embedded_") {
				if _, ok := texKeyToData[texKey]; !ok {
					texKeyToData[texKey] = data.textureBytes[texKey]
				}
				data.meshes[i].Textures[texType] = texKey
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
	for i := range data.meshes {
		for texType, texKey := range data.meshes[i].Textures {
			if depId, ok := texKeyToDepId[texKey]; ok {
				data.meshes[i].Textures[texType] = depId
			} else if strings.HasPrefix(texKey, "embedded_") {
				data.meshes[i].Textures[texType] = ""
			}
		}
	}
	textureURIs := make(map[string]map[string]string, len(data.set.Meshes))
	for i := range data.set.Meshes {
		if i >= len(data.meshes) {
			continue
		}
		uris := meshTextureURIs(data.meshes[i].Textures, res, fs, cache, cc.Config.SrcPath)
		data.set.Meshes[i].Textures = cloneTextureMap(uris)
		if len(uris) > 0 {
			textureURIs[data.set.Meshes[i].Key] = uris
		}
	}
	if err := writeMeshSetTextureURIs(data.set, res.ContentPath().String(), fs, textureURIs); err != nil {
		slog.Error("failed to write mesh GLB texture references", "id", res.Id, "error", err)
	}
	matchTexture := func(srcPath string) rendering.MaterialTextureData {
		if depId := meshTextureDependencyId(srcPath, res, fs, cache, cc.Config.SrcPath); depId != "" {
			return rendering.MaterialTextureData{Texture: depId, Filter: "Linear"}
		}
		return rendering.MaterialTextureData{}
	}
	materials := make(map[string]string, len(data.set.Meshes))
	for i := range data.set.Meshes {
		if i >= len(data.meshes) {
			continue
		}
		mat := meshMaterialData(data.meshes[i].Textures, meshImportIsAnimated(data, i), matchTexture)
		matName := meshImportMaterialName(cc.Config.Name, len(data.set.Meshes), data.set.Meshes[i].Name)
		matId, err := importOrFindMeshMaterial(mat, matName, res, fs, cache, linkedId)
		if err != nil {
			return err
		}
		materials[data.set.Meshes[i].Key] = matId
		data.set.Meshes[i].Material = matId
	}
	cc.Config.Mesh = &MeshConfig{Submeshes: meshConfigSubmeshes(data.set, materials, nil)}
	if err := WriteConfig(cc.Path, cc.Config, fs); err != nil {
		return err
	}
	cache.IndexCachedContent(cc)
	return nil
}

func meshMaterialData(textures map[string]string, isAnimated bool, matchTexture func(string) rendering.MaterialTextureData) rendering.MaterialData {
	remaining := cloneTextureMap(textures)
	var mat rendering.MaterialData
	if _, ok := remaining["metallicRoughness"]; ok {
		mat = rendering.MaterialData{
			Shader:          "pbr.shader",
			RenderPass:      "opaque.renderpass",
			ShaderPipeline:  "basic.shaderpipeline",
			Textures:        make([]rendering.MaterialTextureData, 0, len(remaining)),
			IsLit:           true,
			ReceivesShadows: true,
			CastsShadows:    true,
		}
		if isAnimated {
			mat.Shader = "pbr_skinned.shader"
		}
		for _, slot := range []string{"baseColor", "normal", "metallicRoughness", "emissive"} {
			if tex, ok := remaining[slot]; ok {
				mat.Textures = append(mat.Textures, matchTexture(tex))
				delete(remaining, slot)
			} else {
				mat.Textures = append(mat.Textures, rendering.MaterialTextureData{
					Texture: assets.TextureSquare, Filter: "Linear"})
			}
		}
	} else {
		mat = rendering.MaterialData{
			Shader:         "basic.shader",
			RenderPass:     "opaque.renderpass",
			ShaderPipeline: "basic.shaderpipeline",
			Textures:       make([]rendering.MaterialTextureData, 0, len(remaining)),
		}
		if isAnimated {
			mat.Shader = "basic_skinned.shader"
		}
	}
	keys := make([]string, 0, len(remaining))
	for key := range remaining {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		mat.Textures = append(mat.Textures, matchTexture(remaining[key]))
	}
	return mat
}

func importOrFindMeshMaterial(mat rendering.MaterialData, name string, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) (string, error) {
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
			return options[i].Id(), nil
		}
	}
	f, err := os.CreateTemp("", "*-kaiju-mat.material")
	if err != nil {
		return "", err
	}
	tempPath := f.Name()
	defer os.Remove(tempPath)
	if err = json.NewEncoder(f).Encode(mat); err != nil {
		f.Close()
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	matRes, err := Import(tempPath, fs, cache, linkedId)
	if err != nil {
		return "", err
	}
	res.Dependencies = append(res.Dependencies, matRes[0])
	ccMat, err := cache.Read(matRes[0].Id)
	if err != nil {
		return "", err
	}
	_, err = cache.Rename(ccMat.Id(), name, fs)
	if !errors.Is(err, CacheContentNameEqual) {
		return "", err
	}
	return matRes[0].Id, nil
}

func meshImportMaterialName(assetName string, meshCount int, meshName string) string {
	if meshCount <= 1 || strings.TrimSpace(meshName) == "" {
		return fmt.Sprintf("%s_mat", assetName)
	}
	name := strings.NewReplacer("/", "-", "\\", "-", ":", "-").Replace(meshName)
	return fmt.Sprintf("%s_%s_mat", assetName, name)
}

func meshImportIsAnimated(data meshImportPostProcData, index int) bool {
	return index >= 0 && index < len(data.isAnimated) && data.isAnimated[index]
}

func cloneTextureMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func meshConfigSubmeshes(set kaiju_mesh.KaijuMeshSet, materials map[string]string, old *MeshConfig) []MeshSubmeshConfig {
	oldByKey := map[string]MeshSubmeshConfig{}
	if old != nil {
		for i := range old.Submeshes {
			oldByKey[old.Submeshes[i].Key] = old.Submeshes[i]
		}
	}
	out := make([]MeshSubmeshConfig, 0, len(set.Meshes)+len(oldByKey))
	seen := make(map[string]bool, len(set.Meshes))
	for i := range set.Meshes {
		mesh := &set.Meshes[i]
		mat := materials[mesh.Key]
		if mat == "" {
			if oldSubmesh, ok := oldByKey[mesh.Key]; ok {
				mat = oldSubmesh.Material
			}
		}
		out = append(out, MeshSubmeshConfig{
			Key:      mesh.Key,
			Name:     mesh.Name,
			Material: mat,
			NodeName: mesh.Node.Name,
			Position: mesh.Node.Position,
			Rotation: mesh.Node.Rotation,
			Scale:    mesh.Node.Scale,
		})
		seen[mesh.Key] = true
	}
	if old != nil {
		for i := range old.Submeshes {
			submesh := old.Submeshes[i]
			if seen[submesh.Key] {
				continue
			}
			submesh.Missing = true
			out = append(out, submesh)
		}
	}
	return out
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
	data, ok := proc.postProcessData.(meshImportPostProcData)
	if !ok || len(proc.Variants) == 0 {
		return nil
	}
	cc, err := cache.Read(res.Id)
	if err != nil {
		return err
	}
	preserveMeshSetKeys(&data.set, cc.Config.Mesh)
	data.set.EnsureBVH()
	serialized, err := data.set.Serialize()
	if err != nil {
		return err
	}
	if err := fs.WriteFile(res.ContentPath().String(), serialized, os.ModePerm); err != nil {
		return err
	}
	cc.Config.Mesh = &MeshConfig{Submeshes: meshConfigSubmeshes(data.set, nil, cc.Config.Mesh)}
	if err := WriteConfig(cc.Path, cc.Config, fs); err != nil {
		return err
	}
	cache.IndexCachedContent(cc)
	return nil
}

func preserveMeshSetKeys(set *kaiju_mesh.KaijuMeshSet, old *MeshConfig) {
	if set == nil || old == nil || len(old.Submeshes) == 0 {
		return
	}
	oldByKey := make(map[string]MeshSubmeshConfig, len(old.Submeshes))
	oldByName := make(map[string]MeshSubmeshConfig, len(old.Submeshes))
	oldByNodeName := make(map[string]MeshSubmeshConfig, len(old.Submeshes))
	for i := range old.Submeshes {
		sub := old.Submeshes[i]
		if sub.Key != "" {
			oldByKey[sub.Key] = sub
		}
		if sub.Name != "" {
			if _, exists := oldByName[sub.Name]; !exists {
				oldByName[sub.Name] = sub
			}
		}
		if sub.NodeName != "" {
			if _, exists := oldByNodeName[sub.NodeName]; !exists {
				oldByNodeName[sub.NodeName] = sub
			}
		}
	}
	used := make(map[string]bool, len(set.Meshes))
	for i := range set.Meshes {
		if _, ok := oldByKey[set.Meshes[i].Key]; ok {
			used[set.Meshes[i].Key] = true
		}
	}
	for i := range set.Meshes {
		if used[set.Meshes[i].Key] {
			continue
		}
		oldSubmesh, ok := oldByName[set.Meshes[i].Name]
		if !ok && set.Meshes[i].Node.Name != "" {
			oldSubmesh, ok = oldByNodeName[set.Meshes[i].Node.Name]
		}
		if !ok || oldSubmesh.Key == "" || used[oldSubmesh.Key] {
			used[set.Meshes[i].Key] = true
			continue
		}
		set.Meshes[i].Key = oldSubmesh.Key
		used[oldSubmesh.Key] = true
	}
}

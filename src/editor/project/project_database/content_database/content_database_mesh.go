/******************************************************************************/
/* content_database_mesh.go                                                   */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package content_database

import (
	"encoding/json"
	"errors"
	"fmt"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/assets"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/loaders"
	"kaiju/rendering/loaders/kaiju_mesh"
	"kaiju/rendering/loaders/load_result"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
func (Mesh) ExtNames() []string { return []string{".gltf", ".glb", ".obj"} }

type meshImportPostProcData struct {
	mesh       load_result.Mesh
	isAnimated bool
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
		postProcData[v.Name] = meshImportPostProcData{res.Meshes[i], res.IsTreeAnimated(int(res.Meshes[i].Node.Id))}
	}
	p.postProcessData = postProcData
	for i := range res.Meshes {
		t := res.Meshes[i].Textures
		p.Dependencies = slices.Grow(p.Dependencies, len(p.Dependencies)+len(t))
		for k, v := range t {
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
	if err = json.NewEncoder(f).Encode(mat); err != nil {
		return err
	}
	f.Close()
	matRes, err := Import(f.Name(), fs, cache, linkedId)
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

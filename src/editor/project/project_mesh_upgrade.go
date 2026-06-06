/******************************************************************************/
/* project_mesh_upgrade.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

const meshGLBContentEditorVersion = 0.0019

func (p *Project) upgradeMeshContentToGLB() error {
	if len(p.cacheDatabase.List()) == 0 {
		if err := p.cacheDatabase.Build(&p.fileSystem); err != nil {
			return err
		}
	}
	meshes := p.cacheDatabase.ListByType((content_database.Mesh{}).TypeName())
	for i := range meshes {
		contentPath := meshes[i].ContentPath()
		data, err := p.fileSystem.ReadFile(contentPath)
		if err != nil {
			return err
		}
		if kaiju_mesh.IsGLB(data) {
			continue
		}
		km, err := kaiju_mesh.Deserialize(data)
		if err != nil {
			return err
		}
		textureURIs := p.meshUpgradeTextureURIs(meshes[i])
		out, err := km.SerializeWithOptions(kaiju_mesh.SerializeOptions{TextureURIs: textureURIs})
		if err != nil {
			return err
		}
		if err = p.fileSystem.WriteFile(contentPath, out, os.ModePerm); err != nil {
			return err
		}
		slog.Info("upgraded legacy mesh content to GLB", "id", meshes[i].Id(), "path", contentPath)
	}
	return nil
}

func (p *Project) meshUpgradeTextureURIs(mesh content_database.CachedContent) map[string]string {
	material, ok := p.meshUpgradeMaterial(mesh)
	if !ok {
		return nil
	}
	data, err := p.fileSystem.ReadFile(material.ContentPath())
	if err != nil {
		return nil
	}
	var mat rendering.MaterialData
	if err = json.Unmarshal(data, &mat); err != nil {
		return nil
	}
	out := make(map[string]string, len(mat.Textures))
	slots := []string{"baseColor", "normal", "metallicRoughness", "emissive"}
	for i := range mat.Textures {
		tex := mat.Textures[i]
		if tex.Texture == "" {
			continue
		}
		slot := tex.Label
		if slot == "" && i < len(slots) {
			slot = slots[i]
		}
		if slot == "" {
			continue
		}
		texContent, err := p.cacheDatabase.Read(tex.Texture)
		if err != nil || texContent.Config.Type != (content_database.Texture{}).TypeName() {
			continue
		}
		uri, err := filepath.Rel(filepath.Dir(mesh.ContentPath()), texContent.ContentPath())
		if err != nil {
			continue
		}
		out[slot] = filepath.ToSlash(uri)
	}
	return out
}

func (p *Project) meshUpgradeMaterial(mesh content_database.CachedContent) (content_database.CachedContent, bool) {
	linked, err := p.cacheDatabase.ReadLinked(mesh.Id())
	if err != nil {
		return content_database.CachedContent{}, false
	}
	fallback := -1
	preferredName := mesh.Config.Name + "_mat"
	for i := range linked {
		if linked[i].Config.Type != (content_database.Material{}).TypeName() {
			continue
		}
		if linked[i].Config.Name == preferredName {
			return linked[i], true
		}
		if fallback < 0 {
			fallback = i
		}
	}
	if fallback >= 0 {
		return linked[fallback], true
	}
	return content_database.CachedContent{}, false
}

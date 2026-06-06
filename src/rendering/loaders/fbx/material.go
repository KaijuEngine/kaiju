/******************************************************************************/
/* material.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"path/filepath"
	"strconv"
	"strings"
)

type fbxMaterialResolver struct {
	index        SceneIndex
	sourcePath   string
	textureBytes map[string][]byte
}

func (r fbxMaterialResolver) TexturesForBinding(binding fbxGeometryBinding) map[string]string {
	textures := make(map[string]string)
	material := r.materialForBinding(binding)
	if material == nil {
		return textures
	}
	extraSlot := 0
	for _, connection := range r.index.Connections.ChildrenByParent[material.ID] {
		if connection.Type != "OP" && connection.Type != "OO" {
			continue
		}
		texture := r.index.Texture[connection.Child]
		if texture == nil {
			continue
		}
		textureKey := r.textureKey(texture, connection.Property)
		if textureKey == "" {
			continue
		}
		slot := textureUsageKey(connection.Property)
		if slot == "" {
			slot = textureUsageKey(texture.Name)
		}
		if slot == "" || textures[slot] != "" {
			slot = nextFallbackTextureSlot(textures, &extraSlot)
		}
		textures[slot] = textureKey
	}
	return textures
}

func (r fbxMaterialResolver) materialForBinding(binding fbxGeometryBinding) *Object {
	if binding.modelObject != nil {
		if material := r.firstMaterialChild(binding.modelObject.ID); material != nil {
			return material
		}
	}
	return r.firstMaterialChild(binding.geometry.ID)
}

func (r fbxMaterialResolver) firstMaterialChild(parentID int64) *Object {
	for _, connection := range r.index.Connections.ChildrenByParent[parentID] {
		if connection.Type != "OO" {
			continue
		}
		if material := r.index.Material[connection.Child]; material != nil {
			return material
		}
	}
	return nil
}

func (r fbxMaterialResolver) textureKey(texture *Object, usage string) string {
	if video := r.textureVideo(texture); video != nil {
		if content := objectBytes(video, "Content"); len(content) > 0 {
			key := "embedded_" + strconv.FormatInt(video.ID, 10)
			if slot := textureUsageKey(usage); slot != "" {
				key += "_" + slot
			}
			r.textureBytes[key] = append([]byte(nil), content...)
			return key
		}
		if path := objectString(video, "RelativeFilename", "FileName", "Filename"); path != "" {
			return r.resolveTexturePath(path)
		}
	}
	if path := objectString(texture, "RelativeFilename", "FileName", "Filename"); path != "" {
		return r.resolveTexturePath(path)
	}
	return ""
}

func (r fbxMaterialResolver) textureVideo(texture *Object) *Object {
	for _, connection := range r.index.Connections.ChildrenByParent[texture.ID] {
		if connection.Type != "OO" {
			continue
		}
		if video := r.index.Video[connection.Child]; video != nil {
			return video
		}
	}
	return nil
}

func (r fbxMaterialResolver) resolveTexturePath(path string) string {
	path = cleanFBXPath(path)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) || r.sourcePath == "" {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(filepath.Join(filepath.Dir(r.sourcePath), path))
}

func textureUsageKey(property string) string {
	name := strings.ToLower(strings.TrimSpace(property))
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	switch {
	case name == "diffusecolor" || name == "diffuse" ||
		name == "basecolor" || strings.HasSuffix(name, "|basecolor"):
		return "baseColor"
	case name == "normalmap" || name == "normal" || name == "bump" ||
		name == "bumpmap" || strings.Contains(name, "normalcamera"):
		return "normal"
	case name == "metallicroughness" || strings.Contains(name, "metallicroughness") ||
		name == "metallic" || name == "metalness" || name == "roughness" ||
		strings.HasSuffix(name, "|metallic") || strings.HasSuffix(name, "|roughness"):
		return "metallicRoughness"
	case name == "emissivecolor" || name == "emissive" || name == "incandescence" ||
		strings.HasSuffix(name, "|emissive"):
		return "emissive"
	default:
		return ""
	}
}

func nextFallbackTextureSlot(textures map[string]string, next *int) string {
	for {
		key := "texture" + strconv.Itoa(*next)
		(*next)++
		if textures[key] == "" {
			return key
		}
	}
}

func objectString(object *Object, names ...string) string {
	if object == nil {
		return ""
	}
	for _, name := range names {
		if value, ok := object.Properties.String(name); ok && value != "" {
			return cleanFBXPath(value)
		}
	}
	for _, name := range names {
		if object.Node == nil {
			continue
		}
		if child := childNode(object.Node, name); child != nil && len(child.Properties) > 0 {
			if value, ok := child.Properties[0].Value.(string); ok && value != "" {
				return cleanFBXPath(value)
			}
		}
	}
	return ""
}

func objectBytes(object *Object, name string) []byte {
	if object == nil {
		return nil
	}
	if object.Node != nil {
		if child := childNode(object.Node, name); child != nil && len(child.Properties) > 0 {
			if value, ok := child.Properties[0].Value.([]byte); ok {
				return value
			}
		}
	}
	if prop, ok := object.Properties.Get(name); ok && len(prop.Values) > 0 {
		if value, ok := prop.Values[0].([]byte); ok {
			return value
		}
	}
	return nil
}

func cleanFBXPath(path string) string {
	path = strings.TrimSpace(strings.Trim(path, "\x00"))
	path = strings.ReplaceAll(path, "\\", string(filepath.Separator))
	if path == "." {
		return ""
	}
	return path
}

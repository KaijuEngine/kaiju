package asset_info

import "strings"

var metadataStructureMap = make(map[string]any, 0)

func RegisterMetadataStructure(typeExtension string, metadata any) bool {
	typeExtension = strings.TrimPrefix(typeExtension, ".")
	if _, ok := metadataStructureMap[typeExtension]; ok {
		return false
	}
	metadataStructureMap[typeExtension] = metadata
	return true
}

func init() {
	RegisterMetadataStructure("png", &ImageMetadata{})
	RegisterMetadataStructure("glb", &MeshMetadata{})
	RegisterMetadataStructure("gltf", &MeshMetadata{})
	RegisterMetadataStructure("obj", &MeshMetadata{})
	RegisterMetadataStructure("html", &HtmlMetadata{})
	RegisterMetadataStructure("material", &MaterialMetadata{})
	RegisterMetadataStructure("renderpass", &RenderPassMetadata{})
	RegisterMetadataStructure("shader", &ShaderMetadata{})
	RegisterMetadataStructure("shaderpipeline", &ShaderPipelineMetadata{})
	RegisterMetadataStructure("stage", &StageMetadata{})
}

// Rather than using interfaces, doing casting, and all that nonsense, we're
// just going to have a flat structure for asset metadata that is easy to
// access, easy to fill out, and easy to write. Since this is offline from
// the game, there is no need to try and abstract this out.
type AssetMetadata struct {
	HTML           HtmlMetadata
	Image          ImageMetadata
	Material       MaterialMetadata
	Mesh           MeshMetadata
	RenderPass     RenderPassMetadata
	Shader         ShaderMetadata
	ShaderPipeline ShaderPipelineMetadata
	Stage          StageMetadata
}

func (a *AssetDatabaseInfo) MetadataStructure() any {
	if m, ok := metadataStructureMap[a.Type]; ok {
		return m
	} else {
		return nil
	}
}

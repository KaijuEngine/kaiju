/******************************************************************************/
/* drawing_reader.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package framework

import (
	"fmt"
	"sort"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/load_result"
)

const pbrMaterialKey = assets.MaterialDefinitionPBR
const basicMaterialKey = assets.MaterialDefinitionBasic
const unlitMaterialKey = assets.MaterialDefinitionUnlit
const unlitTransparentMaterialKey = assets.MaterialDefinitionUnlitTransparent

var pbrTextureSlots = []string{"baseColor", "normal", "metallicRoughness", "emissive"}

type ModelDrawing struct {
	Node     *load_result.Node
	MeshName string
	Drawing  rendering.Drawing
}

type ModelDrawingSlice []ModelDrawing

func (s ModelDrawingSlice) AllForNode(node *load_result.Node) []ModelDrawing {
	defer tracing.NewRegion("framework.AllForNode").End()
	part := []ModelDrawing{}
	for i := range s {
		if s[i].Node == node {
			part = append(part, s[i])
		}
	}
	return part
}
func (s ModelDrawingSlice) AllDrawings() []rendering.Drawing {
	defer tracing.NewRegion("framework.AllDrawings").End()
	drawings := make([]rendering.Drawing, len(s))
	for i := range s {
		drawings[i] = s[i].Drawing
	}
	return drawings
}

func createDrawings(host *engine.Host, res load_result.Result, materialKey string, minimumTextures int, textureSlots []string, shaderData func(*load_result.Mesh) rendering.DrawInstance) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.createDrawings").End()
	drawings := ModelDrawingSlice{}
	for i := range res.Meshes {
		m := res.Meshes[i]
		matKey := materialKey
		if matVal, ok := m.Node.Attributes["material"]; ok {
			if mat, ok := matVal.(string); ok {
				matKey = mat
			}
		}
		var tForm matrix.Transform
		tForm.Initialize(host.WorkGroup())
		tForm.SetLocalPosition(m.Node.Position)
		tForm.SetRotation(m.Node.Rotation.ToEuler())
		tForm.SetScale(m.Node.Scale)
		mesh, ok := host.MeshCache().FindMesh(m.MeshName)
		if !ok {
			mesh = rendering.NewMesh(m.MeshName, m.Verts, m.Indexes)
			host.MeshCache().AddMesh(mesh)
		}
		textures := []*rendering.Texture{}
		textureKeys := textureKeysForSlots(m.Textures, textureSlots)
		for i := range textureKeys {
			tex, err := host.TextureCache().Texture(textureKeys[i], rendering.TextureFilterLinear)
			if err != nil {
				return drawings, fmt.Errorf("failed to load mesh texture %q: %w", textureKeys[i], err)
			}
			textures = append(textures, tex)
		}
		for i := len(textures); i < minimumTextures; i++ {
			tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
			if err != nil {
				return drawings, fmt.Errorf("failed to load fallback texture %q: %w", assets.TextureSquare, err)
			}
			textures = append(textures, tex)
		}
		mat, err := host.MaterialCache().Material(matKey)
		if err != nil {
			return drawings, err
		}
		mat = mat.CreateInstance(textures)
		drawings = append(drawings, ModelDrawing{
			Node:     m.Node,
			MeshName: m.Name,
			Drawing: rendering.Drawing{
				Material:   mat,
				Mesh:       mesh,
				Transform:  &tForm,
				ViewCuller: &host.Cameras.Primary,
				ShaderData: shaderData(&m),
			},
		})
	}
	if len(drawings) == 0 {
		return drawings, fmt.Errorf("no drawings to load from the mesh load result")
	}
	return drawings, nil
}

func textureKeysForSlots(textures map[string]string, textureSlots []string) []string {
	if len(textures) == 0 && len(textureSlots) == 0 {
		return nil
	}
	if len(textureSlots) > 0 {
		keys := make([]string, 0, len(textureSlots))
		for i := range textureSlots {
			if key, ok := textures[textureSlots[i]]; ok && key != "" {
				keys = append(keys, key)
			} else {
				keys = append(keys, textureFallbackForSlot(textureSlots[i]))
			}
		}
		return keys
	}
	keys := make([]string, 0, len(textures))
	slots := make([]string, 0, len(textures))
	for slot := range textures {
		slots = append(slots, slot)
	}
	sort.Strings(slots)
	for i := range slots {
		if textures[slots[i]] != "" {
			keys = append(keys, textures[slots[i]])
		}
	}
	return keys
}

func textureFallbackForSlot(slot string) string {
	switch slot {
	case "normal":
		return assets.TexturePBRDefaultNormal
	case "metallicRoughness":
		return assets.TexturePBRDefaultMetallicRough
	case "emissive":
		return assets.TextureBlankSquare
	default:
		return assets.TextureSquare
	}
}

func CreateDrawingsUnlit(host *engine.Host, res load_result.Result) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawingsUnlit").End()
	return createDrawings(host, res, unlitMaterialKey, 1, nil, func(*load_result.Mesh) rendering.DrawInstance {
		return &shader_data_registry.ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			UVs:            matrix.NewVec4(0, 0, 1, 1),
		}
	})
}

func CreateDrawingsUnlitTransparent(host *engine.Host, res load_result.Result) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawingsUnlitTransparent").End()
	return createDrawings(host, res, unlitTransparentMaterialKey, 1, nil, func(*load_result.Mesh) rendering.DrawInstance {
		return &shader_data_registry.ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			UVs:            matrix.NewVec4(0, 0, 1, 1),
		}
	})
}

func CreateDrawingsBasic(host *engine.Host, res load_result.Result) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawingsBasic").End()
	return createDrawings(host, res, basicMaterialKey, 1, nil, func(*load_result.Mesh) rendering.DrawInstance {
		return &shader_data_registry.ShaderDataStandard{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	})
}

func CreateDrawingsPBR(host *engine.Host, res load_result.Result) (ModelDrawingSlice, error) {
	defer tracing.NewRegion("framework.CreateDrawingsPBR").End()
	drawings, err := createDrawings(host, res, pbrMaterialKey, 4, pbrTextureSlots, func(mesh *load_result.Mesh) rendering.DrawInstance {
		return &shader_data_registry.ShaderDataPBR{
			ShaderDataBase: rendering.NewShaderDataBase(),
			VertColors:     matrix.ColorWhite(),
			MeRoEmAo:       matrix.NewVec4(1, 1, 0, 1),
			LightIds:       [...]int32{-1, -1, -1, -1},
		}
	})
	for i := range drawings {
		drawings[i].Drawing.Material.CastsShadows = true
		drawings[i].Drawing.Material.ReceivesShadows = true
		drawings[i].Drawing.Material.IsLit = true
	}
	return drawings, err
}

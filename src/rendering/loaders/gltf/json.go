package gltf

import (
	"encoding/json"
	"kaiju/matrix"
	"strings"
)

type Asset struct {
	Generator string `json:"generator"`
	Version   string `json:"version"`
}

type Scene struct {
	Name  string  `json:"name"`
	Nodes []int32 `json:"nodes"`
}

type Node struct {
	Name string `json:"name"`
	Mesh int32  `json:"mesh"`
}

type TextureId struct {
	Index int32 `json:"index"`
}

type PBRMetallicRoughness struct {
	BaseColorTexture         *TextureId `json:"baseColorTexture"`
	MetallicRoughnessTexture *TextureId `json:"metallicRoughnessTexture"`
	MetallicFactor           float32    `json:"metallicFactor"`
	RoughnessFactor          float32    `json:"roughnessFactor"`
}

type Materials struct {
	Name                 string               `json:"name"`
	DoubleSided          bool                 `json:"doubleSided"`
	NormalTexture        *TextureId           `json:"normalTexture"`
	OcclusionTexture     *TextureId           `json:"occlusionTexture"`
	EmissiveTexture      *TextureId           `json:"emissiveTexture"`
	PBRMetallicRoughness PBRMetallicRoughness `json:"pbrMetallicRoughness"`
}

type Target struct {
	POSITION   *int32 `json:"POSITION"`
	NORMAL     *int32 `json:"NORMAL"`
	TANGENT    *int32 `json:"TANGENT"`
	TEXCOORD_0 *int32 `json:"TEXCOORD_0"`
	TEXCOORD_1 *int32 `json:"TEXCOORD_1"`
	COLOR_0    *int32 `json:"COLOR_0"`
	COLOR_1    *int32 `json:"COLOR_1"`
}

type Primitive struct {
	Attributes map[string]uint32 `json:"attributes"`
	Indices    int32             `json:"indices"`
	Material   *int32            `json:"material"`
	Mode       int32             `json:"mode"`
	Targets    []Target          `json:"targets"`
	Extensions interface{}       `json:"extensions"`
	Extras     interface{}       `json:"extras"`
}

type Mesh struct {
	Name       string      `json:"name"`
	Primitives []Primitive `json:"primitives"`
}

type Texture struct {
	Sampler int32 `json:"sampler"`
	Source  int32 `json:"source"`
}

type Image struct {
	Name     string `json:"name"`
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
}

type Accessor struct {
	BufferView    int32         `json:"bufferView"`
	ComponentType ComponentType `json:"componentType"`
	Count         int32         `json:"count"`
	Max           matrix.Vec3   `json:"max"`
	Min           matrix.Vec3   `json:"min"`
	Type          AccessorType  `json:"type"`
}

type BufferView struct {
	Buffer     int32 `json:"buffer"`
	ByteLength int32 `json:"byteLength"`
	ByteOffset int32 `json:"byteOffset"`
	Target     int32 `json:"target"`
}

type Sampler struct {
	MagFilter int32 `json:"magFilter"`
	MinFilter int32 `json:"minFilter"`
	WrapS     int32 `json:"wrapS"`
	WrapT     int32 `json:"wrapT"`
}

type Buffer struct {
	ByteLength int32  `json:"byteLength"`
	URI        string `json:"uri"`
}

type GLTF struct {
	Asset       Asset        `json:"asset"`
	Scene       int32        `json:"scene"`
	Scenes      []Scene      `json:"scenes"`
	Nodes       []Node       `json:"nodes"`
	Materials   []Materials  `json:"materials"`
	Meshes      []Mesh       `json:"meshes"`
	Textures    []Texture    `json:"textures"`
	Images      []Image      `json:"images"`
	Accessors   []Accessor   `json:"accessors"`
	BufferViews []BufferView `json:"bufferViews"`
	Samplers    []Sampler    `json:"samplers"`
	Buffers     []Buffer     `json:"buffers"`
}

func LoadGLTF(jsonStr string) (GLTF, error) {
	var gltf GLTF
	err := json.NewDecoder(strings.NewReader(jsonStr)).Decode(&gltf)
	return gltf, err
}

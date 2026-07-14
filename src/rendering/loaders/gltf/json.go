/******************************************************************************/
/* json.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gltf

import (
	"encoding/json"
	"strings"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/loaders/load_result"
)

type Asset struct {
	Generator string `json:"generator"`
	Version   string `json:"version"`
	FilePath  string
}

type Scene struct {
	Name  string  `json:"name"`
	Nodes []int32 `json:"nodes"`
}

type Node struct {
	Name        string         `json:"name"`
	Children    []int32        `json:"children"`
	Mesh        *int32         `json:"mesh"`
	Camera      *int32         `json:"camera"`
	Skin        *int32         `json:"skin"`
	Matrix      *matrix.Mat4   `json:"matrix"`
	Rotation    *matrix.Vec4   `json:"rotation"` // Vec4 because glTF XYZW on quat
	Scale       *matrix.Vec3   `json:"scale"`
	Translation *matrix.Vec3   `json:"translation"`
	Weights     []float32      `json:"weights"`
	Extras      map[string]any `json:"extras"`
	//Extensions  interface{}       `json:"extensions"`
}

type ChannelTarget struct {
	Node    int32  `json:"node"`
	PathStr string `json:"path"`
}

func (t *ChannelTarget) Path() load_result.AnimationPathType {
	switch t.PathStr {
	case "translation":
		return load_result.AnimPathTranslation
	case "rotation":
		return load_result.AnimPathRotation
	case "scale":
		return load_result.AnimPathScale
	case "weights":
		return load_result.AnimPathWeights
	}
	return -1
}

type AnimationChannel struct {
	Sampler int32         `json:"sampler"`
	Target  ChannelTarget `json:"target"`
}

type AnimationSampler struct {
	Input            int32  `json:"input"`
	InterpolationStr string `json:"interpolation"`
	Output           int32  `json:"output"`
}

func (a *AnimationSampler) Interpolation() load_result.AnimationInterpolation {
	switch a.InterpolationStr {
	case "LINEAR":
		return load_result.AnimInterpolateLinear
	case "STEP":
		return load_result.AnimInterpolateStep
	case "CUBICSPLINE":
		return load_result.AnimInterpolateCubicSpline
	}
	return -1
}

type Animation struct {
	Name     string             `json:"name"`
	Channels []AnimationChannel `json:"channels"`
	Samplers []AnimationSampler `json:"samplers"`
}

type TextureId struct {
	Index int32 `json:"index"`
}

type PBRMetallicRoughness struct {
	BaseColorTexture         *TextureId    `json:"baseColorTexture"`
	MetallicRoughnessTexture *TextureId    `json:"metallicRoughnessTexture"`
	MetallicFactor           float32       `json:"metallicFactor"`
	RoughnessFactor          float32       `json:"roughnessFactor"`
	BaseColorFactor          *matrix.Color `json:"baseColorFactor"`
}

type Material struct {
	Name                 string               `json:"name"`
	DoubleSided          bool                 `json:"doubleSided"`
	AlphaMode            string               `json:"alphaMode"`
	AlphaCutoff          *float32             `json:"alphaCutoff"`
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

type Skin struct {
	Name                string  `json:"name"`
	InverseBindMatrices int32   `json:"inverseBindMatrices"`
	Joints              []int32 `json:"joints"`
}

type Texture struct {
	Sampler int32 `json:"sampler"`
	Source  int32 `json:"source"`
}

type Image struct {
	Name       string `json:"name"`
	URI        string `json:"uri"`
	MimeType   string `json:"mimeType"`
	BufferView *int32 `json:"bufferView"`
}

type Accessor struct {
	BufferView    int32         `json:"bufferView"`
	ByteOffset    int32         `json:"byteOffset"`
	ComponentType ComponentType `json:"componentType"`
	Count         int32         `json:"count"`
	Normalized    bool          `json:"normalized"`
	Max           matrix.Vec3   `json:"max"`
	Min           matrix.Vec3   `json:"min"`
	Type          AccessorType  `json:"type"`
}

type BufferView struct {
	Buffer     int32 `json:"buffer"`
	ByteLength int32 `json:"byteLength"`
	ByteOffset int32 `json:"byteOffset"`
	ByteStride int32 `json:"byteStride"`
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
	Asset          Asset        `json:"asset"`
	ExtensionsUsed []string     `json:"extensionsUsed"`
	Scene          int32        `json:"scene"`
	Scenes         []Scene      `json:"scenes"`
	Nodes          []Node       `json:"nodes"`
	Animations     []Animation  `json:"animations"`
	Materials      []Material   `json:"materials"`
	Meshes         []Mesh       `json:"meshes"`
	Skins          []Skin       `json:"skins"`
	Textures       []Texture    `json:"textures"`
	Images         []Image      `json:"images"`
	Accessors      []Accessor   `json:"accessors"`
	BufferViews    []BufferView `json:"bufferViews"`
	Samplers       []Sampler    `json:"samplers"`
	Buffers        []Buffer     `json:"buffers"`
}

func LoadGLTF(jsonStr string) (GLTF, error) {
	var gltf GLTF
	err := json.NewDecoder(strings.NewReader(jsonStr)).Decode(&gltf)
	return gltf, err
}

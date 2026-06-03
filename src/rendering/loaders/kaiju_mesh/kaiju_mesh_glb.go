/******************************************************************************/
/* kaiju_mesh_glb.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package kaiju_mesh

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	meshloaders "kaijuengine.com/rendering/loaders"
)

const (
	glbJSONChunkType = 0x4e4f534a
	glbBINChunkType  = 0x004e4942

	glbArrayBufferTarget        = 34962
	glbElementArrayBufferTarget = 34963

	glbComponentUnsignedShort = 5123
	glbComponentUnsignedInt   = 5125
	glbComponentFloat         = 5126

	glbTypeScalar = "SCALAR"
	glbTypeVec2   = "VEC2"
	glbTypeVec3   = "VEC3"
	glbTypeVec4   = "VEC4"
	glbTypeMat4   = "MAT4"
)

var glbMagic = [4]byte{'g', 'l', 'T', 'F'}

type glbAsset struct {
	Generator string `json:"generator,omitempty"`
	Version   string `json:"version"`
}

type glbScene struct {
	Name  string `json:"name,omitempty"`
	Nodes []int  `json:"nodes,omitempty"`
}

type glbNode struct {
	Name        string    `json:"name,omitempty"`
	Children    []int     `json:"children,omitempty"`
	Mesh        *int      `json:"mesh,omitempty"`
	Skin        *int      `json:"skin,omitempty"`
	Rotation    []float32 `json:"rotation,omitempty"`
	Scale       []float32 `json:"scale,omitempty"`
	Translation []float32 `json:"translation,omitempty"`
}

type glbPrimitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    int            `json:"indices"`
	Material   *int           `json:"material,omitempty"`
	Mode       int            `json:"mode"`
	Targets    []glbTarget    `json:"targets,omitempty"`
}

type glbTarget map[string]int

type glbMesh struct {
	Name       string         `json:"name,omitempty"`
	Primitives []glbPrimitive `json:"primitives"`
}

type glbTextureID struct {
	Index int `json:"index"`
}

type glbPBRMetallicRoughness struct {
	BaseColorTexture         *glbTextureID `json:"baseColorTexture,omitempty"`
	MetallicRoughnessTexture *glbTextureID `json:"metallicRoughnessTexture,omitempty"`
	BaseColorFactor          []float32     `json:"baseColorFactor,omitempty"`
	MetallicFactor           float32       `json:"metallicFactor,omitempty"`
	RoughnessFactor          float32       `json:"roughnessFactor,omitempty"`
}

type glbMaterial struct {
	Name                 string                  `json:"name,omitempty"`
	NormalTexture        *glbTextureID           `json:"normalTexture,omitempty"`
	OcclusionTexture     *glbTextureID           `json:"occlusionTexture,omitempty"`
	EmissiveTexture      *glbTextureID           `json:"emissiveTexture,omitempty"`
	PBRMetallicRoughness glbPBRMetallicRoughness `json:"pbrMetallicRoughness,omitempty"`
}

type glbTexture struct {
	Source int `json:"source"`
}

type glbImage struct {
	Name string `json:"name,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type glbAccessor struct {
	BufferView    int       `json:"bufferView"`
	ByteOffset    int       `json:"byteOffset,omitempty"`
	ComponentType int       `json:"componentType"`
	Count         int       `json:"count"`
	Max           []float32 `json:"max,omitempty"`
	Min           []float32 `json:"min,omitempty"`
	Type          string    `json:"type"`
}

type glbBufferView struct {
	Buffer     int `json:"buffer"`
	ByteLength int `json:"byteLength"`
	ByteOffset int `json:"byteOffset,omitempty"`
	ByteStride int `json:"byteStride,omitempty"`
	Target     int `json:"target,omitempty"`
}

type glbBuffer struct {
	ByteLength int `json:"byteLength"`
}

type glbSkin struct {
	Name                string `json:"name,omitempty"`
	InverseBindMatrices int    `json:"inverseBindMatrices"`
	Joints              []int  `json:"joints"`
}

type glbAnimationChannelTarget struct {
	Node int    `json:"node"`
	Path string `json:"path"`
}

type glbAnimationChannel struct {
	Sampler int                       `json:"sampler"`
	Target  glbAnimationChannelTarget `json:"target"`
}

type glbAnimationSampler struct {
	Input         int    `json:"input"`
	Interpolation string `json:"interpolation,omitempty"`
	Output        int    `json:"output"`
}

type glbAnimation struct {
	Name     string                `json:"name,omitempty"`
	Channels []glbAnimationChannel `json:"channels,omitempty"`
	Samplers []glbAnimationSampler `json:"samplers,omitempty"`
}

type glbDocument struct {
	Asset       glbAsset        `json:"asset"`
	Scene       int             `json:"scene"`
	Scenes      []glbScene      `json:"scenes,omitempty"`
	Nodes       []glbNode       `json:"nodes,omitempty"`
	Animations  []glbAnimation  `json:"animations,omitempty"`
	Materials   []glbMaterial   `json:"materials,omitempty"`
	Meshes      []glbMesh       `json:"meshes,omitempty"`
	Skins       []glbSkin       `json:"skins,omitempty"`
	Textures    []glbTexture    `json:"textures,omitempty"`
	Images      []glbImage      `json:"images,omitempty"`
	Accessors   []glbAccessor   `json:"accessors,omitempty"`
	BufferViews []glbBufferView `json:"bufferViews,omitempty"`
	Buffers     []glbBuffer     `json:"buffers,omitempty"`
	Extras      *glbExtras      `json:"extras,omitempty"`
}

type glbExtras struct {
	Kaiju glbKaijuExtras `json:"kaiju"`
}

type glbKaijuExtras struct {
	Version int          `json:"version"`
	Blobs   *glbBlobRefs `json:"blobs,omitempty"`
}

type glbBlobRefs struct {
	TriangleBVH *glbBlobRef `json:"triangleBVH,omitempty"`
}

type glbBlobRef struct {
	BufferView *int   `json:"bufferView,omitempty"`
	Format     string `json:"format,omitempty"`
}

type glbWriter struct {
	doc glbDocument
	bin []byte
}

type singleAssetDatabase struct {
	key  string
	data []byte
}

func (singleAssetDatabase) Cache(string, []byte)                                 {}
func (singleAssetDatabase) CacheRemove(string)                                   {}
func (singleAssetDatabase) CacheClear()                                          {}
func (singleAssetDatabase) Close()                                               {}
func (singleAssetDatabase) PostWindowCreate(assets.PostWindowCreateHandle) error { return nil }

func (d singleAssetDatabase) Exists(key string) bool { return key == d.key }

func (d singleAssetDatabase) Read(key string) ([]byte, error) {
	if key != d.key {
		return nil, fmt.Errorf("asset %q not found", key)
	}
	return d.data, nil
}

func (d singleAssetDatabase) ReadText(key string) (string, error) {
	data, err := d.Read(key)
	return string(data), err
}

func IsGLB(data []byte) bool {
	return len(data) >= 4 && bytes.Equal(data[:4], glbMagic[:])
}

func serializeGLB(k KaijuMesh, options SerializeOptions) ([]byte, error) {
	if len(k.Verts) == 0 {
		return nil, errors.New("kaiju glb serialization requires vertices")
	}
	if len(k.Indexes) == 0 {
		return nil, errors.New("kaiju glb serialization requires indexes")
	}
	w := glbWriter{
		doc: glbDocument{
			Asset: glbAsset{
				Generator: "Kaiju Engine",
				Version:   "2.0",
			},
			Scene: 0,
		},
	}
	w.doc.Buffers = []glbBuffer{{}}
	primitive, err := w.meshPrimitive(k, options)
	if err != nil {
		return nil, err
	}
	w.doc.Meshes = []glbMesh{{
		Name:       k.Name,
		Primitives: []glbPrimitive{primitive},
	}}
	meshNode := w.buildNodes(k)
	w.doc.Nodes[meshNode].Mesh = ptrInt(0)
	if len(w.doc.Skins) > 0 {
		w.doc.Nodes[meshNode].Skin = ptrInt(0)
	}
	w.doc.Scenes = []glbScene{{
		Name:  "Scene",
		Nodes: glbSceneNodes(w.doc.Nodes, meshNode),
	}}
	if k.BVH != nil {
		view := w.addBufferView(serializeTriangleBVHBlob(k.BVH), 0)
		w.doc.Extras = &glbExtras{Kaiju: glbKaijuExtras{
			Version: 1,
			Blobs: &glbBlobRefs{TriangleBVH: &glbBlobRef{
				BufferView: ptrInt(view),
				Format:     "kaiju.triangle_bvh.le.v1",
			}},
		}}
	} else {
		w.doc.Extras = &glbExtras{Kaiju: glbKaijuExtras{Version: 1}}
	}
	w.doc.Buffers[0].ByteLength = len(w.bin)
	jsonBytes, err := json.Marshal(w.doc)
	if err != nil {
		return nil, err
	}
	return encodeGLB(jsonBytes, w.bin), nil
}

func (w *glbWriter) meshPrimitive(k KaijuMesh, options SerializeOptions) (glbPrimitive, error) {
	attrs := map[string]int{}
	attrs["POSITION"] = w.addVec3Accessor(vertexVec3Bytes(k.Verts, func(v rendering.Vertex) matrix.Vec3 {
		return v.Position
	}), glbArrayBufferTarget, k.Verts, func(v rendering.Vertex) matrix.Vec3 {
		return v.Position
	})
	attrs["NORMAL"] = w.addVec3Accessor(vertexVec3Bytes(k.Verts, func(v rendering.Vertex) matrix.Vec3 {
		return v.Normal
	}), glbArrayBufferTarget, k.Verts, func(v rendering.Vertex) matrix.Vec3 {
		return v.Normal
	})
	attrs["TANGENT"] = w.addAccessor(
		w.addBufferView(vertexVec4Bytes(k.Verts, func(v rendering.Vertex) matrix.Vec4 {
			return v.Tangent
		}), glbArrayBufferTarget),
		glbComponentFloat, len(k.Verts), glbTypeVec4, nil, nil)
	attrs["TEXCOORD_0"] = w.addAccessor(
		w.addBufferView(vertexVec2Bytes(k.Verts, func(v rendering.Vertex) matrix.Vec2 {
			return v.UV0
		}), glbArrayBufferTarget),
		glbComponentFloat, len(k.Verts), glbTypeVec2, nil, nil)
	attrs["COLOR_0"] = w.addAccessor(
		w.addBufferView(vertexColorBytes(k.Verts), glbArrayBufferTarget),
		glbComponentFloat, len(k.Verts), glbTypeVec4, nil, nil)
	if shouldWriteSkinAttributes(k) {
		jointBytes, err := vertexJointBytes(k.Verts)
		if err != nil {
			return glbPrimitive{}, err
		}
		attrs["JOINTS_0"] = w.addAccessor(
			w.addBufferView(jointBytes, glbArrayBufferTarget),
			glbComponentUnsignedShort, len(k.Verts), glbTypeVec4, nil, nil)
		attrs["WEIGHTS_0"] = w.addAccessor(
			w.addBufferView(vertexVec4Bytes(k.Verts, func(v rendering.Vertex) matrix.Vec4 {
				return v.JointWeights
			}), glbArrayBufferTarget),
			glbComponentFloat, len(k.Verts), glbTypeVec4, nil, nil)
	}
	indices := w.addAccessor(
		w.addBufferView(indexBytes(k.Indexes), glbElementArrayBufferTarget),
		glbComponentUnsignedInt, len(k.Indexes), glbTypeScalar, nil, nil)
	primitive := glbPrimitive{
		Attributes: attrs,
		Indices:    indices,
		Mode:       4,
	}
	if shouldWriteMorphTarget(k.Verts) {
		targetAccessor := w.addVec3Accessor(vertexVec3Bytes(k.Verts, func(v rendering.Vertex) matrix.Vec3 {
			return v.MorphTarget
		}), glbArrayBufferTarget, k.Verts, func(v rendering.Vertex) matrix.Vec3 {
			return v.MorphTarget
		})
		primitive.Targets = []glbTarget{{"POSITION": targetAccessor}}
	}
	if material := w.addMaterial(k, options); material != nil {
		primitive.Material = material
	}
	return primitive, nil
}

func (w *glbWriter) buildNodes(k KaijuMesh) int {
	maxNode := -1
	for i := range k.Joints {
		maxNode = max(maxNode, int(k.Joints[i].Id), int(k.Joints[i].Parent))
	}
	for i := range k.Animations {
		for j := range k.Animations[i].Frames {
			for b := range k.Animations[i].Frames[j].Bones {
				maxNode = max(maxNode, k.Animations[i].Frames[j].Bones[b].NodeIndex)
			}
		}
	}
	meshNode := maxNode + 1
	if meshNode < 0 {
		meshNode = 0
	}
	w.doc.Nodes = make([]glbNode, meshNode+1)
	for i := 0; i < meshNode; i++ {
		w.doc.Nodes[i].Name = fmt.Sprintf("Node_%d", i)
	}
	w.doc.Nodes[meshNode].Name = k.Name
	for i := range k.Joints {
		j := &k.Joints[i]
		if j.Id < 0 || int(j.Id) >= len(w.doc.Nodes) {
			continue
		}
		node := &w.doc.Nodes[j.Id]
		node.Name = fmt.Sprintf("Joint_%d", j.Id)
		node.Translation = vec3JSON(j.Position)
		node.Scale = vec3JSON(j.Scale)
		q := matrix.QuaternionFromEuler(j.Rotation)
		node.Rotation = quatXYZWJSON(q)
		if j.Parent >= 0 && int(j.Parent) < len(w.doc.Nodes) {
			w.doc.Nodes[j.Parent].Children = appendUniqueInt(w.doc.Nodes[j.Parent].Children, int(j.Id))
		}
	}
	if len(k.Joints) > 0 {
		invBind := make([]byte, 0, len(k.Joints)*16*4)
		jointNodes := make([]int, len(k.Joints))
		for i := range k.Joints {
			jointNodes[i] = int(k.Joints[i].Id)
			for _, v := range k.Joints[i].Skin {
				invBind = appendF32(invBind, float32(v))
			}
		}
		ibm := w.addAccessor(w.addBufferView(invBind, glbArrayBufferTarget),
			glbComponentFloat, len(k.Joints), glbTypeMat4, nil, nil)
		w.doc.Skins = []glbSkin{{
			Name:                "Skin",
			InverseBindMatrices: ibm,
			Joints:              jointNodes,
		}}
	}
	w.addAnimations(k)
	return meshNode
}

func (w *glbWriter) addAnimations(k KaijuMesh) {
	for i := range k.Animations {
		anim := &k.Animations[i]
		out := glbAnimation{Name: anim.Name}
		absTimes := animationAbsoluteTimes(anim)
		type channelKey struct {
			node          int
			path          AnimationPathType
			interpolation AnimationInterpolation
		}
		groups := make(map[channelKey][]int)
		for f := range anim.Frames {
			for b := range anim.Frames[f].Bones {
				bone := &anim.Frames[f].Bones[b]
				key := channelKey{
					node:          bone.NodeIndex,
					path:          bone.PathType,
					interpolation: bone.Interpolation,
				}
				groups[key] = append(groups[key], f)
			}
		}
		keys := make([]channelKey, 0, len(groups))
		for key := range groups {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(a, b int) bool {
			if keys[a].node == keys[b].node {
				return keys[a].path < keys[b].path
			}
			return keys[a].node < keys[b].node
		})
		for _, key := range keys {
			frames := groups[key]
			if len(frames) == 0 {
				continue
			}
			timeBytes := make([]byte, 0, len(frames)*4)
			valueBytes := make([]byte, 0, len(frames)*16)
			for _, frameIndex := range frames {
				timeBytes = appendF32(timeBytes, absTimes[frameIndex])
				bone := findAnimationBone(&anim.Frames[frameIndex], key.node, key.path, key.interpolation)
				if bone == nil {
					continue
				}
				switch key.path {
				case AnimPathTranslation, AnimPathScale:
					valueBytes = appendF32(valueBytes, float32(bone.Data[0]))
					valueBytes = appendF32(valueBytes, float32(bone.Data[1]))
					valueBytes = appendF32(valueBytes, float32(bone.Data[2]))
				case AnimPathRotation:
					q := matrix.Quaternion(bone.Data)
					valueBytes = appendF32(valueBytes, float32(q.X()))
					valueBytes = appendF32(valueBytes, float32(q.Y()))
					valueBytes = appendF32(valueBytes, float32(q.Z()))
					valueBytes = appendF32(valueBytes, float32(q.W()))
				default:
					continue
				}
			}
			path, accessorType := animationPathString(key.path)
			if path == "" {
				continue
			}
			input := w.addAccessor(w.addBufferView(timeBytes, glbArrayBufferTarget),
				glbComponentFloat, len(frames), glbTypeScalar, nil, nil)
			output := w.addAccessor(w.addBufferView(valueBytes, glbArrayBufferTarget),
				glbComponentFloat, len(frames), accessorType, nil, nil)
			sampler := len(out.Samplers)
			out.Samplers = append(out.Samplers, glbAnimationSampler{
				Input:         input,
				Interpolation: animationInterpolationString(key.interpolation),
				Output:        output,
			})
			out.Channels = append(out.Channels, glbAnimationChannel{
				Sampler: sampler,
				Target: glbAnimationChannelTarget{
					Node: key.node,
					Path: path,
				},
			})
		}
		if len(out.Channels) > 0 {
			w.doc.Animations = append(w.doc.Animations, out)
		}
	}
}

func (w *glbWriter) addMaterial(k KaijuMesh, options SerializeOptions) *int {
	textureURIs := k.Textures
	if len(options.TextureURIs) > 0 {
		textureURIs = options.TextureURIs
	}
	if len(textureURIs) == 0 {
		return nil
	}
	textureID := func(slot string) *glbTextureID {
		uri := textureURIs[slot]
		if uri == "" {
			return nil
		}
		image := len(w.doc.Images)
		w.doc.Images = append(w.doc.Images, glbImage{Name: slot, URI: uri})
		texture := len(w.doc.Textures)
		w.doc.Textures = append(w.doc.Textures, glbTexture{Source: image})
		return &glbTextureID{Index: texture}
	}
	mat := glbMaterial{
		Name: "Kaiju Material",
		PBRMetallicRoughness: glbPBRMetallicRoughness{
			BaseColorFactor: []float32{1, 1, 1, 1},
			RoughnessFactor: 1,
		},
	}
	mat.PBRMetallicRoughness.BaseColorTexture = textureID("baseColor")
	mat.PBRMetallicRoughness.MetallicRoughnessTexture = textureID("metallicRoughness")
	mat.NormalTexture = textureID("normal")
	mat.OcclusionTexture = textureID("occlusion")
	mat.EmissiveTexture = textureID("emissive")
	if mat.PBRMetallicRoughness.BaseColorTexture == nil &&
		mat.PBRMetallicRoughness.MetallicRoughnessTexture == nil &&
		mat.NormalTexture == nil &&
		mat.OcclusionTexture == nil &&
		mat.EmissiveTexture == nil {
		return nil
	}
	idx := len(w.doc.Materials)
	w.doc.Materials = append(w.doc.Materials, mat)
	return &idx
}

func (w *glbWriter) addVec3Accessor(data []byte, target int, verts []rendering.Vertex, value func(rendering.Vertex) matrix.Vec3) int {
	minV, maxV := vertexVec3MinMax(verts, value)
	return w.addAccessor(w.addBufferView(data, target), glbComponentFloat, len(verts),
		glbTypeVec3, vec3JSON(minV), vec3JSON(maxV))
}

func (w *glbWriter) addBufferView(data []byte, target int) int {
	w.alignBin()
	offset := len(w.bin)
	w.bin = append(w.bin, data...)
	view := glbBufferView{
		Buffer:     0,
		ByteOffset: offset,
		ByteLength: len(data),
		Target:     target,
	}
	idx := len(w.doc.BufferViews)
	w.doc.BufferViews = append(w.doc.BufferViews, view)
	return idx
}

func (w *glbWriter) addAccessor(view, componentType, count int, accessorType string, minV, maxV []float32) int {
	idx := len(w.doc.Accessors)
	w.doc.Accessors = append(w.doc.Accessors, glbAccessor{
		BufferView:    view,
		ComponentType: componentType,
		Count:         count,
		Min:           minV,
		Max:           maxV,
		Type:          accessorType,
	})
	return idx
}

func (w *glbWriter) alignBin() {
	for len(w.bin)%4 != 0 {
		w.bin = append(w.bin, 0)
	}
}

func deserializeGLB(data []byte) (KaijuMesh, error) {
	db := singleAssetDatabase{key: "mesh.glb", data: data}
	res, err := meshloaders.GLTF("mesh.glb", db)
	if err != nil {
		return KaijuMesh{}, err
	}
	meshes := LoadedResultToKaijuMesh(res)
	if len(meshes) == 0 {
		return KaijuMesh{}, errors.New("glb contains no meshes")
	}
	_, doc, bin, err := decodeGLB(data)
	if err != nil {
		return KaijuMesh{}, err
	}
	if bvh, err := glbTriangleBVH(&doc, bin); err != nil {
		return KaijuMesh{}, err
	} else if bvh != nil {
		meshes[0].BVH = bvh
	}
	if len(doc.Meshes) > 0 && doc.Meshes[0].Name != "" {
		meshes[0].Name = doc.Meshes[0].Name
	}
	return meshes[0], nil
}

func encodeGLB(jsonBytes, bin []byte) []byte {
	jsonBytes = appendPadded(jsonBytes, ' ')
	bin = appendPadded(bin, 0)
	totalLen := 12 + 8 + len(jsonBytes) + 8 + len(bin)
	out := make([]byte, 0, totalLen)
	out = append(out, glbMagic[:]...)
	out = binary.LittleEndian.AppendUint32(out, 2)
	out = binary.LittleEndian.AppendUint32(out, uint32(totalLen))
	out = binary.LittleEndian.AppendUint32(out, uint32(len(jsonBytes)))
	out = binary.LittleEndian.AppendUint32(out, glbJSONChunkType)
	out = append(out, jsonBytes...)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(bin)))
	out = binary.LittleEndian.AppendUint32(out, glbBINChunkType)
	out = append(out, bin...)
	return out
}

func decodeGLB(data []byte) ([]byte, glbDocument, []byte, error) {
	if len(data) < 20 || !IsGLB(data) {
		return nil, glbDocument{}, nil, errors.New("invalid glb file")
	}
	if version := binary.LittleEndian.Uint32(data[4:8]); version != 2 {
		return nil, glbDocument{}, nil, fmt.Errorf("unsupported glb version %d", version)
	}
	if length := binary.LittleEndian.Uint32(data[8:12]); int(length) > len(data) {
		return nil, glbDocument{}, nil, errors.New("glb length exceeds data")
	}
	pos := 12
	jsonLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
	jsonType := binary.LittleEndian.Uint32(data[pos+4 : pos+8])
	pos += 8
	if jsonType != glbJSONChunkType || pos+jsonLen > len(data) {
		return nil, glbDocument{}, nil, errors.New("invalid glb JSON chunk")
	}
	jsonBytes := bytes.TrimRight(data[pos:pos+jsonLen], " ")
	pos += jsonLen
	if pos+8 > len(data) {
		return jsonBytes, glbDocument{}, nil, errors.New("missing glb BIN chunk")
	}
	binLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
	binType := binary.LittleEndian.Uint32(data[pos+4 : pos+8])
	pos += 8
	if binType != glbBINChunkType || pos+binLen > len(data) {
		return nil, glbDocument{}, nil, errors.New("invalid glb BIN chunk")
	}
	var doc glbDocument
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		return nil, glbDocument{}, nil, err
	}
	return jsonBytes, doc, data[pos : pos+binLen], nil
}

func glbTriangleBVH(doc *glbDocument, bin []byte) (*graviton.TriangleBVH, error) {
	if doc.Extras == nil || doc.Extras.Kaiju.Blobs == nil ||
		doc.Extras.Kaiju.Blobs.TriangleBVH == nil ||
		doc.Extras.Kaiju.Blobs.TriangleBVH.BufferView == nil {
		return nil, nil
	}
	ref := doc.Extras.Kaiju.Blobs.TriangleBVH
	if ref.Format != "kaiju.triangle_bvh.le.v1" {
		return nil, fmt.Errorf("unsupported triangle BVH blob format %q", ref.Format)
	}
	idx := *ref.BufferView
	if idx < 0 || idx >= len(doc.BufferViews) {
		return nil, fmt.Errorf("invalid triangle BVH bufferView %d", idx)
	}
	view := doc.BufferViews[idx]
	end := view.ByteOffset + view.ByteLength
	if view.ByteOffset < 0 || view.ByteLength < 0 || end > len(bin) {
		return nil, errors.New("triangle BVH bufferView exceeds BIN chunk")
	}
	return deserializeTriangleBVHBlob(bin[view.ByteOffset:end])
}

func appendPadded(data []byte, pad byte) []byte {
	out := make([]byte, len(data), len(data)+3)
	copy(out, data)
	for len(out)%4 != 0 {
		out = append(out, pad)
	}
	return out
}

func appendF32(out []byte, v float32) []byte {
	return binary.LittleEndian.AppendUint32(out, math.Float32bits(v))
}

func appendU16(out []byte, v uint16) []byte {
	return binary.LittleEndian.AppendUint16(out, v)
}

func indexBytes(indexes []uint32) []byte {
	out := make([]byte, 0, len(indexes)*4)
	for _, idx := range indexes {
		out = binary.LittleEndian.AppendUint32(out, idx)
	}
	return out
}

func vertexVec2Bytes(verts []rendering.Vertex, value func(rendering.Vertex) matrix.Vec2) []byte {
	out := make([]byte, 0, len(verts)*2*4)
	for _, vert := range verts {
		v := value(vert)
		out = appendF32(out, float32(v.X()))
		out = appendF32(out, float32(v.Y()))
	}
	return out
}

func vertexVec3Bytes(verts []rendering.Vertex, value func(rendering.Vertex) matrix.Vec3) []byte {
	out := make([]byte, 0, len(verts)*3*4)
	for _, vert := range verts {
		v := value(vert)
		out = appendF32(out, float32(v.X()))
		out = appendF32(out, float32(v.Y()))
		out = appendF32(out, float32(v.Z()))
	}
	return out
}

func vertexVec4Bytes(verts []rendering.Vertex, value func(rendering.Vertex) matrix.Vec4) []byte {
	out := make([]byte, 0, len(verts)*4*4)
	for _, vert := range verts {
		v := value(vert)
		out = appendF32(out, float32(v.X()))
		out = appendF32(out, float32(v.Y()))
		out = appendF32(out, float32(v.Z()))
		out = appendF32(out, float32(v.W()))
	}
	return out
}

func vertexColorBytes(verts []rendering.Vertex) []byte {
	out := make([]byte, 0, len(verts)*4*4)
	for _, vert := range verts {
		out = appendF32(out, float32(vert.Color.R()))
		out = appendF32(out, float32(vert.Color.G()))
		out = appendF32(out, float32(vert.Color.B()))
		out = appendF32(out, float32(vert.Color.A()))
	}
	return out
}

func vertexJointBytes(verts []rendering.Vertex) ([]byte, error) {
	out := make([]byte, 0, len(verts)*4*2)
	for _, vert := range verts {
		for _, id := range vert.JointIds {
			if id < 0 || id > math.MaxUint16 {
				return nil, fmt.Errorf("joint id %d exceeds glTF UNSIGNED_SHORT", id)
			}
			out = appendU16(out, uint16(id))
		}
	}
	return out, nil
}

func vertexVec3MinMax(verts []rendering.Vertex, value func(rendering.Vertex) matrix.Vec3) (matrix.Vec3, matrix.Vec3) {
	minV := value(verts[0])
	maxV := minV
	for i := 1; i < len(verts); i++ {
		v := value(verts[i])
		minV = matrix.Vec3Min(minV, v)
		maxV = matrix.Vec3Max(maxV, v)
	}
	return minV, maxV
}

func shouldWriteSkinAttributes(k KaijuMesh) bool {
	if len(k.Joints) > 0 {
		return true
	}
	for i := range k.Verts {
		if !matrix.Vec4Approx(k.Verts[i].JointWeights, matrix.Vec4Zero()) {
			return true
		}
		for _, id := range k.Verts[i].JointIds {
			if id != 0 {
				return true
			}
		}
	}
	return false
}

func shouldWriteMorphTarget(verts []rendering.Vertex) bool {
	for i := range verts {
		if !matrix.Vec3Approx(verts[i].MorphTarget, matrix.Vec3Zero()) &&
			!matrix.Vec3Approx(verts[i].MorphTarget, verts[i].Position) {
			return true
		}
	}
	return false
}

func vec3JSON(v matrix.Vec3) []float32 {
	return []float32{float32(v.X()), float32(v.Y()), float32(v.Z())}
}

func quatXYZWJSON(q matrix.Quaternion) []float32 {
	return []float32{float32(q.X()), float32(q.Y()), float32(q.Z()), float32(q.W())}
}

func appendUniqueInt(values []int, value int) []int {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func glbSceneNodes(nodes []glbNode, meshNode int) []int {
	hasParent := make([]bool, len(nodes))
	for i := range nodes {
		for _, child := range nodes[i].Children {
			if child >= 0 && child < len(hasParent) {
				hasParent[child] = true
			}
		}
	}
	out := make([]int, 0, len(nodes))
	for i := range nodes {
		if i != meshNode && !hasParent[i] &&
			(nodes[i].Name != "" || len(nodes[i].Children) > 0) {
			out = appendUniqueInt(out, i)
		}
	}
	return appendUniqueInt(out, meshNode)
}

func ptrInt(v int) *int { return &v }

func animationAbsoluteTimes(anim *KaijuMeshAnimation) []float32 {
	out := make([]float32, len(anim.Frames))
	var total float32
	for i := range anim.Frames {
		out[i] = total
		total += anim.Frames[i].Time
	}
	return out
}

func findAnimationBone(frame *AnimKeyFrame, node int, path AnimationPathType, interpolation AnimationInterpolation) *AnimBone {
	for i := range frame.Bones {
		bone := &frame.Bones[i]
		if bone.NodeIndex == node && bone.PathType == path && bone.Interpolation == interpolation {
			return bone
		}
	}
	return nil
}

func animationPathString(path AnimationPathType) (string, string) {
	switch path {
	case AnimPathTranslation:
		return "translation", glbTypeVec3
	case AnimPathRotation:
		return "rotation", glbTypeVec4
	case AnimPathScale:
		return "scale", glbTypeVec3
	default:
		return "", ""
	}
}

func animationInterpolationString(interpolation AnimationInterpolation) string {
	switch interpolation {
	case AnimInterpolateStep:
		return "STEP"
	default:
		return "LINEAR"
	}
}

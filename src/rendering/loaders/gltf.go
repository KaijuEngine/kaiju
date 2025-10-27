/******************************************************************************/
/* gltf.go                                                                    */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package loaders

import (
	"encoding/binary"
	"errors"
	"fmt"
	"kaiju/engine/assets"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/loaders/gltf"
	"kaiju/rendering/loaders/load_result"
	"path/filepath"
	"slices"
	"strings"
	"unsafe"
)

type fullGLTF struct {
	path string
	glTF gltf.GLTF
	bins [][]byte
}

type rawMeshData struct {
	verts   []rendering.Vertex
	indices []uint32
}

func readFileGLB(file string, assetDB assets.Database) (fullGLTF, error) {
	defer tracing.NewRegion("loaders.readFileGLB").End()
	const headSize = 12
	const chunkHeadSize = 8
	g := fullGLTF{path: file}
	data, err := assetDB.Read(file)
	if err != nil {
		return g, err
	}
	if len(data) < headSize {
		return g, errors.New("invalid glb file")
	}
	magic := data[:4]
	//version := binary.LittleEndian.Uint32(data[4:8])
	//length := binary.LittleEndian.Uint32(data[8:12])
	if string(magic) != "glTF" {
		return g, errors.New("invalid glb file")
	}
	jsonData := data[headSize:]
	if len(jsonData) < 8 {
		return g, errors.New("invalid glb file")
	}
	chunkLen := binary.LittleEndian.Uint32(jsonData[:4])
	chunkType := jsonData[4:8]
	if string(chunkType) != "JSON" {
		return g, errors.New("invalid glb file")
	}
	jsonData = jsonData[:chunkHeadSize+chunkLen]
	jsonStr := string(jsonData[chunkHeadSize:])
	g.glTF, err = gltf.LoadGLTF(jsonStr)
	if err != nil {
		return g, err
	}
	g.glTF.Asset.FilePath = file
	bins := data[headSize+len(jsonData):]
	if len(bins) < chunkHeadSize {
		return g, errors.New("invalid glb file")
	}
	chunkLen = binary.LittleEndian.Uint32(bins[:4])
	chunkType = bins[4:8]
	if string(chunkType) != "BIN\000" {
		return g, errors.New("invalid glb file")
	}
	bins = bins[chunkHeadSize:]
	g.bins = make([][]byte, len(g.glTF.Buffers))
	for i, buffer := range g.glTF.Buffers {
		if buffer.ByteLength == 0 {
			continue
		}
		if len(bins) < int(buffer.ByteLength) {
			return g, errors.New("invalid glb file")
		}
		g.bins[i] = bins[:buffer.ByteLength]
		bins = bins[buffer.ByteLength:]
	}
	return g, nil
}

func readFileGLTF(file string, assetDB assets.Database) (fullGLTF, error) {
	defer tracing.NewRegion("loaders.readFileGLTF").End()
	g := fullGLTF{path: file}
	str, err := assetDB.ReadText(file)
	if err != nil {
		return g, err
	}
	g.glTF, err = gltf.LoadGLTF(str)
	if err != nil {
		return g, err
	}
	g.glTF.Asset.FilePath = file
	g.bins = make([][]byte, len(g.glTF.Buffers))
	root := filepath.Dir(file)
	for i, path := range g.glTF.Buffers {
		uri := filepath.Join(root, path.URI)
		if !assetDB.Exists(uri) {
			return g, errors.New("bin file (" + uri + ") does not exist")
		}
		g.bins[i], err = assetDB.Read(uri)
		if err != nil {
			return g, err
		}
	}
	return g, nil
}

func GLTF(path string, assetDB assets.Database) (load_result.Result, error) {
	defer tracing.NewRegion("loaders.GLTF").End()
	if !assetDB.Exists(path) {
		return load_result.Result{}, errors.New("file does not exist")
	} else if filepath.Ext(path) == ".glb" {
		if g, err := readFileGLB(path, assetDB); err != nil {
			return load_result.Result{}, err
		} else {
			return gltfParse(&g)
		}
	} else if filepath.Ext(path) == ".gltf" {
		if g, err := readFileGLTF(path, assetDB); err != nil {
			return load_result.Result{}, err
		} else {
			return gltfParse(&g)
		}
	} else {
		return load_result.Result{}, errors.New("invalid file extension")
	}
}

func gltfParse(doc *fullGLTF) (load_result.Result, error) {
	defer tracing.NewRegion("loaders.gltfParse").End()
	res := load_result.Result{}
	res.Nodes = make([]load_result.Node, len(doc.glTF.Nodes))
	for i := range res.Nodes {
		res.Nodes[i].Parent = -1
		res.Nodes[i].Attributes = make(map[string]any)
	}
	// TODO:  Deal with multiple skins
	if len(doc.glTF.Skins) > 0 {
		skinAcc := doc.glTF.Accessors[doc.glTF.Skins[0].InverseBindMatrices]
		bv := doc.glTF.BufferViews[skinAcc.BufferView]
		bin := klib.ByteSliceToFloat32Slice(gltfViewBytes(doc, &bv))
		for _, id := range doc.glTF.Skins[0].Joints {
			if !strings.HasPrefix(doc.glTF.Nodes[id].Name, "DRV_") &&
				!strings.HasPrefix(doc.glTF.Nodes[id].Name, "CTRL_") {
				res.Joints = append(res.Joints, load_result.Joint{
					Id:   id,
					Skin: matrix.Mat4FromSlice(bin),
				})
			}
			bin = bin[16:]
		}
	}
	meshDatas := map[int32]rawMeshData{}
	for i := range doc.glTF.Nodes {
		n := &doc.glTF.Nodes[i]
		res.Nodes[i].Name = n.Name
		res.Nodes[i].Transform.SetupRawTransform()
		res.Nodes[i].Transform.Identifier = uint8(i)
		res.Nodes[i].Attributes = n.Extras
		for j := range n.Children {
			res.Nodes[n.Children[j]].Parent = i
			res.Nodes[n.Children[j]].Transform.SetParent(&res.Nodes[i].Transform)
		}
		// TODO:  Come back for this scenario
		//if n.Matrix != nil {
		//}
		if n.Scale != nil {
			res.Nodes[i].Transform.SetScale(*n.Scale)
		}
		if n.Rotation != nil {
			q := matrix.QuaternionFromXYZW(*n.Rotation)
			res.Nodes[i].Transform.SetRotation(q.ToEuler())
		}
		if n.Translation != nil {
			res.Nodes[i].Transform.SetPosition(*n.Translation)
		}
		if n.Mesh == nil {
			continue
		}
		m := &doc.glTF.Meshes[*n.Mesh]
		rmd, ok := meshDatas[*n.Mesh]
		if !ok {
			if verts, err := gltfReadMeshVerts(m, doc); err != nil {
				return res, err
			} else if indices, err := gltfReadMeshIndices(m, doc); err != nil {
				return res, err
			} else {
				rmd.verts = verts
				rmd.indices = indices
			}
			meshDatas[*n.Mesh] = rmd
		}
		textures := gltfReadMeshTextures(m, &doc.glTF)
		key := fmt.Sprintf("%s/%s", doc.path, m.Name)
		res.Add(n.Name, key, rmd.verts, rmd.indices, textures, &res.Nodes[i])
	}
	res.Animations = gltfReadAnimations(doc)
	return res, nil
}

func gltfAttr(primitive []gltf.Primitive, cmp string) (uint32, bool) {
	defer tracing.NewRegion("loaders.gltfAttr").End()
	idx, ok := primitive[0].Attributes[cmp]
	return idx, ok
}

func gltfViewBytes(doc *fullGLTF, view *gltf.BufferView) []byte {
	defer tracing.NewRegion("loaders.gltfViewBytes").End()
	return doc.bins[view.Buffer][view.ByteOffset : view.ByteOffset+view.ByteLength]
}

func gltfReadMeshMorphTargets(mesh *gltf.Mesh, doc *fullGLTF, verts []rendering.Vertex) klib.ErrorList {
	defer tracing.NewRegion("loaders.gltfReadMeshMorphTargets").End()
	errs := klib.NewErrorList()
	for _, target := range mesh.Primitives[0].Targets {
		if target.POSITION == nil {
			continue
		}
		acc := doc.glTF.Accessors[*target.POSITION]
		if len(doc.glTF.BufferViews) <= int(acc.BufferView) {
			errs.AddAny(errors.New("invalid buffer view index"))
		}
		view := doc.glTF.BufferViews[acc.BufferView]
		if acc.Count <= 0 {
			errs.AddAny(errors.New("invalid accessor count"))
			continue
		}
		targets := gltfViewBytes(doc, &view)
		const v3Size = int(unsafe.Sizeof([3]float32{}))
		if int(acc.Count) != len(verts) || len(targets)/v3Size != len(verts) {
			errs.AddAny(errors.New("morph targets do not match vert count"))
			continue
		}
		floats := klib.ConvertByteSliceType[float32](targets)
		for i := 0; i < len(verts); i++ {
			verts[i].MorphTarget = matrix.Vec3{
				floats[i*3+0],
				floats[i*3+1],
				floats[i*3+2],
			}
		}
	}
	return errs
}

func gltfReadMeshVerts(mesh *gltf.Mesh, doc *fullGLTF) ([]rendering.Vertex, error) {
	defer tracing.NewRegion("loaders.gltfReadMeshVerts").End()
	var pos, nml, tan, tex0, tex1, jnt0, wei0 *gltf.BufferView
	var posAcc, nmlAcc, tanAcc, tex0Acc, tex1Acc, jnt0Acc, wei0Acc *gltf.Accessor
	g := &doc.glTF
	if idx, ok := gltfAttr(mesh.Primitives, gltf.POSITION); ok {
		pos = &g.BufferViews[idx]
		posAcc = &g.Accessors[idx]
	}
	if idx, ok := gltfAttr(mesh.Primitives, gltf.NORMAL); ok {
		nml = &g.BufferViews[idx]
		nmlAcc = &g.Accessors[idx]
	}
	if idx, ok := gltfAttr(mesh.Primitives, gltf.TANGENT); ok {
		tan = &g.BufferViews[idx]
		tanAcc = &g.Accessors[idx]
	}
	if idx, ok := gltfAttr(mesh.Primitives, gltf.TEXCOORD_0); ok {
		tex0 = &g.BufferViews[idx]
		tex0Acc = &g.Accessors[idx]
	}
	if idx, ok := gltfAttr(mesh.Primitives, gltf.TEXCOORD_1); ok {
		tex1 = &g.BufferViews[idx]
		tex1Acc = &g.Accessors[idx]
	}
	if idx, ok := gltfAttr(mesh.Primitives, gltf.JOINTS_0); ok {
		jnt0 = &g.BufferViews[idx]
		jnt0Acc = &g.Accessors[idx]
	}
	if idx, ok := gltfAttr(mesh.Primitives, gltf.WEIGHTS_0); ok {
		wei0 = &g.BufferViews[idx]
		wei0Acc = &g.Accessors[idx]
	}

	// TODO:  Probably need to support multiple buffers, but they are NULL?
	verts := klib.ByteSliceToFloat32Slice(gltfViewBytes(doc, pos))
	vertNormals := klib.ByteSliceToFloat32Slice(gltfViewBytes(doc, nml))
	var texCoords0 []float32
	var tangent []float32
	if tex0 != nil {
		texCoords0 = klib.ByteSliceToFloat32Slice(gltfViewBytes(doc, tex0))
	} else {
		texCoords0 = nil
	}
	if tan != nil {
		tangent = klib.ByteSliceToFloat32Slice(gltfViewBytes(doc, tan))
	} else {
		tangent = nil
	}
	//const uint8_t* vertColors = col0 != NULL
	//	? (uint8_t*)gltfData.bin + col0.data.buffer_view.offset : NULL;
	jointIds := make([]byte, 0)
	weights := make([]float32, 0)
	if jnt0 != nil {
		jointIds = gltfViewBytes(doc, jnt0)
		weights = klib.ByteSliceToFloat32Slice(gltfViewBytes(doc, wei0))
	}

	//size_t vertNormalsSize = nml.data.buffer_view.size;
	//size_t texCoords0Size = tex0.data.buffer_view.size;
	//size_t texCoords1Size = tex1 == NULL ? 0 : tex1.data.buffer_view.size;
	//size_t vertTangentSize = tan == NULL ? 0 : tan.data.buffer_view.size;
	//size_t vertColorsSize = col0 == NULL ? 0 : col0.data.buffer_view.size;
	vertCount := posAcc.Count
	if !(vertCount > 0) {
		return []rendering.Vertex{}, errors.New("vertCount <= 0")
	}
	if !(posAcc.ComponentType == gltf.FLOAT && posAcc.Type == gltf.VEC3) {
		return []rendering.Vertex{}, errors.New("posAcc.ComponentType != gltf.ComponentFloat || posAcc.Type != gltf.AccessorVec3")
	}
	if !(wei0 == nil || wei0Acc.ComponentType == gltf.FLOAT && wei0Acc.Type == gltf.VEC4) {
		return []rendering.Vertex{}, errors.New("wei0 == NULL || wei0Acc.ComponentType == gltf.ComponentFloat && wei0Acc.Type == gltf.AccessorVec4")
	}
	if !(nmlAcc.ComponentType == gltf.FLOAT && nmlAcc.Type == gltf.VEC3) {
		return []rendering.Vertex{}, errors.New("nmlAcc.ComponentType != gltf.ComponentFloat || nmlAcc.Type != gltf.AccessorVec3")
	}
	if !(tan == nil || tanAcc.ComponentType == gltf.FLOAT && tanAcc.Type == gltf.VEC4) {
		return []rendering.Vertex{}, errors.New("tan == NULL || tanAcc.ComponentType == gltf.ComponentFloat && tanAcc.Type == gltf.AccessorVec4")
	}
	if !(tex0 == nil || tex0Acc.ComponentType == gltf.FLOAT && tex0Acc.Type == gltf.VEC2) {
		return []rendering.Vertex{}, errors.New("tex0 == NULL || tex0Acc.ComponentType == gltf.ComponentFloat && tex0Acc.Type == gltf.AccessorVec2")
	}
	if !(tex1 == nil || tex1Acc.ComponentType == gltf.FLOAT && tex1Acc.Type == gltf.VEC2) {
		return []rendering.Vertex{}, errors.New("tex1 == NULL || tex1Acc.ComponentType == gltf.ComponentFloat && tex1Acc.Type == gltf.AccessorVec2")
	}
	vertData := make([]rendering.Vertex, vertCount)
	vertColor := matrix.ColorWhite()
	if mesh.Primitives[0].Material != nil && len(doc.glTF.Materials) > int(*mesh.Primitives[0].Material) {
		mat := doc.glTF.Materials[*mesh.Primitives[0].Material]
		if mat.PBRMetallicRoughness.BaseColorFactor != nil {
			vertColor = *mat.PBRMetallicRoughness.BaseColorFactor
		}
	}
	for i := int32(0); i < vertCount; i++ {
		vertData[i].Position = matrix.Vec3FromSlice(verts)
		verts = verts[3:]
		vertData[i].Color = vertColor
		vertData[i].MorphTarget = vertData[i].Position
		// NAN is being exported for colors, so skipping this line
		//vertData[j].color = (vertColors != NULL ? ((color*)vertColors)[j] : color_white());
		vertData[i].Color.MultiplyAssign(vertColor)
		joint := [4]int32{0, 0, 0, 0}
		const jointSize = uint64(unsafe.Sizeof(joint))
		if len(jointIds) > 0 {
			switch jnt0Acc.ComponentType {
			case gltf.UNSIGNED_BYTE:
				joint[0] = int32(jointIds[0])
				joint[1] = int32(jointIds[1])
				joint[2] = int32(jointIds[2])
				joint[3] = int32(jointIds[3])
				jointIds = jointIds[4:]
			case gltf.UNSIGNED_SHORT:
				ptr := klib.ByteSliceToUInt16Slice(jointIds)
				joint[0] = int32(ptr[0])
				joint[1] = int32(ptr[1])
				joint[2] = int32(ptr[2])
				joint[3] = int32(ptr[3])
				jointIds = jointIds[4*2:]
			default:
				klib.Memcpy(unsafe.Pointer(&joint[0]), unsafe.Pointer(&jointIds[0]), jointSize)
				jointIds = jointIds[jointSize:]
			}
		}
		vertData[i].JointIds = matrix.Vec4i{joint[0], joint[1], joint[2], joint[3]}
		if len(weights) > 0 {
			vertData[i].JointWeights = matrix.Vec4FromSlice(weights)
			weights = weights[4:]
		} else {
			vertData[i].JointWeights = matrix.Vec4Zero()
		}
		vertData[i].Normal = matrix.Vec3FromSlice(vertNormals)
		vertNormals = vertNormals[3:]
		if tangent != nil {
			vertData[i].Tangent = matrix.Vec4FromSlice(tangent)
			tangent = tangent[4:]
		} else {
			vertData[i].Tangent = matrix.Vec4Zero()
		}
		if texCoords0 != nil {
			vertData[i].UV0 = matrix.Vec2FromSlice(texCoords0)
			texCoords0 = texCoords0[2:]
		} else {
			vertData[i].UV0 = matrix.Vec2Zero()
		}
		for vertData[i].UV0.X() > 1.0 {
			vertData[i].UV0[matrix.Vx] -= 1.0
		}
		for vertData[i].UV0.Y() > 1.0 {
			vertData[i].UV0[matrix.Vy] -= 1.0
		}
	}
	errs := gltfReadMeshMorphTargets(mesh, doc, vertData)
	return vertData, errs.First()
}

func gltfReadMeshIndices(mesh *gltf.Mesh, doc *fullGLTF) ([]uint32, error) {
	defer tracing.NewRegion("loaders.gltfReadMeshIndices").End()
	idx := mesh.Primitives[0].Indices
	view := doc.glTF.BufferViews[idx]
	acc := doc.glTF.Accessors[idx]
	indices := doc.bins[view.Buffer][view.ByteOffset:]
	indicesSize := view.ByteLength
	if !(indicesSize > 0) {
		return []uint32{}, errors.New("indicesCount > 0")
	}
	var convertedIndices []uint32
	switch acc.ComponentType {
	case gltf.BYTE:
		fallthrough
	case gltf.UNSIGNED_BYTE:
		indicesCount := indicesSize
		convertedIndices = make([]uint32, indicesSize)
		for i := int32(0); i < indicesCount; i++ {
			convertedIndices[i] = uint32(indices[i])
		}
	case gltf.SHORT:
		fallthrough
	case gltf.UNSIGNED_SHORT:
		indicesCount := indicesSize / 2
		convertedIndices = make([]uint32, indicesCount)
		vals := unsafe.Slice((*uint16)(unsafe.Pointer(&indices[0])), indicesCount)
		for i := int32(0); i < indicesCount; i++ {
			convertedIndices[i] = uint32(vals[i])
		}
	case gltf.UNSIGNED_INT:
		fallthrough
	case gltf.FLOAT:
		indicesCount := indicesSize / 4
		convertedIndices = make([]uint32, indicesCount)
		vals := unsafe.Slice((*uint32)(unsafe.Pointer(&indices[0])), indicesCount)
		for i := int32(0); i < indicesCount; i++ {
			convertedIndices[i] = uint32(vals[i])
		}
	default:
		return []uint32{}, errors.New("invalid component type")
	}
	return convertedIndices, nil
}

func gltfReadMeshTextures(mesh *gltf.Mesh, doc *gltf.GLTF) map[string]string {
	defer tracing.NewRegion("loaders.gltfReadMeshTextures").End()
	textures := make(map[string]string)
	if len(doc.Materials) == 0 || mesh.Primitives[0].Material == nil {
		return textures
	}
	uri := func(path string) string {
		return filepath.ToSlash(filepath.Join(filepath.Dir(doc.Asset.FilePath), path))
	}
	mat := doc.Materials[*mesh.Primitives[0].Material]
	if mat.PBRMetallicRoughness.BaseColorTexture != nil {
		textures["baseColor"] = uri(doc.Images[mat.PBRMetallicRoughness.BaseColorTexture.Index].URI)
	}
	if mat.PBRMetallicRoughness.MetallicRoughnessTexture != nil {
		textures["metallicRoughness"] = uri(doc.Images[mat.PBRMetallicRoughness.MetallicRoughnessTexture.Index].URI)
	}
	if mat.NormalTexture != nil {
		textures["normal"] = uri(doc.Images[mat.NormalTexture.Index].URI)
	}
	if mat.OcclusionTexture != nil {
		textures["occlusion"] = uri(doc.Images[mat.OcclusionTexture.Index].URI)
	}
	if mat.EmissiveTexture != nil {
		textures["emissive"] = uri(doc.Images[mat.EmissiveTexture.Index].URI)
	}
	return textures
}

func gltfReadAnimations(doc *fullGLTF) []load_result.Animation {
	defer tracing.NewRegion("loaders.gltfReadAnimations").End()
	anims := make([]load_result.Animation, len(doc.glTF.Animations))
	for i := range doc.glTF.Animations {
		a := &doc.glTF.Animations[i]
		anims[i] = load_result.Animation{
			Name:   a.Name,
			Frames: make([]load_result.AnimKeyFrame, 0),
		}
		for j := range doc.glTF.Animations[i].Channels {
			c := a.Channels[j]
			sampler := &a.Samplers[c.Sampler]
			inAcc := &doc.glTF.Accessors[sampler.Input]
			outAcc := &doc.glTF.Accessors[sampler.Output]
			// Times ([]float32) of the key frames of the animation
			in := gltfViewBytes(doc, &doc.glTF.BufferViews[inAcc.BufferView])
			// Values for the animated properties at the respective key frames
			out := gltfViewBytes(doc, &doc.glTF.BufferViews[outAcc.BufferView])
			fIn := klib.ByteSliceToFloat32Slice(in)
			fOut := klib.ByteSliceToFloat32Slice(out)
			for k := 0; k < len(fIn); k++ {
				var key *load_result.AnimKeyFrame = nil
				for l := range anims[i].Frames {
					if matrix.Approx(anims[i].Frames[l].Time, fIn[k]) {
						key = &anims[i].Frames[l]
						break
					}
				}
				if key == nil {
					anims[i].Frames = append(anims[i].Frames, load_result.AnimKeyFrame{
						Bones: make([]load_result.AnimBone, 0),
						Time:  fIn[k],
					})
					key = &anims[i].Frames[len(anims[i].Frames)-1]
				}
				bone := load_result.AnimBone{
					PathType:      c.Target.Path(),
					Interpolation: sampler.Interpolation(),
					NodeIndex:     int(c.Target.Node),
				}
				switch bone.PathType {
				case load_result.AnimPathTranslation:
					bone.Data = matrix.Vec3FromSlice(fOut).AsAligned16()
					fOut = fOut[3:]
				case load_result.AnimPathRotation:
					// glTF has the specification as XYZW instead of WXYZ
					bone.Data = matrix.QuaternionFromXYZWSlice(fOut)
					fOut = fOut[4:]
				case load_result.AnimPathScale:
					bone.Data = matrix.Vec3FromSlice(fOut).AsAligned16()
					fOut = fOut[3:]
				case load_result.AnimPathWeights:
					// TODO:  Implement reading weights data
				}
				key.Bones = append(key.Bones, bone)
			}
		}
		slices.SortFunc(anims[i].Frames, func(a, b load_result.AnimKeyFrame) int {
			return int((a.Time - b.Time) * 10000)
		})
		// Convert frame from absolute time to relative time length
		for j := range anims[i].Frames[:len(anims[i].Frames)-1] {
			anims[i].Frames[j].Time = anims[i].Frames[j+1].Time - anims[i].Frames[j].Time
		}
		// Last frame should be a goal and not have time?
		anims[i].Frames[len(anims[i].Frames)-1].Time = 0.0
	}
	return anims
}

/******************************************************************************/
/* gltf.go                                                                    */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package loaders

import (
	"encoding/binary"
	"errors"
	"kaiju/assets"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders/gltf"
	"kaiju/rendering/loaders/load_result"
	"path/filepath"
	"slices"
	"strings"
	"unsafe"
)

type fullGLTF struct {
	glTF gltf.GLTF
	bins [][]byte
}

func readFileGLB(file string, assetDB *assets.Database) (fullGLTF, error) {
	const headSize = 12
	const chunkHeadSize = 8
	g := fullGLTF{}
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

func readFileGLTF(file string, assetDB *assets.Database) (fullGLTF, error) {
	g := fullGLTF{}
	str, err := assetDB.ReadText(file)
	if err != nil {
		return g, err
	}
	g.glTF, err = gltf.LoadGLTF(str)
	if err != nil {
		return g, err
	}
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

func GLTF(path string, assetDB *assets.Database) (load_result.Result, error) {
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
	res := load_result.NewResult()
	res.Nodes = make([]load_result.Node, len(doc.glTF.Nodes))
	for i := range res.Nodes {
		res.Nodes[i].Parent = -1
	}
	// TODO:  Deal with multiple skins
	if len(doc.glTF.Skins) > 0 {
		bv := doc.glTF.BufferViews[doc.glTF.Skins[0].InverseBindMatrices]
		bin := klib.ByteSliceToFloat32Slice(doc.bins[bv.Buffer][bv.ByteOffset : bv.ByteOffset+bv.ByteLength])
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
	for i := range doc.glTF.Nodes {
		n := &doc.glTF.Nodes[i]
		res.Nodes[i].Name = n.Name
		res.Nodes[i].Transform = matrix.NewTransform()
		res.Nodes[i].Transform.Identifier = uint8(i)
		for j := range n.Children {
			res.Nodes[n.Children[j]].Parent = i
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
			res.Nodes[i].Transform.SetRotation(*n.Translation)
		}
		if n.Mesh == nil {
			continue
		}
		m := &doc.glTF.Meshes[*n.Mesh]
		if verts, err := gltfReadMeshVerts(m, doc); err != nil {
			return res, err
		} else if indices, err := gltfReadMeshIndices(m, doc); err != nil {
			return res, err
		} else {
			textures := gltfReadMeshTextures(m, &doc.glTF)
			res.Add(n.Name, m.Name, verts, indices, klib.MapValues(textures))
		}
		res.Animations = gltfReadAnimations(doc)
	}
	return res, nil
}

func gltfAttr(primitive []gltf.Primitive, cmp string) (uint32, bool) {
	idx, ok := primitive[0].Attributes[cmp]
	return idx, ok
}

func gltfViewBytes(doc *fullGLTF, view *gltf.BufferView) []byte {
	return doc.bins[view.Buffer][view.ByteOffset : view.ByteOffset+view.ByteLength]
}

func gltfReadMeshMorphTargets(mesh *gltf.Mesh, doc *fullGLTF, verts []rendering.Vertex) klib.ErrorList {
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
	verts := gltfViewBytes(doc, pos)
	vertNormals := gltfViewBytes(doc, nml)
	var texCoords0 []byte
	var tangent []byte
	if tex0 != nil {
		texCoords0 = gltfViewBytes(doc, tex0)
	} else {
		texCoords0 = nil
	}
	if tan != nil {
		tangent = gltfViewBytes(doc, tan)
	} else {
		tangent = nil
	}
	//const uint8_t* vertColors = col0 != NULL
	//	? (uint8_t*)gltfData.bin + col0.data.buffer_view.offset : NULL;
	jointIds := make([]byte, 0)
	weights := make([]byte, 0)
	if jnt0 != nil {
		jointIds = gltfViewBytes(doc, jnt0)
		weights = gltfViewBytes(doc, wei0)
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
	const v4size = int32(unsafe.Sizeof(matrix.Vec4{}))
	const v3size = int32(unsafe.Sizeof(matrix.Vec3{}))
	const v2size = int32(unsafe.Sizeof(matrix.Vec2{}))
	for i := int32(0); i < vertCount; i++ {
		klib.Memcpy(unsafe.Pointer(&vertData[i].Position), unsafe.Pointer(&verts[i*v3size]), uint64(unsafe.Sizeof(vertData[i].Position)))
		vertData[i].Color = vertColor
		vertData[i].MorphTarget = vertData[i].Position
		// NAN is being exported for colors, so skipping this line
		//vertData[j].color = (vertColors != NULL ? ((color*)vertColors)[j] : color_white());
		vertData[i].Color.MultiplyAssign(vertColor)
		joint := [4]int{0, 0, 0, 0}
		const jointSize = uint64(unsafe.Sizeof(joint))
		if len(jointIds) > 0 {
			switch jnt0Acc.ComponentType {
			case gltf.UNSIGNED_BYTE:
				ptr := jointIds[i*4:]
				joint[0] = int(ptr[0])
				joint[1] = int(ptr[1])
				joint[2] = int(ptr[2])
				joint[3] = int(ptr[3])
			case gltf.UNSIGNED_SHORT:
				ptr := jointIds[i*4*2:]
				joint[0] = int(ptr[0])
				joint[1] = int(ptr[1])
				joint[2] = int(ptr[2])
				joint[3] = int(ptr[3])
			default:
				klib.Memcpy(unsafe.Pointer(&joint[0]), unsafe.Pointer(&jointIds[i]), jointSize)
			}
		}
		klib.Memcpy(unsafe.Pointer(&vertData[i].JointIds), unsafe.Pointer(&joint[0]), jointSize)
		if len(weights) > 0 {
			klib.Memcpy(unsafe.Pointer(&vertData[i].JointWeights), unsafe.Pointer(&weights[i*v4size]), uint64(v4size))
		} else {
			vertData[i].JointWeights = matrix.Vec4Zero()
		}
		klib.Memcpy(unsafe.Pointer(&vertData[i].Normal), unsafe.Pointer(&vertNormals[i*v3size]), uint64(v3size))
		if tangent != nil {
			klib.Memcpy(unsafe.Pointer(&vertData[i].Tangent), unsafe.Pointer(&tangent[i*v4size]), uint64(v4size))
		} else {
			vertData[i].Tangent = matrix.Vec4Zero()
		}
		if texCoords0 != nil {
			klib.Memcpy(unsafe.Pointer(&vertData[i].UV0), unsafe.Pointer(&texCoords0[i*v2size]), uint64(v2size))
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
	textures := make(map[string]string)
	if len(doc.Materials) == 0 || mesh.Primitives[0].Material == nil {
		return textures
	}
	mat := doc.Materials[*mesh.Primitives[0].Material]
	if mat.PBRMetallicRoughness.BaseColorTexture != nil {
		textures["baseColor"] = doc.Images[mat.PBRMetallicRoughness.BaseColorTexture.Index].URI
	}
	if mat.PBRMetallicRoughness.MetallicRoughnessTexture != nil {
		textures["metallicRoughness"] = doc.Images[mat.PBRMetallicRoughness.MetallicRoughnessTexture.Index].URI
	}
	if mat.NormalTexture != nil {
		textures["normal"] = doc.Images[mat.NormalTexture.Index].URI
	}
	if mat.OcclusionTexture != nil {
		textures["occlusion"] = doc.Images[mat.OcclusionTexture.Index].URI
	}
	if mat.EmissiveTexture != nil {
		textures["emissive"] = doc.Images[mat.EmissiveTexture.Index].URI
	}
	return textures
}

func gltfReadAnimations(doc *fullGLTF) []load_result.Animation {
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
			iv := &doc.glTF.BufferViews[sampler.Input]
			ov := &doc.glTF.BufferViews[sampler.Output]
			// Times ([]float32) of the key frames of the animation
			in := doc.bins[iv.Buffer][iv.ByteOffset : iv.ByteOffset+iv.ByteLength]
			// Values for the animated properties at the respective key frames
			out := doc.bins[ov.Buffer][ov.ByteOffset : ov.ByteOffset+ov.ByteLength]
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

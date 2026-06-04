/******************************************************************************/
/* content_database_mesh_fast_gltf.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

const (
	fastGLBJSONChunkType = 0x4e4f534a
	fastGLBBINChunkType  = 0x004e4942
)

var fastGLBMagic = [4]byte{'g', 'l', 'T', 'F'}

type meshFastGLTFPostProcData struct {
	doc          map[string]any
	bin          []byte
	submeshes    []meshFastGLTFSubmesh
	imageSources map[int]meshFastGLTFImageSource
	textureBytes map[string][]byte
}

type meshFastGLTFSubmesh struct {
	Key           string
	Name          string
	Material      string
	NodeName      string
	NodeIndex     int
	Position      matrix.Vec3
	Rotation      matrix.Vec3
	Scale         matrix.Vec3
	Textures      map[string]string
	TextureImages map[string]int
	IsAnimated    bool
}

type meshFastGLTFImageSource struct {
	Key      string
	Path     string
	Embedded bool
}

func meshFastImportGLTF(src string) (ProcessedImport, error) {
	doc, buffers, err := meshFastReadGLTFDocument(src)
	if err != nil {
		return ProcessedImport{}, err
	}
	imageSources, imageBufferViews, textureBytes, err := meshFastGLTFImageSources(doc, buffers, src)
	if err != nil {
		return ProcessedImport{}, err
	}
	submeshes, err := meshFastGLTFSubmeshes(doc, imageSources)
	if err != nil {
		return ProcessedImport{}, err
	}
	if len(submeshes) == 0 {
		return ProcessedImport{}, NoMeshesInFileError{Path: src}
	}
	bin, err := meshFastGLTFImportBIN(src, doc, buffers, imageBufferViews, textureBytes)
	if err != nil {
		return ProcessedImport{}, err
	}
	meshFastGLTFApplyKaijuExtras(doc, submeshes, nil)
	data, err := meshFastEncodeGLB(doc, bin)
	if err != nil {
		return ProcessedImport{}, err
	}
	proc := ProcessedImport{
		Variants: []ImportVariant{{
			Name: fileNameNoExt(src),
			Data: data,
		}},
		postProcessData: meshFastGLTFPostProcData{
			doc:          doc,
			bin:          bin,
			submeshes:    submeshes,
			imageSources: imageSources,
			textureBytes: textureBytes,
		},
	}
	for _, source := range imageSources {
		if source.Embedded {
			continue
		}
		proc.Dependencies = klib.AppendUnique(proc.Dependencies, source.Path)
	}
	return proc, nil
}

func meshFastReadGLTFDocument(src string) (map[string]any, [][]byte, error) {
	switch strings.ToLower(filepath.Ext(src)) {
	case ".glb":
		return meshFastReadGLBDocument(src)
	case ".gltf":
		return meshFastReadJSONGLTFDocument(src)
	default:
		return nil, nil, errors.New("invalid glTF file extension")
	}
}

func meshFastReadJSONGLTFDocument(src string) (map[string]any, [][]byte, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, nil, err
	}
	doc, err := meshFastDecodeJSON(data)
	if err != nil {
		return nil, nil, err
	}
	buffers, err := meshFastReadGLTFBuffers(doc, src, nil)
	return doc, buffers, err
}

func meshFastReadGLBDocument(src string) (map[string]any, [][]byte, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, nil, err
	}
	if len(data) < 12 || !bytes.Equal(data[:4], fastGLBMagic[:]) {
		return nil, nil, errors.New("invalid glb file")
	}
	if version := binary.LittleEndian.Uint32(data[4:8]); version != 2 {
		return nil, nil, fmt.Errorf("unsupported glb version %d", version)
	}
	totalLen := int(binary.LittleEndian.Uint32(data[8:12]))
	if totalLen > len(data) {
		return nil, nil, errors.New("glb length exceeds data")
	}
	var jsonBytes []byte
	var binBytes []byte
	for pos := 12; pos+8 <= totalLen; {
		chunkLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
		chunkType := binary.LittleEndian.Uint32(data[pos+4 : pos+8])
		pos += 8
		if chunkLen < 0 || pos+chunkLen > totalLen {
			return nil, nil, errors.New("invalid glb chunk length")
		}
		chunk := data[pos : pos+chunkLen]
		pos += chunkLen
		switch chunkType {
		case fastGLBJSONChunkType:
			jsonBytes = bytes.TrimRight(chunk, " ")
		case fastGLBBINChunkType:
			binBytes = chunk
		}
	}
	if len(jsonBytes) == 0 {
		return nil, nil, errors.New("missing glb JSON chunk")
	}
	doc, err := meshFastDecodeJSON(jsonBytes)
	if err != nil {
		return nil, nil, err
	}
	buffers, err := meshFastReadGLTFBuffers(doc, src, binBytes)
	return doc, buffers, err
}

func meshFastDecodeJSON(data []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	doc := map[string]any{}
	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func meshFastReadGLTFBuffers(doc map[string]any, src string, binChunk []byte) ([][]byte, error) {
	bufferValues := meshFastArrayField(doc, "buffers")
	buffers := make([][]byte, len(bufferValues))
	binOffset := 0
	for i := range bufferValues {
		buffer, ok := meshFastMap(bufferValues[i])
		if !ok {
			return nil, fmt.Errorf("buffer %d is not an object", i)
		}
		byteLen, _ := meshFastIntField(buffer, "byteLength")
		if byteLen < 0 {
			return nil, fmt.Errorf("buffer %d has negative byteLength", i)
		}
		if uri, ok := meshFastStringField(buffer, "uri"); ok && uri != "" {
			data, err := meshFastReadURI(src, uri)
			if err != nil {
				return nil, err
			}
			if len(data) < byteLen {
				return nil, fmt.Errorf("buffer %d data is shorter than byteLength", i)
			}
			buffers[i] = data[:byteLen]
			continue
		}
		if byteLen == 0 {
			buffers[i] = nil
			continue
		}
		if binOffset+byteLen > len(binChunk) {
			return nil, fmt.Errorf("buffer %d exceeds glb BIN chunk", i)
		}
		buffers[i] = binChunk[binOffset : binOffset+byteLen]
		binOffset += byteLen
	}
	return buffers, nil
}

func meshFastReadURI(src, rawURI string) ([]byte, error) {
	if strings.HasPrefix(rawURI, "data:") {
		return meshFastDecodeDataURI(rawURI)
	}
	path := meshFastExternalURIPath(src, rawURI)
	return os.ReadFile(path)
}

func meshFastDecodeDataURI(rawURI string) ([]byte, error) {
	comma := strings.Index(rawURI, ",")
	if comma < 0 {
		return nil, errors.New("invalid data URI")
	}
	meta := rawURI[:comma]
	payload := rawURI[comma+1:]
	if strings.Contains(meta, ";base64") {
		return base64.StdEncoding.DecodeString(payload)
	}
	decoded, err := url.PathUnescape(payload)
	if err != nil {
		return nil, err
	}
	return []byte(decoded), nil
}

func meshFastExternalURIPath(src, rawURI string) string {
	decoded, err := url.PathUnescape(rawURI)
	if err == nil {
		rawURI = decoded
	}
	rawURI = filepath.FromSlash(rawURI)
	if filepath.IsAbs(rawURI) {
		return filepath.Clean(rawURI)
	}
	return filepath.Clean(filepath.Join(filepath.Dir(src), rawURI))
}

func meshFastGLTFImageSources(
	doc map[string]any,
	buffers [][]byte,
	src string,
) (map[int]meshFastGLTFImageSource, map[int]bool, map[string][]byte, error) {
	imageValues := meshFastArrayField(doc, "images")
	sources := make(map[int]meshFastGLTFImageSource, len(imageValues))
	imageBufferViews := make(map[int]bool)
	textureBytes := map[string][]byte{}
	for i := range imageValues {
		image, ok := meshFastMap(imageValues[i])
		if !ok {
			continue
		}
		key := fmt.Sprintf("embedded_%d", i)
		if viewIndex, ok := meshFastIntField(image, "bufferView"); ok {
			data, err := meshFastGLTFBufferViewBytes(doc, buffers, viewIndex)
			if err != nil {
				return nil, nil, nil, err
			}
			imageBufferViews[viewIndex] = true
			textureBytes[key] = slices.Clone(data)
			sources[i] = meshFastGLTFImageSource{Key: key, Embedded: true}
			delete(image, "bufferView")
			delete(image, "mimeType")
			continue
		}
		uri, ok := meshFastStringField(image, "uri")
		if !ok || uri == "" {
			continue
		}
		if strings.HasPrefix(uri, "data:") {
			data, err := meshFastDecodeDataURI(uri)
			if err != nil {
				return nil, nil, nil, err
			}
			textureBytes[key] = data
			sources[i] = meshFastGLTFImageSource{Key: key, Embedded: true}
			delete(image, "uri")
			delete(image, "mimeType")
			continue
		}
		path := meshFastExternalURIPath(src, uri)
		if _, err := os.Stat(path); err != nil {
			return nil, nil, nil, MeshInvalidTextureError{src, uri, path}
		}
		sources[i] = meshFastGLTFImageSource{Key: path, Path: path}
	}
	return sources, imageBufferViews, textureBytes, nil
}

func meshFastGLTFBufferViewBytes(doc map[string]any, buffers [][]byte, viewIndex int) ([]byte, error) {
	viewValues := meshFastArrayField(doc, "bufferViews")
	if viewIndex < 0 || viewIndex >= len(viewValues) {
		return nil, fmt.Errorf("invalid bufferView %d", viewIndex)
	}
	view, ok := meshFastMap(viewValues[viewIndex])
	if !ok {
		return nil, fmt.Errorf("bufferView %d is not an object", viewIndex)
	}
	bufferIndex, _ := meshFastIntField(view, "buffer")
	byteOffset, _ := meshFastIntField(view, "byteOffset")
	byteLength, _ := meshFastIntField(view, "byteLength")
	if bufferIndex < 0 || bufferIndex >= len(buffers) {
		return nil, fmt.Errorf("bufferView %d references invalid buffer %d", viewIndex, bufferIndex)
	}
	end := byteOffset + byteLength
	if byteOffset < 0 || byteLength < 0 || end > len(buffers[bufferIndex]) {
		return nil, fmt.Errorf("bufferView %d exceeds buffer bounds", viewIndex)
	}
	return buffers[bufferIndex][byteOffset:end], nil
}

func meshFastGLTFCompactBuffers(
	doc map[string]any,
	buffers [][]byte,
	skipViews map[int]bool,
) ([]byte, error) {
	viewValues := meshFastArrayField(doc, "bufferViews")
	type compactView struct {
		view   map[string]any
		data   []byte
		offset int
	}
	views := make([]compactView, len(viewValues))
	totalLen := 0
	for i := range viewValues {
		view, ok := meshFastMap(viewValues[i])
		if !ok {
			return nil, fmt.Errorf("bufferView %d is not an object", i)
		}
		if skipViews[i] {
			views[i] = compactView{view: view, offset: totalLen}
			continue
		}
		data, err := meshFastGLTFBufferViewBytes(doc, buffers, i)
		if err != nil {
			return nil, err
		}
		totalLen = meshFastAlign4(totalLen)
		views[i] = compactView{view: view, data: data, offset: totalLen}
		totalLen += len(data)
	}
	bin := make([]byte, totalLen)
	for i := range views {
		view := views[i].view
		view["buffer"] = 0
		view["byteOffset"] = views[i].offset
		view["byteLength"] = len(views[i].data)
		if len(views[i].data) > 0 {
			copy(bin[views[i].offset:], views[i].data)
		}
	}
	doc["buffers"] = []any{map[string]any{"byteLength": len(bin)}}
	return bin, nil
}

func meshFastGLTFImportBIN(
	src string,
	doc map[string]any,
	buffers [][]byte,
	imageBufferViews map[int]bool,
	textureBytes map[string][]byte,
) ([]byte, error) {
	if meshFastGLTFCanReuseGLBBIN(src, doc, buffers, imageBufferViews, textureBytes) {
		bufferValues := meshFastArrayField(doc, "buffers")
		buffer, _ := meshFastMap(bufferValues[0])
		buffer["byteLength"] = len(buffers[0])
		delete(buffer, "uri")
		return buffers[0], nil
	}
	return meshFastGLTFCompactBuffers(doc, buffers, imageBufferViews)
}

func meshFastGLTFCanReuseGLBBIN(
	src string,
	doc map[string]any,
	buffers [][]byte,
	imageBufferViews map[int]bool,
	textureBytes map[string][]byte,
) bool {
	if strings.ToLower(filepath.Ext(src)) != ".glb" || len(buffers) != 1 ||
		len(imageBufferViews) != 0 || len(textureBytes) != 0 {
		return false
	}
	bufferValues := meshFastArrayField(doc, "buffers")
	if len(bufferValues) != 1 {
		return false
	}
	buffer, ok := meshFastMap(bufferValues[0])
	if !ok {
		return false
	}
	if uri, ok := meshFastStringField(buffer, "uri"); ok && uri != "" {
		return false
	}
	for i, value := range meshFastArrayField(doc, "bufferViews") {
		view, ok := meshFastMap(value)
		if !ok {
			return false
		}
		bufferIndex, _ := meshFastIntField(view, "buffer")
		byteOffset, _ := meshFastIntField(view, "byteOffset")
		byteLength, _ := meshFastIntField(view, "byteLength")
		if bufferIndex != 0 || byteOffset < 0 || byteLength < 0 ||
			byteOffset+byteLength > len(buffers[0]) {
			return false
		}
		if imageBufferViews[i] {
			return false
		}
	}
	return true
}

func meshFastAlign4(n int) int {
	if rem := n % 4; rem != 0 {
		return n + (4 - rem)
	}
	return n
}

func meshFastGLTFSubmeshes(
	doc map[string]any,
	imageSources map[int]meshFastGLTFImageSource,
) ([]meshFastGLTFSubmesh, error) {
	nodes := meshFastArrayField(doc, "nodes")
	meshes := meshFastArrayField(doc, "meshes")
	animated := meshFastGLTFAnimatedNodes(doc)
	usedKeys := map[string]int{}
	out := []meshFastGLTFSubmesh{}
	for nodeIndex := range nodes {
		node, ok := meshFastMap(nodes[nodeIndex])
		if !ok {
			continue
		}
		meshIndex, ok := meshFastIntField(node, "mesh")
		if !ok {
			continue
		}
		if meshIndex < 0 || meshIndex >= len(meshes) {
			return nil, fmt.Errorf("node %d references invalid mesh %d", nodeIndex, meshIndex)
		}
		mesh, ok := meshFastMap(meshes[meshIndex])
		if !ok {
			return nil, fmt.Errorf("mesh %d is not an object", meshIndex)
		}
		primitives := meshFastArrayField(mesh, "primitives")
		nodeName, _ := meshFastStringField(node, "name")
		meshName, _ := meshFastStringField(mesh, "name")
		for primitiveIndex := range primitives {
			primitive, ok := meshFastMap(primitives[primitiveIndex])
			if !ok {
				continue
			}
			keyName := nodeName
			if keyName == "" {
				keyName = meshName
			}
			name := meshName
			if name == "" {
				name = nodeName
			}
			if name == "" {
				name = fmt.Sprintf("mesh_%d", len(out))
			}
			if primitiveIndex > 0 {
				name = fmt.Sprintf("%s_%d", name, primitiveIndex+1)
			}
			textures, textureImages := meshFastGLTFPrimitiveTextures(doc, primitive, imageSources)
			position, rotation, scale := meshFastGLTFNodeTransform(node)
			out = append(out, meshFastGLTFSubmesh{
				Key:           kaiju_mesh.StableMeshKey(keyName, len(out), usedKeys),
				Name:          name,
				NodeName:      nodeName,
				NodeIndex:     nodeIndex,
				Position:      position,
				Rotation:      rotation,
				Scale:         scale,
				Textures:      textures,
				TextureImages: textureImages,
				IsAnimated:    animated[nodeIndex],
			})
		}
	}
	return out, nil
}

func meshFastGLTFPrimitiveTextures(
	doc map[string]any,
	primitive map[string]any,
	imageSources map[int]meshFastGLTFImageSource,
) (map[string]string, map[string]int) {
	materialIndex, ok := meshFastIntField(primitive, "material")
	if !ok {
		return nil, nil
	}
	materials := meshFastArrayField(doc, "materials")
	if materialIndex < 0 || materialIndex >= len(materials) {
		return nil, nil
	}
	material, ok := meshFastMap(materials[materialIndex])
	if !ok {
		return nil, nil
	}
	textures := map[string]string{}
	textureImages := map[string]int{}
	addTexture := func(slot string, textureInfo any) {
		textureMap, ok := meshFastMap(textureInfo)
		if !ok {
			return
		}
		textureIndex, ok := meshFastIntField(textureMap, "index")
		if !ok {
			return
		}
		imageIndex, ok := meshFastGLTFTextureImage(doc, textureIndex)
		if !ok {
			return
		}
		source, ok := imageSources[imageIndex]
		if !ok {
			return
		}
		textures[slot] = source.Key
		textureImages[slot] = imageIndex
	}
	if pbr, ok := meshFastMap(material["pbrMetallicRoughness"]); ok {
		addTexture("baseColor", pbr["baseColorTexture"])
		addTexture("metallicRoughness", pbr["metallicRoughnessTexture"])
	}
	addTexture("normal", material["normalTexture"])
	addTexture("occlusion", material["occlusionTexture"])
	addTexture("emissive", material["emissiveTexture"])
	if len(textures) == 0 {
		return nil, nil
	}
	return textures, textureImages
}

func meshFastGLTFTextureImage(doc map[string]any, textureIndex int) (int, bool) {
	textures := meshFastArrayField(doc, "textures")
	if textureIndex < 0 || textureIndex >= len(textures) {
		return 0, false
	}
	texture, ok := meshFastMap(textures[textureIndex])
	if !ok {
		return 0, false
	}
	return meshFastIntField(texture, "source")
}

func meshFastGLTFAnimatedNodes(doc map[string]any) map[int]bool {
	nodes := meshFastArrayField(doc, "nodes")
	parent := make([]int, len(nodes))
	for i := range parent {
		parent[i] = -1
	}
	for i := range nodes {
		node, ok := meshFastMap(nodes[i])
		if !ok {
			continue
		}
		for _, childValue := range meshFastArrayField(node, "children") {
			child, ok := meshFastInt(childValue)
			if ok && child >= 0 && child < len(parent) {
				parent[child] = i
			}
		}
	}
	animated := map[int]bool{}
	for _, animationValue := range meshFastArrayField(doc, "animations") {
		animation, ok := meshFastMap(animationValue)
		if !ok {
			continue
		}
		for _, channelValue := range meshFastArrayField(animation, "channels") {
			channel, ok := meshFastMap(channelValue)
			if !ok {
				continue
			}
			target, ok := meshFastMap(channel["target"])
			if !ok {
				continue
			}
			nodeIndex, ok := meshFastIntField(target, "node")
			if !ok || nodeIndex < 0 || nodeIndex >= len(parent) {
				continue
			}
			for nodeIndex >= 0 {
				animated[nodeIndex] = true
				nodeIndex = parent[nodeIndex]
			}
		}
	}
	return animated
}

func meshFastGLTFNodeTransform(node map[string]any) (matrix.Vec3, matrix.Vec3, matrix.Vec3) {
	position := meshFastVec3Field(node, "translation", matrix.Vec3Zero())
	scale := meshFastVec3Field(node, "scale", matrix.Vec3One())
	rotation := matrix.Vec3Zero()
	if quat, ok := meshFastVec4Field(node, "rotation"); ok {
		rotation = matrix.QuaternionFromXYZW(quat).ToEuler()
	}
	return position, rotation, scale
}

func meshFastGLTFProcessedData(
	data meshFastGLTFPostProcData,
	imageURIs map[int]string,
	materials map[string]string,
) ([]byte, error) {
	meshFastGLTFApplyImageURIs(data.doc, imageURIs)
	meshFastGLTFApplyKaijuExtras(data.doc, data.submeshes, materials)
	return meshFastEncodeGLB(data.doc, data.bin)
}

func meshFastGLTFWriteProcessed(
	data meshFastGLTFPostProcData,
	path string,
	fs *project_file_system.FileSystem,
	imageURIs map[int]string,
	materials map[string]string,
) error {
	out, err := meshFastGLTFProcessedData(data, imageURIs, materials)
	if err != nil {
		return err
	}
	return fs.WriteFile(path, out, os.ModePerm)
}

func meshFastGLTFFinalizeImport(
	data meshFastGLTFPostProcData,
	res *ImportResult,
	fs *project_file_system.FileSystem,
	cache *Cache,
	linkedId string,
) ([]byte, error) {
	cc, err := cache.Read(res.Id)
	if err != nil {
		return nil, err
	}
	meshFastGLTFImportEmbeddedTextures(&data, res, fs, cache, linkedId)
	textureURIs := make(map[string]map[string]string, len(data.submeshes))
	for i := range data.submeshes {
		uris := meshTextureURIs(data.submeshes[i].Textures, res, fs, cache, cc.Config.SrcPath)
		if len(uris) > 0 {
			textureURIs[data.submeshes[i].Key] = uris
		}
	}
	matchTexture := func(srcPath string) rendering.MaterialTextureData {
		if depId := meshTextureDependencyId(srcPath, res, fs, cache, cc.Config.SrcPath); depId != "" {
			return rendering.MaterialTextureData{Texture: depId, Filter: "Linear"}
		}
		return rendering.MaterialTextureData{}
	}
	materialCache := newMeshMaterialSignatureCache()
	materials, err := meshFastGLTFMaterials(data, nil, res, fs, cache, linkedId, cc.Config.Name, matchTexture, materialCache)
	if err != nil {
		return nil, err
	}
	for i := range data.submeshes {
		data.submeshes[i].Material = materials[data.submeshes[i].Key]
	}
	imageURIs := meshFastGLTFTextureImageURIs(data, textureURIs, res, fs, cache, cc.Config.SrcPath)
	out, err := meshFastGLTFProcessedData(data, imageURIs, materials)
	if err != nil {
		return nil, err
	}
	cc.Config.Mesh = &MeshConfig{Submeshes: meshFastGLTFConfigSubmeshes(data.submeshes, materials, nil)}
	if err := WriteConfig(cc.Path, cc.Config, fs); err != nil {
		return nil, err
	}
	cache.IndexCachedContent(cc)
	return out, nil
}

func meshFastGLTFPostImportProcessing(
	data meshFastGLTFPostProcData,
	res *ImportResult,
	fs *project_file_system.FileSystem,
	cache *Cache,
	linkedId string,
) error {
	out, err := meshFastGLTFFinalizeImport(data, res, fs, cache, linkedId)
	if err != nil {
		return err
	}
	return fs.WriteFile(res.ContentPath().String(), out, os.ModePerm)
}

func meshFastGLTFPostReimportProcessing(
	data meshFastGLTFPostProcData,
	res *ImportResult,
	fs *project_file_system.FileSystem,
	cache *Cache,
) error {
	if len(data.submeshes) == 0 {
		return nil
	}
	cc, err := cache.Read(res.Id)
	if err != nil {
		return err
	}
	meshFastGLTFPreserveSubmeshKeys(data.submeshes, cc.Config.Mesh)
	res.Dependencies = meshLinkedTextureDependencies(res.Id, cache)
	textureURIs := make(map[string]map[string]string, len(data.submeshes))
	for i := range data.submeshes {
		uris := meshTextureURIs(data.submeshes[i].Textures, res, fs, cache, cc.Config.SrcPath)
		if len(uris) > 0 {
			textureURIs[data.submeshes[i].Key] = uris
		}
	}
	matchTexture := func(srcPath string) rendering.MaterialTextureData {
		if depId := meshTextureDependencyId(srcPath, res, fs, cache, cc.Config.SrcPath); depId != "" {
			return rendering.MaterialTextureData{Texture: depId, Filter: "Linear"}
		}
		return rendering.MaterialTextureData{}
	}
	materialCache := newMeshMaterialSignatureCache()
	materials, err := meshFastGLTFMaterials(data, cc.Config.Mesh, res, fs, cache, cc.Config.LinkedId, cc.Config.Name, matchTexture, materialCache)
	if err != nil {
		return err
	}
	for i := range data.submeshes {
		data.submeshes[i].Material = materials[data.submeshes[i].Key]
	}
	imageURIs := meshFastGLTFTextureImageURIs(data, textureURIs, res, fs, cache, cc.Config.SrcPath)
	if err := meshFastGLTFWriteProcessed(data, res.ContentPath().String(), fs, imageURIs, materials); err != nil {
		return err
	}
	cc.Config.Mesh = &MeshConfig{Submeshes: meshFastGLTFConfigSubmeshes(data.submeshes, materials, cc.Config.Mesh)}
	if err := WriteConfig(cc.Path, cc.Config, fs); err != nil {
		return err
	}
	cache.IndexCachedContent(cc)
	return nil
}

func meshFastGLTFImportEmbeddedTextures(
	data *meshFastGLTFPostProcData,
	res *ImportResult,
	fs *project_file_system.FileSystem,
	cache *Cache,
	linkedId string,
) {
	texKeyToDepId := make(map[string]string, len(data.textureBytes))
	for texKey, textureData := range data.textureBytes {
		ext := meshEmbeddedTextureExtension(textureData)
		tf, err := os.CreateTemp("", "*-kaiju-texture"+ext)
		if err != nil {
			continue
		}
		if _, err := tf.Write(textureData); err != nil {
			tf.Close()
			os.Remove(tf.Name())
			continue
		}
		tf.Close()
		texRes, err := Import(tf.Name(), fs, cache, linkedId)
		os.Remove(tf.Name())
		if err != nil || len(texRes) == 0 {
			continue
		}
		res.Dependencies = append(res.Dependencies, texRes[0])
		texKeyToDepId[texKey] = texRes[0].Id
	}
	for i := range data.submeshes {
		for slot, texKey := range data.submeshes[i].Textures {
			if depId, ok := texKeyToDepId[texKey]; ok {
				data.submeshes[i].Textures[slot] = depId
			} else if strings.HasPrefix(texKey, "embedded_") {
				data.submeshes[i].Textures[slot] = ""
			}
		}
	}
	for imageIndex, source := range data.imageSources {
		if !source.Embedded {
			continue
		}
		if depId := texKeyToDepId[source.Key]; depId != "" {
			source.Key = depId
			data.imageSources[imageIndex] = source
		}
	}
}

func meshFastGLTFMaterials(
	data meshFastGLTFPostProcData,
	old *MeshConfig,
	res *ImportResult,
	fs *project_file_system.FileSystem,
	cache *Cache,
	linkedId string,
	assetName string,
	matchTexture func(string) rendering.MaterialTextureData,
	materialCache *meshMaterialSignatureCache,
) (map[string]string, error) {
	oldByKey := make(map[string]string)
	if old != nil {
		for i := range old.Submeshes {
			if old.Submeshes[i].Key != "" && old.Submeshes[i].Material != "" {
				oldByKey[old.Submeshes[i].Key] = old.Submeshes[i].Material
			}
		}
	}
	materials := make(map[string]string, len(data.submeshes))
	for i := range data.submeshes {
		key := data.submeshes[i].Key
		if matId := oldByKey[key]; matId != "" {
			materials[key] = matId
			continue
		}
		mat := meshMaterialData(data.submeshes[i].Textures, data.submeshes[i].IsAnimated, matchTexture)
		matName := meshImportMaterialName(assetName, len(data.submeshes), data.submeshes[i].Name)
		matId, err := importOrFindMeshMaterial(mat, matName, res, fs, cache, linkedId, materialCache)
		if err != nil {
			return nil, err
		}
		materials[key] = matId
	}
	return materials, nil
}

func meshFastGLTFApplyImageURIs(doc map[string]any, imageURIs map[int]string) {
	images := meshFastArrayField(doc, "images")
	for imageIndex, uri := range imageURIs {
		if imageIndex < 0 || imageIndex >= len(images) || uri == "" {
			continue
		}
		image, ok := meshFastMap(images[imageIndex])
		if !ok {
			continue
		}
		image["uri"] = uri
		delete(image, "bufferView")
		delete(image, "mimeType")
	}
}

func meshFastGLTFApplyKaijuExtras(
	doc map[string]any,
	submeshes []meshFastGLTFSubmesh,
	materials map[string]string,
) {
	extras, ok := meshFastMap(doc["extras"])
	if !ok {
		extras = map[string]any{}
		doc["extras"] = extras
	}
	kaiju, ok := meshFastMap(extras["kaiju"])
	if !ok {
		kaiju = map[string]any{}
		extras["kaiju"] = kaiju
	}
	if _, ok := kaiju["version"]; !ok {
		kaiju["version"] = 1
	}
	oldExtras := meshFastGLTFExistingMeshExtras(kaiju)
	meshExtras := make([]any, len(submeshes))
	for i := range submeshes {
		extra := meshFastCloneMap(oldExtras[i])
		extra["key"] = submeshes[i].Key
		extra["name"] = submeshes[i].Name
		extra["mesh"] = i
		extra["node"] = submeshes[i].NodeIndex
		material := submeshes[i].Material
		if materials != nil && materials[submeshes[i].Key] != "" {
			material = materials[submeshes[i].Key]
		}
		if material != "" {
			extra["material"] = material
		} else {
			delete(extra, "material")
		}
		meshExtras[i] = extra
	}
	kaiju["meshes"] = meshExtras
}

func meshFastGLTFExistingMeshExtras(kaiju map[string]any) map[int]map[string]any {
	out := map[int]map[string]any{}
	for i, value := range meshFastArrayField(kaiju, "meshes") {
		extra, ok := meshFastMap(value)
		if !ok {
			continue
		}
		meshIndex, ok := meshFastIntField(extra, "mesh")
		if !ok {
			meshIndex = i
		}
		out[meshIndex] = extra
	}
	return out
}

func meshFastGLTFConfigSubmeshes(
	submeshes []meshFastGLTFSubmesh,
	materials map[string]string,
	old *MeshConfig,
) []MeshSubmeshConfig {
	set := kaiju_mesh.KaijuMeshSet{Meshes: make([]kaiju_mesh.KaijuMesh, len(submeshes))}
	for i := range submeshes {
		set.Meshes[i] = kaiju_mesh.KaijuMesh{
			Key:      submeshes[i].Key,
			Name:     submeshes[i].Name,
			Material: submeshes[i].Material,
			Node: kaiju_mesh.KaijuMeshNode{
				Name:     submeshes[i].NodeName,
				Position: submeshes[i].Position,
				Rotation: submeshes[i].Rotation,
				Scale:    submeshes[i].Scale,
			},
		}
	}
	return meshConfigSubmeshes(set, materials, old)
}

func meshFastGLTFPreserveSubmeshKeys(submeshes []meshFastGLTFSubmesh, old *MeshConfig) {
	set := kaiju_mesh.KaijuMeshSet{Meshes: make([]kaiju_mesh.KaijuMesh, len(submeshes))}
	for i := range submeshes {
		set.Meshes[i] = kaiju_mesh.KaijuMesh{
			Key:  submeshes[i].Key,
			Name: submeshes[i].Name,
			Node: kaiju_mesh.KaijuMeshNode{Name: submeshes[i].NodeName},
		}
	}
	preserveMeshSetKeys(&set, old)
	for i := range submeshes {
		submeshes[i].Key = set.Meshes[i].Key
	}
}

func meshFastGLTFTextureImageURIs(
	data meshFastGLTFPostProcData,
	textureURIs map[string]map[string]string,
	res *ImportResult,
	fs *project_file_system.FileSystem,
	cache *Cache,
	meshSrcPath string,
) map[int]string {
	out := map[int]string{}
	for i := range data.submeshes {
		uris := textureURIs[data.submeshes[i].Key]
		for slot, imageIndex := range data.submeshes[i].TextureImages {
			if uri := uris[slot]; uri != "" {
				out[imageIndex] = uri
			}
		}
	}
	for imageIndex, source := range data.imageSources {
		if out[imageIndex] != "" {
			continue
		}
		if uri := meshFastTextureURI(source.Key, res, fs, cache, meshSrcPath); uri != "" {
			out[imageIndex] = uri
		}
	}
	return out
}

func meshFastTextureURI(
	texKey string,
	res *ImportResult,
	fs *project_file_system.FileSystem,
	cache *Cache,
	meshSrcPath string,
) string {
	depId := meshTextureDependencyId(texKey, res, fs, cache, meshSrcPath)
	if depId == "" {
		return ""
	}
	uri, err := filepath.Rel(filepath.Dir(res.ContentPath().String()),
		project_file_system.AsContentPath(filepath.Join(
			project_file_system.ContentFolder,
			project_file_system.ContentTextureFolder,
			depId)).String())
	if err != nil {
		return ""
	}
	return filepath.ToSlash(uri)
}

func meshFastEncodeGLB(doc map[string]any, bin []byte) ([]byte, error) {
	if buffers := meshFastArrayField(doc, "buffers"); len(buffers) > 0 {
		if buffer, ok := meshFastMap(buffers[0]); ok {
			buffer["byteLength"] = len(bin)
			delete(buffer, "uri")
		}
	}
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	jsonBytes = meshFastPadded(jsonBytes, ' ')
	bin = meshFastPadded(bin, 0)
	totalLen := 12 + 8 + len(jsonBytes) + 8 + len(bin)
	out := make([]byte, 0, totalLen)
	out = append(out, fastGLBMagic[:]...)
	out = binary.LittleEndian.AppendUint32(out, 2)
	out = binary.LittleEndian.AppendUint32(out, uint32(totalLen))
	out = binary.LittleEndian.AppendUint32(out, uint32(len(jsonBytes)))
	out = binary.LittleEndian.AppendUint32(out, fastGLBJSONChunkType)
	out = append(out, jsonBytes...)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(bin)))
	out = binary.LittleEndian.AppendUint32(out, fastGLBBINChunkType)
	out = append(out, bin...)
	return out, nil
}

func meshFastPadded(data []byte, pad byte) []byte {
	out := slices.Clone(data)
	for len(out)%4 != 0 {
		out = append(out, pad)
	}
	return out
}

func meshFastArrayField(m map[string]any, key string) []any {
	if value, ok := m[key]; ok {
		if out, ok := value.([]any); ok {
			return out
		}
	}
	return nil
}

func meshFastMap(value any) (map[string]any, bool) {
	out, ok := value.(map[string]any)
	return out, ok
}

func meshFastCloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func meshFastStringField(m map[string]any, key string) (string, bool) {
	value, ok := m[key].(string)
	return value, ok
}

func meshFastIntField(m map[string]any, key string) (int, bool) {
	return meshFastInt(m[key])
}

func meshFastInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return int(i), true
		}
		f, err := v.Float64()
		if err == nil {
			return int(f), true
		}
	}
	return 0, false
}

func meshFastFloat(value any) (matrix.Float, bool) {
	switch v := value.(type) {
	case float32:
		return matrix.Float(v), true
	case float64:
		return matrix.Float(v), true
	case int:
		return matrix.Float(v), true
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return matrix.Float(f), true
		}
	}
	return 0, false
}

func meshFastVec3Field(m map[string]any, key string, fallback matrix.Vec3) matrix.Vec3 {
	values, ok := m[key].([]any)
	if !ok || len(values) < 3 {
		return fallback
	}
	out := fallback
	for i := 0; i < 3; i++ {
		if value, ok := meshFastFloat(values[i]); ok {
			out[i] = value
		}
	}
	return out
}

func meshFastVec4Field(m map[string]any, key string) (matrix.Vec4, bool) {
	values, ok := m[key].([]any)
	if !ok || len(values) < 4 {
		return matrix.Vec4{}, false
	}
	out := matrix.Vec4{}
	for i := 0; i < 4; i++ {
		value, ok := meshFastFloat(values[i])
		if !ok {
			return matrix.Vec4{}, false
		}
		out[i] = value
	}
	return out, true
}

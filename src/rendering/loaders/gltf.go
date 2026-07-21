/******************************************************************************/
/* gltf.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package loaders

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/gltf"
	"kaijuengine.com/rendering/loaders/load_result"
)

type fullGLTF struct {
	path     string
	glTF     gltf.GLTF
	bins     [][]byte
	textures map[int32][]byte
}

type GLTFLoadOptions struct {
	Workers int
}

func (o GLTFLoadOptions) workerCount() int {
	if o.Workers > 0 {
		return o.Workers
	}
	return max(1, runtime.GOMAXPROCS(0))
}

type gltfPrimitiveTask struct {
	Index          int
	NodeIndex      int
	MeshIndex      int
	PrimitiveIndex int
	NodeName       string
	Key            string
}

type gltfPrimitiveResult struct {
	Task         gltfPrimitiveTask
	Verts        []rendering.Vertex
	Indices      []uint32
	Textures     map[string]string
	TextureBytes map[string][]byte
	Err          error
}

type gltfAccessorView struct {
	accessor       *gltf.Accessor
	view           *gltf.BufferView
	data           []byte
	start          int
	stride         int
	componentSize  int
	componentCount int
	elementSize    int
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
	// version := binary.LittleEndian.Uint32(data[4:8])
	// length := binary.LittleEndian.Uint32(data[8:12])
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
	g.textures = make(map[int32][]byte)
	if err = gltfExtractEmbeddedTextures(&g); err != nil {
		return g, err
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
	g.textures = make(map[int32][]byte)
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
	if err = gltfExtractEmbeddedTextures(&g); err != nil {
		return g, err
	}
	return g, nil
}

func gltfExtractEmbeddedTextures(g *fullGLTF) error {
	for i := range g.glTF.Images {
		img := &g.glTF.Images[i]
		if img.BufferView != nil {
			if *img.BufferView < 0 || int(*img.BufferView) >= len(g.glTF.BufferViews) {
				return fmt.Errorf("image %d references invalid bufferView %d", i, *img.BufferView)
			}
			view := g.glTF.BufferViews[*img.BufferView]
			if view.Buffer < 0 || int(view.Buffer) >= len(g.bins) {
				return fmt.Errorf("image %d bufferView references invalid buffer %d", i, view.Buffer)
			}
			bin := g.bins[view.Buffer]
			end := view.ByteOffset + view.ByteLength
			if view.ByteOffset < 0 || view.ByteLength < 0 || end > int32(len(bin)) {
				return fmt.Errorf("image %d bufferView exceeds buffer bounds", i)
			}
			g.textures[int32(i)] = bin[view.ByteOffset:end]
		} else if strings.HasPrefix(img.URI, "data:") {
			comma := strings.Index(img.URI, ",")
			if comma == -1 {
				continue
			}
			b64 := img.URI[comma+1:]
			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err == nil {
				g.textures[int32(i)] = decoded
			}
		}
	}
	return nil
}

func GLTF(path string, assetDB assets.Database) (load_result.Result, error) {
	defer tracing.NewRegion("loaders.GLTF").End()
	return GLTFWithOptions(path, assetDB, GLTFLoadOptions{})
}

func GLTFWithOptions(path string, assetDB assets.Database, opts GLTFLoadOptions) (load_result.Result, error) {
	defer tracing.NewRegion("loaders.GLTFWithOptions").End()
	if !assetDB.Exists(path) {
		return load_result.Result{}, errors.New("file does not exist")
	} else if filepath.Ext(path) == ".glb" {
		if g, err := readFileGLB(path, assetDB); err != nil {
			return load_result.Result{}, err
		} else {
			return gltfParse(&g, opts.workerCount())
		}
	} else if filepath.Ext(path) == ".gltf" {
		if g, err := readFileGLTF(path, assetDB); err != nil {
			return load_result.Result{}, err
		} else {
			return gltfParse(&g, opts.workerCount())
		}
	} else {
		return load_result.Result{}, errors.New("invalid file extension")
	}
}

func gltfParse(doc *fullGLTF, workers int) (load_result.Result, error) {
	defer tracing.NewRegion("loaders.gltfParse").End()
	res := load_result.Result{}
	res.TextureBytes = make(map[string][]byte)
	res.Nodes = make([]load_result.Node, len(doc.glTF.Nodes))
	for i := range res.Nodes {
		res.Nodes[i].Parent = -1
		res.Nodes[i].Attributes = make(map[string]any)
	}
	// TODO:  Deal with multiple skins
	if len(doc.glTF.Skins) > 0 {
		skin, err := gltfAccessor(doc, doc.glTF.Skins[0].InverseBindMatrices)
		if err != nil {
			return res, err
		}
		if err = gltfValidateAccessor(skin, gltf.FLOAT, gltf.MAT4, "inverse bind matrices"); err != nil {
			return res, err
		}
		for i, id := range doc.glTF.Skins[0].Joints {
			if i >= int(skin.accessor.Count) {
				return res, errors.New("skin joint count exceeds inverse bind matrix count")
			}
			mat := make([]matrix.Float, 16)
			for j := range mat {
				mat[j] = skin.float(i, j)
			}
			res.Joints = append(res.Joints, load_result.Joint{
				Id:   id,
				Skin: matrix.Mat4FromSlice(mat),
			})
		}
	}
	tasks := make([]gltfPrimitiveTask, 0)
	for i := range doc.glTF.Nodes {
		n := &doc.glTF.Nodes[i]
		res.Nodes[i].Id = int32(i)
		res.Nodes[i].Name = n.Name
		res.Nodes[i].Attributes = n.Extras
		for j := range n.Children {
			cid := n.Children[j]
			if cid < 0 || int(cid) >= len(res.Nodes) {
				return res, fmt.Errorf("node %d references invalid child %d", i, cid)
			}
			res.Nodes[cid].Parent = i
		}
		// TODO:  Come back for this scenario
		//if n.Matrix != nil {
		//}
		if n.Scale != nil {
			res.Nodes[i].Scale = *n.Scale
		} else {
			res.Nodes[i].Scale = matrix.Vec3One()
		}
		if n.Rotation != nil {
			res.Nodes[i].Rotation = matrix.QuaternionFromXYZW(*n.Rotation)
		} else {
			res.Nodes[i].Rotation = matrix.QuaternionIdentity()
		}
		if n.Translation != nil {
			res.Nodes[i].Position = *n.Translation
		}
		if n.Mesh == nil {
			continue
		}
		if int(*n.Mesh) < 0 || int(*n.Mesh) >= len(doc.glTF.Meshes) {
			return res, fmt.Errorf("node %d references invalid mesh %d", i, *n.Mesh)
		}
		m := &doc.glTF.Meshes[*n.Mesh]
		for p := range m.Primitives {
			key := fmt.Sprintf("%s/%s", doc.path, m.Name)
			if p > 0 {
				key += fmt.Sprintf("_%d", p+1)
			}
			tasks = append(tasks, gltfPrimitiveTask{
				Index:          len(tasks),
				NodeIndex:      i,
				MeshIndex:      int(*n.Mesh),
				PrimitiveIndex: p,
				NodeName:       n.Name,
				Key:            key,
			})
		}
	}
	meshResults := gltfReadPrimitiveTasks(doc, tasks, workers)
	for i := range meshResults {
		if meshResults[i].Err != nil {
			return res, meshResults[i].Err
		}
		for key, bytes := range meshResults[i].TextureBytes {
			res.TextureBytes[key] = bytes
		}
		task := meshResults[i].Task
		res.Add(task.NodeName, task.Key, meshResults[i].Verts,
			meshResults[i].Indices, meshResults[i].Textures, &res.Nodes[task.NodeIndex])
		loadedMesh := &res.Meshes[len(res.Meshes)-1]
		loadedMesh.MaterialAlphaMode, loadedMesh.MaterialAlphaCutoff, loadedMesh.MaterialDoubleSided =
			gltfPrimitiveMaterial(doc, task.MeshIndex, task.PrimitiveIndex)
	}
	var err error
	res.Animations, err = gltfReadAnimations(doc, workers)
	if err != nil {
		return res, err
	}
	for i := range doc.glTF.Animations {
		for j := range doc.glTF.Animations[i].Channels {
			nid := doc.glTF.Animations[i].Channels[j].Target.Node
			if nid < 0 || int(nid) >= len(res.Nodes) {
				return res, fmt.Errorf("animation %d channel %d references invalid node %d", i, j, nid)
			}
			res.Nodes[nid].IsAnimated = true
			p := res.Nodes[nid].Parent
			for p >= 0 {
				res.Nodes[p].IsAnimated = true
				p = res.Nodes[p].Parent
			}
		}
	}
	return res, nil
}

func gltfPrimitiveMaterial(doc *fullGLTF, meshIndex, primitiveIndex int) (string, matrix.Float, bool) {
	const defaultAlphaCutoff = matrix.Float(0.5)
	if meshIndex < 0 || meshIndex >= len(doc.glTF.Meshes) ||
		primitiveIndex < 0 || primitiveIndex >= len(doc.glTF.Meshes[meshIndex].Primitives) {
		return "OPAQUE", defaultAlphaCutoff, false
	}
	materialIndex := doc.glTF.Meshes[meshIndex].Primitives[primitiveIndex].Material
	if materialIndex == nil || *materialIndex < 0 || int(*materialIndex) >= len(doc.glTF.Materials) {
		return "OPAQUE", defaultAlphaCutoff, false
	}
	material := &doc.glTF.Materials[*materialIndex]
	alphaMode := material.AlphaMode
	if alphaMode == "" {
		alphaMode = "OPAQUE"
	}
	alphaCutoff := defaultAlphaCutoff
	if material.AlphaCutoff != nil {
		alphaCutoff = matrix.Float(*material.AlphaCutoff)
	}
	return alphaMode, alphaCutoff, material.DoubleSided
}

func gltfAttr(primitive gltf.Primitive, cmp string) (uint32, bool) {
	defer tracing.NewRegion("loaders.gltfAttr").End()
	idx, ok := primitive.Attributes[cmp]
	return idx, ok
}

func gltfViewBytes(doc *fullGLTF, view *gltf.BufferView) []byte {
	defer tracing.NewRegion("loaders.gltfViewBytes").End()
	return doc.bins[view.Buffer][view.ByteOffset : view.ByteOffset+view.ByteLength]
}

func gltfAccessor(doc *fullGLTF, index int32) (gltfAccessorView, error) {
	if index < 0 || int(index) >= len(doc.glTF.Accessors) {
		return gltfAccessorView{}, fmt.Errorf("invalid accessor index %d", index)
	}
	acc := &doc.glTF.Accessors[index]
	if acc.BufferView < 0 || int(acc.BufferView) >= len(doc.glTF.BufferViews) {
		return gltfAccessorView{}, fmt.Errorf("accessor %d references invalid buffer view %d", index, acc.BufferView)
	}
	view := &doc.glTF.BufferViews[acc.BufferView]
	if view.Buffer < 0 || int(view.Buffer) >= len(doc.bins) {
		return gltfAccessorView{}, fmt.Errorf("buffer view %d references invalid buffer %d", acc.BufferView, view.Buffer)
	}
	if acc.Count < 0 || view.ByteOffset < 0 || view.ByteLength < 0 || acc.ByteOffset < 0 || view.ByteStride < 0 {
		return gltfAccessorView{}, fmt.Errorf("accessor %d has negative sizing metadata", index)
	}
	componentSize, err := gltfComponentSize(acc.ComponentType)
	if err != nil {
		return gltfAccessorView{}, err
	}
	componentCount, err := gltfAccessorComponentCount(acc.Type)
	if err != nil {
		return gltfAccessorView{}, err
	}
	elementSize := componentSize * componentCount
	stride := elementSize
	if view.ByteStride > 0 {
		stride = int(view.ByteStride)
	}
	if stride < elementSize {
		return gltfAccessorView{}, fmt.Errorf("accessor %d stride %d is smaller than element size %d", index, stride, elementSize)
	}
	data := doc.bins[view.Buffer]
	start := int64(view.ByteOffset) + int64(acc.ByteOffset)
	viewStart := int64(view.ByteOffset)
	viewEnd := viewStart + int64(view.ByteLength)
	if start < viewStart || start > viewEnd || viewEnd > int64(len(data)) {
		return gltfAccessorView{}, fmt.Errorf("accessor %d byte range exceeds buffer", index)
	}
	if acc.Count > 0 {
		lastEnd := start + int64(acc.Count-1)*int64(stride) + int64(elementSize)
		if lastEnd > viewEnd || lastEnd > int64(len(data)) {
			return gltfAccessorView{}, fmt.Errorf("accessor %d element range exceeds buffer view", index)
		}
	}
	return gltfAccessorView{
		accessor:       acc,
		view:           view,
		data:           data,
		start:          int(start),
		stride:         stride,
		componentSize:  componentSize,
		componentCount: componentCount,
		elementSize:    elementSize,
	}, nil
}

func gltfPrimitiveAccessor(doc *fullGLTF, primitive gltf.Primitive, semantic string) (gltfAccessorView, bool, error) {
	idx, ok := gltfAttr(primitive, semantic)
	if !ok {
		return gltfAccessorView{}, false, nil
	}
	acc, err := gltfAccessor(doc, int32(idx))
	return acc, true, err
}

func gltfRequiredPrimitiveAccessor(doc *fullGLTF, primitive gltf.Primitive, semantic string) (gltfAccessorView, error) {
	acc, ok, err := gltfPrimitiveAccessor(doc, primitive, semantic)
	if err != nil {
		return gltfAccessorView{}, err
	}
	if !ok {
		return gltfAccessorView{}, fmt.Errorf("missing required glTF attribute %s", semantic)
	}
	return acc, nil
}

func gltfValidateAccessor(acc gltfAccessorView, componentType gltf.ComponentType, accessorType gltf.AccessorType, name string) error {
	if acc.accessor.ComponentType != componentType || acc.accessor.Type != accessorType {
		return fmt.Errorf("%s accessor must be %s/%d", name, accessorType, componentType)
	}
	return nil
}

func gltfValidateAccessorCount(acc gltfAccessorView, count int32, name string) error {
	if acc.accessor.Count != count {
		return fmt.Errorf("%s accessor count %d does not match position count %d", name, acc.accessor.Count, count)
	}
	return nil
}

func gltfComponentSize(componentType gltf.ComponentType) (int, error) {
	switch componentType {
	case gltf.BYTE, gltf.UNSIGNED_BYTE:
		return 1, nil
	case gltf.SHORT, gltf.UNSIGNED_SHORT:
		return 2, nil
	case gltf.UNSIGNED_INT, gltf.FLOAT:
		return 4, nil
	default:
		return 0, fmt.Errorf("invalid glTF component type %d", componentType)
	}
}

func gltfAccessorComponentCount(accessorType gltf.AccessorType) (int, error) {
	switch accessorType {
	case gltf.SCALAR:
		return 1, nil
	case gltf.VEC2:
		return 2, nil
	case gltf.VEC3:
		return 3, nil
	case gltf.VEC4:
		return 4, nil
	case gltf.MAT2:
		return 4, nil
	case gltf.MAT3:
		return 9, nil
	case gltf.MAT4:
		return 16, nil
	default:
		return 0, fmt.Errorf("invalid glTF accessor type %q", accessorType)
	}
}

func (a gltfAccessorView) componentBytes(element, component int) []byte {
	offset := a.start + element*a.stride + component*a.componentSize
	return a.data[offset : offset+a.componentSize]
}

func (a gltfAccessorView) float(element, component int) matrix.Float {
	bytes := a.componentBytes(element, component)
	return matrix.Float(math.Float32frombits(binary.LittleEndian.Uint32(bytes)))
}

func (a gltfAccessorView) normalizedFloat(element, component int) matrix.Float {
	bytes := a.componentBytes(element, component)
	switch a.accessor.ComponentType {
	case gltf.BYTE:
		return max(matrix.Float(int8(bytes[0]))/127, -1)
	case gltf.UNSIGNED_BYTE:
		return matrix.Float(bytes[0]) / 255
	case gltf.SHORT:
		return max(matrix.Float(int16(binary.LittleEndian.Uint16(bytes)))/32767, -1)
	case gltf.UNSIGNED_SHORT:
		return matrix.Float(binary.LittleEndian.Uint16(bytes)) / 65535
	case gltf.FLOAT:
		return matrix.Float(math.Float32frombits(binary.LittleEndian.Uint32(bytes)))
	default:
		return 0
	}
}

func (a gltfAccessorView) int32(element, component int) int32 {
	bytes := a.componentBytes(element, component)
	switch a.accessor.ComponentType {
	case gltf.BYTE:
		return int32(int8(bytes[0]))
	case gltf.UNSIGNED_BYTE:
		return int32(bytes[0])
	case gltf.SHORT:
		return int32(int16(binary.LittleEndian.Uint16(bytes)))
	case gltf.UNSIGNED_SHORT:
		return int32(binary.LittleEndian.Uint16(bytes))
	case gltf.UNSIGNED_INT:
		return int32(binary.LittleEndian.Uint32(bytes))
	default:
		return 0
	}
}

func (a gltfAccessorView) uint32(element, component int) uint32 {
	bytes := a.componentBytes(element, component)
	switch a.accessor.ComponentType {
	case gltf.BYTE, gltf.UNSIGNED_BYTE:
		return uint32(bytes[0])
	case gltf.SHORT, gltf.UNSIGNED_SHORT:
		return uint32(binary.LittleEndian.Uint16(bytes))
	case gltf.UNSIGNED_INT:
		return binary.LittleEndian.Uint32(bytes)
	default:
		return 0
	}
}

func gltfParallelFor(count, workers, minItemsPerWorker int, work func(from, to int)) {
	if count <= 0 {
		return
	}
	if workers <= 1 || count < minItemsPerWorker {
		work(0, count)
		return
	}
	workerCount := min(workers, count/minItemsPerWorker)
	if workerCount <= 1 {
		work(0, count)
		return
	}
	group := sync.WaitGroup{}
	group.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		from := i * count / workerCount
		to := (i + 1) * count / workerCount
		go func() {
			defer group.Done()
			work(from, to)
		}()
	}
	group.Wait()
}

func gltfReadPrimitiveTasks(doc *fullGLTF, tasks []gltfPrimitiveTask, workers int) []gltfPrimitiveResult {
	results := make([]gltfPrimitiveResult, len(tasks))
	if len(tasks) == 0 {
		return results
	}
	workers = max(1, workers)
	taskWorkers := min(workers, len(tasks))
	chunkWorkers := max(1, workers/taskWorkers)
	if taskWorkers <= 1 {
		for i := range tasks {
			results[i] = gltfReadPrimitiveTask(doc, tasks[i], workers)
		}
		return results
	}
	jobs := make(chan int)
	group := sync.WaitGroup{}
	group.Add(taskWorkers)
	for range taskWorkers {
		go func() {
			defer group.Done()
			for idx := range jobs {
				results[idx] = gltfReadPrimitiveTask(doc, tasks[idx], chunkWorkers)
			}
		}()
	}
	for i := range tasks {
		jobs <- i
	}
	close(jobs)
	group.Wait()
	return results
}

func gltfReadPrimitiveTask(doc *fullGLTF, task gltfPrimitiveTask, workers int) gltfPrimitiveResult {
	mesh := &doc.glTF.Meshes[task.MeshIndex]
	verts, err := gltfReadMeshVerts(mesh, doc, task.PrimitiveIndex, workers)
	if err != nil {
		return gltfPrimitiveResult{Task: task, Err: err}
	}
	indices, err := gltfReadMeshIndices(mesh, doc, task.PrimitiveIndex, workers)
	if err != nil {
		return gltfPrimitiveResult{Task: task, Err: err}
	}
	textures, textureBytes := gltfReadMeshTextures(mesh, doc, task.PrimitiveIndex)
	return gltfPrimitiveResult{
		Task:         task,
		Verts:        verts,
		Indices:      indices,
		Textures:     textures,
		TextureBytes: textureBytes,
	}
}

func gltfTextureKey(doc *fullGLTF, imageIdx int32, texType string) (string, bool) {
	if imageIdx < 0 || int(imageIdx) >= len(doc.glTF.Images) {
		return "", false
	}
	if _, ok := doc.textures[imageIdx]; ok {
		return "embedded_" + strconv.FormatInt(int64(imageIdx), 10) + "_" + texType, true
	}
	img := doc.glTF.Images[imageIdx]
	if img.URI == "" {
		return "", false
	}
	return filepath.ToSlash(filepath.Join(filepath.Dir(doc.path), img.URI)), false
}

func gltfReadMeshMorphTargets(mesh *gltf.Mesh, doc *fullGLTF, primitive int, verts []rendering.Vertex, workers int) error {
	defer tracing.NewRegion("loaders.gltfReadMeshMorphTargets").End()
	for _, target := range mesh.Primitives[primitive].Targets {
		if target.POSITION == nil {
			continue
		}
		acc, err := gltfAccessor(doc, *target.POSITION)
		if err != nil {
			return err
		}
		if err = gltfValidateAccessor(acc, gltf.FLOAT, gltf.VEC3, "morph target position"); err != nil {
			return err
		}
		if int(acc.accessor.Count) != len(verts) {
			return errors.New("morph targets do not match vert count")
		}
		if acc.accessor.Count <= 0 {
			continue
		}
		gltfParallelFor(len(verts), workers, 8192, func(from, to int) {
			for i := from; i < to; i++ {
				verts[i].MorphTarget = matrix.Vec3{
					acc.float(i, 0),
					acc.float(i, 1),
					acc.float(i, 2),
				}
			}
		})
	}
	return nil
}

func gltfReadMeshVerts(mesh *gltf.Mesh, doc *fullGLTF, primitive int, workers int) ([]rendering.Vertex, error) {
	defer tracing.NewRegion("loaders.gltfReadMeshVerts").End()
	prim := mesh.Primitives[primitive]
	pos, err := gltfRequiredPrimitiveAccessor(doc, prim, gltf.POSITION)
	if err != nil {
		return []rendering.Vertex{}, err
	}
	nml, err := gltfRequiredPrimitiveAccessor(doc, prim, gltf.NORMAL)
	if err != nil {
		return []rendering.Vertex{}, err
	}
	tan, hasTan, err := gltfPrimitiveAccessor(doc, prim, gltf.TANGENT)
	if err != nil {
		return []rendering.Vertex{}, err
	}
	tex0, hasTex0, err := gltfPrimitiveAccessor(doc, prim, gltf.TEXCOORD_0)
	if err != nil {
		return []rendering.Vertex{}, err
	}
	tex1, hasTex1, err := gltfPrimitiveAccessor(doc, prim, gltf.TEXCOORD_1)
	if err != nil {
		return []rendering.Vertex{}, err
	}
	col0, hasCol0, err := gltfPrimitiveAccessor(doc, prim, gltf.COLOR_0)
	if err != nil {
		return []rendering.Vertex{}, err
	}
	jnt0, hasJnt0, err := gltfPrimitiveAccessor(doc, prim, gltf.JOINTS_0)
	if err != nil {
		return []rendering.Vertex{}, err
	}
	wei0, hasWei0, err := gltfPrimitiveAccessor(doc, prim, gltf.WEIGHTS_0)
	if err != nil {
		return []rendering.Vertex{}, err
	}

	vertCount := pos.accessor.Count
	if !(vertCount > 0) {
		return []rendering.Vertex{}, errors.New("vertCount <= 0")
	}
	if err = gltfValidateAccessor(pos, gltf.FLOAT, gltf.VEC3, "position"); err != nil {
		return []rendering.Vertex{}, err
	}
	if err = gltfValidateAccessor(nml, gltf.FLOAT, gltf.VEC3, "normal"); err != nil {
		return []rendering.Vertex{}, err
	}
	if err = gltfValidateAccessorCount(nml, vertCount, "normal"); err != nil {
		return []rendering.Vertex{}, err
	}
	if hasTan {
		if err = gltfValidateAccessor(tan, gltf.FLOAT, gltf.VEC4, "tangent"); err != nil {
			return []rendering.Vertex{}, err
		}
		if err = gltfValidateAccessorCount(tan, vertCount, "tangent"); err != nil {
			return []rendering.Vertex{}, err
		}
	}
	if hasTex0 {
		if err = gltfValidateAccessor(tex0, gltf.FLOAT, gltf.VEC2, "texcoord0"); err != nil {
			return []rendering.Vertex{}, err
		}
		if err = gltfValidateAccessorCount(tex0, vertCount, "texcoord0"); err != nil {
			return []rendering.Vertex{}, err
		}
	}
	if hasTex1 {
		if err = gltfValidateAccessor(tex1, gltf.FLOAT, gltf.VEC2, "texcoord1"); err != nil {
			return []rendering.Vertex{}, err
		}
		if err = gltfValidateAccessorCount(tex1, vertCount, "texcoord1"); err != nil {
			return []rendering.Vertex{}, err
		}
	}
	if hasCol0 {
		validComponent := col0.accessor.ComponentType == gltf.FLOAT ||
			col0.accessor.ComponentType == gltf.UNSIGNED_BYTE ||
			col0.accessor.ComponentType == gltf.UNSIGNED_SHORT
		if !validComponent || (col0.accessor.Type != gltf.VEC3 && col0.accessor.Type != gltf.VEC4) {
			return []rendering.Vertex{}, errors.New("COLOR_0 must be FLOAT, normalized UNSIGNED_BYTE, or normalized UNSIGNED_SHORT VEC3/VEC4")
		}
		if col0.accessor.ComponentType != gltf.FLOAT && !col0.accessor.Normalized {
			return []rendering.Vertex{}, errors.New("integer COLOR_0 accessors must be normalized")
		}
		if err = gltfValidateAccessorCount(col0, vertCount, "color0"); err != nil {
			return []rendering.Vertex{}, err
		}
	}
	if hasJnt0 != hasWei0 {
		return []rendering.Vertex{}, errors.New("JOINTS_0 and WEIGHTS_0 must both be present")
	}
	if hasJnt0 {
		if jnt0.accessor.Type != gltf.VEC4 {
			return []rendering.Vertex{}, errors.New("JOINTS_0 must be VEC4")
		}
		if jnt0.accessor.ComponentType != gltf.UNSIGNED_BYTE &&
			jnt0.accessor.ComponentType != gltf.UNSIGNED_SHORT &&
			jnt0.accessor.ComponentType != gltf.UNSIGNED_INT {
			return []rendering.Vertex{}, errors.New("JOINTS_0 must use an unsigned integer component type")
		}
		if err = gltfValidateAccessorCount(jnt0, vertCount, "joints0"); err != nil {
			return []rendering.Vertex{}, err
		}
		if err = gltfValidateAccessor(wei0, gltf.FLOAT, gltf.VEC4, "weights0"); err != nil {
			return []rendering.Vertex{}, err
		}
		if err = gltfValidateAccessorCount(wei0, vertCount, "weights0"); err != nil {
			return []rendering.Vertex{}, err
		}
	}
	vertData := make([]rendering.Vertex, vertCount)
	vertColor := matrix.ColorWhite()
	if prim.Material != nil && *prim.Material >= 0 && int(*prim.Material) < len(doc.glTF.Materials) {
		mat := doc.glTF.Materials[*prim.Material]
		if mat.PBRMetallicRoughness.BaseColorFactor != nil {
			vertColor = *mat.PBRMetallicRoughness.BaseColorFactor
		}
	}
	gltfParallelFor(int(vertCount), workers, 8192, func(from, to int) {
		for i := from; i < to; i++ {
			vertData[i].Position = matrix.Vec3{
				pos.float(i, 0),
				pos.float(i, 1),
				pos.float(i, 2),
			}
			vertData[i].Color = vertColor
			vertData[i].MorphTarget = vertData[i].Position
			if hasJnt0 {
				vertData[i].JointIds = matrix.Vec4i{
					jnt0.int32(i, 0),
					jnt0.int32(i, 1),
					jnt0.int32(i, 2),
					jnt0.int32(i, 3),
				}
				vertData[i].JointWeights = matrix.Vec4{
					wei0.float(i, 0),
					wei0.float(i, 1),
					wei0.float(i, 2),
					wei0.float(i, 3),
				}
			} else {
				vertData[i].JointIds = matrix.Vec4i{0, 0, 0, 0}
				vertData[i].JointWeights = matrix.Vec4Zero()
			}
			vertData[i].Normal = matrix.Vec3{
				nml.float(i, 0),
				nml.float(i, 1),
				nml.float(i, 2),
			}
			if hasTan {
				vertData[i].Tangent = matrix.Vec4{
					tan.float(i, 0),
					tan.float(i, 1),
					tan.float(i, 2),
					tan.float(i, 3),
				}
			} else {
				vertData[i].Tangent = matrix.Vec4Zero()
			}
			if hasTex0 {
				vertData[i].UV0 = matrix.Vec2{
					tex0.float(i, 0),
					tex0.float(i, 1),
				}
			} else {
				vertData[i].UV0 = matrix.Vec2Zero()
			}
			if hasCol0 {
				alpha := matrix.Float(1)
				if col0.componentCount >= 4 {
					alpha = col0.normalizedFloat(i, 3)
				}
				vertData[i].Color = matrix.Color{
					col0.normalizedFloat(i, 0),
					col0.normalizedFloat(i, 1),
					col0.normalizedFloat(i, 2),
					alpha,
				}
			}
			for vertData[i].UV0.X() > 1.0 {
				vertData[i].UV0[matrix.Vx] -= 1.0
			}
			for vertData[i].UV0.Y() > 1.0 {
				vertData[i].UV0[matrix.Vy] -= 1.0
			}
		}
	})
	if err = gltfReadMeshMorphTargets(mesh, doc, primitive, vertData, workers); err != nil {
		return []rendering.Vertex{}, err
	}
	return vertData, nil
}

func gltfReadMeshIndices(mesh *gltf.Mesh, doc *fullGLTF, primitive int, workers int) ([]uint32, error) {
	defer tracing.NewRegion("loaders.gltfReadMeshIndices").End()
	idx := mesh.Primitives[primitive].Indices
	acc, err := gltfAccessor(doc, idx)
	if err != nil {
		return []uint32{}, err
	}
	if acc.accessor.Type != gltf.SCALAR {
		return []uint32{}, errors.New("index accessor must be SCALAR")
	}
	if acc.accessor.ComponentType != gltf.UNSIGNED_BYTE &&
		acc.accessor.ComponentType != gltf.UNSIGNED_SHORT &&
		acc.accessor.ComponentType != gltf.UNSIGNED_INT {
		return []uint32{}, errors.New("index accessor must use an unsigned integer component type")
	}
	if !(acc.accessor.Count > 0) {
		return []uint32{}, errors.New("indicesCount > 0")
	}
	convertedIndices := make([]uint32, acc.accessor.Count)
	gltfParallelFor(int(acc.accessor.Count), workers, 16384, func(from, to int) {
		for i := from; i < to; i++ {
			convertedIndices[i] = acc.uint32(i, 0)
		}
	})
	return convertedIndices, nil
}

func gltfReadMeshTextures(mesh *gltf.Mesh, doc *fullGLTF, primitive int) (map[string]string, map[string][]byte) {
	defer tracing.NewRegion("loaders.gltfReadMeshTextures").End()
	textures := make(map[string]string)
	textureBytes := make(map[string][]byte)
	if len(doc.glTF.Materials) == 0 || mesh.Primitives[primitive].Material == nil {
		return textures, textureBytes
	}
	if *mesh.Primitives[primitive].Material < 0 ||
		int(*mesh.Primitives[primitive].Material) >= len(doc.glTF.Materials) {
		return textures, textureBytes
	}
	resolver := func(textureIndex int32, texType string) string {
		if textureIndex < 0 || int(textureIndex) >= len(doc.glTF.Textures) {
			return ""
		}
		srcIdx := doc.glTF.Textures[textureIndex].Source
		key, embedded := gltfTextureKey(doc, srcIdx, texType)
		if embedded {
			textureBytes[key] = doc.textures[srcIdx]
		}
		return key
	}
	addTexture := func(slot string, texture *gltf.TextureId, texType string) {
		if texture == nil {
			return
		}
		key := resolver(texture.Index, texType)
		if key != "" {
			textures[slot] = key
		}
	}
	mat := doc.glTF.Materials[*mesh.Primitives[primitive].Material]
	addTexture("baseColor", mat.PBRMetallicRoughness.BaseColorTexture, "baseColor")
	addTexture("metallicRoughness", mat.PBRMetallicRoughness.MetallicRoughnessTexture, "metallicRoughness")
	addTexture("normal", mat.NormalTexture, "normal")
	addTexture("occlusion", mat.OcclusionTexture, "occlusion")
	addTexture("emissive", mat.EmissiveTexture, "emissive")
	return textures, textureBytes
}

func gltfReadAnimations(doc *fullGLTF, workers int) ([]load_result.Animation, error) {
	defer tracing.NewRegion("loaders.gltfReadAnimations").End()
	anims := make([]load_result.Animation, len(doc.glTF.Animations))
	var firstErr error
	errMutex := sync.Mutex{}
	gltfParallelFor(len(doc.glTF.Animations), workers, 1, func(from, to int) {
		for i := from; i < to; i++ {
			anim, err := gltfReadAnimation(doc, i)
			if err != nil {
				errMutex.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMutex.Unlock()
				continue
			}
			anims[i] = anim
		}
	})
	if firstErr != nil {
		return nil, firstErr
	}
	return anims, nil
}

func gltfReadAnimation(doc *fullGLTF, index int) (load_result.Animation, error) {
	a := &doc.glTF.Animations[index]
	anim := load_result.Animation{
		Name:   a.Name,
		Frames: make([]load_result.AnimKeyFrame, 0),
	}
	for j := range a.Channels {
		c := a.Channels[j]
		if c.Sampler < 0 || int(c.Sampler) >= len(a.Samplers) {
			return anim, fmt.Errorf("animation %d channel %d references invalid sampler %d", index, j, c.Sampler)
		}
		sampler := &a.Samplers[c.Sampler]
		inAcc, err := gltfAccessor(doc, sampler.Input)
		if err != nil {
			return anim, err
		}
		outAcc, err := gltfAccessor(doc, sampler.Output)
		if err != nil {
			return anim, err
		}
		if err = gltfValidateAccessor(inAcc, gltf.FLOAT, gltf.SCALAR, "animation input"); err != nil {
			return anim, err
		}
		boneTemplate := load_result.AnimBone{
			PathType:      c.Target.Path(),
			Interpolation: sampler.Interpolation(),
			NodeIndex:     int(c.Target.Node),
		}
		switch boneTemplate.PathType {
		case load_result.AnimPathTranslation, load_result.AnimPathScale:
			if err = gltfValidateAccessor(outAcc, gltf.FLOAT, gltf.VEC3, "animation output"); err != nil {
				return anim, err
			}
		case load_result.AnimPathRotation:
			if err = gltfValidateAccessor(outAcc, gltf.FLOAT, gltf.VEC4, "animation output"); err != nil {
				return anim, err
			}
		case load_result.AnimPathWeights:
			// TODO:  Implement reading weights data
			continue
		default:
			continue
		}
		if outAcc.accessor.Count < inAcc.accessor.Count {
			return anim, errors.New("animation output count is smaller than input count")
		}
		for k := 0; k < int(inAcc.accessor.Count); k++ {
			time := inAcc.float(k, 0)
			key := gltfAnimationFrame(&anim, time)
			bone := boneTemplate
			switch bone.PathType {
			case load_result.AnimPathTranslation:
				bone.Data = matrix.Vec3{
					outAcc.float(k, 0),
					outAcc.float(k, 1),
					outAcc.float(k, 2),
				}.AsAligned16()
			case load_result.AnimPathRotation:
				// glTF stores quaternions as XYZW.
				bone.Data = matrix.QuaternionFromXYZW([4]matrix.Float{
					outAcc.float(k, 0),
					outAcc.float(k, 1),
					outAcc.float(k, 2),
					outAcc.float(k, 3),
				})
			case load_result.AnimPathScale:
				bone.Data = matrix.Vec3{
					outAcc.float(k, 0),
					outAcc.float(k, 1),
					outAcc.float(k, 2),
				}.AsAligned16()
			}
			key.Bones = append(key.Bones, bone)
		}
	}
	if len(anim.Frames) == 0 {
		return anim, nil
	}
	slices.SortFunc(anim.Frames, func(a, b load_result.AnimKeyFrame) int {
		return int((a.Time - b.Time) * 10000)
	})
	for j := range anim.Frames[:len(anim.Frames)-1] {
		anim.Frames[j].Time = anim.Frames[j+1].Time - anim.Frames[j].Time
	}
	anim.Frames[len(anim.Frames)-1].Time = 0.0
	return anim, nil
}

func gltfAnimationFrame(anim *load_result.Animation, time matrix.Float) *load_result.AnimKeyFrame {
	for i := range anim.Frames {
		if matrix.Approx(anim.Frames[i].Time, float32(time)) {
			return &anim.Frames[i]
		}
	}
	anim.Frames = append(anim.Frames, load_result.AnimKeyFrame{
		Bones: make([]load_result.AnimBone, 0),
		Time:  float32(time),
	})
	return &anim.Frames[len(anim.Frames)-1]
}

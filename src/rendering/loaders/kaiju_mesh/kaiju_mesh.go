/******************************************************************************/
/* kaiju_mesh.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package kaiju_mesh

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"runtime"
	"slices"
	"strings"
	"sync"

	"kaijuengine.com/debug"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/load_result"
)

type AnimationPathType = int
type AnimationInterpolation = int

const (
	AnimPathInvalid AnimationPathType = iota - 1
	AnimPathTranslation
	AnimPathRotation
	AnimPathScale
	AnimPathWeights
)

const (
	AnimInterpolateInvalid AnimationInterpolation = iota - 1
	AnimInterpolateLinear
	AnimInterpolateStep
	AnimInterpolateCubicSpline
)

// KaijuMesh is a base primitive representing a single mesh. This is the
// archived format of the authored meshes. Typically this structure is created
// by loading in a mesh using something like [loaders.GLTF] and then converting
// the result of that using [LoadedResultToKaijuMesh]. From this point, it is
// typically serialized and stored into the content database. When reading a
// mesh from the content database, it will return a KaijuMesh.
type KaijuMesh struct {
	Key        string
	Name       string
	Node       KaijuMeshNode
	Material   string
	Verts      []rendering.Vertex
	Indexes    []uint32
	Textures   map[string]string
	BVH        *graviton.TriangleBVH
	Animations []KaijuMeshAnimation
	Joints     []KaijuMeshJoint
}

type KaijuMeshNode struct {
	Name     string
	Position matrix.Vec3
	Rotation matrix.Vec3
	Scale    matrix.Vec3
}

type KaijuMeshSet struct {
	Name   string
	Meshes []KaijuMesh
}

type SerializeOptions struct {
	TextureURIs     map[string]string
	MeshTextureURIs map[string]map[string]string
}

type MeshRef struct {
	Asset string
	Key   string
}

// LoadedResultToKaijuMesh will take in a [load_result.Result] and convert every
// mesh contained within the structure to our built in version known as
// [KaijuMesh]. This is typically used for the editor, but games/applications
// may find some use for it.
func LoadedResultToKaijuMesh(res load_result.Result) []KaijuMesh {
	defer tracing.NewRegion("kaiju_mesh.LoadedResultToKaijuMesh").End()
	out := make([]KaijuMesh, len(res.Meshes))
	keys := meshKeysForResult(res)
	build := func(meshIndex int) {
		m := &res.Meshes[meshIndex]
		out[meshIndex] = KaijuMesh{
			Key:        keys[meshIndex],
			Name:       m.MeshName,
			Node:       meshNodeFromLoadResult(m.Node),
			Verts:      slices.Clone(m.Verts),
			Indexes:    slices.Clone(m.Indexes),
			Textures:   cloneStringMap(m.Textures),
			Animations: make([]KaijuMeshAnimation, len(res.Animations)),
			Joints:     make([]KaijuMeshJoint, len(res.Joints)),
		}
		for jointIndex := range res.Joints {
			out[meshIndex].Joints[jointIndex].fromLoadResult(&res, &res.Joints[jointIndex])
		}
		for animIndex := range res.Animations {
			out[meshIndex].Animations[animIndex].fromLoadResult(&res.Animations[animIndex])
		}
	}
	workers := min(runtime.GOMAXPROCS(0), len(res.Meshes))
	if workers <= 1 {
		for i := range res.Meshes {
			build(i)
		}
	} else {
		jobs := make(chan int)
		group := sync.WaitGroup{}
		group.Add(workers)
		for range workers {
			go func() {
				defer group.Done()
				for idx := range jobs {
					build(idx)
				}
			}()
		}
		for i := range res.Meshes {
			jobs <- i
		}
		close(jobs)
		group.Wait()
	}
	debug.Ensure(len(out) == len(res.Meshes))
	return out
}

func LoadedResultToKaijuMeshSet(name string, res load_result.Result) KaijuMeshSet {
	defer tracing.NewRegion("kaiju_mesh.LoadedResultToKaijuMeshSet").End()
	return KaijuMeshSet{
		Name:   name,
		Meshes: LoadedResultToKaijuMesh(res),
	}
}

// Serialize will convert a [KaijuMesh] into a byte array for saving to the
// database or later use.
func (k KaijuMesh) Serialize() ([]byte, error) {
	defer tracing.NewRegion("KaijuMesh.Serialize").End()
	return k.SerializeWithOptions(SerializeOptions{})
}

func (k KaijuMesh) SerializeWithOptions(options SerializeOptions) ([]byte, error) {
	defer tracing.NewRegion("KaijuMesh.SerializeWithOptions").End()
	return serializeGLB(k, options)
}

func (s KaijuMeshSet) Serialize() ([]byte, error) {
	defer tracing.NewRegion("KaijuMeshSet.Serialize").End()
	return s.SerializeWithOptions(SerializeOptions{})
}

func (s KaijuMeshSet) SerializeWithOptions(options SerializeOptions) ([]byte, error) {
	defer tracing.NewRegion("KaijuMeshSet.SerializeWithOptions").End()
	if len(s.Meshes) == 1 {
		return s.Meshes[0].SerializeWithOptions(options)
	}
	return serializeGLBSet(s, options)
}

// Deserialize will construct a [KaijuMesh] from the given array of bytes. This
// supports GLB mesh content and falls back to native/gob for legacy assets.
func Deserialize(data []byte) (KaijuMesh, error) {
	defer tracing.NewRegion("kaiju_mesh.Deserialize").End()
	if IsGLB(data) {
		set, err := deserializeGLBSet(data)
		if err != nil {
			return KaijuMesh{}, err
		}
		if len(set.Meshes) == 0 {
			return KaijuMesh{}, errors.New("mesh set contains no meshes")
		}
		return set.Meshes[0], nil
	}
	if isNativeMesh(data) {
		return deserializeNative(data)
	}
	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)
	var km KaijuMesh
	err := dec.Decode(&km)
	return km, err
}

func DeserializeSet(data []byte) (KaijuMeshSet, error) {
	defer tracing.NewRegion("kaiju_mesh.DeserializeSet").End()
	if IsGLB(data) {
		return deserializeGLBSet(data)
	}
	km, err := Deserialize(data)
	if err != nil {
		return KaijuMeshSet{}, err
	}
	if km.Key == "" {
		km.Key = "mesh"
	}
	return KaijuMeshSet{Meshes: []KaijuMesh{km}}, nil
}

func (s KaijuMeshSet) MeshByKey(key string) (KaijuMesh, bool) {
	if len(s.Meshes) == 0 {
		return KaijuMesh{}, false
	}
	if key == "" {
		return s.Meshes[0], true
	}
	for i := range s.Meshes {
		if s.Meshes[i].Key == key {
			return s.Meshes[i], true
		}
	}
	return KaijuMesh{}, false
}

func (s KaijuMeshSet) EnsureBVH() {
	for i := range s.Meshes {
		s.Meshes[i].EnsureBVH()
	}
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func meshKeysForResult(res load_result.Result) []string {
	out := make([]string, len(res.Meshes))
	used := make(map[string]int, len(res.Meshes))
	for i := range res.Meshes {
		name := res.Meshes[i].Name
		if res.Meshes[i].Node != nil && res.Meshes[i].Node.Name != "" {
			name = res.Meshes[i].Node.Name
		}
		if name == "" {
			name = res.Meshes[i].MeshName
		}
		out[i] = StableMeshKey(name, i, used)
	}
	return out
}

func StableMeshKey(name string, index int, used map[string]int) string {
	return stableMeshKey(name, index, used)
}

func stableMeshKey(name string, index int, used map[string]int) string {
	name = strings.ToLower(strings.TrimSpace(name))
	sb := strings.Builder{}
	lastUnderscore := false
	for _, r := range name {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			sb.WriteRune(r)
			lastUnderscore = false
		} else if !lastUnderscore && sb.Len() > 0 {
			sb.WriteByte('_')
			lastUnderscore = true
		}
	}
	key := strings.Trim(sb.String(), "_")
	if key == "" {
		key = fmt.Sprintf("mesh_%d", index)
	}
	if count := used[key]; count > 0 {
		used[key] = count + 1
		key = fmt.Sprintf("%s_%d", key, count+1)
	} else {
		used[key] = 1
	}
	return key
}

func meshNodeFromLoadResult(node *load_result.Node) KaijuMeshNode {
	out := KaijuMeshNode{
		Scale: matrix.Vec3One(),
	}
	if node == nil {
		return out
	}
	out.Name = node.Name
	out.Position = node.Position
	out.Rotation = node.Rotation.ToEuler()
	out.Scale = node.Scale
	if out.Scale.IsZero() {
		out.Scale = matrix.Vec3One()
	}
	return out
}

func ReadMesh(ref string, host *engine.Host) (KaijuMesh, error) {
	defer tracing.NewRegion("kaiju_mesh.ReadMesh").End()
	meshRef := ParseMeshRef(ref)
	data, err := host.AssetDatabase().Read(meshRef.Asset)
	if err != nil {
		slog.Error("failed to read the mesh", "id", meshRef.Asset, "error", err)
		return KaijuMesh{}, err
	}
	set, err := DeserializeSet(data)
	if err != nil {
		return KaijuMesh{}, err
	}
	if mesh, ok := set.MeshByKey(meshRef.Key); ok {
		return mesh, nil
	}
	return KaijuMesh{}, fmt.Errorf("mesh %q not found in %q", meshRef.Key, meshRef.Asset)
}

func ParseMeshRef(ref string) MeshRef {
	asset, key, ok := strings.Cut(ref, "#mesh=")
	if !ok {
		return MeshRef{Asset: ref}
	}
	if decoded, err := url.QueryUnescape(key); err == nil {
		key = decoded
	}
	return MeshRef{Asset: asset, Key: key}
}

func MeshRefString(asset, key string) string {
	if key == "" {
		return asset
	}
	return asset + "#mesh=" + url.QueryEscape(key)
}

func (r MeshRef) String() string { return MeshRefString(r.Asset, r.Key) }

func (r MeshRef) IsSubmesh() bool { return r.Key != "" }

func (k *KaijuMesh) EnsureBVH() {
	defer tracing.NewRegion("KaijuMesh.EnsureBVH").End()
	if k.BVH == nil {
		k.BVH = k.GenerateBVHArchive()
	}
}

func (k *KaijuMesh) GenerateBVHArchive() *graviton.TriangleBVH {
	defer tracing.NewRegion("KaijuMesh.GenerateBVHArchive").End()
	return graviton.NewTriangleBVH(k.generateBVH(nil, nil, nil))
}

func (k *KaijuMesh) GenerateBVH(threads *concurrent.Threads, transform *matrix.Transform, data any) *graviton.BVH {
	defer tracing.NewRegion("KaijuMesh.GenerateBVH").End()
	if k.BVH == nil {
		k.BVH = k.GenerateBVHArchive()
		if k.BVH == nil {
			return nil
		}
	}
	bvh := k.BVH.ToBVH(transform, data)
	bvh.Refit()
	return bvh
}

func (k KaijuMesh) generateBVH(threads *concurrent.Threads, transform *matrix.Transform, data any) *graviton.BVH {
	defer tracing.NewRegion("KaijuMesh.generateBVH").End()
	tris := k.bvhTriangles(threads)
	return graviton.NewBVH(tris, transform, data)
}

func (k KaijuMesh) bvhTriangles(threads *concurrent.Threads) []graviton.HitObject {
	defer tracing.NewRegion("KaijuMesh.bvhTriangles").End()
	tris := make([]graviton.HitObject, len(k.Indexes)/3)
	if len(tris) == 0 {
		return tris
	}
	construct := func(from, to int) {
		for tri := from; tri < to; tri++ {
			i := tri * 3
			points := [3]matrix.Vec3{
				k.Verts[k.Indexes[i]].Position,
				k.Verts[k.Indexes[i+1]].Position,
				k.Verts[k.Indexes[i+2]].Position,
			}
			tris[tri] = graviton.DetailedTriangleFromPoints(points)
		}
	}
	if threads == nil || threads.ThreadCount() == 0 || len(tris) == 1 {
		construct(0, len(tris))
		return tris
	}
	group := sync.WaitGroup{}
	workCount := min(threads.ThreadCount(), len(tris))
	work := make([]func(int), workCount)
	group.Add(workCount)
	for i := range work {
		from := i * len(tris) / workCount
		to := (i + 1) * len(tris) / workCount
		work[i] = func(int) {
			construct(from, to)
			group.Done()
		}
	}
	threads.AddWork(work)
	group.Wait()
	return tris
}

/******************************************************************************/
/* kaiju_mesh.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package kaiju_mesh

import (
	"bytes"
	"encoding/gob"
	"log/slog"
	"runtime"
	"slices"
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
	Name       string
	Verts      []rendering.Vertex
	Indexes    []uint32
	BVH        *graviton.TriangleBVH
	Animations []KaijuMeshAnimation
	Joints     []KaijuMeshJoint
}

// LoadedResultToKaijuMesh will take in a [load_result.Result] and convert every
// mesh contained within the structure to our built in version known as
// [KaijuMesh]. This is typically used for the editor, but games/applications
// may find some use for it.
func LoadedResultToKaijuMesh(res load_result.Result) []KaijuMesh {
	defer tracing.NewRegion("kaiju_mesh.LoadedResultToKaijuMesh").End()
	out := make([]KaijuMesh, len(res.Meshes))
	build := func(meshIndex int) {
		m := &res.Meshes[meshIndex]
		out[meshIndex] = KaijuMesh{
			Name:       m.MeshName,
			Verts:      slices.Clone(m.Verts),
			Indexes:    slices.Clone(m.Indexes),
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

// Serialize will convert a [KaijuMesh] into a byte array for saving to the
// database or later use.
func (k KaijuMesh) Serialize() ([]byte, error) {
	defer tracing.NewRegion("KaijuMesh.Serialize").End()
	return serializeNative(k)
}

// Deserialize will construct a [KaijuMesh] from the given array of bytes. This
// supports the current native mesh format and falls back to gob for legacy
// assets.
func Deserialize(data []byte) (KaijuMesh, error) {
	defer tracing.NewRegion("kaiju_mesh.Deserialize").End()
	if isNativeMesh(data) {
		return deserializeNative(data)
	}
	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)
	var km KaijuMesh
	err := dec.Decode(&km)
	return km, err
}

func ReadMesh(id string, host *engine.Host) (KaijuMesh, error) {
	defer tracing.NewRegion("kaiju_mesh.ReadMesh").End()
	data, err := host.AssetDatabase().Read(id)
	if err != nil {
		slog.Error("failed to read the mesh", "id", id, "error", err)
		return KaijuMesh{}, err
	}
	return Deserialize(data)
}

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

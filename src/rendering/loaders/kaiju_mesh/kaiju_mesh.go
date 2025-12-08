/******************************************************************************/
/* kaiju_mesh.go                                                              */
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

package kaiju_mesh

import (
	"bytes"
	"encoding/gob"
	"slices"
	"sync"

	"github.com/KaijuEngine/kaiju/debug"
	"github.com/KaijuEngine/kaiju/engine/collision"
	"github.com/KaijuEngine/kaiju/matrix"
	"github.com/KaijuEngine/kaiju/platform/concurrent"
	"github.com/KaijuEngine/kaiju/rendering"
	"github.com/KaijuEngine/kaiju/rendering/loaders/load_result"
)

// KaijuMesh is a base primitive representing a single mesh. This is the
// archived format of the authored meshes. Typically this structure is created
// by loading in a mesh using something like [loaders.GLTF] and then converting
// the result of that using [LoadedResultToKaijuMesh]. From this point, it is
// typically serialized and stored into the content database. When reading a
// mesh from the content database, it will return a KaijuMesh.
type KaijuMesh struct {
	Name    string
	Verts   []rendering.Vertex
	Indexes []uint32
}

// LoadedResultToKaijuMesh will take in a [load_result.Result] and convert every
// mesh contained within the structure to our built in version known as
// [KaijuMesh]. This is typically used for the editor, but games/applications
// may find some use for it.
func LoadedResultToKaijuMesh(res load_result.Result) []KaijuMesh {
	out := make([]KaijuMesh, 0, len(res.Meshes))
	for i := range res.Meshes {
		m := &res.Meshes[i]
		out = append(out, KaijuMesh{
			Name:    m.MeshName,
			Verts:   slices.Clone(m.Verts),
			Indexes: slices.Clone(m.Indexes),
		})
	}
	debug.Ensure(len(out) == len(res.Meshes))
	return out
}

// Serialize will convert a [KaijuMesh] into a byte array for saving to the
// database or later use. This serialization uses the built-in [gob.Encoder]
func (k KaijuMesh) Serialize() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(w)
	err := enc.Encode(k)
	return w.Bytes(), err
}

// Deserialize will construct a [KaijuMesh] from the given array of bytes. This
// deserialization uses the built-in [gob.Decoder]
func Deserialize(data []byte) (KaijuMesh, error) {
	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)
	var km KaijuMesh
	err := dec.Decode(&km)
	return km, err
}

func (k KaijuMesh) GenerateBVH(threads *concurrent.Threads, transform *matrix.Transform, data any) *collision.BVH {
	tris := make([]collision.HitObject, len(k.Indexes)/3)
	group := sync.WaitGroup{}
	construct := func(from, to int) {
		for i := from; i < to; i += 3 {
			for i := 0; i < len(k.Indexes); i += 3 {
				points := [3]matrix.Vec3{
					k.Verts[k.Indexes[i]].Position,
					k.Verts[k.Indexes[i+1]].Position,
					k.Verts[k.Indexes[i+2]].Position,
				}
				tris[i/3] = collision.DetailedTriangleFromPoints(points)
			}
		}
		group.Done()
	}
	work := make([]func(int), len(tris))
	group.Add(len(work))
	for i := range work {
		work[i] = func(int) { construct(i*3, (i+3)*3) }
	}
	threads.AddWork(work)
	group.Wait()
	return collision.NewBVH(tris, transform, data)
}

package kaiju_mesh

import (
	"bytes"
	"encoding/gob"
	"kaiju/debug"
	"kaiju/rendering"
	"kaiju/rendering/loaders/load_result"
	"slices"
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

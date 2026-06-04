/******************************************************************************/
/* kaiju_mesh_bvh_blob.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package kaiju_mesh

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

var triangleBVHBlobMagic = [8]byte{'K', 'J', 'T', 'B', 'V', 'H', 1, 0}

type triangleBVHBlobWriter struct {
	data []byte
}

type triangleBVHBlobReader struct {
	data []byte
	pos  int
}

func serializeTriangleBVHBlob(bvh *graviton.TriangleBVH) []byte {
	w := triangleBVHBlobWriter{data: make([]byte, 0)}
	w.data = append(w.data, triangleBVHBlobMagic[:]...)
	w.triangleBVH(bvh)
	return w.data
}

func deserializeTriangleBVHBlob(data []byte) (*graviton.TriangleBVH, error) {
	r := triangleBVHBlobReader{data: data}
	magic, err := r.bytes(len(triangleBVHBlobMagic))
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(magic, triangleBVHBlobMagic[:]) {
		return nil, errors.New("invalid triangle BVH blob")
	}
	bvh, err := r.triangleBVH()
	if err != nil {
		return nil, err
	}
	if r.pos != len(r.data) {
		return nil, fmt.Errorf("triangle BVH blob has %d unread bytes", len(r.data)-r.pos)
	}
	return bvh, nil
}

func (w *triangleBVHBlobWriter) triangleBVH(bvh *graviton.TriangleBVH) {
	w.bool(bvh != nil)
	if bvh == nil {
		return
	}
	w.vec3(bvh.Bounds.Center)
	w.vec3(bvh.Bounds.Extent)
	w.bool(bvh.HasTriangle)
	if bvh.HasTriangle {
		for i := range bvh.Triangle.Points {
			w.vec3(bvh.Triangle.Points[i])
		}
	}
	w.triangleBVH(bvh.Left)
	w.triangleBVH(bvh.Right)
}

func (w *triangleBVHBlobWriter) bool(v bool) {
	if v {
		w.data = append(w.data, 1)
	} else {
		w.data = append(w.data, 0)
	}
}

func (w *triangleBVHBlobWriter) f32(v matrix.Float) {
	w.data = binary.LittleEndian.AppendUint32(w.data, math.Float32bits(float32(v)))
}

func (w *triangleBVHBlobWriter) vec3(v matrix.Vec3) {
	w.f32(v.X())
	w.f32(v.Y())
	w.f32(v.Z())
}

func (r *triangleBVHBlobReader) triangleBVH() (*graviton.TriangleBVH, error) {
	hasBVH, err := r.bool()
	if err != nil {
		return nil, err
	}
	if !hasBVH {
		return nil, nil
	}
	center, err := r.vec3()
	if err != nil {
		return nil, err
	}
	extent, err := r.vec3()
	if err != nil {
		return nil, err
	}
	hasTriangle, err := r.bool()
	if err != nil {
		return nil, err
	}
	out := &graviton.TriangleBVH{
		Bounds:      graviton.NewAABB(center, extent),
		HasTriangle: hasTriangle,
	}
	if out.HasTriangle {
		points := [3]matrix.Vec3{}
		for i := range points {
			if points[i], err = r.vec3(); err != nil {
				return nil, err
			}
		}
		out.Triangle = graviton.DetailedTriangleFromPoints(points)
	}
	if out.Left, err = r.triangleBVH(); err != nil {
		return nil, err
	}
	if out.Right, err = r.triangleBVH(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *triangleBVHBlobReader) bytes(count int) ([]byte, error) {
	if count < 0 {
		return nil, errors.New("negative byte count")
	}
	if len(r.data)-r.pos < count {
		return nil, fmt.Errorf("triangle BVH blob ended early at byte %d", r.pos)
	}
	out := r.data[r.pos : r.pos+count]
	r.pos += count
	return out, nil
}

func (r *triangleBVHBlobReader) bool() (bool, error) {
	b, err := r.bytes(1)
	if err != nil {
		return false, err
	}
	switch b[0] {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("invalid triangle BVH bool %d", b[0])
	}
}

func (r *triangleBVHBlobReader) f32() (matrix.Float, error) {
	b, err := r.bytes(4)
	if err != nil {
		return 0, err
	}
	return matrix.Float(math.Float32frombits(binary.LittleEndian.Uint32(b))), nil
}

func (r *triangleBVHBlobReader) vec3() (matrix.Vec3, error) {
	x, err := r.f32()
	if err != nil {
		return matrix.Vec3{}, err
	}
	y, err := r.f32()
	if err != nil {
		return matrix.Vec3{}, err
	}
	z, err := r.f32()
	if err != nil {
		return matrix.Vec3{}, err
	}
	return matrix.Vec3{x, y, z}, nil
}

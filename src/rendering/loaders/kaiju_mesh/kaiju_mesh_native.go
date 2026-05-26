/******************************************************************************/
/* kaiju_mesh_native.go                                                       */
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
	"unsafe"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const nativeMeshVersion = uint16(1)

var nativeMeshMagic = [8]byte{'K', 'A', 'I', 'J', 'U', 'M', 'S', 'H'}

type nativeMeshWriter struct {
	data []byte
	err  error
}

type nativeMeshReader struct {
	data []byte
	pos  int
}

func isNativeMesh(data []byte) bool {
	return len(data) >= len(nativeMeshMagic) && bytes.Equal(data[:len(nativeMeshMagic)], nativeMeshMagic[:])
}

func serializeNative(k KaijuMesh) ([]byte, error) {
	if !isLittleEndian() {
		return nil, errors.New("kaiju mesh native serialization requires a little-endian platform")
	}
	w := nativeMeshWriter{data: make([]byte, 0, nativeMeshSizeHint(k))}
	w.bytes(nativeMeshMagic[:])
	w.u16(nativeMeshVersion)
	w.u8(uint8(unsafe.Sizeof(matrix.Float(0))))
	w.u32(uint32(unsafe.Sizeof(rendering.Vertex{})))
	w.string(k.Name)
	writeRawSlice(&w, k.Verts)
	writeRawSlice(&w, k.Indexes)
	w.triangleBVH(k.BVH)
	w.meshJoints(k.Joints)
	w.meshAnimations(k.Animations)
	return w.data, w.err
}

func deserializeNative(data []byte) (KaijuMesh, error) {
	if !isLittleEndian() {
		return KaijuMesh{}, errors.New("kaiju mesh native deserialization requires a little-endian platform")
	}
	r := nativeMeshReader{data: data}
	if _, err := r.bytes(len(nativeMeshMagic)); err != nil {
		return KaijuMesh{}, err
	}
	version, err := r.u16()
	if err != nil {
		return KaijuMesh{}, err
	}
	if version != nativeMeshVersion {
		return KaijuMesh{}, fmt.Errorf("unsupported kaiju mesh native version %d", version)
	}
	floatSize, err := r.u8()
	if err != nil {
		return KaijuMesh{}, err
	}
	if floatSize != uint8(unsafe.Sizeof(matrix.Float(0))) {
		return KaijuMesh{}, fmt.Errorf("kaiju mesh float size mismatch: file=%d runtime=%d", floatSize, unsafe.Sizeof(matrix.Float(0)))
	}
	vertexSize, err := r.u32()
	if err != nil {
		return KaijuMesh{}, err
	}
	if vertexSize != uint32(unsafe.Sizeof(rendering.Vertex{})) {
		return KaijuMesh{}, fmt.Errorf("kaiju mesh vertex size mismatch: file=%d runtime=%d", vertexSize, unsafe.Sizeof(rendering.Vertex{}))
	}
	name, err := r.string()
	if err != nil {
		return KaijuMesh{}, err
	}
	verts, err := readRawSlice[rendering.Vertex](&r)
	if err != nil {
		return KaijuMesh{}, err
	}
	indexes, err := readRawSlice[uint32](&r)
	if err != nil {
		return KaijuMesh{}, err
	}
	bvh, err := r.triangleBVH()
	if err != nil {
		return KaijuMesh{}, err
	}
	joints, err := r.meshJoints()
	if err != nil {
		return KaijuMesh{}, err
	}
	animations, err := r.meshAnimations()
	if err != nil {
		return KaijuMesh{}, err
	}
	if r.pos != len(r.data) {
		return KaijuMesh{}, fmt.Errorf("kaiju mesh native data has %d unread bytes", len(r.data)-r.pos)
	}
	return KaijuMesh{
		Name:       name,
		Verts:      verts,
		Indexes:    indexes,
		BVH:        bvh,
		Animations: animations,
		Joints:     joints,
	}, nil
}

func nativeMeshSizeHint(k KaijuMesh) int {
	size := len(nativeMeshMagic) + 64 + len(k.Name)
	if len(k.Verts) > 0 {
		size += len(k.Verts) * int(unsafe.Sizeof(k.Verts[0]))
	}
	if len(k.Indexes) > 0 {
		size += len(k.Indexes) * int(unsafe.Sizeof(k.Indexes[0]))
	}
	size += len(k.Joints) * (2*4 + 25*int(unsafe.Sizeof(matrix.Float(0))))
	return size
}

func isLittleEndian() bool {
	var x uint16 = 1
	return *(*byte)(unsafe.Pointer(&x)) == 1
}

func (w *nativeMeshWriter) bytes(data []byte) {
	if w.err != nil {
		return
	}
	w.data = append(w.data, data...)
}

func (w *nativeMeshWriter) u8(v uint8) {
	if w.err != nil {
		return
	}
	w.data = append(w.data, v)
}

func (w *nativeMeshWriter) bool(v bool) {
	if v {
		w.u8(1)
	} else {
		w.u8(0)
	}
}

func (w *nativeMeshWriter) u16(v uint16) {
	if w.err != nil {
		return
	}
	w.data = binary.LittleEndian.AppendUint16(w.data, v)
}

func (w *nativeMeshWriter) u32(v uint32) {
	if w.err != nil {
		return
	}
	w.data = binary.LittleEndian.AppendUint32(w.data, v)
}

func (w *nativeMeshWriter) i32(v int32) {
	w.u32(uint32(v))
}

func (w *nativeMeshWriter) f32(v float32) {
	w.u32(math.Float32bits(v))
}

func (w *nativeMeshWriter) string(v string) {
	if uint64(len(v)) > uint64(^uint32(0)) {
		w.err = errors.New("string too long to serialize")
		return
	}
	w.u32(uint32(len(v)))
	w.bytes(unsafe.Slice(unsafe.StringData(v), len(v)))
}

func writeRaw[T any](w *nativeMeshWriter, v *T) {
	if w.err != nil {
		return
	}
	w.data = append(w.data, unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))...)
}

func writeRawSlice[T any](w *nativeMeshWriter, values []T) {
	if uint64(len(values)) > uint64(^uint32(0)) {
		w.err = errors.New("slice too long to serialize")
		return
	}
	w.u32(uint32(len(values)))
	if w.err != nil || len(values) == 0 {
		return
	}
	size := len(values) * int(unsafe.Sizeof(values[0]))
	w.bytes(unsafe.Slice((*byte)(unsafe.Pointer(&values[0])), size))
}

func (w *nativeMeshWriter) triangleBVH(bvh *graviton.TriangleBVH) {
	w.bool(bvh != nil)
	if w.err != nil || bvh == nil {
		return
	}
	writeRaw(w, &bvh.Bounds)
	w.bool(bvh.HasTriangle)
	if bvh.HasTriangle {
		writeRaw(w, &bvh.Triangle)
	}
	w.triangleBVH(bvh.Left)
	w.triangleBVH(bvh.Right)
}

func (w *nativeMeshWriter) meshJoints(joints []KaijuMeshJoint) {
	if uint64(len(joints)) > uint64(^uint32(0)) {
		w.err = errors.New("too many joints to serialize")
		return
	}
	w.u32(uint32(len(joints)))
	for i := range joints {
		w.i32(joints[i].Id)
		w.i32(joints[i].Parent)
		writeRaw(w, &joints[i].Skin)
		writeRaw(w, &joints[i].Position)
		writeRaw(w, &joints[i].Rotation)
		writeRaw(w, &joints[i].Scale)
	}
}

func (w *nativeMeshWriter) meshAnimations(animations []KaijuMeshAnimation) {
	if uint64(len(animations)) > uint64(^uint32(0)) {
		w.err = errors.New("too many animations to serialize")
		return
	}
	w.u32(uint32(len(animations)))
	for i := range animations {
		w.string(animations[i].Name)
		if uint64(len(animations[i].Frames)) > uint64(^uint32(0)) {
			w.err = errors.New("too many animation frames to serialize")
			return
		}
		w.u32(uint32(len(animations[i].Frames)))
		for f := range animations[i].Frames {
			frame := &animations[i].Frames[f]
			w.f32(frame.Time)
			if uint64(len(frame.Bones)) > uint64(^uint32(0)) {
				w.err = errors.New("too many animation bones to serialize")
				return
			}
			w.u32(uint32(len(frame.Bones)))
			for b := range frame.Bones {
				bone := &frame.Bones[b]
				w.i32(int32(bone.NodeIndex))
				w.i32(int32(bone.PathType))
				w.i32(int32(bone.Interpolation))
				writeRaw(w, &bone.Data)
			}
		}
	}
}

func (r *nativeMeshReader) bytes(count int) ([]byte, error) {
	if count < 0 {
		return nil, errors.New("negative byte count")
	}
	if len(r.data)-r.pos < count {
		return nil, fmt.Errorf("kaiju mesh native data ended early at byte %d", r.pos)
	}
	out := r.data[r.pos : r.pos+count]
	r.pos += count
	return out, nil
}

func (r *nativeMeshReader) u8() (uint8, error) {
	b, err := r.bytes(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *nativeMeshReader) bool() (bool, error) {
	v, err := r.u8()
	if err != nil {
		return false, err
	}
	switch v {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("invalid boolean value %d", v)
	}
}

func (r *nativeMeshReader) u16() (uint16, error) {
	b, err := r.bytes(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

func (r *nativeMeshReader) u32() (uint32, error) {
	b, err := r.bytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *nativeMeshReader) i32() (int32, error) {
	v, err := r.u32()
	return int32(v), err
}

func (r *nativeMeshReader) f32() (float32, error) {
	v, err := r.u32()
	return math.Float32frombits(v), err
}

func (r *nativeMeshReader) string() (string, error) {
	count, err := r.u32()
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "", nil
	}
	if uint64(count) > uint64(maxInt()) {
		return "", errors.New("string is too large for this platform")
	}
	b, err := r.bytes(int(count))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readRaw[T any](r *nativeMeshReader, v *T) error {
	size := int(unsafe.Sizeof(*v))
	b, err := r.bytes(size)
	if err != nil {
		return err
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(v)), size), b)
	return nil
}

func readRawSlice[T any](r *nativeMeshReader) ([]T, error) {
	count, err := r.u32()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return []T{}, nil
	}
	size := int(unsafe.Sizeof(*new(T)))
	if uint64(count) > uint64(maxInt()/size) {
		return nil, errors.New("slice is too large for this platform")
	}
	values := make([]T, int(count))
	bytesCount := int(count) * size
	b, err := r.bytes(bytesCount)
	if err != nil {
		return nil, err
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&values[0])), bytesCount), b)
	return values, nil
}

func (r *nativeMeshReader) triangleBVH() (*graviton.TriangleBVH, error) {
	hasBVH, err := r.bool()
	if err != nil {
		return nil, err
	}
	if !hasBVH {
		return nil, nil
	}
	bvh := &graviton.TriangleBVH{}
	if err := readRaw(r, &bvh.Bounds); err != nil {
		return nil, err
	}
	hasTriangle, err := r.bool()
	if err != nil {
		return nil, err
	}
	bvh.HasTriangle = hasTriangle
	if bvh.HasTriangle {
		if err := readRaw(r, &bvh.Triangle); err != nil {
			return nil, err
		}
	}
	if bvh.Left, err = r.triangleBVH(); err != nil {
		return nil, err
	}
	if bvh.Right, err = r.triangleBVH(); err != nil {
		return nil, err
	}
	return bvh, nil
}

func (r *nativeMeshReader) meshJoints() ([]KaijuMeshJoint, error) {
	count, err := r.u32()
	if err != nil {
		return nil, err
	}
	if uint64(count) > uint64(maxInt()) {
		return nil, errors.New("joint count is too large for this platform")
	}
	joints := make([]KaijuMeshJoint, int(count))
	for i := range joints {
		if joints[i].Id, err = r.i32(); err != nil {
			return nil, err
		}
		if joints[i].Parent, err = r.i32(); err != nil {
			return nil, err
		}
		if err = readRaw(r, &joints[i].Skin); err != nil {
			return nil, err
		}
		if err = readRaw(r, &joints[i].Position); err != nil {
			return nil, err
		}
		if err = readRaw(r, &joints[i].Rotation); err != nil {
			return nil, err
		}
		if err = readRaw(r, &joints[i].Scale); err != nil {
			return nil, err
		}
	}
	return joints, nil
}

func (r *nativeMeshReader) meshAnimations() ([]KaijuMeshAnimation, error) {
	count, err := r.u32()
	if err != nil {
		return nil, err
	}
	if uint64(count) > uint64(maxInt()) {
		return nil, errors.New("animation count is too large for this platform")
	}
	animations := make([]KaijuMeshAnimation, int(count))
	for i := range animations {
		if animations[i].Name, err = r.string(); err != nil {
			return nil, err
		}
		frameCount, err := r.u32()
		if err != nil {
			return nil, err
		}
		if uint64(frameCount) > uint64(maxInt()) {
			return nil, errors.New("animation frame count is too large for this platform")
		}
		animations[i].Frames = make([]AnimKeyFrame, int(frameCount))
		for f := range animations[i].Frames {
			frame := &animations[i].Frames[f]
			if frame.Time, err = r.f32(); err != nil {
				return nil, err
			}
			boneCount, err := r.u32()
			if err != nil {
				return nil, err
			}
			if uint64(boneCount) > uint64(maxInt()) {
				return nil, errors.New("animation bone count is too large for this platform")
			}
			frame.Bones = make([]AnimBone, int(boneCount))
			for b := range frame.Bones {
				bone := &frame.Bones[b]
				nodeIndex, err := r.i32()
				if err != nil {
					return nil, err
				}
				pathType, err := r.i32()
				if err != nil {
					return nil, err
				}
				interpolation, err := r.i32()
				if err != nil {
					return nil, err
				}
				bone.NodeIndex = int(nodeIndex)
				bone.PathType = AnimationPathType(pathType)
				bone.Interpolation = AnimationInterpolation(interpolation)
				if err = readRaw(r, &bone.Data); err != nil {
					return nil, err
				}
			}
		}
	}
	return animations, nil
}

func maxInt() int {
	return int(^uint(0) >> 1)
}

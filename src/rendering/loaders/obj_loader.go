package loaders

import (
	"bufio"
	"errors"
	"fmt"
	"kaiju/matrix"
	"kaiju/rendering"
	"math"
	"strings"
)

type objBuilder struct {
	name     string
	material string
	points   []matrix.Vec3
	uvs      []matrix.Vec2
	normals  []matrix.Vec3
	vIndexes []uint32
	tIndexes []uint32
	nIndexes []uint32
}

type objLineType = int

const (
	objLineTypeNotSupported = iota
	objLineTypeComment
	objLineTypeMaterialLib
	objLineTypeObject
	objLineTypeVertex
	objLineTypeUv
	objLineTypeNormal
	objLineTypeMaterial
	objLineTypeFace
	//objLineTypeGroup = 9
)

func objDecipherLine(str string) objLineType {
	runes := []rune(str)
	switch runes[0] {
	case '#':
		return objLineTypeComment
	case 'm':
		return objLineTypeMaterialLib
	case 'o':
		return objLineTypeObject
	case 'v':
		{
			switch runes[1] {
			case 't':
				return objLineTypeUv
			case 'n':
				return objLineTypeNormal
			default:
				return objLineTypeVertex
			}
		}
	case 'u':
		return objLineTypeMaterial
	case 'f':
		return objLineTypeFace
	default:
		return objLineTypeNotSupported
	}
}

func objNewObject(line string) objBuilder {
	obj := objBuilder{}
	obj.points = make([]matrix.Vec3, 0)
	obj.uvs = make([]matrix.Vec2, 0)
	obj.normals = make([]matrix.Vec3, 0)
	obj.vIndexes = make([]uint32, 0)
	obj.tIndexes = make([]uint32, 0)
	obj.nIndexes = make([]uint32, 0)
	obj.material = ""
	var tmp [2]rune
	fmt.Sscanf(line, "%s %s", tmp, obj.name)
	return obj
}

func (obj *objBuilder) readVertex(line string) {
	var p matrix.Vec3
	fmt.Sscanf(line, "v %f %f %f", p.PX(), p.PY(), p.PZ())
	obj.points = append(obj.points, p)
}

func (obj *objBuilder) readUv(line string) {
	var uv matrix.Vec2
	fmt.Sscanf(line, "vt %f %f", uv.PX(), uv.PY())
	obj.uvs = append(obj.uvs, uv)
}

func (obj *objBuilder) readNormal(line string) {
	var n matrix.Vec3
	fmt.Sscanf(line, "vn %f %f %f", n.PX(), n.PY(), n.PZ())
	obj.normals = append(obj.normals, n)
}

func (obj *objBuilder) readMaterial(line string) {
	fmt.Sscanf(line, "mtllib %s", obj.material)
}

func (obj *objBuilder) readFace(line string) error {
	var v0, vt0, vn0,
		v1, vt1, vn1,
		v2, vt2, vn2 uint32
	fmt.Sscanf(line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
		&v0, &vt0, &vn0, &v1, &vt1, &vn1, &v2, &vt2, &vn2)
	v0--
	v1--
	v2--
	vt0--
	vt1--
	vt2--
	vn0--
	vn1--
	vn2--
	if v0 > math.MaxUint32 || v1 > math.MaxUint32 || v2 > math.MaxUint32 {
		return errors.New("Expected 32 bit unsigned int")
	}
	obj.vIndexes = append(obj.vIndexes, v0)
	obj.vIndexes = append(obj.vIndexes, v1)
	obj.vIndexes = append(obj.vIndexes, v2)
	obj.tIndexes = append(obj.tIndexes, vt0)
	obj.tIndexes = append(obj.tIndexes, vt1)
	obj.tIndexes = append(obj.tIndexes, vt2)
	obj.nIndexes = append(obj.nIndexes, vn0)
	obj.nIndexes = append(obj.nIndexes, vn1)
	obj.nIndexes = append(obj.nIndexes, vn2)
	return nil
}

func Obj(renderer rendering.Renderer, key, objData string) []*rendering.Mesh {
	builders := ObjToRaw(objData)
	meshes := make([]*rendering.Mesh, 0)
	for _, builder := range builders {
		verts := make([]rendering.Vertex, 0)
		for i := 0; i < len(builder.points); i++ {
			v := rendering.Vertex{}
			v.Position = builder.points[builder.vIndexes[i]]
			v.UV0 = builder.uvs[builder.tIndexes[i]]
			v.Normal = builder.normals[builder.nIndexes[i]]
			v.Color = matrix.ColorWhite()
			verts = append(verts, v)
		}
		mesh := rendering.NewMesh(key, verts, builder.vIndexes)
		meshes = append(meshes, mesh)
	}
	return meshes
}

func ObjToRaw(objData string) []objBuilder {
	var matLib string
	var builders []objBuilder
	builder := objBuilder{}
	builderSet := false
	scan := bufio.NewScanner(strings.NewReader(objData))
	for scan.Scan() {
		line := scan.Text()
		lineType := objDecipherLine(line)
		switch lineType {
		case objLineTypeMaterialLib:
			fmt.Sscanf(line, "usemtl %s", matLib)
		case objLineTypeObject:
			if builderSet {
				builders = append(builders, builder)
			}
			builder = objNewObject(line)
			builderSet = true
		case objLineTypeVertex:
			builder.readVertex(line)
		case objLineTypeUv:
			builder.readUv(line)
		case objLineTypeNormal:
			builder.readNormal(line)
		case objLineTypeMaterial:
			builder.readMaterial(line)
		case objLineTypeFace:
			builder.readFace(line)
		case objLineTypeNotSupported:
		case objLineTypeComment:
			break
		}
	}
	builders = append(builders, builder)
	return builders
}

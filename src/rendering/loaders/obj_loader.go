/******************************************************************************/
/* obj_loader.go                                                              */
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
	"bufio"
	"fmt"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders/load_result"
	"strings"
)

type objBuilder struct {
	name                 string
	material             string
	points               []matrix.Vec3
	colors               []matrix.Color
	uvs                  []matrix.Vec2
	normals              []matrix.Vec3
	vIndexes             []uint32
	tIndexes             []uint32
	nIndexes             []uint32
	complainedAboutQuads bool
}

func (o *objBuilder) fromVertIdx(idx int) int {
	for i, v := range o.vIndexes {
		if v == uint32(idx) {
			return i
		}
	}
	return 0
}

func (o *objBuilder) complainAboutQuads() {
	if !o.complainedAboutQuads {
		klib.NotYetImplemented(139)
		o.complainedAboutQuads = true
	}
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
	c := matrix.ColorWhite()
	spaceCount := strings.Count(line, " ")
	if spaceCount == 3 {
		fmt.Sscanf(line, "v %f %f %f", p.PX(), p.PY(), p.PZ())
	} else if spaceCount == 6 {
		fmt.Sscanf(line, "v %f %f %f %f %f %f", p.PX(), p.PY(), p.PZ(), c.PR(), c.PG(), c.PB())
	}
	obj.points = append(obj.points, p)
	obj.colors = append(obj.colors, c)
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

func (obj *objBuilder) readFace(line string) {
	var v, vt, vn [4]uint32
	spaceCount := strings.Count(line, " ")
	if spaceCount == 3 {
		fmt.Sscanf(line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
			&v[0], &vt[0], &vn[0], &v[1], &vt[1], &vn[1], &v[2], &vt[2], &vn[2])
		for i := 0; i < 3; i++ {
			v[i]--
			vt[i]--
			vn[i]--
		}
		obj.vIndexes = append(obj.vIndexes, v[:3]...)
		obj.tIndexes = append(obj.tIndexes, vt[:3]...)
		obj.nIndexes = append(obj.nIndexes, vn[:3]...)
	} else if spaceCount == 4 {
		obj.complainAboutQuads()
		fmt.Sscanf(line, "f %d/%d/%d %d/%d/%d %d/%d/%d %d/%d/%d",
			&v[0], &vt[0], &vn[0], &v[1], &vt[1], &vn[1], &v[2], &vt[2], &vn[2], &v[3], &vt[3], &vn[3])
		for i := 0; i < 4; i++ {
			v[i]--
			vt[i]--
			vn[i]--
		}
		obj.vIndexes = append(obj.vIndexes, v[:]...)
		obj.tIndexes = append(obj.tIndexes, vt[:]...)
		obj.nIndexes = append(obj.nIndexes, vn[:]...)
	}
}

func OBJ(objData string) load_result.Result {
	builders := ObjToRaw(objData)
	res := load_result.NewResult()
	for i := range builders {
		builder := &builders[i]
		verts := make([]rendering.Vertex, len(builder.points))
		for j := range builder.points {
			vi := builder.fromVertIdx(j)
			verts[j] = rendering.Vertex{
				Position: builder.points[j],
				UV0:      builder.uvs[builder.tIndexes[vi]],
				Normal:   builder.normals[builder.nIndexes[vi]],
				Color:    builder.colors[builder.vIndexes[vi]],
			}
		}
		res.Add("", builder.name, verts, builder.vIndexes, nil)
	}
	return res
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

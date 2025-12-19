/******************************************************************************/
/* obj_loader.go                                                              */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package loaders

import (
	"bufio"
	"errors"
	"fmt"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/loaders/load_result"
	"path/filepath"
	"strconv"
	"strings"
)

type objLibrary struct {
	points    []matrix.Vec3
	colors    []matrix.Color
	uvs       []matrix.Vec2
	normals   []matrix.Vec3
	materials []string
}

type objBuilder struct {
	name     string
	material string
	points   []matrix.Vec3
	colors   []matrix.Color
	vIndexes []uint32
	tIndexes []uint32
	nIndexes []uint32
}

func (o *objBuilder) fromVertIdx(idx int) int {
	for i, v := range o.vIndexes {
		if v == uint32(idx) {
			return i
		}
	}
	return 0
}

type objOffsets struct {
	pointsOffset  uint32
	uvsOffset     uint32
	normalsOffset uint32
	colorsOffset  uint32
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

func objNewObject(line string) *objBuilder {
	obj := objBuilder{}
	obj.points = make([]matrix.Vec3, 0)
	obj.vIndexes = make([]uint32, 0)
	obj.tIndexes = make([]uint32, 0)
	obj.nIndexes = make([]uint32, 0)
	obj.material = ""
	var tmp [2]rune
	fmt.Sscanf(line, "%s %s", tmp, obj.name)
	return &obj
}

func (objLib *objLibrary) readVertex(line string) error {
	var p matrix.Vec3
	c := matrix.ColorWhite()
	spaceCount := strings.Count(line, " ")
	if spaceCount == 3 {
		fmt.Sscanf(line, "v %f %f %f", p.PX(), p.PY(), p.PZ())
	} else if spaceCount == 6 {
		fmt.Sscanf(line, "v %f %f %f %f %f %f", p.PX(), p.PY(), p.PZ(), c.PR(), c.PG(), c.PB())
	} else {
		return errors.New("invalid OBJ file")
	}
	objLib.points = append(objLib.points, p)
	objLib.colors = append(objLib.colors, c)
	return nil
}

func (objLib *objLibrary) readUv(line string) error {
	var uv matrix.Vec2
	spaceCount := strings.Count(line, " ")
	if spaceCount == 2 {
		fmt.Sscanf(line, "vt %f %f", uv.PX(), uv.PY())
	} else {
		return errors.New("invalid OBJ file")
	}
	objLib.uvs = append(objLib.uvs, uv)
	return nil
}

func (objLib *objLibrary) readNormal(line string) error {
	var n matrix.Vec3
	spaceCount := strings.Count(line, " ")
	if spaceCount == 3 {
		fmt.Sscanf(line, "vn %f %f %f", n.PX(), n.PY(), n.PZ())
	} else {
		return errors.New("invalid OBJ file")
	}
	objLib.normals = append(objLib.normals, n)
	return nil
}

func (objLib *objLibrary) readMaterial(line string) error {
	var materialName string
	fmt.Sscanf(line, "mtllib %s", &materialName)
	objLib.materials = append(objLib.materials, materialName)
	return nil
}

func (obj *objBuilder) readFace(line string, objLib objLibrary) error {
	fields := strings.Fields(line)[1:]
	v := make([]uint32, 0, len(fields))
	vt := make([]uint32, 0, len(fields))
	vn := make([]uint32, 0, len(fields))

	for _, field := range fields {
		parts := strings.Split(field, "/")
		if len(parts) < 1 || parts[0] == "" {
			return errors.New("invalid OBJ file")
		}
		vIdx, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return errors.New("invalid OBJ file")
		}
		if uint32(len(objLib.points)) < uint32(vIdx) {
			return errors.New("invalid OBJ file")
		}
		// We are creating new vertices for each face for proper normals
		v = append(v, uint32(len(obj.points)))
		obj.points = append(obj.points, objLib.points[vIdx-1])
		obj.colors = append(obj.colors, objLib.colors[vIdx-1])

		if len(parts) >= 2 && parts[1] != "" {
			vtIdx, err := strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return errors.New("invalid OBJ file")
			}
			vt = append(vt, uint32(vtIdx-1))
		} else {
			// TODO: What if nothing is specified?
			vt = append(vt, 0)
		}

		if len(parts) >= 3 && parts[2] != "" {
			vnIdx, err := strconv.ParseUint(parts[2], 10, 32)
			if err != nil {
				return errors.New("invalid OBJ file")
			}
			vn = append(vn, uint32(vnIdx-1))
		} else {
			// TODO: What if nothing is specified?
			vn = append(vn, 0)
		}
	}

	if len(fields) >= 3 {
		// Fan triangulation
		pIdx := 1
		for i := 2; i < len(v); i++ {
			obj.vIndexes = append(obj.vIndexes, v[0])
			obj.vIndexes = append(obj.vIndexes, v[pIdx])
			obj.vIndexes = append(obj.vIndexes, v[i])
			obj.tIndexes = append(obj.tIndexes, vt[0])
			obj.tIndexes = append(obj.tIndexes, vt[pIdx])
			obj.tIndexes = append(obj.tIndexes, vt[i])
			obj.nIndexes = append(obj.nIndexes, vn[0])
			obj.nIndexes = append(obj.nIndexes, vn[pIdx])
			obj.nIndexes = append(obj.nIndexes, vn[i])
			pIdx = i
		}
	} else {
		return errors.New("invalid OBJ file")
	}
	return nil
}

func OBJ(path string, assetDB assets.Database) (load_result.Result, error) {
	defer tracing.NewRegion("loaders.OBJ").End()
	if !assetDB.Exists(path) {
		return load_result.Result{}, errors.New("file does not exist")
	} else if filepath.Ext(path) == ".obj" {
		objData, err := assetDB.ReadText(path)
		if err != nil {
			return load_result.Result{}, err
		}
		builders, library, err := ObjToRaw(objData)
		if err != nil {
			return load_result.Result{}, err
		}
		res := load_result.Result{}
		for i := range builders {
			builder := builders[i]
			verts := make([]rendering.Vertex, len(builder.points))
			for j := range builder.points {
				vi := builder.fromVertIdx(j)
				verts[j] = rendering.Vertex{
					Position: builder.points[j],
					UV0:      library.uvs[builder.tIndexes[vi]],
					Normal:   library.normals[builder.nIndexes[vi]],
					Color:    builder.colors[j],
				}
			}
			// TODO:  Read the .obj material file for textures
			res.Add(builder.name, builder.name, verts, builder.vIndexes, map[string]string{}, nil)
		}
		return res, nil
	} else {
		return load_result.Result{}, errors.New("invalid file extension")
	}
}

func ObjToRaw(objData string) ([]*objBuilder, objLibrary, error) {
	var currentMaterial string
	var builders []*objBuilder
	var surfaces map[string]*objBuilder
	var builder *objBuilder
	library := objLibrary{}
	builderSet := false
	scan := bufio.NewScanner(strings.NewReader(objData))
	for scan.Scan() {
		line := scan.Text()
		lineType := objDecipherLine(line)
		switch lineType {
		case objLineTypeMaterial:
			fmt.Sscanf(line, "usemtl %s", &currentMaterial)
			if len(surfaces) == 0 {
				surfaces = make(map[string]*objBuilder, 1)
			}
			if len(builder.points) == 0 {
				builder.material = strings.Clone(currentMaterial)
				surfaces[currentMaterial] = builder
			} else {
				if val, ok := surfaces[currentMaterial]; ok {
					builder = val
				} else {
					builder = objNewObject("O " + builder.name)
					if len(line) > 2 {
						builder.name = line[2:]
					}
					builder.material = strings.Clone(currentMaterial)
					surfaces[currentMaterial] = builder
				}
			}
		case objLineTypeObject:
			if builderSet {
				if len(surfaces) > 0 {
					for _, v := range surfaces {
						builders = append(builders, v)
					}
				} else {
					builders = append(builders, builder)
				}
			}
			builder = objNewObject(line)
			if len(line) > 2 {
				builder.name = line[2:]
			}
			if currentMaterial != "" {
				surfaces = make(map[string]*objBuilder, 1)
				surfaces[currentMaterial] = builder
			}
			builder.material = strings.Clone(currentMaterial)
			builderSet = true
		case objLineTypeVertex:
			err := library.readVertex(line)
			if err != nil {
				return nil, objLibrary{}, err
			}
		case objLineTypeUv:
			err := library.readUv(line)
			if err != nil {
				return nil, objLibrary{}, err
			}
		case objLineTypeNormal:
			err := library.readNormal(line)
			if err != nil {
				return nil, objLibrary{}, err
			}
		case objLineTypeMaterialLib:
			library.readMaterial(line)
		case objLineTypeFace:
			err := builder.readFace(line, library)
			if err != nil {
				return nil, objLibrary{}, err
			}
		case objLineTypeNotSupported:
		case objLineTypeComment:
			break
		}
	}
	if len(surfaces) > 0 {
		for _, v := range surfaces {
			builders = append(builders, v)
		}
	} else {
		builders = append(builders, builder)
	}
	return builders, library, nil
}

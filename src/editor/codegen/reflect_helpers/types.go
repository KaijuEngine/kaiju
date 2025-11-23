/******************************************************************************/
/* types.go                                                                   */
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

package reflect_helpers

import (
	"kaiju/klib"
	"kaiju/matrix"
	"log/slog"
	"strconv"
	"strings"
	"unsafe"
)

func StringToTypeValue(typeName, v string) any {
	switch typeName {
	case "string":
		return v
	case "bool":
		switch strings.ToLower(v) {
		case "false":
			return false
		case "true":
			return true
		default:
			slog.Warn("unexpected tag string value for bool, expected 'true' or 'false'", "value", v)
			return true
		}
	case "int":
		return int(klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int(0))))))
	case "int8":
		return int8(klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int8(0))))))
	case "int16":
		return int16(klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int16(0))))))
	case "int32":
		return int32(klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int32(0))*8))))
	case "int64":
		return klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int64(0))*8)))
	case "uint":
		return uint(klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint(0))*8))))
	case "uint8":
		return uint8(klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint8(0))*8))))
	case "uint16":
		return uint16(klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint16(0))*8))))
	case "uint32":
		return uint32(klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint32(0))*8))))
	case "uint64":
		return klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint64(0)))))
	case "float32":
		return float32(klib.ShouldReturn(strconv.ParseFloat(v, int(unsafe.Sizeof(float32(0))*8))))
	case "float64":
		return klib.ShouldReturn(strconv.ParseFloat(v, int(unsafe.Sizeof(float64(0))*8)))
	case "uintptr":
		return klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint64(0))*8)))
	case "complex64":
		return complex64(klib.ShouldReturn(strconv.ParseComplex(v, int(unsafe.Sizeof(complex64(0))*8))))
	case "complex128":
		return klib.ShouldReturn(strconv.ParseComplex(v, int(unsafe.Sizeof(complex128(0))*8)))
	case "Vec2":
		out := matrix.Vec2{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			out[i] = float32(klib.ShouldReturn(strconv.ParseFloat(parts[i], int(unsafe.Sizeof(float32(0))*8))))
		}
		return out
	case "Vec3":
		out := matrix.Vec3{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			p := strings.TrimSpace(parts[i])
			out[i] = float32(klib.ShouldReturn(strconv.ParseFloat(p, int(unsafe.Sizeof(float32(0))*8))))
		}
		return out
	case "Vec4":
		out := matrix.Vec4{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			out[i] = float32(klib.ShouldReturn(strconv.ParseFloat(parts[i], int(unsafe.Sizeof(float32(0))*8))))
		}
		return out
	}
	return nil
}

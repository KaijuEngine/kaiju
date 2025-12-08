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
	"log/slog"
	"strconv"
	"strings"
	"unsafe"

	"github.com/KaijuEngine/kaiju/klib"
	"github.com/KaijuEngine/kaiju/matrix"
)

func bitSize[T klib.Number]() int { return int(unsafe.Sizeof(T(0))) * 8 }

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
		v = klib.CleanNumString(v)
		return int(klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int]())))
	case "int8":
		v = klib.CleanNumString(v)
		return int8(klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int8]())))
	case "int16":
		v = klib.CleanNumString(v)
		return int16(klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int16]())))
	case "int32":
		v = klib.CleanNumString(v)
		return int32(klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int32]())))
	case "int64":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int64]()))
	case "uint":
		v = klib.CleanNumString(v)
		return uint(klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint]())))
	case "uint8":
		v = klib.CleanNumString(v)
		return uint8(klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint8]())))
	case "uint16":
		v = klib.CleanNumString(v)
		return uint16(klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint16]())))
	case "uint32":
		v = klib.CleanNumString(v)
		return uint32(klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint32]())))
	case "uint64":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint64]()))
	case "float32":
		v = klib.CleanNumString(v)
		return float32(klib.ShouldReturn(strconv.ParseFloat(v, bitSize[float32]())))
	case "float64":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseFloat(v, bitSize[float64]()))
	case "uintptr":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uintptr]()))
	case "complex64":
		v = klib.CleanNumString(v)
		return complex64(klib.ShouldReturn(strconv.ParseComplex(v, bitSize[complex64]())))
	case "complex128":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseComplex(v, bitSize[complex128]()))
	case "Vec2":
		out := matrix.Vec2{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			p := klib.CleanNumString(parts[i])
			out[i] = float32(klib.ShouldReturn(strconv.ParseFloat(p, bitSize[float32]())))
		}
		return out
	case "Vec3":
		out := matrix.Vec3{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			p := klib.CleanNumString(parts[i])
			out[i] = float32(klib.ShouldReturn(strconv.ParseFloat(p, bitSize[float32]())))
		}
		return out
	case "Vec4":
		out := matrix.Vec4{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			p := klib.CleanNumString(parts[i])
			out[i] = float32(klib.ShouldReturn(strconv.ParseFloat(p, bitSize[float32]())))
		}
		return out
	}
	return nil
}

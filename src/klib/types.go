/******************************************************************************/
/* types.go                                                                   */
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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package klib

import (
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
		return int(ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int(0))))))
	case "int8":
		return int8(ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int8(0))))))
	case "int16":
		return int16(ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int16(0))))))
	case "int32":
		return int32(ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int32(0))))))
	case "int64":
		return ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int64(0)))))
	case "uint":
		return uint(ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint(0))))))
	case "uint8":
		return uint8(ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint8(0))))))
	case "uint16":
		return uint16(ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint16(0))))))
	case "uint32":
		return uint32(ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint32(0))))))
	case "uint64":
		return ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint64(0)))))
	case "float32":
		return float32(ShouldReturn(strconv.ParseFloat(v, int(unsafe.Sizeof(float32(0))))))
	case "float64":
		return ShouldReturn(strconv.ParseFloat(v, int(unsafe.Sizeof(float64(0)))))
	case "uintptr":
		return ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint64(0)))))
	case "complex64":
		return complex64(ShouldReturn(strconv.ParseComplex(v, int(unsafe.Sizeof(complex64(0))))))
	case "complex128":
		return ShouldReturn(strconv.ParseComplex(v, int(unsafe.Sizeof(complex128(0)))))
	}
	return nil
}

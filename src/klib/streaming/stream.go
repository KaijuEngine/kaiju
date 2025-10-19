/******************************************************************************/
/* stream.go                                                                  */
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

package streaming

import (
	"encoding/binary"
	"io"
)

func StreamWrite(w io.Writer, data ...any) error {
	var err error
	for i := 0; i < len(data) && err == nil; i++ {
		switch d := data[i].(type) {
		case string:
			err = binary.Write(w, binary.LittleEndian, int32(len(d)))
			if err == nil {
				err = binary.Write(w, binary.LittleEndian, []byte(d))
			}
		case int:
			err = binary.Write(w, binary.LittleEndian, int32(d))
		default:
			err = binary.Write(w, binary.LittleEndian, data[i])
		}
	}
	return err
}

func StreamRead(r io.Reader, outData ...any) error {
	var err error
	for i := 0; i < len(outData) && err == nil; i++ {
		switch d := outData[i].(type) {
		case *string:
			size := int32(0)
			err = binary.Read(r, binary.LittleEndian, &size)
			if err == nil {
				data := make([]byte, size)
				err = binary.Read(r, binary.LittleEndian, data)
				if err == nil {
					*d = string(data)
				}
			}
		case *int:
			out := int32(0)
			err = binary.Read(r, binary.LittleEndian, &out)
			if err == nil {
				*d = int(out)
			}
		default:
			err = binary.Read(r, binary.LittleEndian, outData[i])
		}
	}
	return err
}

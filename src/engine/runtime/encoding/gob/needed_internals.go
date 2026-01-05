/******************************************************************************/
/* needed_internals.go                                                        */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com                           */
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

package gob

import (
	"io"
	"unsafe"
)

const chunk = 10 << 20 // 10M

func saferioSliceCap[E any](c uint64) int {
	var v E
	size := uint64(unsafe.Sizeof(v))
	return saferioSliceCapWithSize(size, c)
}

func saferioSliceCapWithSize(size, c uint64) int {
	if int64(c) < 0 || c != uint64(int(c)) {
		return -1
	}
	if size > 0 && c > (1<<64-1)/size {
		return -1
	}
	if c*size > chunk {
		c = chunk / size
		if c == 0 {
			c = 1
		}
	}
	return int(c)
}

func saferioReadData(r io.Reader, n uint64) ([]byte, error) {
	if int64(n) < 0 || n != uint64(int(n)) {
		// n is too large to fit in int, so we can't allocate
		// a buffer large enough. Treat this as a read failure.
		return nil, io.ErrUnexpectedEOF
	}

	if n < chunk {
		buf := make([]byte, n)
		_, err := io.ReadFull(r, buf)
		if err != nil {
			return nil, err
		}
		return buf, nil
	}

	var buf []byte
	buf1 := make([]byte, chunk)
	for n > 0 {
		next := n
		if next > chunk {
			next = chunk
		}
		_, err := io.ReadFull(r, buf1[:next])
		if err != nil {
			if len(buf) > 0 && err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
		buf = append(buf, buf1[:next]...)
		n -= next
	}
	return buf, nil
}

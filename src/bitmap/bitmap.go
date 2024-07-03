/******************************************************************************/
/* bitmap.go                                                                  */
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

package bitmap

import "math"

const bitsInByte = 8

type Bitmap []byte

// New creates a new bitmap of the specified length. A bitmap is a slice of
// bytes where each bit represents a boolean value. The length is the number
// of bits that the bitmap will represent. The length is rounded up to the
// nearest byte.
func New(length int) Bitmap {
	return make([]byte, LengthFor(length))
}

// NewTrue creates a new bitmap of the specified length with all bits true
func NewTrue(length int) Bitmap {
	b := make([]byte, LengthFor(length))
	for i := 0; i < len(b); i++ {
		b[i] = 0xFF
	}
	return b
}

// LengthFor returns the number of bytes needed to represent the specified
// number of bits.
func LengthFor(byteCount int) int {
	return int(math.Ceil(float64(byteCount) / float64(bitsInByte)))
}

// Set sets the bit at the specified index to true.
func (b Bitmap) Set(index int) {
	b[index/bitsInByte] |= 0x01 << (index % bitsInByte)
}

// Assign sets the bit at the specified index to the specified value.
func (b Bitmap) Assign(index int, value bool) {
	if value {
		b.Set(index)
	} else {
		b.Reset(index)
	}
}

// Reset sets the bit at the specified index to false.
func (b Bitmap) Reset(index int) {
	b[index/bitsInByte] &= ^(0x01 << (index % bitsInByte))
}

// Toggle flips the value of the bit at the specified index.
func (b Bitmap) Toggle(index int) {
	b[index/bitsInByte] ^= 0x01 << (index % bitsInByte)
}

// CountInverse returns the number of bits that are false.
func (b Bitmap) CountInverse() int {
	return len(b)*bitsInByte - Count(b)
}

// Clear sets all bits to false.
func (b Bitmap) Clear() {
	clear(b)
}

/*****************************************************************************/
/* bitmap.go                                                                 */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package bitmap

const bitsInByte = 8

type Bitmap []byte

func New(length int) Bitmap {
	return make([]byte, LengthFor(length))
}

func NewTrue(length int) Bitmap {
	b := make([]byte, LengthFor(length))
	for i := 0; i < len(b); i++ {
		b[i] = 0xFF
	}
	return b
}

func LengthFor(byteCount int) int {
	return (byteCount / bitsInByte) + 1
}

func (b Bitmap) Check(index int) bool {
	return b[index/bitsInByte]&0x01<<(index%bitsInByte) != 0
}

func (b Bitmap) Set(index int) {
	b[index/bitsInByte] |= 0x01 << (index % bitsInByte)
}

func (b Bitmap) Assign(index int, value bool) {
	if value {
		b.Set(index)
	} else {
		b.Reset(index)
	}
}

func (b Bitmap) Reset(index int) {
	b[index/bitsInByte] &= ^(0x01 << (index % bitsInByte))
}

func (b Bitmap) Toggle(index int) {
	b[index/bitsInByte] ^= 0x01 << (index % bitsInByte)
}

func (b Bitmap) Count() int {
	count := 0
	length := len(b) * bitsInByte
	for i := 0; i < length; i++ {
		if b.Check(i) {
			count++
		}
	}
	return count
}

func (b Bitmap) CountInverse() int {
	return len(b)*bitsInByte - b.Count()
}

func (b Bitmap) Clear() {
	clear(b)
}

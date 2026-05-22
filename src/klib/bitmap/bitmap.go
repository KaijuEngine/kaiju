/******************************************************************************/
/* bitmap.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bitmap

import (
	"math"
	"math/bits"
)

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

// IsSet will return true if the given bit is set to 1, otherwise false
func (b Bitmap) IsSet(index int) bool {
	return b[index/bitsInByte]&(0x01<<(index%bitsInByte)) != 0
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

// Count returns the number of bits that are true.
func Count(b Bitmap) int {
	count := 0
	for i := 0; i < len(b); i++ {
		count += bits.OnesCount(uint(b[i]))
	}
	return count
}

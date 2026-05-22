//go:build !amd64

/******************************************************************************/
/* bitmap.pure.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bitmap

// Check returns the value of the bit at the specified index.
func Check(b Bitmap, index int) bool {
	return (b[index/bitsInByte] & (0x01 << (index % bitsInByte))) != 0
}

func CountASM(b Bitmap) int { return Count(b) }

func CountASMUsingTable(b Bitmap) int { return Count(b) }

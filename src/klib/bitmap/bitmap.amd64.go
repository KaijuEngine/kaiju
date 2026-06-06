//go:build amd64

/******************************************************************************/
/* bitmap.amd64.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bitmap

//go:noescape
func Check(b Bitmap, index int) bool

//go:noescape
func CountASM(b Bitmap) int

//go:noescape
func CountASMUsingTable(b Bitmap) int

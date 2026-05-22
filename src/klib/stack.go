/******************************************************************************/
/* stack.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"fmt"
	"runtime"
)

func PrintStack() {
	buf := [4096]byte{}
	n := runtime.Stack(buf[:], false) // false: only current goroutine
	fmt.Printf("Stack trace:\n%s\n", buf[:n])
}

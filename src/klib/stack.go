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

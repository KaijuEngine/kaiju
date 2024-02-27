package klib

import "runtime"

func ExeExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

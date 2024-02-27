package klib

import (
	"os/exec"
	"runtime"
)

func OpenWebsite(url string) {
	cmd := "open"
	if runtime.GOOS == "windows" {
		cmd = "explorer"
	}
	exec.Command(cmd, url).Run()
}

package main

import (
	_ "embed"
	"kaiju/filesystem"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed ignore.txt
var ignore string

func findRoot() (string, error) {
	wd, err := os.Getwd()
	if _, goMain, _, ok := runtime.Caller(0); ok {
		if newWd, pathErr := filepath.Abs(filepath.Dir(goMain)); pathErr == nil {
			wd = filepath.Dir(newWd + "/../../")
		}
	} else if err != nil {
		return "", err
	}
	return wd, nil
}

func main() {
	root, err := findRoot()
	if err != nil {
		panic(err)
	}
	entries, err := filesystem.ListFoldersRecursive(root)
	if err != nil {
		panic(err)
	}
	ignoreEntries := strings.Split(ignore, "\n")
	for i := range ignoreEntries {
		ignoreEntries[i] = strings.TrimSpace(ignoreEntries[i])
	}
	os.Chdir(root + "/interpreter")
	for _, entry := range entries {
		entry = strings.Replace(entry, root, "", 1)
		entry = strings.TrimPrefix(strings.TrimPrefix(entry, "/"), "\\")
		skip := strings.HasPrefix(entry, ".") || len(strings.TrimSpace(entry)) == 0
		for i := 0; i < len(ignoreEntries) && !skip; i++ {
			skip = strings.HasPrefix(entry, ignoreEntries[i])
		}
		if !skip {
			pkg := "kaiju/" + entry
			println("Extracting " + pkg)
			err = exec.Command("yaegi", "extract", pkg).Run()
			if err != nil {
				panic(err)
			}
		}
	}
}

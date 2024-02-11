package main

import (
	"archive/zip"
	_ "embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

//go:embed ignore.txt
var ignore string

//go:embed launch.json.txt
var vsLaunch string

//go:embed settings.json.txt
var vsSettings string

//go:embed go.mod.txt
var goMod string

func findRootFolder() (string, error) {
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
	root, err := findRootFolder()
	if err != nil {
		panic(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}
	ignoreEntries := strings.Split(ignore, "\n")
	for i := range ignoreEntries {
		ignoreEntries[i] = strings.TrimSpace(ignoreEntries[i])
	}
	addFiles := map[string]string{
		".vscode/launch.json":   vsLaunch,
		".vscode/settings.json": vsSettings,
		"src/go.mod":            goMod,
	}
	if err := os.Chdir(root); err != nil {
		panic(err)
	}
	zipTemplate("../project_template.zip", entries, ignoreEntries, addFiles)
}

func zipTemplate(outPath string, entries []fs.DirEntry, ignore []string, explicitFiles map[string]string) {
	file, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	w := zip.NewWriter(file)
	defer w.Close()
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || slices.Contains(ignore, path) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		f, err := w.Create("src/" + path)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}
		return nil
	}
	for _, entry := range entries {
		err = filepath.Walk(entry.Name(), walker)
		if err != nil {
			panic(err)
		}
	}
	for to, text := range explicitFiles {
		f, err := w.Create(to)
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(text))
		if err != nil {
			panic(err)
		}
	}
}

package main

import (
	_ "embed"
	"kaiju/filesystem"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed ignore.txt
var ignore string

//go:embed launch.json.txt
var launch string

//go:embed build.bat.txt
var buildBat string

//go:embed build.js.bat.txt
var buildJSBat string

//go:embed build.sh.txt
var buildSh string

//go:embed build.js.sh.txt
var buildJSSh string

//go:embed go.mod.txt
var goMod string

func findRootAndProjectTemplateFolders() (string, string, error) {
	wd, err := os.Getwd()
	if _, goMain, _, ok := runtime.Caller(0); ok {
		if newWd, pathErr := filepath.Abs(filepath.Dir(goMain)); pathErr == nil {
			wd = filepath.Dir(newWd + "/../../")
		}
	} else if err != nil {
		return "", "", err
	}
	return wd, wd + "/project_template", nil
}

func writeLaunchFiles(projTemplateFolder string) error {
	files := map[string]string{
		".vscode/launch.json": launch,
		"build/build.bat":     buildBat,
		"build/build.js.bat":  buildJSBat,
		"build/build.sh":      buildSh,
		"build/build.js.sh":   buildJSSh,
		"go.mod":              goMod,
	}
	if err := os.Mkdir(filepath.Join(projTemplateFolder, ".vscode"), 0655); err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Join(projTemplateFolder, "build"), 0655); err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Join(projTemplateFolder, "bin"), 0655); err != nil {
		return err
	}
	for file, content := range files {
		f, err := os.Create(filepath.Join(projTemplateFolder, file))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(content)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	root, projTemplateFolder, err := findRootAndProjectTemplateFolders()
	if err != nil {
		panic(err)
	}
	if err = os.RemoveAll(projTemplateFolder); err != nil {
		panic(err)
	}
	if err = os.Mkdir(projTemplateFolder, 0655); err != nil {
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
	for _, entry := range entries {
		skip := false
		for i := 0; i < len(ignoreEntries) && !skip; i++ {
			skip = entry.Name() == ignoreEntries[i]
		}
		if !skip {
			from := filepath.Join(root, entry.Name())
			to := filepath.Join(projTemplateFolder, entry.Name())
			if entry.IsDir() {
				if err = filesystem.CopyDirectory(from, to); err != nil {
					println("Error copying directory: " + from + " to: " + to + " error: " + err.Error())
				}
			} else {
				if err = filesystem.CopyFile(from, to); err != nil {
					println("Error copying file: " + from + " to: " + to + " error: " + err.Error())
				}
			}
		}
	}
	if err = writeLaunchFiles(projTemplateFolder); err != nil {
		panic(err)
	}
}

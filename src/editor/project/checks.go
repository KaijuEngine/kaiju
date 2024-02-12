package project

import (
	"kaiju/filesystem"
	"os"
	"path/filepath"
	"strings"
)

func IsProjectDirectory(path string) bool {
	expected := []string{".vscode/launch.json", ".vscode/launch.json", "src/main.go", "src/source"}
	for _, file := range expected {
		if _, err := os.Stat(filepath.Join(path, file)); err != nil {
			return false
		}
	}
	if src, err := filesystem.ReadTextFile(filepath.Join(path, "src/main.go")); err != nil {
		return false
	} else if !strings.Contains(src, "KAIJU ENGINE") {
		return false
	}
	return true
}

package project_file_system

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func (pfs *FileSystem) copyStockContent() error {
	const root = "editor/editor_embedded_content/editor_content"
	top, err := CodeFS.ReadDir(root)
	if err != nil {
		return err
	}
	all := []string{}
	var readSubDir func(path string) error
	readSubDir = func(path string) error {
		if strings.HasSuffix(path, "renderer/src") {
			return nil
		}
		entries, err := CodeFS.ReadDir(path)
		if err != nil {
			return err
		}
		for i := range entries {
			subPath := filepath.ToSlash(filepath.Join(path, entries[i].Name()))
			if entries[i].IsDir() {
				if err := readSubDir(subPath); err != nil {
					return err
				}
				continue
			}
			all = append(all, subPath)
		}
		return nil
	}
	skip := []string{"editor", "meshes"}
	for i := range top {
		if !top[i].IsDir() {
			continue
		}
		name := top[i].Name()
		if slices.Contains(skip, name) {
			continue
		}
		if err := readSubDir(filepath.ToSlash(filepath.Join(root, name))); err != nil {
			return err
		}
	}
	for i := range all {
		outPath := filepath.Join(StockFolder, filepath.Base(all[i]))
		data, err := CodeFS.ReadFile(all[i])
		if err != nil {
			return err
		}
		if err := pfs.WriteFile(outPath, data, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

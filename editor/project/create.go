package project

import (
	"errors"
	"kaiju/filesystem"
	"os"
	"path/filepath"
)

const projectTemplateFolder = "project_template"

func createSource(projTemplateFolder string) error {
	sourceDir := filepath.Join(projTemplateFolder, "/source")
	err := os.Mkdir(sourceDir, 0655)
	if err != nil {
		return err
	}
	mainFile := filepath.Join(sourceDir, "/source.go")
	_, err = os.Stat(mainFile)
	if err == nil {
		return errors.New("source file already exists and should not")
	}
	f, err := os.Create(mainFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(`package source

func Main() {
	// TODO:  Bootstrap your project here
}
`)
	return err
}

func CreateNewProject(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if err = os.MkdirAll(path, 0755); err != nil {
			return err
		}
	} else if !stat.IsDir() {
		return os.ErrExist
	}
	if err = filesystem.CopyDirectory(projectTemplateFolder, path); err != nil {
		return err
	}
	return createSource(path)
}

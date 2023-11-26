package project

import (
	"errors"
	"kaiju/filesystem"
	"os"
	"path/filepath"
	"strings"
)

const projectTemplateFolder = "project_template"

func createSource(projTemplateFolder string) error {
	sourceDir := filepath.Join(projTemplateFolder, "/source")
	err := os.Mkdir(sourceDir, 0755)
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

import "kaiju/engine"

func Main(host *engine.Host) {
	
}
`)
	return err
}

func setupBuildScripts(projectName, projTemplateFolder string) error {
	buildDir := filepath.Join(projTemplateFolder, "/build")
	files, err := filesystem.ListFilesRecursive(buildDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		src, err := filesystem.ReadTextFile(file)
		if err != nil {
			return err
		}
		src = strings.ReplaceAll(src, "[PROJECT_NAME]", projectName)
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(src)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateNewProject(projectName, path string) error {
	if filepath.Base(path) != projectName {
		return errors.New("project name and path do not match")
	}
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
	if err = setupBuildScripts(projectName, projectTemplateFolder); err != nil {
		return err
	}
	return createSource(path)
}

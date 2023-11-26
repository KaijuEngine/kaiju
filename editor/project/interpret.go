package project

import (
	"kaiju/engine"
	"kaiju/filesystem"
	"kaiju/interpreter"
	"path/filepath"

	"github.com/KaijuEngine/yaegi/interp"
	"github.com/KaijuEngine/yaegi/stdlib"
)

func InterpretSource(projectPath string, host *engine.Host) error {
	//project.CreateNewProject(projectPath)
	itp := interp.New(interp.Options{ImportSub: interp.ImportSubstitution{
		Prefix:  "kaiju/source",
		Replace: projectPath + "/source",
	}})
	if err := itp.Use(stdlib.Symbols); err != nil {
		return err
	}
	if err := itp.Use(interpreter.Symbols); err != nil {
		return err
	}
	src, err := filesystem.ReadTextFile(filepath.Join(projectPath, "source/source.go"))
	if err != nil {
		return err
	}
	if _, err := itp.Eval(src); err != nil {
		return err
	}
	var entry func(*engine.Host)
	if sourceMain, err := itp.Eval("source.Main"); err != nil {
		return err
	} else {
		entry = sourceMain.Interface().(func(*engine.Host))
	}
	entry(host)
	return nil
}

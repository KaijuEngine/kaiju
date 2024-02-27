/******************************************************************************/
/* run_project.go                                                             */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package editor

import (
	"bufio"
	"errors"
	"kaiju/klib"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
)

func (ed *Editor) runProject(isDebug bool) {
	go func() {
		ed.statusBar.SetMessage("Running code tidy...")
		if err := ed.tidyProjectCode(); err != nil {
			slog.Error(err.Error())
			return
		}
		ed.statusBar.SetMessage("Compiling code...")
		if err := ed.compileProjectCode(isDebug); err != nil {
			slog.Error(err.Error())
			return
		}
		ed.statusBar.SetMessage("Compilation completed successfully, launching project...")
		ed.launchProject(isDebug)
	}()
}

func (ed *Editor) codeCompilerPath() (string, error) {
	const releasePath = "/bin/go/bin/go"
	const developPath = "/../go/bin/go"
	kaijuCompiler := filepath.Join(ed.editorDir, releasePath+klib.ExeExtension())
	if _, err := os.Stat(kaijuCompiler); os.IsNotExist(err) {
		kaijuCompiler = filepath.Join(ed.editorDir, developPath+klib.ExeExtension())
	}
	if _, err := os.Stat(kaijuCompiler); os.IsNotExist(err) {
		return "", errors.New("failed to find the Kaiju Go compiler")
	}
	os.MkdirAll("bin/obj", os.ModePerm)
	return kaijuCompiler, nil
}

func (ed *Editor) tidyProjectCode() error {
	kaijuCompiler, err := ed.codeCompilerPath()
	if err != nil {
		return err
	}
	cmd := exec.Command(kaijuCompiler, "mod", "tidy")
	return ed.runCodeCommand(cmd)
}

func (ed *Editor) compileProjectCode(isDebug bool) error {
	kaijuCompiler, err := ed.codeCompilerPath()
	if err != nil {
		return err
	}
	args := []string{"build", "-v"}
	if isDebug {
		args = append(args, `-ldflags=-s -w`, "-tags=debug")
	}
	args = append(args, "-o", "../bin/kaiju"+klib.ExeExtension(), "main.go")
	cmd := exec.Command(kaijuCompiler, args...)
	if runtime.GOOS == "windows" {
		cmd.Env = append(cmd.Env, `CGO_LDFLAGS="-lgdi32 -lXInput"`)
	} else if runtime.GOOS == "linux" {
		cmd.Env = append(cmd.Env, `CGO_LDFLAGS="-lX11"`)
	}
	cmd.Env = slices.Clone(os.Environ())
	cmd.Env = append(cmd.Env, `GOTMPDIR=../bin/obj`)
	return ed.runCodeCommand(cmd)
}

func (ed *Editor) launchProject(isDebug bool) {
	args := []string{}
	if isDebug {
		args = append(args, `-stage=`+ed.stageManager.StageName())
	}
	cmd := exec.Command("bin/kaiju"+klib.ExeExtension(), args...)
	// TODO:  Create in/out pipes for bi-directional communication
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		slog.Error("failed to start the project")
	}
	cmd.Wait()
}

func (ed *Editor) runCodeCommand(cmd *exec.Cmd) error {
	cmd.Dir = "src"
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return errors.New("failed to create a pipe for the Kaiju Go compiler")
	}
	errScanner := bufio.NewScanner(errPipe)
	if err := cmd.Start(); err != nil {
		return errors.New("failed to start the Kaiju Go compiler")
	}
	hasErrors := false
	for errScanner.Scan() {
		// TODO:  Output this to a visual window and to "the" log file
		msg := errScanner.Text()
		ed.statusBar.SetMessage(msg)
		hasErrors = hasErrors || (len(msg) > 0 && msg[0] == '#')
	}
	if hasErrors {
		return errors.New("failed to compile the project")
	} else {
		return nil
	}
}

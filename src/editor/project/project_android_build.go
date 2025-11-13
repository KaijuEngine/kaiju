/******************************************************************************/
/* project_android_build.go                                                   */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package project

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

func (p *Project) BuildAndroid(ndkHome, javaHome string, tags []string) error {
	defer tracing.NewRegion("Project.BuildAndroid").End()
	if err := p.Package(); err != nil {
		return err
	}
	if err := p.copyAndroidProjectTemplate(); err != nil {
		return err
	}
	if err := p.buildKaijuAndroidLibrary(ndkHome, tags); err != nil {
		return err
	}
	if err := p.copyAndroidContentToAssets(); err != nil {
		return err
	}
	if err := p.buildAPK(javaHome, tags); err != nil {
		return err
	}
	slog.Info("completed full Android build")
	return nil
}

func (p *Project) copyAndroidProjectTemplate() error {
	defer tracing.NewRegion("Project.copyAndroidProjectTemplate").End()
	if !p.fileSystem.Exists(project_file_system.ProjectBuildAndroidFolder) {
		slog.Info("the Android project was not found, creating it",
			"folder", project_file_system.ProjectBuildAndroidFolder)
		err := project_file_system.EngineFS.CopyFolder(&p.fileSystem,
			"editor/project_templates/android",
			project_file_system.ProjectBuildAndroidFolder, []string{})
		if err != nil {
			slog.Error("failed to copy the android template project to build folder")
			return err
		}
		slog.Info("project template copy complete")
	}
	return nil
}

func (p *Project) buildKaijuAndroidLibrary(ndkHome string, tags []string) error {
	defer tracing.NewRegion("Project.buildKaijuAndroidLibrary").End()
	if ndkHome == "" {
		return errors.New("the NDK folder path hasn't yet been setup in the editor settings")
	}
	if s, err := os.Stat(ndkHome); err != nil || !s.IsDir() {
		return errors.New("the NDK folder specified in the editor settings is not valid")
	}
	arch := "arm64-v8a"
	plat := ""
	switch runtime.GOOS {
	case "windows":
		plat = "windows-x86_64"
	default:
		return errors.New("the NDK platform for this platform hasn't yet been set, this requires an engine code edit")
	}
	slog.Info("building Android native libraries")
	androidPath := p.fileSystem.FullPath(project_file_system.ProjectBuildAndroidFolder)
	outPath := filepath.Join(androidPath, "app/src/main/jniLibs/", arch)
	args := []string{
		"build",
		"-buildmode=c-shared",
	}
	if len(tags) > 0 {
		args = append(args, fmt.Sprintf("-tags=%s", strings.Join(tags, ",")))
	}
	args = append(args, "-o", filepath.Join(outPath, "/libkaiju_android.so"), ".")
	cmd := exec.Command("go", args...)
	env := os.Environ()
	env = append(env,
		"CC="+filepath.Join(ndkHome, "toolchains/llvm/prebuilt", plat, "bin/aarch64-linux-android21-clang"),
		"CXX="+filepath.Join(ndkHome, "toolchains/llvm/prebuilt", plat, "bin/aarch64-linux-android21-clang++"),
		"AR="+filepath.Join(ndkHome, "toolchains/llvm/prebuilt", plat, "bin/llvm-ar"),
		"RANLIB="+filepath.Join(ndkHome, "toolchains/llvm/prebuilt", plat, "bin/llvm-ranlib"),
		"GOOS=android",
		"CGO_ENABLED=1",
		"GOARCH=arm64",
		"CGO_CFLAGS=-D__android__=1 -I"+ndkHome+"/sources/android/native_app_glue",
		"CGO_LDFLAGS=-landroid",
	)
	cmd.Env = env
	cmd.Dir = p.fileSystem.FullPath(project_file_system.ProjectCodeFolder)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer out.Close()
	scanner := bufio.NewScanner(out)
	slog.Info("starting build of Go code for Android")
	if err = cmd.Start(); err != nil {
		return err
	}
	for scanner.Scan() {
		slog.Info(scanner.Text())
	}
	if err := cmd.Wait(); err != nil {
		slog.Error("failed to build the Go code", "message", string(stderr.String()), "error", err)
		return err
	}
	const hFile = "libkaiju_android.h"
	h := filepath.Join(outPath, hFile)
	if _, err := os.Stat(h); err == nil {
		if err := os.Rename(h, filepath.Join(outPath, "../../cpp", hFile)); err != nil {
			slog.Error("failed to copy the header file into the cpp folder")
			return err
		}
	}
	return nil
}

func (p *Project) copyAndroidContentToAssets() error {
	defer tracing.NewRegion("Project.copyAndroidContentToAssets").End()
	slog.Info("copying content to android assets")
	from := p.packagePath()
	toDir := filepath.Join(
		p.fileSystem.FullPath(project_file_system.ProjectBuildAndroidFolder),
		"app/src/main/assets")
	if s, err := os.Stat(toDir); err != nil {
		if err := os.Mkdir(toDir, os.ModePerm); err != nil {
			return err
		}
	} else if !s.IsDir() {
		return fmt.Errorf("the asset path is not a folder: %s", toDir)
	}
	to := filepath.Join(toDir, filepath.Base(from))
	return filesystem.CopyFile(from, to)
}

func (p *Project) buildAPK(javaHome string, tags []string) error {
	defer tracing.NewRegion("Project.buildAPK").End()
	slog.Info("building android APK")
	gradle := filepath.Join(project_file_system.ProjectBuildAndroidFolder, "/gradlew")
	if runtime.GOOS == "windows" {
		gradle += ".bat"
	}
	gradle = p.fileSystem.FullPath(gradle)
	var cmd *exec.Cmd
	if slices.Contains(tags, "debug") {
		cmd = exec.Command(gradle, "assembleDebug")
	} else {
		cmd = exec.Command(gradle, "assembleRelease")
	}
	cmd.Dir = filepath.Dir(gradle)
	cmd.Env = os.Environ()
	if os.Getenv("JAVA_HOME") == "" {
		if javaHome == "" {
			return errors.New("the JAVA_HOME folder path hasn't yet been setup in the editor settings")
		} else {
			cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_HOME=%s", javaHome))
		}
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("failed to get stdout pipe for Gradle", "error", err)
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		slog.Error("failed to get stderr pipe for Gradle", "error", err)
		return err
	}
	if err = cmd.Start(); err != nil {
		slog.Error("failed to start Gradle build", "error", err)
		return err
	}
	scanAndLog := func(pipe io.Reader, level string) {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			text := scanner.Text()
			switch level {
			case "info":
				slog.Info(text)
			case "error":
				slog.Error(text)
			}
		}
	}
	// goroutine
	go scanAndLog(stdoutPipe, "info")
	// goroutine
	go scanAndLog(stderrPipe, "error")
	if err = cmd.Wait(); err != nil {
		slog.Error("Gradle build failed", "error", err)
		return err
	}
	slog.Info("Gradle build completed successfully")
	outFolder := filepath.Join(cmd.Dir, "app/build/outputs/apk")
	if slices.Contains(tags, "debug") {
		outFolder = filepath.Join(outFolder, "debug")
	} else {
		outFolder = filepath.Join(outFolder, "release")
	}
	filesystem.OpenFileBrowserToFolder(outFolder)
	return nil
}

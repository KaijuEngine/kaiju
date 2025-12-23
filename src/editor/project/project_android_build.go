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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/assets/content_archive"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
)

func (p *Project) BuildRunAndroid(reader content_archive.FileReader, ndkHome, javaHome string, tags []string) error {
	if err := p.BuildAndroid(reader, ndkHome, javaHome, tags); err != nil {
		return err
	}
	if err := p.deployAndroidAPK(ndkHome, tags); err != nil {
		return err
	}
	slog.Info("application has been deployed on the target Android device")
	return nil
}

func (p *Project) BuildAndroid(reader content_archive.FileReader, ndkHome, javaHome string, tags []string) error {
	defer tracing.NewRegion("Project.BuildAndroid").End()
	sdkHome := filepath.Join(ndkHome, "../../")
	if err := p.Package(reader); err != nil {
		return err
	}
	if err := p.copyAndroidProjectTemplate(); err != nil {
		return err
	}
	if err := p.cleanKaijuAndroidLibrary(sdkHome, javaHome); err != nil {
		return err
	}
	if err := p.buildKaijuAndroidLibrary(sdkHome, ndkHome, tags); err != nil {
		return err
	}
	if err := p.copyAndroidContentToAssets(); err != nil {
		return err
	}
	if err := p.buildAPK(sdkHome, javaHome, tags); err != nil {
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
	return p.updateAndroidProjectStrings()
}

func (p *Project) cleanKaijuAndroidLibrary(sdkHome, javaHome string) error {
	defer tracing.NewRegion("Project.cleanKaijuAndroidLibrary").End()
	gradle := filepath.Join(project_file_system.ProjectBuildAndroidFolder, "/gradlew")
	if runtime.GOOS == "windows" {
		gradle += ".bat"
	}
	gradle = p.fileSystem.FullPath(gradle)
	cmd := exec.Command(gradle, "clean")
	cmd.Dir = filepath.Dir(gradle)
	cmd.Env = os.Environ()
	if os.Getenv("JAVA_HOME") == "" {
		if javaHome == "" {
			return errors.New("the JAVA_HOME folder path hasn't yet been setup in the editor settings")
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_HOME=%s", javaHome))
	}
	if os.Getenv("ANDROID_HOME") == "" {
		if sdkHome == "" {
			return errors.New("the ANDROID_HOME folder path hasn't yet been setup in the editor settings")
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("ANDROID_HOME=%s", sdkHome))
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("failed to get stdout pipe for Gradle clean", "error", err)
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		slog.Error("failed to get stderr pipe for Gradle clean", "error", err)
		return err
	}
	if err = cmd.Start(); err != nil {
		slog.Error("failed to start Gradle clean", "error", err)
		return err
	}
	scanAndLog := func(pipe io.Reader, level string) {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			text := scanner.Text()
			if level == "info" {
				slog.Info(text)
			} else {
				slog.Error(text)
			}
		}
	}
	go scanAndLog(stdoutPipe, "info")
	go scanAndLog(stderrPipe, "error")
	if err = cmd.Wait(); err != nil {
		slog.Error("Gradle clean failed", "error", err)
		return err
	}
	slog.Info("Gradle clean completed successfully")
	return nil
}

func (p *Project) buildKaijuAndroidLibrary(sdkHome, ndkHome string, tags []string) error {
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
		"ANDROID_HOME="+sdkHome,
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
	return filesystem.CopyFileOverwrite(from, to)
}

func (p *Project) buildAPK(sdkHome, javaHome string, tags []string) error {
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
	if os.Getenv("ANDROID_HOME") == "" {
		if sdkHome == "" {
			return errors.New("the ANDROID_HOME folder path hasn't yet been setup in the editor settings")
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("ANDROID_HOME=%s", sdkHome))
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

func (p *Project) deployAndroidAPK(ndkHome string, tags []string) error {
	adb := filepath.Join(ndkHome, "../../platform-tools/adb")
	if runtime.GOOS == "windows" {
		adb += ".exe"
	}
	outFolder := filepath.Join(p.fileSystem.FullPath(
		project_file_system.ProjectBuildAndroidFolder),
		"app/build/outputs/apk")
	if slices.Contains(tags, "debug") {
		outFolder = filepath.Join(outFolder, "debug")
	} else {
		outFolder = filepath.Join(outFolder, "release")
	}
	apkPath := ""
	entries, err := os.ReadDir(outFolder)
	if err != nil {
		slog.Error("failed to read APK output folder", "folder", outFolder, "error", err)
		return err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".apk") {
			apkPath = filepath.Join(outFolder, e.Name())
			break
		}
	}
	if apkPath == "" {
		return errors.New("no APK file found in " + outFolder)
	}
	manifestPath := filepath.Join(p.fileSystem.FullPath(
		project_file_system.ProjectBuildAndroidFolder),
		"app/src/main/AndroidManifest.xml")
	packageName, err := extractAndroidPackageName(manifestPath)
	if err != nil {
		slog.Error("failed to extract package name from manifest", "manifest", manifestPath, "error", err)
		return err
	}
	installCmd := exec.Command(adb, "install", "-r", apkPath)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	slog.Info("installing APK via ADB", "apk", apkPath)
	if err = installCmd.Run(); err != nil {
		slog.Error("ADB install failed", "error", err)
		return err
	}
	launchCmd := exec.Command(adb, "shell", "monkey", "-p", packageName,
		"-c", "android.intent.category.LAUNCHER", "1")
	launchCmd.Stdout = os.Stdout
	launchCmd.Stderr = os.Stderr
	slog.Info("launching app via ADB", "package", packageName)
	if err = launchCmd.Run(); err != nil {
		slog.Error("ADB launch failed", "error", err)
		return err
	}
	return nil
}

func extractAndroidPackageName(manifestPath string) (string, error) {
	type manifest struct {
		XMLName xml.Name `xml:"manifest"`
		Package string   `xml:"package,attr"`
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", err
	}
	var m manifest
	if err = xml.Unmarshal(data, &m); err == nil && m.Package != "" {
		return m.Package, nil
	}
	gradlePath := filepath.Join(filepath.Dir(filepath.Dir(manifestPath)), "build.gradle")
	if _, err := os.Stat(gradlePath); err == nil {
		if id, err := parseAndroidAppIdFromGradle(gradlePath); err == nil && id != "" {
			return id, nil
		}
	}
	type activity struct {
		Name string `xml:"name,attr"`
	}
	type app struct {
		Activities []activity `xml:"application>activity"`
	}
	var a app
	if err = xml.Unmarshal(data, &a); err == nil && len(a.Activities) > 0 {
		if fullName := a.Activities[0].Name; fullName != "" {
			if idx := strings.LastIndex(fullName, "."); idx > 0 {
				return fullName[:idx], nil
			}
		}
	}
	return "", errors.New("package name not found in manifest, Gradle, or activity")
}

func parseAndroidAppIdFromGradle(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// matches: applicationId "com.example.app"
	re := regexp.MustCompile(`applicationId\s+["']([^"']+)["']`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) == 2 {
		return matches[1], nil
	}
	return "", errors.New("applicationId not found in Gradle file")
}

func (p *Project) updateAndroidProjectStrings() error {
	if err := p.updateAndroidSettingsGradleKTS(); err != nil {
		return err
	}
	if err := p.updateAndroidAppBuildGradleKTS(); err != nil {
		return err
	}
	if err := p.updateAndroidAppSrcMainAndroidManifestXML(); err != nil {
		return err
	}
	if err := p.updateAndroidAppSrcMainJavaComKaijuengineKaijuengine(); err != nil {
		return err
	}
	if err := p.updateAndroidAppSrcMainResValuesStringsXML(); err != nil {
		return err
	}
	return nil
}

func (p *Project) updateAndroidSettingsGradleKTS() error {
	// Set rootProject.name to p.settings.Android.RootProjectName in settings.gradle.kts
	settingsPath := filepath.Join(
		p.fileSystem.FullPath(project_file_system.ProjectBuildAndroidFolder),
		"settings.gradle.kts",
	)
	stat, err := os.Stat(settingsPath)
	if err != nil {
		slog.Error("the settings.gradle.kts doesn't exist", "path", settingsPath, "error", err)
		return err
	}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		slog.Error("failed to read settings.gradle.kts", "path", settingsPath, "error", err)
		// If the file cannot be read we simply return – the rest of the TODOs can be
		// handled later.
		return err
	}
	newRootName := fmt.Sprintf(`rootProject.name = "%s"`, p.settings.Android.RootProjectName)
	re := regexp.MustCompile(`(?m)^(\s*)rootProject\.name\s*=.*$`)
	if re.Match(data) {
		data = re.ReplaceAllFunc(data, func(m []byte) []byte {
			submatches := re.FindSubmatch(m)
			return []byte(fmt.Sprintf(`%s%s`, submatches[1], newRootName))
		})
	} else {
		// If the line is not present, prepend it to the file
		data = append([]byte(newRootName+"\n"), data...)
	}
	if err = os.WriteFile(settingsPath, data, stat.Mode().Perm()); err != nil {
		slog.Error("failed to write updated settings.gradle.kts", "path", settingsPath, "error", err)
		return err
	}
	return nil
}

func (p *Project) updateAndroidAppBuildGradleKTS() error {
	if strings.TrimSpace(p.settings.Android.ApplicationId) == "" {
		slog.Warn("the ApplicationId was not set in the project settings")
		return nil
	}
	// Set namespace to p.settings.Android.ApplicationId in app/build.gradle.kts
	// Set applicationId to p.settings.Android.ApplicationId in app/build.gradle.kts
	gradlePath := filepath.Join(
		p.fileSystem.FullPath(project_file_system.ProjectBuildAndroidFolder),
		"app/build.gradle.kts",
	)
	stat, err := os.Stat(gradlePath)
	if err != nil {
		slog.Error("build.gradle.kts not found", "path", gradlePath, "error", err)
		return err
	}
	data, err := os.ReadFile(gradlePath)
	if err != nil {
		slog.Error("failed to read build.gradle.kts", "path", gradlePath, "error", err)
		return err
	}
	appID := p.settings.Android.ApplicationId
	// Update the namespace
	nsRe := regexp.MustCompile(`(?m)^(\s*)namespace\s*=.*$`)
	newNS := fmt.Sprintf(`namespace = "%s"`, appID)
	if nsRe.Match(data) {
		data = nsRe.ReplaceAllFunc(data, func(m []byte) []byte {
			submatches := nsRe.FindSubmatch(m)
			return []byte(fmt.Sprintf(`%s%s`, submatches[1], newNS))
		})
	} else {
		data = append([]byte(newNS+"\n"), data...)
	}
	// Update the applicationId
	appIdRe := regexp.MustCompile(`(?m)^(\s*)applicationId\s*=.*$`)
	newAppId := fmt.Sprintf(`applicationId = "%s"`, appID)
	if appIdRe.Match(data) {
		data = appIdRe.ReplaceAllFunc(data, func(m []byte) []byte {
			submatches := appIdRe.FindSubmatch(m)
			return []byte(fmt.Sprintf(`%s%s`, submatches[1], newAppId))
		})
	} else {
		slog.Error("failed to find the applicationId in build.gradle.kts", "path", gradlePath)
		return nil
	}
	if err = os.WriteFile(gradlePath, data, stat.Mode().Perm()); err != nil {
		slog.Error("failed to write updated build.gradle.kts", "path", gradlePath, "error", err)
		return err
	}
	return nil
}

func (p *Project) updateAndroidAppSrcMainAndroidManifestXML() error {
	if strings.TrimSpace(p.settings.Android.ApplicationId) == "" {
		slog.Warn("the ApplicationId was not set in the project settings")
		return nil
	}
	// Set android:name to p.settings.Android.ApplicationId+".MainActivity" in app/src/main/AndroidManifest.xml
	manifestPath := filepath.Join(
		p.fileSystem.FullPath(project_file_system.ProjectBuildAndroidFolder),
		"app/src/main/AndroidManifest.xml",
	)
	stat, err := os.Stat(manifestPath)
	if err != nil {
		slog.Error("the AndroidManifest.xml doesn't exist", "path", manifestPath, "error", err)
		return err
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		slog.Error("failed to read AndroidManifest.xml", "path", manifestPath, "error", err)
		return err
	}
	re := regexp.MustCompile(`android:name\s*=\s*"[^"]*.MainActivity"`)
	newName := fmt.Sprintf(`android:name="%s.MainActivity"`, p.settings.Android.ApplicationId)
	if re.Match(data) {
		data = re.ReplaceAll(data, []byte(newName))
	} else {
		slog.Error("android:name attribute not found and no <activity> tag to insert into", "path", manifestPath)
		return errors.New("android:name attribute not found in AndroidManifest.xml")
	}
	if err = os.WriteFile(manifestPath, data, stat.Mode().Perm()); err != nil {
		slog.Error("failed to write updated AndroidManifest.xml", "path", manifestPath, "error", err)
		return err
	}
	return nil
}

func (p *Project) updateAndroidAppSrcMainJavaComKaijuengineKaijuengine() error {
	if strings.TrimSpace(p.settings.Android.ApplicationId) == "" {
		slog.Warn("the ApplicationId was not set in the project settings")
		return nil
	}
	// Set android:name to p.settings.Android.ApplicationId+".MainActivity" in app/src/main/java/com/kaijuengine/kaijuengine/MainActivity.java
	manifestPath := filepath.Join(
		p.fileSystem.FullPath(project_file_system.ProjectBuildAndroidFolder),
		"app/src/main/java/com/kaijuengine/kaijuengine/MainActivity.java",
	)
	stat, err := os.Stat(manifestPath)
	if err != nil {
		slog.Error("the MainActivity.java doesn't exist", "path", manifestPath, "error", err)
		return err
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		slog.Error("failed to read MainActivity.java", "path", manifestPath, "error", err)
		return err
	}
	re := regexp.MustCompile(`package ([\w\.]+);`)
	newName := fmt.Sprintf(`package %s;`, p.settings.Android.ApplicationId)
	if re.Match(data) {
		data = re.ReplaceAll(data, []byte(newName))
	} else {
		slog.Error("package line not found", "path", manifestPath)
		return errors.New("package line not found in MainActivity.java")
	}
	if err = os.WriteFile(manifestPath, data, stat.Mode().Perm()); err != nil {
		slog.Error("failed to write updated MainActivity.java", "path", manifestPath, "error", err)
		return err
	}
	return nil
}

func (p *Project) updateAndroidAppSrcMainResValuesStringsXML() error {
	// Set app_name to p.settings.Android.RootProjectName in app/src/main/res/values/strings.xml
	stringsPath := filepath.Join(
		p.fileSystem.FullPath(project_file_system.ProjectBuildAndroidFolder),
		"app/src/main/res/values/strings.xml",
	)
	stat, err := os.Stat(stringsPath)
	if err != nil {
		slog.Error("the strings.xml doesn't exist", "path", stringsPath, "error", err)
		return err
	}
	data, err := os.ReadFile(stringsPath)
	if err != nil {
		slog.Error("failed to read strings.xml", "path", stringsPath, "error", err)
		return err
	}
	re := regexp.MustCompile(`(?s)<string\s+name\s*=\s*"app_name"\s*>.*?</string>`)
	newEntry := fmt.Sprintf(`<string name="app_name">%s</string>`, p.settings.Android.RootProjectName)
	if re.Match(data) {
		data = re.ReplaceAll(data, []byte(newEntry))
	} else {
		slog.Error(`failed to find the "app_name" in strings.xml`, "path", stringsPath)
		return nil
	}
	if err = os.WriteFile(stringsPath, data, stat.Mode().Perm()); err != nil {
		slog.Error("failed to write updated strings.xml", "path", stringsPath, "error", err)
		return err
	}
	return nil
}

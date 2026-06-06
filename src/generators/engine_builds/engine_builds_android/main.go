/******************************************************************************/
/* build_android.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"kaijuengine.com/klib"
)

func main() {
	// TODO:  Pull the architecture
	arch := "arm64-v8a"
	ndk := os.Getenv("NDK_HOME")
	// TODO:  Try to locate where NDK_HOME is, the usual folders per-paltform
	if ndk == "" {
		panic("NDK_HOME not set in environment")
	}
	plat := ""
	if runtime.GOOS == "windows" {
		plat = "windows-x86_64"
	}
	if plat == "" {
		panic("build platform not yet specified")
	}
	outPath := filepath.Join("editor/project_templates/android/app/src/main/jniLibs/", arch)
	args := []string{
		"build",
		"-buildmode=c-shared",
		"-tags=debug,editor,filedrop",
		"-o", filepath.Join(outPath, "/libkaiju_android.so"),
		".",
	}
	cmd := exec.Command("go", args...)
	env := os.Environ()
	env = append(env,
		"CC="+filepath.Join(ndk, "toolchains/llvm/prebuilt", plat, "bin/aarch64-linux-android21-clang"),
		"CXX="+filepath.Join(ndk, "toolchains/llvm/prebuilt", plat, "bin/aarch64-linux-android21-clang++"),
		"AR="+filepath.Join(ndk, "toolchains/llvm/prebuilt", plat, "bin/llvm-ar"),
		"RANLIB="+filepath.Join(ndk, "toolchains/llvm/prebuilt", plat, "bin/llvm-ranlib"),
		"GOOS=android",
		"CGO_ENABLED=1",
		"GOARCH=arm64",
		"CGO_CFLAGS=-D__android__=1 -I"+ndk+"/sources/android/native_app_glue",
		"CGO_LDFLAGS=-landroid",
	)
	cmd.Env = env
	if _, err := os.Stat("src"); err == nil {
		cmd.Dir = "src"
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out := klib.MustReturn(cmd.StdoutPipe())
	scanner := bufio.NewScanner(out)
	println("CWD:", klib.MustReturn(filepath.Abs(cmd.Dir)))
	println("Running:", cmd.String())
	klib.Must(cmd.Start())
	for scanner.Scan() {
		println(scanner.Text())
	}
	if err := cmd.Wait(); err != nil {
		println(string(stderr.String()))
		log.Fatal(err)
	}
	const hFile = "libkaiju_android.h"
	outPath = filepath.Join(cmd.Dir, outPath)
	h := filepath.Join(outPath, hFile)
	if _, err := os.Stat(h); err == nil {
		if err := os.Rename(h, filepath.Join(outPath, "../../cpp", hFile)); err != nil {
			log.Fatal(err)
		}
	}
	println("Android compile successful")
}

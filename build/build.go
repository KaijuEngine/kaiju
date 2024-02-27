package main

import (
	"flag"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	// TODO:  Run the tests when we have some: go test -timeout 30s -v ./...

	fs := flag.NewFlagSet("Kaiju Build Args", flag.ContinueOnError)
	isEditor := fs.Bool("editor", false, "Builds the editor, otherwise builds the runtime")
	renderer := fs.String("renderer", "", "vk (Vulkan default)")
	fs.Parse(os.Args[1:])
	tags := []string{}       // tags
	cgoLDFLAGS := []string{} // CGO_LDFLAGS
	cgoCFLAGS := []string{}  // CGO_CFLAGS
	goOS := runtime.GOOS
	goArch := runtime.GOARCH
	cgoEnabled := true
	outExtension := ""
	if runtime.GOOS == "windows" {
		outExtension = ".exe"
		cgoLDFLAGS = append(cgoLDFLAGS, "-lgdi32", "-lXInput")
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cgoLDFLAGS = append(cgoLDFLAGS, "-lX11")
	}
	if *isEditor {
		tags = append(tags, "editor")
	}
	switch *renderer {
	case "vk":
		fallthrough
	default:
	}
	args := []string{
		"build",
		"-tags", strings.Join(tags, ","),
		"-o", "./bin/kaiju" + outExtension,
		`-ldflags="-s -w"`,
		"./src/main.go",
	}
	cmd := exec.Command("go", args...)
	if !cgoEnabled {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	}
	cmd.Env = append(cmd.Env, "GOOS="+goOS, "GOARCH="+goArch)
	cmd.Env = append(cmd.Env, "CGO_LDFLAGS="+strings.Join(cgoLDFLAGS, " "))
	cmd.Env = append(cmd.Env, "CGO_CFLAGS="+strings.Join(cgoCFLAGS, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}

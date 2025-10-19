/******************************************************************************/
/* main.go                                                                    */
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

package main

import (
	"bufio"
	"flag"
	"kaiju/klib"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func compile(args ...string) error {
	cmd := exec.Command("glslc", args...)
	outPipe := klib.MustReturn(cmd.StderrPipe())
	scanner := bufio.NewScanner(outPipe)
	err := cmd.Start()
	if err != nil {
		vp := os.Getenv("VK_SDK_PATH")
		if vp != "" {
			cmd = exec.Command(filepath.Join(vp, "Bin", "glslc"), args...)
			outPipe = klib.MustReturn(cmd.StdoutPipe())
			scanner = bufio.NewScanner(outPipe)
			err = cmd.Start()
		}
		if err != nil {
			panic("Failed to run glslc, make sure you have the Vulkan 'Bin' folder in your environment path")
		}
	}
	for scanner.Scan() {
		println(scanner.Text())
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	println("Compiled " + args[2])
	return nil
}

func hasOIT(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	src := string(data)
	return strings.Contains(src, `"inc_fragment_oit_block.inl"`) ||
		strings.Contains(src, "#ifdef OIT")
}

func main() {
	fs := flag.NewFlagSet("Kaiju Spir-V compile", flag.ContinueOnError)
	dbg := fs.Bool("d", false, "Compile the shader for debugging")
	out := fs.String("o", "", "The output path for the compiled shader")
	in := fs.String("i", "", "The path of the shader to be compiled")
	fs.Parse(os.Args[1:])
	if *in == "" {
		panic("Expected -i=... input, run with arg -h for help")
	}
	outName := *out
	if outName == "" {
		outName = filepath.Dir(*in)
	}
	if !strings.HasSuffix(*out, ".spv") {
		outName = filepath.Join(*out, filepath.Base(*in)+".spv")
	}
	args := []string{*in,
		"-o", outName,
	}
	if *dbg {
		args = append(args, "-g")
	}
	err := compile(args...)
	if err == nil && hasOIT(*in) {
		args[2] = strings.TrimSuffix(args[2], ".spv") + ".oit.spv"
		args = append(args, "-DOIT")
		err = compile(args...)
	}
	if err != nil {
		println("Exiting due to compile error")
	}
}

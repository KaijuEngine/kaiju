/******************************************************************************/
/* pprof.go                                                                   */
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

package profiler

import (
	"kaiju/build"
	"kaiju/klib/contexts"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
)

const (
	pprofProcName = "pprof"
)

var runningPprof *exec.Cmd = nil

func StartPGOProfiler() error {
	pprofFile, err := os.Create(pprofMergeFile)
	if err != nil {
		return err
	}
	return pprof.StartCPUProfile(pprofFile)
}

func StopPGOProfiler() {
	pprof.StopCPUProfile()
}

func StartDefaultProfiler() error {
	if build.Debug {
		pprofFile, err := os.Create(pprofCPUFile)
		if err != nil {
			return err
		}
		pprof.StartCPUProfile(pprofFile)
		return nil
	} else {
		return nil
	}
}

func StopDefaultProfiler() {
	if build.Debug {
		pprof.StopCPUProfile()
	}
}

func ShowDefaultProfilerInWeb() error {
	if build.Debug {
		if err := CleanupProfiler(); err != nil {
			return err
		}
		ctx := contexts.NewCancellable()
		runningPprof = exec.CommandContext(ctx, "go", []string{"tool", pprofProcName, "-http=:" + pprofWebPort, pprofCPUFile}...)
		return runningPprof.Start()
	} else {
		return nil
	}
}

func ProfileCallWithDefaultProfiler(call func()) error {
	if build.Debug {
		if err := StartDefaultProfiler(); err != nil {
			return err
		}
		call()
		StopDefaultProfiler()
		return ShowDefaultProfilerInWeb()
	} else {
		return nil
	}
}

func CleanupProfiler() error {
	if build.Debug {
		if runningPprof != nil {
			runningPprof.Process.Kill()
			if runtime.GOOS == "windows" {
				return exec.Command("taskkill", "/F", "/IM", pprofProcName+".exe").Run()
			} else {
				return exec.Command("pkill", pprofProcName).Run()
			}
		}
	}
	return nil
}

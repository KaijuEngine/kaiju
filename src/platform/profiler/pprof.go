/******************************************************************************/
/* pprof.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package profiler

import (
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"

	"kaijuengine.com/build"
	"kaijuengine.com/klib/contexts"
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

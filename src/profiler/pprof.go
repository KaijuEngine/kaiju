package profiler

import (
	"os"
	"runtime/pprof"
)

func StartDefaultProfiler() error {
	pprofFile, err := os.Create(pprofCPUFile)
	if err != nil {
		return err
	}
	pprof.StartCPUProfile(pprofFile)
	return nil
}

func StopDefaultProfiler() {
	pprof.StopCPUProfile()
}

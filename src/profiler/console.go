package profiler

import (
	"bufio"
	"fmt"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/systems/console"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
)

const (
	pprofCPU  = "cpu.prof"
	pprofHeap = "heap.prof"
)

func consoleTop(host *engine.Host) string {
	cmd := exec.Command("go", "tool", "pprof", "-top", pprofCPU)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out := klib.MustReturn(cmd.StdoutPipe())
	scanner := bufio.NewScanner(out)
	err := cmd.Start()
	if err != nil {
		return err.Error()
	}
	sb := strings.Builder{}
	for scanner.Scan() {
		sb.WriteString(scanner.Text() + "\n")
	}
	return sb.String()
}

func SetupConsole(host *engine.Host) {
	var pprofFile *os.File = nil
	console.For(host).AddCommand("pprof", func(arg string) string {
		if arg == "start" {
			pprofFile = klib.MustReturn(os.Create(pprofCPU))
			pprof.StartCPUProfile(pprofFile)
			return "CPU profile started"
		} else if arg == "stop" {
			if pprofFile == nil {
				return "CPU profile not yet started"
			}
			pprof.StopCPUProfile()
			pprofFile.Close()
			return "CPU profile written to " + pprofCPU
		} else if arg == "heap" {
			hp := klib.MustReturn(os.Create(pprofHeap))
			pprof.WriteHeapProfile(hp)
			hp.Close()
			return "Heap profile written to " + pprofHeap
		} else if arg == "top" {
			return consoleTop(host)
		} else {
			return ""
		}
	})
	console.For(host).AddCommand("GC", func(string) string {
		runtime.GC()
		return "Garbage collection done"
	})
	console.For(host).AddCommand("MemStats", func(string) string {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		return fmt.Sprintf("Alloc: %d, TotalAlloc: %d, Sys: %d, NumGC: %d", mem.Alloc, mem.TotalAlloc, mem.Sys, mem.NumGC)
	})
}

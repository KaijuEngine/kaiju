package profiler

import (
	"bufio"
	"fmt"
	"kaiju/contexts"
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
	pprofCPU     = "cpu.prof"
	pprofHeap    = "heap.prof"
	pprofMerge   = "default.pgo"
	pprofWebPort = "9382"
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

func consoleMerge(host *engine.Host, argStr string) string {
	// First arg in split will be "merge" and can be skipped
	args := strings.Split(argStr, " ")[1:]
	cmdArgs := make([]string, 0, len(args)+5)
	cmdArgs = append(cmdArgs, "tool", "pprof", "-proto")
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, ">", pprofMerge)
	cmd := exec.Command("go", cmdArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Start()
	if err != nil {
		return err.Error()
	}
	cmd.Wait()
	return "Files merged into " + pprofMerge
}

func launchWeb(host *engine.Host) (*contexts.Cancellable, error) {
	ctx := contexts.NewCancellable()
	cmd := exec.CommandContext(ctx, "go", "tool", "pprof", "-http=:"+pprofWebPort, pprofCPU)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	go func() {
		console.For(host).Write("Starting server on localhost:" + pprofWebPort)
		<-ctx.Done()
		cmd.Process.Kill()
		if ctx.Err() == nil {
			console.For(host).Write("Failed to start web server, make sure you have go and graphviz installed.")
			ctx.Cancel()
		}
	}()
	return ctx, err
}

func SetupConsole(host *engine.Host) {
	var pprofFile *os.File = nil
	var ctx *contexts.Cancellable = nil
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
		} else if arg == "web" {
			var err error
			if ctx != nil {
				ctx.Cancel()
				ctx = nil
			}
			if ctx, err = launchWeb(host); err != nil {
				return err.Error()
			} else {
				return "Starting up web server..."
			}
		} else if arg == "webstop" {
			if ctx != nil {
				ctx.Cancel()
				ctx = nil
				return "Web server stopped"
			} else {
				return "Web server not running"
			}
		} else if strings.HasPrefix(arg, "merge") {
			return consoleMerge(host, arg)
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

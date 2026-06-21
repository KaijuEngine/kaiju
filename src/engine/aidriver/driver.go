/******************************************************************************/
/* driver.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

// Package aidriver exposes a localhost HTTP control surface that lets an
// external process (e.g. an AI agent driving the game through curl) capture a
// screenshot of the rendered frame and inject mouse/keyboard input.
//
// The package is a plain library: it is only linked into a binary that is
// wired up from the bootstrap layer under the `ai_driver` build tag, so default
// builds carry none of this code. See bootstrap/bootstrap_aidriver*.go.
//
// Threading model: the HTTP server runs on its own goroutine and only ever
// enqueues commands onto a buffered channel. Every command is executed on the
// game-loop thread by a drain that re-registers itself each frame via
// host.RunNextFrame, running right after Window.Poll and before the UI update —
// the same point at which real OS input is applied. This keeps all engine state
// mutation single-threaded with no locks.
package aidriver

import (
	"kaijuengine.com/engine"
)

// command is one unit of work executed on the game-loop thread by the drain.
// Implementations carry their own reply channel so the originating HTTP handler
// can block until the result is ready.
type command interface {
	apply(d *driver)
}

type driver struct {
	host    *engine.Host
	queue   chan command
	closing chan struct{}
	// focused mirrors the window's activation state. It is written only from
	// OnActivate/OnDeactivate (game-loop thread) and read only by stateCommand
	// (also game-loop thread), so it needs no synchronization.
	focused bool
}

const queueDepth = 256

func newDriver(host *engine.Host) *driver {
	return &driver{
		host:    host,
		queue:   make(chan command, queueDepth),
		closing: make(chan struct{}),
		focused: true,
	}
}

// enqueue offers a command to the drain without blocking. It returns false when
// the queue is saturated so the HTTP layer can answer with backpressure rather
// than stalling the server goroutine.
func (d *driver) enqueue(cmd command) bool {
	select {
	case d.queue <- cmd:
		return true
	default:
		return false
	}
}

// drainOnce runs on the game-loop thread. It executes every queued command for
// this frame, then re-arms itself for the next frame. The closing channel stops
// the re-registration during teardown.
func (d *driver) drainOnce() {
	for {
		select {
		case <-d.closing:
			return
		case cmd := <-d.queue:
			cmd.apply(d)
		default:
			d.host.RunNextFrame(d.drainOnce)
			return
		}
	}
}

// framebufferSize returns the swap-chain pixel dimensions (what a screenshot
// measures). It falls back to the logical window size when no renderer is
// available, which keeps coordinate handling sane in headless scenarios. It
// reads a CPU-side field and is only called on the game-loop thread.
func (d *driver) framebufferSize(winW, winH int) (int, int) {
	fbW, fbH := winW, winH
	if w := d.host.Window; w != nil {
		if gi := w.GpuInstance; gi != nil && gi.IsValid() {
			ext := gi.PrimaryDevice().LogicalDevice.SwapChain.Extent
			if ext.X() > 0 && ext.Y() > 0 {
				fbW, fbH = int(ext.X()), int(ext.Y())
			}
		}
	}
	return fbW, fbH
}

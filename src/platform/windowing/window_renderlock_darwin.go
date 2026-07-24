//go:build darwin && !ios

/******************************************************************************/
/* window_renderlock_darwin.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package windowing

/*
#include "cocoa_window.h"
*/
import "C"

// RenderLock / RenderUnlock serialize the render goroutine's GPU work against
// AppKit's main-thread resize of the window's CAMetalLayer. MoltenVK and
// CAMetalLayer are not safe for that concurrent access, so the render loop
// brackets a frame (acquire/submit/present) with these and the MetalView brackets
// its layer geometry change with the same underlying dispatch_semaphore (no thread
// ownership, so it stays correct even if the goroutine migrates OS threads between
// the paired cgo calls).
func (w *Window) RenderLock()   { C.cocoa_render_lock(w.handle) }
func (w *Window) RenderUnlock() { C.cocoa_render_unlock(w.handle) }

// IsInLiveResize reports whether AppKit is mid live (interactive) window resize. The
// render loop skips rendering while true so it never touches the CAMetalLayer
// concurrently with AppKit's main-thread resize of it.
func (w *Window) IsInLiveResize() bool { return C.cocoa_in_live_resize() != 0 }

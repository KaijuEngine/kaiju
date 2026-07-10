//go:build !darwin || ios

/******************************************************************************/
/* window_renderlock_other.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package windowing

// RenderLock / RenderUnlock are a no-op off macOS. The render/resize
// serialization they provide is only needed for AppKit + CAMetalLayer + MoltenVK;
// other backends do not resize the rendering surface from a separate thread while
// the render loop runs.
func (w *Window) RenderLock()   {}
func (w *Window) RenderUnlock() {}

// IsInLiveResize is always false off macOS (no AppKit live-resize concept).
func (w *Window) IsInLiveResize() bool { return false }

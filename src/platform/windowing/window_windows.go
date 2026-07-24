/******************************************************************************/
/* window_windows.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package windowing

import (
	"errors"
	"image"
	"image/draw"
	"unicode/utf16"
	"unsafe"

	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"

	"golang.design/x/clipboard"
)

/*
#cgo LDFLAGS: -lgdi32 -lXInput -ldwmapi
#cgo noescape get_toggle_key_state
#cgo noescape window_main
#cgo noescape window_show
#cgo noescape window_destroy
#cgo noescape window_cursor_standard
#cgo noescape window_cursor_ibeam
#cgo noescape window_dpi
#cgo noescape screen_width_mm
#cgo noescape screen_height_mm
#cgo noescape screen_count
#cgo noescape screen_resolutions
#cgo noescape window_focus
#cgo noescape window_position
#cgo noescape window_set_position
#cgo noescape window_set_size
#cgo noescape window_remove_border
#cgo noescape window_add_border
#cgo noescape window_poll_controller
#cgo noescape window_poll
#cgo noescape window_show_cursor
#cgo noescape window_hide_cursor
#cgo noescape window_lock_cursor
#cgo noescape window_unlock_cursor
#cgo noescape window_set_fullscreen
#cgo noescape window_set_windowed
#cgo noescape window_enable_raw_mouse
#cgo noescape window_disable_raw_mouse
#cgo noescape window_set_title
#cgo noescape window_set_title_bar_mode
#cgo noescape window_set_cursor_position
#cgo noescape window_set_icon

#include "windowing.h"
*/
import "C"

//export goProcessEvents
func goProcessEvents(goWindow C.uint64_t, events unsafe.Pointer, eventCount C.uint32_t) {
	goProcessEventsCommon(uint64(goWindow), events, uint32(eventCount))
}

func (w *Window) checkToggleKeyState() map[hid.KeyboardKey]bool {
	mask := C.get_toggle_key_state()
	caps := (mask & 1) != 0
	num := (mask & 2) != 0
	scroll := (mask & 4) != 0
	return map[hid.KeyboardKey]bool{
		hid.KeyboardKeyCapsLock:   caps,
		hid.KeyboardKeyNumLock:    num,
		hid.KeyboardKeyScrollLock: scroll,
	}
}

func (w *Window) createWindow(windowName string, x, y int, _ any) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	w.lookupId = nextLookupId.Add(1)
	windowLookup.Store(w.lookupId, w)
	C.window_main(title, C.int(w.width), C.int(w.height),
		C.int(x), C.int(y), C.uint64_t(w.lookupId))
}

func (w *Window) showWindow() {
	C.window_show(w.handle)
}

func destroyWindow(handle unsafe.Pointer) {
	C.window_destroy(handle)
}

func (w *Window) pollController() {
	defer tracing.NewRegion("Window.pollController").End()
	C.window_poll_controller(w.handle)
}

func (w *Window) pollEvents() {
	defer tracing.NewRegion("Window.pollEvents").End()
	C.window_poll(w.handle)
}

func (w *Window) poll() {
	w.pollController()
	w.pollEvents()
}

func (w *Window) cursorStandard() {
	C.window_cursor_standard(w.handle)
}

func (w *Window) cursorHand() {
	C.window_cursor_hand(w.handle)
}

func (w *Window) cursorIbeam() {
	C.window_cursor_ibeam(w.handle)
}

func (w *Window) cursorSizeAll() {
	C.window_cursor_size_all(w.handle)
}

func (w *Window) cursorSizeNS() {
	C.window_cursor_size_ns(w.handle)
}

func (w *Window) cursorSizeWE() {
	C.window_cursor_size_we(w.handle)
}

func (w *Window) copyToClipboard(text string) {
	clipboard.Write(clipboard.FmtText, []byte(text))
}

func (w *Window) clipboardContents() string {
	return string(clipboard.Read(clipboard.FmtText))
}

func (w *Window) invalidateMonitorCache() {}

func (w *Window) dotsPerMillimeter() float64 {
	dpi := float64(C.window_dpi(w.handle))
	return dpi / 25.4
}

func (w *Window) monitorCount() int {
	return int(C.screen_count())
}

func (w *Window) monitorResolutions() []MonitorResolution {
	count := int(C.screen_resolutions(nil, 0))
	if count <= 0 {
		return nil
	}
	native := make([]C.MonitorResolution, count)
	found := int(C.screen_resolutions(&native[0], C.int(len(native))))
	if found > len(native) {
		native = make([]C.MonitorResolution, found)
		found = int(C.screen_resolutions(&native[0], C.int(len(native))))
	}
	found = min(found, len(native))
	resolutions := make([]MonitorResolution, found)
	for i := range found {
		resolutions[i] = MonitorResolution{
			Width:  int(native[i].width),
			Height: int(native[i].height),
		}
	}
	return resolutions
}

func (w *Window) sizeMM() (int, int, error) {
	dpmm := float64(C.window_dpi(w.handle)) / 25.4
	if dpmm <= 0 {
		return 0, 0, errors.New("invalid dpmm")
	}
	return int(float64(w.width) / dpmm), int(float64(w.height) / dpmm), nil
}

func (w *Window) screenSizeMM() (int, int, error) {
	width := int(C.screen_width_mm(w.handle))
	height := int(C.screen_height_mm(w.handle))
	var err error
	if width == -1 {
		err = errors.New("width: failed to get the device context for HWND")
	} else if height == -1 {
		err = errors.New("height: failed to get the device context for HWND")
	}
	return width, height, err
}

func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }

func (w *Window) focus() {
	C.window_focus(w.handle)
}

func (w *Window) position() (x, y int) {
	C.window_position(w.handle, (*C.int)(unsafe.Pointer(&x)), (*C.int)(unsafe.Pointer(&y)))
	return x, y
}

func (w *Window) setPosition(x, y int) {
	C.window_set_position(w.handle, C.int(x), C.int(y))
}

func (w *Window) setSize(width, height int) {
	C.window_set_size(w.handle, C.int(width), C.int(height))
}

func (w *Window) removeBorder() {
	C.window_remove_border(w.handle)
}

func (w *Window) addBorder() {
	C.window_add_border(w.handle)
}

func (w *Window) showCursor() {
	C.window_show_cursor(w.handle)
}

func (w *Window) hideCursor() {
	C.window_hide_cursor(w.handle)
}

func (w *Window) lockCursor(x, y int) {
	C.window_lock_cursor(w.handle, C.int(x), C.int(y))
}

func (w *Window) unlockCursor() {
	C.window_unlock_cursor(w.handle)
}

func (w *Window) setFullscreen() {
	C.window_set_fullscreen(w.handle)
}

func (w *Window) setWindowed(width, height int) {
	C.window_set_windowed(w.handle, C.int(width), C.int(height))
}

func (w *Window) enableRawMouse() {
	C.window_enable_raw_mouse(w.handle)
}

func (w *Window) disableRawMouse() {
	C.window_disable_raw_mouse(w.handle)
}

func (w *Window) setTitle(newTitle string) {
	windowTitle := utf16.Encode(append([]rune(newTitle), []rune("\x00\x00")...))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	C.window_set_title(w.handle, title)
}

func (w *Window) setTitleBarMode(mode TitleBarMode) {
	C.window_set_title_bar_mode(w.handle, C.int(mode))
}

func (w *Window) getTitleBarMode() TitleBarMode {
	return w.titleBarMode
}

func (w *Window) setCursorPosition(x, y int) {
	C.window_set_cursor_position(w.handle, C.int(x), C.int(y))
}

func (w *Window) setIcon(img image.Image) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)
	bgra := make([]byte, width*height*4)
	for i := 0; i < len(rgba.Pix); i += 4 {
		bgra[i] = rgba.Pix[i+2]
		bgra[i+1] = rgba.Pix[i+1]
		bgra[i+2] = rgba.Pix[i]
		bgra[i+3] = rgba.Pix[i+3]
	}
	C.window_set_icon(w.handle, C.int(width), C.int(height), (*C.uint8_t)(&bgra[0]))
}

func (w *Window) readApplicationAsset(path string) ([]byte, error) {
	return []byte{}, errors.New("windows doesn't support application assets")
}

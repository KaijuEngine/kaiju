/******************************************************************************/
/* x11.h                                                                      */
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

#ifndef WINDOWING_X11_H
#define WINDOWING_X11_H

#include <X11/Xlib.h>
#include <X11/extensions/Xrandr.h>
#include <linux/joystick.h>
#include <linux/input.h>
#include <stdbool.h>
#include "shared_mem.h"

#ifndef MAX_CONTROLLERS
#define MAX_CONTROLLERS 4
#endif

#ifndef EVIOCGABS
#define EVIOCGABS(axis) _IOR('E', 0x20 + (axis), struct input_absinfo)
#endif

#ifndef EVIOCGKEY
#define EVIOCGKEY(len) _IOR('E', 0x2f, unsigned char[len])
#endif

typedef struct {
	int fd;
	bool connected;
	char name[128];
	uint8_t numAxes;
	uint8_t numButtons;
	uint16_t buttonState;
	int16_t axisState[8];
} ControllerInfo;

typedef struct {
	float dpmm;
	int mm_width;
	int mm_height;
	int px_width;
	int px_height;
	int x;
	int y;
	int found;
} MonitorInfo;

typedef struct {
	SharedMem sm;
	Window w;
	Display* d;
	Atom WM_DELETE_WINDOW;
	Atom TARGETS;
	Atom TEXT;
	Atom UTF8_STRING;
	Atom CLIPBOARD;
	MonitorInfo monitorCache;
	int monitorCacheDirty;
	ControllerInfo controllers[MAX_CONTROLLERS];
} X11State;

unsigned int get_toggle_key_state();
void window_main(const char* windowTitle,
	int width, int height, int x, int y, uint64_t goWindow);
void window_show(void* x11State);
void window_poll_controller(void* x11State);
void window_poll(void* x11State);
void window_destroy(void* x11State);
void* display(void* x11State);
void* window(void* x11State);
void window_focus(void* state);
void window_position(void* state, int* x, int* y);
void window_set_position(void* state, int x, int y);
void window_set_size(void* state, int width, int height);
int window_width_mm(void* state);
int window_height_mm(void* state);
int screen_count(void* state);
void window_invalidate_monitor_cache(void* state);
void window_cursor_standard(void* state);
void window_cursor_ibeam(void* state);
void window_cursor_size_all(void* state);
void window_cursor_size_ns(void* state);
void window_cursor_size_we(void* state);
void window_show_cursor(void* state);
void window_hide_cursor(void* state);
float window_dpi(void* state);
void window_set_title(void* state, const char* windowTitle);
void window_set_full_screen(void* state);
void window_set_windowed(void* state, int width, int height);
void window_lock_cursor(void* state, int x, int y);
void window_unlock_cursor(void* state);
void window_set_cursor_position(void* state, int x, int y);
void window_set_icon(void* state, int width, int height, const unsigned char* rgba);

#endif

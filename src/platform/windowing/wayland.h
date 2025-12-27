/******************************************************************************/
/* wayland.h                                                                  */
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

#ifndef WINDOWING_WAYLAND_H
#define WINDOWING_WAYLAND_H

#include <wayland-client.h>
#include <xkbcommon/xkbcommon.h>
#include "shared_mem.h"

typedef struct {
	SharedMem sm;
	struct wl_display* display;
	struct wl_registry* registry;
	struct wl_compositor* compositor;
	struct wl_surface* surface;
	struct wl_seat* seat;
	struct wl_keyboard* keyboard;
	struct wl_pointer* pointer;
	struct wl_shm* shm;
	struct wl_output* output;
	struct xdg_wm_base* xdg_wm_base;
	struct xdg_surface* xdg_surface;
	struct xdg_toplevel* xdg_toplevel;
	struct zxdg_decoration_manager_v1* decoration_manager;
	struct zxdg_toplevel_decoration_v1* toplevel_decoration;
	struct wl_cursor_theme* cursor_theme;
	struct wl_cursor* current_cursor;
	struct wl_surface* cursor_surface;
	struct xkb_context* xkb_context;
	struct xkb_keymap* xkb_keymap;
	struct xkb_state* xkb_state;
	int32_t output_scale;
	int32_t output_width;
	int32_t output_height;
	int32_t output_width_mm;
	int32_t output_height_mm;
	bool configured;
	bool closed;
} WaylandState;

unsigned int get_toggle_key_state();
void window_main(const char* windowTitle,
	int width, int height, int x, int y, uint64_t goWindow);
void window_show(void* waylandState);
void window_poll_controller(void* waylandState);
void window_poll(void* waylandState);
void window_destroy(void* waylandState);
void* display(void* waylandState);
void* window(void* waylandState);
void window_focus(void* state);
void window_position(void* state, int* x, int* y);
void window_set_position(void* state, int x, int y);
void window_set_size(void* state, int width, int height);
int window_width_mm(void* state);
int window_height_mm(void* state);
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

#endif

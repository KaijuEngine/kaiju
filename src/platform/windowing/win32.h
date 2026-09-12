/******************************************************************************/
/* win32.h                                                                    */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

#ifndef WINDOWING_WIN32_H
#define WINDOWING_WIN32_H

#include <wchar.h>
#include <stdint.h>
#include <stdbool.h>

typedef struct {
	int width;
	int height;
} MonitorResolution;

unsigned int get_toggle_key_state();
void window_main(const wchar_t* windowTitle,
	int width, int height, int x, int y, uint64_t goWindow);
void window_show(void* hwnd);
void window_poll_controller(void* hwnd);
void window_poll(void* hwnd);
void window_destroy(void* hwnd);
void window_cursor_standard(void* hwnd);
void window_cursor_hand(void* hwnd);
void window_cursor_ibeam(void* hwnd);
void window_cursor_size_all(void* hwnd);
void window_cursor_size_ns(void* hwnd);
void window_cursor_size_we(void* hwnd);
float window_dpi(void* hwnd);
int screen_width_mm(void* hwnd);
int screen_height_mm(void* hwnd);
int screen_count();
int screen_resolutions(MonitorResolution* resolutions, int capacity);
void window_focus(void* hwnd);
void window_position(void* hwnd, int* x, int* y);
void window_set_position(void* hwnd, int x, int y);
void window_set_size(void* hwnd, int width, int height);
void window_remove_border(void* hwnd);
void window_add_border(void* hwnd);
void window_show_cursor(void* hwnd);
void window_hide_cursor(void* hwnd);
void window_lock_cursor(void* hwnd, int x, int y);
void window_unlock_cursor(void* hwnd);
void window_set_fullscreen(void* hwnd);
void window_set_windowed(void* hwnd, int width, int height);
void window_enable_raw_mouse(void* hwnd);
void window_disable_raw_mouse(void* hwnd);
void window_set_title(void* hwnd, const wchar_t* windowTitle);
void window_set_title_bar_mode(void* hwnd, int mode);
void window_set_cursor_position(void* hwnd, int x, int y);
void window_set_icon(void* hwnd, int width, int height, const uint8_t* pixelData);
#if KAIJU_ENABLE_FILEDROP
void window_set_file_drop_enabled(void* hwnd, bool enabled);
#endif

#endif

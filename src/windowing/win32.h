/******************************************************************************/
/* win32.h                                                                   */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

#ifndef WINDOWING_WIN32_H
#define WINDOWING_WIN32_H

#include <wchar.h>
#include <stdint.h>
#include <stdbool.h>

void window_main(const wchar_t* windowTitle,
	int width, int height, int x, int y, uint64_t goWindow);
void window_show(void* hwnd);
void window_poll_controller(void* hwnd);
void window_poll(void* hwnd);
void window_destroy(void* hwnd);
void window_cursor_standard(void* hwnd);
void window_cursor_ibeam(void* hwnd);
void window_cursor_size_all(void* hwnd);
void window_cursor_size_ns(void* hwnd);
void window_cursor_size_we(void* hwnd);
float window_dpi(void* hwnd);
void window_focus(void* hwnd);
void window_position(void* hwnd, int* x, int* y);
void window_set_position(void* hwnd, int x, int y);
void window_set_size(void* hwnd, int width, int height);
void window_remove_border(void* hwnd);
void window_add_border(void* hwnd);

#endif
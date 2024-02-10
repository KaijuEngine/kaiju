#ifndef WINDOWING_WIN32_H
#define WINDOWING_WIN32_H

#include <wchar.h>
#include <stdint.h>
#include <stdbool.h>

void window_main(const wchar_t* windowTitle, int width, int height, void* evtSharedMem, int size);
void window_show(void* hwnd);
uint32_t window_poll_controller(void* hwnd);
uint32_t window_poll(void* hwnd);
void window_destroy(void* hwnd);
void window_cursor_standard(void* hwnd);
void window_cursor_ibeam(void* hwnd);

#ifdef OPENGL
void window_create_gl_context(void* winHWND, void* evtSharedMem, int size);
#endif

#endif
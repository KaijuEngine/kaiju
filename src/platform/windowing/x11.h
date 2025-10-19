

#ifndef WINDOWING_X11_H
#define WINDOWING_X11_H

#include <X11/Xlib.h>
#include "shared_mem.h"

typedef struct {
	SharedMem sm;
	Window w;
	Display* d;
	Atom WM_DELETE_WINDOW;
	Atom TARGETS;
	Atom TEXT;
	Atom UTF8_STRING;
	Atom CLIPBOARD;
} X11State;

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
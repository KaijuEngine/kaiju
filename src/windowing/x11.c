/******************************************************************************/
/* x11.c                                                                     */
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

#if (defined(__linux__) || defined(__unix__)) && !defined(__ANDROID__)

#include "x11.h"
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <pthread.h>
#include <X11/Xlib.h>
#include <X11/Xcursor/Xcursor.h>

// Cursor docs
// https://tronche.com/gui/x/xlib/appendix/b/
//#include <X11/cursorfont.h>

// XLib docs
// https://www.x.org/releases/X11R7.7/doc/libX11/libX11/libX11.html

#define EVT_MASK	ExposureMask | KeyPressMask | KeyReleaseMask | ButtonPressMask | ButtonReleaseMask | PointerMotionMask | FocusChangeMask | StructureNotifyMask

Atom XA_ATOM = 4, XA_STRING = 31;

static bool isExtensionSupported(const char* extList, const char* extension) {
	const char* start;
	const char* where, *terminator;
	where = strchr(extension, ' ');
	if (where || *extension == '\0') {
		return false;
	}
	for (start = extList;;) {
		where = strstr(start, extension);
		if (!where) {
			break;
		}
		terminator = where + strlen(extension);
		if (where == start || *(where - 1) == ' ') {
			if (*terminator == ' ' || *terminator == '\0') {
				return true;
			}
		}
		start = terminator;
	}
	return false;
}

void window_main(const char* windowTitle,
	int width, int height, int x, int y, uint64_t goWindow)
{
	X11State* x11State = calloc(1, sizeof(X11State));
	x11State->sm.goWindow = (void*)goWindow;
	XInitThreads();
	Display* d = XOpenDisplay(NULL);
	if (d == NULL) {
		printf("Failed to open display"); fflush(stdout);
		shared_mem_add_event(&x11State->sm, (WindowEvent) {
			.type = WINDOW_EVENT_TYPE_FATAL,
		});
		shared_mem_flush_events(&x11State->sm);
		free(x11State);
		return;
	}
	if (x < 0) {
		x = 10;
	}
	if (y < 0) {
		y = 10;
	}
	Window w = XCreateSimpleWindow(d, RootWindow(d, DefaultScreen(d)), x, y,
		width, height, 1, BlackPixel(d, DefaultScreen(d)), WhitePixel(d, DefaultScreen(d)));
	if (w == None) {
		printf("Failed to create window"); fflush(stdout);
		shared_mem_add_event(&x11State->sm, (WindowEvent) {
			.type = WINDOW_EVENT_TYPE_FATAL,
		});
		shared_mem_flush_events(&x11State->sm);
		free(x11State);
		return;
	}
	XStoreName(d, w, windowTitle);
	XSetIconName(d, w, windowTitle);
	XSelectInput(d, w, EVT_MASK);
	XMapWindow(d, w);
	x11State->w = w;
	x11State->d = d;
	x11State->WM_DELETE_WINDOW = XInternAtom(d, "WM_DELETE_WINDOW", False);
	x11State->TARGETS = XInternAtom(d, "TARGETS", 0);
	x11State->TEXT = XInternAtom(d, "TEXT", 0);
	x11State->UTF8_STRING = XInternAtom(d, "UTF8_STRING", 1);
	if (x11State->UTF8_STRING== None) {
		x11State->UTF8_STRING = XA_STRING;
	}
	x11State->CLIPBOARD = XInternAtom(d, "CLIPBOARD", 0);
	XSetWMProtocols(d, w, &x11State->WM_DELETE_WINDOW, 1);
	shared_mem_add_event(&x11State->sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_SET_HANDLE,
		.setHandle = {
			.hwnd = x11State,
		}
	});
	shared_mem_flush_events(&x11State->sm);
}

void window_show(void* x11State) {
	X11State* s = x11State;
	// Flush initial events
	//while (window_poll(x11State) != Expose) {}
	//while (window_poll(x11State) != 0) {}
}

void window_poll_controller(void* x11State) {
	// TODO:  Implement for controllers
}

void window_poll(void* x11State) {
	X11State* s = x11State;
	XEvent e = { 0 };
	while (true) {
		if (!XCheckMaskEvent(s->d, EVT_MASK, &e)) {
			if (!XCheckTypedEvent(s->d, ClientMessage, &e)) {
				break;
			}
		}
		bool filtered = XFilterEvent(&e, s->w);
		switch (e.type) {
			case DestroyNotify:
				shared_mem_add_event(&s->sm, (WindowEvent) {
					.type = WINDOW_EVENT_TYPE_ACTIVITY,
					.windowActivity = {
						.activityType = WINDOW_EVENT_ACTIVITY_TYPE_CLOSE,
					}
				});
				shared_mem_flush_events(&s->sm);
				break;
			case Expose:
				break;
			case FocusIn:
				shared_mem_add_event(&s->sm, (WindowEvent) {
					.type = WINDOW_EVENT_TYPE_ACTIVITY,
					.windowActivity = {
						.activityType = WINDOW_EVENT_ACTIVITY_TYPE_FOCUS,
					}
				});
				break;
			case FocusOut:
				shared_mem_add_event(&s->sm, (WindowEvent) {
					.type = WINDOW_EVENT_TYPE_ACTIVITY,
					.windowActivity = {
						.activityType = WINDOW_EVENT_ACTIVITY_TYPE_BLUR,
					}
				});
				break;
			case ConfigureNotify:
				if (s->x == 0 || s->y == 0 || s->width == 0 || s->height == 0) {
					s->x = e.xconfigure.x;
					s->y = e.xconfigure.y;
					s->width = e.xconfigure.width;
					s->height = e.xconfigure.height;
				}
				if (s->x != e.xconfigure.x || s->y != e.xconfigure.y) {
					s->x = e.xconfigure.x;
					s->y = e.xconfigure.y;
					shared_mem_add_event(&s->sm, (WindowEvent) {
						.type = WINDOW_EVENT_TYPE_MOVE,
						.windowMove = {
							.x = s->x,
							.y = s->y,
						}
					});
				}
				if (s->width != e.xconfigure.width || s->height != e.xconfigure.height) {
					s->width = e.xconfigure.width;
					s->height = e.xconfigure.height;
					shared_mem_add_event(&s->sm, (WindowEvent) {
						.type = WINDOW_EVENT_TYPE_RESIZE,
						.windowResize = {
							.width = e.xconfigure.width,
							.height = e.xconfigure.height,
							.left = e.xconfigure.x,
							.top = e.xconfigure.y,
							.right = e.xconfigure.x + e.xconfigure.width,
							.bottom = e.xconfigure.y + e.xconfigure.height,
						}
					});
				}
				break;
			case KeyPress:
			case KeyRelease:
				shared_mem_add_event(&s->sm, (WindowEvent) {
					.type = WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
					.keyboardButton = {
						.action = e.type == KeyPress
							? WINDOW_EVENT_BUTTON_TYPE_DOWN : WINDOW_EVENT_BUTTON_TYPE_UP,
						.buttonId = XLookupKeysym(&e.xkey, 0),
					}
				});
				break;
			case ButtonPress:
			case ButtonRelease:
			{
				WindowEvent evt = {
					.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
					.mouseButton.action = e.type == ButtonPress
						? WINDOW_EVENT_BUTTON_TYPE_DOWN : WINDOW_EVENT_BUTTON_TYPE_UP,
				};
				switch (e.xbutton.button) {
					case Button1:
						evt.mouseButton.buttonId = MOUSE_BUTTON_LEFT;
						break;
					case Button2:
						evt.mouseButton.buttonId = MOUSE_BUTTON_MIDDLE;
						break;
					case Button3:
						evt.mouseButton.buttonId = MOUSE_BUTTON_RIGHT;
						break;
					case Button4:
						evt.mouseButton.buttonId = MOUSE_BUTTON_X1;
						break;
					case Button5:
						evt.mouseButton.buttonId = MOUSE_BUTTON_X2;
						break;
				}
				evt.mouseButton.x = e.xbutton.x;
				evt.mouseButton.y = e.xbutton.y;
				shared_mem_add_event(&s->sm, evt);
				break;
			}
			case MotionNotify:
				shared_mem_add_event(&s->sm, (WindowEvent) {
					.type = WINDOW_EVENT_TYPE_MOUSE_MOVE,
					.mouseMove = {
						.x = e.xmotion.x,
						.y = e.xmotion.y,
					}
				});
				break;
			case ClientMessage:
				if (!filtered) {
					const Atom protocol = e.xclient.data.l[0];
					if (protocol == s->WM_DELETE_WINDOW) {
						XDestroyWindow(s->d, s->w);
						shared_mem_add_event(&s->sm, (WindowEvent) {
							.type = WINDOW_EVENT_TYPE_ACTIVITY,
							.windowActivity = {
								.activityType = WINDOW_EVENT_ACTIVITY_TYPE_CLOSE,
							}
						});
						shared_mem_flush_events(&s->sm);
					}
				}
				break;
		}
	}
	shared_mem_flush_events(&s->sm);
}

void window_destroy(void* x11State) {
	X11State* s = x11State;
	XDestroyWindow(s->d, s->w);
	XCloseDisplay(s->d);
	free(s);
}

void* display(void* x11State) { return ((X11State*)x11State)->d; }
void* window(void* x11State) { return &((X11State*)x11State)->w; }

void window_focus(void* state) {
	X11State* s = state;
	XRaiseWindow(s->d, s->w);
	XSetInputFocus(s->d, s->w, RevertToParent, CurrentTime);
}

int window_width_mm(void* state) {
	X11State* s = state;
	int sid = DefaultScreen(s->d);
	return XDisplayWidthMM(s->d, sid);
}

int window_height_mm(void* state) {
	X11State* s = state;
	int sid = DefaultScreen(s->d);
	return XDisplayHeightMM(s->d, sid);
}

void window_cursor_standard(void* state) {
	X11State* s = state;
	Cursor c = XcursorLibraryLoadCursor(s->d, "arrow");
	XDefineCursor(s->d, s->w, c);
}

void window_cursor_ibeam(void* state) {
	X11State* s = state;
	Cursor c = XcursorLibraryLoadCursor(s->d, "xterm");
	XDefineCursor(s->d, s->w, c);
}

void window_cursor_size_all(void* state) {
	X11State* s = state;
	Cursor c = XcursorLibraryLoadCursor(s->d, "sizing");
	XDefineCursor(s->d, s->w, c);
}

void window_cursor_size_ns(void* state) {
	X11State* s = state;
	Cursor c = XcursorLibraryLoadCursor(s->d, "sb_v_double_arrow");
	XDefineCursor(s->d, s->w, c);
}

void window_cursor_size_we(void* state) {
	X11State* s = state;
	Cursor c = XcursorLibraryLoadCursor(s->d, "sb_h_double_arrow");
	XDefineCursor(s->d, s->w, c);
}

void window_position(void* state, int* x, int* y) {
	X11State* s = state;
	XWindowAttributes attributes;
	XGetWindowAttributes(s->d, s->w, &attributes);
	*x = attributes.x;
	*y = attributes.y;
}

void window_set_position(void* state, int x, int y) {
	X11State* s = state;
	XMoveWindow(s->d, s->w, x, y);
}

void window_set_size(void* state, int width, int height) {
	X11State* s = state;
	XResizeWindow(s->d, s->w, width, height);
}

#endif

//go:build linux && !android

/******************************************************************************/
/* x11.c                                                                      */
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

#if (defined(__linux__) || defined(__unix__)) && !(defined(__android__) || defined(__ANDROID__))

#include "x11.h"
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <pthread.h>
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>
#include <X11/Xlib.h>
#include <X11/Xcursor/Xcursor.h>
#include <X11/XKBlib.h>
#include <X11/Xatom.h>
#include <X11/extensions/Xrandr.h>
#include <linux/joystick.h>

// Cursor docs
// https://tronche.com/gui/x/xlib/appendix/b/
//#include <X11/cursorfont.h>

// XLib docs
// https://www.x.org/releases/X11R7.7/doc/libX11/libX11/libX11.html

#define EVT_MASK	ExposureMask | KeyPressMask | KeyReleaseMask | ButtonPressMask | ButtonReleaseMask | PointerMotionMask | FocusChangeMask | StructureNotifyMask

// Atom XA_ATOM = 4, XA_STRING = 31;

unsigned int get_toggle_key_state() {
    Display *d = XOpenDisplay(NULL);
    if (!d) return 0;

    unsigned int state = 0;
    XkbGetIndicatorState(d, XkbUseCoreKbd, &state);
    XCloseDisplay(d);

    return state;
}

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
	x11State->sm.x = x;
	x11State->sm.y = y;
	x11State->sm.windowWidth = width;
	x11State->sm.windowHeight = height;
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
	if (x < 0 || y < 0) {
		int screen = DefaultScreen(d);               // primary screen
		int screenWidth = DisplayWidth(d, screen);   // primary screen
		int screenHeight = DisplayHeight(d, screen); // primary screen
		x = (screenWidth - width) / 2;
		y = (screenHeight - height) / 2;
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
	// Initialize controller states
	for (int i = 0; i < MAX_CONTROLLERS; i++) {
		x11State->controllers[i].fd = -1;
		x11State->controllers[i].connected = false;
	}
	// Scan for available game controllers
	for (int i = 0; i < MAX_CONTROLLERS; i++) {
		char devicePath[64];
		snprintf(devicePath, sizeof(devicePath), "/dev/input/js%d", i);
		int fd = open(devicePath, O_RDONLY | O_NONBLOCK);
		if (fd >= 0) {
			x11State->controllers[i].fd = fd;
			x11State->controllers[i].connected = true;
			x11State->controllers[i].buttonState = 0;
			for (int axis = 0; axis < 8; axis++) {
				x11State->controllers[i].axisState[axis] = 0;
			}
			// Get controller name via ioctl
			if (ioctl(fd, JSIOCGNAME(sizeof(x11State->controllers[i].name)), x11State->controllers[i].name) < 0) {
				strncpy(x11State->controllers[i].name, "Unknown Controller", sizeof(x11State->controllers[i].name) - 1);
			}
			// Get number of axes and buttons
			uint8_t numAxes = 0, numButtons = 0;
			ioctl(fd, JSIOCGAXES, &numAxes);
			ioctl(fd, JSIOCGBUTTONS, &numButtons);
			x11State->controllers[i].numAxes = numAxes;
			x11State->controllers[i].numButtons = numButtons;
		}
	}
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

static inline void set_cursor_position_relative_to_window(X11State* s, int x, int y) {
	int borderSize = ((s->sm.right-s->sm.left)-s->sm.clientRect.right) / 2;
	int titleSize = (s->sm.bottom-s->sm.top)-s->sm.clientRect.bottom-borderSize;
	int wx = s->sm.left + x + borderSize;
	int wy = s->sm.top + y + titleSize;
	XWarpPointer(s->d, None, s->w, 0, 0, 0, 0, wx, wy);
}

static inline void lock_cursor_position(X11State* s) {
	set_cursor_position_relative_to_window(s, s->sm.lockCursor.x, s->sm.lockCursor.y);
}

// Linux joystick deadzone values (matching XInput)
#define JOYSTICK_DEADZONE_AXIS 7849  // ~24% of 32768
#define JOYSTICK_DEADZONE_TRIGGER 30  // ~12% of 255
#define JOYSTICK_HAT_DEADZONE 16384   // half of full hat axis range

static inline int16_t apply_axis_deadzone(int16_t value, int16_t deadzone) {
	if (value < 0) {
		if (-value < deadzone) return 0;
		return value;
	}
	if (value < deadzone) return 0;
	return value;
}

static inline uint8_t apply_trigger_deadzone(uint8_t value, uint8_t deadzone) {
	if (value < deadzone) return 0;
	return value;
}

void window_poll_controller(void* x11State) {
	X11State* s = x11State;
	struct js_event evt;
	for (int i = 0; i < MAX_CONTROLLERS; i++) {
		if (!s->controllers[i].connected || s->controllers[i].fd < 0) {
			// Try to reconnect
			char devicePath[64];
			snprintf(devicePath, sizeof(devicePath), "/dev/input/js%d", i);
			int fd = open(devicePath, O_RDONLY | O_NONBLOCK);
			if (fd >= 0) {
				s->controllers[i].fd = fd;
				s->controllers[i].connected = true;
				uint8_t numAxes = 0, numButtons = 0;
				ioctl(fd, JSIOCGAXES, &numAxes);
				ioctl(fd, JSIOCGBUTTONS, &numButtons);
				s->controllers[i].numAxes = numAxes;
				s->controllers[i].numButtons = numButtons;
				// Initialize state (this was the missing piece — analogs were garbage/old memory,
				// deadzone turned everything to 0; triggers start released at -32768)
				s->controllers[i].buttonState = 0;
				if (numAxes >= 1) {
					s->controllers[i].axisState[0] = 0;
				}
				if (numAxes >= 2) {
					s->controllers[i].axisState[1] = 0;
				}
				if (numAxes >= 3) {
					s->controllers[i].axisState[2] = 0;
				}
				if (numAxes >= 4) {
					s->controllers[i].axisState[3] = 0;
				}
				if (numAxes >= 5) {
					s->controllers[i].axisState[4] = -32768; // left trigger released
				}
				if (numAxes >= 6) {
					s->controllers[i].axisState[5] = -32768; // right trigger released
				}
				if (numAxes >= 7) {
					s->controllers[i].axisState[6] = 0;     // hat X
				}
				if (numAxes >= 8) {
					s->controllers[i].axisState[7] = 0;     // hat Y
				}
				// Flush any initialization events from the device
				while (read(fd, &evt, sizeof(evt)) > 0) { /* flush */ }
				shared_mem_add_event(&s->sm, (WindowEvent) {
					.type = WINDOW_EVENT_TYPE_CONTROLLER_STATE,
					.controllerState = {
						.controllerId = i,
						.connectionType = WINDOW_EVENT_CONTROLLER_CONNECTION_TYPE_CONNECTED,
					}
				});
			}
			continue;
		}
		// Read all available events for this controller to keep state current
		fd_set fdset;
		struct timeval tv;
		int16_t thumbLX = 0, thumbLY = 0, thumbRX = 0, thumbRY = 0;
		uint8_t leftTrigger = 0, rightTrigger = 0;
		uint16_t buttons = s->controllers[i].buttonState;
		while (true) {
			FD_ZERO(&fdset);
			FD_SET(s->controllers[i].fd, &fdset);
			tv.tv_sec = 0;
			tv.tv_usec = 0;
			int ready = select(s->controllers[i].fd + 1, &fdset, NULL, NULL, &tv);
			if (ready <= 0) break;
			int bytesRead = read(s->controllers[i].fd, &evt, sizeof(evt));
			if (bytesRead != sizeof(evt)) break;
			unsigned char type = evt.type & ~JS_EVENT_INIT;
			if (type == JS_EVENT_BUTTON && evt.number < 16) {
				if (evt.value) {
					s->controllers[i].buttonState |= (1u << evt.number);
				} else {
					s->controllers[i].buttonState &= ~(1u << evt.number);
				}
				buttons = s->controllers[i].buttonState;
			} else if (type == JS_EVENT_AXIS && evt.number < 8) {
				s->controllers[i].axisState[evt.number] = (int16_t)evt.value;
			}
		}
		if (s->controllers[i].numAxes > 0) {
			thumbLX = apply_axis_deadzone(s->controllers[i].axisState[0], JOYSTICK_DEADZONE_AXIS);
			if (s->controllers[i].numAxes > 1) {
				thumbLY = -apply_axis_deadzone(s->controllers[i].axisState[1], JOYSTICK_DEADZONE_AXIS);
			}
			if (s->controllers[i].numAxes > 2) {
				leftTrigger = apply_trigger_deadzone((uint8_t)((s->controllers[i].axisState[2] + 32768) >> 8), JOYSTICK_DEADZONE_TRIGGER);
			}
			if (s->controllers[i].numAxes > 3) {
				thumbRX = apply_axis_deadzone(s->controllers[i].axisState[3], JOYSTICK_DEADZONE_AXIS);
			}
			if (s->controllers[i].numAxes > 4) {
				thumbRY = -apply_axis_deadzone(s->controllers[i].axisState[4], JOYSTICK_DEADZONE_AXIS);
			}
			if (s->controllers[i].numAxes > 5) {
				rightTrigger = apply_trigger_deadzone((uint8_t)((s->controllers[i].axisState[5] + 32768) >> 8), JOYSTICK_DEADZONE_TRIGGER);
			}
			if (s->controllers[i].numAxes > 6) {
				int16_t hatX = s->controllers[i].axisState[6];
				if (hatX < -JOYSTICK_HAT_DEADZONE) {
					buttons |= (1u << 14); // D-Pad Left
				}
				if (hatX > JOYSTICK_HAT_DEADZONE)  {
					buttons |= (1u << 15); // D-Pad Right
				}
			}
			if (s->controllers[i].numAxes > 7) {
				int16_t hatY = s->controllers[i].axisState[7];
				if (hatY < -JOYSTICK_HAT_DEADZONE) {
					buttons |= (1u << 12); // D-Pad Up
				}
				if (hatY > JOYSTICK_HAT_DEADZONE) {
					buttons |= (1u << 13); // D-Pad Down
				}
			}
		}
		shared_mem_add_event(&s->sm, (WindowEvent) {
			.type = WINDOW_EVENT_TYPE_CONTROLLER_STATE,
			.controllerState = {
				.controllerId = i,
				.connectionType = WINDOW_EVENT_CONTROLLER_CONNECTION_TYPE_CONNECTED,
				.buttons = buttons,
				.thumbLX = thumbLX,
				.thumbLY = thumbLY,
				.thumbRX = thumbRX,
				.thumbRY = thumbRY,
				.leftTrigger = leftTrigger,
				.rightTrigger = rightTrigger,
			}
		});
	}
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
				// No need to trigger a resize on init
				if (s->sm.windowWidth == 0 || s->sm.windowHeight == 0) {
					s->sm.windowWidth = e.xconfigure.width;
					s->sm.windowHeight = e.xconfigure.height;
				}
				if (s->sm.x != e.xconfigure.x || s->sm.y != e.xconfigure.y) {
					s->sm.x = e.xconfigure.x;
					s->sm.y = e.xconfigure.y;
					shared_mem_add_event(&s->sm, (WindowEvent) {
						.type = WINDOW_EVENT_TYPE_MOVE,
						.windowMove = {
							.x = s->sm.x,
							.y = s->sm.y,
						}
					});
				}
				if (s->sm.windowWidth != e.xconfigure.width || s->sm.windowHeight != e.xconfigure.height) {
					s->sm.windowWidth = e.xconfigure.width;
					s->sm.windowHeight = e.xconfigure.height;
					shared_mem_add_event(&s->sm, (WindowEvent) {
						.type = WINDOW_EVENT_TYPE_RESIZE,
						.windowResize = {
							.width = e.xconfigure.width,
							.height = e.xconfigure.height,
							.left = s->sm.x,
							.top = s->sm.y,
							.right = s->sm.x + e.xconfigure.width,
							.bottom = s->sm.y + e.xconfigure.height,
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
			{
				int btn = e.xbutton.button;
				if (btn >= 4 && btn <= 7) {
					int deltaX = 0;
					int deltaY = 0;
					if (btn == 4) {
						deltaY =  1;
					} else if (btn == 5) {
						deltaY = -1;
					} else if (btn == 6) {
						deltaX = -1;
					} else if (btn == 7) {
						deltaX =  1;
					}
					WindowEvent evt = {
						.type = WINDOW_EVENT_TYPE_MOUSE_SCROLL,
						.mouseScroll = {
							.deltaX = deltaX,
							.deltaY = deltaY,
							.x = e.xbutton.x,
							.y = e.xbutton.y,
						}
					};
					shared_mem_add_event(&s->sm, evt);
				}
			}
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
					// case Button4:
					// 	evt.mouseButton.buttonId = MOUSE_BUTTON_X1;
					// 	break;
					// case Button5:
					// 	evt.mouseButton.buttonId = MOUSE_BUTTON_X2;
					// 	break;
				}
				evt.mouseButton.x = e.xbutton.x;
				evt.mouseButton.y = e.xbutton.y;
				shared_mem_add_event(&s->sm, evt);
				break;
			}
			case MotionNotify:
				if (s->sm.lockCursor.active) {
					if (e.xmotion.x != s->sm.lockCursor.x || e.xmotion.y != s->sm.lockCursor.y) {
						shared_mem_add_event(&s->sm, (WindowEvent) {
							.type = WINDOW_EVENT_TYPE_MOUSE_MOVE,
							.mouseMove = {
								.x = e.xmotion.x,
								.y = e.xmotion.y,
							}
						});
					}
					lock_cursor_position(s);
				} else {
					shared_mem_add_event(&s->sm, (WindowEvent) {
						.type = WINDOW_EVENT_TYPE_MOUSE_MOVE,
						.mouseMove = {
							.x = e.xmotion.x,
							.y = e.xmotion.y,
						}
					});
				}
				break;
			case ClientMessage:
				if (!filtered) {
					const Atom protocol = e.xclient.data.l[0];
					if (protocol == s->WM_DELETE_WINDOW) {
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
	// Close any open controller file descriptors
	for (int i = 0; i < MAX_CONTROLLERS; i++) {
		if (s->controllers[i].fd >= 0) {
			close(s->controllers[i].fd);
			s->controllers[i].fd = -1;
			s->controllers[i].connected = false;
		}
	}
	if (s->w) {
		XDestroyWindow(s->d, s->w);
	}
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

static MonitorInfo query_monitor_info(Display* d, Window w) {
	MonitorInfo info = {0};
	XWindowAttributes attrs;
	XGetWindowAttributes(d, w, &attrs);
	int wx = 0, wy = 0;
	Window child;
	XTranslateCoordinates(d, w, DefaultRootWindow(d), 0, 0, &wx, &wy, &child);
	int wcx = wx + attrs.width / 2;
	int wcy = wy + attrs.height / 2;
	XRRScreenResources* sr = XRRGetScreenResourcesCurrent(d, DefaultRootWindow(d));
	if (!sr) return info;
	for (int i = 0; i < sr->ncrtc; i++) {
		XRRCrtcInfo* ci = XRRGetCrtcInfo(d, sr, sr->crtcs[i]);
		if (!ci || ci->width == 0 || ci->noutput == 0) {
			if (ci) XRRFreeCrtcInfo(ci);
			continue;
		}
		if (wcx >= ci->x && wcx < ci->x + (int)ci->width &&
		    wcy >= ci->y && wcy < ci->y + (int)ci->height) {
			XRROutputInfo* oi = XRRGetOutputInfo(d, sr, ci->outputs[0]);
			if (oi && oi->mm_width > 0 && oi->mm_height > 0) {
				info.dpmm = (float)ci->width / (float)oi->mm_width;
				info.mm_width = (int)oi->mm_width;
				info.mm_height = (int)oi->mm_height;
				info.px_width = (int)ci->width;
				info.px_height = (int)ci->height;
				info.x = ci->x;
				info.y = ci->y;
				info.found = 1;
			}
			if (oi) XRRFreeOutputInfo(oi);
			XRRFreeCrtcInfo(ci);
			break;
		}
		XRRFreeCrtcInfo(ci);
	}
	XRRFreeScreenResources(sr);
	return info;
}

static MonitorInfo find_monitor_info(X11State* s) {
	if (!s->monitorCacheDirty && s->monitorCache.found) {
		return s->monitorCache;
	}
	s->monitorCache = query_monitor_info(s->d, s->w);
	s->monitorCacheDirty = 0;
	return s->monitorCache;
}

void window_invalidate_monitor_cache(void* state) {
	X11State* s = state;
	s->monitorCacheDirty = 1;
}

int window_width_mm(void* state) {
	X11State* s = state;
	MonitorInfo mi = find_monitor_info(s);
	if (mi.found) return mi.mm_width;
	int sid = DefaultScreen(s->d);
	return XDisplayWidthMM(s->d, sid);
}

int window_height_mm(void* state) {
	X11State* s = state;
	MonitorInfo mi = find_monitor_info(s);
	if (mi.found) return mi.mm_height;
	int sid = DefaultScreen(s->d);
	return XDisplayHeightMM(s->d, sid);
}

int screen_count(void* state) {
	(void)state;
	return 1; // TODO
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

void window_show_cursor(void* state) {
	X11State* s = state;
	XUndefineCursor(s->d, s->w);
	XFlush(s->d);
}

void window_hide_cursor(void* state) {
	X11State* s = state;
	Pixmap blank = XCreatePixmap(s->d, s->w, 1, 1, 1);
	XColor dummy_color;
	Cursor cursor = XCreatePixmapCursor(s->d, blank, blank, &dummy_color, &dummy_color, 0, 0);
	XDefineCursor(s->d, s->w, cursor);
	XFreePixmap(s->d, blank);
	XFreeCursor(s->d, cursor);
	XFlush(s->d);
}

void window_lock_cursor(void* state, int x, int y) {
	X11State* s = state;
	s->sm.lockCursor.x = x;
	s->sm.lockCursor.y = y;
	s->sm.lockCursor.active = true;
	set_cursor_position_relative_to_window(state, x, y);
}

void window_unlock_cursor(void* state) {
	X11State* s = state;
	s->sm.lockCursor.active = false;
}

void window_set_cursor_position(void* state, int x, int y) {
	X11State* s = state;
	set_cursor_position_relative_to_window(s, x, y);
}

float window_dpi(void* state) {
	X11State* s = state;
	MonitorInfo mi = find_monitor_info(s);
	if (mi.found) return mi.dpmm;
	int screen = XDefaultScreen(s->d);
	return (float)DisplayWidth(s->d, screen) / (float)DisplayWidthMM(s->d, screen);
}

void window_set_title(void* state, const char* windowTitle) {
	X11State* s = state;
	XStoreName(s->d, s->w, windowTitle);
}

void window_set_full_screen(void* state) {
	X11State* s = state;
	XWindowAttributes attrs;
	XGetWindowAttributes(s->d, s->w, &attrs);
	s->sm.savedState.rect.left = attrs.x;
	s->sm.savedState.rect.top = attrs.y;
	s->sm.savedState.rect.right = attrs.x + attrs.width;
	s->sm.savedState.rect.bottom = attrs.y + attrs.height;
	// TODO:  Save the border state
	s->sm.savedState.borderWidth = attrs.border_width;
	s->sm.savedState.overrideRedirect = attrs.override_redirect;
	int fx, fy, fw, fh;
	MonitorInfo mi = find_monitor_info(s);
	if (mi.found) {
		fx = mi.x;
		fy = mi.y;
		fw = mi.px_width;
		fh = mi.px_height;
	} else {
		int screen = DefaultScreen(s->d);
		fx = 0;
		fy = 0;
		fw = DisplayWidth(s->d, screen);
		fh = DisplayHeight(s->d, screen);
	}
	XClientMessageEvent ev = { 0 };
	ev.type = ClientMessage;
	ev.window = s->w;
	ev.message_type = XInternAtom(s->d, "_NET_WM_STATE", False);
	ev.format = 32;
	ev.data.l[0] = 1;
	ev.data.l[1] = XInternAtom(s->d, "_NET_WM_STATE_FULLSCREEN", False);
	ev.data.l[2] = 0;
	ev.data.l[3] = 1;
	XSendEvent(s->d, DefaultRootWindow(s->d), False,
	           SubstructureRedirectMask | SubstructureNotifyMask, (XEvent*)&ev);
	XSetWindowAttributes attr = { 0 };
	XChangeWindowAttributes(s->d, s->w, CWOverrideRedirect, &attr);
	XSetWindowBorderWidth(s->d, s->w, 0);
	XWindowChanges changes;
	changes.x = fx;
	changes.y = fy;
	changes.width = fw;
	changes.height = fh;
	changes.border_width = 0;
	unsigned int value_mask = CWX | CWY | CWWidth | CWHeight | CWBorderWidth;
	XConfigureWindow(s->d, s->w, value_mask, &changes);
	XFlush(s->d);
}

void window_set_windowed(void* state, int width, int height) {
    X11State* s = state;
    XSetWindowAttributes attr = { 0 };
    attr.override_redirect = s->sm.savedState.overrideRedirect;
    XChangeWindowAttributes(s->d, s->w, CWOverrideRedirect, &attr);
    XSetWindowBorderWidth(s->d, s->w, s->sm.savedState.borderWidth);
    XClientMessageEvent ev = { 0 };
    ev.type = ClientMessage;
    ev.window = s->w;
    ev.message_type = XInternAtom(s->d, "_NET_WM_STATE", False);
    ev.format = 32;
    ev.data.l[0] = 0;
    ev.data.l[1] = XInternAtom(s->d, "_NET_WM_STATE_FULLSCREEN", False);
    ev.data.l[2] = 0;
    ev.data.l[3] = 1;
    XSendEvent(s->d, DefaultRootWindow(s->d), False,
               SubstructureRedirectMask | SubstructureNotifyMask, (XEvent*)&ev);
    XWindowChanges changes;
    changes.x = s->sm.savedState.rect.left;
    changes.y = s->sm.savedState.rect.top;
    changes.width = width;
    changes.height = height;
    changes.border_width = s->sm.savedState.borderWidth;
    unsigned int value_mask = CWX | CWY | CWWidth | CWHeight | CWBorderWidth;
    XConfigureWindow(s->d, s->w, value_mask, &changes);
    XFlush(s->d);
    XSync(s->d, False);
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

void window_set_icon(void* state, int width, int height, const unsigned char* rgba) {
	X11State* s = state;
	Atom netWmIcon = XInternAtom(s->d, "_NET_WM_ICON", False);
	if (netWmIcon == None) {
		return;
	}
	int dataLen = 2 + (width * height);
	unsigned long* iconData = calloc(dataLen, sizeof(unsigned long));
	if (!iconData) {
		return;
	}
	iconData[0] = width;
	iconData[1] = height;
	for (int i = 0; i < width * height; i++) {
		unsigned char r = rgba[i * 4 + 0];
		unsigned char g = rgba[i * 4 + 1];
		unsigned char b = rgba[i * 4 + 2];
		unsigned char a = rgba[i * 4 + 3];
		iconData[2 + i] = ((unsigned long)a << 24) | ((unsigned long)b << 16) | ((unsigned long)g << 8) | r;
	}
	XChangeProperty(s->d, s->w, netWmIcon, XA_CARDINAL, 32, PropModeReplace,
		(unsigned char*)iconData, dataLen);
	XFlush(s->d);
	free(iconData);
}

#endif

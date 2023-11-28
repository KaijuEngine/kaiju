#ifndef WINDOWING_X11_H
#define WINDOWING_X11_H

#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <X11/Xlib.h>

void window_main(const char* windowTitle, void* evtSharedMem, int size) {
	char* esm = evtSharedMem;
	Display* d = XOpenDisplay(NULL);
	if (d == NULL) {
		esm[0] = SHARED_MEM_FATAL;
		return;
	}
	int s = DefaultScreen(d);
	Window w = XCreateSimpleWindow(d, RootWindow(d, s), 10, 10,
		1280, 720,1, BlackPixel(d, s), WhitePixel(d, s));
	XStoreName(d, w, windowTitle);
	XSetIconName(d, w, windowTitle);
	XSelectInput(d, w, ExposureMask | KeyPressMask | KeyReleaseMask | ButtonPressMask | ButtonReleaseMask | PointerMotionMask);
	XMapWindow(d, w);
	Atom WM_DELETE_WINDOW = XInternAtom(d, "WM_DELETE_WINDOW", False);
	XSetWMProtocols(d, w, &WM_DELETE_WINDOW, 1);
	SharedMem sm = {evtSharedMem, size};
	shared_memory_set_write_state(&sm, SHARED_MEM_WRITTEN);
	XEvent e;
	while (esm[0] != SHARED_MEM_QUIT) {
		while (esm[0] != SHARED_MEM_AVAILABLE) {}
		shared_memory_set_write_state(&sm, SHARED_MEM_WRITING);
		XNextEvent(d, &e);
		bool filtered = XFilterEvent(&e, None);
		uint32_t msgType = e.type;
		switch (e.type) {
			case Expose:
				break;
			case KeyPress:
			case KeyRelease:
				sm.evt->keyId = XLookupKeysym(&e.xkey, 0);
				break;
			case ButtonPress:
			case ButtonRelease:
				switch (e.xbutton.button) {
					case Button1:
						sm.evt->mouseButtonId = MOUSE_BUTTON_LEFT;
						break;
					case Button2:
						sm.evt->mouseButtonId = MOUSE_BUTTON_MIDDLE;
						break;
					case Button3:
						sm.evt->mouseButtonId = MOUSE_BUTTON_RIGHT;
						break;
					case Button4:
						sm.evt->mouseButtonId = MOUSE_BUTTON_X1;
						break;
					case Button5:
						sm.evt->mouseButtonId = MOUSE_BUTTON_X2;
						break;
				}
				sm.evt->mouseX = e.xbutton.x;
				sm.evt->mouseY = e.xbutton.y;
				break;
			case MotionNotify:
				sm.evt->mouseButtonId = -1;
				sm.evt->mouseX = e.xmotion.x;
				sm.evt->mouseY = e.xmotion.y;
				break;
			case ClientMessage:
				if (filtered) {
					return;
				}
				const Atom protocol = e.xclient.data.l[0];
				if (protocol == WM_DELETE_WINDOW) {
					shared_memory_set_write_state(&sm, SHARED_MEM_QUIT);
				}
				break;
		}
		shared_memory_set_write_state(&sm, SHARED_MEM_WRITTEN);
	}
	XDestroyWindow(d, w);
	XCloseDisplay(d);
}

#endif
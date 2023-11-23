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
	esm[0] = SHARED_MEM_WRITTEN;
	XEvent e;
	while (esm[0] != SHARED_MEM_QUIT) {
		while (esm[0] != SHARED_MEM_AVAILABLE) {}
		esm[0] = SHARED_MEM_WRITING;
		void* esmData = esm + SHARED_MEM_DATA_START;
		XNextEvent(d, &e);
		bool filtered = XFilterEvent(&e, None);
		uint32_t msgType = e.type;
		memcpy(esmData, &msgType, sizeof(msgType));
		esmData += sizeof(msgType);
		InputEvent ie;
		switch (e.type) {
			case Expose:
				break;
			case KeyPress:
				break;
			case KeyRelease:
				break;
			case ButtonPress:
			case ButtonRelease:
				ie.mouseX = e.xbutton.x;
				ie.mouseY = e.xbutton.y;
				switch (e.xbutton.button) {
					case Button1:
						ie.mouseButtonId = MOUSE_BUTTON_LEFT;
						break;
					case Button2:
						ie.mouseButtonId = MOUSE_BUTTON_MIDDLE;
						break;
					case Button3:
						ie.mouseButtonId = MOUSE_BUTTON_RIGHT;
						break;
					case Button4:
						ie.mouseButtonId = MOUSE_BUTTON_X1;
						break;
					case Button5:
						ie.mouseButtonId = MOUSE_BUTTON_X2;
						break;
				}
				break;
			case MotionNotify:
				ie.mouseX = e.xmotion.x;
				ie.mouseY = e.xmotion.y;
				ie.mouseButtonId = -1;
				break;
			case ClientMessage:
				if (filtered) {
					return;
				}
				const Atom protocol = e.xclient.data.l[0];
				if (protocol == WM_DELETE_WINDOW) {
					esm[0] = SHARED_MEM_QUIT;
				}
				break;
		}
		if (esm[0] == SHARED_MEM_WRITING) {
			memcpy(esmData, &ie, sizeof(ie));
			esm[0] = SHARED_MEM_WRITTEN;
		}
	}
	XDestroyWindow(d, w);
	XCloseDisplay(d);
}

#endif
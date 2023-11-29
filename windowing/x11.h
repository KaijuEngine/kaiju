#ifndef WINDOWING_X11_H
#define WINDOWING_X11_H

#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <X11/Xlib.h>
#include "../gl/glad.h"
#include <GL/glx.h>

typedef GLXContext (*glXCreateContextAttribsARBProc)(Display*, GLXFBConfig, GLXContext, Bool, const int*);

typedef struct {
	Display* d;
	Window w;
	GLXFBConfig bestFbcConfig;
	GLXContext ctx;
} X11State;

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

void window_create_gl_context(void* state, void* evtSharedMem, int size) {
	X11State* x11State = state;
	char* esm = evtSharedMem;
	const char* glxExts = glXQueryExtensionsString(x11State->d, DefaultScreen(x11State->d));
	glXCreateContextAttribsARBProc glXCreateContextAttribsARB = (glXCreateContextAttribsARBProc)
		glXGetProcAddressARB((const GLubyte*)"glXCreateContextAttribsARB");
	x11State->ctx = NULL;
	// TODO:  Deal with ctx errors
	if (!isExtensionSupported(glxExts, "GLX_ARB_create_context") || !glXCreateContextAttribsARB) {
		x11State->ctx = glXCreateNewContext(x11State->d, x11State->bestFbcConfig, GLX_RGBA_TYPE, 0, True);
	} else {
		int contextAttrs[] = {
			GLX_CONTEXT_MAJOR_VERSION_ARB, 3,
			GLX_CONTEXT_MINOR_VERSION_ARB, 3,
			GLX_CONTEXT_FLAGS_ARB, GLX_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB,
			GLX_CONTEXT_PROFILE_MASK_ARB, GLX_CONTEXT_CORE_PROFILE_BIT_ARB,
			None
		};
		x11State->ctx = glXCreateContextAttribsARB(x11State->d, x11State->bestFbcConfig, 0, True, contextAttrs);
		XSync(x11State->d, False);
		// TODO:  Check ctx errors and if so, then do the following
		//context_attribs[1] = 1;
		//context_attribs[3] = 0;
		//x11State->ctx = glXCreateContextAttribsARB(d, bestFbc, 0, True, context_attribs);
	}
	XSync(x11State->d, False);
	//XSetErrorHandler(oldHandler);
	// TODO:  Check context error as well as ctx
	if (x11State->ctx == NULL) {
		write_fatal(evtSharedMem, size, "Failed to create GL context");
		return;
	}
	glXMakeCurrent(x11State->d, x11State->w, x11State->ctx);
	if (gladLoadGL() == 0) {
		write_fatal(evtSharedMem, size, "Failed to load OpenGL");
		return;
	}
}

void window_main(const char* windowTitle, void* evtSharedMem, int size) {
	char* esm = evtSharedMem;
	SharedMem sm = {evtSharedMem, size};
	Display* d = XOpenDisplay(NULL);
	if (d == NULL) {
		write_fatal(evtSharedMem, size, "Failed to open display");
		return;
	}
	int visAttrs[] = {
		GLX_X_RENDERABLE, True,
		GLX_DRAWABLE_TYPE, GLX_WINDOW_BIT,
		GLX_RENDER_TYPE, GLX_RGBA_BIT,
		GLX_X_VISUAL_TYPE, GLX_TRUE_COLOR,
		GLX_RED_SIZE, 8,
		GLX_GREEN_SIZE, 8,
		GLX_BLUE_SIZE, 8,
		GLX_ALPHA_SIZE, 8,
		GLX_DEPTH_SIZE, 24,
		GLX_STENCIL_SIZE, 8,
		GLX_DOUBLEBUFFER, True,
		None
	};
	int glxMajor, glxMinor;
	if (!glXQueryVersion(d, &glxMajor, &glxMinor) ||
		((glxMajor == 1) && (glxMinor < 3)) || (glxMajor < 1))
	{
		write_fatal(evtSharedMem, size, "Invalid GLX version");
		return;
	}
	int fbCount;
	GLXFBConfig* fbc = glXChooseFBConfig(d, DefaultScreen(d), visAttrs, &fbCount);
	int bestFbc = -1, worstFbc = -1, bestNumSamp = -1, worstNumSamp = 999;
	for (int i = 0; i < fbCount; i++) {
		XVisualInfo* vi = glXGetVisualFromFBConfig(d, fbc[i]);
		if (vi != NULL) {
			int sampBuf, samples;
			glXGetFBConfigAttrib(d, fbc[i], GLX_SAMPLE_BUFFERS, &sampBuf);
			glXGetFBConfigAttrib(d, fbc[i], GLX_SAMPLES, &samples);
			if (bestFbc < 0 || (sampBuf && samples > bestNumSamp)) {
				bestFbc = i;
				bestNumSamp = samples;
			}
			if (worstFbc < 0 || !sampBuf || samples < worstNumSamp) {
				worstFbc = i;
				worstNumSamp = samples;
			}
		}
		XFree(vi);
	}
	GLXFBConfig bestFbcConfig = fbc[bestFbc];
	XFree(fbc);
	XVisualInfo* vi = glXGetVisualFromFBConfig(d, bestFbcConfig);
	XSetWindowAttributes swa;
	Colormap cmap = XCreateColormap(d, RootWindow(d, vi->screen), vi->visual, AllocNone);
	swa.colormap = cmap;
	swa.background_pixmap = None;
	swa.border_pixel = 0;
	swa.event_mask = StructureNotifyMask;
	Window w = XCreateWindow(d, RootWindow(d, vi->screen), 10, 10, 1280, 720,
		0, vi->depth, InputOutput, vi->visual, CWBorderPixel | CWColormap | CWEventMask, &swa);
	if (w == None) {
		write_fatal(evtSharedMem, size, "Failed to create window");
		return;
	}
	XFree(vi);
	XStoreName(d, w, windowTitle);
	XSetIconName(d, w, windowTitle);
	XSelectInput(d, w, ExposureMask | KeyPressMask | KeyReleaseMask | ButtonPressMask | ButtonReleaseMask | PointerMotionMask);
	XMapWindow(d, w);
	X11State x11State = {d, w, bestFbcConfig};
	X11State* cpyState = &x11State;
	memcpy(esm+SHARED_MEM_DATA_START, &cpyState, sizeof(X11State**));
	shared_memory_set_write_state(&sm, SHARED_MEM_AWAITING_CONTEXT);
	// Context should be created in Go here on go main thread
	shared_memory_wait_for_available(&sm);
	Atom WM_DELETE_WINDOW = XInternAtom(d, "WM_DELETE_WINDOW", False);
	XSetWMProtocols(d, w, &WM_DELETE_WINDOW, 1);
	shared_memory_set_write_state(&sm, SHARED_MEM_AWAITING_START);
	XEvent e;
	while (esm[0] != SHARED_MEM_QUIT) {
		while (esm[0] != SHARED_MEM_AVAILABLE) {}
		shared_memory_set_write_state(&sm, SHARED_MEM_WRITING);
		XNextEvent(d, &e);
		bool filtered = XFilterEvent(&e, None);
		sm.evt->evtType = e.type;
		switch (e.type) {
			case Expose:
				// Setting this to 0 will make Go think there are no more
				// events to process and continue. Then next time we come
				// back around in go for availability we'll continue from
				// here.
				sm.evt->evtType = 0;
				shared_memory_set_write_state(&sm, SHARED_MEM_WRITTEN);
				shared_memory_wait_for_available(&sm);
				shared_memory_set_write_state(&sm, SHARED_MEM_WRITING);
				sm.evt->evtType = e.type;
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
	glXMakeCurrent(d, 0, 0);
	glXDestroyContext(d, x11State.ctx);
	XDestroyWindow(d, w);
	XFreeColormap(d, cmap);
	XCloseDisplay(d);
}

#endif
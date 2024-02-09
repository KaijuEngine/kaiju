#if defined(__linux__) || defined(__unix__) || defined(__APPLE__)

#include "x11.h"
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <pthread.h>
#include <X11/Xlib.h>

#ifdef OPENGL
#include "../gl/dist/glad.h"
#include <GL/glx.h>

typedef GLXContext (*glXCreateContextAttribsARBProc)(Display*, GLXFBConfig, GLXContext, Bool, const int*);
#endif

#define EVT_MASK	ExposureMask | KeyPressMask | KeyReleaseMask | ButtonPressMask | ButtonReleaseMask | PointerMotionMask

int shared_mem_set_thread_priority(SharedMem* sm) {
	// TODO:  Get current thread priority and set the current thread priority to idle
	return 0;
}

void shared_mem_reset_thread_priority(SharedMem* sm, int priority) {
	// TODO:  Set the current thread priority to the given priority
}

void shared_mem_wait(SharedMem* sm) {
	sched_yield();
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

#ifdef OPENGL
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
#endif

void window_main(const char* windowTitle, int width, int height, void* evtSharedMem, int size) {
	Display* d = XOpenDisplay(NULL);
	if (d == NULL) {
		write_fatal(evtSharedMem, size, "Failed to open display");
		return;
	}
#ifdef OPENGL
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
	Colormap cmap = XCreateColormap(d, RootWindow(d, vi->screen), NULL, AllocNone);
	swa.colormap = cmap;
	swa.background_pixmap = None;
	swa.border_pixel = 0;
	swa.event_mask = StructureNotifyMask;
	Window w = XCreateWindow(d, RootWindow(d, vi->screen), 10, 10, width, height,
		0, vi->depth, InputOutput, vi->visual, CWBorderPixel | CWColormap | CWEventMask, &swa);
#else
	Window w = XCreateSimpleWindow(d, RootWindow(d, DefaultScreen(d)), 10, 10,
		width, height, 1, BlackPixel(d, DefaultScreen(d)), WhitePixel(d, DefaultScreen(d)));
#endif
	if (w == None) {
		write_fatal(evtSharedMem, size, "Failed to create window");
		return;
	}
#ifdef OPENGL
	XFree(vi);
#endif
	XStoreName(d, w, windowTitle);
	XSetIconName(d, w, windowTitle);
	XSelectInput(d, w, EVT_MASK);
	XMapWindow(d, w);
	X11State* x11State = malloc(sizeof(X11State));
	x11State->sm.sharedMem = evtSharedMem;
	x11State->sm.size = size;
	x11State->w = w;
	x11State->d = d;
#ifdef OPENGL
	x11State->bestFbcConfig = bestFbcConfig;
	x11State->cmap = cmap;
#endif
	x11State->WM_DELETE_WINDOW = XInternAtom(d, "WM_DELETE_WINDOW", False);
	XSetWMProtocols(d, w, &x11State->WM_DELETE_WINDOW, 1);
	memcpy(evtSharedMem+SHARED_MEM_DATA_START, &x11State, sizeof(x11State));
}

void window_show(void* x11State) {
	X11State* s = x11State;
	// Flush initial events
	while (window_poll(x11State) != Expose) {}
	while (window_poll(x11State) != 0) {}
}

int window_poll_controller(void* x11State) {
	// TODO:  Implement for controllers
	return 0;
}

int window_poll(void* x11State) {
	X11State* s = x11State;
	XEvent e = {};

	if (!XCheckMaskEvent(s->d, EVT_MASK, &e)) {
		return 0;
	}
	bool filtered = XFilterEvent(&e, None);
	s->sm.evt->evtType = e.type;
	switch (e.type) {
		case Expose:
			//return 0;
			break;
		case KeyPress:
		case KeyRelease:
			s->sm.evt->keyboard.keyId = XLookupKeysym(&e.xkey, 0);
			break;
		case ButtonPress:
		case ButtonRelease:
			switch (e.xbutton.button) {
				case Button1:
					s->sm.evt->mouse.mouseButtonId = MOUSE_BUTTON_LEFT;
					break;
				case Button2:
					s->sm.evt->mouse.mouseButtonId = MOUSE_BUTTON_MIDDLE;
					break;
				case Button3:
					s->sm.evt->mouse.mouseButtonId = MOUSE_BUTTON_RIGHT;
					break;
				case Button4:
					s->sm.evt->mouse.mouseButtonId = MOUSE_BUTTON_X1;
					break;
				case Button5:
					s->sm.evt->mouse.mouseButtonId = MOUSE_BUTTON_X2;
					break;
			}
			s->sm.evt->mouse.mouseX = e.xbutton.x;
			s->sm.evt->mouse.mouseY = e.xbutton.y;
			break;
		case MotionNotify:
			s->sm.evt->mouse.mouseButtonId = -1;
			s->sm.evt->mouse.mouseX = e.xmotion.x;
			s->sm.evt->mouse.mouseY = e.xmotion.y;
			break;
		case ClientMessage:
			if (!filtered) {
				const Atom protocol = e.xclient.data.l[0];
				if (protocol == s->WM_DELETE_WINDOW) {
					shared_memory_set_write_state(&s->sm, SHARED_MEM_QUIT);
				}
			}
			break;
	}
	return e.type;
}

void window_destroy(void* x11State) {
	X11State* s = x11State;
#ifdef OPENGL
	glXMakeCurrent(s->d, 0, 0);
	glXDestroyContext(s->d, s->ctx);
#endif
	XDestroyWindow(s->d, s->w);
#ifdef OPENGL
	XFreeColormap(s->d, s->cmap);
#endif
	XCloseDisplay(s->d);
}

void* display(void* x11State) { return ((X11State*)x11State)->d; }
void* window(void* x11State) { return &((X11State*)x11State)->w; }

#endif

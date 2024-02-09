#ifndef WINDOWING_X11_H
#define WINDOWING_X11_H

#include <X11/Xlib.h>
#include "shared_mem.h"

#ifdef OPENGL
#include "../gl/dist/glad.h"
#include <GL/glx.h>
#endif

typedef struct {
	SharedMem sm;
	Window w;
	Display* d;
	Atom WM_DELETE_WINDOW;
#ifdef OPENGL
	GLXFBConfig bestFbcConfig;
	Colormap cmap
	GLXContext ctx;
#endif
} X11State;

void window_main(const char* windowTitle, int width, int height, void* evtSharedMem, int size);
void window_show(void* x11State);
int window_poll_controller(void* x11State);
int window_poll(void* x11State);
void window_destroy(void* x11State);
void* display(void* x11State);
void* window(void* x11State);

#ifdef OPENGL
void window_create_gl_context(void* state, void* evtSharedMem, int size);
#endif

#endif
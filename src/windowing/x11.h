#ifndef WINDOWING_X11_H
#define WINDOWING_X11_H

#include <X11/Xlib.h>

#ifdef OPENGL
#include "../gl/dist/glad.h"
#include <GL/glx.h>
#endif

typedef struct {
	Window* w;
	Display* d;
#ifdef OPENGL
	GLXFBConfig bestFbcConfig;
	GLXContext ctx;
#endif
} X11State;

void window_main(const char* windowTitle, int width, int height, void* evtSharedMem, int size);

#ifdef OPENGL
void window_create_gl_context(void* state, void* evtSharedMem, int size);
#endif

#endif
#ifndef WINDOWING_X11_H
#define WINDOWING_X11_H

void window_main(const char* windowTitle, int width, int height, void* evtSharedMem, int size);

#ifdef OPENGL
void window_create_gl_context(void* state, void* evtSharedMem, int size);
#endif

#endif
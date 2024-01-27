#ifndef WINDOWING_WIN32_H
#define WINDOWING_WIN32_H

#include <wchar.h>

void window_main(const wchar_t* windowTitle, int width, int height, void* evtSharedMem, int size);

#ifdef OPENGL
void window_create_gl_context(void* winHWND, void* evtSharedMem, int size);
#endif

#endif
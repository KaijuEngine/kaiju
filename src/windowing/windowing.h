#ifndef WINDOWING_H
#define WINDOWING_H

#if defined(_WIN32) || defined(_WIN64)
#include "win32.h"
#elif defined(__linux__) || defined(__unix__) || defined(__APPLE__)
#include "x11.h"
#endif

#endif
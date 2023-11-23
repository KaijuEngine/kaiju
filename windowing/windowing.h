#ifndef WINDOWING_H
#define WINDOWING_H

#include <stdint.h>

#define SHARED_MEM_AVAILABLE	0
#define SHARED_MEM_WRITING		1
#define SHARED_MEM_WRITTEN		2
#define SHARED_MEM_FATAL		0xFE
#define SHARED_MEM_QUIT			0xFF
#define SHARED_MEM_DATA_START	4

typedef struct {
	union {
		int32_t mouseX;
		int32_t key;
	};
	int32_t mouseY;
	int32_t mouseXButton;
} InputEvent;

#if defined(_WIN32) || defined(_WIN64)
#include "win32.h"
#elif defined(__linux__) || defined(__unix__) || defined(__APPLE__)
#include "x11.h"
#endif

#endif
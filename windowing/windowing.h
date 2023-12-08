#ifndef WINDOWING_H
#define WINDOWING_H

#include <stdint.h>
#include <string.h>

#define SHARED_MEM_AVAILABLE		0
#define SHARED_MEM_WRITING			1
#define SHARED_MEM_WRITTEN			2
#define SHARED_MEM_AWAITING_CONTEXT	0xFC
#define SHARED_MEM_AWAITING_START	0xFD
#define SHARED_MEM_FATAL			0xFE
#define SHARED_MEM_QUIT				0xFF
#define SHARED_MEM_DATA_START		4

#define MOUSE_BUTTON_LEFT			0
#define MOUSE_BUTTON_MIDDLE			1
#define MOUSE_BUTTON_RIGHT			2
#define MOUSE_BUTTON_X1				3
#define MOUSE_BUTTON_X2				4

typedef struct {
	union {
		uint8_t writeState;
		int32_t buffer;
	};
	uint32_t evtType;
	union {
		int32_t mouseButtonId;
		int32_t keyId;
	};
	int32_t mouseX;
	int32_t mouseY;
} InputEvent;

typedef struct {
	union {
		char* sharedMem;
		InputEvent* evt;
	};
	int size;
} SharedMem;

int shared_mem_set_thread_priority(SharedMem* sm);
void shared_mem_reset_thread_priority(SharedMem* sm, int priority);
void shared_mem_wait(SharedMem* sm);

void shared_memory_wait_for_available(SharedMem* sm) {
	int priority = shared_mem_set_thread_priority(sm);
	while (sm->evt->writeState != SHARED_MEM_AVAILABLE) {
		shared_mem_wait(sm);
	}
	shared_mem_reset_thread_priority(sm, priority);
}

void shared_memory_set_write_state(SharedMem* sm, uint8_t state) {
	uint8_t smState = sm->evt->writeState;
	if (smState != SHARED_MEM_QUIT && smState != SHARED_MEM_FATAL) {
		sm->evt->writeState = state;
	}
}

void write_fatal(char* evtSharedMem, int size, const char* msg) {
	int msgLen = strlen(msg);
	if (msgLen > size - SHARED_MEM_DATA_START) {
		msgLen = size - SHARED_MEM_DATA_START - 1;
	}
	memcpy(evtSharedMem + SHARED_MEM_DATA_START, msg, msgLen);
	evtSharedMem[SHARED_MEM_DATA_START + msgLen] = '\0';
	evtSharedMem[0] = SHARED_MEM_FATAL;
}

#if defined(_WIN32) || defined(_WIN64)
#include "win32.h"
#elif defined(__linux__) || defined(__unix__) || defined(__APPLE__)
#include "x11.h"
#endif

#endif
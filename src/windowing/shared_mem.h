#ifndef SHARED_MEM_H
#define SHARED_MEM_H

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

#define EVENT_TYPE_CONTROLLER		-2

#define MOUSE_BUTTON_LEFT			0
#define MOUSE_BUTTON_MIDDLE			1
#define MOUSE_BUTTON_RIGHT			2
#define MOUSE_BUTTON_X1				3
#define MOUSE_BUTTON_X2				4

#if defined(_WIN32) || defined(_WIN64)
#include <xinput.h>
#define MAX_CONTROLLERS				XUSER_MAX_COUNT
#else
// TODO:  Get the correct value for X11
#define MAX_CONTROLLERS				4
#endif

typedef struct {
	int32_t mouseButtonId;
	int32_t mouseX;
	int32_t mouseY;
} MouseEvent;

typedef struct {
	int32_t keyId;
} KeyboardEvent;

typedef struct {
	int32_t width;
	int32_t height;
} ResizeEvent;

typedef struct {
	uint16_t buttons;
	int16_t thumbLX;
	int16_t thumbLY;
	int16_t thumbRX;
	int16_t thumbRY;
	uint8_t leftTrigger;
	uint8_t rightTrigger;
	uint8_t isConnected;
} ControllerState;

typedef struct {
	ControllerState states[MAX_CONTROLLERS];
} ControllerEvent;

typedef struct {
	union {
		uint8_t writeState;
		int32_t buffer;
	};
	uint32_t evtType;
	union {
		MouseEvent mouse;
		KeyboardEvent keyboard;
		ResizeEvent resize;
		ControllerEvent controllers;
	};
} InputEvent;

typedef struct {
	union {
		uint8_t* sharedMem;
		InputEvent* evt;
	};
	int size;
} SharedMem;

int shared_mem_set_thread_priority(SharedMem* sm);
void shared_mem_reset_thread_priority(SharedMem* sm, int priority);
void shared_mem_wait(SharedMem* sm);

static inline void shared_memory_wait_for_available(SharedMem* sm) {
	int priority = shared_mem_set_thread_priority(sm);
	while (sm->evt->writeState != SHARED_MEM_AVAILABLE) {
		shared_mem_wait(sm);
	}
	shared_mem_reset_thread_priority(sm, priority);
}

static inline void shared_memory_set_write_state(SharedMem* sm, uint8_t state) {
	uint8_t smState = sm->evt->writeState;
	if (smState != SHARED_MEM_QUIT && smState != SHARED_MEM_FATAL) {
		sm->evt->writeState = state;
	}
}

static inline void write_fatal(char* evtSharedMem, int size, const char* msg) {
	int msgLen = strlen(msg);
	if (msgLen > size - SHARED_MEM_DATA_START) {
		msgLen = size - SHARED_MEM_DATA_START - 1;
	}
	memcpy(evtSharedMem + SHARED_MEM_DATA_START, msg, msgLen);
	evtSharedMem[SHARED_MEM_DATA_START + msgLen] = '\0';
	evtSharedMem[0] = SHARED_MEM_FATAL;
}

#endif

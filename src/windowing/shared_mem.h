/******************************************************************************/
/* shared_mem.h                                                              */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

#ifndef SHARED_MEM_H
#define SHARED_MEM_H

#include <stdint.h>
#include <string.h>

#define SHARED_MEM_WINDOW_ACTIVITY	0xF9
#define SHARED_MEM_WINDOW_MOVE		0xFA
#define SHARED_MEM_WINDOW_RESIZE	0xFB
#define SHARED_MEM_AWAITING_CONTEXT	0xFC
#define SHARED_MEM_AWAITING_START	0xFD
#define SHARED_MEM_FATAL			0xFE
#define SHARED_MEM_QUIT				0xFF
#define SHARED_MEM_DATA_START		4

#define EVENT_TYPE_CONTROLLER		-2

#define MOUSE_WHEEL_HORIZONTAL		-2
#define MOUSE_WHEEL_VERTICAL		-1
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
	int32_t value;
} EnumEvent;

typedef struct {
	int32_t mouseButtonId;
	int32_t mouseX;
	int32_t mouseY;
	int32_t wheelDelta;
} MouseEvent;

typedef struct {
	int32_t keyId;
} KeyboardEvent;

typedef struct {
	int32_t width;
	int32_t height;
} ResizeEvent;

typedef struct {
	int32_t x;
	int32_t y;
} MoveEvent;

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
		EnumEvent enumEvent;
		MouseEvent mouse;
		KeyboardEvent keyboard;
		MoveEvent move;
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
	int windowWidth;
	int windowHeight;
} SharedMem;

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

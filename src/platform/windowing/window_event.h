/******************************************************************************/
/* window_event.h                                                             */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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

#ifndef WINDOW_EVENT_H
#define WINDOW_EVENT_H

#include <stdint.h>
#include <stdalign.h>

#define WINDOW_EVENT_BUFFER_SIZE	128

typedef enum {
	WINDOW_EVENT_TYPE_SET_HANDLE = 1,
	WINDOW_EVENT_TYPE_ACTIVITY = 2,
	WINDOW_EVENT_TYPE_MOVE = 3,
	WINDOW_EVENT_TYPE_RESIZE = 4,
	WINDOW_EVENT_TYPE_MOUSE_MOVE = 5,
	WINDOW_EVENT_TYPE_MOUSE_SCROLL = 6,
	WINDOW_EVENT_TYPE_MOUSE_BUTTON = 7,
	WINDOW_EVENT_TYPE_KEYBOARD_BUTTON = 8,
	WINDOW_EVENT_TYPE_CONTROLLER_STATE = 9,
	WINDOW_EVENT_TYPE_TOUCH_STATE = 10,
	WINDOW_EVENT_TYPE_STYLUS_STATE = 11,
	WINDOW_EVENT_TYPE_FATAL = 12,
} WindowEventType;

typedef enum {
	WINDOW_EVENT_ACTIVITY_TYPE_MINIMIZE = 1,
	WINDOW_EVENT_ACTIVITY_TYPE_MAXIMIZE = 2,
	WINDOW_EVENT_ACTIVITY_TYPE_CLOSE = 3,
	WINDOW_EVENT_ACTIVITY_TYPE_FOCUS = 4,
	WINDOW_EVENT_ACTIVITY_TYPE_BLUR = 5,
} WindowEventActivityType;

typedef enum {
	WINDOW_EVENT_BUTTON_TYPE_DOWN = 1,
	WINDOW_EVENT_BUTTON_TYPE_UP = 2,
} WindowEventButtonType;

typedef enum {
	WINDOW_EVENT_CONTROLLER_CONNECTION_TYPE_DISCONNECTED = 1,
	WINDOW_EVENT_CONTROLLER_CONNECTION_TYPE_CONNECTED = 2,
} WindowEventControllerConnectionType;

typedef enum {
	TOUCH_ACTION_STATE_TYPE_UP = 1,
	TOUCH_ACTION_STATE_TYPE_MOVE = 2,
	TOUCH_ACTION_STATE_TYPE_CANCEL = 3,
} TouchActionStateType;

typedef enum {
	STYLUS_ACTION_STATE_TYPE_NONE = 1,
	STYLUS_ACTION_STATE_TYPE_HOVER_ENTER = 2,
	STYLUS_ACTION_STATE_TYPE_HOVER_MOVE = 3,
	STYLUS_ACTION_STATE_TYPE_HOVER_EXIT = 4,
	STYLUS_ACTION_STATE_TYPE_DOWN = 5,
	STYLUS_ACTION_STATE_TYPE_MOVE = 6,
	STYLUS_ACTION_STATE_TYPE_UP = 7,
} StylusActionStateType;

typedef struct {
	alignas(8) void* hwnd;
	alignas(8) void* instance;
} SetHandleEvent;

typedef struct {
	alignas(4) WindowEventActivityType activityType;
	char _[4];
} WindowActivityEvent;

typedef struct {
	alignas(4) int32_t x;
	alignas(4) int32_t y;
} WindowMoveEvent;

typedef struct {
	alignas(4) int32_t width;
	alignas(4) int32_t height;
	alignas(4) int32_t left;
	alignas(4) int32_t top;
	alignas(4) int32_t right;
	alignas(4) int32_t bottom;
} WindowResizeEvent;

typedef struct {
	alignas(4) int32_t x;
	alignas(4) int32_t y;
} MouseMoveWindowEvent;

typedef struct {
	alignas(4) int32_t deltaX;
	alignas(4) int32_t deltaY;
	alignas(4) int32_t x;
	alignas(4) int32_t y;
} MouseScrollWindowEvent;

typedef struct {
	alignas(4) int32_t buttonId;
	alignas(4) WindowEventButtonType action;
	alignas(4) int32_t x;
	alignas(4) int32_t y;
} MouseButtonWindowEvent;

typedef struct {
	alignas(4) int32_t buttonId;
	alignas(4) WindowEventButtonType action;
} KeyboardButtonWindowEvent;

typedef struct {
	alignas(1) uint8_t controllerId;
	alignas(1) uint8_t leftTrigger;
	alignas(1) uint8_t rightTrigger;
	char _0[1];
	alignas(4) WindowEventControllerConnectionType connectionType;
	alignas(2) uint16_t buttons;
	alignas(2) int16_t thumbLX;
	alignas(2) int16_t thumbLY;
	alignas(2) int16_t thumbRX;
	alignas(2) int16_t thumbRY;
	char _1[6];
} ControllerStateWindowEvent;

typedef struct {
	alignas(4) float x;
	alignas(4) float y;
	alignas(4) float pressure;
	alignas(4) int index;
	alignas(4) int actionState;
	char _1[4];
} TouchStateWindowEvent;

typedef struct {
	alignas(4) float x;
	alignas(4) float y;
	alignas(4) float pressure;
	alignas(4) float distance;
	alignas(4) int actionState;
	char _1[4];
} StylusStateWindowEvent;

typedef struct {
	alignas(8) WindowEventType type;
	union {
		SetHandleEvent setHandle;
		WindowActivityEvent windowActivity;
		WindowMoveEvent windowMove;
		WindowResizeEvent windowResize;
		MouseMoveWindowEvent mouseMove;
		MouseScrollWindowEvent mouseScroll;
		MouseButtonWindowEvent mouseButton;
		KeyboardButtonWindowEvent keyboardButton;
		ControllerStateWindowEvent controllerState;
		TouchStateWindowEvent touchState;
		StylusStateWindowEvent stylusState;
	};
} WindowEvent;

#endif

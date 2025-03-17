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
	WINDOW_EVENT_TYPE_FATAL = 10,
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
	};
} WindowEvent;

#endif

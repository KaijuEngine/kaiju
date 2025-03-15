#ifndef WINDOW_EVENT_H
#define WINDOW_EVENT_H

#include <stdint.h>

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
	void* hwnd;
	void* instance;
} SetHandleEvent;

typedef struct {
	WindowEventActivityType activityType;
} WindowActivityEvent;

typedef struct {
	int32_t x;
	int32_t y;
} WindowMoveEvent;

typedef struct {
	int32_t width;
	int32_t height;
	int32_t left;
	int32_t top;
	int32_t right;
	int32_t bottom;
} WindowResizeEvent;

typedef struct {
	int32_t x;
	int32_t y;
} MouseMoveWindowEvent;

typedef struct {
	int32_t deltaX;
	int32_t deltaY;
	int32_t x;
	int32_t y;
} MouseScrollWindowEvent;

typedef struct {
	int32_t buttonId;
	WindowEventButtonType action;
	int32_t x;
	int32_t y;
} MouseButtonWindowEvent;

typedef struct {
	int32_t buttonId;
	WindowEventButtonType action;
} KeyboardButtonWindowEvent;

typedef struct {
	uint8_t controllerId;
	WindowEventControllerConnectionType connectionType;
	uint16_t buttons;
	int16_t thumbLX;
	int16_t thumbLY;
	int16_t thumbRX;
	int16_t thumbRY;
	uint8_t leftTrigger;
	uint8_t rightTrigger;
} ControllerStateWindowEvent;

typedef struct {
	//WindowEventType type;
	uint32_t type;
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

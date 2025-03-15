#ifndef WINDOW_EVENT_H
#define WINDOW_EVENT_H

#include <stdint.h>

#define WINDOW_EVENT_BUFFER_SIZE	100

typedef enum {
	WINDOW_EVENT_TYPE_WINDOW_ACTIVITY,
	WINDOW_EVENT_TYPE_WINDOW_MOVE,
	WINDOW_EVENT_TYPE_WINDOW_RESIZE,
	WINDOW_EVENT_TYPE_MOUSE_MOVE,
	WINDOW_EVENT_TYPE_MOUSE_SCROLL,
	WINDOW_EVENT_TYPE_MOUSE_BUTTON,
	WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
	WINDOW_EVENT_TYPE_CONTROLLER_CONNECTED,
	WINDOW_EVENT_TYPE_CONTROLLER_BUTTON,
	WINDOW_EVENT_TYPE_CONTROLLER_STICK,
	WINDOW_EVENT_TYPE_CONTROLLER_TRIGGER,
} WindowEventType;

typedef enum {
	WINDOW_EVENT_ACTIVITY_TYPE_MINIMIZE,
	WINDOW_EVENT_ACTIVITY_TYPE_MAXIMIZE,
	WINDOW_EVENT_ACTIVITY_TYPE_CLOSE,
	WINDOW_EVENT_ACTIVITY_TYPE_FOCUS,
	WINDOW_EVENT_ACTIVITY_TYPE_BLUR,
} WindowEventActivityType;

typedef enum {
	WINDOW_EVENT_BUTTON_TYPE_DOWN,
	WINDOW_EVENT_BUTTON_TYPE_UP,
} WindowEventButtonType;

typedef struct {
	WindowEventActivityType activityType;
} WindowActivityEvent;

typedef struct {
	int32_t x;
	int32_t y;
} MouseMoveWindowEvent;

typedef struct {
	int32_t buttonId;
	WindowEventType action;
} MouseButtonWindowEvent;

typedef struct {
	int32_t buttonId;
	WindowEventType action;
} KeyboardButtonWindowEvent;

typedef struct {
	WindowEventType type;
	union {
		MouseMoveWindowEvent mouseMove;
		MouseButtonWindowEvent mouseButton;
		KeyboardButtonWindowEvent keyboardButton;

	}
} WindowEvent;



#endif

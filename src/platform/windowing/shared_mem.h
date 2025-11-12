

#ifndef SHARED_MEM_H
#define SHARED_MEM_H

#include <stdbool.h>

#include "window_event.h"

#define SHARED_MEM_NONE				0x00
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

extern void goProcessEvents(uint64_t goWindow, void* events, uint32_t eventCount);

#if __linux__
typedef long LONG;
typedef struct {
	LONG left;
	LONG top;
	LONG right;
	LONG bottom;
} RECT;
#endif

typedef struct {
	void* goWindow;
	int x;
	int y;
	int windowWidth;
	int windowHeight;
	int left;
	int top;
	int right;
	int bottom;
	RECT clientRect;
	int mouseX;
	int mouseY;
	bool rawInputFailed;
	bool rawInputRequested;
	struct {
		LONG style;
		LONG exStyle;
		RECT rect;
		int borderWidth;
		int overrideRedirect;
	} savedState;
	uint32_t eventCount;
	struct {
		int x;
		int y;
		bool active;
	} lockCursor;
	WindowEvent events[WINDOW_EVENT_BUFFER_SIZE];
#if defined(__android__)
	const ASensor* accelerometer;
	ASensorManager* sensorManager;
	ASensorEventQueue* sensorQueue;
#else
	size_t _0[3]; // Keep structure size consistant between platforms
#endif
} SharedMem;

static inline void shared_mem_flush_events(SharedMem* mem) {
	if (mem->eventCount == 0) {
		return;
	}
	goProcessEvents((uint64_t)mem->goWindow, mem->events, mem->eventCount);
	mem->eventCount = 0;
}

static inline void shared_mem_add_event(SharedMem* mem, WindowEvent evt) {
	mem->events[mem->eventCount++] = evt;
	if (mem->eventCount == WINDOW_EVENT_BUFFER_SIZE) {
		shared_mem_flush_events(mem);
	}
}

#endif

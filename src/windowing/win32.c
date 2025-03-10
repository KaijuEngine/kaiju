/******************************************************************************/
/* win32.c                                                                   */
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

#if defined(_WIN32) || defined(_WIN64)

#ifndef WIN32_LEAN_AND_MEAN
#define WIN32_LEAN_AND_MEAN
#endif

#ifndef WINVER
#define WINVER 0x0605
#endif

#ifndef UNICODE
#define UNICODE
#endif

#include "shared_mem.h"
#include "strings.h"

#include "win32.h"
#include <string.h>
#include <windows.h>
#include <windowsx.h>

/*
* Messages defined here are NOT to be sent to other windows
* https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-registerwindowmessagea#remarks
*/
#define UWM_SET_CURSOR		(WM_USER + 0x0001)
#define CURSOR_ARROW        1
#define CURSOR_IBEAM        2
#define CURSOR_WAIT         3
#define CURSOR_CROSS        4
#define CURSOR_UPARROW      5
#define CURSOR_SIZE_NWSE    6
#define CURSOR_SIZE_NESW    7
#define CURSOR_SIZE_WE      8
#define CURSOR_SIZE_NS      9
#define CURSOR_SIZE_ALL     10
#define CURSOR_NO           11
#define CURSOR_HAND         12
#define CURSOR_APP_STARTING 13
#define CURSOR_HELP         14
#define CURSOR_PIN          15
#define CURSOR_PERSON       16

void setMouseEvent(InputEvent* evt, LPARAM lParam, int buttonId) {
	evt->mouse.mouseButtonId = buttonId;
	evt->mouse.mouseX = GET_X_LPARAM(lParam);
	evt->mouse.mouseY = GET_Y_LPARAM(lParam);
}

bool obtainControllerStates(SharedMem* sm) {
	bool readControllerStates = false;
	DWORD dwResult;
	memset(&sm->evt->controllers, 0, sizeof(ControllerEvent));
	for (DWORD i = 0; i < MAX_CONTROLLERS; i++) {
		XINPUT_STATE state;
		ZeroMemory(&state, sizeof(XINPUT_STATE));
		// Simply get the state of the controller from XInput.
		dwResult = XInputGetState(i, &state);
		if(dwResult == ERROR_SUCCESS) {
			sm->evt->controllers.states[i].buttons = state.Gamepad.wButtons;
			sm->evt->controllers.states[i].leftTrigger = state.Gamepad.bLeftTrigger;
			sm->evt->controllers.states[i].rightTrigger = state.Gamepad.bRightTrigger;
			sm->evt->controllers.states[i].thumbLX = state.Gamepad.sThumbLX;
			sm->evt->controllers.states[i].thumbLY = state.Gamepad.sThumbLY;
			sm->evt->controllers.states[i].thumbRX = state.Gamepad.sThumbRX;
			sm->evt->controllers.states[i].thumbRY = state.Gamepad.sThumbRY;
			sm->evt->controllers.states[i].isConnected = 1;
			readControllerStates = true;
		} else {
			// TODO:  readControllerStates would be true here too, but
			// no need to spam the event if no controllers are available?
			// Probably means the state of the controllers need tracking in C...
			sm->evt->controllers.states[i].isConnected = 0;
		}
	}
	return readControllerStates;
}

LRESULT CALLBACK window_proc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
	SharedMem* sm = (SharedMem*)GetWindowLongPtrA(hwnd, GWLP_USERDATA);
	switch (uMsg) {
		case WM_DESTROY:
			if (sm != NULL) {
				shared_memory_set_write_state(sm, SHARED_MEM_QUIT);
			}
			PostQuitMessage(0);
			return 0;
		case WM_ACTIVATE:
			switch (LOWORD(wParam)) {
				case WA_ACTIVE:
				case WA_CLICKACTIVE:
					sm->evt->enumEvent.value = 1;
					break;
				case WA_INACTIVE:
					sm->evt->enumEvent.value = 0;
					break;
			}
			shared_memory_set_write_state(sm, SHARED_MEM_WINDOW_ACTIVITY);
			break;
		case WM_MOVE:
			sm->evt->move.x = (int)(short)LOWORD(lParam);
			sm->evt->move.y = (int)(short)HIWORD(lParam);
			shared_memory_set_write_state(sm, SHARED_MEM_WINDOW_MOVE);
			break;
		case WM_SIZE:
			if (sm != NULL) {
				RECT clientArea;
				GetClientRect(hwnd, &clientArea);
				LONG width = clientArea.right-clientArea.left;
				LONG height = clientArea.bottom-clientArea.top;
				if (sm->windowWidth != width || sm->windowHeight != height) {
					sm->windowWidth = width;
					sm->windowHeight = height;
					sm->evt->resize.width = width;
					sm->evt->resize.height = height;
					shared_memory_set_write_state(sm, SHARED_MEM_WINDOW_RESIZE);
				}
			}
			PostMessage(hwnd, WM_PAINT, 0, 0);
			break;
	}
	return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

void process_message(SharedMem* sm, MSG *msg) {
	sm->evt->evtType = msg->message;
	switch (msg->message) {
		case WM_QUIT:
		case WM_DESTROY:
			shared_memory_set_write_state(sm, SHARED_MEM_QUIT);
			break;
		case WM_MOUSEMOVE:
			setMouseEvent(sm->evt, msg->lParam, -1);
			break;
		case WM_LBUTTONDOWN:
			SetCapture(msg->hwnd);
			setMouseEvent(sm->evt, msg->lParam, MOUSE_BUTTON_LEFT);
			break;
		case WM_LBUTTONUP:
			ReleaseCapture();
			setMouseEvent(sm->evt, msg->lParam, MOUSE_BUTTON_LEFT);
			break;
		case WM_MBUTTONDOWN:
			SetCapture(msg->hwnd);
			setMouseEvent(sm->evt, msg->lParam, MOUSE_BUTTON_MIDDLE);
			break;
		case WM_MBUTTONUP:
			ReleaseCapture();
			setMouseEvent(sm->evt, msg->lParam, MOUSE_BUTTON_MIDDLE);
			break;
		case WM_RBUTTONDOWN:
			SetCapture(msg->hwnd);
			setMouseEvent(sm->evt, msg->lParam, MOUSE_BUTTON_RIGHT);
			break;
		case WM_RBUTTONUP:
			ReleaseCapture();
			setMouseEvent(sm->evt, msg->lParam, MOUSE_BUTTON_RIGHT);
			break;
		case WM_XBUTTONDOWN:
		case WM_XBUTTONUP:
			if (msg->wParam & 0x0010000) {
				setMouseEvent(sm->evt, msg->lParam, MOUSE_BUTTON_X1);
			} else if (msg->wParam & 0x0020000) {
				setMouseEvent(sm->evt, msg->lParam, MOUSE_BUTTON_X2);
			}
			break;
		case WM_MOUSEWHEEL:
			setMouseEvent(sm->evt, msg->lParam, MOUSE_WHEEL_VERTICAL);
			sm->evt->mouse.wheelDelta = GET_WHEEL_DELTA_WPARAM(msg->wParam);
			break;
		case WM_MOUSEHWHEEL:
			setMouseEvent(sm->evt, msg->lParam, MOUSE_WHEEL_HORIZONTAL);
			sm->evt->mouse.wheelDelta = GET_WHEEL_DELTA_WPARAM(msg->wParam);
			break;
		case WM_KEYDOWN:
		case WM_SYSKEYDOWN:
		case WM_KEYUP:
		case WM_SYSKEYUP:
			switch (msg->wParam) {
				case VK_SHIFT:
					UINT scancode = (msg->lParam & 0x00FF0000) >> 16;
					sm->evt->keyboard.keyId = MapVirtualKey(scancode, MAPVK_VSC_TO_VK_EX);
					break;
				case VK_CONTROL:
					if (msg->lParam & 0x01000000) {
						sm->evt->keyboard.keyId = VK_RCONTROL;
					} else {
						sm->evt->keyboard.keyId = VK_LCONTROL;
					}
					break;
				case VK_MENU:
					if (msg->lParam & 0x01000000) {
						sm->evt->keyboard.keyId = VK_RMENU;
					} else {
						sm->evt->keyboard.keyId = VK_LMENU;
					}
					break;
				default:
					sm->evt->keyboard.keyId = msg->wParam;
					break;
			}
			break;
		case UWM_SET_CURSOR:
		{
			HCURSOR c = NULL;
			switch (msg->wParam) {
				case CURSOR_ARROW:
					c = LoadCursor(NULL, IDC_ARROW);
					break;
				case CURSOR_IBEAM:
					c = LoadCursor(NULL, IDC_IBEAM);
					break;
				case CURSOR_WAIT:
					c = LoadCursor(NULL, IDC_WAIT);
					break;
				case CURSOR_CROSS:
					c = LoadCursor(NULL, IDC_CROSS);
					break;
				case CURSOR_UPARROW:
					c = LoadCursor(NULL, IDC_UPARROW);
					break;
				case CURSOR_SIZE_NWSE:
					c = LoadCursor(NULL, IDC_SIZENWSE);
					break;
				case CURSOR_SIZE_NESW:
					c = LoadCursor(NULL, IDC_SIZENESW);
					break;
				case CURSOR_SIZE_WE:
					c = LoadCursor(NULL, IDC_SIZEWE);
					break;
				case CURSOR_SIZE_NS:
					c = LoadCursor(NULL, IDC_SIZENS);
					break;
				case CURSOR_SIZE_ALL:
					c = LoadCursor(NULL, IDC_SIZEALL);
					break;
				case CURSOR_NO:
					c = LoadCursor(NULL, IDC_NO);
					break;
				case CURSOR_HAND:
					c = LoadCursor(NULL, IDC_HAND);
					break;
				case CURSOR_APP_STARTING:
					c = LoadCursor(NULL, IDC_APPSTARTING);
					break;
				case CURSOR_HELP:
					c = LoadCursor(NULL, IDC_HELP);
					break;
				//case CURSOR_PIN:
				//	c = LoadCursor(NULL, IDC_PIN);
				//	break;
				//case CURSOR_PERSON:
				//	c = LoadCursor(NULL, IDC_PERSON);
				//	break;
			}
			if (c != NULL) {
				SetCursor(c);
				SetClassLongPtr(msg->hwnd, GCLP_HCURSOR, (LONG_PTR)c);
			}
			break;
		}
	}
}

void window_main(const wchar_t* windowTitle, int width, int height,
	int x, int y, void* evtSharedMem, int size)
{
	char* esm = evtSharedMem;
	// Register the window class.
	HMODULE hInstance = GetModuleHandle(NULL);
    const wchar_t className[]  = L"Kaiju Window Class";
    WNDCLASS wc = { 0 };
    wc.lpfnWndProc   = window_proc;
    wc.hInstance     = hInstance;
    wc.lpszClassName = className;
	wc.hCursor		 = LoadCursor(NULL, IDC_ARROW);
	wc.hIcon		 = LoadIcon(NULL, IDI_APPLICATION);
    RegisterClass(&wc);
	RECT clientArea = {0, 0, width, height};
	AdjustWindowRectEx(&clientArea, WS_OVERLAPPEDWINDOW, FALSE, 0);
	width = clientArea.right-clientArea.left;
	height = clientArea.bottom-clientArea.top;
	if (x < 0) {
		x = CW_USEDEFAULT;
	}
	if (y < 0) {
		x = CW_USEDEFAULT;
	}
    // Create the window.
    HWND hwnd = CreateWindowEx(
        0,									// Optional window styles.
        className,							// Window class
        windowTitle,						// Window text
        WS_OVERLAPPEDWINDOW,				// Window style
		x, y, width, height,				// Position & size
        NULL,								// Parent window
        NULL,								// Menu
        hInstance,							// Instance handle
        NULL								// Additional application data
	);
    if (hwnd == NULL) {
		write_fatal(esm, size, "Failed to create window.");
		return;
    }
	window_cursor_standard(hwnd);
	memcpy(esm+SHARED_MEM_DATA_START, &hwnd, sizeof(HWND*));
	memcpy(esm+SHARED_MEM_DATA_START+sizeof(&hwnd), &hInstance, sizeof(HMODULE*));
	SharedMem* sm = malloc(sizeof(SharedMem));
	sm->sharedMem = evtSharedMem;
	sm->size = size;
	sm->windowWidth = width;
	sm->windowHeight = height;
	SetWindowLongPtrA(hwnd, GWLP_USERDATA, (LONG_PTR)sm);
}

void window_show(void* hwnd) {
	ShowWindow(hwnd, SW_SHOW);
}

uint32_t window_poll_controller(void* hwnd) {
	SharedMem* sm = (SharedMem*)GetWindowLongPtrA(hwnd, GWLP_USERDATA);
	if (obtainControllerStates(sm)) {
		return EVENT_TYPE_CONTROLLER;
	}
	return 0;
}

uint32_t window_poll(void* hwnd) {
	SharedMem* sm = (SharedMem*)GetWindowLongPtrA(hwnd, GWLP_USERDATA);
 	// Run the message loop.
    MSG msg = { 0 };
	if (PeekMessage(&msg, hwnd, 0, 0, PM_REMOVE) > 0) {
		TranslateMessage(&msg);
		// TODO:  Window resize happens in here, but would be clobbered by &msg which is different
		DispatchMessage(&msg);
		process_message(sm, &msg);
		return msg.message;
	} else {
		return 0;
	}
}

void window_destroy(void* hwnd) {
	SharedMem* sm = (SharedMem*)GetWindowLongPtrA(hwnd, GWLP_USERDATA);
	DestroyWindow(hwnd);
	free(sm);
}

void window_cursor_standard(void* hwnd) {
	PostMessageA(hwnd, UWM_SET_CURSOR, CURSOR_ARROW, 0);
}

void window_cursor_ibeam(void* hwnd) {
	PostMessageA(hwnd, UWM_SET_CURSOR, CURSOR_IBEAM, 0);
}

void window_cursor_size_all(void* hwnd) {
	PostMessageA(hwnd, UWM_SET_CURSOR, CURSOR_SIZE_ALL, 0);
}

void window_cursor_size_ns(void* hwnd) {
	PostMessageA(hwnd, UWM_SET_CURSOR, CURSOR_SIZE_NS, 0);
}

void window_cursor_size_we(void* hwnd) {
	PostMessageA(hwnd, UWM_SET_CURSOR, CURSOR_SIZE_WE, 0);
}

float window_dpi(void* hwnd) {
	return ((float)GetDpiForWindow(hwnd)) / 25.4F;
}

void window_focus(void* hwnd) {
	BringWindowToTop(hwnd);
	SetFocus(hwnd);
}

void window_position(void* hwnd, int* x, int* y) {
	WINDOWPLACEMENT wp;
	wp.length = sizeof(WINDOWPLACEMENT);
	if (GetWindowPlacement(hwnd, &wp)) {
		*x = wp.rcNormalPosition.left;
		*y = wp.rcNormalPosition.top;
	} else {
		*x = -1;
		*y = -1;
	}
}

void window_set_position(void* hwnd, int x, int y) {
	SetWindowPos(hwnd, NULL, x, y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
}

void window_set_size(void* hwnd, int width, int height) {
	SetWindowPos(hwnd, NULL, 0, 0, width, height, SWP_NOMOVE | SWP_NOZORDER);
}

void window_remove_border(void* hwnd) {
	LONG style = GetWindowLong(hwnd, GWL_STYLE);
	style &= ~WS_CAPTION;
	style &= ~WS_THICKFRAME;
	style &= ~WS_MINIMIZEBOX;
	style &= ~WS_MAXIMIZEBOX;
	style &= ~WS_SYSMENU;
	SetWindowLong(hwnd, GWL_STYLE, style);
}

void window_add_border(void* hwnd) {
	LONG style = GetWindowLong(hwnd, GWL_STYLE);
	style |= WS_CAPTION;
	style |= WS_THICKFRAME;
	style |= WS_MINIMIZEBOX;
	style |= WS_MAXIMIZEBOX;
	style |= WS_SYSMENU;
	SetWindowLong(hwnd, GWL_STYLE, style);
}

#endif

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

void readMousePosition(LPARAM lParam, int32_t* x, int32_t* y) {
	*x = GET_X_LPARAM(lParam);
	*y = GET_Y_LPARAM(lParam);
}

double now_seconds() {
    FILETIME ft;
    GetSystemTimeAsFileTime(&ft);
    ULARGE_INTEGER uli;
    uli.LowPart = ft.dwLowDateTime;
    uli.HighPart = ft.dwHighDateTime;
    return (uli.QuadPart / 10000000.0) - 11644473600.0;
}

bool obtainControllerStates(SharedMem* sm) {
	static double last = 0;
	static double connectedControllers[MAX_CONTROLLERS] = { 0 };
	bool readControllerStates = false;
	double now = now_seconds();
	double delta = now - last;
	DWORD dwResult;
	for (DWORD i = 0; i < MAX_CONTROLLERS; i++) {
		// Don't check disconnected controllers every frame, bad perf
		if (connectedControllers[i] > 0) {
			connectedControllers[i] -= delta;
			continue;
		}
		XINPUT_STATE state;
		ZeroMemory(&state, sizeof(XINPUT_STATE));
		// Simply get the state of the controller from XInput.
		dwResult = XInputGetState(i, &state);
		WindowEvent evt = { WINDOW_EVENT_TYPE_CONTROLLER_STATE };
		evt.controllerState.controllerId = i;
		if(dwResult == ERROR_SUCCESS) {
			evt.controllerState.buttons = state.Gamepad.wButtons;
			evt.controllerState.leftTrigger = state.Gamepad.bLeftTrigger;
			evt.controllerState.rightTrigger = state.Gamepad.bRightTrigger;
			evt.controllerState.thumbLX = state.Gamepad.sThumbLX;
			evt.controllerState.thumbLY = state.Gamepad.sThumbLY;
			evt.controllerState.thumbRX = state.Gamepad.sThumbRX;
			evt.controllerState.thumbRY = state.Gamepad.sThumbRY;
			evt.controllerState.connectionType = WINDOW_EVENT_CONTROLLER_CONNECTION_TYPE_CONNECTED;
			readControllerStates = true;
			connectedControllers[i] = 0; // Check this controller next frame
		} else {
			// TODO:  readControllerStates would be true here too, but
			// no need to spam the event if no controllers are available?
			// Probably means the state of the controllers need tracking in C...
			evt.controllerState.connectionType = WINDOW_EVENT_CONTROLLER_CONNECTION_TYPE_DISCONNECTED;
			connectedControllers[i] = 3.0; // Wait a few seconds
		}
		shared_mem_add_event(sm, evt);
	}
	last = now;
	return readControllerStates;
}

LRESULT CALLBACK window_proc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
	SharedMem* sm = (SharedMem*)GetWindowLongPtrA(hwnd, GWLP_USERDATA);
	switch (uMsg) {
		case WM_DESTROY:
			if (sm != NULL) {
				shared_mem_add_event(sm, (WindowEvent) {
					.type = WINDOW_EVENT_TYPE_ACTIVITY,
					.windowActivity = { WINDOW_EVENT_ACTIVITY_TYPE_CLOSE }
				});
				shared_mem_flush_events(sm);
			}
			PostQuitMessage(0);
			return 0;
		case WM_ACTIVATE:
			switch (LOWORD(wParam)) {
				case WA_ACTIVE:
				case WA_CLICKACTIVE:
					shared_mem_add_event(sm, (WindowEvent) {
						.type = WINDOW_EVENT_TYPE_ACTIVITY,
						.windowActivity = { WINDOW_EVENT_ACTIVITY_TYPE_FOCUS }
					});
					break;
				case WA_INACTIVE:
					shared_mem_add_event(sm, (WindowEvent) {
						.type = WINDOW_EVENT_TYPE_ACTIVITY,
						.windowActivity = { WINDOW_EVENT_ACTIVITY_TYPE_BLUR }
					});
					break;
			}
			shared_mem_flush_events(sm);
			break;
		case WM_MOVE:
			// TODO:  Should handle this better, but move is called on focus too
			shared_mem_add_event(sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_MOVE,
				.windowMove = {
					.x = (int32_t)(short)LOWORD(lParam),
					.y = (int32_t)(short)HIWORD(lParam),
				}
			});
			shared_mem_flush_events(sm);
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
					WindowEvent evt = (WindowEvent) {
						.type = WINDOW_EVENT_TYPE_RESIZE,
						.windowResize = {
							.width = width,
							.height = height,
						}
					};
					RECT windowRect;
					if (GetWindowRect(hwnd, &windowRect)) {
						sm->left = windowRect.left;
						sm->top = windowRect.top;
						sm->right = windowRect.right;
						sm->bottom = windowRect.bottom;
						evt.windowResize.left = sm->left;
						evt.windowResize.top = sm->top;
						evt.windowResize.right = sm->right;
						evt.windowResize.bottom = sm->bottom;
					}
					shared_mem_add_event(sm, evt);
					shared_mem_flush_events(sm);
				}
			}
			PostMessage(hwnd, WM_PAINT, 0, 0);
			break;
	}
	return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

void process_message(SharedMem* sm, MSG *msg) {
	switch (msg->message) {
		case WM_QUIT:
		case WM_DESTROY:
		{
			shared_mem_add_event(sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_ACTIVITY,
				.windowActivity = { WINDOW_EVENT_ACTIVITY_TYPE_CLOSE }
			});
			shared_mem_flush_events(sm);
			break;
		}
		case WM_MOUSEMOVE:
		{
			WindowEvent evt = { WINDOW_EVENT_TYPE_MOUSE_MOVE };
			readMousePosition(msg->lParam, &evt.mouseMove.x, &evt.mouseMove.y);
			shared_mem_add_event(sm, evt);
			break;
		}
		case WM_LBUTTONDOWN:
		{
			WindowEvent evt = {
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = {
					.buttonId = MOUSE_BUTTON_LEFT,
					.action = WINDOW_EVENT_BUTTON_TYPE_DOWN,
				}
			};
			readMousePosition(msg->lParam, &evt.mouseButton.x, &evt.mouseButton.y);
			shared_mem_add_event(sm, evt);
			SetCapture(msg->hwnd);
			break;
		}
		case WM_LBUTTONUP:
		{
			WindowEvent evt = {
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = {
					.buttonId = MOUSE_BUTTON_LEFT,
					.action = WINDOW_EVENT_BUTTON_TYPE_UP,
				}
			};
			readMousePosition(msg->lParam, &evt.mouseButton.x, &evt.mouseButton.y);
			shared_mem_add_event(sm, evt);
			ReleaseCapture();
			break;
		}
		case WM_MBUTTONDOWN:
		{
			WindowEvent evt = {
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = {
					.buttonId = MOUSE_BUTTON_MIDDLE,
					.action = WINDOW_EVENT_BUTTON_TYPE_DOWN,
				}
			};
			readMousePosition(msg->lParam, &evt.mouseButton.x, &evt.mouseButton.y);
			shared_mem_add_event(sm, evt);
			SetCapture(msg->hwnd);
			break;
		}
		case WM_MBUTTONUP:
		{
			WindowEvent evt = {
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = {
					.buttonId = MOUSE_BUTTON_MIDDLE,
					.action = WINDOW_EVENT_BUTTON_TYPE_UP,
				}
			};
			readMousePosition(msg->lParam, &evt.mouseButton.x, &evt.mouseButton.y);
			shared_mem_add_event(sm, evt);
			ReleaseCapture();
			break;
		}
		case WM_RBUTTONDOWN:
		{
			WindowEvent evt = {
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = {
					.buttonId = MOUSE_BUTTON_RIGHT,
					.action = WINDOW_EVENT_BUTTON_TYPE_DOWN,
				}
			};
			readMousePosition(msg->lParam, &evt.mouseButton.x, &evt.mouseButton.y);
			shared_mem_add_event(sm, evt);
			SetCapture(msg->hwnd);
			break;
		}
		case WM_RBUTTONUP:
		{
			WindowEvent evt = {
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = {
					.buttonId = MOUSE_BUTTON_RIGHT,
					.action = WINDOW_EVENT_BUTTON_TYPE_UP,
				}
			};
			readMousePosition(msg->lParam, &evt.mouseButton.x, &evt.mouseButton.y);
			shared_mem_add_event(sm, evt);
			ReleaseCapture();
			break;
		}
		case WM_XBUTTONDOWN:
		case WM_XBUTTONUP:
		{
			WindowEvent evt = { WINDOW_EVENT_TYPE_MOUSE_BUTTON };
			if (msg->wParam & 0x0010000) {
				evt.mouseButton.buttonId = MOUSE_BUTTON_X1;
			} else if (msg->wParam & 0x0020000) {
				evt.mouseButton.buttonId = MOUSE_BUTTON_X2;
			}
			readMousePosition(msg->lParam, &evt.mouseButton.x, &evt.mouseButton.y);
			if (msg->message == WM_XBUTTONDOWN) {
				evt.mouseButton.action = WINDOW_EVENT_BUTTON_TYPE_DOWN;
				SetCapture(msg->hwnd);
			} else {
				evt.mouseButton.action = WINDOW_EVENT_BUTTON_TYPE_UP;
				ReleaseCapture();
			}
			shared_mem_add_event(sm, evt);
			break;
		}
		case WM_MOUSEWHEEL:
		{
			WindowEvent evt = {
				.type = WINDOW_EVENT_TYPE_MOUSE_SCROLL,
				.mouseScroll = {
					.deltaY = GET_WHEEL_DELTA_WPARAM(msg->wParam),
				}
			};
			readMousePosition(msg->lParam, &evt.mouseScroll.x, &evt.mouseScroll.y);
			shared_mem_add_event(sm, evt);
			break;
		}
		case WM_MOUSEHWHEEL:
		{
			WindowEvent evt = {
				.type = WINDOW_EVENT_TYPE_MOUSE_SCROLL,
				.mouseScroll = {
					.deltaX = GET_WHEEL_DELTA_WPARAM(msg->wParam),
				}
			};
			readMousePosition(msg->lParam, &evt.mouseScroll.x, &evt.mouseScroll.y);
			shared_mem_add_event(sm, evt);
			break;
		}
		case WM_KEYDOWN:
		case WM_SYSKEYDOWN:
		case WM_KEYUP:
		case WM_SYSKEYUP:
		{
			WindowEvent evt = { WINDOW_EVENT_TYPE_KEYBOARD_BUTTON };
			if (msg->message == WM_KEYDOWN || msg->message == WM_SYSKEYDOWN) {
				evt.keyboardButton.action = WINDOW_EVENT_BUTTON_TYPE_DOWN;
			} else {
				evt.keyboardButton.action = WINDOW_EVENT_BUTTON_TYPE_UP;
			}
			switch (msg->wParam) {
				case VK_SHIFT:
					UINT scancode = (msg->lParam & 0x00FF0000) >> 16;
					evt.keyboardButton.buttonId = MapVirtualKey(scancode, MAPVK_VSC_TO_VK_EX);
					break;
				case VK_CONTROL:
					if (msg->lParam & 0x01000000) {
						evt.keyboardButton.buttonId = VK_RCONTROL;
					} else {
						evt.keyboardButton.buttonId = VK_LCONTROL;
					}
					break;
				case VK_MENU:
					if (msg->lParam & 0x01000000) {
						evt.keyboardButton.buttonId = VK_RMENU;
					} else {
						evt.keyboardButton.buttonId = VK_LMENU;
					}
					break;
				default:
					evt.keyboardButton.buttonId = msg->wParam;
					break;
			}
			shared_mem_add_event(sm, evt);
			break;
		}
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

void window_main(const wchar_t* windowTitle,
	int width, int height, int x, int y, uint64_t goWindow)
{
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
	SharedMem* sm = calloc(1, sizeof(SharedMem));
	sm->goWindow = (void*)goWindow;
    if (hwnd == NULL) {
		shared_mem_add_event(sm, (WindowEvent) {
			.type = WINDOW_EVENT_TYPE_FATAL,
			.setHandle = {
				.hwnd = hwnd,
				.instance = hInstance,
			}
		});
		shared_mem_flush_events(sm);
		return;
    }
	window_cursor_standard(hwnd);
	sm->windowWidth = width;
	sm->windowHeight = height;
	shared_mem_add_event(sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_SET_HANDLE,
		.setHandle = {
			.hwnd = hwnd,
			.instance = hInstance,
		}
	});
	shared_mem_flush_events(sm);
	SetWindowLongPtrA(hwnd, GWLP_USERDATA, (LONG_PTR)sm);
}

void window_show(void* hwnd) {
	SharedMem* sm = (SharedMem*)GetWindowLongPtrA(hwnd, GWLP_USERDATA);
	ShowWindow(hwnd, SW_SHOW);
	RECT windowRect;
	if (GetWindowRect(hwnd, &windowRect)) {
		if (sm->left != windowRect.left || sm->top != windowRect.top
			|| sm->right != windowRect.right || sm->bottom != windowRect.bottom)
		{
			WindowEvent evt = (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_RESIZE,
				.windowResize = {
					.width = sm->windowWidth,
					.height = sm->windowHeight,
					.left = windowRect.left,
					.top = windowRect.top,
					.right = windowRect.right,
					.bottom = windowRect.bottom,
				}
			};
			shared_mem_add_event(sm, evt);
			shared_mem_flush_events(sm);
		}
	}
}

void window_poll_controller(void* hwnd) {
	SharedMem* sm = (SharedMem*)GetWindowLongPtrA(hwnd, GWLP_USERDATA);
	obtainControllerStates(sm);
}

void window_poll(void* hwnd) {
	SharedMem* sm = (SharedMem*)GetWindowLongPtrA(hwnd, GWLP_USERDATA);
    MSG msg = { 0 };
	while (PeekMessage(&msg, hwnd, 0, 0, PM_REMOVE) > 0) {
		TranslateMessage(&msg);
		DispatchMessage(&msg);
		process_message(sm, &msg);
	}
	shared_mem_flush_events(sm);
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
	return ((float)GetDpiForWindow(hwnd));
}

int screen_width_mm(void* hwnd) {
    HDC hdc = GetDC(hwnd);
    if (hdc == NULL) {
		return -1;
    }
    int widthMM = GetDeviceCaps(hdc, HORZSIZE);
    ReleaseDC(NULL, hdc);
	return widthMM;
}

int screen_height_mm(void* hwnd) {
    HDC hdc = GetDC(hwnd);
    if (hdc == NULL) {
		return -1;
    }
    int heightMM = GetDeviceCaps(hdc, VERTSIZE);
    ReleaseDC(NULL, hdc);
	return heightMM;
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

void window_show_cursor(void* hwnd) {
	ShowCursor(TRUE);
}

void window_hide_cursor(void* hwnd) {
	ShowCursor(FALSE);
}

#endif

#if defined(_WIN32) || defined(_WIN64)

#ifndef UNICODE
#define UNICODE
#endif

#include "win32.h"
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <windows.h>
#include <windowsx.h>
#include "shared_mem.h"

#include <XInput.h>

#ifdef OPENGL
#include "../gl/dist/glad_wgl.h"
#endif

int shared_mem_set_thread_priority(SharedMem* sm) {
	int priority = GetThreadPriority(GetCurrentThread());
	if (sm->evt->writeState != SHARED_MEM_WRITTEN) {
		SetThreadPriority(GetCurrentThread(), THREAD_PRIORITY_IDLE);
	}
	return priority;
}

void shared_mem_reset_thread_priority(SharedMem* sm, int priority) {
	SetThreadPriority(GetCurrentThread(), priority);
}

void shared_mem_wait(SharedMem* sm) {
	SwitchToThread();
}

void setMouseEvent(InputEvent* evt, LPARAM lParam, int buttonId) {
	evt->mouse.mouseButtonId = buttonId;
	evt->mouse.mouseX = GET_X_LPARAM(lParam);
	evt->mouse.mouseY = GET_Y_LPARAM(lParam);
}

void setSizeEvent(InputEvent* evt, LPARAM lParam) {
	evt->resize.width = LOWORD(lParam);
	evt->resize.height = HIWORD(lParam);
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
	if (sm != NULL) {
		sm->evt->evtType = uMsg;
		shared_memory_set_write_state(sm, SHARED_MEM_WRITING);
		switch (uMsg) {
			case WM_DESTROY:
				PostQuitMessage(0);
				shared_memory_set_write_state(sm, SHARED_MEM_QUIT);
				return 0;
						case WM_SIZE:
				setSizeEvent(sm->evt, lParam);
				PostMessage(hwnd, WM_PAINT, 0, 0);
				break;
			case WM_PAINT:
			{
				//PAINTSTRUCT ps;
				//BeginPaint(hwnd, &ps);
				//EndPaint(hwnd, &ps);
				break;
			}
			case WM_MOUSEMOVE:
				setMouseEvent(sm->evt, lParam, -1);
				break;
			case WM_LBUTTONDOWN:
			case WM_LBUTTONUP:
				setMouseEvent(sm->evt, lParam, MOUSE_BUTTON_LEFT);
				break;
			case WM_MBUTTONDOWN:
			case WM_MBUTTONUP:
				setMouseEvent(sm->evt, lParam, MOUSE_BUTTON_MIDDLE);
				break;
			case WM_RBUTTONDOWN:
			case WM_RBUTTONUP:
				setMouseEvent(sm->evt, lParam, MOUSE_BUTTON_RIGHT);
				break;
			case WM_XBUTTONDOWN:
			case WM_XBUTTONUP:
				if (wParam & 0x0010000) {
					setMouseEvent(sm->evt, lParam, MOUSE_BUTTON_X1);
				} else if (wParam & 0x0020000) {
					setMouseEvent(sm->evt, lParam, MOUSE_BUTTON_X2);
				}
				break;
			case WM_MOUSEWHEEL:
				// TODO:  Add wheel code
				break;
			case WM_KEYDOWN:
			case WM_SYSKEYDOWN:
			case WM_KEYUP:
			case WM_SYSKEYUP:
				switch (wParam) {
					case VK_SHIFT:
						UINT scancode = (lParam & 0x00FF0000) >> 16;
						sm->evt->keyboard.keyId = MapVirtualKey(scancode, MAPVK_VSC_TO_VK_EX);
						break;
					case VK_CONTROL:
						if (lParam & 0x01000000) {
							sm->evt->keyboard.keyId = VK_RCONTROL;
						} else {
							sm->evt->keyboard.keyId = VK_LCONTROL;
						}
						break;
					case VK_MENU:
						if (lParam & 0x01000000) {
							sm->evt->keyboard.keyId = VK_RMENU;
						} else {
							sm->evt->keyboard.keyId = VK_LMENU;
						}
						break;
					default:
						sm->evt->keyboard.keyId = wParam;
						break;
				}
				break;
		}
		shared_memory_set_write_state(sm, SHARED_MEM_WRITTEN);
	}
    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

#ifdef OPENGL
const char* setup_gl_context(HWND hwnd) {
	PIXELFORMATDESCRIPTOR pfd = { 0 };
	pfd.nSize = sizeof(pfd);
	pfd.nVersion = 1;
	pfd.dwFlags = PFD_DRAW_TO_WINDOW|PFD_SUPPORT_OPENGL|PFD_DOUBLEBUFFER;
	pfd.iPixelType = PFD_TYPE_RGBA;
	pfd.cColorBits = 32;
	pfd.cDepthBits = 24;
	pfd.cStencilBits = 8;
	pfd.iLayerType = PFD_MAIN_PLANE;
	HDC hdc = GetDC(hwnd);
	int pixelFormat = ChoosePixelFormat(hdc, &pfd);
	if (pixelFormat == 0) {
		return "Failed to find a suitable pixel format.";
	}
	if (!SetPixelFormat(hdc, pixelFormat, &pfd)) {
		return "Failed to set the pixel format.";
	}
	HGLRC legacyCtx = wglCreateContext(hdc);
	if (legacyCtx == NULL) {
		return "Failed to create an OpenGL context.";
	}
	if (!wglMakeCurrent(hdc, legacyCtx)) {
		return "Failed to make the OpenGL context current.";
	}
	const int ctxAttr[] = {
		WGL_CONTEXT_MAJOR_VERSION_ARB, 3,
		WGL_CONTEXT_MINOR_VERSION_ARB, 3,
		WGL_CONTEXT_FLAGS_ARB, WGL_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB,
		WGL_CONTEXT_PROFILE_MASK_ARB, WGL_CONTEXT_CORE_PROFILE_BIT_ARB,
		0
	};
	PFNWGLCREATECONTEXTATTRIBSARBPROC wglCreateContextAttribsARB =
		(PFNWGLCREATECONTEXTATTRIBSARBPROC)wglGetProcAddress("wglCreateContextAttribsARB");
	if (wglCreateContextAttribsARB == NULL) {
		return "Failed to get wglCreateContextAttribsARB.";
	}
	HGLRC renderingCtx = wglCreateContextAttribsARB(hdc, 0, ctxAttr);
	if (renderingCtx == NULL) {
		return "Failed to create the rendering context.";
	}
	BOOL res = wglMakeCurrent(hdc, renderingCtx);
	if (!res) {
		return "Failed to make the rendering context current.";
	}
	res = wglDeleteContext(legacyCtx);
	if (!res) {
		return "Failed to delete the legacy context.";
	}
	if (!wglMakeCurrent(hdc, renderingCtx)) {
		return "Failed to make the rendering context current.";
	}
	// VSync
	const int frameVSyncSkipCount = 1;
	PFNWGLSWAPINTERVALEXTPROC wglSwapIntervalEXT =
		(PFNWGLSWAPINTERVALEXTPROC)wglGetProcAddress("wglSwapIntervalEXT");
	if (wglSwapIntervalEXT != NULL) {
		wglSwapIntervalEXT(frameVSyncSkipCount);
	}
	if (gladLoadGL() == 0) {
		return "Failed to load OpenGL.";
	}
	return NULL;
}

void window_create_gl_context(void* winHWND, void* evtSharedMem, int size) {
	HWND hwnd = winHWND;
	char* esm = evtSharedMem;
	const char* err = setup_gl_context(hwnd);
	if (err != NULL) {
		write_fatal(esm, size, err);
		return;
	}
}
#endif

void window_main(const wchar_t* windowTitle, int width, int height, void* evtSharedMem, int size) {
	char* esm = evtSharedMem;
	// Register the window class.
	HMODULE hInstance = GetModuleHandle(NULL);
    const wchar_t className[]  = L"Kaiju Window Class";
    WNDCLASS wc = { };
    wc.lpfnWndProc   = window_proc;
    wc.hInstance     = hInstance;
    wc.lpszClassName = className;
	wc.hIcon		 = LoadIcon(NULL, IDI_APPLICATION);
	wc.hCursor		 = LoadCursor(NULL, IDC_ARROW);
    RegisterClass(&wc);
	RECT clientArea = {0, 0, width, height};
	AdjustWindowRectEx(&clientArea, WS_OVERLAPPEDWINDOW, FALSE, 0);
    // Create the window.
    HWND hwnd = CreateWindowEx(
        0,									// Optional window styles.
        className,							// Window class
        windowTitle,						// Window text
        WS_OVERLAPPEDWINDOW,				// Window style
        CW_USEDEFAULT, CW_USEDEFAULT,		// Position
		clientArea.right-clientArea.left,	// Width
		clientArea.bottom-clientArea.top,	// Height
        NULL,								// Parent window
        NULL,								// Menu
        hInstance,							// Instance handle
        NULL								// Additional application data
	);
    if (hwnd == NULL) {
		write_fatal(esm, size, "Failed to create window.");
		return;
    }
	SharedMem sm = {esm, size};
	memcpy(esm+SHARED_MEM_DATA_START, &hwnd, sizeof(HWND*));
	memcpy(esm+SHARED_MEM_DATA_START+sizeof(&hwnd), &hInstance, sizeof(HMODULE*));
	shared_memory_set_write_state(&sm, SHARED_MEM_AWAITING_CONTEXT);
	// Context should be created in Go here on go main thread
	shared_memory_wait_for_available(&sm);
	ShowWindow(hwnd, SW_SHOW);
	SetWindowLongPtrA(hwnd, GWLP_USERDATA, (LONG_PTR)&sm);
	shared_memory_set_write_state(&sm, SHARED_MEM_AWAITING_START);
    // Run the message loop.
    MSG msg = { };
	while (esm[0] != SHARED_MEM_QUIT) {
		shared_memory_wait_for_available(&sm);
		if (obtainControllerStates(&sm)) {
			sm.evt->evtType = EVENT_TYPE_CONTROLLER;
			shared_memory_set_write_state(&sm, SHARED_MEM_WRITTEN);
			shared_memory_wait_for_available(&sm);
		}
		do {
			if (PeekMessage(&msg, NULL, 0, 0, PM_REMOVE) > 0) {
				TranslateMessage(&msg);
				DispatchMessage(&msg);
			} else {
				sm.evt->evtType = 0;
				shared_memory_set_write_state(&sm, SHARED_MEM_WRITTEN);
			}
		} while(sm.evt->evtType != 0);
	}
}

#endif
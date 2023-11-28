#ifndef WINDOWING_WIN32_H
#define WINDOWING_WIN32_H

#ifndef UNICODE
#define UNICODE
#endif

#include <stdint.h>
#include <string.h>
#include <windows.h>
#include <windowsx.h>

void write_fatal(char* evtSharedMem, int size, const char* msg) {
	strcpy_s(evtSharedMem + SHARED_MEM_DATA_START, size - SHARED_MEM_DATA_START, msg);
	evtSharedMem[0] = SHARED_MEM_FATAL;
}

void setMouseEvent(InputEvent* evt, LPARAM lParam, int buttonId) {
	evt->mouseButtonId = buttonId;
	evt->mouseX = GET_X_LPARAM(lParam);
	evt->mouseY = GET_Y_LPARAM(lParam);
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
			case WM_PAINT:
			{
				PAINTSTRUCT ps;
				HDC hdc = BeginPaint(hwnd, &ps);
				// All painting occurs here, between BeginPaint and EndPaint.
				FillRect(hdc, &ps.rcPaint, (HBRUSH) (COLOR_WINDOW+1));
				EndPaint(hwnd, &ps);
				break;
			}
			case WM_MOUSEMOVE:
				setMouseEvent(sm->evt, uMsg, -1);
				break;
			case WM_LBUTTONDOWN:
			case WM_LBUTTONUP:
				setMouseEvent(sm->evt, uMsg, MOUSE_BUTTON_LEFT);
				break;
			case WM_MBUTTONDOWN:
			case WM_MBUTTONUP:
				setMouseEvent(sm->evt, uMsg, MOUSE_BUTTON_MIDDLE);
				break;
			case WM_RBUTTONDOWN:
			case WM_RBUTTONUP:
				setMouseEvent(sm->evt, uMsg, MOUSE_BUTTON_RIGHT);
				break;
			case WM_XBUTTONDOWN:
			case WM_XBUTTONUP:
				if (wParam & 0x0010000) {
					setMouseEvent(sm->evt, uMsg, MOUSE_BUTTON_X1);
				} else if (wParam & 0x0020000) {
					setMouseEvent(sm->evt, uMsg, MOUSE_BUTTON_X2);
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
						sm->evt->keyId = MapVirtualKey(scancode, MAPVK_VSC_TO_VK_EX);
						break;
					case VK_CONTROL:
						if (lParam & 0x01000000) {
							sm->evt->keyId = VK_RCONTROL;
						} else {
							sm->evt->keyId = VK_LCONTROL;
						}
						break;
					case VK_MENU:
						if (lParam & 0x01000000) {
							sm->evt->keyId = VK_RMENU;
						} else {
							sm->evt->keyId = VK_LMENU;
						}
						break;
					default:
						sm->evt->keyId = wParam;
						break;
				}
				break;
		}
		shared_memory_set_write_state(sm, SHARED_MEM_WRITTEN);
	}
    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

void window_main(const wchar_t* windowTitle, void* evtSharedMem, int size) {
	char* esm = evtSharedMem;
	// Register the window class.
	HMODULE hInstance = GetModuleHandle(NULL);
    const wchar_t className[]  = L"Kaiju Window Class";
    WNDCLASS wc = { };
    wc.lpfnWndProc   = window_proc;
    wc.hInstance     = hInstance;
    wc.lpszClassName = className;
    RegisterClass(&wc);
    // Create the window.
    HWND hwnd = CreateWindowEx(
        0,						// Optional window styles.
        className,				// Window class
        windowTitle,			// Window text
        WS_OVERLAPPEDWINDOW,	// Window style
        // Size and position
        CW_USEDEFAULT, CW_USEDEFAULT, CW_USEDEFAULT, CW_USEDEFAULT,
        NULL,					// Parent window
        NULL,					// Menu
        hInstance,				// Instance handle
        NULL					// Additional application data
	);
    if (hwnd == NULL) {
		write_fatal(esm, size, "Failed to create window.");
		return;
    }
	SharedMem sm = {esm, size};
	SetWindowLongPtrA(hwnd, GWLP_USERDATA, (LONG_PTR)&sm);
    ShowWindow(hwnd, SW_SHOW);
	shared_memory_set_write_state(&sm, SHARED_MEM_WRITTEN);
    // Run the message loop.
    MSG msg = { };
	while (esm[0] != SHARED_MEM_QUIT) {
		if (PeekMessage(&msg, NULL, 0, 0, PM_REMOVE) > 0) {
			shared_memory_wait_for_available(&sm);
			TranslateMessage(&msg);
			DispatchMessage(&msg);
		} else {
			sm.evt->evtType = 0;
			shared_memory_set_write_state(&sm, SHARED_MEM_WRITTEN);
		}
	}
}

#endif
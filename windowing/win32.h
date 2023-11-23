#ifndef WINDOWING_WIN32_H
#define WINDOWING_WIN32_H

#ifndef UNICODE
#define UNICODE
#endif

#include <stdint.h>
#include <windows.h>
#include <windowsx.h>

LRESULT CALLBACK window_proc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
    switch (uMsg) {
		case WM_DESTROY:
			PostQuitMessage(0);
			return 0;
		case WM_PAINT:
		{
			PAINTSTRUCT ps;
			HDC hdc = BeginPaint(hwnd, &ps);
			// All painting occurs here, between BeginPaint and EndPaint.
			FillRect(hdc, &ps.rcPaint, (HBRUSH) (COLOR_WINDOW+1));
			EndPaint(hwnd, &ps);
		}
		return 0;
    }
    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

InputEvent setMouseEvent(MSG msg, int buttonId) {
	InputEvent ie;
	ie.mouseX = GET_X_LPARAM(msg.lParam);
	ie.mouseY = GET_Y_LPARAM(msg.lParam);
	ie.mouseButtonId = buttonId;
	return ie;
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
		esm[0] = SHARED_MEM_FATAL;
		return;
    }
    ShowWindow(hwnd, SW_SHOW);
	esm[0] = SHARED_MEM_WRITTEN;
    // Run the message loop.
    MSG msg = { };
	while (esm[0] != SHARED_MEM_QUIT) {
		void* esmData = esm + SHARED_MEM_DATA_START;
		uint32_t msgType = 0;
		if (PeekMessage(&msg, NULL, 0, 0, PM_REMOVE) > 0) {
			while (esm[0] != SHARED_MEM_AVAILABLE) {}
			esm[0] = SHARED_MEM_WRITING;
			if (msg.message == WM_QUIT) {
				esm[0] = SHARED_MEM_QUIT;
				break;
			} else {
				msgType = msg.message;
				memcpy(esmData, &msgType, sizeof(msgType));
				esmData += sizeof(msgType);
				InputEvent ie;
				switch (msg.message) {
					case WM_MOUSEMOVE:
						ie = setMouseEvent(msg, -1);
						break;
					case WM_LBUTTONDOWN:
					case WM_LBUTTONUP:
						ie = setMouseEvent(msg, MOUSE_BUTTON_LEFT);
						break;
					case WM_MBUTTONDOWN:
					case WM_MBUTTONUP:
						ie = setMouseEvent(msg, MOUSE_BUTTON_MIDDLE);
						break;
					case WM_RBUTTONDOWN:
					case WM_RBUTTONUP:
						ie = setMouseEvent(msg, MOUSE_BUTTON_RIGHT);
						break;
					case WM_XBUTTONDOWN:
					case WM_XBUTTONUP:
						if (msg.wParam & 0x0010000) {
							ie = setMouseEvent(msg, MOUSE_BUTTON_X1);
						} else if (msg.wParam & 0x0020000) {
							ie = setMouseEvent(msg, MOUSE_BUTTON_X2);
						}
						break;
				}
				memcpy(esmData, &ie, sizeof(ie));
				esm[0] = SHARED_MEM_WRITTEN;
			}
			TranslateMessage(&msg);
			DispatchMessage(&msg);
		} else {
			memcpy(esmData, &msgType, sizeof(msgType));
			esm[0] = SHARED_MEM_WRITTEN;
		}
	}
}

#endif
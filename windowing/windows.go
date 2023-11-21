//go:build windows

package windowing

import (
	"unicode/utf16"
	"unsafe"
)

/*
#ifndef UNICODE
#define UNICODE
#endif

#include <windows.h>
#include <windowsx.h>

#define SHARED_MEM_AVAILABLE	0
#define SHARED_MEM_WRITING		1
#define SHARED_MEM_WRITTEN		2
#define SHARED_MEM_FATAL		0xFE
#define SHARED_MEM_QUIT			0xFF
#define SHARED_MEM_DATA_START	4

typedef struct {
	union {
		int mouseX;
		int key;
	};
	int mouseY;
	int mouseXButton;
} InputEvent;

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
    // Run the message loop.
    MSG msg = { };
    while (GetMessage(&msg, NULL, 0, 0) > 0) {
		esm[0] = SHARED_MEM_WRITING;
		if (msg.message == WM_QUIT) {
			break;
		} else {
			void* esmData = esm + SHARED_MEM_DATA_START;
			memcpy(esmData, &msg.message, sizeof(msg.message));
			esmData += sizeof(msg.message);
			InputEvent ie;
			switch (msg.message) {
				case WM_LBUTTONDOWN:
				case WM_LBUTTONUP:
				case WM_MOUSEMOVE:
					ie.mouseX = GET_X_LPARAM(msg.lParam);
					ie.mouseY = GET_Y_LPARAM(msg.lParam);
					//ie.mouseXButton = msg.wParam & 0x0020; (x button 1)
					//ie.mouseXButton = msg.wParam & 0x0040; (x button 2)
					break;
			}
			memcpy(esmData, &ie, sizeof(ie));
			esm[0] = SHARED_MEM_WRITTEN;
		}
		TranslateMessage(&msg);
		DispatchMessage(&msg);
	}
	esm[0] = SHARED_MEM_QUIT;
}
*/
import "C"

const (
	sharedMemAvailable = iota
	sharedMemWriting
	sharedMemWritten
	sharedMemFatal = 0xFE
	sharedMemQuit  = 0xFF
)

const evtSharedMemSize = 256

type evtMem [evtSharedMemSize]byte

func (e *evtMem) AsPointer() unsafe.Pointer { return unsafe.Pointer(&e[0]) }
func (e evtMem) IsFatal() bool              { return e[0] == sharedMemFatal }
func (e evtMem) IsReady() bool              { return e[0] == sharedMemAvailable }
func (e evtMem) IsQuit() bool               { return e[0] == sharedMemQuit }
func (e *evtMem) MakeAvailable()            { e[0] = sharedMemAvailable }
func (e evtMem) HasEvent() bool             { return e.EventType() != 0 }
func (e evtMem) EventType() uint32 {
	return *(*uint32)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func createWindow(windowName string) {
	var evtSharedMem evtMem
	windowTitle := utf16.Encode([]rune(windowName))
	go C.window_main((*C.wchar_t)(unsafe.Pointer(&windowTitle[0])), evtSharedMem.AsPointer(), evtSharedMemSize)
	for !evtSharedMem.IsQuit() && !evtSharedMem.IsFatal() {

	}
}

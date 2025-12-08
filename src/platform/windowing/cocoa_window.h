#ifndef COCOA_WINDOW_H
#define COCOA_WINDOW_H

#include <objc/objc.h>
#include "shared_mem.h"

// Creates a Cocoa window with CAMetalLayer-backed NSView
// Returns: NSView* (the view with Metal layer attached)
// Also sets outWindow to the NSWindow* if needed
void* cocoa_create_window(const char* title, int x, int y, int width, int height, void** outWindow, void* goWindow);

// Destroy the window
void cocoa_destroy_window(void* nsWindow);

// Show the window
void cocoa_show_window(void* nsWindow);

// Poll events
void cocoa_poll_events(void* nsWindow);

// Get screen DPI (dots per inch)
float cocoa_get_dpi(void* nsWindow);

// Get screen size in millimeters
void cocoa_screen_size_mm(void* nsWindow, int* width, int* height);

#endif // COCOA_WINDOW_H

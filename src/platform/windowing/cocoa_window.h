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

// Window position and size
void cocoa_get_position(void* nsWindow, int* x, int* y);
void cocoa_set_position(void* nsWindow, int x, int y);
void cocoa_set_size(void* nsWindow, int width, int height);

// Window title
void cocoa_set_title(void* nsWindow, const char* title);

// Clipboard
void cocoa_copy_to_clipboard(const char* text);
char* cocoa_clipboard_contents(void);

// Cursor variants
void cocoa_cursor_standard(void);
void cocoa_cursor_ibeam(void);
void cocoa_cursor_size_all(void);
void cocoa_cursor_size_ns(void);
void cocoa_cursor_size_we(void);
void cocoa_show_cursor(void);
void cocoa_hide_cursor(void);

// Window focus
void cocoa_focus_window(void* nsWindow);

// Get bundle resource path (for reading app assets)
void cocoa_get_bundle_resource_path(const char* resourceName, void** outPath);

#endif // COCOA_WINDOW_H

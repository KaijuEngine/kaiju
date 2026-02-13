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

// Get screen pixel width
int cocoa_get_screen_pixel_width(void* nsWindow);

// Get screen pixel height
int cocoa_get_screen_pixel_height(void* nsWindow);

// Get backing scale factor
double cocoa_get_backing_scale_factor(void* nsWindow);

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

// Window icon
void cocoa_set_icon(void* nsWindow, int width, int height, const uint8_t* pixelData);

// Window border manipulation
void cocoa_remove_border(void* nsWindow);
void cocoa_add_border(void* nsWindow);

// Fullscreen mode
void cocoa_set_fullscreen(void* nsWindow);
void cocoa_set_windowed(void* nsWindow, int width, int height);

// Cursor lock/confinement
void cocoa_lock_cursor(void* nsWindow, int x, int y);
void cocoa_unlock_cursor(void* nsWindow);

// Raw mouse input (delta mode)
void cocoa_enable_raw_mouse(void* nsWindow);
void cocoa_disable_raw_mouse(void* nsWindow);

// Keyboard inputs
bool cocoa_get_caps_lock_toggle_key_state(void);

void cocoa_run_app(void);

#endif // COCOA_WINDOW_H

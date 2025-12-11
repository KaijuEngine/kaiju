#import <Cocoa/Cocoa.h>
#import <QuartzCore/CAMetalLayer.h>
#import <objc/runtime.h>
#include <string.h>
#include "cocoa_window.h"
#include "window_event.h"

// Custom NSView subclass with CAMetalLayer backing
@interface MetalView : NSView
@end

@implementation MetalView

- (instancetype)initWithFrame:(NSRect)frame {
    self = [super initWithFrame:frame];
    if (self) {
        self.wantsLayer = YES;
        self.layer = [[CAMetalLayer alloc] init];
        ((CAMetalLayer*)self.layer).pixelFormat = MTLPixelFormatBGRA8Unorm;
    }
    return self;
}

- (CALayer*)makeBackingLayer {
    return [[CAMetalLayer alloc] init];
}

- (BOOL)wantsUpdateLayer {
    return YES;
}

@end

// Application delegate to handle dock icon clicks when window is minimized
@interface AppDelegate : NSObject <NSApplicationDelegate>
@end

@implementation AppDelegate

- (BOOL)applicationShouldHandleReopen:(NSApplication *)sender hasVisibleWindows:(BOOL)flag {
    
    // When dock icon is clicked, restore and activate windows
    BOOL didRestoreWindow = NO;
    
    for (NSWindow* window in [NSApp windows]) {
        if ([window isMiniaturized]) {
            [window deminiaturize:nil];
            didRestoreWindow = YES;
        }
        // Bring window to front regardless
        [window makeKeyAndOrderFront:nil];
    }
    
    // Always activate the application
    [NSApp activateIgnoringOtherApps:YES];
    return YES;
}

@end

// Helper function to retrieve SharedMem from window
static inline SharedMem* getSharedMem(NSWindow* window) {
    NSValue* value = objc_getAssociatedObject(window, "sharedMem");
    return (value != nil) ? [value pointerValue] : NULL;
}

// Window delegate to handle resize events.
// Uses NSWindowDelegate to receive resize notifications only when size actually changes,
// preventing continuous swapchain recreation that would occur with polling-based detection.
@interface WindowDelegate : NSObject <NSWindowDelegate>
@end

@implementation WindowDelegate

- (void)windowDidResize:(NSNotification *)notification {
    NSWindow* window = [notification object];
    SharedMem* sm = getSharedMem(window);
    if (sm == NULL) return;
    
    NSRect contentRect = [window contentRectForFrameRect:[window frame]];
    int newWidth = (int)contentRect.size.width;
    int newHeight = (int)contentRect.size.height;
    
    if (sm->windowWidth != newWidth || sm->windowHeight != newHeight) {
        sm->windowWidth = newWidth;
        sm->windowHeight = newHeight;
        
        NSRect frameRect = [window frame];
        sm->left = (int)frameRect.origin.x;
        sm->bottom = (int)frameRect.origin.y;
        sm->right = sm->left + (int)frameRect.size.width;
        sm->top = sm->bottom + (int)frameRect.size.height;
        
        shared_mem_add_event(sm, (WindowEvent) {
            .type = WINDOW_EVENT_TYPE_RESIZE,
            .windowResize = {
                .width = newWidth,
                .height = newHeight,
                .left = sm->left,
                .top = sm->top,
                .right = sm->right,
                .bottom = sm->bottom,
            }
        });
        shared_mem_flush_events(sm);
    }
}

- (void)windowDidResignKey:(NSNotification *)notification {
    NSWindow* window = [notification object];
    // Unlock cursor when window loses focus to prevent it from getting stuck
    cocoa_unlock_cursor((__bridge void*)window);
}

- (void)windowWillClose:(NSNotification *)notification {
    NSWindow* window = [notification object];
    SharedMem* sm = getSharedMem(window);
    
    // Send close event to notify Go that window is closing
    if (sm != NULL) {
        shared_mem_add_event(sm, (WindowEvent) {
            .type = WINDOW_EVENT_TYPE_ACTIVITY,
            .windowActivity = { .activityType = WINDOW_EVENT_ACTIVITY_TYPE_CLOSE }
        });
        shared_mem_flush_events(sm);
    }
    
    // Unlock cursor when window closes
    cocoa_unlock_cursor((__bridge void*)window);
}

@end

void* cocoa_create_window(const char* title, int x, int y, int width, int height, void** outWindow, void* goWindow) {
    @autoreleasepool {
        // Ensure NSApplication is initialized
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
        
        // Finish launching to enable full app behavior
        static dispatch_once_t launchToken;
        dispatch_once(&launchToken, ^{
            [NSApp finishLaunching];
        });
        
        // Set application delegate to handle dock icon clicks (only once)
        static dispatch_once_t delegateToken;
        dispatch_once(&delegateToken, ^{
            if ([NSApp delegate] == nil) {
                AppDelegate* appDelegate = [[AppDelegate alloc] init];
                [NSApp setDelegate:appDelegate];
            } else {
            }
        });
        
        // Create window
        NSRect frame = NSMakeRect(x, y, width, height);
        NSWindowStyleMask styleMask = NSWindowStyleMaskTitled | 
                                      NSWindowStyleMaskClosable | 
                                      NSWindowStyleMaskMiniaturizable | 
                                      NSWindowStyleMaskResizable;
        
        NSWindow* window = [[NSWindow alloc] initWithContentRect:frame
                                                        styleMask:styleMask
                                                          backing:NSBackingStoreBuffered
                                                            defer:NO];
        
        [window setTitle:[NSString stringWithUTF8String:title]];
        [window makeKeyAndOrderFront:nil];
        
        // Set up window delegate for resize events
        WindowDelegate* delegate = [[WindowDelegate alloc] init];
        [window setDelegate:delegate];
        
        // Create Metal view
        MetalView* view = [[MetalView alloc] initWithFrame:frame];
        [window setContentView:view];
        
        // Allocate and initialize SharedMem
        SharedMem* sm = calloc(1, sizeof(SharedMem));
        sm->goWindow = goWindow;
        sm->windowWidth = width;
        sm->windowHeight = height;
        sm->x = x;
        sm->y = y;
        
        // Store SharedMem pointer in window for event handling
        objc_setAssociatedObject(window, "sharedMem", [NSValue valueWithPointer:sm], OBJC_ASSOCIATION_RETAIN);
        
        // CFBridgingRetain transfers ownership to C, preventing ARC from deallocating.
        // These pointers must be released later with CFBridgingRelease.
        void* retainedWindow = (void*)CFBridgingRetain(window);
        void* retainedView = (void*)CFBridgingRetain(view);
        
        // Send SET_HANDLE event to communicate handles back to Go (matches Windows pattern).
        // This ensures handles are set through the event system rather than manual assignment,
        // maintaining consistent initialization order across platforms.
        shared_mem_add_event(sm, (WindowEvent) {
            .type = WINDOW_EVENT_TYPE_SET_HANDLE,
            .setHandle = {
                .hwnd = retainedView,      // NSView* - required for Vulkan surface creation
                .instance = retainedWindow, // NSWindow* - required for window operations
            }
        });
        shared_mem_flush_events(sm);
        
        // Store window pointer if requested
        if (outWindow != NULL) {
            *outWindow = retainedWindow;
        }
        
        return retainedView;
    }
}

void cocoa_destroy_window(void* nsView) {
    @autoreleasepool {
        if (nsView == NULL) return;
        
        // nsView is actually the MetalView, get the window from it
        NSView* view = (__bridge NSView*)(nsView);
        NSWindow* window = [view window];
        
        if (window == NULL) return;
        
        // Free SharedMem if allocated
        SharedMem* sm = getSharedMem(window);
        if (sm != NULL) {
            free(sm);
        }
        
        [window close];
        CFBridgingRelease(nsView);
    }
}

void cocoa_show_window(void* nsWindow) {
    @autoreleasepool {
        if (nsWindow == NULL) return;
        
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        [window makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:YES];
    }
}

void cocoa_poll_events(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        SharedMem* sm = getSharedMem(window);
        if (sm == NULL) return;
        
        NSEvent* event;
        while ((event = [NSApp nextEventMatchingMask:NSEventMaskAny
                                          untilDate:[NSDate distantPast]
                                             inMode:NSDefaultRunLoopMode
                                            dequeue:YES]) != nil) {
            
            // Process mouse events
            switch ([event type]) {
                case NSEventTypeMouseMoved:
                case NSEventTypeLeftMouseDragged:
                case NSEventTypeRightMouseDragged:
                case NSEventTypeOtherMouseDragged: {
                    NSPoint location = [event locationInWindow];
                    // Cocoa uses bottom-left origin; convert to top-left origin to match Windows/Linux
                    int32_t x = (int32_t)location.x;
                    int32_t y = sm->windowHeight - (int32_t)location.y;
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_MOUSE_MOVE,
                        .mouseMove = {
                            .x = x,
                            .y = y,
                        }
                    });
                    
                    // Apply cursor lock if active (matches Windows behavior)
                    if (sm->lockCursor.active) {
                        NSPoint windowPoint = NSMakePoint(sm->lockCursor.x, sm->lockCursor.y);
                        NSPoint screenPoint = [window convertPointToScreen:windowPoint];
                        CGWarpMouseCursorPosition(NSPointToCGPoint(screenPoint));
                    }
                    break;
                }
                
                case NSEventTypeLeftMouseDown: {
                    NSPoint location = [event locationInWindow];
                    int32_t x = (int32_t)location.x;
                    int32_t y = sm->windowHeight - (int32_t)location.y;
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
                        .mouseButton = {
                            .buttonId = MOUSE_BUTTON_LEFT,
                            .action = WINDOW_EVENT_BUTTON_TYPE_DOWN,
                            .x = x,
                            .y = y,
                        }
                    });
                    break;
                }
                
                case NSEventTypeLeftMouseUp: {
                    NSPoint location = [event locationInWindow];
                    int32_t x = (int32_t)location.x;
                    int32_t y = sm->windowHeight - (int32_t)location.y;
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
                        .mouseButton = {
                            .buttonId = MOUSE_BUTTON_LEFT,
                            .action = WINDOW_EVENT_BUTTON_TYPE_UP,
                            .x = x,
                            .y = y,
                        }
                    });
                    break;
                }
                
                case NSEventTypeRightMouseDown: {
                    NSPoint location = [event locationInWindow];
                    int32_t x = (int32_t)location.x;
                    int32_t y = sm->windowHeight - (int32_t)location.y;
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
                        .mouseButton = {
                            .buttonId = MOUSE_BUTTON_RIGHT,
                            .action = WINDOW_EVENT_BUTTON_TYPE_DOWN,
                            .x = x,
                            .y = y,
                        }
                    });
                    break;
                }
                
                case NSEventTypeRightMouseUp: {
                    NSPoint location = [event locationInWindow];
                    int32_t x = (int32_t)location.x;
                    int32_t y = sm->windowHeight - (int32_t)location.y;
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
                        .mouseButton = {
                            .buttonId = MOUSE_BUTTON_RIGHT,
                            .action = WINDOW_EVENT_BUTTON_TYPE_UP,
                            .x = x,
                            .y = y,
                        }
                    });
                    break;
                }
                
                case NSEventTypeOtherMouseDown: {
                    NSPoint location = [event locationInWindow];
                    int32_t x = (int32_t)location.x;
                    int32_t y = sm->windowHeight - (int32_t)location.y;
                    int32_t buttonId = MOUSE_BUTTON_MIDDLE;
                    if ([event buttonNumber] == 3) buttonId = MOUSE_BUTTON_X1;
                    else if ([event buttonNumber] == 4) buttonId = MOUSE_BUTTON_X2;
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
                        .mouseButton = {
                            .buttonId = buttonId,
                            .action = WINDOW_EVENT_BUTTON_TYPE_DOWN,
                            .x = x,
                            .y = y,
                        }
                    });
                    break;
                }
                
                case NSEventTypeOtherMouseUp: {
                    NSPoint location = [event locationInWindow];
                    int32_t x = (int32_t)location.x;
                    int32_t y = sm->windowHeight - (int32_t)location.y;
                    int32_t buttonId = MOUSE_BUTTON_MIDDLE;
                    if ([event buttonNumber] == 3) buttonId = MOUSE_BUTTON_X1;
                    else if ([event buttonNumber] == 4) buttonId = MOUSE_BUTTON_X2;
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
                        .mouseButton = {
                            .buttonId = buttonId,
                            .action = WINDOW_EVENT_BUTTON_TYPE_UP,
                            .x = x,
                            .y = y,
                        }
                    });
                    break;
                }
                
                case NSEventTypeScrollWheel: {
                    NSPoint location = [event locationInWindow];
                    // NSEvent provides deltaX/Y in points. Multiply by 120 to match Windows scroll wheel units.
                    int32_t deltaX = (int32_t)([event scrollingDeltaX] * 120.0);
                    int32_t deltaY = (int32_t)([event scrollingDeltaY] * 120.0);
                    
                    if (deltaX != 0 || deltaY != 0) {
                        shared_mem_add_event(sm, (WindowEvent) {
                            .type = WINDOW_EVENT_TYPE_MOUSE_SCROLL,
                            .mouseScroll = {
                                .deltaX = deltaX,
                                .deltaY = deltaY,
                                .x = (int32_t)location.x,
                                .y = (int32_t)location.y,
                            }
                        });
                    }
                    break;
                }
                
                case NSEventTypeFlagsChanged: {
                    // Handle modifier key changes (Cmd, Ctrl, Shift, Alt, etc.)
                    NSEventModifierFlags flags = [event modifierFlags];
                    unsigned short keyCode = [event keyCode];
                    
                    // Determine if this is a press or release by checking the flag state
                    // For modifier keys, we need to check if the corresponding flag is set
                    BOOL isPressed = NO;
                    switch (keyCode) {
                        case 0x37: // Left Command
                        case 0x36: // Right Command
                            isPressed = (flags & NSEventModifierFlagCommand) != 0;
                            break;
                        case 0x3B: // Left Control
                        case 0x3E: // Right Control
                            isPressed = (flags & NSEventModifierFlagControl) != 0;
                            break;
                        case 0x38: // Left Shift
                        case 0x3C: // Right Shift
                            isPressed = (flags & NSEventModifierFlagShift) != 0;
                            break;
                        case 0x3A: // Left Option/Alt
                        case 0x3D: // Right Option/Alt
                            isPressed = (flags & NSEventModifierFlagOption) != 0;
                            break;
                        default:
                            break;
                    }
                    
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
                        .keyboardButton = {
                            .buttonId = keyCode,
                            .action = isPressed ? WINDOW_EVENT_BUTTON_TYPE_DOWN : WINDOW_EVENT_BUTTON_TYPE_UP,
                        }
                    });
                    break;
                }
                
                case NSEventTypeKeyDown: {
                    unsigned short keyCode = [event keyCode];
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
                        .keyboardButton = {
                            .buttonId = keyCode,
                            .action = WINDOW_EVENT_BUTTON_TYPE_DOWN,
                        }
                    });
                    break;
                }
                
                case NSEventTypeKeyUp: {
                    unsigned short keyCode = [event keyCode];
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
                        .keyboardButton = {
                            .buttonId = keyCode,
                            .action = WINDOW_EVENT_BUTTON_TYPE_UP,
                        }
                    });
                    break;
                }
                
                default:
                    break;
            }
            
            // Forward event to application for standard handling
            [NSApp sendEvent:event];
        }
        
        // Flush any accumulated events
        shared_mem_flush_events(sm);
    }
}

float cocoa_get_dpi(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        NSScreen* screen = [window screen];
        if (screen == nil) {
            screen = [NSScreen mainScreen];
        }
        
        NSDictionary* description = [screen deviceDescription];
        NSSize displayPixelSize = [[description objectForKey:NSDeviceSize] sizeValue];
        CGSize displayPhysicalSize = CGDisplayScreenSize(
            [[description objectForKey:@"NSScreenNumber"] unsignedIntValue]);
        
        // Physical size is in millimeters, convert to inches
        float widthInInches = displayPhysicalSize.width / 25.4;
        float dpi = displayPixelSize.width / widthInInches;
        
        return dpi > 0 ? dpi : 72.0; // Default to 72 DPI if calculation fails
    }
}

void cocoa_screen_size_mm(void* nsWindow, int* width, int* height) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        NSScreen* screen = [window screen];
        if (screen == nil) {
            screen = [NSScreen mainScreen];
        }
        
        NSDictionary* description = [screen deviceDescription];
        CGSize displayPhysicalSize = CGDisplayScreenSize(
            [[description objectForKey:@"NSScreenNumber"] unsignedIntValue]);
        
        *width = (int)displayPhysicalSize.width;
        *height = (int)displayPhysicalSize.height;
    }
}

void cocoa_get_position(void* nsWindow, int* x, int* y) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        NSRect frame = [window frame];
        *x = (int)frame.origin.x;
        *y = (int)frame.origin.y;
    }
}

void cocoa_set_position(void* nsWindow, int x, int y) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        NSPoint point = NSMakePoint(x, y);
        [window setFrameOrigin:point];
    }
}

void cocoa_set_size(void* nsWindow, int width, int height) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        NSRect frame = [window frame];
        frame.size.width = width;
        frame.size.height = height;
        [window setFrame:frame display:YES animate:NO];
    }
}

void cocoa_set_title(void* nsWindow, const char* title) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        [window setTitle:[NSString stringWithUTF8String:title]];
    }
}

void cocoa_copy_to_clipboard(const char* text) {
    @autoreleasepool {
        NSPasteboard* pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        [pasteboard setString:[NSString stringWithUTF8String:text] forType:NSPasteboardTypeString];
    }
}

char* cocoa_clipboard_contents(void) {
    @autoreleasepool {
        NSPasteboard* pasteboard = [NSPasteboard generalPasteboard];
        NSString* content = [pasteboard stringForType:NSPasteboardTypeString];
        if (content == nil) {
            return strdup("");
        }
        return strdup([content UTF8String]);
    }
}

void cocoa_cursor_standard(void) {
    @autoreleasepool {
        [[NSCursor arrowCursor] set];
    }
}

void cocoa_cursor_ibeam(void) {
    @autoreleasepool {
        [[NSCursor IBeamCursor] set];
    }
}

void cocoa_cursor_size_all(void) {
    @autoreleasepool {
        [[NSCursor closedHandCursor] set];
    }
}

void cocoa_cursor_size_ns(void) {
    @autoreleasepool {
        [[NSCursor resizeUpDownCursor] set];
    }
}

void cocoa_cursor_size_we(void) {
    @autoreleasepool {
        [[NSCursor resizeLeftRightCursor] set];
    }
}

void cocoa_show_cursor(void) {
    @autoreleasepool {
        [NSCursor unhide];
    }
}

void cocoa_hide_cursor(void) {
    @autoreleasepool {
        [NSCursor hide];
    }
}

void cocoa_focus_window(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        [window makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:YES];
    }
}

void cocoa_get_bundle_resource_path(const char* resourceName, void** outPath) {
    @autoreleasepool {
        NSString* name = [NSString stringWithUTF8String:resourceName];
        NSString* path = [[NSBundle mainBundle] pathForResource:name ofType:nil];
        
        if (path == nil) {
            // Try with extension separated
            NSString* extension = [name pathExtension];
            NSString* basename = [name stringByDeletingPathExtension];
            path = [[NSBundle mainBundle] pathForResource:basename ofType:extension];
        }
        
        if (path != nil) {
            *outPath = (void*)strdup([path UTF8String]);
        } else {
            *outPath = NULL;
        }
    }
}

void cocoa_remove_border(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        NSWindowStyleMask styleMask = [window styleMask];
        // Remove title bar and border decorations
        styleMask &= ~(NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | 
                       NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable);
        styleMask |= NSWindowStyleMaskBorderless;
        [window setStyleMask:styleMask];
    }
}

void cocoa_add_border(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        NSWindowStyleMask styleMask = [window styleMask];
        // Add title bar and border decorations
        styleMask &= ~NSWindowStyleMaskBorderless;
        styleMask |= (NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | 
                      NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable);
        [window setStyleMask:styleMask];
    }
}

void cocoa_set_fullscreen(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        // Check if already in fullscreen
        if (([window styleMask] & NSWindowStyleMaskFullScreen) == 0) {
            [window toggleFullScreen:nil];
        }
    }
}

void cocoa_set_windowed(void* nsWindow, int width, int height) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        // Exit fullscreen if needed
        if (([window styleMask] & NSWindowStyleMaskFullScreen) != 0) {
            [window toggleFullScreen:nil];
        }
        // Set the requested window size
        NSRect frame = [window frame];
        frame.size.width = width;
        frame.size.height = height;
        [window setFrame:frame display:YES animate:YES];
    }
}

void cocoa_lock_cursor(void* nsWindow, int x, int y) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        SharedMem* sm = getSharedMem(window);
        if (sm == NULL) return;
        
        sm->lockCursor.x = x;
        sm->lockCursor.y = y;
        sm->lockCursor.active = true;
    }
}

void cocoa_unlock_cursor(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        SharedMem* sm = getSharedMem(window);
        if (sm == NULL) return;
        
        sm->lockCursor.active = false;
    }
}


// Enable raw mouse input for game mode (mouselook): hides and decouples cursor
void cocoa_enable_raw_mouse(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        SharedMem* sm = getSharedMem(window);
        if (sm == NULL) return;
        if (sm->rawInputRequested) {
            CGAssociateMouseAndMouseCursorPosition(NO);
            [NSCursor hide];
        }
    }
}

// Disable raw mouse input: restores normal cursor behavior
void cocoa_disable_raw_mouse(void* nsWindow) {
    @autoreleasepool {
        NSWindow* window = (__bridge NSWindow*)(nsWindow);
        SharedMem* sm = getSharedMem(window);
        if (sm == NULL) return;
        sm->rawInputRequested = false;
        CGAssociateMouseAndMouseCursorPosition(YES);
        [NSCursor unhide];
    }
}

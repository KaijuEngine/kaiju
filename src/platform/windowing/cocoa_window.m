#import <Cocoa/Cocoa.h>
#import <QuartzCore/CAMetalLayer.h>
#import <objc/runtime.h>
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

// Window delegate to handle resize events.
// Uses NSWindowDelegate to receive resize notifications only when size actually changes,
// preventing continuous swapchain recreation that would occur with polling-based detection.
@interface WindowDelegate : NSObject <NSWindowDelegate>
@end

@implementation WindowDelegate

- (void)windowDidResize:(NSNotification *)notification {
    NSWindow* window = [notification object];
    NSValue* value = objc_getAssociatedObject(window, "sharedMem");
    if (value == nil) return;
    SharedMem* sm = [value pointerValue];
    
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

@end

void* cocoa_create_window(const char* title, int x, int y, int width, int height, void** outWindow, void* goWindow) {
    @autoreleasepool {
        // Ensure NSApplication is initialized
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
        
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

void cocoa_destroy_window(void* nsWindow) {
    @autoreleasepool {
        if (nsWindow != NULL) {
            NSWindow* window = (NSWindow*)CFBridgingRelease(nsWindow);
            [window close];
        }
    }
}

void cocoa_show_window(void* nsWindow) {
    @autoreleasepool {
        if (nsWindow != NULL) {
            NSWindow* window = (__bridge NSWindow*)(nsWindow);
            [window makeKeyAndOrderFront:nil];
            [NSApp activateIgnoringOtherApps:YES];
        }
    }
}

void cocoa_poll_events(void* nsWindow) {
    NSWindow* window = (__bridge NSWindow*)(nsWindow);
    NSValue* value = objc_getAssociatedObject(window, "sharedMem");
    if (value == nil) return;
    SharedMem* sm = [value pointerValue];
    @autoreleasepool {
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
                    // Cocoa uses bottom-left origin, convert to top-left to match Windows/Linux.
                    // Y-axis is flipped: top = 0, bottom = height
                    int32_t x = (int32_t)location.x;
                    int32_t y = sm->windowHeight - (int32_t)location.y;
                    shared_mem_add_event(sm, (WindowEvent) {
                        .type = WINDOW_EVENT_TYPE_MOUSE_MOVE,
                        .mouseMove = {
                            .x = x,
                            .y = y,
                        }
                    });
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

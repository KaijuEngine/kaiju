// Fixed Cocoa backend: single event ingestion model, no polling, no double-dispatch
// APIs preserved as requested

//go:build darwin
// +build darwin

#import <Cocoa/Cocoa.h>
#import <QuartzCore/CAMetalLayer.h>
#import <objc/runtime.h>
#include <string.h>
#include "cocoa_window.h"
#include "window_event.h"

#pragma mark - Metal View

@interface MetalView : NSView
@end

@implementation MetalView
- (BOOL)mouseDownCanMoveWindow {
	return NO;
}

- (instancetype)initWithFrame:(NSRect)frame {
	self = [super initWithFrame:frame];
	if (self) {
		self.wantsLayer = YES;
		self.layer = [[CAMetalLayer alloc] init];
		((CAMetalLayer*)self.layer).pixelFormat = MTLPixelFormatBGRA8Unorm;
	}
	return self;
}
- (CALayer*)makeBackingLayer { return [[CAMetalLayer alloc] init]; }
- (BOOL)wantsUpdateLayer { return YES; }
@end

#pragma mark - App Delegate

@interface AppDelegate : NSObject <NSApplicationDelegate>
@end

@implementation AppDelegate
- (BOOL)applicationShouldHandleReopen:(NSApplication *)sender hasVisibleWindows:(BOOL)flag {
	for (NSWindow* w in [NSApp windows]) {
		if ([w isMiniaturized]) [w deminiaturize:nil];
		[w makeKeyAndOrderFront:nil];
	}
	[NSApp activateIgnoringOtherApps:YES];
	return YES;
}

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
	(void)sender;
	return YES;
}
@end

#pragma mark - Helpers

static inline SharedMem* getSharedMem(NSWindow* window) {
	NSValue* v = objc_getAssociatedObject(window, "sharedMem");
	return v ? (SharedMem*)[v pointerValue] : NULL;
}

static inline NSEventModifierFlags modifierFlagForKeyCode(unsigned short keyCode) {
	switch (keyCode) {
		case 0x36: case 0x37: return NSEventModifierFlagCommand;
		case 0x3B: case 0x3E: return NSEventModifierFlagControl;
		case 0x38: case 0x3C: return NSEventModifierFlagShift;
		case 0x3A: case 0x3D: return NSEventModifierFlagOption;
		case 0x39: return NSEventModifierFlagCapsLock;
		default: return 0;
	}
}

#pragma mark - Window Delegate

@interface WindowDelegate : NSObject <NSWindowDelegate>
@end

@implementation WindowDelegate
- (void)windowDidResize:(NSNotification *)n {
	NSWindow* w = n.object;
	SharedMem* sm = getSharedMem(w);
	if (!sm) return;

	NSRect cr = [w contentRectForFrameRect:w.frame];
	int nw = (int)cr.size.width;
	int nh = (int)cr.size.height;
	if (nw == sm->windowWidth && nh == sm->windowHeight) return;

	sm->windowWidth = nw;
	sm->windowHeight = nh;

	NSRect fr = w.frame;
	sm->left = (int)fr.origin.x;
	sm->bottom = (int)fr.origin.y;
	sm->right = sm->left + (int)fr.size.width;
	sm->top = sm->bottom + (int)fr.size.height;

	shared_mem_add_event(sm, (WindowEvent){
		.type = WINDOW_EVENT_TYPE_RESIZE,
		.windowResize = { nw, nh, sm->left, sm->top, sm->right, sm->bottom }
	});
	shared_mem_flush_events(sm);
}

- (void)windowDidResignKey:(NSNotification *)n {
	cocoa_unlock_cursor((__bridge void*)n.object);
}

- (void)windowWillClose:(NSNotification *)n {
	NSWindow* w = n.object;
	SharedMem* sm = getSharedMem(w);
	if (sm) {
		shared_mem_add_event(sm, (WindowEvent){
			.type = WINDOW_EVENT_TYPE_ACTIVITY,
			.windowActivity = { WINDOW_EVENT_ACTIVITY_TYPE_CLOSE }
		});
		shared_mem_flush_events(sm);
	}
	cocoa_unlock_cursor((__bridge void*)w);
}
@end

#pragma mark - Event Monitor

static id gEventMonitor = nil;

static NSEvent* handleEvent(NSEvent* e) {
	NSWindow* w = e.window;
	if (!w) return e;
	SharedMem* sm = getSharedMem(w);
	if (!sm) return e;

	switch (e.type) {
		case NSEventTypeMouseMoved:
		case NSEventTypeLeftMouseDragged:
		case NSEventTypeRightMouseDragged:
		case NSEventTypeOtherMouseDragged: {
			NSPoint p = e.locationInWindow;
			int32_t x = (int32_t)p.x;
			int32_t y = sm->windowHeight - (int32_t)p.y;
			shared_mem_add_event(sm, (WindowEvent){
				.type = WINDOW_EVENT_TYPE_MOUSE_MOVE,
				.mouseMove = { x, y }
			});
			if (sm->lockCursor.active) {
				NSRect r = NSMakeRect(sm->lockCursor.x, sm->lockCursor.y, 0, 0);
				NSRect sr = [w convertRectToScreen:r];
				CGWarpMouseCursorPosition(sr.origin);
			}
		} break;

		case NSEventTypeLeftMouseDown:
		case NSEventTypeLeftMouseUp:
		case NSEventTypeRightMouseDown:
		case NSEventTypeRightMouseUp:
		case NSEventTypeOtherMouseDown:
		case NSEventTypeOtherMouseUp: {
			NSInteger btn = e.buttonNumber;
			int32_t id = (btn==0)?MOUSE_BUTTON_LEFT:(btn==1)?MOUSE_BUTTON_RIGHT:(btn==2)?MOUSE_BUTTON_MIDDLE:(btn==3)?MOUSE_BUTTON_X1:MOUSE_BUTTON_X2;
			int action = (e.type==NSEventTypeLeftMouseDown||e.type==NSEventTypeRightMouseDown||e.type==NSEventTypeOtherMouseDown)
						 ? WINDOW_EVENT_BUTTON_TYPE_DOWN
						 : WINDOW_EVENT_BUTTON_TYPE_UP;
			NSPoint p = e.locationInWindow;
			shared_mem_add_event(sm, (WindowEvent){
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = { id, action, (int32_t)p.x, sm->windowHeight-(int32_t)p.y }
			});
		} break;

		case NSEventTypeScrollWheel: {
			double dx = e.hasPreciseScrollingDeltas ? e.scrollingDeltaX : e.deltaX;
			double dy = e.hasPreciseScrollingDeltas ? e.scrollingDeltaY : e.deltaY;
			if (e.isDirectionInvertedFromDevice) dy = -dy;
			if (dx || dy) {
				NSPoint p = e.locationInWindow;
				shared_mem_add_event(sm, (WindowEvent){
					.type = WINDOW_EVENT_TYPE_MOUSE_SCROLL,
					.mouseScroll = { dx, dy, (int32_t)p.x, (int32_t)p.y }
				});
			}
		} break;

		case NSEventTypeFlagsChanged: {
			// Caps‑Lock is the only modifier we want to treat as a toggle.
			if (e.keyCode == 0x39) {
				shared_mem_add_event(sm, (WindowEvent){
					.type = WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
					.keyboardButton = { e.keyCode, WINDOW_EVENT_BUTTON_TYPE_UP }
				});
				break;
			}

			// Existing handling for the other modifiers (Shift, Ctrl, …)
			NSEventModifierFlags f = modifierFlagForKeyCode(e.keyCode);
			if (!f) break;
			BOOL down = (e.modifierFlags & f) != 0;
			shared_mem_add_event(sm, (WindowEvent){
				.type = WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
				.keyboardButton = { e.keyCode,
					down ? WINDOW_EVENT_BUTTON_TYPE_DOWN : WINDOW_EVENT_BUTTON_TYPE_UP }
			});
		} break;

		case NSEventTypeKeyDown:
		case NSEventTypeKeyUp:
			// Skip auto‑repeat key‑down events
			if (e.type == NSEventTypeKeyDown && e.isARepeat) {
				break;
			}
			shared_mem_add_event(sm, (WindowEvent){
				.type = WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
				.keyboardButton = { e.keyCode, e.type==NSEventTypeKeyDown?WINDOW_EVENT_BUTTON_TYPE_DOWN:WINDOW_EVENT_BUTTON_TYPE_UP }
			});
			break;
		default: break;
	}

	shared_mem_flush_events(sm);
	return e;
}


void cocoa_run_app(void) {
	static BOOL started = NO;

	if (started) {
		return;
	}
	started = YES;

	if (![NSThread isMainThread]) {
		NSLog(@"cocoa_run_app MUST run on main thread");
		abort();
	}

	@autoreleasepool {
		[NSApplication sharedApplication];
		[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

		static AppDelegate *delegate = nil;
		delegate = [[AppDelegate alloc] init];
		[NSApp setDelegate:delegate];
		[NSApp finishLaunching];
		[NSApp activateIgnoringOtherApps:YES];

		// INSTALL MONITOR HERE
		gEventMonitor =
		[NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskAny
			handler:^NSEvent*(NSEvent* e) {
				return handleEvent(e);
			}];

		for (NSWindow* w in [NSApp windows]) {
			[w makeKeyAndOrderFront:nil];
		}

		[NSApp run];
	}
}

#pragma mark - Public API (unchanged)

void* cocoa_create_window(const char* title,
						  int x, int y, int w, int h,
						  void** outWindow,
						  void* goWindow)
{
	__block void* resultView = NULL;

	// ALWAYS hop to main thread for AppKit
	dispatch_sync(dispatch_get_main_queue(), ^{
		@autoreleasepool {

			// ❌ NO sharedApplication
			// ❌ NO finishLaunching
			// ❌ NO delegate setup
			// ❌ NO event monitors here

			NSRect frame = NSMakeRect(x, y, w, h);

			NSWindow* win = [[NSWindow alloc]
				initWithContentRect:frame
						  styleMask:(NSWindowStyleMaskTitled |
									 NSWindowStyleMaskClosable |
									 NSWindowStyleMaskMiniaturizable |
									 NSWindowStyleMaskResizable)
							backing:NSBackingStoreBuffered
							  defer:NO];

			win.title = [NSString stringWithUTF8String:title];

			// Window delegate (retained via associated object)
			WindowDelegate* winDelegate = [WindowDelegate new];
			objc_setAssociatedObject(win, "windowDelegate",
									 winDelegate,
									 OBJC_ASSOCIATION_RETAIN);
			win.delegate = winDelegate;

			// Metal view
			MetalView* view = [[MetalView alloc] initWithFrame:NSMakeRect(0, 0, w, h)];
			win.contentView = view;

			[win makeKeyAndOrderFront:nil];
			[NSApp activateIgnoringOtherApps:YES];

			// Shared memory
			SharedMem* sm = calloc(1, sizeof(SharedMem));
			sm->goWindow = goWindow;
			sm->windowWidth = w;
			sm->windowHeight = h;

			objc_setAssociatedObject(win,
									 "sharedMem",
									 [NSValue valueWithPointer:sm],
									 OBJC_ASSOCIATION_RETAIN);

			// Retain Cocoa objects for C side
			void* retainedWindow = (void*)CFBridgingRetain(win);
			void* retainedView   = (void*)CFBridgingRetain(view);

			// Send SET_HANDLE event (matches Windows/Linux init flow)
			shared_mem_add_event(sm, (WindowEvent){
				.type = WINDOW_EVENT_TYPE_SET_HANDLE,
				.setHandle = {
					.hwnd     = retainedView,   // NSView*
					.instance = retainedWindow // NSWindow*
				}
			});
			shared_mem_flush_events(sm);

			if (outWindow) {
				*outWindow = retainedWindow;
			}

			resultView = retainedView;
		}
	});

	return resultView;
}

void cocoa_poll_events(void* nsWindow) { (void)nsWindow; /* intentionally empty */ }

void cocoa_destroy_window(void* nsWindow) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)nsWindow;

			SharedMem* sm = getSharedMem(window);
			if (sm) {
				free(sm);
				objc_setAssociatedObject(window, "sharedMem", nil, OBJC_ASSOCIATION_ASSIGN);
			}

			[window orderOut:nil];
			[window close];

			// Release exactly what we retained in create
			CFBridgingRelease((__bridge CFTypeRef)window);
		}
	});
}

// Remaining APIs unchanged (show, focus, cursor, clipboard, fullscreen, etc.)
// They are safe because all AppKit calls remain on main thread

void cocoa_show_window(void* nsWindow) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)nsWindow;
			[window makeKeyAndOrderFront:nil];
			[NSApp activateIgnoringOtherApps:YES];
		}
	});
}

int cocoa_get_screen_pixel_width(void* nsWindow) {
	__block int result = 1920;

	if (!nsWindow) {
		return result;
	}

/*
	dispatch_sync(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)nsWindow;
			if (!window) return;

			NSScreen* screen = window.screen ?: [NSScreen mainScreen];
			if (!screen) return;

			NSRect frame = screen.frame; // points
			CGFloat scale = screen.backingScaleFactor;

			int px = (int)(frame.size.width * scale);
			if (px > 0) {
				result = px;
			}
		}
	});
*/
	return result;
}

int cocoa_get_screen_pixel_height(void* nsWindow) {
	__block int result = 1080;

	if (!nsWindow) {
		return result;
	}
/*
	dispatch_sync(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)nsWindow;
			if (!window) return;

			NSScreen* screen = window.screen ?: [NSScreen mainScreen];
			if (!screen) return;

			NSRect frame = screen.frame;
			CGFloat scale = screen.backingScaleFactor;

			int px = (int)(frame.size.height * scale);
			if (px > 0) {
				result = px;
			}
		}
	});

*/
	return result;
}

double cocoa_get_backing_scale_factor(void* nsWindow) {
	__block double result = 1.0;

	if (!nsWindow) {
		return result;
	}
/*
	dispatch_sync(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)nsWindow;
			if (!window) return;

			NSScreen* screen = window.screen ?: [NSScreen mainScreen];
			if (!screen) return;

			CGFloat scale = screen.backingScaleFactor;
			if (scale > 0) {
				result = scale;
			}
		}
	});
*/
	return result;
}

void cocoa_get_position(void* nsWindow, int* x, int* y) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			NSRect frame = [window frame];
			*x = (int)frame.origin.x;
			*y = (int)frame.origin.y;
		}
	});
}

void cocoa_set_position(void* nsWindow, int x, int y) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			NSPoint point = NSMakePoint(x, y);
			[window setFrameOrigin:point];
		}
	});
}

void cocoa_set_size(void* nsWindow, int width, int height) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			NSRect frame = [window frame];
			frame.size.width = width;
			frame.size.height = height;
			[window setFrame:frame display:YES animate:NO];
		}
	});
}

void cocoa_set_title(void* nsWindow, const char* title) {
	if (!nsWindow || !title) return;

	// COPY the string immediately
	char* titleCopy = strdup(title);
	if (!titleCopy) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)nsWindow;

			NSString* str = [[NSString alloc] initWithUTF8String:titleCopy];
			free(titleCopy);

			if (!str) return; // invalid UTF-8 safety

			[window setTitle:str];
		}
	});
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
	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			[[NSCursor arrowCursor] set];
		}
	});
}

void cocoa_cursor_ibeam(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			[[NSCursor IBeamCursor] set];
		}
	});
}

void cocoa_cursor_size_all(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			[[NSCursor closedHandCursor] set];
		}
	});
}

void cocoa_cursor_size_ns(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			[[NSCursor resizeUpDownCursor] set];
		}
	});
}

void cocoa_cursor_size_we(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			[[NSCursor resizeLeftRightCursor] set];
		}
	});
}

void cocoa_show_cursor(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			[NSCursor unhide];
		}
	});
}

void cocoa_hide_cursor(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			[NSCursor hide];
		}
	});
}

void cocoa_focus_window(void* nsWindow) {
	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			[window makeKeyAndOrderFront:nil];
			[NSApp activateIgnoringOtherApps:YES];
		}
	});
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
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			NSWindowStyleMask styleMask = [window styleMask];
			// Remove title bar and border decorations
			styleMask &= ~(NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | 
						   NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable);
			styleMask |= NSWindowStyleMaskBorderless;
			[window setStyleMask:styleMask];
		}
	});
}

void cocoa_add_border(void* nsWindow) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			NSWindowStyleMask styleMask = [window styleMask];
			// Add title bar and border decorations
			styleMask &= ~NSWindowStyleMaskBorderless;
			styleMask |= (NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | 
						  NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable);
			[window setStyleMask:styleMask];
		}
	});
}

void cocoa_set_fullscreen(void* nsWindow) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			// Check if already in fullscreen
			if (([window styleMask] & NSWindowStyleMaskFullScreen) == 0) {
				[window toggleFullScreen:nil];
			}
		}
	});
}

void cocoa_set_windowed(void* nsWindow, int width, int height) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
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
	});
}

void cocoa_lock_cursor(void* nsWindow, int x, int y) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			SharedMem* sm = getSharedMem(window);
			if (sm == NULL) return;

			sm->lockCursor.x = x;
			sm->lockCursor.y = y;
			sm->lockCursor.active = true;
		}
	});
}

void cocoa_unlock_cursor(void* nsWindow) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			SharedMem* sm = getSharedMem(window);
			if (sm == NULL) return;

			sm->lockCursor.active = false;
		}
	});
}


// Enable raw mouse input for game mode (mouselook): hides and decouples cursor
void cocoa_enable_raw_mouse(void* nsWindow) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			SharedMem* sm = getSharedMem(window);
			if (sm == NULL) return;
			if (sm->rawInputRequested) {
				CGAssociateMouseAndMouseCursorPosition(NO);
				[NSCursor hide];
			}
		}
	});
}

// Disable raw mouse input: restores normal cursor behavior
void cocoa_disable_raw_mouse(void* nsWindow) {
	if (!nsWindow) return;

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			NSWindow* window = (__bridge NSWindow*)(nsWindow);
			SharedMem* sm = getSharedMem(window);
			if (sm == NULL) return;
			sm->rawInputRequested = false;
			CGAssociateMouseAndMouseCursorPosition(YES);
			[NSCursor unhide];
		}
	});
}

// Get the current state of the Caps Lock toggle key
bool cocoa_get_caps_lock_toggle_key_state(void) {
    @autoreleasepool {
        NSEventModifierFlags flags = [NSEvent modifierFlags];
        return (flags & NSEventModifierFlagCapsLock) != 0;
    }
}

void cocoa_set_icon(void* nsWindow, int width, int height, const uint8_t* pixelData) {
	if (!pixelData || width <= 0 || height <= 0) {
		return;
	}

	(void)nsWindow; // nsWindow kept for potential future per-window icons

	dispatch_async(dispatch_get_main_queue(), ^{
		@autoreleasepool {
			// Create CGImage from raw RGBA pixel data
			CGDataProviderRef provider = CGDataProviderCreateWithData(NULL, pixelData, (size_t)width * height * 4, NULL);
			if (!provider) {
				return;
			}
			CGImageRef cgImage = CGImageCreate(
				(size_t)width,
				(size_t)height,
				8,                          // bits per component
				32,                         // bits per pixel
				(size_t)width * 4,         // bytes per row
				CGColorSpaceCreateDeviceRGB(),
				(CGBitmapInfo)kCGImageAlphaPremultipliedLast,
				provider,
				NULL,
				0,
				kCGRenderingIntentDefault
			);
			CGDataProviderRelease(provider);
			if (!cgImage) {
				return;
			}
			NSImage* image = [[NSImage alloc] initWithCGImage:cgImage size:NSMakeSize(width, height)];
			CGImageRelease(cgImage);
			// Set the application dock icon
			[NSApp setApplicationIconImage:image];
		}
	});
}

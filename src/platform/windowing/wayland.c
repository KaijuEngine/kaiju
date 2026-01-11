//go:build linux && !android && wayland

/******************************************************************************/
/* wayland.c                                                                  */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

#if defined(__linux__) && defined(USE_WAYLAND) && !defined(__ANDROID__)

#include "wayland.h"
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <stdio.h>
#include <unistd.h>
#include <sys/mman.h>
#include <wayland-client.h>
#include <wayland-cursor.h>
#include <xkbcommon/xkbcommon.h>
#include "xdg-shell-client-protocol.h"
#include "xdg-decoration-client-protocol.h"

// Wayland docs:
// https://wayland-book.com/
// https://wayland.freedesktop.org/docs/html/

static WaylandState* global_state = NULL;

// XDG WM Base listener
static void xdg_wm_base_ping(void* data, struct xdg_wm_base* xdg_wm_base,
	uint32_t serial)
{
	xdg_wm_base_pong(xdg_wm_base, serial);
}

static const struct xdg_wm_base_listener xdg_wm_base_listener = {
	.ping = xdg_wm_base_ping,
};

// XDG Surface listener
static void xdg_surface_configure(void* data, struct xdg_surface* xdg_surface,
	uint32_t serial)
{
	WaylandState* s = data;
	xdg_surface_ack_configure(xdg_surface, serial);
	s->configured = true;
}

static const struct xdg_surface_listener xdg_surface_listener = {
	.configure = xdg_surface_configure,
};

// XDG Toplevel listener
static void xdg_toplevel_configure(void* data, struct xdg_toplevel* toplevel,
	int32_t width, int32_t height, struct wl_array* states)
{
	WaylandState* s = data;
	if (width > 0 && height > 0) {
		if (s->sm.windowWidth != width || s->sm.windowHeight != height) {
			s->sm.windowWidth = width;
			s->sm.windowHeight = height;
			shared_mem_add_event(&s->sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_RESIZE,
				.windowResize = {
					.width = width,
					.height = height,
					.left = s->sm.x,
					.top = s->sm.y,
					.right = s->sm.x + width,
					.bottom = s->sm.y + height,
				}
			});
		}
	}
}

static void xdg_toplevel_close(void* data, struct xdg_toplevel* toplevel) {
	WaylandState* s = data;
	s->closed = true;
	shared_mem_add_event(&s->sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_ACTIVITY,
		.windowActivity = {
			.activityType = WINDOW_EVENT_ACTIVITY_TYPE_CLOSE,
		}
	});
	shared_mem_flush_events(&s->sm);
}

static void xdg_toplevel_configure_bounds(void* data,
	struct xdg_toplevel* toplevel, int32_t width, int32_t height)
{
	// Optional: can be used to constrain window size
}

static void xdg_toplevel_wm_capabilities(void* data,
	struct xdg_toplevel* toplevel, struct wl_array* capabilities)
{
	// Optional: indicates which capabilities the compositor supports
}

static const struct xdg_toplevel_listener xdg_toplevel_listener = {
	.configure = xdg_toplevel_configure,
	.close = xdg_toplevel_close,
	.configure_bounds = xdg_toplevel_configure_bounds,
	.wm_capabilities = xdg_toplevel_wm_capabilities,
};

// Keyboard listener
static void keyboard_keymap(void* data, struct wl_keyboard* keyboard,
	uint32_t format, int fd, uint32_t size)
{
	WaylandState* s = data;
	if (format != WL_KEYBOARD_KEYMAP_FORMAT_XKB_V1) {
		close(fd);
		return;
	}
	char* map_str = mmap(NULL, size, PROT_READ, MAP_PRIVATE, fd, 0);
	if (map_str == MAP_FAILED) {
		close(fd);
		return;
	}
	if (s->xkb_keymap) {
		xkb_keymap_unref(s->xkb_keymap);
	}
	if (s->xkb_state) {
		xkb_state_unref(s->xkb_state);
	}
	s->xkb_keymap = xkb_keymap_new_from_string(s->xkb_context, map_str,
		XKB_KEYMAP_FORMAT_TEXT_V1, XKB_KEYMAP_COMPILE_NO_FLAGS);
	munmap(map_str, size);
	close(fd);
	if (s->xkb_keymap) {
		s->xkb_state = xkb_state_new(s->xkb_keymap);
	}
}

static void keyboard_enter(void* data, struct wl_keyboard* keyboard,
	uint32_t serial, struct wl_surface* surface, struct wl_array* keys)
{
	WaylandState* s = data;
	shared_mem_add_event(&s->sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_ACTIVITY,
		.windowActivity = {
			.activityType = WINDOW_EVENT_ACTIVITY_TYPE_FOCUS,
		}
	});
}

static void keyboard_leave(void* data, struct wl_keyboard* keyboard,
	uint32_t serial, struct wl_surface* surface)
{
	WaylandState* s = data;
	shared_mem_add_event(&s->sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_ACTIVITY,
		.windowActivity = {
			.activityType = WINDOW_EVENT_ACTIVITY_TYPE_BLUR,
		}
	});
}

static void keyboard_key(void* data, struct wl_keyboard* keyboard,
	uint32_t serial, uint32_t time, uint32_t key, uint32_t state)
{
	WaylandState* s = data;
	if (!s->xkb_state) return;
	// Convert evdev keycode to XKB keycode (add 8)
	xkb_keysym_t sym = xkb_state_key_get_one_sym(s->xkb_state, key + 8);
	shared_mem_add_event(&s->sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_KEYBOARD_BUTTON,
		.keyboardButton = {
			.action = state == WL_KEYBOARD_KEY_STATE_PRESSED
				? WINDOW_EVENT_BUTTON_TYPE_DOWN : WINDOW_EVENT_BUTTON_TYPE_UP,
			.buttonId = sym,
		}
	});
}

static void keyboard_modifiers(void* data, struct wl_keyboard* keyboard,
	uint32_t serial, uint32_t mods_depressed, uint32_t mods_latched,
	uint32_t mods_locked, uint32_t group)
{
	WaylandState* s = data;
	if (s->xkb_state) {
		xkb_state_update_mask(s->xkb_state, mods_depressed, mods_latched,
			mods_locked, 0, 0, group);
	}
}

static void keyboard_repeat_info(void* data, struct wl_keyboard* keyboard,
	int32_t rate, int32_t delay)
{
	// Optional: key repeat configuration
}

static const struct wl_keyboard_listener keyboard_listener = {
	.keymap = keyboard_keymap,
	.enter = keyboard_enter,
	.leave = keyboard_leave,
	.key = keyboard_key,
	.modifiers = keyboard_modifiers,
	.repeat_info = keyboard_repeat_info,
};

// Pointer listener
static void pointer_enter(void* data, struct wl_pointer* pointer,
	uint32_t serial, struct wl_surface* surface,
	wl_fixed_t sx, wl_fixed_t sy)
{
	WaylandState* s = data;
	// Set cursor when entering surface
	if (s->current_cursor && s->cursor_surface) {
		struct wl_cursor_image* image = s->current_cursor->images[0];
		wl_pointer_set_cursor(pointer, serial, s->cursor_surface,
			image->hotspot_x, image->hotspot_y);
	}
}

static void pointer_leave(void* data, struct wl_pointer* pointer,
	uint32_t serial, struct wl_surface* surface)
{
	// Pointer left the surface
}

static void pointer_motion(void* data, struct wl_pointer* pointer,
	uint32_t time, wl_fixed_t sx, wl_fixed_t sy)
{
	WaylandState* s = data;
	int x = wl_fixed_to_int(sx);
	int y = wl_fixed_to_int(sy);
	shared_mem_add_event(&s->sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_MOUSE_MOVE,
		.mouseMove = {
			.x = x,
			.y = y,
		}
	});
	s->sm.mouseX = x;
	s->sm.mouseY = y;
}

static void pointer_button(void* data, struct wl_pointer* pointer,
	uint32_t serial, uint32_t time, uint32_t button, uint32_t state)
{
	WaylandState* s = data;
	WindowEvent evt = {
		.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
		.mouseButton.action = state == WL_POINTER_BUTTON_STATE_PRESSED
			? WINDOW_EVENT_BUTTON_TYPE_DOWN : WINDOW_EVENT_BUTTON_TYPE_UP,
		.mouseButton.x = s->sm.mouseX,
		.mouseButton.y = s->sm.mouseY,
	};
	// Linux input event codes: BTN_LEFT=272, BTN_RIGHT=273, BTN_MIDDLE=274
	switch (button) {
		case 272: // BTN_LEFT
			evt.mouseButton.buttonId = MOUSE_BUTTON_LEFT;
			break;
		case 273: // BTN_RIGHT
			evt.mouseButton.buttonId = MOUSE_BUTTON_RIGHT;
			break;
		case 274: // BTN_MIDDLE
			evt.mouseButton.buttonId = MOUSE_BUTTON_MIDDLE;
			break;
		case 275: // BTN_SIDE
			evt.mouseButton.buttonId = MOUSE_BUTTON_X1;
			break;
		case 276: // BTN_EXTRA
			evt.mouseButton.buttonId = MOUSE_BUTTON_X2;
			break;
		default:
			return; // Unknown button, ignore
	}
	shared_mem_add_event(&s->sm, evt);
}

static void pointer_axis(void* data, struct wl_pointer* pointer,
	uint32_t time, uint32_t axis, wl_fixed_t value)
{
	WaylandState* s = data;
	int delta = wl_fixed_to_int(value);
	int deltaX = 0;
	int deltaY = 0;
	// Axis 0 = vertical, 1 = horizontal
	// Wayland uses opposite sign convention from X11
	if (axis == 0) {
		deltaY = -delta;
	} else {
		deltaX = -delta;
	}
	shared_mem_add_event(&s->sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_MOUSE_SCROLL,
		.mouseScroll = {
			.deltaX = deltaX,
			.deltaY = deltaY,
			.x = s->sm.mouseX,
			.y = s->sm.mouseY,
		}
	});
}

static void pointer_frame(void* data, struct wl_pointer* pointer) {
	// Optional: indicates end of pointer event sequence
}

static void pointer_axis_source(void* data, struct wl_pointer* pointer,
	uint32_t axis_source)
{
	// Optional: indicates axis source (wheel, finger, etc.)
}

static void pointer_axis_stop(void* data, struct wl_pointer* pointer,
	uint32_t time, uint32_t axis)
{
	// Optional: indicates axis movement stopped
}

static void pointer_axis_discrete(void* data, struct wl_pointer* pointer,
	uint32_t axis, int32_t discrete)
{
	// Optional: discrete scroll events
}

static void pointer_axis_value120(void* data, struct wl_pointer* pointer,
	uint32_t axis, int32_t value120)
{
	// Optional: high-resolution scroll events
}

static const struct wl_pointer_listener pointer_listener = {
	.enter = pointer_enter,
	.leave = pointer_leave,
	.motion = pointer_motion,
	.button = pointer_button,
	.axis = pointer_axis,
	.frame = pointer_frame,
	.axis_source = pointer_axis_source,
	.axis_stop = pointer_axis_stop,
	.axis_discrete = pointer_axis_discrete,
	.axis_value120 = pointer_axis_value120,
};

// Seat listener
static void seat_capabilities(void* data, struct wl_seat* seat,
	uint32_t capabilities)
{
	WaylandState* s = data;
	bool have_keyboard = capabilities & WL_SEAT_CAPABILITY_KEYBOARD;
	bool have_pointer = capabilities & WL_SEAT_CAPABILITY_POINTER;

	if (have_keyboard && !s->keyboard) {
		s->keyboard = wl_seat_get_keyboard(seat);
		wl_keyboard_add_listener(s->keyboard, &keyboard_listener, s);
	} else if (!have_keyboard && s->keyboard) {
		wl_keyboard_release(s->keyboard);
		s->keyboard = NULL;
	}

	if (have_pointer && !s->pointer) {
		s->pointer = wl_seat_get_pointer(seat);
		wl_pointer_add_listener(s->pointer, &pointer_listener, s);
	} else if (!have_pointer && s->pointer) {
		wl_pointer_release(s->pointer);
		s->pointer = NULL;
	}
}

static void seat_name(void* data, struct wl_seat* seat, const char* name) {
	// Optional: seat name
}

static const struct wl_seat_listener seat_listener = {
	.capabilities = seat_capabilities,
	.name = seat_name,
};

// Output listener
static void output_geometry(void* data, struct wl_output* output,
	int32_t x, int32_t y, int32_t physical_width, int32_t physical_height,
	int32_t subpixel, const char* make, const char* model, int32_t transform)
{
	WaylandState* s = data;
	s->output_width_mm = physical_width;
	s->output_height_mm = physical_height;
}

static void output_mode(void* data, struct wl_output* output,
	uint32_t flags, int32_t width, int32_t height, int32_t refresh)
{
	WaylandState* s = data;
	// WL_OUTPUT_MODE_CURRENT means this is the current mode
	if (flags & WL_OUTPUT_MODE_CURRENT) {
		s->output_width = width;
		s->output_height = height;
	}
}

static void output_done(void* data, struct wl_output* output) {
	// Output configuration complete
}

static void output_scale(void* data, struct wl_output* output, int32_t factor) {
	WaylandState* s = data;
	s->output_scale = factor;
}

static void output_name(void* data, struct wl_output* output,
	const char* name)
{
	// Output name
}

static void output_description(void* data, struct wl_output* output,
	const char* description)
{
	// Output description
}

static const struct wl_output_listener output_listener = {
	.geometry = output_geometry,
	.mode = output_mode,
	.done = output_done,
	.scale = output_scale,
	.name = output_name,
	.description = output_description,
};

// Decoration listener
static void toplevel_decoration_configure(void* data,
	struct zxdg_toplevel_decoration_v1* decoration, uint32_t mode)
{
	// Mode received from compositor - we requested server-side
}

static const struct zxdg_toplevel_decoration_v1_listener decoration_listener = {
	.configure = toplevel_decoration_configure,
};

// Registry listener
static void registry_global(void* data, struct wl_registry* registry,
	uint32_t name, const char* interface, uint32_t version)
{
	WaylandState* s = data;
	if (strcmp(interface, wl_compositor_interface.name) == 0) {
		s->compositor = wl_registry_bind(registry, name,
			&wl_compositor_interface, 4);
	} else if (strcmp(interface, wl_seat_interface.name) == 0) {
		s->seat = wl_registry_bind(registry, name, &wl_seat_interface, 5);
		wl_seat_add_listener(s->seat, &seat_listener, s);
	} else if (strcmp(interface, wl_shm_interface.name) == 0) {
		s->shm = wl_registry_bind(registry, name, &wl_shm_interface, 1);
	} else if (strcmp(interface, wl_output_interface.name) == 0) {
		s->output = wl_registry_bind(registry, name, &wl_output_interface, 4);
		wl_output_add_listener(s->output, &output_listener, s);
	} else if (strcmp(interface, xdg_wm_base_interface.name) == 0) {
		s->xdg_wm_base = wl_registry_bind(registry, name,
			&xdg_wm_base_interface, 1);
		xdg_wm_base_add_listener(s->xdg_wm_base, &xdg_wm_base_listener, s);
	} else if (strcmp(interface, zxdg_decoration_manager_v1_interface.name) == 0) {
		s->decoration_manager = wl_registry_bind(registry, name,
			&zxdg_decoration_manager_v1_interface, 1);
	}
}

static void registry_global_remove(void* data, struct wl_registry* registry,
	uint32_t name)
{
	// Object removed from registry
}

static const struct wl_registry_listener registry_listener = {
	.global = registry_global,
	.global_remove = registry_global_remove,
};

static void set_cursor(WaylandState* s, const char* cursor_name) {
	if (!s->cursor_theme) {
		s->cursor_theme = wl_cursor_theme_load(NULL, 24, s->shm);
		if (!s->cursor_theme) return;
	}
	if (!s->cursor_surface) {
		s->cursor_surface = wl_compositor_create_surface(s->compositor);
		if (!s->cursor_surface) return;
	}
	struct wl_cursor* cursor = wl_cursor_theme_get_cursor(s->cursor_theme,
		cursor_name);
	if (!cursor) {
		cursor = wl_cursor_theme_get_cursor(s->cursor_theme, "left_ptr");
	}
	if (cursor && cursor->image_count > 0) {
		s->current_cursor = cursor;
		struct wl_cursor_image* image = cursor->images[0];
		struct wl_buffer* buffer = wl_cursor_image_get_buffer(image);
		wl_surface_attach(s->cursor_surface, buffer, 0, 0);
		wl_surface_damage_buffer(s->cursor_surface, 0, 0,
			image->width, image->height);
		wl_surface_commit(s->cursor_surface);
	}
}

unsigned int get_toggle_key_state() {
	// Wayland doesn't provide direct access to LED state like X11
	// Return 0 for now - toggle key state must be tracked via key events
	if (global_state && global_state->xkb_state) {
		unsigned int state = 0;
		xkb_mod_index_t caps = xkb_keymap_mod_get_index(global_state->xkb_keymap,
			XKB_MOD_NAME_CAPS);
		xkb_mod_index_t num = xkb_keymap_mod_get_index(global_state->xkb_keymap,
			XKB_MOD_NAME_NUM);
		if (caps != XKB_MOD_INVALID &&
			xkb_state_mod_index_is_active(global_state->xkb_state, caps,
				XKB_STATE_MODS_LOCKED)) {
			state |= 1;
		}
		if (num != XKB_MOD_INVALID &&
			xkb_state_mod_index_is_active(global_state->xkb_state, num,
				XKB_STATE_MODS_LOCKED)) {
			state |= 2;
		}
		return state;
	}
	return 0;
}

void window_main(const char* windowTitle,
	int width, int height, int x, int y, uint64_t goWindow)
{
	WaylandState* s = calloc(1, sizeof(WaylandState));
	global_state = s;
	s->sm.goWindow = (void*)goWindow;
	s->sm.x = x;
	s->sm.y = y;
	s->sm.windowWidth = width;
	s->sm.windowHeight = height;
	s->output_scale = 1;

	s->display = wl_display_connect(NULL);
	if (!s->display) {
		printf("Failed to connect to Wayland display\n"); fflush(stdout);
		shared_mem_add_event(&s->sm, (WindowEvent) {
			.type = WINDOW_EVENT_TYPE_FATAL,
		});
		shared_mem_flush_events(&s->sm);
		free(s);
		return;
	}

	s->xkb_context = xkb_context_new(XKB_CONTEXT_NO_FLAGS);
	if (!s->xkb_context) {
		printf("Failed to create XKB context\n"); fflush(stdout);
		wl_display_disconnect(s->display);
		shared_mem_add_event(&s->sm, (WindowEvent) {
			.type = WINDOW_EVENT_TYPE_FATAL,
		});
		shared_mem_flush_events(&s->sm);
		free(s);
		return;
	}

	s->registry = wl_display_get_registry(s->display);
	wl_registry_add_listener(s->registry, &registry_listener, s);
	wl_display_roundtrip(s->display);

	if (!s->compositor || !s->xdg_wm_base) {
		printf("Failed to get required Wayland interfaces\n"); fflush(stdout);
		wl_display_disconnect(s->display);
		shared_mem_add_event(&s->sm, (WindowEvent) {
			.type = WINDOW_EVENT_TYPE_FATAL,
		});
		shared_mem_flush_events(&s->sm);
		free(s);
		return;
	}

	s->surface = wl_compositor_create_surface(s->compositor);
	s->xdg_surface = xdg_wm_base_get_xdg_surface(s->xdg_wm_base, s->surface);
	xdg_surface_add_listener(s->xdg_surface, &xdg_surface_listener, s);

	s->xdg_toplevel = xdg_surface_get_toplevel(s->xdg_surface);
	xdg_toplevel_add_listener(s->xdg_toplevel, &xdg_toplevel_listener, s);
	xdg_toplevel_set_title(s->xdg_toplevel, windowTitle);
	xdg_toplevel_set_app_id(s->xdg_toplevel, "kaiju");

	// Request server-side decorations (title bar, buttons) if available
	if (s->decoration_manager) {
		s->toplevel_decoration = zxdg_decoration_manager_v1_get_toplevel_decoration(
			s->decoration_manager, s->xdg_toplevel);
		zxdg_toplevel_decoration_v1_add_listener(s->toplevel_decoration,
			&decoration_listener, s);
		zxdg_toplevel_decoration_v1_set_mode(s->toplevel_decoration,
			ZXDG_TOPLEVEL_DECORATION_V1_MODE_SERVER_SIDE);
	}

	wl_surface_commit(s->surface);
	wl_display_roundtrip(s->display);

	// Wait for initial configure
	while (!s->configured && !s->closed) {
		wl_display_dispatch(s->display);
	}

	set_cursor(s, "left_ptr");

	shared_mem_add_event(&s->sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_SET_HANDLE,
		.setHandle = {
			.hwnd = s,
		}
	});
	shared_mem_flush_events(&s->sm);
}

void window_show(void* waylandState) {
	// Wayland surfaces are implicitly shown when committed
}

void window_poll_controller(void* waylandState) {
	// TODO: Implement controller support for Wayland
}

void window_poll(void* waylandState) {
	WaylandState* s = waylandState;
	if (s->closed) return;

	// Non-blocking dispatch
	wl_display_dispatch_pending(s->display);
	wl_display_flush(s->display);

	// Handle cursor locking
	if (s->sm.lockCursor.active) {
		// Wayland doesn't support cursor warping in the same way as X11
		// Pointer constraints protocol would be needed for proper implementation
	}

	shared_mem_flush_events(&s->sm);
}

void window_destroy(void* waylandState) {
	WaylandState* s = waylandState;
	if (s->cursor_surface) {
		wl_surface_destroy(s->cursor_surface);
	}
	if (s->cursor_theme) {
		wl_cursor_theme_destroy(s->cursor_theme);
	}
	if (s->xkb_state) {
		xkb_state_unref(s->xkb_state);
	}
	if (s->xkb_keymap) {
		xkb_keymap_unref(s->xkb_keymap);
	}
	if (s->xkb_context) {
		xkb_context_unref(s->xkb_context);
	}
	if (s->keyboard) {
		wl_keyboard_release(s->keyboard);
	}
	if (s->pointer) {
		wl_pointer_release(s->pointer);
	}
	if (s->toplevel_decoration) {
		zxdg_toplevel_decoration_v1_destroy(s->toplevel_decoration);
	}
	if (s->xdg_toplevel) {
		xdg_toplevel_destroy(s->xdg_toplevel);
	}
	if (s->xdg_surface) {
		xdg_surface_destroy(s->xdg_surface);
	}
	if (s->surface) {
		wl_surface_destroy(s->surface);
	}
	if (s->decoration_manager) {
		zxdg_decoration_manager_v1_destroy(s->decoration_manager);
	}
	if (s->xdg_wm_base) {
		xdg_wm_base_destroy(s->xdg_wm_base);
	}
	if (s->seat) {
		wl_seat_release(s->seat);
	}
	if (s->output) {
		wl_output_release(s->output);
	}
	if (s->shm) {
		wl_shm_destroy(s->shm);
	}
	if (s->compositor) {
		wl_compositor_destroy(s->compositor);
	}
	if (s->registry) {
		wl_registry_destroy(s->registry);
	}
	if (s->display) {
		wl_display_disconnect(s->display);
	}
	if (global_state == s) {
		global_state = NULL;
	}
	free(s);
}

void* display(void* waylandState) {
	return ((WaylandState*)waylandState)->display;
}

void* window(void* waylandState) {
	return ((WaylandState*)waylandState)->surface;
}

void window_focus(void* state) {
	// Wayland doesn't allow clients to focus themselves
	// This is a compositor-only operation
}

void window_position(void* state, int* x, int* y) {
	WaylandState* s = state;
	// Wayland doesn't expose window position to clients
	*x = s->sm.x;
	*y = s->sm.y;
}

void window_set_position(void* state, int x, int y) {
	// Wayland doesn't allow clients to set their own position
	// This is handled by the compositor
}

void window_set_size(void* state, int width, int height) {
	WaylandState* s = state;
	// Request a specific size - compositor may or may not honor it
	s->sm.windowWidth = width;
	s->sm.windowHeight = height;
	wl_surface_commit(s->surface);
}

int window_width_mm(void* state) {
	WaylandState* s = state;
	// Return the physical width, or estimate based on ~96 DPI if unknown
	if (s->output_width_mm > 0) {
		return s->output_width_mm;
	}
	// Fallback: assume 96 DPI = ~3.78 pixels per mm
	return s->output_width > 0 ? s->output_width * 254 / 960 : 508;
}

int window_height_mm(void* state) {
	WaylandState* s = state;
	// Return the physical height, or estimate based on ~96 DPI if unknown
	if (s->output_height_mm > 0) {
		return s->output_height_mm;
	}
	// Fallback: assume 96 DPI = ~3.78 pixels per mm
	return s->output_height > 0 ? s->output_height * 254 / 960 : 285;
}

void window_cursor_standard(void* state) {
	set_cursor(state, "left_ptr");
}

void window_cursor_ibeam(void* state) {
	set_cursor(state, "xterm");
}

void window_cursor_size_all(void* state) {
	set_cursor(state, "all-scroll");
}

void window_cursor_size_ns(void* state) {
	set_cursor(state, "ns-resize");
}

void window_cursor_size_we(void* state) {
	set_cursor(state, "ew-resize");
}

void window_show_cursor(void* state) {
	WaylandState* s = state;
	if (s->current_cursor) {
		set_cursor(s, "left_ptr");
	}
}

void window_hide_cursor(void* state) {
	WaylandState* s = state;
	if (s->pointer && s->cursor_surface) {
		wl_surface_attach(s->cursor_surface, NULL, 0, 0);
		wl_surface_commit(s->cursor_surface);
		// Set cursor to empty surface
		// Note: This requires the pointer to enter the surface again
		// to properly hide. A better solution would use zwp_pointer_constraints
	}
}

float window_dpi(void* state) {
	WaylandState* s = state;
	// Wayland uses scale factor instead of DPI directly
	// A scale of 2 typically means ~192 DPI (2x 96)
	return (float)s->output_scale * 96.0f / 25.4f;
}

void window_set_title(void* state, const char* windowTitle) {
	WaylandState* s = state;
	xdg_toplevel_set_title(s->xdg_toplevel, windowTitle);
}

void window_set_full_screen(void* state) {
	WaylandState* s = state;
	// Save current state for restoration
	s->sm.savedState.rect.left = s->sm.x;
	s->sm.savedState.rect.top = s->sm.y;
	s->sm.savedState.rect.right = s->sm.x + s->sm.windowWidth;
	s->sm.savedState.rect.bottom = s->sm.y + s->sm.windowHeight;
	xdg_toplevel_set_fullscreen(s->xdg_toplevel, NULL);
}

void window_set_windowed(void* state, int width, int height) {
	WaylandState* s = state;
	xdg_toplevel_unset_fullscreen(s->xdg_toplevel);
	// Request preferred size
	s->sm.windowWidth = width;
	s->sm.windowHeight = height;
	wl_surface_commit(s->surface);
}

void window_lock_cursor(void* state, int x, int y) {
	WaylandState* s = state;
	s->sm.lockCursor.x = x;
	s->sm.lockCursor.y = y;
	s->sm.lockCursor.active = true;
	// Proper cursor locking requires zwp_pointer_constraints protocol
}

void window_unlock_cursor(void* state) {
	WaylandState* s = state;
	s->sm.lockCursor.active = false;
}

#endif

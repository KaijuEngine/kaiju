/******************************************************************************/
/* android.c                                                                  */
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

#if defined(__android__)


#include <jni.h>
#include <memory.h>
#include <stdlib.h>
#include <android/log.h>
#include <android/sensor.h>
#include <android/window.h>
#include <android/choreographer.h>
#include <android_native_app_glue.h>

#include "android.h"
#include "shared_mem.h"

#define log_verbose(...) (void)__android_log_print(ANDROID_LOG_VERBOSE, "KaijuEngineLogC" __VA_OPT__(,) __VA_ARGS__)
#define log_info(...) (void)__android_log_print(ANDROID_LOG_INFO, "KaijuEngineLogC" __VA_OPT__(,) __VA_ARGS__)
#define log_warn(...) (void)__android_log_print(ANDROID_LOG_WARN, "KaijuEngineLogC" __VA_OPT__(,) __VA_ARGS__)
#define log_err(...) (void)__android_log_print(ANDROID_LOG_ERROR, "KaijuEngineLogC" __VA_OPT__(,) __VA_ARGS__)
#define debug(...) (void)__android_log_print(ANDROID_LOG_INFO, "KaijuEngineLogC" __VA_OPT__(,) __VA_ARGS__)

#define ANDROID_MOUSE_MOVE			7
#define ANDROID_MOUSE_UP			9
#define ANDROID_MOUSE_DOWN			10
#define ANDROID_MOUSE_HELD_START	11
#define ANDROID_MOUSE_HELD_END		12
#define ANDROID_MOUSE_EVENT_BEGIN	ANDROID_MOUSE_MOVE

static inline SharedMem* local_shared_memory(struct android_app* state) { return state->userData; }

static inline bool local_wait_for_window_init(struct android_app* state) {
	struct android_poll_source* source;
	while (state->userData == NULL) {
		while (ALooper_pollOnce(0, NULL, NULL, (void**)&source) >= 0) {
			// Process this event.
			if (source != NULL) {
				source->process(state, source);
			}
			// Check if we are exiting.
			if (state->destroyRequested != 0) {
				return false;
			}
		}
	}
	return true;
}

static void local_set_touch(AInputEvent* event, SharedMem* sm) {
	//for (int i = 0; i < count; i++) {
	//	touch_set_pressure(touch, i, AMotionEvent_getPressure(event, i));
	//	touch_set_position(touch, i, AMotionEvent_getX(event, i),
	//		AMotionEvent_getY(event, i),
	//		(float)win->width, (float)win->height);
	//	//AMotionEvent_getRawX(event, i),
	//	//AMotionEvent_getRawY(event, i));
	//}
	int32_t action = AMotionEvent_getAction(event);
	int32_t actionType = action & AMOTION_EVENT_ACTION_MASK;
	//int32_t pointerIndex = (action & AMOTION_EVENT_ACTION_POINTER_INDEX_MASK) >> AMOTION_EVENT_ACTION_POINTER_INDEX_SHIFT;
	if (actionType == AMOTION_EVENT_ACTION_CANCEL) {
		shared_mem_add_event(sm, (WindowEvent) {
			.type = WINDOW_EVENT_TYPE_TOUCH_STATE,
			.touchState = {
				.index = 0,
				.actionState = TOUCH_ACTION_STATE_TYPE_CANCEL,
			}
		});
	} else {
		float w = (float)sm->windowWidth;
		float h = (float)sm->windowHeight;
		int apc = (int)AMotionEvent_getPointerCount(event);
		int count = apc < MAX_TOUCH_POINTERS_AVAILABLE ? apc : MAX_TOUCH_POINTERS_AVAILABLE;
		if (actionType == AMOTION_EVENT_ACTION_UP && count == 1) {
			float x = AMotionEvent_getX(event, 0);
			float y = AMotionEvent_getY(event, 0);
			shared_mem_add_event(sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_TOUCH_STATE,
				.touchState = {
					.index = 0,
					.x = x,
					.y = y,
					.pressure = 0,
					.actionState = TOUCH_ACTION_STATE_TYPE_UP,
				}
			});
		} else {
			for (int i = 0; i < count; i++) {
				shared_mem_add_event(sm, (WindowEvent) {
					.type = WINDOW_EVENT_TYPE_TOUCH_STATE,
					.touchState = {
						.index = i,
						.x = AMotionEvent_getX(event, i),
						.y = AMotionEvent_getY(event, i),
						.pressure = AMotionEvent_getPressure(event, i),
						.actionState = TOUCH_ACTION_STATE_TYPE_MOVE,
					}
				});
			}
		}
	}
}

static void local_set_stylus(AInputEvent* event, SharedMem* sm) {
	StylusActionStateType s = STYLUS_ACTION_STATE_TYPE_NONE;
	switch (AMotionEvent_getAction(event)) {
		// case AMOTION_EVENT_ACTION_IDLE:
		// 	s = STYLUS_ACTION_STATE_TYPE_NONE;
		// 	break;
		case AMOTION_EVENT_ACTION_HOVER_ENTER:
			s = STYLUS_ACTION_STATE_TYPE_HOVER_ENTER;
			break;
		case AMOTION_EVENT_ACTION_HOVER_MOVE:
			s = STYLUS_ACTION_STATE_TYPE_HOVER_MOVE;
			break;
		case AMOTION_EVENT_ACTION_HOVER_EXIT:
			s = STYLUS_ACTION_STATE_TYPE_HOVER_EXIT;
			break;
		case AMOTION_EVENT_ACTION_DOWN:
			s = STYLUS_ACTION_STATE_TYPE_DOWN;
			break;
		case AMOTION_EVENT_ACTION_MOVE:
			s = STYLUS_ACTION_STATE_TYPE_MOVE;
			break;
		case AMOTION_EVENT_ACTION_UP:
			s = STYLUS_ACTION_STATE_TYPE_UP;
			break;
	}
	shared_mem_add_event(sm, (WindowEvent) {
		.type = WINDOW_EVENT_TYPE_STYLUS_STATE,
		.stylusState = {
			.x = AMotionEvent_getX(event, 0),
			.y = AMotionEvent_getY(event, 0),
			.pressure = AMotionEvent_getPressure(event, 0),
			.distance = AMotionEvent_getAxisValue(event, AMOTION_EVENT_AXIS_DISTANCE, 0),
			.actionState = s,
		}
	});
}

static void local_set_mouse(AInputEvent* event, SharedMem* sm) {
	double x = AMotionEvent_getRawX(event, 0);
	double y = AMotionEvent_getRawY(event, 0);
	int32_t a = AMotionEvent_getAction(event);
	switch (a) {
		case ANDROID_MOUSE_DOWN:
			shared_mem_add_event(sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = {
					.x = (int32_t)x,
					.y = (int32_t)y,
					.buttonId = MOUSE_BUTTON_LEFT,
					.action = WINDOW_EVENT_BUTTON_TYPE_DOWN,
				}
			});
			break;
		case ANDROID_MOUSE_HELD_START:
		case ANDROID_MOUSE_HELD_END:
		case AMOTION_EVENT_ACTION_MOVE:
		case ANDROID_MOUSE_MOVE:
			// Repeat is managed internally, just update the position
			shared_mem_add_event(sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_MOUSE_MOVE,
				.mouseMove = {
					.x = (int32_t)x,
					.y = (int32_t)y,
				}
			});
			break;
		case ANDROID_MOUSE_UP:
			shared_mem_add_event(sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_MOUSE_BUTTON,
				.mouseButton = {
					.x = (int32_t)x,
					.y = (int32_t)y,
					.buttonId = MOUSE_BUTTON_LEFT,
					.action = WINDOW_EVENT_BUTTON_TYPE_UP,
				}
			});
			break;
	}
}

static int32_t local_input_handle(struct android_app* app, AInputEvent* event) {
	SharedMem* sm = local_shared_memory(app);
	int32_t evtType = AInputEvent_getType(event);
	if (evtType == AINPUT_EVENT_TYPE_MOTION) {
		int32_t action = AMotionEvent_getAction(event);
		int idx = (action & AMOTION_EVENT_ACTION_POINTER_INDEX_MASK) >> AMOTION_EVENT_ACTION_POINTER_INDEX_SHIFT;
		int32_t toolType = AMotionEvent_getToolType(event, idx);
		switch (toolType) {
			case AMOTION_EVENT_TOOL_TYPE_FINGER:
			{
				//if (action >= ANDROID_MOUSE_EVENT_BEGIN
				//	|| (eng->host->window->mouse.button_held(MouseButton::LEFT)
				//		&& action != AMOTION_EVENT_ACTION_UP)) {
				//	goto android_input_mouse_tool_type_switch;
				//} else {
					local_set_touch(event, sm);
					return 1;
				//}
			}
			case AMOTION_EVENT_TOOL_TYPE_STYLUS:
			{
				local_set_stylus(event, sm);
				// TODO:  Do any stylus event processing
				return 1;
			}
			case AMOTION_EVENT_TOOL_TYPE_MOUSE:
			android_input_mouse_tool_type_switch:
			{
				local_set_mouse(event, sm);
				// TODO:  Do any mouse event processing
				return 1;
			}
		}
	} else if (evtType == AINPUT_EVENT_TYPE_KEY) {
		//int32_t action = AKeyEvent_getAction(event);
		//int32_t keyCode = AKeyEvent_getKeyCode(event);
		//int32_t flags = AKeyEvent_getFlags(event);
		//int32_t scanCode = AKeyEvent_getScanCode(event);
		//KeyboardKey key = window_handle_to_keyboard_key(keyCode);
		//log_verbose("Key event: %d, %d, %d, %d", action, keyCode, flags, scanCode);
		//if (action == AKEY_EVENT_ACTION_DOWN) {
		//	// TODO:  Only do this if using a software keyboard
		//	keyboard_set_key_down_up(&eng->host->window->keyboard, key);
		//}
		////valk_keyboard_set_key_down(host->platform->keyboard, key);
		////else if (action == AKEY_EVENT_ACTION_UP)
		////valk_keyboard_set_key_up(host->platform->keyboard, key);
		return 1;
	}
	return 0;
}

static void local_handle_cmd(struct android_app* app, int32_t cmd) {
	SharedMem* sm = local_shared_memory(app);
	switch (cmd) {
		case APP_CMD_SAVE_STATE:
			log_info("Application save state requested");
			// TODO:  Save the state of the game
			//app->savedState
			break;
		case APP_CMD_INIT_WINDOW:
		{
			log_info("Window initialize requested");
			if (app->window != NULL) {
				log_info("Application window found, checking for shared data");
				if (app->userData == NULL) {
					log_info("Shared data was not found, creating it");
					SharedMem* sm = calloc(1, sizeof(SharedMem));
					app->userData = sm;
					if (app->savedState != NULL) {
						// TODO:  Need to do this later
						log_info("Saved state was found, loading state");
					}
				}
			}
			break;
		}
		case APP_CMD_TERM_WINDOW:
			log_info("Window terminate requested");
			// The window is being hidden or closed, clean it up.
			//shared_mem_add_event(sm, (WindowEvent) {
			//	.type = WINDOW_EVENT_TYPE_ACTIVITY,
			//	.windowActivity = { WINDOW_EVENT_ACTIVITY_TYPE_CLOSE }
			//});
			//shared_mem_flush_events(sm);
			break;
		case APP_CMD_PAUSE:
		case APP_CMD_STOP:
			log_info("Application pause/stop requested");
			// TODO:  Probably should be managed similar to minimized on desktop
			// however, it should not continue to update/animate
			break;
		case APP_CMD_GAINED_FOCUS:
		{
			log_info("Window focus gain requested");
			// TODO:  Continue update/animate
			const ASensor* accelerometer = sm->accelerometer;
			ASensorEventQueue* q = sm->sensorQueue;
			if (accelerometer != NULL) {
				ASensorEventQueue_enableSensor(q, accelerometer);
				// We'd like to get 60 events per second (in us).
				ASensorEventQueue_setEventRate(q, accelerometer, (1000L / 60) * 1000);
			}
			ALooper_wake(ALooper_forThread());
			// TODO:  Reconstruct the render system
			break;
		}
		case APP_CMD_WINDOW_RESIZED:
		{
			log_info("Window resize requested");
			int w = ANativeWindow_getWidth(app->window);
			int h = ANativeWindow_getHeight(app->window);
			if (sm->windowWidth != w && sm->windowHeight != h) {
				log_info("Window is resizing from <%d, %d> to <%d, %d>",
					sm->windowWidth, sm->windowHeight, w, h);
				sm->windowWidth = w;
				sm->windowHeight = h;
			}
			shared_mem_add_event(sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_RESIZE,
				.windowResize = {
					.width = w,
					.height = h,
				}
			});
			shared_mem_flush_events(sm);
			break;
		}
		case APP_CMD_LOST_FOCUS:
		{
			log_info("Window lost focus requested");
			// When our app loses focus, we stop monitoring the accelerometer.
			// This is to avoid consuming battery while not being used.
			const ASensor* accelerometer = sm->accelerometer;
			ASensorEventQueue* q = sm->sensorQueue;
			if (accelerometer != NULL) {
				ASensorEventQueue_disableSensor(q, accelerometer);
			}
			// TODO:  Also likely will need to disable animation
			break;
		}
		case APP_CMD_CONFIG_CHANGED:
			log_info("Application config change requested");
			// TODO:  Resolution change is probable
			break;
		case APP_CMD_DESTROY:
		{
			log_info("Application destroy requested");
			shared_mem_add_event(sm, (WindowEvent) {
				.type = WINDOW_EVENT_TYPE_ACTIVITY,
				.windowActivity = { WINDOW_EVENT_ACTIVITY_TYPE_CLOSE }
			});
			shared_mem_flush_events(sm);
			break;
		}
	}
}

static int local_sensor_callback(int fd, int events, void* data) {
	SharedMem* sm = (SharedMem*)data;
	// Poll
	if (sm->accelerometer != NULL) {
		ASensorEvent event;
		while (ASensorEventQueue_getEvents(sm->sensorQueue, &event, 1) > 0) {
			//log_info("accelerometer: x=%f y=%f z=%f",
			//	event.acceleration.x, event.acceleration.y,
			//	event.acceleration.z);
		}
	}
    return 1;
}

void window_main(void* androidApp, uint64_t goWindow) {
	struct android_app* state = androidApp;
	log_info("Entering native application");
	state->userData = NULL;
	state->onAppCmd = local_handle_cmd;
	if (state->savedState != NULL) {
		log_info("Engine loading from save state");
		// TODO:  Load from saved state
	}
	// Read all pending events.
	if (!local_wait_for_window_init(state)) {
		log_info("Kill requested before construction!");
		return;
	}
	log_info("Setting up the input event handler");
	state->onInputEvent = local_input_handle;
	SharedMem* sm = local_shared_memory(state);
	sm->goWindow = (void*)goWindow;
	log_info("Getting the sensor manager");
	sm->sensorManager = ASensorManager_getInstance();
	log_info("Getting the accelerometer");
	sm->accelerometer = ASensorManager_getDefaultSensor(sm->sensorManager, ASENSOR_TYPE_ACCELEROMETER);
	log_info("Getting the sensor queue");
	sm->sensorQueue = ASensorManager_createEventQueue(sm->sensorManager,
		state->looper, LOOPER_ID_USER, local_sensor_callback, sm);
	//ANativeActivity_setWindowFlags(state->activity, AWINDOW_FLAG_KEEP_SCREEN_ON, 0);
	//ALooper_acquire(ALooper_forThread());
}

void* pull_android_window(void* androidApp) {
	struct android_app* state = androidApp;
	log_info("Returning the android window: %p", state->window);
	return state->window;
}

void window_poll(void* androidApp) {
	struct android_app* state = androidApp;
	struct android_poll_source* source;
	int result;
	while ((result = ALooper_pollOnce(0, NULL, NULL, (void**)&source)) != ALOOPER_POLL_TIMEOUT) {
		if (result == ALOOPER_POLL_ERROR) {
			log_err("ALooper_pollOnce returned an error");
			break;
		}
		if (source != NULL) {
			source->process(state, source);
		}
		source = NULL;
	}
	shared_mem_flush_events(local_shared_memory(state));
}

void window_size_mm(void* androidApp, int* widthMM, int* heightMM) {
	struct android_app* state = androidApp;
	JNIEnv* env = NULL;
	const ANativeActivity* activity = state->activity;
	(*activity->vm)->AttachCurrentThread(activity->vm, &env, NULL);
	jclass activityClass = (*env)->GetObjectClass(env, state->activity->clazz);
	jmethodID width_mm = (*env)->GetMethodID(env, activityClass, "width_mm", "()F");
	jmethodID height_mm = (*env)->GetMethodID(env, activityClass, "height_mm", "()F");
	jfloat wmm = (*env)->CallFloatMethod(env, state->activity->clazz, width_mm);
	jfloat hmm = (*env)->CallFloatMethod(env, state->activity->clazz, height_mm);
	*widthMM = (int)wmm;
	*heightMM = (int)hmm;
	(*env)->DeleteLocalRef(env, activityClass);
	(*activity->vm)->DetachCurrentThread(activity->vm);
}

void window_open_website(void* androidApp, const char* url) {
	struct android_app* state = androidApp;
	JNIEnv* env = NULL;
	const ANativeActivity* activity = state->activity;
	(*activity->vm)->AttachCurrentThread(activity->vm, &env, NULL);
	jclass activityClass = (*env)->GetObjectClass(env, state->activity->clazz);
	jobject urlStr = (*env)->NewStringUTF(env, (jstring)url);
	jmethodID open_web_url = (*env)->GetMethodID(env, activityClass, "open_web_url", "(Ljava/lang/String;)V");
	(*env)->CallVoidMethod(env, state->activity->clazz, open_web_url, urlStr);
	(*env)->DeleteLocalRef(env, urlStr);
	(*env)->DeleteLocalRef(env, activityClass);
	(*activity->vm)->DetachCurrentThread(activity->vm);
}

bool window_asset_exists(void* androidApp, const char* path) {
	struct android_app* state = androidApp;
	AAsset* f = AAssetManager_open(state->activity->assetManager, path, AASSET_MODE_BUFFER);
	bool exists = f != NULL;
	AAsset_close(f);
	return exists;
}

int64_t window_asset_length(void* androidApp, const char* path) {
	struct android_app* state = androidApp;
	AAsset* f = AAssetManager_open(state->activity->assetManager, path, AASSET_MODE_BUFFER);
	if (f == NULL) {
		return 0;
	}
	off64_t len = AAsset_getLength64(f);
	AAsset_close(f);
	return (int64_t)len;
}

int64_t window_asset_read(void* androidApp, const char* path, void* outData) {
	struct android_app* state = androidApp;
	AAsset* f = AAssetManager_open(state->activity->assetManager, path, AASSET_MODE_BUFFER);
	if (f == NULL) {
		return 0;
	}
	off64_t len = AAsset_getLength64(f);
	const void* buff = AAsset_getBuffer(f);
	memcpy(outData, buff, len);
	AAsset_close(f);
	return (int64_t)len;
}

#endif

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
#include <stdlib.h>
#include <android/window.h>
#include <android/choreographer.h>
#include <android_native_app_glue.h>

#include "log.h"

static inline bool local_wait_for_engine(struct android_app* state) {
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

static int32_t local_input_handle(struct android_app* app, AInputEvent* event) {
	return 0;
}

static void local_handle_cmd(struct android_app* app, int32_t cmd) {
	switch (cmd) {
		case APP_CMD_SAVE_STATE:
			break;
		case APP_CMD_INIT_WINDOW:
			break;
		case APP_CMD_TERM_WINDOW:
			break;
		case APP_CMD_PAUSE:
		case APP_CMD_STOP:
			break;
		case APP_CMD_GAINED_FOCUS:
			break;
		case APP_CMD_WINDOW_RESIZED:
			break;
		case APP_CMD_LOST_FOCUS:
			break;
		case APP_CMD_CONFIG_CHANGED:
			break;
		case APP_CMD_DESTROY:
			break;
	}
}

void window_main(void* androidApp) {
	struct android_app* state = androidApp;
	log_info("Entering native application");
	state->userData = NULL;
	state->onAppCmd = local_handle_cmd;
	if (state->savedState != NULL) {
		log_info("Engine loading from save state");
		// TODO:  Load from saved state
	}
	// Read all pending events.
    if (!local_wait_for_engine(state)) {
	    log_info("Kill requested before host constructed!");
        return;
    }
	state->onInputEvent = local_input_handle;
	//ANativeActivity_setWindowFlags(state->activity, AWINDOW_FLAG_KEEP_SCREEN_ON, 0);
	//Engine* eng = (Engine*)state->userData;
	//engine_frame(eng); // Catch frame up to just before rendering
	log_info("Presenting splash screen");
	//ALooper_acquire(ALooper_forThread());
    struct android_poll_source* source;
	log_info("Beginning device event loop");
    while (!state->destroyRequested) {
        int result = ALooper_pollOnce(-1, NULL, NULL, (void**)&source);
        if (result == ALOOPER_POLL_ERROR) {
            log_err("ALooper_pollOnce returned an error");
        }
        if (source != NULL) {
            source->process(state, source);
        }
    }
}

#endif

/******************************************************************************/
/* log.h                                                                      */
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

#ifndef STDLIB_LOG_H
#define STDLIB_LOG_H

#include <stdio.h>
#include <errno.h>

extern FILE* gLogFileHandle;

extern struct GLocCallback {
	void* state;
	void(*call)(char mode, const char* message, void* state);
} gLogCallback;

#if defined(_DEBUG) && defined(__linux__)
#include <signal.h>
#if !__android__
#include <bits/signum-generic.h>
#endif
#endif


#if _DEBUG
static inline void log_break() {
#if defined(__linux__) || defined(STEAM_DECK)
	//__asm__ volatile("int $0x03");
	raise(SIGTRAP);
#elif __windows__ || defined(XBOX)
	/*asm("nop");*/ __debugbreak();
#elif __android__
	raise(SIGTRAP);
#elif defined(MACOS)
	//__asm__("int $3");
	__builtin_trap();
#elif __ios__
	__builtin_trap();
#elif defined(SWITCH)
	__builtin_debugtrap();
#elif defined(SONY)
	SCE_BREAK();
#endif
}
#else
#define log_break() do{}while(0)
#endif

#if _DEBUG || ENABLE_LOGGING
	#if __android__
		#include <android/log.h>
		#define log_err(...) (void)__android_log_print(ANDROID_LOG_ERROR, "NativeActivity" __VA_OPT__(,) __VA_ARGS__)
		#define log_warn(...) (void)__android_log_print(ANDROID_LOG_WARN, "NativeActivity" __VA_OPT__(,) __VA_ARGS__)
		#define log_info(...) (void)__android_log_print(ANDROID_LOG_INFO, "NativeActivity" __VA_OPT__(,) __VA_ARGS__)
		#define log_verbose(...) (void)__android_log_print(ANDROID_LOG_VERBOSE, "NativeActivity" __VA_OPT__(,) __VA_ARGS__)
		#define debug(...) (void)__android_log_print(ANDROID_LOG_INFO, "NativeActivity" __VA_OPT__(,) __VA_ARGS__)
		#define log_ensure(A, M, ...) do { if (!(A)) { log_err(M __VA_OPT__(,) __VA_ARGS__); log_break(); } } while(0)
	#else
		#define log_cpp_write_file(M, ...) fprintf(gLogFileHandle, "[E] (%s:%d:) " M "\n", __FILE__, __LINE__ __VA_OPT__(,) __VA_ARGS__), fflush(gLogFileHandle)
		#define log_err(M, ...) log_write(stderr, 'E', "[E] (%s:%d:) " M "\n", __FILE__, __LINE__ __VA_OPT__(,) __VA_ARGS__), fflush(stderr), log_cpp_write_file(M, __VA_ARGS__)
		#define log_warn(M, ...) log_write(stderr, 'W', "[W] (%s:%d:) " M "\n", __FILE__, __LINE__ __VA_OPT__(,) __VA_ARGS__), fflush(stderr), log_cpp_write_file(M, __VA_ARGS__)
		#define log_info(M, ...) log_write(stdout, 'I', "[I] (%s:%d:) " M "\n", __FILE__, __LINE__ __VA_OPT__(,) __VA_ARGS__), fflush(stdout), log_cpp_write_file(M, __VA_ARGS__)
		#define log_verbose(M, ...) log_write(stdout, 'V', "[V] (%s:%d:) " M "\n", __FILE__, __LINE__ __VA_OPT__(,) __VA_ARGS__), fflush(stdout), log_cpp_write_file(M, __VA_ARGS__)
		#define debug(M, ...) log_write(stdout, 'D', "[D] %s:%d:\n\t" M "\n", __FILE__, __LINE__ __VA_OPT__(,) __VA_ARGS__), fflush(stdout), log_cpp_write_file(M, __VA_ARGS__)
		#define log_ensure(A, M, ...) do { if (!(A)) { log_err(M __VA_OPT__(,) __VA_ARGS__); log_break(); } } while(0)
	#endif
#else
	#define log_err(M, ...) do{}while(0)
	#define log_warn(M, ...) do{}while(0)
	#define log_info(M, ...) do{}while(0)
	#define log_verbose(M, ...) do{}while(0)
	#define debug(M, ...) do{}while(0)
	#define log_ensure(C, M, ...) do{}while(0)
#endif

#if _DEBUG
#define check(A, M, ...) do { if(!(A)) { log_err(M __VA_OPT__(,) __VA_ARGS__); errno=0; goto error; } } while(0)
#define check_mem(A) check((A), "Out of memory.")
#define sentinel(M, ...)  do { log_err(M __VA_OPT__(,) __VA_ARGS__); errno=0; goto error; } while(0)
#define check_debug(A, M, ...) do { if(!(A)) { debug(M __VA_OPT__(,) __VA_ARGS__); errno=0; goto error; } } while(0)
#define ensure(A) do { if (!(A)) { log_break(); } } while(0)
#else
#define check(A, M, ...) do{}while(0)
#define sentinel(M, ...) do{}while(0)
#define check_mem(A) do{}while(0)
#define check_debug(A, M, ...) do{}while(0)
#define ensure(A) do{}while(0)
#endif

#endif

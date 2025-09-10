/******************************************************************************/
/* soloud.c.go                                                                */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package audio

/*
#cgo windows LDFLAGS: -L../../../libs -lsoloud_win32 -lstdc++ -lwinmm -lole32 -luuid
#cgo linux LDFLAGS: -L../../../libs -lsoloud_nix -lasound -lstdc++
#include <stdlib.h>
#include "soloud_c.h"
*/
import "C"
import (
	"log/slog"
	"unsafe"
)

type SoloudHandle = *C.Soloud
type SoloudWav = *C.Wav

func errToString(soloud SoloudHandle, errCode int) string {
	return C.GoString(C.Soloud_getErrorString(soloud, C.int(errCode)))
}

func create() SoloudHandle {
	return C.Soloud_create()
}

func initialize(soloud SoloudHandle) int {
	return int(C.Soloud_init(soloud))
}

func deinitialize(soloud SoloudHandle) {
	C.Soloud_deinit(soloud)
}

func destroy(soloud SoloudHandle) {
	C.Soloud_destroy(soloud)
}

func wavCreate() SoloudWav {
	return C.Wav_create()
}

func wavLoad(path string, wav SoloudWav) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	res := int(C.Wav_load(wav, cPath))
	if res != 0 {
		slog.Error("there was an error loading the sound file", "file", path, "code", res)
	}
}

func wavDestroy(wav SoloudWav) {
	C.Wav_destroy(wav)
}

func wavSetVolume(wav SoloudWav, volume float32) {
	C.Wav_setVolume(wav, C.float(volume))
}

func setVolume(soloud SoloudHandle, handle uint32, volume float32) {
	C.Soloud_setVolume(soloud, C.uint(handle), C.float(volume))
}

func play(soloud SoloudHandle, wav SoloudWav) uint32 {
	return uint32(C.Soloud_play(soloud, (*C.AudioSource)(wav)))
}

func stopAudioSource(soloud SoloudHandle, wav SoloudWav) {
	C.Soloud_stopAudioSource(soloud, (*C.AudioSource)(wav))
}

func setLooping(soloud SoloudHandle, handle uint32, loop bool) {
	if loop {
		C.Soloud_setLooping(soloud, C.uint(handle), C.int(1))
	} else {
		C.Soloud_setLooping(soloud, C.uint(handle), C.int(0))
	}
}

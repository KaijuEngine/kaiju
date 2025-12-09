/******************************************************************************/
/* soloud.c.go                                                                */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
#cgo windows LDFLAGS: -L../../libs -lsoloud_win32 -lstdc++ -lwinmm -lole32 -luuid
#cgo android LDFLAGS: -L../../libs -lsoloud_android
#cgo linux,!android LDFLAGS: -L../../libs -lsoloud_nix -lasound -lstdc++
#cgo darwin,!ios LDFLAGS: -L../../libs -lsoloud_darwin -lstdc++ -framework AudioToolbox -framework CoreAudio
#include <stdlib.h>
#include "soloud_c.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type SoloudHandle = *C.Soloud
type SoloudWav = *C.Wav
type VoiceHandle = uint32

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

func wavLoadMem(wav SoloudWav, data []byte) error {
	res := int(C.Wav_loadMemEx(wav, (*C.uchar)(unsafe.Pointer(&data[0])), C.uint(len(data)), C.int(1), C.int(0)))
	if res != 0 {
		return fmt.Errorf("there was an error loading the audio memory: %d", res)
	}
	return nil
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

func clipLength(wav SoloudWav) float64 {
	return float64(C.Wav_getLength(wav))
}

func play(soloud SoloudHandle, wav SoloudWav) VoiceHandle {
	return VoiceHandle(C.Soloud_play(soloud, (*C.AudioSource)(wav)))
}

func stopAudioSource(soloud SoloudHandle, wav SoloudWav) {
	C.Soloud_stopAudioSource(soloud, (*C.AudioSource)(wav))
}

func isValidVoiceHandle(soloud SoloudHandle, handle VoiceHandle) bool {
	return int(C.Soloud_isValidVoiceHandle(soloud, (C.uint)(handle))) != 0
}

func seek(soloud SoloudHandle, handle VoiceHandle, seconds float64) int {
	return int(C.Soloud_seek(soloud, (C.uint)(handle), C.double(seconds)))
}

func setLooping(soloud SoloudHandle, handle uint32, loop bool) {
	if loop {
		C.Soloud_setLooping(soloud, C.uint(handle), C.int(1))
	} else {
		C.Soloud_setLooping(soloud, C.uint(handle), C.int(0))
	}
}

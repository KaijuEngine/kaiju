/******************************************************************************/
/* soloud.c.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package audio

/*
#cgo windows LDFLAGS: -L../../libs -lsoloud_win32 -lstdc++ -lwinmm -lole32 -luuid
#cgo android LDFLAGS: -L../../libs -lsoloud_android
#cgo linux,!android LDFLAGS: -L../../libs -lsoloud_nix -lasound -lstdc++
#cgo darwin,!ios,arm64 LDFLAGS: -L../../libs -lsoloud_darwin_arm64 -lstdc++ -framework AudioToolbox -framework CoreAudio
#cgo darwin,!ios,amd64 LDFLAGS: -L../../libs -lsoloud_darwin_amd64 -lstdc++ -framework AudioToolbox -framework CoreAudio
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

const InvalidVoiceHandle = VoiceHandle(0)

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

func stopAudio(soloud SoloudHandle, handle VoiceHandle) {
	C.Soloud_stop(soloud, (C.uint)(handle))
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

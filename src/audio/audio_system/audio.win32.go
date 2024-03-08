//go:build windows

/******************************************************************************/
/* audio.win32.go                                                             */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package audio_system

/*
#cgo LDFLAGS: -lole32
#cgo noescape macro_IMMDeviceEnumerator_GetDefaultAudioEndpoint
#cgo noescape macro_IMMDevice_Activate
#cgo noescape macro_IAudioClient_GetMixFormat
#cgo noescape macro_IAudioClient_Initialize
#cgo noescape macro_IAudioClient_GetBufferSize
#cgo noescape macro_IAudioClient_GetService
#cgo noescape macro_IMMDeviceEnumerator_Release
#cgo noescape macro_IMMDevice_Release
#cgo noescape macro_IAudioClient_Release
#cgo noescape macro_IAudioClient_Start
#cgo noescape macro_IAudioClient_Stop
#cgo noescape macro_IAudioClient_GetCurrentPadding
#cgo noescape macro_IAudioRenderClient_GetBuffer
#cgo noescape macro_IAudioRenderClient_ReleaseBuffer

#if defined(_WIN32) || defined(_WIN64)
#define WIN32_LEAN_AND_MEAN
#define WINDOWS_NETWORKING
#define _CRT_SECURE_NO_DEPRECATE
#define COBJMACROS
#include <windows.h>
#include <ole2.h>
#include <Audioclient.h>
#include <audiopolicy.h>
#include <mmdeviceapi.h>
#include <devicetopology.h>
#include <endpointvolume.h>
#include <AudioSessionTypes.h>
#include <functiondiscoverykeys_devpkey.h>
#endif

HRESULT macro_IMMDeviceEnumerator_GetDefaultAudioEndpoint(IMMDeviceEnumerator* enumerator, IMMDevice** device) {
	return IMMDeviceEnumerator_GetDefaultAudioEndpoint(enumerator, eRender, eConsole, device);
}

HRESULT macro_IMMDevice_Activate(IMMDevice* device, REFIID iid, DWORD dwClsCtx, PROPVARIANT* pActivationParams, void** iface) {
	return IMMDevice_Activate(device, iid, dwClsCtx, pActivationParams, iface);
}

HRESULT macro_IAudioClient_GetMixFormat(IAudioClient* client, WAVEFORMATEX** format) {
	return IAudioClient_GetMixFormat(client, format);
}

HRESULT macro_IAudioClient_Initialize(IAudioClient* client, AUDCLNT_SHAREMODE mode, DWORD flags, REFERENCE_TIME tps, DWORD buffer, WAVEFORMATEX* format, GUID* session) {
	return IAudioClient_Initialize(client, mode, flags, tps, buffer, format, session);
}

HRESULT macro_IAudioClient_GetBufferSize(IAudioClient* client, UINT32* buffer) {
	return IAudioClient_GetBufferSize(client, buffer);
}

HRESULT macro_IAudioClient_GetService(IAudioClient* client, REFIID iid, void** iface) {
	return IAudioClient_GetService(client, iid, iface);
}

HRESULT macro_IMMDeviceEnumerator_Release(IMMDeviceEnumerator* enumerator) {
	return IMMDeviceEnumerator_Release(enumerator);
}

HRESULT macro_IMMDevice_Release(IMMDevice* device) {
	return IMMDevice_Release(device);
}

HRESULT macro_IAudioClient_Release(IAudioClient* client) {
	return IAudioClient_Release(client);
}

HRESULT macro_IAudioClient_Start(IAudioClient* client) {
	return IAudioClient_Start(client);
}

HRESULT macro_IAudioClient_Stop(IAudioClient* client) {
	return IAudioClient_Stop(client);
}

HRESULT macro_IAudioClient_GetCurrentPadding(IAudioClient* client, UINT32* padding) {
	return IAudioClient_GetCurrentPadding(client, padding);
}

HRESULT macro_IAudioRenderClient_GetBuffer(IAudioRenderClient* client, UINT32 numFramesRequested, BYTE** data) {
	return IAudioRenderClient_GetBuffer(client, numFramesRequested, data);
}

HRESULT macro_IAudioRenderClient_ReleaseBuffer(IAudioRenderClient* client, UINT32 numFramesWritten, DWORD flags) {
	return IAudioRenderClient_ReleaseBuffer(client, numFramesWritten, flags);
}
*/
import "C"
import (
	"errors"
	"math"
	"unsafe"
)

const (
	opusSampleRate = 48000
	micDecodeSize  = 100000
	nsUnits        = 100
	nsToMs         = 1000000 / nsUnits
)

var (
	// BCDE0395-E52F-467C-8E3D-C4579291692E
	CLSID_MMDeviceEnumerator = C.CLSID{0xBCDE0395, 0xE52F, 0x467C, [8]C.uchar{C.uchar(0x8E), C.uchar(0x3D), C.uchar(0xC4), C.uchar(0x57), C.uchar(0x92), C.uchar(0x91), C.uchar(0x69), C.uchar(0x2E)}}
	// A95664D2-9614-4F35-A746-DE8DB63617E6
	IID_IMMDeviceEnumerator = C.IID{0xA95664D2, 0x9614, 0x4F35, [8]C.uchar{C.uchar(0xA7), C.uchar(0x46), C.uchar(0xDE), C.uchar(0x8D), C.uchar(0xB6), C.uchar(0x36), C.uchar(0x17), C.uchar(0xE6)}}
	// 1BE09788-6894-4089-8586-9A2A6C265AC5
	IID_IMMEndpoint = C.IID{0x1BE09788, 0x6894, 0x4089, [8]C.uchar{C.uchar(0x85), C.uchar(0x86), C.uchar(0x9A), C.uchar(0x2A), C.uchar(0x6C), C.uchar(0x26), C.uchar(0x5A), C.uchar(0xC5)}}
	// 1CB9AD4C-DBFA-4c32-B178-C2F568A703B2
	IID_IAudioClient = C.IID{0x1CB9AD4C, 0xDBFA, 0x4c32, [8]C.uchar{C.uchar(0xB1), C.uchar(0x78), C.uchar(0xC2), C.uchar(0xF5), C.uchar(0x68), C.uchar(0xA7), C.uchar(0x03), C.uchar(0xB2)}}
	// C8ADBD64-E71E-48a0-A4DE-185C395CD317
	IID_IAudioCaptureClient = C.IID{0xC8ADBD64, 0xE71E, 0x48a0, [8]C.uchar{C.uchar(0xA4), C.uchar(0xDE), C.uchar(0x18), C.uchar(0x5C), C.uchar(0x39), C.uchar(0x5C), C.uchar(0xD3), C.uchar(0x17)}}
	// F294ACFC-3146-4483-A7BF-ADDCA7C260E2
	IID_IAudioRenderClient = C.IID{0xF294ACFC, 0x3146, 0x4483, [8]C.uchar{C.uchar(0xA7), C.uchar(0xBF), C.uchar(0xAD), C.uchar(0xDC), C.uchar(0xA7), C.uchar(0xC2), C.uchar(0x60), C.uchar(0xE2)}}
)

type SpeakerDevice struct {
	enumerator       *C.IMMDeviceEnumerator
	device           *C.IMMDevice
	mixFormat        *C.WAVEFORMATEX
	audioClient      *C.IAudioClient
	bufferFrameCount uint32
	renderClient     *C.IAudioRenderClient
	// TODO:  OPUS
	//decoder1 *C.OpusDecoder;
	//decoder2 *C.OpusDecoder;
	channelMixBuffer [micDecodeSize]byte
	resampleBuffer   [micDecodeSize]byte
	numFramesPadding uint32
	samples          int32
	opusApplication  int
	wavType          int
}

func initialize() error {
	if int(C.CoInitialize(nil)) != 0 {
		return errors.New("audio_init failed")
	}
	return nil
}

func quit() {
	C.CoUninitialize()
}

func NewSpeakerDevice(msBufferLen uintptr) (*SpeakerDevice, error) {
	speaker := &SpeakerDevice{}
	// TODO:  OPUS
	//speaker.opusApplication = OPUS_APPLICATION_VOIP;
	speaker.wavType = 3

	hr := C.CoCreateInstance(&CLSID_MMDeviceEnumerator, nil,
		C.CLSCTX_ALL, &IID_IMMDeviceEnumerator, (*C.LPVOID)(unsafe.Pointer(&speaker.enumerator)))
	if hr < 0 {
		speaker.Free()
		return nil, errors.New("could not enumerate audio devices")
	}
	hr = C.macro_IMMDeviceEnumerator_GetDefaultAudioEndpoint(
		speaker.enumerator, &speaker.device)
	if hr < 0 {
		speaker.Free()
		return nil, errors.New("could not get default audio endpoint")
	}
	hr = C.macro_IMMDevice_Activate(speaker.device, &IID_IAudioClient, C.CLSCTX_ALL,
		nil, (*unsafe.Pointer)(unsafe.Pointer(&speaker.audioClient)))
	if hr < 0 {
		speaker.Free()
		return nil, errors.New("could not activate audio client")
	}
	hr = C.macro_IAudioClient_GetMixFormat(speaker.audioClient, &speaker.mixFormat)
	if hr < 0 {
		speaker.Free()
		return nil, errors.New("could not get mix format")
	}
	tps := (C.REFERENCE_TIME)(nsToMs * msBufferLen)
	hr = C.macro_IAudioClient_Initialize(speaker.audioClient, C.AUDCLNT_SHAREMODE_SHARED,
		0, tps, 0, speaker.mixFormat, nil)
	if hr < 0 {
		speaker.Free()
		return nil, errors.New("could not initialize audio client")
	}
	hr = C.macro_IAudioClient_GetBufferSize(speaker.audioClient, (*C.uint)(unsafe.Pointer(&speaker.bufferFrameCount)))
	if hr < 0 {
		speaker.Free()
		return nil, errors.New("could not get buffer size")
	}
	hr = C.macro_IAudioClient_GetService(speaker.audioClient, &IID_IAudioRenderClient, (*unsafe.Pointer)(unsafe.Pointer(&speaker.renderClient)))
	if hr < 0 {
		speaker.Free()
		return nil, errors.New("could not get render client")
	}
	// TODO:  Problem if we are not working with float or PCM-16
	switch speaker.mixFormat.wFormatTag {
	case C.WAVE_FORMAT_PCM:
		speaker.wavType = 1
		break
	case C.WAVE_FORMAT_IEEE_FLOAT:
		speaker.wavType = 3
		break
	case C.WAVE_FORMAT_EXTENSIBLE:
		{
			ex := (*C.WAVEFORMATEXTENSIBLE)(unsafe.Pointer(speaker.mixFormat))
			if ex.SubFormat.Data1 == 1 {
				//ex.SubFormat == KSDATAFORMAT_SUBTYPE_PCM (00000001-0000-0010-8000-00aa00389b71)
				speaker.wavType = 1
			} else if ex.SubFormat.Data1 == 3 {
				//ex.SubFormat == KSDATAFORMAT_SUBTYPE_IEEE_FLOAT (00000003-0000-0010-8000-00aa00389b71)
				speaker.wavType = 3
			}
			break
		}
	}

	speaker.samples = opusSampleRate
	//speaker.samples = speaker.mixFormat.nAvgBytesPerSec;
	//if (speaker.samples > 24000)
	//	speaker.samples = 48000;
	//else if (speaker.samples > 16000)
	//	speaker.samples = 24000;
	//else if (speaker.samples > 12000)
	//	speaker.samples = 16000;
	//else if (speaker.samples > 8000)
	//	speaker.samples = 12000;
	//else
	//	speaker.samples = 8000;

	// TODO:  OPUS
	//int error;
	//speaker.decoder1 = opus_decoder_create(speaker.samples, 1, &error);
	//speaker.decoder2 = opus_decoder_create(speaker.samples, 2, &error);
	//if (error) {
	//	audio_speaker_free(speaker);
	//	return NULL;
	//}
	return speaker, nil
}

func (s *SpeakerDevice) Free() {
	C.CoTaskMemFree((C.LPVOID)(unsafe.Pointer(s.mixFormat)))
	C.macro_IMMDeviceEnumerator_Release(s.enumerator)
	C.macro_IMMDevice_Release(s.device)
	C.macro_IAudioClient_Release(s.audioClient)
	// TODO:  OPUS
	//opus_decoder_destroy(s.decoder1);
	//opus_decoder_destroy(speas.decoder2);
}

func (s *SpeakerDevice) Start() error {
	hr := C.macro_IAudioClient_Start(s.audioClient)
	if hr < 0 {
		return errors.New("could not start audio client")
	}
	return nil
}

func (s *SpeakerDevice) Stop() error {
	hr := C.macro_IAudioClient_Stop(s.audioClient)
	if hr < 0 {
		return errors.New("could not stop audio client")
	}
	return nil
}

func (s *SpeakerDevice) LoadWavData(wav *Wav) error {
	// See how much buffer space is available.
	hr := C.macro_IAudioClient_GetCurrentPadding(s.audioClient, (*C.uint)(unsafe.Pointer(&s.numFramesPadding)))
	if hr < 0 {
		return errors.New("could not get current padding")
	}
	hr = C.macro_IAudioClient_GetBufferSize(s.audioClient, (*C.uint)(unsafe.Pointer(&s.bufferFrameCount)))
	if hr < 0 {
		return errors.New("could not get buffer size")
	}
	numFramesAvailable := s.bufferFrameCount - s.numFramesPadding
	if numFramesAvailable > 0 {
		var data *C.BYTE
		hr = C.macro_IAudioRenderClient_GetBuffer(s.renderClient, C.uint(numFramesAvailable), &data)
		if hr < 0 {
			return errors.New("could not get buffer")
		}
		ds := unsafe.Sizeof(int16(0))
		if wav.formatType == 3 {
			ds = unsafe.Sizeof(float32(0))
		}
		samples := int(wav.dataSize / int32(wav.channels) / int32(ds))
		sampleSize := samples * int(wav.channels)

		ratio := float64(s.mixFormat.nSamplesPerSec) / float64(wav.sampleRate)
		resampleTotal := int32(math.Ceil(float64(wav.dataSize)*ratio) + float64(ds))
		wavResample := make([]byte, resampleTotal)
		resample(wavResample, wav.wavData,
			int32(s.mixFormat.nSamplesPerSec), wav.sampleRate, wav.dataSize,
			wav.channels, wav.formatType)

		speakerChannels := s.mixFormat.nChannels

		// TODO:  Check to make sure this is correct
		rechannelData := *(*[]byte)(unsafe.Pointer(data))
		rechannelData = rechannelData[:numFramesAvailable:numFramesAvailable]

		if s.wavType == 1 && wav.formatType == 1 {
			rechannel(rechannelData, wavResample, int16(speakerChannels), wav.channels, sampleSize)
		} else if s.wavType == 3 && wav.formatType == 3 {
			rechannelFloat(rechannelData, wavResample, int16(speakerChannels), wav.channels, sampleSize)
		} else if s.wavType == 1 && wav.formatType == 3 {
			rechannelFl2pcm(rechannelData, wavResample, int16(speakerChannels), wav.channels, sampleSize)
		} else if s.wavType == 3 && wav.formatType == 1 {
			rechannelPcm2fl(rechannelData, wavResample, int16(speakerChannels), wav.channels, sampleSize)
		}

		framesWritten := numFramesAvailable
		if samples > 0 {
			framesWritten = uint32(samples)
		}
		hr := C.macro_IAudioRenderClient_ReleaseBuffer(s.renderClient, C.uint(framesWritten), 0)
		if hr < 0 {
			return errors.New("could not release buffer")
		}
		// Sleep to test
		//Sleep(wav.msDuration);
	}
	return nil
}

func playWav(wav *Wav) {
	//s, err := NewSpeakerDevice(1000)
	//if err != nil {
	//	return
	//}

	//C.speaker_start(s)
	//w := cWav(wav)
	//C.speaker_load_wav_data(s, &w)
	//go func() {
	//	time.Sleep(10000)
	//	C.speaker_stop(s)
	//	C.audio_speaker_free(s)
	//}()
}

func resample(out, in []byte, outRate, inRate, total int32, channels, wavType int16) int {
	// TODO:  Can this be skipped if the inRate and outRate are equal?
	offset := 0
	ratio := float64(outRate / inRate)
	resampleTotal := int(math.Floor(float64(total) * ratio))
	if wavType == 1 {
		iOutLen := len(out) / int(unsafe.Sizeof(int16(0)))
		iInLen := len(in) / int(unsafe.Sizeof(int16(0)))
		iOut := *(*[]int16)(unsafe.Pointer(&out))
		iOut = iOut[:iOutLen:iOutLen]
		iIn := *(*[]int16)(unsafe.Pointer(&in))
		iIn = iIn[:iInLen:iInLen]
		// TODO:  This needs to be changed just like the float block below
		length := resampleTotal / int(unsafe.Sizeof(int16(0)))
		for i := 0; i < length; i++ {
			idx := int(float64(i) / ratio)
			sample := iIn[idx]
			if idx+offset != i {
				// Average the two
				sample = (iOut[i-1] + sample) / 2
				offset++
			}
			iOut[i] = sample
		}
	} else {
		fOutLen := len(out) / int(unsafe.Sizeof(float32(0)))
		fInLen := len(in) / int(unsafe.Sizeof(float32(0)))
		fOut := *(*[]float32)(unsafe.Pointer(&out))
		fOut = fOut[:fOutLen:fOutLen]
		fIn := *(*[]float32)(unsafe.Pointer(&in))
		fIn = fIn[:fInLen:fInLen]
		length := resampleTotal / int(unsafe.Sizeof(float32(0)))
		if channels == 1 {
			for i := 0; i < length; i++ {
				idx := int(float64(i) / ratio)
				sample := fIn[idx]
				if idx+offset != i {
					sample = (fOut[i-1] + sample) / 2.0
					offset++
				}
				fOut[i] = sample
			}
		} else {
			// TODO:  If Opus changes to support more than 2 channels, review
			for i := 0; i < length; i += int(channels) {
				idx := int(float64(i)/ratio) & -1
				if idx+offset != i && idx+2 < length {
					fOut[i] = (fOut[i-2] + fIn[idx+2]) / 2.0
					fOut[i+1] = (fOut[i-1] + fIn[idx+3]) / 2.0
					offset += int(channels)
				} else {
					fOut[i] = fIn[idx]
					fOut[i+1] = fIn[idx+1]
				}
			}
		}
	}
	return resampleTotal
}

func rechannel(out, in []byte, outChannels, inChannels int16, sampleSize int) {
	if len(in) == 0 || len(out) == 0 || &in[0] == &out[0] {
		return
	}
	iOutLen := len(out) / int(unsafe.Sizeof(int16(0)))
	iInLen := len(in) / int(unsafe.Sizeof(int16(0)))
	d := *(*[]int16)(unsafe.Pointer(&out))
	d = d[:iOutLen:iOutLen]
	td := *(*[]int16)(unsafe.Pointer(&in))
	td = td[:iInLen:iInLen]
	idx := 0
	if outChannels == 1 && inChannels > 1 {
		for i := 0; i < sampleSize; i += 2 {
			d[idx] = int16(float64(td[i])*0.5) + int16(float64(td[i+1])+0.5)
			idx++
		}
	} else if outChannels > inChannels {
		for i := 0; i < sampleSize; i += int(inChannels) {
			val := int16(float64(td[i]) * 0.5)
			if inChannels > 1 {
				val += int16(float64(td[i+1]) * 0.5)
			}
			for j := 0; j < int(outChannels); j++ {
				d[idx] = val
				idx++
			}
		}
	} else {
		length := sampleSize * int(unsafe.Sizeof(int16(0)))
		copy(out[:length], in[:length])
	}
}

func rechannelFloat(out, in []byte, outChannels, inChannels int16, sampleSize int) {
	if len(in) == 0 || len(out) == 0 || &in[0] == &out[0] {
		return
	}
	fOutLen := len(out) / int(unsafe.Sizeof(float32(0)))
	fInLen := len(in) / int(unsafe.Sizeof(float32(0)))
	d := *(*[]float32)(unsafe.Pointer(&out))
	d = d[:fOutLen:fOutLen]
	td := *(*[]float32)(unsafe.Pointer(&in))
	td = td[:fInLen:fInLen]
	idx := 0
	if outChannels == 1 && inChannels > 1 {
		for i := 0; i < int(sampleSize); i += 2 {
			d[idx] = (td[i] * 0.5) + (td[i+1] + 0.5)
			idx++
		}
	} else if outChannels > inChannels {
		for i := 0; i < sampleSize; i += int(inChannels) {
			val := td[i] * 0.5
			if inChannels > 1 {
				val += td[i+1] * 0.5
			}
			for j := 0; j < int(outChannels); j++ {
				d[idx] = val
				idx++
			}
		}
	} else {
		length := sampleSize * int(unsafe.Sizeof(float32(0)))
		copy(out[:length], in[:length])
	}
}

func rechannelFl2pcm(out, in []byte, outChannels, inChannels int16, sampleSize int) {
	// TODO:  Test this
	if len(in) == 0 || len(out) == 0 || &in[0] == &out[0] {
		return
	}
	iOutLen := len(out) / int(unsafe.Sizeof(int16(0)))
	fInLen := len(in) / int(unsafe.Sizeof(float32(0)))
	d := *(*[]int16)(unsafe.Pointer(&out))
	d = d[:iOutLen:iOutLen]
	td := *(*[]float32)(unsafe.Pointer(&in))
	td = td[:fInLen:fInLen]
	idx := 0
	if outChannels == 1 && inChannels > 1 {
		for i := 0; i < sampleSize; i += 2 {
			d[idx] = int16((td[i] * 0.5) + (td[i+1]+0.5)*math.MaxInt16)
			idx++
		}
	} else if outChannels > inChannels {
		for i := 0; i < sampleSize; i += int(inChannels) {
			val := td[i] * 0.5
			if inChannels > 1 {
				val += td[i+1] * 0.5
			}
			for j := 0; j < int(outChannels); j++ {
				d[idx] = int16(val * math.MaxInt16)
				idx++
			}
		}
	} else {
		for i := 0; i < sampleSize; i++ {
			d[i] = int16(td[i] * math.MaxInt16)
		}
	}
}

func rechannelPcm2fl(out, in []byte, outChannels, inChannels int16, sampleSize int) {
	// TODO:  Test this
	if len(in) == 0 || len(out) == 0 || &in[0] == &out[0] {
		return
	}
	fOutLen := len(out) / int(unsafe.Sizeof(float32(0)))
	iInLen := len(in) / int(unsafe.Sizeof(int16(0)))
	d := *(*[]float32)(unsafe.Pointer(&out))
	d = d[:fOutLen:fOutLen]
	td := *(*[]int16)(unsafe.Pointer(&in))
	td = td[:iInLen:iInLen]
	idx := 0
	if outChannels == 1 && inChannels > 1 {
		for i := 0; i < int(sampleSize); i += 2 {
			d[idx] = ((float32(td[i]) * 0.5) + (float32(td[i+1]) + 0.5)) / math.MaxInt16
			idx++
		}
	} else if outChannels > inChannels {
		for i := 0; i < int(sampleSize); i += int(inChannels) {
			val := float32(td[i]) * 0.5
			if inChannels > 1 {
				val += float32(td[i+1]) * 0.5
			}
			for j := 0; j < int(outChannels); j++ {
				d[idx] = val / math.MaxInt16
				idx++
			}
		}
	} else {
		for i := 0; i < int(sampleSize); i++ {
			d[i] = float32(td[i]) / float32(math.MaxInt16)
		}
	}
}

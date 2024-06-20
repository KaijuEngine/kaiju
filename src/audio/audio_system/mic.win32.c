/******************************************************************************/
/* mic.win32.c                                                                */
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

#if defined(_WIN32) || defined(_WIN64)

#include "mic.win32.h"

#ifndef TODO_OPUS

#include <opus/opus.h>

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
#ifdef _VS
#pragma comment(lib, "ole32.lib")
#endif
#define NS_UNITS			100
#define NS_TO_MS			1000000 / NS_UNITS
#endif

#define OPUS_SAMPLE_RATE	48000
#define MIC_PACKET_SIZE		100000
#define MIC_DECODE_SIZE		100000
#define PACKET_MS			20

// BCDE0395-E52F-467C-8E3D-C4579291692E
const CLSID CLSID_MMDeviceEnumerator = { 0xBCDE0395,0xE52F,0x467C,0x8E,0x3D,0xC4,0x57,0x92,0x91,0x69,0x2E };
// A95664D2-9614-4F35-A746-DE8DB63617E6
const IID IID_IMMDeviceEnumerator = { 0xA95664D2,0x9614,0x4F35,0xA7,0x46,0xDE,0x8D,0xB6,0x36,0x17,0xE6 };
// 1BE09788-6894-4089-8586-9A2A6C265AC5
const IID IID_IMMEndpoint = { 0x1BE09788,0x6894,0x4089,0x85,0x86,0x9A,0x2A,0x6C,0x26,0x5A,0xC5 };
// 1CB9AD4C-DBFA-4c32-B178-C2F568A703B2
const IID IID_IAudioClient = { 0x1CB9AD4C, 0xDBFA, 0x4c32, 0xB1, 0x78, 0xC2, 0xF5, 0x68, 0xA7, 0x03, 0xB2 };
// C8ADBD64-E71E-48a0-A4DE-185C395CD317
const IID IID_IAudioCaptureClient = { 0xC8ADBD64, 0xE71E, 0x48a0, 0xA4, 0xDE, 0x18, 0x5C, 0x39, 0x5C, 0xD3, 0x17 };
// F294ACFC-3146-4483-A7BF-ADDCA7C260E2
const IID IID_IAudioRenderClient = { 0xF294ACFC, 0x3146, 0x4483, 0xA7, 0xBF, 0xAD, 0xDC, 0xA7, 0xC2, 0x60, 0xE2 };

struct MicrophoneDevice {
	IMMDeviceEnumerator* enumerator;
	IMMDevice* device;
	WAVEFORMATEX* mixFormat;
	IAudioClient* audioClient;
	IAudioCaptureClient* captureClient;
	OpusEncoder* encoder1;
	OpusEncoder* encoder2;
	uint8_t* packetBuffer;
	uint8_t* resampleBuffer;
	opus_int32 packetBufferSize;
	int32_t actualPacketLen;
	int32_t bufferwriteTargetSize;
	UINT32 bufferFrameCount;
	int32_t samples;
	int opusApplication;
	int wavType;
};

MicrophoneDevice* audio_mic_new(size_t msBufferLen) {
	MicrophoneDevice* mic = calloc(1, sizeof(MicrophoneDevice));
	mic->opusApplication = OPUS_APPLICATION_VOIP;
	mic->wavType = 3;
	mic->packetBufferSize = MIC_PACKET_SIZE;
	mic->packetBuffer = malloc(mic->packetBufferSize);
	mic->resampleBuffer = malloc(mic->packetBufferSize);

	HRESULT hr = CoCreateInstance(&CLSID_MMDeviceEnumerator, NULL,
		CLSCTX_ALL, &IID_IMMDeviceEnumerator, (void**)&mic->enumerator);
	if (FAILED(hr)) {
		audio_mic_free(mic);
		return NULL;
	}
	hr = IMMDeviceEnumerator_GetDefaultAudioEndpoint(
		mic->enumerator, eCapture, eConsole, &mic->device);
	if (FAILED(hr)) {
		audio_mic_free(mic);
		return NULL;
	}
	hr = IMMDevice_Activate(mic->device, &IID_IAudioClient, CLSCTX_ALL,
		NULL, (void**)&mic->audioClient);
	if (FAILED(hr)) {
		audio_mic_free(mic);
		return NULL;
	}
	hr = IAudioClient_GetMixFormat(mic->audioClient, &mic->mixFormat);
	if (FAILED(hr)) {
		audio_mic_free(mic);
		return NULL;
	}
	REFERENCE_TIME tps = (REFERENCE_TIME)(NS_TO_MS * msBufferLen);
	hr = IAudioClient_Initialize(mic->audioClient, AUDCLNT_SHAREMODE_SHARED, 0,
		tps, 0, mic->mixFormat, NULL);
	if (FAILED(hr)) {
		audio_mic_free(mic);
		return NULL;
	}
	hr = IAudioClient_GetBufferSize(mic->audioClient, &mic->bufferFrameCount);
	if (FAILED(hr)) {
		audio_mic_free(mic);
		return NULL;
	}
	hr = IAudioClient_GetService(mic->audioClient, &IID_IAudioCaptureClient,
		(void**)&mic->captureClient);
	if (FAILED(hr)) {
		audio_mic_free(mic);
		return NULL;
	}
	// TODO:  Problem if we are not working with float or PCM-16
	switch (mic->mixFormat->wFormatTag) {
		case WAVE_FORMAT_PCM:
			mic->wavType = 1;
			break;
		case WAVE_FORMAT_IEEE_FLOAT:
			mic->wavType = 3;
			break;
		case WAVE_FORMAT_EXTENSIBLE:
		{
			WAVEFORMATEXTENSIBLE* ex = (WAVEFORMATEXTENSIBLE*)mic->mixFormat;
			//ex->SubFormat == KSDATAFORMAT_SUBTYPE_PCM (00000001-0000-0010-8000-00aa00389b71)
			if (ex->SubFormat.Data1 == 1)
				mic->wavType = 1;
			//ex->SubFormat == KSDATAFORMAT_SUBTYPE_IEEE_FLOAT (00000003-0000-0010-8000-00aa00389b71)
			else if (ex->SubFormat.Data1 == 3)
				mic->wavType = 3;
			break;
		}
	}
	mic->samples = OPUS_SAMPLE_RATE;
	//mic->samples = mic->mixFormat->nSamplesPerSec;
	//if (mic->samples > 24000)
	//	mic->samples = 48000;
	//else if (mic->samples > 16000)
	//	mic->samples = 24000;
	//else if (mic->samples > 12000)
	//	mic->samples = 16000;
	//else if (mic->samples > 8000)
	//	mic->samples = 12000;
	//else
	//	mic->samples = 8000;

	mic->bufferwriteTargetSize = (mic->samples / (1000 / PACKET_MS))
		* mic->mixFormat->nChannels * (mic->wavType == 3
		? sizeof(float) : sizeof(opus_int16));

	int error = 0;
	//OPUS_APPLICATION_AUDIO - good for music
	//OPUS_APPLICATION_RESTRICTED_LOWDELAY - disables the speech-optimized mode for low delay
	mic->encoder1 = opus_encoder_create(mic->samples, 1, mic->opusApplication, &error);
	mic->encoder2 = opus_encoder_create(mic->samples, 2, mic->opusApplication, &error);
	/* NOTE:  Regardless of the sampling rate and number channels selected, the Opus encoder can switch to a lower audio bandwidth or number of channels if the bitrate selected is too low. This also means that it is safe to always use 48 kHz stereo input and let the encoder optimize the encoding.
	*/
	/* Only change if necessary
		opus_encoder_ctl(enc, OPUS_SET_BITRATE(bitrate));
		opus_encoder_ctl(enc, OPUS_SET_COMPLEXITY(complexity));
		opus_encoder_ctl(enc, OPUS_SET_SIGNAL(signal_type));
	*/
	if (error) {
		audio_mic_free(mic);
		return NULL;
	}
	return mic;
}

void audio_mic_free(MicrophoneDevice* mic) {
	CoTaskMemFree(mic->mixFormat);
	IMMDeviceEnumerator_Release(mic->enumerator);
	IMMDevice_Release(mic->device);
	IAudioClient_Release(mic->audioClient);
	IAudioCaptureClient_Release(mic->captureClient);
	opus_encoder_destroy(mic->encoder1);
	opus_encoder_destroy(mic->encoder2);
	free(mic->packetBuffer);
	free(mic->resampleBuffer);
	free(mic);
}

int mic_start(MicrophoneDevice* mic) {
	HRESULT hr = IAudioClient_Start(mic->audioClient);
	if (FAILED(hr))
		return -1;
	else
		return 0;
}

int mic_stop(MicrophoneDevice* mic) {
	HRESULT hr = IAudioClient_Stop(mic->audioClient);
	if (FAILED(hr))
		return -1;
	else
		return 0;
}

int mic_encode(MicrophoneDevice* mic, int32_t* outReadLen) {
	mic->actualPacketLen = 0;
	BYTE* data;
	DWORD flags;
	UINT32 numFramesAvailable;
	UINT32 packetLength = 0;

	// TODO: If the mic is more channels than 2 we need to down channel
	//int channels = mic->mixFormat->nChannels == 1 ? 1 : 2;

	HRESULT hr = IAudioCaptureClient_GetNextPacketSize(mic->captureClient, &packetLength);
	HRESULT packetHR;
	if (!FAILED(hr)) {
		int32_t writeLen = 0;
		while (writeLen < mic->bufferwriteTargetSize) {
			// Get the available data in the shared buffer.
			packetHR = IAudioCaptureClient_GetBuffer(mic->captureClient, &data,
				&numFramesAvailable, &flags, NULL, NULL);
			if (!FAILED(packetHR) && numFramesAvailable > 0) {
				int32_t total = numFramesAvailable * mic->mixFormat->nBlockAlign;
				writeLen += local_resample(mic->resampleBuffer + writeLen, data,
					mic->samples, mic->mixFormat->nSamplesPerSec, total,
					mic->mixFormat->nChannels, mic->wavType);
				if (flags & AUDCLNT_BUFFERFLAGS_SILENT)
				{
					// TODO:  Write 0 audio
				}
			}
			packetHR = IAudioCaptureClient_ReleaseBuffer(mic->captureClient, numFramesAvailable);
			if (FAILED(packetHR))
				break;
			packetHR = IAudioCaptureClient_GetNextPacketSize(mic->captureClient, &packetLength);
			if (FAILED(packetHR))
				break;
		}
		mic->packetBuffer[0] = mic->wavType;
		if (mic->wavType == 1) {
			int32_t size = writeLen / mic->mixFormat->nChannels / sizeof(opus_int16);
			if (mic->mixFormat->nChannels == 1) {
				mic->actualPacketLen = opus_encode(mic->encoder1, (opus_int16*)mic->resampleBuffer,
					size, mic->packetBuffer + 1, mic->packetBufferSize - 1);
			} else {
				mic->actualPacketLen = opus_encode(mic->encoder2, (opus_int16*)mic->resampleBuffer,
					size, mic->packetBuffer + 1, mic->packetBufferSize - 1);
			}
		} else if (mic->wavType == 3) {
			int32_t size = writeLen / mic->mixFormat->nChannels / sizeof(float);
			if (mic->mixFormat->nChannels == 1) {
				mic->actualPacketLen = opus_encode_float(mic->encoder1, (float*)mic->resampleBuffer,
					size, mic->packetBuffer + 1, mic->packetBufferSize - 1);
			} else {
				mic->actualPacketLen = opus_encode_float(mic->encoder2, (float*)mic->resampleBuffer,
					size, mic->packetBuffer + 1, mic->packetBufferSize - 1);
			}
		}
	}
	if (mic->actualPacketLen < 0 || FAILED(packetHR)) {
		mic->actualPacketLen = 0;
		*outReadLen = 0;
		return -1;
	} else {
		mic->actualPacketLen += 1;	// wavType
		*outReadLen = mic->actualPacketLen;
		return 0;
	}
}

int speaker_mic_decode(SpeakerDevice* speaker, const uint8_t* packet, int32_t len) {
	// See how much buffer space is available.
	HRESULT hr = IAudioClient_GetCurrentPadding(speaker->audioClient, &speaker->numFramesPadding);
	if (FAILED(hr))
		return -1;
	
	hr = IAudioClient_GetBufferSize(speaker->audioClient, &speaker->bufferFrameCount);
	if (FAILED(hr))
		return -2;

	int32_t numFramesAvailable = speaker->bufferFrameCount - speaker->numFramesPadding;
	if (numFramesAvailable > 0) {
		BYTE* data;
		hr = IAudioRenderClient_GetBuffer(speaker->renderClient,
			numFramesAvailable, &data);

		uint8_t wavType = packet[0];
		unsigned char* buf = (unsigned char*)(packet + 1);
		int32_t size = numFramesAvailable * speaker->mixFormat->nBlockAlign;
		int speakerChannels = speaker->mixFormat->nChannels;
		int streamChannels = opus_packet_get_nb_channels(buf);
		int32_t samples = 0;
		size_t sampleSize = 0;
		size_t sampleDataSize = 0;

		void* targetData = speakerChannels == streamChannels ? data : speaker->channelMixBuffer;
		if (wavType == 1) {
			if (streamChannels == 1) {
				samples = opus_decode(speaker->decoder1,
					buf, len - 1, (opus_int16*)targetData, size, 0);
			} else {
				samples = opus_decode(speaker->decoder2,
					buf, len - 1, (opus_int16*)targetData, size, 0);
			}
			sampleSize = samples * streamChannels;
			sampleDataSize = sampleSize * sizeof(opus_int16);
		}
		else {
			if (streamChannels == 1) {
				samples = opus_decode_float(speaker->decoder1,
					buf, len - 1, (float*)targetData, size, 0);
			} else {
				samples = opus_decode_float(speaker->decoder2,
					buf, len - 1, (float*)targetData, size, 0);
			}
			sampleSize = samples * streamChannels;
			sampleDataSize = sampleSize * sizeof(float);
		}

		size_t resampleTotal = local_resample(speaker->resampleBuffer, targetData,
			speaker->mixFormat->nSamplesPerSec, OPUS_SAMPLE_RATE,
			sampleDataSize, streamChannels, wavType);
		samples = (int)(resampleTotal / sizeof(float)) / streamChannels;
		sampleSize = samples * streamChannels;
		targetData = speaker->resampleBuffer;

		if (speaker->wavType == 1 && wavType == 1)
			local_rechannel(data, targetData, speakerChannels, streamChannels, sampleSize);
		else if (speaker->wavType == 3 && wavType == 3)
			local_rechannel_float(data, targetData, speakerChannels, streamChannels, sampleSize);
		else if (speaker->wavType == 1 && wavType == 3)
			local_rechannel_fl2pcm(data, targetData, speakerChannels, streamChannels, sampleSize);
		else if (speaker->wavType == 3 && wavType == 1)
			local_rechannel_pcm2fl(data, targetData, speakerChannels, streamChannels, sampleSize);
		
		int framesWritten = samples > 0 ? samples : numFramesAvailable;
		hr = IAudioRenderClient_ReleaseBuffer(
			speaker->renderClient, framesWritten, 0);

		if (FAILED(hr))
			return -3;
		else if (samples < 0)	// decode error
			return -4;
		else
			return 0;
	} else
		return -100;
}

const uint8_t* mic_get_packet(const MicrophoneDevice* mic, int32_t* len) {
	*len = mic->actualPacketLen;
	return mic->packetBuffer;
}

#endif	// TODO_OPUS

#endif	// defined(_WIN32) || defined(_WIN64)

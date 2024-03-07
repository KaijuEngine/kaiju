/******************************************************************************/
/* audio.win32.c                                                              */
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

#include <math.h>
#include <time.h>
#include "audio.win32.h"
#include <stdio.h>
#include <float.h>
#include <stdint.h>

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
#define MIC_DECODE_SIZE		100000

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

// TODO:  OPUS
typedef int64_t opus_int16;

struct SpeakerDevice {
	IMMDeviceEnumerator* enumerator;
	IMMDevice* device;
	WAVEFORMATEX* mixFormat;
	IAudioClient* audioClient;
	UINT32 bufferFrameCount;
	IAudioRenderClient* renderClient;
	// TODO:  OPUS
	//OpusDecoder* decoder1;
	//OpusDecoder* decoder2;
	uint8_t* channelMixBuffer;
	uint8_t* resampleBuffer;
	UINT32 numFramesPadding;
	int32_t samples;
	int opusApplication;
	int wavType;
};

static int32_t local_resample(void* out, const void* in, size_t outRate,
	size_t inRate, size_t total, int channels, int wavType);
static void local_rechannel(void* out, const void* in,
	int16_t outChannels, int16_t inChannels, size_t sampleSize);
static void local_rechannel_float(void* out, void* in,
	int16_t outChannels, int16_t inChannels, size_t sampleSize);
static void local_rechannel_fl2pcm(void* out, void* in,
	int16_t outChannels, int16_t inChannels, size_t sampleSize);
static void local_rechannel_pcm2fl(void* out, void* in,
	int16_t outChannels, int16_t inChannels, size_t sampleSize);

int audio_init() {
	CoInitialize(NULL);
	return 0;
}

void audio_quit() {
	CoUninitialize();
}

SpeakerDevice* audio_speaker_new(size_t msBufferLen) {
	SpeakerDevice* speaker = calloc(1, sizeof(SpeakerDevice));
	// TODO:  OPUS
	//speaker->opusApplication = OPUS_APPLICATION_VOIP;
	speaker->wavType = 3;
	speaker->channelMixBuffer = malloc(MIC_DECODE_SIZE);
	speaker->resampleBuffer = malloc(MIC_DECODE_SIZE);

	HRESULT hr = CoCreateInstance(&CLSID_MMDeviceEnumerator, NULL,
		CLSCTX_ALL, &IID_IMMDeviceEnumerator, (void**)&speaker->enumerator);
	if (FAILED(hr)) {
		audio_speaker_free(speaker);
		return NULL;
	}
	hr = IMMDeviceEnumerator_GetDefaultAudioEndpoint(
		speaker->enumerator, eRender, eConsole, &speaker->device);
	if (FAILED(hr)) {
		audio_speaker_free(speaker);
		return NULL;
	}
	hr = IMMDevice_Activate(speaker->device, &IID_IAudioClient, CLSCTX_ALL,
		NULL, (void**)&speaker->audioClient);
	if (FAILED(hr)) {
		audio_speaker_free(speaker);
		return NULL;
	}
	hr = IAudioClient_GetMixFormat(speaker->audioClient, &speaker->mixFormat);
	if (FAILED(hr)) {
		audio_speaker_free(speaker);
		return NULL;
	}
	REFERENCE_TIME tps = (REFERENCE_TIME)(NS_TO_MS * msBufferLen);
	hr = IAudioClient_Initialize(speaker->audioClient, AUDCLNT_SHAREMODE_SHARED, 0,
		tps, 0, speaker->mixFormat, NULL);
	if (FAILED(hr)) {
		audio_speaker_free(speaker);
		return NULL;
	}
	hr = IAudioClient_GetBufferSize(speaker->audioClient, &speaker->bufferFrameCount);
	if (FAILED(hr)) {
		audio_speaker_free(speaker);
		return NULL;
	}
	hr = IAudioClient_GetService(speaker->audioClient, &IID_IAudioRenderClient,
		(void**)&speaker->renderClient);
	if (FAILED(hr)) {
		audio_speaker_free(speaker);
		return NULL;
	}
	// TODO:  Problem if we are not working with float or PCM-16
	switch (speaker->mixFormat->wFormatTag) {
		case WAVE_FORMAT_PCM:
			speaker->wavType = 1;
			break;
		case WAVE_FORMAT_IEEE_FLOAT:
			speaker->wavType = 3;
			break;
		case WAVE_FORMAT_EXTENSIBLE:
		{
			WAVEFORMATEXTENSIBLE* ex = (WAVEFORMATEXTENSIBLE*)speaker->mixFormat;
			//ex->SubFormat == KSDATAFORMAT_SUBTYPE_PCM (00000001-0000-0010-8000-00aa00389b71)
			if (ex->SubFormat.Data1 == 1)
				speaker->wavType = 1;
			//ex->SubFormat == KSDATAFORMAT_SUBTYPE_IEEE_FLOAT (00000003-0000-0010-8000-00aa00389b71)
			else if (ex->SubFormat.Data1 == 3)
				speaker->wavType = 3;
			break;
		}
	}

	speaker->samples = OPUS_SAMPLE_RATE;
	//speaker->samples = speaker->mixFormat->nAvgBytesPerSec;
	//if (speaker->samples > 24000)
	//	speaker->samples = 48000;
	//else if (speaker->samples > 16000)
	//	speaker->samples = 24000;
	//else if (speaker->samples > 12000)
	//	speaker->samples = 16000;
	//else if (speaker->samples > 8000)
	//	speaker->samples = 12000;
	//else
	//	speaker->samples = 8000;

	int error;
	// TODO:  OPUS
	//speaker->decoder1 = opus_decoder_create(speaker->samples, 1, &error);
	//speaker->decoder2 = opus_decoder_create(speaker->samples, 2, &error);
	if (error) {
		audio_speaker_free(speaker);
		return NULL;
	}
	return speaker;
}

void audio_speaker_free(SpeakerDevice* speaker) {
	CoTaskMemFree(speaker->mixFormat);
	IMMDeviceEnumerator_Release(speaker->enumerator);
	IMMDevice_Release(speaker->device);
	IAudioClient_Release(speaker->audioClient);
	// TODO:  OPUS
	//opus_decoder_destroy(speaker->decoder1);
	//opus_decoder_destroy(speaker->decoder2);
	free(speaker->channelMixBuffer);
	free(speaker->resampleBuffer);
	free(speaker);
}

int speaker_start(SpeakerDevice* speaker) {
	HRESULT hr = IAudioClient_Start(speaker->audioClient);
	if (FAILED(hr))
		return -1;
	else
		return 0;
}

int speaker_stop(SpeakerDevice* speaker) {
	HRESULT hr = IAudioClient_Stop(speaker->audioClient);
	if (FAILED(hr))
		return -1;
	else
		return 0;
}

int speaker_load_wav_data(SpeakerDevice* speaker, const AudioWav* wav) {
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

		size_t ds = wav->formatType == 3 ? sizeof(float) : sizeof(opus_int16);
		int32_t samples = (int32_t)(wav->dataSize / wav->channels / ds);
		int sampleSize = samples * wav->channels;

		const double ratio = (double)speaker->mixFormat->nSamplesPerSec / wav->sampleRate;
		const int32_t resampleTotal = (int32_t)(ceil(wav->dataSize * ratio) + ds);
		uint8_t* wavResample = malloc(resampleTotal);
		local_resample(wavResample, wav->wavData,
			speaker->mixFormat->nSamplesPerSec, wav->sampleRate, wav->dataSize,
			wav->channels, wav->formatType);

		int speakerChannels = speaker->mixFormat->nChannels;
		if (speaker->wavType == 1 && wav->formatType == 1)
			local_rechannel(data, wavResample, speakerChannels, wav->channels, sampleSize);
		else if (speaker->wavType == 3 && wav->formatType == 3)
			local_rechannel_float(data, wavResample, speakerChannels, wav->channels, sampleSize);
		else if (speaker->wavType == 1 && wav->formatType == 3)
			local_rechannel_fl2pcm(data, wavResample, speakerChannels, wav->channels, sampleSize);
		else if (speaker->wavType == 3 && wav->formatType == 1)
			local_rechannel_pcm2fl(data, wavResample, speakerChannels, wav->channels, sampleSize);

		int framesWritten = samples > 0 ? samples : numFramesAvailable;
		hr = IAudioRenderClient_ReleaseBuffer(
			speaker->renderClient, framesWritten, 0);
		// Sleep to test
		Sleep(wav->msDuration);
		free(wavResample);
	}
	return 0;
}

int32_t local_resample(void* out, const void* in, size_t outRate,
	size_t inRate, size_t total, int channels, int wavType) {
	// TODO:  Can this be skipped if the inRate and outRate are equal?
	size_t offset = 0;
	const double ratio = (double)outRate / inRate;
	const int32_t resampleTotal = (int32_t)floor(total * ratio);
	if (wavType == 1) {
		// TODO:  This needs to be changed just like the float block below
		const size_t len = resampleTotal / sizeof(opus_int16);
		for (size_t i = 0; i < len; ++i) {
			size_t idx = (size_t)(i / ratio);
			opus_int16 sample = ((opus_int16*)in)[idx];
			if (idx + offset != i) {
				// Average the two
				sample = (((opus_int16*)out)[i - 1] + sample) / 2;
				offset++;
			}
			((opus_int16*)out)[i] = sample;
		}
	} else {
		float* fOut = (float*)out;
		float* fIn = (float*)in;
		const size_t len = resampleTotal / sizeof(float);
		if (channels == 1) {
			for (size_t i = 0; i < len; ++i) {
				size_t idx = (size_t)(i / ratio);
				float sample = fIn[idx];
				if (idx + offset != i)
				{
					sample = (fOut[i - 1] + sample) / 2.0F;
					offset++;
				}
				fOut[i] = sample;
			}
		}
		// TODO:  If Opus changes to support more than 2 channels, review
		else
		{
			for (size_t i = 0; i < len; i += channels) {
				size_t idx = (size_t)(i / ratio) & (SIZE_MAX - 1);
				if (idx + offset != i && idx + 2 < len) {
					fOut[i] = (fOut[i - 2] + fIn[idx + 2]) / 2.0F;
					fOut[i + 1] = (fOut[i - 1] + fIn[idx + 3]) / 2.0F;
					offset += channels;
				} else {
					fOut[i] = fIn[idx];
					fOut[i + 1] = fIn[idx + 1];
				}
			}
		}
	}
	return resampleTotal;
}

void local_rechannel(void* out, const void* in,
	int16_t outChannels, int16_t inChannels, size_t sampleSize)
{
	if (in == out)
		return;
	else if (outChannels == 1 && inChannels > 1) {
		size_t idx = 0;
		opus_int16* d = (opus_int16*)out;
		opus_int16* td = (opus_int16*)in;
		for (size_t i = 0; i < sampleSize; i += 2)
			d[idx++] = (opus_int16)(td[i] * 0.5f) + (opus_int16)(td[i + 1] + 0.5f);
	} else if (outChannels > inChannels) {
		opus_int16* d = (opus_int16*)out;
		opus_int16* td = (opus_int16*)in;
		size_t dIdx = 0;
		for (size_t i = 0; i < sampleSize; i += inChannels) {
			opus_int16 val = (opus_int16)(td[i] * 0.5F);
			if (inChannels > 1)
				val += (opus_int16)(td[i + 1] * 0.5F);
			for (int16_t j = 0; j < outChannels; ++j)
				d[dIdx++] = val;
		}
	} else
		memcpy_s(out, sampleSize * sizeof(opus_int16), in, sampleSize * sizeof(opus_int16));
}

void local_rechannel_float(void* out, void* in,
	int16_t outChannels, int16_t inChannels, size_t sampleSize)
{
	if (in == out)
		return;
	else if (outChannels == 1 && inChannels > 1) {
		size_t idx = 0;
		float* d = (float*)out;
		float* td = (float*)in;
		for (size_t i = 0; i < sampleSize; i += 2)
			d[idx++] = (td[i] * 0.5f) + (td[i + 1] + 0.5f);
	} else if (outChannels > inChannels) {
		float* d = (float*)out;
		float* td = (float*)in;
		size_t dIdx = 0;
		for (size_t i = 0; i < sampleSize; i += inChannels)
		{
			float val = td[i] * 0.5F;
			if (inChannels > 1)
				val += td[i + 1] * 0.5F;
			for (int16_t j = 0; j < outChannels; ++j)
				d[dIdx++] = val;
		}
	} else
		memcpy_s(out, sampleSize * sizeof(float), in, sampleSize * sizeof(float));
}

void local_rechannel_fl2pcm(void* out, void* in,
	int16_t outChannels, int16_t inChannels, size_t sampleSize)
{
	// TODO:  Test this
	if (in == out)
		return;
	else if (outChannels == 1 && inChannels > 1) {
		size_t idx = 0;
		opus_int16* d = (opus_int16*)out;
		float* td = (float*)in;
		for (size_t i = 0; i < sampleSize; i += 2)
			d[idx++] = (opus_int16)((td[i] * 0.5f) + (td[i + 1] + 0.5f) * INT16_MAX);
	} else if (outChannels > inChannels) {
		opus_int16* d = (opus_int16*)out;
		float* td = (float*)in;
		size_t dIdx = 0;
		for (size_t i = 0; i < sampleSize; i += inChannels)
		{
			float val = td[i] * 0.5F;
			if (inChannels > 1)
				val += td[i + 1] * 0.5F;
			for (int16_t j = 0; j < outChannels; ++j)
				d[dIdx++] = (opus_int16)(val * INT16_MAX);
		}
	} else {
		opus_int16* d = (opus_int16*)out;
		float* td = (float*)in;
		for (size_t i = 0; i < sampleSize; ++i)
			d[i] = (opus_int16)(td[i] * INT16_MAX);
	}
}

void local_rechannel_pcm2fl(void* out, void* in,
	int16_t outChannels, int16_t inChannels, size_t sampleSize)
{
	// TODO:  Test this
	if (in == out)
		return;
	else if (outChannels == 1 && inChannels > 1) {
		size_t idx = 0;
		float* d = (float*)out;
		opus_int16* td = (opus_int16*)in;
		for (size_t i = 0; i < sampleSize; i += 2)
			d[idx++] = ((td[i] * 0.5f) + (td[i + 1] + 0.5f)) / INT16_MAX;
	} else if (outChannels > inChannels) {
		float* d = (float*)out;
		opus_int16* td = (opus_int16*)in;
		size_t dIdx = 0;
		for (size_t i = 0; i < sampleSize; i += inChannels)
		{
			float val = td[i] * 0.5F;
			if (inChannels > 1)
				val += td[i + 1] * 0.5F;
			for (int16_t j = 0; j < outChannels; ++j)
				d[dIdx++] = val / INT16_MAX;
		}
	} else {
		float* d = (float*)out;
		opus_int16* td = (opus_int16*)in;
		for (size_t i = 0; i < sampleSize; ++i)
			d[i] = (float)td[i] / INT16_MAX;
	}
}

/******************************************************************************/
/* mic.win32.h                                                                */
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

#ifndef MIC_H
#define MIC_H

#if defined(_WIN32) || defined(_WIN64)

#define TODO_OPUS 1
#ifndef TODO_OPUS
#include <stdint.h>

typedef struct SpeakerDevice SpeakerDevice;
typedef struct MicrophoneDevice MicrophoneDevice;

MicrophoneDevice* audio_mic_new(size_t msBufferLen);
void audio_mic_free(MicrophoneDevice* mic);
int mic_start(MicrophoneDevice* mic);
int mic_stop(MicrophoneDevice* mic);
int mic_encode(MicrophoneDevice* mic, int32_t* outReadLen);
int speaker_mic_decode(SpeakerDevice* speaker, const uint8_t* packet, int32_t len);
const uint8_t* mic_get_packet(const MicrophoneDevice* mic, int32_t* len);
*/

#endif	// TODO_OPUS
#endif	// defined(_WIN32) || defined(_WIN64)

#endif

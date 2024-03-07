/******************************************************************************/
/* wav.go                                                                     */
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

import (
	"errors"
	"kaiju/engine"
	"kaiju/klib"
	"unsafe"
)

const (
	wavFormatPcm  = 1
	wavFormatFlt  = 3
	wavHeaderRiff = "RIFF"
	wavHeaderWave = "WAVE"
	wavHeaderFmt  = "fmt "
	wavHeaderData = "data"
	wavHeaderFact = "fact"
)

type Wav struct {
	rawData              []byte
	wavData              []byte // Data at top because we need to align on 64-bit
	riff                 [4]byte
	size                 int32
	wave                 [4]byte
	fmt                  [4]byte
	fmtLen               int32
	formatType           int16
	channels             int16
	sampleRate           int32
	averageSample        int32
	bitsPerSampleChannel int16
	bitsPerSample        int16
	data                 [4]byte
	dataSize             int32
	msDuration           int32
}

func LoadWav(host *engine.Host, wavFile string) (*Wav, error) {
	data, err := host.AssetDatabase().Read(wavFile)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("empty file")
	}
	if len(data) < 44 {
		return nil, errors.New("file too small")
	}
	wav := new(Wav)
	startOffset := unsafe.Offsetof(wav.riff)
	cpySize := unsafe.Sizeof(*wav) - startOffset
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(wav)) + startOffset)
	klib.Memcpy(ptr, unsafe.Pointer(&data[0]), uint64(cpySize))
	hRiff := *(*int32)(unsafe.Pointer(&[]byte(wavHeaderRiff)[0]))
	hWave := *(*int32)(unsafe.Pointer(&[]byte(wavHeaderWave)[0]))
	hFmt := *(*int32)(unsafe.Pointer(&[]byte(wavHeaderFmt)[0]))
	hData := *(*int32)(unsafe.Pointer(&[]byte(wavHeaderData)[0]))
	if *(*int32)(unsafe.Pointer(&wav.riff[0])) != hRiff {
		return nil, errors.New("invalid riff")
	}
	if *(*int32)(unsafe.Pointer(&wav.wave[0])) != hWave {
		return nil, errors.New("invalid wave")
	}
	if *(*int32)(unsafe.Pointer(&wav.fmt[0])) != hFmt {
		return nil, errors.New("invalid fmt")
	}
	if *(*int32)(unsafe.Pointer(&wav.data[0])) != hData {
		return nil, errors.New("invalid data")
	}
	offset := int(cpySize)
	for i := offset; i < len(data); i++ {
		if data[i] == 'd' && *(*int32)(unsafe.Pointer(&data[i])) == hData {
			offset = i + 4
			break
		}
	}
	wav.rawData = data
	wav.wavData = data[offset:]
	wav.dataSize = int32(len(data) - offset)
	ds := int(unsafe.Sizeof(float32(0)))
	if wav.formatType == 1 {
		ds = int(unsafe.Sizeof(int16(0)))
	}
	samples := wav.dataSize / int32(wav.channels) / int32(ds)
	wav.msDuration = int32(float32(samples) / float32(wav.sampleRate) * 1000.0)
	return wav, nil
}

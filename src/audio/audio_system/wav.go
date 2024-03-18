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
	"kaiju/assets"
	"kaiju/klib"
	"math"
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

type WavFormat = int16

const (
	WavFormatPcm   WavFormat = 1
	WavFormatFloat WavFormat = 3
)

type Wav struct {
	rawData              []byte
	WavData              []byte // Data at top because we need to align on 64-bit
	riff                 [4]byte
	size                 int32
	wave                 [4]byte
	fmt                  [4]byte
	fmtLen               int32
	FormatType           WavFormat
	Channels             int16
	SampleRate           int32
	averageSample        int32
	bitsPerSampleChannel int16
	bitsPerSample        int16
	data                 [4]byte
	dataSize             int32
	msDuration           int32
}

func LoadWav(assetDatabase *assets.Database, wavFile string) (*Wav, error) {
	data, err := assetDatabase.Read(wavFile)
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
	fhData := *(*int32)(unsafe.Pointer(&[]byte(wavHeaderFact)[0]))
	if *(*int32)(unsafe.Pointer(&wav.riff[0])) != hRiff {
		return nil, errors.New("invalid riff")
	}
	if *(*int32)(unsafe.Pointer(&wav.wave[0])) != hWave {
		return nil, errors.New("invalid wave")
	}
	if *(*int32)(unsafe.Pointer(&wav.fmt[0])) != hFmt {
		return nil, errors.New("invalid fmt")
	}
	if *(*int32)(unsafe.Pointer(&wav.data[0])) != hData &&
		*(*int32)(unsafe.Pointer(&wav.data[0])) != fhData {
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
	wav.WavData = data[offset:]
	wav.dataSize = int32(len(data) - offset)
	ds := int(unsafe.Sizeof(float32(0)))
	if wav.FormatType == 1 {
		ds = int(unsafe.Sizeof(int16(0)))
	}
	samples := wav.dataSize / int32(wav.Channels) / int32(ds)
	wav.msDuration = int32(float32(samples) / float32(wav.SampleRate) * 1000.0)
	return wav, nil
}

func Resample(w *Wav, sampleRate int32) []byte {
	if w.SampleRate == sampleRate {
		return w.WavData
	}
	newData := make([]byte, int(float64(w.dataSize)*float64(sampleRate)/float64(w.SampleRate)))
	resample(newData, w.WavData, sampleRate, w.SampleRate, w.dataSize, w.Channels, w.FormatType)
	return newData
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

func Rechannel(w *Wav, channels int16) []byte {
	if w.Channels == channels {
		return w.WavData
	}
	newData := make([]byte, int(w.dataSize)*int(channels)/int(w.Channels))
	rechannel(newData, w.WavData, channels, w.Channels, len(newData)/int(unsafe.Sizeof(int16(0))))
	return newData
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

func Pcm2Float(wav *Wav) []byte {
	if wav.FormatType == WavFormatFloat {
		return wav.WavData
	}
	newData := make([]byte, int(wav.dataSize)*int(unsafe.Sizeof(float32(0)))/int(unsafe.Sizeof(int16(0))))
	rechannelPcm2fl(newData, wav.WavData, wav.Channels, wav.Channels, len(newData)/int(unsafe.Sizeof(float32(0))))
	return newData
}

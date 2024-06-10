/******************************************************************************/
/* audio.go                                                                   */
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

package audio

import (
	"bytes"
	"kaiju/audio/audio_system"
	"log/slog"
	"time"

	"github.com/ebitengine/oto/v3"
)

type Audio struct {
	otoCtx  *oto.Context
	options oto.NewContextOptions
}

func NewAudio() (Audio, error) {
	a := Audio{
		options: oto.NewContextOptions{},
	}
	a.options.SampleRate = 48000
	a.options.ChannelCount = 2
	a.options.Format = oto.FormatFloat32LE
	otoCtx, readyChan, err := oto.NewContext(&a.options)
	if err != nil {
		return Audio{}, err
	}
	a.otoCtx = otoCtx
	<-readyChan
	return a, nil
}

func (a *Audio) Play(wav *audio_system.Wav) {
	if wav == nil {
		slog.Error("Wav is nil")
		return
	}
	data := wav.WavData
	// TODO:  Rather than doing this real-time, it should be a part of the
	// import process of the asset.  This is a temporary solution.
	if int(wav.Channels) != a.options.ChannelCount {
		slog.Warn("Rechanneling audio, this is a temporary solution",
			slog.Int("channels", int(wav.Channels)),
			slog.Int("target", int(a.options.ChannelCount)))
		data = audio_system.Rechannel(wav, int16(a.options.ChannelCount))
	}
	if int(wav.SampleRate) != a.options.SampleRate {
		slog.Warn("Resampling audio, this is a temporary solution",
			slog.Int("sampleRate", int(wav.SampleRate)),
			slog.Int("target", int(a.options.SampleRate)))
		data = audio_system.Resample(wav, int32(a.options.SampleRate))
	}
	if wav.FormatType != audio_system.WavFormatFloat {
		slog.Warn("Converting audio to float, this is a temporary solution",
			slog.String("format", "Float"))
		data = audio_system.Pcm2Float(wav)
	}
	player := a.otoCtx.NewPlayer(bytes.NewReader(data))
	player.Play()
	go func() {
		for player.IsPlaying() {
			time.Sleep(time.Millisecond)
		}
		if err := player.Close(); err != nil {
			slog.Error(err.Error())
		}
	}()
}

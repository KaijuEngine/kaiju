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

func NewAudio() (*Audio, error) {
	a := &Audio{
		options: oto.NewContextOptions{},
	}
	a.options.SampleRate = 44100
	a.options.ChannelCount = 2
	a.options.Format = oto.FormatFloat32LE
	otoCtx, readyChan, err := oto.NewContext(&a.options)
	if err != nil {
		return nil, err
	}
	a.otoCtx = otoCtx
	<-readyChan
	return a, nil
}

func (a *Audio) Play(wav *audio_system.Wav) {
	// Resample if needed to a.options.SampleRate
	// Rechannel if needed to a.options.ChannelCount
	// Reformat if needed to a.options.Format
	player := a.otoCtx.NewPlayer(bytes.NewReader(wav.Data))
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

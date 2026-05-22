/******************************************************************************/
/* play_sound_entity_data.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_audio

import (
	"log/slog"
	"time"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/content_id"
)

var soundBindingKey = ""

func init() {
	engine.RegisterEntityData(PlaySoundEntityData{})
}

func SoundBindingKey() string {
	if soundBindingKey == "" {
		soundBindingKey = pod.QualifiedNameForLayout(PlaySoundEntityData{})
	}
	return soundBindingKey
}

type PlaySoundEntityData struct {
	SoundId      content_id.Sound
	DelaySeconds float32
}

func (c PlaySoundEntityData) Init(e *engine.Entity, host *engine.Host) {
	adb := host.AssetDatabase()
	if !adb.Exists(string(c.SoundId)) {
		slog.Error("the sound could not be found", "id", c.SoundId)
		return
	}
	a := host.Audio()
	clip, err := a.LoadSound(adb, string(c.SoundId))
	if err != nil {
		slog.Error("failed to load the sound clip", "id", c.SoundId, "error", err)
		return
	}
	if c.DelaySeconds <= 0 {
		a.Play(clip)
	} else {
		ms := c.DelaySeconds * 1000
		host.RunAfterTime(time.Millisecond*time.Duration(ms), func() {
			a.Play(clip)
		})
	}
}

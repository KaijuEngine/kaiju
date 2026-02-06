/******************************************************************************/
/* play_sound_entity_data.go                                                  */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package engine_entity_data_audio

import (
	"kaiju/engine"
	"kaiju/engine/encoding/pod"
	"kaiju/engine_entity_data/content_id"
	"log/slog"
	"time"
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

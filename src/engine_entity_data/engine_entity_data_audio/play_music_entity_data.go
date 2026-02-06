/******************************************************************************/
/* play_music_entity_data.go                                                  */
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
	"kaiju/platform/audio"
	"log/slog"
	"weak"
)

var musicBindingKey = ""

func init() {
	engine.RegisterEntityData(PlayMusicEntityData{})
}

func MusicBindingKey() string {
	if musicBindingKey == "" {
		musicBindingKey = pod.QualifiedNameForLayout(PlayMusicEntityData{})
	}
	return musicBindingKey
}

type PlayMusicEntityData struct {
	MusicId content_id.Music
	Loop    bool
}

type MusicPlayer struct {
	host   weak.Pointer[engine.Host]
	Clip   *audio.AudioClip
	Handle audio.VoiceHandle
}

func (c PlayMusicEntityData) Init(e *engine.Entity, host *engine.Host) {
	adb := host.AssetDatabase()
	if !adb.Exists(string(c.MusicId)) {
		slog.Error("the music could not be found", "id", c.MusicId)
		return
	}
	a := host.Audio()
	clip, err := a.LoadMusic(adb, string(c.MusicId))
	if err != nil {
		slog.Error("failed to load the music clip", "id", c.MusicId, "error", err)
		return
	}
	player := &MusicPlayer{
		host:   weak.Make(host),
		Clip:   clip,
		Handle: a.Play(clip),
	}
	if c.Loop {
		a.SetLooping(player.Handle, c.Loop)
	}
	e.AddNamedData(MusicBindingKey(), player)
	e.OnDestroy.Add(player.Stop)
}

func (p *MusicPlayer) Stop() {
	host := p.host.Value()
	if host == nil {
		return
	}
	a := host.Audio()
	if !a.IsValidVoiceHandle(p.Handle) {
		return
	}
	a.Stop(p.Handle)
}

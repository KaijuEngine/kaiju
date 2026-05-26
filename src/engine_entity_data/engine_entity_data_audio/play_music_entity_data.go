/******************************************************************************/
/* play_music_entity_data.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_audio

import (
	"log/slog"
	"weak"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/platform/audio"
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

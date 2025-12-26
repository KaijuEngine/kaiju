package engine_entity_data_audio

import (
	"kaiju/engine"
	"kaiju/engine_entity_data/content_id"
	"kaiju/platform/audio"
	"log/slog"
	"weak"
)

const MusicBindingKey = "kaiju.PlayMusicEntityData"

func init() {
	engine.RegisterEntityData(MusicBindingKey, PlayMusicEntityData{})
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
	a := host.Audio()
	clip, err := a.LoadSound(host.AssetDatabase(), string(c.MusicId))
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
	e.AddNamedData(MusicBindingKey, player)
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

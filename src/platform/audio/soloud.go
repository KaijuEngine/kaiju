package audio

import (
	"errors"
	"fmt"
	"kaiju/klib"
	"math"
	"path/filepath"
	"runtime"
)

type AudioClip struct {
	wav     SoloudWav
	key     string
	handles []uint32
	isSFX   bool
}

type Audio struct {
	soloud           SoloudHandle
	sfxVolume        float32
	bgmVolume        float32
	sfxUnmutedVolume float32
	bgmUnmutedVolume float32
	sfx              map[string]*AudioClip
	bgm              map[string]*AudioClip
}

func New() (*Audio, error) {
	audio := &Audio{
		sfx:    make(map[string]*AudioClip),
		bgm:    make(map[string]*AudioClip),
		soloud: create(),
	}
	if audio.soloud == nil {
		return audio, errors.New("failed to create an instance of soloud")
	}
	errCode := initialize(audio.soloud)
	if errCode != 0 {
		return audio, fmt.Errorf("failed to initialize soloud: (%d) %s",
			errCode, errToString(audio.soloud, errCode))
	}
	audio.SetSoundVolume(0.5)
	audio.SetMusicVolume(0.5)
	runtime.AddCleanup(audio, func(soloud SoloudHandle) {
		deinitialize(soloud)
		destroy(soloud)
	}, audio.soloud)
	return audio, nil
}

func NewClip(path string) *AudioClip {
	// TODO:  This should use the asset database to load the wav rather than
	// the file path to the audio file
	clip := &AudioClip{
		key: path,
		wav: wavCreate(),
	}
	wavLoad(path, clip.wav)
	type ClipFreeState struct {
		audio *Audio
		wav   SoloudWav
	}
	runtime.AddCleanup(clip, func(wav SoloudWav) {
		wavDestroy(wav)
	}, clip.wav)
	return clip
}

func (a *Audio) SoundVolume() float32 {
	return a.sfxVolume
}

func (a *Audio) MusicVolume() float32 {
	return a.bgmVolume
}

func (a *Audio) IsSoundMuted() bool {
	return a.sfxVolume <= math.SmallestNonzeroFloat32
}

func (a *Audio) IsMusicMuted() bool {
	return a.bgmVolume <= math.SmallestNonzeroFloat32
}

func (a *Audio) UnloadClip(clip *AudioClip) {
	if clip.isSFX {
		delete(a.sfx, clip.key)
	} else {
		delete(a.bgm, clip.key)
	}
}

func (a *Audio) MuteSounds() {
	a.sfxUnmutedVolume = a.sfxVolume
	a.SetSoundVolume(0)
}

func (a *Audio) UnmuteSounds() {
	a.SetSoundVolume(a.sfxUnmutedVolume)
}

func (a *Audio) MuteMusic() {
	a.bgmUnmutedVolume = a.bgmVolume
	a.SetMusicVolume(0)
}

func (a *Audio) UnmuteMusic() {
	a.SetMusicVolume(a.bgmUnmutedVolume)
}

func (a *Audio) LoadClip(soundPath string) *AudioClip {
	clip := NewClip(soundPath)
	if filepath.Ext(soundPath) == ".wav" {
		clip.isSFX = true
		a.sfx[clip.key] = clip
		wavSetVolume(clip.wav, a.sfxVolume)
	} else {
		a.bgm[clip.key] = clip
		wavSetVolume(clip.wav, a.bgmVolume)
	}
	return clip
}

func (a *Audio) Play(clip *AudioClip) bool {
	return play(a.soloud, clip.wav) == 0
}

func (a *Audio) Stop(clip *AudioClip) {
	stopAudioSource(a.soloud, clip.wav)
	clip.handles = clip.handles[:0]
}

func (a *Audio) PlaySound(key string) (*AudioClip, uint32) {
	if sfx, ok := a.sfx[key]; ok {
		return sfx, play(a.soloud, sfx.wav)
	}
	return nil, 0
}

func (a *Audio) PlayMusic(key string) (*AudioClip, uint32) {
	if bgm, ok := a.bgm[key]; ok {
		handle := play(a.soloud, bgm.wav)
		setLooping(a.soloud, handle, true)
		bgm.handles = append(bgm.handles, uint32(handle))
		return bgm, uint32(handle)
	}
	return nil, 0
}

func (a *Audio) SetSoundVolume(volume float32) {
	a.sfxVolume = klib.Clamp(volume, 0.0, 1.0)
	for k := range a.sfx {
		wavSetVolume(a.sfx[k].wav, volume)
	}
}

func (a *Audio) SetMusicVolume(volume float32) {
	a.bgmVolume = klib.Clamp(volume, 0.0, 1.0)
	for _, v := range a.bgm {
		wavSetVolume(v.wav, volume)
		for i := range v.handles {
			setVolume(a.soloud, v.handles[i], volume)
		}
	}
}

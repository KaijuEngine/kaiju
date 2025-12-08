/******************************************************************************/
/* soloud.go                                                                  */
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

package audio

import (
	"errors"
	"fmt"
	"kaiju/engine/assets"
	"kaiju/klib"
	"math"
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

func (a *Audio) MusicById(id string) (*AudioClip, bool) {
	c, ok := a.bgm[id]
	return c, ok
}

func (a *Audio) SoundById(id string) (*AudioClip, bool) {
	c, ok := a.sfx[id]
	return c, ok
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

func (a *Audio) LoadMusic(adb assets.Database, key string) (*AudioClip, error) {
	if c, ok := a.bgm[key]; ok {
		return c, nil
	}
	data, err := adb.Read(key)
	if err != nil {
		return nil, err
	}
	clip := newClip(a, key, data)
	a.bgm[clip.key] = clip
	wavSetVolume(clip.wav, a.bgmVolume)
	return clip, nil
}

func (a *Audio) LoadSound(adb assets.Database, key string) (*AudioClip, error) {
	if c, ok := a.sfx[key]; ok {
		return c, nil
	}
	data, err := adb.Read(key)
	if err != nil {
		return nil, err
	}
	clip := newClip(a, key, data)
	clip.isSFX = true
	a.sfx[clip.key] = clip
	wavSetVolume(clip.wav, a.sfxVolume)
	return clip, nil
}

func (a *Audio) Play(clip *AudioClip) VoiceHandle {
	return play(a.soloud, clip.wav)
}

func (a *Audio) Stop(clip *AudioClip) {
	stopAudioSource(a.soloud, clip.wav)
	clip.handles = clip.handles[:0]
}

func (a *Audio) IsValidVoiceHandle(handle VoiceHandle) bool {
	return isValidVoiceHandle(a.soloud, handle)
}

func (a *Audio) Seek(handle VoiceHandle, seconds float64) bool {
	return seek(a.soloud, handle, seconds) != 0
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

func (c *AudioClip) Length() float64 {
	return clipLength(c.wav)
}

func newClip(a *Audio, key string, data []byte) *AudioClip {
	// TODO:  This should use the asset database to load the wav rather than
	// the file path to the audio file
	clip := &AudioClip{
		key: key,
		wav: wavCreate(),
	}
	wavLoadMem(clip.wav, data)
	type ClipFreeState struct {
		audio *Audio // Hold the audio pointer so the system isn't cleaned up before wav
		wav   SoloudWav
	}
	runtime.AddCleanup(clip, func(s ClipFreeState) {
		wavDestroy(s.wav)
	}, ClipFreeState{a, clip.wav})
	return clip
}

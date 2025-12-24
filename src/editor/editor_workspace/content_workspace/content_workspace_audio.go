/******************************************************************************/
/* content_workspace_audio.go                                                 */
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

package content_workspace

import (
	"fmt"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/audio"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"weak"
)

type ContentAudioView struct {
	workspace   weak.Pointer[ContentWorkspace]
	audioPlayer *document.Element
	lastPlayed  *audio.AudioClip
	playing     *audio.AudioClip
	handle      audio.VoiceHandle
	duration    float64
	seconds     float64
	lastId      string
}

func audioPlayButton(e *document.Element) *ui.Button {
	return e.Children[1].UI.ToButton()
}

func audioTimeLabel(e *document.Element) *ui.Label {
	return e.Children[2].InnerLabel()
}

func audioSlider(e *document.Element) *ui.Slider {
	return e.Children[3].UI.ToSlider()
}

func (v *ContentAudioView) setAudioPanelVisibility(target *document.Element) {
	defer tracing.NewRegion("ContentAudioView.setAudioPanelVisibility").End()
	id := target.Attribute("id")
	cc, err := v.workspace.Value().cache.Read(id)
	if err != nil {
		slog.Error("failed to find the config to add tag to content", "id", id, "error", err)
		return
	}
	if isAudioType(cc) {
		v.audioPlayer.UI.Show()
		v.playAudioId(id)
		v.stopAudio()
		v.updateTimeLabel()
		if id != v.lastId {
			v.seconds = 0
		}
		v.lastId = id
		v.setSliderPosition()
	} else {
		v.audioPlayer.UI.Hide()
	}
}

func (v *ContentAudioView) playAudioId(id string) {
	defer tracing.NewRegion("ContentAudioView.playAudioId").End()
	w := v.workspace.Value()
	cc, err := w.cache.Read(id)
	if err != nil {
		slog.Error("failed to find the config to add tag to content", "id", id, "error", err)
		return
	}
	if cc.Config.Type == (content_database.Music{}).TypeName() {
		a := w.Host.Audio()
		clip, err := a.LoadMusic(w.Host.AssetDatabase(), cc.Id())
		if err == nil {
			v.playAudio(clip)
		} else {
			slog.Error("failed to load the music data", "error", err)
		}
	} else if cc.Config.Type == (content_database.Sound{}).TypeName() {
		a := w.Host.Audio()
		clip, err := a.LoadSound(w.Host.AssetDatabase(), cc.Id())
		if err == nil {
			v.playAudio(clip)
		} else {
			slog.Error("failed to load the sound data", "error", err)
		}
	} else {
		slog.Error("the selected content is not audio", "type", cc.Config.Type, "id", id)
		return
	}
}

func (v *ContentAudioView) playAudio(clip *audio.AudioClip) {
	defer tracing.NewRegion("ContentAudioView.playAudio").End()
	if clip == nil {
		return
	}
	w := v.workspace.Value()
	a := w.Host.Audio()
	shouldPlay := v.playing != clip
	v.stopAudio()
	if !shouldPlay {
		return
	}
	v.handle = a.Play(clip)
	if v.handle == 0 {
		slog.Error("failed to play the audio clip")
		return
	}
	a.Seek(v.handle, v.seconds)
	v.setSliderPosition()
	v.playing = clip
	v.lastPlayed = clip
	v.duration = clip.Length()
	audioPlayButton(v.audioPlayer).Label().SetText("Stop")
}

func (v *ContentAudioView) setAudioPosition(position float32) {
	defer tracing.NewRegion("ContentAudioView.setAudioPosition").End()
	v.seconds = v.duration * float64(position)
	v.updateTimeLabel()
	if v.playing == nil {
		return
	}
	a := v.workspace.Value().Host.Audio()
	if !a.IsValidVoiceHandle(v.handle) {
		clip := v.playing
		v.playing = nil
		v.handle = 0
		v.playAudio(clip)
	}
	a.Seek(v.handle, v.seconds)
}

func (v *ContentAudioView) stopAudio() {
	defer tracing.NewRegion("ContentAudioView.stopAudio").End()
	if v.playing == nil {
		return
	}
	a := v.workspace.Value().Host.Audio()
	a.Stop(v.playing)
	v.playing = nil
	v.handle = 0
	audioPlayButton(v.audioPlayer).Label().SetText("Play")
}

func (v *ContentAudioView) update(deltaTime float64) {
	defer tracing.NewRegion("ContentAudioView.update").End()
	if v.playing == nil {
		return
	}
	v.seconds += deltaTime
	if v.seconds > v.duration {
		return
	}
	v.setSliderPosition()
	v.updateTimeLabel()
}

func (v *ContentAudioView) setSliderPosition() {
	defer tracing.NewRegion("ContentAudioView.setSliderPosition").End()
	audioSlider(v.audioPlayer).SetValueWithoutEvent(float32(v.seconds / v.duration))
}

func (v *ContentAudioView) updateTimeLabel() {
	defer tracing.NewRegion("ContentAudioView.updateTimeLabel").End()
	ch, cm, cs := secondsToHMS(v.seconds)
	dh, dm, ds := secondsToHMS(v.duration)
	audioTimeLabel(v.audioPlayer).SetText(fmt.Sprintf("%03d:%02d:%02d - %03d:%02d:%02d",
		ch, cm, cs, dh, dm, ds))
}

func secondsToHMS(seconds float64) (h int, m int, s int) {
	total := int(seconds)
	h = total / 3600
	m = (total % 3600) / 60
	s = total % 60
	return
}

func isAudioType(cc content_database.CachedContent) bool {
	return cc.Config.Type == (content_database.Music{}).TypeName() ||
		cc.Config.Type == (content_database.Sound{}).TypeName()
}

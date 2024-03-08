package audio

import "kaiju/audio/audio_system"

func Init() {
	audio_system.Init()
}

func Quit() {
	audio_system.Quit()
}

func Play(wav *audio_system.Wav) {
	audio_system.PlayWav(wav)
}
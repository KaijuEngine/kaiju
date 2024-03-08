package audio_system

func Init() error {
	return initialize()
}

func Quit() {
	quit()
}

func PlayWav(wav *Wav) {
	playWav(wav)
}

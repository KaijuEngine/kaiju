//go:build !windows

package audio_system

func initialize() error { return nil }
func quit()             {}
func playWav(wav *Wav)  {}

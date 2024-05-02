//go:build !windows

package audio_system

/*
// This is here just to import CGO to make the compiler happy
*/
import "C"

func initialize() error { return nil }
func quit()             {}
func playWav(wav *Wav)  {}

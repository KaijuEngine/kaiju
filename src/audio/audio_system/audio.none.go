//go:build !windows

package audio_system

func initialize() error { return nil }
func quit()             {}

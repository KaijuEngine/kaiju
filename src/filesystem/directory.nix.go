//go:build linux || darwin

package filesystem

import (
	"os"
	"path/filepath"
)

func imageDirectory() (string, error) {
	userFolder, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userFolder, "Pictures"), nil
}

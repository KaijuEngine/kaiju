package content_opener

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func IsATextFile(file string) bool {
	return strings.HasSuffix(file, ".html") ||
		strings.HasSuffix(file, ".css") ||
		strings.HasSuffix(file, ".ini") ||
		strings.HasSuffix(file, ".json") ||
		strings.HasSuffix(file, ".txt") ||
		strings.HasSuffix(file, ".md")
}

func EditTextFile(file string) error {
	var cmd *exec.Cmd
	cmd = exec.Command("code", file)
	if err := cmd.Run(); err != nil {
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", "-a", "TextEdit", file)
		case "linux":
			cmd = exec.Command("gedit", file)
		case "windows":
			cmd = exec.Command("notepad", file)
		default:
			return fmt.Errorf("unsupported OS")
		}
	}
	return cmd.Run()
}

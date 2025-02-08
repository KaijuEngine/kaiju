package content_opener

import (
	"fmt"
	"os/exec"
	"runtime"
)

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

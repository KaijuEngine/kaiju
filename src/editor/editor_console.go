package editor

import (
	"kaiju/engine"
	"kaiju/systems/console"
	"strings"
)

func setupConsole(ed *Editor) {
	console.For(ed.container.Host).AddCommand("lua",
		"Show plugin vms that are running", func(*engine.Host, string) string {
			sb := strings.Builder{}
			for i := range ed.luaVMs {
				sb.WriteString("VM: ")
				sb.WriteString(ed.luaVMs[i].PluginPath)
				sb.WriteRune('\n')
			}
			return sb.String()
		})
}

package about_window

import (
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"os/exec"
	"runtime"
)

func openContributions(*document.DocElement) {
	cmd := "open"
	if runtime.GOOS == "windows" {
		cmd = "explorer"
	}
	exec.Command(cmd, "https://github.com/KaijuEngine/kaiju/graphs/contributors").Run()
}

func New() {
	container := host_container.New("About Window", nil)
	go container.Run(500, 300)
	<-container.PrepLock
	html := klib.MustReturn(container.Host.AssetDatabase().ReadText("ui/editor/about_window.html"))
	markup.DocumentFromHTMLString(container.Host, html, "", nil, map[string]func(*document.DocElement){
		"openContributions": openContributions,
	})
}

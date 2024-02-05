package html_preview

import (
	"fmt"
	"kaiju/engine"
	"kaiju/filesystem"
	"kaiju/host_container"
	"kaiju/systems/console"
	"kaiju/uimarkup"
	"kaiju/uimarkup/markup"
	"os"
	"time"
)

type Preview struct {
	doc          *markup.Document
	path         string
	lastModified time.Time
}

func (p *Preview) readHTML(container *host_container.HostContainer) {
	container.RunFunction(func() {
		if html, err := filesystem.ReadTextFile(p.path); err == nil {
			if p.doc != nil {
				for _, elm := range p.doc.Elements {
					elm.UI.Entity().Destroy()
				}
			}
			p.doc = uimarkup.DocumentFromHTMLString(container.Host, html, "", nil, nil)
		}
	})
	p.lastModified = time.Now()
}

func startPreview(previewContainer *host_container.HostContainer, path string) {
	preview := Preview{path: path}
	preview.readHTML(previewContainer)
	for !previewContainer.Host.Closing {
		time.Sleep(time.Second * 1)
		if s, err := os.Stat(path); err != nil {
			// TODO:  Should be able to signal a close to the window
			return
		} else if s.ModTime().After(preview.lastModified) {
			preview.readHTML(previewContainer)
		}
	}
}

func SetupConsole(host *engine.Host) {
	console.For(host).AddCommand("preview", func(filePath string) string {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Sprintf("File not found: %s", filePath)
		}
		previewContainer, err := host_container.New()
		if err != nil {
			return err.Error()
		}
		go previewContainer.Run()
		go startPreview(previewContainer, filePath)
		return fmt.Sprintf("Previewing file: %s", filePath)
	})
}

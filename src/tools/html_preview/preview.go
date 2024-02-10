package html_preview

import (
	"fmt"
	"kaiju/engine"
	"kaiju/filesystem"
	"kaiju/host_container"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/systems/console"
	"os"
	"time"
)

type Preview struct {
	doc         *document.Document
	html        string
	css         string
	bindingData any
	lastMod     time.Time
}

func (p *Preview) fileChanged() bool {
	hs, hErr := os.Stat(p.html)
	if hErr != nil {
		return false
	}
	if p.css != "" {
		cs, cErr := os.Stat(p.css)
		if cErr != nil {
			return false
		}
		return hs.ModTime().After(p.lastMod) || cs.ModTime().After(p.lastMod)
	} else {
		return hs.ModTime().After(p.lastMod)
	}
}

func (p *Preview) readHTML(container *host_container.HostContainer) {
	container.RunFunction(func() {
		if html, err := filesystem.ReadTextFile(p.html); err == nil {
			css := ""
			if p.css != "" {
				css, _ = filesystem.ReadTextFile(p.css)
			}
			if p.doc != nil {
				for _, elm := range p.doc.Elements {
					elm.UI.Entity().Destroy()
				}
			}
			p.doc = markup.DocumentFromHTMLString(
				container.Host, html, css, p.bindingData, nil)
		}
	})
	p.lastMod = time.Now()
}

func startPreview(previewContainer *host_container.HostContainer, htmlFile, cssFile string, bindingData any) {
	preview := Preview{
		html:        htmlFile,
		css:         cssFile,
		bindingData: bindingData,
	}
	preview.readHTML(previewContainer)
	for !previewContainer.Host.Closing {
		time.Sleep(time.Second * 1)
		if preview.fileChanged() {
			preview.readHTML(previewContainer)
		}
	}
}

func New(htmlFile, cssFile string, bindingData any) (*host_container.HostContainer, error) {
	c := host_container.New("HTML Preview")
	c.Host.SetFrameRateLimit(60)
	go c.Run(engine.DefaultWindowWidth, engine.DefaultWindowHeight)
	<-c.PrepLock
	go startPreview(c, htmlFile, cssFile, bindingData)
	return c, nil
}

func SetupConsole(host *engine.Host) {
	console.For(host).AddCommand("preview", func(_ *engine.Host, filePath string) string {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Sprintf("File not found: %s", filePath)
		}
		if _, err := New(filePath, "", nil); err != nil {
			return fmt.Sprintf("Error creating preview: %s", err)
		}
		return fmt.Sprintf("Previewing file: %s", filePath)
	})
}

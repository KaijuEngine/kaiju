/******************************************************************************/
/* preview.go                                                                 */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package html_preview

import (
	"encoding/json"
	"fmt"
	"kaiju/engine"
	"kaiju/filesystem"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/systems/console"
	"os"
	"strings"
	"time"
)

type Preview struct {
	doc         *document.Document
	html        string
	styles      []string
	bindingData any
	lastMod     time.Time
}

func (p *Preview) filesChanged() bool {
	hs, hErr := os.Stat(p.html)
	if hErr != nil {
		return false
	}
	if hs.ModTime().After(p.lastMod) {
		return true
	}
	for f := range p.styles {
		if s, e := os.Stat(p.styles[f]); e == nil && s.ModTime().After(p.lastMod) {
			return true
		}
	}
	return false
}

func (p *Preview) pullStyles() {
	p.styles = p.styles[:0]
	for i := range p.doc.HeadElements {
		if p.doc.HeadElements[i].Data() == "link" {
			if p.doc.HeadElements[i].Attribute("rel") == "stylesheet" {
				cssPath := p.doc.HeadElements[i].Attribute("href")
				p.styles = append(p.styles, "content/"+cssPath)
			}
		}
	}
}

func (p *Preview) readHTML(container *host_container.Container) {
	container.RunFunction(func() {
		if html, err := filesystem.ReadTextFile(p.html); err == nil {
			if p.doc != nil {
				for _, elm := range p.doc.Elements {
					elm.UI.Entity().Destroy()
				}
			}
			p.doc = markup.DocumentFromHTMLString(
				container.Host, html, "", p.bindingData, nil)
			p.pullStyles()
		}
	})
	p.lastMod = time.Now()
}

func loadBindingData(htmlFile string) any {
	bindingFile := htmlFile + ".json"
	if _, err := os.Stat(bindingFile); os.IsNotExist(err) {
		return nil
	}
	bindingData, err := filesystem.ReadTextFile(bindingFile)
	if err != nil {
		return nil
	}
	var out any
	err = klib.JsonDecode(json.NewDecoder(strings.NewReader(bindingData)), &out)
	if err != nil {
		return nil
	}
	return out
}

func startPreview(previewContainer *host_container.Container, htmlFile string) {
	preview := Preview{
		html:        htmlFile,
		bindingData: loadBindingData(htmlFile),
	}
	preview.readHTML(previewContainer)
	for !previewContainer.Host.Closing {
		time.Sleep(time.Second * 1)
		if preview.filesChanged() {
			preview.readHTML(previewContainer)
		}
	}
}

func New(htmlFile string) (*host_container.Container, error) {
	c := host_container.New("HTML Preview", nil)
	c.Host.SetFrameRateLimit(60)
	go c.Run(engine.DefaultWindowWidth, engine.DefaultWindowHeight, -1, -1)
	<-c.PrepLock
	go startPreview(c, htmlFile)
	return c, nil
}

func SetupConsole(host *engine.Host) {
	console.For(host).AddCommand("preview", func(_ *engine.Host, filePath string) string {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Sprintf("File not found: %s", filePath)
		}
		if _, err := New(filePath); err != nil {
			return fmt.Sprintf("Error creating preview: %s", err)
		}
		return fmt.Sprintf("Previewing file: %s", filePath)
	})
}

/*****************************************************************************/
/* preview.go                                                                */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

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

func (p *Preview) readHTML(container *host_container.Container) {
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

func startPreview(previewContainer *host_container.Container, htmlFile, cssFile string, bindingData any) {
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

func New(htmlFile, cssFile string, bindingData any) (*host_container.Container, error) {
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

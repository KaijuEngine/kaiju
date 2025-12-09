/******************************************************************************/
/* table_of_contents_overlay.go                                               */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package table_of_contents_overlay

import (
	"kaiju/engine"
	"kaiju/engine/assets/table_of_contents"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
)

type TableOfContentsOverlay struct {
	doc     *document.Document
	uiMan   ui.Manager
	config  Config
	changed bool
}

type Config struct {
	TOC       table_of_contents.TableOfContents
	OnChanged func(toc table_of_contents.TableOfContents)
	OnClose   func()
}

func Show(host *engine.Host, config Config) (*TableOfContentsOverlay, error) {
	defer tracing.NewRegion("table_of_contents_overlay.Show").End()
	o := &TableOfContentsOverlay{
		config: config,
	}
	o.uiMan.Init(host)
	var err error
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/table_of_contents.go.html",
		config.TOC, map[string]func(*document.Element){
			"clickRemove": o.clickRemove,
			"clickMiss":   o.clickMiss,
		})
	if err != nil {
		return o, err
	}
	return o, err
}

func (o *TableOfContentsOverlay) Close() {
	defer tracing.NewRegion("TableOfContentsOverlay.Close").End()
	o.doc.Destroy()
}

func (o *TableOfContentsOverlay) clickRemove(e *document.Element) {
	defer tracing.NewRegion("TableOfContentsOverlay.clickRemove").End()
	o.config.TOC.Remove(e.Attribute("id"))
	o.changed = true
	o.doc.RemoveElement(e.Parent.Value())
}

func (o *TableOfContentsOverlay) clickMiss(*document.Element) {
	defer tracing.NewRegion("TableOfContentsOverlay.clickMiss").End()
	o.Close()
	if o.changed && o.config.OnChanged != nil {
		o.config.OnChanged(o.config.TOC)
	}
	if o.config.OnClose != nil {
		o.config.OnClose()
	}
}

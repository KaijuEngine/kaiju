/******************************************************************************/
/* data_picker.go                                                             */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package details_window

import (
	"kaiju/editor/codegen"
	"kaiju/editor/interfaces"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
	"strconv"
	"strings"
)

type DataPicker struct {
	editor    interfaces.Editor
	container *host_container.Container
	doc       *document.Document
	uiMan     ui.Manager
	picked    bool
	lock      chan int
}

func NewDataPicker(host *engine.Host, types []codegen.GeneratedType) chan int {
	const html = "editor/ui/data_picker.html"
	dp := &DataPicker{
		container: host_container.New("Data Select", nil),
		lock:      make(chan int),
	}
	cx, cy := host.Window.Center()
	dp.uiMan.Init(dp.container.Host)
	go dp.container.Run(300, 600, cx-150, cy-300)
	<-dp.container.PrepLock
	dp.container.RunFunction(func() {
		dp.doc, _ = markup.DocumentFromHTMLAsset(&dp.uiMan, html, types,
			map[string]func(*document.Element){
				"pick":   dp.pick,
				"search": dp.search,
			})
	})
	dp.container.Host.OnClose.Add(func() {
		if !dp.picked {
			dp.lock <- -1
		}
	})
	dp.container.Host.Window.Focus()
	return dp.lock
}

func (dp *DataPicker) pick(elm *document.Element) {
	dp.picked = true
	idx, _ := strconv.Atoi(elm.Attribute("id"))
	dp.lock <- idx
	dp.container.Close()
}

func (dp *DataPicker) search(elm *document.Element) {
	input, _ := dp.doc.GetElementById("search")
	query := strings.ToLower(input.UI.ToInput().Text())
	for i := range dp.doc.Elements {
		name := dp.doc.Elements[i].Attribute("data-name")
		if name != "" {
			if strings.Contains(strings.ToLower(name), query) {
				dp.doc.Elements[i].UI.Entity().Activate()
			} else {
				dp.doc.Elements[i].UI.Entity().Deactivate()
			}
		}
	}
}

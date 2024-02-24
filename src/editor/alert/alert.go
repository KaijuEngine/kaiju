/******************************************************************************/
/* alert.go                                                                   */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                      */
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

package alert

import (
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
)

type alertMsg struct {
	Title       string
	Description string
	Placeholder string
	StrValue    string
	Ok          string
	Cancel      string
	block       chan bool
	inputBlock  chan string
	container   *host_container.Container
	doc         *document.Document
}

func (a *alertMsg) done(isOkay bool) {
	if a.block != nil {
		a.block <- true
		close(a.block)
	} else if a.inputBlock != nil {
		if !isOkay {
			a.inputBlock <- ""
		} else {
			input, _ := a.doc.GetElementById("str")
			a.inputBlock <- input.UI.(*ui.Input).Text()
		}
		close(a.inputBlock)
	}
	a.container.Close()
}

func create(title, description, placeholder, value, ok, cancel string, host *engine.Host) alertMsg {
	container := host_container.New("!!! Alert !!!", nil)
	a := alertMsg{
		Title:       title,
		Description: description,
		Ok:          ok,
		Cancel:      cancel,
		container:   container,
		Placeholder: placeholder,
		StrValue:    value,
	}
	if placeholder != "" {
		a.inputBlock = make(chan string)
	} else {
		a.block = make(chan bool)
	}
	x, y := host.Window.Center()
	go container.Run(300, 200, x-150, y-100)
	<-container.PrepLock
	a.doc = klib.MustReturn(markup.DocumentFromHTMLAsset(container.Host,
		"editor/ui/alert_window.html", a, map[string]func(*document.DocElement){
			"okClick":     func(*document.DocElement) { a.done(true) },
			"cancelClick": func(*document.DocElement) { a.done(false) },
		}))
	return a
}

func New(title, description, ok, cancel string, host *engine.Host) chan bool {
	return create(title, description, "", "", ok, cancel, host).block
}

func NewInput(title, hint, value, ok, cancel string, host *engine.Host) chan string {
	return create(title, "", hint, value, ok, cancel, host).inputBlock
}

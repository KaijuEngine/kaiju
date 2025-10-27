/******************************************************************************/
/* input_prompt_overlay.go                                                    */
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

package input_prompt

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
	"log/slog"
)

type InputPrompt struct {
	doc    *document.Document
	uiMan  ui.Manager
	config Config
	input  *document.Element
}

type Config struct {
	Title       string
	Description string
	Placeholder string
	Value       string
	ConfirmText string
	CancelText  string
	OnConfirm   func(string)
	OnCancel    func()
}

func Show(host *engine.Host, config Config) (*InputPrompt, error) {
	defer tracing.NewRegion("input_prompt.Show").End()
	ip := &InputPrompt{
		config: config,
	}
	ip.uiMan.Init(host)
	var err error
	ip.doc, err = markup.DocumentFromHTMLAsset(&ip.uiMan, "editor/ui/overlay/input_prompt.go.html",
		ip.config, map[string]func(*document.Element){
			"confirm": ip.confirm,
			"cancel":  ip.cancel,
		})
	if err != nil {
		return ip, err
	}
	ip.input, _ = ip.doc.GetElementById("input")
	return ip, err
}

func (ip *InputPrompt) Close() {
	defer tracing.NewRegion("InputPrompt.Close").End()
	ip.doc.Destroy()
}

func (ip *InputPrompt) confirm(e *document.Element) {
	defer tracing.NewRegion("InputPrompt.confirm").End()
	txt := ip.input.UI.ToInput().Text()
	ip.Close()
	if ip.config.OnConfirm == nil {
		slog.Error("the input prompt didn't have a OnConfirm set, nothing to do")
		return
	}
	ip.config.OnConfirm(txt)
}

func (ip *InputPrompt) cancel(e *document.Element) {
	defer tracing.NewRegion("InputPrompt.cancel").End()
	ip.Close()
	if ip.config.OnCancel != nil {
		ip.config.OnCancel()
	}
}

/******************************************************************************/
/* ai_prompt_overlay.go                                                       */
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

package ai_prompt

import (
	"log/slog"
	"strings"

	"github.com/KaijuEngine/kaiju/engine"
	"github.com/KaijuEngine/kaiju/engine/ui"
	"github.com/KaijuEngine/kaiju/engine/ui/markup"
	"github.com/KaijuEngine/kaiju/engine/ui/markup/document"
	"github.com/KaijuEngine/kaiju/ollama"
	"github.com/KaijuEngine/kaiju/platform/hid"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
)

const systemPrompt = `You are a helpful assistant for the Kaiju game engine with access to tools. When calling a tool, respond ONLY with valid JSON in this exact format:
{
  "tool_name": "your_tool_name",
  "arguments": {
    "param1": value1,
    "param2": value2
  }
}
Do not add extra text, explanations, or incomplete objects. Ensure all keys have values and the JSON is closed properly.
Game developers will be asking you to do tasks and you will try your best to do so.
If asked by the developer about more information about this game engine (Kaiju engine) or how to use it, consult the "docs" tool.
Before assuming you know the answer, first call the appropriate tool to ensure you have all the information.
You are to always call a tool, never respond without calling a tool first. If you can't run a tool first, reply that you don't have a tool for that action.`

type AIPrompt struct {
	doc     *document.Document
	uiMan   ui.Manager
	keyKb   hid.KeyCallbackId
	onClose func()
}

func Show(host *engine.Host, onClose func()) (*AIPrompt, error) {
	defer tracing.NewRegion("ai_prompt.Show").End()
	o := &AIPrompt{onClose: onClose}
	o.uiMan.Init(host)
	var err error
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/ai_prompt.go.html",
		nil, map[string]func(*document.Element){
			"submitPrompt": o.submitPrompt,
		})
	if err != nil {
		return o, err
	}
	o.keyKb = host.Window.Keyboard.AddKeyCallback(func(keyId int, keyState hid.KeyState) {
		if keyId == hid.KeyboardKeyEscape {
			o.Close()
		}
	})
	p, _ := o.doc.GetElementById("aiPrompt")
	p.UI.ToInput().Focus()
	return o, err
}

func (o *AIPrompt) Close() {
	defer tracing.NewRegion("ConfirmPrompt.Close").End()
	o.uiMan.Host.Window.CursorStandard()
	o.doc.Destroy()
	o.uiMan.Host.Window.Keyboard.RemoveKeyCallback(o.keyKb)
	if o.onClose == nil {
		slog.Warn("onClose was not set on the AIPrompt")
		return
	}
	o.onClose()
}

func (o *AIPrompt) submitPrompt(e *document.Element) {
	defer tracing.NewRegion("ConfirmPrompt.confirm").End()
	p := strings.TrimSpace(e.UI.ToInput().Text())
	o.Close()
	if p == "" {
		return
	}
	go func() {
		res, err := ollama.Chat("http://127.0.0.1:11434", ollama.APIRequest{
			Model:  "gpt-oss:latest",
			System: systemPrompt,
			Messages: []ollama.Message{
				{
					Role:    "system",
					Content: systemPrompt,
				},
				{
					Role:    "user",
					Content: p,
				},
			},
			Think: true,
			Options: ollama.APIRequestOptions{
				NumCtx:      65536,
				Temperature: 1.0,
			},
		})
		if err != nil {
			slog.Error("ai prompt failed", "prompt", p, "error", err)
			return
		}
		slog.Info(res.Message.Content)
	}()
}

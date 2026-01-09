/******************************************************************************/
/* content_selector_overlay.go                                                */
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

package content_selector

import (
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"strings"
)

type ContentSelector struct {
	doc      *document.Document
	uiMan    ui.Manager
	keyKb    hid.KeyCallbackId
	onSelect func(id string)
	onClose  func()
	list     *document.Element
}

type contentSelectorData struct {
	Options []contentSelectorEntry
}

type contentSelectorEntry struct {
	Id      string
	Name    string
	Texture string
}

func Show(host *engine.Host, typeName string, cache *content_database.Cache, onSelect func(id string), onClose func()) (*ContentSelector, error) {
	defer tracing.NewRegion("content_selector.Show").End()
	o := &ContentSelector{onSelect: onSelect, onClose: onClose}
	o.uiMan.Init(host)
	var err error
	all := cache.ListByType(typeName)
	data := contentSelectorData{
		Options: make([]contentSelectorEntry, len(all), len(all)+2),
	}
	for i := range all {
		data.Options[i].Id = all[i].Id()
		data.Options[i].Name = all[i].Config.Name
		if all[i].Config.Type == (content_database.Texture{}).TypeName() {
			data.Options[i].Texture = data.Options[i].Id
		} else {
			data.Options[i].Texture = "editor/textures/icons/file.png"
		}
	}
	if typeName == (content_database.Texture{}).TypeName() {
		data.Options = append(data.Options, contentSelectorEntry{
			Id:      assets.TextureSquare,
			Name:    assets.TextureSquare,
			Texture: assets.TextureSquare,
		})
	}
	data.Options = append(data.Options, contentSelectorEntry{
		Name: "None",
		// TODO:  Make a special blank texture for this
		Texture: "editor/textures/icons/file.png",
	})
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/content_selector_overlay.go.html",
		data, map[string]func(*document.Element){
			"search":        o.search,
			"selectContent": o.selectContent,
		})
	if err != nil {
		return o, err
	}
	o.keyKb = host.Window.Keyboard.AddKeyCallback(func(keyId int, keyState hid.KeyState) {
		if keyId == hid.KeyboardKeyEscape {
			o.Close()
		}
	})
	o.list, _ = o.doc.GetElementById("list")
	return o, err
}

func (o *ContentSelector) Close() {
	defer tracing.NewRegion("ContentSelector.Close").End()
	o.closeInternal()
	if o.onClose == nil {
		slog.Warn("onClose was not set on the ContentSelector")
		return
	}
	o.onClose()
}

func (o *ContentSelector) closeInternal() {
	o.uiMan.Host.Window.CursorStandard()
	o.doc.Destroy()
	o.uiMan.Host.Window.Keyboard.RemoveKeyCallback(o.keyKb)
}

func (o *ContentSelector) search(e *document.Element) {
	defer tracing.NewRegion("ContentSelector.search").End()
	q := strings.ToLower(e.UI.ToInput().Text())
	for _, c := range o.list.Children {
		lbl := strings.ToLower(c.Children[1].InnerLabel().Text())
		if strings.Contains(lbl, q) {
			c.UI.Show()
		} else {
			c.UI.Hide()
		}
	}
}

func (o *ContentSelector) selectContent(e *document.Element) {
	defer tracing.NewRegion("ContentSelector.selectContent").End()
	id := e.Attribute("id")
	o.closeInternal()
	o.onSelect(id)
}

/******************************************************************************/
/* sponsors_overlay.go                                                        */
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

package sponsors

import (
	"encoding/json"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
	"log/slog"
	"net/http"
	"slices"
)

type Sponsors struct {
	doc     *document.Document
	uiMan   ui.Manager
	onClose func()
}

type Sponsor struct {
	Name    string
	GitHub  string
	Website string
	Support int
}

func Show(host *engine.Host, onClose func()) (*Sponsors, error) {
	defer tracing.NewRegion("sponsors.Show").End()
	o := &Sponsors{onClose: onClose}
	o.uiMan.Init(host)
	sponsors := readSponsorList()
	slices.SortFunc(sponsors, func(a, b Sponsor) int {
		return b.Support - a.Support
	})
	var err error
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/sponsors_overlay.go.html",
		sponsors, map[string]func(*document.Element){
			"missClose": func(*document.Element) { o.Close() },
			"clickLink": clickLink,
		})
	if err != nil {
		return o, err
	}
	return o, err
}

func (o *Sponsors) Close() {
	defer tracing.NewRegion("Sponsors.Close").End()
	o.doc.Destroy()
	if o.onClose == nil {
		slog.Warn("onClose was not set on the Sponsors")
		return
	}
	o.onClose()
}

func clickLink(e *document.Element) {
	defer tracing.NewRegion("sponsors.clickLink").End()
	klib.OpenWebsite(e.Attribute("data-path"))
}

func readSponsorList() []Sponsor {
	const url = `https://raw.githubusercontent.com/KaijuEngine/kaiju/refs/heads/master/sponsors.json`
	sponsors := []Sponsor{}
	response, err := http.Get(url)
	if err != nil {
		slog.Error("could not communicate with GitHub for sponsor list")
		return sponsors
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&sponsors); err != nil {
		slog.Error("could not deserialize list of sponsors from GitHub")
		return sponsors
	}
	return sponsors
}

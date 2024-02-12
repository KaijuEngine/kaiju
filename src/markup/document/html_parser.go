/*****************************************************************************/
/* html_parser.go                                                            */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
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

package document

import (
	"html/template"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup/elements"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/ui"
	"strconv"
	"strings"
)

type TemplateIndexedAny struct {
	Idx   int
	Value any
}

var funcMap = template.FuncMap{
	"add":    func(a, b int) int { return a + b },
	"sub":    func(a, b int) int { return a - b },
	"mul":    func(a, b int) int { return a * b },
	"div":    func(a, b int) int { return a / b },
	"muladd": func(a, b, c int) int { return a*b + c },
	"addf":   func(a, b float32) float32 { return a + b },
	"subf":   func(a, b float32) float32 { return a - b },
	"mulf":   func(a, b float32) float32 { return a * b },
	"divf":   func(a, b float32) float32 { return a / b },
	"indexed": func(idx int, value any) TemplateIndexedAny {
		return TemplateIndexedAny{idx, value}
	},
	"count": func(count int) []int {
		out := make([]int, count)
		for i := 0; i < count; i++ {
			out[i] = i
		}
		return out
	},
}

type DocElement struct {
	UI      ui.UI
	UIPanel *ui.Panel
	HTML    *Element
}

func (d DocElement) InnerLabel() *ui.Label {
	if len(d.HTML.Children) > 0 {
		if lbl := d.HTML.Children[0].DocumentElement.UI.(*ui.Label); lbl != nil {
			return lbl
		}
	}
	return nil
}

type Document struct {
	Elements      []DocElement
	TopElements   []*DocElement
	HeadElements  []Element
	groups        map[string][]DocElement
	ids           map[string]DocElement
	classElements map[string][]DocElement
	tagElements   map[string][]DocElement
}

type parseInfo struct {
	doc          *Document
	textureCache *rendering.TextureCache
	elm          DocElement
	attrValue    string
	funcMap      map[string]func(DocElement, *Document)
}

func (h Document) GetElementById(id string) (DocElement, bool) {
	if e, ok := h.ids[id]; ok {
		return e, ok
	} else {
		return DocElement{}, ok
	}
}

func (h Document) GetElementsByGroup(group string) []DocElement {
	if e, ok := h.groups[group]; ok {
		return e
	} else {
		return []DocElement{}
	}
}

func (h Document) GetElementsByClass(class string) []DocElement {
	if e, ok := h.classElements[class]; ok {
		return e
	} else {
		return []DocElement{}
	}
}

func (h Document) GetElementsByTagName(tag string) []DocElement {
	if e, ok := h.tagElements[tag]; ok {
		return e
	} else {
		return []DocElement{}
	}
}

func TransformHTML(htmlStr string, withData any) string {
	tpl := template.Must(template.New("html").Funcs(funcMap).Parse(htmlStr))
	sb := strings.Builder{}
	tpl.ExecuteTemplate(&sb, "html", withData)
	htmlStr = sb.String()
	return htmlStr
}

func (d *Document) createUIElement(host *engine.Host, e *Element, parent *ui.Panel) {
	appendElement := func(uiElm ui.UI, panel *ui.Panel) DocElement {
		entry := DocElement{
			UI:      uiElm,
			UIPanel: panel,
			HTML:    e,
		}
		e.DocumentElement = &entry
		d.Elements = append(d.Elements, entry)
		parent.AddChild(uiElm)
		return entry
	}
	if e.IsText() {
		anchor := ui.AnchorTopLeft
		txt := strings.TrimSpace(e.Data())
		txt = strings.ReplaceAll(txt, "\r", "")
		txt = strings.ReplaceAll(txt, "\n", " ")
		txt = strings.ReplaceAll(txt, "\t", " ")
		txt = klib.ReplaceStringRecursive(txt, "  ", " ")
		label := ui.NewLabel(host, txt, anchor)
		label.SetJustify(rendering.FontJustifyLeft)
		label.SetBaseline(rendering.FontBaselineTop)
		label.SetBGColor(matrix.ColorTransparent())
		appendElement(label, nil)
	} else if tag, ok := elements.ElementMap[strings.ToLower(e.Data())]; ok {
		var panel *ui.Panel
		if e.IsImage() {
			tex, err := host.TextureCache().Texture(
				e.Attribute("src"), rendering.TextureFilterLinear)
			if err != nil {
				panic(err)
			}
			img := ui.NewImage(host, tex, ui.AnchorTopLeft)
			panel = (*ui.Panel)(img)
		} else {
			panel = ui.NewPanel(host, nil, ui.AnchorTopLeft)
			panel.SetOverflow(ui.OverflowVisible)
		}
		var uiElm ui.UI = panel
		if e.IsInput() {
			inputType := e.Attribute("type")
			switch inputType {
			case "checkbox":
				cb := panel.ConvertToCheckbox()
				if e.Attribute("checked") != "" {
					cb.SetChecked(true)
				}
				uiElm = cb
			case "slider":
				slider := panel.ConvertToSlider()
				if a := e.Attribute("value"); a != "" {
					if f, err := strconv.ParseFloat(a, 32); err == nil {
						slider.SetValue(float32(f))
					}
				}
				uiElm = slider
			case "text":
				input := panel.ConvertToInput(e.Attribute("placeholder"))
				input.SetText(e.Attribute("value"))
				uiElm = input
			}
		}
		entry := appendElement(uiElm, panel)
		for i := range e.Children {
			d.createUIElement(host, &e.Children[i], panel)
		}
		id := e.Attribute("id")
		group := e.Attribute("group")
		if len(id) > 0 {
			d.ids[id] = entry
			uiElm.Entity().SetName(id)
		}
		if len(group) > 0 {
			d.groups[group] = append(d.groups[group], entry)
		}
		classList := strings.Split(e.Attribute("class"), " ")
		for _, c := range classList {
			if len(c) == 0 {
				continue
			}
			if m, ok := d.classElements[c]; ok {
				d.classElements[c] = append(m, entry)
			} else {
				d.classElements[c] = []DocElement{entry}
			}
		}
		if m, ok := d.tagElements[tag.Key()]; ok {
			d.tagElements[tag.Key()] = append(m, entry)
		} else {
			d.tagElements[tag.Key()] = []DocElement{entry}
		}
	}
}

func (d *Document) setupBody(h *Element, host *engine.Host) *Element {
	body := h.Body()
	bodyPanel := ui.NewPanel(host, nil, ui.AnchorCenter)
	bodyPanel.Layout().AddFunction(func(l *ui.Layout) {
		w, h := float32(host.Window.Width()), float32(host.Window.Height())
		l.Scale(w, h)
	})
	bodyPanel.DontFitContent()
	bodyPanel.Clean()
	body.DocumentElement = &DocElement{
		UI:      bodyPanel,
		UIPanel: bodyPanel,
		HTML:    body,
	}
	d.Elements = append(d.Elements, *body.DocumentElement)
	d.tagElements["body"] = []DocElement{*body.DocumentElement}
	bodyClasses := strings.Split(body.Attribute("class"), " ")
	for _, c := range bodyClasses {
		if len(c) == 0 {
			continue
		}
		if m, ok := d.classElements[c]; ok {
			d.classElements[c] = append(m, *body.DocumentElement)
		} else {
			d.classElements[c] = []DocElement{*body.DocumentElement}
		}
	}
	return body
}

func DocumentFromHTMLString(host *engine.Host, htmlStr string, withData any, funcMap map[string]func(*DocElement)) *Document {
	parsed := &Document{
		Elements:      make([]DocElement, 0),
		groups:        map[string][]DocElement{},
		ids:           map[string]DocElement{},
		classElements: map[string][]DocElement{},
		tagElements:   map[string][]DocElement{},
		HeadElements:  make([]Element, 0),
	}
	h := NewHTML(TransformHTML(htmlStr, withData))
	body := parsed.setupBody(h, host)
	bodyPanel := body.DocumentElement.UIPanel
	bodyPanel.Entity().SetName("body")
	for i := range body.Children {
		idx := len(parsed.Elements)
		parsed.createUIElement(host, &body.Children[i], bodyPanel)
		if idx < len(parsed.Elements) {
			parsed.TopElements = append(parsed.TopElements, &parsed.Elements[idx])
		}
	}
	for i := range parsed.Elements {
		setupEvents(&parsed.Elements[i], funcMap)
	}
	for _, elm := range h.Children[len(h.Children)-1].Children {
		if elm.Data() == "head" {
			for _, child := range elm.Children {
				parsed.HeadElements = append(parsed.HeadElements, child)
			}
		}
	}
	return parsed
}

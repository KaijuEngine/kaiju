/*****************************************************************************/
/* html.go                                                                   */
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
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package markup

import (
	"kaiju/engine"
	"kaiju/markup/css"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func DocumentFromHTML(host *engine.Host, htmlPath, cssPath string, withData any, funcMap map[string]func(*document.DocElement)) *document.Document {
	htmlBytes, err := host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		panic("Failed to read markup file: " + htmlPath)
	}
	if cssPath == "" {
		return DocumentFromHTMLString(host, string(htmlBytes), "", withData, funcMap)
	} else {
		cssBytes, err := host.AssetDatabase().ReadText(cssPath)
		if err != nil {
			panic("Failed to read css file: " + cssPath)
		}
		return DocumentFromHTMLString(host, string(htmlBytes), string(cssBytes), withData, funcMap)
	}
}

func sizeTexts(doc *document.Document, host *engine.Host) {
	for i := range doc.Elements {
		elm := &doc.Elements[i]
		e := elm.HTML
		if e.IsText() {
			label := elm.UI.(*ui.Label)
			parentWidth := float32(-1.0)
			updateSize := func(l *ui.Layout) {
				if p := ui.FirstOnEntity(label.Entity().Parent); p != nil {
					newParentWidth := p.Layout().PixelSize().Width()
					height := l.PixelSize().Height()
					if newParentWidth != parentWidth {
						parentWidth = newParentWidth
						textSize := host.FontCache().MeasureStringWithin(
							label.FontFace(), e.Data(), label.FontSize(),
							parentWidth, label.LineHeight())
						height = textSize.Height()
					}
					l.Scale(parentWidth, height)
				}
			}
			updateSize(label.Layout())
			label.Layout().AddFunction(updateSize)
		}
		height := elm.UI.Layout().PixelSize().Y()
		p := elm.HTML.Parent
		for p != nil && p.DocumentElement != nil {
			pPanel := p.DocumentElement.UIPanel
			if pPanel.FittingContent() && pPanel.Layout().PixelSize().Y() < height {
				pPanel.Layout().ScaleHeight(height)
			}
			p = p.Parent
		}
	}
}

func DocumentFromHTMLString(host *engine.Host, html, cssStr string, withData any, funcMap map[string]func(*document.DocElement)) *document.Document {
	doc := document.DocumentFromHTMLString(host, html, withData, funcMap)
	s := rules.NewStyleSheet()
	s.Parse(css.DefaultCSS)
	s.Parse(cssStr)
	for i := range doc.HeadElements {
		if doc.HeadElements[i].Data() == "style" {
			s.Parse(doc.HeadElements[i].Children[0].Data())
		} else if doc.HeadElements[i].Data() == "link" {
			if doc.HeadElements[i].Attribute("rel") == "stylesheet" {
				cssPath := doc.HeadElements[i].Attribute("href")
				css, err := host.AssetDatabase().ReadText(cssPath)
				if err != nil {
					continue
				}
				s.Parse(css)
			}

		}
	}
	css.Apply(s, doc, host)
	sizeTexts(doc, host)
	return doc
}

/******************************************************************************/
/* html.go                                                                    */
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

package markup

import (
	"kaiju/engine"
	"kaiju/markup/css"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
	"weak"
)

func sizeTexts(doc *document.Document, host *engine.Host) {
	for i := range doc.Elements {
		e := doc.Elements[i]
		if e.IsText() {
			parentWidth := float32(-1.0)
			updateSize := func(l *ui.Layout) {
				if p := ui.FirstOnEntity(l.Ui().Entity().Parent); p != nil {
					newParentWidth := p.Layout().PixelSize().Width()
					height := l.PixelSize().Height()
					if newParentWidth != parentWidth {
						parentWidth = newParentWidth
						lbl := l.Ui().ToLabel()
						textSize := host.FontCache().MeasureStringWithin(
							lbl.FontFace(), e.Data(), lbl.FontSize(),
							parentWidth, lbl.LineHeight())
						height = textSize.Height()
					}
					l.Scale(parentWidth, height)
				}
			}
			label := e.UI.ToLabel()
			updateSize(label.Base().Layout())
			label.Base().Layout().AddFunction(updateSize)
		}
		height := e.UI.Layout().PixelSize().Y()
		p := e.Parent.Value()
		for p != nil && p.UIPanel != nil {
			pPanel := p.UIPanel
			if pPanel.FittingContentHeight() && pPanel.Base().Layout().PixelSize().Y() < height {
				pPanel.Base().Layout().ScaleHeight(height)
			}
			p = p.Parent.Value()
		}
	}
}

func DocumentFromHTMLAsset(uiMan *ui.Manager, htmlPath string, withData any, funcMap map[string]func(*document.Element)) (*document.Document, error) {
	m, err := uiMan.Host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		return nil, err
	}
	return DocumentFromHTMLString(uiMan, m, "", withData, funcMap, nil), nil
}

func DocumentFromHTMLAssetRooted(uiMan *ui.Manager, htmlPath string, withData any, funcMap map[string]func(*document.Element), root *document.Element) (*document.Document, error) {
	m, err := uiMan.Host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		return nil, err
	}
	return DocumentFromHTMLString(uiMan, m, "", withData, funcMap, root), nil
}

func DocumentFromHTMLString(uiMan *ui.Manager, html, cssStr string, withData any, funcMap map[string]func(*document.Element), root *document.Element) *document.Document {
	host := uiMan.Host
	doc := document.DocumentFromHTMLString(uiMan, html, withData, funcMap)
	if root != nil {
		for i := range doc.TopElements {
			root.UIPanel.AddChild(doc.TopElements[i].UI)
			root.Children = append(root.Children, doc.TopElements[i])
			doc.TopElements[i].Parent = weak.Make(root)
		}
	}
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
	doc.SetupStylizer(s, host, css.Apply)
	sizeTexts(doc, host)
	return doc
}

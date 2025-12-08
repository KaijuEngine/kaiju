/******************************************************************************/
/* html.go                                                                    */
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

package markup

import (
	"kaiju/build"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/document"
	"log/slog"
	"strings"
	"weak"

	gohtml "golang.org/x/net/html"
)

func DocumentFromHTMLAsset(uiMan *ui.Manager, htmlPath string, withData any, funcMap map[string]func(*document.Element)) (*document.Document, error) {
	host := uiMan.Host
	m, err := host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		return nil, err
	}
	doc := DocumentFromHTMLString(uiMan, m, "", withData, funcMap, nil)
	//if build.Debug {
	//	doc.Debug.ReloadEventId = document.Debug.ReloadStylesEvent.Add(func() {
	//		reloadDocumentStyles(doc, []string{htmlPath}, []string{}, host)
	//	})
	//	host.OnClose.Add(func() { document.Debug.ReloadStylesEvent.Clear() })
	//}
	return doc, nil
}

func DocumentFromHTMLAssetRooted(uiMan *ui.Manager, htmlPath string, withData any, funcMap map[string]func(*document.Element), root *document.Element) (*document.Document, error) {
	host := uiMan.Host
	m, err := host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		return nil, err
	}
	doc := DocumentFromHTMLString(uiMan, m, "", withData, funcMap, root)
	//if build.Debug {
	//	doc.Debug.ReloadEventId = document.Debug.ReloadStylesEvent.Add(func() {
	//		reloadDocumentStyles(doc, []string{htmlPath}, []string{}, host)
	//	})
	//}
	return doc, nil
}

func DocumentFromHTMLString(uiMan *ui.Manager, html, cssStr string, withData any, funcMap map[string]func(*document.Element), root *document.Element) *document.Document {
	host := uiMan.Host
	window := host.Window
	doc := document.DocumentFromHTMLString(uiMan, html, withData, funcMap)
	if root != nil {
		for i := range doc.TopElements {
			root.UIPanel.AddChild(doc.TopElements[i].UI)
			root.Children = append(root.Children, doc.TopElements[i])
			doc.TopElements[i].Parent = weak.Make(root)
		}
	}
	s := rules.NewStyleSheet()
	s.Parse(css.DefaultCSS, window)
	s.Parse(cssStr, window)
	for i := range doc.HeadElements {
		if doc.HeadElements[i].Data == "style" {
			s.Parse(doc.HeadElements[i].Children[0].Data, window)
		} else if doc.HeadElements[i].Data == "link" {
			if doc.HeadElements[i].Attribute("rel") == "stylesheet" {
				cssPath := doc.HeadElements[i].Attribute("href")
				css, err := host.AssetDatabase().ReadText(cssPath)
				if err != nil {
					continue
				}
				s.Parse(css, window)
			}
		}
	}
	doc.SetupStyle(s, host, css.Stylizer{Window: uiMan.Host.Window})
	return doc
}

func reloadDocumentStyles(doc *document.Document, files []string, raw []string, host *engine.Host) {
	if !build.Debug {
		slog.Error("reloadDocumentStyles should not be called in a non-debug build")
		return
	}
	findAttr := func(n *gohtml.Node, key string) string {
		for i := range n.Attr {
			if n.Attr[i].Key == key {
				return n.Attr[i].Val
			}
		}
		return ""
	}
	window := host.Window
	s := rules.NewStyleSheet()
	s.Parse(css.DefaultCSS, window)
	for i := range files {
		data, err := host.AssetDatabase().Read(files[i])
		if err != nil {
			slog.Error("reloadDocumentStyles failed to read the file", "file", files[i], "error", err)
			continue
		}
		if strings.HasSuffix(files[i], ".html") {
			tpl, err := gohtml.Parse(strings.NewReader(string(data)))
			if err != nil {
				slog.Error("reloadDocumentStyles failed to parse the html string", "file", files[i], "error", err)
				continue
			}
			for root := range tpl.ChildNodes() {
				if root.Data == "html" {
					for top := range root.ChildNodes() {
						if top.Data == "head" {
							for c := range top.ChildNodes() {
								if c.Data == "style" {
									s.Parse(c.FirstChild.Data, window)
								} else if c.Data == "link" {
									if findAttr(c, "rel") == "stylesheet" {
										cssPath := findAttr(c, "href")
										css, err := host.AssetDatabase().ReadText(cssPath)
										if err != nil {
											continue
										}
										s.Parse(css, window)
									}
								}
							}
						}
					}
				}
			}
		} else if strings.HasSuffix(files[i], ".css") {
			s.Parse(string(data), window)
		} else {
			slog.Error("failed to reloadDocumentStyles for file", "file", files[i])
		}
	}
	for i := range raw {
		s.Parse(raw[i], window)
	}
	doc.SetupStyle(s, host, css.Stylizer{})
}

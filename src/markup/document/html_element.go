/******************************************************************************/
/* html_element.go                                                            */
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

package document

import (
	"kaiju/matrix"
	"kaiju/ui"
	"strings"

	"golang.org/x/net/html"
)

type Element struct {
	UI       *ui.UI
	UIPanel  *ui.Panel
	Parent   *Element
	Children []*Element
	node     *html.Node
	attr     map[string]*html.Attribute
}

func (d Element) InnerLabel() *ui.Label {
	if len(d.Children) > 0 {
		if lbl := d.Children[0].UI.ToLabel(); lbl != nil {
			return lbl
		}
	}
	return nil
}

func (d *Element) EnforceColor(color matrix.Color) {
	d.UIPanel.EnforceColor(color)
	setChildTextBackgroundColor(d, color)
}

func (d *Element) UnEnforceColor() {
	d.UIPanel.UnEnforceColor()
	color := d.UIPanel.Base().ShaderData().FgColor
	setChildTextBackgroundColor(d, color)
}

func setChildTextBackgroundColor(elm *Element, color matrix.Color) {
	for _, c := range elm.Children {
		if c.IsText() {
			c.UI.ToLabel().SetBGColor(color)
		}
		setChildTextBackgroundColor(c, color)
	}
}

func (e *Element) IsText() bool {
	return e.node.Type == html.TextNode
}

func (e *Element) IsButton() bool {
	return e.node.Data == "button"
}

func (e *Element) IsInput() bool {
	return e.node.Data == "input"
}

func (e *Element) IsImage() bool {
	return e.node.Data == "img"
}

func NewHTML(htmlStr string) *Element {
	doc, _ := html.Parse(strings.NewReader(htmlStr))
	return createElement(doc)
}

func createElement(node *html.Node) *Element {
	root := toElement(node)
	root.setParents(nil)
	return root
}

func (e *Element) setParents(parent *Element) {
	e.Parent = parent
	for i := range e.Children {
		e.Children[i].setParents(e)
	}
}

func toElement(node *html.Node) *Element {
	elm := Element{
		node:     node,
		attr:     make(map[string]*html.Attribute),
		Children: make([]*Element, 0),
	}
	for i := 0; i < len(node.Attr); i++ {
		elm.attr[node.Attr[i].Key] = &node.Attr[i]
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if len(strings.TrimSpace(c.Data)) > 0 {
			elm.Children = append(elm.Children, toElement(c))
		}
	}
	return &elm
}

func (e *Element) Root() *Element {
	if e.Parent == nil {
		return e
	}
	return e.Parent.Root()
}

func (e *Element) Html() *Element {
	if e.Parent == nil {
		return e.Children[len(e.Children)-1]
	}
	return e.Parent.Html()
}

func (e *Element) Head() *Element {
	res := e.Html()
	for i, c := range res.Children {
		if c.node.Data == "head" {
			return res.Children[i]
		}
	}
	return nil
}

func (e *Element) Body() *Element {
	res := e.Html()
	for i, c := range res.Children {
		if c.node.Data == "body" {
			return res.Children[i]
		}
	}
	return nil
}

func (e *Element) Data() string {
	return e.node.Data
}

func (e *Element) Attribute(key string) string {
	if a, ok := e.attr[key]; ok {
		return a.Val
	}
	return ""
}

func (e *Element) FindElementById(id string) *Element {
	if e.Attribute("id") == id {
		return e
	}
	for _, c := range e.Children {
		if elm := c.FindElementById(id); elm != nil {
			return elm
		}
	}
	return nil
}

func (e *Element) FindElementLabelById(id string) *ui.Label {
	elm := e.FindElementById(id)
	return elm.InnerLabel()
}

func (e *Element) FindElementByTag(tag string) *Element {
	if !e.IsText() && e.node.Data == tag {
		return e
	}
	for _, c := range e.Children {
		if elm := c.FindElementByTag(tag); elm != nil {
			return elm
		}
	}
	return nil
}

func (e *Element) FindElementsByTag(tag string) []*Element {
	elements := make([]*Element, 0)
	if !e.IsText() && e.node.Data == tag {
		elements = append(elements, e)
	}
	for _, c := range e.Children {
		elements = append(elements, c.FindElementsByTag(tag)...)
	}
	return elements
}

/******************************************************************************/
/* html_element.go                                                            */
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

package document

import (
	"kaiju/engine"
	"kaiju/engine/systems/events"
	"kaiju/engine/ui"
	"kaiju/matrix"
	"slices"
	"strings"
	"weak"

	"golang.org/x/net/html"
)

type Element struct {
	Type       html.NodeType
	Data       string
	namespace  string
	attr       []html.Attribute
	UI         *ui.UI
	UIPanel    *ui.Panel
	Parent     weak.Pointer[Element]
	Children   []*Element
	attrMap    map[string]*html.Attribute
	Stylizer   ElementLayoutStylizer
	UIEventIds [ui.EventTypeEnd][]events.Id
}

func (e *Element) ClassList() []string {
	return strings.Split(e.Attribute("class"), " ")
}

func (e *Element) HasClass(class string) bool {
	all := e.ClassList()
	for i := range all {
		if all[i] == class {
			return true
		}
	}
	return false
}

// SetAttribute sets an attribute on the HTML element. If the attribute key is
// "class", it parses the value as space-separated class names and sets them
// using SetClasses. Otherwise, it updates the existing attribute or adds a new
// one to the element's attribute list.
func (e *Element) SetAttribute(key, value string) {
	if key == "class" {
		e.SetClasses(strings.Split(value, " ")...)
	} else {
		found := false
		for i := 0; i < len(e.attr) && !found; i++ {
			if e.attr[i].Key == key {
				found = true
				e.attr[i].Val = value
			}
		}
		if !found {
			e.attr = append(e.attr, html.Attribute{
				Key: key,
				Val: value,
			})
			e.recacheAttrs()
		}
	}
}

func (e *Element) SetClasses(classes ...string) {
	const classKey = "class"
	for i := range e.attr {
		if e.attr[i].Key == classKey {
			e.attr = slices.Delete(e.attr, i, i+1)
			break
		}
	}
	e.attr = append(e.attr, html.Attribute{
		Key: classKey,
		Val: strings.Join(classes, " "),
	})
	e.attrMap[classKey] = &e.attr[len(e.attr)-1]
}

func (d *Element) IndexOfChild(child *Element) int {
	for i := range d.Children {
		if d.Children[i] == child {
			return i
		}
	}
	return -1
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
	return e.Type == html.TextNode && (e.Parent.Value() == nil || e.Parent.Value().Data != "option")
}

func (e *Element) IsButton() bool {
	return e.Data == "button" ||
		(e.Data == "input" && e.Attribute("type") == "submit")
}

func (e *Element) IsInput() bool {
	return e.Data == "input"
}

func (e *Element) IsImage() bool {
	return e.Data == "img"
}

func (e *Element) IsSelect() bool {
	return e.Data == "select"
}

func (e *Element) IsSelectOption() bool {
	return e.Data == "option"
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
	e.Parent = weak.Make(parent)
	for i := range e.Children {
		e.Children[i].setParents(e)
	}
}

func toElement(node *html.Node) *Element {
	elm := &Element{
		Type:      node.Type,
		Data:      node.Data,
		namespace: node.Namespace,
		attr:      node.Attr,
		attrMap:   make(map[string]*html.Attribute),
		Children:  make([]*Element, 0),
	}
	elm.Stylizer = ElementLayoutStylizer{element: weak.Make(elm)}
	elm.recacheAttrs()
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if len(strings.TrimSpace(c.Data)) > 0 {
			elm.Children = append(elm.Children, toElement(c))
		}
	}
	return elm
}

func (e *Element) Root() *Element {
	if e.Parent.Value() == nil {
		return e
	}
	return e.Parent.Value().Root()
}

func (e *Element) Html() *Element {
	if e.Parent.Value() == nil {
		return e.Children[len(e.Children)-1]
	}
	return e.Parent.Value().Html()
}

func (e *Element) Head() *Element {
	res := e.Html()
	for i, c := range res.Children {
		if c.Data == "head" {
			return res.Children[i]
		}
	}
	return nil
}

func (e *Element) Body() *Element {
	res := e.Html()
	for i, c := range res.Children {
		if c.Data == "body" {
			return res.Children[i]
		}
	}
	return nil
}

func (e *Element) Attribute(key string) string {
	if a, ok := e.attrMap[key]; ok {
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
	if !e.IsText() && e.Data == tag {
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
	if !e.IsText() && e.Data == tag {
		elements = append(elements, e)
	}
	for _, c := range e.Children {
		elements = append(elements, c.FindElementsByTag(tag)...)
	}
	return elements
}

func (e *Element) Clone(parent *Element) *Element {
	elm := &Element{
		Type:      e.Type,
		Data:      e.Data,
		namespace: e.namespace,
	}
	elm.attr = append(elm.attr, e.attr...)
	elm.recacheAttrs()
	for i := range e.UIEventIds {
		elm.UIEventIds[i] = make([]events.Id, 0, len(e.UIEventIds[i]))
	}
	if parent != nil {
		parent.Children = append(parent.Children, elm)
		elm.Parent = weak.Make(parent)
	}
	var eParent *engine.Entity
	if parent != nil {
		eParent = parent.UI.Entity()
	}
	elm.UI = e.UI.Clone(eParent)
	if !elm.UI.IsType(ui.ElementTypeLabel) {
		elm.UIPanel = elm.UI.ToPanel()
	}
	elm.Children = make([]*Element, 0, len(e.Children))
	for i := range e.Children {
		e.Children[i].Clone(elm)
	}
	elm.Stylizer = e.Stylizer.clone(elm)
	return elm
}

func (e *Element) recacheAttrs() {
	e.attrMap = make(map[string]*html.Attribute)
	for i := range e.attr {
		e.attrMap[e.attr[i].Key] = &e.attr[i]
	}
}

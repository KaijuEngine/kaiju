package markup

import (
	"kaiju/ui"
	"strings"

	"golang.org/x/net/html"
)

type Element struct {
	Parent          *Element
	Children        []Element
	node            *html.Node
	attr            map[string]*html.Attribute
	DocumentElement *DocElement
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

func NewHTML(htmlStr string) *Element {
	doc, _ := html.Parse(strings.NewReader(htmlStr))
	return createElement(doc)
}

func createElement(node *html.Node) *Element {
	root := toElement(node)
	root.setParents(nil)
	return &root
}

func (e *Element) setParents(parent *Element) {
	e.Parent = parent
	for i := range e.Children {
		e.Children[i].setParents(e)
	}
}

func toElement(node *html.Node) Element {
	elm := Element{
		node:     node,
		attr:     make(map[string]*html.Attribute),
		Children: make([]Element, 0),
	}
	for i := 0; i < len(node.Attr); i++ {
		elm.attr[node.Attr[i].Key] = &node.Attr[i]
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if len(strings.TrimSpace(c.Data)) > 0 {
			elm.Children = append(elm.Children, toElement(c))
		}
	}
	return elm
}

func (e *Element) Root() *Element {
	if e.Parent == nil {
		return e
	}
	return e.Parent.Root()
}

func (e *Element) Html() *Element {
	if e.Parent == nil {
		return &e.Children[len(e.Children)-1]
	}
	return e.Parent.Html()
}

func (e *Element) Head() *Element {
	res := e.Html()
	for i, c := range res.Children {
		if c.node.Data == "head" {
			return &res.Children[i]
		}
	}
	return nil
}

func (e *Element) Body() *Element {
	res := e.Html()
	for i, c := range res.Children {
		if c.node.Data == "body" {
			return &res.Children[i]
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
	return elm.DocumentElement.InnerLabel()
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

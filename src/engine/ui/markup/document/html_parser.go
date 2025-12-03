/******************************************************************************/
/* html_parser.go                                                             */
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
	"html/template"
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/engine/systems/events"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/elements"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"log/slog"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"weak"
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

//var Debug = struct {
//	ReloadStylesEvent events.Event
//}{}

type Document struct {
	host             weak.Pointer[engine.Host]
	Elements         []*Element
	TopElements      []*Element
	HeadElements     []*Element
	onWindowResizeId events.Id
	groups           map[string][]*Element
	ids              map[string]*Element
	classElements    map[string][]*Element
	tagElements      map[string][]*Element
	style            rules.StyleSheet
	stylizer         Stylizer
	// TODO:  Should this be here?
	firstInput *ui.Input
	lastInput  *ui.Input
	funcMap    map[string]func(*Element)
	//Debug      struct {
	//	ReloadEventId events.Id
	//}
}

func (d *Document) SetupStyle(style rules.StyleSheet, host *engine.Host, stylizer Stylizer) {
	d.style = style
	d.stylizer = stylizer
	d.host = weak.Make(host)
	d.stylizer.ApplyStyles(d.style, d)
	wd := weak.Make(d)
	d.onWindowResizeId = host.Window.OnResize.Add(func() {
		sd := wd.Value()
		if sd != nil {
			sd.ApplyStyles()
		}
	})
	type documentCleanup struct {
		host weak.Pointer[engine.Host]
		eid  events.Id
	}
	runtime.AddCleanup(d, func(dc documentCleanup) {
		h := dc.host.Value()
		if h != nil {
			h.Window.OnResize.Remove(dc.eid)
		}
	}, documentCleanup{d.host, d.onWindowResizeId})
}

func (h *Document) GetElementById(id string) (*Element, bool) {
	if e, ok := h.ids[id]; ok {
		return e, ok
	} else {
		return nil, ok
	}
}

func (h *Document) GetElementsByGroup(group string) []*Element {
	if e, ok := h.groups[group]; ok {
		return e
	} else {
		return []*Element{}
	}
}

func (h *Document) reloadElementCaches() {
	h.recacheElementIds()
	h.recacheElementTags()
	h.recacheElementGroups()
	h.recacheElementClasses()
}

func (h *Document) recacheElementClasses() {
	clear(h.classElements)
	for i := range h.Elements {
		e := h.Elements[i]
		classList := e.ClassList()
		for _, c := range classList {
			if len(c) == 0 {
				continue
			}
			if m, ok := h.classElements[c]; ok {
				h.classElements[c] = append(m, e)
			} else {
				h.classElements[c] = []*Element{e}
			}
		}
	}
}

func (h *Document) recacheElementGroups() {
	clear(h.groups)
	for i := range h.Elements {
		e := h.Elements[i]
		if group := e.Attribute("group"); group != "" {
			g, ok := h.groups[group]
			if !ok {
				g = make([]*Element, 0, 1)
			}
			g = append(g, e)
			h.groups[group] = g
		}
	}
}

func (h *Document) recacheElementTags() {
	clear(h.tagElements)
	for i := range h.Elements {
		e := h.Elements[i]
		if tag, ok := elements.ElementMap[strings.ToLower(e.Data)]; ok {
			h.tagElement(e, tag.Key())
		}
	}
}

func (h *Document) recacheElementIds() {
	clear(h.ids)
	for i := range h.Elements {
		e := h.Elements[i]
		if id := e.Attribute("id"); id != "" {
			h.ids[id] = e
		}
	}
}

func (h *Document) GetElementsByClass(class string) []*Element {
	if e, ok := h.classElements[class]; ok {
		return e
	} else {
		return []*Element{}
	}
}

func (h *Document) GetElementsByTagName(tag string) []*Element {
	if e, ok := h.tagElements[tag]; ok {
		return e
	} else {
		return []*Element{}
	}
}

func TransformHTML(htmlStr string, withData any) (string, error) {
	tpl, err := template.New("html").Funcs(funcMap).Parse(htmlStr)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	if err := tpl.ExecuteTemplate(&sb, "html", withData); err != nil {
		slog.Error("there was an error in the html template", "error", err)
	}
	return sb.String(), nil
}

func (d *Document) createUIElement(uiMan *ui.Manager, e *Element, parent *ui.Panel) {
	if e.IsSelectOption() {
		return
	}
	appendElement := func(uiElm *ui.UI, panel *ui.Panel) *Element {
		e.UI = uiElm
		e.UIPanel = panel
		d.Elements = append(d.Elements, e)
		parent.AddChild(uiElm)
		return e
	}
	if e.IsText() {
		txt := strings.TrimSpace(e.Data)
		txt = strings.ReplaceAll(txt, "\r", "")
		txt = strings.ReplaceAll(txt, "\n", " ")
		txt = strings.ReplaceAll(txt, "\t", " ")
		txt = klib.ReplaceStringRecursive(txt, "  ", " ")
		label := uiMan.Add().ToLabel()
		label.Init(txt)
		label.SetJustify(rendering.FontJustifyLeft)
		label.SetBaseline(rendering.FontBaselineTop)
		label.SetBGColor(matrix.ColorTransparent())
		appendElement(label.Base(), nil)
	} else if tag, ok := elements.ElementMap[strings.ToLower(e.Data)]; ok {
		panel := uiMan.Add().ToPanel()
		host := uiMan.Host
		if e.IsImage() {
			src := e.Attribute("src")
			var tex *rendering.Texture
			var err error
			var spriteJSON string
			if strings.HasSuffix(src, ".gif") {
				root := strings.TrimSuffix(src, ".gif")
				pngSrc := root + ".png"
				jsonSrc := pngSrc + ".json"
				assets := host.AssetDatabase()
				if !assets.Exists(jsonSrc) {
					slog.Error("failed to find the JSON for the sprite sheet", "png", pngSrc, "json", jsonSrc)
					return
				}
				if tex, err = host.TextureCache().Texture(pngSrc, rendering.TextureFilterLinear); err != nil {
					slog.Error("failed to read the sprite sheet PNG file", "png", pngSrc, "error", err)
					return
				}
				if spriteJSON, err = assets.ReadText(jsonSrc); err != nil {
					slog.Error("failed to read the sprite sheet JSON file", "json", jsonSrc, "error", err)
					return
				}
				assets.Cache(jsonSrc, []byte(spriteJSON))
			} else {
				tex, err = host.TextureCache().Texture(src, rendering.TextureFilterLinear)
			}
			if err != nil {
				slog.Error(err.Error())
				return
			}
			img := panel.Base().ToImage()
			if strings.HasSuffix(src, ".gif") {
				img.InitSpriteSheet(12, tex, spriteJSON)
				img.PlayAnimation()
			} else {
				img.Init(tex)
			}
			panel = (*ui.Panel)(img)
		} else if e.IsInput() {
			inputType := e.Attribute("type")
			switch inputType {
			case "checkbox":
				cb := panel.Base().ToCheckbox()
				cb.Init()
				if e.Attribute("checked") != "" {
					cb.SetChecked(true)
				}
			case "slider":
				slider := panel.Base().ToSlider()
				slider.Init()
				panel.DontFitContent()
				if a := e.Attribute("value"); a != "" {
					if f, err := strconv.ParseFloat(a, 32); err == nil {
						slider.SetValue(float32(f))
					}
				}
			case "text", "number":
				input := panel.Base().ToInput()
				input.Init(e.Attribute("placeholder"))
				input.SetTextWithoutEvent(e.Attribute("value"))
				if d.firstInput == nil {
					d.firstInput = input
				}
				if d.lastInput != nil {
					d.lastInput.SetNextFocusedInput(input)
				}
				d.lastInput = input
				input.SetNextFocusedInput(d.firstInput)
			}
			panel.SetOverflow(ui.OverflowVisible)
		} else if e.IsSelect() {
			sel := panel.Base().ToSelect()
			sel.Init("", []ui.SelectOption{})
			selectStartValue := ""
			if a := e.Attribute("value"); a != "" {
				selectStartValue = a
			}
			for i := range e.Children {
				child := e.Children[i]
				if child.IsSelectOption() {
					childText := ""
					if len(child.Children) > 0 {
						childText = child.Children[0].Data
					}
					val := child.Attribute("value")
					sel.AddOption(childText, val)
					if val == selectStartValue {
						sel.PickOption(i)
					} else if val == "" && childText == selectStartValue {
						sel.PickOption(i)
					}
				}
			}
		} else {
			panel.Init(nil, ui.ElementTypePanel)
			panel.SetOverflow(ui.OverflowVisible)
		}
		entry := appendElement(panel.Base(), panel)
		for i := range e.Children {
			d.createUIElement(uiMan, e.Children[i], panel)
		}
		id := e.Attribute("id")
		group := e.Attribute("group")
		if len(id) > 0 {
			d.ids[id] = entry
			panel.Base().Entity().SetName(id)
		} else {
			panel.Base().Entity().SetName(e.Attribute("name"))
		}
		if len(group) > 0 {
			d.groups[group] = append(d.groups[group], entry)
		}
		classList := e.ClassList()
		for _, c := range classList {
			if len(c) == 0 {
				continue
			}
			if m, ok := d.classElements[c]; ok {
				d.classElements[c] = append(m, entry)
			} else {
				d.classElements[c] = []*Element{entry}
			}
		}
		d.tagElement(entry, tag.Key())
	}
}

func (d *Document) tagElement(elm *Element, tag string) {
	if m, ok := d.tagElements[tag]; ok {
		d.tagElements[tag] = append(m, elm)
	} else {
		d.tagElements[tag] = []*Element{elm}
	}
}

func (d *Document) setupBody(h *Element, uiMan *ui.Manager) *Element {
	body := h.Body()
	bodyPanel := uiMan.Add().ToPanel()
	bodyPanel.Init(nil, ui.ElementTypePanel)
	bodyPanel.DontFitContent()
	bodyPanel.Base().Clean()
	body.UI = bodyPanel.Base()
	body.UIPanel = bodyPanel
	d.Elements = append(d.Elements, body)
	d.tagElements["body"] = []*Element{body}
	bodyClasses := body.ClassList()
	for _, c := range bodyClasses {
		if len(c) == 0 {
			continue
		}
		if m, ok := d.classElements[c]; ok {
			d.classElements[c] = append(m, body)
		} else {
			d.classElements[c] = []*Element{body}
		}
	}
	return body
}

func DocumentFromHTMLString(uiMan *ui.Manager, htmlStr string, withData any, funcMap map[string]func(*Element)) *Document {
	parsed := &Document{
		Elements:      make([]*Element, 0),
		groups:        map[string][]*Element{},
		ids:           map[string]*Element{},
		classElements: map[string][]*Element{},
		tagElements:   map[string][]*Element{},
		HeadElements:  make([]*Element, 0),
		funcMap:       funcMap,
	}
	transformed, err := TransformHTML(htmlStr, withData)
	if err != nil {
		slog.Error("failed to parse the html file", "error", err)
		return parsed
	}
	h := NewHTML(transformed)
	body := parsed.setupBody(h, uiMan)
	bodyPanel := body.UIPanel
	bodyPanel.Base().Entity().SetName("htmlBody")
	for i := range body.Children {
		idx := len(parsed.Elements)
		parsed.createUIElement(uiMan, body.Children[i], bodyPanel)
		if idx < len(parsed.Elements) {
			parsed.TopElements = append(parsed.TopElements, parsed.Elements[idx])
		}
	}
	for i := range parsed.Elements {
		setupEvents(parsed.Elements[i], parsed.funcMap)
	}
	for _, elm := range h.Children[len(h.Children)-1].Children {
		if elm.Data == "head" {
			for _, child := range elm.Children {
				parsed.HeadElements = append(parsed.HeadElements, child)
			}
		}
	}
	return parsed
}

func (d *Document) IsActive() bool {
	anyActive := false
	for i := range d.Elements {
		anyActive = anyActive || d.Elements[i].UI.Entity().IsActive()
	}
	return anyActive
}

func (d *Document) Activate() {
	for i := range d.Elements {
		d.Elements[i].UI.Entity().Activate()
	}
}

func (d *Document) Deactivate() {
	for i := range d.Elements {
		d.Elements[i].UI.Entity().Deactivate()
	}
}

func (d *Document) Destroy() {
	for _, e := range d.TopElements {
		if e.Parent.Value() != nil {
			for i, c := range e.Parent.Value().Children {
				if c == e {
					parent := e.Parent.Value()
					parent.Children = slices.Delete(parent.Children, i, i+1)
					parent.UI.SetDirty(ui.DirtyTypeLayout)
					break
				}
			}
		}
	}
	for _, e := range d.Elements {
		e.UI.Entity().Destroy()
	}
	clear(d.funcMap)
	*d = Document{}
	//if build.Debug {
	//	Debug.ReloadStylesEvent.Remove(d.Debug.ReloadEventId)
	//}
}

func (d *Document) MarkDirty() {
	for i := range d.Elements {
		d.Elements[i].UI.SetDirty(ui.DirtyTypeGenerated)
	}
}

func (d *Document) Clean() {
	if len(d.Elements) > 0 {
		d.Elements[0].UI.Clean()
	}
}

func (d *Document) indexElement(elm *Element) {
	d.Elements = append(d.Elements, elm)
	if id := elm.Attribute("id"); id != "" {
		d.ids[id] = elm
	}
	if group := elm.Attribute("group"); group != "" {
		d.groups[group] = append(d.groups[group], elm)
	}
	if tag, ok := elements.ElementMap[strings.ToLower(elm.Data)]; ok {
		d.tagElement(elm, tag.Key())
	}
	for _, c := range elm.ClassList() {
		if len(c) == 0 {
			continue
		}
		if m, ok := d.classElements[c]; ok {
			d.classElements[c] = append(m, elm)
		} else {
			d.classElements[c] = []*Element{elm}
		}
	}
}

func (d *Document) removeIndexedElement(elm *Element) {
	for i, e := range d.Elements {
		if e == elm {
			d.Elements = klib.RemoveUnordered(d.Elements, i)
			break
		}
	}
	delete(d.ids, elm.Attribute("id"))
	if group := elm.Attribute("group"); group != "" {
		for i := range d.groups[group] {
			if d.groups[group][i] == elm {
				// TODO:  Is sorted remove required here?
				d.groups[group] = slices.Delete(d.groups[group], i, i+1)
				break
			}
		}
	}
	if tag, ok := elements.ElementMap[strings.ToLower(elm.Data)]; ok {
		if _, ok := d.tagElements[tag.Key()]; ok {
			for i := range d.tagElements[tag.Key()] {
				if d.tagElements[tag.Key()][i] == elm {
					// TODO:  Is sorted remove required here?
					d.tagElements[tag.Key()] = slices.Delete(d.tagElements[tag.Key()], i, i+1)
					break
				}
			}
		}
	}
	for _, c := range elm.ClassList() {
		if len(c) == 0 {
			continue
		}
		if _, ok := d.classElements[c]; !ok {
			continue
		}
		for i := range d.classElements[c] {
			if d.classElements[c][i] == elm {
				// TODO:  Is sorted remove required here?
				d.classElements[c] = slices.Delete(d.classElements[c], i, i+1)
				break
			}
		}
	}
}

func (d *Document) AddChildElement(parent *Element, elm *Element) {
	parent.Children = append(parent.Children, elm)
	d.indexElement(elm)
}

// RemoveElement removes the specified element from the document by first
// recursively removing all child elements, then removing the element from
// its parent's children list, destroying the UI entity, and updating the
// document's element caches. Finally, it reapplies all document styles.
//
// Parameters:
//   - elm: pointer to the Element to be removed from the document
//
// The function performs the following operations:
// 1. Recursively removes all child elements starting from the last child
// 2. Removes the element from its parent's Children list if it has a parent
// 3. Destroys the UI entity associated with the element
// 4. Updates the parent's layout dirty flag
// 5. Removes the element from document's indexed elements
// 6. Applies all document styles to reflect changes
func (d *Document) RemoveElement(elm *Element) {
	d.RemoveElementWithoutApplyStyles(elm)
	d.stylizer.ApplyStyles(d.style, d)
}

func (d *Document) RemoveElementWithoutApplyStyles(elm *Element) {
	for i := len(elm.Children) - 1; i >= 0; i-- {
		d.RemoveElement(elm.Children[i])
	}
	if elm.Parent.Value() != nil {
		for i, c := range elm.Parent.Value().Children {
			if c == elm {
				c.UI.Entity().Destroy()
				parent := elm.Parent.Value()
				parent.Children = slices.Delete(parent.Children, i, i+1)
				parent.UI.SetDirty(ui.DirtyTypeLayout)
				break
			}
		}
	}
	d.removeIndexedElement(elm)
}

// SetElementClassesWithoutApply updates the class list of the given element
// without applying styles. It removes the element from its previous class
// lists, sets the new classes, and updates the document's classElements map
// to reflect the new class assignments.
//
// Parameters:
//   - elm: pointer to the Element whose classes will be updated
//   - classes: variadic string parameters representing the new class names
//
// The function performs the following operations:
// 1. Sorts both the input classes and existing element classes
// 2. Returns early if classes are identical
// 3. Removes element from previous class lists in classElements map
// 4. Sets the new class list on the element
// 5. Adds element to the appropriate class lists in classElements map
func (d *Document) SetElementClassesWithoutApply(elm *Element, classes ...string) {
	elmClasses := elm.ClassList()
	sort.Strings(classes)
	sort.Strings(elmClasses)
	if klib.SlicesAreTheSame(classes, elmClasses) {
		return
	}
	for i := range elmClasses {
		elms := d.classElements[elmClasses[i]]
		target := slices.Index(elms, elm)
		if target >= 0 {
			d.classElements[elmClasses[i]] = slices.Delete(elms, target, target+1)
		}
	}
	elm.SetClasses(classes...)
	for _, c := range classes {
		if len(c) == 0 {
			continue
		}
		if m, ok := d.classElements[c]; ok {
			d.classElements[c] = append(m, elm)
		} else {
			d.classElements[c] = []*Element{elm}
		}
	}
}

// SetElementClasses updates the class list of the given element and applies
// style changes to the entire document. It calls SetElementClassesWithoutApply
// to update classes, then applies all styles in the document.
//
// Parameters:
//   - elm: pointer to the Element whose classes will be updated
//   - classes: variadic string parameters representing the new class names
func (d *Document) SetElementClasses(elm *Element, classes ...string) {
	d.SetElementClassesWithoutApply(elm, classes...)
	elm.UI.Layout().ClearStyles()
	d.stylizer.ApplyStyles(d.style, d)
}

// ApplyStyles will go through and apply styles to all elements within the
// document. This is typically used after [SetElementClassesWithoutApply]. The
// typical flow is to call [SetElementClassesWithoutApply] in a loop to change
// styles of many elements at the same time, then apply styles after.
func (d *Document) ApplyStyles() { d.stylizer.ApplyStyles(d.style, d) }

// DuplicateElement will create a duplicate of a given element, nesting it under
// the same parent as the given element (at the end). If you wish to just
// duplicate an element and use one of the Insert functions, then use
// Element.Clone followed by an Insert function instead
func (d *Document) DuplicateElement(elm *Element) *Element {
	cpy := elm.Clone(elm.Parent.Value())
	if elm.Attribute("id") != "" {
		cpy.SetAttribute("id", "")
	}
	d.appendElement(cpy)
	d.stylizer.ApplyStyles(d.style, d)
	return cpy
}

func (d *Document) DuplicateElementToParent(elm, parent *Element) *Element {
	cpy := elm.Clone(elm.Parent.Value())
	if elm.Attribute("id") != "" {
		cpy.SetAttribute("id", "")
	}
	d.appendElement(cpy)
	d.ChangeElementParentWithoutApply(cpy, parent)
	d.stylizer.ApplyStyles(d.style, d)
	return cpy
}

// DuplicateElementRepeat is the same as [DuplicateElement], but will duplicate
// the element a specified number of times. This is an optimization to avoid
// calling [ApplyStyles] on each duplicated element and instead call it at the
// end, after all copies are created.
func (d *Document) DuplicateElementRepeat(elm *Element, count int) []*Element {
	elms := d.DuplicateElementRepeatWithoutApplyStyles(elm, count)
	d.stylizer.ApplyStyles(d.style, d)
	return elms
}

func (d *Document) DuplicateElementRepeatWithoutApplyStyles(elm *Element, count int) []*Element {
	elms := make([]*Element, count)
	for i := range count {
		elms[i] = elm.Clone(elm.Parent.Value())
		d.appendElement(elms[i])
	}
	d.stylizer.ApplyStyles(d.style, d)
	return elms
}

func (d *Document) SetElementIdWithoutApplyStyles(elm *Element, id string) {
	defer tracing.NewRegion("Document.SetElementIdWithoutApplyStyles").End()
	currentId := elm.Attribute("id")
	if currentId != "" {
		delete(d.ids, currentId)
	}
	elm.SetAttribute("id", id)
	d.ids[id] = elm
}

func (d *Document) SetElementId(elm *Element, id string) {
	defer tracing.NewRegion("Document.SetElementId").End()
	d.SetElementIdWithoutApplyStyles(elm, id)
	d.ApplyStyles()
}

func (d *Document) ChangeElementParent(child, parent *Element) {
	d.ChangeElementParentWithoutApply(child, parent)
	d.ApplyStyles()
}

func (d *Document) ChangeElementParentWithoutApply(child, parent *Element) {
	// Check to see if anywhere in newParent's hierarchy is this entity
	// if so, then set it's direct descendant to take this entity's parent.
	{
		p := parent
		for p != nil {
			if p.Parent.Value() == child {
				d.ChangeElementParentWithoutApply(p, child.Parent.Value())
				break
			}
			p = p.Parent.Value()
		}
	}
	current := child.Parent.Value()
	if current != nil {
		current.Children = klib.SlicesRemoveElement(current.Children, child)
		child.Parent = weak.Make[Element](nil)
		current.UI.ToPanel().RemoveChild(child.UI)
	}
	parent.Children = append(parent.Children, child)
	child.Parent = weak.Make(parent)
	parent.UIPanel.AddChild(child.UI)
}

func (d *Document) appendElement(elm *Element) {
	var addChildren func(target *Element)
	addChildren = func(target *Element) {
		d.Elements = append(d.Elements, target)
		setupEvents(target, d.funcMap)
		for i := range target.Children {
			addChildren(target.Children[i])
		}
	}
	addChildren(elm)
	d.reloadElementCaches()
}

func (d *Document) isElementInDocument(elm *Element) bool {
	for i := range d.Elements {
		if d.Elements[i] == elm {
			return true
		}
	}
	return false
}

// InsertElementBefore inserts the given element elm into the document before
// the specified element before. It handles the removal of elm from its current
// parent if it has one, and then inserts it into the children of before's
// parent at the correct position. The function ensures that elm is properly
// added to the document's element list and updates all necessary caches and
// styles after the insertion.
//
// Parameters:
//   - elm: the element to be inserted
//   - before: the element before which elm should be inserted
//
// Preconditions:
//   - Both elm and before must not be nil (enforced by debug.Ensure)
func (d *Document) InsertElementBefore(elm *Element, before *Element) {
	debug.Ensure(elm != nil)
	debug.Ensure(before != nil)
	parent := before.Parent.Value()
	idx := -1
	if parent != nil {
		idx = parent.IndexOfChild(before)
	}
	d.insertElementAt(elm, before.Parent.Value(), idx)
}

// InsertElementAfter inserts the given element elm into the document after
// the specified element after. It handles the removal of elm from its current
// parent if it has one, and then inserts it into the children of after's
// parent at the correct position. The function ensures that elm is properly
// added to the document's element list and updates all necessary caches and
// styles after the insertion.
//
// Parameters:
//   - elm: the element to be inserted
//   - after: the element after which elm should be inserted
//
// Preconditions:
//   - Both elm and after must not be nil (enforced by debug.Ensure)
func (d *Document) InsertElementAfter(elm *Element, after *Element) {
	debug.Ensure(elm != nil)
	debug.Ensure(after != nil)
	parent := after.Parent.Value()
	idx := -1
	if parent != nil {
		idx = parent.IndexOfChild(after) + 1
	}
	d.insertElementAt(elm, after.Parent.Value(), idx)
}

func (d *Document) insertElementAt(elm *Element, parent *Element, index int) {
	fromParent := elm.Parent.Value()
	if fromParent != nil {
		idx := fromParent.IndexOfChild(elm)
		fromParent.Children = slices.Delete(fromParent.Children, idx, idx+1)
		elm.Parent = weak.Make[Element](nil)
	}
	if parent != nil {
		parent.Children = slices.Insert(parent.Children, index, elm)
		elm.Parent = weak.Make(parent)
		parent.UI.ToPanel().AddChild(elm.UI)
	}
	if !d.isElementInDocument(elm) {
		d.appendElement(elm)
	}
	d.stylizer.ApplyStyles(d.style, d)
}

/******************************************************************************/
/* html_element_layout_stylizer.go                                            */
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
	"errors"
	"kaiju/engine"
	"kaiju/engine/systems/events"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/klib"
	"slices"
	"weak"
)

var (
	LinkedPropertyMap    map[string]CSSProperty
	selfDestructingRules = map[string]struct{}{
		"visibility": {},
		"display":    {},
	}
)

type CSSProperty interface {
	Key() string
	Process(panel *ui.Panel, elm *Element, values []rules.PropertyValue, host *engine.Host) error
	Sort() int
	Preprocess(values []rules.PropertyValue, rules []rules.Rule) ([]rules.PropertyValue, []rules.Rule)
}

type ElementLayoutStylizer struct {
	element         weak.Pointer[Element]
	styleRules      []rules.Rule
	activateEvtId   events.Id
	deactivateEvtId events.Id
	hoverEvtId      events.Id
	hoverExitEvtId  events.Id
	activeEvt       struct {
		enterId events.Id
		downId  events.Id
		upId    events.Id
		exitId  events.Id
	}
	currentInvoke rules.RuleInvoke
}

func (s *ElementLayoutStylizer) ClearRules() {
	s.styleRules = s.styleRules[:0]
	e := s.element.Value()
	if e == nil {
		return
	}
	e.UI.RemoveEvent(ui.EventTypeEnter, s.hoverEvtId)
	e.UI.RemoveEvent(ui.EventTypeExit, s.hoverExitEvtId)
	e.UI.RemoveEvent(ui.EventTypeEnter, s.activeEvt.enterId)
	e.UI.RemoveEvent(ui.EventTypeExit, s.activeEvt.exitId)
	e.UI.RemoveEvent(ui.EventTypeDown, s.activeEvt.downId)
	e.UI.RemoveEvent(ui.EventTypeUp, s.activeEvt.upId)
	if !e.UI.IsType(ui.ElementTypeLabel) {
		l := e.UI.Layout()
		e.UI.ToPanel().FitContent()
		l.SetInnerOffset(0, 0, 0, 0)
		l.SetLocalInnerOffset(0, 0, 0, 0)
		l.SetMargin(0, 0, 0, 0)
	}
	entity := e.UI.Entity()
	entity.OnActivate.Remove(s.activateEvtId)
	entity.OnDeactivate.Remove(s.deactivateEvtId)
	s.hoverEvtId = 0
	s.hoverExitEvtId = 0
	s.activeEvt.enterId = 0
	s.activeEvt.exitId = 0
	s.activeEvt.downId = 0
	s.activeEvt.upId = 0
	s.activateEvtId = 0
	s.deactivateEvtId = 0
}

func (s *ElementLayoutStylizer) AddRule(rule rules.Rule) {
	elm := s.element.Value()
	if elm == nil {
		return
	}
	_, rule.SelfDestruct = selfDestructingRules[rule.Property]
	rule.Sort = LinkedPropertyMap[rule.Property].Sort()
	s.styleRules = append(s.styleRules, rule)
	switch rule.Invocation {
	case rules.RuleInvokeHover:
		if s.hoverEvtId == 0 {
			s.hoverEvtId = elm.UI.AddEvent(ui.EventTypeEnter, func() {
				s.currentInvoke = rules.RuleInvokeHover
				elm.UI.SetDirty(ui.DirtyTypeGenerated)
			})
			s.hoverExitEvtId = elm.UI.AddEvent(ui.EventTypeExit, func() {
				s.currentInvoke = rules.RuleInvokeImmediate
				elm.UI.SetDirty(ui.DirtyTypeGenerated)
			})
		}
	case rules.RuleInvokeActive:
		if s.activeEvt.enterId == 0 {
			s.activeEvt.enterId = elm.UI.AddEvent(ui.EventTypeEnter, func() {
				if elm.UI.IsDown() {
					s.currentInvoke = rules.RuleInvokeActive
					elm.UI.SetDirty(ui.DirtyTypeGenerated)
				}
			})
			s.activeEvt.downId = elm.UI.AddEvent(ui.EventTypeDown, func() {
				s.currentInvoke = rules.RuleInvokeActive
				elm.UI.SetDirty(ui.DirtyTypeGenerated)
			})
			s.activeEvt.upId = elm.UI.AddEvent(ui.EventTypeUp, func() {
				s.currentInvoke = rules.RuleInvokeHover
				elm.UI.SetDirty(ui.DirtyTypeGenerated)
			})
			s.activeEvt.exitId = elm.UI.AddEvent(ui.EventTypeExit, func() {
				s.currentInvoke = rules.RuleInvokeImmediate
				elm.UI.SetDirty(ui.DirtyTypeGenerated)
			})
			elm.UIEventIds[ui.EventTypeEnter] = append(elm.UIEventIds[ui.EventTypeEnter], s.activeEvt.enterId)
			elm.UIEventIds[ui.EventTypeDown] = append(elm.UIEventIds[ui.EventTypeDown], s.activeEvt.downId)
			elm.UIEventIds[ui.EventTypeUp] = append(elm.UIEventIds[ui.EventTypeUp], s.activeEvt.upId)
			elm.UIEventIds[ui.EventTypeExit] = append(elm.UIEventIds[ui.EventTypeExit], s.activeEvt.exitId)
		}
	}
}

func (s *ElementLayoutStylizer) ProcessStyle(layout *ui.Layout) []error {
	return s.processRules(layout, s.currentInvoke)
}

func (s *ElementLayoutStylizer) processRules(layout *ui.Layout, invoke rules.RuleInvoke) []error {
	problems := make([]error, 0)
	elm := s.element.Value()
	if elm == nil {
		return []error{errors.New("missing element when processing rules")}
	}
	host := elm.UI.Host()
	a := make([]rules.Rule, 0, len(s.styleRules))
	b := make([]rules.Rule, 0, len(s.styleRules))
	for i := 0; i < len(s.styleRules); i++ {
		if s.currentInvoke != rules.RuleInvokeImmediate && s.styleRules[i].Invocation == rules.RuleInvokeImmediate {
			a = append(a, s.styleRules[i])
		} else if s.styleRules[i].Invocation == s.currentInvoke {
			b = append(b, s.styleRules[i])
		}
		if s.styleRules[i].SelfDestruct {
			s.styleRules = slices.Delete(s.styleRules, i, i+1)
			i--
		}
	}
	for j := 0; j < len(a); j++ {
		for i := range b {
			if a[j].Property == b[i].Property {
				a = klib.RemoveUnordered(a, j)
				j--
				break
			}
		}
	}
	all := append(a, b...)
	// Look ahead to see if any upcoming properties can be merged
	for i := 0; i < len(all); i++ {
		if p, ok := LinkedPropertyMap[all[i].Property]; ok {
			subRules := all[i:]
			all[i].Values, subRules = p.Preprocess(all[i].Values, subRules)
			for j := range subRules {
				all[i+j] = subRules[j]
			}
			all = all[:i+len(subRules)]
		}
	}
	slices.SortFunc(all, func(x, y rules.Rule) int { return x.Sort - y.Sort })
	for i := range all {
		if len(all[i].Values) == 1 && all[i].Values[0].Str == "revert" {
			continue
		}
		if p, ok := LinkedPropertyMap[all[i].Property]; ok {
			if err := p.Process(layout.Ui().ToPanel(), elm, all[i].Values, host); err != nil {
				problems = append(problems, err)
			}
		}
	}
	if elm.UI.Entity().Name() == "openProjectBtn" {
		println("...")
	}
	return problems
}

func (s *ElementLayoutStylizer) clone(newElm *Element) ElementLayoutStylizer {
	out := ElementLayoutStylizer{
		element: weak.Make(newElm),
	}
	for i := range s.styleRules {
		out.AddRule(s.styleRules[i].Clone())
	}
	newElm.UI.Layout().Stylizer = s
	return out
}

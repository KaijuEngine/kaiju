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

package document

import (
	"errors"
	"slices"
	"weak"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/klib"
)

var (
	LinkedPropertyMap    map[string]CSSProperty
	selfDestructingRules = map[string]struct{}{
		"visibility": {},
		"display":    {},
	}
	panelOnlyProperties = map[string]struct{}{
		"width":        {},
		"height":       {},
		"min-width":    {},
		"max-width":    {},
		"min-height":   {},
		"max-height":   {},
		"aspect-ratio": {},
		"box-sizing":   {},
	}
)

func isPanelOnlyProperty(property string) bool {
	_, ok := panelOnlyProperties[property]
	return ok
}

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
	focusEvt        struct {
		clickId events.Id
		missId  events.Id
	}
	activeEvt struct {
		enterId events.Id
		downId  events.Id
		upId    events.Id
		exitId  events.Id
	}
	currentState     rules.RuleInvoke
	interestedStates rules.RuleInvoke
}

func (s *ElementLayoutStylizer) HasRule(rule string) bool {
	for i := range s.styleRules {
		if s.styleRules[i].Property == rule {
			return true
		}
	}
	return false
}

func (s *ElementLayoutStylizer) ClearRules() {
	s.styleRules = s.styleRules[:0]
	e := s.element.Value()
	if e == nil {
		return
	}
	e.UI.RemoveEvent(ui.EventTypeEnter, s.hoverEvtId)
	e.UI.RemoveEvent(ui.EventTypeExit, s.hoverExitEvtId)
	e.UI.RemoveEvent(ui.EventTypeClick, s.focusEvt.clickId)
	e.UI.RemoveEvent(ui.EventTypeMiss, s.focusEvt.missId)
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
	s.focusEvt.clickId = 0
	s.focusEvt.missId = 0
	s.activeEvt.enterId = 0
	s.activeEvt.exitId = 0
	s.activeEvt.downId = 0
	s.activeEvt.upId = 0
	s.activateEvtId = 0
	s.deactivateEvtId = 0
	s.interestedStates = rules.RuleInvokeImmediate
}

func (s *ElementLayoutStylizer) AddRule(rule rules.Rule) {
	elm := s.element.Value()
	if elm == nil {
		return
	}
	_, rule.SelfDestruct = selfDestructingRules[rule.Property]
	p, ok := LinkedPropertyMap[rule.Property]
	if !ok {
		return
	}
	rule.Sort = p.Sort()
	s.styleRules = append(s.styleRules, rule)
	s.interestedStates = s.interestedStates.With(rule.Invocation)
	if rule.Invocation&rules.RuleInvokeHover != 0 {
		if s.hoverEvtId == 0 {
			s.hoverEvtId = elm.UI.AddEvent(ui.EventTypeEnter, func() {
				s.setState(rules.RuleInvokeHover, true)
			})
			s.hoverExitEvtId = elm.UI.AddEvent(ui.EventTypeExit, func() {
				s.setState(rules.RuleInvokeHover, false)
			})
		}
	}
	if rule.Invocation&rules.RuleInvokeFocus != 0 {
		if s.focusEvt.clickId == 0 {
			s.focusEvt.clickId = elm.UI.AddEvent(ui.EventTypeClick, func() {
				s.setState(rules.RuleInvokeFocus, true)
			})
			s.focusEvt.missId = elm.UI.AddEvent(ui.EventTypeMiss, func() {
				s.setState(rules.RuleInvokeFocus, false)
			})
		}
	}
	if rule.Invocation&rules.RuleInvokeActive != 0 {
		if s.activeEvt.enterId == 0 {
			s.activeEvt.enterId = elm.UI.AddEvent(ui.EventTypeEnter, func() {
				if elm.UI.IsDown() {
					s.setState(rules.RuleInvokeActive, true)
				}
			})
			s.activeEvt.downId = elm.UI.AddEvent(ui.EventTypeDown, func() {
				s.setState(rules.RuleInvokeActive, true)
			})
			s.activeEvt.upId = elm.UI.AddEvent(ui.EventTypeUp, func() {
				s.setState(rules.RuleInvokeActive, false)
			})
			s.activeEvt.exitId = elm.UI.AddEvent(ui.EventTypeExit, func() {
				s.setState(rules.RuleInvokeActive, false)
			})
			elm.UIEventIds[ui.EventTypeEnter] = append(elm.UIEventIds[ui.EventTypeEnter], s.activeEvt.enterId)
			elm.UIEventIds[ui.EventTypeDown] = append(elm.UIEventIds[ui.EventTypeDown], s.activeEvt.downId)
			elm.UIEventIds[ui.EventTypeUp] = append(elm.UIEventIds[ui.EventTypeUp], s.activeEvt.upId)
			elm.UIEventIds[ui.EventTypeExit] = append(elm.UIEventIds[ui.EventTypeExit], s.activeEvt.exitId)
		}
	}
	s.syncValidationState()
}

func (s *ElementLayoutStylizer) setState(state rules.RuleInvoke, enabled bool) {
	if s.interestedStates&state == 0 {
		return
	}
	elm := s.element.Value()
	if elm == nil {
		return
	}
	next := s.currentState
	if enabled {
		next = next.With(state)
	} else {
		next &^= state
	}
	if next != s.currentState {
		s.currentState = next
		elm.UI.SetDirty(ui.DirtyTypeGenerated)
	}
}

func (s *ElementLayoutStylizer) syncValidationState() {
	if s.interestedStates&(rules.RuleInvokeInvalid|rules.RuleInvokeValid) == 0 {
		return
	}
	elm := s.element.Value()
	if elm == nil || !elm.UI.IsType(ui.ElementTypeInput) {
		return
	}
	if elm.UI.ToInput().IsValid() {
		s.currentState &^= rules.RuleInvokeInvalid
		s.currentState = s.currentState.With(rules.RuleInvokeValid)
	} else {
		s.currentState &^= rules.RuleInvokeValid
		s.currentState = s.currentState.With(rules.RuleInvokeInvalid)
	}
}

func (s *ElementLayoutStylizer) ProcessStyle(layout *ui.Layout) []error {
	return s.processRules(layout)
}

func (s *ElementLayoutStylizer) processRules(layout *ui.Layout) []error {
	problems := make([]error, 0)
	elm := s.element.Value()
	if elm == nil {
		return []error{errors.New("missing element when processing rules")}
	}
	s.syncValidationState()
	host := elm.UI.Host()
	a := make([]rules.Rule, 0, len(s.styleRules))
	b := make([]rules.Rule, 0, len(s.styleRules))
	for i := 0; i < len(s.styleRules); i++ {
		if s.currentState != rules.RuleInvokeImmediate && s.styleRules[i].Invocation == rules.RuleInvokeImmediate {
			a = append(a, s.styleRules[i])
		} else if s.styleRules[i].Invocation.Matches(s.currentState) {
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
	slices.SortStableFunc(all, func(x, y rules.Rule) int { return x.Sort - y.Sort })
	isLabel := layout.Ui().IsType(ui.ElementTypeLabel)
	for i := range all {
		if isLabel && isPanelOnlyProperty(all[i].Property) {
			continue
		}
		if p, ok := LinkedPropertyMap[all[i].Property]; ok {
			if err := p.Process(layout.Ui().ToPanel(), elm, all[i].Values, host); err != nil {
				problems = append(problems, err)
			}
		}
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

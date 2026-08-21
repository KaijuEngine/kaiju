/******************************************************************************/
/* html_element_layout_stylizer.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package document

import (
	"errors"
	"slices"
	"strings"
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
	StyleImpact() StyleImpact
}

type CSSPropertyResetter interface {
	Reset(panel *ui.Panel, elm *Element, host *engine.Host) error
}

type StyleImpact uint8

const (
	StyleImpactLayout StyleImpact = iota
	StyleImpactPaint
)

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
		focusId events.Id
		blurId  events.Id
	}
	activeEvt struct {
		enterId events.Id
		downId  events.Id
		upId    events.Id
		exitId  events.Id
	}
	currentState     rules.RuleInvoke
	interestedStates rules.RuleInvoke
	appliedRules     []rules.Rule
	pendingLayout    bool
	pendingPaint     bool
}

func (s *ElementLayoutStylizer) HasRule(rule string) bool {
	for i := range s.styleRules {
		if s.styleRules[i].Property == rule {
			return true
		}
	}
	return false
}

func (s *ElementLayoutStylizer) clearRuleBindings() {
	e := s.element.Value()
	if e == nil {
		return
	}
	e.UI.RemoveEvent(ui.EventTypeEnter, s.hoverEvtId)
	e.UI.RemoveEvent(ui.EventTypeExit, s.hoverExitEvtId)
	e.UI.RemoveEvent(ui.EventTypeClick, s.focusEvt.clickId)
	e.UI.RemoveEvent(ui.EventTypeMiss, s.focusEvt.missId)
	e.UI.RemoveEvent(ui.EventTypeFocus, s.focusEvt.focusId)
	e.UI.RemoveEvent(ui.EventTypeBlur, s.focusEvt.blurId)
	e.UI.RemoveEvent(ui.EventTypeEnter, s.activeEvt.enterId)
	e.UI.RemoveEvent(ui.EventTypeExit, s.activeEvt.exitId)
	e.UI.RemoveEvent(ui.EventTypeDown, s.activeEvt.downId)
	e.UI.RemoveEvent(ui.EventTypeUp, s.activeEvt.upId)
	for evtType := range e.UIEventIds {
		for _, evtId := range e.UIEventIds[evtType] {
			e.UI.RemoveEvent(evtType, evtId)
		}
		e.UIEventIds[evtType] = e.UIEventIds[evtType][:0]
	}
	entity := e.UI.Entity()
	entity.OnActivate.Remove(s.activateEvtId)
	entity.OnDeactivate.Remove(s.deactivateEvtId)
	s.hoverEvtId = 0
	s.hoverExitEvtId = 0
	s.focusEvt.clickId = 0
	s.focusEvt.missId = 0
	s.focusEvt.focusId = 0
	s.focusEvt.blurId = 0
	s.activeEvt.enterId = 0
	s.activeEvt.exitId = 0
	s.activeEvt.downId = 0
	s.activeEvt.upId = 0
	s.activateEvtId = 0
	s.deactivateEvtId = 0
	s.interestedStates = rules.RuleInvokeImmediate
}

// ClearRules queues removal of all CSS rules. It intentionally does not clear
// live layout state; that happens only if the computed layout rules differ.
func (s *ElementLayoutStylizer) ClearRules() {
	s.ReplaceRules(nil)
}

// ReplaceRules installs the next cascaded rule set and compares it with the
// last successfully applied computed rules. No UI dirty flag is set here.
func (s *ElementLayoutStylizer) ReplaceRules(next []rules.Rule) {
	if sourceRuleListsEqual(s.styleRules, next) {
		s.queueComputedDiff(s.computedRules())
		return
	}
	s.clearRuleBindings()
	s.styleRules = s.styleRules[:0]
	for i := range next {
		s.AddRule(next[i].Clone())
	}
	s.queueComputedDiff(s.computedRules())
}

func sourceRuleListsEqual(a, b []rules.Rule) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Property != b[i].Property || a[i].Invocation != b[i].Invocation ||
			!propertyValuesEqual(a[i].Values, b[i].Values) {
			return false
		}
	}
	return true
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
			s.focusEvt.focusId = elm.UI.AddEvent(ui.EventTypeFocus, func() {
				s.setState(rules.RuleInvokeFocus, true)
			})
			s.focusEvt.blurId = elm.UI.AddEvent(ui.EventTypeBlur, func() {
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
		s.queueComputedDiff(s.computedRules())
	}
}

func (s *ElementLayoutStylizer) syncValidationState() {
	if s.interestedStates&(rules.RuleInvokeInvalid|rules.RuleInvokeValid) == 0 {
		return
	}
	elm := s.element.Value()
	if elm == nil {
		return
	}
	valid := false
	switch elm.UI.Type() {
	case ui.ElementTypeInput:
		valid = elm.UI.ToInput().IsValid()
	case ui.ElementTypeTextArea:
		valid = elm.UI.ToTextArea().IsValid()
	default:
		return
	}
	if valid {
		s.currentState &^= rules.RuleInvokeInvalid
		s.currentState = s.currentState.With(rules.RuleInvokeValid)
	} else {
		s.currentState &^= rules.RuleInvokeValid
		s.currentState = s.currentState.With(rules.RuleInvokeInvalid)
	}
}

func (s *ElementLayoutStylizer) ProcessStyle(layout *ui.Layout) []error {
	computed := s.computedRules()
	problems := s.processRuleList(layout, computed)
	s.appliedRules = rules.CloneRules(computed)
	s.pendingLayout = false
	s.pendingPaint = false
	return problems
}

func (s *ElementLayoutStylizer) HasPendingStyle() bool {
	return s.pendingLayout || s.pendingPaint
}

func (s *ElementLayoutStylizer) ProcessPendingStyle(layout *ui.Layout) []error {
	computed := s.computedRules()
	s.queueComputedDiff(computed)
	if !s.HasPendingStyle() {
		return nil
	}
	problems := make([]error, 0)
	if s.pendingLayout {
		problems = append(problems, s.resetRemovedRules(layout, computed)...)
		if !layout.Ui().IsType(ui.ElementTypeLabel) {
			layout.Ui().ToPanel().ClearLayoutStyles()
		}
		problems = append(problems, s.processRuleList(layout, computed)...)
	} else {
		changed, removed := changedPaintRules(s.appliedRules, computed)
		elm := s.element.Value()
		if elm != nil {
			for _, property := range removed {
				if p, ok := LinkedPropertyMap[property]; ok {
					if resetter, ok := p.(CSSPropertyResetter); ok {
						if err := resetter.Reset(layout.Ui().ToPanel(), elm, elm.UI.Host()); err != nil {
							problems = append(problems, err)
						}
					}
				}
			}
		}
		problems = append(problems, s.processRuleList(layout, changed)...)
	}
	s.appliedRules = rules.CloneRules(computed)
	s.pendingLayout = false
	s.pendingPaint = false
	return problems
}

func (s *ElementLayoutStylizer) resetRemovedRules(layout *ui.Layout, computed []rules.Rule) []error {
	nextProperties := make(map[string]struct{}, len(computed))
	for i := range computed {
		nextProperties[computed[i].Property] = struct{}{}
	}
	elm := s.element.Value()
	if elm == nil {
		return []error{errors.New("missing element when resetting rules")}
	}
	problems := make([]error, 0)
	for i := range s.appliedRules {
		property := s.appliedRules[i].Property
		if _, stillPresent := nextProperties[property]; stillPresent {
			continue
		}
		if p, ok := LinkedPropertyMap[property]; ok {
			if resetter, ok := p.(CSSPropertyResetter); ok {
				if err := resetter.Reset(layout.Ui().ToPanel(), elm, elm.UI.Host()); err != nil {
					problems = append(problems, err)
				}
			}
		}
	}
	return problems
}

func (s *ElementLayoutStylizer) computedRules() []rules.Rule {
	elm := s.element.Value()
	if elm == nil {
		return nil
	}
	s.syncValidationState()
	a := make([]rules.Rule, 0, len(s.styleRules))
	b := make([]rules.Rule, 0, len(s.styleRules))
	for i := range s.styleRules {
		if s.currentState != rules.RuleInvokeImmediate && s.styleRules[i].Invocation == rules.RuleInvokeImmediate {
			a = append(a, s.styleRules[i].Clone())
		} else if s.styleRules[i].Invocation.Matches(s.currentState) {
			b = append(b, s.styleRules[i].Clone())
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
	for i := range all {
		all[i].Invocation = rules.RuleInvokeImmediate
		all[i].SelfDestruct = false
	}
	return all
}

func (s *ElementLayoutStylizer) processRuleList(layout *ui.Layout, all []rules.Rule) []error {
	problems := make([]error, 0)
	elm := s.element.Value()
	if elm == nil {
		return []error{errors.New("missing element when processing rules")}
	}
	host := elm.UI.Host()
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

func (s *ElementLayoutStylizer) queueComputedDiff(next []rules.Rule) {
	layoutChanged, paintChanged := computedRuleDiff(s.appliedRules, next)
	s.pendingLayout = layoutChanged
	s.pendingPaint = paintChanged
}

func computedRuleDiff(previous, next []rules.Rule) (layoutChanged, paintChanged bool) {
	previousByProperty := make(map[string]rules.Rule, len(previous))
	nextByProperty := make(map[string]rules.Rule, len(next))
	for i := range previous {
		previousByProperty[previous[i].Property] = previous[i]
	}
	for i := range next {
		nextByProperty[next[i].Property] = next[i]
	}
	allProperties := make(map[string]struct{}, len(previousByProperty)+len(nextByProperty))
	for property := range previousByProperty {
		allProperties[property] = struct{}{}
	}
	for property := range nextByProperty {
		allProperties[property] = struct{}{}
	}
	for property := range allProperties {
		before, hadBefore := previousByProperty[property]
		after, hasAfter := nextByProperty[property]
		if hadBefore && hasAfter && ruleComputedEqual(before, after) {
			continue
		}
		impact := propertyChangeImpact(property, before, hadBefore, after, hasAfter)
		if impact == StyleImpactLayout {
			layoutChanged = true
		} else {
			paintChanged = true
		}
	}
	return
}

func changedPaintRules(previous, next []rules.Rule) (changed []rules.Rule, removed []string) {
	previousByProperty := make(map[string]rules.Rule, len(previous))
	nextByProperty := make(map[string]rules.Rule, len(next))
	for i := range previous {
		previousByProperty[previous[i].Property] = previous[i]
	}
	for i := range next {
		nextByProperty[next[i].Property] = next[i]
	}
	for property, after := range nextByProperty {
		before, hadBefore := previousByProperty[property]
		if propertyChangeImpact(property, before, hadBefore, after, true) == StyleImpactPaint &&
			(!hadBefore || !ruleComputedEqual(before, after)) {
			changed = append(changed, after.Clone())
		}
	}
	for property := range previousByProperty {
		if _, stillPresent := nextByProperty[property]; stillPresent {
			continue
		}
		if p, ok := LinkedPropertyMap[property]; ok && p.StyleImpact() == StyleImpactPaint {
			removed = append(removed, property)
		}
	}
	if len(removed) > 0 {
		changed = changed[:0]
		for property, after := range nextByProperty {
			if p, ok := LinkedPropertyMap[property]; ok && p.StyleImpact() == StyleImpactPaint {
				changed = append(changed, after.Clone())
			}
		}
	}
	slices.SortStableFunc(changed, func(x, y rules.Rule) int { return x.Sort - y.Sort })
	return
}

func propertyChangeImpact(property string, before rules.Rule, hadBefore bool, after rules.Rule, hasAfter bool) StyleImpact {
	p, ok := LinkedPropertyMap[property]
	if !ok {
		return StyleImpactLayout
	}
	if hadBefore && hasAfter && isBorderShorthand(property) &&
		propertyValuesEqual(borderLayoutValues(before.Values), borderLayoutValues(after.Values)) {
		return StyleImpactPaint
	}
	return p.StyleImpact()
}

func isBorderShorthand(property string) bool {
	switch property {
	case "border", "border-left", "border-top", "border-right", "border-bottom":
		return true
	default:
		return false
	}
}

// The implemented border shorthands only let width tokens affect layout.
// Comparing that normalized subset lets `border: 1px solid red` transition to
// `border: 1px solid blue` through the paint-only path.
func borderLayoutValues(values []rules.PropertyValue) []rules.PropertyValue {
	out := make([]rules.PropertyValue, 0, len(values))
	for i := range values {
		v := values[i]
		switch v.Str {
		case "thin", "medium", "thick", "initial", "inherit":
			out = append(out, v)
		default:
			if strings.HasSuffix(v.Str, "px") {
				out = append(out, v)
			}
		}
	}
	return out
}

func propertyValuesEqual(a, b []rules.PropertyValue) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !propertyValueEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func ruleComputedEqual(a, b rules.Rule) bool {
	if a.Property != b.Property || len(a.Values) != len(b.Values) {
		return false
	}
	for i := range a.Values {
		if !propertyValueEqual(a.Values[i], b.Values[i]) {
			return false
		}
	}
	return true
}

func propertyValueEqual(a, b rules.PropertyValue) bool {
	return a.Str == b.Str && a.Num == b.Num &&
		slices.Equal(a.Args, b.Args) && slices.Equal(a.ArgNums, b.ArgNums)
}

func (s *ElementLayoutStylizer) clone(newElm *Element) ElementLayoutStylizer {
	out := ElementLayoutStylizer{
		element: weak.Make(newElm),
	}
	out.ReplaceRules(s.styleRules)
	return out
}

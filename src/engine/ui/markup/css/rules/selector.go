/******************************************************************************/
/* selector.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rules

type RuleState = int

const (
	ReadingTag = iota
	ReadingId
	ReadingClass
	ReadingDescendant
	ReadingChild
	ReadingSibling
	ReadingAdjacent
	ReadingCondition
	ReadingConditionAssignment
	ReadingPseudo
	ReadingPseudoFunction
	ReadingProperty
	ReadingPropertyValue
	ReadingPropertyFunction
)

type SelectorPart struct {
	Name       string
	Args       []string
	SelectType RuleState
}

type Selector struct {
	Parts []SelectorPart
}

type MediaQuery struct {
	Key   string
	Value string
}

type SelectorGroup struct {
	Selectors  []Selector
	Rules      []Rule
	MediaQuery MediaQuery
}

func (m *MediaQuery) IsValid() bool { return m.Key != "" }

func (m *MediaQuery) Clear() {
	m.Key = ""
	m.Value = ""
}

func (s *SelectorGroup) AddRule(r Rule) {
	s.Rules = append(s.Rules, r)
}

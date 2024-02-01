package rules

type RuleState = int

const (
	ReadingTag = iota
	ReadingId
	ReadingClass
	ReadingChild
	ReadingSibling
	ReadingAdjacent
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

type SelectorGroup struct {
	Selectors []Selector
	Rules     []Rule
}

func (s *SelectorGroup) AddRule(r Rule) {
	s.Rules = append(s.Rules, r)
}

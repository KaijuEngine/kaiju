package rules

import (
	"bytes"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

type StyleSheet struct {
	Groups     []SelectorGroup
	CustomVars map[string][]string
	state      RuleState
}

func (s *StyleSheet) addGroup() {
	s.Groups = append(s.Groups, SelectorGroup{
		Selectors: make([]Selector, 0),
		Rules:     make([]Rule, 0),
	})
}

func (s *StyleSheet) removeLastGroup() {
	s.Groups = s.Groups[:len(s.Groups)-1]
}

func (s *StyleSheet) currentGroup() *SelectorGroup {
	return &s.Groups[len(s.Groups)-1]
}

func (s *StyleSheet) readSelector(cssParser *css.Parser) {
	sel := Selector{
		Parts: make([]SelectorPart, 0),
	}
	for _, val := range cssParser.Values() {
		switch val.TokenType {
		case css.IdentToken:
			fallthrough
		case css.NumberToken:
			if s.state == ReadingPseudoFunction {
				idx := len(sel.Parts) - 1
				sel.Parts[idx].Args = append(sel.Parts[idx].Args, string(val.Data))
			} else {
				sel.Parts = append(sel.Parts, SelectorPart{
					Name:       string(val.Data),
					SelectType: s.state,
				})
			}
		case css.HashToken:
			id := strings.TrimPrefix(string(val.Data), "#")
			sel.Parts = append(sel.Parts, SelectorPart{
				Name:       id,
				SelectType: ReadingId,
			})
		case css.ColonToken:
			s.state = ReadingPseudo
		case css.FunctionToken:
			s.state = ReadingPseudoFunction
			sel.Parts = append(sel.Parts, SelectorPart{
				Name:       strings.TrimSuffix(string(val.Data), "("),
				SelectType: ReadingId,
			})
		case css.RightParenthesisToken:
			s.state = ReadingPseudo
		case css.DelimToken:
			switch string(val.Data) {
			case "#":
				s.state = ReadingId
			case ".":
				s.state = ReadingClass
			case ">":
				s.state = ReadingChild
			case "~":
				s.state = ReadingSibling
			case "+":
				s.state = ReadingAdjacent
			case ":":
				s.state = ReadingPseudo
			}
		}
	}
	idx := len(s.Groups) - 1
	s.Groups[idx].Selectors = append(s.Groups[idx].Selectors, sel)
}

func (s *StyleSheet) readProperty(prop string, cssParser *css.Parser) {
	r := Rule{
		Property: prop,
		Values:   make([]PropertyValue, 0),
	}
	for _, val := range cssParser.Values() {
		switch val.TokenType {
		case css.FunctionToken:
			s.state = ReadingPropertyFunction
			r.Values = append(r.Values, PropertyValue{
				Str:  strings.TrimSuffix(string(val.Data), "("),
				Args: make([]string, 0),
			})
		case css.CommaToken:
		case css.CommentToken:
		case css.WhitespaceToken:
		case css.RightParenthesisToken:
			s.state = ReadingProperty
		default:
			if s.state == ReadingPropertyFunction {
				r.Values[len(r.Values)-1].Args = append(r.Values[len(r.Values)-1].Args, string(val.Data))
			} else {
				r.Values = append(r.Values, PropertyValue{
					Str:  string(val.Data),
					Args: make([]string, 0),
				})
			}
		}
	}
	s.currentGroup().AddRule(r)
}

func NewStyleSheet() StyleSheet {
	return StyleSheet{
		Groups:     make([]SelectorGroup, 0),
		state:      ReadingTag,
		CustomVars: make(map[string][]string),
	}
}

func (s *StyleSheet) Parse(cssStr string) {
	cssParser := css.NewParser(parse.NewInput(bytes.NewBufferString(cssStr)), false)
	exit := false
	s.addGroup()
	for !exit {
		gt, _, propData := cssParser.Next()
		switch gt {
		case css.ErrorGrammar:
			exit = true
		case css.CommentGrammar:
			// Do nothing
		case css.BeginAtRuleGrammar:
		case css.AtRuleGrammar:
		case css.EndAtRuleGrammar:
		case css.QualifiedRuleGrammar:
			if s.state < ReadingProperty {
				s.readSelector(cssParser)
			}
		case css.BeginRulesetGrammar:
			s.readSelector(cssParser)
			s.state = ReadingProperty
		case css.EndRulesetGrammar:
			s.state = ReadingTag
			s.addGroup()
		case css.DeclarationGrammar:
			s.readProperty(string(propData), cssParser)
		case css.TokenGrammar:
		case css.CustomPropertyGrammar:
			name := string(propData)
			vals := make([]string, 0)
			for _, val := range cssParser.Values() {
				vals = append(vals, string(val.Data))
			}
			s.CustomVars[name] = vals
		}
	}
	s.removeLastGroup()
}

func (s *StyleSheet) ParseInline(cssStr string) *SelectorGroup {
	cssParser := css.NewParser(parse.NewInput(bytes.NewBufferString(cssStr)), true)
	exit := false
	s.addGroup()
	for !exit {
		gt, _, propData := cssParser.Next()
		switch gt {
		case css.ErrorGrammar:
			exit = true
		case css.CommentGrammar:
			// Do nothing
		case css.DeclarationGrammar:
			s.readProperty(string(propData), cssParser)
		}
	}
	group := s.currentGroup()
	s.removeLastGroup()
	return group
}

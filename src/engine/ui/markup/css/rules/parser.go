/******************************************************************************/
/* parser.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rules

import (
	"bytes"
	"slices"
	"strings"

	"kaijuengine.com/engine/ui/markup/css/helpers"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

type StyleSheet struct {
	Groups         []SelectorGroup
	CustomVars     map[string][]string
	state          RuleState
	stateFuncDepth int
}

// varRefSentinel prefixes a deferred custom-property reference that is stored
// inside a PropertyValue's Str field or a function value's Args slice (e.g.
// var() inside calc()). The NUL byte cannot appear in real CSS source, so this
// marker can never collide with a genuine token.
const varRefSentinel = "\x00var:"

// makeVarRef encodes a deferred reference to the given custom property name.
func makeVarRef(name string) string { return varRefSentinel + name }

// parseVarRef returns the custom property name and true when s is a deferred
// var reference produced by makeVarRef.
func parseVarRef(s string) (string, bool) {
	if strings.HasPrefix(s, varRefSentinel) {
		return s[len(varRefSentinel):], true
	}
	return "", false
}

func (s *StyleSheet) addGroup() {
	g := SelectorGroup{
		Selectors: make([]Selector, 0),
		Rules:     make([]Rule, 0),
	}
	if len(s.Groups) > 0 {
		g.MediaQuery = s.Groups[len(s.Groups)-1].MediaQuery
	}
	s.Groups = append(s.Groups, g)
}

func (s *StyleSheet) setGroupMediaQuery(mediaQuery MediaQuery) {
	s.Groups[len(s.Groups)-1].MediaQuery = mediaQuery
}

func (s *StyleSheet) clearGroupMediaQuery() {
	s.Groups[len(s.Groups)-1].MediaQuery.Clear()
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
	appendCombinator := func(selectType RuleState, name string) {
		if len(sel.Parts) == 0 {
			return
		}
		idx := len(sel.Parts) - 1
		switch sel.Parts[idx].SelectType {
		case ReadingDescendant, ReadingChild, ReadingSibling, ReadingAdjacent:
			sel.Parts[idx] = SelectorPart{
				Name:       name,
				SelectType: selectType,
			}
		default:
			sel.Parts = append(sel.Parts, SelectorPart{
				Name:       name,
				SelectType: selectType,
			})
		}
	}
	pseudoFunctionDepth := 0
	appendPseudoArg := func(data string) bool {
		if pseudoFunctionDepth == 0 {
			return false
		}
		idx := len(sel.Parts) - 1
		sel.Parts[idx].Args = append(sel.Parts[idx].Args, data)
		return true
	}
	for _, val := range cssParser.Values() {
		switch val.TokenType {
		case css.IdentToken:
			fallthrough
		case css.StringToken:
			fallthrough
		case css.NumberToken:
			if appendPseudoArg(string(val.Data)) {
			} else {
				d := string(val.Data)
				if s.state == ReadingConditionAssignment {
					d = strings.Trim(d, `"`)
				}
				sel.Parts = append(sel.Parts, SelectorPart{
					Name:       d,
					SelectType: s.state,
				})
			}
		case css.HashToken:
			id := strings.TrimPrefix(string(val.Data), "#")
			if appendPseudoArg("#" + id) {
			} else {
				sel.Parts = append(sel.Parts, SelectorPart{
					Name:       id,
					SelectType: ReadingId,
				})
			}
		case css.ColonToken:
			if appendPseudoArg(":") {
			} else {
				s.state = ReadingPseudo
			}
		case css.FunctionToken:
			name := strings.TrimSuffix(string(val.Data), "(")
			if appendPseudoArg(name + "(") {
				pseudoFunctionDepth++
			} else {
				s.state = ReadingPseudoFunction
				pseudoFunctionDepth = 1
				sel.Parts = append(sel.Parts, SelectorPart{
					Name:       name,
					SelectType: ReadingPseudoFunction,
				})
			}
		case css.RightParenthesisToken:
			if pseudoFunctionDepth > 1 {
				appendPseudoArg(")")
				pseudoFunctionDepth--
			} else {
				pseudoFunctionDepth = 0
				s.state = ReadingPseudo
			}
		case css.CommaToken:
			appendPseudoArg(",")
		case css.WhitespaceToken:
			if appendPseudoArg(" ") {
			} else if pseudoFunctionDepth == 0 {
				appendCombinator(ReadingDescendant, " ")
				s.state = ReadingTag
			}
		case css.LeftBracketToken:
			if appendPseudoArg("[") {
			} else {
				s.state = ReadingCondition
			}
		case css.RightBracketToken:
			if appendPseudoArg("]") {
			} else {
				s.state = ReadingTag
			}
		case css.DelimToken:
			delim := string(val.Data)
			if appendPseudoArg(delim) {
			} else {
				switch delim {
				case "#":
					s.state = ReadingId
				case ".":
					s.state = ReadingClass
				case ">":
					appendCombinator(ReadingChild, ">")
					s.state = ReadingTag
				case "~":
					s.state = ReadingSibling
				case "+":
					s.state = ReadingAdjacent
				case ":":
					s.state = ReadingPseudo
				case "=":
					if s.state == ReadingCondition {
						s.state = ReadingConditionAssignment
					}
				}
			}
		}
	}
	idx := len(s.Groups) - 1
	s.Groups[idx].Selectors = append(s.Groups[idx].Selectors, sel)
}

func (s *StyleSheet) readProperty(prop string, cssParser *css.Parser, _ helpers.WindowDimensions) {
	r := Rule{
		Property: prop,
		Values:   make([]PropertyValue, 0),
	}
	for _, val := range cssParser.Values() {
		switch val.TokenType {
		case css.FunctionToken:
			s.stateFuncDepth++
			s.state = ReadingPropertyFunction
			r.Values = append(r.Values, PropertyValue{
				Str: strings.TrimSuffix(string(val.Data), "("),
			})
		case css.CommaToken:
		case css.CommentToken:
		case css.WhitespaceToken:
		case css.RightParenthesisToken:
			s.stateFuncDepth = max(0, s.stateFuncDepth-1)
			if s.stateFuncDepth == 0 {
				s.state = ReadingProperty
			}
		default:
			if s.state == ReadingPropertyFunction {
				last := &r.Values[len(r.Values)-1]
				str := string(val.Data)
				if last.Str == "var" {
					// Drop the placeholder "var" function value and record a
					// deferred reference to the custom property. Resolution is
					// performed after the whole sheet is parsed so the final
					// (last-:root-wins) value of every custom property is used,
					// matching CSS computed-value semantics.
					r.Values = r.Values[0 : len(r.Values)-1]
					if s.stateFuncDepth > 1 && len(r.Values) > 0 {
						// var() nested inside another function (e.g. calc()):
						// record the deferred reference in the enclosing
						// function's argument list, preserving argument order.
						last = &r.Values[len(r.Values)-1]
						last.Args = append(last.Args, makeVarRef(str))
					} else {
						// Top-level var(): record a deferred placeholder value
						// that will expand into zero or more values later.
						r.Values = append(r.Values, PropertyValue{
							Str: makeVarRef(str),
						})
					}
				} else {
					last.Args = append(last.Args, str)
				}
			} else {
				r.Values = append(r.Values, PropertyValue{
					Str: string(val.Data),
				})
			}
		}
	}
	// Numeric resolution is intentionally deferred to resolveRuleVars (called
	// post-parse) because deferred var references are not yet substituted here.
	s.currentGroup().AddRule(r)
}

// resolveVars walks every parsed rule in the sheet and substitutes the final
// value of each deferred custom-property reference, then computes numeric forms.
// It must be called once after the entire sheet has been parsed, when
// s.CustomVars holds the last-wins value for every custom property.
func (s *StyleSheet) resolveVars(window helpers.WindowDimensions) {
	for gi := range s.Groups {
		g := &s.Groups[gi]
		for ri := range g.Rules {
			s.resolveRuleVars(&g.Rules[ri], window)
		}
	}
}

// resolveRuleVars substitutes deferred var references in a single rule and then
// computes the numeric forms of every value. An unknown custom property resolves
// to nothing, preserving the previous eager behavior.
func (s *StyleSheet) resolveRuleVars(r *Rule, window helpers.WindowDimensions) {
	// Expand top-level deferred var placeholders. A custom property can expand
	// to multiple tokens, so the value slice is rebuilt.
	resolved := make([]PropertyValue, 0, len(r.Values))
	for i := range r.Values {
		v := r.Values[i]
		if name, ok := parseVarRef(v.Str); ok {
			for _, sub := range s.CustomVars[name] {
				resolved = append(resolved, PropertyValue{Str: sub})
			}
			continue
		}
		// Expand deferred var references stored inside function arguments
		// (e.g. var() nested in calc()), preserving argument order.
		if len(v.Args) > 0 {
			args := make([]string, 0, len(v.Args))
			for _, a := range v.Args {
				if name, ok := parseVarRef(a); ok {
					args = append(args, s.CustomVars[name]...)
					continue
				}
				args = append(args, a)
			}
			v.Args = args
		}
		resolved = append(resolved, v)
	}
	r.Values = resolved

	for i := range r.Values {
		v := &r.Values[i]
		if len(v.Args) > 0 {
			v.ArgNums = make([]float32, len(v.Args))
			for j := range v.Args {
				v.ArgNums[j] = helpers.NumFromLength(v.Args[j], window)
			}
		} else {
			v.Num = helpers.NumFromLength(v.Str, window)
		}
	}
}

func NewStyleSheet() StyleSheet {
	return StyleSheet{
		Groups:     make([]SelectorGroup, 0),
		state:      ReadingTag,
		CustomVars: make(map[string][]string),
	}
}

func (s *StyleSheet) Parse(cssStr string, window helpers.WindowDimensions) {
	cssParser := css.NewParser(parse.NewInput(bytes.NewBufferString(cssStr)), false)
	exit := false
	s.addGroup()
	qualifiedGroupStart := -1
	for !exit {
		gt, _, propData := cssParser.Next()
		switch gt {
		case css.ErrorGrammar:
			exit = true
		case css.CommentGrammar:
			// Do nothing
		case css.BeginAtRuleGrammar:
			q := MediaQuery{}
			for _, val := range cssParser.Values() {
				if val.TokenType == css.WhitespaceToken {
					continue
				}
				v := string(val.Data)
				switch v {
				case "(", ":", ")":
				default:
					if q.Key == "" {
						q.Key = v
					} else {
						q.Value = v
					}
				}
			}
			s.setGroupMediaQuery(q)
		case css.AtRuleGrammar:
		case css.QualifiedRuleGrammar:
			if qualifiedGroupStart < 0 {
				qualifiedGroupStart = len(s.Groups) - 1
			}
			if s.state < ReadingProperty {
				s.readSelector(cssParser)
			}
			s.addGroup()
		case css.BeginRulesetGrammar:
			s.readSelector(cssParser)
			s.state = ReadingProperty
		case css.EndAtRuleGrammar:
			s.state = ReadingTag
			s.addGroup()
			s.clearGroupMediaQuery()
		case css.EndRulesetGrammar:
			s.state = ReadingTag
			if qualifiedGroupStart >= 0 {
				last := &s.Groups[len(s.Groups)-1]
				for i := len(s.Groups) - 2; i >= qualifiedGroupStart; i-- {
					s.Groups[i].Rules = append(s.Groups[i].Rules, slices.Clone(last.Rules)...)
					s.Groups[i].Selectors = append(s.Groups[i].Selectors, slices.Clone(last.Selectors)...)
					// Likely don't need to copy the media query, would be weird
					// s.Groups[i].MediaQuery = last.MediaQuery
				}
				qualifiedGroupStart = -1
			}
			s.addGroup()
		case css.DeclarationGrammar:
			s.readProperty(string(propData), cssParser, window)
		case css.TokenGrammar:
		case css.CustomPropertyGrammar:
			name := string(propData)
			vals := make([]string, 0)
			for _, val := range cssParser.Values() {
				vals = append(vals, strings.TrimSpace(string(val.Data)))
			}
			s.CustomVars[name] = vals
		}
	}
	// All custom properties are now known with their final (last :root wins)
	// values; substitute deferred var references and compute numeric forms.
	s.resolveVars(window)
	s.removeLastGroup()
}

func (s *StyleSheet) ParseInline(cssStr string, window helpers.WindowDimensions) *SelectorGroup {
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
			s.readProperty(string(propData), cssParser, window)
		}
	}
	// Resolve any deferred var references using whatever custom properties are
	// known on this style sheet (e.g. those declared by an already-parsed
	// :root block) and compute numeric forms.
	group := s.currentGroup()
	for ri := range group.Rules {
		s.resolveRuleVars(&group.Rules[ri], window)
	}
	s.removeLastGroup()
	return group
}

/******************************************************************************/
/* rule.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rules

import (
	"slices"

	"kaijuengine.com/matrix"
)

type RuleInvoke uint16

const (
	RuleInvokeImmediate RuleInvoke = 0
	RuleInvokeHover     RuleInvoke = 1 << (iota - 1)
	RuleInvokeActive
	RuleInvokeFocus
	RuleInvokeVisited
	RuleInvokeInvalid
	RuleInvokeValid
)

func (r RuleInvoke) Matches(state RuleInvoke) bool {
	return r == RuleInvokeImmediate || state&r == r
}

func (r RuleInvoke) With(state RuleInvoke) RuleInvoke {
	return r | state
}

type PropertyValue struct {
	Str     string
	Num     matrix.Float
	Args    []string
	ArgNums []matrix.Float
}

func (p *PropertyValue) Clone() PropertyValue {
	return PropertyValue{
		Str:     p.Str,
		Num:     p.Num,
		Args:    slices.Clone(p.Args),
		ArgNums: slices.Clone(p.ArgNums),
	}
}

func (p PropertyValue) IsFunction() bool {
	return len(p.Args) > 0
}

type Rule struct {
	Property     string
	Values       []PropertyValue
	Invocation   RuleInvoke
	Sort         int
	SelfDestruct bool
}

func (r *Rule) Clone() Rule {
	out := Rule{
		Property:     r.Property,
		Invocation:   r.Invocation,
		Sort:         r.Sort,
		SelfDestruct: r.SelfDestruct,
		Values:       make([]PropertyValue, len(r.Values)),
	}
	for i := range r.Values {
		out.Values[i] = r.Values[i].Clone()
	}
	return out
}

func CloneRules(in []Rule) []Rule {
	out := make([]Rule, len(in))
	for i := range in {
		out[i] = in[i].Clone()
	}
	return out
}

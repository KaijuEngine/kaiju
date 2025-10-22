/******************************************************************************/
/* rule.go                                                                    */
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

package rules

import "slices"

type RuleInvoke = int

const (
	RuleInvokeImmediate RuleInvoke = iota
	RuleInvokeHover
	RuleInvokeActive
)

type PropertyValue struct {
	Str     string
	Num     float32
	Args    []string
	ArgNums []float32
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
		Property:   r.Property,
		Invocation: r.Invocation,
		Values:     make([]PropertyValue, len(r.Values)),
	}
	for i := range r.Values {
		out.Values[i] = r.Values[i].Clone()
	}
	return out
}

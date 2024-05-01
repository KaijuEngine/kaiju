/******************************************************************************/
/* css_pseudo.go                                                              */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package pseudos

import (
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
)

type Pseudo interface {
	Key() string
	IsFunction() bool
	Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error)
	AlterRules(rules []rules.Rule) []rules.Rule
}

var PseudoMap = map[string]Pseudo{
	"active":             Active{},
	"any-link":           AnyLink{},
	"autofill":           Autofill{},
	"blank":              Blank{},
	"checked":            Checked{},
	"current":            Current{},
	"default":            Default{},
	"defined":            Defined{},
	"dir":                Dir{},
	"disabled":           Disabled{},
	"empty":              Empty{},
	"enabled":            Enabled{},
	"first":              First{},
	"first-child":        FirstChild{},
	"first-of-type":      FirstOfType{},
	"fullscreen":         Fullscreen{},
	"future":             Future{},
	"focus":              Focus{},
	"focus-visible":      FocusVisible{},
	"focus-within":       FocusWithin{},
	"has":                Has{},
	"host":               Host{},
	"host-context":       HostContext{},
	"hover":              Hover{},
	"indeterminate":      Indeterminate{},
	"in-range":           InRange{},
	"invalid":            Invalid{},
	"is":                 Is{},
	"lang":               Lang{},
	"last-child":         LastChild{},
	"last-of-type":       LastOfType{},
	"left":               Left{},
	"link":               Link{},
	"local-link":         LocalLink{},
	"modal":              Modal{},
	"not":                Not{},
	"nth-child":          NthChild{},
	"nth-col":            NthCol{},
	"nth-last-child":     NthLastChild{},
	"nth-last-col":       NthLastCol{},
	"nth-last-of-type":   NthLastOfType{},
	"nth-of-type":        NthOfType{},
	"only-child":         OnlyChild{},
	"only-of-type":       OnlyOfType{},
	"optional":           Optional{},
	"out-of-range":       OutOfRange{},
	"past":               Past{},
	"picture-in-picture": PictureInPicture{},
	"placeholder-shown":  PlaceholderShown{},
	"paused":             Paused{},
	"playing":            Playing{},
	"read-only":          ReadOnly{},
	"read-write":         ReadWrite{},
	"required":           Required{},
	"right":              Right{},
	"root":               Root{},
	"scope":              Scope{},
	"state":              State{},
	"target":             Target{},
	"target-within":      TargetWithin{},
	"user-invalid":       UserInvalid{},
	"valid":              Valid{},
	"visited":            Visited{},
	"where":              Where{},
}

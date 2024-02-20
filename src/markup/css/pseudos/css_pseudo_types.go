/******************************************************************************/
/* css_pseudo_types.go                                                        */
/******************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/******************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/******************************************************************************/

package pseudos

import "kaiju/markup/css/rules"

// https://developer.mozilla.org/en-US/docs/Web/CSS/:active
type Active struct{}

func (p Active) Key() string                                { return "active" }
func (p Active) IsFunction() bool                           { return false }
func (p Active) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:any-link
type AnyLink struct{}

func (p AnyLink) Key() string                                { return "any-link" }
func (p AnyLink) IsFunction() bool                           { return false }
func (p AnyLink) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:autofill
type Autofill struct{}

func (p Autofill) Key() string                                { return "autofill" }
func (p Autofill) IsFunction() bool                           { return false }
func (p Autofill) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:blank
type Blank struct{}

func (p Blank) Key() string                                { return "blank" }
func (p Blank) IsFunction() bool                           { return false }
func (p Blank) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:checked
type Checked struct{}

func (p Checked) Key() string                                { return "checked" }
func (p Checked) IsFunction() bool                           { return false }
func (p Checked) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:current
type Current struct{}

func (p Current) Key() string                                { return "current" }
func (p Current) IsFunction() bool                           { return false }
func (p Current) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:default
type Default struct{}

func (p Default) Key() string                                { return "default" }
func (p Default) IsFunction() bool                           { return false }
func (p Default) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:defined
type Defined struct{}

func (p Defined) Key() string                                { return "defined" }
func (p Defined) IsFunction() bool                           { return false }
func (p Defined) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:dir
type Dir struct{}

func (p Dir) Key() string                                { return "dir" }
func (p Dir) IsFunction() bool                           { return true }
func (p Dir) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:disabled
type Disabled struct{}

func (p Disabled) Key() string                                { return "disabled" }
func (p Disabled) IsFunction() bool                           { return false }
func (p Disabled) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:empty
type Empty struct{}

func (p Empty) Key() string                                { return "empty" }
func (p Empty) IsFunction() bool                           { return false }
func (p Empty) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:enabled
type Enabled struct{}

func (p Enabled) Key() string                                { return "enabled" }
func (p Enabled) IsFunction() bool                           { return false }
func (p Enabled) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:first
type First struct{}

func (p First) Key() string                                { return "first" }
func (p First) IsFunction() bool                           { return false }
func (p First) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:first-child
type FirstChild struct{}

func (p FirstChild) Key() string                                { return "first-child" }
func (p FirstChild) IsFunction() bool                           { return false }
func (p FirstChild) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:first-of-type
type FirstOfType struct{}

func (p FirstOfType) Key() string                                { return "first-of-type" }
func (p FirstOfType) IsFunction() bool                           { return false }
func (p FirstOfType) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:fullscreen
type Fullscreen struct{}

func (p Fullscreen) Key() string                                { return "fullscreen" }
func (p Fullscreen) IsFunction() bool                           { return false }
func (p Fullscreen) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:future
type Future struct{}

func (p Future) Key() string                                { return "future" }
func (p Future) IsFunction() bool                           { return false }
func (p Future) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:focus
type Focus struct{}

func (p Focus) Key() string                                { return "focus" }
func (p Focus) IsFunction() bool                           { return false }
func (p Focus) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:focus-visible
type FocusVisible struct{}

func (p FocusVisible) Key() string                                { return "focus-visible" }
func (p FocusVisible) IsFunction() bool                           { return false }
func (p FocusVisible) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:focus-within
type FocusWithin struct{}

func (p FocusWithin) Key() string                                { return "focus-within" }
func (p FocusWithin) IsFunction() bool                           { return false }
func (p FocusWithin) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:has
type Has struct{}

func (p Has) Key() string                                { return "has" }
func (p Has) IsFunction() bool                           { return true }
func (p Has) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:host
type Host struct{}

func (p Host) Key() string                                { return "host" }
func (p Host) IsFunction() bool                           { return true }
func (p Host) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:host-context
type HostContext struct{}

func (p HostContext) Key() string                                { return "host-context" }
func (p HostContext) IsFunction() bool                           { return true }
func (p HostContext) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:hover
type Hover struct{}

func (p Hover) Key() string      { return "hover" }
func (p Hover) IsFunction() bool { return false }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:indeterminate
type Indeterminate struct{}

func (p Indeterminate) Key() string                                { return "indeterminate" }
func (p Indeterminate) IsFunction() bool                           { return false }
func (p Indeterminate) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:in-range
type InRange struct{}

func (p InRange) Key() string                                { return "in-range" }
func (p InRange) IsFunction() bool                           { return false }
func (p InRange) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:invalid
type Invalid struct{}

func (p Invalid) Key() string                                { return "invalid" }
func (p Invalid) IsFunction() bool                           { return false }
func (p Invalid) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:is
type Is struct{}

func (p Is) Key() string                                { return "is" }
func (p Is) IsFunction() bool                           { return true }
func (p Is) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:lang
type Lang struct{}

func (p Lang) Key() string                                { return "lang" }
func (p Lang) IsFunction() bool                           { return true }
func (p Lang) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:last-child
type LastChild struct{}

func (p LastChild) Key() string                                { return "last-child" }
func (p LastChild) IsFunction() bool                           { return false }
func (p LastChild) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:last-of-type
type LastOfType struct{}

func (p LastOfType) Key() string                                { return "last-of-type" }
func (p LastOfType) IsFunction() bool                           { return false }
func (p LastOfType) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:left
type Left struct{}

func (p Left) Key() string                                { return "left" }
func (p Left) IsFunction() bool                           { return false }
func (p Left) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:link
type Link struct{}

func (p Link) Key() string                                { return "link" }
func (p Link) IsFunction() bool                           { return false }
func (p Link) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:local-link
type LocalLink struct{}

func (p LocalLink) Key() string                                { return "local-link" }
func (p LocalLink) IsFunction() bool                           { return false }
func (p LocalLink) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:modal
type Modal struct{}

func (p Modal) Key() string                                { return "modal" }
func (p Modal) IsFunction() bool                           { return false }
func (p Modal) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:not
type Not struct{}

func (p Not) Key() string                                { return "not" }
func (p Not) IsFunction() bool                           { return true }
func (p Not) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-child
type NthChild struct{}

func (p NthChild) Key() string                                { return "nth-child" }
func (p NthChild) IsFunction() bool                           { return true }
func (p NthChild) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-col
type NthCol struct{}

func (p NthCol) Key() string                                { return "nth-col" }
func (p NthCol) IsFunction() bool                           { return true }
func (p NthCol) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-last-child
type NthLastChild struct{}

func (p NthLastChild) Key() string                                { return "nth-last-child" }
func (p NthLastChild) IsFunction() bool                           { return true }
func (p NthLastChild) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-last-col
type NthLastCol struct{}

func (p NthLastCol) Key() string                                { return "nth-last-col" }
func (p NthLastCol) IsFunction() bool                           { return true }
func (p NthLastCol) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-last-of-type
type NthLastOfType struct{}

func (p NthLastOfType) Key() string                                { return "nth-last-of-type" }
func (p NthLastOfType) IsFunction() bool                           { return true }
func (p NthLastOfType) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-of-type
type NthOfType struct{}

func (p NthOfType) Key() string                                { return "nth-of-type" }
func (p NthOfType) IsFunction() bool                           { return true }
func (p NthOfType) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:only-child
type OnlyChild struct{}

func (p OnlyChild) Key() string                                { return "only-child" }
func (p OnlyChild) IsFunction() bool                           { return false }
func (p OnlyChild) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:only-of-type
type OnlyOfType struct{}

func (p OnlyOfType) Key() string                                { return "only-of-type" }
func (p OnlyOfType) IsFunction() bool                           { return false }
func (p OnlyOfType) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:optional
type Optional struct{}

func (p Optional) Key() string                                { return "optional" }
func (p Optional) IsFunction() bool                           { return false }
func (p Optional) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:out-of-range
type OutOfRange struct{}

func (p OutOfRange) Key() string                                { return "out-of-range" }
func (p OutOfRange) IsFunction() bool                           { return false }
func (p OutOfRange) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:past
type Past struct{}

func (p Past) Key() string                                { return "past" }
func (p Past) IsFunction() bool                           { return false }
func (p Past) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:picture-in-picture
type PictureInPicture struct{}

func (p PictureInPicture) Key() string                                { return "picture-in-picture" }
func (p PictureInPicture) IsFunction() bool                           { return false }
func (p PictureInPicture) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:placeholder-shown
type PlaceholderShown struct{}

func (p PlaceholderShown) Key() string                                { return "placeholder-shown" }
func (p PlaceholderShown) IsFunction() bool                           { return false }
func (p PlaceholderShown) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:paused
type Paused struct{}

func (p Paused) Key() string                                { return "paused" }
func (p Paused) IsFunction() bool                           { return false }
func (p Paused) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:playing
type Playing struct{}

func (p Playing) Key() string                                { return "playing" }
func (p Playing) IsFunction() bool                           { return false }
func (p Playing) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:read-only
type ReadOnly struct{}

func (p ReadOnly) Key() string                                { return "read-only" }
func (p ReadOnly) IsFunction() bool                           { return false }
func (p ReadOnly) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:read-write
type ReadWrite struct{}

func (p ReadWrite) Key() string                                { return "read-write" }
func (p ReadWrite) IsFunction() bool                           { return false }
func (p ReadWrite) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:required
type Required struct{}

func (p Required) Key() string                                { return "required" }
func (p Required) IsFunction() bool                           { return false }
func (p Required) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:right
type Right struct{}

func (p Right) Key() string                                { return "right" }
func (p Right) IsFunction() bool                           { return false }
func (p Right) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:root
type Root struct{}

func (p Root) Key() string                                { return "root" }
func (p Root) IsFunction() bool                           { return false }
func (p Root) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:scope
type Scope struct{}

func (p Scope) Key() string                                { return "scope" }
func (p Scope) IsFunction() bool                           { return false }
func (p Scope) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:state
type State struct{}

func (p State) Key() string                                { return "state" }
func (p State) IsFunction() bool                           { return true }
func (p State) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:target
type Target struct{}

func (p Target) Key() string                                { return "target" }
func (p Target) IsFunction() bool                           { return false }
func (p Target) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:target-within
type TargetWithin struct{}

func (p TargetWithin) Key() string                                { return "target-within" }
func (p TargetWithin) IsFunction() bool                           { return false }
func (p TargetWithin) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:user-invalid
type UserInvalid struct{}

func (p UserInvalid) Key() string                                { return "user-invalid" }
func (p UserInvalid) IsFunction() bool                           { return false }
func (p UserInvalid) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:valid
type Valid struct{}

func (p Valid) Key() string                                { return "valid" }
func (p Valid) IsFunction() bool                           { return false }
func (p Valid) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:visited
type Visited struct{}

func (p Visited) Key() string                                { return "visited" }
func (p Visited) IsFunction() bool                           { return false }
func (p Visited) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

// https://developer.mozilla.org/en-US/docs/Web/CSS/:where
type Where struct{}

func (p Where) Key() string                                { return "where" }
func (p Where) IsFunction() bool                           { return true }
func (p Where) AlterRules(rules []rules.Rule) []rules.Rule { return rules }

package pseudos

import (
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

type Pseudo interface {
	Key() string
	IsFunction() bool
	Process(elm markup.DocElement, value rules.SelectorPart) ([]markup.DocElement, error)
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

/******************************************************************************/
/* css_font.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"fmt"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

var (
	fontStyleValues = map[string]struct{}{
		"italic":  {},
		"oblique": {},
	}
	fontVariantValues = map[string]struct{}{
		"small-caps": {},
	}
	fontWeightValues = map[string]struct{}{
		"bold":    {},
		"bolder":  {},
		"lighter": {},
		"100":     {},
		"200":     {},
		"300":     {},
		"400":     {},
		"500":     {},
		"600":     {},
		"700":     {},
		"800":     {},
		"900":     {},
	}
	fontStretchValues = map[string]struct{}{
		"ultra-condensed": {},
		"extra-condensed": {},
		"condensed":       {},
		"semi-condensed":  {},
		"semi-expanded":   {},
		"expanded":        {},
		"extra-expanded":  {},
		"ultra-expanded":  {},
	}
	fontSizeKeywords = map[string]struct{}{
		"xx-small": {},
		"x-small":  {},
		"small":    {},
		"medium":   {},
		"large":    {},
		"x-large":  {},
		"xx-large": {},
		"smaller":  {},
		"larger":   {},
	}
	fontSizeKeywordLengths = map[string]rules.PropertyValue{
		"xx-small": {Str: "9px", Num: 9},
		"x-small":  {Str: "10px", Num: 10},
		"small":    {Str: "13px", Num: 13},
		"medium":   {Str: "14px", Num: 14},
		"large":    {Str: "18px", Num: 18},
		"x-large":  {Str: "24px", Num: 24},
		"xx-large": {Str: "32px", Num: 32},
		"smaller":  {Str: "0.8em"},
		"larger":   {Str: "1.2em"},
	}
)

type fontShorthand struct {
	style     string
	weight    string
	size      rules.PropertyValue
	line      rules.PropertyValue
	family    []rules.PropertyValue
	hasLine   bool
	hasSize   bool
	hasFamily bool
}

func isFontSizeToken(v rules.PropertyValue) bool {
	str := strings.TrimSpace(v.Str)
	if str == "" || str == "/" {
		return false
	}
	if _, ok := fontSizeKeywords[str]; ok {
		return true
	}
	if v.Num != 0 || str == "0" {
		return true
	}
	for _, suffix := range []string{
		"px", "em", "rem", "ex", "ch", "vw", "vh", "vmin", "vmax",
		"cm", "mm", "in", "pt", "pc", "%",
	} {
		if strings.HasSuffix(str, suffix) {
			return true
		}
	}
	return false
}

func isFontLineHeightToken(v rules.PropertyValue) bool {
	str := strings.TrimSpace(v.Str)
	if str == "" || str == "/" {
		return false
	}
	if str == "normal" {
		return true
	}
	return isFontSizeToken(v)
}

func isFontNormalToken(v rules.PropertyValue) bool {
	return strings.TrimSpace(v.Str) == "normal"
}

func fontSizeValue(v rules.PropertyValue) rules.PropertyValue {
	if size, ok := fontSizeKeywordLengths[strings.TrimSpace(v.Str)]; ok {
		return size
	}
	return v
}

func parseFontShorthand(values []rules.PropertyValue) (fontShorthand, error) {
	out := fontShorthand{
		style:  "normal",
		weight: "normal",
	}
	if len(values) == 0 {
		return out, errors.New("font requires values")
	}
	for i := 0; i < len(values); i++ {
		str := strings.TrimSpace(values[i].Str)
		if isFontSizeToken(values[i]) {
			out.size = fontSizeValue(values[i])
			out.hasSize = true
			i++
			if i < len(values) && strings.TrimSpace(values[i].Str) == "/" {
				i++
				if i >= len(values) || !isFontLineHeightToken(values[i]) {
					return out, errors.New("font line-height requires a value after /")
				}
				out.line = values[i]
				out.hasLine = true
				i++
			}
			if i >= len(values) {
				return out, errors.New("font requires a font-family after font-size")
			}
			out.family = append(out.family, values[i:]...)
			out.hasFamily = true
			break
		}
		if isFontNormalToken(values[i]) {
			continue
		}
		if _, ok := fontStyleValues[str]; ok {
			out.style = str
			continue
		}
		if _, ok := fontVariantValues[str]; ok {
			continue
		}
		if _, ok := fontWeightValues[str]; ok {
			out.weight = str
			continue
		}
		if _, ok := fontStretchValues[str]; ok {
			continue
		}
		return out, fmt.Errorf("unexpected font shorthand token before font-size: %s", str)
	}
	if !out.hasSize {
		return out, errors.New("font requires a font-size")
	}
	if !out.hasFamily {
		return out, errors.New("font requires a font-family")
	}
	return out, nil
}

func (p Font) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	font, err := parseFontShorthand(values)
	if err != nil {
		return err
	}
	// Apply the family first so subsequent style/weight operations select the
	// matching face variant from that family.
	if err := (FontFamily{}).Process(panel, elm, font.family, host); err != nil {
		return err
	}
	if err := (FontStyle{}).Process(panel, elm, []rules.PropertyValue{{Str: font.style}}, host); err != nil {
		return err
	}
	if err := (FontWeight{}).Process(panel, elm, []rules.PropertyValue{{Str: font.weight}}, host); err != nil {
		return err
	}
	if err := (FontSize{}).Process(panel, elm, []rules.PropertyValue{font.size}, host); err != nil {
		return err
	}
	if font.hasLine {
		if err := (LineHeight{}).Process(panel, elm, []rules.PropertyValue{font.line}, host); err != nil {
			return err
		}
	}
	return nil
}

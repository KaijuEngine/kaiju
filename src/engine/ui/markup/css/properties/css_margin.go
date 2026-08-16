/******************************************************************************/
/* css_margin.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"slices"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/windowing"
)

func marginSizeFromStr(str string, window *windowing.Window) matrix.Float {
	if val, ok := borderSizes[str]; ok {
		return val
	}
	return helpers.NumFromLength(str, window)
}

func preprocLeftTopRightBottom(values []rules.PropertyValue, ruleList []rules.Rule, propName string) ([]rules.PropertyValue, []rules.Rule) {
	for i := 1; i < len(ruleList); i++ {
		if ruleList[i].Property == propName {
			return values, ruleList[1:]
		}
	}
	return values, ruleList
}

func (Margin) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	switch len(values) {
	case 1:
		for i := range 3 {
			values = append(values, values[i])
		}
	case 2:
		values = append(values, values[0])
		values = append(values, values[1])
	case 3:
		values = append(values, values[1])
	}

	for i := 1; i < len(ruleList); i++ {
		removeRule := false
		switch ruleList[i].Property {
		case "margin-top":
			values[0] = ruleList[i].Values[0]
			removeRule = true
		case "margin-right":
			values[1] = ruleList[i].Values[0]
			removeRule = true
		case "margin-bottom":
			values[2] = ruleList[i].Values[0]
			removeRule = true
		case "margin-left":
			values[3] = ruleList[i].Values[0]
			removeRule = true
		}
		if removeRule {
			ruleList = slices.Delete(ruleList, i, i+1)
			i--
		}
	}
	ruleList[0].Values = values
	return values, ruleList
}

func (Margin) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 1 {
		if err := (MarginLeft{}).Process(panel, elm, values, host); err != nil {
			return err
		}
		if err := (MarginTop{}).Process(panel, elm, values, host); err != nil {
			return err
		}
		if err := (MarginRight{}).Process(panel, elm, values, host); err != nil {
			return err
		}
		if err := (MarginBottom{}).Process(panel, elm, values, host); err != nil {
			return err
		}
	} else if len(values) == 2 {
		if err := (MarginTop{}).Process(panel, elm, values[:1], host); err != nil {
			return err
		}
		if err := (MarginBottom{}).Process(panel, elm, values[:1], host); err != nil {
			return err
		}
		if err := (MarginLeft{}).Process(panel, elm, values[1:], host); err != nil {
			return err
		}
		if err := (MarginRight{}).Process(panel, elm, values[1:], host); err != nil {
			return err
		}
	} else if len(values) == 3 {
		if err := (MarginTop{}).Process(panel, elm, values[:1], host); err != nil {
			return err
		}
		if err := (MarginRight{}).Process(panel, elm, values[1:2], host); err != nil {
			return err
		}
		if err := (MarginLeft{}).Process(panel, elm, values[1:2], host); err != nil {
			return err
		}
		if err := (MarginBottom{}).Process(panel, elm, values[2:], host); err != nil {
			return err
		}
	} else if len(values) == 4 {
		if err := (MarginTop{}).Process(panel, elm, values[:1], host); err != nil {
			return err
		}
		if err := (MarginRight{}).Process(panel, elm, values[1:2], host); err != nil {
			return err
		}
		if err := (MarginBottom{}).Process(panel, elm, values[2:3], host); err != nil {
			return err
		}
		if err := (MarginLeft{}).Process(panel, elm, values[3:], host); err != nil {
			return err
		}
	} else {
		return errors.New("Margin requires 1-4 values")
	}
	return nil
}

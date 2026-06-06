/******************************************************************************/
/* css_sizing_constraints.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"strconv"
	"strings"

	"kaijuengine.com/engine/ui"
)

type cssSizingConstraints struct {
	MinWidth      float32
	MaxWidth      float32
	MinHeight     float32
	MaxHeight     float32
	AspectRatio   float32
	UsesBoxSizing bool
}

func (c cssSizingConstraints) HasMinWidth() bool {
	return c.MinWidth > 0
}

func (c cssSizingConstraints) HasMaxWidth() bool {
	return c.MaxWidth >= 0
}

func (c cssSizingConstraints) HasMinHeight() bool {
	return c.MinHeight > 0
}

func (c cssSizingConstraints) HasMaxHeight() bool {
	return c.MaxHeight >= 0
}

func (c cssSizingConstraints) HasAspectRatio() bool {
	return c.AspectRatio > 0
}

func (c cssSizingConstraints) HasBoxSizing() bool {
	return c.UsesBoxSizing
}

func (c cssSizingConstraints) UsesBorderBox() bool {
	return c.UsesBoxSizing
}

func currentSizingConstraints(panel *ui.Panel) cssSizingConstraints {
	return cssSizingConstraints{
		MinWidth:      panel.GetMinSize().X(),
		MaxWidth:      panel.GetMaxSize().X(),
		MinHeight:     panel.GetMinSize().Y(),
		MaxHeight:     panel.GetMaxSize().Y(),
		AspectRatio:   panel.GetAspectRatio(),
		UsesBoxSizing: panel.GetUsesBorderBox(),
	}
}

func enableMinWidth(panel *ui.Panel, v float32) {
	panel.SetMinWidth(v)
}

func disableMinWidth(panel *ui.Panel) {
	panel.SetMinWidth(0)
}

func enableMaxWidth(panel *ui.Panel, v float32) {
	panel.SetMaxWidth(v)
}

func disableMaxWidth(panel *ui.Panel) {
	panel.SetMaxWidth(-1)
}

func enableMinHeight(panel *ui.Panel, v float32) {
	panel.SetMinHeight(v)
}

func disableMinHeight(panel *ui.Panel) {
	panel.SetMinHeight(0)
}

func enableMaxHeight(panel *ui.Panel, v float32) {
	panel.SetMaxHeight(v)
}

func disableMaxHeight(panel *ui.Panel) {
	panel.SetMaxHeight(-1)
}

func enableAspectRatio(panel *ui.Panel, ratio float32) {
	panel.SetAspectRatio(ratio)
}

func disableAspectRatio(panel *ui.Panel) {
	panel.SetAspectRatio(0)
}

func enableBorderBoxSizing(panel *ui.Panel) {
	panel.SetUsesBorderBox(true)
}

func enableContentBoxSizing(panel *ui.Panel) {
	panel.SetUsesBorderBox(false)
}

func applyWidthConstraints(panel *ui.Panel, width float32) float32 {
	c := currentSizingConstraints(panel)
	if c.HasMinWidth() && width < c.MinWidth {
		return c.MinWidth
	}
	if c.HasMaxWidth() && width > c.MaxWidth {
		return c.MaxWidth
	}
	return width
}

func applyHeightConstraints(panel *ui.Panel, height float32) float32 {
	c := currentSizingConstraints(panel)
	if c.HasMinHeight() && height < c.MinHeight {
		return c.MinHeight
	}
	if c.HasMaxHeight() && height > c.MaxHeight {
		return c.MaxHeight
	}
	return height
}

func parseRatio(values []string) (float32, bool) {
	if len(values) == 1 {
		r := strings.TrimSpace(values[0])
		if r == "auto" || r == "initial" {
			return 0, false
		}
		left, right, ok := strings.Cut(r, "/")
		if ok {
			left = strings.TrimSpace(left)
			right = strings.TrimSpace(right)
			if left != "" && right != "" {
				lv := parseSimpleFloat(left)
				rv := parseSimpleFloat(right)
				if rv > 0 {
					return lv / rv, true
				}
			}
		}
	}
	if len(values) == 3 && values[1] == "/" {
		lv := parseSimpleFloat(values[0])
		rv := parseSimpleFloat(values[2])
		if rv > 0 {
			return lv / rv, true
		}
	}
	return 0, false
}

func parseSimpleFloat(v string) float32 {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	out, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return 0
	}
	return float32(out)
}
